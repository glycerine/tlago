package tlc

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	bufSize := diskStateQueueBufferSize()
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

func (q *DiskByteArrayQueue) SetDiskDir(diskdir string) {
	q.diskdir = diskdir
	q.loFile = q.poolName(q.loPool)
	if q.reader != nil {
		q.reader.Restart(q.poolName(q.lastLoPool), q.lastLoPool < q.hiPool)
	}
}

func (q *DiskByteArrayQueue) Enqueue(state *TLCStateMut) {
	q.enqueueRaw(mustStateToBytes(state))
	q.len++
}

func (q *DiskByteArrayQueue) Dequeue() *TLCStateMut {
	if q.IsEmpty() {
		return nil
	}
	bytes := q.dequeueRaw()
	q.len--
	return mustBytesToState(bytes)
}

func (q *DiskByteArrayQueue) SEnqueue(state *TLCStateMut) {
	raw := mustStateToBytes(state)
	q.mu.Lock()
	defer q.mu.Unlock()
	q.enqueueRaw(raw)
	q.len++
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskByteArrayQueue) SEnqueueAll(states []*TLCStateMut) {
	raw := make([][]byte, 0, len(states))
	for _, state := range states {
		raw = append(raw, mustStateToBytes(state))
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, state := range raw {
		q.enqueueRaw(state)
	}
	q.len += int64(len(raw))
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskByteArrayQueue) SEnqueueVec(states *StateVec) {
	if states == nil || states.Size() == 0 {
		return
	}
	raw := make([][]byte, states.Size())
	n := states.Size()
	for i := 0; i < states.Size(); i++ {
		if state := states.At(i); state != nil {
			raw[n-1] = mustStateToBytes(state)
			n--
		}
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, state := range raw {
		q.enqueueRaw(state)
	}
	q.len += int64(len(raw))
	if q.numWaiting.Load() > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskByteArrayQueue) SPeek() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.isAvailLocked() {
		return nil
	}
	return mustBytesToState(q.peekRaw())
}

func (q *DiskByteArrayQueue) SDequeue() *TLCStateMut {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.isAvailLocked() {
		return nil
	}
	raw := q.dequeueRaw()
	q.len--
	return mustBytesToState(raw)
}

func (q *DiskByteArrayQueue) SDequeueMany(cnt int) []*TLCStateMut {
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
		out = append(out, mustBytesToState(q.dequeueRaw()))
		q.len--
	}
	return out
}

func (q *DiskByteArrayQueue) FinishAll() {
	q.mu.Lock()
	q.finish.Store(true)
	if q.writer != nil {
		q.writer.SetFinished()
	}
	if q.reader != nil {
		q.reader.SetFinished()
	}
	if q.cleaner != nil {
		q.cleaner.SetFinished()
	}
	q.cond.Broadcast()
	q.mu.Unlock()
	q.suspend.signal()
}

func (q *DiskByteArrayQueue) SuspendAll() bool {
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
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stop = false
	q.cond.Broadcast()
}

func (q *DiskByteArrayQueue) WakeAllWaiters() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cond.Broadcast()
}

func (q *DiskByteArrayQueue) ResumeAllStuck() {
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

func (q *DiskByteArrayQueue) Size() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.len
}

func (q *DiskByteArrayQueue) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.len < 1
}

