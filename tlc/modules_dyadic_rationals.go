package tlc

func DyadicRationalsReduce(value Value) (Value, error) {
	record, ok := value.(*RecordValue)
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
	den, ok := denValue.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Half", "record", ValuesPPR(value))
	}
	num, ok := numValue.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "Half", "record", ValuesPPR(value))
	}

	gcd := dyadicGCDInt32(num.Val, den.Val)
	if gcd == 1 {
		return record, nil
	}
	if gcd == 0 {
		return nil, newTLCError(ECGeneral, "DyadicRationals!Reduce cannot reduce %s with gcd 0", value)
	}

	names := []*UniqueString{UniqueStringOf("den"), UniqueStringOf("num")}
	values := []Value{
		NewIntValue(int32(int64(den.Val) / gcd)),
		NewIntValue(int32(int64(num.Val) / gcd)),
	}
	return NewRecordValue(names, values, false), nil
}

func dyadicGCDInt32(a int32, b int32) int64 {
	x := dyadicAbs64(int64(a))
	y := dyadicAbs64(int64(b))
	for y != 0 {
		x, y = y, x%y
	}
	return x
}

func dyadicAbs64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
