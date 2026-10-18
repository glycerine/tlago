package tlc

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

const (
	memObjectInitialSize = 4096
	diskObjectBufSize    = 8192
)

type objectStackFields struct {
	mu  sync.Mutex
	len int64
}

// ObjectStack.len is a plain Java int. Atomic snapshots avoid a Go data race
// without turning its separate read/modify/write updates into atomic additions.
func objectStackLengthLoad(length *int64) int {
	return int(int32(atomic.LoadInt64(length)))
}

func objectStackLengthStore(length *int64, value int) {
	atomic.StoreInt64(length, int64(int32(value)))
}

func objectStackLengthAdd(length *int64, delta int) {
	objectStackLengthStore(length, objectStackLengthLoad(length)+delta)
}

type MemObjectStack struct {
	objectStackFields
	states   []any
	filename string
}

func NewMemObjectStack(metadir string, name string) *MemObjectStack {
	return &MemObjectStack{
		states:   make([]any, memObjectInitialSize),
		filename: metadir + string(os.PathSeparator) + name,
	}
}

func (s *MemObjectStack) Push(state any) {
	s.require()
	s.enqueueInner(state)
	objectStackLengthAdd(&s.len, 1)
}

func (s *MemObjectStack) Pop() any {
	s.require()
	if objectStackLengthLoad(&s.len) == 0 {
		return nil
	}
	state := s.dequeueInner()
	objectStackLengthAdd(&s.len, -1)
	return state
}

func (s *MemObjectStack) SPush(state any) {
	s.require()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enqueueInner(state)
	objectStackLengthAdd(&s.len, 1)
}

