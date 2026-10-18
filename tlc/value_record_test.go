package tlc

import (
	"strings"
	"testing"
)

func TestRecordValueDeepCopyIndependentOfSourceNormalization(t *testing.T) {
	assertTrue := func(condition bool) {
		t.Helper()
		if !condition {
			t.Fatal("Java RecordValue assertion failed")
		}
	}
	internTable := NewInternTable(2)
	a := internTable.Put("a")
	b := internTable.Put("b")
	aVal := NewStringValue("aVal")
	bVal := NewStringValue("bVal")
	orig := NewRecordValue([]*UniqueString{b, a}, []Value{bVal, aVal}, false)
	assertTrue(orig.Names[0].Equal(b))
	assertTrue(orig.Names[1].Equal(a))
	assertTrue(mustValueEqual(orig.Values[0], bVal))
	assertTrue(mustValueEqual(orig.Values[1], aVal))
	deepCopy := orig.DeepCopy().(*RecordValue)
	orig.DeepNormalize()
	assertTrue(orig.Names[0].Equal(a))
	assertTrue(orig.Names[1].Equal(b))
	assertTrue(mustValueEqual(orig.Values[0], aVal))
	assertTrue(mustValueEqual(orig.Values[1], bVal))
	assertTrue(deepCopy.Names[0].Equal(b))
	assertTrue(deepCopy.Names[1].Equal(a))
	assertTrue(mustValueEqual(deepCopy.Values[0], bVal))
	assertTrue(mustValueEqual(deepCopy.Values[1], aVal))
}

func TestRecordValueApplyReportsMissingAndNonStringFields(t *testing.T) {
	assertTrue := func(condition bool) {
		t.Helper()
		if !condition {
			t.Fatal("Java RecordValue assertion failed")
		}
	}
	aVal := NewStringValue("aVal")
	recVal := NewSingletonRecordValue(UniqueStringOf("a"), aVal)
	if recVal.IsNormalized() {
		t.Fatal("two-argument singleton constructor should default to unnormalized")
	}
	if recVal.ToRcd() != recVal {
		t.Fatal("ToRcd did not preserve receiver identity")
	}
	if normalized := NewSingletonRecordValue(UniqueStringOf("a"), aVal, true); !normalized.IsNormalized() {
		t.Fatal("explicitly normalized singleton constructor lost its flag")
	}
	if _, err := recVal.Apply(NewStringValue("b")); err != nil {
		if !isJavaEvalOrRuntimeException(err) {
			t.Fatal(err)
		}
		assertTrue(strings.Contains(err.Error(), "Attempted to access nonexistent field 'b' of record\n[a |-> \"aVal\"]"))
	}
	if _, err := recVal.Apply(NewIntValue(0)); err != nil {
		if !isJavaEvalOrRuntimeException(err) {
			t.Fatal(err)
		}
		assertTrue(strings.Contains(err.Error(), "Attempted to access record by a non-string argument: 0"))
	}
	if _, err := recVal.Select(NewIntValue(0)); err != nil {
		if !isJavaEvalOrRuntimeException(err) {
			t.Fatal(err)
		}
		assertTrue(strings.Contains(err.Error(), "Attempted to access record by a non-string argument: 0"))
	}
}
