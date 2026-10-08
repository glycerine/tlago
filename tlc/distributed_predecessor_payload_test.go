package tlc

import "testing"

// TLCStateMutExt's predecessor is an ordinary state reference. No direct
// upstream test exercises its transfer; these checks cover the native graph.
func predecessorPayloadStates() []*TLCStateMut {
	shared := NewTupleValue([]Value{NewIntValue(7)})
	parent := &TLCStateMut{UID: 41, level: 2, values: []Value{shared}}
	first := &TLCStateMut{UID: 42, level: 3, values: []Value{shared}, pred: parent}
	second := &TLCStateMut{UID: 43, level: 3, values: []Value{shared}, pred: parent}
	cycle := &TLCStateMut{UID: 44, level: 10, values: []Value{shared}}
	cycle.pred = cycle
	return []*TLCStateMut{first, second, parent, first, cycle}
}

func requirePredecessorPayloadGraph(t *testing.T, original, copied []*TLCStateMut) {
	t.Helper()
	if len(copied) != 5 || copied[0] == original[0] || copied[0] != copied[3] || copied[0].Predecessor() != copied[2] || copied[1].Predecessor() != copied[2] || copied[2].TracePredecessor() != nil || copied[4].Predecessor() != copied[4] {
		t.Fatal("predecessor graph lost sharing, cycles, nulls or receiver ownership")
	}
	for i, state := range copied {
		if state.UID != original[i].UID || state.Level() != original[i].Level() || state.values[0] != copied[0].values[0] || state.values[0] == original[i].values[0] {
			t.Fatal("predecessor transfer changed stored metadata or shared value identity")
		}
	}
	copied[2].UID = 99
	if copied[0].Predecessor().UID != 99 || copied[1].Predecessor().UID != 99 || original[2].UID != 41 {
		t.Fatal("receiver predecessor mutation lost sharing or touched the sender")
	}
}

func TestDistributedPredecessorPayloadGraph(t *testing.T) {
	original := predecessorPayloadStates()
	requirePredecessorPayloadGraph(t, original, distributedPayloadRoundTrip(t, original))
	// The source parent need not be an explicit root of the invocation.
	root := original[0]
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{root})[0]
	if got.Predecessor() == nil || got.Predecessor() == original[2] || got.Predecessor().UID != 41 || got.Level() != 3 {
		t.Fatal("non-root predecessor was dropped or recomputed")
	}
	root.pred = (*TLCStateMut)(nil)
	if got := distributedPayloadRoundTrip(t, []*TLCStateMut{root})[0]; got.TracePredecessor() != nil {
		t.Fatal("typed null predecessor became a graph reference")
	}
}

func TestDistributedPredecessorPayloadValidation(t *testing.T) {
	for _, reference := range []int{-1, 2} {
		payload := &DistributedStatePayload{States: []DistributedStateNode{{Predecessor: reference}}, Roots: []int{1}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid predecessor reference %d accepted", reference)
		}
	}
	root := &TLCStateMut{pred: &TLCStateMut{action: &Action{}}}
	if _, err := EncodeDistributedStates([]*TLCStateMut{root}); err == nil {
		t.Fatal("unsupported ancestor evaluator metadata was silently dropped")
	}
}

func TestWorkerRPCPredecessorPayloadGraph(t *testing.T) {
	worker := &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if len(states) == 2 {
			return nil, NewWorkerException("predecessor failure", states[0], states[1], true)
		}
		if len(states) != 5 || states[0].Predecessor() != states[2] || states[1].Predecessor() != states[2] || states[4].Predecessor() != states[4] {
			t.Error("worker request lost predecessor graph")
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(states)}, []*LongVec{NewLongVec()}, 1, 5), nil
	}}
	_, client := startWorkerRPC(t, worker)
	original := predecessorPayloadStates()
	result, err := client.GetNextStates(original)
	if err != nil || result == nil || len(result.NextStates) != 1 {
		t.Fatalf("native predecessor graph result = %v/%v", result, err)
	}
	requirePredecessorPayloadGraph(t, original, result.NextStates[0].ToSlice())
	result, err = client.GetNextStates([]*TLCStateMut{original[0], original[4]})
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || !failure.KeepCallStack || failure.State1 == nil || failure.State2 == nil {
		t.Fatalf("native predecessor failure context = %v/%v", result, err)
	}
	if failure.State1 == original[0] || failure.State1.Predecessor() == nil || failure.State1.Predecessor().UID != 41 || failure.State2.Predecessor() != failure.State2 || failure.State1.values[0] != failure.State2.values[0] {
		t.Fatal("worker exception lost predecessor/value graph ownership or cycles")
	}
}
