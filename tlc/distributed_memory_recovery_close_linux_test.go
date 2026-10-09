package tlc

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
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

// No original memory FPSet method tests recovery's final close. Inject at the
// real file syscall, preserving ordinary constructors, readers and manager work.
func TestDistributedMemoryRecoveryCloseFailure(t *testing.T) {
	strace, err := exec.LookPath("strace")
	if err != nil {
		t.Skip("native recovery close-failure check requires strace")
	}
	for _, backend := range []string{"mem", "mem1", "mem2"} {
		for _, phase := range []string{"valid", "truncated", "duplicate"} {
			if backend == "mem1" && phase == "duplicate" {
				continue // SetOfLong recovery accepts duplicates, unlike the other stores.
			}
			for _, fault := range []bool{false, true} {
				t.Run(backend+"/"+phase+"/fault="+strconv.FormatBool(fault), func(t *testing.T) {
					directory := t.TempDir()
					target := filepath.Join(directory, "job.fp.chkpt")
					var data [24]byte
					for i, fp := range []uint64{41, 41, 43} {
						binary.BigEndian.PutUint64(data[8*i:], fp)
					}
					contents := data[:8]
					if phase == "duplicate" {
						contents = data[:]
					} else if phase == "truncated" {
						contents = data[:9]
					}
					if backend == "mem1" {
						contents = memory1RecoveryBytes(1, 9, 4, false, 41)
						if phase == "truncated" {
							contents = contents[:len(contents)-1]
						}
					}
					if err := os.WriteFile(target, contents, 0600); err != nil {
						t.Fatal(err)
					}
					logfile := filepath.Join(directory, "syscalls.log")
					args := []string{"-f", "-yy", "-s", "4096", "-o", logfile, "-P", target, "-e", "trace=close", "-e", "signal=none"}
					if fault {
						args = append(args, "-e", "inject=close:error=EIO:when=1")
					}
					args = append(args, os.Args[0], "-test.run=^TestDistributedMemoryRecoveryCloseProcess$", "-test.v")
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					command := exec.CommandContext(ctx, strace, args...)
					command.Env = append(os.Environ(), "TLAGO_MEMORY_RECOVERY_CLOSE_DIRECTORY="+directory, "TLAGO_MEMORY_RECOVERY_CLOSE_BACKEND="+backend, "TLAGO_MEMORY_RECOVERY_CLOSE_PHASE="+phase, "TLAGO_MEMORY_RECOVERY_CLOSE_FAULT="+strconv.FormatBool(fault))
					command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
					command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
					command.WaitDelay = time.Second
					t.Log("Checking real memory fingerprint recovery-file close")
					output, runErr := command.CombinedOutput()
					log, logErr := os.ReadFile(logfile)
					if logErr != nil || runErr != nil || ctx.Err() != nil || !bytes.Contains(output, []byte("MEMORY_RECOVERY_CLOSE_VERIFIED")) {
						t.Fatalf("recovery close process = %v/%v/%v\n%s\n%s", runErr, ctx.Err(), logErr, output, log)
					}
					var closes []string
					for _, line := range strings.Split(string(log), "\n") {
						if strings.Contains(line, "close(") {
							closes = append(closes, line)
						}
					}
					if len(closes) != 1 || !strings.Contains(closes[0], "<"+target+">") || strings.Contains(closes[0], "(INJECTED)") != fault {
						t.Fatalf("fault missed exact recovery-file close\n%s", log)
					}
					if fault && !strings.Contains(closes[0], "= -1 EIO") || !fault && !strings.HasSuffix(closes[0], "= 0") {
						t.Fatalf("unexpected recovery-file close result\n%s", log)
					}
				})
			}
		}
	}
}

type memoryRecoveryCloseEndpoint struct {
	*LocalFingerprintEndpoint
	failure error
}

func (e *memoryRecoveryCloseEndpoint) RecoverFile(name string) error {
	e.failure = e.LocalFingerprintEndpoint.RecoverFile(name)
	return e.failure
}

