// Copyright (c) 2003 Compaq Corporation. All rights reserved.
package tlc

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
)

// BufferedDataInputStream preserves the source's eager refill and EOF state.
// Like the Java class, callers provide synchronization when sharing a stream.
type BufferedDataInputStream struct {
	in           io.Reader
	buff         [8192]byte
	length, curr int
	temp         [8]byte
	pending      error
}

func NewBufferedDataInputStream(input ...io.Reader) (*BufferedDataInputStream, error) {
	s := &BufferedDataInputStream{length: -1}
	if len(input) > 0 {
		s.in = input[0]
		if err := s.refill(); err != nil {
			return nil, err
		}
	}
	return s, nil
}
func NewBufferedDataInputStreamFile(name string) (*BufferedDataInputStream, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	}
	s, err := NewBufferedDataInputStream(f)
	if err != nil {
		_ = f.Close()
	}
	return s, err
}
func (s *BufferedDataInputStream) refill() error {
	if s.in == nil {
		panic(NewNullPointerException())
	}
	if s.pending != nil {
		err := s.pending
		s.pending = nil
		if errors.Is(err, io.EOF) {
			s.length = -1
			return nil
		}
		return bufferedRandomAccessFileIOError(err)
	}
	n, err := s.in.Read(s.buff[:])
	// Go permits bytes and EOF in the same read; Java reports EOF next time.
	if n > 0 {
		s.length = n
		s.pending = err
		return nil
	}
	if errors.Is(err, io.EOF) {
		s.length = -1
		return nil
	}
	if err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	s.length = 0
	panic(NewTLCRuntimeException(ECSystemStreamEmpty))
}
func (s *BufferedDataInputStream) Open(input io.Reader) error {
	if s.in != nil {
		panic(NewTLCRuntimeException(ECSystemStreamEmpty))
	}
	s.in = input
	s.pending = nil
	return s.refill()
}
func (s *BufferedDataInputStream) OpenFile(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	return s.Open(f)
}
func (s *BufferedDataInputStream) Close() error {
	if s.in == nil {
		panic(NewNullPointerException())
	}
	if closer, ok := s.in.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return bufferedRandomAccessFileIOError(err)
		}
	}
	s.in = nil
	return nil
}
func (s *BufferedDataInputStream) AtEOF() bool { return s.length < 0 }
func (s *BufferedDataInputStream) Read(p []byte) (int, error) {
	n, err := s.ReadBytes(p, 0, len(p))
	if n < 0 {
		return 0, io.EOF
	}
	return n, err
}

