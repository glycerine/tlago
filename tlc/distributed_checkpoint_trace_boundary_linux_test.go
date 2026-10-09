package tlc

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// This native Linux check uses syscall injection, not a production lifecycle
// hook or a timing race. No enabled original test covers this interruption.
func TestDistributedCheckpointTraceBeforeInternInterruption(t *testing.T) {
	strace, err := exec.LookPath("strace")
	if err != nil {
		t.Skip("native syscall-boundary check requires strace")
	}
	for _, interrupted := range []bool{false, true} {
		t.Run(map[bool]string{false: "complete", true: "interrupted"}[interrupted], func(t *testing.T) {
			directory := t.TempDir()
			logfile := filepath.Join(t.TempDir(), "syscalls.log")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			args := []string{"-f", "-s", "4096", "-o", logfile, "-e", "trace=unlinkat,renameat,rmdir", "-e", "signal=none"}
			if interrupted {
				// Both checkpoints execute on one locked native thread. The
				// second queue/trace deletions precede the third unlinkat, vars.
				// Replacing that syscall prevents deletion even if signal delivery
				// follows syscall exit. SIGKILL then prevents further commit work.
				args = append(args, "-e", "inject=unlinkat:error=EIO:signal=SIGKILL:when=3")
			}
			args = append(args, os.Args[0], "-test.run=^TestDistributedCheckpointTraceBoundaryProcess$", "-test.v")
			command := exec.CommandContext(ctx, strace, args...)
			command.Env = append(os.Environ(), "TLAGO_TRACE_BOUNDARY_DIRECTORY="+directory)
			command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
			command.WaitDelay = time.Second
			t.Log("Running concrete checkpoint owners under native syscall tracing")
			output, runErr := command.CombinedOutput()
			log, err := os.ReadFile(logfile)
			if err != nil {
				t.Fatal(err)
			}
			if ctx.Err() != nil || (runErr != nil) != interrupted || !bytes.Contains(output, []byte("BASELINE_COMMITTED")) || bytes.Contains(output, []byte("SECOND_COMMITTED")) == interrupted {
				t.Fatalf("checkpoint process = %v/%v\n%s\n%s", runErr, ctx.Err(), output, log)
			}
			if interrupted {
				var deletion string
				status := command.ProcessState.Sys().(syscall.WaitStatus)
				if !(status.Signaled() && status.Signal() == syscall.SIGKILL || status.Exited() && status.ExitStatus() == 128+int(syscall.SIGKILL)) {
					t.Fatalf("checkpoint process did not terminate with SIGKILL: %v", command.ProcessState)
				}
				var deletes []string
				for _, line := range strings.Split(string(log), "\n") {
					if strings.Contains(line, "unlinkat(") {
						deletes = append(deletes, line)
						if strings.Contains(line, filepath.Join(directory, "vars.chkpt")) {
							deletion = line
						}
					}
				}
				// SIGKILL prevents syscall exit, so strace prints "= ?" rather
				// than its completed-fault INJECTED annotation. Require the exact
				// third entry on the same native thread, plus retained file bytes.
				if deletion == "" || len(deletes) != 3 || deletes[2] != deletion || !strings.HasSuffix(deletion, "= ?") {
					t.Fatalf("fault missed the intern deletion boundary\n%s", log)
				}
				for _, line := range deletes[:2] {
					if strings.Fields(line)[0] != strings.Fields(deletion)[0] || !strings.HasSuffix(line, "= 0") {
						t.Fatalf("fault did not follow the two completed deletes on one thread\n%s", log)
					}
				}
				if !strings.Contains(deletes[0], filepath.Join(directory, "queue.chkpt")) || !strings.Contains(deletes[1], filepath.Join(directory, "Spec.st.chkpt")) {
					t.Fatalf("the two earlier deletes did not belong to queue and trace\n%s", log)
				}
			}
			for _, name := range []string{"queue.chkpt", "Spec.st.chkpt", "vars.chkpt", "Spec.fp.chkpt"} {
				before, err := os.ReadFile(filepath.Join(directory, name+".before"))
				if err != nil {
					t.Fatal(err)
				}
				current, err := os.ReadFile(filepath.Join(directory, name))
				if err != nil {
					t.Fatal(err)
				}
				wantOld := interrupted && (name == "vars.chkpt" || name == "Spec.fp.chkpt")
				if bytes.Equal(before, current) != wantOld {
					t.Fatalf("%s does not retain its expected commit generation", name)
				}
				pending := strings.TrimSuffix(name, ".chkpt") + ".tmp"
				data, pendingErr := os.ReadFile(filepath.Join(directory, pending))
				if wantOld {
					if pendingErr != nil || bytes.Equal(before, data) {
						t.Fatalf("%s lost its unpromoted new generation: %v", pending, pendingErr)
					}
				} else if !os.IsNotExist(pendingErr) {
					t.Fatalf("%s was not promoted: %v", pending, pendingErr)
				}
			}
			oldVariables, oldCount, oldEmpty := stateVariables, UniqueStringVariableCount(), EmptyState
			stateVariables, EmptyState = nil, nil
			SetUniqueStringVariableCount(0)
			defer func() { stateVariables, EmptyState = oldVariables, oldEmpty; SetUniqueStringVariableCount(oldCount) }()
			queue := NewMemStateQueue(directory)
			if err := queue.Recover(); err != nil || queue.Size() != 2 {
				t.Fatalf("committed frontier recovery = %d/%v", queue.Size(), err)
			}
			trace := NewTLCTrace(directory, "Spec")
			defer trace.Close()
			if err := trace.Recover(); err != nil {
				t.Fatal(err)
			}
			metadata, err := os.ReadFile(trace.chkptName("chkpt"))
			if err != nil || len(metadata) != 16 || trace.lastPtr != int64(binary.BigEndian.Uint64(metadata[8:])) || trace.lastPtr <= 0 {
				t.Fatal("trace recovery did not use the newly committed pointer", err)
			}
			store := NewMemFPSet()
			store.Init(1, directory, "Spec")
			if err := store.RecoverFile("Spec"); err != nil {
				t.Fatal(err)
			}
			wantSize := uint64(2)
			if interrupted {
				wantSize = 1
			}
			if store.Size() != wantSize || !store.Contains(41) || store.Contains(43) == interrupted {
				t.Fatal("fingerprint recovery promoted an uncommitted generation")
			}
			intern := NewInternTable(16)
			if err := intern.Recover(directory); err != nil || intern.Find("before") == nil || (intern.Find("after") == nil) != interrupted {
				t.Fatal("intern recovery changed the retained generation", err)
			}
		})
	}
}

