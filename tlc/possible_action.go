package tlc

func NewPossibleAction(pred SemanticNode, con *Context, userPredicate *OpDefNode) *Action {
	name := ""
	if userPredicate != nil && userPredicate.Name != nil {
		name = userPredicate.Name.String()
	}
	action := NewAction(pred, con, name)
	action.Possible = userPredicate
	return action
}

func (t *Tool) evalPossibleTrackNode(node *PossibleTrackNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if node == nil {
		return BoolTrue, nil
	}
	value, err := t.Eval(node.Pred, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	boolValue, ok := value.(*BoolValue)
	if !ok {
		if value == nil {
			return nil, newTLCError(ECGeneral, "_POSSIBLE predicate %s evaluated to nil", node.Name)
		}
		return nil, newTLCError(ECGeneral, "_POSSIBLE predicate %s evaluated to %s instead of a boolean", node.Name, value.KindString())
	}
	next, err := possibleCountsWith(node.Name, boolValue.Val, possibleLocalCounts(s0))
	if err != nil {
		return nil, err
	}
	setPossibleLocalCounts(s0, next)
	return BoolTrue, nil
}

func (t *Tool) evalPossibleCheckNode(node *PossibleCheckNode) (Value, error) {
	if node == nil {
		return BoolTrue, nil
	}
	counts := PossibleCounts()
	count := possibleCountForName(counts, node.Name)
	if count == nil || count.Val == 0 {
		TLCPrintT(counts)
		return BoolFalse, nil
	}
	return BoolTrue, nil
}

func possibleLocalCounts(state *TLCStateMut) Value {
	workerID := workerIDFromState(state)
	if checker := MainChecker(); checker != nil {
		if value := checker.GetNamedValue(workerID, possibleCountsKey); value != nil {
			return value
		}
	}
	if simulator := CurrentSimulator(); simulator != nil {
		if value := simulator.GetLocalNamedValue(possibleCountsKey); value != nil {
			return value
		}
	}
	return EmptyFcn
}

func setPossibleLocalCounts(state *TLCStateMut, value Value) {
	if checker := MainChecker(); checker != nil {
		checker.SetNamedValue(workerIDFromState(state), possibleCountsKey, value)
		return
	}
	if simulator := CurrentSimulator(); simulator != nil {
		simulator.SetAllNamedValues(possibleCountsKey, value)
	}
}

func possibleCountsWith(name string, witnessed bool, current Value) (*FcnRcdValue, error) {
	fcn := asFcnRcdValue(current)
	if fcn == nil {
		fcn = EmptyFcn
	}
	domain := append([]Value(nil), fcn.DomainAsValues()...)
	values := append([]Value(nil), fcn.Values...)
	nameValue := NewStringValue(name)
	idx := -1
	for i, value := range domain {
		eq, err := value.Equal(nameValue)
		if err != nil {
			return nil, err
		}
		if eq {
			idx = i
			break
		}
	}
	if idx < 0 {
		count := int32(0)
		if witnessed {
			count = 1
		}
		domain = append(domain, nameValue)
		values = append(values, NewIntValue(count))
		return NewFcnRcdValue(domain, values, false), nil
	}
	if witnessed {
		old, ok := values[idx].(*IntValue)
		if !ok {
			return nil, newTLCError(ECGeneral, "_POSSIBLE count for %s is not an integer", name)
		}
		values[idx] = NewIntValue(old.Val + 1)
	}
	return NewFcnRcdValue(domain, values, false), nil
}

func possibleCountForName(counts Value, name string) *IntValue {
	fcn := asFcnRcdValue(counts)
	if fcn == nil {
		return nil
	}
	value, err := fcn.Select(NewStringValue(name))
	if err != nil {
		return nil
	}
	count, _ := value.(*IntValue)
	return count
}
