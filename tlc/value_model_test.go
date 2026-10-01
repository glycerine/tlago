package tlc

import "testing"

func resetModelValueTestState() {
	UniqueStringInitialize()
	ModelValueInit()
}

func TestModelValueUntypedEqualityAndOrdering(t *testing.T) {
	resetModelValueTestState()
	u1 := MakeModelValue("u1")
	u2 := MakeModelValue("u2")

	if eq, err := u1.Equal(u2); err != nil || eq {
		t.Fatalf("u1 Equal u2 = %v/%v, want false/nil", eq, err)
	}
	if eq, err := u1.Equal(u1); err != nil || !eq {
		t.Fatalf("u1 Equal u1 = %v/%v, want true/nil", eq, err)
	}
	if cmp, err := u1.Compare(u2); err != nil || cmp >= 0 {
		t.Fatalf("u1 Compare u2 = %d/%v, want negative/nil", cmp, err)
	}
	if cmp, err := u2.Compare(u1); err != nil || cmp <= 0 {
		t.Fatalf("u2 Compare u1 = %d/%v, want positive/nil", cmp, err)
	}
}

func TestModelValueUntypedComparedWithOrdinaryValues(t *testing.T) {
	resetModelValueTestState()
	untyped := MakeModelValue("untyped")
	values := []Value{
		NewStringValue("str"),
		IntZero,
		BoolFalse,
		NewIntervalValue(0, 1),
		NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, true),
		NewTupleValue([]Value{NewStringValue("foo")}),
	}
	for _, value := range values {
		if eq, err := value.Equal(untyped); err != nil || eq {
			t.Fatalf("%T Equal untyped model = %v/%v, want false/nil", value, eq, err)
		}
		if eq, err := untyped.Equal(value); err != nil || eq {
			t.Fatalf("untyped model Equal %T = %v/%v, want false/nil", value, eq, err)
		}
		if cmp, err := value.Compare(untyped); err != nil || cmp <= 0 {
			t.Fatalf("%T Compare untyped model = %d/%v, want positive/nil", value, cmp, err)
		}
		if cmp, err := untyped.Compare(value); err != nil || cmp >= 0 {
			t.Fatalf("untyped model Compare %T = %d/%v, want negative/nil", value, cmp, err)
		}
	}
}

func TestModelValueTypedEqualityAndOrdering(t *testing.T) {
	resetModelValueTestState()
	a := MakeModelValue("A_a")
	b := MakeModelValue("A_b")
	untyped := MakeModelValue("untyped")

	if eq, err := a.Equal(b); err != nil || eq {
		t.Fatalf("A_a Equal A_b = %v/%v, want false/nil", eq, err)
	}
	if eq, err := a.Equal(a); err != nil || !eq {
		t.Fatalf("A_a Equal A_a = %v/%v, want true/nil", eq, err)
	}
	if cmp, err := a.Compare(b); err != nil || cmp >= 0 {
		t.Fatalf("A_a Compare A_b = %d/%v, want negative/nil", cmp, err)
	}
	if cmp, err := b.Compare(a); err != nil || cmp <= 0 {
		t.Fatalf("A_b Compare A_a = %d/%v, want positive/nil", cmp, err)
	}

	if eq, err := a.Equal(untyped); err != nil || eq {
		t.Fatalf("typed Equal untyped = %v/%v, want false/nil", eq, err)
	}
	if eq, err := untyped.Equal(a); err != nil || eq {
		t.Fatalf("untyped Equal typed = %v/%v, want false/nil", eq, err)
	}
	if cmp, err := a.Compare(untyped); err != nil || cmp >= 0 {
		t.Fatalf("typed Compare later untyped = %d/%v, want negative/nil", cmp, err)
	}
	if cmp, err := untyped.Compare(a); err != nil || cmp <= 0 {
		t.Fatalf("later untyped Compare typed = %d/%v, want positive/nil", cmp, err)
	}
}

func TestModelValueTypedRejectsDifferentTypeAndOrdinaryValues(t *testing.T) {
	resetModelValueTestState()
	typed := MakeModelValue("B_b")
	otherTyped := MakeModelValue("A_a")
	values := []Value{
		otherTyped,
		NewStringValue("str"),
		IntZero,
		BoolFalse,
		NewIntervalValue(0, 1),
		NewRecordValue([]*UniqueString{UniqueStringOf("foo")}, []Value{NewStringValue("bar")}, true),
		NewTupleValue([]Value{NewStringValue("foo")}),
	}
	for _, value := range values {
		if _, err := value.Equal(typed); err == nil {
			t.Fatalf("%T Equal typed model returned nil error", value)
		}
		if _, err := typed.Equal(value); err == nil {
			t.Fatalf("typed model Equal %T returned nil error", value)
		}
		if _, err := value.Compare(typed); err == nil {
			t.Fatalf("%T Compare typed model returned nil error", value)
		}
		if _, err := typed.Compare(value); err == nil {
			t.Fatalf("typed model Compare %T returned nil error", value)
		}
	}
}
