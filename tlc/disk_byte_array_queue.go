package tlc

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type DiskByteArrayQueue struct {
	mu            sync.Mutex
	cond          *sync.Cond
	len           int64
	numWaiting    atomic.Int32
	finish        atomic.Bool
	stop          bool
	suspend       stateQueueSuspendBarrier
	diskdir       string
	deqBuf        [][]byte
	enqBuf        [][]byte
	deqIndex      int
	enqIndex      int
	reader        *ByteArrayPoolReader
	writer        *ByteArrayPoolWriter
	loPool        int
	hiPool        int
	lastLoPool    int
	newLastLoPool int
	loFile        string
	cleaner       *ByteArrayPoolCleaner
}

func NewDiskByteArrayQueue(metaDir string) *DiskByteArrayQueue {
	bufSize := diskByteArrayQueueSizeSnapshot.get()
	if bufSize < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(bufSize)))
	}
	q := &DiskByteArrayQueue{
		diskdir:  metaDir,
		deqBuf:   make([][]byte, bufSize),
		enqBuf:   make([][]byte, bufSize),
		deqIndex: bufSize,
		loPool:   1,
	}
	q.cond = sync.NewCond(&q.mu)
	q.suspend.ensure()
	q.reader = NewByteArrayPoolReader(bufSize, q.poolName(0))
	q.reader.Start()
	q.writer = NewByteArrayPoolWriter(bufSize, q.reader)
	q.writer.Start()
	q.cleaner = NewByteArrayPoolCleaner(q)
	q.cleaner.Start()
	q.loFile = q.poolName(q.loPool)
	return q
}

func (q *DiskByteArrayQueue) requireDiskByteArrayQueue() {
	if q == nil {
		panic(NewNullPointerException())
	}
}

func (q *DiskByteArrayQueue) SetDiskDir(diskdir string) {
	q.requireDiskByteArrayQueue()
	q.diskdir = diskdir
	q.loFile = q.poolName(q.loPool)
	if q.reader != nil {
		q.reader.Restart(q.poolName(q.lastLoPool), q.lastLoPool < q.hiPool)
	}
}

func (q *DiskByteArrayQueue) Enqueue(state *TLCStateMut) {
	q.requireDiskByteArrayQueue()
	q.enqueueRaw(mustStateToBytes(state))
	stateQueueLengthAdd(&q.len, 1)
}

func (q *DiskByteArrayQueue) Dequeue() *TLCStateMut {
	q.requireDiskByteArrayQueue()
	if q.IsEmpty() {
		return nil
	}
	bytes := q.dequeueRaw()
	stateQueueLengthAdd(&q.len, -1)
	if bytes == nil {
		return nil
	}
	return mustBytesToState(bytes)
}

func (q *DiskByteArrayQueue) SEnqueue(state *TLCStateMut) {
	q.requireDiskByteArrayQueue()
	raw := mustStateToBytes(state)
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enqueueRaw(raw)
	stateQueueLengthAdd(&q.len, 1)
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskByteArrayQueue) SEnqueueAll(states []*TLCStateMut) {
	q.requireDiskByteArrayQueue()
	if states == nil {
		panic(NewNullPointerException())
	}
	raw := make([][]byte, 0, len(states))
	for _, state := range states {
		raw = append(raw, mustStateToBytes(state))
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, state := range raw {
		q.enqueueRaw(state)
	}
	stateQueueLengthAdd(&q.len, int64(len(raw)))
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskByteArrayQueue) SEnqueueVec(states *StateVec) {
	q.requireDiskByteArrayQueue()
	if states == nil {
		panic(NewNullPointerException())
	}
	q.SEnqueueVecCount(states, states.Size())
}

func (q *DiskByteArrayQueue) SEnqueueVecCount(states *StateVec, count int) {
	q.requireDiskByteArrayQueue()
	count = int(int32(count))
	if count == 0 {
		return
	}
	if count < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(count)))
	}
	raw := make([][]byte, count)
	out := newByteValueOutputStream()
	n := count
	for i := 0; i < states.Size(); i++ {
		if state := states.At(i); state != nil {
			encoded := out.writeState(state)
			index := n - 1
			if index < 0 || index >= len(raw) {
				panic(NewArrayIndexOutOfBoundsException(index, len(raw)))
			}
			raw[index] = encoded
			n--
		}
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, state := range raw {
		q.enqueueRaw(state)
	}
	stateQueueLengthAdd(&q.len, int64(len(raw)))
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskByteArrayQueue) SPeek() *TLCStateMut {
	q.requireDiskByteArrayQueue()
	raw := q.sPeekRaw()
	if raw == nil {
		return nil
	}
	return mustBytesToState(raw)
}

