package tlc

import (
	"errors"
	"fmt"
	"testing"
)

// No original method directly covers failures escaping final-state recovery.
func TestWorkerPostconditionReconstructionFailureStopsEvaluation(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, transition := range []bool{false, true} {
		for _, failure := range []error{errors.New("state recovery failed"), NewOutOfMemoryError("state recovery allocation failed")} {
			t.Run(fmt.Sprintf("transition-%v/%T", transition, failure), func(t *testing.T) {
				parent := checkerTestState(1)
				tool := NewTool()
				tool.InitStates = []*TLCStateMut{parent}
				var successor *TLCStateMut
				if transition {
					successor = checkerTestState(2).SetPredecessor(parent)
					tool.Actions = []*Action{{Name: "Next"}}
					installTestNextStateGenerator(tool, func(*Tool, *Action, *TLCStateMut) (*StateVec, error) { return nil, failure })
				} else {
					tool.GetStateFunc = func(*Tool, uint64, ...any) (*TLCStateInfo, error) { return nil, failure }
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
					if got := recover(); got != failure || aliases != 0 || posts != 0 {
						t.Fatalf("recovery failure lost: %v, aliases %d, postconditions %d", got, aliases, posts)
					}
				}()
				checker.checkPostConditionWithErrorTrace(parent, successor, true)
			})
		}
	}
}
