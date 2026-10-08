package tlc

import "testing"

type retryFailureQueue struct {
	*MemStateQueue
	failure error
}

func (q *retryFailureQueue) SEnqueueAll(states []*TLCStateMut) {
	q.SEnqueue(states[0])
	panic(q.failure)
}

// Upstream has no direct retry-queue tests. Its catch body requires requeueing
// before reducing the transfer limit, and a queue failure escapes that catch.
func TestDistributedRetryRequiresQueueBeforeLimitUpdate(t *testing.T) {
	for _, kind := range []string{"absent", "io", "runtime", "fatal", "healthy"} {
		t.Run(kind, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			defer RemoveMessageRecorder(recorder)
			selector := NewLimitingBlockSelector(&TLCServer{}, 100)
			remote := localWorkerRemoteException(NewRemoteException(javaString("memory"), NewOutOfMemoryError()))
			thread := &TLCServerThread{Selector: selector, Worker: NewDistributedWorkerSmartProxy(&rpcTestWorker{
				next: func([]*TLCStateMut) (*NextStateResult, error) { return nil, remote },
			})}
			thread.cleanupGlobals.Store(true)
			states := []*TLCStateMut{{UID: 1}, {UID: 2}}
			thread.setStates(states)
			memoryQueue := NewMemStateQueue()
			var queue StateQueue
			var failure error
			switch kind {
			case "io":
				failure = NewIOException("requeue failed")
			case "runtime":
				failure = NewRuntimeException("requeue failed")
			case "fatal":
				failure = NewOutOfMemoryError("requeue failed")
			case "healthy":
				queue = memoryQueue
			}
			if failure != nil {
				queue = &retryFailureQueue{MemStateQueue: memoryQueue, failure: failure}
			}
			var result *NextStateResult
			var proceed bool
			err := invokeDistributedServerOperation(func() error { result, proceed = thread.computeBlock(queue); return nil })
			if kind == "healthy" {
				if err != nil || result != nil || !proceed || selector.getMaximum() != 1 || memoryQueue.Size() != 2 || memoryQueue.SDequeue() != states[0] || memoryQueue.SDequeue() != states[1] {
					t.Fatalf("healthy retry changed: %v/%v, error %v", result, proceed, err)
				}
			} else {
				if kind == "absent" {
					if _, ok := err.(*NullPointerException); !ok {
						t.Fatalf("missing queue suppressed failure: %T/%v", err, err)
					}
				} else if err != failure || memoryQueue.Size() != 1 || memoryQueue.SDequeue() != states[0] {
					t.Fatalf("requeue failure changed identity or preceding mutation: %T/%v", err, err)
				}
				if proceed || result != nil || selector.getMaximum() != 100 {
					t.Fatal("failed requeue continued or reduced the transfer limit")
				}
			}
			if !thread.cleanupGlobals.Load() || thread.currentStates()[0] != states[0] || thread.currentStates()[1] != states[1] {
				t.Fatal("retry queue failure entered worker-loss cleanup or cleared assigned work")
			}
			if len(recorder.Records(ECTLCDistributedExceedBlocksize)) != 1 || recorder.Recorded(ECTLCDistributedWorkerLost) || recorder.Recorded(ECGeneral) {
				t.Fatal("retry queue operation changed preceding diagnostic or nested catch boundary")
			}
		})
	}
}
