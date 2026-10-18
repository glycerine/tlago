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
	if positions<<3 < 0 {
		panic(NewIllegalArgumentException())
	}
	return &LongArray{data: make([]uint64, positions)}
}

func NewLongArrayFrom(values []int64) *LongArray {
	if values == nil {
		panic(NewNullPointerException())
	}
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
	if a == nil {
		panic(NewNullPointerException())
	}
	if len(numThreads) > 1 {
		panic(NewIllegalArgumentException("ZeroMemory accepts at most one thread count"))
	}
	threads := 1
	if len(numThreads) > 0 {
		threads = int(int32(numThreads[0]))
		if threads == 0 {
			panic(NewArithmeticException("/ by zero"))
		}
		if threads < 0 {
			panic(NewIllegalArgumentException())
		}
	}
	if len(a.data) == 0 {
		return nil
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
	// Source unsafe word accesses may overlap when one partition's iterator
	// skips wrapped entries while another marks them. Represent those native
	// word accesses atomically in Go, alongside TrySet's compare-and-swap.
	atomic.StoreUint64(&a.data[position], uint64(value))
}

func (a *LongArray) Get(position int64) int64 {
	a.rangeCheck(position)
	return int64(atomic.LoadUint64(&a.data[position]))
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
		panic(NewNullPointerException())
	}
	return int64(len(a.data))
}

func (a *LongArray) String() string {
	if a == nil {
		panic(NewNullPointerException())
	}
	return a.StringRange(0, a.Size()-1)
}

func (a *LongArray) StringRange(start int64, end int64) string {
	if a == nil {
		panic(NewNullPointerException())
	}
	if end == -1 {
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
		panic(NewNullPointerException())
	}
	out := make([]int64, len(a.data))
	for i := range a.data {
		out[i] = a.Get(int64(i))
	}
	return out
}

func (a *LongArray) rangeCheck(position int64) {
	if a == nil {
		panic(NewNullPointerException())
	}
	if position < 0 || position >= int64(len(a.data)) {
		panic(NewAssertionError())
	}
}

type LongComparator func(lo int64, loPos int64, hi int64, hiPos int64) int

func NewLongArraysLongComparator() LongComparator {
	return DefaultLongComparator
}

func (c LongComparator) Compare(lo int64, loPos int64, hi int64, hiPos int64) int {
	if c == nil {
		panic(NewNullPointerException())
	}
	return c(lo, loPos, hi, hiPos)
}

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
		panic(NewNullPointerException())
	}
	LongArraysSortRange(array, 0, array.Size()-1)
}

func LongArraysSortRange(array *LongArray, left int64, right int64, comparators ...LongComparator) {
	if len(comparators) > 1 {
		panic(NewIllegalArgumentException("LongArraysSortRange accepts at most one comparator"))
	}
	if array == nil {
		panic(NewNullPointerException())
	}
	if array.Size() == 0 {
		return
	}
	cmp := NewLongArraysLongComparator()
	if len(comparators) == 1 {
		cmp = comparators[0]
	}
	size := array.Size()
	for i, j := left, left; i < right; {
		lo := (i + 1) % size
		ai := array.Get(lo)
		for {
			hiPos := j % size
			hi := array.Get(hiPos)
			if cmp == nil {
				panic(NewNullPointerException())
			}
			if cmp(ai, lo, hi, hiPos) > -1 {
				break
			}
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
		panic(NewNullPointerException())
	}
	return array.ToArray()
}