func TestDistributedMemoryRecoveryCloseProcess(t *testing.T) {
	directory := os.Getenv("TLAGO_MEMORY_RECOVERY_CLOSE_DIRECTORY")
	if directory == "" {
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	captureFailoverToolIO(t, ToolIOTool)
	backend := os.Getenv("TLAGO_MEMORY_RECOVERY_CLOSE_BACKEND")
	phase := os.Getenv("TLAGO_MEMORY_RECOVERY_CLOSE_PHASE")
	fault := os.Getenv("TLAGO_MEMORY_RECOVERY_CLOSE_FAULT") == "true"
	var store FPSet
	switch backend {
	case "mem":
		store = NewMemFPSet()
	case "mem1":
		store = NewMemFPSet1(NewFPSetConfiguration())
	case "mem2":
		store = NewMemFPSet2(NewFPSetConfiguration())
	default:
		t.Fatal("unknown memory backend")
	}
	store.Init(1, directory, "Spec")
	store.Put(17)
	healthyDirectory := t.TempDir()
	var healthyData [8]byte
	binary.BigEndian.PutUint64(healthyData[:], 97)
	if err := os.WriteFile(filepath.Join(healthyDirectory, "job.fp.chkpt"), healthyData[:], 0600); err != nil {
		t.Fatal(err)
	}
	healthy := NewMemFPSet()
	healthy.Init(1, healthyDirectory, "Healthy")
	endpoint := &memoryRecoveryCloseEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(store)}
	manager := NewDistributedFPSetManager(endpoint, NewLocalFingerprintEndpoint(healthy))
	manager.fpSets[0].hostname = "close-owner"
	registration := manager.entry(0)
	err := manager.Recover("job")
	readRuntime := phase == "duplicate" || phase == "truncated" && backend == "mem"
	wantIO := phase == "valid" && fault || phase == "truncated" && backend != "mem"
	if readRuntime {
		want := NewTLCRuntimeException(ECTLCFPNotInSet)
		if phase == "truncated" {
			want = NewTLCRuntimeException(ECSystemDiskIOErrorForFile, "checkpoints")
		}
		if err == nil || endpoint.failure != err || err.Error() != want.Error() || isJavaIOException(err) || errors.Is(err, syscall.EIO) {
			t.Fatalf("close replaced original runtime recovery failure: %v/%v", err, endpoint.failure)
		}
	} else if err != nil || isJavaIOException(endpoint.failure) != wantIO {
		t.Fatalf("manager recovery/close classification = %v/%v, want I/O %v", err, endpoint.failure, wantIO)
	}
	if phase == "valid" {
		if fault && backend == "mem1" {
			// The existing buffered stream translates native file errors to
			// IOException text, matching its source stream boundary.
			if _, ok := endpoint.failure.(*IOException); !ok || endpoint.failure.Error() != "Input/output error" || javaThrowableCause(endpoint.failure) != nil {
				t.Fatalf("buffered recovery close failure changed: %v", endpoint.failure)
			}
		} else if errors.Is(endpoint.failure, syscall.EIO) != fault {
			t.Fatalf("successful scan discarded or fabricated final close failure: %v", endpoint.failure)
		}
	}
	if phase == "truncated" && backend == "mem1" {
		if _, ok := endpoint.failure.(*EOFException); !ok {
			t.Fatalf("close replaced original table recovery EOF: %v", endpoint.failure)
		}
	}
	if phase == "truncated" && backend == "mem2" && endpoint.failure.Error() != "MemFPSet2.recover: failed." {
		t.Fatalf("close replaced original packed recovery I/O: %v", endpoint.failure)
	}
	const warning = "Error: Failed to checkpoint the fingerprint server at close-owner. This server might be down."
	warnings := strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning)
	wantWarnings := 0
	if wantIO {
		wantWarnings = 1
	}
	if warnings != wantWarnings || healthy.Contains(97) != !readRuntime || manager.entry(0) != registration || manager.NumOfAliveServers() != 2 {
		t.Fatal("recovery close changed warning, healthy continuation or endpoint availability")
	}
	wantCount := uint64(2)
	if backend == "mem1" && phase == "truncated" {
		wantCount = 1 // Serialized count assigned; incomplete key has not been inserted.
	}
	wantRecoveredKey := backend != "mem1" || phase != "truncated"
	if store.Size() != wantCount || store.Contains(41) != wantRecoveredKey || store.Contains(43) || store.Contains(17) != (backend != "mem1") {
		t.Fatal("recovery close changed source partial membership or counts")
	}
	t.Log("MEMORY_RECOVERY_CLOSE_VERIFIED")
}
