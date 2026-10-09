package tlc

import (
	"fmt"
	"testing"
)

// ModelValue's attached data is retained by source network transfer. Native
// value-array data uses the same graph as state/tuple storage; no original
// ModelValueTest method directly exercises attached-data transport.
func modelDataArrayStates() []*TLCStateMut {
	first := &ModelValue{Val: &UniqueString{s: "first", tok: 1}, Index: 0}
	second := &ModelValue{Val: &UniqueString{s: "second", tok: 2}, Index: 1}
	array := []Value{first, NewIntValue(41), nil}
	first.Data, second.Data = array, array
	tuple := &TupleValue{Elems: array}
	return []*TLCStateMut{{level: 1, values: []Value{first, second, tuple}}, {level: 1, values: array}}
}

func checkModelDataArrayStates(states []*TLCStateMut) error {
	if len(states) != 2 || states[0] == nil || states[1] == nil || len(states[0].values) != 3 || len(states[1].values) != 3 {
		return fmt.Errorf("state graph shape changed")
	}
	first, ok := states[0].values[0].(*ModelValue)
	if !ok {
		return fmt.Errorf("first model value missing")
	}
	second, ok := states[0].values[1].(*ModelValue)
	if !ok {
		return fmt.Errorf("second model value missing")
	}
	tuple, ok := states[0].values[2].(*TupleValue)
	if !ok {
		return fmt.Errorf("tuple missing")
	}
	array, ok := first.Data.([]Value)
	if !ok || len(array) != 3 {
		return fmt.Errorf("attached value array missing")
	}
	other, ok := second.Data.([]Value)
	if !ok || len(other) != 3 || len(tuple.Elems) != 3 {
		return fmt.Errorf("shared value array missing")
	}
	if array[0] != first || array[2] != nil || &array[0] != &other[0] || &array[0] != &tuple.Elems[0] || &array[0] != &states[1].values[0] {
		return fmt.Errorf("attached array lost cycles, null elements or backing-array sharing")
	}
	return nil
}

func TestDistributedModelDataValueArrays(t *testing.T) {
	for _, array := range [][]Value{nil, {}, {NewIntValue(41), nil}} {
		model := &ModelValue{Data: array}
		got := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{model}}})[0].values[0].(*ModelValue)
		copied, ok := got.Data.([]Value)
		if !ok || (copied == nil) != (array == nil) || len(copied) != len(array) {
			t.Fatal("attached array lost its type, nil/empty distinction or length")
		}
	}
	states := modelDataArrayStates()
	got := distributedPayloadRoundTrip(t, states)
	if err := checkModelDataArrayStates(got); err != nil {
		t.Fatal(err)
	}
	array := got[0].values[0].(*ModelValue).Data.([]Value)
	array[1] = NewIntValue(97)
	if got[1].values[1].(*IntValue).Val != 97 || states[1].values[1].(*IntValue).Val != 41 {
		t.Fatal("decoded array did not share receiver storage or retained sender storage")
	}
}

func TestDistributedModelDataValueArraysRejectInvalidReferences(t *testing.T) {
	for _, id := range []int{-1, 1000000} {
		payload, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: []Value{NewIntValue(41)}}}}})
		if err != nil {
			t.Fatal(err)
		}
		payload.Values[0].DataArray = id
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("accepted invalid attached array reference %d", id)
		}
	}
}

func TestWorkerRPCModelDataValueArraySharing(t *testing.T) {
	states := modelDataArrayStates()
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if err := checkModelDataArrayStates(received); err != nil {
			return nil, err
		}
		if received[0] == states[0] || &received[1].values[0] == &states[1].values[0] {
			return nil, fmt.Errorf("request retained sender storage")
		}
		received[1].values[1] = NewIntValue(97)
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received)}, []*LongVec{NewLongVec()}, 1, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	got := []*TLCStateMut{result.NextStates[0].At(0), result.NextStates[0].At(1)}
	if err := checkModelDataArrayStates(got); err != nil {
		t.Fatal(err)
	}
	if got[1].values[1].(*IntValue).Val != 97 || states[1].values[1].(*IntValue).Val != 41 {
		t.Fatal("RPC request/reply lost attached array mutation or changed sender")
	}
}
