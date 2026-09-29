package tlc

import "math"

var (
	NatValue    = NewUserValue(naturalsObj{})
	IntValueSet = NewUserValue(integersObj{})
)

func Nat() Value { return NatValue }
func Int() Value { return IntValueSet }

func NatPlus(x, y *IntValue) (*IntValue, error) {
	res := int64(x.Val) + int64(y.Val)
	if res < math.MinInt32 || res > math.MaxInt32 {
		return nil, newTLCError(ECGeneral, "integer overflow in %d+%d", x.Val, y.Val)
	}
	return NewIntValue(int32(res)), nil
}

func NatMinus(x, y *IntValue) (*IntValue, error) {
	res := int64(x.Val) - int64(y.Val)
	if res < math.MinInt32 || res > math.MaxInt32 {
		return nil, newTLCError(ECGeneral, "integer overflow in %d-%d", x.Val, y.Val)
	}
	return NewIntValue(int32(res)), nil
}

func NatTimes(x, y *IntValue) (*IntValue, error) {
	res := int64(x.Val) * int64(y.Val)
	if res < math.MinInt32 || res > math.MaxInt32 {
		return nil, newTLCError(ECGeneral, "integer overflow in %d*%d", x.Val, y.Val)
	}
	return NewIntValue(int32(res)), nil
}

func NatLT(x, y Value) (*BoolValue, error) {
	return intComparison("<", x, y, func(a, b int32) bool { return a < b })
}
func NatLE(x, y Value) (*BoolValue, error) {
	return intComparison("\\leq", x, y, func(a, b int32) bool { return a <= b })
}
func NatGT(x, y Value) (*BoolValue, error) {
	return intComparison(">", x, y, func(a, b int32) bool { return a > b })
}
func NatGEQ(x, y Value) (*BoolValue, error) {
	return intComparison("\\geq", x, y, func(a, b int32) bool { return a >= b })
}

func DotDot(x, y *IntValue) *IntervalValue {
	return NewIntervalValue(x.Val, y.Val)
}

func IntNeg(x *IntValue) (*IntValue, error) {
	if x.Val == math.MinInt32 {
		return nil, newTLCError(ECGeneral, "integer overflow in --2147483648")
	}
	return NewIntValue(-x.Val), nil
}

func NatDivide(x, y *IntValue) (*IntValue, error) {
	if y.Val == 0 {
		return nil, newTLCError(ECGeneral, "division by zero")
	}
	q := x.Val / y.Val
	if q < 0 && q*y.Val != x.Val {
		q--
	}
	return NewIntValue(q), nil
}

func IntDivide(x, y *IntValue) (*IntValue, error) {
	if y.Val == 0 {
		return nil, newTLCError(ECGeneral, "division by zero")
	}
	if x.Val == math.MinInt32 && y.Val == -1 {
		return nil, newTLCError(ECGeneral, "integer overflow in -2147483648 \\div -1")
	}
	q := x.Val / y.Val
	if ((x.Val < 0 && y.Val > 0) || (x.Val > 0 && y.Val < 0)) && q*y.Val != x.Val {
		q--
	}
	return NewIntValue(q), nil
}

func NatMod(x, y *IntValue) (*IntValue, error) {
	if y.Val <= 0 {
		return nil, newTLCError(ECGeneral, "second argument of %% must be a positive number, got %s", y)
	}
	r := x.Val % y.Val
	if r < 0 {
		r += y.Val
	}
	return NewIntValue(r), nil
}

func NatExpt(x, y *IntValue) (*IntValue, error) {
	if y.Val < 0 {
		return nil, newTLCError(ECGeneral, "second argument of ^ must be a natural number, got %s", y)
	}
	if y.Val == 0 {
		if x.Val == 0 {
			return nil, newTLCError(ECGeneral, "0^0 is undefined")
		}
		return IntOne, nil
	}
	res := int64(x.Val)
	for i := int32(1); i < y.Val; i++ {
		res *= int64(x.Val)
		if res < math.MinInt32 || res > math.MaxInt32 {
			return nil, newTLCError(ECGeneral, "integer overflow in %d^%d", x.Val, y.Val)
		}
	}
	return NewIntValue(int32(res)), nil
}

func intComparison(op string, x, y Value, cmp func(int32, int32) bool) (*BoolValue, error) {
	ix, ok := x.(*IntValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "first argument of %s must be an integer, got %s", op, x)
	}
	iy, ok := y.(*IntValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "second argument of %s must be an integer, got %s", op, y)
	}
	return NewBoolValue(cmp(ix.Val, iy.Val)), nil
}

type naturalsObj struct{}

func (naturalsObj) Compare(val Value) (int, error) {
	if uv, ok := val.(*UserValue); ok {
		switch uv.UserObj.(type) {
		case naturalsObj:
			return 0, nil
		case integersObj:
			return -1, nil
		}
	}
	if _, ok := val.(*ModelValue); ok {
		return 1, nil
	}
	return 0, newTLCError(ECGeneral, "attempted to compare Nat with %s", val)
}

func (naturalsObj) Member(val Value) (bool, error) {
	if iv, ok := val.(*IntValue); ok {
		return iv.Val >= 0, nil
	}
	if mv, ok := val.(*ModelValue); ok {
		return mv.modelValueMember(NatValue)
	}
	return false, newTLCError(ECGeneral, "attempted to check if %s is in Nat", val)
}

func (naturalsObj) IsFinite() (bool, error) { return false, nil }
func (naturalsObj) IsEmpty() (bool, error)  { return false, nil }
func (naturalsObj) String() string          { return "Nat" }

type integersObj struct{}

func (integersObj) Compare(val Value) (int, error) {
	if uv, ok := val.(*UserValue); ok {
		switch uv.UserObj.(type) {
		case integersObj:
			return 0, nil
		case naturalsObj:
			return 1, nil
		}
	}
	if _, ok := val.(*ModelValue); ok {
		return 1, nil
	}
	return 0, newTLCError(ECGeneral, "attempted to compare Int with %s", val)
}

func (integersObj) Member(val Value) (bool, error) {
	if _, ok := val.(*IntValue); ok {
		return true, nil
	}
	if mv, ok := val.(*ModelValue); ok {
		return mv.modelValueMember(IntValueSet)
	}
	return false, newTLCError(ECGeneral, "attempted to check if %s is in Int", val)
}

func (integersObj) IsFinite() (bool, error) { return false, nil }
func (integersObj) IsEmpty() (bool, error)  { return false, nil }
func (integersObj) String() string          { return "Int" }
