package tlc

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

var errNilStateQueueBatch = errors.New("cannot enqueue a nil state batch")

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

const memStateQueueInitialSize = 4096
const maxJavaQueueLength = int64(1<<31 - 1)

func NewMemStateQueue(metaDir ...string) *MemStateQueue {
	diskdir := ""
	if len(metaDir) > 0 {
		diskdir = metaDir[0]
	}
	q := &MemStateQueue{states: make([]*TLCStateMut, memStateQueueInitialSize), diskdir: diskdir}
	q.cond = sync.NewCond(&q.mu)
	q.suspend.ensure()
	return q
}

func (q *MemStateQueue) Enqueue(state *TLCStateMut) {
	q.enqueueInner(state)
	q.len++
}

func (q *MemStateQueue) Dequeue() *TLCStateMut {
	if q.IsEmpty() {
		return nil
	}
	state := q.dequeueInner()
	q.len--
	return state
}

func (q *MemStateQueue) SEnqueue(state *TLCStateMut) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enqueueInner(state)
	q.len++
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *MemStateQueue) SEnqueueAll(states []*TLCStateMut) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		panic(errNilStateQueueBatch)
	}
	for _, state := range states {
		q.enqueueInner(state)
	}
	q.len += int64(len(states))
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *MemStateQueue) SEnqueueVec(states *StateVec) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		panic(errNilStateQueueBatch)
	}
	var count int64
	for i := 0; i < states.Size(); i++ {
		if state := states.At(i); state != nil {
			q.enqueueInner(state)
			count++
		}
	}
	q.len += count
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *MemStateQueue) SDequeue() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isAvailLocked() {
		state := q.dequeueInner()
		q.len--
		return state
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
	q.mu.Lock()
	defer q.mu.Unlock()
	if cnt <= 0 {
		panic(NewAssertionError("Nonpositive number of states requested."))
	}
	if !q.isAvailLocked() {
		return nil
	}
	if int64(cnt) > q.len {
		cnt = int(q.len)
	}
	out := make([]*TLCStateMut, 0, cnt)
	for len(out) < cnt && q.len > 0 {
		out = append(out, q.dequeueInner())
		q.len--
	}
	return out
}

func (q *MemStateQueue) FinishAll() {
	q.mu.Lock()
	q.finish.Store(true)
	q.cond.Broadcast()
	q.mu.Unlock()
	q.suspend.signal()
}

func (q *MemStateQueue) SuspendAll() bool {
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
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stop = false
	q.cond.Broadcast()
}

func (q *MemStateQueue) WakeAllWaiters() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cond.Broadcast()
}

