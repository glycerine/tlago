package tlc

import "sync"

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
		return
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
		return
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
		return
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
		return
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

func (l *SingleThreadedReadersWriterLock) BeginRead()  {}
func (l *SingleThreadedReadersWriterLock) EndRead()    {}
func (l *SingleThreadedReadersWriterLock) BeginWrite() {}
func (l *SingleThreadedReadersWriterLock) EndWrite()   {}

type Striped struct {
	locks []*sync.RWMutex
}

func NewStriped(lockCnt int) *Striped {
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
	return s.locks[lockIndex]
}

func (s *Striped) Size() int {
	if s == nil {
		return 0
	}
	return len(s.locks)
}

func (s *Striped) ReleaseAllLocks() {
	if s == nil {
		return
	}
	for i := s.Size() - 1; i >= 0; i-- {
		s.locks[i].Unlock()
	}
}

func (s *Striped) AcquireAllLocks() {
	if s == nil {
		return
	}
	for i := 0; i < s.Size(); i++ {
		s.locks[i].Lock()
	}
}