func TestDistributedCheckpointTraceBoundaryProcess(t *testing.T) {
	directory := os.Getenv("TLAGO_TRACE_BOUNDARY_DIRECTORY")
	if directory == "" {
		return
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	SetNumWorkers(0) // No workers are registered in this concrete checkpoint fixture.
	stateVariables, EmptyState = nil, nil
	SetUniqueStringVariableCount(0)
	queue := NewMemStateQueue(directory)
	trace := NewTLCTrace(directory, "Spec")
	defer trace.Close()
	store := NewMemFPSet()
	store.Init(1, directory, "Spec")
	server := NewTLCServer("Spec", "Spec", directory, NewNonDistributedFPSetManager(store, "local", trace), queue, trace)
	server.InternTable = NewInternTable(16)
	server.InternTable.Put("before")
	first := &TLCStateMut{level: 1}
	if err := trace.WriteInitState(first, 41); err != nil {
		t.Fatal(err)
	}
	queue.Enqueue(first)
	store.Put(41)
	if err := server.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"queue.chkpt", "Spec.st.chkpt", "vars.chkpt", "Spec.fp.chkpt"} {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name+".before"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("BASELINE_COMMITTED")
	second := &TLCStateMut{level: 2}
	if _, err := trace.WriteStateRecord(first, 43, second); err != nil {
		t.Fatal(err)
	}
	queue.Enqueue(second)
	store.Put(43)
	server.InternTable.Put("after")
	if err := server.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	t.Log("SECOND_COMMITTED")
}
