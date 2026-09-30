package tlc

import (
	"math"
	"testing"
)

func TestNaturalsAndIntegersDivisionUseTLAPlusFloorSemantics(t *testing.T) {
	cases := []struct {
		x    int32
		y    int32
		want int32
	}{
		{7, 3, 2},
		{-7, 3, -3},
		{7, -3, -3},
		{-7, -3, 2},
		{-6, 3, -2},
	}
	for _, tc := range cases {
		got, err := IntDivide(NewIntValue(tc.x), NewIntValue(tc.y))
		if err != nil {
			t.Fatalf("IntDivide(%d,%d) returned error: %v", tc.x, tc.y, err)
		}
		if got.Val != tc.want {
			t.Fatalf("IntDivide(%d,%d) = %d, want %d", tc.x, tc.y, got.Val, tc.want)
		}
		got, err = NatDivide(NewIntValue(tc.x), NewIntValue(tc.y))
		if err != nil {
			t.Fatalf("NatDivide(%d,%d) returned error: %v", tc.x, tc.y, err)
		}
		if got.Val != tc.want {
			t.Fatalf("NatDivide(%d,%d) = %d, want %d", tc.x, tc.y, got.Val, tc.want)
		}
	}
}

func TestNaturalsModuloAlwaysReturnsNonnegativeRemainderForPositiveDivisor(t *testing.T) {
	cases := []struct {
		x    int32
		y    int32
		want int32
	}{
		{7, 3, 1},
		{-7, 3, 2},
		{-6, 3, 0},
	}
	for _, tc := range cases {
		got, err := NatMod(NewIntValue(tc.x), NewIntValue(tc.y))
		if err != nil {
			t.Fatalf("NatMod(%d,%d) returned error: %v", tc.x, tc.y, err)
		}
		if got.Val != tc.want {
			t.Fatalf("NatMod(%d,%d) = %d, want %d", tc.x, tc.y, got.Val, tc.want)
		}
	}
}

func TestIntegersAndNaturalsArithmeticErrorsMatchJavaEdges(t *testing.T) {
	assertNumberError(t, func() (*IntValue, error) { return NatPlus(NewIntValue(math.MaxInt32), IntOne) })
	assertNumberError(t, func() (*IntValue, error) { return NatMinus(NewIntValue(math.MinInt32), IntOne) })
	assertNumberError(t, func() (*IntValue, error) { return NatTimes(NewIntValue(math.MaxInt32), NewIntValue(2)) })
	assertNumberError(t, func() (*IntValue, error) { return IntNeg(NewIntValue(math.MinInt32)) })
	assertNumberError(t, func() (*IntValue, error) { return IntDivide(NewIntValue(math.MinInt32), NewIntValue(-1)) })
	assertNumberError(t, func() (*IntValue, error) { return NatDivide(IntOne, IntZero) })
	assertNumberError(t, func() (*IntValue, error) { return NatMod(IntOne, IntZero) })
	assertNumberError(t, func() (*IntValue, error) { return NatExpt(IntZero, IntZero) })
	assertNumberError(t, func() (*IntValue, error) { return NatExpt(NewIntValue(2), NewIntValue(31)) })
}

func TestNaturalsAndIntegersSetMembershipAndOrdering(t *testing.T) {
	if ok, err := Nat().Member(IntZero); err != nil || !ok {
		t.Fatalf("0 \\in Nat = %v/%v, want true/nil", ok, err)
	}
	if ok, err := Nat().Member(IntNegOne); err != nil || ok {
		t.Fatalf("-1 \\in Nat = %v/%v, want false/nil", ok, err)
	}
	if ok, err := Int().Member(IntNegOne); err != nil || !ok {
		t.Fatalf("-1 \\in Int = %v/%v, want true/nil", ok, err)
	}
	cmp, err := Nat().Compare(Int())
	if err != nil || cmp >= 0 {
		t.Fatalf("Nat compare Int = %d/%v, want negative/nil", cmp, err)
	}
	cmp, err = Int().Compare(Nat())
	if err != nil || cmp <= 0 {
		t.Fatalf("Int compare Nat = %d/%v, want positive/nil", cmp, err)
	}
}

func assertNumberError(t *testing.T, fn func() (*IntValue, error)) {
	t.Helper()
	if value, err := fn(); err == nil {
		t.Fatalf("operation returned value %v and nil error, want error", value)
	}
}
