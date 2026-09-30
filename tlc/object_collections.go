package tlc

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	memObjectInitialSize = 4096
	diskObjectBufSize    = 8192
)

type objectStackFields struct {
	mu  sync.Mutex
	len int
}

type MemObjectStack struct {
	objectStackFields
	states   []any
	filename string
}

func NewMemObjectStack(metadir string, name string) *MemObjectStack {
	return &MemObjectStack{
		states:   make([]any, memObjectInitialSize),
		filename: filepath.Join(metadir, name),
	}
}

func (s *MemObjectStack) Push(state any) {
	s.enqueueInner(state)
	s.len++
}

func (s *MemObjectStack) Pop() any {
	if s.len == 0 {
		return nil
	}
	state := s.dequeueInner()
	s.len--
	return state
}

func (s *MemObjectStack) SPush(state any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enqueueInner(state)
	s.len++
}

func (s *MemObjectStack) SPushAll(states []any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, state := range states {
		s.enqueueInner(state)
		s.len++
	}
}

func (s *MemObjectStack) SPop() any {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.dequeueInner()
	s.len--
	return state
}

func (s *MemObjectStack) SPopMany(cnt int) []any {
	s.mu.Lock()
	defer s.mu.Unlock()
	states := make([]any, 0, cnt)
	for idx := 0; idx < cnt && s.len > 0; idx++ {
		states = append(states, s.dequeueInner())
		s.len--
	}
	return states
}

func (s *MemObjectStack) Size() int {
	if s == nil {
		return 0
	}
	return s.len
}

func (s *MemObjectStack) enqueueInner(state any) {
	if s.len == len(s.states) {
		newLen := max(1, s.len*2)
		newStates := make([]any, newLen)
		copy(newStates, s.states[:s.len])
		s.states = newStates
	}
	s.states[s.len] = state
}

func (s *MemObjectStack) dequeueInner() any {
	head := s.len - 1
	res := s.states[head]
	s.states[head] = nil
	return res
}

func (s *MemObjectStack) BeginChkpt() error {
	if s == nil || s.filename == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.filename), 0o755); err != nil {
		return err
	}
	file, err := os.Create(s.filename + ".tmp")
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(file)
	if err := enc.Encode(s.len); err != nil {
		_ = file.Close()
		return err
	}
	for i := 0; i < s.len; i++ {
		if err := enc.Encode(&s.states[i]); err != nil {
			_ = file.Close()
			return err
		}
	}
	return file.Close()
}

func (s *MemObjectStack) CommitChkpt() error {
	if s == nil || s.filename == "" {
		return nil
	}
	oldName := s.filename + ".chkpt"
	newName := s.filename + ".tmp"
	if err := os.Rename(newName, oldName); err != nil {
		return fmt.Errorf("MemObjectStack.CommitChkpt: cannot rename %s to %s: %w", newName, oldName, err)
	}
	return nil
}

func (s *MemObjectStack) Recover() error {
	if s == nil || s.filename == "" {
		return nil
	}
	file, err := os.Open(s.filename + ".chkpt")
	if err != nil {
		return err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	if err := dec.Decode(&s.len); err != nil {
		return err
	}
	if len(s.states) < s.len {
		s.states = make([]any, max(memObjectInitialSize, s.len))
	}
	for i := 0; i < s.len; i++ {
		if err := dec.Decode(&s.states[i]); err != nil {
			return err
		}
	}
	for i := s.len; i < len(s.states); i++ {
		s.states[i] = nil
	}
	return nil
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
	if q == nil || q.len == 0 {
		return nil
	}
	res := q.states[q.start]
	q.states[q.start] = nil
	q.start = (q.start + 1) % len(q.states)
	q.len--
	return res
}

func (q *MemObjectQueue) Size() int {
	if q == nil {
		return 0
	}
	return q.len
}

func (q *MemObjectQueue) BeginChkpt() error {
	if q == nil || q.diskdir == "" {
		return nil
	}
	if err := os.MkdirAll(q.diskdir, 0o755); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(q.diskdir, "queue.tmp"))
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(file)
	if err := enc.Encode(q.len); err != nil {
		_ = file.Close()
		return err
	}
	index := q.start
	for i := 0; i < q.len; i++ {
		if err := enc.Encode(&q.states[index]); err != nil {
			_ = file.Close()
			return err
		}
		index++
		if index == len(q.states) {
			index = 0
		}
	}
	return file.Close()
}

func (q *MemObjectQueue) CommitChkpt() error {
	if q == nil || q.diskdir == "" {
		return nil
	}
	oldName := filepath.Join(q.diskdir, "queue.chkpt")
	newName := filepath.Join(q.diskdir, "queue.tmp")
	if err := os.Remove(oldName); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("MemObjectQueue.CommitChkpt: cannot delete %s: %w", oldName, err)
	}
	if err := os.Rename(newName, oldName); err != nil {
		return fmt.Errorf("MemObjectQueue.CommitChkpt: cannot rename %s to %s: %w", newName, oldName, err)
	}
	return nil
}

