package tlc

func DyadicRationalsReduce(value Value) (Value, error) {
	if isNil(value) {
		return nil, NewNullPointerException(`Cannot invoke "tlc2.value.impl.Value.toString()" because "<parameter1>" is null`)
	}
	record, ok := asConcreteRecordValue(value)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Half", "record", ValuesPPR(value))
	}
	record.Normalize()

	denValue, err := record.Apply(NewStringValue("den"))
	if err != nil {
		return nil, err
	}
	numValue, err := record.Apply(NewStringValue("num"))
	if err != nil {
		return nil, err
	}
	den, denOK := denValue.(*IntValue)
	if !isNil(denValue) && !denOK {
		return nil, valueStreamClassCast(denValue, "tlc2.value.impl.IntValue")
	}
	num, numOK := numValue.(*IntValue)
	if !isNil(numValue) && !numOK {
		return nil, valueStreamClassCast(numValue, "tlc2.value.impl.IntValue")
	}
	if isNil(num) {
		return nil, NewNullPointerException(`Cannot read field "val" because "<local3>" is null`)
	}
	if isNil(den) {
		return nil, NewNullPointerException(`Cannot read field "val" because "<local2>" is null`)
	}

	gcd, err := dyadicGCDInt32(num.ownedVal(), den.ownedVal())
	if err != nil {
		return nil, err
	}
	if gcd == 1 {
		return value, nil
	}
	if gcd == 0 {
		return nil, NewArithmeticException("/ by zero")
	}

	names := []*UniqueString{UniqueStringOf("den"), UniqueStringOf("num")}
	values := []Value{
		NewIntValue(int32(int64(den.ownedVal()) / gcd)),
		NewIntValue(int32(int64(num.ownedVal()) / gcd)),
	}
	return NewRecordValue(names, values, false), nil
}

func dyadicGCDInt32(a int32, b int32) (int64, error) {
	x := dyadicAbs64(int64(a))
	y := dyadicAbs64(int64(b))
	for y != 0 {
		x, y = y, x%y
	}
	if x > int64(^uint32(0)>>1) {
		message := "overflow: gcd(" + MessageNumberFormat(int64(a)) + ", " + MessageNumberFormat(int64(b)) + ") is 2^31"
		return 0, newDyadicMathArithmeticException(message)
	}
	return x, nil
}

func dyadicAbs64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

type dyadicMathArithmeticException struct {
	*ArithmeticException
}

func newDyadicMathArithmeticException(message string) *dyadicMathArithmeticException {
	return &dyadicMathArithmeticException{ArithmeticException: NewArithmeticException(message)}
}

func (*dyadicMathArithmeticException) diagnosticClassName() string {
	return "org.apache.commons.math3.exception.MathArithmeticException"
}
