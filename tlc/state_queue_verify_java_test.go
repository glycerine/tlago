//go:build tlc_verify

/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
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
