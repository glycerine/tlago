package tlc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Complete mechanical translation of tlc2.tool.queue.StateQueueTest.
// Queue operations use DummyTLCState only as an opaque identity. A fresh,
// zero-valued Go state supplies the same identity without unused Java stubs.
func javaStateQueueSetup(t *testing.T) *MemStateQueue {
	t.Helper()
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })
	return NewMemStateQueue("")
}

func requireJavaStateQueueSize(t *testing.T, q *MemStateQueue, expected int64) {
	t.Helper()
	if actual := q.Size(); actual != expected {
		t.Fatalf("queue size = %d, want %d", actual, expected)
	}
}

func TestJavaStateQueueEnqueue(t *testing.T) {
	q := javaStateQueueSetup(t)
	expected := &TLCStateMut{}
	q.Enqueue(expected)
	if actual := q.SDequeue(); actual != expected {
		t.Fatalf("sDequeue = %p, want %p", actual, expected)
	}
}

func TestJavaStateQueueSDequeueEmpty(t *testing.T) {
	q := javaStateQueueSetup(t)
	if state := q.SDequeue(); state != nil {
		t.Fatalf("sDequeue = %p, want nil", state)
	}
}

func TestJavaStateQueueDequeueEmpty(t *testing.T) {
	q := javaStateQueueSetup(t)
	if state := q.Dequeue(); state != nil {
		t.Fatalf("dequeue = %p, want nil", state)
	}
}

func TestJavaStateQueueSDequeueNotEmpty(t *testing.T) {
	q := javaStateQueueSetup(t)
	expected := &TLCStateMut{}
	q.SEnqueue(expected)
	requireJavaStateQueueSize(t, q, 1)
	actual := q.SDequeue()
	requireJavaStateQueueSize(t, q, 0)
	if actual != expected {
		t.Fatalf("sDequeue = %p, want %p", actual, expected)
	}
}

func TestJavaStateQueueDequeueNotEmpty(t *testing.T) {
	q := javaStateQueueSetup(t)
	expected := &TLCStateMut{}
	q.Enqueue(expected)
	requireJavaStateQueueSize(t, q, 1)
	actual := q.Dequeue()
	requireJavaStateQueueSize(t, q, 0)
	if actual != expected {
		t.Fatalf("dequeue = %p, want %p", actual, expected)
	}
}

func TestJavaStateQueueEnqueueAddNotSame(t *testing.T) {
	q := javaStateQueueSetup(t)
	const j = 10
	for i := 0; i < j; i++ {
		q.SEnqueue(&TLCStateMut{})
	}
	requireJavaStateQueueSize(t, q, j)
}

func TestJavaStateQueueEnqueueAddSame(t *testing.T) {
	q := javaStateQueueSetup(t)
	state := &TLCStateMut{}
	const j = 10
	for i := 0; i < j; i++ {
		q.SEnqueue(state)
	}
	requireJavaStateQueueSize(t, q, j)
}

func TestJavaStateQueueSDequeueAbuseEmpty(t *testing.T) {
	q := javaStateQueueSetup(t)
	expectPanic(t, func() { q.SDequeueMany(0) })
	expectPanic(t, func() { q.SDequeueMany(-1) })
	expectPanic(t, func() { q.SDequeueMany(-2147483648) })
	if actual := q.SDequeueMany(2147483647); actual != nil {
		t.Fatalf("sDequeue(Integer.MAX_VALUE) = %v, want nil", actual)
	}
}

func TestJavaStateQueueSDequeueAbuseNonEmpty(t *testing.T) {
	q := javaStateQueueSetup(t)
	q.SEnqueue(&TLCStateMut{})
	expectPanic(t, func() { q.SDequeueMany(0) })
	expectPanic(t, func() { q.SDequeueMany(-1) })
	expectPanic(t, func() { q.SDequeueMany(-2147483648) })
	if actual := q.SDequeueMany(2147483647); len(actual) != 1 {
		t.Fatalf("sDequeue(Integer.MAX_VALUE).length = %d, want 1", len(actual))
	}
}

func TestStateDequeSDequeueManyUpdatesSizeLikeJavaStateQueue(t *testing.T) {
	initTLCCheckerTest(t)
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })

	q := NewStateDeque()
	first := checkerTestState(1)
	second := checkerTestState(2)
	q.SEnqueue(first)
	q.SEnqueue(second)
	got := q.SDequeueMany(2)
	if len(got) != 2 || got[0] != second || got[1] != first {
		t.Fatalf("SDequeueMany = %#v, want LIFO batch", got)
	}
	if q.Size() != 0 {
		t.Fatalf("size after StateDeque SDequeueMany = %d, want 0", q.Size())
	}
}

