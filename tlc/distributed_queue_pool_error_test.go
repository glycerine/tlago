package tlc

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// StateQueue.sEnqueue and ByteArrayQueue.sEnqueue publish len after the entire
// batch. No original queue test exercises a spill failure after a batch prefix.
func TestDistributedBulkQueueSpillFailureRetainsUncountedPrefix(t *testing.T) {
	initTLCCheckerTest(t)
	for _, backend := range []string{"states", "bytes"} {
		for _, batch := range []string{"array", "vector"} {
			t.Run(backend+"/"+batch, func(t *testing.T) {
				// An existing directory cannot be opened as a pool output file.
				pending := t.TempDir()
				initial, first, second := checkerTestState(1), checkerTestState(2), checkerTestState(3)
				states := []*TLCStateMut{first, second}
				var queue StateQueue
				var checkPrefix func()
				if backend == "states" {
					q := &DiskStateQueue{diskdir: t.TempDir(), len: 1, hiPool: 3, enqIndex: 1, enqBuf: []*TLCStateMut{initial, nil}}
					q.writer = NewStatePoolWriter(2, nil)
					if _, err := q.writer.DoWork([]*TLCStateMut{initial, initial}, pending); err != nil {
						t.Fatal(err)
					}
					queue = q
					checkPrefix = func() {
						if q.enqIndex != 2 || q.hiPool != 3 || q.enqBuf[0] != initial || q.enqBuf[1] != first {
							t.Fatal("failed spill changed existing states or discarded the enqueued prefix")
						}
					}
				} else {
					raw := mustStateToBytes(initial)
					q := &DiskByteArrayQueue{diskdir: t.TempDir(), len: 1, hiPool: 3, enqIndex: 1, enqBuf: [][]byte{raw, nil}}
					q.writer = NewByteArrayPoolWriter(2, nil)
					if _, err := q.writer.DoWork([][]byte{raw, raw}, pending); err != nil {
						t.Fatal(err)
					}
					queue = q
					checkPrefix = func() {
						// ByteArrayQueue reverses StateVec entries before enqueuing.
						prefix := first
						if batch == "vector" {
							prefix = second
						}
						if q.enqIndex != 2 || q.hiPool != 3 || !bytes.Equal(q.enqBuf[0], raw) || !bytes.Equal(q.enqBuf[1], mustStateToBytes(prefix)) {
							t.Fatal("failed spill changed existing bytes or discarded the enqueued prefix")
						}
					}
				}
				var failure any
				func() {
					defer func() { failure = recover() }()
					if batch == "array" {
						queue.SEnqueueAll(states)
					} else {
						queue.SEnqueueVec(NewStateVecFrom(states))
					}
				}()
				if failure == nil || !strings.Contains(panicValueAsError(failure).Error(), pending) {
					t.Fatalf("expected pending pool write failure, got %v", failure)
				}
				checkPrefix()
				if got := queue.Size(); got != 1 {
					t.Fatalf("failed batch published prefix count: size = %d, want unchanged 1", got)
				}
			})
		}
	}
}

// No enabled original method tests the queue's synchronous pool-error catches.
func TestDistributedDiskQueuePoolFailureUsesCodedRuntimeError(t *testing.T) {
	for _, operation := range []string{"write", "read", "peek"} {
		t.Run(operation, func(t *testing.T) {
			name := filepath.Join(t.TempDir(), "missing", "0")
			state := &TLCStateMut{UID: 41}
			queue := &DiskStateQueue{diskdir: filepath.Dir(name), len: 1, loPool: 1, hiPool: 3, loFile: name, deqIndex: 1, deqBuf: []*TLCStateMut{state}, enqIndex: 1, enqBuf: []*TLCStateMut{state}}
			queue.reader = NewStatePoolReader(1, name)
			queue.writer = NewStatePoolWriter(1, nil)
			if operation == "write" {
				if _, err := queue.writer.DoWork([]*TLCStateMut{state}, name); err != nil {
					t.Fatal(err)
				}
			}
			var failure any
			func() {
				defer func() { failure = recover() }()
				switch operation {
				case "write":
					queue.Enqueue(&TLCStateMut{UID: 97})
				case "read":
					queue.Dequeue()
				case "peek":
					queue.peekInner()
				}
			}()
			code := ECSystemErrorReadingStates
			if operation == "write" {
				code = ECSystemErrorWritingStates
			}
			coded, ok := failure.(*TLCError)
			if !ok || !coded.Runtime || coded.Code != code || len(coded.Params) != 2 || coded.Params[0] != "queue" || !strings.Contains(coded.Params[1], name) || coded.Cause != nil {
				t.Fatalf("queue %s failure lost source assertion contract: %T %v", operation, failure, failure)
			}
			if queue.len != 1 || queue.enqIndex != 1 || queue.deqIndex != 1 || queue.hiPool != 3 || queue.loPool != 1 || queue.enqBuf[0] != state || queue.deqBuf[0] != state {
				t.Fatal("failed pool operation advanced queue or replaced existing states")
			}
		})
	}
}
