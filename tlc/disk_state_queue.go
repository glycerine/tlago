package tlc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const diskStateQueueBufferSize = 8192

type DiskStateQueue struct {
	mu            sync.Mutex
	cond          *sync.Cond
	len           int64
	numWaiting    int
	finish        bool
	stop          bool
	diskdir       string
	deqBuf        []*TLCStateMut
	enqBuf        []*TLCStateMut
	deqIndex      int
	enqIndex      int
	loPool        int
	hiPool        int
	lastLoPool    int
	newLastLoPool int
}

func NewDiskStateQueue(metaDir string) *DiskStateQueue {
	if metaDir == "" {
		metaDir = filepath.Join(os.TempDir(), "DiskStateQueue")
	}
	q := &DiskStateQueue{
		diskdir:  metaDir,
		deqBuf:   make([]*TLCStateMut, diskStateQueueBufferSize),
		enqBuf:   make([]*TLCStateMut, diskStateQueueBufferSize),
		deqIndex: diskStateQueueBufferSize,
		loPool:   1,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *DiskStateQueue) Enqueue(state *TLCStateMut) {
	q.enqueueInner(state)
	q.len++
}

func (q *DiskStateQueue) Dequeue() *TLCStateMut {
	if q.IsEmpty() {
		return nil
	}
	state := q.dequeueInner()
	q.len--
	return state
}

func (q *DiskStateQueue) SEnqueue(state *TLCStateMut) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enqueueInner(state)
	q.len++
	if q.numWaiting > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskStateQueue) SEnqueueAll(states []*TLCStateMut) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, state := range states {
		q.enqueueInner(state)
		q.len++
	}
	if q.numWaiting > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskStateQueue) SEnqueueVec(states *StateVec) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if states == nil {
		return
	}
	for i := 0; i < states.Size(); i++ {
		state := states.At(i)
		if state != nil {
			q.enqueueInner(state)
			q.len++
		}
	}
	if q.numWaiting > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskStateQueue) SPeek() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.isAvailLocked() {
		return nil
	}
	return q.peekInner()
}

func (q *DiskStateQueue) SDequeue() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.isAvailLocked() {
		return nil
	}
	state := q.dequeueInner()
	q.len--
	return state
}

func (q *DiskStateQueue) SDequeueMany(cnt int) []*TLCStateMut {
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
		q.len--
	}
	return out
}

func (q *DiskStateQueue) FinishAll() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.finish = true
	q.cond.Broadcast()
}

func (q *DiskStateQueue) SuspendAll() bool {
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

func (q *DiskStateQueue) ResumeAll() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stop = false
	q.cond.Broadcast()
}

func (q *DiskStateQueue) ResumeAllStuck() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cond.Broadcast()
}

func (q *DiskStateQueue) Size() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.len
}

func (q *DiskStateQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.len < 1
}

func (q *DiskStateQueue) BeginChkpt() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if err := os.MkdirAll(q.diskdir, 0o755); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(q.diskdir, "queue.tmp"))
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	if err := out.WriteLongNat(q.len); err != nil {
		_ = out.Close()
		return err
	}
	for _, value := range []int{q.loPool, q.hiPool, q.enqIndex, q.deqIndex} {
		if err := out.WriteInt(int32(value)); err != nil {
			_ = out.Close()
			return err
		}
	}
	for i := 0; i < q.enqIndex; i++ {
		if err := q.enqBuf[i].Write(out); err != nil {
			_ = out.Close()
			return err
		}
	}
	for i := q.deqIndex; i < len(q.deqBuf); i++ {
		if q.deqBuf[i] == nil {
			continue
		}
		if err := q.deqBuf[i].Write(out); err != nil {
			_ = out.Close()
			return err
		}
	}
	q.newLastLoPool = q.loPool - 1
	return out.Close()
}

