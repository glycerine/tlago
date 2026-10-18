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

// MemFPSet1 uses FileUtil.newDFOS, unlike the other buffered memory stores.
// No dedicated upstream test covers its checkpoint write/close I/O failures.
func TestDistributedMemory1CheckpointWriteFailure(t *testing.T) {
	strace, err := exec.LookPath("strace")
	if err != nil {
		t.Skip("native checkpoint write-failure check requires strace")
	}
	for fault := 0; fault <= 7; fault++ {
		t.Run(strconv.Itoa(fault), func(t *testing.T) {
			directory := t.TempDir()
			target := filepath.Join(directory, "job.fp.tmp")
			logfile := filepath.Join(directory, "syscalls.log")
			args := []string{"-f", "-yy", "-s", "4096", "-o", logfile, "-P", target, "-e", "trace=write,close", "-e", "signal=none"}
			if fault > 0 && fault <= 6 {
				args = append(args, "-e", "inject=write:error=EIO:when="+strconv.Itoa(fault))
			} else if fault == 7 {
				args = append(args, "-e", "inject=close:error=EIO:when=1")
			}
			args = append(args, os.Args[0], "-test.run=^TestDistributedMemory1CheckpointWriteProcess$", "-test.v")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, strace, args...)
			command.Env = append(os.Environ(), "TLAGO_MEMORY1_WRITE_DIRECTORY="+directory, "TLAGO_MEMORY1_WRITE_FAULT="+strconv.Itoa(fault))
			command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
			command.WaitDelay = time.Second
			t.Logf("Checking checkpoint fault %d (0 normal; 1–6 write; 7 close)", fault)
			output, runErr := command.CombinedOutput()
			log, logErr := os.ReadFile(logfile)
			if runErr != nil || ctx.Err() != nil || logErr != nil || !bytes.Contains(output, []byte("MEMORY1_WRITE_VERIFIED")) {
				t.Fatalf("checkpoint write process = %v/%v/%v\n%s\n%s", runErr, ctx.Err(), logErr, output, log)
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
			wantWrites := fault
			if fault == 0 || fault == 7 {
				wantWrites = 6
			}
			if len(writes) != wantWrites || len(closes) != 1 || !strings.Contains(closes[0], "<"+target+">") {
				t.Fatalf("checkpoint write replayed or failed to close once\n%s", log)
			}
			widths := []int{4, 4, 4, 1, 8, 8}
			for i, line := range writes {
				faulted := fault > 0 && fault <= 6 && i == fault-1
				if !strings.Contains(line, "<"+target+">") || !strings.Contains(line, ", "+strconv.Itoa(widths[i])+")") || strings.Contains(line, "(INJECTED)") != faulted {
					t.Fatalf("write missed the expected field boundary\n%s", log)
				}
				if faulted && !strings.Contains(line, "= -1 EIO") || !faulted && !strings.HasSuffix(line, "= "+strconv.Itoa(widths[i])) {
					t.Fatalf("unexpected checkpoint write result\n%s", log)
				}
			}
			if fault == 7 && (!strings.Contains(closes[0], "= -1 EIO") || !strings.Contains(closes[0], "(INJECTED)")) || fault != 7 && !strings.HasSuffix(closes[0], "= 0") {
				t.Fatalf("unexpected checkpoint close result\n%s", log)
			}
			if fault != 0 {
				prefix := []int{0, 0, 4, 8, 12, 13, 21, 29}[fault]
				data := memory1RecoveryBytes(2, 9, 4, false, 41, 97)
				temporary, err := os.ReadFile(target)
				if err != nil || !bytes.Equal(temporary, data[:prefix]) {
					t.Fatalf("failed checkpoint prefix = %x/%v, want %x", temporary, err, data[:prefix])
				}
			}
		})
	}
}

type memory1CheckpointWriteEndpoint struct {
	*LocalFingerprintEndpoint
	failure error
	commits int
}

func (e *memory1CheckpointWriteEndpoint) BeginChkptFile(name string) error {
	e.failure = e.LocalFingerprintEndpoint.BeginChkptFile(name)
	return e.failure
}

func (e *memory1CheckpointWriteEndpoint) CommitChkptFile(name string) error {
	e.commits++
	return e.LocalFingerprintEndpoint.CommitChkptFile(name)
}

func TestDistributedMemory1CheckpointWriteProcess(t *testing.T) {
	directory := os.Getenv("TLAGO_MEMORY1_WRITE_DIRECTORY")
	if directory == "" {
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fault, err := strconv.Atoi(os.Getenv("TLAGO_MEMORY1_WRITE_FAULT"))
	if err != nil {
		t.Fatal(err)
	}
	captureFailoverToolIO(t, ToolIOTool)
	store := NewMemFPSet1(NewFPSetConfiguration())
	store.Init(1, directory, "Spec")
	store.set = NewSetOfLong(9)
	store.Put(41)
	store.Put(97)
	set := store.set
	old := []byte("previous checkpoint")
	checkpoint := filepath.Join(directory, "job.fp.chkpt")
	if err := os.WriteFile(checkpoint, old, 0600); err != nil {
		t.Fatal(err)
	}
	healthy := NewMemFPSet()
	healthy.Init(1, t.TempDir(), "Healthy")
	healthy.Put(101)
	endpoint := &memory1CheckpointWriteEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(store)}
	manager := NewDistributedFPSetManager(endpoint, NewLocalFingerprintEndpoint(healthy))
	manager.fpSets[0].hostname = "write-owner"
	registration := manager.entry(0)
	if err := manager.Checkpoint("job"); err != nil {
		t.Fatal(err)
	}
	if errors.Is(endpoint.failure, syscall.EIO) != (fault != 0) || isJavaIOException(endpoint.failure) != (fault != 0) {
		t.Fatalf("checkpoint write/close failure = %v", endpoint.failure)
	}
	data := memory1RecoveryBytes(2, 9, 4, false, 41, 97)
	current, err := os.ReadFile(checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	if fault == 0 {
		if endpoint.commits != 1 || !bytes.Equal(current, data) {
			t.Fatal("successful checkpoint was not promoted exactly once")
		}
		if _, err := os.Stat(filepath.Join(directory, "job.fp.tmp")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("successful checkpoint retained temporary file: %v", err)
		}
	} else {
		if endpoint.commits != 0 || !bytes.Equal(current, old) {
			t.Fatal("failed checkpoint was promoted or changed the old checkpoint")
		}
	}
	healthyData, err := os.ReadFile(filepath.Join(healthy.metadir, "job.fp.chkpt"))
	if err != nil || len(healthyData) != 8 || binary.BigEndian.Uint64(healthyData) != 101 {
		t.Fatalf("healthy checkpoint did not continue: %x/%v", healthyData, err)
	}
	const warning = "Error: Failed to checkpoint the fingerprint server at write-owner. This server might be down."
	wantWarnings := 0
	if fault != 0 {
		wantWarnings = 1
	}
	if strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning) != wantWarnings || manager.entry(0) != registration || manager.NumOfAliveServers() != 2 {
		t.Fatal("write failure changed warning or registration")
	}
	if store.set != set || store.Size() != 2 || !store.Contains(41) || !store.Contains(97) || set.length != 9 || set.thresh != 4 || set.hasZero {
		t.Fatal("checkpoint changed memory membership or table metadata")
	}
	t.Log("MEMORY1_WRITE_VERIFIED")
}
