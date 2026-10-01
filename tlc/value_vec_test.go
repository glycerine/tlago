package tlc

import (
	"strings"
	"testing"
)

func TestValueVecAddAtMatchesJavaDirectSlotFill(t *testing.T) {
	vec := NewValueVec(3)
	vec.Add(IntOne)
	vec.Add(NewIntValue(3))

	two := NewIntValue(2)
	vec.AddAt(two, 1)

	if got := vec.Len(); got != 3 {
		t.Fatalf("Len = %d, want 3", got)
	}
	if got := vec.At(0); got != IntOne {
		t.Fatalf("At(0) = %v, want 1", got)
	}
	if got := vec.At(1); got != two {
		t.Fatalf("At(1) = %v, want overwritten direct slot 2", got)
	}
	if got := vec.At(2); got != nil {
		t.Fatalf("At(2) = %v, want nil because Java addElementAt does not shift", got)
	}
}

func TestValueVecAddSortedUniqueKeepsJavaInsertionBehavior(t *testing.T) {
	vec := NewValueVec(0)
	if err := vec.AddSortedUnique(NewIntValue(3)); err != nil {
		t.Fatalf("AddSortedUnique(3): %v", err)
	}
	if err := vec.AddSortedUnique(IntOne); err != nil {
		t.Fatalf("AddSortedUnique(1): %v", err)
	}
	if err := vec.AddSortedUnique(NewIntValue(3)); err != nil {
		t.Fatalf("AddSortedUnique(duplicate 3): %v", err)
	}

	if got := vec.String(); got != "{1, 3}" {
		t.Fatalf("String = %q, want sorted unique Java order", got)
	}
}

func TestValueVecGrowthHonorsJavaSetBound(t *testing.T) {
	oldBound := Globals.SetBound
	Globals.SetBound = 2
	t.Cleanup(func() { Globals.SetBound = oldBound })

	vec := NewValueVec(2)
	vec.Add(IntZero)
	vec.Add(IntOne)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Add beyond set bound did not panic")
		} else if err, ok := r.(error); !ok || !strings.Contains(err.Error(), "Attempted to construct a set with too many elements (>2).") {
			t.Fatalf("panic = %v, want Java set-bound message", r)
		}
	}()
	vec.Add(NewIntValue(2))
}
