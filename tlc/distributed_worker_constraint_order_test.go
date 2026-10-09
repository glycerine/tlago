package tlc

import (
	"reflect"
	"sync"
	"testing"
)

type workerConstraintLookupEndpoint struct {
	*LocalFingerprintEndpoint
	onLookup func()
}

func (e *workerConstraintLookupEndpoint) ContainsBlock(fps *LongVec) (*BitVector, error) {
	e.onLookup()
	return e.LocalFingerprintEndpoint.ContainsBlock(fps)
}

// Java has no direct TLCWorker.getNextStates test. Preserve lookup-before-check
// and invariant-before-constraint order, including when a constraint excludes
// the successor. Exercise the actual TLCApp delegates and native TCP boundary.
func TestDistributedWorkerConstraintCheckOrder(t *testing.T) {
	for _, remote := range []bool{false, true} {
		transport := "local"
		if remote {
			transport = "tcp"
		}
		for _, phase := range []string{"seen", "invariant-false", "invariant-error", "model-false", "model-error", "actions-false", "actions-error", "accepted"} {
			t.Run(transport+"/"+phase, func(t *testing.T) {
				var mu sync.Mutex
				var events []string
				appendEvent := func(event string) { mu.Lock(); events = append(events, event); mu.Unlock() }
				storage := NewMemFPSet()
				var endpoint DistributedFingerprintEndpoint = &workerConstraintLookupEndpoint{
					LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(storage), onLookup: func() { appendEvent("lookup") },
				}
				if remote {
					_, endpoint = startFingerprintRPC(t, endpoint)
				}
				worker, _ := managerOwnershipWorker(t, 1, NewDistributedFPSetManager(endpoint))
				worker.OverallStatesComputed.Store(7)
				predecessor := &TLCStateMut{UID: 31, level: 4}
				successor := &TLCStateMut{UID: 53, level: 5, values: []Value{NewIntValue(17)}}
				fp := successor.FingerPrint()
				if phase == "seen" {
					storage.Put(fp)
				}
				tool := worker.App.Tool
				tool.GetNextStatesFunc = func(*Tool, *Action, *TLCStateMut) (*StateVec, error) {
					return NewStateVecFrom([]*TLCStateMut{successor}), nil
				}
				worker.App.Invariants = []*Action{{Name: "Inv"}}
				tool.InvariantNames = []string{"Inv"}
				failure := NewRuntimeException("constraint evaluation failed")
				tool.IsValidStateFunc = func(*Tool, *Action, *TLCStateMut) (bool, error) {
					appendEvent("invariant")
					if phase == "invariant-error" {
						return false, failure
					}
					return phase != "invariant-false", nil
				}
				tool.IsInModelFunc = func(*Tool, *TLCStateMut) (bool, error) {
					appendEvent("model")
					if phase == "model-error" {
						return false, failure
					}
					// A false invariant must be reported even though this
					// constraint would exclude the offending successor.
					return phase != "model-false" && phase != "invariant-false", nil
				}
				tool.IsInActionsFunc = func(*Tool, *TLCStateMut, *TLCStateMut) (bool, error) {
					appendEvent("actions")
					if phase == "actions-error" {
						return false, failure
					}
					return phase != "actions-false", nil
				}
				invoke := worker.GetNextStates
				if remote {
					_, client := startWorkerRPC(t, NewLocalWorkerEndpoint(worker))
					invoke = client.GetNextStates
				}
				result, err := invoke([]*TLCStateMut{predecessor})
				want := []string{"lookup"}
				if phase != "seen" {
					want = append(want, "invariant")
					if phase != "invariant-false" && phase != "invariant-error" {
						want = append(want, "model")
						if phase != "model-false" && phase != "model-error" {
							want = append(want, "actions")
						}
					}
				}
				mu.Lock()
				got := append([]string(nil), events...)
				mu.Unlock()
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("check order = %q, want %q", got, want)
				}
				failed := phase == "invariant-false" || phase == "invariant-error" || phase == "model-error" || phase == "actions-error"
				if failed {
					wrapped, ok := err.(*WorkerException)
					if result != nil || !ok || wrapped.State1 == nil || wrapped.State1.UID != 31 || wrapped.State2 == nil || wrapped.State2.UID != 53 || wrapped.State2.FingerPrint() != fp {
						t.Fatalf("worker failure lost predecessor/successor context: %v/%v", result, err)
					}
					if phase == "invariant-false" {
						if wrapped.Cause != nil || wrapped.KeepCallStack || wrapped.Error() != "Error: Invariant Inv is violated." {
							t.Fatalf("invariant violation became a constraint exclusion: %+v", wrapped)
						}
					} else if wrapped.Cause == nil || wrapped.Cause.Error() != failure.Error() || !wrapped.KeepCallStack {
						t.Fatalf("evaluation failure changed cause/replay flag: %+v", wrapped)
					}
				} else {
					if err != nil || result == nil || result.StatesComputed != 1 || len(result.NextStates) != 1 || len(result.NextFingerprints) != 1 {
						t.Fatalf("successful result changed: %v/%v", result, err)
					}
					size := 0
					if phase == "accepted" {
						size = 1
						if result.NextStates[0].At(0).UID != 31 || result.NextFingerprints[0].ElementAt(0) != int64(fp) {
							t.Fatal("accepted successor lost predecessor UID or fingerprint alignment")
						}
					}
					if result.NextStates[0].Size() != size || result.NextFingerprints[0].Size() != size {
						t.Fatal("excluded or visited successor was returned as new")
					}
				}
				if phase != "accepted" && successor.UID != 53 {
					t.Fatal("worker rewrote UID before all checks/constraints accepted the state")
				}
				wantStored := uint64(0)
				if phase == "seen" {
					wantStored = 1
				}
				if storage.Size() != wantStored || worker.OverallStatesComputed.Load() != 8 || worker.Computing.Load() || worker.LastInvocation.Load() == 0 {
					t.Fatal("worker inserted fingerprints or changed generation/finally statistics")
				}
			})
		}
	}
}
