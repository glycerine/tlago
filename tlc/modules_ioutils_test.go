package tlc

import "testing"

func TestIOUtilsEmptyCommandMatchesJavaFailure(t *testing.T) {
	value, err := IOUtilsIOExec(EmptyTuple)
	if value != nil {
		t.Fatalf("IOExec empty command value = %v, want nil", value)
	}
	failure, ok := err.(*ArrayIndexOutOfBoundsException)
	if !ok || failure.Error() != "Index 0 out of bounds for length 0" {
		t.Fatalf("IOExec empty command failure = %T/%v, want ArrayIndexOutOfBoundsException", err, err)
	}
}
