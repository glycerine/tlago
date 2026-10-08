package tlc

import (
	"errors"
	"fmt"
	"testing"
)

// No original method directly covers a failure escaping postcondition ALIAS.
func TestWorkerPostconditionAliasFailureStopsEvaluation(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, failure := range []error{errors.New("alias backend failed"), NewOutOfMemoryError("alias allocation failed")} {
		for _, failAt := range []int{1, 2} {
			t.Run(fmt.Sprintf("%T/alias-%d", failure, failAt), func(t *testing.T) {
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
					if aliases == failAt {
						return nil, failure
					}
					return state, nil
				}
				tool.CheckPostConditionCEFunc = func(*Tool, Value) int { posts++; return NoError }
				checker := NewModelChecker(tool, t.TempDir(), false)
				defer checker.ConcurrentTrace.Close()
				defer func() {
					if got := recover(); got != failure || aliases != failAt || posts != 0 {
						t.Fatalf("alias failure lost: got %v, aliases %d, postconditions %d", got, aliases, posts)
					}
				}()
				checker.checkPostConditionWithErrorTrace(parent, successor, true)
			})
		}
	}
}
