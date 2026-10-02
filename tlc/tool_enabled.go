package tlc

func (t *Tool) EnabledImpl(pred SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (state *TLCStateMut, err error) {
	done := t.callStackEnter(pred)
	defer func() { done(err) }()
	if c == nil {
		c = EmptyContext
	}
	if acts == nil {
		acts = EmptyActionItemList
	}
	if s1 == nil {
		s1 = NewEmptyState()
	}
	switch pred := pred.(type) {
	case *OpApplNode:
		return t.EnabledAppl(pred, acts, c, s0, s1, cm)
	case *LetInNode:
		return t.EnabledImpl(pred.Body, acts, letDefinitionsContextWithCostModel(c, pred.Lets, cm, pred.Bindings...), s0, s1, cm)
	case *SubstInNode:
		c1 := c
		for _, sub := range pred.Substs {
			subCM := cm
			if CoverageEnabled() {
				subCM = cm.GetSubst(sub)
			}
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, subCM))
		}
		return t.EnabledImpl(pred.Body, acts, c1, s0, s1, cm)
	case *APSubstInNode:
		c1 := c
		for _, sub := range pred.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, cm))
		}
		return t.EnabledImpl(pred.Body, acts, c1, s0, s1, cm)
	case *LabelNode:
		return t.EnabledImpl(pred.Body, acts, c, s0, s1, cm)
	default:
		return nil, newTLCError(ECGeneral, "attempted to compute ENABLED on a non-boolean expression: %s", SemanticString(pred))
	}
}

func (t *Tool) EnabledFromActionList(acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	if acts == nil || acts.IsEmpty() {
		return s1, nil
	}
	kind := acts.CarKind()
	pred := acts.CarPred()
	c := acts.CarContext()
	acts1 := acts.Cdr()
	switch {
	case kind > ActionItemConjunct || kind == ActionItemPred:
		return t.EnabledImpl(pred, acts1, c, s0, s1, acts.CM)
	case kind == ActionItemUnchanged:
		return t.EnabledUnchanged(pred, acts1, c, s0, s1, acts.CM)
	case kind == ActionItemChanged:
		v1, err := t.Eval(pred, c, s0, EmptyState, EvalEnabled, acts.CM)
		if err != nil {
			return nil, err
		}
		v2, err := t.Eval(pred, c, s1, nil, EvalPrimed, acts.CM)
		if err != nil {
			return nil, err
		}
		eq, err := v1.Equal(v2)
		if err != nil || eq {
			return nil, err
		}
		return t.EnabledFromActionList(acts1, s0, s1, cm)
	default:
		return t.EnabledImpl(pred, acts1, c, s0, s1, acts.CM)
	}
}

