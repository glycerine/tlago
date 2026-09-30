package tlc

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	BufferedRandomAccessFileLogBuffSz = 13
	BufferedRandomAccessFileBuffSz    = 1 << BufferedRandomAccessFileLogBuffSz
	bufferedRandomAccessFileBuffMask  = ^int64(BufferedRandomAccessFileBuffSz - 1)
)

type BufferedRandomAccessFile struct {
	file     *os.File
	writable bool
	dirty    bool
	closed   bool
	length   int64
	curr     int64
	lo       int64
	buff     []byte
	diskPos  int64
	mark     int64
}

func NewBufferedRandomAccessFile(name string, mode string) (*BufferedRandomAccessFile, error) {
	flag := os.O_RDONLY
	writable := false
	switch mode {
	case "r":
	case "rw":
		flag = os.O_RDWR | os.O_CREATE
		writable = true
	default:
		return nil, newTLCError(ECGeneral, "unsupported random access file mode %q", mode)
	}
	file, err := os.OpenFile(name, flag, 0o666)
	if err != nil {
		return nil, err
	}
	raf := &BufferedRandomAccessFile{
		file:     file,
		writable: writable,
		buff:     make([]byte, BufferedRandomAccessFileBuffSz),
	}
	if err := raf.init(); err != nil {
		_ = file.Close()
		return nil, err
	}
	return raf, nil
}

func (f *BufferedRandomAccessFile) init() error {
	f.dirty = false
	f.lo = 0
	f.curr = 0
	f.diskPos = 0
	length, err := f.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	f.length = length
	if _, err := f.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return f.fillBuffer()
}

func (f *BufferedRandomAccessFile) requireOpenFile() error {
	if f == nil || f.closed || f.file == nil {
		return errors.New("File handle closed")
	}
	return nil
}

func (f *BufferedRandomAccessFile) InvalidateBufferedData() error {
	if err := f.Flush(); err != nil {
		return err
	}
	return f.fillBuffer()
}

func (f *BufferedRandomAccessFile) Close() error {
	if f == nil || f.closed {
		return nil
	}
	var err error
	if flushErr := f.Flush(); flushErr != nil {
		err = flushErr
	}
	if closeErr := f.file.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	f.closed = true
	return err
}

func (f *BufferedRandomAccessFile) Flush() error {
	if err := f.requireOpenFile(); err != nil {
		return err
	}
	_, err := f.flushBuffer()
	return err
}

func (f *BufferedRandomAccessFile) flushBuffer() (bool, error) {
	if !f.dirty {
		return false, nil
	}
	if !f.writable {
		return false, errors.New("file not open for writing")
	}
	length := min64(f.length-f.lo, BufferedRandomAccessFileBuffSz)
	if length > 0 {
		if f.diskPos != f.lo {
			if _, err := f.file.Seek(f.lo, io.SeekStart); err != nil {
				return false, err
			}
			f.diskPos = f.lo
		}
		n, err := f.file.Write(f.buff[:length])
		f.diskPos += int64(n)
		if err != nil {
			return false, err
		}
		if n != int(length) {
			return false, io.ErrShortWrite
		}
	}
	f.dirty = false
	return true, nil
}

