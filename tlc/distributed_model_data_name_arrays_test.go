package tlc

import (
	"fmt"
	"testing"
)

// Source ModelValue's attached data is not transient, and UniqueString is
// transferable. Reuse native name-array identities already used by records.
func modelDataNameArrayStates() []*TLCStateMut {
	name := &UniqueString{s: "name\x00世界", tok: 97, loc: 3, unregistered: true}
	names := []*UniqueString{name, {s: "other", tok: 98, loc: 4}, nil}
	values := []Value{NewIntValue(1), NewIntValue(2), NewIntValue(3)}
	return []*TLCStateMut{{level: 1, values: []Value{
		&ModelValue{Val: name, Data: names},
		NewRecordValue(names, values, false),
		&SetOfRcdsValue{Names: names, Values: values},
		&ModelValue{Data: map[string]any{"names": names, "nested": []any{names}}},
		&ModelValue{Data: append([]*UniqueString(nil), names...)},
		&ModelValue{Data: []*UniqueString(nil)},
		&ModelValue{Data: []*UniqueString{}},
	}}}
}

func checkModelDataNameArrays(states []*TLCStateMut) error {
	values := states[0].values
	model := values[0].(*ModelValue)
	names, ok := model.Data.([]*UniqueString)
	if !ok || len(names) != 3 || names[0] != model.Val || names[2] != nil || names[0].s != "name\x00世界" || names[0].tok != 97 || names[0].loc != 3 || !names[0].unregistered {
		return fmt.Errorf("attached names lost native type, identity, null entry or metadata")
	}
	record := values[1].(*RecordValue)
	set := values[2].(*SetOfRcdsValue)
	attached := values[3].(*ModelValue).Data.(map[string]any)
	for _, shared := range [][]*UniqueString{record.Names, set.Names, attached["names"].([]*UniqueString), attached["nested"].([]any)[0].([]*UniqueString)} {
		if len(shared) != len(names) || &shared[0] != &names[0] {
			return fmt.Errorf("attached name array lost sharing with record or nested attachment")
		}
	}
	separate := values[4].(*ModelValue).Data.([]*UniqueString)
	if &separate[0] == &names[0] || separate[0] != names[0] || separate[1] != names[1] || separate[2] != nil {
		return fmt.Errorf("separate equal array merged or lost shared name elements")
	}
	nilNames, ok := values[5].(*ModelValue).Data.([]*UniqueString)
	if !ok || nilNames != nil {
		return fmt.Errorf("nil attached name array lost native type")
	}
	empty, ok := values[6].(*ModelValue).Data.([]*UniqueString)
	if !ok || empty == nil || len(empty) != 0 {
		return fmt.Errorf("empty attached name array became null")
	}
	return nil
}

func checkModelDataNameArrayMutation(t *testing.T, received, sender []*TLCStateMut) {
	t.Helper()
	if err := checkModelDataNameArrays(received); err != nil {
		t.Fatal(err)
	}
	values := received[0].values
	names := values[0].(*ModelValue).Data.([]*UniqueString)
	names[0] = names[1]
	if values[1].(*RecordValue).Names[0] != names[1] || values[2].(*SetOfRcdsValue).Names[0] != names[1] || values[4].(*ModelValue).Data.([]*UniqueString)[0] == names[1] {
		t.Fatal("receiver name-array mutation lost sharing or changed the independent array")
	}
	if err := checkModelDataNameArrays(sender); err != nil {
		t.Fatalf("receiver mutation changed sender graph: %v", err)
	}
}

func TestDistributedModelDataNameArrays(t *testing.T) {
	states := modelDataNameArrayStates()
	checkModelDataNameArrayMutation(t, distributedPayloadRoundTrip(t, states), states)
}

func TestWorkerRPCModelDataNameArrays(t *testing.T) {
	states := modelDataNameArrayStates()
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if err := checkModelDataNameArrays(received); err != nil {
			return nil, err
		}
		if received[0].UID == 1 {
			return nil, NewWorkerException("attached name-array failure", received[0], received[0], true)
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received)}, []*LongVec{NewLongVec()}, 1, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	checkModelDataNameArrayMutation(t, []*TLCStateMut{result.NextStates[0].At(0)}, states)
	states[0].UID = 1
	result, err = client.GetNextStates(states)
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("attached array failure lost state identity: %v/%v", result, err)
	}
	checkModelDataNameArrayMutation(t, []*TLCStateMut{failure.State1}, states)
}

func TestDistributedModelDataNameArraysInvalid(t *testing.T) {
	for _, id := range []int{-1, 2} {
		payload := &DistributedStatePayload{NameArrays: [][]int{{0}}, Values: []DistributedValueNode{{Kind: "model", DataKind: "nameArray", DataArray: id}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid attached name-array reference %d accepted", id)
		}
		payload.Values = nil
		payload.ObjectArrays = [][]DistributedObjectDataNode{{{Kind: "nameArray", Array: id}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid nested name-array reference %d accepted", id)
		}
	}
	if _, err := DecodeDistributedStates(&DistributedStatePayload{NameArrays: [][]int{{1}}, Values: []DistributedValueNode{{Kind: "model", DataKind: "nameArray", DataArray: 1}}}); err == nil {
		t.Fatal("invalid name element reference accepted")
	}
}