func (q *DiskStateQueue) CommitChkpt() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := q.lastLoPool; i < q.newLastLoPool; i++ {
		if err := os.Remove(q.poolName(i)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	q.lastLoPool = q.newLastLoPool
	oldName := filepath.Join(q.diskdir, "queue.chkpt")
	newName := filepath.Join(q.diskdir, "queue.tmp")
	if err := os.Remove(oldName); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(newName, oldName)
}

func (q *DiskStateQueue) Recover() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	file, err := os.Open(filepath.Join(q.diskdir, "queue.chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	defer in.Close()
	length, err := in.ReadLongNat()
	if err != nil {
		return err
	}
	q.len = length
	values := []*int{&q.loPool, &q.hiPool, &q.enqIndex, &q.deqIndex}
	for _, ptr := range values {
		value, err := in.ReadInt()
		if err != nil {
			return err
		}
		*ptr = int(value)
	}
	q.lastLoPool = q.loPool - 1
	for i := range q.enqBuf {
		q.enqBuf[i] = nil
	}
	for i := range q.deqBuf {
		q.deqBuf[i] = nil
	}
	for i := 0; i < q.enqIndex; i++ {
		state := NewEmptyState()
		if err := state.Read(in); err != nil {
			return err
		}
		q.enqBuf[i] = state
	}
	for i := q.deqIndex; i < len(q.deqBuf); i++ {
		state := NewEmptyState()
		if err := state.Read(in); err != nil {
			if errors.Is(err, io.EOF) {
				return io.ErrUnexpectedEOF
			}
			return err
		}
		q.deqBuf[i] = state
	}
	return nil
}

func (q *DiskStateQueue) Delete() error {
	q.FinishAll()
	if q.diskdir == "" {
		return nil
	}
	return os.RemoveAll(q.diskdir)
}

func (q *DiskStateQueue) enqueueInner(state *TLCStateMut) {
	if q.enqIndex == len(q.enqBuf) {
		if err := q.spillEnqueueBuffer(); err != nil {
			panic(err)
		}
	}
	q.enqBuf[q.enqIndex] = state
	q.enqIndex++
}

func (q *DiskStateQueue) dequeueInner() *TLCStateMut {
	if q.deqIndex == len(q.deqBuf) {
		if err := q.fillDequeueBuffer(); err != nil {
			panic(err)
		}
	}
	state := q.deqBuf[q.deqIndex]
	q.deqBuf[q.deqIndex] = nil
	q.deqIndex++
	return state
}

func (q *DiskStateQueue) peekInner() *TLCStateMut {
	if q.deqIndex == len(q.deqBuf) {
		if err := q.fillDequeueBuffer(); err != nil {
			panic(err)
		}
	}
	return q.deqBuf[q.deqIndex]
}

func (q *DiskStateQueue) fillDequeueBuffer() error {
	if q.loPool+1 <= q.hiPool {
		if err := q.readPool(q.loPool, q.deqBuf); err != nil {
			return err
		}
		q.deqIndex = 0
		q.loPool++
		return nil
	}
	q.deqIndex = len(q.deqBuf) - q.enqIndex
	for i := range q.deqBuf {
		q.deqBuf[i] = nil
	}
	copy(q.deqBuf[q.deqIndex:], q.enqBuf[:q.enqIndex])
	for i := 0; i < q.enqIndex; i++ {
		q.enqBuf[i] = nil
	}
	q.enqIndex = 0
	return nil
}

func (q *DiskStateQueue) spillEnqueueBuffer() error {
	if err := os.MkdirAll(q.diskdir, 0o755); err != nil {
		return err
	}
	if err := q.writePool(q.hiPool, q.enqBuf); err != nil {
		return err
	}
	for i := range q.enqBuf {
		q.enqBuf[i] = nil
	}
	q.hiPool++
	q.enqIndex = 0
	return nil
}

func (q *DiskStateQueue) writePool(pool int, states []*TLCStateMut) error {
	file, err := os.Create(q.poolName(pool))
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	for _, state := range states {
		if state == nil {
			state = NewEmptyState()
		}
		if err := state.Write(out); err != nil {
			_ = out.Close()
			return err
		}
	}
	return out.Close()
}

func (q *DiskStateQueue) readPool(pool int, states []*TLCStateMut) error {
	file, err := os.Open(q.poolName(pool))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	defer in.Close()
	for i := range states {
		state := NewEmptyState()
		if err := state.Read(in); err != nil {
			return err
		}
		states[i] = state
	}
	return nil
}

func (q *DiskStateQueue) isAvailLocked() bool {
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

func (q *DiskStateQueue) poolName(pool int) string {
	return filepath.Join(q.diskdir, fmt.Sprint(pool))
}
