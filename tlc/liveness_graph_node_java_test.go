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

// Original GraphNodeTest.testAllocateRealign.
func TestJavaGraphNodeAllocateRealign(t *testing.T) {
	node := NewGraphNode(0, 0)
	sizeHint := 5
	node.AddTransition(1, -1, -1, -1, nil, 0, sizeHint)
	sizeHint--
	node.AddTransition(2, -1, -1, -1, nil, 0, sizeHint)
	sizeHint--
	node.AddTransition(3, -1, -1, -1, nil, 0, sizeHint)
	sizeHint--
	node.AddTransition(4, -1, -1, -1, nil, 0, sizeHint)
	sizeHint--
	node.AddTransition(5, -1, -1, -1, nil, 0, sizeHint)
	sizeHint--
	overallocated := node.Realign()
	if overallocated != 0 {
		t.Fatal("Allocation overallocated")
	}
	if !node.TransExists(1, -1) {
		t.Fatal("Lost a transition during the allocation business")
	}
	if !node.TransExists(2, -1) {
		t.Fatal("Lost a transition during the allocation business")
	}
	if !node.TransExists(3, -1) {
		t.Fatal("Lost a transition during the allocation business")
	}
	if !node.TransExists(4, -1) {
		t.Fatal("Lost a transition during the allocation business")
	}
	if !node.TransExists(5, -1) {
		t.Fatal("Lost a transition during the allocation business")
	}
}

// Original GraphNodeTest.testRealign.
func TestJavaGraphNodeRealign(t *testing.T) {
	node := NewGraphNode(0, 0)
	node.AddTransition(1, -1, -1, -1, nil, 0, 64)
	overallocated := node.Realign()
	if overallocated != 63 {
		t.Fatal("Allocation overallocated")
	}
	overallocated = node.Realign()
	if overallocated != 0 {
		t.Fatal("Allocation overallocated")
	}
	if !node.TransExists(1, -1) {
		t.Fatal("Lost a transition during the allocation business")
	}
}

// Original GraphNodeTest.testAllocateNested.
func TestJavaGraphNodeAllocateNested(t *testing.T) {
	node := NewGraphNode(0, 0)
	cnt := 0
	for i := 0; i < 5; i++ {
		for j := 0; j < 10; j++ {
			for k := 0; k < 15; k++ {
				l := 5 * 10 * 15
				fp, hint := cnt, l-cnt
				cnt++
				node.AddTransition(uint64(fp), -1, -1, -1, nil, 0, hint)
			}
		}
	}
	overallocated := node.Realign()
	if overallocated != 0 {
		t.Fatal("Nested allocation overallocated")
	}
	for i := 0; i < cnt; i++ {
		if !node.TransExists(uint64(i), -1) {
			t.Fatal("Lost a transition during this allocation business")
		}
	}
}

// Original GraphNodeTest.testAllocateNestedRandom; Java Random seed4711.
func TestJavaGraphNodeAllocateNestedRandom(t *testing.T) {
	node := NewGraphNode(0, 0)
	rnd := NewJavaRandom(4711)
	verificationSet := []int{}
	cnt := 0
	for i := 0; i < 5; i++ {
		x := int(rnd.NextIntN(10))
		for j := 0; j < x; j++ {
			y := int(rnd.NextIntN(15))
			for k := 0; k < y; k++ {
				l := 5 * x * y
				allocationHint := l - cnt
				cnt++
				node.AddTransition(uint64(cnt), -1, -1, -1, nil, 0, allocationHint)
				verificationSet = append(verificationSet, cnt)
			}
		}
	}
	for _, i := range verificationSet {
		if !node.TransExists(uint64(i), -1) {
			t.Fatal("Lost a transition during this allocation business")
		}
	}
}

// Original GraphNodeTest.testAllocateNegative.
func TestJavaGraphNodeAllocateNegative(t *testing.T) {
	node := NewGraphNode(0, 0)
	node.AddTransition(0, 0, 0, 0, nil, 0, -1)
	if node.Realign() != 0 {
		t.Fatal("overallocated")
	}
}

// Original GraphNodeTest.testAllocateAndSuccessorSize.
func TestJavaGraphNodeAllocateAndSuccessorSize(t *testing.T) {
	node := NewGraphNode(0, 0)
	node.AddTransition(0, 0, 0, 0, nil, 0, 100)
	if node.SuccSize() != 1 {
		t.Fatalf("succSize=%d, want1", node.SuccSize())
	}
}
