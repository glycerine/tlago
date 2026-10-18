package tlc

func FunctionsIsInjective(value Value) (*BoolValue, error) {
	if value == nil {
		panic(NewNullPointerException())
	}
	switch v := value.(type) {
	case *TupleValue:
		return functionsIsInjectiveNonDestructive(v.GetElems())
	case *FcnRcdValue:
		if v.ownedIntv() != nil && v.ownedIntv().Low == 1 {
			return functionsIsInjectiveNonDestructive(v.ownedValues())
		}
	}

	if tuple := asTupleValue(value); tuple != nil {
		return functionsIsInjectiveDestructive(tuple.GetElems())
	}
	switch v := value.(type) {
	case *SetOfRcdsValue:
		return functionsIsInjectiveNonDestructive(v.ownedValues())
	case *FcnRcdValue:
		return functionsIsInjectiveNonDestructive(v.ownedValues())
	}
	return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "IsInjective", "function", ValuesPPR(value))
}

func functionsIsInjectiveDestructive(values []Value) (*BoolValue, error) {
	if values == nil {
		panic(NewNullPointerException())
	}
	if err := javaSortValues(values); err != nil {
		return nil, err
	}
	for i := 1; i < len(values); i++ {
		eq, err := values[i-1].Equal(values[i])
		if err != nil {
			return nil, err
		}
		if eq {
			return BoolFalse, nil
		}
	}
	return BoolTrue, nil
}

func functionsIsInjectiveNonDestructive(values []Value) (*BoolValue, error) {
	if values == nil {
		panic(NewNullPointerException())
	}
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] == nil {
				panic(NewNullPointerException())
			}
			eq, err := values[i].Equal(values[j])
			if err != nil {
				return nil, err
			}
			if eq {
				return BoolFalse, nil
			}
		}
	}
	return BoolTrue, nil
}

func FunctionsAntiFunction(value Value) (Value, error) {
	if value == nil {
		panic(NewNullPointerException())
	}
	fcn := asFcnRcdValue(value.Normalize())
	if fcn == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "AntiFunction", "functions", ValuesPPR(value))
	}
	rangeValues := fcn.GetDomainAsValues()
	if fcn.ownedIntv() == nil {
		if rangeValues == nil {
			panic(NewNullPointerException())
		}
		rangeCopy := make([]Value, len(rangeValues))
		copy(rangeCopy, rangeValues)
		rangeValues = rangeCopy
	}
	if fcn.ownedValues() == nil {
		panic(NewNullPointerException())
	}
	domain := make([]Value, len(fcn.ownedValues()))
	copy(domain, fcn.ownedValues())
	return NewFcnRcdValue(domain, rangeValues, false).Normalize(), nil
}

func FunctionsFoldFunction(op Value, base Value, fun Value) (Value, error) {
	getDomain, _, ok := functionsFunctionAccess(fun)
	if !ok {
		if fun == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldFunction", "function", ValuesPPR(fun))
	}
	domain, err := getDomain()
	if err != nil {
		return nil, err
	}
	return FunctionsFoldFunctionOnSet(op, base, fun, domain)
}

func FunctionsFoldFunctionOnSet(op Value, base Value, fun Value, subdomain Value) (Value, error) {
	_, apply, ok := functionsFunctionAccess(fun)
	if !ok {
		if fun == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldFunctionOnSet", "function", ValuesPPR(fun))
	}
	enumerable, ok := asEnumerable(subdomain)
	if !ok {
		if subdomain == nil {
			panic(NewNullPointerException())
		}
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "fourth", "FoldFunctionOnSet", "set", ValuesPPR(subdomain))
	}

	args := []Value{nil, base}
	enum := enumerable.Elements()
	for {
		elem := nextEnumerationElement(enum)
		if elem == nil {
			return args[1], nil
		}
		value, err := apply(elem)
		if err != nil {
			return nil, err
		}
		args[0] = value
		args[1], err = sequenceOperatorEval(op, args)
		if err != nil {
			return nil, err
		}
	}
}

// Selecting a function must not read or normalize its domain. Only FoldFunction
// invokes the domain getter; FoldFunctionOnSet uses the caller's subdomain.
func functionsFunctionAccess(value Value) (func() (Value, error), func(Value) (Value, error), bool) {
	switch v := value.(type) {
	case *CounterExample:
		record := asRecordValue(v)
		return func() (Value, error) { return record.DomainValue(), nil }, record.Apply, true
	case FunctionValue:
		return v.GetDomain, func(arg Value) (Value, error) {
			return v.ApplyWithControl(arg, EvalClear)
		}, true
	default:
		return nil, nil, false
	}
}
