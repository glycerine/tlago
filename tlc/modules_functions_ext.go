package tlc

import "sort"

func FunctionsIsInjective(value Value) (*BoolValue, error) {
	switch v := value.(type) {
	case *TupleValue:
		return functionsIsInjectiveNonDestructive(v.Elems)
	case *FcnRcdValue:
		if v.Intv != nil && v.Intv.Low == 1 {
			return functionsIsInjectiveNonDestructive(v.Values)
		}
	}

	if tuple := asTupleValue(value); tuple != nil {
		return functionsIsInjectiveDestructive(tuple.Elems)
	}
	switch v := value.(type) {
	case *SetOfRcdsValue:
		return functionsIsInjectiveNonDestructive(v.Values)
	case *FcnRcdValue:
		return functionsIsInjectiveNonDestructive(v.Values)
	}
	return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "IsInjective", "function", ValuesPPR(value))
}

func functionsIsInjectiveDestructive(values []Value) (*BoolValue, error) {
	var sortErr error
	sort.Slice(values, func(i int, j int) bool {
		if sortErr != nil {
			return false
		}
		cmp, err := values[i].Compare(values[j])
		if err != nil {
			sortErr = err
			return false
		}
		return cmp < 0
	})
	if sortErr != nil {
		return nil, sortErr
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
	for i := 0; i < len(values); i++ {
		for j := i + 1; j < len(values); j++ {
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
	fcn := asFcnRcdValue(value.Normalize())
	if fcn == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "AntiFunction", "functions", ValuesPPR(value))
	}
	rangeValues := fcn.DomainAsValues()
	domain := make([]Value, len(fcn.Values))
	copy(domain, fcn.Values)
	rangeCopy := make([]Value, len(rangeValues))
	copy(rangeCopy, rangeValues)
	return NewFcnRcdValue(domain, rangeCopy, false).Normalize(), nil
}

func FunctionsFoldFunction(op Value, base Value, fun Value) (Value, error) {
	domain, _, ok, err := functionsFunctionAccess(fun)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldFunction", "function", ValuesPPR(fun))
	}
	return FunctionsFoldFunctionOnSet(op, base, fun, domain)
}

func FunctionsFoldFunctionOnSet(op Value, base Value, fun Value, subdomain Value) (Value, error) {
	_, apply, ok, err := functionsFunctionAccess(fun)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldFunctionOnSet", "function", ValuesPPR(fun))
	}
	enumerable, ok := asEnumerable(subdomain)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "fourth", "FoldFunctionOnSet", "set", ValuesPPR(subdomain))
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
		value, err := apply(elem)
		if err != nil {
			return nil, err
		}
		args[0] = value
		args[1], err = EvalOperatorValue(op, args, EvalClear)
		if err != nil {
			return nil, err
		}
	}
}

func functionsFunctionAccess(value Value) (Value, func(Value) (Value, error), bool, error) {
	switch v := value.(type) {
	case *FcnLambdaValue:
		domain, err := v.GetDomain()
		return domain, v.Apply, true, err
	case *FcnRcdValue:
		return v.DomainValue(), v.Apply, true, nil
	case *TupleValue:
		return v.Domain(), v.Apply, true, nil
	case *RecordValue:
		return v.DomainValue(), v.Apply, true, nil
	default:
		return nil, nil, false, nil
	}
}
