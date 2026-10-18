package tlc

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type StateQueue interface {
	Enqueue(state *TLCStateMut)
	Dequeue() *TLCStateMut
	SEnqueue(state *TLCStateMut)
	// Bulk enqueue rejects nil batches; explicit empty batches are valid.
	SEnqueueAll(states []*TLCStateMut)
	SEnqueueVec(states *StateVec)
	SPeek() *TLCStateMut
	SDequeue() *TLCStateMut
	SDequeueMany(cnt int) []*TLCStateMut
	FinishAll()
	SuspendAll() bool
	ResumeAll()
	ResumeAllStuck()
	// WakeAllWaiters wakes consumers without changing suspension or completion.
	WakeAllWaiters()
	Size() int64
	IsEmpty() bool
	BeginChkpt() error
	CommitChkpt() error
	Recover() error
	Delete() error
}

func deleteQueueDirLikeJava(path string) error {
	if path == "" {
		return nil
	}
	// Java's DiskStateQueue.delete/DiskByteArrayQueue.delete call
	// File.delete() on the queue directory and ignore the boolean result.
	// This is intentionally non-recursive; ModelChecker cleanup owns the
	// success-vs-failure recursive metadata-directory policy.
	_ = os.Remove(path)
	return nil
}

func NewStateQueue(metaDir string) StateQueue {
	if stateQueuePropertyBool(modelCheckerBAQueueProperty, "TLAGO_MODEL_CHECKER_BAQUEUE") {
		return NewDiskByteArrayQueue(metaDir)
	}
	switch GetStateQueueName() {
	case "MemStateQueue":
		return NewMemStateQueue(metaDir)
	case "StateDeque":
		return NewStateDeque()
	case "DiskByteArrayQueue":
		return NewDiskByteArrayQueue(metaDir)
	default:
		return NewDiskStateQueue(metaDir)
	}
}

func GetStateQueueName() string {
	if name, ok := stateQueueProperty("tlc2.tool.queue.IStateQueue", "TLAGO_STATE_QUEUE"); ok {
		return name
	}
	return "DiskStateQueue"
}

func stateQueueProperty(name string, aliases ...string) (string, bool) {
	if value, ok := tlcLookupSystemProperty(name); ok {
		return value, true
	}
	for _, key := range aliases {
		if value, ok := os.LookupEnv(key); ok {
			return value, true
		}
	}
	return "", false
}

func stateQueuePropertyBool(name string, aliases ...string) bool {
	if value, ok := stateQueueProperty(name, aliases...); ok {
		return javaBooleanProperty(value)
	}
	return false
}

func stateQueueInitialized(queue StateQueue) bool {
	switch q := queue.(type) {
	case nil:
		return false
	case *MemStateQueue:
		return q.diskdir != ""
	case *DiskStateQueue:
		return q.diskdir != ""
	case *DiskByteArrayQueue:
		return q.diskdir != ""
	case *StateDeque:
		return true
	default:
		return true
	}
}

func setStateQueueDir(queue StateQueue, diskdir string) {
	switch q := queue.(type) {
	case *MemStateQueue:
		q.diskdir = diskdir
	case *DiskStateQueue:
		q.diskdir = diskdir
	case *DiskByteArrayQueue:
		q.SetDiskDir(diskdir)
	case *StateDeque:
	}
}

type stateQueueSuspendBarrier struct {
	once sync.Once
	mu   sync.Mutex
	cond *sync.Cond
}

func (b *stateQueueSuspendBarrier) ensure() {
	b.once.Do(func() {
		b.cond = sync.NewCond(&b.mu)
	})
}

func (b *stateQueueSuspendBarrier) signal() {
	b.ensure()
	b.mu.Lock()
	b.cond.Signal()
	b.mu.Unlock()
}

func (b *stateQueueSuspendBarrier) broadcast() {
	b.ensure()
	b.mu.Lock()
	b.cond.Broadcast()
	b.mu.Unlock()
}

type MemStateQueue struct {
	mu         sync.Mutex
	cond       *sync.Cond
	states     []*TLCStateMut
	start      int
	len        int64
	numWaiting atomic.Int32
	finish     atomic.Bool
	stop       bool
	suspend    stateQueueSuspendBarrier
	diskdir    string
}

