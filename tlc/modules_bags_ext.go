package tlc

func BagsExtFoldBag(op Value, base Value, bag Value) (Value, error) {
	if bag == nil {
		panic(NewNullPointerException())
	}
	fcn := asFcnRcdValue(bag)
	if fcn == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldBag", "bag", ValuesPPR(bag))
	}

	domain := fcn.DomainAsValues()
	values := fcn.Values
	if domain == nil {
		panic(NewNullPointerException())
	}
	args := []Value{base, nil}
	for i, elem := range domain {
		args[1] = elem
		value := fcnParameterDomain(values, i)
		count, ok := value.(*IntValue)
		if !ok || count.Val <= 0 {
			if value == nil {
				panic(NewNullPointerException())
			}
			countText := ValuesPPR(value)
			if elem == nil {
				panic(NewNullPointerException())
			}
			domainText := ValuesPPR(elem)
			lastValue := fcnParameterDomain(values, i)
			if lastValue == nil {
				panic(NewNullPointerException())
			}
			return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "FoldBag", "an element of Nat", countText+" (in: "+domainText+":>"+ValuesPPR(lastValue)+")")
		}
		for j := int32(0); j < bagMultiplicity(values, i); j++ {
			value, err := sequenceOperatorEval(op, args)
			if err != nil {
				return nil, err
			}
			args[0] = value
		}
	}
	return args[0], nil
}

// The source captures the values array, but re-reads its slot for each bound.
func bagMultiplicity(values []Value, index int) int32 {
	value := fcnParameterDomain(values, index)
	if value == nil {
		panic(NewNullPointerException())
	}
	count, ok := value.(*IntValue)
	if !ok {
		panic(valueStreamClassCast(value, "tlc2.value.impl.IntValue"))
	}
	return count.Val
}
