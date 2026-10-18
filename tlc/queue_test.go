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

func requireJavaStateQueueSize(t *testing.T, q StateQueue, expected int64) {
	t.Helper()
	if actual := q.Size(); actual != expected {
		t.Fatalf("queue size = %d, want %d", actual, expected)
	}
}

func TestJavaStateQueueEnqueue(t *testing.T) {
	javaStateQueueOriginalMethod(t, javaStateQueueSetup(t), "testEnqueue")
}

func TestJavaStateQueueSDequeueEmpty(t *testing.T) {
	javaStateQueueOriginalMethod(t, javaStateQueueSetup(t), "testsDequeueEmpty")
}

func TestJavaStateQueueDequeueEmpty(t *testing.T) {
	javaStateQueueOriginalMethod(t, javaStateQueueSetup(t), "testDequeueEmpty")
}

func TestJavaStateQueueSDequeueNotEmpty(t *testing.T) {
	javaStateQueueOriginalMethod(t, javaStateQueueSetup(t), "testsDequeueNotEmpty")
}

func TestJavaStateQueueDequeueNotEmpty(t *testing.T) {
	javaStateQueueOriginalMethod(t, javaStateQueueSetup(t), "testDequeueNotEmpty")
}

func TestJavaStateQueueEnqueueAddNotSame(t *testing.T) {
	javaStateQueueOriginalMethod(t, javaStateQueueSetup(t), "testEnqueueAddNotSame")
}

func TestJavaStateQueueEnqueueAddSame(t *testing.T) {
	javaStateQueueOriginalMethod(t, javaStateQueueSetup(t), "testEnqueueAddSame")
}

func TestJavaStateQueueSDequeueAbuseEmpty(t *testing.T) {
	javaStateQueueOriginalMethod(t, javaStateQueueSetup(t), "testsDequeueAbuseEmpty")
}

func TestJavaStateQueueSDequeueAbuseNonEmpty(t *testing.T) {
	javaStateQueueOriginalMethod(t, javaStateQueueSetup(t), "testsDequeueAbuseNonEmpty")
}

var javaStateQueueOriginalMethods = []string{
	"testEnqueue",
	"testsDequeueEmpty",
	"testDequeueEmpty",
	"testsDequeueNotEmpty",
	"testDequeueNotEmpty",
	"testEnqueueAddNotSame",
	"testEnqueueAddSame",
	"testsDequeueAbuseEmpty",
	"testsDequeueAbuseNonEmpty",
}

// Whole original StateQueueTest methods, shared by its concrete subclasses.
func javaStateQueueOriginalMethod(t *testing.T, q StateQueue, method string) {
	t.Helper()
	switch method {
	case "testEnqueue":
		expected := &TLCStateMut{}
		q.Enqueue(expected)
		if actual := q.SDequeue(); actual != expected {
			t.Fatalf("sDequeue = %p, want %p", actual, expected)
		}
	case "testsDequeueEmpty":
		if state := q.SDequeue(); state != nil {
			t.Fatalf("sDequeue = %p, want nil", state)
		}
	case "testDequeueEmpty":
		if state := q.Dequeue(); state != nil {
			t.Fatalf("dequeue = %p, want nil", state)
		}
	case "testsDequeueNotEmpty":
		expected := &TLCStateMut{}
		q.SEnqueue(expected)
		requireJavaStateQueueSize(t, q, 1)
		actual := q.SDequeue()
		requireJavaStateQueueSize(t, q, 0)
		if actual != expected {
			t.Fatalf("sDequeue = %p, want %p", actual, expected)
		}
	case "testDequeueNotEmpty":
		expected := &TLCStateMut{}
		q.Enqueue(expected)
		requireJavaStateQueueSize(t, q, 1)
		actual := q.Dequeue()
		requireJavaStateQueueSize(t, q, 0)
		if actual != expected {
			t.Fatalf("dequeue = %p, want %p", actual, expected)
		}
	case "testEnqueueAddNotSame":
		const j = 10
		for i := 0; i < j; i++ {
			q.SEnqueue(&TLCStateMut{})
		}
		requireJavaStateQueueSize(t, q, j)
	case "testEnqueueAddSame":
		state := &TLCStateMut{}
		const j = 10
		for i := 0; i < j; i++ {
			q.SEnqueue(state)
		}
		requireJavaStateQueueSize(t, q, j)
	case "testsDequeueAbuseEmpty":
		javaStateQueueExpectRuntimeOrAssertion(t, func() { q.SDequeueMany(0) })
		javaStateQueueExpectRuntimeOrAssertion(t, func() { q.SDequeueMany(-1) })
		javaStateQueueExpectRuntimeOrAssertion(t, func() { q.SDequeueMany(-2147483648) })
		if actual := q.SDequeueMany(2147483647); actual != nil {
			t.Fatalf("sDequeue(Integer.MAX_VALUE) = %v, want nil", actual)
		}
	case "testsDequeueAbuseNonEmpty":
		q.SEnqueue(&TLCStateMut{})
		javaStateQueueExpectRuntimeOrAssertion(t, func() { q.SDequeueMany(0) })
		javaStateQueueExpectRuntimeOrAssertion(t, func() { q.SDequeueMany(-1) })
		javaStateQueueExpectRuntimeOrAssertion(t, func() { q.SDequeueMany(-2147483648) })
		if actual := q.SDequeueMany(2147483647); len(actual) != 1 {
			t.Fatalf("sDequeue(Integer.MAX_VALUE).length = %d, want 1", len(actual))
		}
	default:
		t.Fatalf("unknown original queue method %s", method)
	}
}

