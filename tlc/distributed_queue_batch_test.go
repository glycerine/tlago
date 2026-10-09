package tlc

import "testing"

// StateQueue and ByteArrayQueue dereference each supplied batch before
// enqueueing anything. No original queue test covers null batch arguments.
func TestDistributedQueueRejectsNilBatches(t *testing.T) {
	initTLCCheckerTest(t)
	for _, backend := range []string{"memory", "deque", "disk", "bytes"} {
		for _, batch := range []string{"array", "vector"} {
			t.Run(backend+"/"+batch, func(t *testing.T) {
				var queue StateQueue
				switch backend {
				case "memory":
					queue = NewMemStateQueue(t.TempDir())
				case "deque":
					queue = NewStateDeque()
				case "disk":
					queue = NewDiskStateQueue(t.TempDir())
				case "bytes":
					queue = NewDiskByteArrayQueue(t.TempDir())
				}
				t.Cleanup(queue.FinishAll)
				initial := checkerTestState(17)
				initial.UID = 17 // Disk state headers encode nonnegative trace locations.
				queue.SEnqueue(initial)
				var failure any
				func() {
					defer func() { failure = recover() }()
					if batch == "array" {
						queue.SEnqueueAll(nil)
					} else {
						queue.SEnqueueVec(nil)
					}
				}()
				if _, ok := failure.(error); !ok {
					t.Fatalf("nil batch must fail with a Go error, got %T: %v", failure, failure)
				}
				// These calls also verify the synchronized methods released the lock.
				if queue.Size() != 1 {
					t.Fatal("rejected batch changed queue length")
				}
				queue.SEnqueueAll(make([]*TLCStateMut, 0))
				queue.SEnqueueVec(NewStateVec(0))
				if queue.Size() != 1 {
					t.Fatal("empty batch changed queue length")
				}
				if got := queue.SDequeue(); got == nil || got.Lookup(UniqueStringOf("x")).(*IntValue).Val != 17 {
					t.Fatalf("rejected/empty batch changed the existing state: %v", got)
				}
				next := checkerTestState(23)
				next.UID = 23
				queue.SEnqueue(next)
				if queue.Size() != 1 || queue.SDequeue().Lookup(UniqueStringOf("x")).(*IntValue).Val != 23 || queue.Size() != 0 {
					t.Fatal("queue did not remain usable after rejecting the batch")
				}
			})
		}
	}
}
