package tlc

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The source's buffered stores stop at failed writes. Native resource cleanup
// must close the descriptor without flushing the failed buffer a second time.
func TestDistributedBufferedMemoryCheckpointFailure(t *testing.T) {
	strace, err := exec.LookPath("strace")
	if err != nil {
		t.Skip("native buffered checkpoint failure check requires strace")
	}
	for _, backend := range []string{"mem", "mem2"} {
		for fault := 0; fault <= 3; fault++ {
			t.Run(backend+"/"+strconv.Itoa(fault), func(t *testing.T) {
				directory := t.TempDir()
				target := filepath.Join(directory, "job.fp.tmp")
				logfile := filepath.Join(directory, "syscalls.log")
				args := []string{"-f", "-yy", "-s", "32", "-o", logfile, "-P", target, "-e", "trace=write,close", "-e", "signal=none"}
				if fault == 1 || fault == 2 {
					args = append(args, "-e", "inject=write:error=EIO:when="+strconv.Itoa(fault))
				} else if fault == 3 {
					args = append(args, "-e", "inject=close:error=EIO:when=1")
				}
				args = append(args, os.Args[0], "-test.run=^TestDistributedBufferedMemoryCheckpointProcess$", "-test.v")
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				command := exec.CommandContext(ctx, strace, args...)
				command.Env = append(os.Environ(), "TLAGO_BUFFERED_MEMORY_DIRECTORY="+directory, "TLAGO_BUFFERED_MEMORY_BACKEND="+backend, "TLAGO_BUFFERED_MEMORY_FAULT="+strconv.Itoa(fault))
				command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
				command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
				command.WaitDelay = time.Second
				t.Logf("Checking %s checkpoint fault %d (0 normal; 1 full-buffer write; 2 final flush; 3 close)", backend, fault)
				output, runErr := command.CombinedOutput()
				log, logErr := os.ReadFile(logfile)
				if runErr != nil || ctx.Err() != nil || logErr != nil || !bytes.Contains(output, []byte("BUFFERED_MEMORY_CHECKPOINT_VERIFIED")) {
					t.Fatalf("buffered checkpoint process = %v/%v/%v\n%s\n%s", runErr, ctx.Err(), logErr, output, log)
				}
				var writes, closes []string
				for _, line := range strings.Split(string(log), "\n") {
					if strings.Contains(line, "write(") {
						writes = append(writes, line)
					}
					if strings.Contains(line, "close(") {
						closes = append(closes, line)
					}
				}
				wantWrites := 2
				if fault == 1 {
					wantWrites = 1
				}
				if len(writes) != wantWrites || len(closes) != 1 || !strings.Contains(closes[0], "<"+target+">") {
					t.Fatalf("failed buffer was replayed or descriptor not closed once\n%s", log)
				}
				for i, line := range writes {
					width := []int{8192, 8}[i]
					faulted := (fault == 1 || fault == 2) && i == fault-1
					if !strings.Contains(line, "<"+target+">") || !strings.Contains(line, ", "+strconv.Itoa(width)+")") || strings.Contains(line, "(INJECTED)") != faulted {
						t.Fatalf("checkpoint buffering changed\n%s", log)
					}
					if faulted && !strings.Contains(line, "= -1 EIO") || !faulted && !strings.HasSuffix(line, "= "+strconv.Itoa(width)) {
						t.Fatalf("unexpected buffered write result\n%s", log)
					}
				}
				if fault == 3 && (!strings.Contains(closes[0], "= -1 EIO") || !strings.Contains(closes[0], "(INJECTED)")) || fault != 3 && !strings.HasSuffix(closes[0], "= 0") {
					t.Fatalf("unexpected buffered close result\n%s", log)
				}
				path := target
				if fault == 0 {
					path = filepath.Join(directory, "job.fp.chkpt")
				}
				data, err := os.ReadFile(path)
				prefix := []int{8200, 0, 8192, 8200}[fault]
				if err != nil || !bytes.Equal(data, bufferedMemoryCheckpointBytes()[:prefix]) {
					t.Fatalf("checkpoint prefix = %d bytes/%v, want %d", len(data), err, prefix)
				}
			})
		}
	}
}

