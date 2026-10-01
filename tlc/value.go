package tlc

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
)

type ValueKind byte

const (
	BoolValueKind ValueKind = iota
	IntValueKind
	RealValueKind
	StringValueKind
	RecordValueKind
	SetEnumValueKind
	SetPredValueKind
	TupleValueKind
	FcnLambdaValueKind
	FcnRcdValueKind
	OpLambdaValueKind
	OpRcdValueKind
	MethodValueKind
	SetOfFcnsValueKind
	SetOfRcdsValueKind
	SetOfTuplesValueKind
	SubsetValueKind
	SetDiffValueKind
	SetCapValueKind
	SetCupValueKind
	UnionValueKind
	ModelValueKind
	UserValueKind
	IntervalValueKind
	UndefValueKind
	LazyValueKind
	DummyValueKind
)

var valueKindImage = []string{
	"a Boolean value",
	"an integer",
	"a real",
	"a string",
	"a record",
	"a set of the form {e1, ... ,eN}",
	"a set of the form {x \\in S : expr}",
	"a tuple",
	"a function of the form  [x \\in S |-> expr]",
	"a function  of the form (d1 :> e1 @@ ... @@ dN :> eN)",
	"an operator",
	"a constant operator",
	"a java method",
	"a set of the form [S -> T]",
	"a set of the form [d1 : S1, ... , dN : SN]",
	"a cartesian product",
	"a set of the form SUBSET S",
	"a set of the form S \\ T",
	"a set of the form S \\cap T",
	"a set of the form S \\cup T",
	"a set of the form UNION  S",
	"a model value",
	"a special set constant",
	"a set of the form i..j",
	"an undefined value",
	"a value represented in lazy form",
	"a dummy for not-a-value",
}

type Value interface {
	fmt.Stringer
	Kind() ValueKind
	KindString() string
	Compare(Value) (int, error)
	Equal(Value) (bool, error)
	Member(Value) (bool, error)
	IsFinite() (bool, error)
	Size() (int, error)
	Normalize() Value
	DeepNormalize()
	IsNormalized() bool
	IsDefined() bool
	DeepCopy() Value
	FingerPrint(uint64) uint64
	Permute(*MVPerm) Value
	TakeExcept(ValueExcept) (Value, error)
	TakeExcepts([]ValueExcept) (Value, error)
}

func InitializeValue(value Value) Value {
	if value == nil {
		return nil
	}
	value.DeepNormalize()
	value.FingerPrint(0)
	return value
}

func ValueJavaHashCode(value Value) int32 {
	if value == nil {
		return 0
	}
	fp := value.FingerPrint(FP64New())
	return int32(uint32(fp>>32) ^ uint32(fp))
}

type ValueExcept struct {
	Path  []Value
	Index int
	Value Value
}

type BaseValue struct {
	source SemanticNode
}

func (BaseValue) KindStringFor(kind ValueKind) string {
	if int(kind) >= 0 && int(kind) < len(valueKindImage) {
		return valueKindImage[kind]
	}
	return "an unknown value"
}

func (BaseValue) unsupported(format string, args ...any) error {
	return newTLCError(ECGeneral, format, args...)
}

func (v *BaseValue) SetSource(source SemanticNode) {
	if v != nil {
		v.source = source
	}
}

func (v *BaseValue) GetSource() SemanticNode {
	if v == nil {
		return nil
	}
	return v.source
}

func (v *BaseValue) HasSource() bool {
	return v != nil && v.source != nil
}

type BoolValue struct {
	BaseValue
	Val bool
}

var (
	BoolFalse = &BoolValue{Val: false}
	BoolTrue  = &BoolValue{Val: true}
)

func NewBoolValue(v bool) *BoolValue {
	if v {
		return BoolTrue
	}
	return BoolFalse
}

func (v *BoolValue) Kind() ValueKind    { return BoolValueKind }
func (v *BoolValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *BoolValue) Compare(other Value) (int, error) {
	o, ok := other.(*BoolValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		return 0, v.unsupported("Attempted to compare boolean %s with non-boolean:\n%s", v, other)
	}
	x, y := 0, 0
	if v.Val {
		x = 1
	}
	if o.Val {
		y = 1
	}
	return x - y, nil
}

