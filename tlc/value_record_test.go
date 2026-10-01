package tlc

import (
	"strings"
	"testing"
)

func TestRecordValueDeepCopyIndependentOfSourceNormalization(t *testing.T) {
	a := UniqueStringOf("a")
	b := UniqueStringOf("b")
	aVal := NewStringValue("aVal")
	bVal := NewStringValue("bVal")

	orig := NewRecordValue([]*UniqueString{b, a}, []Value{bVal, aVal}, false)
	if !orig.Names[0].Equal(b) || !orig.Names[1].Equal(a) {
		t.Fatalf("original names before copy = %v, want [b a]", orig.Names)
	}
	if !mustValueEqual(orig.Values[0], bVal) || !mustValueEqual(orig.Values[1], aVal) {
		t.Fatalf("original values before copy = %v, want [bVal aVal]", orig.Values)
	}

	copied, ok := orig.DeepCopy().(*RecordValue)
	if !ok {
		t.Fatalf("DeepCopy returned %T, want *RecordValue", orig.DeepCopy())
	}

	orig.DeepNormalize()
	if !orig.Names[0].Equal(a) || !orig.Names[1].Equal(b) {
		t.Fatalf("normalized original names = %v, want [a b]", orig.Names)
	}
	if !mustValueEqual(orig.Values[0], aVal) || !mustValueEqual(orig.Values[1], bVal) {
		t.Fatalf("normalized original values = %v, want [aVal bVal]", orig.Values)
	}

	if !copied.Names[0].Equal(b) || !copied.Names[1].Equal(a) {
		t.Fatalf("copy names after source normalization = %v, want [b a]", copied.Names)
	}
	if !mustValueEqual(copied.Values[0], bVal) || !mustValueEqual(copied.Values[1], aVal) {
		t.Fatalf("copy values after source normalization = %v, want [bVal aVal]", copied.Values)
	}
}

func TestRecordValueApplyReportsMissingAndNonStringFields(t *testing.T) {
	rec := NewRecordValue([]*UniqueString{UniqueStringOf("a")}, []Value{NewStringValue("aVal")}, true)

	_, err := rec.Apply(NewStringValue("b"))
	if err == nil {
		t.Fatalf("Apply missing field returned nil error")
	}
	if msg := err.Error(); !strings.Contains(msg, "Attempted to access nonexistent field 'b' of record") ||
		!strings.Contains(msg, `[a |-> "aVal"]`) {
		t.Fatalf("missing field error = %q", msg)
	}

	_, err = rec.Apply(IntZero)
	if err == nil {
		t.Fatalf("Apply non-string field returned nil error")
	}
	if msg := err.Error(); !strings.Contains(msg, "Attempted to access record by a non-string argument: 0") {
		t.Fatalf("non-string field error = %q", msg)
	}
}
