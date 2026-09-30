package tlc

import "fmt"

type IntQueue struct {
	m    []int32
	head int
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

func (q *IntQueue) String() string {
	return fmt.Sprintf("IntQueue{size:%d}", q.Size())
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
