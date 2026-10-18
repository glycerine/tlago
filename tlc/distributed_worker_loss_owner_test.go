package tlc

import "testing"

// No original Java method directly covers missing coordinator ownership.
// Cleanup claims its one-time flag before coordinator removal; the timer must
// resolve the coordinator queue before it can enter that cleanup operation.
func TestDistributedWorkerLossRequiresCoordinator(t *testing.T) {
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	for _, phase := range []string{"direct", "timer_not_alive", "timer_remote_failure", "timer_alive"} {
		t.Run(phase, func(t *testing.T) {
			SetNumWorkers(2)
			queue := NewMemStateQueue()
			states := []*TLCStateMut{{UID: 7}, {UID: 11}}
			worker := &timerStatusWorker{rpcTestWorker: &rpcTestWorker{}, alive: phase == "timer_alive"}
			if phase == "timer_remote_failure" {
				worker.failure = distributedTestRemoteFailure("worker unavailable")
			}
			thread := &TLCServerThread{Worker: NewDistributedWorkerSmartProxy(worker), keepAliveDone: make(chan struct{})}
			thread.cleanupGlobals.Store(true)
			thread.setStates(states)
			err := invokeDistributedServerOperation(func() error {
				if phase == "direct" {
					thread.HandleRemoteWorkerLost(queue)
				} else {
					(&TLCTimerTask{Thread: thread}).Run()
				}
				return nil
			})
			if phase == "timer_alive" {
				if err != nil {
					t.Fatal("successful status call accessed missing coordinator", err)
				}
			} else if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("missing coordinator failure %T/%v", err, err)
			}
			claimed := phase == "direct"
			if thread.cleanupGlobals.Load() == claimed || thread.keepAliveStopped.Load() != claimed || NumWorkers() != 2 || queue.Size() != 0 || thread.GetCurrentSize() != len(states) {
				t.Fatal("missing coordinator changed cleanup ordering, assigned work or worker count")
			}
			for i, state := range states {
				if thread.currentStates()[i] != state {
					t.Fatal("missing coordinator replaced assigned state")
				}
			}
			if claimed {
				// The source flag is consumed even though removal failed. A
				// later report must not restart partial cleanup or lose work.
				thread.Server = &TLCServer{StateQueue: queue}
				thread.HandleRemoteWorkerLost(queue)
				if queue.Size() != 0 || thread.GetCurrentSize() != 2 || NumWorkers() != 2 {
					t.Fatal("duplicate report resumed failed one-time cleanup")
				}
			}
		})
	}
}
