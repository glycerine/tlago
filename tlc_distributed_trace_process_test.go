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
	return runNativeDistributedModel(t, model, remote, false)
}

// Source-harness mode retains the default worker count and the original Ant
// off-heap/512 KiB profile, and starts the worker before the coordinator.
func runNativeDistributedModel(t *testing.T, model string, remote, sourceHarness bool) string {
	t.Helper()
	return runNativeDistributedModelWithCheckFailure(t, model, remote, sourceHarness, "")
}

// Fault mode is native-only; original model checks keep their zero-GENERAL gate.
func runNativeDistributedModelWithCheckFailure(t *testing.T, model string, remote, sourceHarness bool, checkFailure string) string {
	t.Helper()
	if checkFailure != "" && (!remote || sourceHarness || model != "EWD840") {
		t.Fatal("final-check fault requires the native remote EWD840 model")
	}
	survivingCheckHost := checkFailure == "reply-loss-survivor" || checkFailure == "reply-loss-survivor-lsb" || checkFailure == "reply-loss-survivor-msb" || checkFailure == "reply-loss-survivor-offheap"
	fingerprintImplementation := "tlc2.tool.fp.MemFPSet"
	if checkFailure == "reply-loss-survivor-lsb" {
		fingerprintImplementation = "tlc2.tool.fp.LSBDiskFPSet"
	} else if checkFailure == "reply-loss-survivor-msb" {
		fingerprintImplementation = "tlc2.tool.fp.MSBDiskFPSet"
	} else if checkFailure == "reply-loss-survivor-offheap" {
		fingerprintImplementation = "tlc2.tool.fp.OffHeapDiskFPSet"
	}
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
		if role == "fpserver-check-reply-loss" || (survivingCheckHost && role == "fpserver-check-survivor") {
			// Separate hosts own separate storage, including simultaneous starts
			// and a killed role that cannot perform its normal cleanup.
			command.Env = append(command.Env, "TMPDIR="+t.TempDir())
			if fingerprintImplementation != "tlc2.tool.fp.MemFPSet" {
				command.Env = append(command.Env, "GOMEMLIMIT=64MiB")
				if fingerprintImplementation == "tlc2.tool.fp.OffHeapDiskFPSet" {
					// Two children per host must evict before the roughly 28k-entry
					// model partitions complete, retaining both disk and memory entries.
					command.Env = append(command.Env, "TLAGO_MAX_DIRECT_MEMORY=256k")
				}
			}
		}
		if sourceHarness {
			command.Env = append(command.Env, "TMPDIR="+t.TempDir(),
				tlc.FPSetImplProperty+"=tlc2.tool.fp.OffHeapDiskFPSet", "TLAGO_MAX_DIRECT_MEMORY=512k")
		}
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
	if model == "EWD840" {
		spec = "MC06"
	}
	serverArgs := []string{"-tool", "-deadlock", "-metadir", t.TempDir(), spec}
	if remote {
		count := 1
		if survivingCheckHost {
			count = 2
		}
		serverArgs = append([]string{fmt.Sprintf("-Dtlc2.tool.distributed.TLCServer.expectedFPSetCount=%d", count)}, serverArgs...)
	}
	if sourceHarness {
		start("worker", "127.0.0.1")
	}
	server := start("server", serverArgs...)
	var fingerprint, survivor *nativeDistributedTestProcess
	if remote {
		if sourceHarness {
			start("fpserver", "127.0.0.1")
		} else {
			role := "fpserver"
			if checkFailure != "" {
				role = "fpserver-check-io-" + checkFailure
				if checkFailure == "reply-loss" || survivingCheckHost {
					role = "fpserver-check-reply-loss"
				}
			}
			fingerprint = start(role, "-Dtlc2.tool.fp.FPSet.impl="+fingerprintImplementation, "127.0.0.1")
		}
	}
	if survivingCheckHost {
		// Fix slot order before starting the second host. The source statistics
		// loop skips a failed slot after reassignment rather than retrying it.
		waitForNativeDistributedMarker(t, ctx, server, "first fingerprint registration", func() bool {
			return len(nativeDistributedMessages(server.output.String(), tlc.ECTLCDistributedServerFPSetRegistered)) == 1
		})
		survivor = start("fpserver-check-survivor", "-Dtlc2.tool.fp.FPSet.impl="+fingerprintImplementation, "127.0.0.1")
	}
	if !sourceHarness {
		start("worker", "-Dtlc2.tool.distributed.TLCWorker.threadCount=1", "127.0.0.1")
	}
	if survivingCheckHost {
		waitForNativeFinalFingerprintCheckPartition(t, ctx, server, fingerprint)
	} else if checkFailure == "reply-loss" {
		waitForNativeFinalFingerprintCheck(t, ctx, server, fingerprint)
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
		// Java's recorder is shared by all roles in its single process. In
		// the native harness, retain that scope across each child's output.
		output := process.output.String()
		expectedGeneral := 0
		if checkFailure != "" && process == server {
			expectedGeneral = 1
		}
		messages := nativeDistributedMessages(output, tlc.ECGeneral)
		expectedFailure := nativeFinalFingerprintCheckFailure
		if checkFailure == "reply-loss" || survivingCheckHost {
			expectedFailure = "unexpected EOF"
		}
		if len(messages) != expectedGeneral || (expectedGeneral == 1 && !strings.Contains(messages[0], expectedFailure)) {
			t.Fatalf("%s GENERAL messages = %q, want %d", process.output.role, messages, expectedGeneral)
		}
		if checkFailure != "" && strings.HasPrefix(process.output.role, "fpserver-check-io-") {
			markers := []string{"NATIVE_FINAL_FP_CHECK_COUNT=114942", "NATIVE_FINAL_FP_STATES_SEEN=", "NATIVE_FINAL_FP_EXIT_CLEANUP=true COUNT=114942"}
			previous := -1
			for _, marker := range markers {
				index := strings.Index(output, marker)
				if strings.Count(output, marker) != 1 || index <= previous {
					t.Fatalf("final fingerprint operations missing, repeated or reordered: %s", output)
				}
				previous = index
			}
		}
		expectedEOF := 0
		if (checkFailure == "reply-loss" || survivingCheckHost) && process == server {
			expectedEOF = 1
		}
		if strings.Count(output, "unexpected EOF") != expectedEOF {
			t.Fatalf("%s lost an unexpected RPC reply", process.output.role)
		}
		t.Logf("%s exited normally", process.output.role)
	}
	if survivingCheckHost {
		return server.output.String() + "\n" + survivor.output.String() + "\n" + fingerprint.output.String()
	}
	return server.output.String()
}
