package tlc

import "testing"

func singletonSet(value Value) *SetEnumValue {
	return NewSetEnumValue([]Value{value}, true)
}

func expectEmptyValue(t *testing.T, label string, value Value, want bool) {
	t.Helper()
	got, err := IsEmptyValue(value)
	if err != nil {
		t.Fatalf("%s IsEmptyValue returned error: %v", label, err)
	}
	if got != want {
		t.Fatalf("%s IsEmptyValue = %v, want %v", label, got, want)
	}
}

func TestValueIsEmptyMatchesJavaSetConstructors(t *testing.T) {
	one := IntOne
	two := NewIntValue(2)

	expectEmptyValue(t, "{}", EmptySet, true)
	expectEmptyValue(t, "{1}", singletonSet(one), false)

	expectEmptyValue(t, "1..0", NewIntervalValue(1, 0), true)
	expectEmptyValue(t, "1..1", NewIntervalValue(1, 1), false)

	expectEmptyValue(t, "{1} \\cap {2}", NewSetCapValue(singletonSet(one), singletonSet(two)), true)
	expectEmptyValue(t, "{1} \\cap {1}", NewSetCapValue(singletonSet(one), singletonSet(one)), false)

	expectEmptyValue(t, "{} \\cup {}", NewSetCupValue(EmptySet, EmptySet), true)
	expectEmptyValue(t, "{} \\cup {1}", NewSetCupValue(EmptySet, singletonSet(one)), false)

	expectEmptyValue(t, "{1} \\ {1}", NewSetDiffValue(singletonSet(one), singletonSet(one)), true)
	expectEmptyValue(t, "{1} \\ {2}", NewSetDiffValue(singletonSet(one), singletonSet(two)), false)

	expectEmptyValue(t, "[{1} -> {}]", NewSetOfFcnsValue(singletonSet(one), EmptySet), true)
	expectEmptyValue(t, "[{} -> {}]", NewSetOfFcnsValue(EmptySet, EmptySet), false)

	rcdsEmpty, err := NewSetOfRcdsValue([]*UniqueString{UniqueStringOf("n1")}, []Value{EmptySet}, false)
	if err != nil {
		t.Fatalf("NewSetOfRcdsValue empty returned error: %v", err)
	}
	expectEmptyValue(t, "[n1: {}]", rcdsEmpty, true)
	rcdsNonEmpty, err := NewSetOfRcdsValue([]*UniqueString{UniqueStringOf("n1")}, []Value{singletonSet(one)}, false)
	if err != nil {
		t.Fatalf("NewSetOfRcdsValue nonempty returned error: %v", err)
	}
	expectEmptyValue(t, "[n1: {1}]", rcdsNonEmpty, false)

	expectEmptyValue(t, "{} \\X {1}", NewSetOfTuplesValue([]Value{EmptySet, singletonSet(one)}), true)
	expectEmptyValue(t, "{1} \\X {1}", NewSetOfTuplesValue([]Value{singletonSet(one), singletonSet(one)}), false)

	expectEmptyValue(t, "SUBSET {}", NewSubsetValue(EmptySet), false)
	expectEmptyValue(t, "SUBSET {1}", NewSubsetValue(singletonSet(one)), false)

	expectEmptyValue(t, "UNION {}", NewUnionValue(EmptySet), true)
	expectEmptyValue(t, "UNION {{}}", NewUnionValue(singletonSet(EmptySet)), true)
	expectEmptyValue(t, "UNION {{1}}", NewUnionValue(singletonSet(singletonSet(one))), false)

	expectEmptyValue(t, "Nat", Nat(), false)
	expectEmptyValue(t, "Int", Int(), false)
	expectEmptyValue(t, "STRING", STRING(), false)
	expectEmptyValue(t, "Seq({})", Seq(EmptySet), false)
}

func TestValueIsEmptyRejectsUnspecifiedNonSetValues(t *testing.T) {
	if _, err := IsEmptyValue(AnySetValue); err == nil {
		t.Fatalf("IsEmptyValue(ANY) returned nil error")
	}
	if _, err := IsEmptyValue(IntOne); err == nil {
		t.Fatalf("IsEmptyValue(1) returned nil error")
	}
}
