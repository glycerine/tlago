package tlc

import "testing"

// No original method directly covers metadata ownership or absent owners in
// distributed initial-state publication. Source writes only the UID to state.
func TestDistributedInitPublicationPreservesMetadataAndFailureOrder(t *testing.T) {
	oldTool := stateTool
	stateTool = nil
	t.Cleanup(func() { stateTool = oldTool })
	for _, kind := range []string{"healthy", "missing-trace", "missing-queue", "trace-io", "seen", "excluded"} {
		t.Run(kind, func(t *testing.T) {
			parent := &TLCStateMut{UID: 37, level: 4}
			action := &Action{}
			state := &TLCStateMut{UID: 53, WorkerID: TLCStateInitWorkerID, level: 5,
				pred: parent, action: action, values: []Value{NewIntValue(17)}}
			fp := state.FingerPrint()
			manager := NewDistributedFPSetManagerFromFPSet(NewMemFPSet())
			queue := NewMemStateQueue()
			trace := NewTLCTrace()
			server := &TLCServer{FPSetManager: manager, Trace: trace, StateQueue: queue}
			checks := 0
			app := &TLCApp{Invariants: []*Action{{}}, Tool: &Tool{
				IsInModelFunc: func(*Tool, *TLCStateMut) (bool, error) { return kind != "excluded", nil },
				IsValidStateFunc: func(_ *Tool, _ *Action, got *TLCStateMut) (bool, error) {
					checks++
					if got != state || state.WorkerID != TLCStateInitWorkerID || state.TracePredecessor() != parent || state.GetAction() != action || state.Level() != 5 {
						t.Fatal("initial property check received rewritten state metadata")
					}
					return true, nil
				},
			}}
			var expected error
			switch kind {
			case "missing-trace":
				server.Trace = nil
			case "missing-queue":
				server.StateQueue = nil
			case "trace-io":
				expected = NewIOException("initial trace failure")
				trace.traceErr = expected
			case "seen":
				manager.Put(fp)
				server.Trace, server.StateQueue = nil, nil
			case "excluded":
				server.Trace, server.StateQueue, server.FPSetManager = nil, nil, nil
			}
			functor := &distributedDoInitFunctor{server: server, app: app}
			got, err := functor.AddElement(state)
			if got != state || err != nil || state.WorkerID != TLCStateInitWorkerID || state.TracePredecessor() != parent || state.GetAction() != action || state.Level() != 5 {
				t.Fatalf("publication changed result or metadata: %v", err)
			}
			failed := kind == "missing-trace" || kind == "missing-queue" || kind == "trace-io"
			if failed {
				if expected != nil {
					if functor.err != expected {
						t.Fatalf("trace failure identity changed: %T/%v", functor.err, functor.err)
					}
				} else if _, ok := functor.err.(*NullPointerException); !ok {
					t.Fatalf("missing owner failure: %T/%v", functor.err, functor.err)
				}
				if !server.IsDone() || server.ErrState != state || !server.KeepCallStack || checks != 0 || queue.Size() != 0 || manager.Size() != 1 {
					t.Fatal("publication failure changed preceding insertion or subsequent checks")
				}
				later := &TLCStateMut{UID: 99, values: []Value{NewIntValue(18)}}
				if got, err := functor.AddElement(later); got != later || err != nil || later.UID != 99 || manager.Size() != 1 || server.ErrState != state {
					t.Fatal("recorded initialization failure did not suppress later publication")
				}
			} else if functor.err != nil || server.IsDone() || checks != map[string]int{"healthy": 1, "seen": 0, "excluded": 1}[kind] {
				t.Fatalf("publication property-check behavior changed: %v, checks %d", functor.err, checks)
			}
			written := kind == "healthy" || kind == "missing-queue"
			if written {
				records := trace.Records()
				if len(records) != 1 || records[0].PreviousUID != 1 || records[0].FP != fp || state.UID != 0 {
					t.Fatal("root trace write or preceding UID assignment changed")
				}
			} else if state.UID != 53 || len(trace.Records()) != 0 {
				t.Fatal("skipped/failed trace write changed UID or trace records")
			}
			if kind == "healthy" && (queue.Size() != 1 || queue.SDequeue() != state) {
				t.Fatal("initial publication did not queue original state")
			}
			if kind == "excluded" && (manager.Size() != 0 || queue.Size() != 0) {
				t.Fatal("excluded initial state touched publication owners")
			}
		})
	}
}
