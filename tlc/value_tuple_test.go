package tlc

import "testing"

func TestTupleCompareAndEqualConvertOtherValueToTupleLikeJava(t *testing.T) {
	tuple := NewTupleValue([]Value{IntOne, NewIntValue(2)})
	fcn := NewFcnRcdIntervalValue(NewIntervalValue(1, 2), []Value{IntOne, NewIntValue(2)})

	cmp, err := tuple.Compare(fcn)
	if err != nil {
		t.Fatalf("Compare function record: %v", err)
	}
	if cmp != 0 {
		t.Fatalf("Compare function record = %d, want 0", cmp)
	}

	eq, err := tuple.Equal(fcn)
	if err != nil {
		t.Fatalf("Equal function record: %v", err)
	}
	if !eq {
		t.Fatalf("Equal function record = false, want true")
	}
}

func TestEmptyTupleEqualsEmptyRecordLikeJavaToTuple(t *testing.T) {
	eq, err := EmptyTuple.Equal(EmptyRecord)
	if err != nil {
		t.Fatalf("Equal empty record: %v", err)
	}
	if !eq {
		t.Fatalf("empty tuple should equal empty record through Java toTuple")
	}
}

func TestTupleValuePublicCarrierAndConversionSurface(t *testing.T) {
	one := NewSingletonTupleValue(IntOne)
	if got := one.GetElem(0); got != IntOne {
		t.Fatalf("singleton GetElem(0) = %v, want IntOne", got)
	}
	if got := one.GetElems(); len(got) != 1 || &got[0] != &one.Elems[0] {
		t.Fatal("GetElems did not return the tuple's backing array")
	}
	if one.ToTuple() != one {
		t.Fatal("ToTuple did not preserve receiver identity")
	}
	if one.ToRcd() != nil {
		t.Fatal("non-empty ToRcd should return nil")
	}
	pair := NewPairTupleValue(IntZero, IntOne)
	if len(pair.Elems) != 2 || pair.GetElem(0) != IntZero || pair.GetElem(1) != IntOne {
		t.Fatalf("pair elements = %v, want [0 1]", pair.Elems)
	}
	if EmptyTuple.ToRcd() != EmptyRecord {
		t.Fatal("empty ToRcd did not return EmptyRecord")
	}
}
