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
