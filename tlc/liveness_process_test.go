package tlc

import "testing"

func TestLiveExprNonBooleanPredicateUsesJavaErrorCode(t *testing.T) {
	expr := NewLNAction("bad", NewValueNode(NewIntValue(1)), EmptyContext, nil)
	_, err := expr.Eval(&Tool{}, EmptyState, EmptyState)
	if err == nil {
		t.Fatalf("Eval returned nil error for non-boolean liveness predicate")
	}
	tlcErr, ok := err.(*TLCError)
	if !ok {
		t.Fatalf("Eval error = %T, want TLCError", err)
	}
	if tlcErr.Code != ECTLCLiveEncounteredNonboolPredicate {
		t.Fatalf("Eval error code = %d, want %d", tlcErr.Code, ECTLCLiveEncounteredNonboolPredicate)
	}
	if got, want := tlcErr.Error(), "Encountered an action predicate that's not a boolean."; got != want {
		t.Fatalf("Eval error = %q, want Java message %q", got, want)
	}
}

func TestLiveExprStateNonBooleanPredicateUsesJavaErrorCode(t *testing.T) {
	expr := NewLNState("bad-state", NewValueNode(NewIntValue(1)), EmptyContext, nil)
	_, err := expr.Eval(&Tool{}, EmptyState, EmptyState)
	if err == nil {
		t.Fatalf("Eval returned nil error for non-boolean state predicate")
	}
	tlcErr, ok := err.(*TLCError)
	if !ok {
		t.Fatalf("Eval error = %T, want TLCError", err)
	}
	if tlcErr.Code != ECTLCLiveStatePredicateNonBool {
		t.Fatalf("Eval error code = %d, want %d", tlcErr.Code, ECTLCLiveStatePredicateNonBool)
	}
	if got, want := tlcErr.Error(), "A state predicate was evaluated to a non-boolean value."; got != want {
		t.Fatalf("Eval error = %q, want Java message %q", got, want)
	}
}

