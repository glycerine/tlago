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

func TestStateFunctorGetStatesReturnsSetOfStatesLikeJava(t *testing.T) {
	initTLCCheckerTest(t)
	state := checkerTestState(1)
	set := NewSetOfStates(1)
	set.Put(state)
	functor := &StateFunctor{GetStatesFunc: func() *SetOfStates { return set }}

	if got := functor.GetStates(); got != set {
		t.Fatalf("GetStates returned %p, want configured set %p", got, set)
	}
	if got := (*StateFunctor)(nil).GetStates(); got == nil || got.Size() != 0 {
		t.Fatalf("nil functor GetStates = %#v, want empty set", got)
	}
}

func TestToolSymmetryPermutationsAreConcreteAndCopied(t *testing.T) {
	ModelValueInit()
	a := AddModelValue("A")
	b := AddModelValue("B")
	swap := NewMVPerm()
	swap.Put(a, b)
	swap.Put(b, a)

	tool := NewTool()
	if tool.HasSymmetry() {
		t.Fatalf("new tool unexpectedly has symmetry")
	}
	tool.SetSymmetryPermutations([]*MVPerm{swap})
	if !tool.HasSymmetry() {
		t.Fatalf("tool with concrete permutations reports no symmetry")
	}

	perms := tool.GetSymmetryPerms()
	if len(perms) != 1 || perms[0] != swap {
		t.Fatalf("GetSymmetryPerms = %#v, want stored swap", perms)
	}
	perms[0] = nil
	if tool.GetSymmetryPerms()[0] != swap {
		t.Fatalf("GetSymmetryPerms returned slice aliases internal storage")
	}
}

func TestToolHasSymmetryUsesModelConfigNameLikeJava(t *testing.T) {
	cfg, err := ParseModelConfigSource("MC.cfg", "SYMMETRY Symmetry")
	if err != nil {
		t.Fatalf("ParseModelConfigSource returned error: %v", err)
	}
	tool := NewToolWithModelConfig(cfg)
	if !tool.HasSymmetry() {
		t.Fatalf("tool with config SYMMETRY reports no symmetry")
	}
}

func TestSetTLCStateToolInstallsToolSymmetryPermutations(t *testing.T) {
	UniqueStringInitialize()
	ModelValueInit()
	SetStateVariables([]string{"x", "y"})
	x := UniqueStringOf("x")
	y := UniqueStringOf("y")
	a := AddModelValue("A")
	b := AddModelValue("B")

	swap := NewMVPerm()
	swap.Put(a, b)
	swap.Put(b, a)
	tool := NewTool()
	tool.SetSymmetryPermutations([]*MVPerm{swap})
	SetTLCStateTool(tool)
	t.Cleanup(func() { SetTLCStateTool(nil) })

	representative := NewEmptyState().Bind(x, a).Bind(y, b)
	symmetric := NewEmptyState().Bind(x, b).Bind(y, a)
	if representative.FingerPrint() != symmetric.FingerPrint() {
		t.Fatalf("tool-installed symmetry did not affect state fingerprints")
	}
}
