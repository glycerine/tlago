package tlc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type IntQueue struct {
	m        []int32
	head     int
	diskdir  string
	filename string
}

func NewIntQueue() *IntQueue {
	return NewIntQueueWithCapacity(0)
}

func NewIntQueueWithCapacity(capacity int) *IntQueue {
	if capacity < 0 {
		capacity = 0
	}
	return &IntQueue{m: make([]int32, 0, capacity)}
}

func NewIntQueueWithDisk(metadir string, filename string, capacity int) *IntQueue {
	q := NewIntQueueWithCapacity(capacity)
	q.diskdir = metadir
	q.filename = filename
	return q
}

func (q *IntQueue) EnqueueInt(elem int32) {
	q.compactIfWorthwhile()
	q.m = append(q.m, elem)
}

func (q *IntQueue) EnqueueLong(elem int64) {
	q.EnqueueInt(int32(uint64(elem) >> 32))
	q.EnqueueInt(int32(uint64(elem) & 0xffffffff))
}

func (q *IntQueue) DequeueInt() int32 {
	if q.Size() < 1 {
		panic("IntQueue is empty")
	}
	res := q.m[q.head]
	q.head++
	if q.head == len(q.m) {
		q.m = q.m[:0]
		q.head = 0
	}
	return res
}

func (q *IntQueue) DequeueLong() int64 {
	high := int64(q.DequeueInt())
	low := int64(q.DequeueInt())
	return (high << 32) | (low & 0xffffffff)
}

func (q *IntQueue) PopInt() int32 {
	if q.Size() < 1 {
		panic("IntQueue is empty")
	}
	last := len(q.m) - 1
	res := q.m[last]
	q.m = q.m[:last]
	if q.head == len(q.m) {
		q.m = q.m[:0]
		q.head = 0
	}
	return res
}

func (q *IntQueue) PopLong() int64 {
	low := int64(q.PopInt())
	high := int64(q.PopInt())
	return (high << 32) | (low & 0xffffffff)
}

func (q *IntQueue) Size() int {
	return len(q.m) - q.head
}

func (q *IntQueue) HasElements() bool {
	return q.Size() > 0
}

func (q *IntQueue) Reset() {
	var zero int32
	for i := range q.m {
		q.m[i] = zero
	}
	q.m = q.m[:0]
	q.head = 0
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
		q.filename = "queue"
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
	for i := q.head; i < len(q.m); i++ {
		if err := out.WriteInt(q.m[i]); err != nil {
			_ = out.Close()
			return err
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
		return fmt.Errorf("IntQueue.CommitChkpt: cannot delete %s: %w", oldName, err)
	}
	if err := os.Rename(newName, oldName); err != nil {
		return fmt.Errorf("IntQueue.CommitChkpt: cannot rename %s to %s: %w", newName, oldName, err)
	}
	return nil
}

func (q *IntQueue) Recover() error {
	if q.diskdir == "" {
		return nil
	}
	if q.filename == "" {
		q.filename = "queue"
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
	q.Reset()
	if cap(q.m) < int(size) {
		q.m = make([]int32, 0, int(size))
	}
	for i := int32(0); i < size; i++ {
		value, err := in.ReadInt()
		if errors.Is(err, io.EOF) {
			return io.ErrUnexpectedEOF
		}
		if err != nil {
			return err
		}
		q.m = append(q.m, value)
	}
	q.head = 0
	return nil
}

func (q *IntQueue) String() string {
	return fmt.Sprintf("IntQueue{size:%d}", q.Size())
}

func (q *IntQueue) chkptName(ext string) string {
	filename := q.filename
	if filename == "" {
		filename = "queue"
	}
	return filepath.Join(q.diskdir, filename+"."+ext)
}

func (q *IntQueue) compactIfWorthwhile() {
	if q.head == 0 {
		return
	}
	if q.head == len(q.m) {
		q.m = q.m[:0]
		q.head = 0
		return
	}
	if q.head < len(q.m)/2 {
		return
	}
	copy(q.m, q.m[q.head:])
	q.m = q.m[:len(q.m)-q.head]
	q.head = 0
}
