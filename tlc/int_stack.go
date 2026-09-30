package tlc

type IntStack struct {
	m []int32
}

func NewIntStack() *IntStack {
	return NewIntStackWithCapacity(0)
}

func NewIntStackWithCapacity(capacity int) *IntStack {
	if capacity < 0 {
		capacity = 0
	}
	return &IntStack{m: make([]int32, 0, capacity)}
}

func (s *IntStack) Size() int {
	return len(s.m)
}

func (s *IntStack) PushInt(x int32) {
	s.m = append(s.m, x)
}

func (s *IntStack) PushLong(x int64) {
	s.PushInt(int32(uint64(x) & 0xffffffff))
	s.PushInt(int32(uint64(x) >> 32))
}

func (s *IntStack) PopInt() int32 {
	if len(s.m) == 0 {
		panic("IntStack is empty")
	}
	last := len(s.m) - 1
	x := s.m[last]
	s.m = s.m[:last]
	return x
}

func (s *IntStack) PopLong() int64 {
	high := int64(s.PopInt())
	low := int64(s.PopInt())
	return (high << 32) | (low & 0xffffffff)
}

func (s *IntStack) PeekInt() int32 {
	return s.PeekIntAt(len(s.m) - 1)
}

func (s *IntStack) PeekIntAt(pos int) int32 {
	if pos < 0 || pos >= len(s.m) {
		panic("IntStack index out of range")
	}
	return s.m[pos]
}

func (s *IntStack) PeekLong() int64 {
	return s.PeekLongAt(len(s.m) - 2)
}

func (s *IntStack) PeekLongAt(pos int) int64 {
	high := int64(s.PeekIntAt(pos + 1))
	low := int64(s.PeekIntAt(pos))
	return (high << 32) | (low & 0xffffffff)
}

func (s *IntStack) Reset() {
	s.m = s.m[:0]
}
