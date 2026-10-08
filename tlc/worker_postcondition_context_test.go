package tlc

import "testing"

func TestWorkerPostconditionRequiresCurrentStateAndTool(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, scenario := range []string{"missing-current", "missing-current-with-successor", "missing-tool"} {
		t.Run(scenario, func(t *testing.T) {
			state := checkerTestState(1)
			tool := NewTool()
			tool.InitStates = []*TLCStateMut{state}
			calls := 0
			tool.CheckPostConditionFunc = func(*Tool) int { calls++; return NoError }
			tool.CheckPostConditionCEFunc = func(*Tool, Value) int { calls++; return NoError }
			checker := NewModelChecker(tool, t.TempDir(), false)
			defer checker.ConcurrentTrace.Close()
			var current, successor *TLCStateMut
			if scenario == "missing-current-with-successor" {
				successor = state
			}
			if scenario == "missing-tool" {
				current, tool = state, nil
			}
			defer func() {
				if _, ok := recover().(*NullPointerException); !ok || calls != 0 {
					t.Fatalf("missing context accepted: postconditions %d", calls)
				}
			}()
			checker.checkPostConditionWithErrorTraceTool(tool, current, successor, true)
		})
	}
}

func TestWorkerPostconditionTracePrefixUsesOwnerAndKeepsInitialState(t *testing.T) {
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, scenario := range []string{"initial", "missing-owner-initial", "missing-owner-noninitial"} {
		t.Run(scenario, func(t *testing.T) {
			state := checkerTestState(1)
			checker := NewModelChecker(NewTool(), t.TempDir(), false)
			defer checker.ConcurrentTrace.Close()
			if scenario != "initial" {
				checker.ConcurrentTrace, checker.Trace = nil, nil
				if scenario == "missing-owner-noninitial" {
					state.SetPredecessor(checkerTestState(0))
				}
				defer func() {
					if _, ok := recover().(*NullPointerException); !ok {
						t.Fatal("missing trace owner was substituted")
					}
				}()
				checker.traceInfoPrefix(state)
				return
			}
			prefix := checker.traceInfoPrefix(state)
			if len(prefix) != 1 || prefix[0].State != state {
				t.Fatalf("initial prefix removed: %v", prefix)
			}
		})
	}
}

func TestWorkerPostconditionRequiresLastPrefixState(t *testing.T) {
	for _, scenario := range []string{"empty", "missing-last"} {
		t.Run(scenario, func(t *testing.T) {
			var prefix []*TLCStateInfo
			if scenario == "missing-last" {
				prefix = []*TLCStateInfo{nil}
			}
			defer func() {
				failure := recover()
				if scenario == "empty" {
					if _, ok := failure.(*NoSuchElementException); !ok {
						t.Fatalf("empty prefix failure %T", failure)
					}
				} else if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("missing last-state failure %T", failure)
				}
			}()
			lastTraceState(prefix)
		})
	}
}