func (t *Tool) EnabledAppl(pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (state *TLCStateMut, err error) {
	done := t.callStackEnter(pred)
	defer func() { done(err) }()
	if CoverageEnabled() {
		cm = cm.Get(pred)
	}
	args := pred.Args
	opNode := pred.Operator
	opcode := GetOpCode(opNode.Name)
	if opcode == 0 {
		val := t.Lookup(opNode, c, s0, false)
		switch v := val.(type) {
		case *OpDefNode:
			opcode = GetOpCode(v.Name)
			if opcode == 0 {
				c1, err := t.GetOpContext(v, args, c, true, cm)
				if err != nil {
					return nil, err
				}
				return t.EnabledImpl(v.Body, acts, c1, s0, s1, cm)
			}
		case *ThmOrAssumpDefNode:
			c1, err := t.GetThmOrAssumpContext(v, args, c, true)
			if err != nil {
				return nil, err
			}
			return t.EnabledImpl(v.Body, acts, c1, s0, s1, cm)
		case *LazyValue:
			return t.EnabledImpl(v.Expr, acts, v.Con, s0, s1, v.CM)
		case *LazySupplierValue:
			lv := asLazyValue(v)
			return t.EnabledImpl(lv.Expr, acts, lv.Con, s0, s1, lv.CM)
		case Value:
			if isOperatorValue(v) {
				bval, err := EvalOperatorValueWithTool(v, t, args, c, s0, s1, EvalEnabled, cm)
				if err != nil {
					return nil, err
				}
				return t.enabledContinueIfBool(pred, bval, acts, s0, s1, cm)
			}
			return t.enabledContinueIfBool(pred, v, acts, s0, s1, cm)
		default:
			if val == nil {
				return nil, newTLCError(ECGeneral, "undefined operator in ENABLED expression: %s", opNode)
			}
			return nil, newTLCError(ECGeneral, "ENABLED operator %s evaluated to %T", opNode, val)
		}
	}

	switch opcode {
	case OpcodeAA:
		acts1 := acts.Cons(args[1], c, cm, ActionItemChanged)
		return t.EnabledImpl(args[0], acts1, c, s0, s1, cm)
	case OpcodeBE:
		enum, err := t.Contexts(pred, c, s0, s1, EvalEnabled, cm)
		if err != nil {
			return nil, err
		}
		for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
			s2, err := t.EnabledImpl(args[0], acts, c1, s0, s1, cm)
			if err != nil || s2 != nil {
				return s2, err
			}
		}
		return nil, enum.Err()
	case OpcodeBF:
		return t.enabledBoundedForall(pred, acts, c, s0, s1, cm)
	case OpcodeCase:
		return t.enabledCase(pred, acts, c, s0, s1, cm)
	case OpcodeFA:
		return t.enabledFcnApply(pred, acts, c, s0, s1, cm)
	case OpcodeCL, OpcodeLand:
		if len(args) == 0 {
			return t.EnabledFromActionList(acts, s0, s1, cm)
		}
		acts1 := acts
		for i := len(args) - 1; i > 0; i-- {
			acts1 = acts1.Cons(args[i], c, cm, i)
		}
		return t.EnabledImpl(args[0], acts1, c, s0, s1, cm)
	case OpcodeDL, OpcodeLor:
		if len(args) == 0 {
			return nil, nil
		}
		for _, arg := range args {
			s2, err := t.EnabledImpl(arg, acts, c, s0, s1, cm)
			if err != nil || s2 != nil {
				return s2, err
			}
		}
		return nil, nil
	case OpcodeITE:
		guard, err := t.Eval(args[0], c, s0, s1, EvalEnabled, cm)
		if err != nil {
			return nil, err
		}
		bguard, ok := guard.(*BoolValue)
		if !ok {
			return nil, newTLCError(ECGeneral, "In computing ENABLED, a non-boolean expression(%s) was used as the guard condition of an IF.\n%s", guard.KindString(), SemanticString(pred))
		}
		idx := 2
		if bguard.Val {
			idx = 1
		}
		return t.EnabledImpl(args[idx], acts, c, s0, s1, cm)
	case OpcodeSA:
		s2, err := t.EnabledImpl(args[0], acts, c, s0, s1, cm)
		if err != nil || s2 != nil {
			return s2, err
		}
		return t.EnabledUnchanged(args[1], acts, c, s0, s1, cm)
	case OpcodeUnchanged:
		return t.EnabledUnchanged(args[0], acts, c, s0, s1, cm)
	case OpcodeEq:
		return t.enabledEquality(pred, args[0], args[1], acts, c, s0, s1, cm)
	case OpcodeSubseteq:
		return t.enabledSubsetEq(pred, args[0], args[1], acts, c, s0, s1, cm)
	case OpcodeIn:
		return t.enabledMembership(pred, args[0], args[1], acts, c, s0, s1, cm)
	case OpcodeImplies:
		value, err := t.Eval(args[0], c, s0, s1, EvalEnabled, cm)
		if err != nil {
			return nil, err
		}
		bval, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCError(ECGeneral, "While computing ENABLED of an expression of the form P => Q, P was %s.\n%s", valueKindString(value), SemanticString(pred))
		}
		if bval.Val {
			return t.EnabledImpl(args[1], acts, c, s0, s1, cm)
		}
		return t.EnabledFromActionList(acts, s0, s1, cm)
	case OpcodeNop:
		return t.EnabledImpl(args[0], acts, c, s0, s1, cm)
	case OpcodeCdot:
		return t.enabledActionComposition(args, acts, c, s0, s1, cm)
	case OpcodeTE, OpcodeTF:
		return nil, newTLCError(ECGeneral, "In computing ENABLED, TLC encountered temporal quantifier.\n%s", SemanticString(pred))
	case OpcodeUC:
		return nil, newTLCError(ECGeneral, "In computing ENABLED, TLC encountered unbounded CHOOSE. Make sure that the expression is of form CHOOSE x \\in S: P(x).\n%s", SemanticString(pred))
	case OpcodeUE:
		return nil, newTLCError(ECGeneral, "In computing ENABLED, TLC encountered unbounded quantifier. Make sure that the expression is of form \\E x \\in S: P(x).\n%s", SemanticString(pred))
	case OpcodeUF:
		return nil, newTLCError(ECGeneral, "In computing ENABLED, TLC encountered unbounded quantifier. Make sure that the expression is of form \\A x \\in S: P(x).\n%s", SemanticString(pred))
	case OpcodeSF:
		return nil, newTLCErrorCode(ECTLCEnabledWrongFormula, "SF", SemanticString(pred))
	case OpcodeWF:
		return nil, newTLCErrorCode(ECTLCEnabledWrongFormula, "WF", SemanticString(pred))
	case OpcodeBox:
		return nil, newTLCErrorCode(ECTLCEnabledWrongFormula, "[]", SemanticString(pred))
	case OpcodeDiamond:
		return nil, newTLCErrorCode(ECTLCEnabledWrongFormula, "<>", SemanticString(pred))
	case OpcodeLeadsto:
		return nil, newTLCError(ECGeneral, "In computing ENABLED, TLC encountered a temporal formula (a ~> b).\n%s", SemanticString(pred))
	case OpcodeArrow:
		return nil, newTLCError(ECGeneral, "In computing ENABLED, TLC encountered a temporal formula (a -+-> formula).\n%s", SemanticString(pred))
	default:
		bval, err := t.Eval(pred, c, s0, s1, EvalEnabled, cm)
		if err != nil {
			return nil, err
		}
		return t.enabledContinueIfBool(pred, bval, acts, s0, s1, cm)
	}
}

