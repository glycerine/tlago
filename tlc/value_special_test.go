package tlc

import (
	"strings"
	"testing"
)

func TestSpecialValuesCannotBeFingerprintedOrPermutedLikeJava(t *testing.T) {
	ModelValueInit()
	SetModelValues()
	for _, tc := range []struct {
		name string
		val  Value
		want string
	}{
		{name: "undef", val: ValUndef, want: "UNDEF"},
		{name: "user", val: AnySetValue, want: "ANY"},
	} {
		t.Run(tc.name+"/fingerprint", func(t *testing.T) {
			expectSpecialValuePanic(t, tc.name+" FingerPrint", tc.want, func() {
				_ = tc.val.FingerPrint(FP64New())
			})
		})
		t.Run(tc.name+"/permute", func(t *testing.T) {
			expectSpecialValuePanic(t, tc.name+" Permute", tc.want, func() {
				_ = tc.val.Permute(NewMVPerm())
			})
		})
	}
}

func expectSpecialValuePanic(t *testing.T, label string, want string, body func()) {
	t.Helper()
	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatalf("%s did not panic", label)
		}
		err, ok := recovered.(error)
		if !ok {
			t.Fatalf("%s panic = %v, want error", label, recovered)
		}
		msg := err.Error()
		if !strings.Contains(msg, "TLC has found a state in which the value of a variable contains") ||
			!strings.Contains(msg, want) {
			t.Fatalf("%s panic = %q", label, msg)
		}
	}()
	body()
}
