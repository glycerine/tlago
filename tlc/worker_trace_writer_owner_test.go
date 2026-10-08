package tlc

import "testing"

// No original method directly covers missing writer owners or stale open errors.
func TestWorkerTraceWritesRequireExistingOwner(t *testing.T) {
	for _, kind := range []string{"initial", "successor"} {
		for _, scenario := range []string{"missing", "memory-only", "closed", "healthy-prior-error"} {
			t.Run(kind+"/"+scenario, func(t *testing.T) {
				worker := NewWorker(2)
				if scenario != "memory-only" {
					worker.SetTraceContext(t.TempDir(), "Spec")
				}
				defer worker.CloseTrace()
				if scenario == "closed" || scenario == "healthy-prior-error" {
					if err := worker.ensureTraceRAF(); err != nil {
						t.Fatal(err)
					}
					if err := worker.traceRAF.WriteFull(make([]byte, 9)); err != nil {
						t.Fatal(err)
					}
				}
				owner := worker.traceRAF
				if scenario == "closed" {
					if err := worker.CloseTrace(); err != nil {
						t.Fatal(err)
					}
				}
				if scenario == "healthy-prior-error" {
					worker.traceErr = NewIOException("earlier creation failure")
				}
				worker.lastPtr, worker.UnseenSuccessorStates = 500, 17
				worker.SetLevel(3)
				state := &TLCStateMut{UID: 314, WorkerID: 7, level: 5}
				parent := &TLCStateMut{UID: 37, WorkerID: 1, level: 7}
				err := invokeDistributedServerOperation(func() error {
					if kind == "initial" {
						return worker.WriteInitState(state, 73)
					}
					return worker.WriteNextState(parent, 73, state, nil)
				})
				if worker.traceRAF != owner {
					t.Fatal("write opened or replaced its trace owner")
				}
				level := 3
				if kind == "successor" {
					level = 8
				}
				if worker.GetMaxLevel() != level {
					t.Fatal("owner access preceded depth update")
				}
				if scenario == "healthy-prior-error" {
					count := 17
					if kind == "successor" {
						count++
					}
					if err != nil || worker.lastPtr != 9 || state.UID != 9 || state.WorkerID != 2 || owner.curr != 22 || worker.UnseenSuccessorStates != count {
						t.Fatalf("existing owner blocked by stale error: %v", err)
					}
				} else {
					if scenario == "closed" {
						if !isJavaIOException(err) {
							t.Fatalf("closed owner error: %T/%v", err, err)
						}
					} else if _, ok := err.(*NullPointerException); !ok {
						t.Fatalf("missing owner was suppressed: %T/%v", err, err)
					}
					if worker.lastPtr != 500 || state.UID != 314 || state.WorkerID != 7 || state.Level() != 5 || worker.UnseenSuccessorStates != 17 {
						t.Fatal("owner failure published state metadata")
					}
				}
				if !worker.mu.TryLock() {
					t.Fatal("write retained worker lock")
				}
				worker.mu.Unlock()
			})
		}
	}
}
