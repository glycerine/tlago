package tlc

import "testing"

func TestTLCStateEmptyIdentityAndStringLookup(t *testing.T) {
	EmptyState = nil
	if !IsEmptyTLCState(nil) {
		t.Fatal("null state was not empty while TLCState.Empty was null")
	}
	var typedNil *TLCStateMut
	if !IsEmptyTLCState(typedNil) {
		t.Fatal("typed null state was not empty while TLCState.Empty was null")
	}

	UniqueStringInitialize()
	SetStateVariables([]string{"x"})
	empty := EmptyState
	state := NewEmptyState().Bind(UniqueStringOf("x"), NewIntValue(7))
	if !IsEmptyTLCState(empty) || IsEmptyTLCState(state) || IsEmptyTLCState(nil) || IsEmptyTLCState(typedNil) {
		t.Fatal("empty-state check did not preserve source reference identity")
	}
	if value := state.LookupName("x"); value.(*IntValue).Val != 7 {
		t.Fatalf("string lookup = %v, want 7", value)
	}

	var nilState *TLCStateMut
	defer func() {
		if _, ok := recover().(*NullPointerException); !ok {
			t.Fatal("null string-lookup receiver did not raise NullPointerException")
		}
	}()
	_ = nilState.LookupName("x")
}

func TestTLCStateVariableNamesReturnFreshSourceArray(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"x", "y"})
	state := NewEmptyState()
	names := state.GetVarsAsStrings()
	if len(names) != 2 || names[0] != "x" || names[1] != "y" {
		t.Fatalf("variable names = %#v, want [x y]", names)
	}
	names[0] = "changed"
	if next := state.GetVarsAsStrings(); next[0] != "x" {
		t.Fatalf("variable-name result retained caller mutation: %#v", next)
	}

	var nilState *TLCStateMut
	defer func() {
		if _, ok := recover().(*NullPointerException); !ok {
			t.Fatal("null variable-name receiver did not raise NullPointerException")
		}
	}()
	_ = nilState.GetVarsAsStrings()
}

func TestTLCStateVariableDeclarationsRetainSourceArray(t *testing.T) {
	UniqueStringInitialize()
	x, y, z := NewVariableSymbolNode("x"), NewVariableSymbolNode("y"), NewVariableSymbolNode("z")
	declarations := []*SymbolNode{x, y}
	SetStateVariableDeclarations(declarations)
	state := NewEmptyState()
	if vars := state.GetVars(); len(vars) != 2 || vars[0] != x || vars[1] != y {
		t.Fatalf("variable declarations = %#v, want supplied array", vars)
	}
	state.GetVars()[0] = z
	if declarations[0] != z {
		t.Fatal("getVars copied the retained declaration array")
	}
	if names := state.GetVarsAsStrings(); len(names) != 2 || names[0] != "z" || names[1] != "y" {
		t.Fatalf("names after declaration mutation = %#v, want [z y]", names)
	}

	var nilState *TLCStateMut
	defer func() {
		if _, ok := recover().(*NullPointerException); !ok {
			t.Fatal("null declaration receiver did not raise NullPointerException")
		}
	}()
	_ = nilState.GetVars()
}

func TestTLCStatePredecessorInfoUsesOriginalState(t *testing.T) {
	initTLCCheckerTest(t)
	original := &TLCStateMut{level: 4, extended: true}
	alias := &TLCStateMut{level: 40, extended: true}
	info := AliasTLCStateInfo(alias, NewTLCStateInfo(original))
	target := &TLCStateMut{extended: true}
	if target.SetPredecessorInfo(info) != target || target.GetPredecessor() != original || target.GetLevel() != 5 {
		t.Fatalf("predecessor info selected %p at level %d, want original %p at level 5", target.Predecessor(), target.Level(), original)
	}

	for name, call := range map[string]func(){
		"null_info": func() { (&TLCStateMut{}).SetPredecessorInfo(nil) },
		"null_receiver": func() {
			var state *TLCStateMut
			state.SetPredecessorInfo(info)
		},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if _, ok := recover().(*NullPointerException); !ok {
					t.Fatal("failure is not NullPointerException")
				}
			}()
			call()
		})
	}
}

func TestTLCStateMetadataNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var state *TLCStateMut
	for name, call := range map[string]func(){
		"get_cached":        func() { _ = state.GetCached(0) },
		"set_cached":        func() { _ = state.SetCached(0, BoolTrue) },
		"set_callable":      func() { state.SetCallable(nil) },
		"exec_callable":     func() { _, _ = state.ExecCallable() },
		"unset_predecessor": func() { _ = state.UnsetPredecessor() },
		"predecessor":       func() { _ = state.Predecessor() },
		"get_predecessor":   func() { _ = state.GetPredecessor() },
		"trace_predecessor": func() { _ = state.TracePredecessor() },
		"level":             func() { _ = state.Level() },
		"get_level":         func() { _ = state.GetLevel() },
		"initial":           func() { _ = state.IsInitial() },
		"set_action":        func() { _ = state.SetAction(nil) },
		"has_action":        func() { _ = state.HasAction() },
		"get_action":        func() { _ = state.GetAction() },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
			}()
			call()
		})
	}
}

func TestTLCStateCoreNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var state *TLCStateMut
	for name, call := range map[string]func(){
		"create_empty":   func() { _ = state.CreateEmpty() },
		"copy":           func() { _ = state.Copy() },
		"deep_copy":      func() { _ = state.DeepCopy() },
		"add_to_vec":     func() { _ = state.AddToVec(nil) },
		"deep_normalize": func() { state.DeepNormalize() },
		"all_assigned":   func() { _ = state.AllAssigned() },
		"none_assigned":  func() { _ = state.NoneAssigned() },
		"unassigned":     func() { _ = state.Unassigned() },
		"get_unassigned": func() { _ = state.GetUnassigned() },
		"equal":          func() { _ = state.Equal(nil) },
		"hash_code":      func() { _ = state.HashCode() },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
			}()
			call()
		})
	}

	valid := &TLCStateMut{}
	if valid.Equal(nil) {
		t.Fatal("valid state compared equal to null")
	}
	var typedNil *TLCStateMut
	if valid.Equal(typedNil) {
		t.Fatal("valid state compared equal to typed null")
	}
}

// Print wrappers and functional states inherit the source base metadata methods,
// even when the process's mutable state factory selects TLCStateMutExt.
func TestBaseStateMetadataInExtendedMode(t *testing.T) {
	initTLCCheckerTest(t)
	oldPolicy := statePreserveMetadata
	functionalEmpty := NewFunctionalState()
	oldFunctionalEmpty := *functionalEmpty
	statePreserveMetadata = true
	t.Cleanup(func() {
		statePreserveMetadata = oldPolicy
		*functionalEmpty = oldFunctionalEmpty
	})
	for _, kind := range []string{"print", "functional"} {
		t.Run(kind, func(t *testing.T) {
			state := NewFunctionalState()
			if kind == "print" {
				state = EmptyRecord.ToState()
			}
			calls := 0
			state.SetAction(&Action{Name: "Ignored"})
			state.SetCallable(func() (any, error) { calls++; return BoolTrue, nil })
			result, err := state.ExecCallable()
			if err != nil || result != nil || calls != 0 || state.HasAction() || state.callable != nil {
				t.Fatal("base state retained extended action or callable behavior")
			}
			parent := &TLCStateMut{level: 5}
			state.SetPredecessor(parent)
			if state.Level() != 6 || state.TracePredecessor() != nil {
				t.Fatal("base predecessor setter did not update only the level")
			}
			if kind == "print" {
				if _, err := EncodeDistributedStates([]*TLCStateMut{state}); err != nil {
					t.Fatalf("ignored evaluator metadata prevented print-state transfer: %v", err)
				}
			}
		})
	}
}

// No upstream method directly tests the base/extended state cache boundary.
func TestStateCacheRespectsSourceStateKind(t *testing.T) {
	oldPolicy := statePreserveMetadata
	t.Cleanup(func() { statePreserveMetadata = oldPolicy })
	for _, kind := range []string{"ordinary", "extended", "print", "functional"} {
		t.Run(kind, func(t *testing.T) {
			statePreserveMetadata = kind != "ordinary"
			state := &TLCStateMut{level: TLCStateInitLevel, extended: kind == "extended"}
			if kind == "print" {
				state.printRecord = EmptyRecord
			}
			state.functional = kind == "functional"
			value := NewIntValue(17)
			stored := state.SetCached(3, value)
			if kind == "extended" {
				if stored != value || state.GetCached(3) != value {
					t.Fatal("extended state did not cache value")
				}
				if state.SetCached(3, nil) != nil || state.GetCached(3) != nil {
					t.Fatal("null cache replacement did not retain source behavior")
				}
			} else if stored != nil || state.GetCached(3) != nil || state.cached != nil {
				t.Fatal("base state allocated or returned a cache")
			}
		})
	}
}