// ReadBytes returns Java's -1 at EOF, retaining the original offset/count API.
func (s *BufferedDataInputStream) ReadBytes(p []byte, off, n int) (int, error) {
	if s.length < 0 {
		return -1, nil
	}
	initial := off
	for n > 0 && s.length > 0 {
		count := min(n, s.length-s.curr)
		copy(p[off:off+count], s.buff[s.curr:s.curr+count])
		s.curr += count
		off += count
		n -= count
		if s.curr == s.length {
			if err := s.refill(); err != nil {
				return off - initial, err
			}
			s.curr = 0
		}
	}
	return off - initial, nil
}
func (s *BufferedDataInputStream) ReadFully(p []byte, rangeArgs ...int) error {
	off, n := 0, len(p)
	if len(rangeArgs) > 0 {
		off, n = rangeArgs[0], rangeArgs[1]
	}
	for n > 0 {
		read, err := s.ReadBytes(p, off, n)
		if err != nil {
			return err
		}
		if read < 0 {
			return NewEOFException()
		}
		off += read
		n -= read
	}
	return nil
}
func (s *BufferedDataInputStream) ReadByte() (int8, error) {
	if s.length < 0 {
		return 0, NewEOFException()
	}
	value := int8(s.buff[s.curr])
	s.curr++
	if s.curr == s.length {
		if err := s.refill(); err != nil {
			return 0, err
		}
		s.curr = 0
	}
	return value, nil
}
func (s *BufferedDataInputStream) ReadBoolean() (bool, error) {
	v, err := s.ReadByte()
	return v != 0, err
}
func (s *BufferedDataInputStream) ReadShort() (int16, error) {
	if err := s.ReadFully(s.temp[:], 0, 2); err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(s.temp[:2])), nil
}
func (s *BufferedDataInputStream) ReadInt() (int32, error) {
	if err := s.ReadFully(s.temp[:], 0, 4); err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(s.temp[:4])), nil
}
func (s *BufferedDataInputStream) ReadLong() (int64, error) {
	if err := s.ReadFully(s.temp[:], 0, 8); err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(s.temp[:])), nil
}
func (s *BufferedDataInputStream) ReadString(n int) (string, error) {
	if n < 0 {
		panic(NewNegativeArraySizeException())
	}
	units := make([]uint16, n)
	off := 0
	for n > 0 {
		if s.length < 0 {
			return "", NewEOFException()
		}
		for n > 0 && s.length > 0 {
			count := min(n, s.length-s.curr)
			for i := 0; i < count; i++ {
				units[off+i] = uint16(int16(int8(s.buff[s.curr+i])))
			}
			s.curr += count
			off += count
			n -= count
			if s.curr == s.length {
				if err := s.refill(); err != nil {
					return "", err
				}
				s.curr = 0
			}
		}
	}
	return javaStringFromUTF16(units), nil
}
func (s *BufferedDataInputStream) Skip(n int) error {
	for s.length > 0 && s.curr+n >= s.length {
		n -= s.length - s.curr
		if err := s.refill(); err != nil {
			return err
		}
		s.curr = 0
	}
	if n > 0 && s.length < 0 {
		return NewEOFException()
	}
	s.curr += n
	if !(s.length < 0 || s.curr < s.length) {
		panic(NewTLCRuntimeException(ECSystemIndexError))
	}
	return nil
}
func (s *BufferedDataInputStream) ReadLine() (*string, error) {
	var result *string
	appendPart := func(data []byte) error {
		part, err := javaCharsetDecode(data, javaDefaultCharset(), true)
		if err != nil {
			return err
		}
		if result == nil {
			result = &part
		} else {
			*result += part
		}
		return nil
	}
	for s.length > 0 {
		for i := s.curr; i < s.length; i++ {
			if s.buff[i] == '\n' || s.buff[i] == '\r' {
				eol := s.buff[i]
				if err := appendPart(s.buff[s.curr:i]); err != nil {
					return nil, err
				}
				if err := s.Skip(i + 1 - s.curr); err != nil {
					return nil, err
				}
				if eol == '\r' && s.length > 0 && s.buff[s.curr] == '\n' {
					if _, err := s.ReadByte(); err != nil {
						return nil, err
					}
				}
				return result, nil
			}
		}
		if err := appendPart(s.buff[s.curr:s.length]); err != nil {
			return nil, err
		}
		if err := s.Skip(s.length - s.curr); err != nil {
			return nil, err
		}
	}
	return result, nil
}

type BufferedDataOutputStream struct {
	out    io.Writer
	buff   [8192]byte
	length int
	temp   [8]byte
}

