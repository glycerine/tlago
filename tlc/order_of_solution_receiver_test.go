package tlc

import (
	"io"
	"testing"
)

func TestOrderOfSolutionNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var solution *OrderOfSolution
	for name, call := range map[string]func(){
		"has_tableau":       func() { _ = solution.HasTableau() },
		"tableau":           func() { _ = solution.GetTableau() },
		"promises":          func() { _ = solution.GetPromises() },
		"check_state":       func() { _ = solution.GetCheckState() },
		"check_action":      func() { _ = solution.GetCheckAction() },
		"pems":              func() { _ = solution.GetPems() },
		"print_promises":    func() { solution.PrintPromises(io.Discard) },
		"set_check_state":   func() { solution.SetCheckState(nil) },
		"set_check_action":  func() { solution.SetCheckAction(nil) },
		"set_pems":          func() { solution.SetPEMs(nil) },
		"check_state_value": func() { _, _ = solution.CheckState(nil, nil) },
		"check_action_value": func() {
			_, _ = solution.CheckAction(nil, nil, nil)
		},
		"check_action_bits": func() {
			_, _ = solution.CheckActionBitVector(nil, nil, nil, nil, 0)
		},
		"has_empty_pem": func() { _ = solution.HasEmptyPEM() },
		"pem_stuttering": func() {
			_ = solution.IsPEMSatisfiedByStuttering(nil, nil)
		},
		"empty_pem_box_free": func() { _ = solution.HasEmptyPEMAndBoxFreePromises() },
		"string":             func() { _ = solution.String() },
	} {
		assertOrderOfSolutionPanic(t, name, call)
	}
}

func TestOrderOfSolutionNullArraysFailAtJavaUseBoundaries(t *testing.T) {
	assertOrderOfSolutionPanic(t, "constructor_array", func() {
		_ = NewOrderOfSolution(nil, nil)
	})
	assertOrderOfSolutionPanic(t, "constructor_element", func() {
		_ = NewOrderOfSolution(nil, []*LiveExprNode{nil})
	})

	solution := &OrderOfSolution{}
	for name, call := range map[string]func(){
		"print_promises":    func() { solution.PrintPromises(io.Discard) },
		"check_state":       func() { _, _ = solution.CheckState(nil, nil) },
		"check_action":      func() { _, _ = solution.CheckAction(nil, nil, nil) },
		"check_action_bits": func() { _, _ = solution.CheckActionBitVector(nil, nil, nil, nil, 0) },
		"has_empty_pem":     func() { _ = solution.HasEmptyPEM() },
		"pem_stuttering":    func() { _ = solution.IsPEMSatisfiedByStuttering(nil, nil) },
		"empty_pem_box_free": func() {
			_ = solution.HasEmptyPEMAndBoxFreePromises()
		},
		"string": func() { _ = solution.String() },
	} {
		assertOrderOfSolutionPanic(t, name, call)
	}

	solution.Promises = make([]*LiveExprNode, 0)
	solution.checkStateExprs = make([]*LiveExprNode, 0)
	solution.checkActionExprs = make([]*LiveExprNode, 0)
	solution.PEMs = make([]*PossibleErrorModel, 0)
	if got := solution.String(); got != "" {
		t.Fatalf("empty solution string = %q, want empty", got)
	}
	if _, err := solution.CheckState(nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.CheckAction(nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := solution.CheckActionBitVector(nil, nil, nil, nil, 0); err != nil {
		t.Fatal(err)
	}
}

func TestOrderOfSolutionNullCheckExpressionFailsLikeJava(t *testing.T) {
	stateSolution := &OrderOfSolution{checkStateExprs: []*LiveExprNode{nil}}
	assertOrderOfSolutionPanic(t, "state", func() {
		_, _ = stateSolution.CheckState(nil, nil)
	})
	actionSolution := &OrderOfSolution{checkActionExprs: []*LiveExprNode{nil}}
	assertOrderOfSolutionPanic(t, "action_values", func() {
		_, _ = actionSolution.CheckAction(nil, nil, nil)
	})
	assertOrderOfSolutionPanic(t, "action_bits", func() {
		_, _ = actionSolution.CheckActionBitVector(nil, nil, nil, NewBitVector(1), 0)
	})
}

func assertOrderOfSolutionPanic(t *testing.T, name string, call func()) {
	t.Helper()
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
