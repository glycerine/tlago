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

import "testing"

// The original four empty-queue catches accept NoSuchElementException only.
func javaIntQueueCatch(t *testing.T, body func()) {
	t.Helper()
	defer func() {
		caught := recover()
		if caught == nil {
			t.Fatal("Returned element where there should be none.")
		}
		if _, ok := caught.(*NoSuchElementException); !ok {
			panic(caught)
		}
	}()
	body()
}
func javaIntQueueEquals(t *testing.T, want, got int64) {
	t.Helper()
	if want != got {
		t.Fatalf("got %d, want %d", got, want)
	}
}

// Complete translation of tlc2.util.MemIntQueueTest, retaining the source zero
// long and exact ring-wrap/growth sequence without additional assertions.
func TestJavaMemIntQueue(t *testing.T) {
	t.Run("testDequeuePastLastElement", func(t *testing.T) {
		queue := NewIntQueueWithDisk("irrelevant", "irrelevant", intQueueInitialSize)
		queue.EnqueueInt(1)
		javaIntQueueEquals(t, 1, int64(queue.DequeueInt()))
		javaIntQueueCatch(t, func() { queue.DequeueInt() })
	})
	t.Run("testEnqueueZeros", func(t *testing.T) {
		queue := NewIntQueueWithDisk("irrelevant", "irrelevant", intQueueInitialSize)
		queue.EnqueueInt(0)
		queue.EnqueueInt(0)
		queue.EnqueueInt(0)
		javaIntQueueEquals(t, 0, int64(queue.DequeueInt()))
		javaIntQueueEquals(t, 0, int64(queue.DequeueInt()))
		javaIntQueueEquals(t, 0, int64(queue.DequeueInt()))
		javaIntQueueCatch(t, func() { queue.DequeueInt() })
	})
	t.Run("testEnqueueLong", func(t *testing.T) {
		queue := NewIntQueueWithDisk("irrelevant", "irrelevant", intQueueInitialSize)
		queue.EnqueueLong(0)
		javaIntQueueEquals(t, 0, int64(queue.DequeueInt()))
		javaIntQueueEquals(t, 0, int64(queue.DequeueInt()))
		javaIntQueueCatch(t, func() { queue.DequeueInt() })
	})
	t.Run("testEnqueueDequeueLong", func(t *testing.T) {
		queue := NewIntQueueWithDisk("irrelevant", "irrelevant", intQueueInitialSize)
		queue.EnqueueLong(0)
		javaIntQueueEquals(t, 0, queue.DequeueLong())
		javaIntQueueCatch(t, func() { queue.DequeueLong() })
	})
	t.Run("testGrow", func(t *testing.T) {
		queue := NewIntQueueWithDisk("irrelevant", "irrelevant", 4)
		queue.EnqueueInt(0)
		queue.EnqueueInt(1)
		queue.EnqueueInt(2)
		queue.EnqueueInt(3)
		javaIntQueueEquals(t, 4, int64(queue.Size()))
		queue.DequeueInt()
		queue.DequeueInt()
		queue.DequeueInt()
		queue.DequeueInt()
		javaIntQueueEquals(t, 0, int64(queue.Size()))
		queue.EnqueueInt(4)
		queue.EnqueueInt(5)
		queue.EnqueueInt(6)
		queue.EnqueueInt(7)
		javaIntQueueEquals(t, 4, int64(queue.Size()))
		queue.DequeueInt()
		queue.EnqueueInt(8)
		queue.EnqueueInt(9)
		for i := 5; i < 10; i++ {
			javaIntQueueEquals(t, int64(i), int64(queue.DequeueInt()))
		}
		javaIntQueueEquals(t, 0, int64(queue.Size()))
	})
}
