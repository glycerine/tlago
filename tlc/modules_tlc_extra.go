package tlc

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
		panic(NewNullPointerException())
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
	if value == nil {
		panic(NewNullPointerException())
	}
	switch v := value.(type) {
	case *SetOfFcnsValue:
		if v == nil {
			panic(NewNullPointerException())
		}
		v.Normalize()
		if v.Domain == nil {
			panic(NewNullPointerException())
		}
		domain, err := tryToSetEnumValue(v.Domain)
		if err != nil {
			return nil, err
		}
		if domain == nil {
			return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "RandomElement", "a finite set", ValuesPPR(value))
		}
		domain.Normalize()
		elems := domain.Elems
		dom := make([]Value, valueStreamArrayLength(int32(elems.Len())))
		vals := make([]Value, valueStreamArrayLength(int32(elems.Len())))
		for i := range vals {
			dom[i] = elems.At(i)
			elem, err := RandomElement(v.Range)
			if err != nil {
				return nil, err
			}
			vals[i] = elem
		}
		return NewFcnRcdValue(dom, vals, true), nil
	case *SetOfRcdsValue:
		if v == nil {
			panic(NewNullPointerException())
		}
		v.Normalize()
		if v.Names == nil {
			panic(NewNullPointerException())
		}
		vals := make([]Value, len(v.Names))
		for i := range vals {
			elem, err := RandomElement(v.fieldValue(i))
			if err != nil {
				return nil, err
			}
			vals[i] = elem
		}
		return NewRecordValue(v.Names, vals, true), nil
	case *SetOfTuplesValue:
		if v == nil {
			panic(NewNullPointerException())
		}
		v.Normalize()
		if v.Sets == nil {
			panic(NewNullPointerException())
		}
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
		if v == nil {
			panic(NewNullPointerException())
		}
		return v.RandomElement()
	default:
		if set, ok := value.(*SetEnumValue); ok && set == nil {
			panic(NewNullPointerException())
		}
		set, err := tryToSetEnumValue(value)
		if err != nil {
			return nil, err
		}
		if set == nil {
			return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "RandomElement", "a finite set", ValuesPPR(value))
		}
		return set.RandomElement()
	}
}

func Any() Value {
	return AnySetValue
}
