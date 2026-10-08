package tlc

import "testing"

// No original method directly passes a missing initial state to this callback.
// Source constraints precede fingerprinting; excluded states still reach checks.
func TestDistributedInitMissingStatePreservesConstraintBoundary(t *testing.T) {
	oldTool := stateTool
	stateTool = nil
	t.Cleanup(func() { stateTool = oldTool })
	for _, kind := range []string{"in-model", "excluded", "constraint-error"} {
		t.Run(kind, func(t *testing.T) {
			storage := NewMemFPSet()
			manager := NewDistributedFPSetManagerFromFPSet(storage)
			queue := NewMemStateQueue()
			trace := NewTLCTrace(t.TempDir(), "Spec")
			t.Cleanup(func() { _ = trace.Close() })
			server := &TLCServer{FPSetManager: manager, StateQueue: queue, Trace: trace}
			constraints, checks := 0, 0
			constraintFailure := NewIOException("initial constraint failure")
			app := &TLCApp{Invariants: []*Action{{}}, Tool: &Tool{
				IsInModelFunc: func(_ *Tool, state *TLCStateMut) (bool, error) {
					constraints++
					if state != nil {
						t.Fatal("later state reached constraints after recorded failure")
					}
					if kind == "constraint-error" {
						return false, constraintFailure
					}
					return kind != "excluded", nil
				},
				IsValidStateFunc: func(_ *Tool, _ *Action, state *TLCStateMut) (bool, error) {
					checks++
					if state != nil {
						t.Fatal("property check replaced missing state")
					}
					return true, nil
				},
			}}
			functor := &distributedDoInitFunctor{server: server, app: app}
			var missing *TLCStateMut
			got, err := functor.AddElement(missing)
			if got != missing || err != nil || constraints != 1 || storage.Size() != 0 || queue.Size() != 0 || len(trace.Records()) != 0 {
				t.Fatalf("missing initial state changed callback/publication ordering: %v/%v", got, err)
			}
			if kind == "excluded" {
				if functor.err != nil || server.IsDone() || checks != 1 {
					t.Fatalf("excluded state was fingerprinted or skipped checks: %v", functor.err)
				}
				return
			}
			if kind == "constraint-error" {
				if functor.err != constraintFailure {
					t.Fatalf("constraint failure replaced: %v", functor.err)
				}
			} else if !isDistributedNullFailure(functor.err) {
				t.Fatalf("missing state was fingerprinted as an empty state: %v", functor.err)
			}
			if !server.IsDone() || !server.KeepCallStack || server.ErrState != nil || server.PredErrState != nil || checks != 0 {
				t.Fatal("initial failure changed source error state or property ordering")
			}
			later := &TLCStateMut{UID: 99, values: []Value{NewIntValue(1)}}
			if got, err := functor.AddElement(later); got != later || err != nil || constraints != 1 || checks != 0 || later.UID != 99 || storage.Size() != 0 || queue.Size() != 0 {
				t.Fatal("recorded initial failure did not suppress later publication")
			}
		})
	}
}

func TestMissingStateFingerprintDoesNotBecomeEmptyState(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		name := "global-tool"
		if explicit {
			name = "explicit-tool"
		}
		t.Run(name, func(t *testing.T) {
			var state *TLCStateMut
			err := invokeDistributedServerOperation(func() error {
				if explicit {
					state.FingerPrintWithTool(nil)
				} else {
					state.FingerPrint()
				}
				return nil
			})
			if !isDistributedNullFailure(err) {
				t.Fatalf("missing fingerprint failure: %T/%v", err, err)
			}
		})
	}
}
