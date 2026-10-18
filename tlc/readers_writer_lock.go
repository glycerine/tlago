package tlc

import (
	"fmt"
	"sync"
)

type ReadersWriterLock struct {
	mu             sync.Mutex
	cond           *sync.Cond
	numReaders     int
	hasWriter      bool
	waitingWriters int
}

func NewReadersWriterLock() *ReadersWriterLock {
	lock := &ReadersWriterLock{}
	lock.cond = sync.NewCond(&lock.mu)
	return lock
}

func (l *ReadersWriterLock) ensureCond() {
	if l.cond == nil {
		l.cond = sync.NewCond(&l.mu)
	}
}

func (l *ReadersWriterLock) BeginRead() {
	if l == nil {
		panic(NewNullPointerException())
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ensureCond()
	for l.hasWriter || l.waitingWriters > 0 {
		l.cond.Wait()
	}
	l.numReaders++
}

func (l *ReadersWriterLock) EndRead() {
	if l == nil {
		panic(NewNullPointerException())
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ensureCond()
	l.numReaders--
	if l.numReaders == 0 {
		l.cond.Broadcast()
	}
}

func (l *ReadersWriterLock) BeginWrite() {
	if l == nil {
		panic(NewNullPointerException())
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ensureCond()
	for l.numReaders > 0 || l.hasWriter {
		l.waitingWriters++
		l.cond.Wait()
		l.waitingWriters--
	}
	l.hasWriter = true
}

func (l *ReadersWriterLock) EndWrite() {
	if l == nil {
		panic(NewNullPointerException())
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ensureCond()
	l.hasWriter = false
	l.cond.Broadcast()
}

type SingleThreadedReadersWriterLock struct{}

func NewSingleThreadedReadersWriterLock() *SingleThreadedReadersWriterLock {
	return &SingleThreadedReadersWriterLock{}
}

func (l *SingleThreadedReadersWriterLock) BeginRead() {
	if l == nil {
		panic(NewNullPointerException())
	}
}

func (l *SingleThreadedReadersWriterLock) EndRead() {
	if l == nil {
		panic(NewNullPointerException())
	}
}

func (l *SingleThreadedReadersWriterLock) BeginWrite() {
	if l == nil {
		panic(NewNullPointerException())
	}
}

func (l *SingleThreadedReadersWriterLock) EndWrite() {
	if l == nil {
		panic(NewNullPointerException())
	}
}

type Striped struct {
	locks []*sync.RWMutex
}

func NewStriped(lockCnt int) *Striped {
	lockCnt = int(int32(lockCnt))
	if lockCnt < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(lockCnt)))
	}
	striped := &Striped{locks: make([]*sync.RWMutex, lockCnt)}
	for i := range striped.locks {
		striped.locks[i] = &sync.RWMutex{}
	}
	return striped
}

func StripedReadWriteLock(lockCnt int) *Striped {
	return NewStriped(lockCnt)
}

func (s *Striped) GetAt(lockIndex int) *sync.RWMutex {
	if s == nil {
		panic(NewNullPointerException())
	}
	lockIndex = int(int32(lockIndex))
	if lockIndex < 0 || lockIndex >= len(s.locks) {
		panic(NewArrayIndexOutOfBoundsException(lockIndex, len(s.locks)))
	}
	return s.locks[lockIndex]
}

func (s *Striped) Size() int {
	if s == nil {
		panic(NewNullPointerException())
	}
	return len(s.locks)
}

func (s *Striped) ReleaseAllLocks() {
	if s == nil {
		panic(NewNullPointerException())
	}
	for i := s.Size() - 1; i >= 0; i-- {
		s.locks[i].Unlock()
	}
}

func (s *Striped) AcquireAllLocks() {
	if s == nil {
		panic(NewNullPointerException())
	}
	for i := 0; i < s.Size(); i++ {
		s.locks[i].Lock()
	}
}

func (s *Striped) AcquireAllLocksExcept(skip int) {
	if s == nil {
		return
	}
	for i := 0; i < s.Size(); i++ {
		if i == skip {
			continue
		}
		s.locks[i].Lock()
	}
}

func (s *Striped) ReleaseAllLocksExcept(skip int) {
	if s == nil {
		return
	}
	for i := s.Size() - 1; i >= 0; i-- {
		if i == skip {
			continue
		}
		s.locks[i].Unlock()
	}
}
