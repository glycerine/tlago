package tlc

import "strings"

type UndefValue struct {
	BaseValue
}

var ValUndef = &UndefValue{}

func (v *UndefValue) Kind() ValueKind    { return UndefValueKind }
func (v *UndefValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *UndefValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	if _, ok := other.(*UndefValue); ok {
		return 0, nil
	}
	return 1, nil
}

func (v *UndefValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	_, ok := other.(*UndefValue)
	return ok, nil
}

func (v *UndefValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	if elem == nil {
		panic(NewNullPointerException())
	}
	return false, v.runtimeFailure("Attempted to check if the value:\n" + ValuesPPR(elem) + "\nis an element " + ValuesPPR(v))
}

func (v *UndefValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return false, v.runtimeFailure("Attempted to check if the value " + ValuesPPR(v) + " is a finite set.")
}

func (v *UndefValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure("Attempted to compute the number of elements in the value " + ValuesPPR(v) + ".")
}

func (v *UndefValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT construct to the value " + ValuesPPR(v) + ".")
	}
	return ex.Value, nil
}

func (v *UndefValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT construct to the value " + ValuesPPR(v) + ".")
	}
	return v, nil
}

func (v *UndefValue) IsNormalized() bool           { return true }
func (v *UndefValue) Normalize() Value             { return v }
func (v *UndefValue) DeepNormalize()               {}
func (v *UndefValue) IsDefined() bool              { return false }
func (v *UndefValue) DeepCopy() Value              { return v }
func (v *UndefValue) FingerPrint(fp uint64) uint64 { return unsupportedValueFingerprint(v) }
func (v *UndefValue) Permute(*MVPerm) Value        { return unsupportedValuePermutation(v) }
func (v *UndefValue) String() string {
	return ValueToString(v, "", true)
}

func (v *UndefValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
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

func nonEnumerableErrorMsg(value Value, expr SemanticNode) string {
	switch v := value.(type) {
	case *UserValue:
		if msg := userObjNonEnumerableErrorMsg(v.UserObj, expr); msg != "" {
			return msg
		}
	}
	return "TLC encountered a non-enumerable quantifier bound\n" +
		ValuesPPR(value) + ".\n" + SemanticString(expr)
}

func userObjNonEnumerableErrorMsg(obj UserObj, expr SemanticNode) string {
	switch o := obj.(type) {
	case naturalsObj:
		return naturalsNonEnumerableErrorMsg(expr)
	case integersObj:
		return integersNonEnumerableErrorMsg(expr)
	case *sequencesObj:
		return o.nonEnumerableErrorMsg(expr)
	default:
		return ""
	}
}

type UserValue struct {
	BaseValue
	UserObj UserObj
}

func NewUserValue(obj UserObj) *UserValue {
	return &UserValue{UserObj: obj}
}

func (v *UserValue) Kind() ValueKind    { return UserValueKind }
func (v *UserValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *UserValue) Compare(other Value) (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	if _, ok := other.(*UserValue); ok {
		return v.UserObj.Compare(other)
	}
	if _, ok := other.(*ModelValue); ok {
		return 1, nil
	}
	message := "Attempted to compare overridden value " + ValuesPPR(v) + " with non-overridden value:\n"
	if other == nil {
		panic(NewNullPointerException())
	}
	return 0, v.runtimeFailure(message + ValuesPPR(other))
}

func (v *UserValue) Equal(other Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	cmp, err := v.Compare(other)
	return cmp == 0, err
}

func (v *UserValue) Member(elem Value) (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return v.UserObj.Member(elem)
}

func (v *UserValue) IsFinite() (resultBool bool, err error) {
	defer catchValueFailure(v, &err)
	return v.UserObj.IsFinite()
}

func (v *UserValue) Size() (resultInt int, err error) {
	defer catchValueFailure(v, &err)
	return 0, v.runtimeFailure("Attempted to compute the number of elements in the overridden value " + ValuesPPR(v) + ".")
}

func (v *UserValue) TakeExcept(ex *ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if ex == nil {
		panic(NewNullPointerException())
	}
	if ex.Index < len(ex.Path) {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the overridden value " + ValuesPPR(v) + ".")
	}
	return ex.Value, nil
}

func (v *UserValue) TakeExcepts(exs []*ValueExcept) (resultValue Value, err error) {
	defer catchValueFailure(v, &err)
	if len(exs) != 0 {
		return nil, v.runtimeFailure("Attempted to apply EXCEPT to the overridden value " + ValuesPPR(v) + ".")
	}
	return v, nil
}

func (v *UserValue) IsNormalized() bool           { return true }
func (v *UserValue) Normalize() Value             { return v }
func (v *UserValue) DeepNormalize()               {}
func (v *UserValue) IsDefined() bool              { return true }
func (v *UserValue) DeepCopy() Value              { return v }
func (v *UserValue) FingerPrint(fp uint64) uint64 { return unsupportedValueFingerprint(v) }
func (v *UserValue) Permute(*MVPerm) Value        { return unsupportedValuePermutation(v) }
func (v *UserValue) String() string {
	return ValueToString(v, "", true)
}

func (v *UserValue) ToString(sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	defer catchValueFailure(v, nil)
	if printer, ok := v.UserObj.(ValueStringPrinter); ok {
		return printer.ToString(sb, offset, swallow)
	}
	sb.WriteString(v.UserObj.String())
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
