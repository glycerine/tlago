package tlc

import (
	"fmt"
	"testing"
)

func operatorSharingStates() []*TLCStateMut {
	rows := [][]Value{{NewIntValue(1)}, nil, {}}
	values := []Value{NewIntValue(7), NewBoolValue(false), NewBoolValue(true)}
	first, second := NewOpRcdValueFrom(rows, values), NewOpRcdValueFrom(rows, values)
	return []*TLCStateMut{{level: 1, values: []Value{first, second, &ModelValue{Data: rows}, &ModelValue{Data: values}}}}
}

func checkOperatorSharingStates(states []*TLCStateMut) error {
	first, second := states[0].values[0].(*OpRcdValue), states[0].values[1].(*OpRcdValue)
	if len(first.Domain) != 3 || len(first.Values) != 3 || first.Domain[1] != nil || first.Domain[2] == nil {
		return fmt.Errorf("operator rows lost nil/empty or count distinctions")
	}
	if &first.Domain[0] != &second.Domain[0] || &first.Values[0] != &second.Values[0] {
		return fmt.Errorf("operator constructor's shared containers were copied independently")
	}
	rows := states[0].values[2].(*ModelValue).Data.([][]Value)
	values := states[0].values[3].(*ModelValue).Data.([]Value)
	if &rows[0] != &first.Domain[0] || &values[0] != &first.Values[0] {
		return fmt.Errorf("operator containers lost sharing with model data")
	}
	first.Domain[0] = []Value{NewIntValue(2)}
	first.Values[0] = NewIntValue(9)
	for _, operator := range []*OpRcdValue{first, second} {
		result, err := operator.Eval([]Value{NewIntValue(2)}, 0)
		if err != nil || result != first.Values[0] {
			return fmt.Errorf("shared operator update changed application: %v/%v", result, err)
		}
	}
	return nil
}

func TestWorkerRPCOperatorContainerSharing(t *testing.T) {
	states := operatorSharingStates()
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(received []*TLCStateMut) (*NextStateResult, error) {
		if err := checkOperatorSharingStates(received); err != nil {
			return nil, err
		}
		if received[0].UID == 1 {
			return nil, NewWorkerException("shared operator failure", received[0], received[0], true)
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(received)}, []*LongVec{NewLongVec()}, 1, 2), nil
	}})
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkOperatorSharingStates([]*TLCStateMut{result.NextStates[0].At(0)}); err != nil {
		t.Fatal(err)
	}
	first := states[0].values[0].(*OpRcdValue)
	if first.Domain[0][0].(*IntValue).Val != 1 || first.Values[0].(*IntValue).Val != 7 {
		t.Fatal("worker changes reached sender operator")
	}
	states[0].UID = 1
	result, err = client.GetNextStates(states)
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.State1 != failure.State2 || !failure.KeepCallStack {
		t.Fatalf("shared operator failure lost state identity: %v/%v", result, err)
	}
	if err := checkOperatorSharingStates([]*TLCStateMut{failure.State1}); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedOperatorContainerSharingInvalid(t *testing.T) {
	for _, node := range []DistributedValueNode{
		{Kind: "operatorRecord", OperatorDomainArray: -1},
		{Kind: "operatorRecord", OperatorDomainArray: 2},
		{Kind: "operatorRecord", OperatorDomainArray: 1, OperatorDomainNil: true},
		{Kind: "operatorRecord", OperatorDomainArray: 1, OperatorDomain: []DistributedValueReferences{{}}},
		{Kind: "model", DataKind: "valueRows", DataArray: -1},
		{Kind: "model", DataKind: "valueRows", DataArray: 2},
	} {
		if _, err := DecodeDistributedStates(&DistributedStatePayload{Values: []DistributedValueNode{node}, ValueRows: [][]int{{}}}); err == nil {
			t.Fatalf("invalid shared operator row reference accepted: %+v", node)
		}
	}
	for _, row := range []int{-1, 1} {
		if _, err := DecodeDistributedStates(&DistributedStatePayload{ValueRows: [][]int{{row}}}); err == nil {
			t.Fatalf("invalid inner operator row reference %d accepted", row)
		}
	}
	for _, rows := range [][][]Value{nil, {}, {nil, {}}} {
		got := distributedPayloadRoundTrip(t, []*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: rows}}}})[0].values[0].(*ModelValue).Data.([][]Value)
		if (rows == nil) != (got == nil) || len(rows) != len(got) {
			t.Fatal("attached value rows lost nil/empty distinction")
		}
		if len(got) == 2 && (got[0] != nil || got[1] == nil) {
			t.Fatal("attached inner rows lost nil/empty distinction")
		}
	}
}

// No direct Java test covers finite operator containers crossing the network.
// The constructor retains supplied Vect objects; the native slices must remain
// shared rather than changing operator application after a round trip.
func TestDistributedOperatorContainerSharing(t *testing.T) {
	states := operatorSharingStates()
	got := distributedPayloadRoundTrip(t, states)
	if err := checkOperatorSharingStates(got); err != nil {
		t.Fatal(err)
	}
	first := states[0].values[0].(*OpRcdValue)
	if first.Domain[0][0].(*IntValue).Val != 1 || first.Values[0].(*IntValue).Val != 7 {
		t.Fatal("receiver operator changes reached sender")
	}
}
