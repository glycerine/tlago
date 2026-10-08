package tlc

import "testing"

// The source tuple/record/function fields use default object serialization;
// their shared Value[] is one mutable array, not merely equal element values.
// No enabled original method directly tests this network graph boundary.
func TestDistributedValueArraySharing(t *testing.T) {
	shared := []Value{NewIntValue(1), NewIntValue(2)}
	separate := append([]Value(nil), shared...)
	first, second := NewTupleValue(shared), NewTupleValue(shared)
	record := NewRecordValue([]*UniqueString{{s: "a"}, {s: "b"}}, shared, true)
	function := NewFcnRcdValue(shared, shared, true)
	operator := NewOpRcdValueFrom([][]Value{shared, shared}, []Value{NewIntValue(3), NewIntValue(4)})
	product := &SetOfTuplesValue{Sets: shared}
	recordSet := &SetOfRcdsValue{Names: record.Names, Values: shared}
	input := []*TLCStateMut{{level: 1, values: []Value{first, second, record, function, NewTupleValue(separate), operator, product, recordSet}}}
	got := distributedPayloadRoundTrip(t, input)[0].values
	a, b := got[0].(*TupleValue), got[1].(*TupleValue)
	r, f := got[2].(*RecordValue), got[3].(*FcnRcdValue)
	independent := got[4].(*TupleValue)
	op := got[5].(*OpRcdValue)
	p, rs := got[6].(*SetOfTuplesValue), got[7].(*SetOfRcdsValue)
	if a == b || a == first || &a.Elems[0] != &b.Elems[0] || &a.Elems[0] != &r.Values[0] || &a.Elems[0] != &f.Values[0] || &a.Elems[0] != &f.Domain[0] || &a.Elems[0] == &shared[0] || &a.Elems[0] == &independent.Elems[0] {
		t.Fatal("value-array identity or isolated receiver ownership was lost")
	}
	a.Elems[0] = NewIntValue(9)
	for _, array := range [][]Value{b.Elems, r.Values, f.Values, f.Domain, op.Domain[0], op.Domain[1], p.Sets, rs.Values} {
		if array[0] != a.Elems[0] {
			t.Fatal("receiver mutation did not reach shared value array")
		}
	}
	if shared[0].(*IntValue).Val != 1 || independent.Elems[0].(*IntValue).Val != 1 {
		t.Fatal("receiver mutation changed sender or separate equal-content array")
	}
}

func TestDistributedValueArrayCyclesAndNulls(t *testing.T) {
	cycle := &TupleValue{}
	cycle.Elems = []Value{cycle}
	alias := &TupleValue{Elems: cycle.Elems}
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{cycle, alias, &TupleValue{}, &TupleValue{Elems: []Value{}}}}})[0].values
	first, second := got[0].(*TupleValue), got[1].(*TupleValue)
	if first.Elems[0] != first || second.Elems[0] != first || &first.Elems[0] != &second.Elems[0] {
		t.Fatal("recursive value array identity was lost")
	}
	if got[2].(*TupleValue).Elems != nil || got[3].(*TupleValue).Elems == nil {
		t.Fatal("nil and initialized empty arrays were conflated")
	}
}

func TestDistributedValueArrayInvalidReferences(t *testing.T) {
	for _, test := range []struct {
		array     int
		nilValues bool
		inline    []int
	}{
		{-1, false, nil}, {2, false, nil}, {1, true, nil}, {1, false, []int{0}},
	} {
		payload := &DistributedStatePayload{
			ValueArrays: [][]int{{0}}, Values: []DistributedValueNode{{Kind: "tuple", ReferencesArray: test.array, ReferencesNil: test.nilValues, References: test.inline}},
		}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatal("invalid/conflicting value array reference accepted")
		}
	}
	if _, err := DecodeDistributedStates(&DistributedStatePayload{ValueArrays: [][]int{{1}}}); err == nil {
		t.Fatal("invalid value reference in array table accepted")
	}
}

func TestDistributedValueArrayWorkerRPC(t *testing.T) {
	shared := []Value{NewIntValue(1)}
	states := []*TLCStateMut{
		{level: 1, values: []Value{NewTupleValue(shared)}},
		{level: 2, values: []Value{NewTupleValue(shared)}},
	}
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		first, second := received[0].values[0].(*TupleValue), received[1].values[0].(*TupleValue)
		if &first.Elems[0] != &second.Elems[0] {
			return nil, NewRuntimeException("worker request lost shared value array")
		}
		first.Elems[0] = NewIntValue(7)
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received[:1]), NewStateVecFrom(received[1:])}, []*LongVec{NewLongVec(), NewLongVec()}, 0, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	first := result.NextStates[0].At(0).values[0].(*TupleValue)
	second := result.NextStates[1].At(0).values[0].(*TupleValue)
	if first == second || &first.Elems[0] != &second.Elems[0] || first.Elems[0].(*IntValue).Val != 7 || shared[0].(*IntValue).Val != 1 || &first.Elems[0] == &shared[0] {
		t.Fatal("worker result lost shared array or isolated ownership")
	}
}