func (t *Tool) enabledContinueIfBool(pred SemanticNode, value Value, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	bval, ok := value.(*BoolValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCExpectedExpressionInComputing, "ENABLED", "boolean", value.String(), SemanticString(pred))
	}
	if bval.Val {
		return t.EnabledFromActionList(acts, s0, s1, cm)
	}
	return nil, nil
}

func (t *Tool) enabledActionComposition(args []SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	if !actionCompositionEnabled() {
		return nil, newTLCError(ECGeneral, actionCompositionUnsupportedMessage)
	}
	if len(args) < 2 {
		return nil, newTLCError(ECGeneral, "malformed action composition")
	}
	tState := s0.CopyWith(s1)
	intermediate, _, err := t.actionCompositionIntermediateStates(UnknownAction, args[0], EmptyActionItemList, c, s0, tState, cm)
	if err != nil {
		return nil, err
	}
	for i := 0; i < intermediate.Size(); i++ {
		s2, err := t.EnabledImpl(args[1], acts, c, intermediate.At(i), s1, cm)
		if err != nil || s2 != nil {
			return s2, err
		}
	}
	return nil, nil
}

func (t *Tool) enabledFcnApply(pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	if len(pred.Args) < 2 {
		return nil, newTLCError(ECGeneral, "malformed function application in ENABLED expression: %s", SemanticString(pred))
	}
	fval, err := t.Eval(pred.Args[0], c, s0, s1, EvalSetKeepLazy(EvalEnabled), cm)
	if err != nil {
		return nil, err
	}
	if fcn, ok := fval.(*FcnLambdaValue); ok && fcn.FcnRcd == nil {
		c1, err := t.getFcnContext(fcn, pred, c, s0, s1, EvalEnabled, cm)
		if err != nil {
			return nil, err
		}
		return t.EnabledImpl(fcn.Body, acts, c1, s0, s1, cm)
	}
	bval, err := t.applyPredicateFunction("ENABLED", pred, fval, c, s0, s1, EvalEnabled, cm)
	if err != nil {
		return nil, err
	}
	if !bval.Val {
		return nil, nil
	}
	return t.EnabledFromActionList(acts, s0, s1, cm)
}

