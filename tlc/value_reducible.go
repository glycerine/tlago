package tlc

func (v *SetEnumValue) Diff(other Value) (resultValue Value, err error) {
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
	defer catchValueFailure(v, &err)
	size := v.Elems.Len()
	if size == 0 {
		return other, nil
	}
	if !isReducibleValue(other) {
		return NewSetCupValue(v, other, v.CM), nil
	}
	values := NewValueVec(0)
	for i := 0; i < size; i++ {
		values.Add(v.Elems.At(i))
	}
	enumerable, _ := asEnumerable(other)
	enum := enumerable.Elements()
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
	defer catchValueFailure(v, &err)
	size, err := v.Size()
	if err != nil {
		return nil, err
	}
	if size == 0 {
		return other, nil
	}
	if !isReducibleValue(other) {
		return NewSetCupValue(v, other, v.CM), nil
	}
	values := NewValueVec(0)
	for i := 0; i < size; i++ {
		values.Add(NewIntValue(v.Low + int32(i)))
	}
	enumerable, _ := asEnumerable(other)
	enum := enumerable.Elements()
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
