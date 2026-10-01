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
