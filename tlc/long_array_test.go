package tlc

import (
	"reflect"
	"testing"
)

func TestLongArrayJavaUtilityBehavior(t *testing.T) {
	array := NewLongArray(5)
	if array.Size() != 5 {
		t.Fatalf("size = %d, want 5", array.Size())
	}
	if got := array.String(); got != "[e, e, e, e, e]" {
		t.Fatalf("zero string = %q, want Java empty-slot rendering", got)
	}

	values := []int64{5, 8, 1, 7, 3}
	for i, value := range values {
		array.Set(int64(i), value)
	}
	for i, value := range values {
		if got := array.Get(int64(i)); got != value {
			t.Fatalf("get[%d] = %d, want %d", i, got, value)
		}
	}

	if !array.TrySet(2, 1, -9) {
		t.Fatalf("TrySet with expected value failed")
	}
	if array.TrySet(2, 1, 99) {
		t.Fatalf("TrySet with stale expected value succeeded")
	}
	if got := array.Get(2); got != -9 {
		t.Fatalf("after TrySet get[2] = %d, want -9", got)
	}

	array.SwapCopy(0, 4)
	if got, want := array.ToArray(), []int64{3, 8, -9, 7, 5}; !reflect.DeepEqual(got, want) {
		t.Fatalf("after SwapCopy = %v, want %v", got, want)
	}

	LongArraysSort(array)
	if got, want := array.ToArray(), []int64{-9, 3, 5, 7, 8}; !reflect.DeepEqual(got, want) {
		t.Fatalf("after LongArraysSort = %v, want %v", got, want)
	}
}

func TestLongArrayNegativeCapacityFailsLikeJava(t *testing.T) {
	requirePanic(t, func() { NewLongArray(-1) })
}