func (s *MemObjectStack) SPushAll(states []any) {
	s.require()
	if states == nil {
		panic(NewNullPointerException())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, state := range states {
		s.enqueueInner(state)
	}
	objectStackLengthAdd(&s.len, len(states))
}

func (s *MemObjectStack) SPop() any {
	s.require()
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.dequeueInner()
	objectStackLengthAdd(&s.len, -1)
	return state
}

func (s *MemObjectStack) SPopMany(cnt int) []any {
	s.require()
	cnt = int(int32(cnt))
	if cnt < 0 {
		panic(NewNegativeArraySizeException(strconv.Itoa(cnt)))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	states := make([]any, cnt)
	idx := 0
	for ; idx < cnt && objectStackLengthLoad(&s.len) > 0; idx++ {
		states[idx] = s.dequeueInner()
		objectStackLengthAdd(&s.len, -1)
	}
	if idx == cnt {
		return states
	}
	result := make([]any, idx)
	copy(result, states)
	return result
}

func (s *MemObjectStack) Size() int {
	s.require()
	return objectStackLengthLoad(&s.len)
}

func (s *MemObjectStack) enqueueInner(state any) {
	length := objectStackLengthLoad(&s.len)
	if length == len(s.states) {
		newLen := max(1, length*2)
		newStates := make([]any, newLen)
		copy(newStates, s.states[:length])
		s.states = newStates
	}
	s.states[length] = state
}

func (s *MemObjectStack) dequeueInner() any {
	head := objectStackLengthLoad(&s.len) - 1
	if head < 0 || head >= len(s.states) {
		panic(NewArrayIndexOutOfBoundsException(head, len(s.states)))
	}
	res := s.states[head]
	s.states[head] = nil
	return res
}

func (s *MemObjectStack) BeginChkpt() error {
	s.require()
	file, err := openPoolFile(s.filename+".tmp", true)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(file)
	length := objectStackLengthLoad(&s.len)
	if err := enc.Encode(length); err != nil {
		return err
	}
	for i := 0; i < length; i += 2 {
		if err := enc.Encode(&s.states[i]); err != nil {
			return err
		}
	}
	return file.Close()
}

func (s *MemObjectStack) CommitChkpt() error {
	s.require()
	oldName := s.filename + ".chkpt"
	newName := s.filename + ".tmp"
	if err := os.Rename(newName, oldName); err != nil {
		return NewIOException(fmt.Sprintf("MemObjectStack.commitChkpt: cannot delete %s", oldName))
	}
	return nil
}

func (s *MemObjectStack) Recover() error {
	s.require()
	file, err := openPoolFile(s.filename+".chkpt", false)
	if err != nil {
		return err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	var length int
	if err := dec.Decode(&length); err != nil {
		return err
	}
	objectStackLengthStore(&s.len, length)
	for i := 0; i < objectStackLengthLoad(&s.len); i++ {
		var state any
		if err := dec.Decode(&state); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return NewEOFException()
			}
			return err
		}
		if i < 0 || i >= len(s.states) {
			return NewArrayIndexOutOfBoundsException(i, len(s.states))
		}
		s.states[i] = state
	}
	return nil
}

func (s *MemObjectStack) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

type MemObjectQueue struct {
	len     int
	states  []any
	start   int
	diskdir string
}

func NewMemObjectQueue(metadir string) *MemObjectQueue {
	return &MemObjectQueue{states: make([]any, memObjectInitialSize), diskdir: metadir}
}

func (q *MemObjectQueue) Enqueue(state any) {
	q.require()
	if q.len == len(q.states) {
		newLen := max(1, q.len*2)
		newStates := make([]any, newLen)
		copyLen := len(q.states) - q.start
		copy(newStates, q.states[q.start:])
		copy(newStates[copyLen:], q.states[:q.start])
		q.states = newStates
		q.start = 0
	}
	last := (q.start + q.len) % len(q.states)
	q.states[last] = state
	q.len++
}

func (q *MemObjectQueue) Dequeue() any {
	q.require()
	if q.len == 0 {
		return nil
	}
	res := q.states[q.start]
	q.states[q.start] = nil
	q.start = (q.start + 1) % len(q.states)
	q.len--
	return res
}

func (q *MemObjectQueue) Size() int {
	q.require()
	return q.len
}

func (q *MemObjectQueue) BeginChkpt() error {
	q.require()
	file, err := openPoolFile(q.checkpointPath("tmp"), true)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(file)
	if err := enc.Encode(q.len); err != nil {
		return err
	}
	index := q.start
	for i := 0; i < q.len; i++ {
		if err := enc.Encode(&q.states[index]); err != nil {
			return err
		}
		index++
		if index == len(q.states) {
			index = 0
		}
	}
	return bufferedRandomAccessFileIOError(file.Close())
}

func (q *MemObjectQueue) CommitChkpt() error {
	q.require()
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

func (q *MemObjectQueue) Recover() error {
	q.require()
	file, err := openPoolFile(q.checkpointPath("chkpt"), false)
	if err != nil {
		return err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	if err := dec.Decode(&q.len); err != nil {
		return err
	}
	for i := 0; i < q.len; i++ {
		// Read the object before checking the array assignment, as Java does.
		var state any
		if err := dec.Decode(&state); err != nil {
			return err
		}
		if i >= len(q.states) {
			return NewArrayIndexOutOfBoundsException(i, len(q.states))
		}
		q.states[i] = state
	}
	return bufferedRandomAccessFileIOError(file.Close())
}

func (q *MemObjectQueue) checkpointPath(ext string) string {
	return filenameNormalizeFile(q.diskdir + string(os.PathSeparator) + "queue." + ext)
}

func (q *MemObjectQueue) require() {
	if q == nil {
		panic(NewNullPointerException())
	}
}

type ObjectPoolStack struct {
	filePrefix string
	buf        []any
	hiPool     int
	idle       bool
	poolFile   string
}

func NewObjectPoolStack(bufSize int, filePrefix string) *ObjectPoolStack {
	bufSize = int(int32(bufSize))
	if bufSize < 0 {
		panic(NewNegativeArraySizeException(strconv.Itoa(bufSize)))
	}
	return &ObjectPoolStack{filePrefix: filePrefix, buf: make([]any, bufSize), idle: true}
}

func (p *ObjectPoolStack) Write(outBuf []any) ([]any, error) {
	p.require()
	res := p.buf
	p.buf = outBuf
	p.poolFile = p.filePrefix + fmt.Sprint(p.hiPool)
	p.hiPool++
	p.idle = false
	return res, NewIllegalMonitorStateException("current thread is not owner")
}

func (p *ObjectPoolStack) Read(inBuf []any) ([]any, error) {
	p.require()
	if p.hiPool == 0 {
		return nil, nil
	}
	res := p.buf
	p.buf = inBuf
	p.hiPool--
	if p.hiPool > 0 {
		if err := readGobObjectSlice(p.filePrefix+fmt.Sprint(p.hiPool-1), p.buf); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (p *ObjectPoolStack) BeginChkpt(enc *gob.Encoder) error {
	p.require()
	if enc == nil {
		panic(NewNullPointerException())
	}
	return enc.Encode(p.hiPool)
}

func (p *ObjectPoolStack) Recover(dec *gob.Decoder) error {
	p.require()
	if dec == nil {
		panic(NewNullPointerException())
	}
	if err := dec.Decode(&p.hiPool); err != nil {
		return err
	}
	if p.hiPool > 0 {
		p.poolFile = p.filePrefix + fmt.Sprint(p.hiPool-1)
		p.idle = false
		return NewIllegalMonitorStateException("current thread is not owner")
	}
	return nil
}

func (p *ObjectPoolStack) require() {
	if p == nil {
		panic(NewNullPointerException())
	}
}

type DiskObjectStack struct {
	objectStackFields
	filePrefix string
	buf1       []any
	buf2       []any
	buf        []any
	index      int
	diskStack  *ObjectPoolStack
}

func NewDiskObjectStack(diskdir string, name string) *DiskObjectStack {
	filePrefix := diskdir + string(os.PathSeparator) + name
	out := &DiskObjectStack{
		filePrefix: filePrefix,
		buf1:       make([]any, diskObjectBufSize),
		buf2:       make([]any, diskObjectBufSize),
		index:      0,
		diskStack:  NewObjectPoolStack(diskObjectBufSize, filePrefix),
	}
	out.buf = out.buf1
	return out
}

func (s *DiskObjectStack) Push(state any) {
	s.require()
	s.enqueueInner(state)
	objectStackLengthAdd(&s.len, 1)
}

func (s *DiskObjectStack) Pop() any {
	s.require()
	if objectStackLengthLoad(&s.len) == 0 {
		return nil
	}
	state := s.dequeueInner()
	objectStackLengthAdd(&s.len, -1)
	return state
}

func (s *DiskObjectStack) SPush(state any) {
	s.require()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enqueueInner(state)
	objectStackLengthAdd(&s.len, 1)
}

func (s *DiskObjectStack) SPop() any {
	s.require()
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.dequeueInner()
	objectStackLengthAdd(&s.len, -1)
	return state
}

func (s *DiskObjectStack) SPushAll(states []any) {
	s.require()
	if states == nil {
		panic(NewNullPointerException())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, state := range states {
		s.enqueueInner(state)
	}
	objectStackLengthAdd(&s.len, len(states))
}

func (s *DiskObjectStack) SPopMany(cnt int) []any {
	s.require()
	cnt = int(int32(cnt))
	if cnt < 0 {
		panic(NewNegativeArraySizeException(strconv.Itoa(cnt)))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	states := make([]any, cnt)
	idx := 0
	for ; idx < cnt && objectStackLengthLoad(&s.len) > 0; idx++ {
		states[idx] = s.dequeueInner()
		objectStackLengthAdd(&s.len, -1)
	}
	if idx == cnt {
		return states
	}
	result := make([]any, idx)
	copy(result, states)
	return result
}

func (s *DiskObjectStack) Size() int {
	s.require()
	return objectStackLengthLoad(&s.len)
}

func (s *DiskObjectStack) enqueueInner(state any) {
	if s.index == diskObjectBufSize && sameObjectSlice(s.buf, s.buf2) {
		buf, err := s.diskStack.Write(s.buf1)
		if err != nil {
			panic(NewTLCRuntimeException(ECSystemErrorWritingStates, "stack", err.Error()))
		}
		s.buf = buf
		s.buf1 = s.buf2
		s.buf2 = s.buf
		s.index = 0
	}
	index := s.index
	s.index = int(int32(s.index) + 1)
	if index < 0 || index >= len(s.buf) {
		panic(NewArrayIndexOutOfBoundsException(index, len(s.buf)))
	}
	s.buf[index] = state
}

func (s *DiskObjectStack) dequeueInner() any {
	if sameObjectSlice(s.buf, s.buf1) && s.index < diskObjectBufSize/2 {
		tempBuf, err := s.diskStack.Read(s.buf)
		if err != nil {
			panic(NewTLCRuntimeException(ECSystemErrorReadingStates, "stack", err.Error()))
		}
		if tempBuf != nil {
			s.buf2 = s.buf1
			s.buf1 = tempBuf
			s.buf = s.buf2
		}
	}
	s.index--
	if s.index < 0 || s.index >= len(s.buf) {
		panic(NewArrayIndexOutOfBoundsException(s.index, len(s.buf)))
	}
	res := s.buf[s.index]
	s.buf[s.index] = nil
	return res
}

func (s *DiskObjectStack) BeginChkpt() error {
	s.require()
	file, err := openPoolFile(s.filePrefix+".tmp", true)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(file)
	index1 := diskObjectBufSize
	index2 := s.index
	if sameObjectSlice(s.buf, s.buf1) {
		index1 = s.index
		index2 = 0
	}
	for _, value := range []int{objectStackLengthLoad(&s.len), index1, index2} {
		if err := enc.Encode(value); err != nil {
			return err
		}
	}
	for i := 0; i < index1; i++ {
		if i < 0 || i >= len(s.buf1) {
			return NewArrayIndexOutOfBoundsException(i, len(s.buf1))
		}
		if err := enc.Encode(&s.buf1[i]); err != nil {
			return err
		}
	}
	for i := 0; i < index2; i++ {
		if i < 0 || i >= len(s.buf2) {
			return NewArrayIndexOutOfBoundsException(i, len(s.buf2))
		}
		if err := enc.Encode(&s.buf2[i]); err != nil {
			return err
		}
	}
	return file.Close()
}

func (s *DiskObjectStack) CommitChkpt() error {
	s.require()
	oldName := s.filePrefix + ".chkpt"
	newName := s.filePrefix + ".tmp"
	if _, err := os.Stat(oldName); err == nil {
		if err := os.Remove(oldName); err != nil {
			return NewIOException(fmt.Sprintf("DiskObjectStack.commitChkpt: cannot delete %s", oldName))
		}
	}
	if err := os.Rename(newName, oldName); err != nil {
		return NewIOException(fmt.Sprintf("DiskObjectStack.commitChkpt: cannot delete %s", oldName))
	}
	return nil
}

func (s *DiskObjectStack) Recover() error {
	s.require()
	file, err := openPoolFile(s.filePrefix+".chkpt", false)
	if err != nil {
		return err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	var length, index1, index2 int
	if err := decodeObjectStackGob(dec, &length); err != nil {
		return err
	}
	objectStackLengthStore(&s.len, length)
	for _, ptr := range []*int{&index1, &index2} {
		if err := decodeObjectStackGob(dec, ptr); err != nil {
			return err
		}
		*ptr = int(int32(*ptr))
	}
	for i := 0; i < index1; i++ {
		var state any
		if err := decodeObjectStackGob(dec, &state); err != nil {
			return err
		}
		if i < 0 || i >= len(s.buf1) {
			return NewArrayIndexOutOfBoundsException(i, len(s.buf1))
		}
		s.buf1[i] = state
	}
	for i := 0; i < index2; i++ {
		var state any
		if err := decodeObjectStackGob(dec, &state); err != nil {
			return err
		}
		if i < 0 || i >= len(s.buf2) {
			return NewArrayIndexOutOfBoundsException(i, len(s.buf2))
		}
		s.buf2[i] = state
	}
	if index2 == 0 {
		s.buf = s.buf1
		s.index = index1
	} else {
		s.buf = s.buf2
		s.index = index2
	}
	return nil
}

func (s *DiskObjectStack) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

func decodeObjectStackGob(dec *gob.Decoder, destination any) error {
	if err := dec.Decode(destination); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return NewEOFException()
		}
		return err
	}
	return nil
}

func readGobObjectSlice(fileName string, values []any) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	var count int
	if err := dec.Decode(&count); err != nil {
		return err
	}
	if count > len(values) {
		return fmt.Errorf("object pool file %s has %d entries for %d-slot buffer", fileName, count, len(values))
	}
	for i := 0; i < count; i++ {
		if err := dec.Decode(&values[i]); err != nil {
			return err
		}
	}
	for i := count; i < len(values); i++ {
		values[i] = nil
	}
	return nil
}

func sameObjectSlice(a []any, b []any) bool {
	if len(a) == 0 || len(b) == 0 {
		return len(a) == 0 && len(b) == 0
	}
	return &a[0] == &b[0]
}