func (q *DiskByteArrayQueue) sPeekRaw() []byte {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.isAvailLocked() {
		return nil
	}
	return q.peekRaw()
}

func (q *DiskByteArrayQueue) SDequeue() *TLCStateMut {
	q.requireDiskByteArrayQueue()
	raw := q.sDequeueRaw()
	if raw == nil {
		return nil
	}
	return mustBytesToState(raw)
}

func (q *DiskByteArrayQueue) sDequeueRaw() []byte {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.isAvailLocked() {
		return nil
	}
	raw := q.dequeueRaw()
	if raw == nil {
		panic(NewAssertionError("Null state found on queue"))
	}
	stateQueueLengthAdd(&q.len, -1)
	return raw
}

func (q *DiskByteArrayQueue) SDequeueMany(cnt int) []*TLCStateMut {
	q.requireDiskByteArrayQueue()
	cnt = int(int32(cnt))
	raw := q.sDequeueManyRaw(cnt)
	if raw == nil {
		return nil
	}
	// The source wrapper uses the requested count, even when its raw helper
	// clips the available batch. Preserve its decode/access failure boundary.
	out := make([]*TLCStateMut, cnt)
	for i := range out {
		if i >= len(raw) {
			panic(NewArrayIndexOutOfBoundsException(i, len(raw)))
		}
		out[i] = mustBytesToState(raw[i])
	}
	return out
}

func (q *DiskByteArrayQueue) sDequeueManyRaw(cnt int) [][]byte {
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
	out := make([][]byte, 0, cnt)
	for len(out) < cnt && stateQueueLengthLoad(&q.len) > 0 {
		out = append(out, q.dequeueRaw())
		stateQueueLengthAdd(&q.len, -1)
	}
	return out
}

func (q *DiskByteArrayQueue) FinishAll() {
	q.requireDiskByteArrayQueue()
	q.mu.Lock()
	q.finish.Store(true)
	q.cond.Broadcast()
	q.suspend.signal()
	q.mu.Unlock()
	if q.writer != nil {
		q.writer.SetFinished()
	}
	if q.reader != nil {
		q.reader.SetFinished()
	}
	if q.cleaner != nil {
		q.cleaner.SetFinished()
	}
}

