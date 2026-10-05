/*******************************************************************************
 * Copyright (c) 2016 Microsoft Research. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
package tlc

import (
	"math"
	"strconv"
	"testing"
)

func assumeJavaLongArray64Bit(t *testing.T) {
	t.Helper()
	if strconv.IntSize != 64 {
		t.Skip("source Assume: architecture is BIT_64")
	}
}
func requireJavaLongArrayAssertion(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		failure := recover()
		if _, ok := failure.(*AssertionError); !ok {
			t.Fatalf("expected AssertionError, got %T: %v", failure, failure)
		}
	}()
	f()
}
func TestJavaLongArrayGetAndSet(t *testing.T) {
	assumeJavaLongArray64Bit(t)
	const elements = 100
	array := NewLongArray(elements)
	if err := array.ZeroMemory(); err != nil {
		t.Fatal(err)
	}
	for i := int64(0); i < elements; i++ {
		if got := array.Get(i); got != 0 {
			t.Fatalf("get(%d)=%d, want0", i, got)
		}
	}
	for i := int64(0); i < elements; i++ {
		array.Set(i, i)
	}
	for i := int64(0); i < elements; i++ {
		if got := array.Get(i); got != i {
			t.Fatalf("get(%d)=%d, want%d", i, got, i)
		}
	}
	for i := int64(0); i < elements; i++ {
		array.Set(i, math.MaxInt64-i)
	}
	for i := int64(0); i < elements; i++ {
		if got := array.Get(i); got != math.MaxInt64-i {
			t.Fatalf("get(%d)=%d, want%d", i, got, int64(math.MaxInt64)-i)
		}
	}
	for i := int64(0); i < elements; i++ {
		array.Set(i, math.MinInt64+i)
	}
	for i := int64(0); i < elements; i++ {
		if got := array.Get(i); got != math.MinInt64+i {
			t.Fatalf("get(%d)=%d, want%d", i, got, int64(math.MinInt64)+i)
		}
	}
}
func TestJavaLongArrayOutOfRangePositive(t *testing.T) {
	assumeJavaLongArray64Bit(t)
	array := NewLongArray(1)
	requireJavaLongArrayAssertion(t, func() { array.Get(1) })
}
func TestJavaLongArrayOutOfRangeNegative(t *testing.T) {
	assumeJavaLongArray64Bit(t)
	array := NewLongArray(1)
	requireJavaLongArrayAssertion(t, func() { array.Get(-1) })
}
func TestJavaLongArrayGetAndTrySet(t *testing.T) {
	assumeJavaLongArray64Bit(t)
	const elements = 100
	array := NewLongArray(elements)
	if err := array.ZeroMemory(); err != nil {
		t.Fatal(err)
	}
	for i := int64(0); i < elements; i++ {
		if got := array.Get(i); got != 0 {
			t.Fatalf("get(%d)=%d, want0", i, got)
		}
	}
	for i := int64(0); i < elements; i++ {
		if !array.TrySet(i, 0, i) {
			t.Fatalf("trySet(%d) false", i)
		}
	}
	for i := int64(0); i < elements; i++ {
		if got := array.Get(i); got != i {
			t.Fatalf("get(%d)=%d, want%d", i, got, i)
		}
	}
	for i := int64(0); i < elements; i++ {
		array.TrySet(i, i, math.MaxInt64-i)
	}
	for i := int64(0); i < elements; i++ {
		if got := array.Get(i); got != math.MaxInt64-i {
			t.Fatalf("get(%d)=%d, want%d", i, got, int64(math.MaxInt64)-i)
		}
	}
	for i := int64(0); i < elements; i++ {
		array.TrySet(i, math.MaxInt64-i, math.MinInt64+i)
	}
	for i := int64(0); i < elements; i++ {
		if got := array.Get(i); got != math.MinInt64+i {
			t.Fatalf("get(%d)=%d, want%d", i, got, int64(math.MinInt64)+i)
		}
	}
}
func TestJavaLongArrayZeroMemory(t *testing.T) {
	assumeJavaLongArray64Bit(t)
	for k := 1; k < 8; k++ {
		for i := 1; i < 128; i++ {
			array := NewLongArray(int64(i))
			if err := array.ZeroMemory(k); err != nil {
				t.Fatal(err)
			}
			// Preserve both original source loops literally, including their i<j condition and i++ increment.
			for j := 0; i < j; i++ {
				if got := array.Get(int64(j)); got != 0 {
					t.Fatalf("get(%d)=%d, want0", j, got)
				}
			}
			for j := 0; i < j; i++ {
				array.Set(int64(j), -1)
			}
		}
	}
}
func TestJavaLongArraySwap(t *testing.T) {
	assumeJavaLongArray64Bit(t)
	const elements = 10321
	array := NewLongArray(elements)
	if err := array.ZeroMemory(); err != nil {
		t.Fatal(err)
	}
	for i := int64(0); i < elements; i++ {
		value := int64(math.MaxInt64) - i
		array.Set(i, value)
	}
	for i := 0; i < elements/2; i++ {
		array.SwapCopy(int64(i), int64(elements-1-i))
	}
	for i := int64(0); i < elements; i++ {
		if got := array.Get(i); got != math.MaxInt64-(elements-1)+i {
			t.Fatalf("get(%d)=%d, want%d", i, got, int64(math.MaxInt64)-(elements-1)+i)
		}
	}
}
func TestJavaLongArraySwapRandom(t *testing.T) {
	assumeJavaLongArray64Bit(t)
	const elements = 21383
	vals := make([]int64, 0, elements)
	rnd := NewJavaRandomDefault()
	for i := 0; i < elements; i++ {
		vals = append(vals, rnd.NextLong())
	}
	array := NewLongArray(elements)
	if err := array.ZeroMemory(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < elements; i++ {
		array.Set(int64(i), vals[i])
	}
	for i := 0; i < elements/2; i++ {
		array.SwapCopy(int64(i), int64(elements-1-i))
	}
	for lo, hi := 0, len(vals)-1; lo < hi; lo, hi = lo+1, hi-1 {
		vals[lo], vals[hi] = vals[hi], vals[lo]
	}
	for i := 0; i < elements; i++ {
		if got := array.Get(int64(i)); got != vals[i] {
			t.Fatalf("get(%d)=%d, want%d", i, got, vals[i])
		}
	}
}
