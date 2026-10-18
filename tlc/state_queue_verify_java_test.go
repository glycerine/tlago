//go:build tlc_verify

package tlc

import (
	"sync"
	"testing"
)

// Original StateQueueJPFTest.test workload, using native memory storage.
// Explore the bounded source monitor model, then execute the original workers
// and suspension task using actual native synchronization.
func TestJavaStateQueueVerification(t *testing.T) {
	verifyJavaStateQueueSchedules(t)
	const workers = 2
	previousWorkers := NumWorkers()
	SetNumWorkers(workers)
	t.Cleanup(func() { SetNumWorkers(previousWorkers) })

	// The source Dummy stores only the last pointer. This workload repeatedly
	// enqueues the same pointer, so native FIFO storage has the same payload
	// observations. No state fields or checkpoint operations are consulted.
	queue := NewMemStateQueue()
	state := NewEmptyState()
	state.UID = 0 // Original no-argument DummyTLCState constructor.
	previousEmpty := EmptyState
	EmptyState = state
	t.Cleanup(func() { EmptyState = previousEmpty })
	queue.Enqueue(state)

	mainDone := make(chan struct{})
	go func() {
		defer close(mainDone)
		// Original MainTask is started before either worker.
		if queue.SuspendAll() {
			queue.ResumeAll()
		}
	}()
	var workerThreads sync.WaitGroup
	for i := 0; i < workers; i++ {
		workerThreads.Add(1)
		go func() {
			defer workerThreads.Done()
			// Original WorkerTask(queue, state) defaults to three iterations.
			for iteration := 0; iteration < 3; iteration++ {
				if queue.SDequeue() == nil {
					queue.FinishAll()
					return
				}
				queue.SEnqueueAll([]*TLCStateMut{state})
			}
			queue.FinishAll()
		}()
	}
	// Preserve mainThread.join() followed by both worker joins.
	<-mainDone
	workerThreads.Wait()
}
