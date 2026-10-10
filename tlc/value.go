package tlc

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
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
	GetCostModel() CostModel
}

func InitializeValue(value Value) Value {
	if value == nil {
		return nil
	}
	if operator, ok := value.(*OpRcdValue); ok {
		// OpRcdValue.initialize initializes each argument and output; the
		// operator itself has no normalization or fingerprint definition.
		for i, arguments := range operator.Domain {
			if arguments == nil {
				panic(NewNullPointerException())
			}
			for j, argument := range arguments {
				if argument == nil {
					panic(NewNullPointerException())
				}
				operator.Domain[i][j] = InitializeValue(argument)
			}
			if operator.Values[i] == nil {
				panic(NewNullPointerException())
			}
			operator.Values[i] = InitializeValue(operator.Values[i])
		}
		return operator
	}
	value.DeepNormalize()
	switch value.(type) {
	case *MethodValue, *EvaluatingValue, *PriorityEvaluatingValue, *CallableValue:
		// MethodValue/EvaluatingValue.initialize omit unsupported fingerprinting.
		return value
	}
	value.FingerPrint(0)
	return value
}

func ValueJavaHashCode(value Value) int32 {
	defer catchValueFailure(value, nil)
	if value == nil {
		return 0
	}
	fp := value.FingerPrint(FP64New())
	return int32(uint32(fp>>32) ^ uint32(fp))
}

// catchValueFailure is deferred at Java's RuntimeException/OutOfMemoryError
// catch boundaries. Returned errors represent Java throws in Go; pointer-only
// and fingerprint APIs retain panics. An unsourced value leaves failures intact.
func catchValueFailure(value Value, err *error) {
	if failure := recover(); failure != nil {
		if valueSource(value) != nil {
			cause, ok := failure.(error)
			if !ok {
				cause = fmt.Errorf("%v", failure)
			}
			panic(NewFingerprintExceptionHead(value, cause))
		}
		panic(failure)
	}
	if err != nil {
		*err = wrapValueFailure(value, *err)
	}
}

func wrapValueFailure(value Value, err error) error {
	if err != nil && valueSource(value) != nil {
		return NewFingerprintExceptionHead(value, err)
	}
	return err
}

type ValueExcept struct {
	Path  []Value
	Index int
	Value Value
}

type BaseValue struct {
	source SemanticNode
	CM     CostModel
}

// Java reference reads/writes are atomic. Protect the Go interface's two words
// when worker-local lazy values attach source metadata to shared cached values.
var valueSourceMutex sync.RWMutex

func newBaseValue(cms ...CostModel) BaseValue {
	if len(cms) > 0 {
		return BaseValue{CM: cms[0]}
	}
	return BaseValue{}
}

func (v BaseValue) GetCostModel() CostModel { return v.CM }

func (v *BaseValue) SetCostModel(cm CostModel) { v.CM = cm }

func (BaseValue) KindStringFor(kind ValueKind) string {
	if int(kind) >= 0 && int(kind) < len(valueKindImage) {
		return valueKindImage[kind]
	}
	return "an unknown value"
}

func (BaseValue) unsupported(format string, args ...any) error {
	return newTLCError(ECGeneral, format, args...)
}

// runtimeFailure retains Assert.fail(reason, getSource()) metadata before
// the enclosing value operation wraps the failure in a FingerprintException.
func (v *BaseValue) runtimeFailure(message string) error {
	if source := v.GetSource(); source != nil {
		return NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
	}
	return NewTLCRuntimeExceptionMessage(message)
}

func (v *BaseValue) SetSource(source SemanticNode) {
	if v != nil {
		valueSourceMutex.Lock()
		v.source = source
		valueSourceMutex.Unlock()
	}
}

func (v *BaseValue) GetSource() SemanticNode {
	if v == nil {
		return nil
	}
	valueSourceMutex.RLock()
	source := v.source
	valueSourceMutex.RUnlock()
	return source
}

