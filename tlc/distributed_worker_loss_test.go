package tlc

import (
	"sync"
	"testing"
)

// Source handleRemoteWorkerLost explicitly requires cleanup to be idempotent:
// keepalive and the failed RPC can report the same lost worker concurrently.
// Upstream has no enabled direct test of that native concurrency boundary.
func TestDistributedConcurrentWorkerLossRequeuesOnce(t *testing.T) {
	oldWorkers := NumWorkers()
	defer SetNumWorkers(oldWorkers)
	for iteration := 0; iteration < 32; iteration++ {
		SetNumWorkers(2)
		queue := NewDiskStateQueue(t.TempDir())
		t.Cleanup(queue.FinishAll)
		states := []*TLCStateMut{{UID: 11, level: 2}, {UID: 22, level: 3}}
		thread := &TLCServerThread{Server: &TLCServer{StateQueue: queue}, keepAliveDone: make(chan struct{})}
		thread.cleanupGlobals.Store(true)
		thread.setStates(states)
		start := make(chan struct{})
		var reports sync.WaitGroup
		reports.Add(2)
		for reporter := 0; reporter < 2; reporter++ {
			go func() {
				defer reports.Done()
				<-start
				thread.HandleRemoteWorkerLost(queue)
			}()
		}
		close(start)
		reports.Wait()
		if NumWorkers() != 1 || thread.GetCurrentSize() != 0 || queue.Size() != 2 {
			t.Fatalf("duplicate cleanup: workers %d, assigned %d, queued %d", NumWorkers(), thread.GetCurrentSize(), queue.Size())
		}
		if queue.SDequeue() != states[0] || queue.SDequeue() != states[1] {
			t.Fatal("failed block was lost, duplicated or reordered")
		}
	}
}
