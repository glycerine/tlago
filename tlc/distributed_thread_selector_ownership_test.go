package tlc

import (
	"math"
	"testing"
)

// Upstream has no direct tests for the constructor's selector ownership.
func TestDistributedThreadRetainsSuppliedSelector(t *testing.T) {
	server := &TLCServer{}
	server.BlockSelector = NewStaticBlockSelector(server, 3)
	for _, selector := range []*BlockSelector{nil, NewStaticBlockSelector(server, 7)} {
		thread := NewTLCServerThread(&rpcTestWorker{}, "tcp://worker:1234/primary", server, selector)
		thread.cancelKeepAlive()
		if thread.Selector != selector {
			t.Fatal("constructor substituted the supplied selector")
		}
	}
}

func TestDistributedThreadMissingSelectorRunsErrorHandlerAndFinally(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	memoryQueue := NewMemStateQueue()
	queue := &distributedSelectorQueue{StateQueue: memoryQueue}
	server := &TLCServer{StateQueue: queue}
	server.BlockSelector = NewStaticBlockSelector(server, 3)
	waiter := make(chan struct{})
	server.completionWaiter = waiter
	thread := &TLCServerThread{Server: server,
		Worker:            NewDistributedWorkerSmartProxy(&rpcTestWorker{}),
		CacheRateHitRatio: -1, keepAliveDone: make(chan struct{}),
	}
	thread.setStates([]*TLCStateMut{{UID: 7}})
	err := invokeDistributedServerOperation(func() error { thread.Run(); return nil })
	if _, ok := server.LastError.(*NullPointerException); err != nil || !ok || !server.IsDone() || !server.KeepCallStack {
		t.Fatalf("missing selector suppressed model failure: %v, recorded %T/%v", err, server.LastError, server.LastError)
	}
	if thread.Selector != nil || queue.requested != 0 || !memoryQueue.finish.Load() {
		t.Fatal("missing selector borrowed server policy or skipped queue shutdown")
	}
	select {
	case <-waiter:
	default:
		t.Fatal("selector failure skipped completion notification")
	}
	if !math.IsNaN(thread.CacheRateHitRatio) || !thread.keepAliveStopped.Load() || thread.currentStates() == nil || len(thread.currentStates()) != 0 || NumWorkers() != oldWorkers+1 {
		t.Fatal("selector failure changed finally or worker accounting")
	}
}

func TestDistributedThreadRetryRequiresSelectorAfterRequeue(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	failure := localWorkerRemoteException(NewRemoteException(javaString("memory"), NewOutOfMemoryError()))
	thread := &TLCServerThread{Worker: NewDistributedWorkerSmartProxy(&rpcTestWorker{
		next: func([]*TLCStateMut) (*NextStateResult, error) { return nil, failure },
	})}
	states := []*TLCStateMut{{UID: 1}, {UID: 2}}
	thread.setStates(states)
	queue := NewMemStateQueue()
	err := invokeDistributedServerOperation(func() error { thread.computeBlock(queue); return nil })
	if _, ok := err.(*NullPointerException); !ok {
		t.Fatalf("missing retry selector suppressed failure: %T/%v", err, err)
	}
	if queue.Size() != 2 || queue.SDequeue() != states[0] || queue.SDequeue() != states[1] || thread.currentStates()[0] != states[0] {
		t.Fatal("selector failure changed preceding requeue or assigned work")
	}
}
