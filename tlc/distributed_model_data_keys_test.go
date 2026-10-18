package tlc

import (
	"fmt"
	"reflect"
	"testing"
)

func modelDataKeyStates() []*TLCStateMut {
	first, second := &ModelValue{Val: UniqueStringOf("keyFirst")}, &ModelValue{Val: UniqueStringOf("keySecond")}
	data := map[any]any{nil: "nil", int(1): "int", int32(1): "int32", true: []byte{17}, float64(1.25): "float64", float32(1.25): "float32"}
	data[uint16(65)], data[int16(65)], data[uint16(0xd800)] = "character", "short", "surrogate"
	data[first], data[second], data["self"] = first, data, data
	data["array"] = []any{data, first}
	first.Data, second.Data = data, data
	return []*TLCStateMut{{level: 1, values: []Value{first, second}}}
}

func checkModelDataKeyStates(states []*TLCStateMut) error {
	first, second := states[0].values[0].(*ModelValue), states[0].values[1].(*ModelValue)
	data := first.Data.(map[any]any)
	if len(data) != 13 || data[nil] != "nil" || data[int(1)] != "int" || data[int32(1)] != "int32" || data[true].([]byte)[0] != 17 || data[float64(1.25)] != "float64" || data[float32(1.25)] != "float32" || data[first] != first {
		return fmt.Errorf("general map key type, membership or value identity changed")
	}
	if data[uint16(65)] != "character" || data[int16(65)] != "short" || data[uint16(0xd800)] != "surrogate" {
		return fmt.Errorf("character and signed-short map keys were merged or changed")
	}
	identity := reflect.ValueOf(data).Pointer()
	for _, shared := range []any{second.Data, data[second], data["self"], data["array"].([]any)[0]} {
		if reflect.ValueOf(shared).Pointer() != identity {
			return fmt.Errorf("recursive general map sharing changed")
		}
	}
	if data["array"].([]any)[1] != first {
		return fmt.Errorf("general map lost references through nested arrays")
	}
	return nil
}

// ModelValue.data is an Object in Java. This checks its native Go general-map
// representation; it adds no credit to the existing original ModelValue methods.
func TestDistributedModelDataGeneralKeys(t *testing.T) {
	states := modelDataKeyStates()
	got := distributedPayloadRoundTrip(t, states)
	if err := checkModelDataKeyStates(got); err != nil {
		t.Fatal(err)
	}
	got[0].values[0].(*ModelValue).Data.(map[any]any)[true].([]byte)[0] = 97
	if states[0].values[0].(*ModelValue).Data.(map[any]any)[true].([]byte)[0] != 17 {
		t.Fatal("general map receiver aliases sender storage")
	}
	for _, original := range []map[any]any{nil, {}} {
		copied := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: original}}}})[0].values[0].(*ModelValue).Data.(map[any]any)
		if (copied == nil) != (original == nil) || len(copied) != 0 {
			t.Fatal("general maps lost nil versus empty distinction")
		}
	}
	separate := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: map[any]any{1: 2}}, &ModelValue{Data: map[any]any{1: 2}}}}})[0].values
	separate[0].(*ModelValue).Data.(map[any]any)[1] = 3
	if separate[1].(*ModelValue).Data.(map[any]any)[1] != 2 {
		t.Fatal("equal general maps were merged")
	}
}

func TestWorkerRPCModelDataGeneralKeys(t *testing.T) {
	states := modelDataKeyStates()
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if err := checkModelDataKeyStates(received); err != nil {
			return nil, err
		}
		if received[0].UID == 1 {
			return nil, NewWorkerException("general map failure", received[0], received[0], true)
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received)}, []*LongVec{NewLongVec()}, 1, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkModelDataKeyStates([]*TLCStateMut{result.NextStates[0].At(0)}); err != nil {
		t.Fatal(err)
	}
	states[0].UID = 1
	result, err = client.GetNextStates(states)
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("general map failure lost state identity: %v/%v", result, err)
	}
	if err := checkModelDataKeyStates([]*TLCStateMut{failure.State1}); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedModelDataGeneralKeysInvalid(t *testing.T) {
	for _, data := range []map[any]any{{struct{}{}: 1}, {1: struct{}{}}} {
		if _, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: data}}}}); err == nil {
			t.Fatal("unsupported general map entry was silently dropped")
		}
	}
	for _, key := range []DistributedObjectDataNode{
		{Kind: "objectKeyMap", Map: 1}, {Kind: "objectArray"}, {Kind: "bytes"},
		{Kind: "value", Value: 1}, {Kind: "objectKeyMap", Map: -1}, {Kind: "objectKeyMap", Map: 2}, {Kind: "unknown"},
	} {
		payload := &DistributedStatePayload{ObjectKeyMaps: [][]DistributedObjectKeyMapEntry{{{Key: key}}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid general map key accepted: %+v", key)
		}
	}
	for _, entries := range [][]DistributedObjectKeyMapEntry{
		{{}, {}},
		{{Key: DistributedObjectDataNode{Kind: "int", Integer: 1}}, {Key: DistributedObjectDataNode{Kind: "int", Integer: 1}}},
		{{Data: DistributedObjectDataNode{Kind: "objectKeyMap", Map: 2}}},
	} {
		if _, err := DecodeDistributedStates(&DistributedStatePayload{ObjectKeyMaps: [][]DistributedObjectKeyMapEntry{entries}}); err == nil {
			t.Fatal("invalid general map entries accepted")
		}
	}
}
