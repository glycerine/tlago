package tlc

import "strings"

type UndefValue struct {
	BaseValue
}

func NewUndefValue() *UndefValue {
	return &UndefValue{BaseValue: newBaseValue()}
}

func (v *UndefValue) Mutates() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return false
}

var ValUndef = NewUndefValue()

func (v *UndefValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return UndefValueKind
}
func (v *UndefValue) KindString() string { return ValueKindString(v) }
func (v *UndefValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

func (v *UndefValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if _, ok := other.(*UndefValue); ok && !isNil(other) {
		return 0, nil
	}
	return 1, nil
}

func (v *UndefValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	_, ok := other.(*UndefValue)
	return ok && !isNil(other), nil
}

func (v *UndefValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure("Attempted to check if the value:\n" + ValuesPPR(elem) + "\nis an element " + ValuesPPR(v))
}

func (v *UndefValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure("Attempted to check if the value " + ValuesPPR(v) + " is a finite set.")
}

func (v *UndefValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure("Attempted to compute the number of elements in the value " + ValuesPPR(v) + ".")
}

func (v *UndefValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil || ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT construct to the value " + ValuesPPR(v) + ".")
	}
	return ex.Value, nil
}

func (v *UndefValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT construct to the value " + ValuesPPR(v) + ".")
	}
	return v, nil
}

func (v *UndefValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *UndefValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *UndefValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
}
func (v *UndefValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return false
}
func (v *UndefValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *UndefValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	return unsupportedValueFingerprint(v)
}
func (v *UndefValue) Permute(*MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return unsupportedValuePermutation(v)
}
func (v *UndefValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *UndefValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	sb.WriteString("UNDEF")
	return sb
}

type UserObj interface {
	Compare(Value) (int, error)
	Member(Value) (bool, error)
	IsFinite() (bool, error)
	String() string
}

type UserObjWithIsEmpty interface {
	UserObj
	IsEmpty() (bool, error)
}

type UserObjWithNonEnumerableErrorMsg interface {
	UserObj
	GetNonEnumerableErrorMsg(SemanticNode) string
}

// UserObjIsEmpty represents UserObj.isEmpty's overridable default.
func UserObjIsEmpty(obj UserObj) (bool, error) {
	if isNil(obj) {
		panic(NewNullPointerException())
	}
	if provider, ok := obj.(UserObjWithIsEmpty); ok {
		return provider.IsEmpty()
	}
	return false, NewTLCRuntimeExceptionMessage("Shouldn't call isEmpty() on value " + obj.String())
}

// UserObjGetNonEnumerableErrorMsg represents the nullable default diagnostic.
// An empty Go string is the source method's null result.
func UserObjGetNonEnumerableErrorMsg(obj UserObj, expr SemanticNode) string {
	if isNil(obj) {
		panic(NewNullPointerException())
	}
	if provider, ok := obj.(UserObjWithNonEnumerableErrorMsg); ok {
		return provider.GetNonEnumerableErrorMsg(expr)
	}
	return ""
}

// ValueGetNonEnumerableErrorMsg represents Value.getNonEnumerableErrorMsg and
// its UserValue override.
func ValueGetNonEnumerableErrorMsg(value Value, expr SemanticNode) string {
	if isNil(value) {
		panic(NewNullPointerException())
	}
	if user, ok := value.(*UserValue); ok && user != nil {
		return user.GetNonEnumerableErrorMsg(expr)
	}
	return defaultNonEnumerableErrorMsg(value, expr)
}

func defaultNonEnumerableErrorMsg(value Value, expr SemanticNode) string {
	return "TLC encountered a non-enumerable quantifier bound\n" +
		ValuesPPR(value) + ".\n" + SemanticString(expr)
}

type UserValue struct {
	BaseValue
	UserObj UserObj
}

func NewUserValue(obj UserObj) *UserValue {
	return &UserValue{UserObj: obj}
}

func (v *UserValue) userObject() UserObj {
	if v == nil || isNil(v.UserObj) {
		panic(NewNullPointerException())
	}
	return v.UserObj
}

func (v *UserValue) Kind() ValueKind {
	if v == nil {
		panic(NewNullPointerException())
	}
	return UserValueKind
}
func (v *UserValue) KindString() string { return ValueKindString(v) }
func (v *UserValue) GetCostModel() CostModel {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v.BaseValue.GetCostModel()
}