func TestMemStateQueueSEnqueueVecSkipsNilStatesLikeJavaStateVecPath(t *testing.T) {
	initTLCCheckerTest(t)
	q := NewMemStateQueue()
	first := checkerTestState(1)
	second := checkerTestState(2)
	q.SEnqueueVec(NewStateVecFrom([]*TLCStateMut{first, nil, second}))

	if q.Size() != 2 {
		t.Fatalf("size after SEnqueueVec with nil = %d, want 2", q.Size())
	}
	if got := q.Dequeue(); got != first {
		t.Fatalf("first dequeue = %p, want first state %p", got, first)
	}
	if got := q.Dequeue(); got != second {
		t.Fatalf("second dequeue = %p, want second state %p", got, second)
	}
}

func TestMemStateQueueCommitCheckpointErrorUsesJavaText(t *testing.T) {
	q := NewMemStateQueue(t.TempDir())
	err := q.CommitChkpt()
	if err == nil {
		t.Fatalf("CommitChkpt returned nil, want Java-shaped missing tmp rename error")
	}
	want := "MemStateQueue.commitChkpt: cannot delete " + filepath.Join(q.diskdir, "queue.chkpt")
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("CommitChkpt error = %q, want to contain %q", err.Error(), want)
	}
}

func TestDiskByteArrayQueueSEnqueueVecUsesJavaRawSlotLayout(t *testing.T) {
	initTLCCheckerTest(t)
	q := &DiskByteArrayQueue{
		deqBuf:   make([][]byte, 4),
		enqBuf:   make([][]byte, 4),
		deqIndex: 4,
	}
	q.cond = sync.NewCond(&q.mu)

	first := checkerTestState(1)
	second := checkerTestState(2)
	first.UID = 1
	second.UID = 2
	q.SEnqueueVec(NewStateVecFrom([]*TLCStateMut{first, nil, second}))

	if q.len != 3 || q.enqIndex != 3 {
		t.Fatalf("raw queue len/enqIndex = %d/%d, want 3/3", q.len, q.enqIndex)
	}
	if q.enqBuf[0] != nil {
		t.Fatalf("raw slot 0 = %v, want Java nil hole for skipped StateVec element", q.enqBuf[0])
	}
	if got := stateTestXValue(mustBytesToState(q.enqBuf[1])); got != 2 {
		t.Fatalf("raw slot 1 decoded x = %d, want second StateVec state", got)
	}
	if got := stateTestXValue(mustBytesToState(q.enqBuf[2])); got != 1 {
		t.Fatalf("raw slot 2 decoded x = %d, want first StateVec state", got)
	}
}

func stateTestXValue(state *TLCStateMut) int32 {
	if state == nil {
		return 0
	}
	value, _ := state.Lookup(UniqueStringOf("x")).(*IntValue)
	if value == nil {
		return 0
	}
	return value.Val
}

func TestDiskQueueDeleteUsesJavaNonRecursiveDirectoryDelete(t *testing.T) {
	deleteStateQueue := func(dir string) error {
		q := &DiskStateQueue{diskdir: dir}
		q.cond = sync.NewCond(&q.mu)
		return q.Delete()
	}
	deleteByteArrayQueue := func(dir string) error {
		q := &DiskByteArrayQueue{diskdir: dir}
		q.cond = sync.NewCond(&q.mu)
		return q.Delete()
	}

	for _, tc := range []struct {
		name   string
		delete func(string) error
	}{
		{name: "DiskStateQueue", delete: deleteStateQueue},
		{name: "DiskByteArrayQueue", delete: deleteByteArrayQueue},
	} {
		t.Run(tc.name+"/empty", func(t *testing.T) {
			dir := t.TempDir()
			if err := tc.delete(dir); err != nil {
				t.Fatalf("Delete empty queue dir returned error: %v", err)
			}
			if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("empty queue dir exists/error = %v, want Java File.delete-style removal", err)
			}
		})

		t.Run(tc.name+"/non-empty", func(t *testing.T) {
			dir := t.TempDir()
			artifact := filepath.Join(dir, "queue-artifact")
			if err := os.WriteFile(artifact, []byte("preserve"), 0o644); err != nil {
				t.Fatalf("WriteFile artifact: %v", err)
			}
			if err := tc.delete(dir); err != nil {
				t.Fatalf("Delete non-empty queue dir returned error: %v", err)
			}
			if _, err := os.Stat(artifact); err != nil {
				t.Fatalf("Delete removed non-empty queue dir recursively; Java File.delete preserves it: %v", err)
			}
		})
	}
}

func expectPanic(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	f()
}
