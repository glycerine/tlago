package tlc

import "math/bits"

func BitwiseAnd(x Value, y Value, n Value, m Value) (Value, error) {
	xv, err := bitwiseIntArg("first", "And", x)
	if err != nil {
		return nil, err
	}
	yv, err := bitwiseIntArg("second", "And", y)
	if err != nil {
		return nil, err
	}
	if _, err := bitwiseIntArg("third", "And", n); err != nil {
		return nil, err
	}
	if _, err := bitwiseIntArg("fourth", "And", m); err != nil {
		return nil, err
	}
	return NewIntValue(xv.Val & yv.Val), nil
}

func BitwiseOr(x Value, y Value, n Value, m Value) (Value, error) {
	xv, err := bitwiseIntArg("first", "Or", x)
	if err != nil {
		return nil, err
	}
	yv, err := bitwiseIntArg("second", "Or", y)
	if err != nil {
		return nil, err
	}
	if _, err := bitwiseIntArg("third", "Or", n); err != nil {
		return nil, err
	}
	if _, err := bitwiseIntArg("fourth", "Or", m); err != nil {
		return nil, err
	}
	return NewIntValue(xv.Val | yv.Val), nil
}

func BitwiseXor(x Value, y Value, n Value, m Value) (Value, error) {
	xv, err := bitwiseIntArg("first", "Xor", x)
	if err != nil {
		return nil, err
	}
	yv, err := bitwiseIntArg("second", "Xor", y)
	if err != nil {
		return nil, err
	}
	if _, err := bitwiseIntArg("third", "Xor", n); err != nil {
		return nil, err
	}
	if _, err := bitwiseIntArg("fourth", "Xor", m); err != nil {
		return nil, err
	}
	return NewIntValue(xv.Val ^ yv.Val), nil
}

func BitwiseNot(x Value) (Value, error) {
	xv, err := bitwiseIntArg("first", "Not", x)
	if err != nil {
		return nil, err
	}
	shift := uint((bits.LeadingZeros32(uint32(xv.Val)) - 1) & 31)
	mask := int32(0x7fffffff >> shift)
	return NewIntValue(^xv.Val & mask), nil
}

func BitwiseShiftR(n Value, pos Value) (Value, error) {
	nv, err := bitwiseIntArg("first", "shiftR", n)
	if err != nil {
		return nil, err
	}
	pv, err := bitwiseIntArg("second", "shiftR", pos)
	if err != nil {
		return nil, err
	}
	return NewIntValue(int32(uint32(nv.Val) >> uint(pv.Val&31))), nil
}

func bitwiseIntArg(position string, operator string, value Value) (*IntValue, error) {
	intValue, ok := value.(*IntValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, position, operator, "integer", ValuesPPR(value))
	}
	return intValue, nil
}