func (f *BufferedRandomAccessFile) fillBuffer() error {
	if f.diskPos != f.lo {
		if _, err := f.file.Seek(f.lo, io.SeekStart); err != nil {
			return err
		}
		f.diskPos = f.lo
	}
	count := 0
	for count < len(f.buff) {
		n, err := f.file.Read(f.buff[count:])
		if n > 0 {
			count += n
			f.diskPos += int64(n)
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
	}
	return nil
}

func (f *BufferedRandomAccessFile) Seek(pos int64) error {
	_, err := f.Seeek(pos)
	return err
}

func (f *BufferedRandomAccessFile) Seeek(pos int64) (bool, error) {
	if err := f.requireOpenFile(); err != nil {
		return false, err
	}
	f.curr = pos
	if pos < f.lo || pos >= f.lo+BufferedRandomAccessFileBuffSz {
		if _, err := f.flushBuffer(); err != nil {
			return false, err
		}
		f.lo = pos & bufferedRandomAccessFileBuffMask
		if err := f.fillBuffer(); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (f *BufferedRandomAccessFile) GetFilePointer() (int64, error) {
	if err := f.requireOpenFile(); err != nil {
		return 0, err
	}
	return f.curr, nil
}

func (f *BufferedRandomAccessFile) Length() (int64, error) {
	if err := f.requireOpenFile(); err != nil {
		return 0, err
	}
	return f.length, nil
}

func (f *BufferedRandomAccessFile) SetLength(newLength int64) error {
	if err := f.requireOpenFile(); err != nil {
		return err
	}
	if !f.writable {
		return errors.New("file not open for writing")
	}
	if err := f.file.Truncate(newLength); err != nil {
		return err
	}
	f.length = newLength
	if pos, err := f.file.Seek(0, io.SeekCurrent); err == nil {
		f.diskPos = pos
	}
	if f.curr > newLength {
		return f.Seek(newLength)
	}
	return nil
}

func (f *BufferedRandomAccessFile) restoreInvariantsAfterIncreasingCurr() error {
	if f.curr >= f.lo+BufferedRandomAccessFileBuffSz {
		_, err := f.Seeek(f.curr)
		return err
	}
	return nil
}

func (f *BufferedRandomAccessFile) ReadByteValue() (int, error) {
	if err := f.requireOpenFile(); err != nil {
		return -1, err
	}
	if f.curr >= f.length {
		return -1, nil
	}
	result := int(f.buff[f.curr-f.lo])
	f.curr++
	if err := f.restoreInvariantsAfterIncreasingCurr(); err != nil {
		return -1, err
	}
	return result, nil
}

func (f *BufferedRandomAccessFile) Read(p []byte) (int, error) {
	if err := f.requireOpenFile(); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, nil
	}
	readable := min64(f.lo+BufferedRandomAccessFileBuffSz, f.length) - f.curr
	if readable <= 0 {
		return 0, io.EOF
	}
	n := min(len(p), int(readable))
	copy(p, f.buff[f.curr-f.lo:f.curr-f.lo+int64(n)])
	f.curr += int64(n)
	if err := f.restoreInvariantsAfterIncreasingCurr(); err != nil {
		return n, err
	}
	return n, nil
}

func (f *BufferedRandomAccessFile) ReadFull(p []byte) error {
	_, err := io.ReadFull(f, p)
	return err
}

func (f *BufferedRandomAccessFile) ReadSignedByte() (int, error) {
	value, err := f.ReadByteValue()
	if err != nil || value < 0 {
		return value, err
	}
	if value >= 128 {
		value -= 256
	}
	return value, nil
}

func (f *BufferedRandomAccessFile) ReadShort() (int16, error) {
	var buf [2]byte
	if err := f.ReadFull(buf[:]); err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(buf[:])), nil
}

func (f *BufferedRandomAccessFile) ReadInt() (int32, error) {
	var buf [4]byte
	if err := f.ReadFull(buf[:]); err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(buf[:])), nil
}

func (f *BufferedRandomAccessFile) ReadLong() (int64, error) {
	var buf [8]byte
	if err := f.ReadFull(buf[:]); err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(buf[:])), nil
}

func (f *BufferedRandomAccessFile) ReadShortNat() (int, error) {
	res, err := f.ReadSignedByte()
	if err != nil {
		return 0, err
	}
	if res >= 0 {
		return res, nil
	}
	next, err := f.ReadByteValue()
	if err != nil {
		return 0, err
	}
	res = (res << 16) | (next & 0xff)
	return -res, nil
}