func (t *Tool) enabledBoundedForall(pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	enum, err := t.Contexts(pred, c, s0, s1, EvalEnabled, cm)
	if err != nil {
		return nil, err
	}
	c1 := enum.NextElement()
	if c1 == nil {
		return t.EnabledFromActionList(acts, s0, s1, cm)
	}
	acts1 := acts
	for c2 := enum.NextElement(); c2 != nil; c2 = enum.NextElement() {
		acts1 = acts1.Cons(pred.Args[0], c2, cm, ActionItemPred)
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	return t.EnabledImpl(pred.Args[0], acts1, c1, s0, s1, cm)
}

func (t *Tool) enabledCase(pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	var other SemanticNode
	for _, arg := range pred.Args {
		pair, ok := arg.(*OpApplNode)
		if !ok || len(pair.Args) < 2 {
			return nil, newTLCError(ECGeneral, "malformed CASE in ENABLED")
		}
		if pair.Args[0] == nil {
			other = pair.Args[1]
			continue
		}
		value, err := t.Eval(pair.Args[0], c, s0, s1, EvalEnabled, cm)
		if err != nil {
			return nil, err
		}
		bval, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCError(ECGeneral, "In computing ENABLED, a non-boolean expression(%s) was used as a guard condition of a CASE.\n%s", value.KindString(), SemanticString(pair.Args[1]))
		}
		if bval.Val {
			return t.EnabledImpl(pair.Args[1], acts, c, s0, s1, cm)
		}
	}
	if other == nil {
		return nil, newTLCError(ECGeneral, "In computing ENABLED, TLC encountered a CASE with no conditions true.\n%s", SemanticString(pred))
	}
	return t.EnabledImpl(other, acts, c, s0, s1, cm)
}

func (t *Tool) enabledEquality(pred SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	varNode := t.GetPrimedVar(left, c, true)
	if varNode == nil {
		bval, err := t.evalBool(pred, c, s0, s1, EvalEnabled, cm, "ENABLED equality")
		if err != nil || !bval.Val {
			return nil, err
		}
		return t.EnabledFromActionList(acts, s0, s1, cm)
	}
	rval, err := t.Eval(right, c, s0, s1, EvalEnabled, cm)
	if err != nil {
		return nil, err
	}
	lval := s1.Lookup(varNode.Name)
	if lval == nil {
		s1 = s1.Bind(varNode.Name, rval)
		return t.EnabledFromActionList(acts, s0, s1, cm)
	}
	eq, err := lval.Equal(rval)
	if err != nil || !eq {
		return nil, err
	}
	return t.EnabledFromActionList(acts, s0, s1, cm)
}

func (t *Tool) enabledMembership(pred SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	varNode := t.GetPrimedVar(left, c, true)
	if varNode == nil {
		bval, err := t.evalBool(pred, c, s0, s1, EvalEnabled, cm, "ENABLED membership")
		if err != nil || !bval.Val {
			return nil, err
		}
		return t.EnabledFromActionList(acts, s0, s1, cm)
	}
	rval, err := t.Eval(right, c, s0, s1, EvalEnabled, cm)
	if err != nil {
		return nil, err
	}
	return t.enabledEnumerateAssignment(varNode.Name, rval, pred, acts, s0, s1, cm)
}

func (t *Tool) enabledSubsetEq(pred SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	varNode := t.GetPrimedVar(left, c, true)
	if varNode == nil {
		bval, err := t.evalBool(pred, c, s0, s1, EvalEnabled, cm, "ENABLED subset")
		if err != nil || !bval.Val {
			return nil, err
		}
		return t.EnabledFromActionList(acts, s0, s1, cm)
	}
	rset, err := t.Eval(right, c, s0, s1, EvalEnabled, cm)
	if err != nil {
		return nil, err
	}
	return t.enabledEnumerateAssignment(varNode.Name, NewSubsetValue(rset, cm), pred, acts, s0, s1, cm)
}

