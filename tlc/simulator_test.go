package tlc

import "testing"

func TestSimulatorPrintBehaviorAliasUsesPairwiseJavaPath(t *testing.T) {
	initTLCCheckerTest(t)
	cur := checkerTestState(0)
	succ := checkerTestState(1).SetPredecessor(cur)

	pairCalls := 0
	tool := NewTool()
	tool.EvalAliasInfoFunc = func(tl *Tool, current *TLCStateInfo, successor *TLCStateMut, prefix func() []*TLCStateInfo) (*TLCStateInfo, error) {
		t.Fatalf("simulator behavior alias used prefix-aware trace path; Java Simulator.printBehavior calls evalAlias(current, successor)")
		return current, nil
	}
	tool.EvalAliasInfoPairFunc = func(tl *Tool, current *TLCStateInfo, successor *TLCStateMut) (*TLCStateInfo, error) {
		pairCalls++
		if current == nil || successor == nil {
			t.Fatalf("pairwise alias call %d had current=%v successor=%v", pairCalls, current, successor)
		}
		return current, nil
	}

	sim := &Simulator{Tool: tool}
	sim.printBehaviorTrace(NewStateVecFrom([]*TLCStateMut{cur, succ}))

	if pairCalls != 2 {
		t.Fatalf("pairwise alias calls = %d, want one per trace state", pairCalls)
	}
}
