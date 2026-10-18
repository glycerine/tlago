package tlc

import (
	"fmt"
	"math"
	"testing"
)

// ModelValue.data is non-transient source Object data. Mixed arrays/maps of
// transferable entries retain identity and cycles; no original method directly
// tests this attached-data networking boundary.
func modelDataObjectStates() []*TLCStateMut {
	first := &ModelValue{Val: &UniqueString{s: "first", tok: 1}}
	second := &ModelValue{Val: &UniqueString{s: "second", tok: 2}, Index: 1}
	array := make([]any, 5)
	bytes := []byte{0, 17, 255}
	data := map[string]any{
		"array": array, "value": first, "bytes": bytes, "": nil,
		"blue\x00雪": "mixed", "small": int8(-7), "nan": math.Float64frombits(0x7ff8000000000041),
		"values": []Value{first, second}, "valueMap": map[string]Value{"first": first},
		"nilArray": []any(nil), "emptyArray": []any{},
		"nilMap": map[string]any(nil), "emptyMap": map[string]any{},
	}
	data["self"] = data
	array[0], array[1], array[2], array[3], array[4] = data, array, first, bytes, int64(-99)
	first.Data, second.Data = data, array
	return []*TLCStateMut{{level: 1, values: []Value{first, second}}, {level: 2, values: []Value{first, second}}}
}

func checkModelDataObjectStates(states []*TLCStateMut) error {
	if len(states) != 2 || len(states[0].values) != 2 || len(states[1].values) != 2 {
		return fmt.Errorf("mixed graph state shape changed")
	}
	first, ok := states[0].values[0].(*ModelValue)
	if !ok || states[1].values[0] != first || states[1].values[1] != states[0].values[1] {
		return fmt.Errorf("mixed graph model sharing changed")
	}
	second := states[0].values[1].(*ModelValue)
	data, ok := first.Data.(map[string]any)
	if !ok || data["value"] != first || data["blue\x00雪"] != "mixed" || data["small"] != int8(-7) || math.Float64bits(data["nan"].(float64)) != 0x7ff8000000000041 {
		return fmt.Errorf("mixed map data lost scalar/value representation")
	}
	if value, present := data[""]; !present || value != nil {
		return fmt.Errorf("mixed map lost present nil entry")
	}
	array := second.Data.([]any)
	other := data["array"].([]any)
	if len(array) != 5 || &array[0] != &other[0] || &array[0] != &array[1].([]any)[0] || array[2] != first || array[4] != int64(-99) {
		return fmt.Errorf("mixed array sharing or recursive entries changed")
	}
	data["identity"] = true
	if data["self"].(map[string]any)["identity"] != true || array[0].(map[string]any)["identity"] != true {
		return fmt.Errorf("mixed recursive maps were copied independently")
	}
	delete(data, "identity")
	bytes := data["bytes"].([]byte)
	if len(bytes) != 3 || &bytes[0] != &array[3].([]byte)[0] || bytes[1] != 17 {
		return fmt.Errorf("nested byte-array sharing changed")
	}
	values := data["values"].([]Value)
	if values[0] != first || values[1] != second || data["valueMap"].(map[string]Value)["first"] != first {
		return fmt.Errorf("mixed attachments lost references to the value graph")
	}
	if data["nilArray"].([]any) != nil || data["emptyArray"].([]any) == nil || data["nilMap"].(map[string]any) != nil || data["emptyMap"].(map[string]any) == nil {
		return fmt.Errorf("mixed containers lost typed nil/empty distinctions")
	}
	return nil
}