func (t *Tool) enabledEnumerateAssignment(varName *UniqueString, domain Value, pred SemanticNode, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*TLCStateMut, error) {
	lval := s1.Lookup(varName)
	if lval != nil {
		member, err := domain.Member(lval)
		if err != nil || !member {
			return nil, err
		}
		return t.EnabledFromActionList(acts, s0, s1, cm)
	}
	enumerable, ok := asEnumerable(domain)
	if !ok {
		return nil, newTLCError(ECGeneral, "The right side of \\IN is not enumerable.\n%s", SemanticString(pred))
	}
	enum := enumerable.Elements()
	for val := enum.NextElement(); val != nil; val = enum.NextElement() {
		s2State := s1.Bind(varName, val)
		s2, err := t.EnabledFromActionList(acts, s0, s2State, cm)
		if err != nil || s2 != nil {
			return s2, err
		}
	}
	return nil, enum.Err()
}

func (t *Tool) EnabledUnchanged(expr SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (state *TLCStateMut, err error) {
	done := t.callStackEnter(expr)
	defer func() { done(err) }()
	if CoverageEnabled() {
		cm = cm.Get(expr)
	}
	if varNode := t.GetVar(expr, c, true); varNode != nil {
		varName := varNode.Name
		v0, err := t.Eval(expr, c, s0, s1, EvalEnabled, cm)
		if err != nil {
			return nil, err
		}
		v1 := s1.Lookup(varName)
		if v1 == nil {
			s1 = s1.Bind(varName, v0)
			return t.EnabledFromActionList(acts, s0, s1, cm)
		}
		eq, err := v1.Equal(v0)
		if err != nil {
			return nil, err
		}
		if !eq {
			PrintWarning(ECTLCUnchangedVariableChanged, varName.String(), SemanticString(expr))
			return nil, nil
		}
		return t.EnabledFromActionList(acts, s0, s1, cm)
	}
	if appl, ok := expr.(*OpApplNode); ok {
		opName := appl.Operator.Name
		opcode := GetOpCode(opName)
		if opcode == OpcodeTup {
			acts1 := acts
			for i := 1; i < len(appl.Args); i++ {
				acts1 = acts1.Cons(appl.Args[i], c, cm, ActionItemUnchanged)
			}
			if len(appl.Args) == 0 {
				return t.EnabledFromActionList(acts, s0, s1, cm)
			}
			return t.EnabledUnchanged(appl.Args[0], acts1, c, s0, s1, cm)
		}
		if opcode == 0 && len(appl.Args) == 0 {
			val := t.Lookup(appl.Operator, c, EmptyState, false)
			switch v := val.(type) {
			case *LazyValue:
				return t.EnabledUnchanged(v.Expr, acts, v.Con, s0, s1, cm)
			case *LazySupplierValue:
				lv := asLazyValue(v)
				return t.EnabledUnchanged(lv.Expr, acts, lv.Con, s0, s1, cm)
			case *OpDefNode:
				return t.EnabledUnchanged(v.Body, acts, c, s0, s1, cm)
			case nil:
				return nil, newTLCError(ECGeneral, "undefined identifier %s in ENABLED UNCHANGED expression %s", opName, SemanticString(expr))
			default:
				return t.EnabledFromActionList(acts, s0, s1, cm)
			}
		}
	}
	v0, err := t.Eval(expr, c, s0, EmptyState, EvalEnabled, cm)
	if err != nil {
		return nil, err
	}
	v1, err := t.Eval(expr, c, s1, EmptyState, EvalPrimed, cm)
	if err != nil {
		return nil, err
	}
	eq, err := v0.Equal(v1)
	if err != nil || !eq {
		return nil, err
	}
	return t.EnabledFromActionList(acts, s0, s1, cm)
}
