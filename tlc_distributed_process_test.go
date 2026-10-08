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
	for _, remoteFP := range []bool{false, true} {
		name := "coordinator_fingerprints"
		if remoteFP {
			name = "standalone_fingerprints"
		}
		t.Run(name, func(t *testing.T) {
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
			if remoteFP {
				serverArgs = append([]string{"-Dtlc2.tool.distributed.TLCServer.expectedFPSetCount=1"}, serverArgs...)
			}
			server := start("server", serverArgs...)
			if remoteFP {
				// A supported native implementation avoids inheriting the
				// upstream harness's known OffHeap assumption failure.
				start("fpserver", "-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "127.0.0.1")
			}
			start("worker", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
			for _, process := range roles {
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
			if !strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCFinished)) {
				t.Fatal("TLC_FINISHED absent")
			}
			stats := regexp.MustCompile(fmt.Sprintf(`(?s)@!@!@STARTMSG %d:\d+ @!@!@\n(.*?)\n@!@!@ENDMSG %d @!@!@`, tlc.ECTLCStats, tlc.ECTLCStats)).FindStringSubmatch(output)
			if len(stats) != 2 || !regexp.MustCompile(`^\d+ states generated, 114942 distinct states found, 0 states left on queue\.$`).MatchString(stats[1]) {
				t.Fatal("TLC_STATS lacks original distinct/queue assertions")
			}
			if strings.Contains(output, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECGeneral)) {
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
			status := RunCLI(os.Args[i+1:], os.Stdout, os.Stderr)
			tlc.CleanupDistributedFiles()
			os.Exit(status)
		}
	}
	t.Fatal("helper role arguments missing")
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
