package tlc

import (
	"os"
	"path/filepath"
	"sync"
)

type MemStateQueue struct {
	mu         sync.Mutex
	cond       *sync.Cond
	states     []*TLCStateMut
	start      int
	len        int64
	numWaiting int
	finish     bool
	stop       bool
	diskdir    string
}

const memStateQueueInitialSize = 4096

func NewMemStateQueue(metaDir ...string) *MemStateQueue {
	diskdir := ""
	if len(metaDir) > 0 {
		diskdir = metaDir[0]
	}
	q := &MemStateQueue{states: make([]*TLCStateMut, memStateQueueInitialSize), diskdir: diskdir}
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

func (q *MemStateQueue) BeginChkpt() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.diskdir == "" {
		dir, err := os.MkdirTemp("", "MemStateQueue")
		if err != nil {
			return err
		}
		q.diskdir = dir
	}
	if err := os.MkdirAll(q.diskdir, 0o755); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(q.diskdir, "queue.tmp"))
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	if err := out.WriteInt(int32(q.len)); err != nil {
		file.Close()
		return err
	}
	index := q.start
	for i := int64(0); i < q.len; i++ {
		if err := q.states[index].Write(out); err != nil {
			file.Close()
			return err
		}
		index++
		if index == len(q.states) {
			index = 0
		}
	}
	return out.Close()
}

func (q *MemStateQueue) CommitChkpt() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.diskdir == "" {
		return nil
	}
	oldName := filepath.Join(q.diskdir, "queue.chkpt")
	newName := filepath.Join(q.diskdir, "queue.tmp")
	_ = os.Remove(oldName)
	return os.Rename(newName, oldName)
}

func (q *MemStateQueue) Recover() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.diskdir == "" {
		return nil
	}
	file, err := os.Open(filepath.Join(q.diskdir, "queue.chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	length, err := in.ReadInt()
	if err != nil {
		file.Close()
		return err
	}
	if int(length) > len(q.states) {
		q.states = make([]*TLCStateMut, max(memStateQueueInitialSize, int(length)))
	}
	for i := range q.states {
		q.states[i] = nil
	}
	q.start = 0
	q.len = int64(length)
	for i := int32(0); i < length; i++ {
		state := NewEmptyState()
		if err := state.Read(in); err != nil {
			file.Close()
			return err
		}
		q.states[i] = state
	}
	return in.Close()
}

func (q *MemStateQueue) Delete() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.diskdir == "" {
		return nil
	}
	if err := os.Remove(filepath.Join(q.diskdir, "queue.tmp")); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(filepath.Join(q.diskdir, "queue.chkpt")); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
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

type DiskByteArrayQueue struct {
	*MemStateQueue
}

func NewDiskByteArrayQueue(metaDir string) *DiskByteArrayQueue {
	return &DiskByteArrayQueue{MemStateQueue: NewMemStateQueue(metaDir)}
}

type StateDeque struct {
	*MemStateQueue
}

func NewStateDeque() *StateDeque {
	return &StateDeque{MemStateQueue: NewMemStateQueue()}
}

func (q *StateDeque) Push(state *TLCStateMut) {
	q.SEnqueue(state)
}

func (q *StateDeque) Pop() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.len == 0 {
		return nil
	}
	last := (q.start + int(q.len) - 1) % len(q.states)
	state := q.states[last]
	q.states[last] = nil
	q.len--
	return state
}