// GetNonEnumerableErrorMsg ports UserValue.getNonEnumerableErrorMsg. The
// source invokes a non-null UserObj override twice: once to test it and again
// to return it, so retain that observable dispatch order.
func (v *UserValue) GetNonEnumerableErrorMsg(expr SemanticNode) string {
	obj := v.userObject()
	if message := UserObjGetNonEnumerableErrorMsg(obj, expr); message != "" {
		return UserObjGetNonEnumerableErrorMsg(obj, expr)
	}
	return defaultNonEnumerableErrorMsg(v, expr)
}

func (v *UserValue) Compare(other Value) (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if _, ok := other.(*UserValue); ok && !isNil(other) {
		return v.userObject().Compare(other)
	}
	if _, ok := other.(*ModelValue); ok && !isNil(other) {
		return 1, nil
	}
	message := "Attempted to compare overridden value " + ValuesPPR(v) + " with non-overridden value:\n"
	if isNil(other) {
		panic(NewNullPointerException(`Cannot invoke "Object.toString()" because "obj" is null`))
	}
	return 0, v.runtimeFailure(message + ValuesPPR(other))
}

func (v *UserValue) Equal(other Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	cmp, err := v.Compare(other)
	return cmp == 0, err
}

func (v *UserValue) Member(elem Value) (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return v.userObject().Member(elem)
}

func (v *UserValue) IsFinite() (resultBool bool, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return v.userObject().IsFinite()
}

func (v *UserValue) Size() (resultInt int, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure("Attempted to compute the number of elements in the overridden value " + ValuesPPR(v) + ".")
}

func (v *UserValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if ex == nil || ex.Path == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the overridden value " + ValuesPPR(v) + ".")
	}
	return ex.Value, nil
}

func (v *UserValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	if exs == nil {
		panic(NewNullPointerException())
	}
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the overridden value " + ValuesPPR(v) + ".")
	}
	return v, nil
}

func (v *UserValue) IsNormalized() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *UserValue) Normalize() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *UserValue) DeepNormalize() {
	if v == nil {
		panic(NewNullPointerException())
	}
}
func (v *UserValue) IsDefined() bool {
	if v == nil {
		panic(NewNullPointerException())
	}
	return true
}
func (v *UserValue) DeepCopy() Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return v
}
func (v *UserValue) FingerPrint(fp uint64) uint64 {
	if v == nil {
		panic(NewNullPointerException())
	}
	return unsupportedValueFingerprint(v)
}
func (v *UserValue) Permute(*MVPerm) Value {
	if v == nil {
		panic(NewNullPointerException())
	}
	return unsupportedValuePermutation(v)
}
func (v *UserValue) String() string {
	if v == nil {
		panic(NewNullPointerException())
	}
	return ValueToString(v, "", true)
}

func (v *UserValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, nil)
	obj := v.userObject()
	if printer, ok := obj.(ValueStringPrinter); ok {
		return printer.ToString(sb, offset, swallow)
	}
	sb.WriteString(obj.String())
	return sb
}

// invalidStateValueFailure is the source Value fingerprint/permutation Assert.
// Format before reading the source, preserving failures from value formatting.
func invalidStateValueFailure(value Value) error {
	message := "TLC has found a state in which the value of a variable contains " + ValuesPPR(value)
	if source := valueSource(value); source != nil {
		return NewTLCDetailedRuntimeException(ECGeneral, message, source, EmptyContext)
	}
	return NewTLCRuntimeExceptionMessage(message)
}

func unsupportedValueFingerprint(value Value) uint64 {
	defer catchValueFailure(value, nil)
	panic(invalidStateValueFailure(value))
}

func unsupportedValuePermutation(value Value) Value {
	defer catchValueFailure(value, nil)
	panic(invalidStateValueFailure(value))
}

type AnySet struct{}

var AnySetValue = NewUserValue(AnySet{})

func (AnySet) Compare(val Value) (int, error) {
	return 0, newTLCErrorCode(ECTLCModuleCompareValue, "ANY", ValuesPPR(val))
}

func (AnySet) Member(Value) (bool, error) {
	return true, nil
}

func (AnySet) IsFinite() (bool, error) {
	return false, nil
}

func (AnySet) String() string {
	return "ANY"
}

func (AnySet) ToString(sb *strings.Builder, _ int, _ bool) *strings.Builder {
	sb.WriteString("ANY")
	return sb
}
