package tlc

import (
	"bufio"
	"errors"
	"os"
	"sort"
	"strings"
)

const (
	bigSetDefaultMaxSize     = 10000
	bigSetDefaultInitialSize = 840
)

type BigSet struct {
	MaxSize     int
	InitialSize int
	File        string
	FilePtr     int
	Els         *InsMap[string, *BigInt]
}

func NewBigSet(file string) *BigSet {
	return NewBigSetWithSizes(bigSetDefaultMaxSize, bigSetDefaultInitialSize, file)
}

func NewBigSetMax(maxSize int, file string) *BigSet {
	return NewBigSetWithSizes(maxSize, bigSetDefaultInitialSize, file)
}

func NewBigSetWithSizes(maxSize int, initialSize int, file string) *BigSet {
	maxSize = int(int32(maxSize))
	initialSize = int(int32(initialSize))
	if initialSize <= 0 {
		panic(NewIllegalArgumentException())
	}
	set := &BigSet{
		MaxSize:     maxSize,
		InitialSize: initialSize,
		File:        file,
		Els:         NewInsMap[string, *BigInt](),
	}
	if err := touchFile(file + "0"); err != nil {
		printDistributedFileException(err)
	} else if err := touchFile(file + "1"); err != nil {
		printDistributedFileException(err)
	}
	return set
}

func (s *BigSet) Size() int {
	if s == nil || s.Els == nil {
		panic(NewNullPointerException())
	}
	return s.Els.Len()
}

func (s *BigSet) Write() error {
	if s == nil {
		panic(NewNullPointerException())
	}
	values := make([]*BigInt, 0, s.Size())
	for _, value := range s.Els.All() {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Cmp(values[j]) < 0 })
	file, err := openFileOutputStreamAppend(s.File + intString(s.FilePtr))
	if err != nil {
		return err
	}
	buffer := bufio.NewWriterSize(file, 1024)
	if err := WriteInt(buffer, int32(len(values))); err != nil {
		return err
	}
	flushErr := bufferedRandomAccessFileIOError(buffer.Flush())
	closeErr := bufferedRandomAccessFileIOError(file.Close())
	if flushErr != nil {
		javaSuppressCloseError(flushErr, closeErr)
		return flushErr
	}
	if closeErr != nil {
		return closeErr
	}
	if s.FilePtr == 0 {
		s.FilePtr = 1
	} else {
		s.FilePtr = 0
	}
	return nil
}

func (s *BigSet) Clear() error {
	if s == nil {
		panic(NewNullPointerException())
	}
	if err := s.Delete(); err != nil {
		return err
	}
	s.Els = NewInsMap[string, *BigInt]()
	return nil
}

func (s *BigSet) Delete() error {
	if s == nil {
		panic(NewNullPointerException())
	}
	for _, suffix := range []string{"0", "1"} {
		_ = os.Remove(s.File + suffix)
	}
	s.FilePtr = 0
	return nil
}

func (s *BigSet) Put(key *BigInt) error {
	if s == nil || key == nil || s.Els == nil {
		panic(NewNullPointerException())
	}
	s.Els.Set(key.String(), key)
	if s.Size() >= s.MaxSize {
		if err := s.Write(); err != nil {
			return err
		}
		s.Els = NewInsMap[string, *BigInt]()
	}
	return nil
}

func (s *BigSet) Contains(key *BigInt) bool {
	if s == nil || s.Els == nil || key == nil {
		panic(NewNullPointerException())
	}
	_, ok := s.Els.Get2(key.String())
	return ok
}

func touchFile(name string) error {
	name = filenameNormalizeFile(name)
	if strings.IndexByte(name, 0) >= 0 {
		return NewIOException("Invalid file path")
	}
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	return bufferedRandomAccessFileIOError(file.Close())
}

func intString(value int) string {
	if value == 0 {
		return "0"
	}
	return "1"
}
