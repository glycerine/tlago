package tlc

import (
	"io"
	"os"
	"strings"
	"syscall"
	"time"
)

type RandomAccessFile struct {
	file           *os.File
	closed         bool
	SuperSeekCnt   int64
	SuperReadCnt   int64
	SuperWriteCnt  int64
	SuperSeekTime  int64
	SuperReadTime  int64
	SuperWriteTime int64
}

func NewRandomAccessFile(name string, mode string) (*RandomAccessFile, error) {
	name = filenameNormalizeFile(name)
	flag := os.O_RDONLY
	switch mode {
	case "r":
	case "rw", "rws", "rwd":
		flag = os.O_RDWR | os.O_CREATE
		if mode == "rws" {
			flag |= os.O_SYNC
		}
		if mode == "rwd" {
			flag |= fileDataSyncFlag()
		}
	default:
		return nil, NewIllegalArgumentException(`Illegal mode "` + mode + `" must be one of "r", "rw", "rws", or "rwd"`)
	}
	if strings.IndexByte(name, 0) >= 0 {
		return nil, NewFileNotFoundException("Invalid file path")
	}
	file, err := os.OpenFile(name, flag, 0o666)
	if err != nil {
		return nil, bufferedRandomAccessFileOpenError(name, err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, bufferedRandomAccessFileOpenError(name, err)
	}
	if info.IsDir() {
		_ = file.Close()
		return nil, bufferedRandomAccessFileOpenError(name, syscall.EISDIR)
	}
	return &RandomAccessFile{file: file}, nil
}

func (f *RandomAccessFile) Close() error {
	f.require()
	if f.file == nil || f.closed {
		return nil
	}
	f.closed = true
	return bufferedRandomAccessFileIOError(f.file.Close())
}

func (f *RandomAccessFile) Seek(pos int64) error {
	f.require()
	f.SuperSeekCnt++
	startMillis := time.Now().UnixMilli()
	if pos < 0 {
		return NewIOException("Negative seek offset")
	}
	_, err := f.file.Seek(pos, io.SeekStart)
	if err == nil {
		f.SuperSeekTime += time.Now().UnixMilli() - startMillis
	}
	return bufferedRandomAccessFileIOError(err)
}

func (f *RandomAccessFile) ReadByteValue() (int, error) {
	f.require()
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
	f.require()
	f.SuperReadCnt++
	startMillis := time.Now().UnixMilli()
	if p == nil {
		panic(NewNullPointerException())
	}
	n, err := f.file.Read(p)
	if err == nil || err == io.EOF {
		f.SuperReadTime += time.Now().UnixMilli() - startMillis
	}
	if err == io.EOF {
		return n, err
	}
	return n, bufferedRandomAccessFileIOError(err)
}

func (f *RandomAccessFile) WriteByteValue(value int) error {
	f.require()
	_, err := f.Write([]byte{byte(value)})
	return err
}

func (f *RandomAccessFile) Write(p []byte) (int, error) {
	f.require()
	f.SuperWriteCnt++
	startMillis := time.Now().UnixMilli()
	if p == nil {
		panic(NewNullPointerException())
	}
	n, err := f.file.Write(p)
	if err == nil {
		f.SuperWriteTime += time.Now().UnixMilli() - startMillis
	}
	return n, bufferedRandomAccessFileIOError(err)
}

func (f *RandomAccessFile) require() {
	if f == nil {
		panic(NewNullPointerException())
	}
}
