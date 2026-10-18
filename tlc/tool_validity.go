package tlc

func (t *Tool) IsInModelImpl(state *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	for _, constraint := range t.requireConstraintArray(t.GetModelConstraints()) {
		ok, err := t.IsInModelForConstraint(constraint, state)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (t *Tool) IsInModelForConstraintImpl(constraint SemanticNode, state *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.evalConstraint(constraint, state)
}

func (t *Tool) IsInActionsImpl(s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	for _, constraint := range t.requireConstraintArray(t.GetActionConstraints()) {
		ok, err := t.IsInActionsForConstraint(constraint, s1, s2)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func (t *Tool) IsInActionsForConstraintImpl(constraint SemanticNode, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.evalConstraint(constraint, s1, s2)
}

func (t *Tool) evalConstraint(constraint SemanticNode, s0 *TLCStateMut, successor ...*TLCStateMut) (bool, error) {
	if len(successor) > 1 {
		panic(NewIllegalArgumentException("evalConstraint accepts at most one successor state"))
	}
	cm := DoNotRecordCostModel
	if specCoverageEnabled() {
		object := SemanticToolObjectForTool(t, constraint)
		action, ok := object.(*Action)
		if isNil(object) || ok && action == nil {
			panic(NewNullPointerException())
		}
		if !ok {
			panic(valueStreamClassCast(object, "tlc2.tool.Action"))
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
		return false, newTLCDetailedRuntimeExceptionCode(ECTLCExpectedValue, constraint, EmptyContext, "boolean", SemanticString(constraint))
	}
	valid := bval.Val
	if specCoverageEnabled() {
		if valid {
			cm.IncSecondary()
		} else {
			cm.IncInvocations()
		}
	}
	return valid, nil
}

func (t *Tool) EvalRewardImpl(s1 *TLCStateMut, s2 *TLCStateMut, fallback float64) (float64, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	reward := t.GetRLReward()
	if semanticExploreNull(reward) {
		return fallback, nil
	}
	value, err := t.Eval(reward, EmptyContext, s1, s2, EvalClear, DoNotRecordCostModel)
	if err != nil {
		return fallback, err
	}
	intValue, ok := value.(*IntValue)
	if !ok {
		return fallback, newTLCDetailedRuntimeExceptionCode(ECTLCExpectedValue, reward, EmptyContext, "integer", SemanticString(reward))
	}
	return float64(intValue.ownedVal()), nil
}

func (t *Tool) IsValidExprImpl(expr SemanticNode, ctxt *Context) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t.evalPredicateValue(expr, ctxt, EmptyState, EmptyState, EvalConst, EmptyContext, DoNotRecordCostModel)
}

func (t *Tool) IsValidAssumption(expr SemanticNode) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if t.ownedDebugger() != nil {
		return debugToolIsValidAssumption(t, expr)
	}
	return t.IsValidExpr(expr)
}

func (t *Tool) IsValidTransitionImpl(action *Action, s0 *TLCStateMut, s1 *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if action == nil {
		return false, NewNullPointerException()
	}
	pred, con := action.ownedPred(), action.ownedContext()
	return t.evalPredicateValue(pred, con, s0, s1, EvalClear, con, action.CM)
}

func (t *Tool) IsValidStateImpl(action *Action, state *TLCStateMut) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if action == nil {
		return false, NewNullPointerException()
	}
	pred, con := action.ownedPred(), action.ownedContext()
	return t.evalPredicateValue(pred, con, state, EmptyState, EvalClear, con, action.CM)
}

func (t *Tool) IsValidActionImpl(action *Action) (bool, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if action == nil {
		return false, NewNullPointerException()
	}
	pred, con := action.ownedPred(), action.ownedContext()
	return t.evalPredicateValue(pred, con, EmptyState, EmptyState, EvalClear, con, action.CM)
}

func (t *Tool) CheckAssumptionsImpl() int {
	if t == nil {
		panic(NewNullPointerException())
	}
	for _, err := range t.ConfigErrors {
		if err != nil {
			return PrintError(err.Code, err.Params...)
		}
	}
	assumptions := t.GetAssumptions()
	semanticGraphArrayLength(t.ownedSpecProcessor() != nil, assumptions)
	axioms := t.GetAssumptionIsAxiom()
	for i, assumption := range assumptions {
		if t.ownedSpecProcessor() != nil && (axioms == nil || i >= len(axioms)) {
			if semanticExploreNull(assumption) {
				panic(NewNullPointerException())
			}
			var failure error = NewNullPointerException()
			if axioms != nil {
				failure = NewArrayIndexOutOfBoundsException(i, len(axioms))
			}
			return PrintErrorNullable(ECTLCAssumptionEvaluationError, javaString(SemanticString(assumption)), javaThrowableDetailMessage(failure))
		}
		if i < len(axioms) && axioms[i] {
			continue
		}
		if t.ownedSpecProcessor() != nil && semanticExploreNull(assumption) {
			panic(NewNullPointerException())
		}
		description := SemanticString(assumption)
		if t.ownedSpecProcessor() == nil {
			switch assumption.(type) {
			case *NumeralNode, *DecimalNode:
			default:
				if location, ok := semanticNodeSourceLocation(assumption); ok {
					description = location.String()
				}
			}
		}
		ok, err := checkPropertyPredicate(func() (bool, error) {
			return t.IsValidAssumption(assumption)
		})
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
		// Java calls the virtual isValid(expr, ctxt) overload here.  DebugTool
		// overrides that overload to enter constant-evaluation mode, allowing
		// postcondition breakpoints to suspend after action exploration.
		ok, err := checkPropertyPredicate(func() (bool, error) {
			return t.IsValidExpr(post.ownedPred(), ctxt)
		})
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

// Java catches Exception around property evaluation, including exceptions
// thrown directly by a predicate. Error subclasses escape that catch.
func checkPropertyPredicate(eval func() (bool, error)) (valid bool, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			exception, ok := failure.(error)
			if !ok || isJavaError(exception) {
				panic(failure)
			}
			valid, err = false, exception
		}
	}()
	valid, err = eval()
	if isJavaError(err) {
		panic(err)
	}
	return valid, err
}

func (t *Tool) evalPredicateValue(expr SemanticNode, ctxt *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, failureContext *Context, cm CostModel) (bool, error) {
	value, err := t.Eval(expr, ctxt, s0, s1, control, cm)
	if err != nil {
		return false, err
	}
	bval, ok := value.(*BoolValue)
	if !ok {
		return false, newTLCDetailedRuntimeExceptionCode(ECTLCExpectedValue, expr, failureContext, "boolean", SemanticString(expr))
	}
	return bval.Val, nil
}
