package tlc

import (
	"sync"
	"testing"
)

// ByteArrayQueue removes raw work before decoding it outside the monitor.
// No original queue method covers a truncated in-memory state in a batch.
func TestDistributedByteQueueDecodeFailureRetainsRawRemoval(t *testing.T) {
	initTLCCheckerTest(t)
	for _, operation := range []string{"peek", "single", "batch"} {
		t.Run(operation, func(t *testing.T) {
			second, third := checkerTestState(2), checkerTestState(3)
			second.UID, third.UID = 2, 3
			q := &DiskByteArrayQueue{len: 3,
				deqBuf: [][]byte{make([]byte, 0), mustStateToBytes(second), mustStateToBytes(third)}}
			q.cond = sync.NewCond(&q.mu)
			var failure any
			func() {
				defer func() { failure = recover() }()
				switch operation {
				case "peek":
					q.SPeek()
				case "single":
					q.SDequeue()
				case "batch":
					q.SDequeueMany(2)
				}
			}()
			if failure == nil {
				t.Fatal("truncated byte state did not fail decoding")
			}
			removed := map[string]int{"peek": 0, "single": 1, "batch": 2}[operation]
			if q.Size() != int64(3-removed) || q.deqIndex != removed {
				t.Fatalf("decode failure removed %d states, size %d; want %d removed", q.deqIndex, q.Size(), removed)
			}
			for i := 0; i < removed; i++ {
				if q.deqBuf[i] != nil {
					t.Fatal("raw dequeue retained a removed slot")
				}
			}
			if !q.mu.TryLock() {
				t.Fatal("decode failure retained the queue lock")
			}
			q.mu.Unlock()
			if operation != "peek" {
				if got := stateTestXValue(q.SDequeue()); got != int32(removed+1) {
					t.Fatalf("remaining state = %d, want %d", got, removed+1)
				}
			}
		})
	}
}

func TestDistributedByteQueueNullRawEntryBoundaries(t *testing.T) {
	for _, operation := range []string{"peek", "dequeue", "synchronized_dequeue"} {
		t.Run(operation, func(t *testing.T) {
			q := &DiskByteArrayQueue{len: 1, deqBuf: make([][]byte, 1)}
			q.cond = sync.NewCond(&q.mu)
			var state *TLCStateMut
			var failure any
			func() {
				defer func() { failure = recover() }()
				switch operation {
				case "peek":
					state = q.SPeek()
				case "dequeue":
					state = q.Dequeue()
				case "synchronized_dequeue":
					state = q.SDequeue()
				}
			}()
			if state != nil {
				t.Fatal("null raw entry became a state")
			}
			if operation == "synchronized_dequeue" {
				if _, ok := failure.(error); !ok || q.Size() != 1 || q.deqIndex != 1 {
					t.Fatalf("null entry assertion boundary: failure %v, size %d, index %d", failure, q.Size(), q.deqIndex)
				}
			} else {
				removed := 0
				if operation == "dequeue" {
					removed = 1
				}
				if failure != nil || q.Size() != int64(1-removed) || q.deqIndex != removed {
					t.Fatalf("null raw entry boundary: failure %v, size %d, index %d", failure, q.Size(), q.deqIndex)
				}
			}
		})
	}
}

func TestDistributedByteQueueBatchDecodeUsesRequestedCount(t *testing.T) {
	initTLCCheckerTest(t)
	state := checkerTestState(5)
	state.UID = 5
	for _, count := range []int{1, 2} {
		t.Run(intToDecimal(count), func(t *testing.T) {
			q := &DiskByteArrayQueue{len: 1, deqBuf: [][]byte{mustStateToBytes(state)}}
			q.cond = sync.NewCond(&q.mu)
			var states []*TLCStateMut
			var failure any
			func() {
				defer func() { failure = recover() }()
				states = q.SDequeueMany(count)
			}()
			if q.Size() != 0 || q.deqIndex != 1 || q.deqBuf[0] != nil {
				t.Fatal("raw batch was not completely removed before decoding")
			}
			if count == 1 {
				if failure != nil || len(states) != 1 || stateTestXValue(states[0]) != 5 {
					t.Fatalf("available batch: %v/%v", states, failure)
				}
			} else if failure == nil || states != nil {
				t.Fatal("source wrapper's oversized request must fail after raw removal")
			}
		})
	}
}
