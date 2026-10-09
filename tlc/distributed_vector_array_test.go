package tlc

import (
	"fmt"
	"sync/atomic"
	"testing"
)

// The source result retains its supplied vector arrays. Native graph transfer
// must retain aliases to these arrays through attached model data as well.
func attachedVectorArrayState() *TLCStateMut {
	state := &TLCStateMut{level: 1}
	states := make([]*StateVec, 2, 9)
	states[0] = newDistributedStateVec(17).Add(state)
	fps := make([]*LongVec, 2, 11)
	fps[0] = NewLongVecFrom([]int64{41})
	state.values = []Value{&ModelValue{Data: []any{states, states, append([]*StateVec(nil), states...), fps, fps, append([]*LongVec(nil), fps...), ([]*StateVec)(nil), []*StateVec{}, ([]*LongVec)(nil), []*LongVec{}}}}
	return state
}

func checkAttachedVectorArrays(state *TLCStateMut) error {
	data := state.values[0].(*ModelValue).Data.([]any)
	states, repeated, other := data[0].([]*StateVec), data[1].([]*StateVec), data[2].([]*StateVec)
	fps, repeatedFPs, otherFPs := data[3].([]*LongVec), data[4].([]*LongVec), data[5].([]*LongVec)
	if len(states) != 2 || cap(states) != 2 || len(fps) != 2 || cap(fps) != 2 || states[1] != nil || fps[1] != nil {
		return fmt.Errorf("vector array length, capacity or null entries changed")
	}
	if &states[0] != &repeated[0] || &states[0] == &other[0] || &fps[0] != &repeatedFPs[0] || &fps[0] == &otherFPs[0] {
		return fmt.Errorf("shared and distinct vector array identities changed")
	}
	if states[0] != other[0] || states[0].At(0) != state || fps[0] != otherFPs[0] || fps[0].ElementAt(0) != 41 {
		return fmt.Errorf("vector array elements or recursive state identity changed")
	}
	if data[6].([]*StateVec) != nil || data[7].([]*StateVec) == nil || len(data[7].([]*StateVec)) != 0 || data[8].([]*LongVec) != nil || data[9].([]*LongVec) == nil || len(data[9].([]*LongVec)) != 0 {
		return fmt.Errorf("typed null and empty vector arrays changed")
	}
	return nil
}

func TestDistributedAttachedVectorArrays(t *testing.T) {
	sender := attachedVectorArrayState()
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{sender})[0]
	if err := checkAttachedVectorArrays(got); err != nil {
		t.Fatal(err)
	}
	data := got.values[0].(*ModelValue).Data.([]any)
	data[0].([]*StateVec)[1] = data[0].([]*StateVec)[0]
	data[3].([]*LongVec)[1] = data[3].([]*LongVec)[0]
	if data[1].([]*StateVec)[1] == nil || data[4].([]*LongVec)[1] == nil || data[2].([]*StateVec)[1] != nil || data[5].([]*LongVec)[1] != nil || sender.values[0].(*ModelValue).Data.([]any)[0].([]*StateVec)[1] != nil || sender.values[0].(*ModelValue).Data.([]any)[3].([]*LongVec)[1] != nil {
		t.Fatal("vector array mutation lost sharing or receiver isolation")
	}
}

func TestWorkerRPCAttachedVectorArrays(t *testing.T) {
	var fail atomic.Bool
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if err := checkAttachedVectorArrays(states[0]); err != nil {
			return nil, err
		}
		if fail.Load() {
			return nil, NewWorkerException("vector arrays", states[0], states[0], true)
		}
		data := states[0].values[0].(*ModelValue).Data.([]any)
		return NewNextStateResult(data[0].([]*StateVec), data[3].([]*LongVec), 1, 2), nil
	}})
	sender := attachedVectorArrayState()
	result, err := client.GetNextStates([]*TLCStateMut{sender})
	if err != nil {
		t.Fatal(err)
	}
	state := result.NextStates[0].At(0)
	if err := checkAttachedVectorArrays(state); err != nil {
		t.Fatal(err)
	}
	data := state.values[0].(*ModelValue).Data.([]any)
	if &result.NextStates[0] != &data[0].([]*StateVec)[0] || &result.NextFingerprints[0] != &data[3].([]*LongVec)[0] {
		t.Fatal("result root arrays lost sharing with attached data")
	}
	result.NextStates[1], result.NextFingerprints[1] = result.NextStates[0], result.NextFingerprints[0]
	if data[1].([]*StateVec)[1] == nil || data[4].([]*LongVec)[1] == nil || sender.values[0].(*ModelValue).Data.([]any)[0].([]*StateVec)[1] != nil || sender.values[0].(*ModelValue).Data.([]any)[3].([]*LongVec)[1] != nil {
		t.Fatal("root array mutation lost sharing or reached sender")
	}
	fail.Store(true)
	result, err = client.GetNextStates([]*TLCStateMut{sender})
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("vector array failure context changed: %v/%v", result, err)
	}
	if err := checkAttachedVectorArrays(failure.State1); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedAttachedVectorArrayInvalidReferences(t *testing.T) {
	for _, kind := range []string{"stateVectorArray", "longVectorArray"} {
		for _, id := range []int{-1, 2} {
			payload := &DistributedStatePayload{StateVectorArrays: [][]int{{0}}, LongVectorArrays: [][]int{{0}}, Values: []DistributedValueNode{{Kind: "model", DataKind: kind, DataArray: id}}}
			if _, err := DecodeDistributedStates(payload); err == nil {
				t.Fatalf("invalid %s attachment reference %d accepted", kind, id)
			}
			payload.Values = nil
			payload.ObjectArrays = [][]DistributedObjectDataNode{{{Kind: kind, Array: id}}}
			if _, err := DecodeDistributedStates(payload); err == nil {
				t.Fatalf("invalid nested %s reference %d accepted", kind, id)
			}
		}
	}
	for _, id := range []int{-1, 1} {
		if _, err := DecodeDistributedStates(&DistributedStatePayload{StateVectorArrays: [][]int{{id}}}); err == nil {
			t.Fatalf("invalid state-vector array element %d accepted", id)
		}
		if _, err := DecodeDistributedStates(&DistributedStatePayload{LongVectorArrays: [][]int{{id}}}); err == nil {
			t.Fatalf("invalid long-vector array element %d accepted", id)
		}
	}
}
