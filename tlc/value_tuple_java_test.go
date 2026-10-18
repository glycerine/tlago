package tlc

import (
	"strings"
	"testing"
)

// The original catches TLCRuntimeException and checks contains only when thrown;
// it has no fail() after these try blocks. Other exceptions remain uncaught.
func javaTupleCaughtMessage(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		return
	}
	failure := javaRuntimeException(err)
	if failure == nil {
		t.Fatalf("uncaught %T: %v", err, err)
	}
	if !strings.Contains(failure.Error(), want) {
		t.Fatalf("message=%q, want contains %q", failure.Error(), want)
	}
}

func TestJavaTupleValue(t *testing.T) {
	t.Run("testErrorMessages", func(t *testing.T) {
		elems := []Value{NewStringValue("A")}
		tupVal := NewTupleValue(elems)
		_, err := tupVal.Apply(NewIntValue(2))
		javaTupleCaughtMessage(t, err, "Attempted to access index 2 of tuple\n<<\"A\">>\nwhich is out of bounds")
		_, err = tupVal.Apply(NewStringValue("a"))
		javaTupleCaughtMessage(t, err, "Attempted to access tuple at a non integral index: \"a\"")
		_, err = tupVal.Select(NewStringValue("a"))
		javaTupleCaughtMessage(t, err, "Attempted to access tuple at a non integral index: \"a\"")
		args := []Value{NewStringValue("arg1"), NewStringValue("arg2")}
		_, err = tupVal.ApplyArgs(args, EvalClear)
		javaTupleCaughtMessage(t, err, "Attempted to access tuple with 2 arguments when it expects 1.")
	})
}