func TestLiveExprTemporalDirectEvalUsesJavaErrorCode(t *testing.T) {
	for _, tc := range []struct {
		name string
		expr *LiveExprNode
		want string
	}{
		{name: "always", expr: NewLNAll(LNTrue), want: "Can not evaluate a temporal formula []F."},
		{name: "eventually", expr: NewLNEven(LNTrue), want: "Can not evaluate a temporal formula <>F."},
	} {
		_, err := tc.expr.Eval(&Tool{}, EmptyState, EmptyState)
		if err == nil {
			t.Fatalf("%s Eval returned nil error", tc.name)
		}
		tlcErr, ok := err.(*TLCError)
		if !ok {
			t.Fatalf("%s Eval error = %T, want TLCError", tc.name, err)
		}
		if tlcErr.Code != ECTLCLiveCannotEvalFormula {
			t.Fatalf("%s Eval error code = %d, want %d", tc.name, tlcErr.Code, ECTLCLiveCannotEvalFormula)
		}
		if got := tlcErr.Error(); got != tc.want {
			t.Fatalf("%s Eval error = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestTBParPositiveClosureActionUsesJavaErrorCode(t *testing.T) {
	defer func() {
		recovered := recover()
		tlcErr, ok := recovered.(*TLCError)
		if !ok {
			t.Fatalf("panic = %T %v, want TLCError", recovered, recovered)
		}
		if tlcErr.Code != ECTLCLiveEncounteredActions {
			t.Fatalf("panic code = %d, want %d", tlcErr.Code, ECTLCLiveEncounteredActions)
		}
		if got, want := tlcErr.Error(), "TLC encountered actions when computing closure."; got != want {
			t.Fatalf("panic message = %q, want %q", got, want)
		}
	}()
	par := NewTBPar(1)
	par.AddElement(NewLNAction("A", nil, EmptyContext, nil))
	_ = par.PositiveClosure()
}

func TestParseLivenessReturnsNilWithoutFairnessOrProperties(t *testing.T) {
	live, err := ParseLiveness(&Tool{})
	if err != nil {
		t.Fatalf("ParseLiveness error = %v", err)
	}
	if live != nil {
		t.Fatalf("ParseLiveness = %v, want nil", live)
	}
}

func TestParseLivenessNegatesSingleProperty(t *testing.T) {
	prop := NewLNState("P", nil, EmptyContext, nil)
	tool := &Tool{ImpliedTemporals: []*Action{NewAction(prop, EmptyContext, "Prop")}}

	live, err := ParseLiveness(tool)
	if err != nil {
		t.Fatalf("ParseLiveness error = %v", err)
	}
	if live == nil || live.Kind != LiveExprNeg || live.Body != prop {
		t.Fatalf("ParseLiveness = %#v, want negated property", live)
	}
	if tool.LivenessIsTrue() {
		t.Fatalf("Tool.LivenessIsTrue = true, want false with implied temporal property")
	}
}

func TestParseLivenessCombinesFairnessAndMultiplePropertiesLikeJava(t *testing.T) {
	fair := NewLNAll(NewLNEven(NewLNAction("Fair", nil, EmptyContext, nil)))
	propA := NewLNState("A", nil, EmptyContext, nil)
	propB := NewLNState("B", nil, EmptyContext, nil)
	tool := &Tool{
		Temporals: []*Action{
			NewAction(fair, EmptyContext, "Fairness"),
		},
		ImpliedTemporals: []*Action{
			NewAction(propA, EmptyContext, "PropA"),
			NewAction(propB, EmptyContext, "PropB"),
		},
	}

	live, err := ParseLiveness(tool)
	if err != nil {
		t.Fatalf("ParseLiveness error = %v", err)
	}
	if live == nil || live.Kind != LiveExprConj || live.Count() != 2 {
		t.Fatalf("ParseLiveness = %#v, want fairness /\\ property-disjunction", live)
	}
	if live.GetBody(0) != fair {
		t.Fatalf("first conjunct = %#v, want fairness", live.GetBody(0))
	}
	disj := live.GetBody(1)
	if disj == nil || disj.Kind != LiveExprDisj || disj.Count() != 2 {
		t.Fatalf("second conjunct = %#v, want disjunction of negated properties", disj)
	}
	for i, want := range []*LiveExprNode{propA, propB} {
		got := disj.GetBody(i)
		if got == nil || got.Kind != LiveExprNeg || got.Body != want {
			t.Fatalf("disjunct %d = %#v, want negated property %#v", i, got, want)
		}
	}
}

func TestProcessLivenessClassifiesAndBinsChecks(t *testing.T) {
	action := NewLNAction("A", nil, EmptyContext, nil)
	state := NewLNState("S", nil, EmptyContext, nil)
	tf := NewLNAll(NewLNState("T", nil, EmptyContext, nil))
	live := NewLNConj(
		NewLNEven(NewLNAll(action)),
		NewLNAll(NewLNEven(state)),
		tf,
	)
	tool := &Tool{ImpliedTemporals: []*Action{NewAction(NewLNNeg(live), EmptyContext, "Prop")}}

	solutions, err := ProcessLivenessSilent(tool, true)
	if err != nil {
		t.Fatalf("ProcessLivenessSilent error = %v", err)
	}
	if len(solutions) != 1 {
		t.Fatalf("len(solutions) = %d, want 1", len(solutions))
	}
	solution := solutions[0]
	if solution == nil || len(solution.PEMs) != 1 {
		t.Fatalf("solution PEMs = %#v, want one PEM", solution)
	}
	if len(solution.CheckAction) != 1 || solution.CheckAction[0] != action {
		t.Fatalf("CheckAction = %#v, want action bin containing A", solution.CheckAction)
	}
	if len(solution.CheckState) != 1 || solution.CheckState[0] != state {
		t.Fatalf("CheckState = %#v, want state bin containing S", solution.CheckState)
	}
	pem := solution.PEMs[0]
	if len(pem.EAAction) != 1 || pem.EAAction[0] != 0 {
		t.Fatalf("PEM.EAAction = %#v, want [0]", pem.EAAction)
	}
	if len(pem.AEState) != 1 || pem.AEState[0] != 0 {
		t.Fatalf("PEM.AEState = %#v, want [0]", pem.AEState)
	}
	if !solution.HasTableau() {
		t.Fatalf("solution.HasTableau = false, want true for remaining temporal formula")
	}
}
