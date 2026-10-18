package tlc

import (
	"os"
	"strconv"
)

const (
	SynchronousDiskIntStackBufSize    = 8388608
	synchronousDiskIntStackBufSizeMax = SynchronousDiskIntStackBufSize << 5
)

type SynchronousDiskIntStack struct {
	bufSize    int
	filePrefix string
	size       int64
	index      int
	hiPool     int
	buf        []int32
}

func NewSynchronousDiskIntStack(diskdir string, name string, capacity ...int) *SynchronousDiskIntStack {
	if len(capacity) > 1 {
		panic(NewIllegalArgumentException("NewSynchronousDiskIntStack accepts at most one capacity"))
	}
	capacityValue := SynchronousDiskIntStackBufSize
	if len(capacity) > 0 {
		capacityValue = int(int32(capacity[0]))
	}
	if capacityValue > synchronousDiskIntStackBufSizeMax {
		capacityValue = synchronousDiskIntStackBufSizeMax
	}
	if capacityValue < 0 {
		panic(NewNegativeArraySizeException(strconv.Itoa(capacityValue)))
	}
	return &SynchronousDiskIntStack{
		bufSize:    capacityValue,
		filePrefix: diskdir + string(os.PathSeparator) + name,
		buf:        make([]int32, capacityValue),
	}
}

func (s *SynchronousDiskIntStack) Size() int64 {
	s.require()
	return s.size
}

func (s *SynchronousDiskIntStack) PushInt(x int32) {
	s.require()
	if s.index == s.bufSize {
		if err := s.flushPool(); err != nil {
			panic(NewTLCRuntimeException(ECSystemErrorWritingStates, "stack", err.Error()))
		}
	}
	if s.index < 0 || s.index >= len(s.buf) {
		panic(NewArrayIndexOutOfBoundsException(s.index, len(s.buf)))
	}
	s.buf[s.index] = x
	s.index++
	s.size++
}

func (s *SynchronousDiskIntStack) PushLong(x int64) {
	s.require()
	s.PushInt(int32(uint64(x) & 0xffffffff))
	s.PushInt(int32(uint64(x) >> 32))
}

func (s *SynchronousDiskIntStack) PopInt() int32 {
	s.require()
	if s.index == 0 && s.hasPool() {
		if err := s.fillPool(); err != nil {
			panic(NewTLCRuntimeException(ECSystemErrorWritingStates, "stack", err.Error()))
		}
	}
	s.size--
	s.index--
	if s.index < 0 || s.index >= len(s.buf) {
		panic(NewArrayIndexOutOfBoundsException(s.index, len(s.buf)))
	}
	return s.buf[s.index]
}

func (s *SynchronousDiskIntStack) PopLong() int64 {
	s.require()
	high := int64(s.PopInt())
	low := int64(s.PopInt())
	return (high << 32) | (low & 0xffffffff)
}

func (s *SynchronousDiskIntStack) Reset() {
	s.require()
	s.size = 0
	s.index = 0
	s.hiPool = 0
}

func (s *SynchronousDiskIntStack) hasPool() bool {
	return s.hiPool >= 0
}

func (s *SynchronousDiskIntStack) flushPool() error {
	fileName := s.poolName(s.hiPool)
	out, err := NewBufferedDataOutputStreamFile(fileName)
	if err != nil {
		return err
	}
	for _, value := range s.buf {
		if err := out.WriteInt(value); err != nil {
			return err
		}
	}
	if err := out.Close(); err != nil {
		return err
	}
	s.hiPool++
	s.index = 0
	return nil
}

func (s *SynchronousDiskIntStack) fillPool() error {
	fileName := s.poolName(s.hiPool - 1)
	in, err := NewBufferedDataInputStreamFile(fileName)
	if err != nil {
		return err
	}
	for i := range s.buf {
		value, err := in.ReadInt()
		if err != nil {
			return err
		}
		s.buf[i] = value
	}
	if err := in.Close(); err != nil {
		return err
	}
	s.hiPool--
	s.index = len(s.buf)
	return nil
}

func (s *SynchronousDiskIntStack) poolName(pool int) string {
	return s.filePrefix + strconv.Itoa(pool)
}

func (s *SynchronousDiskIntStack) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}
