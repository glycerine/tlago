package tlc

import "testing"

// No original method replaces the checker tool before worker error reporting.
func TestWorkerErrorPostconditionUsesWorkerTool(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, scenario := range []string{"deadlock", "invariant", "implied"} {
		t.Run(scenario, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			parent, successor := checkerTestState(1), checkerTestState(2)
			workerPosts, checkerPosts, workerAliases, checkerAliases := 0, 0, 0, 0
			makeTool := func(posts, aliases *int) *Tool {
				tool := NewTool()
				tool.InitStates = []*TLCStateMut{parent}
				tool.GetStateFunc = func(_ *Tool, fp uint64, _ ...any) (*TLCStateInfo, error) {
					if fp == successor.FingerPrint() {
						return NewTLCStateInfo(successor.Copy()), nil
					}
					return NewTLCStateInfo(parent.Copy()), nil
				}
				if scenario != "deadlock" {
					tool.Actions = []*Action{{Name: "Next"}}
					installTestNextStateGenerator(tool, func(*Tool, *Action, *TLCStateMut) (*StateVec, error) {
						states := NewStateVec(1)
						states.AddElement(successor.Copy())
						return states, nil
					})
				}
				tool.EvalAliasInfoPairFunc = func(_ *Tool, state *TLCStateInfo, _ *TLCStateMut) (*TLCStateInfo, error) {
					*aliases++
					return state, nil
				}
				tool.CheckPostConditionCEFunc = func(_ *Tool, value Value) int {
					*posts++
					if _, ok := value.(*CounterExample); !ok {
						t.Fatalf("counterexample type %T", value)
					}
					return NoError
				}
				return tool
			}
			tool := makeTool(&workerPosts, &workerAliases)
			if scenario == "invariant" {
				tool.Invariants = []*Action{{Name: "Inv"}}
				tool.InvariantNames = []string{"Inv"}
				tool.IsValidStateFunc = func(*Tool, *Action, *TLCStateMut) (bool, error) { return false, nil }
			}
			if scenario == "implied" {
				tool.ImpliedActions = []*Action{{Name: "Implied"}}
				tool.ImpliedActNames = []string{"Implied"}
				tool.IsValidTransitionFunc = func(*Tool, *Action, *TLCStateMut, *TLCStateMut) (bool, error) { return false, nil }
			}
			checker := NewModelChecker(tool, t.TempDir(), scenario == "deadlock")
			defer checker.ConcurrentTrace.Close()
			checker.Tool = makeTool(&checkerPosts, &checkerAliases)
			worker := checker.Workers[0]
			if err := worker.WriteInitState(parent, parent.FingerPrint()); err != nil {
				t.Fatal(err)
			}
			if scenario == "deadlock" {
				if stop, err := worker.DoNext(parent); stop || err != nil {
					t.Fatalf("deadlock outcome %v, %v", stop, err)
				}
			} else {
				if _, err := worker.AddNextElement(parent, &Action{Name: "Next"}, successor); err == nil {
					t.Fatal("violation did not stop successor publication")
				}
			}
			if workerPosts != 1 || checkerPosts != 0 || workerAliases == 0 || checkerAliases != 0 {
				t.Fatalf("postcondition owner changed: worker posts/aliases %d/%d, checker %d/%d", workerPosts, workerAliases, checkerPosts, checkerAliases)
			}
		})
	}
}
