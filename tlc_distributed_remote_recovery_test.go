package tlago

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
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

// This exercises source-supported recovery with already registered endpoints.
// It does not change Java's CLI recovery-before-registration limitation or earn
// credit for its assumption-disabled distributed model harness.
func TestNativeDistributedRemoteCheckpointRestart(t *testing.T) {
	for _, backend := range []string{"mem", "lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			checkNativeDistributedRemoteCheckpointRestart(t, backend, 1, nativeRemoteCheckpointNone)
		})
	}
}

func TestNativeDistributedRemoteCheckpointRestartMultipleWorkers(t *testing.T) {
	for _, backend := range []string{"mem", "lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			checkNativeDistributedRemoteCheckpointRestart(t, backend, 2, nativeRemoteCheckpointNone)
		})
	}
}

// The remote store commits real checkpoint files before closing its connection.
// Recovery must use those files, even though the coordinator never got the reply.
func TestNativeDistributedRemoteCheckpointCommitReplyLoss(t *testing.T) {
	for _, backend := range []string{"mem", "lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			checkNativeDistributedRemoteCheckpointRestart(t, backend, 1, nativeRemoteCheckpointCommitReplyLoss)
		})
	}
}

func TestNativeDistributedRemoteCheckpointCommitReplyLossMultipleWorkers(t *testing.T) {
	for _, backend := range []string{"mem", "lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			checkNativeDistributedRemoteCheckpointRestart(t, backend, 2, nativeRemoteCheckpointCommitReplyLoss)
		})
	}
}

// Lose the acknowledgement after storage restores its committed checkpoint.
// The following size query triggers source failover before workers reconnect.
func TestNativeDistributedRemoteCheckpointRecoverReplyLoss(t *testing.T) {
	for _, backend := range []string{"mem", "lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			checkNativeDistributedRemoteCheckpointRestart(t, backend, 1, nativeRemoteCheckpointRecoverReplyLoss)
		})
	}
}

type nativeRemoteCheckpointFault int

const (
	nativeRemoteCheckpointNone nativeRemoteCheckpointFault = iota
	nativeRemoteCheckpointCommitReplyLoss
	nativeRemoteCheckpointMissingDiskSnapshot
	nativeRemoteCheckpointRecoverReplyLoss
	nativeRemoteCheckpointMissingMemSnapshot
	nativeRemoteCheckpointTraceToIntern
	nativeRemoteCheckpointTruncatedQueue
	nativeRemoteCheckpointTruncatedTrace
)

// Unlike nested disk recovery, a direct memory-store I/O failure is reported
// and ignored by the source manager before recovering the next registration.
func TestNativeDistributedRemoteCheckpointMissingMemSnapshot(t *testing.T) {
	checkNativeDistributedRemoteCheckpointRestart(t, "mem", 1, nativeRemoteCheckpointMissingMemSnapshot)
}

// Queue recovery precedes every remote fingerprint recovery and publication.
func TestNativeDistributedRemoteCheckpointTruncatedQueue(t *testing.T) {
	checkNativeDistributedRemoteCheckpointRestart(t, "lsb", 1, nativeRemoteCheckpointTruncatedQueue)
}

// Trace recovery must win over an independently missing queue checkpoint.
func TestNativeDistributedRemoteCheckpointTruncatedTrace(t *testing.T) {
	checkNativeDistributedRemoteCheckpointRestart(t, "lsb", 1, nativeRemoteCheckpointTruncatedTrace)
}

func TestNativeDistributedRemoteCheckpointMissingDiskSnapshot(t *testing.T) {
	for _, backend := range []string{"lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			checkNativeDistributedRemoteCheckpointRestart(t, backend, 1, nativeRemoteCheckpointMissingDiskSnapshot)
		})
	}
}

