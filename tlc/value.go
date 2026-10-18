package tlc

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type ValueKind byte

var valueCoverageStatics struct {
	sync.Once
	enabled bool
}

func valueCoverageEnabled() bool {
	valueCoverageStatics.Do(func() {
		valueCoverageStatics.enabled = CoverageEnabled()
	})
	return valueCoverageStatics.enabled
}

// InitializeValueStatics represents a fresh Java Value class initialization.
// It requires an idle runtime, like the concrete value-static reset helpers.
func InitializeValueStatics() {
	valueCoverageStatics.Once = sync.Once{}
	valueCoverageStatics.enabled = false
}

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
	TakeExcept(*ValueExcept) (Value, error)
	TakeExcepts([]*ValueExcept) (Value, error)
	GetCostModel() CostModel
}

// ValueSetCostModel represents IValue.setCostModel and returns the receiver.
func ValueSetCostModel(value Value, cm CostModel) Value {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if setter, ok := value.(interface{ SetCostModel(CostModel) }); ok {
		setter.SetCostModel(cm)
	}
	return value
}

// ValueSetSource represents IValue.setSource.
func ValueSetSource(value Value, source SemanticNode) {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if setter, ok := value.(interface{ SetSource(SemanticNode) }); ok {
		setter.SetSource(source)
	}
}

// ValueGetSource represents IValue.getSource.
func ValueGetSource(value Value) SemanticNode {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if sourced, ok := value.(interface{ GetSource() SemanticNode }); ok {
		return sourced.GetSource()
	}
	return nil
}

// ValueHasSource represents IValue.hasSource.
func ValueHasSource(value Value) bool {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if sourced, ok := value.(interface{ HasSource() bool }); ok {
		return sourced.HasSource()
	}
	return false
}

func InitializeValue(value Value) Value {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if initializer, ok := value.(interface{ Initialize() Value }); ok {
		return initializer.Initialize()
	}
	value.DeepNormalize()
	value.FingerPrint(0)
	return value
}

// ValueIsAtom represents IValue.isAtom's interface default.
func ValueIsAtom(value Value) bool {
	if isNil(value) {
		panic(NewNullPointerException(`Cannot invoke "tlc2.value.impl.Value.isAtom()" because "v" is null`))
	}
	if atom, ok := value.(interface{ IsAtom() bool }); ok {
		return atom.IsAtom()
	}
	return false
}

// ValueHasData represents Value.hasData and its ModelValue override.
func ValueHasData(value Value) bool {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if provider, ok := value.(interface{ HasData() bool }); ok {
		return provider.HasData()
	}
	return false
}

// ValueGetData represents Value.getData and its ModelValue override.
func ValueGetData(value Value) any {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if provider, ok := value.(interface{ GetData() any }); ok {
		return provider.GetData()
	}
	return nil
}

// ValueSetData represents Value.setData and its ModelValue override.
func ValueSetData(value Value, data any) any {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if provider, ok := value.(interface{ SetData(any) any }); ok {
		return provider.SetData(data)
	}
	panic(NewWrongInvocationException("Value: Can not set data\n" + ValuesPPR(value)))
}

// ValueToSetEnum represents Value.toSetEnum and its concrete overrides.
func ValueToSetEnum(value Value) (*SetEnumValue, error) {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	return tryToSetEnumValue(value)
}

// ValueToFcnRcd represents Value.toFcnRcd and its concrete overrides.
func ValueToFcnRcd(value Value) *FcnRcdValue {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	return asFcnRcdValue(value)
}

// ValueToRcd represents Value.toRcd and its concrete overrides.
func ValueToRcd(value Value) *RecordValue {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	return asRecordValue(value)
}

// ValueToTuple represents Value.toTuple and its concrete overrides.
func ValueToTuple(value Value) *TupleValue {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	return asTupleValue(value)
}

// ValueToState represents Value.toState and its concrete overrides.
func ValueToState(value Value) *TLCStateMut {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if converter, ok := value.(interface{ ToState() *TLCStateMut }); ok {
		return converter.ToState()
	}
	return nil
}

// ValueToUnquotedString represents Value.toUnquotedString and its overrides.
func ValueToUnquotedString(value Value) string {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if unquoted, ok := value.(interface{ ToUnquotedString() string }); ok {
		return unquoted.ToUnquotedString()
	}
	if unquoted, ok := value.(interface{ UnquotedString() string }); ok {
		return unquoted.UnquotedString()
	}
	return value.String()
}

// ValueGetTypeString represents Value.getTypeString.
func ValueGetTypeString(value Value) string {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	typ := reflect.TypeOf(value)
	name := "Value"
	if typ != nil {
		if typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		name = typ.Name()
	}
	return name + ": " + ValueKindString(value)
}

