package tlc

func (t *Tool) IsInModelImpl(state *TLCStateMut) (bool, error) {
	for _, constraint := range t.requireConstraintArray(t.GetModelConstraints()) {
		ok, err := t.IsInModelForConstraint(constraint, state)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (t *Tool) IsInModelForConstraintImpl(constraint SemanticNode, state *TLCStateMut) (bool, error) {
	return t.evalConstraint(constraint, state)
}

func (t *Tool) IsInActionsImpl(s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	for _, constraint := range t.requireConstraintArray(t.GetActionConstraints()) {
		ok, err := t.IsInActionsForConstraint(constraint, s1, s2)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (t *Tool) IsInActionsForConstraintImpl(constraint SemanticNode, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	return t.evalConstraint(constraint, s1, s2)
}

func (t *Tool) evalConstraint(constraint SemanticNode, s0 *TLCStateMut, successor ...*TLCStateMut) (bool, error) {
	cm := DoNotRecordCostModel
	if CoverageEnabled() {
		object := SemanticToolObjectForTool(t, constraint)
		action, ok := object.(*Action)
		if object == nil || ok && action == nil {
			panic(NewNullPointerException())
		}
		if !ok {
			panic(NewClassCastException())
		}
		cm = action.CM
	}
	// Source isInModel uses the one-state eval overload, which selects State
	// mode in DebugTool. Action constraints use the two-state/control overload.
	args := []any{EmptyContext, s0, cm}
	if len(successor) > 0 {
		args = []any{EmptyContext, s0, successor[0], EvalClear, cm}
	}
	value, err := t.Eval(constraint, args...)
	if err != nil {
		return false, err
	}
	bval, ok := value.(*BoolValue)
	if !ok {
		return false, NewTLCRuntimeException(ECTLCExpectedValue, "boolean", semanticNodeLocationString(constraint))
	}
	valid := bval.Val
	if CoverageEnabled() {
		if valid {
			cm.IncSecondary()
		} else {
			cm.IncInvocations()
		}
	}
	return valid, nil
}

func (t *Tool) EvalRewardImpl(s1 *TLCStateMut, s2 *TLCStateMut, fallback float64) (float64, error) {
	if t == nil || t.RLReward == nil {
		return fallback, nil
	}
	value, err := t.Eval(t.RLReward, EmptyContext, s1, s2, EvalClear, CostModel{})
	if err != nil {
		return fallback, err
	}
	intValue, ok := value.(*IntValue)
	if !ok {
		return fallback, newTLCErrorCode(ECTLCExpectedValue, "integer", SemanticString(t.RLReward))
	}
	return float64(intValue.Val), nil
}

func (t *Tool) IsValidExprImpl(expr SemanticNode, ctxt *Context) (bool, error) {
	if ctxt == nil {
		ctxt = EmptyContext
	}
	return t.evalPredicateValue(expr, ctxt, EmptyState, EmptyState, EvalConst, "constant expression")
}

func (t *Tool) IsValidTransitionImpl(action *Action, s0 *TLCStateMut, s1 *TLCStateMut) (bool, error) {
	if action == nil {
		return false, newTLCError(ECGeneral, "cannot validate nil action")
	}
	return t.evalPredicateValue(action.Pred, action.Con, s0, s1, EvalClear, action.GetName(), action.CM)
}

func (t *Tool) IsValidStateImpl(action *Action, state *TLCStateMut) (bool, error) {
	if action == nil {
		return false, newTLCError(ECGeneral, "cannot validate nil action")
	}
	return t.evalPredicateValue(action.Pred, action.Con, state, EmptyState, EvalClear, action.GetName(), action.CM)
}

func (t *Tool) IsValidActionImpl(action *Action) (bool, error) {
	if action == nil {
		return false, newTLCError(ECGeneral, "cannot validate nil action")
	}
	return t.evalPredicateValue(action.Pred, action.Con, EmptyState, EmptyState, EvalClear, action.GetName(), action.CM)
}

func (t *Tool) CheckAssumptionsImpl() int {
	for _, err := range t.ConfigErrors {
		if err != nil {
			return PrintError(err.Code, err.Params...)
		}
	}
	for i, assumption := range t.Assumptions {
		if i < len(t.AssumptionIsAxiom) && t.AssumptionIsAxiom[i] {
			continue
		}
		description := SemanticString(assumption)
		switch assumption.(type) {
		case *NumeralNode, *DecimalNode:
		default:
			if location, ok := semanticNodeSourceLocation(assumption); ok {
				description = location.String()
			}
		}
		ok, err := t.IsValidExpr(assumption, EmptyContext)
		if err != nil {
			return PrintErrorNullable(ECTLCAssumptionEvaluationError, javaString(description), javaThrowableDetailMessage(err))
		}
		if !ok {
			return PrintError(ECTLCAssumptionFalse, description)
		}
	}
	return NoError
}

func (t *Tool) CheckPostConditionImpl(ctxt *Context) int {
	if ctxt == nil {
		ctxt = EmptyContext
	}
	for _, post := range t.GetPostConditionSpecs() {
		ok, err := t.evalPredicateValue(post.Pred, ctxt, EmptyState, EmptyState, EvalConst, post.GetName())
		if err != nil {
			return PrintErrorNullable(ECTLCPostconditionEvaluationError, javaString(post.GetName()), javaString(postConditionPredicateString(post)), javaThrowableDetailMessage(err))
		}
		if !ok {
			if post.IsPossible() {
				return PrintError(ECTLCPossibleUnwitnessed, post.GetName(), postConditionPredicateString(post))
			}
			return PrintError(ECTLCPostconditionFalse, post.GetName(), postConditionPredicateString(post))
		}
	}
	return NoError
}

func postConditionPredicateString(post *Action) string {
	if post == nil {
		return ""
	}
	return SemanticString(post.GetPred())
}

func (t *Tool) evalPredicateValue(expr SemanticNode, ctxt *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, where string, costModels ...CostModel) (bool, error) {
	if ctxt == nil {
		ctxt = EmptyContext
	}
	cm := DoNotRecordCostModel
	if len(costModels) > 0 {
		cm = costModels[0]
	}
	value, err := t.Eval(expr, ctxt, s0, s1, control, cm)
	if err != nil {
		return false, err
	}
	bval, ok := value.(*BoolValue)
	if !ok {
		return false, NewTLCRuntimeException(ECTLCExpectedValue, "boolean", semanticNodeLocationString(expr))
	}
	return bval.Val, nil
}
