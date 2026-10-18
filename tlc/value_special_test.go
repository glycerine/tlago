package tlc

import (
	"strings"
	"testing"
)

type countingNonEnumerableUserObj struct {
	calls int
}

func (o *countingNonEnumerableUserObj) Compare(Value) (int, error) { return 0, nil }
func (o *countingNonEnumerableUserObj) Member(Value) (bool, error) { return false, nil }
func (o *countingNonEnumerableUserObj) IsFinite() (bool, error)    { return false, nil }
func (o *countingNonEnumerableUserObj) String() string             { return "Counting" }
func (o *countingNonEnumerableUserObj) GetNonEnumerableErrorMsg(SemanticNode) string {
	o.calls++
	return "custom bound diagnostic"
}

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

func TestSpecialValueJavaNullComparison(t *testing.T) {
	var typedUndef *UndefValue
	for _, other := range []Value{nil, typedUndef} {
		if got, err := ValUndef.Compare(other); err != nil || got != 1 {
			t.Fatalf("UNDEF.compareTo(null) = %d, %v; want 1, nil", got, err)
		}
		if got, err := ValUndef.Equal(other); err != nil || got {
			t.Fatalf("UNDEF.equals(null) = %t, %v; want false, nil", got, err)
		}
	}

	var typedUser *UserValue
	var typedModel *ModelValue
	for _, tc := range []struct {
		name  string
		other Value
	}{
		{"null", nil},
		{"typed_user", typedUser},
		{"typed_model", typedModel},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, compare := range []func(){
				func() { _, _ = AnySetValue.Compare(tc.other) },
				func() { _, _ = AnySetValue.Equal(tc.other) },
			} {
				func() {
					defer func() {
						failure, ok := recover().(*NullPointerException)
						want := `Cannot invoke "Object.toString()" because "obj" is null`
						if !ok || failure.Error() != want {
							t.Fatalf("panic = %T %v, want NullPointerException %q", failure, failure, want)
						}
					}()
					compare()
				}()
			}
		})
	}
}

func TestUserValueGetNonEnumerableErrorMsgUsesUserObjOverride(t *testing.T) {
	obj := &countingNonEnumerableUserObj{}
	value := NewUserValue(obj)
	if got := value.GetNonEnumerableErrorMsg(nil); got != "custom bound diagnostic" {
		t.Fatalf("GetNonEnumerableErrorMsg = %q, want custom bound diagnostic", got)
	}
	if obj.calls != 2 {
		t.Fatalf("UserObj diagnostic calls = %d, want 2", obj.calls)
	}
	obj.calls = 0
	if got := ValueGetNonEnumerableErrorMsg(value, nil); got != "custom bound diagnostic" {
		t.Fatalf("evaluator diagnostic = %q, want custom bound diagnostic", got)
	}
	if obj.calls != 2 {
		t.Fatalf("evaluator UserObj diagnostic calls = %d, want 2", obj.calls)
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
