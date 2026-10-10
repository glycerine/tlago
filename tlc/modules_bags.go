package tlc

const javaBagsBagCupSignature = "public static tlc2.value.impl.Value tlc2.module.Bags.BagCup(tlc2.value.impl.Value,tlc2.value.impl.Value)"

func EmptyBag() Value {
	return emptyFcnValue()
}

func IsABag(value Value) (*BoolValue, error) {
	if value == nil {
		panic(NewNullPointerException())
	}
	fcn := asFcnRcdValue(value)
	if fcn == nil {
		return BoolFalse, nil
	}
	values := fcn.Values
	if values == nil {
		panic(NewNullPointerException())
	}
	for _, value := range values {
		count, ok := value.(*IntValue)
		if !ok || count.Val <= 0 {
			return BoolFalse, nil
		}
	}
	return BoolTrue, nil
}

func BagCardinality(value Value) (*IntValue, error) {
	if value == nil {
		panic(NewNullPointerException())
	}
	fcn := asFcnRcdValue(value)
	if fcn == nil {
		return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "BagCardinality", "a function with a finite domain", ValuesPPR(value))
	}
	total := int32(0)
	values := fcn.Values
	if values == nil {
		panic(NewNullPointerException())
	}
	for _, elem := range values {
		count, ok := elem.(*IntValue)
		if !ok || count.Val <= 0 {
			return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "BagCardinality", "a bag", ValuesPPR(value))
		}
		total += count.Val
	}
	return NewIntValue(total), nil
}

func BagIn(elem Value, bag Value) (*BoolValue, error) {
	if bag == nil {
		panic(NewNullPointerException())
	}
	fcn := asFcnRcdValue(bag)
	if fcn == nil {
		panic(NewNullPointerException())
	}
	values := fcn.Values
	domain := fcn.DomainAsValues()
	if domain == nil {
		panic(NewNullPointerException())
	}
	for i, dval := range domain {
		if elem == nil {
			panic(NewNullPointerException())
		}
		eq, err := elem.Equal(dval)
		if err != nil {
			return nil, err
		}
		if eq {
			count, ok := fcnParameterDomain(values, i).(*IntValue)
			if !ok {
				return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "BagIn", "bag", ValuesPPR(bag))
			}
			return NewBoolValue(count.Val > 0), nil
		}
	}
	return BoolFalse, nil
}

func CopiesIn(elem Value, bag Value) (*IntValue, error) {
	if bag == nil {
		panic(NewNullPointerException())
	}
	fcn := asFcnRcdValue(bag)
	if fcn == nil {
		panic(NewNullPointerException())
	}
	values := fcn.Values
	domain := fcn.DomainAsValues()
	if domain == nil {
		panic(NewNullPointerException())
	}
	for i, dval := range domain {
		if elem == nil {
			panic(NewNullPointerException())
		}
		eq, err := elem.Equal(dval)
		if err != nil {
			return nil, err
		}
		if eq {
			count, ok := fcnParameterDomain(values, i).(*IntValue)
			if !ok {
				return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "CopiesIn", "bag", ValuesPPR(bag))
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
	fcn1 := asFcnRcdValue(b1)
	if fcn1 == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "(-)", "bag", ValuesPPR(b1))
	}
	fcn2 := asFcnRcdValue(b2)
	if fcn2 == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "(-)", "bag", ValuesPPR(b2))
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
	if !canConvertToSetEnum(set) {
		return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "BagUnion", "a finite enumerable set", ValuesPPR(set))
	}
	setEnum, err := toSetEnumValue(set)
	if err != nil {
		return nil, err
	}
	setEnum.Normalize()
	if setEnum.Elems.Len() == 0 {
		return emptyFcnValue(), nil
	}
	if setEnum.Elems.Len() == 1 {
		return setEnum.Elems.At(0), nil
	}
	first := asFcnRcdValue(setEnum.Elems.At(0))
	if first == nil {
		return nil, newTLCErrorCode(ECTLCModuleBagUnion1, ValuesPPR(set))
	}
	domain := NewValueVec(0)
	values := NewValueVec(0)
	for i, dval := range first.DomainAsValues() {
		domain.Add(dval)
		values.Add(first.Values[i])
	}
	for i := 1; i < setEnum.Elems.Len(); i++ {
		fcn := asFcnRcdValue(setEnum.Elems.At(i))
		if fcn == nil {
			return nil, newTLCErrorCode(ECTLCModuleBagUnion1, ValuesPPR(set))
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
	if !isOperatorValue(op) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentErrorAn, "first", "BagOfAll", "operator", ValuesPPR(op))
	}
	fcn := asFcnRcdValue(bag)
	if fcn == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "BagOfAll", "function with a finite domain", ValuesPPR(bag))
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
	if !canConvertToSetEnum(set) {
		return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "BagToSet", "a function with a finite domain", ValuesPPR(set))
	}
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
		return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, name, "a function with a finite domain", ValuesPPR(value))
	}
	return fcn, nil
}

func requireBag(position string, operator string, value Value) (*FcnRcdValue, error) {
	fcn := asFcnRcdValue(value)
	if fcn == nil {
		return nil, javaMethodOverrideError(javaBagsBagCupSignature, `Cannot invoke "tlc2.value.impl.Value.toFcnRcd()" because "b" is null`)
	}
	ok, err := IsABag(fcn)
	if err != nil {
		return nil, err
	}
	if !ok.Val {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, position, operator, "bag", ValuesPPR(value))
	}
	return fcn, nil
}