func NewBufferedDataOutputStream(output ...io.Writer) *BufferedDataOutputStream {
	s := &BufferedDataOutputStream{}
	if len(output) > 0 {
		s.out = output[0]
	}
	return s
}
func NewBufferedDataOutputStreamFile(name string) (*BufferedDataOutputStream, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	}
	return NewBufferedDataOutputStream(f), nil
}
func (s *BufferedDataOutputStream) Open(output io.Writer) { s.out = output; s.length = 0 }
func (s *BufferedDataOutputStream) OpenFile(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	s.Open(f)
	return nil
}
func (s *BufferedDataOutputStream) writeBuffer() error {
	if s.out == nil {
		panic(NewNullPointerException())
	}
	n, err := s.out.Write(s.buff[:s.length])
	if err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	if n != s.length {
		return bufferedRandomAccessFileIOError(io.ErrShortWrite)
	}
	return nil
}
func (s *BufferedDataOutputStream) Flush() error {
	if err := s.writeBuffer(); err != nil {
		return err
	}
	if flusher, ok := s.out.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return bufferedRandomAccessFileIOError(err)
		}
	}
	s.length = 0
	return nil
}
func (s *BufferedDataOutputStream) Close() error {
	if err := s.Flush(); err != nil {
		return err
	}
	if closer, ok := s.out.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return bufferedRandomAccessFileIOError(err)
		}
	}
	s.out = nil
	return nil
}
func (s *BufferedDataOutputStream) Write(p []byte) (int, error) {
	err := s.WriteBytes(p, 0, len(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
func (s *BufferedDataOutputStream) WriteBytes(p []byte, off, n int) error {
	for n > 0 {
		count := min(n, len(s.buff)-s.length)
		copy(s.buff[s.length:s.length+count], p[off:off+count])
		s.length += count
		off += count
		n -= count
		if s.length == len(s.buff) {
			if err := s.writeBuffer(); err != nil {
				return err
			}
			s.length = 0
		}
	}
	return nil
}
func (s *BufferedDataOutputStream) WriteByte(value int8) error {
	s.buff[s.length] = byte(value)
	s.length++
	if s.length == len(s.buff) {
		if err := s.writeBuffer(); err != nil {
			return err
		}
		s.length = 0
	}
	return nil
}
func (s *BufferedDataOutputStream) WriteBoolean(value bool) error {
	v := int8(0)
	if value {
		v = 1
	}
	return s.WriteByte(v)
}
func (s *BufferedDataOutputStream) WriteShort(value int16) error {
	binary.BigEndian.PutUint16(s.temp[:2], uint16(value))
	return s.WriteBytes(s.temp[:], 0, 2)
}
func (s *BufferedDataOutputStream) WriteInt(value int32) error {
	binary.BigEndian.PutUint32(s.temp[:4], uint32(value))
	return s.WriteBytes(s.temp[:], 0, 4)
}
func (s *BufferedDataOutputStream) WriteLong(value int64) error {
	binary.BigEndian.PutUint64(s.temp[:], uint64(value))
	return s.WriteBytes(s.temp[:], 0, 8)
}
func (s *BufferedDataOutputStream) WriteFloat(value float32) error {
	bits := math.Float32bits(value)
	if math.IsNaN(float64(value)) {
		bits = 0x7fc00000
	}
	return s.WriteInt(int32(bits))
}
func (s *BufferedDataOutputStream) WriteDouble(value float64) error {
	bits := math.Float64bits(value)
	if math.IsNaN(value) {
		bits = 0x7ff8000000000000
	}
	return s.WriteLong(int64(bits))
}
func (s *BufferedDataOutputStream) WriteString(value string) error {
	units := javaStringUTF16(value)
	return s.WriteChars(units, 0, len(units))
}
func (s *BufferedDataOutputStream) WriteChars(units []uint16, off, n int) error {
	final := off + n
	for off < final {
		end := min(final, off+len(s.buff)-s.length)
		for off < end {
			s.buff[s.length] = byte(units[off])
			s.length++
			off++
		}
		if s.length == len(s.buff) {
			if err := s.writeBuffer(); err != nil {
				return err
			}
			s.length = 0
		}
	}
	return nil
}
func (s *BufferedDataOutputStream) WriteAnyString(value *string) error {
	if value == nil {
		return s.WriteInt(-1)
	}
	if err := s.WriteInt(int32(len(javaStringUTF16(*value)))); err != nil {
		return err
	}
	return s.WriteString(*value)
}
func (s *BufferedDataOutputStream) WriteLine(value string) error {
	if err := s.WriteString(value); err != nil {
		return err
	}
	return s.WriteByte('\n')
}
