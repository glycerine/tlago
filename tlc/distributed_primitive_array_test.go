package tlc

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

// No original Java method directly tests primitive-array attachments on the
// worker boundary. These checks exercise the native payload and TCP transport.
func distributedPrimitiveArrayExamples() []any {
	return []any{
		[]bool{true, false},
		[]int{math.MinInt, 0, math.MaxInt},
		[]int8{math.MinInt8, 0, math.MaxInt8},
		[]int16{math.MinInt16, 0, math.MaxInt16},
		[]int32{math.MinInt32, 0, math.MaxInt32},
		[]int64{math.MinInt64, 0, math.MaxInt64},
		[]uint16{0, 0xd800, 0xdfff, math.MaxUint16},
		[]float32{1, 0, math.Float32frombits(1 << 31), math.Float32frombits(1), math.MaxFloat32, float32(math.Inf(1)), float32(math.Inf(-1)), math.Float32frombits(0x7fc01234)},
		[]float64{1, 0, math.Float64frombits(1 << 63), math.Float64frombits(1), math.MaxFloat64, math.Inf(1), math.Inf(-1), math.Float64frombits(0x7ff8000000001234)},
	}
}

func distributedPrimitiveArrayStates(data []any) []*TLCStateMut {
	states := make([]*TLCStateMut, len(data))
	for i, array := range data {
		v := reflect.ValueOf(array)
		separate := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		reflect.Copy(separate, v)
		attachments := []any{array, separate.Interface(), []any{array}, reflect.MakeSlice(v.Type(), 0, 0).Interface(), reflect.Zero(v.Type()).Interface()}
		state := &TLCStateMut{level: 1}
		for j, attachment := range attachments {
			state.values = append(state.values, &ModelValue{Val: &UniqueString{s: fmt.Sprintf("A%d_%d", i, j)}, Data: attachment})
		}
		states[i] = state
	}
	return states
}

func distributedPrimitiveArraysEqual(got, want any) bool {
	a, b := reflect.ValueOf(got), reflect.ValueOf(want)
	if !a.IsValid() || a.Type() != b.Type() || a.Len() != b.Len() || a.IsNil() != b.IsNil() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		if !distributedScalarDataEqual(a.Index(i).Interface(), b.Index(i).Interface()) {
			return false
		}
	}
	return true
}

func checkDistributedPrimitiveArrays(states []*TLCStateMut, data []any) error {
	if len(states) != len(data) {
		return fmt.Errorf("state count %d, want %d", len(states), len(data))
	}
	for i, expected := range data {
		if len(states[i].values) != 5 {
			return fmt.Errorf("%T attachment count changed", expected)
		}
		attachment := func(j int) any { return states[i].values[j].(*ModelValue).Data }
		a, b, c := attachment(0), attachment(1), attachment(2).([]any)[0]
		for _, array := range []any{a, b, c} {
			if !distributedPrimitiveArraysEqual(array, expected) {
				return fmt.Errorf("%T element type or bits changed", expected)
			}
		}
		pa, pb, pc := reflect.ValueOf(a).Pointer(), reflect.ValueOf(b).Pointer(), reflect.ValueOf(c).Pointer()
		if pa != pc || pa == pb || pa == reflect.ValueOf(expected).Pointer() {
			return fmt.Errorf("%T lost sharing or aliases separate/sender data", expected)
		}
		empty, nilArray := reflect.ValueOf(attachment(3)), reflect.ValueOf(attachment(4))
		if empty.Type() != reflect.TypeOf(expected) || nilArray.Type() != empty.Type() || empty.Len() != 0 || empty.IsNil() || !nilArray.IsNil() {
			return fmt.Errorf("%T nil and empty representations changed", expected)
		}
	}
	return nil
}

func TestDistributedModelPrimitiveArrays(t *testing.T) {
	checkDistributedModelArrays(t, distributedPrimitiveArrayExamples())
}

func checkDistributedModelArrays(t *testing.T, data []any) {
	t.Helper()
	copied := distributedPayloadRoundTrip(t, distributedPrimitiveArrayStates(data))
	if err := checkDistributedPrimitiveArrays(copied, data); err != nil {
		t.Fatal(err)
	}
	for i, original := range data {
		values := copied[i].values
		a := reflect.ValueOf(values[0].(*ModelValue).Data)
		b := reflect.ValueOf(values[1].(*ModelValue).Data)
		c := reflect.ValueOf(values[2].(*ModelValue).Data.([]any)[0])
		old := reflect.ValueOf(original).Index(0).Interface()
		a.Index(0).Set(reflect.Zero(a.Type().Elem()))
		if !reflect.DeepEqual(c.Index(0).Interface(), a.Index(0).Interface()) || !reflect.DeepEqual(b.Index(0).Interface(), old) || !reflect.DeepEqual(reflect.ValueOf(original).Index(0).Interface(), old) {
			t.Fatalf("%T receiver mutation changed ownership", original)
		}
	}
}

func TestDistributedModelPrimitiveArraysWorkerRPC(t *testing.T) {
	checkDistributedModelArraysWorkerRPC(t, distributedPrimitiveArrayExamples())
}

func checkDistributedModelArraysWorkerRPC(t *testing.T, data []any) {
	t.Helper()
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if err := checkDistributedPrimitiveArrays(received, data); err != nil {
			return nil, err
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received)}, []*LongVec{NewLongVec()}, 0, int64(len(received))), nil
	}})
	result, err := client.GetNextStates(distributedPrimitiveArrayStates(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NextStates) != 1 || result.NextStates[0].Size() != len(data) {
		t.Fatal("worker lost result states")
	}
	states := make([]*TLCStateMut, len(data))
	for i := range states {
		states[i] = result.NextStates[0].At(i)
	}
	if err := checkDistributedPrimitiveArrays(states, data); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedModelPrimitiveArraysInvalidPayload(t *testing.T) {
	for _, test := range []struct {
		name, kind, referenceKind string
		bits                      uint64
		reference                 int
	}{
		{"negative_reference", "boolArray", "boolArray", 0, -1},
		{"missing_reference", "boolArray", "boolArray", 0, 2},
		{"kind_mismatch", "boolArray", "int8Array", 0, 1},
		{"unknown_kind", "unknownArray", "boolArray", 0, 1},
		{"bool", "boolArray", "boolArray", 2, 1},
		{"int8_positive", "int8Array", "int8Array", 128, 1},
		{"int8_negative", "int8Array", "int8Array", ^uint64(128), 1},
		{"int16_positive", "int16Array", "int16Array", 32768, 1},
		{"int16_negative", "int16Array", "int16Array", ^uint64(32768), 1},
		{"int32_positive", "int32Array", "int32Array", 1 << 31, 1},
		{"int32_negative", "int32Array", "int32Array", ^uint64(1 << 31), 1},
		{"uint16", "uint16Array", "uint16Array", 1 << 16, 1},
		{"float32", "float32Array", "float32Array", 1 << 32, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			payload := &DistributedStatePayload{
				Roots: []int{1}, States: []DistributedStateNode{{Level: 1, Values: []int{1}}},
				Values:          []DistributedValueNode{{Kind: "model", DataKind: test.referenceKind, DataArray: test.reference}},
				PrimitiveArrays: []DistributedPrimitiveArrayNode{{Kind: test.kind, Bits: []uint64{test.bits}}},
			}
			if _, err := DecodeDistributedStates(payload); err == nil {
				t.Fatal("invalid primitive array payload accepted")
			}
		})
	}
}
