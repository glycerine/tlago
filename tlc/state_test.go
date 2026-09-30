package tlc

import "testing"

func TestTLCStateUnassignedVariablesAreLexicographic(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"z", "a", "m"})
	state := NewEmptyState().Bind(UniqueStringOf("m"), NewIntValue(1))

	unassigned := state.Unassigned()
	if len(unassigned) != 2 {
		t.Fatalf("unassigned len = %d, want 2", len(unassigned))
	}
	if unassigned[0].Name.String() != "a" || unassigned[1].Name.String() != "z" {
		t.Fatalf("unassigned order = %q, %q; want a, z", unassigned[0].Name, unassigned[1].Name)
	}
}

func TestTLCStateCopyPreservesLevelButNotWorkerOrUID(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"x"})
	pred := checkerTestState(0)
	state := checkerTestState(1).SetPredecessor(pred)
	state.WorkerID = 7
	state.UID = 42

	shallow := state.Copy()
	if shallow.Level() != state.Level() {
		t.Fatalf("copy level = %d, want %d", shallow.Level(), state.Level())
	}
	if shallow.WorkerID != TLCStateInitWorkerID {
		t.Fatalf("copy worker id = %d, want default", shallow.WorkerID)
	}
	if shallow.UID != TLCStateInitUID {
		t.Fatalf("copy uid = %d, want %d", shallow.UID, TLCStateInitUID)
	}

	deep := state.DeepCopy()
	if deep.Level() != state.Level() || deep.WorkerID != state.WorkerID || deep.UID != state.UID {
		t.Fatalf("deep copy level/worker/uid = %d/%d/%d, want %d/%d/%d",
			deep.Level(), deep.WorkerID, deep.UID, state.Level(), state.WorkerID, state.UID)
	}
}

func TestTLCStateAddToVecAddsShallowCopy(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"x"})
	state := checkerTestState(1)
	state.WorkerID = 3
	state.UID = 99
	vec := NewStateVec(0)

	result := state.AddToVec(vec)
	if result != vec {
		t.Fatalf("AddToVec returned %p, want vec %p", result, vec)
	}
	if vec.Size() != 1 {
		t.Fatalf("vec size = %d, want 1", vec.Size())
	}
	added := vec.At(0)
	if added == state {
		t.Fatalf("AddToVec added original state, want shallow copy")
	}
	if !added.Equal(state) {
		t.Fatalf("added state values differ from original")
	}
	if added.WorkerID != TLCStateInitWorkerID || added.UID != TLCStateInitUID {
		t.Fatalf("added worker/uid = %d/%d, want default/%d", added.WorkerID, added.UID, TLCStateInitUID)
	}
}

func TestTLCStateSubsetUsesPartialStateSemantics(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"x", "y"})
	x := UniqueStringOf("x")
	y := UniqueStringOf("y")

	universe := NewEmptyState()
	xOne := NewEmptyState().Bind(x, NewIntValue(1))
	xOneYTwo := NewEmptyState().Bind(x, NewIntValue(1)).Bind(y, NewIntValue(2))
	xTwo := NewEmptyState().Bind(x, NewIntValue(2))

	if got := IsStateSubset(xOneYTwo, xOne); got != PartialYes {
		t.Fatalf("x=1,y=2 subset x=1 = %s, want YES", got)
	}
	if got := IsStateSubset(xOne, xOneYTwo); got != PartialNo {
		t.Fatalf("x=1 subset x=1,y=2 = %s, want NO", got)
	}
	if got := IsStateSubset(xOne, universe); got != PartialYes {
		t.Fatalf("x=1 subset universe = %s, want YES", got)
	}
	if got := IsStateSubset(nil, xOne); got != PartialNo {
		t.Fatalf("nil/universe subset x=1 = %s, want NO", got)
	}
	if got := IsStateSubset(xTwo, xOne); got != PartialNo {
		t.Fatalf("x=2 subset x=1 = %s, want NO", got)
	}
}

func TestTLCStateSubsetReturnsMaybeForIncomparableValues(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"x"})
	x := UniqueStringOf("x")

	intState := NewEmptyState().Bind(x, NewIntValue(1))
	boolState := NewEmptyState().Bind(x, BoolTrue)
	if got := IsStateSubset(intState, boolState); got != PartialMaybe {
		t.Fatalf("int subset bool = %s, want MAYBE", got)
	}
}
