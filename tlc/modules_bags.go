package tlc

func EmptyBag() Value {
	return EmptyFcn
}

func IsABag(value Value) (*BoolValue, error) {
	fcn := asFcnRcdValue(value)
	if fcn == nil {
		return BoolFalse, nil
	}
	for _, value := range fcn.Values {
		count, ok := value.(*IntValue)
		if !ok || count.Val <= 0 {
			return BoolFalse, nil
		}
	}
	return BoolTrue, nil
}

func BagCardinality(value Value) (*IntValue, error) {
	fcn, err := requireBagFunction("BagCardinality", value)
	if err != nil {
		return nil, err
	}
	total := int32(0)
	for _, value := range fcn.Values {
		count, ok := value.(*IntValue)
		if !ok || count.Val <= 0 {
			return nil, newTLCError(ECGeneral, "BagCardinality expected a bag, got %s", value)
		}
		total += count.Val
	}
	return NewIntValue(total), nil
}

func BagIn(elem Value, bag Value) (*BoolValue, error) {
	fcn, err := requireBagFunction("BagIn", bag)
	if err != nil {
		return nil, err
	}
	domain := fcn.DomainAsValues()
	for i, dval := range domain {
		eq, err := elem.Equal(dval)
		if err != nil {
			return nil, err
		}
		if eq {
			count, ok := fcn.Values[i].(*IntValue)
			if !ok {
				return nil, newTLCError(ECGeneral, "second argument of BagIn must be a bag, got %s", bag)
			}
			return NewBoolValue(count.Val > 0), nil
		}
	}
	return BoolFalse, nil
}

func CopiesIn(elem Value, bag Value) (*IntValue, error) {
	fcn, err := requireBagFunction("CopiesIn", bag)
	if err != nil {
		return nil, err
	}
	domain := fcn.DomainAsValues()
	for i, dval := range domain {
		eq, err := elem.Equal(dval)
		if err != nil {
			return nil, err
		}
		if eq {
			count, ok := fcn.Values[i].(*IntValue)
			if !ok {
				return nil, newTLCError(ECGeneral, "second argument of CopiesIn must be a bag, got %s", bag)
			}
			return count, nil
		}
	}
	return IntZero, nil
}

func BagCup(b1 Value, b2 Value) (Value, error) {
	fcn1, err := requireBag("first", "(+)", b1)
	if err != nil {
		return nil, err
	}
	fcn2, err := requireBag("second", "(+)", b2)
	if err != nil {
		return nil, err
	}
	domain := NewValueVec(0)
	values := NewValueVec(0)
	for i, dval := range fcn1.DomainAsValues() {
		domain.Add(dval)
		values.Add(fcn1.Values[i])
	}
	domain1Len := domain.Len()
	for i, dval := range fcn2.DomainAsValues() {
		found := false
		for j := 0; j < domain1Len; j++ {
			eq, err := dval.Equal(domain.At(j))
			if err != nil {
				return nil, err
			}
			if eq {
				left := values.At(j).(*IntValue)
				right := fcn2.Values[i].(*IntValue)
				values.Set(j, NewIntValue(left.Val+right.Val))
				found = true
				break
			}
		}
		if !found {
			domain.Add(dval)
			values.Add(fcn2.Values[i])
		}
	}
	return NewFcnRcdValue(domain.ToArray(), values.ToArray(), false), nil
}

func BagDiff(b1 Value, b2 Value) (Value, error) {
	fcn1, err := requireBagFunction("first argument of (-)", b1)
	if err != nil {
		return nil, err
	}
	fcn2, err := requireBagFunction("second argument of (-)", b2)
	if err != nil {
		return nil, err
	}
	domain := NewValueVec(0)
	values := NewValueVec(0)
	domain2 := fcn2.DomainAsValues()
	for i, dval := range fcn1.DomainAsValues() {
		v1 := fcn1.Values[i].(*IntValue).Val
		for j, d2 := range domain2 {
			eq, err := dval.Equal(d2)
			if err != nil {
				return nil, err
			}
			if eq {
				v1 -= fcn2.Values[j].(*IntValue).Val
				break
			}
		}
		if v1 > 0 {
			domain.Add(dval)
			values.Add(NewIntValue(v1))
		}
	}
	return NewFcnRcdValue(domain.ToArray(), values.ToArray(), fcn1.IsNormalized()), nil
}

