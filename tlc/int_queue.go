package tlc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type IntQueue struct {
	elems    []int32
	size     int
	start    int
	diskdir  string
	filename string
}

const intQueueInitialSize = 4096

func NewIntQueue() *IntQueue {
	return NewIntQueueWithCapacity(intQueueInitialSize)
}

func NewIntQueueWithCapacity(capacity int) *IntQueue {
	if capacity < 0 {
		capacity = 0
	}
	return &IntQueue{elems: make([]int32, capacity)}
}

func NewIntQueueWithDisk(metadir string, filename string, capacity int) *IntQueue {
	q := NewIntQueueWithCapacity(capacity)
	q.diskdir = metadir
	if filename == "" {
		filename = "null"
	}
	q.filename = filename
	return q
}

func (q *IntQueue) EnqueueInt(elem int32) {
	if q.size == len(q.elems) {
		newElems := make([]int32, q.ensureCapacity(intQueueInitialSize))
		copyLen := len(q.elems) - q.start
		copy(newElems, q.elems[q.start:])
		copy(newElems[copyLen:], q.elems[:q.start])
		q.elems = newElems
		q.start = 0
	}
	last := (q.start + q.size) % len(q.elems)
	q.elems[last] = elem
	q.size++
}

func (q *IntQueue) EnqueueLong(elem int64) {
	q.EnqueueInt(int32(uint64(elem) >> 32))
	q.EnqueueInt(int32(uint64(elem) & 0xffffffff))
}

func (q *IntQueue) DequeueInt() int32 {
	if q.size < 1 {
		panic("IntQueue is empty")
	}
	res := q.elems[q.start]
	q.size--
	q.start = (q.start + 1) % len(q.elems)
	return res
}

func (q *IntQueue) DequeueLong() int64 {
	high := int64(q.DequeueInt())
	low := int64(q.DequeueInt())
	return (high << 32) | (low & 0xffffffff)
}

func (q *IntQueue) PopInt() int32 {
	if q.size < 1 {
		panic("IntQueue is empty")
	}
	q.size--
	return q.elems[q.size]
}

func (q *IntQueue) PopLong() int64 {
	low := int64(q.PopInt())
	high := int64(q.PopInt())
	return (high << 32) | (low & 0xffffffff)
}

func (q *IntQueue) Size() int {
	return q.size
}

func (q *IntQueue) HasElements() bool {
	return q.Size() > 0
}

func (q *IntQueue) Reset() {
	var zero int32
	for i := range q.elems {
		q.elems[i] = zero
	}
	q.size = 0
	q.start = 0
}

func (q *IntQueue) BeginChkpt() error {
	if q.diskdir == "" {
		dir, err := os.MkdirTemp("", "MemIntQueue")
		if err != nil {
			return err
		}
		q.diskdir = dir
	}
	if q.filename == "" {
		q.filename = "null"
	}
	if err := os.MkdirAll(q.diskdir, 0o755); err != nil {
		return err
	}
	file, err := os.Create(q.chkptName("tmp"))
	if err != nil {
		return err
	}
	out := NewValueOutputStream(file)
	if err := out.WriteInt(int32(q.Size())); err != nil {
		_ = out.Close()
		return err
	}
	index := q.start
	for i := 0; i < q.size; i++ {
		if err := out.WriteInt(q.elems[index]); err != nil {
			_ = out.Close()
			return err
		}
		index++
		if index == len(q.elems) {
			index = 0
		}
	}
	return out.Close()
}

func (q *IntQueue) CommitChkpt() error {
	if q.diskdir == "" {
		return nil
	}
	oldName := q.chkptName("chkpt")
	newName := q.chkptName("tmp")
	if err := os.Remove(oldName); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("MemStateQueue.commitChkpt: cannot delete %s", oldName)
	}
	if err := os.Rename(newName, oldName); err != nil {
		return fmt.Errorf("MemStateQueue.commitChkpt: cannot delete %s", oldName)
	}
	return nil
}

func (q *IntQueue) Recover() error {
	if q.diskdir == "" {
		return nil
	}
	if q.filename == "" {
		q.filename = "null"
	}
	file, err := os.Open(q.chkptName("chkpt"))
	if err != nil {
		return err
	}
	in := NewValueInputStream(file)
	defer in.Close()
	size, err := in.ReadInt()
	if err != nil {
		return err
	}
	q.size = int(size)
	if len(q.elems) < q.size {
		q.elems = make([]int32, q.size)
	}
	for i := int32(0); i < size; i++ {
		value, err := in.ReadInt()
		if errors.Is(err, io.EOF) {
			return io.ErrUnexpectedEOF
		}
		if err != nil {
			return err
		}
		q.elems[i] = value
	}
	return nil
}

func (q *IntQueue) String() string {
	return fmt.Sprintf("IntQueue{size:%d}", q.Size())
}

func (q *IntQueue) chkptName(ext string) string {
	filename := q.filename
	if filename == "" {
		filename = "null"
	}
	return filepath.Join(q.diskdir, filename+"."+ext)
}

func (q *IntQueue) ensureCapacity(minCapacity int) int {
	newSize := int((int64(q.size)*3)/2) + 1
	if min := q.size + minCapacity; newSize < min {
		newSize = min
	}
	return newSize
}
