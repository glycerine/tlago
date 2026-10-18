package tlc

import "testing"

func TestSequencesStringTailConcatAndSubSeqMatchJava(t *testing.T) {
	UniqueStringInitialize()
	tail, err := Tail(NewStringValue("abc"))
	if err != nil {
		t.Fatalf("Tail(\"abc\") returned error: %v", err)
	}
	assertStringValue(t, tail, "bc")

	concat, err := Concat(NewStringValue("abc"), NewStringValue("d"))
	if err != nil {
		t.Fatalf("Concat(strings) returned error: %v", err)
	}
	assertStringValue(t, concat, "abcd")

	subseq, err := SubSeq(NewStringValue("abc"), IntOne, IntOne)
	if err != nil {
		t.Fatalf("SubSeq(\"abc\",1,1) returned error: %v", err)
	}
	assertStringValue(t, subseq, "a")
}

func TestSequencesRejectJavaStringOperationsThatAreNotSequences(t *testing.T) {
	UniqueStringInitialize()
	assertSequenceError(t, func() (Value, error) { return Head(NewStringValue("a")) })
	assertSequenceError(t, func() (Value, error) { return Head(NewStringValue("")) })
	assertSequenceError(t, func() (Value, error) { return Append(NewStringValue(""), NewStringValue("a")) })
	assertSequenceError(t, func() (Value, error) { return Append(NewStringValue("abc"), NewStringValue("d")) })
	assertSequenceError(t, func() (Value, error) { return Append(NewStringValue(""), IntZero) })
	assertSequenceError(t, func() (Value, error) {
		return Concat(NewTupleValue([]Value{NewStringValue("abc")}), NewStringValue("d"))
	})
	assertSequenceError(t, func() (Value, error) {
		return Concat(NewStringValue("abc"), NewTupleValue([]Value{NewStringValue("d")}))
	})
	assertSequenceError(t, func() (Value, error) {
		return Concat(NewTupleValue([]Value{NewStringValue("abc")}), IntOne)
	})
	assertSequenceError(t, func() (Value, error) {
		return Concat(IntOne, NewTupleValue([]Value{NewStringValue("d")}))
	})
	assertSequenceError(t, func() (Value, error) { return Concat(IntOne, IntOne) })
}

func assertStringValue(t *testing.T, value Value, want string) {
	t.Helper()
	stringValue, ok := value.(*StringValue)
	if !ok {
		t.Fatalf("value %T = %v, want StringValue", value, value)
	}
	if got := stringValue.Val.String(); got != want {
		t.Fatalf("string value = %q, want %q", got, want)
	}
}

func assertSequenceError(t *testing.T, fn func() (Value, error)) {
	t.Helper()
	if value, err := fn(); err == nil {
		t.Fatalf("operation returned value %v and nil error, want error", value)
	}
}
