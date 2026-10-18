package tlc

import (
	"math"
	"testing"
)

// Whole original OffHeapIteratorTest.testNext and setup assumption.
func TestJavaOffHeapIteratorNext(t *testing.T) {
	assumeJavaLongArraysSupported(t)
	const elements int64 = 32
	array := NewLongArray(2 * elements)
	for i := int64(0); i < array.Size(); i++ {
		array.Set(i, i+1)
	}
	iterator := NewOffHeapIterator(array, elements, NewInfinitePrecisionOffHeapIndexer(2*elements, 1))
	actual := int64(1)
	for iterator.HasNext() {
		next := iterator.Next()
		if next != actual {
			t.Fatalf("next=%d, want %d", next, actual)
		}
		actual++
	}
	if actual-1 != elements {
		t.Fatalf("count=%d, want %d", actual-1, elements)
	}
	for i := int64(0); i < array.Size(); i++ {
		if got := array.Get(i); got != i+1 {
			t.Fatalf("array[%d]=%d, want %d", i, got, i+1)
		}
	}
}

// Whole original OffHeapIteratorTest.testMarkNext. Preserve the source's repeated
// array.get(elements) assertion in the upper-half loop, rather than substituting
// array.get(i) or inventing a stronger/different test.
func TestJavaOffHeapIteratorMarkNext(t *testing.T) {
	assumeJavaLongArraysSupported(t)
	const elements int64 = 32
	array := NewLongArray(2 * elements)
	for i := int64(0); i < array.Size(); i++ {
		array.Set(i, i+1)
	}
	iterator := NewOffHeapIterator(array, elements, NewInfinitePrecisionOffHeapIndexer(2*elements, 1))
	actual := int64(1)
	for iterator.HasNext() {
		next := iterator.MarkNext()
		if next != actual {
			t.Fatalf("markNext=%d, want %d", next, actual)
		}
		actual++
	}
	if actual-1 != elements {
		t.Fatalf("count=%d, want %d", actual-1, elements)
	}
	for i := elements; i < array.Size(); i++ {
		if got := array.Get(elements); got != elements+1 {
			t.Fatalf("array[%d]=%d, want %d", elements, got, elements+1)
		}
	}
	for i := int64(0); i < elements; i++ {
		expected := (i + 1) | math.MinInt64
		actual = array.Get(i)
		if actual != expected {
			t.Fatalf("marked array[%d]=%d, want %d", i, actual, expected)
		}
	}
}
