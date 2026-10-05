package tlc

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type LongArray struct {
	data []uint64
}

func NewLongArray(positions int64) *LongArray {
	return &LongArray{data: make([]uint64, positions)}
}

func NewLongArrayFrom(values []int64) *LongArray {
	out := NewLongArray(int64(len(values)))
	for i, value := range values {
		out.Set(int64(i), value)
	}
	return out
}

func LongArrayIsSupported() bool {
	return true
}

func (a *LongArray) ZeroMemory(numThreads ...int) error {
	if a == nil || len(a.data) == 0 {
		return nil
	}
	threads := 1
	if len(numThreads) > 0 && numThreads[0] > 1 {
		threads = numThreads[0]
	}
	if threads <= 1 || len(a.data) < threads {
		for i := range a.data {
			a.data[i] = 0
		}
		return nil
	}
	segment := len(a.data) / threads
	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		offset := i
		lower := segment * offset
		upper := (offset + 1) * segment
		if offset == threads-1 {
			upper = len(a.data)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pos := lower; pos < upper; pos++ {
				a.data[pos] = 0
			}
		}()
	}
	wg.Wait()
	return nil
}

func (a *LongArray) TrySet(position int64, expected int64, value int64) bool {
	a.rangeCheck(position)
	return atomic.CompareAndSwapUint64(&a.data[position], uint64(expected), uint64(value))
}

func (a *LongArray) Set(position int64, value int64) {
	a.rangeCheck(position)
	a.data[position] = uint64(value)
}

func (a *LongArray) Get(position int64) int64 {
	a.rangeCheck(position)
	return int64(a.data[position])
}

func (a *LongArray) Swap(position1 int64, position2 int64) {
	a.rangeCheck(position1)
	a.rangeCheck(position2)
	tmp := a.Get(position1)
	a.Set(position1, a.Get(position2))
	a.Set(position2, tmp)
}

func (a *LongArray) SwapCopy(position1 int64, position2 int64) {
	a.Swap(position1, position2)
}

func (a *LongArray) Size() int64 {
	if a == nil {
		return 0
	}
	return int64(len(a.data))
}

func (a *LongArray) String() string {
	if a == nil || a.Size() == 0 {
		return "[]"
	}
	return a.StringRange(0, a.Size()-1)
}

func (a *LongArray) StringRange(start int64, end int64) string {
	if a == nil || end == -1 {
		return "[]"
	}
	a.rangeCheck(start)
	a.rangeCheck(end)
	if start > end {
		return "[]"
	}
	var b strings.Builder
	b.WriteByte('[')
	for i := start; ; i++ {
		value := a.Get(i)
		if value == 0 {
			b.WriteByte('e')
		} else {
			b.WriteString(strconv.FormatInt(value, 10))
		}
		if i == end {
			b.WriteByte(']')
			return b.String()
		}
		b.WriteString(", ")
	}
}

func (a *LongArray) ToArray() []int64 {
	if a == nil {
		return nil
	}
	out := make([]int64, len(a.data))
	for i, value := range a.data {
		out[i] = int64(value)
	}
	return out
}

func (a *LongArray) rangeCheck(position int64) {
	if a == nil || position < 0 || position >= int64(len(a.data)) {
		panic(NewAssertionError())
	}
}

type LongComparator func(lo int64, loPos int64, hi int64, hiPos int64) int

func DefaultLongComparator(lo int64, loPos int64, hi int64, hiPos int64) int {
	switch {
	case lo < hi:
		return -1
	case lo > hi:
		return 1
	default:
		return 0
	}
}

func LongArraysSort(array *LongArray) {
	if array == nil {
		return
	}
	LongArraysSortRange(array, 0, array.Size()-1, DefaultLongComparator)
}

func LongArraysSortRange(array *LongArray, left int64, right int64, cmp LongComparator) {
	if array == nil || array.Size() == 0 {
		return
	}
	if cmp == nil {
		cmp = DefaultLongComparator
	}
	size := array.Size()
	for i, j := left, left; i < right; {
		lo := (i + 1) % size
		ai := array.Get(lo)
		for cmp(ai, lo, array.Get(j%size), j%size) <= -1 {
			array.Set((j+1)%size, array.Get(j%size))
			if j == left {
				j--
				break
			}
			j--
		}
		array.Set((j+1)%size, ai)
		i++
		j = i
	}
}

func LongArraysToArray(array *LongArray) []int64 {
	if array == nil {
		return nil
	}
	return array.ToArray()
}
