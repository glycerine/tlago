package tlc

import (
	"testing"
	"time"
)

// No original method directly covers required owners of batch statistics.
func TestDistributedBatchStatisticsRequireOwnersInSourceOrder(t *testing.T) {
	for _, kind := range []string{"missing-timer", "missing-server", "healthy"} {
		t.Run(kind, func(t *testing.T) {
			server := &TLCServer{}
			server.WorkerStatesGenerated.Store(19)
			result := NewNextStateResult([]*StateVec{
				NewStateVecFrom([]*TLCStateMut{{UID: 1}, {UID: 2}, {UID: 3}}),
				NewStateVecFrom([]*TLCStateMut{{UID: 4}, {UID: 5}}),
			}, []*LongVec{NewLongVecFrom([]int64{1, 2, 3}), NewLongVecFrom([]int64{4, 5})}, 1, 11)
			thread := &TLCServerThread{Server: server, ReceivedStates: 7,
				Worker: NewDistributedWorkerSmartProxy(&rpcTestWorker{
					next: func([]*TLCStateMut) (*NextStateResult, error) { return result, nil },
				})}
			thread.setStates([]*TLCStateMut{{UID: 7}})
			if kind != "missing-timer" {
				thread.TimerTask = &TLCTimerTask{Thread: thread}
				thread.TimerTask.LastInvocation.Store(-1)
			}
			if kind == "missing-server" {
				thread.Server = nil
			}
			before := time.Now().UnixMilli()
			got, err := thread.computeBlockAttempt()
			after := time.Now().UnixMilli()
			if thread.ReceivedStates != 10 {
				t.Fatal("owner failure skipped preceding received-state mutation")
			}
			if kind == "healthy" {
				// Source delta subtracts the partition count, not state count.
				if err != nil || got != result || server.WorkerStatesGenerated.Load() != 28 {
					t.Fatalf("healthy batch statistics changed: %v, count %d", err, server.WorkerStatesGenerated.Load())
				}
			} else {
				if _, ok := err.(*NullPointerException); !ok || got != nil || server.WorkerStatesGenerated.Load() != 19 {
					t.Fatalf("missing owner suppressed failure or changed delta: %T/%v, result %v, count %d", err, err, got, server.WorkerStatesGenerated.Load())
				}
			}
			if thread.TimerTask != nil {
				stamp := thread.TimerTask.LastInvocation.Load()
				if stamp < before || stamp > after {
					t.Fatalf("timestamp skipped or moved after coordinator operation: %d outside [%d, %d]", stamp, before, after)
				}
			}
		})
	}
}

func TestDistributedStatisticsMethodsRejectMissingOwner(t *testing.T) {
	for _, call := range []func(){
		func() { (*TLCTimerTask)(nil).SetLastInvocation(time.UnixMilli(1)) },
		func() { (*TLCServer)(nil).AddStatesGeneratedDelta(0) },
		func() { (*TLCServer)(nil).AddStatesGeneratedDelta(7) },
	} {
		err := invokeDistributedServerOperation(func() error { call(); return nil })
		if _, ok := err.(*NullPointerException); !ok {
			t.Fatalf("missing statistics owner suppressed failure: %T/%v", err, err)
		}
	}
}

func TestDistributedBatchMissingTimerRequeuesAssignedWork(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	SetNumWorkers(2)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	defer RemoveMessageRecorder(recorder)
	queue := NewMemStateQueue()
	server := &TLCServer{StateQueue: queue}
	server.WorkerStatesGenerated.Store(19)
	states := []*TLCStateMut{{UID: 7}, {UID: 8}}
	thread := &TLCServerThread{Server: server, ReceivedStates: 7,
		URI: "tcp://worker:1234/primary", keepAliveDone: make(chan struct{}),
		Worker: NewDistributedWorkerSmartProxy(&rpcTestWorker{
			next: func([]*TLCStateMut) (*NextStateResult, error) {
				return NewNextStateResult([]*StateVec{NewStateVecFrom(states)}, []*LongVec{NewLongVecFrom([]int64{1, 2})}, 1, 5), nil
			},
		})}
	thread.setStates(states)
	thread.cleanupGlobals.Store(true)
	server.RegisterTLCServerThread(thread)
	result, proceed := thread.computeBlock(queue)
	if result != nil || proceed || thread.ReceivedStates != 9 || server.WorkerStatesGenerated.Load() != 19 {
		t.Fatal("missing timer returned a publishable block or changed statistics ordering")
	}
	if queue.Size() != 2 || queue.SDequeue() != states[0] || queue.SDequeue() != states[1] || len(server.GetServerThreads()) != 0 || NumWorkers() != 1 || len(thread.currentStates()) != 0 || !thread.keepAliveStopped.Load() {
		t.Fatal("missing timer skipped worker-loss deregistration or assigned-state recovery")
	}
	if len(recorder.Records(ECTLCDistributedWorkerLost)) != 1 || recorder.Recorded(ECGeneral) || server.IsDone() {
		t.Fatal("statistics null failure crossed the inner catch into model failure")
	}
}