func (f *BufferedRandomAccessFile) ReadNat() (int, error) {
	short, err := f.ReadShort()
	if err != nil {
		return 0, err
	}
	res := int(short)
	if res >= 0 {
		return res, nil
	}
	next, err := f.ReadShort()
	if err != nil {
		return 0, err
	}
	res = (res << 16) | (int(next) & 0xffff)
	return -res, nil
}

func (f *BufferedRandomAccessFile) ReadLongNat() (int64, error) {
	first, err := f.ReadInt()
	if err != nil {
		return 0, err
	}
	res := int64(first)
	if res >= 0 {
		return res, nil
	}
	next, err := f.ReadInt()
	if err != nil {
		return 0, err
	}
	res = (res << 32) | (int64(next) & 0xffffffff)
	return -res, nil
}

func (f *BufferedRandomAccessFile) WriteByteValue(value int) error {
	if err := f.requireOpenFile(); err != nil {
		return err
	}
	if !f.writable {
		return errors.New("file not open for writing")
	}
	f.buff[f.curr-f.lo] = byte(value)
	f.curr++
	f.dirty = true
	if f.curr > f.length {
		f.length = f.curr
	}
	return f.restoreInvariantsAfterIncreasingCurr()
}

func (f *BufferedRandomAccessFile) Write(p []byte) (int, error) {
	if err := f.requireOpenFile(); err != nil {
		return 0, err
	}
	if !f.writable {
		return 0, errors.New("file not open for writing")
	}
	written := 0
	for len(p) > 0 {
		n, err := f.writeAtMost(p)
		written += n
		p = p[n:]
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func (f *BufferedRandomAccessFile) WriteFull(p []byte) error {
	n, err := f.Write(p)
	if err != nil {
		return err
	}
	if n != len(p) {
		return io.ErrShortWrite
	}
	return nil
}

func (f *BufferedRandomAccessFile) WriteShort(value int16) error {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(value))
	return f.WriteFull(buf[:])
}

func (f *BufferedRandomAccessFile) WriteInt(value int32) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(value))
	return f.WriteFull(buf[:])
}

func (f *BufferedRandomAccessFile) WriteLong(value int64) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(value))
	return f.WriteFull(buf[:])
}

func (f *BufferedRandomAccessFile) WriteShortNat(x int) error {
	if x <= 0x7f {
		return f.WriteByteValue(x)
	}
	return f.WriteShort(int16(-x))
}

func (f *BufferedRandomAccessFile) WriteNat(x int) error {
	if x <= 0x7fff {
		return f.WriteShort(int16(x))
	}
	return f.WriteInt(int32(-x))
}

func (f *BufferedRandomAccessFile) WriteLongNat(x int64) error {
	if x <= 0x7fffffff {
		return f.WriteInt(int32(x))
	}
	return f.WriteLong(-x)
}

func (f *BufferedRandomAccessFile) writeAtMost(p []byte) (int, error) {
	writable := min(len(p), int(f.lo+BufferedRandomAccessFileBuffSz-f.curr))
	if writable <= 0 {
		return 0, nil
	}
	copy(f.buff[f.curr-f.lo:f.curr-f.lo+int64(writable)], p[:writable])
	f.dirty = true
	f.curr += int64(writable)
	if f.curr > f.length {
		f.length = f.curr
	}
	return writable, f.restoreInvariantsAfterIncreasingCurr()
}

func (f *BufferedRandomAccessFile) Reset() error {
	if err := f.SetLength(0); err != nil {
		return err
	}
	return f.init()
}

func (f *BufferedRandomAccessFile) GetMark() int64 {
	if f == nil {
		return 0
	}
	return f.mark
}

func (f *BufferedRandomAccessFile) Mark() int64 {
	old := f.mark
	f.mark = f.curr
	return old
}

func (f *BufferedRandomAccessFile) SeekAndMark(pos int64) error {
	f.mark = pos
	return f.Seek(pos)
}

func min64(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
