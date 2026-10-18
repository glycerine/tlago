package tlc

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	if len(input) > 1 {
		panic(NewIllegalArgumentException("NewBufferedDataInputStream accepts at most one input"))
	}
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
	f, err := openFileInputStream(name)
	if err != nil {
		return nil, err
	}
	s, err := NewBufferedDataInputStream(f)
	return s, err
}
func (s *BufferedDataInputStream) refill() error {
	if s == nil {
		panic(NewNullPointerException())
	}
	if isNil(s.in) {
		panic(NewNullPointerException())
	}
	if s.pending != nil {
		err := s.pending
		s.pending = nil
		if bufferedInputEOF(err) {
			s.length = -1
			return nil
		}
		return bufferedRandomAccessFileIOError(err)
	}
	if input, ok := s.in.(*BufferedDataInputStream); ok {
		// Java throws even after the child has copied bytes. Keep those bytes,
		// but do not publish a new length or defer the child's failure.
		n, err := input.ReadBytes(s.buff[:], 0, len(s.buff))
		if err != nil {
			return err
		}
		s.length = n
		if n == 0 {
			panic(NewTLCRuntimeException(ECSystemStreamEmpty))
		}
		return nil
	}
	n, err := s.in.Read(s.buff[:])
	// Go permits bytes and EOF in the same read; Java reports EOF next time.
	if n > 0 {
		s.length = n
		s.pending = err
		return nil
	}
	if bufferedInputEOF(err) {
		s.length = -1
		return nil
	}
	if err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	s.length = 0
	panic(NewTLCRuntimeException(ECSystemStreamEmpty))
}

func bufferedInputEOF(err error) bool {
	// Source checked EOFException is a failed read, distinct from an
	// io.Reader's ordinary end-of-stream result.
	if _, represented := err.(interface{ GetMessage() *string }); represented {
		return false
	}
	return errors.Is(err, io.EOF)
}
func (s *BufferedDataInputStream) Open(input io.Reader) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	if !isNil(s.in) {
		panic(NewTLCRuntimeException(ECSystemStreamEmpty))
	}
	s.in = input
	s.pending = nil
	return s.refill()
}
func (s *BufferedDataInputStream) OpenFile(name string) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	f, err := openFileInputStream(name)
	if err != nil {
		return err
	}
	return s.Open(f)
}
func (s *BufferedDataInputStream) Close() error {
	if s == nil {
		panic(NewNullPointerException())
	}
	if isNil(s.in) {
		panic(NewNullPointerException())
	}
	if closer, ok := s.in.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			// FileInputStream.close succeeds when another owner closed the file.
			_, file := s.in.(*os.File)
			if !file || !errors.Is(err, os.ErrClosed) {
				return bufferedRandomAccessFileIOError(err)
			}
		}
	}
	s.in = nil
	return nil
}
func (s *BufferedDataInputStream) AtEOF() bool {
	if s == nil {
		panic(NewNullPointerException())
	}
	return s.length < 0
}
func (s *BufferedDataInputStream) Read(p []byte) (int, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	if p == nil {
		panic(NewNullPointerException())
	}
	n, err := s.ReadBytes(p, 0, len(p))
	if n < 0 {
		return 0, io.EOF
	}
	return n, err
}