func checkNativeDistributedRemoteCheckpointRestart(t *testing.T, backend string, workerThreads int, fault nativeRemoteCheckpointFault, debuggerSetup ...func(*exec.Cmd)) {
	commitReplyLoss := fault == nativeRemoteCheckpointCommitReplyLoss
	recoverReplyLoss := fault == nativeRemoteCheckpointRecoverReplyLoss
	missingMemSnapshot := fault == nativeRemoteCheckpointMissingMemSnapshot
	traceToIntern := fault == nativeRemoteCheckpointTraceToIntern
	truncatedQueue := fault == nativeRemoteCheckpointTruncatedQueue
	truncatedTrace := fault == nativeRemoteCheckpointTruncatedTrace
	corruptCoordinator := truncatedQueue || truncatedTrace
	var debugger, debuggerScript string
	if traceToIntern {
		var err error
		debugger, err = exec.LookPath("gdb")
		if err != nil {
			t.Fatal("external checkpoint boundary check requires gdb")
		}
		source, err := filepath.Abs("tlc/distributed.go")
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		var line int
		for index, text := range strings.Split(string(data), "\n") {
			if strings.Contains(text, "if err := s.serverInternTable().CommitChkpt(s.Metadir)") {
				line = index + 1
			}
		}
		if line == 0 {
			t.Fatal("intern checkpoint caller line absent")
		}
		debuggerScript = filepath.Join(t.TempDir(), "checkpoint.gdb")
		script := fmt.Sprintf("set pagination off\nset confirm off\nset debuginfod enabled off\nhandle SIGURG nostop noprint pass\nhandle SIGPIPE nostop noprint pass\nbreak %s:%d\nignore 1 1\nstarti\ninfo inferiors\ncontinue\nprintf \"NATIVE_TRACE_INTERN_PC=%%p\\n\", $pc\ninfo line *$pc\nx/i $pc\ndisassemble 'github.com/glycerine/tlago/tlc.(*TLCServer).Checkpoint'\nkill\nprintf \"NATIVE_TRACE_INTERN_KILLED\\n\"\nquit\n", source, line)
		if err := os.WriteFile(debuggerScript, []byte(script), 0o600); err != nil {
			t.Fatal(err)
		}
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	var processes []*nativeDistributedTestProcess
	defer func() {
		cancel()
		for _, process := range processes {
			if !process.joined {
				<-process.done
				process.joined = true
			}
		}
	}()
	common := []string{fmt.Sprintf("-Dtlc2.tool.distributed.TLCServer.port=%d", port), "-Dtlago.distributed.bindHost=127.0.0.1", "-Dtlago.distributed.advertiseHost=127.0.0.1", "-Dtlc2.tool.distributed.TLCServer.expectedFPSetCount=2"}
	start := func(label, role string, environment []string, args ...string) *nativeDistributedTestProcess {
		commandArgs := append([]string{"-test.run=^TestNativeDistributedProcessHelper$", "--", role}, common...)
		commandArgs = append(commandArgs, args...)
		command := exec.CommandContext(ctx, os.Args[0], commandArgs...)
		if traceToIntern && role == "checkpoint-mid-run" {
			debugArgs := append([]string{"-q", "-batch", "-x", debuggerScript, "--args", os.Args[0]}, commandArgs...)
			command = exec.CommandContext(ctx, debugger, debugArgs...)
			if len(debuggerSetup) > 0 {
				debuggerSetup[0](command)
			}
			environment = append(environment, "TLAGO_CHECKPOINT_TRACE_INTERN_BOUNDARY=1", "DEBUGINFOD_URLS=")
		}
		command.Dir = model
		command.Env = append(os.Environ(), "TLAGO_NATIVE_DISTRIBUTED_PROCESS_HELPER=1")
		command.Env = append(command.Env, environment...)
		output := &nativeDistributedTestLog{test: t, role: label}
		command.Stdout, command.Stderr = output, output
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		process := &nativeDistributedTestProcess{command: command, output: output, done: make(chan error, 1)}
		go func() { process.done <- command.Wait() }()
		processes = append(processes, process)
		return process
	}
	waitMarker := func(process *nativeDistributedTestProcess, marker string) {
		t.Helper()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for !strings.Contains(process.output.String(), marker) {
			select {
			case err := <-process.done:
				process.joined = true
				t.Fatalf("%s exited before %s: %v\n%s", process.output.role, marker, err, process.output.String())
			case <-ctx.Done():
				t.Fatal("remote checkpoint watchdog expired")
			case <-ticker.C:
			}
		}
	}
	directories := []string{t.TempDir(), t.TempDir()}
	startHosts := func(prefix string) ([]*nativeDistributedTestProcess, []string) {
		var hosts []*nativeDistributedTestProcess
		var addresses []string
		for i, directory := range directories {
			environment := []string{"TLAGO_CHECKPOINT_FP_DIRECTORY=" + directory, "TLAGO_CHECKPOINT_FP_STORAGE=" + backend}
			if commitReplyLoss && prefix == "original" && i == 0 {
				environment = append(environment, "TLAGO_CHECKPOINT_COMMIT_REPLY_LOSS=1")
			}
			if recoverReplyLoss && prefix == "restarted" && i == 0 {
				environment = append(environment, "TLAGO_CHECKPOINT_RECOVER_REPLY_LOSS=1")
			}
			host := start(fmt.Sprintf("%s-fp-%d", prefix, i), "checkpoint-fp-host", environment)
			waitMarker(host, "NATIVE_CHECKPOINT_FP_READY=")
			match := regexp.MustCompile(`NATIVE_CHECKPOINT_FP_READY=([^\s]+)`).FindStringSubmatch(host.output.String())
			if len(match) != 2 {
				t.Fatal("fingerprint readiness address absent")
			}
			client, err := tlc.DialFingerprintEndpoint(match[1], "primary")
			if err != nil {
				t.Fatal(err)
			}
			size, failure := client.Size()
			_ = client.CloseConnection()
			if failure != nil || size != 0 {
				t.Fatalf("new fingerprint host is not empty: %d/%v", size, failure)
			}
			hosts, addresses = append(hosts, host), append(addresses, match[1])
		}
		return hosts, addresses
	}
	hosts, addresses := startHosts("original")
	environment := []string{"TLAGO_REGISTERED_FP_ENDPOINTS=" + strings.Join(addresses, ",")}
	producer := start("checkpoint-producer", "checkpoint-mid-run", environment, "-tool", "-deadlock", "-metadir", t.TempDir(), "MC06")
	worker := start("original-worker", "worker", nil, fmt.Sprintf("-Dtlc2.tool.distributed.TLCWorker.threadCount=%d", workerThreads), "127.0.0.1")
	if err := <-producer.done; err != nil {
		producer.joined = true
		t.Fatalf("checkpoint producer failed: %v\n%s", err, producer.output.String())
	}
	producer.joined = true
	output := producer.output.String()
	path := regexp.MustCompile(`NATIVE_CHECKPOINT_PATH=([^\r\n]+)`).FindStringSubmatch(output)
	counts := regexp.MustCompile(`NATIVE_CHECKPOINT_COUNTS=(\d+),(\d+)`).FindStringSubmatch(output)
	if traceToIntern {
		begins := regexp.MustCompile(`NATIVE_CHECKPOINT_BEGIN_COUNTS=(\d+),(\d+)`).FindAllStringSubmatch(output, -1)
		if len(begins) != 2 {
			t.Fatal("debugger did not observe exactly two real model checkpoints")
		}
		counts = begins[1]
		checkNativeDistributedTraceInternPC(t, output)
	}
	if len(path) != 2 || len(counts) != 3 {
		t.Fatal("checkpoint path/counts absent")
	}
	if !filepath.IsAbs(path[1]) {
		path[1] = filepath.Join(model, path[1])
	}
	distinct, _ := strconv.Atoi(counts[1])
	queued, _ := strconv.Atoi(counts[2])
	if distinct <= 16384 || queued <= 0 || len(nativeDistributedMessages(output, tlc.ECTLCCheckpointEnd)) != 1 || len(nativeDistributedMessages(output, tlc.ECGeneral)) != 0 || len(nativeDistributedMessages(output, tlc.ECTLCFinished)) != 0 {
		t.Fatal("producer did not checkpoint an unfinished successor frontier")
	}
	expectedStarts := 1
	if traceToIntern {
		expectedStarts = 2
		baseline := regexp.MustCompile(`NATIVE_CHECKPOINT_BASELINE=(\d+),(\d+)`).FindStringSubmatch(output)
		if len(baseline) != 3 {
			t.Fatal("debugger interruption lacks a completed baseline")
		}
		baselineDistinct, _ := strconv.Atoi(baseline[1])
		if distinct <= baselineDistinct || strings.Contains(output, "NATIVE_CHECKPOINT_COUNTS=") {
			t.Fatal("second checkpoint did not advance or returned after debugger interruption")
		}
		for _, name := range []string{"queue", "MC06.st", "vars"} {
			committed, err := os.ReadFile(filepath.Join(path[1], name+".chkpt"))
			before, beforeErr := os.ReadFile(filepath.Join(path[1], name+".chkpt.baseline"))
			pending, pendingErr := os.ReadFile(filepath.Join(path[1], name+".tmp"))
			if err != nil || beforeErr != nil || bytes.Equal(committed, before) != (name == "vars") {
				t.Fatalf("unexpected checkpoint generation for %s: %v/%v", name, err, beforeErr)
			}
			if name == "vars" {
				if pendingErr != nil || len(pending) == 0 {
					t.Fatal("intern pending generation absent at caller boundary")
				}
			} else if !os.IsNotExist(pendingErr) {
				t.Fatal("queue/trace temporary was not promoted before the caller boundary")
			}
		}
	}
	if len(nativeDistributedMessages(output, tlc.ECTLCCheckpointStart)) != expectedStarts || len(nativeDistributedMessages(output, tlc.ECTLCDistributedServerFPSetRegistered)) != 2 || len(nativeDistributedMessages(output, tlc.ECTLCDistributedWorkerRegistered)) != workerThreads {
		t.Fatal("producer checkpoint/registration sequence differs from the real two-store model")
	}
	if commitReplyLoss {
		const diagnostic = "Error: Failed to checkpoint the fingerprint server at checkpoint-host-0. This server might be down."
		marker := "NATIVE_CHECKPOINT_FP_COMMIT_REPLY_LOST=" + filepath.Base(filepath.Clean(path[1]))
		if strings.Count(output, diagnostic) != 1 || strings.Count(hosts[0].output.String(), marker+"\n") != 1 {
			t.Fatalf("completed commit must lose exactly one reply and report its original host\n%s\n%s", output, hosts[0].output.String())
		}
		if strings.Count(output, "NATIVE_CHECKPOINT_ALIVE_FINGERPRINTS=2\n") != 1 {
			t.Fatalf("checkpoint I/O failure must preserve both registrations\n%s", output)
		}
	}
	queueFile, err := os.Open(filepath.Join(path[1], "queue.chkpt"))
	if err != nil {
		t.Fatal(err)
	}
	queueInput, err := tlc.NewValueInputStreamWithGlobalCompression(queueFile)
	if err != nil {
		_ = queueFile.Close()
		t.Fatal(err)
	}
	queueCount, readErr := queueInput.ReadInt()
	closeErr := queueInput.Close()
	if readErr != nil || closeErr != nil || int(queueCount) != queued {
		t.Fatalf("persisted queue count differs from checkpoint frontier: %d/%v/%v", queueCount, readErr, closeErr)
	}
	type snapshot struct {
		filename string
		data     []byte
		highBits uint64
	}
	snapshots := make([][]snapshot, len(hosts))
	basename := filepath.Base(filepath.Clean(path[1]))
	var committed int
	for i, directory := range directories {
		children := 1
		if backend != "mem" {
			children = 2
		}
		for child := 0; child < children; child++ {
			filename := basename + ".fp.chkpt"
			var highBits uint64
			if backend != "mem" {
				filename = fmt.Sprintf("%s_%d.fp.chkpt", basename, child)
				// Disk children store normalized low 63 bits. Their MultiFPSet
				// child index preserves the original fingerprint's high bit.
				highBits = uint64(child) << 63
			}
			data, err := os.ReadFile(filepath.Join(directory, filename))
			if err != nil || len(data) == 0 || len(data)%8 != 0 {
				t.Fatalf("remote committed snapshot %s: %v/%d", filename, err, len(data))
			}
			snapshots[i] = append(snapshots[i], snapshot{filename, data, highBits})
			committed += len(data) / 8
		}
	}
	if committed != distinct {
		t.Fatal("committed remote membership differs from checkpoint frontier")
	}
	for _, process := range append(hosts, worker) {
		killErr := process.command.Process.Kill()
		if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
			t.Fatal(killErr)
		}
		err := <-process.done
		process.joined = true
		if process != worker && (killErr != nil || err == nil) {
			t.Fatal("original fingerprint host did not crash")
		}
	}
	t.Logf("committed frontier %d/%d; original coordinator, worker and both fingerprint hosts are gone", distinct, queued)
	for i, directory := range directories {
		for _, snapshot := range snapshots[i] {
			data, err := os.ReadFile(filepath.Join(directory, snapshot.filename))
			if err != nil || !bytes.Equal(data, snapshot.data) {
				t.Fatal("process crash changed committed fingerprint bytes")
			}
		}
	}
	missingDiskSnapshot := fault == nativeRemoteCheckpointMissingDiskSnapshot
	if missingDiskSnapshot || missingMemSnapshot {
		if backend == "mem" {
			if !missingMemSnapshot {
				t.Fatal("missing disk snapshot case requires nested disk storage")
			}
		} else if missingMemSnapshot {
			t.Fatal("missing memory snapshot case requires direct memory storage")
		}
		if err := os.Remove(filepath.Join(directories[0], snapshots[0][0].filename)); err != nil {
			t.Fatal(err)
		}
	}
	if truncatedQueue {
		// Retain only part of the serialized queue length. Do not regenerate
		// any state or alter trace/intern/fingerprint checkpoint generations.
		if err := os.Truncate(filepath.Join(path[1], "queue.chkpt"), 1); err != nil {
			t.Fatal(err)
		}
	}
	if truncatedTrace {
		if err := os.Truncate(filepath.Join(path[1], "MC06.st.chkpt"), 1); err != nil {
			t.Fatal(err)
		}
		// Missing queue would produce a file-open error if recovery were
		// reordered. Require the trace EOF instead and retain the queue bytes.
		if err := os.Rename(filepath.Join(path[1], "queue.chkpt"), filepath.Join(path[1], "queue.chkpt.unreached")); err != nil {
			t.Fatal(err)
		}
	}
	coordinatorSnapshots := make(map[string][]byte)
	entries, err := os.ReadDir(path[1])
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".chkpt") || entry.Name() == "queue.chkpt.unreached" {
			data, err := os.ReadFile(filepath.Join(path[1], entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			coordinatorSnapshots[entry.Name()] = data
		}
	}
	restarted, addresses := startHosts("restarted")
	server := start("recovered-coordinator", "registered-fp-recovery", []string{"TLAGO_REGISTERED_FP_ENDPOINTS=" + strings.Join(addresses, ",")}, "-tool", "-deadlock", "-recover", path[1], "MC06")
	if corruptCoordinator {
		for _, process := range append([]*nativeDistributedTestProcess{server}, restarted...) {
			err := <-process.done
			process.joined = true
			if err != nil {
				t.Fatalf("%s failed caught-recovery shutdown: %v\n%s", process.output.role, err, process.output.String())
			}
		}
		output := server.output.String()
		failure := nativeDistributedMessages(output, tlc.ECGeneral)
		if len(failure) != 1 || !strings.Contains(failure[0], "EOF") || len(nativeDistributedMessages(output, tlc.ECTLCCheckpointRecoverStart)) != 1 || len(nativeDistributedMessages(output, tlc.ECTLCDistributedServerFPSetRegistered)) != 2 {
			t.Fatalf("corrupt coordinator snapshot lost its recovery failure: %s", output)
		}
		for _, code := range []int{tlc.ECTLCCheckpointRecoverEnd, tlc.ECTLCComputingInit, tlc.ECTLCDistributedServerRunning, tlc.ECTLCDistributedWorkerRegistered, tlc.ECTLCFinished, tlc.ECTLCStats} {
			if len(nativeDistributedMessages(output, code)) != 0 {
				t.Fatalf("coordinator recovery failure crossed phase %d: %s", code, output)
			}
		}
		for i, process := range restarted {
			if strings.Contains(process.output.String(), "NATIVE_CHECKPOINT_FP_RECOVER=") || strings.Count(process.output.String(), "NATIVE_CHECKPOINT_FP_EXIT_SIZE=0\n") != 1 || len(nativeDistributedMessages(process.output.String(), tlc.ECGeneral)) != 0 {
				t.Fatalf("coordinator failure reached fingerprint recovery or broke shutdown: %s", process.output.String())
			}
			for _, snapshot := range snapshots[i] {
				data, err := os.ReadFile(filepath.Join(directories[i], snapshot.filename))
				if err != nil || !bytes.Equal(data, snapshot.data) {
					t.Fatal("coordinator recovery failure changed a remote committed snapshot")
				}
			}
		}
		for filename, expected := range coordinatorSnapshots {
			data, err := os.ReadFile(filepath.Join(path[1], filename))
			if err != nil || !bytes.Equal(data, expected) {
				t.Fatalf("coordinator recovery failure changed checkpoint %s: %v", filename, err)
			}
		}
		if truncatedTrace {
			if _, err := os.Stat(filepath.Join(path[1], "queue.chkpt")); !os.IsNotExist(err) {
				t.Fatal("trace failure regenerated the missing queue checkpoint")
			}
		}
		t.Log("corrupt coordinator snapshot stops before remote recovery and publication; all roles join and checkpoint bytes survive")
		return
	}
	if missingDiskSnapshot {
		// Source main reports this failure and closes the server with cleanup=false;
		// it does not turn the caught exception into a process failure status.
		for _, process := range append([]*nativeDistributedTestProcess{server}, restarted...) {
			err := <-process.done
			process.joined = true
			if err != nil {
				t.Fatalf("%s did not follow source caught-failure shutdown: %v\n%s", process.output.role, err, process.output.String())
			}
		}
		output := server.output.String()
		failure := nativeDistributedMessages(output, tlc.ECGeneral)
		if len(failure) != 1 || !strings.Contains(failure[0], snapshots[0][0].filename) || len(nativeDistributedMessages(output, tlc.ECTLCCheckpointRecoverStart)) != 1 || len(nativeDistributedMessages(output, tlc.ECTLCDistributedServerFPSetRegistered)) != 2 {
			t.Fatalf("missing disk child did not report its recovery failure: %s", output)
		}
		for _, code := range []int{tlc.ECTLCCheckpointRecoverEnd, tlc.ECTLCComputingInit, tlc.ECTLCDistributedServerRunning, tlc.ECTLCDistributedWorkerRegistered, tlc.ECTLCFinished, tlc.ECTLCStats} {
			if len(nativeDistributedMessages(output, code)) != 0 {
				t.Fatalf("failed recovery crossed phase %d: %s", code, output)
			}
		}
		marker := "NATIVE_CHECKPOINT_FP_RECOVER=" + basename + "\n"
		if strings.Count(restarted[0].output.String(), marker) != 1 || strings.Contains(restarted[1].output.String(), marker) {
			t.Fatal("nested disk failure retried or recovered a later host")
		}
		for i, process := range restarted {
			var expected int
			if i == 0 {
				// Native child ownership joins the sibling's recovery before
				// propagating the failed child's operation error.
				expected = len(snapshots[0][1].data) / 8
			}
			if strings.Count(process.output.String(), fmt.Sprintf("NATIVE_CHECKPOINT_FP_EXIT_SIZE=%d\n", expected)) != 1 || len(nativeDistributedMessages(process.output.String(), tlc.ECGeneral)) != 0 {
				t.Fatalf("failed recovery changed partial storage/shutdown for host %d: %s", i, process.output.String())
			}
			for child, snapshot := range snapshots[i] {
				data, err := os.ReadFile(filepath.Join(directories[i], snapshot.filename))
				if i == 0 && child == 0 {
					if !os.IsNotExist(err) {
						t.Fatal("failed recovery recreated the missing committed child")
					}
				} else if err != nil || !bytes.Equal(data, snapshot.data) {
					t.Fatal("failed recovery changed a retained committed child")
				}
			}
		}
		for filename, expected := range coordinatorSnapshots {
			data, err := os.ReadFile(filepath.Join(path[1], filename))
			if err != nil || !bytes.Equal(data, expected) {
				t.Fatalf("failed recovery changed coordinator checkpoint %s: %v", filename, err)
			}
		}
		t.Log("missing committed disk child stops recovery before publication; later host remains empty and all retained checkpoints survive")
		return
	}
	waitMarker(server, fmt.Sprintf("@!@!@STARTMSG %d:", tlc.ECTLCCheckpointRecoverEnd))
	output = server.output.String()
	reportedDistinct := distinct
	if missingMemSnapshot {
		reportedDistinct = len(snapshots[1][0].data) / 8
		const warning = "Error: Failed to checkpoint the fingerprint server at checkpoint-host-0. This server might be down."
		if strings.Count(output, warning) != 1 || strings.Contains(output, "Warning: Failed to connect") {
			t.Fatal("missing memory snapshot did not retain source warning and routing")
		}
		for _, process := range restarted {
			if strings.Count(process.output.String(), "NATIVE_CHECKPOINT_FP_RECOVER="+basename+"\n") != 1 {
				t.Fatal("missing memory snapshot was replayed or stopped later recovery")
			}
		}
	}
	if recoverReplyLoss {
		var survivorSnapshotSize int
		for _, snapshot := range snapshots[1] {
			survivorSnapshotSize += len(snapshot.data) / 8
		}
		// Recovery catches I/O without reassignment. The following size call
		// aliases the failed slot to the survivor, but does not retry that slot.
		// This first count therefore includes only the original survivor slot.
		reportedDistinct = survivorSnapshotSize
		marker := "NATIVE_CHECKPOINT_FP_RECOVER_REPLY_LOST=" + filepath.Base(filepath.Clean(path[1]))
		const warning = "Error: Failed to checkpoint the fingerprint server at checkpoint-host-0. This server might be down."
		if strings.Count(restarted[0].output.String(), marker+"\n") != 1 || strings.Count(output, warning) != 1 || strings.Count(restarted[0].output.String(), "NATIVE_CHECKPOINT_FP_RECOVER=") != 1 || !strings.Contains(output, "to the fp server at checkpoint-host-0.") {
			t.Fatal("completed recovery reply loss was replayed, unreported or failed to reach source size failover")
		}
	}
	if !strings.Contains(output, fmt.Sprintf("Recovery completed. %d states examined. %d states on queue.", reportedDistinct, queued)) || len(nativeDistributedMessages(output, tlc.ECTLCComputingInit)) != 0 {
		t.Fatal("registered remote recovery counts differ or initialization was regenerated")
	}
	var recovered uint64
	for i, address := range addresses {
		client, err := tlc.DialFingerprintEndpoint(address, "primary")
		if err != nil {
			t.Fatal(err)
		}
		size, failure := client.Size()
		fingerprints := tlc.NewLongVec()
		var expectedSize uint64
		for _, snapshot := range snapshots[i] {
			expectedSize += uint64(len(snapshot.data) / 8)
			for offset := 0; offset < len(snapshot.data); offset += 8 {
				fingerprints.AddElement(int64(binary.BigEndian.Uint64(snapshot.data[offset:offset+8]) | snapshot.highBits))
			}
		}
		var expectedMissing int
		if missingMemSnapshot && i == 0 {
			expectedSize = 0
			expectedMissing = fingerprints.Size()
		}
		missing, membershipErr := client.ContainsBlock(fingerprints)
		_ = client.CloseConnection()
		if failure != nil || size != expectedSize || membershipErr != nil || missing == nil || missing.TrueCount() != expectedMissing {
			t.Fatalf("restarted partition %d did not retain its complete snapshot: size %d/%v, membership %v/%v", i, size, failure, missing, membershipErr)
		}
		recovered += size
	}
	if missingMemSnapshot {
		if _, err := os.Stat(filepath.Join(directories[0], snapshots[0][0].filename)); !os.IsNotExist(err) {
			t.Fatal("recovery recreated the missing committed memory snapshot")
		}
	}
	expectedRecovered := distinct
	if missingMemSnapshot {
		expectedRecovered = reportedDistinct
	}
	if recovered != uint64(expectedRecovered) {
		t.Fatal("restarted stores did not recover exactly committed membership before worker startup")
	}
	if traceToIntern {
		pending, err := os.ReadFile(filepath.Join(path[1], "vars.tmp"))
		if err != nil || len(pending) == 0 {
			t.Fatal("fresh recovery promoted or lost the pending intern checkpoint")
		}
		for name, expected := range coordinatorSnapshots {
			actual, err := os.ReadFile(filepath.Join(path[1], name))
			if err != nil || !bytes.Equal(actual, expected) {
				t.Fatalf("fresh recovery rewrote committed checkpoint %s: %v", name, err)
			}
		}
	}
	t.Log("restarted empty stores recovered committed membership; starting actual successor evaluation")
	if recoverReplyLoss {
		// No transport redial: the manager no longer owns the failed endpoint.
		// Its restored membership was observed above through a fresh connection.
		if err := restarted[0].command.Process.Kill(); err != nil {
			t.Fatal(err)
		}
		if err := <-restarted[0].done; err == nil {
			t.Fatal("retained recovered host did not crash")
		}
		restarted[0].joined = true
		restarted = restarted[1:]
	}
	replacement := start("replacement-worker", "worker", nil, fmt.Sprintf("-Dtlc2.tool.distributed.TLCWorker.threadCount=%d", workerThreads), "127.0.0.1")
	for _, process := range append([]*nativeDistributedTestProcess{server, replacement}, restarted...) {
		if err := <-process.done; err != nil {
			process.joined = true
			t.Fatalf("%s exited: %v\n%s", process.output.role, err, process.output.String())
		}
		process.joined = true
		if len(nativeDistributedMessages(process.output.String(), tlc.ECGeneral)) != 0 {
			t.Fatalf("%s emitted GENERAL", process.output.role)
		}
	}
	output = server.output.String()
	if workerThreads > 1 {
		checkNativeDistributedWorkerGroup(t, output, workerThreads)
	}
	stats := nativeDistributedMessages(output, tlc.ECTLCStats)
	finalDistinct := 114942
	if recoverReplyLoss {
		finalDistinct *= 2 // Source statistics count both aliased partition slots.
	}
	if len(nativeDistributedMessages(output, tlc.ECTLCFinished)) != 1 || len(stats) != 1 || !regexp.MustCompile(fmt.Sprintf(`^\d+ states generated, %d distinct states found, 0 states left on queue\.$`, finalDistinct)).MatchString(stats[0]) || len(nativeDistributedMessages(output, tlc.ECTLCCheckpointRecoverEnd)) != 1 || len(nativeDistributedMessages(output, tlc.ECTLCDistributedServerFPSetRegistered)) != 2 {
		t.Fatal("recovered model lacks original completion/count assertions")
	}
}

