package tlc

import (
	"strconv"
	"testing"
)

func TestWorkerPostconditionDoesNotReplaceMissingReconstruction(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, transition := range []bool{false, true} {
		name := "initial"
		if transition {
			name = "transition"
		}
		t.Run(name, func(t *testing.T) {
			parent := checkerTestState(1)
			tool := NewTool()
			tool.InitStates = []*TLCStateMut{parent}
			var successor *TLCStateMut
			if transition {
				successor = checkerTestState(2).SetPredecessor(parent)
			} else {
				tool.GetStateFunc = func(*Tool, uint64, ...any) (*TLCStateInfo, error) { return nil, nil }
			}
			aliases, posts := 0, 0
			tool.EvalAliasInfoPairFunc = func(_ *Tool, state *TLCStateInfo, _ *TLCStateMut) (*TLCStateInfo, error) {
				aliases++
				return state, nil
			}
			tool.CheckPostConditionCEFunc = func(*Tool, Value) int { posts++; return NoError }
			checker := NewModelChecker(tool, t.TempDir(), false)
			defer checker.ConcurrentTrace.Close()
			defer func() {
				if _, ok := recover().(*NullPointerException); !ok || aliases != 0 || posts != 0 {
					t.Fatalf("missing recovery was replaced: aliases %d, postconditions %d", aliases, posts)
				}
			}()
			checker.checkPostConditionWithErrorTrace(parent, successor, true)
		})
	}
}

func TestWorkerPostconditionPublishesMissingAlias(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, missingAt := range []int{1, 2} {
		t.Run(strconv.Itoa(missingAt), func(t *testing.T) {
			parent := checkerTestState(1)
			successor := checkerTestState(2).SetPredecessor(parent)
			tool := NewTool()
			tool.InitStates = []*TLCStateMut{parent}
			tool.Actions = []*Action{{Name: "Next"}}
			installTestNextStateGenerator(tool, func(*Tool, *Action, *TLCStateMut) (*StateVec, error) {
				states := NewStateVec(1)
				states.AddElement(successor.Copy())
				return states, nil
			})
			aliases, posts := 0, 0
			tool.EvalAliasInfoPairFunc = func(_ *Tool, state *TLCStateInfo, _ *TLCStateMut) (*TLCStateInfo, error) {
				aliases++
				if aliases == missingAt {
					return nil, nil
				}
				return state, nil
			}
			tool.CheckPostConditionCEFunc = func(*Tool, Value) int { posts++; return NoError }
			checker := NewModelChecker(tool, t.TempDir(), false)
			defer checker.ConcurrentTrace.Close()
			defer func() {
				if _, ok := recover().(*NullPointerException); !ok || aliases != 2 || posts != 0 {
					t.Fatalf("missing alias was replaced: aliases %d, postconditions %d", aliases, posts)
				}
			}()
			checker.checkPostConditionWithErrorTrace(parent, successor, true)
		})
	}
}

func TestCounterExampleDoesNotReplaceMissingStateInfo(t *testing.T) {
	for _, trace := range [][]*TLCStateInfo{{nil}, {NewTLCStateInfo(checkerTestState(1)), nil}} {
		t.Run(strconv.Itoa(len(trace)), func(t *testing.T) {
			defer func() {
				if _, ok := recover().(*NullPointerException); !ok {
					t.Fatal("missing state info synthesized a counterexample state")
				}
			}()
			NewCounterExample(trace, UnknownAction, 0, true)
		})
	}
}