// ReadBytes returns Java's -1 at EOF, retaining the original offset/count API.
func (s *BufferedDataInputStream) ReadBytes(p []byte, off, n int) (int, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	off = int(int32(off))
	n = int(int32(n))
	if s.length < 0 {
		return -1, nil
	}
	initial := off
	for n > 0 && s.length > 0 {
		count := min(n, int(int32(s.length)-int32(s.curr)))
		bufferedByteArrayCopy(s.buff[:], s.curr, p, off, count)
		s.curr = int(int32(s.curr) + int32(count))
		off = int(int32(off) + int32(count))
		n = int(int32(n) - int32(count))
		if s.curr == s.length {
			if err := s.refill(); err != nil {
				return off - initial, err
			}
			s.curr = 0
		}
	}
	return int(int32(off) - int32(initial)), nil
}
func (s *BufferedDataInputStream) ReadFully(p []byte, rangeArgs ...int) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	if len(rangeArgs) != 0 && len(rangeArgs) != 2 {
		panic(NewIllegalArgumentException("ReadFully accepts either no range or an offset and length"))
	}
	if len(rangeArgs) == 0 && p == nil {
		panic(NewNullPointerException())
	}
	off, n := 0, len(p)
	if len(rangeArgs) > 0 {
		off, n = int(int32(rangeArgs[0])), int(int32(rangeArgs[1]))
	}
	for n > 0 {
		read, err := s.ReadBytes(p, off, n)
		if err != nil {
			return err
		}
		if read < 0 {
			return NewEOFException()
		}
		off = int(int32(off) + int32(read))
		n = int(int32(n) - int32(read))
	}
	return nil
}
func (s *BufferedDataInputStream) ReadByte() (int8, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	if s.length < 0 {
		return 0, NewEOFException()
	}
	index := s.curr
	s.curr = int(int32(s.curr) + 1)
	value := int8(bufferedByteArrayAt(s.buff[:], index))
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
	if s == nil {
		panic(NewNullPointerException())
	}
	if err := s.ReadFully(s.temp[:], 0, 2); err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(s.temp[:2])), nil
}
func (s *BufferedDataInputStream) ReadInt() (int32, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	if err := s.ReadFully(s.temp[:], 0, 4); err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(s.temp[:4])), nil
}
func (s *BufferedDataInputStream) ReadLong() (int64, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	if err := s.ReadFully(s.temp[:], 0, 8); err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(s.temp[:])), nil
}
func (s *BufferedDataInputStream) ReadString(n int) (string, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	n = int(int32(n))
	if n < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(n)))
	}
	units := make([]uint16, n)
	off := 0
	for n > 0 {
		if s.length < 0 {
			return "", NewEOFException()
		}
		for n > 0 && s.length > 0 {
			count := min(n, int(int32(s.length)-int32(s.curr)))
			for i := 0; i < count; i++ {
				index := int(int32(off) + int32(i))
				value := uint16(int16(int8(bufferedByteArrayAt(s.buff[:], int(int32(s.curr)+int32(i))))))
				if index < 0 || index >= len(units) {
					panic(NewArrayIndexOutOfBoundsException(index, len(units)))
				}
				units[index] = value
			}
			s.curr = int(int32(s.curr) + int32(count))
			off = int(int32(off) + int32(count))
			n = int(int32(n) - int32(count))
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
	if s == nil {
		panic(NewNullPointerException())
	}
	n = int(int32(n))
	for s.length > 0 && int32(s.curr)+int32(n) >= int32(s.length) {
		n = int(int32(n) - (int32(s.length) - int32(s.curr)))
		if err := s.refill(); err != nil {
			return err
		}
		s.curr = 0
	}
	if n > 0 && s.length < 0 {
		return NewEOFException()
	}
	s.curr = int(int32(s.curr) + int32(n))
	if !(s.length < 0 || s.curr < s.length) {
		panic(NewTLCRuntimeException(ECSystemIndexError))
	}
	return nil
}
func (s *BufferedDataInputStream) ReadLine() (*string, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	var result *string
	appendPart := func(start, count int) error {
		if start < 0 || count < 0 || start > len(s.buff)-count {
			panic(&StringIndexOutOfBoundsException{IndexOutOfBoundsException: &IndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(javaString(fmt.Sprintf("Range [%d, %d + %d) out of bounds for length %d", start, start, count, len(s.buff))), nil)}})
		}
		data := s.buff[start : start+count]
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
			if bufferedByteArrayAt(s.buff[:], i) == '\n' || bufferedByteArrayAt(s.buff[:], i) == '\r' {
				eol := s.buff[i]
				if err := appendPart(s.curr, i-s.curr); err != nil {
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
		if err := appendPart(s.curr, s.length-s.curr); err != nil {
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
	if len(output) > 1 {
		panic(NewIllegalArgumentException("NewBufferedDataOutputStream accepts at most one output"))
	}
	s := &BufferedDataOutputStream{}
	if len(output) > 0 {
		s.out = output[0]
	}
	return s
}
func NewBufferedDataOutputStreamFile(name string) (*BufferedDataOutputStream, error) {
	f, err := openFileOutputStream(name)
	if err != nil {
		return nil, err
	}
	return NewBufferedDataOutputStream(f), nil
}
func (s *BufferedDataOutputStream) Open(output io.Writer) {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.out = output
	s.length = 0
}
func (s *BufferedDataOutputStream) OpenFile(name string) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	f, err := openFileOutputStream(name)
	if err != nil {
		return err
	}
	s.Open(f)
	return nil
}
func (s *BufferedDataOutputStream) writeBuffer() error {
	if s == nil {
		panic(NewNullPointerException())
	}
	if isNil(s.out) {
		panic(NewNullPointerException())
	}
	// Preserve range delegation through another native buffered stream.
	if buffer, ok := s.out.(*BufferedDataOutputStream); ok {
		return buffer.WriteBytes(s.buff[:], 0, s.length)
	}
	if wrapper, ok := s.out.(valueStreamWriteCloser); ok {
		if gzip, ok := wrapper.Writer.(*zlibGzipWriter); ok {
			return bufferedRandomAccessFileIOError(gzip.WriteBytes(s.buff[:], 0, s.length))
		}
	}
	_, file := s.out.(*os.File)
	if s.length < 0 || s.length > len(s.buff) {
		var message *string
		if !file {
			message = javaString(fmt.Sprintf("Range [0, 0 + %d) out of bounds for length %d", s.length, len(s.buff)))
		}
		panic(&IndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(message, nil)})
	}
	// FileOutputStream's native write returns before accessing the descriptor
	// for an empty range, including after the file has been closed.
	if file && s.length == 0 {
		return nil
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
	if s == nil {
		panic(NewNullPointerException())
	}
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
	if s == nil {
		panic(NewNullPointerException())
	}
	if err := s.Flush(); err != nil {
		return err
	}
	if closer, ok := s.out.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			_, file := s.out.(*os.File)
			if !file || !errors.Is(err, os.ErrClosed) {
				return bufferedRandomAccessFileIOError(err)
			}
		}
	}
	s.out = nil
	return nil
}
func (s *BufferedDataOutputStream) Write(p []byte) (int, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	if p == nil {
		panic(NewNullPointerException())
	}
	err := s.WriteBytes(p, 0, len(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
func (s *BufferedDataOutputStream) WriteBytes(p []byte, off, n int) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	off = int(int32(off))
	n = int(int32(n))
	for n > 0 {
		count := min(n, int(int32(len(s.buff))-int32(s.length)))
		bufferedByteArrayCopy(p, off, s.buff[:], s.length, count)
		s.length = int(int32(s.length) + int32(count))
		off = int(int32(off) + int32(count))
		n = int(int32(n) - int32(count))
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
	if s == nil {
		panic(NewNullPointerException())
	}
	index := s.length
	s.length = int(int32(s.length) + 1)
	if index < 0 || index >= len(s.buff) {
		panic(NewArrayIndexOutOfBoundsException(index, len(s.buff)))
	}
	s.buff[index] = byte(value)
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
	if s == nil {
		panic(NewNullPointerException())
	}
	binary.BigEndian.PutUint16(s.temp[:2], uint16(value))
	return s.WriteBytes(s.temp[:], 0, 2)
}
func (s *BufferedDataOutputStream) WriteInt(value int32) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	binary.BigEndian.PutUint32(s.temp[:4], uint32(value))
	return s.WriteBytes(s.temp[:], 0, 4)
}
func (s *BufferedDataOutputStream) WriteLong(value int64) error {
	if s == nil {
		panic(NewNullPointerException())
	}
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
	if s == nil {
		panic(NewNullPointerException())
	}
	units := javaStringUTF16(value)
	n, off := len(units), 0
	for n > 0 {
		count := min(n, int(int32(len(s.buff))-int32(s.length)))
		end := int(int32(off) + int32(count))
		if off < 0 || end < off || end > len(units) {
			panic(bufferedStringRangeException(fmt.Sprintf("Range [%d, %d) out of bounds for length %d", off, end, len(units))))
		}
		if s.length < 0 || count < 0 || s.length > len(s.buff)-count {
			panic(bufferedStringRangeException(fmt.Sprintf("Range [%d, %d + %d) out of bounds for length %d", s.length, s.length, count, len(s.buff))))
		}
		for i := 0; i < count; i++ {
			s.buff[s.length+i] = byte(units[off+i])
		}
		s.length = int(int32(s.length) + int32(count))
		off = end
		n = int(int32(n) - int32(count))
		if s.length == len(s.buff) {
			if err := s.writeBuffer(); err != nil {
				return err
			}
			s.length = 0
		}
	}
	return nil
}
func (s *BufferedDataOutputStream) WriteChars(units []uint16, off, n int) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	off = int(int32(off))
	n = int(int32(n))
	final := int(int32(off) + int32(n))
	for off < final {
		end := min(final, int(int32(off)+int32(len(s.buff))-int32(s.length)))
		for off < end {
			index := s.length
			s.length = int(int32(s.length) + 1)
			from := off
			off = int(int32(off) + 1)
			if units == nil {
				panic(NewNullPointerException())
			}
			if from < 0 || from >= len(units) {
				panic(NewArrayIndexOutOfBoundsException(from, len(units)))
			}
			value := byte(units[from])
			if index < 0 || index >= len(s.buff) {
				panic(NewArrayIndexOutOfBoundsException(index, len(s.buff)))
			}
			s.buff[index] = value
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

// System.arraycopy validates both byte-array ranges before copying any bytes.
func bufferedByteArrayCopy(source []byte, sourceIndex int, destination []byte, destinationIndex int, length int) {
	if source == nil || destination == nil {
		panic(NewNullPointerException())
	}
	fail := func(message string) {
		panic(&ArrayIndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(javaString("arraycopy: "+message), nil)})
	}
	if sourceIndex < 0 {
		fail(fmt.Sprintf("source index %d out of bounds for byte[%d]", sourceIndex, len(source)))
	}
	if destinationIndex < 0 {
		fail(fmt.Sprintf("destination index %d out of bounds for byte[%d]", destinationIndex, len(destination)))
	}
	if length < 0 {
		fail(fmt.Sprintf("length %d is negative", length))
	}
	if int64(sourceIndex)+int64(length) > int64(len(source)) {
		fail(fmt.Sprintf("last source index %d out of bounds for byte[%d]", int64(sourceIndex)+int64(length), len(source)))
	}
	if int64(destinationIndex)+int64(length) > int64(len(destination)) {
		fail(fmt.Sprintf("last destination index %d out of bounds for byte[%d]", int64(destinationIndex)+int64(length), len(destination)))
	}
	copy(destination[destinationIndex:destinationIndex+length], source[sourceIndex:sourceIndex+length])
}

func bufferedByteArrayAt(values []byte, index int) byte {
	if index < 0 || index >= len(values) {
		panic(NewArrayIndexOutOfBoundsException(index, len(values)))
	}
	return values[index]
}

func bufferedStringRangeException(message string) *StringIndexOutOfBoundsException {
	return &StringIndexOutOfBoundsException{IndexOutOfBoundsException: &IndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(javaString(message), nil)}}
}
