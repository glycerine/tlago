package tlc

type UndefValue struct {
	BaseValue
}

var ValUndef = &UndefValue{}

func (v *UndefValue) Kind() ValueKind    { return UndefValueKind }
func (v *UndefValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *UndefValue) Compare(other Value) (int, error) {
	if _, ok := other.(*UndefValue); ok {
		return 0, nil
	}
	return 1, nil
}

func (v *UndefValue) Equal(other Value) (bool, error) {
	_, ok := other.(*UndefValue)
	return ok, nil
}

func (v *UndefValue) Member(elem Value) (bool, error) {
	return false, v.unsupported("attempted to check if the value:\n%s\nis an element %s", elem, v)
}

func (v *UndefValue) IsFinite() (bool, error) {
	return false, v.unsupported("attempted to check if the value %s is a finite set", v)
}

func (v *UndefValue) Size() (int, error) {
	return 0, v.unsupported("attempted to compute the number of elements in the value %s", v)
}

func (v *UndefValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, v.unsupported("attempted to apply EXCEPT construct to the value %s", v)
	}
	return ex.Value, nil
}

func (v *UndefValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, v.unsupported("attempted to apply EXCEPT construct to the value %s", v)
	}
	return v, nil
}

func (v *UndefValue) IsNormalized() bool           { return true }
func (v *UndefValue) Normalize() Value             { return v }
func (v *UndefValue) DeepNormalize()               {}
func (v *UndefValue) IsDefined() bool              { return false }
func (v *UndefValue) DeepCopy() Value              { return v }
func (v *UndefValue) FingerPrint(fp uint64) uint64 { return unsupportedValueFingerprint(v) }
func (v *UndefValue) Permute(*MVPerm) Value        { return v }
func (v *UndefValue) String() string               { return "UNDEF" }

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

type UserValue struct {
	BaseValue
	UserObj UserObj
}

func NewUserValue(obj UserObj) *UserValue {
	return &UserValue{UserObj: obj}
}

func (v *UserValue) Kind() ValueKind    { return UserValueKind }
func (v *UserValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *UserValue) Compare(other Value) (int, error) {
	if _, ok := other.(*UserValue); ok {
		return v.UserObj.Compare(other)
	}
	if _, ok := other.(*ModelValue); ok {
		return 1, nil
	}
	return 0, v.unsupported("attempted to compare overridden value %s with non-overridden value:\n%s", v, other)
}

func (v *UserValue) Equal(other Value) (bool, error) {
	cmp, err := v.Compare(other)
	return cmp == 0, err
}

func (v *UserValue) Member(elem Value) (bool, error) {
	return v.UserObj.Member(elem)
}

func (v *UserValue) IsFinite() (bool, error) {
	return v.UserObj.IsFinite()
}

func (v *UserValue) Size() (int, error) {
	return 0, v.unsupported("attempted to compute the number of elements in the overridden value %s", v)
}

func (v *UserValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index < len(ex.Path) {
		return nil, v.unsupported("attempted to apply EXCEPT to the overridden value %s", v)
	}
	return ex.Value, nil
}

func (v *UserValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	if len(exs) != 0 {
		return nil, v.unsupported("attempted to apply EXCEPT to the overridden value %s", v)
	}
	return v, nil
}

func (v *UserValue) IsNormalized() bool           { return true }
func (v *UserValue) Normalize() Value             { return v }
func (v *UserValue) DeepNormalize()               {}
func (v *UserValue) IsDefined() bool              { return true }
func (v *UserValue) DeepCopy() Value              { return v }
func (v *UserValue) FingerPrint(fp uint64) uint64 { return unsupportedValueFingerprint(v) }
func (v *UserValue) Permute(*MVPerm) Value        { return v }
func (v *UserValue) String() string               { return v.UserObj.String() }

func unsupportedValueFingerprint(value Value) uint64 {
	panic(newTLCError(ECGeneral, "TLC has found a state in which the value of a variable contains %s", ValuesPPR(value)))
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