// ValueKindString represents Value.getKindString, including Java's signed-byte
// array index and source-aware FingerprintException boundary.
func ValueKindString(value Value) (result string) {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(value, nil)
	index := int(int8(value.Kind()))
	if index < 0 || index >= len(valueKindImage) {
		panic(NewArrayIndexOutOfBoundsException(index, len(valueKindImage)))
	}
	return valueKindImage[index]
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
			panic(FingerprintExceptionGetNewHead(value, cause))
		}
		panic(failure)
	}
	if err != nil {
		*err = wrapValueFailure(value, *err)
	}
}

func wrapValueFailure(value Value, err error) error {
	if err != nil && valueSource(value) != nil {
		return FingerprintExceptionGetNewHead(value, err)
	}
	return err
}

// ValueExcept is shared by pointer: nested EXCEPT application advances its cursor.
// Lazy matching and materialization explicitly clone cursors at source boundaries.
type ValueExcept struct {
	Path  []Value
	Index int
	Value Value
}

func NewValueExcept(path []Value, value Value) *ValueExcept {
	return &ValueExcept{Path: path, Value: value}
}

func NewValueExceptCopy(ex *ValueExcept, indices ...int) *ValueExcept {
	if len(indices) > 1 {
		panic(NewIllegalArgumentException("NewValueExceptCopy accepts at most one index"))
	}
	if ex == nil {
		panic(NewNullPointerException())
	}
	index := ex.Index
	if len(indices) != 0 {
		index = indices[0]
	}
	index = int(int32(index))
	return &ValueExcept{Path: ex.Path, Index: index, Value: ex.Value}
}

type BaseValue struct {
	source SemanticNode
	CM     CostModel
}

// Java reference reads/writes are atomic. Protect the Go interface's two words
// when worker-local lazy values attach source metadata to shared cached values.
var valueSourceMutex sync.RWMutex

func newBaseValue(cms ...CostModel) BaseValue {
	if len(cms) > 1 {
		panic(NewIllegalArgumentException("value constructors accept at most one cost model"))
	}
	if len(cms) > 0 {
		return BaseValue{CM: cms[0]}
	}
	return BaseValue{}
}

func (v *BaseValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.CM
}

func (v *BaseValue) SetCostModel(cm CostModel) {
	if v == nil {
		panic(NewNullPointerException())
	}
	v.CM = cm
}

func (BaseValue) KindStringFor(kind ValueKind) string {
	index := int(int8(kind))
	if index < 0 || index >= len(valueKindImage) {
		panic(NewArrayIndexOutOfBoundsException(index, len(valueKindImage)))
	}
	return valueKindImage[index]
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
	if v == nil {
		panic(NewNullPointerException())
	}
	valueSourceMutex.Lock()
	v.source = source
	valueSourceMutex.Unlock()
}

func (v *BaseValue) GetSource() SemanticNode {
	if v == nil {
		panic(NewNullPointerException())
	}
	valueSourceMutex.RLock()
	source := v.source
	valueSourceMutex.RUnlock()
	return source
}

func (v *BaseValue) HasSource() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.GetSource() != nil
}

type BoolValue struct {
	BaseValue
	Val bool
}

func (v *BoolValue) Mutates() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return false
}

func (v *BoolValue) IsAtom() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}

var (
	BoolFalse = &BoolValue{Val: false}
	BoolTrue  = &BoolValue{Val: true}
)

func NewBoolValue(v bool) *BoolValue {
	return &BoolValue{Val: v}
}

// BoolValueOf returns Java's canonical ValTrue or ValFalse singleton.
func BoolValueOf(v bool) *BoolValue {
	if v {
		return BoolTrue
	}
	return BoolFalse
}

func (v *BoolValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return BoolValueKind
}
func (v *BoolValue) KindString() string { return ValueKindString(v) }
func (v *BoolValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}
func (v *BoolValue) GetVal() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.Val
}

func (v *BoolValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	o, ok := other.(*BoolValue)
	if !ok || o == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueCompareTo(v)
		}
		if isNil(other) {
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
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	o, ok := other.(*BoolValue)
	if !ok || o == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueEquals(v)
		}
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to compare equality of boolean %s with non-boolean:\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return v.Val == o.Val, nil
}

