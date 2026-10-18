package tlc

import (
	"fmt"
	"sync/atomic"
	"testing"
)

// Source TLCState and ModelValue attachments are transferable. The native
// representation must reuse state identities, including unrooted attachments.
func modelDataStateGraph() []*TLCStateMut {
	first := &TLCStateMut{WorkerID: 2, UID: 31, level: 3}
	second := &TLCStateMut{WorkerID: 4, UID: 41, level: 5, pred: first}
	attached := &TLCStateMut{WorkerID: 6, UID: 51, level: 7, pred: second}
	first.pred = attached
	model := &ModelValue{Data: first}
	first.values = []Value{model, &ModelValue{Data: attached}, &ModelValue{Data: map[any]any{first: attached, second: first, (*TLCStateMut)(nil): "nil state", nil: "nil"}}, &ModelValue{Data: []any{attached, first}}, &ModelValue{Data: (*TLCStateMut)(nil)}}
	second.values = []Value{model}
	attached.values = second.values
	attached.cached = map[int]Value{3: model}
	return []*TLCStateMut{first, second}
}

func checkModelDataStateGraph(states []*TLCStateMut) error {
	first, second := states[0], states[1]
	model := first.values[0].(*ModelValue)
	attached, ok := first.values[1].(*ModelValue).Data.(*TLCStateMut)
	if !ok || model.Data != first || attached == first || attached == second || first.pred != attached || second.pred != first || attached.pred != second {
		return fmt.Errorf("attached state identity or predecessor cycle lost")
	}
	if first.WorkerID != 2 || first.UID != 31 || first.level != 3 || second.WorkerID != 4 || second.UID != 41 || second.level != 5 || attached.WorkerID != 6 || attached.UID != 51 || attached.level != 7 {
		return fmt.Errorf("attached/root state metadata changed")
	}
	if second.values[0] != model || attached.values[0] != model || &second.values[0] != &attached.values[0] || attached.cached[3] != model {
		return fmt.Errorf("attached state lost shared values/cache back-reference")
	}
	keys := first.values[2].(*ModelValue).Data.(map[any]any)
	if len(keys) != 4 || keys[first] != attached || keys[second] != first || keys[(*TLCStateMut)(nil)] != "nil state" || keys[nil] != "nil" {
		return fmt.Errorf("attached state map keys lost identity or typed nil")
	}
	array := first.values[3].(*ModelValue).Data.([]any)
	if array[0] != attached || array[1] != first {
		return fmt.Errorf("nested attached states lost root identity")
	}
	nilState, ok := first.values[4].(*ModelValue).Data.(*TLCStateMut)
	if !ok || nilState != nil {
		return fmt.Errorf("null attached state lost native type")
	}
	return nil
}

func checkModelDataStateIsolation(t *testing.T, received, sender []*TLCStateMut) {
	t.Helper()
	if err := checkModelDataStateGraph(received); err != nil {
		t.Fatal(err)
	}
	attached := received[0].values[1].(*ModelValue).Data.(*TLCStateMut)
	attached.UID = 99
	attached.values[0] = NewIntValue(8)
	if received[0].pred.(*TLCStateMut).UID != 99 || received[1].values[0].(*IntValue).Val != 8 {
		t.Fatal("receiver state mutation lost graph sharing")
	}
	if err := checkModelDataStateGraph(sender); err != nil {
		t.Fatalf("receiver state mutation changed sender: %v", err)
	}
}

func TestDistributedModelDataStates(t *testing.T) {
	states := modelDataStateGraph()
	checkModelDataStateIsolation(t, distributedPayloadRoundTrip(t, states), states)
}

func TestWorkerRPCModelDataStates(t *testing.T) {
	var fail atomic.Bool
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if err := checkModelDataStateGraph(states); err != nil {
			return nil, err
		}
		if fail.Load() {
			return nil, NewWorkerException("attached state", states[0], states[1], true)
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(states)}, []*LongVec{NewLongVec()}, 1, 0), nil
	}})
	states := modelDataStateGraph()
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	checkModelDataStateIsolation(t, []*TLCStateMut{result.NextStates[0].At(0), result.NextStates[0].At(1)}, states)
	fail.Store(true)
	result, err = client.GetNextStates(states)
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || !failure.KeepCallStack {
		t.Fatalf("attached state failure lost context: %v/%v", result, err)
	}
	checkModelDataStateIsolation(t, []*TLCStateMut{failure.State1, failure.State2}, states)
}

func TestDistributedModelDataStatesInvalid(t *testing.T) {
	for _, id := range []int{-1, 2} {
		payload := &DistributedStatePayload{States: []DistributedStateNode{{}}, Values: []DistributedValueNode{{Kind: "model", DataKind: "state", DataState: id}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid attached state reference %d accepted", id)
		}
		payload.Values = nil
		payload.ObjectKeyMaps = [][]DistributedObjectKeyMapEntry{{{Key: DistributedObjectDataNode{Kind: "state", State: id}}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid attached state key %d accepted", id)
		}
	}
	payload := &DistributedStatePayload{States: []DistributedStateNode{{Level: -1}}, Values: []DistributedValueNode{{Kind: "model", DataKind: "state", DataState: 1}}}
	if _, err := DecodeDistributedStates(payload); err == nil {
		t.Fatal("invalid attached-only state metadata accepted")
	}
	for _, attached := range []*TLCStateMut{{level: -1}, {level: 1, functional: true}, {level: 1, action: &Action{}}} {
		if _, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: attached}}}}); err == nil {
			t.Fatal("attached state bypassed metadata validation")
		}
	}
}
