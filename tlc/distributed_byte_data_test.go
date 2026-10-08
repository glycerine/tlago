package tlc

import "testing"

// Java serializable model data preserves shared byte-array objects. No enabled
// direct Java test covers this native graph boundary.
func TestDistributedModelByteDataSharing(t *testing.T) {
	shared := []byte{1, 2, 3}
	separate := append([]byte(nil), shared...)
	makeModel := func(name string, data []byte) *ModelValue {
		return &ModelValue{Val: &UniqueString{s: name}, Data: data}
	}
	states := []*TLCStateMut{
		{level: 1, values: []Value{makeModel("A", shared), makeModel("B", separate)}},
		{level: 2, values: []Value{makeModel("C", shared), makeModel("D", []byte{}), makeModel("E", nil)}},
	}
	copied := distributedPayloadRoundTrip(t, states)
	a := copied[0].values[0].(*ModelValue).Data.([]byte)
	b := copied[0].values[1].(*ModelValue).Data.([]byte)
	c := copied[1].values[0].(*ModelValue).Data.([]byte)
	if &a[0] != &c[0] || &a[0] == &b[0] || &a[0] == &shared[0] {
		t.Fatal("shared byte data lost graph identity or aliases sender/separate data")
	}
	a[1] = 9
	if c[1] != 9 || b[1] != 2 || shared[1] != 2 {
		t.Fatal("receiver mutation did not preserve byte-buffer ownership and sharing")
	}
	if copied[1].values[1].(*ModelValue).Data.([]byte) == nil || copied[1].values[2].(*ModelValue).Data.([]byte) != nil {
		t.Fatal("nil and initialized empty byte data were conflated")
	}
}

func TestDistributedModelByteDataWorkerRPC(t *testing.T) {
	shared := []byte{4, 5}
	states := []*TLCStateMut{
		{level: 1, values: []Value{&ModelValue{Val: &UniqueString{s: "A"}, Data: shared}}},
		{level: 2, values: []Value{&ModelValue{Val: &UniqueString{s: "B"}, Data: shared}}},
	}
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		first := received[0].values[0].(*ModelValue).Data.([]byte)
		second := received[1].values[0].(*ModelValue).Data.([]byte)
		if &first[0] != &second[0] {
			return nil, NewRuntimeException("worker request lost shared byte data")
		}
		first[1] = 7
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received[:1]), NewStateVecFrom(received[1:])}, []*LongVec{NewLongVec(), NewLongVec()}, 0, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	first := result.NextStates[0].At(0).values[0].(*ModelValue).Data.([]byte)
	second := result.NextStates[1].At(0).values[0].(*ModelValue).Data.([]byte)
	if &first[0] != &second[0] || first[1] != 7 || shared[1] != 5 || &first[0] == &shared[0] {
		t.Fatal("worker result partitions lost byte data sharing or isolated ownership")
	}
}

func TestDistributedModelByteDataInvalidReference(t *testing.T) {
	for _, reference := range []int{-1, 2} {
		payload, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: []byte{1}}}}})
		if err != nil {
			t.Fatal(err)
		}
		payload.Values[0].DataBytes = reference
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid byte reference %d accepted", reference)
		}
	}
}
