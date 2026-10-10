package tlc

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
	if b1 == nil {
		panic(NewNullPointerException())
	}
	fcn1 := asFcnRcdValue(b1)
	if b2 == nil {
		panic(NewNullPointerException())
	}
	fcn2 := asFcnRcdValue(b2)
	// Both conversions precede validation. A failed conversion stays null until
	// IsABag dereferences it; the first bag's count error can precede that read.
	if fcn1 == nil {
		panic(NewNullPointerException())
	}
	valid, err := IsABag(fcn1)
	if err != nil {
		return nil, err
	}
	if !valid.Val {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "(+)", "bag", ValuesPPR(b1))
	}
	if fcn2 == nil {
		panic(NewNullPointerException())
	}
	valid, err = IsABag(fcn2)
	if err != nil {
		return nil, err
	}
	if !valid.Val {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "(+)", "bag", ValuesPPR(b2))
	}

	domain1 := fcn1.DomainAsValues()
	values1 := fcn1.Values
	domain2 := fcn2.DomainAsValues()
	values2 := fcn2.Values
	if domain1 == nil {
		panic(NewNullPointerException())
	}
	domain := NewValueVec(len(domain1))
	values := NewValueVec(len(domain1))
	for i := 0; i < len(domain1); i++ {
		domain.Add(domain1[i])
		values.Add(fcnParameterDomain(values1, i))
	}
	if domain2 == nil {
		panic(NewNullPointerException())
	}
	for i := 0; i < len(domain2); i++ {
		found := false
		for j := 0; j < len(domain1); j++ {
			// Equality may normalize arrays retained by either operand. Read the
			// captured domain and original multiplicities again after those effects.
			right, left := domain2[i], domain1[j]
			if right == nil {
				panic(NewNullPointerException())
			}
			eq, err := right.Equal(left)
			if err != nil {
				return nil, err
			}
			if eq {
				leftCount := bagMultiplicity(values1, j)
				rightCount := bagMultiplicity(values2, i)
				values.Set(j, NewIntValue(leftCount+rightCount))
				found = true
				break
			}
		}
		if !found {
			domain.Add(domain2[i])
			values.Add(fcnParameterDomain(values2, i))
		}
	}
	return NewFcnRcdValue(domain.ToArray(), values.ToArray(), false), nil
}

func BagDiff(b1 Value, b2 Value) (Value, error) {
	if b1 == nil {
		panic(NewNullPointerException())
	}
	fcn1 := asFcnRcdValue(b1)
	if b2 == nil {
		panic(NewNullPointerException())
	}
	fcn2 := asFcnRcdValue(b2)
	if fcn1 == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "(-)", "bag", ValuesPPR(b1))
	}
	if fcn2 == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "(-)", "bag", ValuesPPR(b2))
	}
	domain1 := fcn1.DomainAsValues()
	values1 := fcn1.Values
	domain2 := fcn2.DomainAsValues()
	values2 := fcn2.Values
	if domain1 == nil {
		panic(NewNullPointerException())
	}
	domain := NewValueVec(len(domain1))
	values := NewValueVec(len(domain1))
	for i := 0; i < len(domain1); i++ {
		// Capture the count before equality can normalize retained arrays.
		v1 := bagMultiplicity(values1, i)
		// An empty first domain bypasses this read; the count cast precedes it.
		if domain2 == nil {
			panic(NewNullPointerException())
		}
		for j := 0; j < len(domain2); j++ {
			left, right := domain1[i], domain2[j]
			if left == nil {
				panic(NewNullPointerException())
			}
			eq, err := left.Equal(right)
			if err != nil {
				return nil, err
			}
			if eq {
				v1 -= bagMultiplicity(values2, j)
				break
			}
		}
		if v1 > 0 {
			domain.Add(domain1[i])
			values.Add(NewIntValue(v1))
		}
	}
	return NewFcnRcdValue(domain.ToArray(), values.ToArray(), fcn1.IsNormalized()), nil
}

func BagUnion(set Value) (Value, error) {
	if set == nil {
		panic(NewNullPointerException())
	}
	setEnum, err := tryToSetEnumValue(set)
	if err != nil {
		return nil, err
	}
	if setEnum == nil {
		return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "BagUnion", "a finite enumerable set", ValuesPPR(set))
	}
	setEnum.Normalize()
	// Normalization can collapse equal bags. The source retains this vector
	// and its size for the remaining conversions and aggregation.
	elems := setEnum.Elems
	size := elems.Len()
	if size == 0 {
		return emptyFcnValue(), nil
	}
	if size == 1 {
		return elems.At(0), nil
	}
	domain := NewValueVec(0)
	values := NewValueVec(0)
	firstValue := elems.At(0)
	if firstValue == nil {
		panic(NewNullPointerException())
	}
	first := asFcnRcdValue(firstValue)
	if first == nil {
		return nil, newTLCErrorCode(ECTLCModuleBagUnion1, ValuesPPR(set))
	}
	currentDomain := first.DomainAsValues()
	currentValues := first.Values
	if currentDomain == nil {
		panic(NewNullPointerException())
	}
	for i := 0; i < len(currentDomain); i++ {
		domain.Add(currentDomain[i])
		values.Add(fcnParameterDomain(currentValues, i))
	}
	for i := 1; i < size; i++ {
		item := elems.At(i)
		if item == nil {
			panic(NewNullPointerException())
		}
		fcn := asFcnRcdValue(item)
		if fcn == nil {
			return nil, newTLCErrorCode(ECTLCModuleBagUnion1, ValuesPPR(set))
		}
		currentDomain = fcn.DomainAsValues()
		currentValues = fcn.Values
		if currentDomain == nil {
			panic(NewNullPointerException())
		}
		for j := 0; j < len(currentDomain); j++ {
			found := false
			for k := 0; k < domain.Len(); k++ {
				key, prior := currentDomain[j], domain.At(k)
				if key == nil {
					panic(NewNullPointerException())
				}
				eq, err := key.Equal(prior)
				if err != nil {
					return nil, err
				}
				if eq {
					leftValue := values.At(k)
					if leftValue == nil {
						panic(NewNullPointerException())
					}
					left, ok := leftValue.(*IntValue)
					if !ok {
						panic(valueStreamClassCast(leftValue, "tlc2.value.impl.IntValue"))
					}
					right := bagMultiplicity(currentValues, j)
					values.Set(k, NewIntValue(left.Val+right))
					found = true
					break
				}
			}
			if !found {
				domain.Add(currentDomain[j])
				values.Add(fcnParameterDomain(currentValues, j))
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
