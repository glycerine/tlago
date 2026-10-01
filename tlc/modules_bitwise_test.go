package tlc

import "testing"

func TestBitwiseCommunityModuleOverrides(t *testing.T) {
	andValue, err := BitwiseAnd(NewIntValue(6), NewIntValue(3), IntZero, NewIntValue(6))
	if err != nil || andValue.(*IntValue).Val != 2 {
		t.Fatalf("And(6,3,0,6) = %v/%v, want 2/nil", andValue, err)
	}
	orValue, err := BitwiseOr(NewIntValue(4), NewIntValue(1), IntZero, NewIntValue(4))
	if err != nil || orValue.(*IntValue).Val != 5 {
		t.Fatalf("Or(4,1,0,4) = %v/%v, want 5/nil", orValue, err)
	}
	xorValue, err := BitwiseXor(NewIntValue(7), NewIntValue(3), IntZero, NewIntValue(7))
	if err != nil || xorValue.(*IntValue).Val != 4 {
		t.Fatalf("Xor(7,3,0,7) = %v/%v, want 4/nil", xorValue, err)
	}
	notValue, err := BitwiseNot(NewIntValue(5))
	if err != nil || notValue.(*IntValue).Val != 2 {
		t.Fatalf("Not(5) = %v/%v, want 2/nil", notValue, err)
	}
	notZero, err := BitwiseNot(IntZero)
	if err != nil || notZero.(*IntValue).Val != 0 {
		t.Fatalf("Not(0) = %v/%v, want 0/nil", notZero, err)
	}
	shifted, err := BitwiseShiftR(NewIntValue(-1), IntOne)
	if err != nil || shifted.(*IntValue).Val != 2147483647 {
		t.Fatalf("shiftR(-1,1) = %v/%v, want 2147483647/nil", shifted, err)
	}
	wrappedShift, err := BitwiseShiftR(NewIntValue(8), NewIntValue(33))
	if err != nil || wrappedShift.(*IntValue).Val != 4 {
		t.Fatalf("shiftR(8,33) = %v/%v, want 4/nil", wrappedShift, err)
	}
	if value, err := BitwiseAnd(NewStringValue("x"), IntOne, IntZero, IntOne); err == nil {
		t.Fatalf("And(\"x\",1,0,1) = %v/nil, want argument error", value)
	}
}