func checkNativeDistributedTraceInternPC(t *testing.T, output string) {
	t.Helper()
	match := regexp.MustCompile(`NATIVE_TRACE_INTERN_PC=(0x[0-9a-f]+)`).FindStringSubmatch(output)
	if len(match) != 2 || strings.Count(output, "NATIVE_TRACE_INTERN_KILLED\n") != 1 {
		t.Fatal("debugger did not stop and kill the owned inferior at the checkpoint caller")
	}
	pc, err := strconv.ParseUint(match[1][2:], 16, 64)
	if err != nil {
		t.Fatal(err)
	}
	var traceCall, internCall uint64
	assembly := regexp.MustCompile(`(?m)^\s*(?:=>\s*)?(0x[0-9a-f]+)\s+<\+\d+>:\s+(.*)$`).FindAllStringSubmatch(output, -1)
	for _, instruction := range assembly {
		address, err := strconv.ParseUint(instruction[1][2:], 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(instruction[2], "call") {
			continue
		}
		if strings.Contains(instruction[2], "(*TLCTrace).CommitChkpt") {
			traceCall = address
		}
		if strings.Contains(instruction[2], "(*InternTable).CommitChkpt") {
			internCall = address
		}
	}
	if traceCall == 0 || internCall == 0 || !(traceCall < pc && pc < internCall) {
		t.Fatalf("stopped PC %#x is not between trace %#x and intern %#x calls", pc, traceCall, internCall)
	}
	t.Logf("inspected checkpoint PC %#x after trace commit %#x and before intern commit %#x", pc, traceCall, internCall)
}

type nativeCheckpointFingerprintEndpoint struct {
	*tlc.LocalFingerprintEndpoint
	exited           chan struct{}
	once             sync.Once
	commitReplyLoss  *tlc.DistributedRPCServer
	commitOnce       sync.Once
	recoverReplyLoss func()
	recoverOnce      sync.Once
}

func (e *nativeCheckpointFingerprintEndpoint) RecoverFile(name string) error {
	fmt.Println("NATIVE_CHECKPOINT_FP_RECOVER=" + name)
	if err := e.LocalFingerprintEndpoint.RecoverFile(name); err != nil {
		return err
	}
	if e.recoverReplyLoss != nil {
		e.recoverOnce.Do(func() {
			fmt.Println("NATIVE_CHECKPOINT_FP_RECOVER_REPLY_LOST=" + name)
			e.recoverReplyLoss()
		})
	}
	return nil
}

// Keep storage/listening alive for observation while discarding the accepted
// recovery's reply. Existing coordinator clients are never redialed or replayed.
type nativeCheckpointConnectionListener struct {
	net.Listener
	mu          sync.Mutex
	connections []net.Conn
}

func (l *nativeCheckpointConnectionListener) Accept() (net.Conn, error) {
	connection, err := l.Listener.Accept()
	if err == nil {
		l.mu.Lock()
		l.connections = append(l.connections, connection)
		l.mu.Unlock()
	}
	return connection, err
}

func (l *nativeCheckpointConnectionListener) closeConnections() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, connection := range l.connections {
		_ = connection.Close()
	}
	l.connections = nil
}

