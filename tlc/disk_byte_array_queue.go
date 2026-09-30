package tlc

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type DiskByteArrayQueue struct {
	mu            sync.Mutex
	cond          *sync.Cond
	len           int64
	numWaiting    int
	finish        bool
	stop          bool
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
}

func NewDiskByteArrayQueue(metaDir string) *DiskByteArrayQueue {
	if metaDir == "" {
		metaDir = filepath.Join(os.TempDir(), "DiskByteArrayQueue")
	}
	q := &DiskByteArrayQueue{
		diskdir:  metaDir,
		deqBuf:   make([][]byte, diskStateQueueBufferSize),
		enqBuf:   make([][]byte, diskStateQueueBufferSize),
		deqIndex: diskStateQueueBufferSize,
		loPool:   1,
	}
	q.cond = sync.NewCond(&q.mu)
	q.reader = NewByteArrayPoolReader(diskStateQueueBufferSize, q.poolName(0))
	q.reader.Start()
	q.writer = NewByteArrayPoolWriter(diskStateQueueBufferSize, q.reader)
	q.writer.Start()
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
	if q.numWaiting > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskByteArrayQueue) SEnqueueAll(states []*TLCStateMut) {
	raw := make([][]byte, 0, len(states))
	for _, state := range states {
		if state != nil {
			raw = append(raw, mustStateToBytes(state))
		}
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, state := range raw {
		q.enqueueRaw(state)
		q.len++
	}
	if q.numWaiting > 0 && !q.stop {
		q.cond.Broadcast()
	}
}

func (q *DiskByteArrayQueue) SEnqueueVec(states *StateVec) {
	if states == nil || states.Size() == 0 {
		return
	}
	raw := make([][]byte, 0, states.Size())
	for i := 0; i < states.Size(); i++ {
		if state := states.At(i); state != nil {
			raw = append(raw, mustStateToBytes(state))
		}
	}
	for i, j := 0, len(raw)-1; i < j; i, j = i+1, j-1 {
		raw[i], raw[j] = raw[j], raw[i]
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, state := range raw {
		q.enqueueRaw(state)
		q.len++
	}
	if q.numWaiting > 0 && !q.stop {
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
		out = append(out, mustBytesToState(q.dequeueRaw()))
		q.len--
	}
	return out
}

func (q *DiskByteArrayQueue) FinishAll() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.finish = true
	if q.writer != nil {
		q.writer.SetFinished()
	}
	if q.reader != nil {
		q.reader.SetFinished()
	}
	q.cond.Broadcast()
}

func (q *DiskByteArrayQueue) SuspendAll() bool {
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

func (q *DiskByteArrayQueue) ResumeAll() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.stop = false
	q.cond.Broadcast()
}

func (q *DiskByteArrayQueue) ResumeAllStuck() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cond.Broadcast()
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

func (q *DiskByteArrayQueue) Recover() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	file, err := os.Open(filepath.Join(q.diskdir, "queue.chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	defer in.Close()
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
	clearByteArrayBuffer(q.enqBuf)
	clearByteArrayBuffer(q.deqBuf)
	if err := readByteArrayEntries(in, q.enqBuf[:q.enqIndex]); err != nil {
		return err
	}
	if err := readByteArrayEntries(in, q.deqBuf[q.deqIndex:]); err != nil {
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
	if q.diskdir == "" {
		return nil
	}
	return os.RemoveAll(q.diskdir)
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
			return nil
		}
	}
	q.deqIndex = len(q.deqBuf) - q.enqIndex
	clearByteArrayBuffer(q.deqBuf)
	copy(q.deqBuf[q.deqIndex:], q.enqBuf[:q.enqIndex])
	clearByteArrayBuffer(q.enqBuf[:q.enqIndex])
	q.enqIndex = 0
	return nil
}

func (q *DiskByteArrayQueue) spillEnqueueBuffer() error {
	if err := os.MkdirAll(q.diskdir, 0o755); err != nil {
		return err
	}
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

func (q *DiskByteArrayQueue) poolName(pool int) string {
	return filepath.Join(q.diskdir, intToDecimal(pool))
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
	if err := readByteArrayPoolFile(file, deqBuf); err != nil {
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
	out := NewValueOutputStream(&buf)
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
	if err := state.Read(NewValueInputStream(bytes.NewReader(raw))); err != nil {
		panic(err)
	}
	return state
}

func writeByteArrayPoolFile(name string, entries [][]byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
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
	in := NewValueInputStream(file)
	defer in.Close()
	return readByteArrayEntries(in, entries)
}

func writeByteArrayEntries(out *ValueOutputStream, entries [][]byte) error {
	for _, entry := range entries {
		if entry == nil {
			entry = []byte{}
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

func readByteArrayEntries(in *ValueInputStream, entries [][]byte) error {
	for i := range entries {
		length, err := in.ReadInt()
		if err != nil {
			return err
		}
		if length < 0 {
			return newTLCError(ECGeneral, "negative byte-array queue entry length %d", length)
		}
		buf := make([]byte, int(length))
		if _, err := io.ReadFull(in.in, buf); err != nil {
			return err
		}
		entries[i] = buf
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