func (q *DiskByteArrayQueue) BeginChkpt() error {
	if q.cleaner != nil {
		q.cleaner.FinishAndWait()
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	file, err := os.Create(q.queuePath("queue.tmp"))
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	if err := out.WriteLong(q.len); err != nil {
		_ = out.Close()
		return err
	}
	for _, value := range []int{q.loPool, q.hiPool, q.enqIndex, q.deqIndex} {
		if err := out.WriteInt(int32(value)); err != nil {
			_ = out.Close()
			return err
		}
	}
	if err := writeByteArrayEntries(out, q.enqBuf[:q.enqIndex]); err != nil {
		_ = out.Close()
		return err
	}
	if err := writeByteArrayEntries(out, q.deqBuf[q.deqIndex:]); err != nil {
		_ = out.Close()
		return err
	}
	q.newLastLoPool = q.loPool - 1
	return out.Close()
}

func (q *DiskByteArrayQueue) CommitChkpt() error {
	q.mu.Lock()
	defer q.mu.Unlock()
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
	q.mu.Lock()
	defer q.mu.Unlock()
	file, err := os.Open(q.queuePath("queue.chkpt"))
	if err != nil {
		return err
	}
	in, err := NewBufferedDataInputStream(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	closed := false
	defer func() {
		if !closed {
			_ = file.Close()
		}
	}()
	length, err := in.ReadLong()
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
	if err := readByteArrayEntries(in, q.enqBuf[:q.enqIndex]); err != nil {
		return err
	}
	if err := readByteArrayEntries(in, q.deqBuf[q.deqIndex:]); err != nil {
		return err
	}
	closed = true
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
	q.FinishAll()
	return deleteQueueDirLikeJava(q.diskdir)
}

func (q *DiskByteArrayQueue) enqueueRaw(state []byte) {
	if q.enqIndex == len(q.enqBuf) {
		if err := q.spillEnqueueBuffer(); err != nil {
			panic(err)
		}
	}
	q.enqBuf[q.enqIndex] = state
	q.enqIndex++
}

func (q *DiskByteArrayQueue) dequeueRaw() []byte {
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

func (q *DiskByteArrayQueue) peekRaw() []byte {
	if q.deqIndex == len(q.deqBuf) {
		if err := q.fillDequeueBuffer(); err != nil {
			panic(err)
		}
	}
	return q.deqBuf[q.deqIndex]
}

func (q *DiskByteArrayQueue) fillDequeueBuffer() error {
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
	clearByteArrayBuffer(q.deqBuf)
	copy(q.deqBuf[q.deqIndex:], q.enqBuf[:q.enqIndex])
	clearByteArrayBuffer(q.enqBuf[:q.enqIndex])
	q.enqIndex = 0
	q.maybeCleanByteArrayPools()
	return nil
}

func (q *DiskByteArrayQueue) spillEnqueueBuffer() error {
	buf, err := q.writer.DoWork(q.enqBuf, q.poolName(q.hiPool))
	if err != nil {
		return err
	}
	q.enqBuf = buf
	clearByteArrayBuffer(q.enqBuf)
	q.hiPool++
	q.enqIndex = 0
	return nil
}

func (q *DiskByteArrayQueue) isAvailLocked() bool {
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

func (c *ByteArrayPoolCleaner) Start() {
	go c.run()
}

func (c *ByteArrayPoolCleaner) DeleteUpTo(pool int) {
	if c == nil {
		return
	}
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
	if c == nil {
		return
	}
	c.mu.Lock()
	c.finished = true
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *ByteArrayPoolCleaner) FinishAndWait() {
	if c == nil {
		return
	}
	c.SetFinished()
	<-c.done
}

func (c *ByteArrayPoolCleaner) run() {
	defer close(c.done)
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
			if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
				if abs, absErr := filepath.Abs(name); absErr == nil {
					name = abs
				}
				PrintWarning(ECSystemErrorCleaningPool, name)
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
	err      error
}

func NewByteArrayPoolReader(bufSize int, file string) *ByteArrayPoolReader {
	r := &ByteArrayPoolReader{buf: make([][]byte, bufSize), poolFile: file}
	r.cond = sync.NewCond(&r.mu)
	return r
}

func (r *ByteArrayPoolReader) Start() { go r.run() }

func (r *ByteArrayPoolReader) Wakeup() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.canRead = true
	r.cond.Signal()
}

func (r *ByteArrayPoolReader) Restart(file string, canRead bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.poolFile = file
	r.isFull = false
	r.canRead = canRead
	r.cond.Signal()
}

func (r *ByteArrayPoolReader) DoWork(deqBuf [][]byte, file string) ([][]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return nil, r.err
	}
	if r.isFull {
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
	if err := readByteArrayPoolFile(r.poolFile, deqBuf); err != nil {
		return nil, err
	}
	return deqBuf, nil
}

func (r *ByteArrayPoolReader) GetCache(deqBuf [][]byte, file string) ([][]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return nil, r.err
	}
	if r.isFull {
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
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finished = true
	r.cond.Broadcast()
}

func (r *ByteArrayPoolReader) run() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for {
		for r.poolFile == "" || r.isFull || !r.canRead {
			if r.finished {
				return
			}
			r.cond.Wait()
		}
		if err := readByteArrayPoolFile(r.poolFile, r.buf); err != nil {
			r.err = err
			r.cond.Broadcast()
			return
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
	err      error
}

func NewByteArrayPoolWriter(bufSize int, reader *ByteArrayPoolReader) *ByteArrayPoolWriter {
	w := &ByteArrayPoolWriter{buf: make([][]byte, bufSize), reader: reader}
	w.cond = sync.NewCond(&w.mu)
	return w
}

func (w *ByteArrayPoolWriter) Start() { go w.run() }

func (w *ByteArrayPoolWriter) DoWork(enqBuf [][]byte, file string) ([][]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err != nil {
		return nil, w.err
	}
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
	w.mu.Lock()
	defer w.mu.Unlock()
	for w.poolFile != "" {
		w.cond.Wait()
	}
	return w.err
}

func (w *ByteArrayPoolWriter) SetFinished() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.finished = true
	w.cond.Broadcast()
}

func (w *ByteArrayPoolWriter) run() {
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
			w.err = err
			w.cond.Broadcast()
			return
		}
		w.poolFile = ""
		w.cond.Broadcast()
		if w.reader != nil {
			w.reader.Wakeup()
		}
	}
}

func mustStateToBytes(state *TLCStateMut) []byte {
	var buf bytes.Buffer
	out := NewValueOutputStreamWithoutHandles(&buf)
	if state == nil {
		state = NewEmptyState()
	}
	if err := state.Write(out); err != nil {
		panic(err)
	}
	return append([]byte(nil), buf.Bytes()...)
}

func mustBytesToState(raw []byte) *TLCStateMut {
	state := NewEmptyState()
	if err := state.Read(NewByteValueInputStream(raw)); err != nil {
		panic(err)
	}
	return state
}

func writeByteArrayPoolFile(name string, entries [][]byte) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	if err := writeByteArrayEntries(out, entries); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func readByteArrayPoolFile(name string, entries [][]byte) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	in, err := NewBufferedDataInputStream(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	closed := false
	defer func() {
		if !closed {
			_ = file.Close()
		}
	}()
	if err := readByteArrayEntries(in, entries); err != nil {
		return err
	}
	closed = true
	return in.Close()
}

func writeByteArrayEntries(out *ValueOutputStream, entries [][]byte) error {
	for i, entry := range entries {
		if entry == nil {
			return fmt.Errorf("byte-array queue write encountered nil entry at slot %d", i)
		}
		if err := out.WriteInt(int32(len(entry))); err != nil {
			return err
		}
		if _, err := out.WriteRaw(entry); err != nil {
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

func clearByteArrayBuffer(buf [][]byte) {
	for i := range buf {
		buf[i] = nil
	}
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
