package tlc

const javaTLCRandomElementSignature = "public static tlc2.value.impl.Value tlc2.module.TLC.RandomElement(tlc2.value.impl.Value)"

func SortSeq(seq Value, cmp Value) (Value, error) {
	tuple := asTupleValue(seq)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "SortSeq", "natural number", ValuesPPR(seq))
	}
	if !isOperatorValue(cmp) {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SortSeq", "operator", ValuesPPR(cmp))
	}
	if len(tuple.Elems) == 0 {
		return tuple, nil
	}
	newElems := make([]Value, len(tuple.Elems))
	newElems[0] = tuple.Elems[0]
	for i := 1; i < len(tuple.Elems); i++ {
		j := i
		arg0 := tuple.Elems[i]
		for j > 0 {
			less, err := compareWithOperator(cmp, arg0, newElems[j-1])
			if err != nil {
				return nil, err
			}
			if !less {
				break
			}
			newElems[j] = newElems[j-1]
			j--
		}
		newElems[j] = arg0
	}
	return NewTupleValue(newElems), nil
}

func compareWithOperator(cmp Value, left Value, right Value) (bool, error) {
	res, err := EvalOperatorValue(cmp, []Value{left, right}, EvalClear)
	if err != nil {
		return false, err
	}
	boolValue, ok := res.(*BoolValue)
	if !ok {
		return false, newTLCErrorCode(ECTLCModuleArgumentError, "second", "SortSeq", "boolean function", ValuesPPR(res))
	}
	return boolValue.Val, nil
}

func TLCEval(value Value) Value {
	converted, err := TLCEvalChecked(value)
	if err != nil {
		panic(err)
	}
	return converted
}

func TLCEvalChecked(value Value) (Value, error) {
	if value == nil {
		return ValUndef, nil
	}
	if canConvertToSetEnum(value) {
		set, err := toSetEnumValue(value)
		if err != nil {
			return nil, err
		}
		if set != nil {
			return set, nil
		}
	}
	if fcn := asFcnRcdValue(value); fcn != nil {
		return fcn, nil
	}
	return value, nil
}

func canConvertToSetEnum(value Value) bool {
	switch value.(type) {
	case *SetEnumValue, *SetPredValue, *IntervalValue, *SetCupValue, *SetCapValue, *SetDiffValue, *UnionValue, *SetOfTuplesValue, *SetOfRcdsValue, *SetOfFcnsValue, *SubsetValue, *KSubsetValue:
		return true
	default:
		return false
	}
}

func RandomElement(value Value) (Value, error) {
	switch v := value.(type) {
	case *SetOfFcnsValue:
		v.Normalize()
		domain, err := toSetEnumValue(v.Domain)
		if err != nil {
			return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "RandomElement", "a finite set", ValuesPPR(value))
		}
		domain.Normalize()
		dom := domain.Elems.ToArray()
		vals := make([]Value, len(dom))
		for i := range vals {
			elem, err := RandomElement(v.Range)
			if err != nil {
				return nil, err
			}
			vals[i] = elem
		}
		return NewFcnRcdValue(dom, vals, true), nil
	case *SetOfRcdsValue:
		v.Normalize()
		vals := make([]Value, len(v.Values))
		for i := range vals {
			elem, err := RandomElement(v.Values[i])
			if err != nil {
				return nil, err
			}
			vals[i] = elem
		}
		return NewRecordValue(v.Names, vals, true), nil
	case *SetOfTuplesValue:
		v.Normalize()
		vals := make([]Value, len(v.Sets))
		for i := range vals {
			elem, err := RandomElement(v.Sets[i])
			if err != nil {
				return nil, err
			}
			vals[i] = elem
		}
		return NewTupleValue(vals), nil
	case *IntervalValue:
		size, err := v.Size()
		if err != nil {
			return nil, err
		}
		if size == 0 {
			return nil, javaMethodOverrideError(javaTLCRandomElementSignature, "Attempted to retrieve out-of-bounds element from the interval value "+ValuesPPR(v)+".")
		}
		index := int(RandomEnumerableGenerator().NextDouble() * float64(size))
		return NewIntValue(v.Low + int32(index)), nil
	default:
		set, err := toSetEnumValue(value)
		if err != nil {
			return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "RandomElement", "a finite set", ValuesPPR(value))
		}
		if _, err := set.normalizeSet(); err != nil {
			return nil, err
		}
		if set.Elems.Len() == 0 {
			return nil, javaMethodOverrideError(javaTLCRandomElementSignature, "Index 0 out of bounds for length 0")
		}
		index := int(RandomEnumerableGenerator().NextDouble() * float64(set.Elems.Len()))
		return set.Elems.At(index), nil
	}
}

func Any() Value {
	return AnySetValue
}
