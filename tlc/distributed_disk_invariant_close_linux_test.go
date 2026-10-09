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

// No original method directly covers scan-close failure. Fault the real native
// close of a separate scan file, without a production hook or timing race.
func TestDistributedDiskInvariantCloseFailure(t *testing.T) {
	strace, err := exec.LookPath("strace")
	if err != nil {
		t.Skip("native close-failure check requires strace")
	}
	for _, backend := range []string{"lsb", "msb"} {
		for _, phase := range []string{"valid", "invalid", "truncated"} {
			for _, fault := range []bool{false, true} {
				t.Run(backend+"/"+phase+"/fault="+strconv.FormatBool(fault), func(t *testing.T) {
					directory := t.TempDir()
					target := filepath.Join(directory, "scan.fp")
					var data [16]byte
					binary.BigEndian.PutUint64(data[:8], 41)
					binary.BigEndian.PutUint64(data[8:], 43)
					contents := data[:]
					if phase == "invalid" {
						binary.BigEndian.PutUint64(data[8:], 41)
					} else if phase == "truncated" {
						contents = contents[:9]
					}
					if err := os.WriteFile(target, contents, 0600); err != nil {
						t.Fatal(err)
					}
					logfile := filepath.Join(directory, "syscalls.log")
					args := []string{"-f", "-yy", "-s", "4096", "-o", logfile, "-P", target, "-e", "trace=close", "-e", "signal=none"}
					if fault {
						args = append(args, "-e", "inject=close:error=EIO:when=1")
					}
					args = append(args, os.Args[0], "-test.run=^TestDistributedDiskInvariantCloseProcess$", "-test.v")
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					command := exec.CommandContext(ctx, strace, args...)
					command.Env = append(os.Environ(), "TLAGO_INVARIANT_CLOSE_DIRECTORY="+directory, "TLAGO_INVARIANT_CLOSE_BACKEND="+backend, "TLAGO_INVARIANT_CLOSE_PHASE="+phase, "TLAGO_INVARIANT_CLOSE_FAULT="+strconv.FormatBool(fault))
					command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
					command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
					command.WaitDelay = time.Second
					t.Log("Checking real invariant scan-file close under native tracing")
					output, runErr := command.CombinedOutput()
					log, err := os.ReadFile(logfile)
					if err != nil || runErr != nil || ctx.Err() != nil || !bytes.Contains(output, []byte("CLOSE_OWNERSHIP_VERIFIED")) {
						t.Fatalf("close process = %v/%v/%v\n%s\n%s", runErr, ctx.Err(), err, output, log)
					}
					var closes []string
					for _, line := range strings.Split(string(log), "\n") {
						if strings.Contains(line, "close(") {
							closes = append(closes, line)
						}
					}
					if len(closes) != 1 || !strings.Contains(closes[0], "<"+target+">") || strings.Contains(closes[0], "(INJECTED)") != fault {
						t.Fatalf("fault missed the exact scan-file close\n%s", log)
					}
					if fault && !strings.Contains(closes[0], "= -1 EIO") || !fault && !strings.HasSuffix(closes[0], "= 0") {
						t.Fatalf("unexpected scan-file close result\n%s", log)
					}
				})
			}
		}
	}
}

func TestDistributedDiskInvariantCloseProcess(t *testing.T) {
	directory := os.Getenv("TLAGO_INVARIANT_CLOSE_DIRECTORY")
	if directory == "" {
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	t.Setenv(DiskFPSetLogLockCntProperty, "1")
	config := NewFPSetConfiguration()
	config.SetMemory(1 << 20)
	var set *DiskFPSet
	if os.Getenv("TLAGO_INVARIANT_CLOSE_BACKEND") == "lsb" {
		set = NewLSBDiskFPSet(config).DiskFPSet
	} else {
		set = NewMSBDiskFPSet(config).DiskFPSet
	}
	set.Init(1, directory, "Spec")
	// Constructor reader ownership remains separate from the scan file, whose
	// first close is the invariant's finally boundary. The parent wrote its bytes.
	set.fpFilename = filepath.Join(directory, "scan.fp")
	valid, err := invokeFingerprintEndpoint(func() (bool, error) { return NewLocalFingerprintEndpoint(set).CheckInvariant() })
	fault := os.Getenv("TLAGO_INVARIANT_CLOSE_FAULT") == "true"
	phase := os.Getenv("TLAGO_INVARIANT_CLOSE_PHASE")
	for i := 0; i < set.rwLock.Size(); i++ {
		lock := set.rwLock.GetAt(i)
		acquired := lock.TryLock()
		lock.Unlock() // Release either the probe or retained fixture ownership.
		if acquired == fault {
			t.Errorf("stripe %d retained = %v, want %v", i, !acquired, fault)
		}
	}
	set.Close()
	if fault {
		if valid || !isJavaIOException(err) || !strings.Contains(err.Error(), "Input/output error") {
			t.Fatalf("close failure did not override scan result: %v/%v", valid, err)
		}
	} else if phase == "truncated" {
		var eof *EOFException
		if valid || !errors.As(err, &eof) {
			t.Fatalf("truncated scan = %v/%v", valid, err)
		}
	} else if err != nil || valid != (phase == "valid") {
		t.Fatalf("scan result = %v/%v", valid, err)
	}
	t.Log("CLOSE_OWNERSHIP_VERIFIED")
}
