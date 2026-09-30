package tlc

import (
	"strconv"
	"strings"
)

type LongVec struct {
	data []int64
}

func NewLongVec() *LongVec {
	return NewLongVecWithCapacity(10)
}

func NewLongVecWithCapacity(capacity int) *LongVec {
	if capacity < 0 {
		capacity = 0
	}
	return &LongVec{data: make([]int64, 0, capacity)}
}

func NewLongVecFrom(values []int64) *LongVec {
	data := make([]int64, len(values))
	copy(data, values)
	return &LongVec{data: data}
}

func (v *LongVec) Add(x int64) {
	v.data = append(v.data, x)
}

func (v *LongVec) AddElement(x int64) {
	v.Add(x)
}

func (v *LongVec) At(index int) int64 {
	v.rangeCheck(index)
	return v.data[index]
}

func (v *LongVec) ElementAt(index int) int64 {
	return v.At(index)
}

func (v *LongVec) Last() int64 {
	return v.data[len(v.data)-1]
}

func (v *LongVec) LastElement() int64 {
	return v.Last()
}

func (v *LongVec) Remove(index int) {
	v.rangeCheck(index)
	v.data[index] = v.data[len(v.data)-1]
	v.data = v.data[:len(v.data)-1]
}

func (v *LongVec) RemoveElement(index int) {
	v.Remove(index)
}

func (v *LongVec) IsEmpty() bool {
	return len(v.data) == 0
}

func (v *LongVec) Size() int {
	return len(v.data)
}

func (v *LongVec) Reset() {
	v.data = v.data[:0]
}

func (v *LongVec) Reverse() *LongVec {
	out := NewLongVecFrom(v.data)
	for left, right := 0, len(out.data)-1; left < right; left, right = left+1, right-1 {
		out.data[left], out.data[right] = out.data[right], out.data[left]
	}
	return out
}

func (v *LongVec) Pack() *LongVec {
	if len(v.data) <= 1 {
		return v
	}
	filtered := v.data[:0]
	for _, x := range v.data {
		if len(filtered) == 0 || filtered[len(filtered)-1] != x {
			filtered = append(filtered, x)
		}
	}
	v.data = filtered
	return v
}

func (v *LongVec) RemoveLastIf(x int64) *LongVec {
	if len(v.data) > 0 && v.data[len(v.data)-1] == x {
		v.data = v.data[:len(v.data)-1]
	}
	return v
}

func (v *LongVec) ToSlice() []int64 {
	out := make([]int64, len(v.data))
	copy(out, v.data)
	return out
}

func (v *LongVec) String() string {
	var b strings.Builder
	b.WriteByte('<')
	for i, x := range v.data {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.FormatInt(x, 10))
	}
	b.WriteByte('>')
	return b.String()
}

func (v *LongVec) rangeCheck(index int) {
	if index < 0 || index >= len(v.data) {
		panic("LongVec index out of bounds")
	}
}
