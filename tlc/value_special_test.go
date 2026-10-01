package tlc

import (
	"strings"
	"testing"
)

func TestSpecialValuesCannotBeFingerprintedLikeJava(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  Value
		want string
	}{
		{name: "undef", val: ValUndef, want: "UNDEF"},
		{name: "user", val: AnySetValue, want: "ANY"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					t.Fatalf("%s FingerPrint did not panic", tc.name)
				}
				err, ok := recovered.(error)
				if !ok {
					t.Fatalf("%s FingerPrint panic = %v, want error", tc.name, recovered)
				}
				msg := err.Error()
				if !strings.Contains(msg, "TLC has found a state in which the value of a variable contains") ||
					!strings.Contains(msg, tc.want) {
					t.Fatalf("%s FingerPrint panic = %q", tc.name, msg)
				}
			}()
			_ = tc.val.FingerPrint(FP64New())
		})
	}
}
