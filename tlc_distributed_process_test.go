package tlago

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glycerine/tlago/tlc"
)

// Upstream DistributedTLCTestCase disables these models unconditionally. This
// native process check reuses EWD840's unchanged model and all test assertions;
// it does not claim translation of the disabled in-JVM/RMI harness. Each role
// needs its own process because FP64 and the tool's interning are process-wide.
func TestNativeDistributedEWD840ProcessRoles(t *testing.T) {
	for _, scenario := range []struct {
		name                                       string
		remoteFP, recovering, workerLoss, combined bool
		allWorkersLost                             bool
		workerReplyLoss                            bool
		duplicateWorkerRegistration                bool
		fingerprintReplyLoss                       bool
		fingerprintLoss                            bool
		fingerprintServers                         int
		workerThreads                              int
		midRunCheckpoint                           bool
		checkpointInterrupted                      bool
		checkpointInterruptedAfterQueue            bool
		checkpointInterruptedAfterIntern           bool
		checkpointInterruptedAfterFirstFP          bool
	}{
		{name: "coordinator_fingerprints"},
		{name: "multiple_worker_threads", workerThreads: 2},
		{name: "standalone_fingerprints", remoteFP: true},
		{name: "partitioned_fingerprints", remoteFP: true, fingerprintServers: 2},
		{name: "fingerprint_server_loss", remoteFP: true, fingerprintServers: 2, fingerprintLoss: true},
		{name: "fingerprint_put_reply_loss", remoteFP: true, fingerprintServers: 2, fingerprintLoss: true, fingerprintReplyLoss: true},
		{name: "combined_worker_fingerprints", remoteFP: true, combined: true},
		{name: "checkpoint_recovery", recovering: true},
		{name: "mid_run_checkpoint_recovery", recovering: true, midRunCheckpoint: true},
		{name: "mid_run_checkpoint_recovery_multiple_workers", recovering: true, midRunCheckpoint: true, workerThreads: 2},
		{name: "checkpoint_interruption_before_commit", recovering: true, midRunCheckpoint: true, checkpointInterrupted: true},
		{name: "checkpoint_interruption_after_queue_commit", recovering: true, midRunCheckpoint: true, checkpointInterrupted: true, checkpointInterruptedAfterQueue: true},
		{name: "checkpoint_interruption_after_intern_commit", recovering: true, midRunCheckpoint: true, checkpointInterrupted: true, checkpointInterruptedAfterIntern: true},
		{name: "checkpoint_interruption_after_first_fingerprint_commit", recovering: true, midRunCheckpoint: true, checkpointInterrupted: true, checkpointInterruptedAfterIntern: true, checkpointInterruptedAfterFirstFP: true},
		{name: "worker_loss", workerLoss: true},
		{name: "all_workers_lost", workerLoss: true, allWorkersLost: true},
		{name: "computed_worker_reply_loss", workerLoss: true, allWorkersLost: true, workerReplyLoss: true},
		{name: "duplicate_worker_registration", duplicateWorkerRegistration: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			model, err := filepath.Abs("tlc/test_vectors/models/EWD840")
			if err != nil {
				t.Fatal(err)
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			port := listener.Addr().(*net.TCPAddr).Port
			if err := listener.Close(); err != nil {
				t.Fatal(err)
			}
			// This is a watchdog, not a state/model bound. Preserve N=7 and
			// the source model/configuration without any exploration cutoff.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			common := []string{fmt.Sprintf("-Dtlc2.tool.distributed.TLCServer.port=%d", port), "-Dtlago.distributed.bindHost=127.0.0.1", "-Dtlago.distributed.advertiseHost=127.0.0.1"}
			var roles []*nativeDistributedTestProcess
			var failedFingerprint *nativeDistributedTestProcess
			var survivingFingerprint tlc.DistributedEndpointReference
			releaseWorker := filepath.Join(t.TempDir(), "resume-worker")
			roleCounts := make(map[string]int)
			start := func(role string, args ...string) *nativeDistributedTestProcess {
				commandArgs := []string{"-test.run=^TestNativeDistributedProcessHelper$", "--", role}
				commandArgs = append(commandArgs, common...)
				commandArgs = append(commandArgs, args...)
				command := exec.CommandContext(ctx, os.Args[0], commandArgs...)
				command.Dir = model
				command.Env = append(os.Environ(), "TLAGO_NATIVE_DISTRIBUTED_PROCESS_HELPER=1")
				if role == "worker-fingerprint-loss" || role == "worker-reply-loss" || role == "worker-register-twice" {
					command.Env = append(command.Env, "TLAGO_NATIVE_WORKER_RELEASE="+releaseWorker)
				}
				if scenario.fingerprintServers > 1 && (role == "fpserver" || role == "fpserver-put-reply-loss") {
					// These roles represent separate hosts. Give each private
					// temporary storage even when they start in the same millisecond.
					command.Env = append(command.Env, "TMPDIR="+t.TempDir())
				}
				roleCounts[role]++
				label := role
				if scenario.fingerprintServers > 1 && (role == "fpserver" || role == "fpserver-put-reply-loss") {
					label = fmt.Sprintf("%s-%d", role, roleCounts[role])
				}
				if scenario.midRunCheckpoint && role == "worker" {
					label = "worker-before-checkpoint"
					if roleCounts[role] > 1 {
						label = "worker-after-checkpoint"
					}
				}
				output := &nativeDistributedTestLog{test: t, role: label}
				command.Stdout, command.Stderr = output, output
				if err := command.Start(); err != nil {
					t.Fatal(err)
				}
				process := &nativeDistributedTestProcess{command: command, output: output, done: make(chan error, 1)}
				go func() { process.done <- command.Wait() }()
				roles = append(roles, process)
				t.Logf("started %s process %d", role, command.Process.Pid)
				return process
			}
			// Join every child before returning, even on a failed assertion.
			defer func() {
				cancel()
				for _, process := range roles {
					if !process.joined {
						<-process.done
					}
				}
			}()
			serverArgs := []string{"-tool", "-deadlock", "-metadir", t.TempDir(), "MC06"}
			recoveryDistinct, recoveryQueue := "16384", "16384"
			if scenario.recovering {
				var previousTraceMetadata []byte
				var previousFingerprintCheckpoints [2][]byte
				producer := "checkpoint-frontier"
				if scenario.midRunCheckpoint {
					producer = "checkpoint-mid-run"
				}
				producerArgs := append([]string{"-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet"}, serverArgs...)
				if scenario.checkpointInterrupted {
					// Establish an older complete checkpoint. The next producer
					// advances it but exits at the selected file-commit boundary.
					baseline := start("checkpoint-frontier", producerArgs...)
					err := <-baseline.done
					baseline.joined = true
					if err != nil {
						t.Fatalf("baseline checkpoint: %v\n%s", err, baseline.output.String())
					}
					path := regexp.MustCompile(`(?m)^NATIVE_CHECKPOINT_PATH=(.+)$`).FindStringSubmatch(baseline.output.String())
					if len(path) != 2 {
						t.Fatal("baseline checkpoint directory absent")
					}
					previousTraceMetadata, err = os.ReadFile(filepath.Join(path[1], "MC06.st.chkpt"))
					if err != nil {
						t.Fatal(err)
					}
					for i := range previousFingerprintCheckpoints {
						previousFingerprintCheckpoints[i], err = os.ReadFile(filepath.Join(path[1], fmt.Sprintf("MC06_%d.fp.chkpt", i)))
						if err != nil {
							t.Fatal(err)
						}
					}
					producer = "checkpoint-interrupt"
					if scenario.checkpointInterruptedAfterQueue {
						producer = "checkpoint-interrupt-after-queue"
					}
					if scenario.checkpointInterruptedAfterIntern {
						producer = "checkpoint-interrupt-after-intern"
					}
					if scenario.checkpointInterruptedAfterFirstFP {
						producer = "checkpoint-interrupt-after-first-fingerprint"
					}
					producerArgs = []string{"-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "-tool", "-deadlock", "-recover", path[1], "MC06"}
				}
				snapshot := start(producer, producerArgs...)
				var snapshotWorker *nativeDistributedTestProcess
				if scenario.midRunCheckpoint {
					snapshotWorker = start("worker", fmt.Sprintf("-Dtlc2.tool.distributed.TLCWorker.threadCount=%d", max(1, scenario.workerThreads)), "127.0.0.1")
				}
				err := <-snapshot.done
				snapshot.joined = true
				if err != nil {
					t.Fatalf("checkpoint producer: %v\n%s", err, snapshot.output.String())
				}
				path := regexp.MustCompile(`(?m)^NATIVE_CHECKPOINT_PATH=(.+)$`).FindStringSubmatch(snapshot.output.String())
				if len(path) != 2 {
					t.Fatal("checkpoint producer did not publish its checkpoint directory")
				}
				if snapshotWorker != nil {
					// The producer deliberately exits at its checkpoint boundary,
					// without finishing exploration. Retire its old worker before restart.
					if err := snapshotWorker.command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
						t.Fatal(err)
					}
					<-snapshotWorker.done
					snapshotWorker.joined = true
					countMarker := "NATIVE_CHECKPOINT_COUNTS"
					if scenario.checkpointInterrupted {
						countMarker = "NATIVE_INTERRUPTED_CHECKPOINT_COUNTS"
					}
					counts := regexp.MustCompile(`(?m)^` + countMarker + `=(\d+),(\d+)$`).FindStringSubmatch(snapshot.output.String())
					if len(counts) != 3 {
						t.Fatal("mid-run checkpoint counts absent")
					}
					distinct, _ := strconv.ParseUint(counts[1], 10, 64)
					queued, _ := strconv.ParseUint(counts[2], 10, 64)
					if distinct <= 16384 || distinct >= 114942 || queued == 0 || queued >= distinct {
						t.Fatalf("checkpoint was not a partially explored frontier: %d/%d", distinct, queued)
					}
					completedCheckpoints := 1
					if scenario.checkpointInterrupted {
						completedCheckpoints = 0
						// The default nested MultiFPSet rebuilds fingerprints from
						// the complete persisted trace, not the old FP checkpoint.
						// Trace.recover seeks to the saved cursor without truncation.
						trace := tlc.NewTLCTrace(path[1], "MC06")
						elements, err := trace.Elements()
						if err != nil {
							t.Fatal(err)
						}
						seen := make(map[uint64]bool)
						for {
							pos, err := elements.NextPos()
							if err != nil {
								t.Fatal(err)
							}
							if pos == -1 {
								break
							}
							fp, err := elements.NextFP()
							if err != nil {
								t.Fatal(err)
							}
							if seen[fp] {
								t.Fatal("persisted interrupted trace contains duplicate fingerprints")
							}
							seen[fp] = true
						}
						if err := elements.Close(); err != nil {
							t.Fatal(err)
						}
						if err := trace.Close(); err != nil {
							t.Fatal(err)
						}
						if uint64(len(seen)) < distinct {
							t.Fatal("interrupted trace lost flushed successor records")
						}
						if scenario.checkpointInterruptedAfterQueue || scenario.checkpointInterruptedAfterIntern {
							recoveryQueue = counts[2]
						}
						// Inspect committed queue and unchanged old trace metadata
						// independently of the producer's live-count markers.
						checkpoint, err := os.Open(filepath.Join(path[1], "queue.chkpt"))
						if err != nil {
							t.Fatal(err)
						}
						input, err := tlc.NewValueInputStreamWithGlobalCompression(checkpoint)
						if err != nil {
							_ = checkpoint.Close()
							t.Fatal(err)
						}
						committedQueue, readErr := input.ReadInt()
						closeErr := input.Close()
						if readErr != nil || closeErr != nil || strconv.Itoa(int(committedQueue)) != recoveryQueue {
							t.Fatalf("committed queue count = %d/%v/%v, want %s", committedQueue, readErr, closeErr, recoveryQueue)
						}
						metadata, err := os.ReadFile(filepath.Join(path[1], "MC06.st.chkpt"))
						if err != nil {
							t.Fatal(err)
						}
						if scenario.checkpointInterruptedAfterIntern {
							if bytes.Equal(metadata, previousTraceMetadata) {
								t.Fatal("trace checkpoint was not committed before interruption")
							}
							for _, name := range []string{"MC06.st.tmp", "vars.tmp"} {
								if _, err := os.Stat(filepath.Join(path[1], name)); !os.IsNotExist(err) {
									t.Fatalf("committed temporary %s still present: %v", name, err)
								}
							}
							if _, err := os.Stat(filepath.Join(path[1], "vars.chkpt")); err != nil {
								t.Fatal("intern checkpoint absent after its commit")
							}
						} else {
							if !bytes.Equal(metadata, previousTraceMetadata) {
								t.Fatal("interrupted producer changed trace checkpoint metadata")
							}
							for _, name := range []string{"MC06.st.tmp", "vars.tmp"} {
								if _, err := os.Stat(filepath.Join(path[1], name)); err != nil {
									t.Fatalf("uncommitted temporary %s: %v", name, err)
								}
							}
						}
						for i, previous := range previousFingerprintCheckpoints {
							name := fmt.Sprintf("MC06_%d.fp", i)
							committed, err := os.ReadFile(filepath.Join(path[1], name+".chkpt"))
							if err != nil {
								t.Fatal(err)
							}
							_, temporaryErr := os.Stat(filepath.Join(path[1], name+".tmp"))
							if scenario.checkpointInterruptedAfterFirstFP && i == 0 {
								if !os.IsNotExist(temporaryErr) {
									t.Fatal("first fingerprint temporary was not promoted")
								}
								if bytes.Equal(committed, previous) {
									t.Fatal("first fingerprint checkpoint did not change after exploration")
								}
							} else {
								if temporaryErr != nil {
									t.Fatalf("uncommitted fingerprint temporary %s: %v", name, temporaryErr)
								}
								if !bytes.Equal(committed, previous) {
									t.Fatalf("later fingerprint checkpoint %s was changed", name)
								}
							}
						}
						_, temporaryErr := os.Stat(filepath.Join(path[1], "queue.tmp"))
						if scenario.checkpointInterruptedAfterQueue || scenario.checkpointInterruptedAfterIntern {
							if !os.IsNotExist(temporaryErr) {
								t.Fatal("queue temporary was not promoted before interruption")
							}
						} else if temporaryErr != nil {
							t.Fatal("pre-commit queue temporary absent")
						}
						recoveryDistinct = strconv.Itoa(len(seen))
						t.Logf("interrupted checkpoint retains committed queue %s and complete trace %s", recoveryQueue, recoveryDistinct)
					} else {
						recoveryDistinct, recoveryQueue = counts[1], counts[2]
					}
					if len(nativeDistributedMessages(snapshot.output.String(), tlc.ECTLCCheckpointStart)) != 1 || len(nativeDistributedMessages(snapshot.output.String(), tlc.ECTLCCheckpointEnd)) != completedCheckpoints || len(nativeDistributedMessages(snapshot.output.String(), tlc.ECGeneral)) != 0 {
						t.Fatal("mid-run producer checkpoint phase or GENERAL assertion failed")
					}
					if len(nativeDistributedMessages(snapshot.output.String(), tlc.ECTLCDistributedWorkerRegistered)) != max(1, scenario.workerThreads) || len(nativeDistributedMessages(snapshot.output.String(), tlc.ECTLCFinished)) != 0 {
						t.Fatal("checkpoint producer did not stop an unfinished run with the configured workers")
					}
				}
				serverArgs = []string{"-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "-tool", "-deadlock", "-recover", path[1], "MC06"}
			}
			if scenario.remoteFP {
				count := max(1, scenario.fingerprintServers)
				serverArgs = append([]string{fmt.Sprintf("-Dtlc2.tool.distributed.TLCServer.expectedFPSetCount=%d", count)}, serverArgs...)
			}
			recoveredRoleStart := len(roles)
			server := start("server", serverArgs...)
			if scenario.combined {
				// Exercise the production launcher and shared native listener:
				// both roles must bootstrap and finish in this one process.
				start("worker-fpserver", "-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
			} else if scenario.remoteFP {
				// A supported native implementation avoids inheriting the
				// upstream harness's known OffHeap assumption failure.
				for index := range max(1, scenario.fingerprintServers) {
					fingerprintRole := "fpserver"
					if scenario.fingerprintReplyLoss && index == 0 {
						fingerprintRole = "fpserver-put-reply-loss"
					}
					fingerprint := start(fingerprintRole, "-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "127.0.0.1")
					if scenario.fingerprintLoss && index == 0 {
						failedFingerprint = fingerprint
						// Register partition zero first: source reassign uses a
						// forward assignment loop, without wrapping its writes.
						ticker := time.NewTicker(10 * time.Millisecond)
						for len(nativeDistributedMessages(server.output.String(), tlc.ECTLCDistributedServerFPSetRegistered)) == 0 {
							select {
							case err := <-server.done:
								server.joined = true
								t.Fatalf("coordinator exited before first FP registration: %v", err)
							case <-ctx.Done():
								t.Fatal("first FP registration watchdog expired")
							case <-ticker.C:
							}
						}
						ticker.Stop()
					}
				}
			}
			if scenario.fingerprintServers > 1 {
				// Before launching the worker, inspect the complete initial FP
				// frontier. This does not truncate or bound the ensuing model run.
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()
				for len(nativeDistributedMessages(server.output.String(), tlc.ECTLCDistributedServerRunning)) == 0 {
					select {
					case err := <-server.done:
						server.joined = true
						t.Fatalf("coordinator exited before initial partition inspection: %v", err)
					case <-ctx.Done():
						t.Fatal("initial partition inspection watchdog expired")
					case <-ticker.C:
					}
				}
				coordinator, err := rpc.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(port)))
				if err != nil {
					t.Fatal(err)
				}
				var reply tlc.DistributedServerReply
				err = coordinator.Call("Coordinator.Call", tlc.DistributedServerRequest{Object: tlc.TLCServerName, Operation: "manager"}, &reply)
				_ = coordinator.Close()
				if err != nil || reply.Failure != nil {
					t.Fatalf("initial manager snapshot = %v/%v", err, reply.Failure)
				}
				manager := reply.Manager
				if manager == nil || len(manager.Nodes) != 2 || len(manager.Partitions) != 2 || manager.Partitions[0] == manager.Partitions[1] {
					t.Fatal("coordinator did not retain two FP partitions")
				}
				if scenario.fingerprintLoss {
					survivingFingerprint = manager.Nodes[manager.Partitions[1]-1].Endpoint
				}
				var initial uint64
				for i, node := range manager.Nodes {
					endpoint, err := tlc.DialFingerprintEndpoint(node.Endpoint.Address, node.Endpoint.Object)
					if err != nil {
						t.Fatal(err)
					}
					size, err := endpoint.Size()
					_ = endpoint.CloseConnection()
					if err != nil || size == 0 {
						t.Fatalf("initial partition %d = %d/%v", i, size, err)
					}
					initial += size
					t.Logf("initial fingerprint partition %d contains %d states", i, size)
				}
				if initial != 16384 {
					t.Fatalf("initial fingerprint partitions total %d, want 16384", initial)
				}
			}
			if !scenario.combined && !scenario.allWorkersLost && !scenario.fingerprintLoss {
				workerRole := "worker"
				if scenario.duplicateWorkerRegistration {
					workerRole = "worker-register-twice"
				}
				start(workerRole, fmt.Sprintf("-Dtlc2.tool.distributed.TLCWorker.threadCount=%d", max(1, scenario.workerThreads)), "127.0.0.1")
			}
			if scenario.fingerprintLoss {
				workerRole := "worker-fingerprint-loss"
				if scenario.fingerprintReplyLoss {
					workerRole = "worker"
				}
				worker := start(workerRole, "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
				markerOutput, marker := worker.output, "NATIVE_WORKER_BLOCK_ASSIGNED"
				if scenario.fingerprintReplyLoss {
					markerOutput, marker = failedFingerprint.output, "NATIVE_FP_PUT_COMPLETED_NEW="
				}
				ticker := time.NewTicker(10 * time.Millisecond)
				for !strings.Contains(markerOutput.String(), marker) {
					select {
					case err := <-failedFingerprint.done:
						failedFingerprint.joined = true
						t.Fatalf("fingerprint host exited before loss barrier: %v", err)
					case err := <-server.done:
						server.joined = true
						t.Fatalf("coordinator exited before fingerprint loss: %v", err)
					case err := <-worker.done:
						worker.joined = true
						t.Fatalf("worker exited before fingerprint loss: %v", err)
					case <-ctx.Done():
						t.Fatal("fingerprint-loss worker assignment watchdog expired")
					case <-ticker.C:
					}
				}
				ticker.Stop()
				if err := failedFingerprint.command.Process.Kill(); err != nil {
					t.Fatal(err)
				}
				if err := <-failedFingerprint.done; err == nil {
					t.Fatal("killed fingerprint host unexpectedly exited successfully")
				}
				failedFingerprint.joined = true
				if scenario.fingerprintReplyLoss {
					completed := regexp.MustCompile(`NATIVE_FP_PUT_COMPLETED_NEW=(\d+)`).FindStringSubmatch(failedFingerprint.output.String())
					if len(completed) != 2 || completed[1] == "0" {
						t.Fatal("fingerprint host did not complete an actual insertion before reply loss")
					}
					t.Logf("killed first fingerprint host after inserting %s new fingerprints, before returning the reply", completed[1])
				} else {
					t.Log("killed first fingerprint host with a worker block assigned; resuming actual successor evaluation")
					if err := os.WriteFile(releaseWorker, nil, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				coordinator, err := rpc.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(port)))
				if err != nil {
					t.Fatal(err)
				}
				defer coordinator.Close()
				ticker = time.NewTicker(10 * time.Millisecond)
				for {
					var reply tlc.DistributedServerReply
					if err := coordinator.Call("Coordinator.Call", tlc.DistributedServerRequest{Object: tlc.TLCServerName, Operation: "manager"}, &reply); err != nil || reply.Failure != nil {
						t.Fatalf("failover manager inspection: %v/%v", err, reply.Failure)
					}
					manager := reply.Manager
					if manager != nil && len(manager.Nodes) == 1 && len(manager.Partitions) == 2 && manager.Partitions[0] == 1 && manager.Partitions[1] == 1 {
						if manager.Broken || !manager.Nodes[0].Available || manager.Nodes[0].Endpoint != survivingFingerprint {
							t.Fatal("failover did not retain the surviving fingerprint registration")
						}
						break
					}
					select {
					case err := <-server.done:
						server.joined = true
						t.Fatalf("coordinator exited before failover inspection: %v", err)
					case <-ctx.Done():
						t.Fatal("fingerprint reassignment inspection watchdog expired")
					case <-ticker.C:
					}
				}
				ticker.Stop()
				t.Log("coordinator retains two partitions sharing the surviving fingerprint registration")
			}
			if scenario.workerLoss {
				role, marker := "worker-failpoint", "NATIVE_WORKER_BLOCK_ASSIGNED"
				if scenario.workerReplyLoss {
					role, marker = "worker-reply-loss", "NATIVE_WORKER_REPLY_COMPUTED"
				}
				failed := start(role, "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
				// Wait for a real RPC block and all registrations. Process-loss
				// rows pause evaluation; reply-loss waits for actual completion.
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()
				registrations := 2
				if scenario.allWorkersLost {
					registrations = 1
				}
				for !strings.Contains(failed.output.String(), marker) || strings.Count(server.output.String(), "@!@!@STARTMSG 7001:") < registrations {
					select {
					case err := <-server.done:
						server.joined = true
						t.Fatalf("coordinator exited before worker assignment: %v\n%s", err, server.output.String())
					case err := <-failed.done:
						failed.joined = true
						t.Fatalf("failpoint worker exited before assigned work: %v", err)
					case <-ctx.Done():
						t.Fatal("worker assignment watchdog expired")
					case <-ticker.C:
					}
				}
				if scenario.workerReplyLoss {
					if err := os.WriteFile(releaseWorker, nil, 0o600); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := failed.command.Process.Kill(); err != nil {
						t.Fatal(err)
					}
					if err := <-failed.done; err == nil {
						t.Fatal("killed worker unexpectedly exited successfully")
					}
					failed.joined = true
				}
				if scenario.allWorkersLost {
					// The finally-block cache warning follows the source worker
					// cleanup and worker-count decrement. Wait for both before
					// permitting any replacement to register.
					for !strings.Contains(server.output.String(), fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCDistributedWorkerDeregistered)) || !strings.Contains(server.output.String(), nativeDistributedLostWorkerCacheWarning) {
						select {
						case err := <-server.done:
							server.joined = true
							t.Fatalf("coordinator exited after losing its last worker: %v\n%s", err, server.output.String())
						case <-ctx.Done():
							t.Fatal("last-worker cleanup watchdog expired")
						case <-ticker.C:
						}
					}
					coordinator, err := tlc.DialServerEndpoint(net.JoinHostPort("127.0.0.1", fmt.Sprint(port)), tlc.TLCServerWorkerName)
					if err != nil {
						t.Fatal(err)
					}
					done, failure := coordinator.IsDone()
					_ = coordinator.CloseConnection()
					if failure != nil || done {
						t.Fatalf("coordinator did not preserve unfinished work with no workers: done %v, error %v", done, failure)
					}
					t.Log("last worker cleanup completed; coordinator remains available with unfinished work")
				}
				if scenario.workerReplyLoss {
					for !strings.Contains(failed.output.String(), "NATIVE_WORKER_TRANSPORT_CLOSED_RUNTIME_ALIVE") {
						select {
						case err := <-failed.done:
							failed.joined = true
							t.Fatalf("worker exited before live-runtime confirmation: %v", err)
						case <-ctx.Done():
							t.Fatal("worker live-runtime confirmation watchdog expired")
						case <-ticker.C:
						}
					}
					if err := os.WriteFile(releaseWorker+".retire", nil, 0o600); err != nil {
						t.Fatal(err)
					}
					if err := <-failed.done; err != nil {
						t.Fatalf("retiring disconnected worker: %v", err)
					}
					failed.joined = true
					t.Log("computed reply lost; live worker retired after coordinator cleanup; starting replacement")
				} else {
					t.Log("killed worker with an unfinished assigned block; starting replacement")
				}
				start("worker", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
			}
			for _, process := range roles {
				if process.joined {
					continue
				}
				err := <-process.done
				process.joined = true
				if err != nil {
					t.Fatalf("%s exited with %v; output:\n%s", process.output.role, err, process.output.String())
				}
				if scenario.duplicateWorkerRegistration && process == server {
					if err := os.WriteFile(releaseWorker, nil, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				t.Logf("%s exited normally", process.output.role)
			}
			// Mechanical EWD840Distributed{WithFPSet}TLCTest assertions:
			// FINISHED, STATS distinct=114942 and queue=0, no GENERAL.
			output := server.output.String()
			if scenario.workerThreads > 1 {
				registered := nativeDistributedMessages(output, tlc.ECTLCDistributedWorkerRegistered)
				stats := nativeDistributedMessages(output, tlc.ECTLCDistributedWorkerStats)
				if len(registered) != scenario.workerThreads || len(stats) != scenario.workerThreads {
					t.Fatal("worker group did not register and report every worker")
				}
				endpoints := make(map[string]bool)
				var host string
				for _, message := range registered {
					// Registration text ends with the date. Its URI identifies
					// distinct workers on the one native process listener.
					match := regexp.MustCompile(`tcp://[^\s]+`).FindString(message)
					endpoint, err := url.Parse(match)
					if err != nil || match == "" || endpoints[match] || endpoint.Host == "" || endpoint.Path == "" {
						t.Fatalf("worker group endpoint identity is invalid: %q", message)
					}
					endpoints[match] = true
					if host != "" && host != endpoint.Host {
						t.Fatal("worker group did not share its process listener")
					}
					host = endpoint.Host
				}
				for _, message := range stats {
					endpoint := regexp.MustCompile(`tcp://[^\s]+`).FindString(message)
					counts := regexp.MustCompile(`Sent: (\d+) Rcvd: (\d+)`).FindStringSubmatch(message)
					// Ensure this row exercised every shared-app worker, rather
					// than merely registering an idle second endpoint.
					if !endpoints[endpoint] || len(counts) != 3 || counts[1] == "0" || counts[2] == "0" {
						t.Fatalf("worker group statistics lack actual work or identity: %q", message)
					}
					delete(endpoints, endpoint)
				}
				for _, process := range roles {
					roleOutput := process.output.String()
					if len(nativeDistributedMessages(roleOutput, tlc.ECGeneral)) != 0 || strings.Contains(roleOutput, "unexpected EOF") {
						t.Fatalf("shared-worker role %s emitted GENERAL or lost a reply", process.output.role)
					}
				}
			}
			if scenario.fingerprintServers > 1 {
				if len(nativeDistributedMessages(output, tlc.ECTLCDistributedServerFPSetRegistered)) != 2 {
					t.Fatal("partitioned coordinator did not accept exactly two fingerprint registrations")
				}
				for _, process := range roles {
					roleOutput := process.output.String()
					expectedEOF := 0
					if scenario.fingerprintReplyLoss && process == server {
						// The accepted put reply is deliberately lost in this row.
						// Source failover reports the transport error before retry.
						expectedEOF = 1
					}
					if len(nativeDistributedMessages(roleOutput, tlc.ECGeneral)) != 0 || strings.Count(roleOutput, "unexpected EOF") != expectedEOF {
						t.Fatalf("partitioned role %s emitted GENERAL or lost an RPC reply", process.output.role)
					}
				}
			}
			if scenario.workerLoss {
				if !strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCDistributedWorkerLost)) || strings.Count(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCDistributedWorkerDeregistered)) != 1 {
					t.Fatal("worker loss was not reported and deregistered exactly once")
				}
				if scenario.workerReplyLoss {
					if len(nativeDistributedMessages(output, tlc.ECTLCDistributedExceedBlocksize)) != 1 {
						t.Fatal("lost computed reply did not preserve the single source EOF smaller-block retry")
					}
					for _, process := range roles {
						if process.output.role == "worker-reply-loss" {
							workerOutput := process.output.String()
							if strings.Count(workerOutput, "NATIVE_WORKER_REPLY_COMPUTED") != 1 || strings.Count(workerOutput, "NATIVE_WORKER_TRANSPORT_CLOSED_RUNTIME_ALIVE") != 1 || len(nativeDistributedMessages(workerOutput, tlc.ECGeneral)) != 0 {
								t.Fatal("disconnected worker did not retain one completed reply and a live runtime without GENERAL")
							}
						}
					}
				}
			}
			if scenario.recovering {
				if !strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCCheckpointRecoverEnd)) || !strings.Contains(output, fmt.Sprintf("Recovery completed. %s states examined. %s states on queue.", recoveryDistinct, recoveryQueue)) {
					t.Fatalf("checkpoint recovery counts absent:\n%s", output)
				}
				if strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCComputingInit)) {
					t.Fatal("recovered coordinator regenerated initial states")
				}
				if scenario.midRunCheckpoint {
					if len(nativeDistributedMessages(output, tlc.ECTLCCheckpointRecoverEnd)) != 1 {
						t.Fatal("mid-run coordinator did not report exactly one recovery")
					}
					for _, process := range roles[recoveredRoleStart:] {
						roleOutput := process.output.String()
						if len(nativeDistributedMessages(roleOutput, tlc.ECGeneral)) != 0 || strings.Contains(roleOutput, "unexpected EOF") {
							t.Fatalf("recovered role %s emitted GENERAL or lost an RPC reply", process.output.role)
						}
					}
				}
			}
			if !strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCFinished)) {
				t.Fatal("TLC_FINISHED absent")
			}
			stats := regexp.MustCompile(fmt.Sprintf(`(?s)@!@!@STARTMSG %d:\d+ @!@!@\n(.*?)\n@!@!@ENDMSG %d @!@!@`, tlc.ECTLCStats, tlc.ECTLCStats)).FindStringSubmatch(output)
			distinct := 114942
			if scenario.fingerprintLoss {
				// Java size() sums both slots after they alias the survivor.
				// Ordinary original model assertions remain unchanged.
				distinct *= 2
				for _, process := range []*nativeDistributedTestProcess{server, roles[len(roles)-1]} {
					if strings.Count(process.output.String(), "Warning: Failed to connect from ") != 1 || strings.Contains(process.output.String(), "Warning: there is no fp server available.") {
						t.Fatalf("%s did not report fingerprint reassignment to surviving storage", process.output.role)
					}
				}
			}
			if len(stats) != 2 || !regexp.MustCompile(fmt.Sprintf(`^\d+ states generated, %d distinct states found, 0 states left on queue\.$`, distinct)).MatchString(stats[1]) {
				t.Fatal("TLC_STATS lacks original distinct/queue assertions")
			}
			if scenario.workerLoss {
				general := regexp.MustCompile(fmt.Sprintf(`(?s)@!@!@STARTMSG %d:(\d+) @!@!@\n(.*?)\n@!@!@ENDMSG %d @!@!@`, tlc.ECGeneral, tlc.ECGeneral)).FindAllStringSubmatch(output, -1)
				if len(general) != 1 || general[0][1] != "3" || general[0][2] != nativeDistributedLostWorkerCacheWarning {
					t.Fatalf("worker-loss GENERAL events differ from source cache warning: %v", general)
				}
			} else if scenario.duplicateWorkerRegistration {
				registered := nativeDistributedMessages(output, tlc.ECTLCDistributedWorkerRegistered)
				if len(registered) != 2 || registered[0] != registered[1] || len(nativeDistributedMessages(output, tlc.ECTLCDistributedWorkerStats)) != 2 || len(nativeDistributedMessages(output, tlc.ECTLCDistributedWorkerLost)) != 0 {
					t.Fatal("duplicate worker registration lost identity, coordinator threads or orderly completion")
				}
				general := regexp.MustCompile(fmt.Sprintf(`(?s)@!@!@STARTMSG %d:(\d+) @!@!@\n(.*?)\n@!@!@ENDMSG %d @!@!@`, tlc.ECGeneral, tlc.ECGeneral)).FindAllStringSubmatch(output, -1)
				ignoredExit, cacheWarning := 0, 0
				for _, message := range general {
					if message[1] != "3" {
						t.Fatal("duplicate registration emitted a GENERAL error")
					}
					switch message[2] {
					case "Ignoring attempt to exit dead worker":
						ignoredExit++
					case nativeDistributedLostWorkerCacheWarning:
						cacheWarning++
					default:
						t.Fatalf("unexpected duplicate-registration diagnostic %q", message[2])
					}
				}
				// The second thread's final cache query can overlap the first
				// thread's exit. Source permits that single cache warning.
				if ignoredExit != 1 || cacheWarning > 1 {
					t.Fatalf("duplicate exit/cache warnings %d/%d", ignoredExit, cacheWarning)
				}
			} else if strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECGeneral)) {
				t.Fatal("GENERAL recorded")
			}
			if !scenario.fingerprintReplyLoss && strings.Contains(output, "unexpected EOF") {
				t.Fatal("native process shutdown lost an accepted RPC reply")
			}
		})
	}
}