// StateQueue.len is a plain Java long. Atomic snapshots avoid a Go data race
// without turning its separate read/modify/write updates into atomic additions.
func stateQueueLengthLoad(length *int64) int64 {
	return atomic.LoadInt64(length)
}

func stateQueueLengthStore(length *int64, value int64) {
	atomic.StoreInt64(length, value)
}

func stateQueueLengthAdd(length *int64, delta int64) {
	stateQueueLengthStore(length, stateQueueLengthLoad(length)+delta)
}

const memStateQueueInitialSize = 4096
const maxJavaQueueLength = int64(1<<31 - 1)

func NewMemStateQueue(metaDir ...string) *MemStateQueue {
	if len(metaDir) > 1 {
		panic(NewIllegalArgumentException("NewMemStateQueue accepts at most one metadata directory"))
	}
	diskdir := ""
	if len(metaDir) > 0 {
		diskdir = metaDir[0]
	}
	q := &MemStateQueue{states: make([]*TLCStateMut, memStateQueueInitialSize), diskdir: diskdir}
	q.cond = sync.NewCond(&q.mu)
	q.suspend.ensure()
	return q
}

func (q *MemStateQueue) requireMemStateQueue() {
	if q == nil {
		panic(NewNullPointerException())
	}
}

func (q *MemStateQueue) Enqueue(state *TLCStateMut) {
	q.requireMemStateQueue()
	q.enqueueInner(state)
	stateQueueLengthAdd(&q.len, 1)
}

func (q *MemStateQueue) Dequeue() *TLCStateMut {
	q.requireMemStateQueue()
	if q.IsEmpty() {
		return nil
	}
	state := q.dequeueInner()
	stateQueueLengthAdd(&q.len, -1)
	return state
}

func (q *MemStateQueue) SEnqueue(state *TLCStateMut) {
	q.requireMemStateQueue()
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enqueueInner(state)
	stateQueueLengthAdd(&q.len, 1)
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *MemStateQueue) SEnqueueAll(states []*TLCStateMut) {
	q.requireMemStateQueue()
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		panic(NewNullPointerException())
	}
	for _, state := range states {
		q.enqueueInner(state)
	}
	stateQueueLengthAdd(&q.len, int64(len(states)))
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *MemStateQueue) SEnqueueVec(states *StateVec) {
	q.requireMemStateQueue()
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		panic(NewNullPointerException())
	}
	var count int64
	for i := 0; i < states.Size(); i++ {
		if state := states.At(i); state != nil {
			q.enqueueInner(state)
			count++
		}
	}
	stateQueueLengthAdd(&q.len, count)
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *MemStateQueue) SDequeue() *TLCStateMut {
	q.requireMemStateQueue()
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isAvailLocked() {
		state := q.dequeueInner()
		if state == nil {
			panic(NewAssertionError("Null state found on queue"))
		}
		stateQueueLengthAdd(&q.len, -1)
		return state
	}
	return nil
}

func (q *MemStateQueue) SPeek() *TLCStateMut {
	q.requireMemStateQueue()
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isAvailLocked() {
		return q.peekInner()
	}
	return nil
}

func (q *MemStateQueue) SDequeueMany(cnt int) []*TLCStateMut {
	q.requireMemStateQueue()
	cnt = int(int32(cnt))
	q.mu.Lock()
	defer q.mu.Unlock()
	if cnt <= 0 {
		panic(NewAssertionError("Nonpositive number of states requested."))
	}
	if !q.isAvailLocked() {
		return nil
	}
	if int64(cnt) > stateQueueLengthLoad(&q.len) {
		cnt = int(stateQueueLengthLoad(&q.len))
	}
	out := make([]*TLCStateMut, 0, cnt)
	for len(out) < cnt && stateQueueLengthLoad(&q.len) > 0 {
		out = append(out, q.dequeueInner())
		stateQueueLengthAdd(&q.len, -1)
	}
	return out
}

func (q *MemStateQueue) FinishAll() {
	q.requireMemStateQueue()
	q.mu.Lock()
	q.finish.Store(true)
	q.cond.Broadcast()
	q.suspend.signal()
	q.mu.Unlock()
}

