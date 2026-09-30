package tlc

import (
	"math"
	"testing"
)

func TestFcnRcdValueSelectEmptyAndNormalizedEmpty(t *testing.T) {
	empty := NewFcnRcdValue(nil, nil, false)
	if got, err := empty.Select(IntNegOne); err != nil || got != nil {
		t.Fatalf("empty Select = %v/%v, want nil/nil", got, err)
	}
	empty.Normalize()
	if got, err := empty.Select(IntNegOne); err != nil || got != nil {
		t.Fatalf("normalized empty Select = %v/%v, want nil/nil", got, err)
	}
}

func TestFcnRcdValueEmptyIntervalDomainToTuple(t *testing.T) {
	emptyIntervalFcn := NewFcnRcdIntervalValue(NewIntervalValue(2, 1), nil)
	tuple := emptyIntervalFcn.ToTuple()
	if tuple == nil || len(tuple.Elems) != 0 {
		t.Fatalf("empty interval function ToTuple = %v, want empty tuple", tuple)
	}
	nonTuple := NewFcnRcdIntervalValue(NewIntervalValue(2, 2), []Value{NewIntValue(42)})
	if tuple := nonTuple.ToTuple(); tuple != nil {
		t.Fatalf("2..2 interval function ToTuple = %v, want nil", tuple)
	}
}

func TestFcnRcdValueSelectExplicitDomainBeforeAndAfterNormalize(t *testing.T) {
	for _, normalize := range []bool{false, true} {
		domain := []Value{NewIntValue(-2), NewIntValue(-1), NewIntValue(0), NewIntValue(1), NewIntValue(2)}
		values := []Value{NewIntValue(1022), NewIntValue(1023), NewIntValue(1024), NewIntValue(1025), NewIntValue(1026)}
		fcn := NewFcnRcdValue(domain, values, false)
		if normalize {
			fcn.Normalize()
		}
		for j := int32(-4); j <= 4; j++ {
			got, err := fcn.Select(NewIntValue(j))
			if err != nil {
				t.Fatalf("Select(%d, normalize=%v) returned error: %v", j, normalize, err)
			}
			if j < -2 || j > 2 {
				if got != nil {
					t.Fatalf("Select(%d, normalize=%v) = %v, want nil", j, normalize, got)
				}
				continue
			}
			if !mustValueEqual(got, NewIntValue(j+1024)) {
				t.Fatalf("Select(%d, normalize=%v) = %v, want %d", j, normalize, got, j+1024)
			}
		}
	}
}

func TestFcnRcdValueMalformedIntervalDoesNotWrap(t *testing.T) {
	zero := IntZero
	intervalFcn := NewFcnRcdIntervalValue(NewIntervalValue(math.MaxInt32, math.MaxInt32), []Value{zero, zero})
	explicitFcn := NewFcnRcdValue(
		[]Value{NewIntValue(math.MaxInt32), NewIntValue(math.MinInt32)},
		[]Value{zero, zero},
		true)
	eq, err := explicitFcn.Equal(intervalFcn)
	if err != nil {
		t.Fatalf("Equal returned error: %v", err)
	}
	if eq {
		t.Fatalf("malformed explicit function should not equal malformed interval function")
	}

	malformed := NewFcnRcdIntervalValue(NewIntervalValue(math.MinInt32, math.MaxInt32), []Value{zero})
	selected, err := malformed.Select(NewIntValue(math.MaxInt32))
	if err != nil {
		t.Fatalf("malformed Select returned error: %v", err)
	}
	if selected != nil {
		t.Fatalf("malformed Select(MaxInt32) = %v, want nil", selected)
	}
	excepted, err := malformed.TakeExcept(ValueExcept{Path: []Value{NewIntValue(math.MaxInt32)}, Value: IntOne})
	if err != nil {
		t.Fatalf("malformed TakeExcept returned error: %v", err)
	}
	if excepted != malformed {
		t.Fatalf("malformed TakeExcept returned %p, want original %p", excepted, malformed)
	}
}

func TestFcnRcdValueEmptyIntervalCompareAndNormalize(t *testing.T) {
	a := NewFcnRcdIntervalValue(NewIntervalValue(2, 1), nil)
	b := NewFcnRcdIntervalValue(NewIntervalValue(3, 2), nil)
	eq, err := a.Equal(b)
	if err != nil {
		t.Fatalf("empty interval function Equal returned error: %v", err)
	}
	if !eq {
		t.Fatalf("empty interval functions should be equal")
	}
	cmp, err := a.Compare(b)
	if err != nil {
		t.Fatalf("empty interval function Compare returned error: %v", err)
	}
	if cmp != 0 {
		t.Fatalf("empty interval function Compare = %d, want 0", cmp)
	}

	emptyTupleFcn := asFcnRcdValue(EmptyTuple)
	eq, err = a.Equal(emptyTupleFcn)
	if err != nil {
		t.Fatalf("empty interval vs empty tuple Equal returned error: %v", err)
	}
	if !eq {
		t.Fatalf("empty interval function should equal empty tuple function")
	}
	cmp, err = emptyTupleFcn.Compare(a)
	if err != nil {
		t.Fatalf("empty tuple vs empty interval Compare returned error: %v", err)
	}
	if cmp != 0 {
		t.Fatalf("empty tuple vs empty interval Compare = %d, want 0", cmp)
	}

	set := NewSetEnumValue([]Value{a, b}, false)
	size, err := set.Size()
	if err != nil {
		t.Fatalf("set size returned error: %v", err)
	}
	if size != 1 {
		t.Fatalf("set of equivalent empty interval functions size = %d, want 1", size)
	}
}
