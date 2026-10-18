package tlc

import (
	"fmt"
	"testing"
)

func modelDataNameStates() []*TLCStateMut {
	name := &UniqueString{s: "name\x00世界", tok: 97, loc: 3, unregistered: true}
	other := &UniqueString{s: name.s, tok: name.tok, loc: name.loc, unregistered: name.unregistered}
	model := &ModelValue{Val: name, Data: name}
	container := &ModelValue{Val: name, Data: map[any]any{name: []any{name, (*UniqueString)(nil)}, other: other}}
	return []*TLCStateMut{{level: 1, values: []Value{model, container, &StringValue{Val: name}}}}
}

func checkModelDataNameStates(states []*TLCStateMut) error {
	model, container := states[0].values[0].(*ModelValue), states[0].values[1].(*ModelValue)
	name := model.Val
	if model.Data != name || container.Val != name || states[0].values[2].(*StringValue).Val != name || name.s != "name\x00世界" || name.tok != 97 || name.loc != 3 || !name.unregistered {
		return fmt.Errorf("attached name lost metadata or sharing with value fields")
	}
	data := container.Data.(map[any]any)
	array, ok := data[name].([]any)
	if !ok || len(data) != 2 || len(array) != 2 || array[0] != name {
		return fmt.Errorf("attached name lost map-key or array sharing")
	}
	if nilName, ok := array[1].(*UniqueString); !ok || nilName != nil {
		return fmt.Errorf("attached name lost typed nil")
	}
	for key, value := range data {
		if key != name {
			other := key.(*UniqueString)
			if other != value || *other != *name {
				return fmt.Errorf("equal distinct names were merged or lost metadata")
			}
		}
	}
	return nil
}

// UniqueString is transferable in the source. The Go attachment reuses its
// native graph entry; it must not flatten names into text or re-intern them.
func TestDistributedModelDataNames(t *testing.T) {
	states := modelDataNameStates()
	got := distributedPayloadRoundTrip(t, states)
	if err := checkModelDataNameStates(got); err != nil {
		t.Fatal(err)
	}
	got[0].values[0].(*ModelValue).Val.loc = 8
	if states[0].values[0].(*ModelValue).Val.loc != 3 {
		t.Fatal("receiver name metadata aliases sender")
	}
	nilData := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: (*UniqueString)(nil)}}}})[0].values[0].(*ModelValue).Data
	if name, ok := nilData.(*UniqueString); !ok || name != nil {
		t.Fatal("root attached name lost typed nil")
	}
	nilKeys := map[any]any{nil: "nil", (*UniqueString)(nil): "typed nil"}
	keys := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: nilKeys}}}})[0].values[0].(*ModelValue).Data.(map[any]any)
	if len(keys) != 2 || keys[nil] != "nil" || keys[(*UniqueString)(nil)] != "typed nil" {
		t.Fatal("nil and typed nil name keys were merged")
	}
}

func TestWorkerRPCModelDataNames(t *testing.T) {
	states := modelDataNameStates()
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if err := checkModelDataNameStates(received); err != nil {
			return nil, err
		}
		if received[0].UID == 1 {
			return nil, NewWorkerException("attached name failure", received[0], received[0], true)
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received)}, []*LongVec{NewLongVec()}, 1, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkModelDataNameStates([]*TLCStateMut{result.NextStates[0].At(0)}); err != nil {
		t.Fatal(err)
	}
	states[0].UID = 1
	result, err = client.GetNextStates(states)
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("attached name failure lost state identity: %v/%v", result, err)
	}
	if err := checkModelDataNameStates([]*TLCStateMut{failure.State1}); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedModelDataNamesInvalid(t *testing.T) {
	if _, err := DecodeDistributedStates(&DistributedStatePayload{Values: []DistributedValueNode{{Kind: "model", DataKind: "uniqueString"}}}); err != nil {
		t.Fatalf("valid nil attached name rejected: %v", err)
	}
	for _, id := range []int{-1, 1} {
		payload := &DistributedStatePayload{Values: []DistributedValueNode{{Kind: "model", DataKind: "uniqueString", DataName: id}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid attached name ID %d accepted", id)
		}
		payload.Values = nil
		payload.ObjectKeyMaps = [][]DistributedObjectKeyMapEntry{{{Key: DistributedObjectDataNode{Kind: "uniqueString", Name: id}}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid attached name key ID %d accepted", id)
		}
	}
}
