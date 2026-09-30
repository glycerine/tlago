package tlc

import "sync"

type MemStateQueue struct {
	mu         sync.Mutex
	cond       *sync.Cond
	states     []*TLCStateMut
	start      int
	len        int64
	numWaiting int
	finish     bool
	stop       bool
}

const memStateQueueInitialSize = 4096

func NewMemStateQueue() *MemStateQueue {
	q := &MemStateQueue{states: make([]*TLCStateMut, memStateQueueInitialSize)}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *MemStateQueue) Enqueue(state *TLCStateMut) {
	q.enqueueInner(state)
}

func (q *MemStateQueue) Dequeue() *TLCStateMut {
	if q.IsEmpty() {
		return nil
	}
	return q.dequeueInner()
}

func (q *MemStateQueue) SEnqueue(state *TLCStateMut) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enqueueInner(state)
	if q.numWaiting > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *MemStateQueue) SEnqueueAll(states []*TLCStateMut) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, state := range states {
		if state != nil {
			q.enqueueInner(state)
		}
	}
	if q.numWaiting > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *MemStateQueue) SEnqueueVec(states *StateVec) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		return
	}
	for i := 0; i < states.Size(); i++ {
		if state := states.At(i); state != nil {
			q.enqueueInner(state)
		}
	}
	if q.numWaiting > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *MemStateQueue) SDequeue() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isAvailLocked() {
		return q.dequeueInner()
	}
	return nil
}

func (q *MemStateQueue) SPeek() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isAvailLocked() {
		return q.peekInner()
	}
	return nil
}

func (q *MemStateQueue) SDequeueMany(cnt int) []*TLCStateMut {
	if cnt <= 0 {
		panic("nonpositive number of states requested")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.isAvailLocked() {
		return nil
	}
	if int64(cnt) > q.len {
		cnt = int(q.len)
	}
	out := make([]*TLCStateMut, 0, cnt)
	for len(out) < cnt && q.len > 0 {
		out = append(out, q.dequeueInner())
	}
	return out
}

func (q *MemStateQueue) FinishAll() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.finish = true
	q.cond.Broadcast()
}

func (q *MemStateQueue) SuspendAll() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.finish {
		return false
	}
	q.stop = true
	for !q.finish && q.numWaiting < NumWorkers() {
		q.cond.Wait()
	}
	return !q.finish
}

func (q *MemStateQueue) ResumeAll() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stop = false
	q.cond.Broadcast()
}

func (q *MemStateQueue) ResumeAllStuck() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cond.Broadcast()
}

func (q *MemStateQueue) Size() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.len
}

func (q *MemStateQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.len < 1
}

func (q *MemStateQueue) isAvailLocked() bool {
	if q.finish {
		return false
	}
	for q.len == 0 || q.stop {
		q.numWaiting++
		if q.numWaiting >= NumWorkers() && q.len == 0 {
			q.numWaiting--
			q.cond.Broadcast()
			return false
		}
		q.cond.Wait()
		q.numWaiting--
		if q.finish {
			return false
		}
	}
	return true
}

func (q *MemStateQueue) enqueueInner(state *TLCStateMut) {
	if q.len == int64(len(q.states)) {
		q.grow()
	}
	last := (q.start + int(q.len)) % len(q.states)
	q.states[last] = state
	q.len++
}

func (q *MemStateQueue) dequeueInner() *TLCStateMut {
	state := q.states[q.start]
	q.states[q.start] = nil
	q.start = (q.start + 1) % len(q.states)
	q.len--
	return state
}

func (q *MemStateQueue) peekInner() *TLCStateMut {
	return q.states[q.start]
}

func (q *MemStateQueue) grow() {
	old := q.states
	newStates := make([]*TLCStateMut, max(1, (len(old)*4)/3+1))
	copyLen := len(old) - q.start
	copy(newStates, old[q.start:])
	copy(newStates[copyLen:], old[:q.start])
	q.states = newStates
	q.start = 0
}