func (v *BoolValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(elem) {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of the boolean %s", ValuesPPR(elem), ValuesPPR(v)))
}

func (v *BoolValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the boolean %s is a finite set.", ValuesPPR(v)))
}

func (v *BoolValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compute the number of elements in the boolean %s.", ValuesPPR(v)))
}

func (v *BoolValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *BoolValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
}
func (v *BoolValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *BoolValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *BoolValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *BoolValue) Permute(*MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *BoolValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	fp = FP64ExtendByte(fp, byte(BoolValueKind))
	if v.Val {
		return FP64ExtendUTF16(fp, []uint16{'t'})
	}
	return FP64ExtendUTF16(fp, []uint16{'f'})
}

func (v *BoolValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the boolean %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *BoolValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
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
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *BoolValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
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

	valCaptured bool
	valOwner    int32
}

func (v *IntValue) Mutates() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return false
}

func (v *IntValue) IsAtom() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}

var intValueCache = func() []*IntValue {
	out := make([]*IntValue, 10)
	for i := range out {
		out[i] = newIntValue(int32(i))
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
	InitializeValueStatics()
	for i := range intValueCache {
		intValueCache[i] = newIntValue(int32(i))
	}
	IntNegOne, IntOne, IntZero = NewIntValue(-1), NewIntValue(1), NewIntValue(0)
}

func NewIntValue(v int32) *IntValue {
	if v >= 0 && int(v) < len(intValueCache) {
		return intValueCache[v]
	}
	return newIntValue(v)
}

func newIntValue(value int32) *IntValue {
	return &IntValue{Val: value, valCaptured: true, valOwner: value}
}

func (v *IntValue) ownedVal() int32 {
	if v.valCaptured {
		return v.valOwner
	}
	return v.Val
}

func (v *IntValue) captureVal(value int32) {
	v.Val, v.valOwner, v.valCaptured = value, value, true
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
	if a == nil {
		panic(NewNullPointerException(`Cannot read field "val" because "n" is null`))
	}
	if b == nil {
		panic(NewNullPointerException(`Cannot read field "val" because "m" is null`))
	}
	return NewIntValue(a.ownedVal() + b.ownedVal())
}

func (v *IntValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return IntValueKind
}
func (v *IntValue) KindString() string { return ValueKindString(v) }
func (v *IntValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}
func (v *IntValue) NBits() int {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	return IntValueNBits(v.ownedVal())
}

func (v *IntValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	o, ok := other.(*IntValue)
	if !ok || o == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueCompareTo(v)
		}
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare integer %s with non-integer:\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	if v.ownedVal() < o.ownedVal() {
		return -1, nil
	}
	if v.ownedVal() > o.ownedVal() {
		return 1, nil
	}
	return 0, nil
}

func (v *IntValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	o, ok := other.(*IntValue)
	if !ok || o == nil {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueEquals(v)
		}
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of integer %s with non-integer:\n%s", ValuesPPR(v), ValuesPPR(other)))
	}
	return v.ownedVal() == o.ownedVal(), nil
}

func (v *IntValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if isNil(elem) {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure("Attempted to check if the value:\n" + ValuesPPR(elem) + "\nis an element of the integer " + ValuesPPR(v))
}

func (v *IntValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the integer %s is a finite set.", ValuesPPR(v)))
}

func (v *IntValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compute the number of elements in the integer %s.", ValuesPPR(v)))
}

func (v *IntValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *IntValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
}
func (v *IntValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *IntValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *IntValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *IntValue) Permute(*MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}

func (v *IntValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	return FP64ExtendInt(FP64ExtendByte(fp, byte(IntValueKind)), v.ownedVal())
}

func (v *IntValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the integer %s.", ValuesPPR(v)))
	}
	return ex.Value, nil
}

func (v *IntValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
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
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *IntValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	sb.WriteString(strconv.FormatInt(int64(v.ownedVal()), 10))
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
	Val   *UniqueString
	owner Value

	valCaptured bool
	valOwner    *UniqueString
}

func (v *StringValue) Mutates() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return false
}

func (v *StringValue) IsAtom() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}

// Inherited StringValue methods retain the concrete Java receiver, including
// virtual diagnostic printing and identity returns for the debugger subclass.
func (v *StringValue) receiver() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	if v.owner != nil {
		return v.owner
	}
	return v
}

func asStringValue(value any) (*StringValue, bool) {
	switch v := value.(type) {
	case *StringValue:
		return v, true
	case *DebuggerValue:
		return v.StringValue, true
	default:
		return nil, false
	}
}

func NewStringValue(s string, cms ...CostModel) *StringValue {
	return NewStringValueFromUnique(UniqueStringOf(s), cms...)
}

func NewStringValueFromUnique(s *UniqueString, cms ...CostModel) *StringValue {
	return &StringValue{BaseValue: newBaseValue(cms...), Val: s, valCaptured: true, valOwner: s}
}

func (v *StringValue) ownedVal() *UniqueString {
	if v.valCaptured {
		return v.valOwner
	}
	return v.Val
}

