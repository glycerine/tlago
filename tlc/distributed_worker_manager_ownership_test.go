package tlc

import "testing"

func managerOwnershipWorker(t *testing.T, count int, manager *DistributedFPSetManager) (*DistributedWorker, *int) {
	t.Helper()
	oldTool := stateTool
	stateTool = nil
	t.Cleanup(func() { stateTool = oldTool })
	generated := 0
	tool := &Tool{Actions: []*Action{{}},
		GetNextStatesFunc: func(*Tool, *Action, *TLCStateMut) (*StateVec, error) {
			generated++
			states := NewStateVec(count)
			for i := 0; i < count; i++ {
				states.Add(&TLCStateMut{values: []Value{NewIntValue(int32(i + 1))}})
			}
			return states, nil
		}, IsGoodStateFunc: func(*Tool, *TLCStateMut) bool { return true }}
	worker := NewDistributedWorker(0, tool, manager, DistributedWorkerAddress{Hostname: "worker", Port: 1234})
	worker.App.checkDeadlock = false
	t.Cleanup(worker.Runtime.executor.Shutdown)
	return worker, &generated
}

// There is no direct original missing-manager test. Source worker construction
// retains null; the first manager dereference follows generation/statistics.
func TestDistributedWorkerRetainsMissingManager(t *testing.T) {
	for _, count := range []int{0, 2} {
		t.Run(fmtInt(count), func(t *testing.T) {
			worker, generated := managerOwnershipWorker(t, count, nil)
			if worker.FPSetManager != nil {
				t.Fatal("worker invented an empty fingerprint manager")
			}
			predecessor := &TLCStateMut{UID: 31, level: 4}
			worker.OverallStatesComputed.Store(7)
			result, err := worker.GetNextStates([]*TLCStateMut{predecessor})
			failure, ok := err.(*WorkerException)
			if result != nil || !ok || !isDistributedNullFailure(failure.Cause) || failure.State1 != predecessor || failure.State2 != nil || !failure.KeepCallStack {
				t.Fatalf("missing manager failure/context %v/%v", result, err)
			}
			if *generated != 1 || worker.OverallStatesComputed.Load() != int64(7+count) || worker.Computing.Load() || worker.LastInvocation.Load() == 0 {
				t.Fatal("missing manager changed generation/statistics/finally ordering")
			}
		})
	}
}

func TestWorkerRPCMissingManagerFailureContext(t *testing.T) {
	worker, generated := managerOwnershipWorker(t, 2, nil)
	_, client := startWorkerRPC(t, NewLocalWorkerEndpoint(worker))
	predecessor := &TLCStateMut{UID: 31, level: 4}
	result, err := client.GetNextStates([]*TLCStateMut{predecessor})
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || !isDistributedNullFailure(failure.Cause) || failure.State1 == nil || failure.State1 == predecessor || failure.State1.UID != 31 || failure.State1.Level() != 4 || failure.State2 != nil || !failure.KeepCallStack {
		t.Fatalf("remote missing manager failure/context %v/%v", result, err)
	}
	if *generated != 1 || worker.OverallStatesComputed.Load() != 2 || worker.Computing.Load() || worker.FPSetManager != nil {
		t.Fatal("remote missing manager changed worker state")
	}
	if alive, err := client.IsAlive(); err != nil || !alive {
		t.Fatalf("failure stopped worker endpoint %v/%v", alive, err)
	}
}

func TestDistributedWorkerRetainsEmptyManager(t *testing.T) {
	manager := NewDistributedFPSetManager()
	worker, generated := managerOwnershipWorker(t, 0, manager)
	result, err := worker.GetNextStates([]*TLCStateMut{{UID: 31}})
	if err != nil || result == nil || len(result.NextStates) != 0 || len(result.NextFingerprints) != 0 || result.StatesComputed != 0 || worker.FPSetManager != manager || *generated != 1 || worker.Computing.Load() {
		t.Fatalf("empty manager conflated with missing manager: %v/%v", result, err)
	}
}

func TestDistributedWorkerGenerationFailurePrecedesMissingManager(t *testing.T) {
	worker, generated := managerOwnershipWorker(t, 2, nil)
	worker.OverallStatesComputed.Store(7)
	worker.App.Tool.IsGoodStateFunc = func(*Tool, *TLCStateMut) bool { return false }
	predecessor := &TLCStateMut{UID: 31, level: 4}
	result, err := worker.GetNextStates([]*TLCStateMut{predecessor})
	failure, ok := err.(*WorkerException)
	if result != nil || !ok || failure.Cause != nil || failure.State1 != predecessor || failure.State2 == nil || failure.KeepCallStack || failure.Error() != "Error: Successor state is not completely specified by the next-state action." {
		t.Fatalf("generation failure replaced by manager failure: %v/%v", result, err)
	}
	if *generated != 1 || worker.OverallStatesComputed.Load() != 7 || worker.Computing.Load() {
		t.Fatal("failed generation entered total-statistics update")
	}
}
