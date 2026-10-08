package tlago

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	}{
		{name: "coordinator_fingerprints"},
		{name: "standalone_fingerprints", remoteFP: true},
		{name: "combined_worker_fingerprints", remoteFP: true, combined: true},
		{name: "checkpoint_recovery", recovering: true},
		{name: "worker_loss", workerLoss: true},
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
			start := func(role string, args ...string) *nativeDistributedTestProcess {
				commandArgs := []string{"-test.run=^TestNativeDistributedProcessHelper$", "--", role}
				commandArgs = append(commandArgs, common...)
				commandArgs = append(commandArgs, args...)
				command := exec.CommandContext(ctx, os.Args[0], commandArgs...)
				command.Dir = model
				command.Env = append(os.Environ(), "TLAGO_NATIVE_DISTRIBUTED_PROCESS_HELPER=1")
				output := &nativeDistributedTestLog{test: t, role: role}
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
			if scenario.recovering {
				snapshot := start("checkpoint-frontier", append([]string{"-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet"}, serverArgs...)...)
				err := <-snapshot.done
				snapshot.joined = true
				if err != nil {
					t.Fatalf("checkpoint producer: %v\n%s", err, snapshot.output.String())
				}
				path := regexp.MustCompile(`(?m)^NATIVE_CHECKPOINT_PATH=(.+)$`).FindStringSubmatch(snapshot.output.String())
				if len(path) != 2 {
					t.Fatal("checkpoint producer did not publish its committed directory")
				}
				serverArgs = []string{"-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "-tool", "-deadlock", "-recover", path[1], "MC06"}
			}
			if scenario.remoteFP {
				serverArgs = append([]string{"-Dtlc2.tool.distributed.TLCServer.expectedFPSetCount=1"}, serverArgs...)
			}
			server := start("server", serverArgs...)
			if scenario.combined {
				// Exercise the production launcher and shared native listener:
				// both roles must bootstrap and finish in this one process.
				start("worker-fpserver", "-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
			} else if scenario.remoteFP {
				// A supported native implementation avoids inheriting the
				// upstream harness's known OffHeap assumption failure.
				start("fpserver", "-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "127.0.0.1")
			}
			if !scenario.combined {
				start("worker", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
			}
			if scenario.workerLoss {
				failed := start("worker-failpoint", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
				// Wait for a real assigned RPC block and both registrations.
				// The failpoint pauses actual successor evaluation, so killing
				// it deterministically discards an unfinished assigned block.
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()
				for !strings.Contains(failed.output.String(), "NATIVE_WORKER_BLOCK_ASSIGNED") || strings.Count(server.output.String(), "@!@!@STARTMSG 7001:") < 2 {
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
			if scenario.workerLoss {
				if !strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCDistributedWorkerLost)) || strings.Count(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCDistributedWorkerDeregistered)) != 1 {
					t.Fatal("worker loss was not reported and deregistered exactly once")
				}
			}
			if scenario.recovering {
				if !strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCCheckpointRecoverEnd)) || !strings.Contains(output, "Recovery completed. 16384 states examined. 16384 states on queue.") {
					t.Fatalf("initial frontier recovery counts absent:\n%s", output)
				}
				if strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCComputingInit)) {
					t.Fatal("recovered coordinator regenerated initial states")
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
				if len(general) != 1 || general[0][1] != "3" || general[0][2] != "Failed to read remote worker cache statistic (Expect to see a negative chache hit rate. Does not invalidate model checking results)" {
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
	env := network.Environment(tlc.DistributedWorkerEnvironment{LoadApp: func(server tlc.DistributedServerEndpoint, resolver *tlc.RMIFilenameToStreamResolver) (*tlc.TLCApp, error) {
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
