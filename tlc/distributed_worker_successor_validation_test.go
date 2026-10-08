package tlc

import "testing"

// Tool.isGoodState dereferences its state in the reference implementation.
// No original Java test directly covers a missing successor. These native
// checks distinguish that evaluator failure from an ordinary incomplete state.
func TestDistributedWorkerSuccessorValidation(t *testing.T) {
	for _, remote := range []bool{false, true} {
		for _, missing := range []bool{false, true} {
			name := "local"
			if remote {
				name = "tcp"
			}
			if missing {
				name += "/missing"
			} else {
				name += "/incomplete"
			}
			t.Run(name, func(t *testing.T) {
				worker, generated := managerOwnershipWorker(t, 1, nil)
				worker.App.Tool.IsGoodStateFunc = nil
				var successor *TLCStateMut
				if !missing {
					successor = &TLCStateMut{values: []Value{nil}}
				}
				worker.App.Tool.GetNextStatesFunc = func(*Tool, *Action, *TLCStateMut) (*StateVec, error) {
					*generated++
					return NewStateVec(1).Add(successor), nil
				}
				worker.OverallStatesComputed.Store(7)
				predecessor := &TLCStateMut{UID: 31, level: 4}
				invoke := worker.GetNextStates
				if remote {
					_, client := startWorkerRPC(t, NewLocalWorkerEndpoint(worker))
					invoke = client.GetNextStates
				}
				result, err := invoke([]*TLCStateMut{predecessor})
				failure, ok := err.(*WorkerException)
				if result != nil || !ok || failure.State1 == nil || failure.State1.UID != 31 || failure.State1.Level() != 4 {
					t.Fatalf("validation lost predecessor context: %v/%v", result, err)
				}
				if !remote && failure.State1 != predecessor {
					t.Fatal("local validation replaced predecessor")
				}
				if missing {
					if !isDistributedNullFailure(failure.Cause) || !failure.KeepCallStack || failure.State2 != nil {
						t.Fatalf("missing successor became ordinary incomplete state: %+v", failure)
					}
				} else if failure.Cause != nil || failure.KeepCallStack || failure.State2 == nil || failure.Error() != "Error: Successor state is not completely specified by the next-state action." {
					t.Fatalf("incomplete successor changed diagnostic: %+v", failure)
				}
				if *generated != 1 || worker.OverallStatesComputed.Load() != 7 || worker.Computing.Load() || worker.LastInvocation.Load() == 0 {
					t.Fatal("validation failure changed generation/statistics/finally ordering")
				}
			})
		}
	}
}
