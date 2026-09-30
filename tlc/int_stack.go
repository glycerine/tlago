package tlc

type IntStack struct {
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
	return s.size
}

func (s *IntStack) PushInt(x int32) {
	if s.size == len(s.elems) {
		newElems := make([]int32, s.ensureCapacity(intStackMinCapacity))
		copy(newElems, s.elems[:s.size])
		s.elems = newElems
	}
	s.elems[s.size] = x
	s.size++
}

func (s *IntStack) PushLong(x int64) {
	s.PushInt(int32(uint64(x) & 0xffffffff))
	s.PushInt(int32(uint64(x) >> 32))
}

func (s *IntStack) PopInt() int32 {
	if s.size == 0 {
		panic("IntStack is empty")
	}
	s.size--
	return s.elems[s.size]
}

func (s *IntStack) PopLong() int64 {
	high := int64(s.PopInt())
	low := int64(s.PopInt())
	return (high << 32) | (low & 0xffffffff)
}

func (s *IntStack) PeekInt() int32 {
	return s.PeekIntAt(s.size - 1)
}

func (s *IntStack) PeekIntAt(pos int) int32 {
	if pos < 0 || pos >= s.size {
		panic("IntStack index out of range")
	}
	return s.elems[pos]
}

func (s *IntStack) PeekLong() int64 {
	return s.PeekLongAt(s.size - 2)
}

func (s *IntStack) PeekLongAt(pos int) int64 {
	high := int64(s.PeekIntAt(pos + 1))
	low := int64(s.PeekIntAt(pos))
	return (high << 32) | (low & 0xffffffff)
}

func (s *IntStack) Reset() {
	s.size = 0
}

func (s *IntStack) ensureCapacity(minCapacity int) int {
	newSize := int((int64(s.size)*3)/2) + 1
	if min := s.size + minCapacity; newSize < min {
		newSize = min
	}
	return newSize
}
