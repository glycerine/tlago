package tlc

import (
	"fmt"
	"sync/atomic"
	"testing"
)

// TLCStateVec's source contract transfers active entries and object references.
// No dedicated original method covers attached or recursive state vectors.
func attachedStateVectorState() *TLCStateMut {
	state := &TLCStateMut{WorkerID: 3, UID: 41, level: 7}
	vector := newDistributedStateVec(17)
	vector.Add(state).Add(nil)
	other := &StateVec{states: vector.states, distributed: true}
	data := []any{vector, vector, other, newDistributedStateVec(19), (*StateVec)(nil), map[any]any{vector: other}}
	state.values = []Value{&ModelValue{Data: data}}
	return state
}

func checkAttachedStateVectors(state *TLCStateMut) error {
	data := state.values[0].(*ModelValue).Data.([]any)
	vector, other := data[0].(*StateVec), data[2].(*StateVec)
	if vector != data[1] || vector == other || vector.Size() != 2 || other.Size() != 2 || cap(vector.states) != 2 || cap(other.states) != 2 {
		return fmt.Errorf("state vector identity, count or active capacity changed")
	}
	if vector.At(0) != state || other.At(0) != state || vector.At(1) != nil || other.At(1) != nil || state.WorkerID != 3 || state.UID != 41 || state.level != 7 {
		return fmt.Errorf("state vector back-reference, null entry or state metadata changed")
	}
	if &vector.states[0] == &other.states[0] || !vector.distributed || !other.distributed {
		return fmt.Errorf("state vector backing isolation or collection policy changed")
	}
	if empty := data[3].(*StateVec); empty == nil || empty.Size() != 0 || cap(empty.states) != 0 {
		return fmt.Errorf("empty state vector retained spare capacity or became null")
	}
	if data[4].(*StateVec) != nil || data[5].(map[any]any)[vector] != other {
		return fmt.Errorf("state vector typed null or map-key identity changed")
	}
	return nil
}

func TestDistributedAttachedStateVectors(t *testing.T) {
	sender := attachedStateVectorState()
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{sender})[0]
	if err := checkAttachedStateVectors(got); err != nil {
		t.Fatal(err)
	}
	data := got.values[0].(*ModelValue).Data.([]any)
	data[0].(*StateVec).Add(got)
	if data[1].(*StateVec).Size() != 3 || data[2].(*StateVec).Size() != 2 || sender.values[0].(*ModelValue).Data.([]any)[0].(*StateVec).Size() != 2 {
		t.Fatal("receiver state vector mutation lost sharing or reached sender")
	}
}

func TestWorkerRPCAttachedStateVectors(t *testing.T) {
	var fail atomic.Bool
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if err := checkAttachedStateVectors(states[0]); err != nil {
			return nil, err
		}
		if fail.Load() {
			return nil, NewWorkerException("attached state vector", states[0], states[0], true)
		}
		vector := states[0].values[0].(*ModelValue).Data.([]any)[0].(*StateVec)
		return NewNextStateResult([]*StateVec{vector, vector}, []*LongVec{NewLongVec()}, 1, 2), nil
	}})
	sender := attachedStateVectorState()
	result, err := client.GetNextStates([]*TLCStateMut{sender})
	if err != nil {
		t.Fatal(err)
	}
	state := result.NextStates[0].At(0)
	if err := checkAttachedStateVectors(state); err != nil {
		t.Fatal(err)
	}
	vector := state.values[0].(*ModelValue).Data.([]any)[0].(*StateVec)
	if result.NextStates[0] != vector || result.NextStates[1] != vector {
		t.Fatal("result state vectors lost sharing with attached data")
	}
	vector.Add(state)
	if result.NextStates[1].Size() != 3 || sender.values[0].(*ModelValue).Data.([]any)[0].(*StateVec).Size() != 2 {
		t.Fatal("result state vector mutation lost sharing or reached sender")
	}
	fail.Store(true)
	result, err = client.GetNextStates([]*TLCStateMut{sender})
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("state vector error context changed: %v/%v", result, err)
	}
	if err := checkAttachedStateVectors(failure.State1); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedAttachedStateVectorInvalidReferences(t *testing.T) {
	for _, id := range []int{-1, 2} {
		payload := &DistributedStatePayload{StateVectors: [][]int{{0}}, Values: []DistributedValueNode{{Kind: "model", DataKind: "stateVector", DataArray: id}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid attached state vector reference %d accepted", id)
		}
		payload.Values = nil
		payload.ObjectKeyMaps = [][]DistributedObjectKeyMapEntry{{{Key: DistributedObjectDataNode{Kind: "stateVector", Array: id}}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid state vector map key %d accepted", id)
		}
		if _, err := DecodeDistributedResult(&DistributedResultPayload{States: &DistributedStatePayload{StateVectors: [][]int{{0}}}, StateVectors: []int{id}}); err == nil {
			t.Fatalf("invalid result state vector reference %d accepted", id)
		}
	}
	for _, id := range []int{-1, 1} {
		if _, err := DecodeDistributedStates(&DistributedStatePayload{StateVectors: [][]int{{id}}}); err == nil {
			t.Fatalf("invalid vector state reference %d accepted", id)
		}
	}
	for _, state := range []*TLCStateMut{{level: -1}, {level: 1, functional: true}, {level: 1, action: &Action{}}} {
		vector := newDistributedStateVec(1).Add(state)
		if _, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: vector}}}}); err == nil {
			t.Fatal("attached vector bypassed state metadata validation")
		}
		if _, err := EncodeDistributedResult(NewNextStateResult([]*StateVec{vector}, nil, 1, 1)); err == nil {
			t.Fatal("result vector bypassed state metadata validation")
		}
	}
}
