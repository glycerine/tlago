package tlc

import "testing"

// No original method directly distinguishes final-state recovery overloads.
func TestWorkerPostconditionReconstructionUsesSourceOverloadAndMetadata(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, route := range []string{"initial", "noninitial", "transition"} {
		t.Run(route, func(t *testing.T) {
			parent, target, generated := checkerTestState(1), checkerTestState(2), checkerTestState(2)
			target.WorkerID, target.UID = 3, 77
			wantWorkerID, wantUID := generated.WorkerID, generated.UID
			action := &Action{Name: "Next"}
			tool := NewTool()
			tool.Actions = []*Action{action}
			fingerprintCalls, transitionCalls := 0, 0
			tool.GetStateFunc = func(*Tool, uint64, ...any) (*TLCStateInfo, error) {
				fingerprintCalls++
				return NewTLCStateInfo(generated), nil
			}
			installTestNextStateGenerator(tool, func(*Tool, *Action, *TLCStateMut) (*StateVec, error) {
				transitionCalls++
				states := NewStateVec(1)
				states.AddElement(generated)
				return states, nil
			})
			checker := NewModelChecker(tool, t.TempDir(), false)
			defer checker.ConcurrentTrace.Close()
			var info *TLCStateInfo
			if route == "initial" {
				info = checker.stateInfoForState(tool, target, nil)
			} else if route == "noninitial" {
				info = checker.stateInfoForState(tool, target, parent)
			} else {
				info = checker.stateInfoForTransition(tool, target, parent)
			}
			wantFP, wantTransition := 0, 1
			if route == "initial" {
				wantFP, wantTransition = 1, 0
			}
			if info == nil || info.State != generated || fingerprintCalls != wantFP || transitionCalls != wantTransition || generated.WorkerID != wantWorkerID || generated.UID != wantUID {
				t.Fatalf("reconstruction diverged: info %v, fingerprint/transition %d/%d, worker/UID %d/%d; want %d/%d, %d/%d", info, fingerprintCalls, transitionCalls, generated.WorkerID, generated.UID, wantFP, wantTransition, wantWorkerID, wantUID)
			}
			if route != "initial" && info.Action() != action {
				t.Fatal("generated transition action was lost")
			}
		})
	}
}
