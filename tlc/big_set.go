package tlc

import (
	"errors"
	"os"
	"sort"
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
	if maxSize <= 0 {
		maxSize = bigSetDefaultMaxSize
	}
	if initialSize <= 0 {
		initialSize = bigSetDefaultInitialSize
	}
	set := &BigSet{
		MaxSize:     maxSize,
		InitialSize: initialSize,
		File:        file,
		Els:         NewInsMap[string, *BigInt](),
	}
	_ = touchFile(file + "0")
	_ = touchFile(file + "1")
	return set
}

func (s *BigSet) Size() int {
	if s == nil || s.Els == nil {
		return 0
	}
	return s.Els.Len()
}

func (s *BigSet) Write() error {
	if s == nil {
		return nil
	}
	values := make([]*BigInt, 0, s.Size())
	for _, value := range s.Els.All() {
		values = append(values, value)
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Cmp(values[j]) < 0 })
	file, err := os.Create(s.File + intString(s.FilePtr))
	if err != nil {
		return err
	}
	if err := WriteInt(file, int32(len(values))); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
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
		return nil
	}
	if err := s.Delete(); err != nil {
		return err
	}
	s.Els = NewInsMap[string, *BigInt]()
	return nil
}

func (s *BigSet) Delete() error {
	if s == nil {
		return nil
	}
	var firstErr error
	for _, suffix := range []string{"0", "1"} {
		if err := os.Remove(s.File + suffix); err != nil && !errors.Is(err, os.ErrNotExist) && firstErr == nil {
			firstErr = err
		}
	}
	s.FilePtr = 0
	return firstErr
}

func (s *BigSet) Put(key *BigInt) error {
	if s == nil || key == nil {
		return nil
	}
	if s.Els == nil {
		s.Els = NewInsMap[string, *BigInt]()
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
		return false
	}
	_, ok := s.Els.Get2(key.String())
	return ok
}

func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0o666)
	if err != nil {
		return err
	}
	return file.Close()
}

func intString(value int) string {
	if value == 0 {
		return "0"
	}
	return "1"
}
