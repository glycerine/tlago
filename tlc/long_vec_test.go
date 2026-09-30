package tlc

import "testing"

func TestLongVecReadAndRemoveBoundsMatchJava(t *testing.T) {
	vec := NewLongVec()
	expectPanic(t, func() { vec.At(0) })
	expectPanic(t, func() { vec.At(-1) })
	expectPanic(t, func() { vec.Remove(0) })
	expectPanic(t, func() { vec.Remove(-1) })

	vec.Add(1)
	if got := vec.At(0); got != 1 {
		t.Fatalf("At(0) = %d, want 1", got)
	}
	expectPanic(t, func() { vec.At(1) })

	vec.Remove(0)
	expectPanic(t, func() { vec.Remove(0) })
}

func TestLongVecRemoveUsesLastElementAndShrinks(t *testing.T) {
	vec := NewLongVec()
	vec.Add(1)
	vec.Add(2)
	vec.Add(3)

	vec.Remove(1)
	if vec.Size() != 2 {
		t.Fatalf("size after Remove = %d, want 2", vec.Size())
	}
	if vec.At(0) != 1 || vec.At(1) != 3 {
		t.Fatalf("contents after Remove = %s, want <1, 3>", vec)
	}
	expectPanic(t, func() { vec.At(2) })
}

func TestLongVecGrowsFromZeroCapacity(t *testing.T) {
	vec := NewLongVecWithCapacity(0)
	vec.Add(1)
	vec.Add(2)

	if vec.Size() != 2 {
		t.Fatalf("size = %d, want 2", vec.Size())
	}
	vec.Remove(1)
	vec.Remove(0)
	if vec.Size() != 0 {
		t.Fatalf("size after removing all = %d, want 0", vec.Size())
	}
	expectPanic(t, func() { vec.Remove(0) })
}

func TestLongVecReversePackAndRemoveLastIf(t *testing.T) {
	vec := NewLongVecFrom([]int64{1, 2, 2, 1, 1, 3})

	reversed := vec.Reverse()
	if reversed.String() != "<3, 1, 1, 2, 2, 1>" {
		t.Fatalf("Reverse = %s", reversed)
	}
	if vec.String() != "<1, 2, 2, 1, 1, 3>" {
		t.Fatalf("Reverse mutated original: %s", vec)
	}

	vec.Pack()
	if vec.String() != "<1, 2, 1, 3>" {
		t.Fatalf("Pack = %s, want <1, 2, 1, 3>", vec)
	}
	vec.RemoveLastIf(4)
	if vec.String() != "<1, 2, 1, 3>" {
		t.Fatalf("RemoveLastIf absent mutated vector: %s", vec)
	}
	vec.RemoveLastIf(3)
	if vec.String() != "<1, 2, 1>" {
		t.Fatalf("RemoveLastIf present = %s, want <1, 2, 1>", vec)
	}
}
