package tlc

import (
	"math"
	"testing"
)

func TestPrimitiveValueJavaAccessors(t *testing.T) {
	if !BoolTrue.GetVal() || BoolFalse.GetVal() {
		t.Fatal("BoolValue.getVal did not return the stored value")
	}
	name := UniqueStringOf("accessor")
	if got := NewStringValueFromUnique(name).GetVal(); got != name {
		t.Fatal("StringValue.getVal did not return the stored UniqueString")
	}
	if got := NewStringValueFromUnique(nil).GetVal(); got != nil {
		t.Fatalf("StringValue.getVal = %v, want nil", got)
	}

	for _, call := range []func(){
		func() { _ = (*BoolValue)(nil).GetVal() },
		func() { _ = (*StringValue)(nil).GetVal() },
	} {
		func() {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("nil accessor panic = %T, want *NullPointerException", failure)
				}
			}()
			call()
		}()
	}
}

func TestIntValueJavaSum(t *testing.T) {
	if got := SumIntValues(IntOne, IntOne); got != NewIntValue(2) {
		t.Fatal("sum did not return the cached IntValue")
	}
	if got := SumIntValues(NewIntValue(math.MaxInt32), IntOne).Val; got != math.MinInt32 {
		t.Fatalf("overflowing sum = %d, want %d", got, int32(math.MinInt32))
	}

	for _, tc := range []struct {
		name string
		call func()
		want string
	}{
		{"left", func() { SumIntValues(nil, IntOne) }, `Cannot read field "val" because "n" is null`},
		{"right", func() { SumIntValues(IntOne, nil) }, `Cannot read field "val" because "m" is null`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				failure, ok := recover().(*NullPointerException)
				if !ok || failure.Error() != tc.want {
					t.Fatalf("panic = %T %v, want NullPointerException %q", failure, failure, tc.want)
				}
			}()
			tc.call()
		})
	}
}