func (q *DiskByteArrayQueue) SuspendAll() bool {
	q.requireDiskByteArrayQueue()
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

func (q *DiskByteArrayQueue) ResumeAll() {
	q.requireDiskByteArrayQueue()
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stop = false
	q.cond.Broadcast()
}

func (q *DiskByteArrayQueue) WakeAllWaiters() {
	q.requireDiskByteArrayQueue()
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cond.Broadcast()
}

func (q *DiskByteArrayQueue) ResumeAllStuck() {
	q.requireDiskByteArrayQueue()
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

func (q *DiskByteArrayQueue) Size() int64 {
	q.requireDiskByteArrayQueue()
	return stateQueueLengthLoad(&q.len)
}

func (q *DiskByteArrayQueue) IsEmpty() bool {
	q.requireDiskByteArrayQueue()
	return stateQueueLengthLoad(&q.len) < 1
}

func (q *DiskByteArrayQueue) BeginChkpt() error {
	q.requireDiskByteArrayQueue()
	if q.cleaner != nil {
		q.cleaner.FinishAndWait()
	}
	file, err := openPoolFile(q.queuePath("queue.tmp"), true)
	if err != nil {
		return err
	}
	out := NewBufferedDataOutputStream(file)
	if err := out.WriteLong(stateQueueLengthLoad(&q.len)); err != nil {
		return err
	}
	for _, value := range []int{q.loPool, q.hiPool, q.enqIndex, q.deqIndex} {
		if err := out.WriteInt(int32(value)); err != nil {
			return err
		}
	}
	if err := writeByteArrayEntries(out, q.enqBuf[:q.enqIndex]); err != nil {
		return err
	}
	if err := writeByteArrayEntries(out, q.deqBuf[q.deqIndex:]); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	q.newLastLoPool = q.loPool - 1
	return nil
}

func (q *DiskByteArrayQueue) CommitChkpt() error {
	q.requireDiskByteArrayQueue()
	for i := q.lastLoPool; i < q.newLastLoPool; i++ {
		oldPool := q.poolName(i)
		if err := os.Remove(oldPool); err != nil {
			return NewIOException(fmt.Sprintf("DiskStateQueue.commitChkpt: cannot delete %s", oldPool))
		}
	}
	q.lastLoPool = q.newLastLoPool
	oldName := q.queuePath("queue.chkpt")
	newName := q.queuePath("queue.tmp")
	if _, err := os.Stat(oldName); err == nil {
		if err := os.Remove(oldName); err != nil {
			return NewIOException(fmt.Sprintf("DiskStateQueue.commitChkpt: cannot delete %s", oldName))
		}
	}
	if err := os.Rename(newName, oldName); err != nil {
		return NewIOException(fmt.Sprintf("DiskStateQueue.commitChkpt: cannot delete %s", oldName))
	}
	return nil
}

func (q *DiskByteArrayQueue) Recover() error {
	q.requireDiskByteArrayQueue()
	file, err := openPoolFile(q.queuePath("queue.chkpt"), false)
	if err != nil {
		return err
	}
	in, err := NewBufferedDataInputStream(file)
	if err != nil {
		return err
	}
	length, err := in.ReadLong()
	if err != nil {
		return err
	}
	stateQueueLengthStore(&q.len, length)
	values := []*int{&q.loPool, &q.hiPool, &q.enqIndex, &q.deqIndex}
	for _, ptr := range values {
		value, err := in.ReadInt()
		if err != nil {
			return err
		}
		*ptr = int(value)
	}
	q.lastLoPool = q.loPool - 1
	if err := readByteArrayEntries(in, q.enqBuf[:q.enqIndex]); err != nil {
		return err
	}
	if err := readByteArrayEntries(in, q.deqBuf[q.deqIndex:]); err != nil {
		return err
	}
	if err := in.Close(); err != nil {
		return err
	}
	if q.reader != nil {
		q.reader.Restart(q.poolName(q.lastLoPool), q.lastLoPool < q.hiPool)
	}
	q.loFile = q.poolName(q.loPool)
	return nil
}

func (q *DiskByteArrayQueue) Delete() error {
	q.requireDiskByteArrayQueue()
	q.FinishAll()
	return deleteQueueDirLikeJava(q.diskdir)
}

func (q *DiskByteArrayQueue) enqueueRaw(state []byte) {
	if q.enqIndex == len(q.enqBuf) {
		if err := q.spillEnqueueBuffer(); err != nil {
			panic(err)
		}
	}
	index := q.enqIndex
	q.enqIndex = int(int32(index) + 1)
	if index < 0 || index >= len(q.enqBuf) {
		panic(NewArrayIndexOutOfBoundsException(index, len(q.enqBuf)))
	}
	q.enqBuf[index] = state
}

func (q *DiskByteArrayQueue) dequeueRaw() []byte {
	if q.deqIndex == len(q.deqBuf) {
		if err := q.fillDequeueBuffer(); err != nil {
			panic(err)
		}
	}
	index := q.deqIndex
	q.deqIndex = int(int32(index) + 1)
	if index < 0 || index >= len(q.deqBuf) {
		panic(NewArrayIndexOutOfBoundsException(index, len(q.deqBuf)))
	}
	return q.deqBuf[index]
}

func (q *DiskByteArrayQueue) peekRaw() []byte {
	if q.deqIndex == len(q.deqBuf) {
		if err := q.fillDequeueBuffer(); err != nil {
			panic(err)
		}
	}
	if q.deqIndex < 0 || q.deqIndex >= len(q.deqBuf) {
		panic(NewArrayIndexOutOfBoundsException(q.deqIndex, len(q.deqBuf)))
	}
	return q.deqBuf[q.deqIndex]
}

func (q *DiskByteArrayQueue) fillDequeueBuffer() (err error) {
	defer catchDiskQueuePoolFailure(ECSystemErrorReadingStates, &err)
	if q.loPool+1 <= q.hiPool {
		if q.loPool+1 >= q.hiPool && q.writer != nil {
			if err := q.writer.EnsureWritten(); err != nil {
				return err
			}
		}
		buf, err := q.reader.DoWork(q.deqBuf, q.loFile)
		if err != nil {
			return err
		}
		q.deqBuf = buf
		q.deqIndex = 0
		q.loPool++
		q.loFile = q.poolName(q.loPool)
		q.maybeCleanByteArrayPools()
		return nil
	}
	if q.writer != nil {
		if err := q.writer.EnsureWritten(); err != nil {
			return err
		}
	}
	if q.reader != nil {
		buf, err := q.reader.GetCache(q.deqBuf, q.loFile)
		if err != nil {
			return err
		}
		if buf != nil {
			q.deqBuf = buf
			q.deqIndex = 0
			q.loPool++
			q.loFile = q.poolName(q.loPool)
			q.maybeCleanByteArrayPools()
			return nil
		}
	}
	q.deqIndex = len(q.deqBuf) - q.enqIndex
	copy(q.deqBuf[q.deqIndex:], q.enqBuf[:q.enqIndex])
	q.enqIndex = 0
	q.maybeCleanByteArrayPools()
	return nil
}

func (q *DiskByteArrayQueue) spillEnqueueBuffer() (err error) {
	defer catchDiskQueuePoolFailure(ECSystemErrorWritingStates, &err)
	buf, err := q.writer.DoWork(q.enqBuf, q.poolName(q.hiPool))
	if err != nil {
		return err
	}
	q.enqBuf = buf
	q.hiPool++
	q.enqIndex = 0
	return nil
}

func (q *DiskByteArrayQueue) isAvailLocked() bool {
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

func (q *DiskByteArrayQueue) needsWaiting() bool {
	return int(q.numWaiting.Load()) < NumWorkers()
}

func (q *DiskByteArrayQueue) poolName(pool int) string {
	return q.queuePath(intToDecimal(pool))
}

func (q *DiskByteArrayQueue) queuePath(name string) string {
	return q.diskdir + string(os.PathSeparator) + name
}

func (q *DiskByteArrayQueue) maybeCleanByteArrayPools() {
	if q.cleaner == nil {
		return
	}
	if q.loPool-q.lastLoPool > diskStateQueueCleanerThreshold {
		q.cleaner.DeleteUpTo(q.loPool - 1)
	}
}

type ByteArrayPoolCleaner struct {
	mu         sync.Mutex
	cond       *sync.Cond
	queue      *DiskByteArrayQueue
	deleteUpTo int
	finished   bool
	started    atomic.Bool
	done       chan struct{}
}

func NewByteArrayPoolCleaner(queue *DiskByteArrayQueue) *ByteArrayPoolCleaner {
	c := &ByteArrayPoolCleaner{
		queue: queue,
		done:  make(chan struct{}),
	}
	c.cond = sync.NewCond(&c.mu)
	return c
}

func (c *ByteArrayPoolCleaner) requireByteArrayPoolCleaner() {
	if c == nil {
		panic(NewNullPointerException())
	}
}

func (c *ByteArrayPoolCleaner) Start() {
	c.requireByteArrayPoolCleaner()
	if !c.started.CompareAndSwap(false, true) {
		panic(NewIllegalThreadStateException())
	}
	go c.run()
}

func (c *ByteArrayPoolCleaner) DeleteUpTo(pool int) {
	c.requireByteArrayPoolCleaner()
	pool = int(int32(pool))
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.finished {
		return
	}
	if pool > c.deleteUpTo {
		c.deleteUpTo = pool
	}
	c.cond.Signal()
}

func (c *ByteArrayPoolCleaner) SetFinished() {
	c.requireByteArrayPoolCleaner()
	c.mu.Lock()
	c.finished = true
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *ByteArrayPoolCleaner) FinishAndWait() {
	c.requireByteArrayPoolCleaner()
	c.SetFinished()
	<-c.done
}

func (c *ByteArrayPoolCleaner) run() {
	defer close(c.done)
	defer func() {
		if failure := recover(); failure != nil {
			statePoolFailureExit(ECSystemErrorCleaningPool, "", failure)
		}
	}()
	for {
		c.mu.Lock()
		for !c.finished && c.deleteUpTo <= 0 {
			c.cond.Wait()
		}
		if c.finished {
			c.mu.Unlock()
			return
		}
		target := c.deleteUpTo
		c.deleteUpTo = 0
		c.mu.Unlock()

		q := c.queue
		if q == nil {
			continue
		}
		q.mu.Lock()
		start := q.lastLoPool
		q.mu.Unlock()
		if target <= start {
			continue
		}
		for i := start; i < target; i++ {
			name := q.poolName(i)
			if err := os.Remove(name); err != nil {
				canonical, err := statePoolCanonicalPath(name)
				if err != nil {
					panic(err)
				}
				PrintWarning(ECSystemErrorCleaningPool, canonical)
			}
		}
		q.mu.Lock()
		if q.lastLoPool < target {
			q.lastLoPool = target
		}
		q.mu.Unlock()
	}
}

type ByteArrayPoolReader struct {
	mu       sync.Mutex
	cond     *sync.Cond
	buf      [][]byte
	poolFile string
	isFull   bool
	canRead  bool
	finished bool
	started  atomic.Bool
	done     chan struct{}
}

func NewByteArrayPoolReader(bufSize int, file string) *ByteArrayPoolReader {
	bufSize = int(int32(bufSize))
	if bufSize < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(bufSize)))
	}
	r := &ByteArrayPoolReader{buf: make([][]byte, bufSize), poolFile: file, done: make(chan struct{})}
	r.cond = sync.NewCond(&r.mu)
	return r
}