// Original expectRuntimeException catches RuntimeException | AssertionError.
// Retain the native Java runtime families and reject Go string/runtime panics.
func javaStateQueueExpectRuntimeOrAssertion(t *testing.T, body func()) {
	t.Helper()
	defer func() {
		switch failure := recover().(type) {
		case *AssertionError, *RuntimeException, *NegativeArraySizeException,
			*IllegalArgumentException, *IllegalStateException, *ArrayIndexOutOfBoundsException,
			*IndexOutOfBoundsException, *StringIndexOutOfBoundsException, *ArithmeticException,
			*NullPointerException, *NoSuchElementException, *UnsupportedOperationException,
			*ClassCastException, *SecurityException, *ConcurrentModificationException,
			*RejectedExecutionException, *InvalidPathException, *NumberFormatException:
			return
		default:
			t.Fatalf("expected RuntimeException or AssertionError, got %T: %v", failure, failure)
		}
	}()
	body()
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

func TestStateDequeSDequeuePreservesStateQueueNullAssertion(t *testing.T) {
	oldWorkers := NumWorkers()
	SetNumWorkers(1)
	t.Cleanup(func() { SetNumWorkers(oldWorkers) })

	q := NewStateDeque()
	q.len = 1
	defer func() {
		failure := recover()
		assertion, ok := failure.(*AssertionError)
		if !ok || assertion.GetMessage() == nil || *assertion.GetMessage() != "Null state found on queue" {
			t.Fatalf("dequeue failure = %T: %v, want AssertionError with source message", failure, failure)
		}
		if q.len != 1 {
			t.Fatalf("failed dequeue changed logical size to %d", q.len)
		}
	}()
	q.SDequeue()
}

func TestStateDequeNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var q *StateDeque
	for name, call := range map[string]func(){
		"enqueue":          func() { q.Enqueue(nil) },
		"dequeue":          func() { _ = q.Dequeue() },
		"s_enqueue":        func() { q.SEnqueue(nil) },
		"s_enqueue_all":    func() { q.SEnqueueAll(nil) },
		"s_enqueue_vec":    func() { q.SEnqueueVec(nil) },
		"s_peek":           func() { _ = q.SPeek() },
		"s_dequeue":        func() { _ = q.SDequeue() },
		"s_dequeue_many":   func() { _ = q.SDequeueMany(0) },
		"finish_all":       func() { q.FinishAll() },
		"suspend_all":      func() { _ = q.SuspendAll() },
		"resume_all":       func() { q.ResumeAll() },
		"resume_all_stuck": func() { q.ResumeAllStuck() },
		"wake_all_waiters": func() { q.WakeAllWaiters() },
		"size":             func() { _ = q.Size() },
		"is_empty":         func() { _ = q.IsEmpty() },
		"begin_checkpoint": func() { _ = q.BeginChkpt() },
		"commit_checkpoint": func() {
			_ = q.CommitChkpt()
		},
		"recover": func() { _ = q.Recover() },
		"delete":  func() { _ = q.Delete() },
		"push":    func() { q.Push(nil) },
		"pop":     func() { _ = q.Pop() },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
			}()
			call()
		})
	}
}

func TestMemStateQueueNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var q *MemStateQueue
	for name, call := range map[string]func(){
		"enqueue":          func() { q.Enqueue(nil) },
		"dequeue":          func() { _ = q.Dequeue() },
		"s_enqueue":        func() { q.SEnqueue(nil) },
		"s_enqueue_all":    func() { q.SEnqueueAll(nil) },
		"s_enqueue_vec":    func() { q.SEnqueueVec(nil) },
		"s_peek":           func() { _ = q.SPeek() },
		"s_dequeue":        func() { _ = q.SDequeue() },
		"s_dequeue_many":   func() { _ = q.SDequeueMany(0) },
		"finish_all":       func() { q.FinishAll() },
		"suspend_all":      func() { _ = q.SuspendAll() },
		"resume_all":       func() { q.ResumeAll() },
		"resume_all_stuck": func() { q.ResumeAllStuck() },
		"wake_all_waiters": func() { q.WakeAllWaiters() },
		"size":             func() { _ = q.Size() },
		"is_empty":         func() { _ = q.IsEmpty() },
		"begin_checkpoint": func() { _ = q.BeginChkpt() },
		"commit_checkpoint": func() {
			_ = q.CommitChkpt()
		},
		"recover": func() { _ = q.Recover() },
		"delete":  func() { _ = q.Delete() },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
			}()
			call()
		})
	}
}

func TestDiskStateQueueNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var q *DiskStateQueue
	for name, call := range map[string]func(){
		"enqueue":          func() { q.Enqueue(nil) },
		"dequeue":          func() { _ = q.Dequeue() },
		"s_enqueue":        func() { q.SEnqueue(nil) },
		"s_enqueue_all":    func() { q.SEnqueueAll(nil) },
		"s_enqueue_vec":    func() { q.SEnqueueVec(nil) },
		"s_peek":           func() { _ = q.SPeek() },
		"s_dequeue":        func() { _ = q.SDequeue() },
		"s_dequeue_many":   func() { _ = q.SDequeueMany(0) },
		"finish_all":       func() { q.FinishAll() },
		"suspend_all":      func() { _ = q.SuspendAll() },
		"resume_all":       func() { q.ResumeAll() },
		"resume_all_stuck": func() { q.ResumeAllStuck() },
		"wake_all_waiters": func() { q.WakeAllWaiters() },
		"size":             func() { _ = q.Size() },
		"is_empty":         func() { _ = q.IsEmpty() },
		"begin_checkpoint": func() { _ = q.BeginChkpt() },
		"commit_checkpoint": func() {
			_ = q.CommitChkpt()
		},
		"recover": func() { _ = q.Recover() },
		"delete":  func() { _ = q.Delete() },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
			}()
			call()
		})
	}
}

func TestDiskByteArrayQueueNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var q *DiskByteArrayQueue
	invalidState := &TLCStateMut{functional: true}
	for name, call := range map[string]func(){
		"set_disk_dir":     func() { q.SetDiskDir("") },
		"enqueue":          func() { q.Enqueue(invalidState) },
		"dequeue":          func() { _ = q.Dequeue() },
		"s_enqueue":        func() { q.SEnqueue(invalidState) },
		"s_enqueue_all":    func() { q.SEnqueueAll([]*TLCStateMut{invalidState}) },
		"s_enqueue_vec":    func() { q.SEnqueueVec(NewStateVecFrom([]*TLCStateMut{invalidState})) },
		"s_peek":           func() { _ = q.SPeek() },
		"s_dequeue":        func() { _ = q.SDequeue() },
		"s_dequeue_many":   func() { _ = q.SDequeueMany(0) },
		"finish_all":       func() { q.FinishAll() },
		"suspend_all":      func() { _ = q.SuspendAll() },
		"resume_all":       func() { q.ResumeAll() },
		"resume_all_stuck": func() { q.ResumeAllStuck() },
		"wake_all_waiters": func() { q.WakeAllWaiters() },
		"size":             func() { _ = q.Size() },
		"is_empty":         func() { _ = q.IsEmpty() },
		"begin_checkpoint": func() { _ = q.BeginChkpt() },
		"commit_checkpoint": func() {
			_ = q.CommitChkpt()
		},
		"recover": func() { _ = q.Recover() },
		"delete":  func() { _ = q.Delete() },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
			}()
			call()
		})
	}
}