func BagUnion(set Value) (Value, error) {
	setEnum, err := toSetEnumValue(set)
	if err != nil {
		return nil, err
	}
	setEnum.Normalize()
	if setEnum.Elems.Len() == 0 {
		return EmptyFcn, nil
	}
	if setEnum.Elems.Len() == 1 {
		return setEnum.Elems.At(0), nil
	}
	first, err := requireBagFunction("BagUnion", setEnum.Elems.At(0))
	if err != nil {
		return nil, err
	}
	domain := NewValueVec(0)
	values := NewValueVec(0)
	for i, dval := range first.DomainAsValues() {
		domain.Add(dval)
		values.Add(first.Values[i])
	}
	for i := 1; i < setEnum.Elems.Len(); i++ {
		fcn, err := requireBagFunction("BagUnion", setEnum.Elems.At(i))
		if err != nil {
			return nil, err
		}
		for j, dval := range fcn.DomainAsValues() {
			found := false
			for k := 0; k < domain.Len(); k++ {
				eq, err := dval.Equal(domain.At(k))
				if err != nil {
					return nil, err
				}
				if eq {
					left := values.At(k).(*IntValue)
					right := fcn.Values[j].(*IntValue)
					values.Set(k, NewIntValue(left.Val+right.Val))
					found = true
					break
				}
			}
			if !found {
				domain.Add(dval)
				values.Add(fcn.Values[j])
			}
		}
	}
	return NewFcnRcdValue(domain.ToArray(), values.ToArray(), false), nil
}

func SqSubseteq(b1 Value, b2 Value) (*BoolValue, error) {
	fcn1, err := requireBagFunction("\\sqsubseteq", b1)
	if err != nil {
		return nil, err
	}
	fcn2, err := requireBagFunction("\\sqsubseteq", b2)
	if err != nil {
		return nil, err
	}
	domain2 := fcn2.DomainAsValues()
	for i, dval := range fcn1.DomainAsValues() {
		v1 := fcn1.Values[i].(*IntValue).Val
		for j, d2 := range domain2 {
			eq, err := dval.Equal(d2)
			if err != nil {
				return nil, err
			}
			if eq {
				v1 -= fcn2.Values[j].(*IntValue).Val
				break
			}
		}
		if v1 > 0 {
			return BoolFalse, nil
		}
	}
	return BoolTrue, nil
}

func BagOfAll(op Value, bag Value) (Value, error) {
	fcn, err := requireBagFunction("BagOfAll", bag)
	if err != nil {
		return nil, err
	}
	domain := NewValueVec(0)
	values := NewValueVec(0)
	for i, dval := range fcn.DomainAsValues() {
		mapped, err := EvalOperatorValue(op, []Value{dval}, EvalClear)
		if err != nil {
			return nil, err
		}
		found := false
		for j := 0; j < domain.Len(); j++ {
			eq, err := mapped.Equal(domain.At(j))
			if err != nil {
				return nil, err
			}
			if eq {
				left := values.At(j).(*IntValue)
				right := fcn.Values[i].(*IntValue)
				values.Set(j, NewIntValue(left.Val+right.Val))
				found = true
				break
			}
		}
		if !found {
			domain.Add(mapped)
			values.Add(fcn.Values[i])
		}
	}
	return NewFcnRcdValue(domain.ToArray(), values.ToArray(), false), nil
}

func BagToSet(bag Value) (Value, error) {
	fcn, err := requireBagFunction("BagToSet", bag)
	if err != nil {
		return nil, err
	}
	return fcn.DomainValue(), nil
}

func SetToBag(set Value) (Value, error) {
	setEnum, err := toSetEnumValue(set)
	if err != nil {
		return nil, err
	}
	if !setEnum.IsNormalized() {
		setEnum.Normalize()
	}
	domain := setEnum.Elems.ToArray()
	values := make([]Value, len(domain))
	for i := range values {
		values[i] = IntOne
	}
	return NewFcnRcdValue(domain, values, setEnum.IsNormalized()), nil
}

func requireBagFunction(name string, value Value) (*FcnRcdValue, error) {
	fcn := asFcnRcdValue(value)
	if fcn == nil {
		return nil, newTLCError(ECGeneral, "%s expected a function with finite domain, got %s", name, value)
	}
	return fcn, nil
}

func requireBag(position string, operator string, value Value) (*FcnRcdValue, error) {
	fcn, err := requireBagFunction(operator, value)
	if err != nil {
		return nil, err
	}
	ok, err := IsABag(fcn)
	if err != nil {
		return nil, err
	}
	if !ok.Val {
		return nil, newTLCError(ECGeneral, "%s argument of %s must be a bag, got %s", position, operator, value)
	}
	return fcn, nil
}
