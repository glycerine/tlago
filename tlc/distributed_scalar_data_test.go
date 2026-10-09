package tlc

import (
	"math"
	"reflect"
	"testing"
)

func distributedScalarData() []any {
	return []any{
		int8(math.MinInt8), int8(0), int8(math.MaxInt8),
		int16(math.MinInt16), int16(0), int16(math.MaxInt16),
		uint16(0), uint16(0x7f), uint16(0xff), uint16(0x8000),
		uint16(0xd800), uint16(0xdc00), uint16(0xdfff), uint16(math.MaxUint16),
		float32(0), math.Float32frombits(1 << 31), float32(0.1),
		float32(math.MaxFloat32), math.Float32frombits(1),
		float32(math.Inf(1)), float32(math.Inf(-1)), float32(math.NaN()),
		float64(0), math.Float64frombits(1 << 63), float64(0.1),
		float64(math.MaxFloat64), math.Float64frombits(1),
		math.Inf(1), math.Inf(-1), math.NaN(),
	}
}

func distributedScalarDataEqual(got, want any) bool {
	switch expected := want.(type) {
	case float32:
		actual, ok := got.(float32)
		return ok && math.Float32bits(actual) == math.Float32bits(expected)
	case float64:
		actual, ok := got.(float64)
		return ok && math.Float64bits(actual) == math.Float64bits(expected)
	default:
		return reflect.DeepEqual(got, want)
	}
}

func requireDistributedScalarData(t *testing.T, got, want any) {
	t.Helper()
	if !distributedScalarDataEqual(got, want) {
		t.Fatalf("scalar data = %T(%v), want %T(%v)", got, got, want, want)
	}
}

// ModelValue.data is an ordinary non-transient Object in Java. Its boxed byte,
// short, character and float data are transferable; Go must retain their scalar
// types too. No enabled upstream test directly covers this transport boundary.
func TestDistributedModelScalarData(t *testing.T) {
	for _, data := range distributedScalarData() {
		state := &TLCStateMut{level: 1, values: []Value{&ModelValue{Data: data}}}
		copied := distributedPayloadRoundTrip(t, []*TLCStateMut{state})
		requireDistributedScalarData(t, copied[0].values[0].(*ModelValue).Data, data)
	}
}

func TestDistributedModelScalarDataWorkerRPC(t *testing.T) {
	data := distributedScalarData()
	states := make([]*TLCStateMut, len(data))
	for i, value := range data {
		states[i] = &TLCStateMut{level: 1, values: []Value{&ModelValue{Data: value}}}
	}
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if len(received) != len(data) {
			return nil, NewRuntimeException("worker scalar state count changed")
		}
		// Assert inside the handler without Fatal/Goexit, which would abandon
		// its RPC response. The caller checks exact values on the return path.
		for i, state := range received {
			if !distributedScalarDataEqual(state.values[0].(*ModelValue).Data, data[i]) {
				return nil, NewRuntimeException("worker scalar data type or value changed")
			}
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received)}, []*LongVec{NewLongVec()}, 0, int64(len(received))), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.NextStates) != 1 || result.NextStates[0].Size() != len(data) {
		t.Fatal("worker result lost scalar states")
	}
	for i, expected := range data {
		requireDistributedScalarData(t, result.NextStates[0].At(i).values[0].(*ModelValue).Data, expected)
	}
}

func TestDistributedModelFloat32InvalidBits(t *testing.T) {
	payload := &DistributedStatePayload{
		Roots: []int{1}, States: []DistributedStateNode{{Level: 1, Values: []int{1}}},
		Values: []DistributedValueNode{{Kind: "model", DataKind: "float32", DataFloatBits: 1 << 32}},
	}
	if _, err := DecodeDistributedStates(payload); err == nil {
		t.Fatal("out-of-range float32 bit pattern accepted")
	}
}

func TestDistributedModelNarrowIntegerBounds(t *testing.T) {
	for _, test := range []struct {
		kind    string
		integer int64
	}{
		{"int8", math.MinInt8 - 1}, {"int8", math.MaxInt8 + 1},
		{"int16", math.MinInt16 - 1}, {"int16", math.MaxInt16 + 1},
		{"uint16", -1}, {"uint16", math.MaxUint16 + 1},
	} {
		payload := &DistributedStatePayload{
			Roots: []int{1}, States: []DistributedStateNode{{Level: 1, Values: []int{1}}},
			Values: []DistributedValueNode{{Kind: "model", DataKind: test.kind, DataInteger: test.integer}},
		}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("out-of-range %s data %d accepted", test.kind, test.integer)
		}
	}
}

func TestDistributedModelCharacterUnits(t *testing.T) {
	data := make([]any, 1<<16)
	for i := range data {
		data[i] = uint16(i)
	}
	state := &TLCStateMut{level: 1, values: []Value{&ModelValue{Data: data}}}
	copied := distributedPayloadRoundTrip(t, []*TLCStateMut{state})
	got := copied[0].values[0].(*ModelValue).Data.([]any)
	if len(got) != len(data) {
		t.Fatal("character-unit range changed length")
	}
	for i := range data {
		requireDistributedScalarData(t, got[i], data[i])
	}
}
