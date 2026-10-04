// Complete mechanical translation of tlc2.value.impl.IntervalValueTest.
package tlc

import (
	"math"
	"strings"
	"testing"
)

func javaIntervalSize(t *testing.T, v Value) int {
	t.Helper()
	size, err := v.Size()
	if err != nil {
		t.Fatal(err)
	}
	return size
}
func javaIntervalCompare(t *testing.T, a, b Value) int {
	t.Helper()
	cmp, err := a.Compare(b)
	if err != nil {
		t.Fatal(err)
	}
	return cmp
}
func javaIntervalEqual(t *testing.T, a, b Value) {
	t.Helper()
	if a == nil || b == nil {
		t.Fatalf("null operand: %T, %T", a, b)
	}
	if eq, err := a.Equal(b); err != nil || !eq {
		t.Fatalf("equality=%v/%v, want true", eq, err)
	}
}
func javaIntervalSizeOverflow(t *testing.T, iv *IntervalValue, assertZero bool) {
	t.Helper()
	size, err := iv.Size()
	// sizeOverflow asserts zero if size() returns normally, then fails because
	// no exception was thrown. Preserve that original assertion as well.
	if err == nil && assertZero && size != 0 {
		t.Fatalf("size=%d, want 0", size)
	}
	failure := javaRuntimeException(err)
	if failure == nil {
		t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
	}
	if !strings.Contains(failure.Error(), "Size of interval value exceeds the maximum representable size (32bits)") {
		t.Fatalf("unexpected exception message: %v", failure)
	}
}

