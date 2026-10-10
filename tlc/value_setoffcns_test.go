package tlc

import "testing"

func stringValues(values ...string) []Value {
	out := make([]Value, len(values))
	for i, value := range values {
		out[i] = NewStringValue(value)
	}
	return out
}

func TestSubsetValueElementsUseJavaCardinalityLayerOrder(t *testing.T) {
	base := NewSetEnumValue(stringValues("a", "b", "c"), true)
	subset := NewSubsetValue(base)
	enum := subset.Elements()
	want := []Value{
		EmptySet,
		NewSetEnumValue(stringValues("a"), true),
		NewSetEnumValue(stringValues("b"), true),
		NewSetEnumValue(stringValues("c"), true),
		NewSetEnumValue(stringValues("a", "b"), true),
		NewSetEnumValue(stringValues("a", "c"), true),
		NewSetEnumValue(stringValues("b", "c"), true),
		NewSetEnumValue(stringValues("a", "b", "c"), true),
	}
	for i, expected := range want {
		got := enum.NextElement()
		if got == nil {
			t.Fatalf("element %d = nil, want %s", i, expected)
		}
		if !mustValueEqual(got, expected) {
			t.Fatalf("element %d = %s, want %s", i, got, expected)
		}
	}
	if got := enum.NextElement(); got != nil {
		t.Fatalf("extra element = %s", got)
	}
	if err := enum.Err(); err != nil {
		t.Fatalf("enumeration error: %v", err)
	}
}

func TestSetOfFcnsValueDomainAndRangeEmptyBehaviors(t *testing.T) {
	rangeValues := NewSetEnumValue(stringValues("a", "b", "c"), true)
	emptyDomain := NewSetEnumValue([]Value{}, true)

	domainEmpty := NewSetOfFcnsValue(emptyDomain, rangeValues)
	if size, err := domainEmpty.Size(); err != nil || size != 1 {
		t.Fatalf("domain-empty size = %d/%v, want 1/nil", size, err)
	}
	enum := domainEmpty.Elements()
	if got := enum.NextElement(); !mustValueEqual(got, EmptyFcn) {
		t.Fatalf("domain-empty first element = %s, want empty function", got)
	}
	if got := enum.NextElement(); got != nil {
		t.Fatalf("domain-empty extra element = %s", got)
	}

	rangeEmpty := NewSetOfFcnsValue(NewIntervalValue(1, 2), EmptySet)
	if size, err := rangeEmpty.Size(); err != nil || size != 0 {
		t.Fatalf("range-empty size = %d/%v, want 0/nil", size, err)
	}
	if got := rangeEmpty.Elements().NextElement(); got != nil {
		t.Fatalf("range-empty first element = %s, want nil", got)
	}

	bothEmpty := NewSetOfFcnsValue(emptyDomain, EmptySet)
	if size, err := bothEmpty.Size(); err != nil || size != 1 {
		t.Fatalf("domain-and-range-empty size = %d/%v, want 1/nil", size, err)
	}
	enum = bothEmpty.Elements()
	if got := enum.NextElement(); !mustValueEqual(got, EmptyFcn) {
		t.Fatalf("domain-and-range-empty first element = %s, want empty function", got)
	}
	if got := enum.NextElement(); got != nil {
		t.Fatalf("domain-and-range-empty extra element = %s", got)
	}
}

func TestSetOfFcnsValueEnumeratesModelValueDomainDeterministically(t *testing.T) {
	ModelValueInit()
	domain := []Value{AddModelValue("m1"), AddModelValue("m2"), AddModelValue("m3")}
	fcns := NewSetOfFcnsValue(NewSetEnumValue(domain, true), NewSetEnumValue(stringValues("a", "b", "c"), true))
	if size, err := fcns.Size(); err != nil || size != 27 {
		t.Fatalf("function set size = %d/%v, want 27/nil", size, err)
	}

	enum := fcns.Elements()
	wantValues := [][]Value{
		stringValues("a", "a", "a"),
		stringValues("a", "a", "b"),
		stringValues("a", "a", "c"),
		stringValues("a", "b", "a"),
		stringValues("a", "b", "b"),
		stringValues("a", "b", "c"),
		stringValues("a", "c", "a"),
		stringValues("a", "c", "b"),
		stringValues("a", "c", "c"),
		stringValues("b", "a", "a"),
		stringValues("b", "a", "b"),
		stringValues("b", "a", "c"),
		stringValues("b", "b", "a"),
		stringValues("b", "b", "b"),
		stringValues("b", "b", "c"),
		stringValues("b", "c", "a"),
		stringValues("b", "c", "b"),
		stringValues("b", "c", "c"),
		stringValues("c", "a", "a"),
		stringValues("c", "a", "b"),
		stringValues("c", "a", "c"),
		stringValues("c", "b", "a"),
		stringValues("c", "b", "b"),
		stringValues("c", "b", "c"),
		stringValues("c", "c", "a"),
		stringValues("c", "c", "b"),
		stringValues("c", "c", "c"),
	}
	for i, values := range wantValues {
		expected := NewFcnRcdValue(domain, values, true)
		got := enum.NextElement()
		if got == nil {
			t.Fatalf("function %d = nil, want %s", i, expected)
		}
		if !mustValueEqual(got, expected) {
			t.Fatalf("function %d = %s, want %s", i, got, expected)
		}
		ok, err := fcns.Member(got)
		if err != nil || !ok {
			t.Fatalf("function %d membership = %v/%v, want true/nil", i, ok, err)
		}
	}
	if got := enum.NextElement(); got != nil {
		t.Fatalf("extra function = %s", got)
	}
}

func TestSetOfFcnsValueRangeSubsetUsesSubsetEnumerationOrder(t *testing.T) {
	ModelValueInit()
	domain := []Value{AddModelValue("m1"), AddModelValue("m2"), AddModelValue("m3")}
	subsets := NewSubsetValue(NewSetEnumValue(stringValues("a", "b", "c"), true))
	fcns := NewSetOfFcnsValue(NewSetEnumValue(domain, true), subsets)
	if size, err := fcns.Size(); err != nil || size != 512 {
		t.Fatalf("function set with subset range size = %d/%v, want 512/nil", size, err)
	}

	enum := fcns.Elements()
	empty := EmptySet
	want := []*FcnRcdValue{
		NewFcnRcdValue(domain, []Value{empty, empty, empty}, true),
		NewFcnRcdValue(domain, []Value{empty, empty, NewSetEnumValue(stringValues("a"), true)}, true),
		NewFcnRcdValue(domain, []Value{empty, empty, NewSetEnumValue(stringValues("b"), true)}, true),
		NewFcnRcdValue(domain, []Value{empty, empty, NewSetEnumValue(stringValues("c"), true)}, true),
		NewFcnRcdValue(domain, []Value{empty, empty, NewSetEnumValue(stringValues("a", "b"), true)}, true),
	}
	for i, expected := range want {
		got := enum.NextElement()
		if got == nil {
			t.Fatalf("function %d = nil, want %s", i, expected)
		}
		if !mustValueEqual(got, expected) {
			t.Fatalf("function %d = %s, want %s", i, got, expected)
		}
	}
}
