package tlc

import (
	"fmt"
	"os"
	"strings"
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
	capacity = int(int32(capacity))
	if capacity < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(capacity)))
	}
	return &IntQueue{elems: make([]int32, capacity)}
}

func NewIntQueueWithDisk(metadir string, filename string, capacity int) *IntQueue {
	q := NewIntQueueWithCapacity(capacity)
	q.diskdir = metadir
	q.filename = filename
	return q
}

func (q *IntQueue) EnqueueInt(elem int32) {
	q.require()
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
	q.require()
	q.EnqueueInt(int32(uint64(elem) >> 32))
	q.EnqueueInt(int32(uint64(elem) & 0xffffffff))
}

func (q *IntQueue) DequeueInt() int32 {
	q.require()
	if q.size < 1 {
		panic(NewNoSuchElementException())
	}
	res := q.elems[q.start]
	q.size--
	q.start = (q.start + 1) % len(q.elems)
	return res
}

func (q *IntQueue) DequeueLong() int64 {
	q.require()
	high := int64(q.DequeueInt())
	low := int64(q.DequeueInt())
	return (high << 32) | (low & 0xffffffff)
}

func (q *IntQueue) PopInt() int32 {
	q.require()
	if q.size < 1 {
		panic(NewNoSuchElementException())
	}
	q.size--
	return q.elems[q.size]
}

func (q *IntQueue) PopLong() int64 {
	q.require()
	low := int64(q.PopInt())
	high := int64(q.PopInt())
	return (high << 32) | (low & 0xffffffff)
}

func (q *IntQueue) Size() int {
	q.require()
	return q.size
}

func (q *IntQueue) HasElements() bool {
	q.require()
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
	q.require()
	file, err := openPoolFile(q.chkptName("tmp"), true)
	if err != nil {
		return err
	}
	out := NewBufferedDataOutputStream(file)
	if err := out.WriteInt(int32(q.Size())); err != nil {
		return err
	}
	index := q.start
	for i := 0; i < q.size; i++ {
		if index < 0 || index >= len(q.elems) {
			return NewArrayIndexOutOfBoundsException(index, len(q.elems))
		}
		if err := out.WriteInt(q.elems[index]); err != nil {
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
	q.require()
	oldName := q.chkptName("chkpt")
	newName := q.chkptName("tmp")
	if _, err := os.Stat(oldName); err == nil {
		if err := os.Remove(oldName); err != nil {
			return NewIOException(fmt.Sprintf("MemStateQueue.commitChkpt: cannot delete %s", oldName))
		}
	}
	if err := os.Rename(newName, oldName); err != nil {
		return NewIOException(fmt.Sprintf("MemStateQueue.commitChkpt: cannot delete %s", oldName))
	}
	return nil
}

func (q *IntQueue) Recover() error {
	q.require()
	file, err := openPoolFile(q.chkptName("chkpt"), false)
	if err != nil {
		return err
	}
	in, err := NewBufferedDataInputStream(file)
	if err != nil {
		return err
	}
	size, err := in.ReadInt()
	if err != nil {
		return err
	}
	q.size = int(size)
	for i := int32(0); i < size; i++ {
		value, err := in.ReadInt()
		if err != nil {
			return err
		}
		// Java evaluates the input read before checking the assignment index.
		if int(i) >= len(q.elems) {
			return NewArrayIndexOutOfBoundsException(int(i), len(q.elems))
		}
		q.elems[i] = value
	}
	return in.Close()
}

func (q *IntQueue) String() string {
	return fmt.Sprintf("IntQueue{size:%d}", q.Size())
}

func (q *IntQueue) require() {
	if q == nil {
		panic(NewNullPointerException())
	}
}

func MemIntQueueDetailedFormatterFPAndTidx(queue *IntQueue) string {
	if queue == nil {
		panic(NewNullPointerException())
	}
	var out strings.Builder
	for i := 0; i < queue.size; i += 3 {
		fp := (int64(intQueueDetailedElement(queue, i)) << 32) |
			(int64(intQueueDetailedElement(queue, i+1)) & 0xffffffff)
		fmt.Fprintf(&out, "fp: %d tidx: %d\n", fp, intQueueDetailedElement(queue, i+2))
	}
	return out.String()
}

func MemIntQueueDetailedFormatterFPAndTidxAndPtr(queue *IntQueue) string {
	if queue == nil {
		panic(NewNullPointerException())
	}
	var out strings.Builder
	for i := 0; i < queue.size; i += 5 {
		fp := (int64(intQueueDetailedElement(queue, i)) << 32) |
			(int64(intQueueDetailedElement(queue, i+1)) & 0xffffffff)
		ptr := (int64(intQueueDetailedElement(queue, i+3)) << 32) |
			(int64(intQueueDetailedElement(queue, i+4)) & 0xffffffff)
		fmt.Fprintf(&out, "fp: %d tidx: %d ptr: %d\n", fp, intQueueDetailedElement(queue, i+2), ptr)
	}
	return out.String()
}

func intQueueDetailedElement(queue *IntQueue, index int) int32 {
	if index < 0 || index >= len(queue.elems) {
		panic(NewArrayIndexOutOfBoundsException(index, len(queue.elems)))
	}
	return queue.elems[index]
}

func (q *IntQueue) chkptName(ext string) string {
	filename := q.filename
	if filename == "" {
		filename = "null"
	}
	return filenameNormalizeFile(q.diskdir + string(os.PathSeparator) + filename + "." + ext)
}

func (q *IntQueue) ensureCapacity(minCapacity int) int {
	// MemBasedSet calculates in long, then casts and adds in Java int.
	newSize := int32((int64(int32(q.size))*3)/2) + 1
	if minimum := int32(q.size) + int32(minCapacity); newSize < minimum {
		newSize = minimum
	}
	if newSize < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(newSize)))
	}
	return int(newSize)
}
