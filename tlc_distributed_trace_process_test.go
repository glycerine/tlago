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
	"testing"
	"time"

	"github.com/glycerine/tlago/tlc"
)

// DistributedTLCTestCase is disabled upstream. These checks retain its model
// assertions over native Go processes, without original-method completion credit.
func TestNativeDistributedErrorTraceModels(t *testing.T) {
	for _, model := range []string{"DieHard", "TSnapShot"} {
		for _, remote := range []bool{false, true} {
			storage := "coordinator_fingerprints"
			if remote {
				storage = "standalone_fingerprints"
			}
			t.Run(model+"/"+storage, func(t *testing.T) {
				output := runNativeDistributedTraceModel(t, model, remote)
				for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCBehaviorUpToThisPoint} {
					if len(nativeDistributedMessages(output, code)) == 0 {
						t.Fatalf("original required event %d absent", code)
					}
				}
				if len(nativeDistributedMessages(output, tlc.ECGeneral)) != 0 {
					t.Fatal("GENERAL recorded")
				}
				if model == "TSnapShot" {
					// TSnapShotDistributedTLCTest.test: STATS parameter 2 = 0.
					stats := nativeDistributedMessages(output, tlc.ECTLCStats)
					if len(stats) != 1 || !regexp.MustCompile(`^\d+ states generated, \d+ distinct states found, 0 states left on queue\.$`).MatchString(stats[0]) {
						t.Fatalf("original zero-queue assertion failed: %q", stats)
					}
					return
				}
				// DieHardDistributedTLCTest.testSpec: exact original seven states.
				want := []string{
					"/\\ action = \"nondet\"\n/\\ smallBucket = 0\n/\\ bigBucket = 0\n/\\ water_to_pour = 0",
					"/\\ action = \"fill big\"\n/\\ smallBucket = 0\n/\\ bigBucket = 5\n/\\ water_to_pour = 0",
					"/\\ action = \"pour big to small\"\n/\\ smallBucket = 3\n/\\ bigBucket = 2\n/\\ water_to_pour = 3",
					"/\\ action = \"empty small\"\n/\\ smallBucket = 0\n/\\ bigBucket = 2\n/\\ water_to_pour = 3",
					"/\\ action = \"pour big to small\"\n/\\ smallBucket = 2\n/\\ bigBucket = 0\n/\\ water_to_pour = 2",
					"/\\ action = \"fill big\"\n/\\ smallBucket = 2\n/\\ bigBucket = 5\n/\\ water_to_pour = 2",
					"/\\ action = \"pour big to small\"\n/\\ smallBucket = 3\n/\\ bigBucket = 4\n/\\ water_to_pour = 1",
				}
				states := nativeDistributedMessages(output, tlc.ECTLCStatePrint2)
				if len(states) != len(want) {
					t.Fatalf("trace length %d, want %d", len(states), len(want))
				}
				for i, state := range states {
					header, body, ok := strings.Cut(state, "\n")
					if !ok || !strings.HasPrefix(header, fmt.Sprintf("%d: ", i+1)) || strings.TrimSpace(body) != want[i] {
						t.Fatalf("trace state %d = %q, want %q", i+1, state, want[i])
					}
				}
			})
		}
	}
}

func nativeDistributedMessages(output string, code int) []string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?s)@!@!@STARTMSG %d:\d+ @!@!@\n(.*?)\n@!@!@ENDMSG %d @!@!@`, code, code))
	var messages []string
	for _, match := range pattern.FindAllStringSubmatch(output, -1) {
		messages = append(messages, match[1])
	}
	return messages
}

func runNativeDistributedTraceModel(t *testing.T, model string, remote bool) string {
	t.Helper()
	directory, err := filepath.Abs(filepath.Join("tlc/test_vectors/models", model))
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
	// Watchdog only, with unchanged fixtures and no exploration cutoff.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	var roles []*nativeDistributedTestProcess
	defer func() {
		cancel()
		for _, process := range roles {
			if !process.joined {
				<-process.done
			}
		}
	}()
	start := func(role string, args ...string) *nativeDistributedTestProcess {
		commandArgs := []string{"-test.run=^TestNativeDistributedProcessHelper$", "--", role,
			fmt.Sprintf("-Dtlc2.tool.distributed.TLCServer.port=%d", port),
			"-Dtlago.distributed.bindHost=127.0.0.1", "-Dtlago.distributed.advertiseHost=127.0.0.1"}
		commandArgs = append(commandArgs, args...)
		command := exec.CommandContext(ctx, os.Args[0], commandArgs...)
		command.Dir = directory
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
	spec := model
	if model == "TSnapShot" {
		spec = "MC"
	}
	serverArgs := []string{"-tool", "-deadlock", "-metadir", t.TempDir(), spec}
	if remote {
		serverArgs = append([]string{"-Dtlc2.tool.distributed.TLCServer.expectedFPSetCount=1"}, serverArgs...)
	}
	server := start("server", serverArgs...)
	if remote {
		start("fpserver", "-Dtlc2.tool.fp.FPSet.impl=tlc2.tool.fp.MemFPSet", "127.0.0.1")
	}
	start("worker", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
	for _, process := range roles {
		err := <-process.done
		process.joined = true
		if err != nil {
			t.Fatalf("%s exited with %v; output:\n%s", process.output.role, err, process.output.String())
		}
		// Java's recorder is shared by all roles in its single process. In
		// the native harness, retain that scope across each child's output.
		output := process.output.String()
		if len(nativeDistributedMessages(output, tlc.ECGeneral)) != 0 {
			t.Fatalf("%s recorded GENERAL", process.output.role)
		}
		if strings.Contains(output, "unexpected EOF") {
			t.Fatalf("%s shutdown lost an accepted RPC reply", process.output.role)
		}
		t.Logf("%s exited normally", process.output.role)
	}
	return server.output.String()
}
