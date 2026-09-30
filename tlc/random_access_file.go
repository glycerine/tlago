package tlc

import (
	"io"
	"os"
	"time"
)

type RandomAccessFile struct {
	file           *os.File
	SuperSeekCnt   int64
	SuperReadCnt   int64
	SuperWriteCnt  int64
	SuperSeekTime  time.Duration
	SuperReadTime  time.Duration
	SuperWriteTime time.Duration
}

func NewRandomAccessFile(name string, mode string) (*RandomAccessFile, error) {
	flag := os.O_RDONLY
	switch mode {
	case "r":
	case "rw":
		flag = os.O_RDWR | os.O_CREATE
	default:
		return nil, newTLCError(ECGeneral, "unsupported random access file mode %q", mode)
	}
	file, err := os.OpenFile(name, flag, 0o666)
	if err != nil {
		return nil, err
	}
	return &RandomAccessFile{file: file}, nil
}

func (f *RandomAccessFile) Close() error {
	if f == nil || f.file == nil {
		return nil
	}
	return f.file.Close()
}

func (f *RandomAccessFile) Seek(pos int64) error {
	f.SuperSeekCnt++
	start := time.Now()
	_, err := f.file.Seek(pos, io.SeekStart)
	f.SuperSeekTime += time.Since(start)
	return err
}

func (f *RandomAccessFile) ReadByteValue() (int, error) {
	var buf [1]byte
	n, err := f.Read(buf[:])
	if err == io.EOF {
		return -1, nil
	}
	if err != nil {
		return -1, err
	}
	if n == 0 {
		return -1, nil
	}
	return int(buf[0]), nil
}

func (f *RandomAccessFile) Read(p []byte) (int, error) {
	f.SuperReadCnt++
	start := time.Now()
	n, err := f.file.Read(p)
	f.SuperReadTime += time.Since(start)
	return n, err
}

func (f *RandomAccessFile) WriteByteValue(value int) error {
	_, err := f.Write([]byte{byte(value)})
	return err
}

func (f *RandomAccessFile) Write(p []byte) (int, error) {
	f.SuperWriteCnt++
	start := time.Now()
	n, err := f.file.Write(p)
	f.SuperWriteTime += time.Since(start)
	return n, err
}