func (q *MemStateQueue) ResumeAllStuck() {
	q.mu.Lock()
	stop := q.stop
	nonempty := q.len >= 1
	waiting := q.numWaiting.Load()
	if !stop && nonempty && waiting > 0 {
		q.cond.Broadcast()
	}
	q.mu.Unlock()
	if stop {
		q.suspend.broadcast()
	}
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
	file, err := os.Create(q.checkpointPath("tmp"))
	if err != nil {
		return err
	}
	out := NewValueOutputStreamWithGlobalCompression(file)
	if err := out.WriteInt(int32(q.len)); err != nil {
		_ = out.Close()
		return err
	}
	index := q.start
	for i := int64(0); i < q.len; i++ {
		if err := q.states[index].Write(out); err != nil {
			_ = out.Close()
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
	q.mu.Lock()
	defer q.mu.Unlock()
	file, err := os.Open(q.checkpointPath("chkpt"))
	if err != nil {
		return err
	}
	in, err := NewValueInputStreamWithGlobalCompression(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	closed := false
	defer func() {
		if !closed {
			_ = in.Close()
		}
	}()
	length, err := in.ReadInt()
	if err != nil {
		return err
	}
	q.len = int64(length)
	for i := int32(0); i < length; i++ {
		state := NewEmptyState()
		if int(i) >= len(q.states) {
			return NewArrayIndexOutOfBoundsException(int(i), len(q.states))
		}
		q.states[i] = state
		if err := state.Read(in); err != nil {
			return err
		}
	}
	closed = true
	return in.Close()
}

func (q *MemStateQueue) checkpointPath(ext string) string {
	return q.diskdir + string(os.PathSeparator) + "queue." + ext
}

func (q *MemStateQueue) Delete() error {
	return nil
}

func (q *MemStateQueue) isAvailLocked() bool {
	if q.finish.Load() {
		return false
	}
	for q.len == 0 || q.stop {
		waiting := q.numWaiting.Add(1)
		if int(waiting) >= NumWorkers() {
			if q.len == 0 {
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
	if q.len > maxJavaQueueLength {
		panic(newTLCErrorCode(ECSystemErrorWritingStates, "queue", "Amount of states exceeds internal storage"))
	}
	if q.len == int64(len(q.states)) {
		q.grow()
	}
	last := (q.start + int(q.len)) % len(q.states)
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

func (q *StateDeque) Enqueue(state *TLCStateMut) {
	q.enqueueInner(state)
	q.len++
}

func (q *StateDeque) Dequeue() *TLCStateMut {
	if q.IsEmpty() {
		return nil
	}
	state := q.dequeueInner()
	q.len--
	return state
}

func (q *StateDeque) SEnqueue(state *TLCStateMut) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enqueueInner(state)
	q.len++
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *StateDeque) SEnqueueAll(states []*TLCStateMut) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		panic(errNilStateQueueBatch)
	}
	for _, state := range states {
		q.enqueueInner(state)
	}
	q.len += int64(len(states))
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *StateDeque) SEnqueueVec(states *StateVec) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		panic(errNilStateQueueBatch)
	}
	var count int64
	for i := 0; i < states.Size(); i++ {
		if state := states.At(i); state != nil {
			q.enqueueInner(state)
			count++
		}
	}
	q.len += count
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *StateDeque) SPeek() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isAvailLocked() {
		return q.peekInner()
	}
	return nil
}

func (q *StateDeque) SDequeue() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isAvailLocked() {
		state := q.dequeueInner()
		q.len--
		return state
	}
	return nil
}

func (q *StateDeque) SDequeueMany(cnt int) []*TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if cnt <= 0 {
		panic(NewAssertionError("Nonpositive number of states requested."))
	}
	if !q.isAvailLocked() {
		return nil
	}
	if int64(cnt) > q.len {
		cnt = int(q.len)
	}
	out := make([]*TLCStateMut, 0, cnt)
	for len(out) < cnt && q.len > 0 {
		out = append(out, q.dequeueInner())
		q.len--
	}
	return out
}

func (q *StateDeque) FinishAll() {
	q.mu.Lock()
	q.finish.Store(true)
	q.cond.Broadcast()
	q.mu.Unlock()
	q.suspend.signal()
}

func (q *StateDeque) SuspendAll() bool {
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
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stop = false
	q.cond.Broadcast()
}

func (q *StateDeque) WakeAllWaiters() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cond.Broadcast()
}

func (q *StateDeque) ResumeAllStuck() {
	q.mu.Lock()
	stop := q.stop
	nonempty := q.len >= 1
	waiting := q.numWaiting.Load()
	if !stop && nonempty && waiting > 0 {
		q.cond.Broadcast()
	}
	q.mu.Unlock()
	if stop {
		q.suspend.broadcast()
	}
}

func (q *StateDeque) Size() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.len
}

func (q *StateDeque) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.len < 1
}

func (q *StateDeque) BeginChkpt() error {
	return errors.New("StateDeque does not support checkpointing.")
}

func (q *StateDeque) CommitChkpt() error {
	return q.BeginChkpt()
}

func (q *StateDeque) Recover() error {
	return q.BeginChkpt()
}

func (q *StateDeque) Delete() error {
	return nil
}

func (q *StateDeque) Push(state *TLCStateMut) {
	q.SEnqueue(state)
}

func (q *StateDeque) Pop() *TLCStateMut {
	return q.SDequeue()
}

func (q *StateDeque) isAvailLocked() bool {
	if q.finish.Load() {
		return false
	}
	for q.len == 0 || q.stop {
		waiting := q.numWaiting.Add(1)
		if int(waiting) >= NumWorkers() {
			if q.len == 0 {
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
		panic("nil state")
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
