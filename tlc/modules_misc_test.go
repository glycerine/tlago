package tlc

import (
	"math"
	"testing"
)

func TestTLCCombineFcnOverridesTupleDomainValues(t *testing.T) {
	left := NewFcnRcdValue([]Value{NewIntValue(3)}, []Value{NewIntValue(11)}, true)
	right := NewTupleValue([]Value{NewIntValue(1), NewIntValue(2), NewIntValue(3)})

	combined, err := CombineFcn(left, right)
	if err != nil {
		t.Fatalf("CombineFcn returned error: %v", err)
	}
	fcn := requireFcnRcdValue(t, combined)
	fcn.Normalize()
	assertFunctionDomainValues(t, fcn,
		[]Value{NewIntValue(1), NewIntValue(2), NewIntValue(3)},
		[]Value{NewIntValue(1), NewIntValue(2), NewIntValue(11)})
}

func TestTLCCombineFcnExtendsTupleDomainValues(t *testing.T) {
	left := NewFcnRcdValue([]Value{NewIntValue(4)}, []Value{NewIntValue(11)}, true)
	right := NewTupleValue([]Value{NewIntValue(1), NewIntValue(2), NewIntValue(3), NewIntValue(11)})

	combined, err := CombineFcn(left, right)
	if err != nil {
		t.Fatalf("CombineFcn returned error: %v", err)
	}
	fcn := requireFcnRcdValue(t, combined)
	fcn.Normalize()
	assertFunctionDomainValues(t, fcn,
		[]Value{NewIntValue(1), NewIntValue(2), NewIntValue(3), NewIntValue(4)},
		[]Value{NewIntValue(1), NewIntValue(2), NewIntValue(3), NewIntValue(11)})
}

func TestTLCCombineFcnMaxIntIntervalKeepsLeftValue(t *testing.T) {
	max := int32(math.MaxInt32)
	intervalLeft := NewFcnRcdIntervalValue(NewIntervalValue(max, max), []Value{IntZero})
	explicitRight := NewFcnRcdValue([]Value{NewIntValue(max)}, []Value{IntOne}, true)
	assertMaxIntCombine(t, intervalLeft, explicitRight)

	explicitLeft := NewFcnRcdValue([]Value{NewIntValue(max)}, []Value{IntZero}, true)
	intervalRight := NewFcnRcdIntervalValue(NewIntervalValue(max, max), []Value{IntOne})
	assertMaxIntCombine(t, explicitLeft, intervalRight)
}

func assertMaxIntCombine(t *testing.T, left Value, right Value) {
	t.Helper()
	combined, err := CombineFcn(left, right)
	if err != nil {
		t.Fatalf("CombineFcn returned error: %v", err)
	}
	fcn := requireFcnRcdValue(t, combined)
	assertFunctionDomainValues(t, fcn, []Value{NewIntValue(math.MaxInt32)}, []Value{IntZero})
}

func requireFcnRcdValue(t *testing.T, value Value) *FcnRcdValue {
	t.Helper()
	fcn, ok := value.(*FcnRcdValue)
	if !ok {
		t.Fatalf("value %T = %v, want FcnRcdValue", value, value)
	}
	return fcn
}

func assertFunctionDomainValues(t *testing.T, fcn *FcnRcdValue, wantDomain []Value, wantValues []Value) {
	t.Helper()
	if len(fcn.Domain) != len(wantDomain) {
		t.Fatalf("domain len = %d, want %d", len(fcn.Domain), len(wantDomain))
	}
	if len(fcn.Values) != len(wantValues) {
		t.Fatalf("values len = %d, want %d", len(fcn.Values), len(wantValues))
	}
	for i := range wantDomain {
		if !mustValueEqual(fcn.Domain[i], wantDomain[i]) {
			t.Fatalf("domain[%d] = %v, want %v", i, fcn.Domain[i], wantDomain[i])
		}
	}
	for i := range wantValues {
		if !mustValueEqual(fcn.Values[i], wantValues[i]) {
			t.Fatalf("values[%d] = %v, want %v", i, fcn.Values[i], wantValues[i])
		}
	}
}

func mustValueEqual(a Value, b Value) bool {
	eq, err := a.Equal(b)
	if err != nil {
		panic(err)
	}
	return eq
}
