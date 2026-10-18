package tlc

import (
	"fmt"
	"math"
	"strings"
)

var (
	NatValue    = NewUserValue(naturalsObj{})
	IntValueSet = NewUserValue(integersObj{})
)

func Nat() Value { return NatValue }
func Int() Value { return IntValueSet }

func NatPlus(x, y *IntValue) (*IntValue, error) {
	res := int64(x.ownedVal()) + int64(y.ownedVal())
	if res < math.MinInt32 || res > math.MaxInt32 {
		return nil, newTLCErrorCode(ECTLCModuleOverflow, fmt.Sprintf("%d+%d", x.ownedVal(), y.ownedVal()))
	}
	return NewIntValue(int32(res)), nil
}

func NatMinus(x, y *IntValue) (*IntValue, error) {
	res := int64(x.ownedVal()) - int64(y.ownedVal())
	if res < math.MinInt32 || res > math.MaxInt32 {
		return nil, newTLCErrorCode(ECTLCModuleOverflow, fmt.Sprintf("%d-%d", x.ownedVal(), y.ownedVal()))
	}
	return NewIntValue(int32(res)), nil
}

func NatTimes(x, y *IntValue) (*IntValue, error) {
	res := int64(x.ownedVal()) * int64(y.ownedVal())
	if res < math.MinInt32 || res > math.MaxInt32 {
		return nil, newTLCErrorCode(ECTLCModuleOverflow, fmt.Sprintf("%d*%d", x.ownedVal(), y.ownedVal()))
	}
	return NewIntValue(int32(res)), nil
}

func NatLT(x, y Value) (*BoolValue, error) {
	return intComparison("<", x, y, func(a, b int32) bool { return a < b })
}
func NatLE(x, y Value) (*BoolValue, error) {
	return intComparison("<=", x, y, func(a, b int32) bool { return a <= b })
}
func NatGT(x, y Value) (*BoolValue, error) {
	return intComparison(">", x, y, func(a, b int32) bool { return a > b })
}
func NatGEQ(x, y Value) (*BoolValue, error) {
	return intComparison(">", x, y, func(a, b int32) bool { return a >= b })
}

func IntGEQ(x, y Value) (*BoolValue, error) {
	return intComparison(">=", x, y, func(a, b int32) bool { return a >= b })
}

func DotDot(x, y *IntValue) *IntervalValue {
	return NewIntervalValue(x.ownedVal(), y.ownedVal())
}

func IntNeg(x *IntValue) (*IntValue, error) {
	if x.ownedVal() == math.MinInt32 {
		return nil, newTLCErrorCode(ECTLCModuleOverflow, "--2147483648")
	}
	return NewIntValue(-x.ownedVal()), nil
}

func NatDivide(x, y *IntValue) (*IntValue, error) {
	n1, n2 := numericIntValue(x), numericIntValue(y)
	if n2 == 0 {
		return nil, newTLCErrorCode(ECTLCModuleDivisionByZero)
	}
	q := n1 / n2
	if q < 0 && q*n2 != n1 {
		q--
	}
	return NewIntValue(q), nil
}

func IntDivide(x, y *IntValue) (*IntValue, error) {
	if y.ownedVal() == 0 {
		return nil, newTLCErrorCode(ECTLCModuleDivisionByZero)
	}
	if x.ownedVal() == math.MinInt32 && y.ownedVal() == -1 {
		return nil, newTLCErrorCode(ECTLCModuleOverflow, "-2147483648 \\div -1")
	}
	q := x.ownedVal() / y.ownedVal()
	if ((x.ownedVal() < 0 && y.ownedVal() > 0) || (x.ownedVal() > 0 && y.ownedVal() < 0)) && q*y.ownedVal() != x.ownedVal() {
		q--
	}
	return NewIntValue(q), nil
}

// Naturals reads both operands before validating the modulus. Integers checks
// the modulus first; retain distinct functions for these source boundaries.
func NatMod(x, y *IntValue) (*IntValue, error) {
	n1, n2 := numericIntValue(x), numericIntValue(y)
	if n2 <= 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "%", "positive number", fmt.Sprint(n2))
	}
	r := n1 % n2
	if r < 0 {
		r += n2
	}
	return NewIntValue(r), nil
}

func IntMod(x, y *IntValue) (*IntValue, error) {
	if y.ownedVal() <= 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "%", "positive number", y.String())
	}
	r := x.ownedVal() % y.ownedVal()
	if r < 0 {
		r += y.ownedVal()
	}
	return NewIntValue(r), nil
}

func NatExpt(x, y *IntValue) (*IntValue, error) {
	n1, n2 := numericIntValue(x), numericIntValue(y)
	return intExponentiation(n1, n2)
}