func TestMemStateQueueSEnqueueVecUsesJavaLengthDependentSlots(t *testing.T) {
	initTLCCheckerTest(t)
	q := NewMemStateQueue()
	first := checkerTestState(1)
	second := checkerTestState(2)
	q.SEnqueueVec(NewStateVecFrom([]*TLCStateMut{first, nil, second}))

	if q.Size() != 2 {
		t.Fatalf("size after SEnqueueVec with nil = %d, want 2", q.Size())
	}
	// StateQueue publishes len after the loop, while MemStateQueue.enqueueInner
	// uses that unchanged len for its slot. Thus the second non-null state
	// overwrites the first; the source counts two entries but stores just one.
	if got := q.Dequeue(); got != second {
		t.Fatalf("first dequeue = %p, want source's overwritten slot %p", got, second)
	}
	if got := q.Dequeue(); got != nil {
		t.Fatalf("second dequeue = %p, want source's unfilled slot", got)
	}
}

func TestStateDequeBulkEnqueueRetainsIndependentStorage(t *testing.T) {
	for _, batch := range []string{"array", "vector"} {
		t.Run(batch, func(t *testing.T) {
			q := NewStateDeque()
			states := make([]*TLCStateMut, memStateQueueInitialSize+3)
			for i := range states {
				states[i] = &TLCStateMut{UID: int64(i)}
			}
			if batch == "array" {
				q.SEnqueueAll(states)
			} else {
				q.SEnqueueVec(NewStateVecFrom(states))
			}
			if q.Size() != int64(len(states)) {
				t.Fatalf("batch size = %d, want %d", q.Size(), len(states))
			}
			for i := len(states) - 1; i >= 0; i-- {
				if got := q.Dequeue(); got != states[i] {
					t.Fatalf("dequeue %d = %p, want %p", i, got, states[i])
				}
			}
			if q.stored != 0 || q.Size() != 0 {
				t.Fatal("dequeue did not empty both storage and logical count")
			}
		})
	}
	q := NewStateDeque()
	initial, prefix := &TLCStateMut{UID: 1}, &TLCStateMut{UID: 2}
	q.Enqueue(initial)
	var failure any
	func() {
		defer func() { failure = recover() }()
		q.SEnqueueAll([]*TLCStateMut{prefix, nil})
	}()
	if failure == nil || q.Size() != 1 || q.stored != 2 || q.peekInner() != prefix {
		t.Fatalf("failed batch lost ArrayDeque prefix/count contract: failure=%v size=%d stored=%d", failure, q.Size(), q.stored)
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

func TestDiskByteArrayQueueSEnqueueVecCountPreservesSourceOverload(t *testing.T) {
	var nilVector *StateVec
	q := &DiskByteArrayQueue{}
	q.SEnqueueVecCount(nilVector, 0)
	if q.len != 0 || q.enqIndex != 0 {
		t.Fatal("zero-count overload changed an empty queue")
	}

	initTLCCheckerTest(t)
	q.enqBuf = make([][]byte, 3)
	q.deqBuf = make([][]byte, 3)
	q.deqIndex = 3
	q.cond = sync.NewCond(&q.mu)
	first, second := checkerTestState(1), checkerTestState(2)
	first.UID, second.UID = 1, 2
	q.SEnqueueVecCount(NewStateVecFrom([]*TLCStateMut{first, nil, second}), 2)
	if q.len != 2 || q.enqIndex != 2 || stateTestXValue(mustBytesToState(q.enqBuf[0])) != 2 || stateTestXValue(mustBytesToState(q.enqBuf[1])) != 1 {
		t.Fatalf("counted raw layout = len/index %d/%d, want two reversed states", q.len, q.enqIndex)
	}

	defer func() {
		if _, ok := recover().(*NegativeArraySizeException); !ok {
			t.Fatal("negative count did not raise NegativeArraySizeException")
		}
	}()
	q.SEnqueueVecCount(nilVector, -1)
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
