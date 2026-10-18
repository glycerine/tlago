package tlc

import "sync"

type IntStack struct {
	mu    sync.Mutex
	elems []int32
	size  int
}

const intStackMinCapacity = 1024

func NewIntStack() *IntStack {
	return NewIntStackWithCapacity(intStackMinCapacity)
}

func NewIntStackWithCapacity(capacity int) *IntStack {
	if capacity < 0 {
		capacity = 0
	}
	return &IntStack{elems: make([]int32, capacity)}
}

func (s *IntStack) Size() int {
	s.require()
	return s.size
}

func (s *IntStack) PushInt(x int32) {
	s.require()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pushIntLocked(x)
}

func (s *IntStack) pushIntLocked(x int32) {
	if s.size == len(s.elems) {
		newElems := make([]int32, s.ensureCapacity(intStackMinCapacity))
		copy(newElems, s.elems[:s.size])
		s.elems = newElems
	}
	if s.size < 0 || s.size >= len(s.elems) {
		panic(NewArrayIndexOutOfBoundsException(s.size, len(s.elems)))
	}
	s.elems[s.size] = x
	s.size++
}

func (s *IntStack) PushLong(x int64) {
	s.require()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pushIntLocked(int32(uint64(x) & 0xffffffff))
	s.pushIntLocked(int32(uint64(x) >> 32))
}

func (s *IntStack) PopInt() int32 {
	s.require()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.popIntLocked()
}

func (s *IntStack) popIntLocked() int32 {
	s.size--
	if s.size < 0 || s.size >= len(s.elems) {
		panic(NewArrayIndexOutOfBoundsException(s.size, len(s.elems)))
	}
	return s.elems[s.size]
}

func (s *IntStack) PopLong() int64 {
	s.require()
	s.mu.Lock()
	defer s.mu.Unlock()
	high := int64(s.popIntLocked())
	low := int64(s.popIntLocked())
	return (high << 32) | (low & 0xffffffff)
}

func (s *IntStack) PeakInt(positions ...int) int32 {
	s.require()
	if len(positions) > 1 {
		panic(NewIllegalArgumentException("PeakInt accepts at most one position"))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	pos := s.size - 1
	if len(positions) == 1 {
		pos = positions[0]
	}
	return s.peakIntLocked(pos)
}

func (s *IntStack) peakIntLocked(pos int) int32 {
	pos = int(int32(pos))
	if pos < 0 || pos >= len(s.elems) {
		panic(NewArrayIndexOutOfBoundsException(pos, len(s.elems)))
	}
	return s.elems[pos]
}

func (s *IntStack) PeekInt() int32          { return s.PeakInt() }
func (s *IntStack) PeekIntAt(pos int) int32 { return s.PeakInt(pos) }

func (s *IntStack) PeakLong(positions ...int) int64 {
	s.require()
	if len(positions) > 1 {
		panic(NewIllegalArgumentException("PeakLong accepts at most one position"))
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(positions) == 0 {
		high := int64(s.peakIntLocked(s.size - 1))
		low := int64(s.peakIntLocked(s.size - 1))
		return (high << 32) | (low & 0xffffffff)
	}
	pos := int(int32(positions[0]))
	high := int64(s.peakIntLocked(int(int32(pos) + 1)))
	low := int64(s.peakIntLocked(pos))
	return (high << 32) | (low & 0xffffffff)
}

func (s *IntStack) PeekLong() int64          { return s.PeakLong() }
func (s *IntStack) PeekLongAt(pos int) int64 { return s.PeakLong(pos) }

func (s *IntStack) Reset() {
	s.require()
	s.size = 0
}

func (s *IntStack) ensureCapacity(minCapacity int) int {
	newSize := int((int64(s.size)*3)/2) + 1
	if min := s.size + minCapacity; newSize < min {
		newSize = min
	}
	return newSize
}

func (s *IntStack) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}
