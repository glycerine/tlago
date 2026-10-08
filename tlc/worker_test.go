package tlc

import "testing"

func TestLivenessErrorNeedsCallStackReplayRecognizesTLCError(t *testing.T) {
	if !livenessErrorNeedsCallStackReplay(newTLCErrorCode(ECTLCLiveStatePredicateNonBool)) {
		t.Fatalf("livenessErrorNeedsCallStackReplay(TLCError) = false, want true like Java EvalException")
	}
}

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
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVecFrom([]*TLCStateMut{seenSucc, newSucc}), nil
	})
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
	if worker.GetStatesGenerated() != 2 {
		t.Fatalf("worker states generated = %d, want 2", worker.GetStatesGenerated())
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
	if worker.UnseenSuccessorStates != 0 {
		t.Fatalf("worker unseen successor count = %d, want reset after out-degree sample", worker.UnseenSuccessorStates)
	}
	if worker.OutDegree == nil {
		t.Fatalf("worker out-degree statistics were not initialized")
	}
	if worker.OutDegree.Observations() != 1 || worker.OutDegree.Median() != 1 {
		t.Fatalf("worker out-degree observations/median = %v/%v, want 1/1", worker.OutDegree.Observations(), worker.OutDegree.Median())
	}
	if worker.GetMaxLevel() != 2 {
		t.Fatalf("worker max level = %d, want 2", worker.GetMaxLevel())
	}
}

func TestWorkerFunctorPathReportsDeadlockWhenNoSuccessorsAreGenerated(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}
	cur := checkerTestState(1)

	tool := NewTool()
	tool.InitStates = []*TLCStateMut{cur}
	tool.Actions = []*Action{action}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVec(0), nil
	})

	mc := NewModelChecker(tool, t.TempDir(), true)
	worker := NewModelCheckingWorker(0, mc, tool)
	stop, err := worker.DoNext(cur)
	if err != nil {
		t.Fatalf("Worker.DoNext returned error: %v", err)
	}
	if stop {
		t.Fatalf("Worker.DoNext stop = true, want false so Worker.Run observes the finished queue like Java")
	}
	if mc.ErrorCode != ECTLCDeadlockReached {
		t.Fatalf("error code = %d, want %d", mc.ErrorCode, ECTLCDeadlockReached)
	}
	if mc.ErrState != cur {
		t.Fatalf("error state = %p, want current state %p", mc.ErrState, cur)
	}
	if worker.OutDegree == nil {
		t.Fatalf("worker out-degree statistics were not initialized")
	}
	if worker.OutDegree.Observations() != 1 || worker.OutDegree.Median() != 0 {
		t.Fatalf("worker deadlock out-degree observations/median = %v/%v, want 1/0", worker.OutDegree.Observations(), worker.OutDegree.Median())
	}
	if dequeued := mc.StateQueue.SDequeue(); dequeued != nil {
		t.Fatalf("SDequeue after deadlock = %p, want nil after finishAll", dequeued)
	}
}

func TestWorkerCollectsInModelSuccessorsForLivenessSet(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}
	cur := checkerTestState(1)
	seenSucc := checkerTestState(1)
	newSucc := checkerTestState(2)

	tool := NewTool()
	tool.Actions = []*Action{action}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVecFrom([]*TLCStateMut{seenSucc, newSucc}), nil
	})

	mc := NewModelChecker(tool, t.TempDir(), true)
	mc.FPSet.Put(cur.FingerPrint())
	worker := NewModelCheckingWorker(0, mc, tool)
	worker.SetOfStates = NewSetOfStates(1)

	stop, err := worker.DoNext(cur)
	if err != nil {
		t.Fatalf("Worker.DoNext returned error: %v", err)
	}
	if stop {
		t.Fatalf("Worker.DoNext stop = true, want false")
	}
	if worker.SetOfStates.Size() != 2 {
		t.Fatalf("collected states = %d, want 2 seen and unseen in-model successors", worker.SetOfStates.Size())
	}
	if !stateSetContains(worker.SetOfStates, seenSucc) || !stateSetContains(worker.SetOfStates, newSucc) {
		t.Fatalf("collected set does not contain both in-model successors")
	}
	if got := mc.GetStateQueueSize(); got != 1 {
		t.Fatalf("queued states = %d, want 1 unseen successor", got)
	}
}

func TestWorkerDoesNotCollectOutOfModelSuccessors(t *testing.T) {
	initTLCCheckerTest(t)
	action := &Action{Name: "Next"}
	cur := checkerTestState(1)
	inModelSucc := checkerTestState(2)
	outOfModelSucc := checkerTestState(99)

	tool := NewTool()
	tool.Actions = []*Action{action}
	installTestNextStateGenerator(tool, func(tl *Tool, a *Action, state *TLCStateMut) (*StateVec, error) {
		return NewStateVecFrom([]*TLCStateMut{inModelSucc, outOfModelSucc}), nil
	})
	tool.IsInModelFunc = func(tl *Tool, state *TLCStateMut) (bool, error) {
		value := state.Lookup(UniqueStringOf("x"))
		intValue, ok := value.(*IntValue)
		return !ok || intValue.Val != 99, nil
	}

	mc := NewModelChecker(tool, t.TempDir(), true)
	worker := NewModelCheckingWorker(0, mc, tool)
	worker.SetOfStates = NewSetOfStates(1)

	stop, err := worker.DoNext(cur)
	if err != nil {
		t.Fatalf("Worker.DoNext returned error: %v", err)
	}
	if stop {
		t.Fatalf("Worker.DoNext stop = true, want false")
	}
	if worker.SetOfStates.Size() != 1 {
		t.Fatalf("collected states = %d, want only the in-model successor", worker.SetOfStates.Size())
	}
	if !stateSetContains(worker.SetOfStates, inModelSucc) {
		t.Fatalf("collected set does not contain the in-model successor")
	}
	if stateSetContains(worker.SetOfStates, outOfModelSucc) {
		t.Fatalf("collected set contains an out-of-model successor")
	}
}

func stateSetContains(set *SetOfStates, want *TLCStateMut) bool {
	for _, state := range set.ToSlice() {
		if state.Equal(want) {
			return true
		}
	}
	return false
}
