package tlc

import "testing"

func TestParseLivenessReturnsNilWithoutFairnessOrProperties(t *testing.T) {
	live, err := parseLiveness(&Tool{})
	if err != nil {
		t.Fatalf("parseLiveness error = %v", err)
	}
	if live != nil {
		t.Fatalf("parseLiveness = %v, want nil", live)
	}
}

func TestParseLivenessNegatesSingleProperty(t *testing.T) {
	prop := NewLNState("P", nil, EmptyContext, nil)
	tool := &Tool{ImpliedTemporals: []*Action{NewAction(prop, EmptyContext, "Prop")}}

	live, err := parseLiveness(tool)
	if err != nil {
		t.Fatalf("parseLiveness error = %v", err)
	}
	if live == nil || live.Kind != LiveExprNeg || live.Body != prop {
		t.Fatalf("parseLiveness = %#v, want negated property", live)
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

	live, err := parseLiveness(tool)
	if err != nil {
		t.Fatalf("parseLiveness error = %v", err)
	}
	if live == nil || live.Kind != LiveExprConj || live.Count() != 2 {
		t.Fatalf("parseLiveness = %#v, want fairness /\\ property-disjunction", live)
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