func bufferedMemoryCheckpointBytes() []byte {
	data := make([]byte, 1025*8)
	for fp := 1; fp <= 1025; fp++ {
		binary.BigEndian.PutUint64(data[(fp-1)*8:], uint64(fp))
	}
	return data
}

func TestDistributedBufferedMemoryCheckpointProcess(t *testing.T) {
	directory := os.Getenv("TLAGO_BUFFERED_MEMORY_DIRECTORY")
	if directory == "" {
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fault, err := strconv.Atoi(os.Getenv("TLAGO_BUFFERED_MEMORY_FAULT"))
	if err != nil {
		t.Fatal(err)
	}
	captureFailoverToolIO(t, ToolIOTool)
	var store FPSet
	switch os.Getenv("TLAGO_BUFFERED_MEMORY_BACKEND") {
	case "mem":
		store = NewMemFPSet()
	case "mem2":
		store = NewMemFPSet2(NewFPSetConfiguration())
	default:
		t.Fatal("unknown buffered memory store")
	}
	store.Init(1, directory, "Spec")
	for fp := 1; fp <= 1025; fp++ {
		store.Put(uint64(fp))
	}
	old := []byte("previous checkpoint")
	checkpoint := filepath.Join(directory, "job.fp.chkpt")
	if err := os.WriteFile(checkpoint, old, 0600); err != nil {
		t.Fatal(err)
	}
	healthy := NewMemFPSet()
	healthy.Init(1, t.TempDir(), "Healthy")
	healthy.Put(2001)
	endpoint := &memory1CheckpointWriteEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(store)}
	manager := NewDistributedFPSetManager(endpoint, NewLocalFingerprintEndpoint(healthy))
	manager.fpSets[0].hostname = "buffer-owner"
	registration := manager.entry(0)
	if err := manager.Checkpoint("job"); err != nil {
		t.Fatal(err)
	}
	if fault != 0 {
		if _, ok := endpoint.failure.(*IOException); !ok || endpoint.failure.Error() != "Input/output error" || javaThrowableCause(endpoint.failure) != nil || endpoint.commits != 0 {
			t.Fatalf("buffered checkpoint failure/promotions = %v/%d", endpoint.failure, endpoint.commits)
		}
		current, err := os.ReadFile(checkpoint)
		if err != nil || !bytes.Equal(current, old) {
			t.Fatalf("failed checkpoint replaced the previous file: %x/%v", current, err)
		}
	} else if endpoint.failure != nil || endpoint.commits != 1 {
		t.Fatalf("normal checkpoint failed or was not promoted once: %v/%d", endpoint.failure, endpoint.commits)
	}
	healthyData, err := os.ReadFile(filepath.Join(healthy.metadir, "job.fp.chkpt"))
	if err != nil || len(healthyData) != 8 || binary.BigEndian.Uint64(healthyData) != 2001 {
		t.Fatalf("healthy checkpoint did not continue: %x/%v", healthyData, err)
	}
	const warning = "Error: Failed to checkpoint the fingerprint server at buffer-owner. This server might be down."
	wantWarnings := 0
	if fault != 0 {
		wantWarnings = 1
	}
	if strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning) != wantWarnings || manager.entry(0) != registration || manager.NumOfAliveServers() != 2 || store.Size() != 1025 {
		t.Fatal("buffered failure changed warning, registration or membership count")
	}
	for fp := 1; fp <= 1025; fp++ {
		if !store.Contains(uint64(fp)) {
			t.Fatalf("checkpoint lost fingerprint %d", fp)
		}
	}
	t.Log("BUFFERED_MEMORY_CHECKPOINT_VERIFIED")
}
