package tlc

import (
	"sync"
	"sync/atomic"
)

// Java synchronized(this) is reentrant, including recorder callbacks which
// query progress while modelCheck or registerWorker already owns the monitor.
// Use the same goroutine identity as the existing Java thread-local carriers.
type distributedServerMonitor struct {
	gate  sync.Mutex
	owner atomic.Uint64
	depth int // Only the owning goroutine accesses depth.
}

func (m *distributedServerMonitor) Lock() {
	id := currentGoroutineID()
	if id != 0 && m.owner.Load() == id {
		m.depth++
		return
	}
	m.gate.Lock()
	m.depth = 1
	m.owner.Store(id)
}

func (m *distributedServerMonitor) Unlock() {
	m.depth--
	if m.depth == 0 {
		m.owner.Store(0)
		m.gate.Unlock()
	}
}

// Object.wait releases every recursive acquisition and restores that depth
// when the waiting thread reacquires the monitor.
func (m *distributedServerMonitor) releaseForWait() int {
	depth := m.depth
	m.depth = 0
	m.owner.Store(0)
	m.gate.Unlock()
	return depth
}

func (m *distributedServerMonitor) reacquireAfterWait(depth int) {
	m.Lock()
	m.depth = depth
}
