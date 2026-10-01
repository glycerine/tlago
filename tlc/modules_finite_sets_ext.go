package tlc

func FiniteSetsExtQuantify(set Value, test Value) (Value, error) {
	enumerable, ok := asEnumerable(set)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "Quantify", "set", ValuesPPR(set))
	}

	size := int32(0)
	args := []Value{nil}
	enum := enumerable.Elements()
	for {
		elem := enum.NextElement()
		if elem == nil {
			if err := enum.Err(); err != nil {
				return nil, err
			}
			return NewIntValue(size), nil
		}
		args[0] = elem
		value, err := EvalOperatorValue(test, args, EvalClear)
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
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "kSubset", "natural number", ValuesPPR(kValue))
	}
	if _, ok := asEnumerable(set); !ok {
		if _, userValue := set.(*UserValue); !userValue {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "kSubset", "set", ValuesPPR(set))
		}
	}
	return NewKSubsetValue(int(k.Val), set), nil
}

func FiniteSetsExtFoldSet(op Value, base Value, set Value) (Value, error) {
	enumerable, ok := asEnumerable(set)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldSet", "set", ValuesPPR(set))
	}

	args := []Value{nil, base}
	enum := enumerable.Elements()
	for {
		elem := enum.NextElement()
		if elem == nil {
			if err := enum.Err(); err != nil {
				return nil, err
			}
			return args[1], nil
		}
		args[0] = elem
		var err error
		args[1], err = EvalOperatorValue(op, args, EvalClear)
		if err != nil {
			return nil, err
		}
	}
}
