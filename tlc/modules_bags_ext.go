package tlc

func BagsExtFoldBag(op Value, base Value, bag Value) (Value, error) {
	fcn := asFcnRcdValue(bag)
	if fcn == nil {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "FoldBag", "bag", ValuesPPR(bag))
	}

	domain := fcn.DomainAsValues()
	args := []Value{base, nil}
	for i, elem := range domain {
		count, ok := fcn.Values[i].(*IntValue)
		if !ok || count.Val <= 0 {
			return nil, newTLCErrorCode(ECTLCModuleApplyingToWrongValue, "FoldBag", "an element of Nat", ValuesPPR(fcn.Values[i])+" (in: "+ValuesPPR(elem)+":>"+ValuesPPR(fcn.Values[i])+")")
		}
		args[1] = elem
		for j := int32(0); j < count.Val; j++ {
			value, err := EvalOperatorValue(op, args, EvalClear)
			if err != nil {
				return nil, err
			}
			args[0] = value
		}
	}
	return args[0], nil
}