func (r *ByteArrayPoolReader) requireByteArrayPoolReader() {
	if r == nil {
		panic(NewNullPointerException())
	}
}

func (r *ByteArrayPoolReader) Start() {
	r.requireByteArrayPoolReader()
	if !r.started.CompareAndSwap(false, true) {
		panic(NewIllegalThreadStateException())
	}
	go r.run()
}

func (r *ByteArrayPoolReader) Wakeup() {
	r.requireByteArrayPoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.canRead = true
	r.cond.Signal()
}

func (r *ByteArrayPoolReader) Restart(file string, canRead bool) {
	r.requireByteArrayPoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.poolFile = file
	r.isFull = false
	r.canRead = canRead
	r.cond.Signal()
}

func (r *ByteArrayPoolReader) DoWork(deqBuf [][]byte, file string) ([][]byte, error) {
	r.requireByteArrayPoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isFull {
		if r.poolFile != "" {
			panic(NewAssertionErrorInt(int32(ECSystemFileNull)))
		}
		out := r.buf
		r.buf = deqBuf
		r.poolFile = file
		r.isFull = false
		r.canRead = true
		r.cond.Signal()
		return out, nil
	}
	if r.poolFile != "" {
		if err := readByteArrayPoolFile(r.poolFile, deqBuf); err != nil {
			return nil, err
		}
		r.poolFile = file
		r.canRead = true
		r.cond.Signal()
		return deqBuf, nil
	}
	// Source's direct branch opens the null pending File, not the argument.
	panic(NewNullPointerException())
}

