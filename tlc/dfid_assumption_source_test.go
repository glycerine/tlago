package tlc

import "testing"

func TestDFIDAssumptionsUseSourceValidityPath(t *testing.T) {
	tool := NewTool()
	tool.Assumptions = []SemanticNode{NewNumeralNode(1)}
	tool.AssumptionIsAxiom = []bool{false}
	tool.ConfigErrors = []*ConfigError{NewConfigError(ECGeneral, "model checker only")}
	tool.CheckAssumptionsFunc = func(*Tool) int {
		t.Fatal("DFID delegated to Tool.checkAssumptions instead of its source loop")
		return ECGeneral
	}
	validityCalls := 0
	tool.IsValidExprFunc = func(_ *Tool, _ SemanticNode, ctxt *Context) (bool, error) {
		validityCalls++
		if ctxt != EmptyContext {
			t.Fatal("DFID assumption did not use the one-expression isValid overload")
		}
		return true, nil
	}
	checker := &DFIDModelChecker{AbstractChecker: &AbstractChecker{Tool: tool}}
	if got := checker.CheckAssumptions(); got != NoError || validityCalls != 1 {
		t.Fatalf("DFID assumptions returned %d with %d validity calls", got, validityCalls)
	}
}