func (v *BoolValue) Equal(other Value) (bool, error) {
	o, ok := other.(*BoolValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		return false, v.unsupported("Attempted to compare equality of boolean %s with non-boolean:\n%s", v, other)
	}
	return v.Val == o.Val, nil
}

func (v *BoolValue) Member(elem Value) (bool, error) {
	return false, v.unsupported("Attempted to check if the value:\n%s\nis an element of the boolean %s", elem, v)
}

func (v *BoolValue) IsFinite() (bool, error) {
	return false, v.unsupported("Attempted to check if the boolean %s is a finite set.", v)
}

func (v *BoolValue) Size() (int, error) {
	return 0, v.unsupported("Attempted to compute the number of elements in the boolean %s.", v)
}

func (v *BoolValue) Normalize() Value      { return v }
func (v *BoolValue) DeepNormalize()        {}
func (v *BoolValue) IsNormalized() bool    { return true }
func (v *BoolValue) IsDefined() bool       { return true }
func (v *BoolValue) DeepCopy() Value       { return v }
func (v *BoolValue) Permute(*MVPerm) Value { return v }

func (v *BoolValue) FingerPrint(fp uint64) uint64 {
	fp = FP64ExtendInt(fp, int32(BoolValueKind))
	if v.Val {
		return FP64ExtendUTF16(fp, []uint16{'t'})
	}
	return FP64ExtendUTF16(fp, []uint16{'f'})
}

func (v *BoolValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, v.unsupported("Attempted to apply EXCEPT construct to the boolean %s.", v)
	}
	return ex.Value, nil
}

func (v *BoolValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, v.unsupported("Attempted to apply EXCEPT construct to the boolean %s.", v)
	}
	return v, nil
}

func (v *BoolValue) String() string {
	if v.Val {
		return "TRUE"
	}
	return "FALSE"
}

type IntValue struct {
	BaseValue
	Val int32
}

var intValueCache = func() []*IntValue {
	out := make([]*IntValue, 10)
	for i := range out {
		out[i] = &IntValue{Val: int32(i)}
	}
	return out
}()

var (
	IntNegOne = NewIntValue(-1)
	IntOne    = NewIntValue(1)
	IntZero   = NewIntValue(0)
)

func NewIntValue(v int32) *IntValue {
	if v >= 0 && int(v) < len(intValueCache) {
		return intValueCache[v]
	}
	return &IntValue{Val: v}
}

func IntValueNBits(tmp int32) int {
	nb := 0
	for tmp != 0 && tmp != -1 {
		nb++
		tmp >>= 1
	}
	return nb + 1
}

func SumIntValues(a, b *IntValue) *IntValue {
	return NewIntValue(a.Val + b.Val)
}

func (v *IntValue) Kind() ValueKind    { return IntValueKind }
func (v *IntValue) KindString() string { return v.KindStringFor(v.Kind()) }
func (v *IntValue) NBits() int         { return IntValueNBits(v.Val) }

func (v *IntValue) Compare(other Value) (int, error) {
	o, ok := other.(*IntValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		return 0, v.unsupported("Attempted to compare integer %s with non-integer:\n%s", v, other)
	}
	if v.Val < o.Val {
		return -1, nil
	}
	if v.Val > o.Val {
		return 1, nil
	}
	return 0, nil
}

func (v *IntValue) Equal(other Value) (bool, error) {
	o, ok := other.(*IntValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		return false, v.unsupported("Attempted to check equality of integer %s with non-integer:\n%s", v, other)
	}
	return v.Val == o.Val, nil
}

func (v *IntValue) Member(elem Value) (bool, error) {
	return false, v.unsupported("Attempted to check if the value:\n%s\nis an element of the integer %s", elem, v)
}

func (v *IntValue) IsFinite() (bool, error) {
	return false, v.unsupported("Attempted to check if the integer %s is a finite set.", v)
}

func (v *IntValue) Size() (int, error) {
	return 0, v.unsupported("Attempted to compute the number of elements in the integer %s.", v)
}

