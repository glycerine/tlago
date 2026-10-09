package tlc

import (
	"fmt"
	"testing"
)

// Source ModelValue.data is a non-transient Object. A map with string keys and
// Value entries transfers with its shared identity and cycles. No direct
// original ModelValueTest method tests attached-data networking.
func modelDataMapStates() []*TLCStateMut {
	first := &ModelValue{Val: &UniqueString{s: "first", tok: 1}}
	second := &ModelValue{Val: &UniqueString{s: "second", tok: 2}, Index: 1}
	data := map[string]Value{"self": first, "other": second, "number": NewIntValue(41), "": nil, "blue\x00雪": NewStringValue("blue")}
	first.Data, second.Data = data, data
	return []*TLCStateMut{{level: 1, values: []Value{first, second}}, {level: 2, values: []Value{first, second}}}
}

func checkModelDataMapStates(states []*TLCStateMut) error {
	if len(states) != 2 || len(states[0].values) != 2 || len(states[1].values) != 2 {
		return fmt.Errorf("map graph state shape changed")
	}
	first, firstOK := states[0].values[0].(*ModelValue)
	second, secondOK := states[0].values[1].(*ModelValue)
	if !firstOK || !secondOK || states[1].values[0] != first || states[1].values[1] != second {
		return fmt.Errorf("map graph model sharing changed")
	}
	data, firstOK := first.Data.(map[string]Value)
	other, secondOK := second.Data.(map[string]Value)
	if !firstOK || !secondOK || len(data) != 5 || len(other) != 5 || data["self"] != first || data["other"] != second || data["blue\x00雪"].(*StringValue).Val.String() != "blue" {
		return fmt.Errorf("map entries or cycles changed")
	}
	if value, present := data[""]; !present || value != nil {
		return fmt.Errorf("nil-valued empty key lost")
	}
	// Mutation verifies shared receiver storage rather than map equality alone.
	number := data["number"]
	data["number"] = NewIntValue(97)
	if other["number"] != data["number"] {
		return fmt.Errorf("attached maps lost shared identity")
	}
	data["number"] = number
	return nil
}

func TestDistributedModelDataValueMaps(t *testing.T) {
	for _, data := range []map[string]Value{nil, {}, {"nil": nil}} {
		got := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: data}}}})[0].values[0].(*ModelValue)
		copied, ok := got.Data.(map[string]Value)
		if !ok || (copied == nil) != (data == nil) || len(copied) != len(data) {
			t.Fatal("attached map lost type, nil/empty distinction or entries")
		}
	}
	states := modelDataMapStates()
	got := distributedPayloadRoundTrip(t, states)
	if err := checkModelDataMapStates(got); err != nil {
		t.Fatal(err)
	}
	got[0].values[0].(*ModelValue).Data.(map[string]Value)["number"] = NewIntValue(97)
	if states[0].values[0].(*ModelValue).Data.(map[string]Value)["number"].(*IntValue).Val != 41 {
		t.Fatal("receiver map mutation changed sender")
	}
	// Equal contents do not imply shared map identity.
	one := &ModelValue{Data: map[string]Value{"number": NewIntValue(41)}}
	two := &ModelValue{Data: map[string]Value{"number": NewIntValue(41)}}
	separate := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{one, two}}})[0].values
	separate[0].(*ModelValue).Data.(map[string]Value)["number"] = NewIntValue(97)
	if separate[1].(*ModelValue).Data.(map[string]Value)["number"].(*IntValue).Val != 41 {
		t.Fatal("distinct maps were merged")
	}
}

func TestDistributedModelDataValueMapsRejectInvalidReferences(t *testing.T) {
	for _, id := range []int{-1, 2} {
		payload, err := EncodeDistributedStates(modelDataMapStates())
		if err != nil {
			t.Fatal(err)
		}
		payload.Values[0].DataMap = id
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid map ID %d accepted", id)
		}
	}
	for _, entries := range [][]DistributedValueMapEntry{
		{{Key: "duplicate"}, {Key: "duplicate"}},
		{{Key: "negative", Value: -1}},
		{{Key: "outside", Value: 1000000}},
	} {
		payload, err := EncodeDistributedStates(modelDataMapStates())
		if err != nil {
			t.Fatal(err)
		}
		payload.ValueMaps[0] = entries
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid map entries accepted: %v", entries)
		}
	}
}

func TestWorkerRPCModelDataValueMaps(t *testing.T) {
	states := modelDataMapStates()
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if len(received) == 1 {
			return nil, NewWorkerException("map failure", received[0], received[0], true)
		}
		if err := checkModelDataMapStates(received); err != nil {
			return nil, err
		}
		received[0].values[0].(*ModelValue).Data.(map[string]Value)["number"] = NewIntValue(97)
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received)}, []*LongVec{NewLongVec()}, 1, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	got := []*TLCStateMut{result.NextStates[0].At(0), result.NextStates[0].At(1)}
	if err := checkModelDataMapStates(got); err != nil {
		t.Fatal(err)
	}
	if got[0].values[1].(*ModelValue).Data.(map[string]Value)["number"].(*IntValue).Val != 97 || states[0].values[0].(*ModelValue).Data.(map[string]Value)["number"].(*IntValue).Val != 41 {
		t.Fatal("RPC map mutation did not retain sharing or touched sender")
	}
	result, err = client.GetNextStates(states[:1])
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 == nil || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("map failure lost state identity: %v/%v", result, err)
	}
	if err := checkModelDataMapStates([]*TLCStateMut{failure.State1, failure.State2}); err != nil {
		t.Fatal(err)
	}
}