func (v *StringValue) captureVal(value *UniqueString) {
	v.Val, v.valOwner, v.valCaptured = value, value, true
}

func (v *StringValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return StringValueKind
}
func (v *StringValue) KindString() string { return ValueKindString(v) }
func (v *StringValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}
func (v *StringValue) GetVal() *UniqueString {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.ownedVal()
}
func (v *StringValue) Length() int {
	defer catchValueFailure(v.receiver(), nil)
	if v.ownedVal() == nil {
		panic(NewNullPointerException())
	}
	return v.ownedVal().Length()
}

func (v *StringValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v.receiver(), &err)
	var o *StringValue
	var ok bool
	if !isNil(other) {
		o, ok = asStringValue(other)
	}
	if !ok {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueCompareTo(v.receiver())
		}
		selfText := ValuesPPR(v.receiver())
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compare string %s with non-string:\n%s", selfText, ValuesPPR(other)))
	}
	if v.ownedVal() == nil || o.ownedVal() == nil {
		panic(NewNullPointerException())
	}
	return v.ownedVal().Compare(o.ownedVal()), nil
}

func (v *StringValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v.receiver(), &err)
	var o *StringValue
	var ok bool
	if !isNil(other) {
		o, ok = asStringValue(other)
	}
	if !ok {
		if mv, ok := other.(*ModelValue); ok && mv != nil {
			return mv.ModelValueEquals(v.receiver())
		}
		selfText := ValuesPPR(v.receiver())
		if isNil(other) {
			panic(NewNullPointerException())
		}
		return false, v.runtimeFailure(fmt.Sprintf("Attempted to check equality of string %s with non-string:\n%s", selfText, ValuesPPR(other)))
	}
	if v.ownedVal() == nil || o.ownedVal() == nil {
		panic(NewNullPointerException())
	}
	return v.ownedVal().Equal(o.ownedVal()), nil
}

func (v *StringValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v.receiver(), &err)
	if isNil(elem) {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the value:\n%s\nis an element of the string %s", ValuesPPR(elem), ValuesPPR(v.receiver())))
}

func (v *StringValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v.receiver(), &err)
	return false, v.runtimeFailure(fmt.Sprintf("Attempted to check if the string %s is a finite set.", ValuesPPR(v.receiver())))
}

func (v *StringValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v.receiver(), &err)
	return 0, v.runtimeFailure(fmt.Sprintf("Attempted to compute the number of elements in the string %s.", ValuesPPR(v.receiver())))
}

func (v *StringValue) Normalize() Value      { return v.receiver() }
func (v *StringValue) DeepNormalize()        { _ = v.receiver() }
func (v *StringValue) IsNormalized() bool    { _ = v.receiver(); return true }
func (v *StringValue) IsDefined() bool       { _ = v.receiver(); return true }
func (v *StringValue) DeepCopy() Value       { return v.receiver() }
func (v *StringValue) Permute(*MVPerm) Value { return v.receiver() }

func (v *StringValue) FingerPrint(fp uint64) uint64 {
	defer catchValueFailure(v.receiver(), nil)
	fp = FP64ExtendByte(fp, byte(StringValueKind))
	if v.ownedVal() == nil {
		panic(NewNullPointerException())
	}
	fp = FP64ExtendInt(fp, int32(v.ownedVal().Length()))
	return FP64ExtendString(fp, v.ownedVal().String())
}

func (v *StringValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v.receiver(), &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the string %s.", ValuesPPR(v.receiver())))
	}
	return ex.Value, nil
}

func (v *StringValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v.receiver(), &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure(fmt.Sprintf("Attempted to apply EXCEPT construct to the string %s.", ValuesPPR(v.receiver())))
	}
	return v.receiver(), nil
}

func (v *StringValue) String() string {
	return ValueToString(v.receiver(), "", true)
}

func (v *StringValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v.receiver(), nil)
	if v.ownedVal() == nil {
		panic(NewNullPointerException())
	}
	sb.WriteString(`"` + tlaStringPrintVersion(v.ownedVal().String()) + `"`)
	return sb
}

// ToUnquotedString is StringValue.toUnquotedString.
func (v *StringValue) ToUnquotedString() string {
	// Java reads val before entering PrintVersion's catch boundary.
	if v == nil || v.ownedVal() == nil {
		panic(NewNullPointerException())
	}
	return tlaStringPrintVersion(v.ownedVal().String())
}

func (v *StringValue) UnquotedString() string {
	return v.ToUnquotedString()
}

func (v *StringValue) RawString() string {
	if v == nil || v.ownedVal() == nil {
		return ""
	}
	return v.ownedVal().String()
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
