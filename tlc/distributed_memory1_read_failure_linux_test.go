package tlc

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// FileUtil.newDFIS is unbuffered: a failed field read retains all earlier
// assignments in SetOfLong.recover. No dedicated upstream test covers read I/O.
func TestDistributedMemory1RecoveryReadFailure(t *testing.T) {
	strace, err := exec.LookPath("strace")
	if err != nil {
		t.Skip("native recovery read-failure check requires strace")
	}
	for read := 0; read <= 6; read++ {
		t.Run(strconv.Itoa(read), func(t *testing.T) {
			directory := t.TempDir()
			target := filepath.Join(directory, "job.fp.chkpt")
			if err := os.WriteFile(target, memory1RecoveryBytes(2, 9, 4, false, 41, 97), 0600); err != nil {
				t.Fatal(err)
			}
			logfile := filepath.Join(directory, "syscalls.log")
			args := []string{"-f", "-yy", "-s", "4096", "-o", logfile, "-P", target, "-e", "trace=read,close", "-e", "signal=none"}
			if read != 0 {
				args = append(args, "-e", "inject=read:error=EIO:when="+strconv.Itoa(read))
			}
			args = append(args, os.Args[0], "-test.run=^TestDistributedMemory1RecoveryReadProcess$", "-test.v")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, strace, args...)
			command.Env = append(os.Environ(), "TLAGO_MEMORY1_READ_DIRECTORY="+directory, "TLAGO_MEMORY1_READ_FAULT="+strconv.Itoa(read))
			command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
			command.WaitDelay = time.Second
			t.Logf("Checking recovery with file read %d faulted (0 means normal)", read)
			output, runErr := command.CombinedOutput()
			log, logErr := os.ReadFile(logfile)
			if runErr != nil || ctx.Err() != nil || logErr != nil || !bytes.Contains(output, []byte("MEMORY1_READ_VERIFIED")) {
				t.Fatalf("recovery read process = %v/%v/%v\n%s\n%s", runErr, ctx.Err(), logErr, output, log)
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
			wantReads := read
			if read == 0 {
				wantReads = 6
			}
			if len(reads) != wantReads || len(closes) != 1 || !strings.Contains(closes[0], "<"+target+">") || !strings.HasSuffix(closes[0], "= 0") {
				t.Fatalf("unexpected read/close ownership\n%s", log)
			}
			widths := []int{4, 4, 4, 1, 8, 8}
			for i, line := range reads {
				faulted := read != 0 && i == read-1
				if !strings.Contains(line, "<"+target+">") || !strings.Contains(line, ", "+strconv.Itoa(widths[i])+")") || strings.Contains(line, "(INJECTED)") != faulted {
					t.Fatalf("read did not stop at the expected field boundary\n%s", log)
				}
				if faulted && !strings.Contains(line, "= -1 EIO") || !faulted && !strings.HasSuffix(line, "= "+strconv.Itoa(widths[i])) {
					t.Fatalf("unexpected field read result\n%s", log)
				}
			}
		})
	}
}

func TestDistributedMemory1RecoveryReadProcess(t *testing.T) {
	directory := os.Getenv("TLAGO_MEMORY1_READ_DIRECTORY")
	if directory == "" {
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	read, err := strconv.Atoi(os.Getenv("TLAGO_MEMORY1_READ_FAULT"))
	if err != nil {
		t.Fatal(err)
	}
	captureFailoverToolIO(t, ToolIOTool)
	store := NewMemFPSet1(NewFPSetConfiguration())
	store.Init(1, directory, "Spec")
	old := []int64{17, 0, 0, 0, 0}
	store.set = &SetOfLong{count: 7, length: 5, thresh: 2, hasZero: true, table: old}
	healthy := NewMemFPSet()
	healthy.Init(1, t.TempDir(), "Healthy")
	healthy.Put(101)
	if err := healthy.BeginChkptFile("job"); err != nil {
		t.Fatal(err)
	}
	if err := healthy.CommitChkptFile("job"); err != nil {
		t.Fatal(err)
	}
	healthyDirectory := healthy.metadir
	healthy = NewMemFPSet()
	healthy.Init(1, healthyDirectory, "Healthy")
	endpoint := &memoryRecoveryCloseEndpoint{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(store)}
	manager := NewDistributedFPSetManager(endpoint, NewLocalFingerprintEndpoint(healthy))
	manager.fpSets[0].hostname = "read-owner"
	registration := manager.entry(0)
	if err := manager.Recover("job"); err != nil {
		t.Fatal(err)
	}
	if errors.Is(endpoint.failure, syscall.EIO) != (read != 0) || isJavaIOException(endpoint.failure) != (read != 0) {
		t.Fatalf("field read failure = %v", endpoint.failure)
	}
	count, length, threshold, zero := 7, 5, 2, true
	if read == 0 || read > 1 {
		count = 2
	}
	if read == 0 || read > 2 {
		length = 9
	}
	if read == 0 || read > 3 {
		threshold = 4
	}
	if read == 0 || read > 4 {
		zero = false
	}
	if read == 6 {
		count = 3
	} else if read == 0 {
		count = 4
	}
	set := store.set
	if set.count != count || set.length != length || set.thresh != threshold || set.hasZero != zero {
		t.Fatalf("failed read changed source partial fields: %#v", set)
	}
	if read >= 1 && read <= 4 {
		if &set.table[0] != &old[0] || !reflect.DeepEqual(set.table, old) {
			t.Fatal("failed header read replaced the old table")
		}
	} else if len(set.table) != 9 || &set.table[0] == &old[0] || set.Contains(41) != (read == 0 || read == 6) || set.Contains(97) != (read == 0) {
		t.Fatal("failed key read lost prior table reconstruction")
	}
	const warning = "Error: Failed to checkpoint the fingerprint server at read-owner. This server might be down."
	wantWarnings := 0
	if read != 0 {
		wantWarnings = 1
	}
	if strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning) != wantWarnings || !healthy.Contains(101) || manager.entry(0) != registration || manager.NumOfAliveServers() != 2 {
		t.Fatal("read failure changed warning, healthy continuation or registration")
	}
	t.Log("MEMORY1_READ_VERIFIED")
}
