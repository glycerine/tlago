package tlc

import (
	"sync"
	"testing"
	"time"
)

type workerLossWakeQueue struct {
	*MemStateQueue
	onWake func()
}

func (q *workerLossWakeQueue) WakeAllWaiters() {
	q.onWake()
	q.MemStateQueue.WakeAllWaiters()
}

func (q *workerLossWakeQueue) ResumeAllStuck() {
	panic("worker loss used checkpoint wakeup")
}

func TestDistributedWorkerLossWakePrecedesWorkerDecrement(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	SetNumWorkers(2)
	server := &TLCServer{}
	thread := &TLCServerThread{Server: server,
		Worker:        NewDistributedWorkerSmartProxy(&rpcTestWorker{}),
		keepAliveDone: make(chan struct{}), URI: "tcp://worker:1234/primary"}
	states := []*TLCStateMut{{UID: 7}, {UID: 8}}
	thread.setStates(states)
	thread.cleanupGlobals.Store(true)
	server.RegisterTLCServerThread(thread)
	wakes := 0
	queue := &workerLossWakeQueue{MemStateQueue: NewMemStateQueue()}
	queue.stop = true
	queue.onWake = func() {
		wakes++
		if NumWorkers() != 2 || len(thread.currentStates()) != 0 || len(server.GetServerThreads()) != 0 || queue.Size() != 2 || !thread.keepAliveStopped.Load() || thread.cleanupGlobals.Load() {
			t.Fatal("wakeup moved before cleanup/requeue or after worker decrement")
		}
	}
	thread.HandleRemoteWorkerLost(queue)
	thread.HandleRemoteWorkerLost(queue)
	if wakes != 1 || NumWorkers() != 1 || queue.Size() != 2 || !queue.stop {
		t.Fatal("duplicate worker loss changed wakeup, worker count or suspension")
	}
}

// Source worker loss notifies the queue monitor even while suspended. It does
// not use resumeAllStuck, whose suspended branch notifies checkpoint waiters.
func TestDistributedWorkerLossWakesSuspendedQueueConsumers(t *testing.T) {
	oldWorkers := NumWorkers()
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	for _, kind := range []string{"memory", "deque", "disk", "bytes"} {
		for _, block := range []string{"empty", "absent"} {
			t.Run(kind+"/"+block, func(t *testing.T) {
				SetNumWorkers(2)
				var queue StateQueue
				var mu *sync.Mutex
				var cond *sync.Cond
				var stopped *bool
				switch kind {
				case "memory":
					q := NewMemStateQueue()
					queue, mu, cond, stopped = q, &q.mu, q.cond, &q.stop
					q.numWaiting.Store(1)
				case "deque":
					q := NewStateDeque()
					queue, mu, cond, stopped = q, &q.mu, q.cond, &q.stop
					q.numWaiting.Store(1)
				case "disk":
					q := NewDiskStateQueue(t.TempDir())
					queue, mu, cond, stopped = q, &q.mu, q.cond, &q.stop
					q.numWaiting.Store(1)
				case "bytes":
					q := NewDiskByteArrayQueue(t.TempDir())
					queue, mu, cond, stopped = q, &q.mu, q.cond, &q.stop
					q.numWaiting.Store(1)
				}
				*stopped = true
				ready, woke := make(chan struct{}), make(chan struct{})
				go func() {
					mu.Lock()
					close(ready)
					cond.Wait()
					mu.Unlock()
					close(woke)
				}()
				<-ready
				t.Cleanup(func() {
					mu.Lock()
					cond.Broadcast()
					mu.Unlock()
					<-woke
				})
				thread := &TLCServerThread{Server: &TLCServer{}, keepAliveDone: make(chan struct{})}
				var states []*TLCStateMut
				if block == "empty" {
					states = []*TLCStateMut{}
				}
				thread.setStates(states)
				thread.cleanupGlobals.Store(true)
				thread.HandleRemoteWorkerLost(queue)
				select {
				case <-woke:
				case <-time.After(time.Second):
					t.Fatal("worker loss did not wake suspended queue consumer")
				}
				mu.Lock()
				stillStopped := *stopped
				mu.Unlock()
				if !stillStopped || queue.Size() != 0 || NumWorkers() != 1 || !thread.keepAliveStopped.Load() || thread.cleanupGlobals.Load() || thread.currentStates() == nil {
					t.Fatal("worker-loss wake resumed the queue or changed cleanup state")
				}
				thread.HandleRemoteWorkerLost(queue)
				if NumWorkers() != 1 || queue.Size() != 0 {
					t.Fatal("duplicate report repeated empty-work cleanup")
				}
			})
		}
	}
}
