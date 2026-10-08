package tlago

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
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
		fingerprintServers                         int
		midRunCheckpoint                           bool
		checkpointInterrupted                      bool
	}{
		{name: "coordinator_fingerprints"},
		{name: "standalone_fingerprints", remoteFP: true},
		{name: "partitioned_fingerprints", remoteFP: true, fingerprintServers: 2},
		{name: "combined_worker_fingerprints", remoteFP: true, combined: true},
		{name: "checkpoint_recovery", recovering: true},
		{name: "mid_run_checkpoint_recovery", recovering: true, midRunCheckpoint: true},
		{name: "checkpoint_interruption_before_commit", recovering: true, midRunCheckpoint: true, checkpointInterrupted: true},
		{name: "worker_loss", workerLoss: true},
		{name: "all_workers_lost", workerLoss: true, allWorkersLost: true},
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
			roleCounts := make(map[string]int)
			start := func(role string, args ...string) *nativeDistributedTestProcess {
				commandArgs := []string{"-test.run=^TestNativeDistributedProcessHelper$", "--", role}
				commandArgs = append(commandArgs, common...)
				commandArgs = append(commandArgs, args...)
				command := exec.CommandContext(ctx, os.Args[0], commandArgs...)
				command.Dir = model
				command.Env = append(os.Environ(), "TLAGO_NATIVE_DISTRIBUTED_PROCESS_HELPER=1")
				if scenario.fingerprintServers > 1 && role == "fpserver" {
					// These roles represent separate hosts. Give each private
					// temporary storage even when they start in the same millisecond.
					command.Env = append(command.Env, "TMPDIR="+t.TempDir())
				}
				roleCounts[role]++
				label := role
				if scenario.fingerprintServers > 1 && role == "fpserver" {
					label = fmt.Sprintf("fpserver-%d", roleCounts[role])
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
				producer := "checkpoint-frontier"
				if scenario.midRunCheckpoint {
					producer = "checkpoint-mid-run"
				}
				producerArgs := append([]string{"-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet"}, serverArgs...)
				if scenario.checkpointInterrupted {
					// Establish an older complete checkpoint. The next producer
					// advances it but exits before committing any replacement file.
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
					producer = "checkpoint-interrupt"
					producerArgs = []string{"-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "-tool", "-deadlock", "-recover", path[1], "MC06"}
				}
				snapshot := start(producer, producerArgs...)
				var snapshotWorker *nativeDistributedTestProcess
				if scenario.midRunCheckpoint {
					snapshotWorker = start("worker", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
				}
				err := <-snapshot.done
				snapshot.joined = true
				if err != nil {
					t.Fatalf("checkpoint producer: %v\n%s", err, snapshot.output.String())
				}
				path := regexp.MustCompile(`(?m)^NATIVE_CHECKPOINT_PATH=(.+)$`).FindStringSubmatch(snapshot.output.String())
				if len(path) != 2 {
					t.Fatal("checkpoint producer did not publish its committed directory")
				}
				if snapshotWorker != nil {
					// The producer deliberately exits after commit, without
					// finishing exploration. Retire its old worker before restart.
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
						elements := trace.Elements()
						seen := make(map[uint64]bool)
						for elements.NextPos() != -1 {
							fp := elements.NextFP()
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
						// The untouched queue.chkpt still contains the old initial
						// frontier. Require both exact independent recovery counts.
						recoveryDistinct = strconv.Itoa(len(seen))
						t.Logf("interrupted checkpoint retains old queue %s and complete trace %s", recoveryQueue, recoveryDistinct)
					} else {
						recoveryDistinct, recoveryQueue = counts[1], counts[2]
					}
					if len(nativeDistributedMessages(snapshot.output.String(), tlc.ECTLCCheckpointStart)) != 1 || len(nativeDistributedMessages(snapshot.output.String(), tlc.ECTLCCheckpointEnd)) != completedCheckpoints || len(nativeDistributedMessages(snapshot.output.String(), tlc.ECGeneral)) != 0 {
						t.Fatal("mid-run producer checkpoint phase or GENERAL assertion failed")
					}
					if len(nativeDistributedMessages(snapshot.output.String(), tlc.ECTLCDistributedWorkerRegistered)) != 1 || len(nativeDistributedMessages(snapshot.output.String(), tlc.ECTLCFinished)) != 0 {
						t.Fatal("checkpoint producer did not stop an unfinished run with one real worker")
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
				for range max(1, scenario.fingerprintServers) {
					start("fpserver", "-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "127.0.0.1")
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
			if !scenario.combined && !scenario.allWorkersLost {
				start("worker", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
			}
			if scenario.workerLoss {
				failed := start("worker-failpoint", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
				// Wait for a real assigned RPC block and all registrations.
				// The failpoint pauses actual successor evaluation, so killing
				// it deterministically discards an unfinished assigned block.
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()
				registrations := 2
				if scenario.allWorkersLost {
					registrations = 1
				}
				for !strings.Contains(failed.output.String(), "NATIVE_WORKER_BLOCK_ASSIGNED") || strings.Count(server.output.String(), "@!@!@STARTMSG 7001:") < registrations {
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
				if err := failed.command.Process.Kill(); err != nil {
					t.Fatal(err)
				}
				if err := <-failed.done; err == nil {
					t.Fatal("killed worker unexpectedly exited successfully")
				}
				failed.joined = true
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
				t.Log("killed worker with an unfinished assigned block; starting replacement")
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
				t.Logf("%s exited normally", process.output.role)
			}
			// Mechanical EWD840Distributed{WithFPSet}TLCTest assertions:
			// FINISHED, STATS distinct=114942 and queue=0, no GENERAL.
			output := server.output.String()
			if scenario.fingerprintServers > 1 {
				if len(nativeDistributedMessages(output, tlc.ECTLCDistributedServerFPSetRegistered)) != 2 {
					t.Fatal("partitioned coordinator did not accept exactly two fingerprint registrations")
				}
				for _, process := range roles {
					roleOutput := process.output.String()
					if len(nativeDistributedMessages(roleOutput, tlc.ECGeneral)) != 0 || strings.Contains(roleOutput, "unexpected EOF") {
						t.Fatalf("partitioned role %s emitted GENERAL or lost an RPC reply", process.output.role)
					}
				}
			}
			if scenario.workerLoss {
				if !strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCDistributedWorkerLost)) || strings.Count(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCDistributedWorkerDeregistered)) != 1 {
					t.Fatal("worker loss was not reported and deregistered exactly once")
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
			if len(stats) != 2 || !regexp.MustCompile(`^\d+ states generated, 114942 distinct states found, 0 states left on queue\.$`).MatchString(stats[1]) {
				t.Fatal("TLC_STATS lacks original distinct/queue assertions")
			}
			if scenario.workerLoss {
				general := regexp.MustCompile(fmt.Sprintf(`(?s)@!@!@STARTMSG %d:(\d+) @!@!@\n(.*?)\n@!@!@ENDMSG %d @!@!@`, tlc.ECGeneral, tlc.ECGeneral)).FindAllStringSubmatch(output, -1)
				if len(general) != 1 || general[0][1] != "3" || general[0][2] != nativeDistributedLostWorkerCacheWarning {
					t.Fatalf("worker-loss GENERAL events differ from source cache warning: %v", general)
				}
			} else if strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECGeneral)) {
				t.Fatal("GENERAL recorded")
			}
			if strings.Contains(output, "unexpected EOF") {
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
			if len(args) > 0 && args[0] == "checkpoint-frontier" {
				if err := nativeDistributedCheckpointFrontier(args[1:]); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "checkpoint-mid-run" {
				if err := nativeDistributedCheckpointMidRun(args[1:], false); err != nil {
					fmt.Fprintln(os.Stderr, err)
				} else {
					status = ExitOK
				}
			} else if len(args) > 0 && args[0] == "checkpoint-interrupt" {
				if err := nativeDistributedCheckpointMidRun(args[1:], true); err != nil {
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
		app.Tool.GetNextStatesFunc = func(*tlc.Tool, *tlc.Action, *tlc.TLCStateMut) (*tlc.StateVec, error) {
			fmt.Println("NATIVE_WORKER_BLOCK_ASSIGNED")
			<-make(chan struct{})
			return nil, fmt.Errorf("unreachable paused worker evaluation")
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
	server    *tlc.TLCServer
	distinct  uint64
	queued    int64
	interrupt bool
}

func (q *nativeMidRunCheckpointQueue) BeginChkpt() error {
	// Checkpoint has already suspended all server threads here. Capture the
	// counts of the persisted frontier, rather than post-resume live counters.
	q.distinct, q.queued = q.server.FPSetManager.Size(), q.Size()
	return q.StateQueue.BeginChkpt()
}

func (q *nativeMidRunCheckpointQueue) CommitChkpt() error {
	if q.interrupt {
		// Queue commit is the first replacement-file commit. Temporary queue,
		// trace, fingerprint and intern files have been written. Recovery still
		// reads the old queue, while MultiFPSet enumerates the full trace file.
		fmt.Printf("NATIVE_INTERRUPTED_CHECKPOINT_COUNTS=%d,%d\n", q.distinct, q.queued)
		fmt.Println("NATIVE_CHECKPOINT_PATH=" + q.server.Metadir)
		os.Exit(ExitOK)
	}
	return q.StateQueue.CommitChkpt()
}

// Run the unchanged model with a real TCP worker until successors have been
// inserted. Commit using the production checkpoint barrier, then abruptly exit
// this process. The parent starts a fresh CLI coordinator and worker to recover.
func nativeDistributedCheckpointMidRun(args []string, interrupt bool) error {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	network := tlc.NewDistributedCoordinatorNetwork("127.0.0.1", "127.0.0.1")
	defer network.Close()
	process := tlc.NewDistributedServerProcess()
	env := tlc.DistributedServerEnvironment{
		CreateServer: func(app *tlc.TLCApp, _ int) (*tlc.TLCServer, error) {
			server, err := tlc.NewTLCServerFromApp(app)
			if err == nil {
				server.ConfigurePublication(network.Publication())
			}
			return server, err
		},
		ModelCheck: func(server *tlc.TLCServer) error {
			queue := &nativeMidRunCheckpointQueue{StateQueue: server.StateQueue, server: server, interrupt: interrupt}
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