func (e *nativeCheckpointFingerprintEndpoint) CommitChkptFile(name string) error {
	if err := e.LocalFingerprintEndpoint.CommitChkptFile(name); err != nil {
		return err
	}
	if e.commitReplyLoss != nil {
		e.commitOnce.Do(func() {
			fmt.Println("NATIVE_CHECKPOINT_FP_COMMIT_REPLY_LOST=" + name)
			// Close before returning to the RPC handler, after storage committed.
			// Keep the storage process alive until the parent crashes it.
			if err := e.commitReplyLoss.Close(); err != nil {
				panic(err)
			}
		})
	}
	return nil
}

func (e *nativeCheckpointFingerprintEndpoint) Exit(cleanup bool) error {
	fmt.Printf("NATIVE_CHECKPOINT_FP_EXIT_SIZE=%d\n", e.Set.Size())
	if err := e.LocalFingerprintEndpoint.Exit(cleanup); err != nil {
		return err
	}
	e.once.Do(func() { close(e.exited) })
	return nil
}
func nativeCheckpointFingerprintHost() error {
	directory := os.Getenv("TLAGO_CHECKPOINT_FP_DIRECTORY")
	if directory == "" {
		return fmt.Errorf("fingerprint directory missing")
	}
	var storage tlc.FPSet
	switch backend := os.Getenv("TLAGO_CHECKPOINT_FP_STORAGE"); backend {
	case "", "mem":
		storage = tlc.NewMemFPSet()
	case "lsb", "msb":
		implementation := "tlc2.tool.fp.LSBDiskFPSet"
		if backend == "msb" {
			implementation = "tlc2.tool.fp.MSBDiskFPSet"
		}
		config := tlc.NewFPSetConfigurationWithRatioAndImplementation(1, implementation)
		config.SetFPBits(1) // Source DistributedFPSet.main factory layout.
		// Budget storage explicitly for this native fixture; model bounds
		// and the source factory's nested routing remain unchanged.
		config.SetMemory(1 << 20)
		storage = tlc.NewFPSet(config)
	default:
		return fmt.Errorf("unknown checkpoint fingerprint storage %q", backend)
	}
	storage.Init(1, directory, "MC06")
	defer storage.Close()
	endpoint := &nativeCheckpointFingerprintEndpoint{LocalFingerprintEndpoint: tlc.NewLocalFingerprintEndpoint(storage), exited: make(chan struct{})}
	host := tlc.NewDistributedRPCServer()
	if os.Getenv("TLAGO_CHECKPOINT_COMMIT_REPLY_LOSS") == "1" {
		endpoint.commitReplyLoss = host
	}
	if err := host.RegisterFingerprint("primary", endpoint); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	done := make(chan error, 1)
	if os.Getenv("TLAGO_CHECKPOINT_RECOVER_REPLY_LOSS") == "1" {
		tracked := &nativeCheckpointConnectionListener{Listener: listener}
		listener = tracked
		endpoint.recoverReplyLoss = tracked.closeConnections
	}
	go func() { done <- host.Serve(listener) }()
	fmt.Println("NATIVE_CHECKPOINT_FP_READY=" + listener.Addr().String())
	<-endpoint.exited
	if err := host.CloseGracefully(); err != nil {
		return err
	}
	if err := <-done; err != nil && !errors.Is(err, net.ErrClosed) {
		return err
	}
	return nil
}

