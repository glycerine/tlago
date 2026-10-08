package tlc

import (
	"reflect"
	"testing"
)

// No original method replaces the checker tool after worker construction.
func TestWorkerSuccessorEvaluationUsesWorkerTool(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, excluded := range []bool{false, true} {
		for _, missing := range []bool{false, true} {
			name := "eligible"
			if excluded {
				name = "excluded"
			}
			if missing {
				name += "/missing-checker-tool"
			} else {
				name += "/replacement-checker-tool"
			}
			t.Run(name, func(t *testing.T) {
				var calls []string
				makeTool := func(prefix string) *Tool {
					tool := NewTool()
					tool.Invariants = []*Action{{Name: "Inv"}}
					tool.ImpliedActions = []*Action{{Name: "Implied"}}
					tool.ModelConstraints = []SemanticNode{"state constraint"}
					tool.ActionConstraints = []SemanticNode{"action constraint"}
					tool.IsGoodStateFunc = func(*Tool, *TLCStateMut) bool { calls = append(calls, prefix+"good"); return true }
					tool.IsInModelFunc = func(*Tool, *TLCStateMut) (bool, error) { calls = append(calls, prefix+"model"); return !excluded, nil }
					tool.IsInActionsFunc = func(*Tool, *TLCStateMut, *TLCStateMut) (bool, error) {
						calls = append(calls, prefix+"actions")
						return true, nil
					}
					tool.IsInModelForConstraintFunc = func(*Tool, SemanticNode, *TLCStateMut) (bool, error) {
						calls = append(calls, prefix+"state reason")
						return false, nil
					}
					tool.IsInActionsForConstraintFn = func(*Tool, SemanticNode, *TLCStateMut, *TLCStateMut) (bool, error) {
						calls = append(calls, prefix+"action reason")
						return false, nil
					}
					tool.IsValidStateFunc = func(*Tool, *Action, *TLCStateMut) (bool, error) {
						calls = append(calls, prefix+"invariant")
						return true, nil
					}
					tool.IsValidTransitionFunc = func(*Tool, *Action, *TLCStateMut, *TLCStateMut) (bool, error) {
						calls = append(calls, prefix+"implied")
						return true, nil
					}
					return tool
				}
				writer := &StateWriter{Constrained: true}
				checker := NewModelChecker(makeTool(""), t.TempDir(), false, WithModelCheckerStateWriter(writer))
				defer checker.ConcurrentTrace.Close()
				checker.Tool = makeTool("replacement ")
				if missing {
					checker.Tool = nil
				}
				parent, successor := checkerTestState(1), checkerTestState(2)
				parent.UID = 37
				_, err := checker.Workers[0].AddNextElement(parent, &Action{Name: "Next"}, successor)
				want := []string{"good", "model", "actions", "invariant", "implied"}
				wantSize := int64(1)
				if excluded {
					want = []string{"good", "model", "state reason", "action reason", "invariant", "implied"}
					wantSize = 0
				}
				if err != nil || !reflect.DeepEqual(calls, want) || checker.StateQueue.Size() != wantSize {
					t.Fatalf("evaluation owner changed: err %v, calls %v, want %v, queue %d", err, calls, want, checker.StateQueue.Size())
				}
			})
		}
	}
}
