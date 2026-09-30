package tlc

import "testing"

func unnormalizedIntSet() *SetEnumValue {
	return NewSetEnumValue([]Value{
		NewIntValue(42),
		NewIntValue(23),
		NewIntValue(4711),
		IntOne,
	}, false)
}

func requireNormalized(t *testing.T, label string, value Value) {
	t.Helper()
	if !value.IsNormalized() {
		t.Fatalf("%s is not normalized", label)
	}
}

func requireNotNormalized(t *testing.T, label string, value Value) {
	t.Helper()
	if value.IsNormalized() {
		t.Fatalf("%s is already normalized", label)
	}
}

func TestInitializeValueNormalizesLazySetOperations(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value Value
		left  *SetEnumValue
		right *SetEnumValue
	}{
		{"cap", NewSetCapValue(unnormalizedIntSet(), unnormalizedIntSet()), nil, nil},
		{"cup", NewSetCupValue(unnormalizedIntSet(), unnormalizedIntSet()), nil, nil},
		{"diff", NewSetDiffValue(unnormalizedIntSet(), unnormalizedIntSet()), nil, nil},
	} {
		switch value := tc.value.(type) {
		case *SetCapValue:
			tc.left, tc.right = value.Set1.(*SetEnumValue), value.Set2.(*SetEnumValue)
		case *SetCupValue:
			tc.left, tc.right = value.Set1.(*SetEnumValue), value.Set2.(*SetEnumValue)
		case *SetDiffValue:
			tc.left, tc.right = value.Set1.(*SetEnumValue), value.Set2.(*SetEnumValue)
		}
		requireNotNormalized(t, tc.name, tc.value)
		requireNotNormalized(t, tc.name+" left", tc.left)
		requireNotNormalized(t, tc.name+" right", tc.right)

		InitializeValue(tc.value)

		requireNormalized(t, tc.name, tc.value)
		requireNormalized(t, tc.name+" left", tc.left)
		requireNormalized(t, tc.name+" right", tc.right)
	}
}

func TestInitializeValueNormalizesUnionAndSubset(t *testing.T) {
	outer := NewSetEnumValue([]Value{
		NewSetEnumValue([]Value{NewIntValue(42)}, false),
		NewSetEnumValue([]Value{NewIntValue(23)}, false),
		NewSetEnumValue([]Value{NewIntValue(4711)}, false),
		NewSetEnumValue([]Value{IntOne}, false),
	}, false)
	union := NewUnionValue(outer)
	requireNotNormalized(t, "union", union)
	requireNotNormalized(t, "union set", outer)

	InitializeValue(union)

	requireNormalized(t, "union", union)
	requireNormalized(t, "union set", outer)

	subsetSet := unnormalizedIntSet()
	subset := NewSubsetValue(subsetSet)
	requireNotNormalized(t, "subset", subset)
	requireNotNormalized(t, "subset set", subsetSet)

	InitializeValue(subset)

	requireNormalized(t, "subset", subset)
	requireNormalized(t, "subset set", subsetSet)
}

func TestInitializeValueNormalizesCompositeChildren(t *testing.T) {
	a := UniqueStringOf("a")
	b := UniqueStringOf("b")
	aVal := unnormalizedIntSet()
	bVal := unnormalizedIntSet()
	record := NewRecordValue([]*UniqueString{b, a}, []Value{bVal, aVal}, false)
	requireNotNormalized(t, "record", record)
	requireNotNormalized(t, "record a value", aVal)
	requireNotNormalized(t, "record b value", bVal)

	InitializeValue(record)

	requireNormalized(t, "record", record)
	requireNormalized(t, "record a value", aVal)
	requireNormalized(t, "record b value", bVal)

	fcnA := unnormalizedIntSet()
	fcnB := unnormalizedIntSet()
	fcn := NewFcnRcdValue([]Value{NewStringValue("B"), NewStringValue("A")}, []Value{fcnB, fcnA}, false)
	requireNotNormalized(t, "function record", fcn)
	requireNotNormalized(t, "function A value", fcnA)
	requireNotNormalized(t, "function B value", fcnB)

	InitializeValue(fcn)

	requireNormalized(t, "function record", fcn)
	requireNormalized(t, "function A value", fcnA)
	requireNormalized(t, "function B value", fcnB)

	tupleValue := unnormalizedIntSet()
	tuple := NewTupleValue([]Value{tupleValue})
	requireNotNormalized(t, "tuple child", tupleValue)

	InitializeValue(tuple)

	requireNormalized(t, "tuple", tuple)
	requireNormalized(t, "tuple child", tupleValue)
}
