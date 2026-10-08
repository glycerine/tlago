package tlc

import "testing"

// Record and record-set constructors retain their name arrays. No enabled
// original test directly covers shared arrays across this transfer boundary.
func TestDistributedNameArraySharing(t *testing.T) {
	shared := []*UniqueString{{s: "a"}, {s: "b"}}
	separate := append([]*UniqueString(nil), shared...)
	values := []Value{NewIntValue(1), NewIntValue(2)}
	first, second := NewRecordValue(shared, values, true), NewRecordValue(shared, values, true)
	set := &SetOfRcdsValue{Names: shared, Values: values}
	input := []*TLCStateMut{{level: 1, values: []Value{first, second, set, NewRecordValue(separate, values, true), &StringValue{Val: shared[0]}, &RecordValue{}, &RecordValue{Names: []*UniqueString{}, Values: []Value{}}}}}
	got := distributedPayloadRoundTrip(t, input)[0].values
	a, b, c := got[0].(*RecordValue), got[1].(*RecordValue), got[2].(*SetOfRcdsValue)
	other := got[3].(*RecordValue)
	if a == b || &a.Names[0] != &b.Names[0] || &a.Names[0] != &c.Names[0] || &a.Names[0] == &other.Names[0] || &a.Names[0] == &shared[0] {
		t.Fatal("name-array identity or receiver isolation lost")
	}
	if a.Names[0] != other.Names[0] || a.Names[0] != got[4].(*StringValue).Val || a.Names[0] == shared[0] {
		t.Fatal("shared name object identity or receiver isolation lost")
	}
	a.Names[0] = &UniqueString{s: "changed"}
	if b.Names[0] != a.Names[0] || c.Names[0] != a.Names[0] || other.Names[0].s != "a" || shared[0].s != "a" {
		t.Fatal("name-array mutation did not retain sharing and sender isolation")
	}
	if got[5].(*RecordValue).Names != nil || got[6].(*RecordValue).Names == nil {
		t.Fatal("nil and initialized empty name arrays conflated")
	}
}

func TestDistributedNameArrayWorkerRPC(t *testing.T) {
	shared := []*UniqueString{{s: "a"}}
	states := []*TLCStateMut{
		{level: 1, values: []Value{NewRecordValue(shared, []Value{NewIntValue(1)}, true)}},
		{level: 2, values: []Value{NewRecordValue(shared, []Value{NewIntValue(2)}, true)}},
	}
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		a, b := received[0].values[0].(*RecordValue), received[1].values[0].(*RecordValue)
		if &a.Names[0] != &b.Names[0] {
			return nil, NewRuntimeException("worker request lost shared name array")
		}
		a.Names[0] = &UniqueString{s: "changed"}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received[:1]), NewStateVecFrom(received[1:])}, []*LongVec{NewLongVec(), NewLongVec()}, 0, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	a := result.NextStates[0].At(0).values[0].(*RecordValue)
	b := result.NextStates[1].At(0).values[0].(*RecordValue)
	if a == b || &a.Names[0] != &b.Names[0] || a.Names[0].s != "changed" || &a.Names[0] == &shared[0] || shared[0].s != "a" {
		t.Fatal("worker result lost shared name array or sender isolation")
	}
}

func TestDistributedNameArrayInvalidReferences(t *testing.T) {
	for _, node := range []DistributedValueNode{
		{Kind: "record", NamesArray: -1},
		{Kind: "record", NamesArray: 2},
		{Kind: "record", NamesArray: 1, NamesNil: true},
		{Kind: "record", NamesArray: 1, Names: []int{0}},
	} {
		if _, err := DecodeDistributedStates(&DistributedStatePayload{Values: []DistributedValueNode{node}, NameArrays: [][]int{{0}}}); err == nil {
			t.Fatal("invalid/conflicting name array reference accepted")
		}
	}
	for _, id := range []int{-1, 1} {
		if _, err := DecodeDistributedStates(&DistributedStatePayload{NameArrays: [][]int{{id}}}); err == nil {
			t.Fatal("invalid name object reference in array table accepted")
		}
	}
}
