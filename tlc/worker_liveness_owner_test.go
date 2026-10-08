package tlc

import (
	"errors"
	"testing"
)

// No original method directly exercises missing liveness-tail owners.
func TestWorkerLivenessPreservesOwnerAccessOrder(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, scenario := range []string{"current", "set", "checker", "writer", "livecheck", "checker-tool", "healthy-tool", "writer-failure"} {
		t.Run(scenario, func(t *testing.T) {
			tool := NewTool()
			originalNoDebug, checkerNoDebug, writes := 0, 0, 0
			tool.NoDebugFunc = func(tl *Tool) *Tool { originalNoDebug++; return tl }
			checkerTool := NewTool()
			checkerTool.NoDebugFunc = func(tl *Tool) *Tool { checkerNoDebug++; return tl }
			failure := errors.New("stuttering writer failed")
			state := checkerTestState(1)
			writer := &StateWriter{WriteTransitionFunc: func(cur, next *TLCStateMut, status StateVisitStatus, _ *Action, _ ...SemanticNode) error {
				writes++
				if cur != state || next != state || status != StateVisitUnseen {
					t.Fatal("stuttering state changed")
				}
				if scenario == "writer-failure" {
					return failure
				}
				return nil
			}}
			checker := NewModelChecker(tool, t.TempDir(), false, WithModelCheckerStateWriter(writer), WithModelCheckerLiveCheck(NewNoOpLiveCheck(tool, "")))
			defer checker.ConcurrentTrace.Close()
			worker := checker.Workers[0]
			worker.SetOfStates = NewSetOfStates(1)
			checker.Tool = checkerTool
			wantSize, wantWrites, wantNoDebug := 1, 0, 0
			var wantErr error
			expectPanic := true
			switch scenario {
			case "current":
				state = nil
				wantSize = 0
			case "set":
				worker.SetOfStates = nil
				wantSize = -1
			case "checker":
				worker.Checker = nil
			case "writer":
				checker.AllStateWriter = nil
			case "livecheck":
				checker.LiveCheck = nil
				wantWrites, wantNoDebug = 1, 1
			case "checker-tool":
				checker.Tool = nil
				wantWrites = 1
			case "healthy-tool":
				expectPanic = false
				wantWrites, wantNoDebug = 1, 1
			case "writer-failure":
				checker.LiveCheck = nil
				expectPanic = false
				wantWrites, wantErr = 1, failure
			}
			var err error
			defer func() {
				got := recover()
				if expectPanic {
					if _, ok := got.(*NullPointerException); !ok {
						t.Fatalf("missing owner failure %T/%v", got, got)
					}
				} else if got != nil {
					t.Fatalf("unexpected panic %v", got)
				}
				size := -1
				if worker.SetOfStates != nil {
					size = worker.SetOfStates.Size()
				}
				if err != wantErr || size != wantSize || writes != wantWrites || checkerNoDebug != wantNoDebug || originalNoDebug != 0 {
					t.Fatalf("liveness order changed: error %v, set %d, writes %d, checker/original tool %d/%d; want %v, %d, %d, %d/0", err, size, writes, checkerNoDebug, originalNoDebug, wantErr, wantSize, wantWrites, wantNoDebug)
				}
			}()
			err = worker.CheckLiveness(state)
		})
	}
}

func TestWorkerLivenessReplayUsesCheckerTool(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	tool, checkerTool := NewTool(), NewTool()
	tool.InvariantNames, checkerTool.InvariantNames = []string{"worker marker"}, []string{"checker marker"}
	originalNoDebug, checkerNoDebug := 0, 0
	tool.NoDebugFunc = func(tl *Tool) *Tool { originalNoDebug++; return tl }
	checkerTool.NoDebugFunc = func(tl *Tool) *Tool { checkerNoDebug++; return tl }
	failure := NewEvalException(ECTLCLiveStatePredicateNonBool)
	var evaluated []*Tool
	stateExpr := NewLNState("fail", nil, EmptyContext, func(tl *Tool, _, _ *TLCStateMut) (bool, error) {
		evaluated = append(evaluated, tl)
		return false, failure
	})
	liveCheck := &LiveCheck{Checkers: []*LiveChecker{{Solution: &OrderOfSolution{CheckState: []*LiveExprNode{stateExpr}}}}}
	checker := NewModelChecker(tool, t.TempDir(), false, WithModelCheckerLiveCheck(liveCheck))
	defer checker.ConcurrentTrace.Close()
	checker.Tool = checkerTool
	worker := checker.Workers[0]
	worker.SetOfStates = NewSetOfStates(1)
	err := worker.CheckLiveness(checkerTestState(1))
	if err != failure || len(evaluated) != 2 || evaluated[0] != checkerTool || len(evaluated[1].InvariantNames) != 1 || evaluated[1].InvariantNames[0] != "checker marker" || evaluated[1].CallStack == nil || checkerNoDebug != 2 || originalNoDebug != 0 {
		t.Fatalf("liveness replay tool changed: error %v, evaluations %d, checker/original noDebug %d/%d", err, len(evaluated), checkerNoDebug, originalNoDebug)
	}
}