// The integer override checks the exponent before reading the base.
func IntExpt(x, y *IntValue) (*IntValue, error) {
	exponent := numericIntValue(y)
	if exponent < 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "^", "natural number", fmt.Sprint(exponent))
	}
	return intExponentiation(numericIntValue(x), exponent)
}

func intExponentiation(base, exponent int32) (*IntValue, error) {
	if exponent < 0 {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "^", "natural number", fmt.Sprint(exponent))
	}
	if exponent == 0 {
		if base == 0 {
			return nil, newTLCErrorCode(ECTLCModuleNullPowerNull)
		}
		return IntOne, nil
	}
	res := int64(base)
	for i := int32(1); i < exponent; i++ {
		res *= int64(base)
		if res < math.MinInt32 || res > math.MaxInt32 {
			return nil, newTLCErrorCode(ECTLCModuleOverflow, fmt.Sprintf("%d^%d", base, exponent))
		}
	}
	return NewIntValue(int32(res)), nil
}

func numericIntValue(value *IntValue) int32 {
	if value == nil {
		panic(NewNullPointerException())
	}
	return value.ownedVal()
}

func intComparison(op string, x, y Value, cmp func(int32, int32) bool) (*BoolValue, error) {
	ix, ok := x.(*IntValue)
	if !ok {
		if x == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentErrorAn, "first", op, "integer", ValuesPPR(x))
	}
	iy, ok := y.(*IntValue)
	if !ok {
		if y == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentErrorAn, "second", op, "integer", ValuesPPR(y))
	}
	return BoolValueOf(cmp(ix.ownedVal(), iy.ownedVal())), nil
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
	return 0, newTLCErrorCode(ECTLCModuleCompareValue, "Nat", ValuesPPR(val))
}

func (naturalsObj) Member(val Value) (bool, error) {
	if iv, ok := val.(*IntValue); ok {
		return iv.ownedVal() >= 0, nil
	}
	if mv, ok := val.(*ModelValue); ok {
		return mv.ModelValueMember(NatValue)
	}
	if val == nil {
		panic(NewNullPointerException())
	}
	return false, newTLCErrorCode(ECTLCModuleCheckMemberOf, ValuesPPR(val), "Nat")
}

func (naturalsObj) IsFinite() (bool, error) { return false, nil }
func (naturalsObj) IsEmpty() (bool, error)  { return false, nil }
func (naturalsObj) String() string          { return "Nat" }
func (naturalsObj) ToString(sb *strings.Builder, _ int, _ bool) *strings.Builder {
	sb.WriteString("Nat")
	return sb
}
func (naturalsObj) GetNonEnumerableErrorMsg(expr SemanticNode) string {
	return naturalsNonEnumerableErrorMsg(expr)
}

func naturalsNonEnumerableErrorMsg(expr SemanticNode) string {
	return "TLC encountered the non-enumerable quantifier bound\n" +
		ValuesPPR(NatValue) + "\n" + SemanticString(expr) + "\n" +
		"The set Nat contains infinitely many elements. As a result, TLC cannot evaluate expressions that\n" +
		"universally (\\A) or existentially (\\E) quantify over " + ValuesPPR(NatValue) + ", because this would require checking an\n" +
		"infinite number of cases. Note that TLC handles set membership like T \\subseteq Nat for any finite set T."
}

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
	return 0, newTLCErrorCode(ECTLCModuleCompareValue, "Int", ValuesPPR(val))
}

func (integersObj) Member(val Value) (bool, error) {
	if _, ok := val.(*IntValue); ok {
		return true, nil
	}
	if mv, ok := val.(*ModelValue); ok {
		return mv.ModelValueMember(IntValueSet)
	}
	if val == nil {
		panic(NewNullPointerException())
	}
	return false, newTLCErrorCode(ECTLCModuleCheckMemberOf, ValuesPPR(val), "Int")
}

func (integersObj) IsFinite() (bool, error) { return false, nil }
func (integersObj) IsEmpty() (bool, error)  { return false, nil }
func (integersObj) String() string          { return "Int" }
func (integersObj) ToString(sb *strings.Builder, _ int, _ bool) *strings.Builder {
	sb.WriteString("Int")
	return sb
}
func (integersObj) GetNonEnumerableErrorMsg(expr SemanticNode) string {
	return integersNonEnumerableErrorMsg(expr)
}

func integersNonEnumerableErrorMsg(expr SemanticNode) string {
	return "TLC encountered the non-enumerable quantifier bound\n" +
		ValuesPPR(IntValueSet) + "\n" + SemanticString(expr) + "\n" +
		"The set Int contains infinitely many elements. As a result, TLC cannot evaluate expressions that\n" +
		"universally (\\A) or existentially (\\E) quantify over " + ValuesPPR(IntValueSet) + ", because this would require checking an\n" +
		"infinite number of cases. Note that TLC handles set membership like T \\subseteq Int for any finite set T."
}