func (q *MemStateQueue) SuspendAll() bool {
	q.requireMemStateQueue()
	q.mu.Lock()
	if q.finish.Load() {
		q.mu.Unlock()
		return false
	}
	q.stop = true
	needWait := q.needsWaiting()
	q.mu.Unlock()
	for needWait {
		q.suspend.ensure()
		q.suspend.mu.Lock()
		if q.finish.Load() {
			q.suspend.mu.Unlock()
			return false
		}
		if !q.needsWaiting() {
			q.suspend.mu.Unlock()
			return true
		}
		q.suspend.cond.Wait()
		q.suspend.mu.Unlock()
		q.mu.Lock()
		if q.finish.Load() {
			q.mu.Unlock()
			return false
		}
		needWait = q.needsWaiting()
		q.mu.Unlock()
	}
	return true
}

func (q *MemStateQueue) ResumeAll() {
	q.requireMemStateQueue()
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stop = false
	q.cond.Broadcast()
}

func (q *MemStateQueue) WakeAllWaiters() {
	q.requireMemStateQueue()
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cond.Broadcast()
}

func (q *MemStateQueue) ResumeAllStuck() {
	q.requireMemStateQueue()
	q.mu.Lock()
	stop := q.stop
	q.mu.Unlock()
	if stop {
		q.suspend.broadcast()
	}
	q.mu.Lock()
	if !q.stop && stateQueueLengthLoad(&q.len) >= 1 && q.numWaiting.Load() > 0 {
		q.cond.Broadcast()
	}
	q.mu.Unlock()
}

func (q *MemStateQueue) Size() int64 {
	q.requireMemStateQueue()
	return stateQueueLengthLoad(&q.len)
}

func (q *MemStateQueue) IsEmpty() bool {
	q.requireMemStateQueue()
	return stateQueueLengthLoad(&q.len) < 1
}