func nativeRegisteredFingerprintServer(app *tlc.TLCApp) (*tlc.TLCServer, error) {
	addresses := strings.Split(os.Getenv("TLAGO_REGISTERED_FP_ENDPOINTS"), ",")
	if len(addresses) != tlc.TLCServerExpectedFPSetCount() {
		return nil, fmt.Errorf("registered fingerprint count mismatch")
	}
	server, err := tlc.NewDistributedFPSetTLCServer(app, len(addresses))
	if err != nil {
		return nil, err
	}
	for i, address := range addresses {
		endpoint, err := tlc.DialFingerprintEndpoint(address, "primary")
		if err != nil {
			return nil, err
		}
		if err := server.RegisterFPSet(endpoint, fmt.Sprintf("checkpoint-host-%d", i)); err != nil {
			_ = endpoint.CloseConnection()
			return nil, err
		}
	}
	return server, nil
}
func nativeRegisteredFingerprintRecovery(args []string) error {
	args, err := tlc.ExtractDistributedStartupProperties(args)
	if err != nil {
		return err
	}
	network := tlc.NewDistributedCoordinatorNetwork("127.0.0.1", "127.0.0.1")
	defer network.Close()
	env := tlc.DistributedServerEnvironment{CreateServer: func(app *tlc.TLCApp, _ int) (*tlc.TLCServer, error) {
		server, err := nativeRegisteredFingerprintServer(app)
		if err == nil {
			server.ConfigurePublication(network.Publication())
		}
		return server, err
	}}
	_, err = RunDistributedServer(tlc.NewDistributedServerProcess(), args, env, tlc.RuntimeParameters{})
	return err
}
