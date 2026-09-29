package tlc

import "testing"

func TestToolAssignsActionIDsWithInitialPredicatesFirst(t *testing.T) {
	initA := &Action{Name: "InitA", IsInitPred: true}
	initB := &Action{Name: "InitB", IsInitPred: true}
	nextA := &Action{Name: "NextA"}
	nextB := &Action{Name: "NextB"}

	tool := NewTool()
	tool.InitStateSpec = []*Action{initA, initB}
	tool.Actions = []*Action{nextA, nextB}
	tool.AssignActionIDs()

	for _, tc := range []struct {
		action *Action
		id     int
	}{
		{action: initA, id: 0},
		{action: initB, id: 1},
		{action: nextA, id: 2},
		{action: nextB, id: 3},
	} {
		if got := tc.action.GetID(); got != tc.id {
			t.Fatalf("%s id = %d, want %d", tc.action.GetName(), got, tc.id)
		}
	}
}

func TestToolSpecActionsPreserveInitThenNextOrder(t *testing.T) {
	init := &Action{Name: "Init"}
	nextA := &Action{Name: "NextA"}
	nextB := &Action{Name: "NextB"}
	tool := NewTool()
	tool.InitStateSpec = []*Action{init}
	tool.Actions = []*Action{nextA, nextB}

	specActions := tool.GetSpecActions()
	if len(specActions) != 3 {
		t.Fatalf("spec actions len = %d, want 3", len(specActions))
	}
	if specActions[0] != init || specActions[1] != nextA || specActions[2] != nextB {
		t.Fatalf("spec actions order = %#v, want init, nextA, nextB", specActions)
	}
	specActions[0] = nil
	if tool.GetInitStateSpec()[0] != init {
		t.Fatalf("GetSpecActions returned slice aliases internal storage")
	}
}

func TestToolConstraintAndSpecGettersReturnCopies(t *testing.T) {
	modelConstraint := "model constraint"
	actionConstraint := "action constraint"
	post := &Action{Name: "Post"}
	next := &Action{Name: "Next"}
	view := "view"

	tool := NewTool()
	tool.NextStateSpec = next
	tool.ViewSpec = view
	tool.ModelConstraints = []SemanticNode{modelConstraint}
	tool.ActionConstraints = []SemanticNode{actionConstraint}
	tool.PostConditionSpecs = []*Action{post}

	if tool.GetNextStateSpec() != next {
		t.Fatalf("next state spec mismatch")
	}
	if tool.GetViewSpec() != view {
		t.Fatalf("view spec mismatch")
	}
	if !tool.HasStateOrActionConstraints() {
		t.Fatalf("HasStateOrActionConstraints = false, want true")
	}

	modelConstraints := tool.GetModelConstraints()
	actionConstraints := tool.GetActionConstraints()
	postConditions := tool.GetPostConditionSpecs()
	modelConstraints[0] = nil
	actionConstraints[0] = nil
	postConditions[0] = nil
	if tool.GetModelConstraints()[0] != modelConstraint {
		t.Fatalf("GetModelConstraints returned slice aliases internal storage")
	}
	if tool.GetActionConstraints()[0] != actionConstraint {
		t.Fatalf("GetActionConstraints returned slice aliases internal storage")
	}
	if tool.GetPostConditionSpecs()[0] != post {
		t.Fatalf("GetPostConditionSpecs returned slice aliases internal storage")
	}
}