func TestTLCStateUnassignedVariablesAreLexicographic(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"z", "a", "m"})
	state := NewEmptyState().Bind(UniqueStringOf("m"), NewIntValue(1))

	unassigned := state.GetUnassigned()
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
	SetTLCStateTool(nil)
	pred := checkerTestState(0)
	action := NewAction(nil, EmptyContext, "Next")
	state := checkerTestState(1).SetPredecessor(pred).SetAction(action)
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
	if shallow.Predecessor() != nil || shallow.GetAction() != nil {
		t.Fatalf("ordinary copy preserved predecessor/action; Java TLCStateMut.copy drops both")
	}

	deep := state.DeepCopy()
	if deep.Level() != state.Level() || deep.WorkerID != state.WorkerID || deep.UID != state.UID {
		t.Fatalf("deep copy level/worker/uid = %d/%d/%d, want %d/%d/%d",
			deep.Level(), deep.WorkerID, deep.UID, state.Level(), state.WorkerID, state.UID)
	}
	if deep.Predecessor() != nil || deep.GetAction() != nil {
		t.Fatalf("ordinary deep copy preserved predecessor/action; Java TLCStateMut.deepCopy drops both")
	}
}

func TestTLCStateCopyPreservesMetadataInExtendedMode(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"x"})
	tool := NewTool().SetMode(ModeSimulation)
	SetTLCStateTool(tool)
	t.Cleanup(func() { SetTLCStateTool(nil) })
	pred := checkerTestState(0)
	action := NewAction(nil, EmptyContext, "Next")
	state := checkerTestState(1).SetPredecessor(pred).SetAction(action)

	shallow := state.Copy()
	if shallow.Predecessor() != pred || shallow.GetAction() != action {
		t.Fatalf("extended copy predecessor/action = %p/%p, want %p/%p",
			shallow.Predecessor(), shallow.GetAction(), pred, action)
	}
	deep := state.DeepCopy()
	if deep.Predecessor() != pred || deep.GetAction() != action {
		t.Fatalf("extended deep copy predecessor/action = %p/%p, want %p/%p",
			deep.Predecessor(), deep.GetAction(), pred, action)
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

	if got := TLCStateIsSubset(xOneYTwo, xOne); got != PartialYes {
		t.Fatalf("x=1,y=2 subset x=1 = %s, want YES", got)
	}
	if got := TLCStateIsSubset(xOne, xOneYTwo); got != PartialNo {
		t.Fatalf("x=1 subset x=1,y=2 = %s, want NO", got)
	}
	if got := TLCStateIsSubset(xOne, universe); got != PartialYes {
		t.Fatalf("x=1 subset universe = %s, want YES", got)
	}
	if got := TLCStateIsSubset(nil, xOne); got != PartialNo {
		t.Fatalf("nil/universe subset x=1 = %s, want NO", got)
	}
	if got := TLCStateIsSubset(xTwo, xOne); got != PartialNo {
		t.Fatalf("x=2 subset x=1 = %s, want NO", got)
	}
}

func TestTLCStateSubsetReturnsMaybeForIncomparableValues(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"x"})
	x := UniqueStringOf("x")

	intState := NewEmptyState().Bind(x, NewIntValue(1))
	boolState := NewEmptyState().Bind(x, BoolTrue)
	if got := TLCStateIsSubset(intState, boolState); got != PartialMaybe {
		t.Fatalf("int subset bool = %s, want MAYBE", got)
	}
}

func TestTLCStateFingerprintUsesSymmetryRepresentative(t *testing.T) {
	UniqueStringInitialize()
	ModelValueInit()
	SetStateVariables([]string{"x", "y"})
	x := UniqueStringOf("x")
	y := UniqueStringOf("y")
	a := AddModelValue("A")
	b := AddModelValue("B")

	representative := NewEmptyState().Bind(x, a).Bind(y, b)
	symmetric := NewEmptyState().Bind(x, b).Bind(y, a)
	if representative.FingerPrint() == symmetric.FingerPrint() {
		t.Fatalf("fingerprints unexpectedly match before symmetry is configured")
	}

	swap := NewMVPerm()
	swap.Put(a, b)
	swap.Put(b, a)
	SetStateSymmetryPermutations([]*MVPerm{swap})
	t.Cleanup(func() { SetStateSymmetryPermutations(nil) })

	if representative.FingerPrint() != symmetric.FingerPrint() {
		t.Fatalf("symmetric states have different fingerprints")
	}
	if symmetric.Lookup(x) != b || symmetric.Lookup(y) != a {
		t.Fatalf("fingerprinting mutated symmetric state to x=%v y=%v", symmetric.Lookup(x), symmetric.Lookup(y))
	}
}
