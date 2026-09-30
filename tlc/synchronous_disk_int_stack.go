package tlc

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
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
	capacityValue := SynchronousDiskIntStackBufSize
	if len(capacity) > 0 {
		capacityValue = capacity[0]
	}
	if capacityValue > synchronousDiskIntStackBufSizeMax {
		capacityValue = synchronousDiskIntStackBufSizeMax
	}
	if capacityValue < 1 {
		capacityValue = 1
	}
	return &SynchronousDiskIntStack{
		bufSize:    capacityValue,
		filePrefix: filepath.Join(diskdir, name),
		buf:        make([]int32, capacityValue),
	}
}

func (s *SynchronousDiskIntStack) Size() int64 {
	if s == nil {
		return 0
	}
	return s.size
}

func (s *SynchronousDiskIntStack) PushInt(x int32) {
	if s.index == s.bufSize {
		if err := s.flushPool(); err != nil {
			panic(err)
		}
	}
	s.buf[s.index] = x
	s.index++
	s.size++
}

func (s *SynchronousDiskIntStack) PushLong(x int64) {
	s.PushInt(int32(uint64(x) & 0xffffffff))
	s.PushInt(int32(uint64(x) >> 32))
}

func (s *SynchronousDiskIntStack) PopInt() int32 {
	if s.index == 0 && s.hasPool() {
		if err := s.fillPool(); err != nil {
			panic(err)
		}
	}
	s.size--
	s.index--
	return s.buf[s.index]
}

func (s *SynchronousDiskIntStack) PopLong() int64 {
	high := int64(s.PopInt())
	low := int64(s.PopInt())
	return (high << 32) | (low & 0xffffffff)
}

func (s *SynchronousDiskIntStack) Reset() {
	s.size = 0
	s.index = 0
	s.hiPool = 0
}

func (s *SynchronousDiskIntStack) hasPool() bool {
	return s.hiPool >= 0
}

func (s *SynchronousDiskIntStack) flushPool() error {
	fileName := s.poolName(s.hiPool)
	if err := os.MkdirAll(filepath.Dir(fileName), 0o755); err != nil {
		return err
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	out := bufio.NewWriter(file)
	for _, value := range s.buf {
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(value))
		if _, err := out.Write(b[:]); err != nil {
			_ = file.Close()
			return err
		}
	}
	if err := out.Flush(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	s.hiPool++
	s.index = 0
	return nil
}

func (s *SynchronousDiskIntStack) fillPool() error {
	fileName := s.poolName(s.hiPool - 1)
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	in := bufio.NewReader(file)
	for i := range s.buf {
		var b [4]byte
		if _, err := io.ReadFull(in, b[:]); err != nil {
			_ = file.Close()
			return err
		}
		s.buf[i] = int32(binary.BigEndian.Uint32(b[:]))
	}
	if err := file.Close(); err != nil {
		return err
	}
	s.hiPool--
	s.index = len(s.buf)
	return nil
}

func (s *SynchronousDiskIntStack) poolName(pool int) string {
	return s.filePrefix + strconv.Itoa(pool)
}
