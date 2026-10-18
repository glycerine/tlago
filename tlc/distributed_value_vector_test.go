package tlc

import "testing"

// ValueVec has no custom source transfer contract: its identity, count and
// complete backing array survive transfer. No enabled original network test
// directly covers this representation.
func TestDistributedValueVectorSharingAndCapacity(t *testing.T) {
	storage := []Value{NewIntValue(1), NewIntValue(2), nil}
	vector := NewValueVecFrom(storage)
	vector.count = 1
	first, second := NewSetEnumValueVec(vector, false), NewSetEnumValueVec(vector, true)
	other := NewValueVecFrom(storage)
	other.count = 2
	storage[2] = first // Recursive reference in the unused backing storage.
	input := []*TLCStateMut{{level: 1, values: []Value{first, second, NewSetEnumValueVec(other, false), NewTupleValue(storage), &SetEnumValue{}, NewSetEnumValueVec(NewValueVec(0), true)}}}
	got := distributedPayloadRoundTrip(t, input)[0].values
	a, b, c := got[0].(*SetEnumValue), got[1].(*SetEnumValue), got[2].(*SetEnumValue)
	tuple := got[3].(*TupleValue)
	if a == b || a.Elems == vector || a.Elems != b.Elems || a.Elems == c.Elems {
		t.Fatal("shared/distinct vector identity or sender isolation lost")
	}
	if a.Elems.Len() != 1 || a.Elems.Cap() != 3 || c.Elems.Len() != 2 || c.Elems.Cap() != 3 {
		t.Fatal("vector count or backing capacity lost")
	}
	if &a.Elems.data[0] != &c.Elems.data[0] || &a.Elems.data[0] != &tuple.Elems[0] || tuple.Elems[2] != a {
		t.Fatal("vector backing array sharing or unused recursive slot lost")
	}
	a.Elems.Add(NewIntValue(7))
	if b.Elems.Len() != 2 || c.Elems.At(1) != a.Elems.At(1) || tuple.Elems[1] != a.Elems.At(1) || storage[1].(*IntValue).Val != 2 {
		t.Fatal("append did not preserve shared vector count/backing storage")
	}
	if got[4].(*SetEnumValue).Elems != nil || got[5].(*SetEnumValue).Elems == nil || got[5].(*SetEnumValue).Elems.Cap() != 0 {
		t.Fatal("null and initialized empty vectors conflated")
	}
}

func TestDistributedValueVectorWorkerRPC(t *testing.T) {
	vector := NewValueVec(4)
	vector.Add(NewIntValue(1))
	states := []*TLCStateMut{
		{level: 1, values: []Value{NewSetEnumValueVec(vector, false)}},
		{level: 2, values: []Value{NewSetEnumValueVec(vector, true)}},
	}
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		a, b := received[0].values[0].(*SetEnumValue), received[1].values[0].(*SetEnumValue)
		if a.Elems != b.Elems || a.Elems.Cap() != 4 {
			return nil, NewRuntimeException("worker request lost vector identity/capacity")
		}
		a.Elems.Add(NewIntValue(7))
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received[:1]), NewStateVecFrom(received[1:])}, []*LongVec{NewLongVec(), NewLongVec()}, 0, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	a := result.NextStates[0].At(0).values[0].(*SetEnumValue)
	b := result.NextStates[1].At(0).values[0].(*SetEnumValue)
	if a == b || a.Elems != b.Elems || a.Elems == vector || a.Elems.Len() != 2 || a.Elems.Cap() != 4 || b.Elems.At(1).(*IntValue).Val != 7 || vector.Len() != 1 {
		t.Fatal("worker result lost vector identity/capacity or sender isolation")
	}
}

// Counts need not fit backing storage: Java default serialization retains them.
func TestDistributedValueVectorInvalidReferences(t *testing.T) {
	for _, node := range []DistributedValueVectorNode{
		{Array: -1}, {Array: 2},
	} {
		if _, err := DecodeDistributedStates(&DistributedStatePayload{ValueArrays: [][]int{{0}}, ValueVectors: []DistributedValueVectorNode{node}}); err == nil {
			t.Fatal("invalid vector backing reference accepted")
		}
	}
	for _, node := range []DistributedValueNode{
		{Kind: "enum", CollectionPresent: true, Vector: -1},
		{Kind: "enum", CollectionPresent: true, Vector: 2},
		{Kind: "enum", Vector: 1},
		{Kind: "enum", CollectionPresent: true, Vector: 1, References: []int{0}},
		{Kind: "enum", CollectionPresent: true, Vector: 1, ReferencesArray: 1},
	} {
		payload := &DistributedStatePayload{ValueArrays: [][]int{{0}}, ValueVectors: []DistributedValueVectorNode{{Array: 1}}, Values: []DistributedValueNode{node}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatal("invalid/conflicting vector reference accepted")
		}
	}
}