func TestDistributedModelDataObjects(t *testing.T) {
	states := modelDataObjectStates()
	got := distributedPayloadRoundTrip(t, states)
	if err := checkModelDataObjectStates(got); err != nil {
		t.Fatal(err)
	}
	data := got[0].values[0].(*ModelValue).Data.(map[string]any)
	data["bytes"].([]byte)[1] = 97
	data["changed"] = true
	sender := states[0].values[0].(*ModelValue).Data.(map[string]any)
	if sender["bytes"].([]byte)[1] != 17 || sender["changed"] != nil {
		t.Fatal("mixed receiver containers alias sender storage")
	}
	for _, empty := range []any{[]any(nil), []any{}, map[string]any(nil), map[string]any{}} {
		copied := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: empty}}}})[0].values[0].(*ModelValue).Data
		switch original := empty.(type) {
		case []any:
			if actual, ok := copied.([]any); !ok || (actual == nil) != (original == nil) || len(actual) != 0 {
				t.Fatal("root object array lost its type or nil distinction")
			}
		case map[string]any:
			if actual, ok := copied.(map[string]any); !ok || (actual == nil) != (original == nil) || len(actual) != 0 {
				t.Fatal("root object map lost its type or nil distinction")
			}
		}
	}
	separate := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{
		&ModelValue{Data: map[string]any{"key": 1}}, &ModelValue{Data: map[string]any{"key": 1}},
		&ModelValue{Data: []any{1}}, &ModelValue{Data: []any{1}},
	}}})[0].values
	separate[0].(*ModelValue).Data.(map[string]any)["key"] = 2
	separate[2].(*ModelValue).Data.([]any)[0] = 2
	if separate[1].(*ModelValue).Data.(map[string]any)["key"] != 1 || separate[3].(*ModelValue).Data.([]any)[0] != 1 {
		t.Fatal("equal mixed containers were merged")
	}
}

func TestDistributedModelDataObjectsInvalid(t *testing.T) {
	for _, data := range []any{[]any{struct{}{}}, map[string]any{"opaque": struct{}{}}} {
		if _, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: data}}}}); err == nil {
			t.Fatal("mixed container silently discarded an opaque entry")
		}
	}
	for _, node := range []DistributedValueNode{
		{DataKind: "objectArray", DataArray: -1}, {DataKind: "objectArray", DataArray: 2},
		{DataKind: "objectMap", DataMap: -1}, {DataKind: "objectMap", DataMap: 2},
		{DataKind: "value", DataValue: 2}, {DataKind: "int8", DataInteger: 128}, {DataKind: "unknown"},
	} {
		payload := &DistributedStatePayload{ObjectArrays: [][]DistributedObjectDataNode{{distributedObjectDataNode(node)}}, ObjectMaps: [][]DistributedObjectMapEntry{{}}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid mixed array entry accepted: %+v", node)
		}
		payload.ObjectArrays[0] = nil
		payload.ObjectMaps[0] = []DistributedObjectMapEntry{{Key: "bad", Data: distributedObjectDataNode(node)}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid mixed map entry accepted: %+v", node)
		}
	}
	if _, err := DecodeDistributedStates(&DistributedStatePayload{ObjectMaps: [][]DistributedObjectMapEntry{{{Key: "same"}, {Key: "same"}}}}); err == nil {
		t.Fatal("duplicate mixed map keys accepted")
	}
}

func TestWorkerRPCModelDataObjects(t *testing.T) {
	states := modelDataObjectStates()
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if len(received) == 1 {
			return nil, NewWorkerException("mixed attachment failure", received[0], received[0], true)
		}
		if err := checkModelDataObjectStates(received); err != nil {
			return nil, err
		}
		received[0].values[0].(*ModelValue).Data.(map[string]any)["rpc"] = int64(97)
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received)}, []*LongVec{NewLongVec()}, 1, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	got := []*TLCStateMut{result.NextStates[0].At(0), result.NextStates[0].At(1)}
	if err := checkModelDataObjectStates(got); err != nil {
		t.Fatal(err)
	}
	if got[0].values[0].(*ModelValue).Data.(map[string]any)["rpc"] != int64(97) || states[0].values[0].(*ModelValue).Data.(map[string]any)["rpc"] != nil {
		t.Fatal("RPC mixed mutation lost ownership")
	}
	result, err = client.GetNextStates(states[:1])
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("mixed failure lost state identity: %v/%v", result, err)
	}
	if err := checkModelDataObjectStates([]*TLCStateMut{failure.State1, failure.State2}); err != nil {
		t.Fatal(err)
	}
}