func (v *IntValue) Normalize() Value      { return v }
func (v *IntValue) DeepNormalize()        {}
func (v *IntValue) IsNormalized() bool    { return true }
func (v *IntValue) IsDefined() bool       { return true }
func (v *IntValue) DeepCopy() Value       { return v }
func (v *IntValue) Permute(*MVPerm) Value { return v }

func (v *IntValue) FingerPrint(fp uint64) uint64 {
	return FP64ExtendInt(FP64ExtendInt(fp, int32(IntValueKind)), v.Val)
}

func (v *IntValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, v.unsupported("Attempted to apply EXCEPT construct to the integer %s.", v)
	}
	return ex.Value, nil
}

func (v *IntValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, v.unsupported("Attempted to apply EXCEPT construct to the integer %s.", v)
	}
	return v, nil
}

func (v *IntValue) String() string {
	return strconv.FormatInt(int64(v.Val), 10)
}

func NarrowToIntValue(value int64) Value {
	if int64(int32(value)) != value {
		return IntNegOne
	}
	return NewIntValue(int32(value))
}

type StringValue struct {
	BaseValue
	Val *UniqueString
}

func NewStringValue(s string) *StringValue {
	return &StringValue{Val: UniqueStringOf(s)}
}

func NewStringValueFromUnique(s *UniqueString) *StringValue {
	return &StringValue{Val: s}
}

func (v *StringValue) Kind() ValueKind    { return StringValueKind }
func (v *StringValue) KindString() string { return v.KindStringFor(v.Kind()) }
func (v *StringValue) Length() int        { return v.Val.Length() }

func (v *StringValue) Compare(other Value) (int, error) {
	o, ok := other.(*StringValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		return 0, v.unsupported("Attempted to compare string %s with non-string:\n%s", v, other)
	}
	return v.Val.Compare(o.Val), nil
}

func (v *StringValue) Equal(other Value) (bool, error) {
	o, ok := other.(*StringValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		return false, v.unsupported("Attempted to check equality of string %s with non-string:\n%s", v, other)
	}
	return v.Val.Equal(o.Val), nil
}

func (v *StringValue) Member(elem Value) (bool, error) {
	return false, v.unsupported("Attempted to check if the value:\n%s\nis an element of the string %s", elem, v)
}

func (v *StringValue) IsFinite() (bool, error) {
	return false, v.unsupported("Attempted to check if the string %s is a finite set.", v)
}

func (v *StringValue) Size() (int, error) {
	return 0, v.unsupported("Attempted to compute the number of elements in the string %s.", v)
}

func (v *StringValue) Normalize() Value      { return v }
func (v *StringValue) DeepNormalize()        {}
func (v *StringValue) IsNormalized() bool    { return true }
func (v *StringValue) IsDefined() bool       { return true }
func (v *StringValue) DeepCopy() Value       { return v }
func (v *StringValue) Permute(*MVPerm) Value { return v }

func (v *StringValue) FingerPrint(fp uint64) uint64 {
	fp = FP64ExtendInt(fp, int32(StringValueKind))
	fp = FP64ExtendInt(fp, int32(v.Val.Length()))
	return FP64ExtendString(fp, v.Val.String())
}

func (v *StringValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, v.unsupported("Attempted to apply EXCEPT construct to the string %s.", v)
	}
	return ex.Value, nil
}

func (v *StringValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, v.unsupported("Attempted to apply EXCEPT construct to the string %s.", v)
	}
	return v, nil
}

func (v *StringValue) String() string {
	return `"` + tlaStringPrintVersion(v.Val.String()) + `"`
}

func (v *StringValue) UnquotedString() string {
	return tlaStringPrintVersion(v.Val.String())
}

func (v *StringValue) RawString() string {
	if v == nil || v.Val == nil {
		return ""
	}
	return v.Val.String()
}

func tlaStringPrintVersion(s string) string {
	var b strings.Builder
	for _, c := range utf16.Encode([]rune(s)) {
		switch c {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\f':
			b.WriteString(`\f`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(rune(c))
		}
	}
	return b.String()
}
