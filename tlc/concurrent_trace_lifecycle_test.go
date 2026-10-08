package tlc

import (
	"encoding/binary"
	"os"
	"testing"
)

// No original method directly covers a missing worker midway through these loops.
func TestConcurrentTraceLifecycleRequiresWorkersInOrder(t *testing.T) {
	for _, operation := range []string{"begin", "commit", "recover", "level"} {
		t.Run(operation, func(t *testing.T) {
			directory := t.TempDir()
			base := NewTLCTrace(directory, "Spec")
			defer base.Close()
			first, last := NewWorker(0), NewWorker(2)
			for _, worker := range []*Worker{first, last} {
				worker.SetTraceContext(directory, "Spec")
				defer worker.CloseTrace()
				if err := worker.ensureTraceRAF(); err != nil {
					t.Fatal(err)
				}
				if err := worker.WriteInitState(NewEmptyState(), 99); err != nil {
					t.Fatal(err)
				}
			}
			trace := &ConcurrentTLCTrace{TLCTrace: base, Workers: []*Worker{first, nil, last}}
			if operation == "commit" {
				for _, worker := range []*Worker{first, last} {
					if err := worker.BeginChkpt(); err != nil {
						t.Fatal(err)
					}
				}
			}
			if operation == "recover" {
				checkpoint := make([]byte, 16)
				binary.BigEndian.PutUint64(checkpoint, 13)
				binary.BigEndian.PutUint64(checkpoint[8:], 73)
				for _, worker := range []*Worker{first, last} {
					if err := os.WriteFile(worker.traceFileBase+".chkpt", checkpoint, 0600); err != nil {
						t.Fatal(err)
					}
				}
				first.lastPtr, last.lastPtr = 500, 600
			}
			first.SetLevel(3)
			last.SetLevel(10)
			err := invokeDistributedServerOperation(func() error {
				switch operation {
				case "begin":
					return trace.BeginChkpt()
				case "commit":
					return trace.CommitChkpt()
				case "recover":
					return trace.Recover()
				case "level":
					trace.GetLevel()
				}
				return nil
			})
			if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("required worker failure: %T/%v", err, err)
			}
			exists := func(path string) bool {
				t.Helper()
				_, err := os.Stat(path)
				if err == nil {
					return true
				}
				if os.IsNotExist(err) {
					return false
				}
				t.Fatal(err)
				return false
			}
			switch operation {
			case "begin":
				if !exists(first.traceFileBase+".tmp") || exists(last.traceFileBase+".tmp") {
					t.Fatal("begin did not preserve earlier write or reached later worker")
				}
			case "commit":
				if !exists(first.traceFileBase+".chkpt") || exists(first.traceFileBase+".tmp") || exists(last.traceFileBase+".chkpt") || !exists(last.traceFileBase+".tmp") || exists(base.chkptName("chkpt")) {
					t.Fatal("commit did not stop before later worker/marker publication")
				}
			case "recover":
				if first.lastPtr != 73 || last.lastPtr != 600 {
					t.Fatalf("recovery mutation order: %d/%d", first.lastPtr, last.lastPtr)
				}
			}
			if !trace.mu.TryLock() {
				t.Fatal("failed lifecycle operation retained monitor")
			}
			trace.mu.Unlock()
		})
	}
}

func TestConcurrentTraceLevelUsesWorkerMaximaOnly(t *testing.T) {
	base := NewTLCTrace()
	base.level = 99
	first, second := NewWorker(0), NewWorker(1)
	trace := &ConcurrentTLCTrace{TLCTrace: base, Workers: []*Worker{first, second}}
	for _, level := range []int{0, 1, 7} {
		second.SetLevel(level)
		want := level
		if want < 1 {
			want = 1
		}
		if got := trace.GetLevel(); got != want {
			t.Fatalf("worker level %d: got %d, want %d", level, got, want)
		}
		if got := trace.GetLevelForReporting(); got != want {
			t.Fatalf("reporting level %d, want %d", got, want)
		}
	}
	trace.Workers = nil
	if got := trace.GetLevel(); got != 1 {
		t.Fatalf("empty-worker level %d, want 1", got)
	}
}
