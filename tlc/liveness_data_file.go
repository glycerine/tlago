package tlc

import (
	"fmt"
	"io"
	"os"
)

// GraphNode and BitVector share the source primitive data protocol between
// value streams and buffered random-access graph files.
type dataOutput interface {
	WriteInt(int32) error
	WriteLong(int64) error
	WriteNat(int32) error
}
type dataInput interface {
	ReadInt() (int32, error)
	ReadLong() (int64, error)
	ReadNat() (int32, error)
}

// Graph storage uses the original BufferedRandomAccessFile rather than creating
// a sequential value stream at every seek. The Go seek convention is adapted
// without changing the buffered file's logical cursor or length.
type livenessDataFile struct{ *BufferedRandomAccessFile }

func newLivenessDataFile(name string) (*livenessDataFile, error) {
	file, err := NewBufferedRandomAccessFile(name, "rw")
	if err != nil {
		return nil, err
	}
	return &livenessDataFile{file}, nil
}
func (f *livenessDataFile) Seek(offset int64, whence int) (int64, error) {
	var base int64
	var err error
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		base, err = f.GetFilePointer()
	case io.SeekEnd:
		base, err = f.Length()
	default:
		return 0, fmt.Errorf("invalid seek whence %d", whence)
	}
	if err != nil {
		return 0, err
	}
	position := base + offset
	if err = f.BufferedRandomAccessFile.Seek(position); err != nil {
		return 0, err
	}
	return f.GetFilePointer()
}
func (f *livenessDataFile) Truncate(length int64) error { return f.SetLength(length) }
func (f *livenessDataFile) Sync() error                 { return f.Flush() }
func (f *livenessDataFile) WriteNat(value int32) error {
	return f.BufferedRandomAccessFile.WriteNat(int(value))
}
func (f *livenessDataFile) ReadNat() (int32, error) {
	value, err := f.BufferedRandomAccessFile.ReadNat()
	return int32(value), err
}

type livenessDataFileInfo struct {
	os.FileInfo
	size int64
}

func (i livenessDataFileInfo) Size() int64 { return i.size }
func (f *livenessDataFile) Stat() (os.FileInfo, error) {
	info, err := f.file.Stat()
	if err != nil {
		return nil, err
	}
	length, err := f.Length()
	if err != nil {
		return nil, err
	}
	return livenessDataFileInfo{info, length}, nil
}