func (q *MemObjectQueue) Recover() error {
	if q == nil || q.diskdir == "" {
		return nil
	}
	file, err := os.Open(filepath.Join(q.diskdir, "queue.chkpt"))
	if err != nil {
		return err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	if err := dec.Decode(&q.len); err != nil {
		return err
	}
	if len(q.states) < q.len {
		q.states = make([]any, max(memObjectInitialSize, q.len))
	}
	q.start = 0
	for i := 0; i < q.len; i++ {
		if err := dec.Decode(&q.states[i]); err != nil {
			return err
		}
	}
	for i := q.len; i < len(q.states); i++ {
		q.states[i] = nil
	}
	return nil
}

type ObjectPoolStack struct {
	filePrefix string
	buf        []any
	hiPool     int
}

func NewObjectPoolStack(bufSize int, filePrefix string) *ObjectPoolStack {
	return &ObjectPoolStack{filePrefix: filePrefix, buf: make([]any, bufSize)}
}

func (p *ObjectPoolStack) Write(outBuf []any) ([]any, error) {
	if p == nil {
		return nil, nil
	}
	res := p.buf
	p.buf = outBuf
	fileName := p.filePrefix + fmt.Sprint(p.hiPool)
	p.hiPool++
	if err := writeGobObjectSlice(fileName, outBuf); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *ObjectPoolStack) Read(inBuf []any) ([]any, error) {
	if p == nil || p.hiPool == 0 {
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
	if p == nil || enc == nil {
		return nil
	}
	return enc.Encode(p.hiPool)
}

func (p *ObjectPoolStack) Recover(dec *gob.Decoder) error {
	if p == nil || dec == nil {
		return nil
	}
	if err := dec.Decode(&p.hiPool); err != nil {
		return err
	}
	if p.hiPool > 0 {
		return readGobObjectSlice(p.filePrefix+fmt.Sprint(p.hiPool-1), p.buf)
	}
	return nil
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
	filePrefix := filepath.Join(diskdir, name)
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
	s.enqueueInner(state)
	s.len++
}

func (s *DiskObjectStack) Pop() any {
	if s.len == 0 {
		return nil
	}
	state := s.dequeueInner()
	s.len--
	return state
}

func (s *DiskObjectStack) SPush(state any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enqueueInner(state)
	s.len++
}

func (s *DiskObjectStack) SPop() any {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.dequeueInner()
	s.len--
	return state
}

func (s *DiskObjectStack) SPushAll(states []any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, state := range states {
		s.enqueueInner(state)
		s.len++
	}
}

func (s *DiskObjectStack) SPopMany(cnt int) []any {
	s.mu.Lock()
	defer s.mu.Unlock()
	states := make([]any, 0, cnt)
	for idx := 0; idx < cnt && s.len > 0; idx++ {
		states = append(states, s.dequeueInner())
		s.len--
	}
	return states
}

func (s *DiskObjectStack) Size() int {
	if s == nil {
		return 0
	}
	return s.len
}

func (s *DiskObjectStack) enqueueInner(state any) {
	if s.index == diskObjectBufSize && sameObjectSlice(s.buf, s.buf2) {
		buf, err := s.diskStack.Write(s.buf1)
		if err != nil {
			panic(err)
		}
		s.buf = buf
		s.buf1 = s.buf2
		s.buf2 = s.buf
		s.index = 0
	}
	s.buf[s.index] = state
	s.index++
}

func (s *DiskObjectStack) dequeueInner() any {
	if sameObjectSlice(s.buf, s.buf1) && s.index < diskObjectBufSize/2 {
		tempBuf, err := s.diskStack.Read(s.buf)
		if err != nil {
			panic(err)
		}
		if tempBuf != nil {
			s.buf2 = s.buf1
			s.buf1 = tempBuf
			s.buf = s.buf2
		}
	}
	s.index--
	res := s.buf[s.index]
	s.buf[s.index] = nil
	return res
}

func (s *DiskObjectStack) BeginChkpt() error {
	if s == nil || s.filePrefix == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.filePrefix), 0o755); err != nil {
		return err
	}
	file, err := os.Create(s.filePrefix + ".tmp")
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
	for _, value := range []int{s.len, index1, index2} {
		if err := enc.Encode(value); err != nil {
			_ = file.Close()
			return err
		}
	}
	for i := 0; i < index1; i++ {
		if err := enc.Encode(&s.buf1[i]); err != nil {
			_ = file.Close()
			return err
		}
	}
	for i := 0; i < index2; i++ {
		if err := enc.Encode(&s.buf2[i]); err != nil {
			_ = file.Close()
			return err
		}
	}
	return file.Close()
}

func (s *DiskObjectStack) CommitChkpt() error {
	if s == nil || s.filePrefix == "" {
		return nil
	}
	oldName := s.filePrefix + ".chkpt"
	newName := s.filePrefix + ".tmp"
	if err := os.Remove(oldName); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("DiskObjectStack.CommitChkpt: cannot delete %s: %w", oldName, err)
	}
	if err := os.Rename(newName, oldName); err != nil {
		return fmt.Errorf("DiskObjectStack.CommitChkpt: cannot rename %s to %s: %w", newName, oldName, err)
	}
	return nil
}

func (s *DiskObjectStack) Recover() error {
	if s == nil || s.filePrefix == "" {
		return nil
	}
	file, err := os.Open(s.filePrefix + ".chkpt")
	if err != nil {
		return err
	}
	defer file.Close()
	dec := gob.NewDecoder(file)
	var index1, index2 int
	for _, ptr := range []*int{&s.len, &index1, &index2} {
		if err := dec.Decode(ptr); err != nil {
			return err
		}
	}
	for i := 0; i < index1; i++ {
		if err := dec.Decode(&s.buf1[i]); err != nil {
			return err
		}
	}
	for i := 0; i < index2; i++ {
		if err := dec.Decode(&s.buf2[i]); err != nil {
			return err
		}
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

func writeGobObjectSlice(fileName string, values []any) error {
	if err := os.MkdirAll(filepath.Dir(fileName), 0o755); err != nil {
		return err
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(file)
	if err := enc.Encode(len(values)); err != nil {
		_ = file.Close()
		return err
	}
	for i := range values {
		if err := enc.Encode(&values[i]); err != nil {
			_ = file.Close()
			return err
		}
	}
	return file.Close()
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
