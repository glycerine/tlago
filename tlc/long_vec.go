// Copyright (c) 2003 Compaq Corporation.  All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation.  All rights reserved.
// Last modified on Mon 30 Apr 2007 at 13:26:33 PST by lamport
//
//	modified on Mon Dec  4 16:20:19 PST 2000 by yuanyu
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
		panic(NewNegativeArraySizeException(strconv.Itoa(capacity)))
	}
	return &LongVec{data: make([]int64, 0, capacity)}
}

func NewLongVecFrom(values []int64) *LongVec {
	data := make([]int64, len(values))
	copy(data, values)
	return &LongVec{data: data}
}

func (v *LongVec) Add(x int64) {
	if len(v.data) == cap(v.data) {
		v.ensureCapacity(len(v.data) + 1)
	}
	v.data = append(v.data, x)
}

// Java LongVec doubles its backing array, including growth from capacity zero.
func (v *LongVec) ensureCapacity(minCapacity int) {
	if cap(v.data) < minCapacity {
		newCapacity := int(int32(cap(v.data)) + int32(cap(v.data)))
		if newCapacity < minCapacity {
			newCapacity = minCapacity
		}
		data := make([]int64, len(v.data), newCapacity)
		copy(data, v.data)
		v.data = data
	}
}

func (v *LongVec) AddElement(x int64) {
	v.Add(x)
}

func (v *LongVec) At(index int) int64 {
	v.rangeCheck(index)
	return v.arrayAt(index)
}

func (v *LongVec) ElementAt(index int) int64 {
	return v.At(index)
}

func (v *LongVec) Last() int64 {
	return v.arrayAt(len(v.data) - 1)
}

func (v *LongVec) LastElement() int64 {
	return v.Last()
}

func (v *LongVec) Remove(index int) {
	v.rangeCheck(index)
	// Array assignment evaluates its right-hand side before checking the
	// destination index, as in the original Java expression.
	last := v.arrayAt(len(v.data) - 1)
	if index < 0 || index >= cap(v.data) {
		panic(NewArrayIndexOutOfBoundsException(index, cap(v.data)))
	}
	v.data[:cap(v.data)][index] = last
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
	filtered := NewLongVecWithCapacity(v.Size())
	for _, x := range v.data {
		if filtered.IsEmpty() || filtered.LastElement() != x {
			filtered.AddElement(x)
		}
	}
	v.data = filtered.data
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
	if index >= len(v.data) {
		message := "Index: " + strconv.Itoa(index) + ", Size: " + strconv.Itoa(len(v.data))
		panic(&IndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(javaString(message), nil)})
	}
}

// Negative indices bypass rangeCheck in Java and fail on the backing array.
func (v *LongVec) arrayAt(index int) int64 {
	if index < 0 || index >= cap(v.data) {
		panic(NewArrayIndexOutOfBoundsException(index, cap(v.data)))
	}
	return v.data[:cap(v.data)][index]
}
