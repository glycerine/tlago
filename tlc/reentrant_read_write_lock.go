package tlc

import "sync"

// The default Java ReentrantReadWriteLock is nonfair: writers can barge when
// the lock is free, while new readers wait if the first queued waiter is a
// writer. Existing readers and the writer can reacquire the read lock; only
// the writer can reacquire the write lock. Read-to-write upgrades still block.
// TLCEval and TLCCache each use a class-wide lock across nested evaluations.
type reentrantReadWriteLock struct {
	gate       sync.Mutex
	changed    *sync.Cond
	writer     uint64
	writeDepth int
	readers    int
	readHolds  map[uint64]int
	waiters    []*reentrantRWWaiter
}

type reentrantRWWaiter struct{ write bool }

func (l *reentrantReadWriteLock) condition() *sync.Cond {
	if l.changed == nil {
		l.changed = sync.NewCond(&l.gate)
	}
	return l.changed
}

func (l *reentrantReadWriteLock) Lock() {
	id := currentGoroutineID()
	l.gate.Lock()
	defer l.gate.Unlock()
	if l.writer == id && l.writeDepth != 0 {
		if l.writeDepth == 65535 {
			panic(NewJavaError("Maximum lock count exceeded"))
		}
		l.writeDepth++
		return
	}
	if l.writer == 0 && l.readers == 0 {
		l.writer, l.writeDepth = id, 1
		return
	}
	waiter := &reentrantRWWaiter{write: true}
	l.waiters = append(l.waiters, waiter)
	condition := l.condition()
	for l.writer != 0 || l.readers != 0 || l.waiters[0] != waiter {
		condition.Wait()
	}
	l.waiters = l.waiters[1:]
	l.writer, l.writeDepth = id, 1
}

func (l *reentrantReadWriteLock) Unlock() {
	id := currentGoroutineID()
	l.gate.Lock()
	defer l.gate.Unlock()
	if l.writer != id || l.writeDepth == 0 {
		panic("write lock is not held by this goroutine")
	}
	l.writeDepth--
	if l.writeDepth == 0 {
		l.writer = 0
		if l.changed != nil {
			l.changed.Broadcast()
		}
	}
}

func (l *reentrantReadWriteLock) RLock() {
	id := currentGoroutineID()
	l.gate.Lock()
	defer l.gate.Unlock()
	ownWrite := l.writer == id && l.writeDepth != 0
	ownRead := l.readHolds[id] > 0
	if (l.writer != 0 && !ownWrite) || (!ownWrite && !ownRead && len(l.waiters) > 0 && l.waiters[0].write) {
		waiter := &reentrantRWWaiter{}
		l.waiters = append(l.waiters, waiter)
		condition := l.condition()
		for l.writer != 0 || l.waiters[0] != waiter {
			condition.Wait()
		}
		l.waiters = l.waiters[1:]
		// The next shared waiter may enter while this reader still holds the lock.
		condition.Broadcast()
	}
	if l.readers == 65535 {
		panic(NewJavaError("Maximum lock count exceeded"))
	}
	if l.readHolds == nil {
		l.readHolds = make(map[uint64]int)
	}
	l.readers++
	l.readHolds[id]++
}

func (l *reentrantReadWriteLock) RUnlock() {
	id := currentGoroutineID()
	l.gate.Lock()
	defer l.gate.Unlock()
	if l.readHolds[id] == 0 {
		panic("read lock is not held by this goroutine")
	}
	l.readHolds[id]--
	if l.readHolds[id] == 0 {
		delete(l.readHolds, id)
	}
	l.readers--
	if l.readers == 0 && l.changed != nil {
		l.changed.Broadcast()
	}
}
