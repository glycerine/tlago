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
