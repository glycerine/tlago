package tlc

import "testing"

// No original method directly covers missing action/predicate evaluation.
// Native checks must not confuse broken evaluator input with a false action.
func TestDistributedWorkerActionValidation(t *testing.T) {
	for _, remote := range []bool{false, true} {
		for _, input := range []string{"missing-action", "missing-predicate", "false-action"} {
			transport := "local"
			if remote {
				transport = "tcp"
			}
			t.Run(transport+"/"+input, func(t *testing.T) {
				worker, generated := managerOwnershipWorker(t, 0, nil)
				tool := worker.App.Tool
				tool.GetNextStatesFunc = nil
				var action *Action
				if input == "missing-predicate" {
					action = NewAction(nil, EmptyContext, "missing")
				} else if input == "false-action" {
					action = NewAction(NewBuiltinOpApplNode(OpEq, NewIntValue(1), NewIntValue(2)), EmptyContext, "false")
				}
				tool.Actions = []*Action{action}
				worker.App.Actions = tool.Actions
				worker.App.checkDeadlock = true
				worker.OverallStatesComputed.Store(7)
				predecessor := &TLCStateMut{UID: 31, level: 4}
				invoke := worker.GetNextStates
				if remote {
					_, client := startWorkerRPC(t, NewLocalWorkerEndpoint(worker))
					invoke = client.GetNextStates
				}
				result, err := invoke([]*TLCStateMut{predecessor})
				failure, ok := err.(*WorkerException)
				if result != nil || !ok || failure.State1 == nil || failure.State1.UID != 31 || failure.State1.Level() != 4 || failure.State2 != nil {
					t.Fatalf("evaluation lost worker context: %v/%v", result, err)
				}
				if !remote && failure.State1 != predecessor {
					t.Fatal("local evaluation replaced predecessor")
				}
				if input == "false-action" {
					if failure.Cause != nil || failure.KeepCallStack || failure.Error() != "Error: deadlock reached." {
						t.Fatalf("false action did not retain ordinary deadlock: %+v", failure)
					}
				} else if !isDistributedNullFailure(failure.Cause) || !failure.KeepCallStack {
					t.Fatalf("missing evaluator input became deadlock: %+v", failure)
				}
				if *generated != 0 || worker.OverallStatesComputed.Load() != 7 || worker.Computing.Load() || worker.LastInvocation.Load() == 0 {
					t.Fatal("evaluation failure changed statistics/finally ordering")
				}
			})
		}
	}
}
