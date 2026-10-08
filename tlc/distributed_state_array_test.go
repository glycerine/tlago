package tlc

import "testing"

// No original method directly exercises state backing arrays at the network
// boundary. These checks cover the native payload's graph and ownership contract.
func TestDistributedStateArraySharing(t *testing.T) {
	shared := []Value{NewIntValue(1), nil}
	tuple := NewTupleValue(shared)
	shared[1] = tuple
	separate := append([]Value(nil), shared...)
	first := &TLCStateMut{level: 1, values: shared}
	second := &TLCStateMut{level: 2, values: shared}
	got := distributedPayloadRoundTrip(t, []*TLCStateMut{first, second, {level: 3, values: separate}, {level: 1}, {level: 1, values: []Value{}}})
	receivedTuple := got[0].values[1].(*TupleValue)
	if got[0] == got[1] || &got[0].values[0] != &got[1].values[0] || &got[0].values[0] != &receivedTuple.Elems[0] || &got[0].values[0] == &shared[0] || &got[0].values[0] == &got[2].values[0] {
		t.Fatal("state/value backing array identity or receiver ownership was lost")
	}
	if receivedTuple.Elems[1] != receivedTuple {
		t.Fatal("recursive state/value array was lost")
	}
	got[1].values[0] = NewIntValue(9)
	if got[0].values[0] != got[1].values[0] || receivedTuple.Elems[0] != got[1].values[0] || shared[0].(*IntValue).Val != 1 || got[2].values[0].(*IntValue).Val != 1 {
		t.Fatal("mutation did not preserve sharing and separate ownership")
	}
	if got[3].values != nil || got[4].values == nil {
		t.Fatal("nil and initialized empty state arrays were conflated")
	}
}

func TestDistributedStateArrayInvalidReferences(t *testing.T) {
	for _, node := range []DistributedStateNode{
		{ValuesArray: -1}, {ValuesArray: 2},
		{ValuesArray: 1, ValuesNil: true}, {ValuesArray: 1, Values: []int{0}},
	} {
		payload := &DistributedStatePayload{ValueArrays: [][]int{{0}}, States: []DistributedStateNode{node}, Roots: []int{1}}
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatalf("invalid/conflicting state array reference accepted: %+v", node)
		}
	}
}

func TestDistributedStateArrayWorkerRPC(t *testing.T) {
	shared := []Value{NewIntValue(1)}
	states := []*TLCStateMut{{level: 1, values: shared}, {level: 2, values: shared}}
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if received[0] == received[1] || &received[0].values[0] != &received[1].values[0] {
			return nil, NewRuntimeException("worker request lost shared state array")
		}
		received[0].values[0] = NewIntValue(7)
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received[:1]), NewStateVecFrom(received[1:])}, []*LongVec{NewLongVec(), NewLongVec()}, 0, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	first, second := result.NextStates[0].At(0), result.NextStates[1].At(0)
	if first == second || &first.values[0] != &second.values[0] || first.values[0].(*IntValue).Val != 7 || shared[0].(*IntValue).Val != 1 || &first.values[0] == &shared[0] {
		t.Fatal("worker result lost shared state array or isolated ownership")
	}
}
