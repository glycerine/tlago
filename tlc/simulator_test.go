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

func TestSimulatorConfigArilKeepsJavaConstructorValue(t *testing.T) {
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	t.Cleanup(func() {
		SetNumWorkers(oldWorkers)
		SetSimulator(nil)
		SetTLCStateTool(nil)
	})

	sim := NewSimulator(NewTool(), true, 1, 1, 17, WithSimulatorAril(5))
	if sim.Aril != 5 {
		t.Fatalf("runtime aril = %d, want 5", sim.Aril)
	}
	config, ok := sim.GetConfig().(*RecordValue)
	if !ok {
		t.Fatalf("config = %T, want *RecordValue", sim.GetConfig())
	}
	aril, err := config.Apply(NewStringValueFromUnique(tlcGetAril))
	if err != nil {
		t.Fatalf("config aril lookup returned error: %v", err)
	}
	if got, want := aril.String(), `"0"`; got != want {
		t.Fatalf("config aril = %s, want %s like Java's eager constructor record", got, want)
	}
}

func TestSimulatorRLFloatPropertiesAcceptJavaSuffixes(t *testing.T) {
	for _, input := range []string{".3d", ".3D", ".3f", ".3F"} {
		got, err := parseJavaDoubleProperty(input)
		if err != nil {
			t.Fatalf("parseJavaDoubleProperty(%q) returned error: %v", input, err)
		}
		if got != 0.3 {
			t.Fatalf("parseJavaDoubleProperty(%q) = %v, want 0.3", input, got)
		}
	}
	if got := simulatorPropertyFloat("missing", "missing", 1.25); got != 1.25 {
		t.Fatalf("simulatorPropertyFloat fallback = %v, want 1.25", got)
	}
}
