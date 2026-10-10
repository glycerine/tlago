package tlc

func FiniteSetsExtQuantify(set Value, test Value) (Value, error) {
	_, ok := asEnumerable(set)
	if !ok {
		if set == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "Quantify", "set", ValuesPPR(set))
	}

	size := int32(0)
	args := []Value{nil}
	converted, err := tryToSetEnumValue(set)
	if err != nil {
		return nil, err
	}
	if converted == nil {
		panic(NewNullPointerException())
	}
	enum := converted.Elements()
	for {
		elem := nextEnumerationElement(enum)
		if elem == nil {
			return NewIntValue(size), nil
		}
		args[0] = elem
		value, err := sequenceOperatorEval(test, args)
		if err != nil {
			return nil, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "Quantify", "boolean-valued operator", ValuesPPR(test))
		}
		if boolValue.Val {
			size++
		}
	}
}

func FiniteSetsExtKSubset(kValue Value, set Value) (Value, error) {
	k, ok := kValue.(*IntValue)
	if !ok {
		if kValue == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "kSubset", "natural number", ValuesPPR(kValue))
	}
	if _, ok := asEnumerable(set); !ok {
		if _, userValue := set.(*UserValue); !userValue {
			if set == nil {
				panic(NewNullPointerException())
			}
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "kSubset", "set", ValuesPPR(set))
		}
	}
	return NewKSubsetValue(int(k.Val), set, set.GetCostModel()), nil
}

func FiniteSetsExtFoldSet(op Value, base Value, set Value) (Value, error) {
	enumerable, ok := asEnumerable(set)
	if !ok {
		if set == nil {
			panic(NewNullPointerException())
		}
		return nil, NewClassCastException("Cannot cast " + javaValueClassName(set) + " to tlc2.value.impl.Enumerable")
	}

	args := []Value{nil, base}
	enum := enumerable.Elements()
	for {
		elem := nextEnumerationElement(enum)
		if elem == nil {
			return args[1], nil
		}
		args[0] = elem
		value, err := sequenceOperatorEval(op, args)
		if err != nil {
			return nil, err
		}
		args[1] = value
	}
}
