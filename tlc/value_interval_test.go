package tlc

import (
	"math"
	"testing"
)

func TestIntervalValueSizeUsesWideArithmeticLikeJava(t *testing.T) {
	cases := []struct {
		low  int32
		high int32
		want int
	}{
		{1, math.MaxInt32, math.MaxInt32},
		{math.MinInt32, -2, math.MaxInt32},
		{math.MaxInt32, math.MaxInt32, 1},
		{math.MinInt32, math.MinInt32, 1},
		{math.MaxInt32 - 10, math.MaxInt32, 11},
		{math.MinInt32, math.MinInt32 + 10, 11},
		{10, 1, 0},
	}
	for _, tc := range cases {
		got, err := NewIntervalValue(tc.low, tc.high).Size()
		if err != nil {
			t.Fatalf("%d..%d Size returned error: %v", tc.low, tc.high, err)
		}
		if got != tc.want {
			t.Fatalf("%d..%d Size = %d, want %d", tc.low, tc.high, got, tc.want)
		}
	}
}

func TestIntervalValueSizeReportsOverflowLikeJava(t *testing.T) {
	if size, err := NewIntervalValue(math.MinInt32, math.MaxInt32).Size(); err == nil {
		t.Fatalf("maximum interval Size = %d with nil error, want overflow error", size)
	}
	if size, err := NewIntervalValue(-989_822_976, 1_157_660_672).Size(); err == nil {
		t.Fatalf("large interval Size = %d with nil error, want overflow error", size)
	}
}

func TestIntervalValueEmptyIntervalsCompareAndEqual(t *testing.T) {
	i1 := NewIntervalValue(10, 1)
	i2 := NewIntervalValue(20, 15)
	eq, err := i1.Equal(i2)
	if err != nil {
		t.Fatalf("empty interval Equal returned error: %v", err)
	}
	if !eq {
		t.Fatalf("empty intervals should compare equal")
	}
	cmp, err := i1.Compare(i2)
	if err != nil {
		t.Fatalf("empty interval Compare returned error: %v", err)
	}
	if cmp != 0 {
		t.Fatalf("empty interval Compare = %d, want 0", cmp)
	}
}

func TestIntervalValueMaxIntSingletonEnumeratesAndFingerprintsAsSet(t *testing.T) {
	iv := NewIntervalValue(math.MaxInt32, math.MaxInt32)
	enum := iv.Elements()
	first := enum.NextElement()
	if !mustValueEqual(first, NewIntValue(math.MaxInt32)) {
		t.Fatalf("first singleton element = %v, want MaxInt32", first)
	}
	if next := enum.NextElement(); next != nil {
		t.Fatalf("second singleton element = %v, want nil", next)
	}
	enum.Reset()
	if !mustValueEqual(enum.NextElement(), NewIntValue(math.MaxInt32)) {
		t.Fatalf("reset singleton element did not return MaxInt32")
	}

	singleton := NewSetEnumValue([]Value{NewIntValue(math.MaxInt32)}, true)
	if iv.FingerPrint(FP64New()) != singleton.FingerPrint(FP64New()) {
		t.Fatalf("singleton interval fingerprint differs from equivalent set")
	}
}
