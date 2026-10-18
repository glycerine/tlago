package tlc

// Reducible represents the eager finite-set operations exposed by Java's
// Reducible interface.
type Reducible interface {
	Size() (int, error)
	Member(Value) (bool, error)
	Diff(Value) (Value, error)
	Cap(Value) (Value, error)
	Cup(Value) (Value, error)
	Elements() ValueEnumeration
}

var (
	_ Reducible = (*SetEnumValue)(nil)
	_ Reducible = (*IntervalValue)(nil)
)

func asReducible(value Value) (Reducible, bool) {
	if isNil(value) {
		return nil, false
	}
	reducible, ok := value.(Reducible)
	return reducible, ok
}

func (v *SetEnumValue) Diff(other Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	size := v.Elems.Len()
	values := NewValueVec(0)
	for i := 0; i < size; i++ {
		elem := v.Elems.At(i)
		member, err := other.Member(elem)
		if err != nil {
			return nil, err
		}
		if !member {
			values.Add(elem)
		}
	}
	return NewSetEnumValueVec(values, v.IsNormalized(), v.CM), nil
}

func (v *SetEnumValue) Cap(other Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	size := v.Elems.Len()
	values := NewValueVec(0)
	for i := 0; i < size; i++ {
		elem := v.Elems.At(i)
		member, err := other.Member(elem)
		if err != nil {
			return nil, err
		}
		if member {
			values.Add(elem)
		}
	}
	return NewSetEnumValueVec(values, v.IsNormalized(), v.CM), nil
}

func (v *SetEnumValue) Cup(other Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	size := v.Elems.Len()
	if size == 0 {
		return other, nil
	}
	reducible, ok := asReducible(other)
	if !ok {
		return NewSetCupValue(v, other, v.CM), nil
	}
	values := NewValueVec(0)
	for i := 0; i < size; i++ {
		values.Add(v.Elems.At(i))
	}
	enum := reducible.Elements()
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		member, err := v.Member(elem)
		if err != nil {
			return nil, err
		}
		if !member {
			values.Add(elem)
		}
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	return NewSetEnumValueVec(values, false), nil
}

func (v *IntervalValue) Diff(other Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	size, err := v.Size()
	if err != nil {
		return nil, err
	}
	values := NewValueVec(0)
	for i := 0; i < size; i++ {
		elem := NewIntValue(v.Low + int32(i))
		member, err := other.Member(elem)
		if err != nil {
			return nil, err
		}
		if !member {
			values.Add(elem)
		}
	}
	return NewSetEnumValueVec(values, true, v.CM), nil
}

func (v *IntervalValue) Cap(other Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	size, err := v.Size()
	if err != nil {
		return nil, err
	}
	values := NewValueVec(0)
	for i := 0; i < size; i++ {
		elem := NewIntValue(v.Low + int32(i))
		member, err := other.Member(elem)
		if err != nil {
			return nil, err
		}
		if member {
			values.Add(elem)
		}
	}
	return NewSetEnumValueVec(values, true, v.CM), nil
}

func (v *IntervalValue) Cup(other Value) (resultValue Value, err error) {
	if v == nil {
		panic(NewNullPointerException())
	}
	defer catchValueFailure(v, &err)
	size, err := v.Size()
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return other, nil
	}
	reducible, ok := asReducible(other)
	if !ok {
		return NewSetCupValue(v, other, v.CM), nil
	}
	values := NewValueVec(0)
	for i := 0; i < size; i++ {
		values.Add(NewIntValue(v.Low + int32(i)))
	}
	enum := reducible.Elements()
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		member, err := v.Member(elem)
		if err != nil {
			return nil, err
		}
		if !member {
			values.Add(elem)
		}
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	return NewSetEnumValueVec(values, false, v.CM), nil
}