func (r *ByteArrayPoolReader) GetCache(deqBuf [][]byte, file string) ([][]byte, error) {
	r.requireByteArrayPoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isFull {
		if r.poolFile != "" {
			panic(NewAssertionErrorInt(int32(ECSystemFileNull)))
		}
		out := r.buf
		r.buf = deqBuf
		r.poolFile = file
		r.isFull = false
		r.canRead = false
		return out, nil
	}
	if r.poolFile != "" && r.canRead {
		if err := readByteArrayPoolFile(r.poolFile, deqBuf); err != nil {
			return nil, err
		}
		r.poolFile = file
		r.canRead = false
		return deqBuf, nil
	}
	return nil, nil
}

func (r *ByteArrayPoolReader) SetFinished() {
	r.requireByteArrayPoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finished = true
	r.cond.Broadcast()
}

func (r *ByteArrayPoolReader) run() {
	defer close(r.done)
	defer func() {
		if failure := recover(); failure != nil {
			r.mu.Lock()
			name := r.poolFile
			r.mu.Unlock()
			statePoolFailureExit(ECSystemErrorReadingPool, name, failure)
		}
	}()
	r.mu.Lock()
	defer r.mu.Unlock()
	for {
		for r.poolFile == "" || r.isFull || !r.canRead {
			if r.finished {
				return
			}
			r.cond.Wait()
			if r.finished {
				return
			}
		}
		if err := readByteArrayPoolFile(r.poolFile, r.buf); err != nil {
			panic(err)
		}
		r.poolFile = ""
		r.isFull = true
	}
}

type ByteArrayPoolWriter struct {
	mu       sync.Mutex
	cond     *sync.Cond
	buf      [][]byte
	poolFile string
	reader   *ByteArrayPoolReader
	finished bool
	started  atomic.Bool
	done     chan struct{}
}

func NewByteArrayPoolWriter(bufSize int, reader *ByteArrayPoolReader) *ByteArrayPoolWriter {
	bufSize = int(int32(bufSize))
	if bufSize < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(bufSize)))
	}
	w := &ByteArrayPoolWriter{buf: make([][]byte, bufSize), reader: reader, done: make(chan struct{})}
	w.cond = sync.NewCond(&w.mu)
	return w
}

func (w *ByteArrayPoolWriter) requireByteArrayPoolWriter() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

