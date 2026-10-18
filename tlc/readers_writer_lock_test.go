package tlc

import (
	"testing"
	"time"
)

func TestReadersWriterLockGivesWaitingWriterPriority(t *testing.T) {
	lock := NewReadersWriterLock()
	lock.BeginRead()

	writerEntered := make(chan struct{})
	writerDone := make(chan struct{})
	go func() {
		lock.BeginWrite()
		close(writerEntered)
		time.Sleep(10 * time.Millisecond)
		lock.EndWrite()
		close(writerDone)
	}()

	time.Sleep(10 * time.Millisecond)
	readerEntered := make(chan struct{})
	go func() {
		lock.BeginRead()
		close(readerEntered)
		lock.EndRead()
	}()

	select {
	case <-readerEntered:
		t.Fatalf("reader entered while writer was waiting")
	default:
	}

	lock.EndRead()
	select {
	case <-writerEntered:
	case <-time.After(time.Second):
		t.Fatalf("writer did not enter after readers drained")
	}
	select {
	case <-readerEntered:
		t.Fatalf("reader entered before waiting writer finished")
	default:
	}
	select {
	case <-writerDone:
	case <-time.After(time.Second):
		t.Fatalf("writer did not finish")
	}
	select {
	case <-readerEntered:
	case <-time.After(time.Second):
		t.Fatalf("reader did not enter after writer finished")
	}
}

func TestReadersWriterLockNilReceiversFailLikeJavaInvocation(t *testing.T) {
	var ordinary *ReadersWriterLock
	var single *SingleThreadedReadersWriterLock
	for name, call := range map[string]func(){
		"ordinary/begin_read":  ordinary.BeginRead,
		"ordinary/end_read":    ordinary.EndRead,
		"ordinary/begin_write": ordinary.BeginWrite,
		"ordinary/end_write":   ordinary.EndWrite,
		"single/begin_read":    single.BeginRead,
		"single/end_read":      single.EndRead,
		"single/begin_write":   single.BeginWrite,
		"single/end_write":     single.EndWrite,
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

func TestStripedAcquireAndReleaseAllLocks(t *testing.T) {
	striped := StripedReadWriteLock(3)
	if striped.Size() != 3 {
		t.Fatalf("size = %d, want 3", striped.Size())
	}
	striped.AcquireAllLocks()
	locked := make(chan struct{})
	go func() {
		striped.GetAt(1).Lock()
		close(locked)
		striped.GetAt(1).Unlock()
	}()
	select {
	case <-locked:
		t.Fatalf("stripe lock acquired while AcquireAllLocks held it")
	default:
	}
	striped.ReleaseAllLocks()
	select {
	case <-locked:
	case <-time.After(time.Second):
		t.Fatalf("stripe lock did not release")
	}
}

func TestStripedNegativeLockCountFailsLikeJavaArrayAllocation(t *testing.T) {
	defer func() {
		failure := recover()
		if _, ok := failure.(*NegativeArraySizeException); !ok {
			t.Fatalf("negative stripe count failure = %T: %v, want NegativeArraySizeException", failure, failure)
		}
	}()
	_ = StripedReadWriteLock(-1)
}

func TestStripedInvalidIndexFailsLikeJavaArrayAccess(t *testing.T) {
	striped := StripedReadWriteLock(2)
	for _, index := range []int{-1, 2} {
		func() {
			defer func() {
				failure := recover()
				if _, ok := failure.(*ArrayIndexOutOfBoundsException); !ok {
					t.Fatalf("getAt(%d) failure = %T: %v, want ArrayIndexOutOfBoundsException", index, failure, failure)
				}
			}()
			_ = striped.GetAt(index)
		}()
	}
}

func TestStripedNilReceiverFailsLikeJavaInvocation(t *testing.T) {
	var striped *Striped
	for name, call := range map[string]func(){
		"get_at":  func() { _ = striped.GetAt(0) },
		"size":    func() { _ = striped.Size() },
		"acquire": striped.AcquireAllLocks,
		"release": striped.ReleaseAllLocks,
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