func (q *MemStateQueue) BeginChkpt() error {
	q.requireMemStateQueue()
	file, err := openPoolFile(q.checkpointPath("tmp"), true)
	if err != nil {
		return err
	}
	out, err := newValueOutputStreamWithCompression(file, UseGZIP())
	if err != nil {
		return err
	}
	if err := out.WriteInt(int32(stateQueueLengthLoad(&q.len))); err != nil {
		return err
	}
	index := q.start
	for i := int64(0); i < stateQueueLengthLoad(&q.len); i++ {
		if err := q.states[index].Write(out); err != nil {
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
	q.requireMemStateQueue()
	oldName := q.checkpointPath("chkpt")
	newName := q.checkpointPath("tmp")
	if _, err := os.Stat(oldName); err == nil {
		if err := os.Remove(oldName); err != nil {
			return NewIOException(fmt.Sprintf("MemStateQueue.commitChkpt: cannot delete %s", oldName))
		}
	}
	if err := os.Rename(newName, oldName); err != nil {
		return NewIOException(fmt.Sprintf("MemStateQueue.commitChkpt: cannot delete %s", oldName))
	}
	return nil
}

func (q *MemStateQueue) Recover() error {
	q.requireMemStateQueue()
	file, err := openPoolFile(q.checkpointPath("chkpt"), false)
	if err != nil {
		return err
	}
	in, err := NewValueInputStreamWithGlobalCompression(file)
	if err != nil {
		return err
	}
	length, err := in.ReadInt()
	if err != nil {
		return err
	}
	stateQueueLengthStore(&q.len, int64(length))
	for i := int32(0); i < length; i++ {
		state := newStateFromEmptyFactory()
		if int(i) >= len(q.states) {
			return NewArrayIndexOutOfBoundsException(int(i), len(q.states))
		}
		q.states[i] = state
		if err := state.Read(in); err != nil {
			return err
		}
	}
	return in.Close()
}

func (q *MemStateQueue) checkpointPath(ext string) string {
	return q.diskdir + string(os.PathSeparator) + "queue." + ext
}

func (q *MemStateQueue) Delete() error {
	q.requireMemStateQueue()
	return nil
}

func (q *MemStateQueue) isAvailLocked() bool {
	if q.finish.Load() {
		return false
	}
	for stateQueueLengthLoad(&q.len) == 0 || q.stop {
		waiting := q.numWaiting.Add(1)
		if int(waiting) >= NumWorkers() {
			if stateQueueLengthLoad(&q.len) == 0 {
				q.numWaiting.Add(-1)
				return false
			}
			q.suspend.signal()
		}
		q.cond.Wait()
		q.numWaiting.Add(-1)
		if q.finish.Load() {
			return false
		}
	}
	return true
}

func (q *MemStateQueue) needsWaiting() bool {
	return int(q.numWaiting.Load()) < NumWorkers()
}

func (q *MemStateQueue) enqueueInner(state *TLCStateMut) {
	if stateQueueLengthLoad(&q.len) > maxJavaQueueLength {
		panic(NewTLCRuntimeException(ECSystemErrorWritingStates, "queue", "Amount of states exceeds internal storage"))
	}
	if stateQueueLengthLoad(&q.len) == int64(len(q.states)) {
		q.grow()
	}
	last := (q.start + int(stateQueueLengthLoad(&q.len))) % len(q.states)
	q.states[last] = state
}

func (q *MemStateQueue) dequeueInner() *TLCStateMut {
	state := q.states[q.start]
	q.states[q.start] = nil
	q.start = (q.start + 1) % len(q.states)
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

type StateDeque struct {
	mu         sync.Mutex
	cond       *sync.Cond
	states     []*TLCStateMut
	start      int
	len        int64
	stored     int // ArrayDeque occupancy is independent of StateQueue.len.
	numWaiting atomic.Int32
	finish     atomic.Bool
	stop       bool
	suspend    stateQueueSuspendBarrier
}

func NewStateDeque() *StateDeque {
	q := &StateDeque{states: make([]*TLCStateMut, memStateQueueInitialSize)}
	q.cond = sync.NewCond(&q.mu)
	q.suspend.ensure()
	return q
}

func (q *StateDeque) requireStateDeque() {
	if q == nil {
		panic(NewNullPointerException())
	}
}

func (q *StateDeque) Enqueue(state *TLCStateMut) {
	q.requireStateDeque()
	q.enqueueInner(state)
	stateQueueLengthAdd(&q.len, 1)
}

func (q *StateDeque) Dequeue() *TLCStateMut {
	q.requireStateDeque()
	if q.IsEmpty() {
		return nil
	}
	state := q.dequeueInner()
	stateQueueLengthAdd(&q.len, -1)
	return state
}

func (q *StateDeque) SEnqueue(state *TLCStateMut) {
	q.requireStateDeque()
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enqueueInner(state)
	stateQueueLengthAdd(&q.len, 1)
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *StateDeque) SEnqueueAll(states []*TLCStateMut) {
	q.requireStateDeque()
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		panic(NewNullPointerException())
	}
	for _, state := range states {
		q.enqueueInner(state)
	}
	stateQueueLengthAdd(&q.len, int64(len(states)))
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *StateDeque) SEnqueueVec(states *StateVec) {
	q.requireStateDeque()
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		panic(NewNullPointerException())
	}
	var count int64
	for i := 0; i < states.Size(); i++ {
		if state := states.At(i); state != nil {
			q.enqueueInner(state)
			count++
		}
	}
	stateQueueLengthAdd(&q.len, count)
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *StateDeque) SPeek() *TLCStateMut {
	q.requireStateDeque()
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isAvailLocked() {
		return q.peekInner()
	}
	return nil
}

func (q *StateDeque) SDequeue() *TLCStateMut {
	q.requireStateDeque()
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isAvailLocked() {
		state := q.dequeueInner()
		if state == nil {
			panic(NewAssertionError("Null state found on queue"))
		}
		stateQueueLengthAdd(&q.len, -1)
		return state
	}
	return nil
}

func (q *StateDeque) SDequeueMany(cnt int) []*TLCStateMut {
	q.requireStateDeque()
	cnt = int(int32(cnt))
	q.mu.Lock()
	defer q.mu.Unlock()
	if cnt <= 0 {
		panic(NewAssertionError("Nonpositive number of states requested."))
	}
	if !q.isAvailLocked() {
		return nil
	}
	if int64(cnt) > stateQueueLengthLoad(&q.len) {
		cnt = int(stateQueueLengthLoad(&q.len))
	}
	out := make([]*TLCStateMut, 0, cnt)
	for len(out) < cnt && stateQueueLengthLoad(&q.len) > 0 {
		out = append(out, q.dequeueInner())
		stateQueueLengthAdd(&q.len, -1)
	}
	return out
}

func (q *StateDeque) FinishAll() {
	q.requireStateDeque()
	q.mu.Lock()
	q.finish.Store(true)
	q.cond.Broadcast()
	q.suspend.signal()
	q.mu.Unlock()
}

func (q *StateDeque) SuspendAll() bool {
	q.requireStateDeque()
	q.mu.Lock()
	if q.finish.Load() {
		q.mu.Unlock()
		return false
	}
	q.stop = true
	needWait := q.needsWaiting()
	q.mu.Unlock()
	for needWait {
		q.suspend.ensure()
		q.suspend.mu.Lock()
		if q.finish.Load() {
			q.suspend.mu.Unlock()
			return false
		}
		if !q.needsWaiting() {
			q.suspend.mu.Unlock()
			return true
		}
		q.suspend.cond.Wait()
		q.suspend.mu.Unlock()
		q.mu.Lock()
		if q.finish.Load() {
			q.mu.Unlock()
			return false
		}
		needWait = q.needsWaiting()
		q.mu.Unlock()
	}
	return true
}

func (q *StateDeque) ResumeAll() {
	q.requireStateDeque()
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stop = false
	q.cond.Broadcast()
}

func (q *StateDeque) WakeAllWaiters() {
	q.requireStateDeque()
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cond.Broadcast()
}

func (q *StateDeque) ResumeAllStuck() {
	q.requireStateDeque()
	q.mu.Lock()
	stop := q.stop
	q.mu.Unlock()
	if stop {
		q.suspend.broadcast()
	}
	q.mu.Lock()
	if !q.stop && stateQueueLengthLoad(&q.len) >= 1 && q.numWaiting.Load() > 0 {
		q.cond.Broadcast()
	}
	q.mu.Unlock()
}

func (q *StateDeque) Size() int64 {
	q.requireStateDeque()
	return stateQueueLengthLoad(&q.len)
}

func (q *StateDeque) IsEmpty() bool {
	q.requireStateDeque()
	return stateQueueLengthLoad(&q.len) < 1
}

func (q *StateDeque) BeginChkpt() error {
	q.requireStateDeque()
	return NewUnsupportedOperationException("StateDeque does not support checkpointing.")
}

func (q *StateDeque) CommitChkpt() error {
	q.requireStateDeque()
	return q.BeginChkpt()
}

func (q *StateDeque) Recover() error {
	q.requireStateDeque()
	return q.BeginChkpt()
}

func (q *StateDeque) Delete() error {
	q.requireStateDeque()
	return nil
}

func (q *StateDeque) Push(state *TLCStateMut) {
	q.requireStateDeque()
	q.SEnqueue(state)
}

func (q *StateDeque) Pop() *TLCStateMut {
	q.requireStateDeque()
	return q.SDequeue()
}

func (q *StateDeque) isAvailLocked() bool {
	if q.finish.Load() {
		return false
	}
	for stateQueueLengthLoad(&q.len) == 0 || q.stop {
		waiting := q.numWaiting.Add(1)
		if int(waiting) >= NumWorkers() {
			if stateQueueLengthLoad(&q.len) == 0 {
				q.numWaiting.Add(-1)
				return false
			}
			q.suspend.signal()
		}
		q.cond.Wait()
		q.numWaiting.Add(-1)
		if q.finish.Load() {
			return false
		}
	}
	return true
}

func (q *StateDeque) needsWaiting() bool {
	return int(q.numWaiting.Load()) < NumWorkers()
}

func (q *StateDeque) enqueueInner(state *TLCStateMut) {
	if state == nil {
		panic(NewNullPointerException())
	}
	if q.stored == len(q.states) {
		q.grow()
	}
	q.start = (q.start - 1 + len(q.states)) % len(q.states)
	q.states[q.start] = state
	q.stored++
}

func (q *StateDeque) dequeueInner() *TLCStateMut {
	if q.stored == 0 {
		return nil
	}
	state := q.states[q.start]
	q.states[q.start] = nil
	q.start = (q.start + 1) % len(q.states)
	q.stored--
	return state
}

func (q *StateDeque) peekInner() *TLCStateMut {
	return q.states[q.start]
}

func (q *StateDeque) grow() {
	old := q.states
	newStates := make([]*TLCStateMut, max(1, (len(old)*4)/3+1))
	copyLen := len(old) - q.start
	copy(newStates, old[q.start:])
	copy(newStates[copyLen:], old[:q.start])
	q.states = newStates
	q.start = 0
}