const nativeDistributedLostWorkerCacheWarning = "Failed to read remote worker cache statistic (Expect to see a negative chache hit rate. Does not invalidate model checking results)"

func TestNativeDistributedProcessHelper(t *testing.T) {
	if os.Getenv("TLAGO_NATIVE_DISTRIBUTED_PROCESS_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			args := os.Args[i+1:]
			status := ExitToolFailure
			if len(args) > 0 && args[0] == "checkpoint-fp-host" {
				if err := nativeCheckpointFingerprintHost(); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "fpserver-put-reply-loss" {
				if err := nativeDistributedFingerprintLostReply(args[1:]); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "worker-register-twice" {
				if err := nativeDistributedRegisterWorkerTwice(args[1:], os.Getenv("TLAGO_NATIVE_WORKER_RELEASE")); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "registered-fp-recovery" {
				if err := nativeRegisteredFingerprintRecovery(args[1:]); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "checkpoint-frontier" {
				if err := nativeDistributedCheckpointFrontier(args[1:]); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "checkpoint-mid-run" {
				if err := nativeDistributedCheckpointMidRun(args[1:], ""); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "checkpoint-interrupt" {
				if err := nativeDistributedCheckpointMidRun(args[1:], "before_queue_commit"); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "checkpoint-interrupt-after-queue" {
				if err := nativeDistributedCheckpointMidRun(args[1:], "after_queue_commit"); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "checkpoint-interrupt-after-intern" {
				if err := nativeDistributedCheckpointMidRun(args[1:], "after_intern_commit"); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "checkpoint-interrupt-after-first-fingerprint" {
				if err := nativeDistributedCheckpointMidRun(args[1:], "after_first_fingerprint_commit"); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "worker-failpoint" {
				if err := nativeDistributedBlockedWorker(args[1:]); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "worker-reply-loss" {
				if err := nativeDistributedLostReplyWorker(args[1:], os.Getenv("TLAGO_NATIVE_WORKER_RELEASE")); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "worker-fingerprint-loss" {
				if err := nativeDistributedPausedWorker(args[1:], os.Getenv("TLAGO_NATIVE_WORKER_RELEASE")); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else {
				status = RunCLI(args, os.Stdout, os.Stderr)
			}
			tlc.CleanupDistributedFiles()
			os.Exit(status)
		}
	}
	t.Fatal("helper role arguments missing")
}

// Use the production worker bootstrap and TCP callbacks. Only evaluation is
// paused, after a coordinator block has actually arrived, until this owned
// test process is killed. No successor response or fingerprint put is faked.
func nativeDistributedBlockedWorker(args []string) error {
	return nativeDistributedPausedWorker(args, "")
}

// Hold an actual completed result before native RPC serialization. The parent
// closes this host while keeping evaluation/runtime ownership in this process.
type nativeDistributedHeldReply struct {
	*tlc.LocalWorkerEndpoint
	host        *tlc.DistributedRPCServer
	releasePath string
}

func (e *nativeDistributedHeldReply) GetNextStates(states []*tlc.TLCStateMut) (*tlc.NextStateResult, error) {
	result, err := e.LocalWorkerEndpoint.GetNextStates(states)
	if err != nil || result == nil {
		return nil, fmt.Errorf("held worker computation: %v", err)
	}
	fmt.Println("NATIVE_WORKER_REPLY_COMPUTED")
	if err := nativeDistributedWaitForFile(e.releasePath); err != nil {
		return nil, err
	}
	if err := e.host.Close(); err != nil {
		return nil, err
	}
	if !e.Worker.IsAlive() {
		return nil, fmt.Errorf("transport closure terminated worker runtime")
	}
	fmt.Println("NATIVE_WORKER_TRANSPORT_CLOSED_RUNTIME_ALIVE")
	if err := nativeDistributedWaitForFile(e.releasePath + ".retire"); err != nil {
		return nil, err
	}
	return result, nil
}

func nativeDistributedWaitForFile(path string) error {
	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func nativeDistributedLostReplyWorker(args []string, releasePath string) error {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	if releasePath == "" {
		return fmt.Errorf("worker reply gate path missing")
	}
	network, err := tlc.NewDistributedWorkerNetwork("127.0.0.1:0", "127.0.0.1")
	if err != nil {
		return err
	}
	defer network.Close()
	process := tlc.NewDistributedWorkerProcess()
	env := network.Environment(tlc.DistributedWorkerEnvironment{LoadApp: func(server tlc.DistributedServerEndpoint, resolver *tlc.DistributedFilenameToStreamResolver) (*tlc.TLCApp, error) {
		app, diagnostics, err := loadDistributedEndpointApp(server, resolver, tlc.RuntimeParameters{})
		if err != nil {
			return nil, err
		}
		if app == nil || diagnostics.HasErrors() {
			return nil, fmt.Errorf("worker model failed to load: %v", diagnostics)
		}
		return app, nil
	}})
	publish := env.PublishWorker
	env.PublishWorker = func(worker *tlc.DistributedWorker) error {
		if err := publish(worker); err != nil {
			return err
		}
		uri, err := url.Parse(worker.GetURI())
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(uri.Path, "/")
		network.Host.UnregisterWorker(name)
		return network.Host.RegisterWorker(name, &nativeDistributedHeldReply{LocalWorkerEndpoint: tlc.NewLocalWorkerEndpoint(worker), host: network.Host, releasePath: releasePath})
	}
	if _, err := RunDistributedWorker(process, args, env, tlc.RuntimeParameters{}); err != nil {
		return err
	}
	if process.Group == nil {
		return fmt.Errorf("worker startup failed")
	}
	if err := nativeDistributedWaitForFile(releasePath + ".retire"); err != nil {
		return err
	}
	if err := process.Shutdown(); err != nil {
		return err
	}
	return process.Runtime.AwaitTermination()
}

func nativeDistributedPausedWorker(args []string, releasePath string) error {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	network, err := tlc.NewDistributedWorkerNetwork("127.0.0.1:0", "127.0.0.1")
	if err != nil {
		return err
	}
	defer network.Close()
	process := tlc.NewDistributedWorkerProcess()
	env := network.Environment(tlc.DistributedWorkerEnvironment{LoadApp: func(server tlc.DistributedServerEndpoint, resolver *tlc.DistributedFilenameToStreamResolver) (*tlc.TLCApp, error) {
		app, diagnostics, err := loadDistributedEndpointApp(server, resolver, tlc.RuntimeParameters{})
		if err != nil {
			return nil, err
		}
		if app == nil || diagnostics.HasErrors() {
			return nil, fmt.Errorf("worker model failed to load: %v", diagnostics)
		}
		var pause sync.Once
		app.Tool.GetNextStatesFunc = func(tool *tlc.Tool, action *tlc.Action, state *tlc.TLCStateMut) (*tlc.StateVec, error) {
			pause.Do(func() {
				fmt.Println("NATIVE_WORKER_BLOCK_ASSIGNED")
				if releasePath == "" {
					<-make(chan struct{})
				}
				for {
					if _, err := os.Stat(releasePath); err == nil {
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			})
			return tool.GetNextStatesImpl(action, state)
		}
		return app, nil
	}})
	if _, err := RunDistributedWorker(process, args, env, tlc.RuntimeParameters{}); err != nil {
		return err
	}
	if process.Group == nil {
		return fmt.Errorf("worker startup failed")
	}
	return process.Runtime.AwaitTermination()
}

// Checkpoint an unchanged model's complete initial frontier before any worker
// starts. A fresh coordinator process must restore it via the production CLI.
func nativeDistributedCheckpointFrontier(args []string) error {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	tlc.SetNumWorkers(0)
	app, diagnostics, err := CreateTLCApp(args, tlc.RuntimeParameters{})
	if err != nil {
		return err
	}
	if app == nil || diagnostics.HasErrors() {
		return fmt.Errorf("checkpoint model did not load: %v", diagnostics)
	}
	server, err := tlc.NewTLCServerFromApp(app)
	if err != nil {
		return err
	}
	defer server.Close(false)
	if code, err := server.DoInit(); err != nil || code != tlc.NoError {
		return fmt.Errorf("initial frontier: code %d, error %v", code, err)
	}
	if server.StateQueue.Size() != 16384 || server.FPSetManager.Size() != 16384 {
		return fmt.Errorf("incomplete original initial frontier")
	}
	if err := server.Checkpoint(); err != nil {
		return err
	}
	fmt.Println("NATIVE_CHECKPOINT_PATH=" + app.GetMetadir())
	return nil
}

type nativeMidRunCheckpointQueue struct {
	tlc.StateQueue
	server       *tlc.TLCServer
	distinct     uint64
	queued       int64
	interruption string
}

func (q *nativeMidRunCheckpointQueue) BeginChkpt() error {
	// Checkpoint has already suspended all server threads here. Capture the
	// counts of the persisted frontier, rather than post-resume live counters.
	q.distinct, q.queued = q.server.FPSetManager.Size(), q.Size()
	return q.StateQueue.BeginChkpt()
}

func (q *nativeMidRunCheckpointQueue) interruptProcess() {
	fmt.Printf("NATIVE_INTERRUPTED_CHECKPOINT_COUNTS=%d,%d\n", q.distinct, q.queued)
	fmt.Println("NATIVE_CHECKPOINT_PATH=" + q.server.Metadir)
	os.Exit(ExitOK)
}

func (q *nativeMidRunCheckpointQueue) CommitChkpt() error {
	if q.interruption == "before_queue_commit" {
		q.interruptProcess()
	}
	if err := q.StateQueue.CommitChkpt(); err != nil {
		return err
	}
	if q.interruption == "after_queue_commit" {
		q.interruptProcess()
	}
	return nil
}

// All storage operations delegate to the source-configured production set.
// Entry into this commit occurs after queue/trace/intern commits in TLCServer.
type nativeInterruptedFingerprintCommit struct {
	tlc.FPSet
	server      *tlc.TLCServer
	afterCommit bool
}

func (s *nativeInterruptedFingerprintCommit) CommitChkpt() error {
	queue, ok := s.server.StateQueue.(*nativeMidRunCheckpointQueue)
	if !ok || (!s.afterCommit && queue.interruption != "after_intern_commit") || (s.afterCommit && queue.interruption != "after_first_fingerprint_commit") {
		return fmt.Errorf("unexpected interrupted fingerprint commit setup")
	}
	if s.afterCommit {
		if err := s.FPSet.CommitChkpt(); err != nil {
			return err
		}
	}
	queue.interruptProcess()
	return nil
}

// Run the unchanged model with a real TCP worker until successors have been
// inserted. Commit using the production checkpoint barrier, then abruptly exit
// this process. The parent starts a fresh CLI coordinator and worker to recover.
func nativeDistributedCheckpointMidRun(args []string, interruption string) error {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	network := tlc.NewDistributedCoordinatorNetwork("127.0.0.1", "127.0.0.1")
	defer network.Close()
	process := tlc.NewDistributedServerProcess()
	env := tlc.DistributedServerEnvironment{
		CreateServer: func(app *tlc.TLCApp, _ int) (*tlc.TLCServer, error) {
			var server *tlc.TLCServer
			var err error
			if os.Getenv("TLAGO_REGISTERED_FP_ENDPOINTS") != "" {
				server, err = nativeRegisteredFingerprintServer(app)
			} else {
				server, err = tlc.NewTLCServerFromApp(app)
			}
			if err == nil {
				server.ConfigurePublication(network.Publication())
				if interruption == "after_intern_commit" || interruption == "after_first_fingerprint_commit" {
					// MemFPSet init owns no open files. Replace the unused initial
					// manager with the same factory/configuration plus a commit
					// failpoint, before recovery or initialization uses either set.
					config := app.GetFPSetConfiguration()
					if config.GetImplementation() != "tlc2.tool.fp.MemFPSet" {
						return nil, fmt.Errorf("interrupted fixture requires memory fingerprint storage")
					}
					set := tlc.NewFPSet(config)
					if interruption == "after_first_fingerprint_commit" {
						multi, ok := set.(*tlc.MultiFPSet)
						if !ok || len(multi.Sets) != 2 {
							return nil, fmt.Errorf("interrupted fixture requires two nested fingerprint sets")
						}
						multi.Sets[0] = &nativeInterruptedFingerprintCommit{FPSet: multi.Sets[0], server: server, afterCommit: true}
					} else {
						set = &nativeInterruptedFingerprintCommit{FPSet: set, server: server}
					}
					set.Init(1, app.GetMetadir(), app.GetFileName())
					server.FPSetManager = tlc.NewNonDistributedFPSetManager(set, server.FPSetManager.GetHostName(), server.Trace)
				}
			}
			return server, err
		},
		ModelCheck: func(server *tlc.TLCServer) error {
			queue := &nativeMidRunCheckpointQueue{StateQueue: server.StateQueue, server: server, interruption: interruption}
			server.StateQueue = queue
			stop, joined := make(chan struct{}), make(chan struct{})
			go func() {
				defer close(joined)
				defer func() {
					if failure := recover(); failure != nil {
						fmt.Fprintln(os.Stderr, "mid-run checkpoint panic:", failure)
						os.Exit(ExitToolFailure)
					}
				}()
				ticker := time.NewTicker(time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-stop:
						return
					case <-ticker.C:
						if server.FPSetManager.Size() <= 16384 {
							continue
						}
					}
					if err := server.Checkpoint(); err != nil {
						fmt.Fprintln(os.Stderr, "mid-run checkpoint:", err)
						os.Exit(ExitToolFailure)
					}
					fmt.Printf("NATIVE_CHECKPOINT_COUNTS=%d,%d\n", queue.distinct, queue.queued)
					fmt.Println("NATIVE_CHECKPOINT_PATH=" + server.Metadir)
					os.Exit(ExitOK)
				}
			}()
			_, err := server.ModelCheck()
			close(stop)
			<-joined
			if err != nil {
				return err
			}
			return fmt.Errorf("model completed before mid-run checkpoint")
		},
	}
	_, err = RunDistributedServer(process, args, env, tlc.RuntimeParameters{})
	return err
}

type nativeDistributedTestProcess struct {
	command *exec.Cmd
	output  *nativeDistributedTestLog
	done    chan error
	joined  bool
}

type nativeDistributedTestLog struct {
	mu      sync.Mutex
	test    *testing.T
	role    string
	output  strings.Builder
	pending string
}

func (log *nativeDistributedTestLog) Write(data []byte) (int, error) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.output.Write(data)
	log.pending += string(data)
	for {
		line, rest, found := strings.Cut(log.pending, "\n")
		if !found {
			break
		}
		log.pending = rest
		log.test.Logf("%s: %s", log.role, line)
	}
	return len(data), nil
}

func (log *nativeDistributedTestLog) String() string {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.output.String()
}
