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

// No dedicated upstream test covers read I/O during memory FPSet recovery.
// Eager refill belongs to readLong, before that fingerprint can be inserted.
func TestDistributedBufferedMemoryRecoveryReadFailure(t *testing.T) {
	strace, err := exec.LookPath("strace")
	if err != nil {
		t.Skip("native buffered recovery failure check requires strace")
	}
	for _, backend := range []string{"mem", "mem2"} {
		for fault := 0; fault <= 3; fault++ {
			t.Run(backend+"/"+strconv.Itoa(fault), func(t *testing.T) {
				directory := t.TempDir()
				target := filepath.Join(directory, "job.fp.chkpt")
				if err := os.WriteFile(target, bufferedMemoryCheckpointBytes(), 0600); err != nil {
					t.Fatal(err)
				}
				logfile := filepath.Join(directory, "syscalls.log")
				args := []string{"-f", "-yy", "-s", "32", "-o", logfile, "-P", target, "-e", "trace=read,close", "-e", "signal=none"}
				if fault != 0 {
					args = append(args, "-e", "inject=read:error=EIO:when="+strconv.Itoa(fault))
				}
				args = append(args, os.Args[0], "-test.run=^TestDistributedBufferedMemoryRecoveryProcess$", "-test.v")
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				command := exec.CommandContext(ctx, strace, args...)
				command.Env = append(os.Environ(), "TLAGO_BUFFERED_RECOVERY_DIRECTORY="+directory, "TLAGO_BUFFERED_RECOVERY_BACKEND="+backend, "TLAGO_BUFFERED_RECOVERY_FAULT="+strconv.Itoa(fault))
				command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
				command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
				command.WaitDelay = time.Second
				t.Logf("Checking %s recovery with refill %d faulted (0 normal)", backend, fault)
				output, runErr := command.CombinedOutput()
				log, logErr := os.ReadFile(logfile)
				if runErr != nil || ctx.Err() != nil || logErr != nil || !bytes.Contains(output, []byte("BUFFERED_RECOVERY_VERIFIED")) {
					t.Fatalf("buffered recovery process = %v/%v/%v\n%s\n%s", runErr, ctx.Err(), logErr, output, log)
				}
				var reads, closes []string
				for _, line := range strings.Split(string(log), "\n") {
					if strings.Contains(line, "read(") {
						reads = append(reads, line)
					}
					if strings.Contains(line, "close(") {
						closes = append(closes, line)
					}
				}
				wantReads := fault
				if fault == 0 {
					wantReads = 3
				}
				if len(reads) != wantReads || len(closes) != 1 || !strings.Contains(closes[0], "<"+target+">") || !strings.HasSuffix(closes[0], "= 0") {
					t.Fatalf("refill count or file cleanup changed\n%s", log)
				}
				for i, line := range reads {
					faulted := fault != 0 && i == fault-1
					if !strings.Contains(line, "<"+target+">") || !strings.Contains(line, ", 8192)") || strings.Contains(line, "(INJECTED)") != faulted {
						t.Fatalf("recovery refill missed source buffer boundary\n%s", log)
					}
					if faulted && !strings.Contains(line, "= -1 EIO") || !faulted && !strings.HasSuffix(line, "= "+strconv.Itoa([]int{8192, 8, 0}[i])) {
						t.Fatalf("unexpected refill result\n%s", log)
					}
				}
			})
		}
	}
}

func TestDistributedBufferedMemoryRecoveryProcess(t *testing.T) {
	directory := os.Getenv("TLAGO_BUFFERED_RECOVERY_DIRECTORY")
	if directory == "" {
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fault, err := strconv.Atoi(os.Getenv("TLAGO_BUFFERED_RECOVERY_FAULT"))
	if err != nil {
		t.Fatal(err)
	}
	captureFailoverToolIO(t, ToolIOTool)
	var store FPSet
	switch os.Getenv("TLAGO_BUFFERED_RECOVERY_BACKEND") {
	case "mem":
		store = NewMemFPSet()
	case "mem2":
		store = NewMemFPSet2(NewFPSetConfiguration())
	default:
		t.Fatal("unknown buffered memory store")
	}
	store.Init(1, directory, "Spec")
	store.Put(2001)
	healthyDirectory := t.TempDir()
	var healthyBytes [8]byte
	binary.BigEndian.PutUint64(healthyBytes[:], 3001)
	if err := os.WriteFile(filepath.Join(healthyDirectory, "job.fp.chkpt"), healthyBytes[:], 0600); err != nil {
		t.Fatal(err)
	}
	healthy := NewMemFPSet()
	healthy.Init(1, healthyDirectory, "Healthy")
	endpoint := &memoryRecoveryCloseEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(store)}
	manager := NewDistributedFPSetManager(endpoint, NewLocalFingerprintEndpoint(healthy))
	manager.fpSets[0].hostname = "refill-owner"
	registration := manager.entry(0)
	if err := manager.Recover("job"); err != nil {
		t.Fatal(err)
	}
	limit := []int{1025, 0, 1023, 1024}[fault]
	if store.Size() != uint64(limit+1) || !store.Contains(2001) {
		t.Fatalf("refill failure retained %d fingerprints, want %d plus previous membership", store.Size(), limit)
	}
	for fp := 1; fp <= 1025; fp++ {
		if store.Contains(uint64(fp)) != (fp <= limit) {
			t.Fatalf("refill failure changed insertion of fingerprint %d", fp)
		}
	}
	if fault != 0 {
		if _, ok := endpoint.failure.(*IOException); !ok || endpoint.failure.Error() != "Input/output error" || javaThrowableCause(endpoint.failure) != nil {
			t.Fatalf("buffered refill I/O changed: %v", endpoint.failure)
		}
	} else if endpoint.failure != nil {
		t.Fatalf("normal recovery failed: %v", endpoint.failure)
	}
	const warning = "Error: Failed to checkpoint the fingerprint server at refill-owner. This server might be down."
	wantWarnings := 0
	if fault != 0 {
		wantWarnings = 1
	}
	if strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning) != wantWarnings || !healthy.Contains(3001) || manager.entry(0) != registration || manager.NumOfAliveServers() != 2 {
		t.Fatal("refill changed source warning, healthy continuation or registration")
	}
	t.Log("BUFFERED_RECOVERY_VERIFIED")
}