func TestJavaIntervalValue(t *testing.T) {
	oldPoly := FP64IrredPoly()
	FP64Init()
	t.Cleanup(func() { FP64InitPoly(oldPoly) })
	t.Run("testElementAt", func(t *testing.T) {
		iv := NewIntervalValue(3, 11)
		for i := 0; i < javaIntervalSize(t, iv); i++ {
			val, err := iv.ElementAt(i)
			if err != nil {
				t.Fatal(err)
			}
			javaIntervalEqual(t, NewIntValue(int32(i+3)), val)
		}
	})
	t.Run("testElementAtOutOfBoundsNegative", func(t *testing.T) {
		iv := NewIntervalValue(3, 11)
		_, err := iv.ElementAt(-1)
		if javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("testElementAtOutOfBoundsSize", func(t *testing.T) {
		iv := NewIntervalValue(3, 11)
		_, err := iv.ElementAt(javaIntervalSize(t, iv))
		if javaRuntimeException(err) == nil {
			t.Fatalf("expected TLCRuntimeException, got %T: %v", err, err)
		}
	})
	t.Run("sizeOverflow", func(t *testing.T) {
		if size := javaIntervalSize(t, NewIntervalValue(1, math.MaxInt32)); size != math.MaxInt32 {
			t.Fatalf("size=%d", size)
		}
		if size := javaIntervalSize(t, NewIntervalValue(math.MinInt32, -2)); size != math.MaxInt32 {
			t.Fatalf("size=%d", size)
		}
		javaIntervalSizeOverflow(t, NewIntervalValue(-989822976, 1157660672), true)
	})
	t.Run("compareToOverflow1", func(t *testing.T) {
		iv := NewIntervalValue(math.MaxInt32-1, math.MaxInt32)
		if size := javaIntervalSize(t, iv); size != 2 {
			t.Fatalf("size=%d", size)
		}
		iv2 := NewIntervalValue(math.MinInt32+1, math.MinInt32+2)
		if size := javaIntervalSize(t, iv2); size != 2 {
			t.Fatalf("size=%d", size)
		}
		if cmp := javaIntervalCompare(t, iv, iv2); cmp != 1 {
			t.Fatalf("compare=%d, want 1", cmp)
		}
	})
	t.Run("testCompareExtremeIntervals", func(t *testing.T) {
		x, y := NewIntervalValue(math.MinInt32, -2), NewIntervalValue(1, math.MaxInt32)
		if xs, ys := javaIntervalSize(t, x), javaIntervalSize(t, y); xs != ys {
			t.Fatalf("sizes=%d,%d", xs, ys)
		}
		if cmp := javaIntervalCompare(t, x, y); cmp >= 0 {
			t.Fatalf("compare(x,y)=%d, want negative", cmp)
		}
		if cmp := javaIntervalCompare(t, y, x); cmp <= 0 {
			t.Fatalf("compare(y,x)=%d, want positive", cmp)
		}
	})
	t.Run("testEmptyIntervalEquality", func(t *testing.T) {
		javaIntervalEqual(t, NewIntervalValue(10, 1), NewIntervalValue(20, 15))
	})
	t.Run("testCompareEmptyIntervals", func(t *testing.T) {
		if cmp := javaIntervalCompare(t, NewIntervalValue(10, 1), NewIntervalValue(20, 15)); cmp != 0 {
			t.Fatalf("compare=%d, want 0", cmp)
		}
	})
	t.Run("testSizeOfMaximumRepresentableInterval", func(t *testing.T) {
		javaIntervalSizeOverflow(t, NewIntervalValue(math.MinInt32, math.MaxInt32), false)
	})
	t.Run("testExtremeIntervalSize", func(t *testing.T) {
		for _, tc := range []struct {
			low, high int32
			want      int
		}{
			{math.MaxInt32, math.MaxInt32, 1}, {math.MinInt32, math.MinInt32, 1},
			{math.MaxInt32 - 10, math.MaxInt32, 11}, {math.MinInt32, math.MinInt32 + 10, 11},
		} {
			if size := javaIntervalSize(t, NewIntervalValue(tc.low, tc.high)); size != tc.want {
				t.Fatalf("size=%d, want %d", size, tc.want)
			}
		}
	})
	t.Run("testMaxIntSingletonEnumerator", func(t *testing.T) {
		iv := NewIntervalValue(math.MaxInt32, math.MaxInt32)
		e := iv.Elements()
		javaIntervalEqual(t, NewIntValue(math.MaxInt32), e.NextElement())
		if next := e.NextElement(); next != nil {
			t.Fatalf("nextElement=%v, want null", next)
		}
		e.Reset()
		javaIntervalEqual(t, NewIntValue(math.MaxInt32), e.NextElement())
		if next := e.NextElement(); next != nil {
			t.Fatalf("nextElement after reset=%v, want null", next)
		}
	})
	t.Run("testMaxIntSingletonSubsetEq", func(t *testing.T) {
		iv := NewIntervalValue(math.MaxInt32, math.MaxInt32)
		singleton := NewSetEnumValue([]Value{NewIntValue(math.MaxInt32)}, true)
		result, err := enumerableSubsetEq(iv, iv, singleton)
		if err != nil {
			t.Fatal(err)
		}
		javaIntervalEqual(t, BoolTrue, result)
	})
	t.Run("testMaxIntSingletonFingerprint", func(t *testing.T) {
		iv := NewIntervalValue(math.MaxInt32, math.MaxInt32)
		singleton := NewSetEnumValue([]Value{NewIntValue(math.MaxInt32)}, true)
		if expected, actual := singleton.FingerPrint(FP64New()), iv.FingerPrint(FP64New()); expected != actual {
			t.Fatalf("fingerprints=%x,%x", expected, actual)
		}
	})
	t.Run("testMaxIntSingletonDiffCapCup", func(t *testing.T) {
		iv := NewIntervalValue(math.MaxInt32, math.MaxInt32)
		singleton := NewSetEnumValue([]Value{NewIntValue(math.MaxInt32)}, true)
		empty := NewSetEnumValue([]Value{}, true)
		value, err := iv.Diff(singleton)
		if err != nil {
			t.Fatal(err)
		}
		if size := javaIntervalSize(t, value); size != 0 {
			t.Fatalf("diff(singleton).size=%d", size)
		}
		value, err = iv.Diff(empty)
		if err != nil {
			t.Fatal(err)
		}
		if size := javaIntervalSize(t, value); size != 1 {
			t.Fatalf("diff(empty).size=%d", size)
		}
		value, err = iv.Cap(singleton)
		if err != nil {
			t.Fatal(err)
		}
		if size := javaIntervalSize(t, value); size != 1 {
			t.Fatalf("cap(singleton).size=%d", size)
		}
		value, err = iv.Cup(empty)
		if err != nil {
			t.Fatal(err)
		}
		if size := javaIntervalSize(t, value); size != 1 {
			t.Fatalf("cup(empty).size=%d", size)
		}
	})
}
