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
	if tool.GetInitStateSpec().ElementAt(0) != init {
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
	if got := (&StateFunctor{}).GetStates(); got == nil || got.Size() != 0 {
		t.Fatalf("default functor GetStates = %#v, want empty set", got)
	}
}

func TestDisabledActionCompositionRetainsJavaFailureContext(t *testing.T) {
	oldCdot := Globals.Cdot
	Globals.Cdot = false
	t.Cleanup(func() { Globals.Cdot = oldCdot })
	t.Setenv(actionCompositionProperty, "false")

	expr := NewBuiltinOpApplNode(OpCdot, NewValueNode(BoolTrue), NewValueNode(BoolTrue))
	context := EmptyContext.Cons(NewSymbolNode("context"), NewIntValue(1))
	tool := NewTool()
	state0, state1 := NewEmptyState(), NewEmptyState()
	requireFailure := func(t *testing.T, err error) {
		t.Helper()
		failure, ok := err.(*TLCError)
		if !ok || !failure.Runtime || failure.Code != ECGeneral || failure.Error() != actionCompositionUnsupportedMessage {
			t.Fatalf("failure = %T/%v, want detailed Java runtime failure", err, err)
		}
		if failure.Expr != expr || failure.Ctxt != context {
			t.Fatalf("failure expression/context = %p/%p, want %p/%p", failure.Expr, failure.Ctxt, expr, context)
		}
	}

	t.Run("eval", func(t *testing.T) {
		_, err := tool.Eval(expr, context, state0, state1, EvalClear, DoNotRecordCostModel)
		requireFailure(t, err)
	})
	t.Run("enabled", func(t *testing.T) {
		_, err := tool.EnabledImpl(expr, EmptyActionItemList, context, state0, state1, DoNotRecordCostModel)
		requireFailure(t, err)
	})
	t.Run("next", func(t *testing.T) {
		_, err := tool.GetNextStatesForPredicate(UnknownAction, expr, EmptyActionItemList, context, state0, state1, NewNextStateFunctor(nil), DoNotRecordCostModel)
		requireFailure(t, err)
	})
}

func TestMakeStateRetainsJavaFailureContext(t *testing.T) {
	tool := NewTool()
	requireFailure := func(t *testing.T, pred SemanticNode, prefix string) {
		t.Helper()
		_, err := tool.MakeStateImpl(pred)
		failure, ok := err.(*TLCError)
		if !ok || !failure.Runtime || failure.Code != ECGeneral || failure.Error() != prefix+SemanticString(pred) {
			t.Fatalf("failure = %T/%v, want detailed Java runtime failure", err, err)
		}
		if failure.Expr != pred || failure.Ctxt != EmptyContext {
			t.Fatalf("failure expression/context = %p/%p, want %p/%p", failure.Expr, failure.Ctxt, pred, EmptyContext)
		}
	}

	SetStateVariables(nil)
	t.Cleanup(func() { SetStateVariables(nil) })
	requireFailure(t, NewBuiltinOpApplNode(OpEq, NewIntValue(1), NewIntValue(2)), "The predicate does not specify a unique state.")

	SetStateVariables([]string{"x"})
	requireFailure(t, NewBuiltinOpApplNode(OpEq, NewIntValue(1), NewIntValue(1)), "The state specified by the predicate is not complete.")
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

func TestResetExtendedTLCStateToolReturnsPreviousToolOnly(t *testing.T) {
	oldStateTool := stateTool
	oldPolicy := statePreserveMetadata
	oldRuntimes := mutableStateRuntimes
	t.Cleanup(func() {
		stateTool = oldStateTool
		statePreserveMetadata = oldPolicy
		mutableStateRuntimes = oldRuntimes
	})

	first := NewTool().SetMode(ModeSimulation)
	first.ViewSpec = NewValueNode(NewIntValue(1))
	first.SetSymmetryPermutations([]*MVPerm{{}})
	SetTLCStateTool(first)
	runtime := &mutableStateRuntimes[1]
	view := runtime.view
	permutations := runtime.permutations
	second := NewTool().SetMode(ModeSimulation)
	if old := ResetExtendedTLCStateTool(second); old != first {
		t.Fatalf("reset old tool = %p, want %p", old, first)
	}
	if runtime.tool != second || stateTool != first || runtime.view != view || len(runtime.permutations) != 1 || runtime.permutations[0] != permutations[0] {
		t.Fatal("reset changed captured view/symmetry or failed to install the tool")
	}
	if old := ResetExtendedTLCStateTool(first); old != second || runtime.tool != first || stateTool != first {
		t.Fatal("reset did not return and restore the temporary tool")
	}
}

func TestToolLookupDoesNotReadStateForUserDefinedOperatorNameClash(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"Clash"})
	tool := NewTool()
	opSym := NewSymbolNode("Clash")
	opDef := NewOpDefNodeForSymbol(opSym, nil, NewValueNode(NewIntValue(2)))
	tool.Define(opSym, opDef)

	state := NewEmptyState()
	state.Bind(UniqueStringOf("Clash"), IntOne)

	if got := tool.Lookup(opSym, EmptyContext, state, false); got != opDef {
		t.Fatalf("operator lookup = %T %[1]v, want OpDefNode despite state variable with same name", got)
	}

	varSym := NewVariableSymbolNode("Clash")
	if got := tool.Lookup(varSym, EmptyContext, state, false); got != IntOne {
		t.Fatalf("variable lookup = %v, want state value", got)
	}
}
