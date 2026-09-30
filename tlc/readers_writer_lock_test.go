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
