package tlc

import (
	"fmt"
	"testing"
)

func javaInitializeValueSetup(t *testing.T) {
	t.Helper()
	previous := FP64IrredPoly()
	FP64Init()
	t.Cleanup(func() { FP64InitPoly(previous) })
}

func requireJavaInitialization(t *testing.T, value Value, expected bool) {
	t.Helper()
	if got := value.IsNormalized(); got != expected {
		t.Fatalf("%T.isNormalized = %v, want %v", value, got, expected)
	}
}

func javaInitializeValueVec() *ValueVec {
	vec := NewValueVec()
	vec.Add(NewIntValue(42))
	vec.Add(NewIntValue(23))
	vec.Add(NewIntValue(4711))
	vec.Add(NewIntValue(1))
	return vec
}

func TestJavaInitializeValueUnion(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := NewValueVec()
	vec.Add(NewSetEnumValue([]Value{NewIntValue(42)}, false))
	vec.Add(NewSetEnumValue([]Value{NewIntValue(23)}, false))
	vec.Add(NewSetEnumValue([]Value{NewIntValue(4711)}, false))
	vec.Add(NewSetEnumValue([]Value{NewIntValue(1)}, false))
	uv := NewUnionValue(NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, uv, false)
	requireJavaInitialization(t, uv.Set, false)
	InitializeValue(uv)
	requireJavaInitialization(t, uv.Set, true)
	requireJavaInitialization(t, uv, true)
}

func TestJavaInitializeValueSetCap(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	scv := NewSetCapValue(NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, scv, false)
	requireJavaInitialization(t, scv.Set1, false)
	requireJavaInitialization(t, scv.Set2, false)
	InitializeValue(scv)
	requireJavaInitialization(t, scv.Set1, true)
	requireJavaInitialization(t, scv.Set2, true)
	requireJavaInitialization(t, scv, true)
}

func TestJavaInitializeValueSetCup(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	scv := NewSetCupValue(NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, scv, false)
	requireJavaInitialization(t, scv.Set1, false)
	requireJavaInitialization(t, scv.Set2, false)
	InitializeValue(scv)
	requireJavaInitialization(t, scv.Set1, true)
	requireJavaInitialization(t, scv.Set2, true)
	requireJavaInitialization(t, scv, true)
}

func TestJavaInitializeValueSetDiff(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	scv := NewSetDiffValue(NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, scv, false)
	requireJavaInitialization(t, scv.Set1, false)
	requireJavaInitialization(t, scv.Set2, false)
	InitializeValue(scv)
	requireJavaInitialization(t, scv.Set1, true)
	requireJavaInitialization(t, scv.Set2, true)
	requireJavaInitialization(t, scv, true)
}

func TestJavaInitializeValueSubset(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	scv := NewSubsetValue(NewSetEnumValueVec(vec, false))
	requireJavaInitialization(t, scv, false)
	requireJavaInitialization(t, scv.Set, false)
	InitializeValue(scv)
	requireJavaInitialization(t, scv.Set, true)
	requireJavaInitialization(t, scv, true)
}

func TestJavaInitializeValueRecord(t *testing.T) {
	javaInitializeValueSetup(t)
	internTable := NewInternTable(2)
	a, b := internTable.Put("a"), internTable.Put("b")
	vec := javaInitializeValueVec()
	aVal, bVal := NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false)
	rcdv := NewRecordValue([]*UniqueString{b, a}, []Value{bVal, aVal}, false)
	requireJavaInitialization(t, rcdv, false)
	for _, v := range rcdv.Values {
		requireJavaInitialization(t, v, false)
	}
	InitializeValue(rcdv)
	for _, v := range rcdv.Values {
		requireJavaInitialization(t, v, true)
	}
	requireJavaInitialization(t, rcdv, true)
}

func TestJavaInitializeValueFcnRecord(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	aVal, bVal := NewSetEnumValueVec(vec, false), NewSetEnumValueVec(vec, false)
	rcdv := NewFcnRcdValue([]Value{NewStringValue("B"), NewStringValue("A")}, []Value{bVal, aVal}, false)
	requireJavaInitialization(t, rcdv, false)
	for _, v := range rcdv.Values {
		requireJavaInitialization(t, v, false)
	}
	InitializeValue(rcdv)
	for _, v := range rcdv.Values {
		requireJavaInitialization(t, v, true)
	}
	requireJavaInitialization(t, rcdv, true)
}

func TestJavaInitializeValueTuple(t *testing.T) {
	javaInitializeValueSetup(t)
	vec := javaInitializeValueVec()
	aVal := NewSetEnumValueVec(vec, false)
	tuple := NewSingletonTupleValue(aVal)
	InitializeValue(tuple)
	requireJavaInitialization(t, tuple, true)
	for _, v := range tuple.Elems {
		requireJavaInitialization(t, v, true)
	}
}

func TestJavaInitializeValueSetOfTuple(t *testing.T) {
	javaInitializeValueSetup(t)
	intVal := NewIntervalValue(1, 2)
	inner := NewSetOfTuplesValue([]Value{intVal, intVal})
	tuples := NewSetOfTuplesValue([]Value{inner, inner})
	InitializeValue(tuples)
	requireJavaInitialization(t, tuples, true)
	for _, v := range tuples.Sets {
		requireJavaInitialization(t, v, true)
	}
}

func TestJavaInitializeValueSetOfRcds(t *testing.T) {
	javaInitializeValueSetup(t)
	values := []Value{
		NewSetEnumValue(javaInitializeRecordValues(7, "a"), true),
		NewIntervalValue(1, 2),
		NewIntervalValue(1, 4),
	}
	setOfRcrds, err := NewSetOfRcdsValue(javaInitializeRecordNames(3), values, true)
	if err != nil {
		t.Fatal(err)
	}
	InitializeValue(setOfRcrds)
	requireJavaInitialization(t, setOfRcrds, true)
	for _, v := range setOfRcrds.Values {
		requireJavaInitialization(t, v, true)
	}
}

func javaInitializeRecordValues(n int, str string) []Value {
	values := make([]Value, n)
	for i := range values {
		values[i] = NewStringValue(fmt.Sprint(str, i))
	}
	return values
}

func javaInitializeRecordNames(n int) []*UniqueString {
	names := make([]*UniqueString, n)
	for i := range names {
		names[i] = UniqueStringOf(fmt.Sprint("N", i))
	}
	return names
}