func (w *ByteArrayPoolWriter) Start() {
	w.requireByteArrayPoolWriter()
	if !w.started.CompareAndSwap(false, true) {
		panic(NewIllegalThreadStateException())
	}
	go w.run()
}

func (w *ByteArrayPoolWriter) DoWork(enqBuf [][]byte, file string) ([][]byte, error) {
	w.requireByteArrayPoolWriter()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.poolFile != "" {
		if err := writeByteArrayPoolFile(w.poolFile, w.buf); err != nil {
			return nil, err
		}
	}
	out := w.buf
	w.buf = enqBuf
	w.poolFile = file
	w.cond.Signal()
	return out, nil
}

func (w *ByteArrayPoolWriter) EnsureWritten() error {
	w.requireByteArrayPoolWriter()
	w.mu.Lock()
	defer w.mu.Unlock()
	for w.poolFile != "" {
		w.cond.Wait()
	}
	return nil
}

func (w *ByteArrayPoolWriter) SetFinished() {
	w.requireByteArrayPoolWriter()
	w.mu.Lock()
	defer w.mu.Unlock()
	w.finished = true
	w.cond.Broadcast()
}

func (w *ByteArrayPoolWriter) run() {
	defer close(w.done)
	defer func() {
		if failure := recover(); failure != nil {
			statePoolFailureExit(ECSystemErrorWritingPool, "", failure)
		}
	}()
	w.mu.Lock()
	defer w.mu.Unlock()
	for {
		for w.poolFile == "" {
			if w.finished {
				return
			}
			w.cond.Wait()
		}
		if err := writeByteArrayPoolFile(w.poolFile, w.buf); err != nil {
			panic(err)
		}
		w.poolFile = ""
		w.cond.Broadcast()
		if w.reader != nil {
			w.reader.Wakeup()
		}
	}
}

type byteValueOutputStream struct {
	buffer bytes.Buffer
	stream *ValueOutputStream
}

func newByteValueOutputStream() *byteValueOutputStream {
	out := &byteValueOutputStream{}
	out.buffer.Grow(16)
	out.stream = NewValueOutputStreamWithoutHandles(&out.buffer)
	return out
}

func (out *byteValueOutputStream) writeState(state *TLCStateMut) []byte {
	if err := state.Write(out.stream); err != nil {
		panic(err)
	}
	encoded := append([]byte(nil), out.buffer.Bytes()...)
	out.buffer.Reset()
	return encoded
}

func mustStateToBytes(state *TLCStateMut) []byte {
	return newByteValueOutputStream().writeState(state)
}

func mustBytesToState(raw []byte) *TLCStateMut {
	state := newStateFromEmptyFactory()
	if err := state.Read(NewByteValueInputStream(raw)); err != nil {
		panic(err)
	}
	return state
}

func writeByteArrayPoolFile(name string, entries [][]byte) error {
	file, err := openPoolFile(name, true)
	if err != nil {
		return err
	}
	out := NewBufferedDataOutputStream(file)
	if entries == nil {
		panic(NewNullPointerException())
	}
	if err := writeByteArrayEntries(out, entries); err != nil {
		return err
	}
	return out.Close()
}

func readByteArrayPoolFile(name string, entries [][]byte) error {
	file, err := openPoolFile(name, false)
	if err != nil {
		return err
	}
	in, err := NewBufferedDataInputStream(file)
	if err != nil {
		return err
	}
	if entries == nil {
		panic(NewNullPointerException())
	}
	if err := readByteArrayEntries(in, entries); err != nil {
		return err
	}
	return in.Close()
}

func writeByteArrayEntries(out *BufferedDataOutputStream, entries [][]byte) error {
	for _, entry := range entries {
		if entry == nil {
			panic(NewNullPointerException())
		}
		if err := out.WriteInt(int32(len(entry))); err != nil {
			return err
		}
		if _, err := out.Write(entry); err != nil {
			return err
		}
	}
	return nil
}

func readByteArrayEntries(in *BufferedDataInputStream, entries [][]byte) error {
	for i := range entries {
		length, err := in.ReadInt()
		if err != nil {
			return err
		}
		if length < 0 {
			panic(NewNegativeArraySizeException(fmt.Sprint(length)))
		}
		entries[i] = make([]byte, int(length))
		// Source read(byte[]) publishes the slot first and ignores its count.
		// A short final entry therefore retains zero padding rather than failing.
		if _, err := in.ReadBytes(entries[i], 0, len(entries[i])); err != nil {
			return err
		}
	}
	return nil
}

func intToDecimal(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var digits [20]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}
