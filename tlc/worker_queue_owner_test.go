package tlc

import "testing"

type workerOwnerQueue struct {
	StateQueue
	dequeued, finished int
	failure            error
}

func (q *workerOwnerQueue) SDequeue() *TLCStateMut {
	q.dequeued++
	if q.failure != nil {
		panic(q.failure)
	}
	return nil
}
func (q *workerOwnerQueue) FinishAll() { q.finished++ }

// No original method directly exercises queue replacement after construction.
func TestWorkerRunUsesCapturedQueue(t *testing.T) {
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, scenario := range []string{"replacement", "missing-replacement", "dequeue-failure"} {
		t.Run(scenario, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			directory := t.TempDir()
			queue := &workerOwnerQueue{StateQueue: NewMemStateQueue(directory)}
			checker := NewModelChecker(NewTool(), directory, false, func(c *ModelChecker) { c.StateQueue = queue })
			defer checker.ConcurrentTrace.Close()
			replacement := &workerOwnerQueue{StateQueue: NewMemStateQueue()}
			checker.StateQueue = replacement
			if scenario == "missing-replacement" {
				checker.StateQueue = nil
			}
			if scenario == "dequeue-failure" {
				queue.failure = NewRuntimeException("dequeue failed")
			}
			err := checker.Workers[0].Run()
			if err != queue.failure || queue.dequeued != 1 || queue.finished != 1 || replacement.dequeued != 0 || replacement.finished != 0 {
				t.Fatalf("worker switched queue owner: err %v, original %d/%d, replacement %d/%d", err, queue.dequeued, queue.finished, replacement.dequeued, replacement.finished)
			}
			if !checker.Done {
				t.Fatal("worker did not publish completion/error")
			}
		})
	}
}

func TestWorkerSuccessorPublicationUsesExecutingWorkerAndQueue(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	defer SetNumWorkers(oldWorkers)
	for _, missingSlot := range []bool{false, true} {
		name := "retained-slot"
		if missingSlot {
			name = "missing-slot"
		}
		t.Run(name, func(t *testing.T) {
			checker := NewModelChecker(NewTool(), t.TempDir(), false)
			defer checker.ConcurrentTrace.Close()
			worker, queue := checker.Workers[0], checker.StateQueue
			replacement := NewMemStateQueue()
			checker.StateQueue = replacement
			if missingSlot {
				checker.Workers[0] = nil
			}
			parent, successor := checkerTestState(1), checkerTestState(2)
			parent.UID = 37
			_, err := worker.AddNextElement(parent, &Action{Name: "Next"}, successor)
			if err != nil || queue.Size() != 1 || queue.SPeek() != successor || replacement.Size() != 0 || worker.UnseenSuccessorStates != 1 {
				t.Fatalf("successor publication switched owner: %v, queue %d, replacement %d, writes %d", err, queue.Size(), replacement.Size(), worker.UnseenSuccessorStates)
			}
		})
	}
}
