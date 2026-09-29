package tlc

import "testing"

func TestWorkerFunctorPathProcessesSuccessorsLikeModelChecker(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}
	invariant := &Action{Name: "Inv"}
	implied := &Action{Name: "Imp"}
	cur := checkerTestState(1)
	seenSucc := checkerTestState(1)
	newSucc := checkerTestState(2)
	invariantChecks := 0
	impliedChecks := 0

	tool := NewTool()
	tool.Actions = []*Action{action}
	tool.Invariants = []*Action{invariant}
	tool.InvariantNames = []string{"Inv"}
	tool.ImpliedActions = []*Action{implied}
	tool.ImpliedActNames = []string{"Imp"}
	tool.GetNextStatesFunc = func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVecFrom([]*TLCStateMut{seenSucc, newSucc}), nil
	}
	tool.IsValidStateFunc = func(tl *Tool, a *Action, state *TLCStateMut) (bool, error) {
		if a == invariant {
			invariantChecks++
		}
		return true, nil
	}
	tool.IsValidTransitionFunc = func(tl *Tool, a *Action, s0 *TLCStateMut, s1 *TLCStateMut) (bool, error) {
		if a == implied {
			impliedChecks++
		}
		return true, nil
	}

	mc := NewModelChecker(tool, t.TempDir(), true)
	mc.FPSet.Put(cur.FingerPrint())
	worker := NewModelCheckingWorker(0, mc, tool)
	stop, err := worker.DoNext(cur)
	if err != nil {
		t.Fatalf("Worker.DoNext returned error: %v", err)
	}
	if stop {
		t.Fatalf("Worker.DoNext stop = true, want false")
	}
	if worker.StatesGenerated != 2 {
		t.Fatalf("worker states generated = %d, want 2", worker.StatesGenerated)
	}
	if mc.NextStatesGenerated != 0 {
		t.Fatalf("checker direct next states generated = %d, want 0 for worker path", mc.NextStatesGenerated)
	}
	if got := mc.GetStatesGenerated(); got != 2 {
		t.Fatalf("total states generated = %d, want 2 from worker accounting", got)
	}
	if got := mc.GetStateQueueSize(); got != 1 {
		t.Fatalf("queued states = %d, want 1 new successor", got)
	}
	if invariantChecks != 1 {
		t.Fatalf("invariant checks = %d, want 1 for the unseen successor only", invariantChecks)
	}
	if impliedChecks != 2 {
		t.Fatalf("implied-action checks = %d, want 2 for seen and unseen successors", impliedChecks)
	}
	if worker.UnseenSuccessorStates != 1 {
		t.Fatalf("worker unseen successor count = %d, want 1", worker.UnseenSuccessorStates)
	}
	if worker.MaxLevel != 2 {
		t.Fatalf("worker max level = %d, want 2", worker.MaxLevel)
	}
}

func TestWorkerFunctorPathReportsDeadlockWhenNoSuccessorsAreGenerated(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}
	cur := checkerTestState(1)

	tool := NewTool()
	tool.Actions = []*Action{action}
	tool.GetNextStatesFunc = func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVec(0), nil
	}

	mc := NewModelChecker(tool, t.TempDir(), true)
	worker := NewModelCheckingWorker(0, mc, tool)
	stop, err := worker.DoNext(cur)
	if err != nil {
		t.Fatalf("Worker.DoNext returned error: %v", err)
	}
	if !stop {
		t.Fatalf("Worker.DoNext stop = false, want true")
	}
	if mc.ErrorCode != ECTLCDeadlockReached {
		t.Fatalf("error code = %d, want %d", mc.ErrorCode, ECTLCDeadlockReached)
	}
	if mc.ErrState != cur {
		t.Fatalf("error state = %p, want current state %p", mc.ErrState, cur)
	}
}
