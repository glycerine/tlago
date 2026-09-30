package tlc

import "testing"

func countValues(t *testing.T, enum ValueEnumeration) int {
	t.Helper()
	count := 0
	for {
		value := enum.NextElement()
		if value == nil {
			if err := enum.Err(); err != nil {
				t.Fatalf("enumeration error after %d values: %v", count, err)
			}
			return count
		}
		count++
	}
}

func TestSetOfTuplesValueLazyStringMatchesJavaProductRendering(t *testing.T) {
	oldEnumBound := Globals.EnumBound
	Globals.EnumBound = 1
	defer func() { Globals.EnumBound = oldEnumBound }()

	ints := NewIntervalValue(1, 2)
	inner := NewSetOfTuplesValue([]Value{ints, ints})
	outer := NewSetOfTuplesValue([]Value{inner, inner})
	want := "((1..2 \\X 1..2) \\X (1..2 \\X 1..2))"
	if got := outer.String(); got != want {
		t.Fatalf("lazy product string = %q, want %q", got, want)
	}
}

func TestSetOfTuplesValueEmptyFiniteComponentShortCircuitsNat(t *testing.T) {
	product := NewSetOfTuplesValue([]Value{Nat(), EmptySet})
	if size, err := product.Size(); err != nil || size != 0 {
		t.Fatalf("Nat x empty size = %d/%v, want 0/nil", size, err)
	}
	if finite, err := product.IsFinite(); err != nil || !finite {
		t.Fatalf("Nat x empty finite = %v/%v, want true/nil", finite, err)
	}
	if count := countValues(t, product.Elements()); count != 0 {
		t.Fatalf("Nat x empty enumerated %d values, want 0", count)
	}
}

func TestSetOfRcdsValueSimpleEnumerationPreservesNamesAndFieldMembership(t *testing.T) {
	names := []*UniqueString{UniqueStringOf("N0"), UniqueStringOf("N1"), UniqueStringOf("N2")}
	values := []Value{
		NewSetEnumValue(stringValues("a0", "a1", "a2", "a3", "a4", "a5", "a6"), true),
		NewIntervalValue(1, 2),
		NewIntervalValue(1, 4),
	}
	rcds, err := NewSetOfRcdsValue(names, values, true)
	if err != nil {
		t.Fatalf("NewSetOfRcdsValue returned error: %v", err)
	}
	if size, err := rcds.Size(); err != nil || size != 56 {
		t.Fatalf("record set size = %d/%v, want 56/nil", size, err)
	}

	enum := rcds.Elements()
	for i := 0; i < 56; i++ {
		value := enum.NextElement()
		if value == nil {
			t.Fatalf("record %d = nil, want value", i)
		}
		record, ok := value.(*RecordValue)
		if !ok {
			t.Fatalf("record %d has type %T, want *RecordValue", i, value)
		}
		for j, name := range names {
			if !record.Names[j].Equal(name) {
				t.Fatalf("record %d name %d = %s, want %s", i, j, record.Names[j], name)
			}
			member, err := values[j].Member(record.Values[j])
			if err != nil || !member {
				t.Fatalf("record %d field %s membership = %v/%v, want true/nil", i, name, member, err)
			}
		}
	}
	if got := enum.NextElement(); got != nil {
		t.Fatalf("extra record = %s", got)
	}
	if err := enum.Err(); err != nil {
		t.Fatalf("record enumeration error: %v", err)
	}
}

func TestSetOfRcdsValueEmptyFiniteFieldShortCircuitsNat(t *testing.T) {
	rcds, err := NewSetOfRcdsValue(
		[]*UniqueString{UniqueStringOf("N0"), UniqueStringOf("N1")},
		[]Value{Nat(), EmptySet},
		false)
	if err != nil {
		t.Fatalf("NewSetOfRcdsValue returned error: %v", err)
	}
	if size, err := rcds.Size(); err != nil || size != 0 {
		t.Fatalf("[N0: Nat, N1: {}] size = %d/%v, want 0/nil", size, err)
	}
	if finite, err := rcds.IsFinite(); err != nil || !finite {
		t.Fatalf("[N0: Nat, N1: {}] finite = %v/%v, want true/nil", finite, err)
	}
	if count := countValues(t, rcds.Elements()); count != 0 {
		t.Fatalf("[N0: Nat, N1: {}] enumerated %d values, want 0", count)
	}
}
