package tlc

import "testing"

// No original method directly tests mutation after worker construction.
func TestWorkerRetainsConstructedDeadlockSetting(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, captured := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[captured], func(t *testing.T) {
			tool := NewTool()
			cur := checkerTestState(1)
			tool.InitStates = []*TLCStateMut{cur}
			tool.Actions = []*Action{{Name: "Next"}}
			installTestNextStateGenerator(tool, func(*Tool, *Action, *TLCStateMut) (*StateVec, error) { return NewStateVec(0), nil })
			checker := NewModelChecker(tool, t.TempDir(), captured)
			defer checker.ConcurrentTrace.Close()
			worker := checker.Workers[0]
			checker.CheckDeadlock = !captured
			stop, err := worker.DoNext(cur)
			if stop || err != nil {
				t.Fatalf("worker iteration: %v/%v", stop, err)
			}
			want := NoError
			if captured {
				want = ECTLCDeadlockReached
			}
			if checker.ErrorCode != want {
				t.Fatalf("deadlock code = %d, want %d", checker.ErrorCode, want)
			}
		})
	}
}

func TestWorkerRetainsConstructedLivenessAndMode(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, setting := range []string{"liveness", "debugger"} {
		for _, captured := range []bool{false, true} {
			name := setting + "/disabled"
			if captured {
				name = setting + "/enabled"
			}
			t.Run(name, func(t *testing.T) {
				tool := NewTool()
				if setting == "debugger" && captured {
					tool.Mode = ModeDebugger
				}
				tool.Actions = []*Action{{Name: "Next"}}
				installTestNextStateGenerator(tool, func(*Tool, *Action, *TLCStateMut) (*StateVec, error) { return NewStateVec(0), nil })
				directory := t.TempDir()
				checker := NewModelChecker(tool, directory, false, func(c *ModelChecker) {
					c.CheckLiveness = setting == "liveness" && captured
					c.LiveCheck = NewNoOpLiveCheck(tool, directory)
				})
				defer checker.ConcurrentTrace.Close()
				worker := checker.Workers[0]
				if setting == "liveness" {
					checker.CheckLiveness = !captured
				} else if captured {
					tool.Mode = ModeMC
				} else {
					tool.Mode = ModeDebugger
				}
				stop, err := worker.DoNext(checkerTestState(1))
				if stop || err != nil {
					t.Fatalf("worker iteration: %v/%v", stop, err)
				}
				if (worker.SetOfStates != nil) != captured {
					t.Fatal("set allocation followed a changed owner setting")
				}
				if captured {
					wantSize := 0
					if setting == "liveness" {
						wantSize = 1
					}
					if worker.SetOfStates.Size() != wantSize {
						t.Fatal("captured liveness did not control stuttering graph insertion")
					}
				}
			})
		}
	}
}