func (v *BaseValue) HasSource() bool {
	return v.GetSource() != nil
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

func (v *BoolValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	o, ok := other.(*BoolValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare boolean %s with non-boolean:\n%s", ValuesPPR(v), ValuesPPR(other)))
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

func (v *BoolValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	o, ok := other.(*BoolValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to compare equality of boolean %s with non-boolean:\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return v.Val == o.Val, nil
}

func (v *BoolValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of the boolean %s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *BoolValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the boolean %s is a finite set.", ValuesPPR(v)))
}

func (v *BoolValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compute the number of elements in the boolean %s.", ValuesPPR(v)))
}

func (v *BoolValue) Normalize() Value      { return v }
func (v *BoolValue) DeepNormalize()        {}
func (v *BoolValue) IsNormalized() bool    { return true }
func (v *BoolValue) IsDefined() bool       { return true }
func (v *BoolValue) DeepCopy() Value       { return v }
func (v *BoolValue) Permute(*MVPerm) Value { return v }

func (v *BoolValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	fp = FP64ExtendByte(fp, byte(BoolValueKind))
	if v.Val {
		return FP64ExtendUTF16(fp, []uint16{'t'})
	}
	return FP64ExtendUTF16(fp, []uint16{'f'})
}

func (v *BoolValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the boolean %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *BoolValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the boolean %s.", ValuesPPR(v)))
	}
	return v, nil
}

func (v *BoolValue) String() string {
	return ValueToString(v, "", true)
}

func (v *BoolValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if v.Val {
		sb.WriteString("TRUE")
	} else {
		sb.WriteString("FALSE")
	}
	return sb
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

// Recreate IntValue's class statics when starting an isolated Java test loader.
// Cached values can retain CallStackTool source metadata from an earlier run.
func InitializeIntValueStatics() {
	for i := range intValueCache {
		intValueCache[i] = &IntValue{Val: int32(i)}
	}
	IntNegOne, IntOne, IntZero = NewIntValue(-1), NewIntValue(1), NewIntValue(0)
}

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
func (v *IntValue) NBits() int {
	defer catchValueFailure(v, nil)
	return IntValueNBits(v.Val)
}

func (v *IntValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	o, ok := other.(*IntValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare integer %s with non-integer:\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	if v.Val < o.Val {
		return -1, nil
	}
	if v.Val > o.Val {
		return 1, nil
	}
	return 0, nil
}

func (v *IntValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	o, ok := other.(*IntValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of integer %s with non-integer:\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return v.Val == o.Val, nil
}

func (v *IntValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure("Attempted to check if the value:\n" + ValuesPPR(elem) + "\nis an element of the integer " + ValuesPPR(v))
}

func (v *IntValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the integer %s is a finite set.", ValuesPPR(v)))
}

func (v *IntValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compute the number of elements in the integer %s.", ValuesPPR(v)))
}

func (v *IntValue) Normalize() Value      { return v }
func (v *IntValue) DeepNormalize()        {}
func (v *IntValue) IsNormalized() bool    { return true }
func (v *IntValue) IsDefined() bool       { return true }
func (v *IntValue) DeepCopy() Value       { return v }
func (v *IntValue) Permute(*MVPerm) Value { return v }

func (v *IntValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	return FP64ExtendInt(FP64ExtendByte(fp, byte(IntValueKind)), v.Val)
}

func (v *IntValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the integer %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *IntValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the integer %s.", ValuesPPR(v)))
	}
	return v, nil
}

func (v *IntValue) String() string {
	return ValueToString(v, "", true)
}

func (v *IntValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	sb.WriteString(strconv.FormatInt(int64(v.Val), 10))
	return sb
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

func NewStringValue(s string, cms ...CostModel) *StringValue {
	return &StringValue{BaseValue: newBaseValue(cms...), Val: UniqueStringOf(s)}
}

func NewStringValueFromUnique(s *UniqueString, cms ...CostModel) *StringValue {
	return &StringValue{BaseValue: newBaseValue(cms...), Val: s}
}

func (v *StringValue) Kind() ValueKind    { return StringValueKind }
func (v *StringValue) KindString() string { return v.KindStringFor(v.Kind()) }
func (v *StringValue) Length() int {
	defer catchValueFailure(v, nil)
	return v.Val.Length()
}

func (v *StringValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	if pending, ok := other.(*DebuggerValue); ok {
		other = pending.StringValue
	}
	o, ok := other.(*StringValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueCompareTo(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare string %s with non-string:\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return v.Val.Compare(o.Val), nil
}

func (v *StringValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if pending, ok := other.(*DebuggerValue); ok {
		other = pending.StringValue
	}
	o, ok := other.(*StringValue)
	if !ok {
		if mv, ok := other.(*ModelValue); ok {
			return mv.modelValueEquals(v)
		}
		if other == nil {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of string %s with non-string:\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return v.Val.Equal(o.Val), nil
}

func (v *StringValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of the string %s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *StringValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the string %s is a finite set.", ValuesPPR(v)))
}

func (v *StringValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compute the number of elements in the string %s.", ValuesPPR(v)))
}

func (v *StringValue) Normalize() Value      { return v }
func (v *StringValue) DeepNormalize()        {}
func (v *StringValue) IsNormalized() bool    { return true }
func (v *StringValue) IsDefined() bool       { return true }
func (v *StringValue) DeepCopy() Value       { return v }
func (v *StringValue) Permute(*MVPerm) Value { return v }

func (v *StringValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v, nil)
	fp = FP64ExtendByte(fp, byte(StringValueKind))
	fp = FP64ExtendInt(fp, int32(v.Val.Length()))
	return FP64ExtendString(fp, v.Val.String())
}

func (v *StringValue) TakeExcept(ex ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the string %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *StringValue) TakeExcepts(exs []ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the string %s.", ValuesPPR(v)))
	}
	return v, nil
}

func (v *StringValue) String() string {
	return ValueToString(v, "", true)
}

func (v *StringValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	sb.WriteString(`"` + tlaStringPrintVersion(v.Val.String()) + `"`)
	return sb
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
	var escaped []uint16
	for _, char := range javaStringUTF16(s) {
		switch char {
		case '"':
			escaped = append(escaped, '\\', '"')
		case '\\':
			escaped = append(escaped, '\\', '\\')
		case '\t':
			escaped = append(escaped, '\\', 't')
		case '\n':
			escaped = append(escaped, '\\', 'n')
		case '\f':
			escaped = append(escaped, '\\', 'f')
		case '\r':
			escaped = append(escaped, '\\', 'r')
		default:
			escaped = append(escaped, char)
		}
	}
	return javaStringFromUTF16(escaped)
}
