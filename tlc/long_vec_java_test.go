/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
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
	"fmt"
	"strings"
	"testing"
)

// Java ArrayIndexOutOfBoundsException and StringIndexOutOfBoundsException are
// subclasses of IndexOutOfBoundsException. Do not accept arbitrary Go panics.
func javaLongVecCatch(t *testing.T, anyException bool, failure string, body func()) {
	t.Helper()
	defer func() {
		caught := recover()
		if caught == nil {
			t.Fatal(failure)
		}
		if anyException {
			if err, ok := caught.(error); ok {
				class := javaThrowableClassName(err)
				if class != fmt.Sprintf("%T", err) && strings.HasSuffix(class, "Exception") {
					return
				}
			}
		} else {
			switch caught.(type) {
			case *IndexOutOfBoundsException, *ArrayIndexOutOfBoundsException, *StringIndexOutOfBoundsException:
				return
			}
		}
		panic(caught)
	}()
	body()
}
func javaLongVecEquals(t *testing.T, want, got int64) {
	t.Helper()
	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

// Complete LongVecTest, including the subclass's inherited methods. The two
// explicit capacity-10 constructions retain the base constructor in both classes.
func javaLongVecMethods(t *testing.T, factory func() *LongVec) {
	t.Run("testReadBeyondCapacity", func(t *testing.T) {
		vec := factory()
		javaLongVecCatch(t, false, "Read beyond capacity", func() { vec.ElementAt(0) })
	})
	t.Run("testAddAndReadBeyondCapacity", func(t *testing.T) {
		vec := factory()
		vec.AddElement(1)
		javaLongVecEquals(t, 1, vec.ElementAt(0))
		javaLongVecCatch(t, false, "Read beyond capacity", func() { vec.ElementAt(1) })
	})
	t.Run("testRemoveBeyondCapacity", func(t *testing.T) {
		vec := NewLongVecWithCapacity(10)
		for i := -1; i <= 10; i++ {
			javaLongVecCatch(t, false, "Read beyond capacity", func() { vec.RemoveElement(0) })
		}
	})
	t.Run("testAddRemoveBeyondCapacity", func(t *testing.T) {
		vec := NewLongVecWithCapacity(10)
		vec.AddElement(1)
		vec.RemoveElement(0)
		javaLongVecCatch(t, false, "Read beyond capacity", func() { vec.RemoveElement(1) })
	})
	t.Run("testRemoveAndGet", func(t *testing.T) {
		vec := factory()
		vec.AddElement(1)
		vec.AddElement(2)
		vec.AddElement(3)
		javaLongVecEquals(t, 1, vec.ElementAt(0))
		javaLongVecEquals(t, 2, vec.ElementAt(1))
		javaLongVecEquals(t, 3, vec.ElementAt(2))
		vec.RemoveElement(1)
		javaLongVecEquals(t, 1, vec.ElementAt(0))
		javaLongVecEquals(t, 3, vec.ElementAt(1))
		javaLongVecCatch(t, false, "A new elements magically appeared in LongVec", func() { vec.ElementAt(2) })
	})
	t.Run("testRemoveWrongOrder", func(t *testing.T) {
		vec := factory()
		vec.AddElement(1)
		vec.AddElement(2)
		javaLongVecEquals(t, 1, vec.ElementAt(0))
		javaLongVecEquals(t, 2, vec.ElementAt(1))
		vec.RemoveElement(0)
		javaLongVecCatch(t, false, "Removed non-existing element", func() { vec.RemoveElement(1) })
	})
	t.Run("testGetNegative", func(t *testing.T) {
		vec := factory()
		javaLongVecCatch(t, false, "Read negative", func() { vec.ElementAt(-1) })
	})
	t.Run("testRemoveNegative", func(t *testing.T) {
		vec := factory()
		javaLongVecCatch(t, false, "Removed negative", func() { vec.RemoveElement(-1) })
	})
}
func TestJavaLongVec(t *testing.T) { javaLongVecMethods(t, NewLongVec) }
func TestJavaGrowingLongVec(t *testing.T) {
	javaLongVecMethods(t, func() *LongVec { return NewLongVecWithCapacity(0) })
	t.Run("testGrowAndShrink", func(t *testing.T) {
		vec := NewLongVecWithCapacity(0)
		vec.AddElement(1)
		vec.AddElement(2)
		javaLongVecEquals(t, 2, int64(vec.Size()))
		vec.RemoveElement(1)
		vec.RemoveElement(0)
		javaLongVecEquals(t, 0, int64(vec.Size()))
		javaLongVecCatch(t, true, "Removed non-existing element", func() { vec.RemoveElement(0) })
	})
}
