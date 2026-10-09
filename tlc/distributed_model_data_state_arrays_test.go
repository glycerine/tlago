package tlc

import (
	"fmt"
	"sync/atomic"
	"testing"
)

func modelDataStateArrayGraph() []*TLCStateMut {
	first, second := &TLCStateMut{UID: 31, level: 1}, &TLCStateMut{UID: 41, level: 2}
	roots := []*TLCStateMut{first, second, nil}
	other := []*TLCStateMut{second, nil, first}
	first.values = []Value{
		&ModelValue{Data: roots}, &ModelValue{Data: other},
		&ModelValue{Data: append([]*TLCStateMut(nil), roots...)},
		&ModelValue{Data: []*TLCStateMut(nil)}, &ModelValue{Data: []*TLCStateMut{}},
		&ModelValue{Data: map[string]any{"roots": roots, "other": other}},
	}
	second.values = []Value{first.values[0]}
	return roots
}

func checkModelDataStateArrayGraph(states []*TLCStateMut, rootArray bool) error {
	first, second := states[0], states[1]
	values := first.values
	roots, ok := values[0].(*ModelValue).Data.([]*TLCStateMut)
	if !ok || len(roots) != 3 || roots[0] != first || roots[1] != second || roots[2] != nil || (rootArray && &roots[0] != &states[0]) || second.values[0] != values[0] {
		return fmt.Errorf("attached root array lost native type, sharing or state cycle")
	}
	other := values[1].(*ModelValue).Data.([]*TLCStateMut)
	separate := values[2].(*ModelValue).Data.([]*TLCStateMut)
	mixed := values[5].(*ModelValue).Data.(map[string]any)
	if len(other) != 3 || other[0] != second || other[1] != nil || other[2] != first || &other[0] != &mixed["other"].([]*TLCStateMut)[0] || &roots[0] != &mixed["roots"].([]*TLCStateMut)[0] {
		return fmt.Errorf("nested state arrays lost shared backing storage")
	}
	if &separate[0] == &roots[0] || separate[0] != first || separate[1] != second || separate[2] != nil {
		return fmt.Errorf("separate state array merged or lost element identity")
	}
	nilStates, ok := values[3].(*ModelValue).Data.([]*TLCStateMut)
	if !ok || nilStates != nil {
		return fmt.Errorf("nil state array lost native type")
	}
	empty, ok := values[4].(*ModelValue).Data.([]*TLCStateMut)
	if !ok || empty == nil || len(empty) != 0 {
		return fmt.Errorf("empty state array became null")
	}
	return nil
}

func TestDistributedModelDataStateArrays(t *testing.T) {
	states := modelDataStateArrayGraph()
	got := distributedPayloadRoundTrip(t, states)
	if err := checkModelDataStateArrayGraph(got, true); err != nil {
		t.Fatal(err)
	}
	first, second := got[0], got[1]
	roots := first.values[0].(*ModelValue).Data.([]*TLCStateMut)
	roots[0] = second
	if got[0] != second || first.values[2].(*ModelValue).Data.([]*TLCStateMut)[0] != first {
		t.Fatal("attached root replacement lost sharing or changed independent array")
	}
	if err := checkModelDataStateArrayGraph(states, true); err != nil {
		t.Fatalf("receiver replacement changed sender: %v", err)
	}
}

func TestWorkerRPCModelDataStateArrays(t *testing.T) {
	var fail atomic.Bool
	_, client := startWorkerRPC(t, &rpcTestWorker{next: func(states []*TLCStateMut) (*NextStateResult, error) {
		if err := checkModelDataStateArrayGraph(states, true); err != nil {
			return nil, err
		}
		if fail.Load() {
			return nil, NewWorkerException("attached state arrays", states[0], states[1], true)
		}
		return NewNextStateResult([]*StateVec{NewStateVecFrom(states)}, []*LongVec{NewLongVec()}, 1, 0), nil
	}})
	states := modelDataStateArrayGraph()
	result, err := client.GetNextStates(states)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkModelDataStateArrayGraph(result.NextStates[0].ToSlice(), false); err != nil {
		t.Fatal(err)
	}
	fail.Store(true)
	result, err = client.GetNextStates(states)
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || !failure.KeepCallStack {
		t.Fatalf("attached arrays lost failure context: %v/%v", result, err)
	}
	if err := checkModelDataStateArrayGraph([]*TLCStateMut{failure.State1, failure.State2}, false); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedModelDataStateArraysInvalid(t *testing.T) {
	for _, payload := range []*DistributedStatePayload{
		{RootsArray: -1}, {RootsArray: 1},
		{Nil: true, RootsArray: 1, StateArrays: [][]int{{}}},
		{Roots: []int{0}, RootsArray: 1, StateArrays: [][]int{{}}},
		{StateArrays: [][]int{{-1}}}, {StateArrays: [][]int{{1}}},
		{Values: []DistributedValueNode{{Kind: "model", DataKind: "stateArray", DataArray: -1}}},
		{Values: []DistributedValueNode{{Kind: "model", DataKind: "stateArray", DataArray: 1}}},
		{ObjectArrays: [][]DistributedObjectDataNode{{{Kind: "stateArray", Array: 1}}}},
	} {
		if _, err := DecodeDistributedStates(payload); err == nil {
			t.Fatal("invalid/conflicting state-array reference accepted")
		}
	}
	for _, attached := range [][]*TLCStateMut{{{level: -1}}, {{level: 1, functional: true}}} {
		if _, err := EncodeDistributedStates([]*TLCStateMut{{level: 1, values: []Value{&ModelValue{Data: attached}}}}); err == nil {
			t.Fatal("attached array bypassed state metadata validation")
		}
	}
}
