package tlc

import "testing"

// No original method directly tests missing disk-trace writer owners.
func TestDistributedTraceWritesRequireExistingOwner(t *testing.T) {
	for _, operation := range []string{"initial", "record", "successor"} {
		for _, scenario := range []string{"missing", "missing-prior-error", "closed", "healthy-prior-error"} {
			t.Run(operation+"/"+scenario, func(t *testing.T) {
				trace := NewTLCTrace(t.TempDir(), "Spec")
				defer trace.Close()
				if err := trace.raf.WriteFull(make([]byte, 9)); err != nil {
					t.Fatal(err)
				}
				owner := trace.raf
				if scenario == "missing" || scenario == "missing-prior-error" {
					if err := owner.Close(); err != nil {
						t.Fatal(err)
					}
					trace.raf, owner = nil, nil
				} else if scenario == "closed" {
					if err := trace.Close(); err != nil {
						t.Fatal(err)
					}
				}
				if scenario == "missing-prior-error" || scenario == "healthy-prior-error" {
					trace.traceErr = NewIOException("earlier creation failure")
				}
				trace.lastPtr, trace.level = 500, 3
				prior := &TLCStateMut{level: 2}
				action := &Action{Name: "Generated"}
				state := &TLCStateMut{UID: 314, WorkerID: 7, level: 5, pred: prior, action: action}
				parent := &TLCStateMut{UID: 37, WorkerID: 1, level: 7}
				err := invokeDistributedServerOperation(func() error {
					switch operation {
					case "initial":
						return trace.WriteInitState(state, 73)
					case "record":
						_, err := trace.WriteStateRecord(parent, 73, state)
						return err
					default:
						return trace.WriteNextState(parent, 73, state, nil)
					}
				})
				if trace.raf != owner {
					t.Fatal("write replaced its file owner")
				}
				if scenario == "healthy-prior-error" {
					if err != nil || trace.lastPtr != 9 || owner.curr != 21 || state.UID != 9 || len(trace.records) != 1 {
						t.Fatalf("existing owner write failed: %v", err)
					}
				} else {
					if scenario == "closed" {
						if !isJavaIOException(err) {
							t.Fatalf("closed owner failure: %T/%v", err, err)
						}
					} else if _, ok := err.(*NullPointerException); !ok {
						t.Fatalf("missing owner failure: %T/%v", err, err)
					}
					if trace.lastPtr != 500 || trace.level != 3 || len(trace.records) != 0 || state.UID != 314 || state.WorkerID != 7 || state.level != 5 || state.pred != prior || state.action != action {
						t.Fatal("owner failure published record or state metadata")
					}
				}
				if !trace.mu.TryLock() {
					t.Fatal("write retained trace lock")
				}
				trace.mu.Unlock()
			})
		}
	}
}

func TestDistributedTraceSuccessorRequiresPredecessorBeforeOwner(t *testing.T) {
	for _, scenario := range []string{"healthy", "missing", "closed"} {
		t.Run(scenario, func(t *testing.T) {
			trace := NewTLCTrace(t.TempDir(), "Spec")
			defer trace.Close()
			if scenario != "healthy" {
				if err := trace.Close(); err != nil {
					t.Fatal(err)
				}
				if scenario == "missing" {
					trace.raf = nil
				}
			}
			owner := trace.raf
			trace.lastPtr = 500
			state := &TLCStateMut{UID: 314, WorkerID: 7, level: 5}
			err := invokeDistributedServerOperation(func() error { return trace.WriteNextState(nil, 73, state, nil) })
			if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("missing predecessor failure: %T/%v", err, err)
			}
			if trace.raf != owner || trace.lastPtr != 500 || len(trace.records) != 0 || state.UID != 314 || state.WorkerID != 7 || state.level != 5 || owner != nil && owner.curr != 0 {
				t.Fatal("missing predecessor reached owner or publication")
			}
		})
	}
}

func TestDistributedTraceWritesRequireTrace(t *testing.T) {
	for _, operation := range []string{"initial", "record", "successor"} {
		t.Run(operation, func(t *testing.T) {
			var trace *TLCTrace
			err := invokeDistributedServerOperation(func() error {
				switch operation {
				case "initial":
					return trace.WriteInitState(nil, 73)
				case "record":
					_, err := trace.WriteStateRecord(nil, 73, nil)
					return err
				default:
					return trace.WriteNextState(nil, 73, nil, nil)
				}
			})
			if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("missing trace failure: %T/%v", err, err)
			}
		})
	}
}
