package tlc

import "fmt"

type toolEvalArgs struct {
	con     *Context
	s0      *TLCStateMut
	s1      *TLCStateMut
	control int
	cm      CostModel
}

func parseEvalArgs(args ...any) (*Context, *TLCStateMut, *TLCStateMut, int, CostModel) {
	parsed := toolEvalArgs{con: EmptyContext, s0: EmptyState, s1: EmptyState, control: EvalClear}
	stateArgs := 0
	for _, arg := range args {
		switch v := arg.(type) {
		case nil:
			if stateArgs == 0 {
				parsed.s0 = nil
				stateArgs++
			} else if stateArgs == 1 {
				parsed.s1 = nil
				stateArgs++
			}
		case *Context:
			if v == nil {
				parsed.con = EmptyContext
			} else {
				parsed.con = v
			}
		case *TLCStateMut:
			if stateArgs == 0 {
				parsed.s0 = v
				stateArgs++
			} else if stateArgs == 1 {
				parsed.s1 = v
				stateArgs++
			}
		case int:
			parsed.control = v
		case CostModel:
			parsed.cm = v
		}
	}
	return parsed.con, parsed.s0, parsed.s1, parsed.control, parsed.cm
}

func (t *Tool) Define(sym *SymbolNode, value any) {
	if t == nil || sym == nil {
		return
	}
	if t.Definitions == nil {
		t.Definitions = make(map[*SymbolNode]any)
	}
	t.Definitions[sym] = value
	if sym.Name != nil {
		if t.DefnsByName == nil {
			t.DefnsByName = make(map[*UniqueString]any)
		}
		t.DefnsByName[sym.Name] = value
	}
}

func (t *Tool) DefineName(name string, value any) *SymbolNode {
	sym := NewSymbolNode(name)
	t.Define(sym, value)
	return sym
}

func (t *Tool) Lookup(sym *SymbolNode, con *Context, state *TLCStateMut, primed bool) any {
	return t.LookupWithCutoff(sym, con, false, state, primed)
}

func (t *Tool) LookupWithCutoff(sym *SymbolNode, con *Context, cutoff bool, state *TLCStateMut, primed bool) any {
	if sym == nil {
		return nil
	}
	if con == nil {
		con = EmptyContext
	}
	if val := con.LookupCutoff(sym, cutoff); val != nil {
		return val
	}
	if sym.Data != nil {
		return muxToolObject(sym.Data, state)
	}
	if sym.Name != nil {
		if t != nil {
			if val := t.Definitions[sym]; val != nil {
				return muxToolObject(val, state)
			}
			if sym.IsUserDefinedOp() {
				if val := t.DefnsByName[sym.Name]; val != nil {
					return muxToolObject(val, state)
				}
				return nil
			}
		}
		if state != nil {
			if val := state.Lookup(sym.Name); val != nil {
				return val
			}
		}
		if t != nil {
			if val := t.DefnsByName[sym.Name]; val != nil {
				return muxToolObject(val, state)
			}
		}
		if primed && state != nil {
			return state.Lookup(sym.Name)
		}
	}
	return nil
}

func muxToolObject(value any, state *TLCStateMut) any {
	switch v := value.(type) {
	case nil:
		return nil
	case *WorkerValue:
		workerID, ok := CurrentWorkerID()
		if !ok {
			workerID = workerIDFromState(state)
		}
		return v.ValueForWorker(workerID)
	default:
		return v
	}
}

func (t *Tool) EvalPure(opDef *OpDefNode, args []SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	c1, err := t.GetOpContext(opDef, args, c, true, cm)
	if err != nil {
		return nil, err
	}
	return t.Eval(opDef.Body, c1, s0, s1, control, cm)
}

func (t *Tool) EvalImpl(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if c == nil {
		c = EmptyContext
	}
	switch expr := expr.(type) {
	case nil:
		return ValUndef, newTLCError(ECGeneral, "attempted to evaluate a nil expression")
	case Value:
		return expr, nil
	case *ValueNode:
		return expr.Value, nil
	case *LabelNode:
		return t.Eval(expr.Body, c, s0, s1, control, cm)
	case *OpApplNode:
		return t.EvalAppl(expr, c, s0, s1, control, cm)
	case *LetInNode:
		return t.evalImplLetInKind(expr, c, s0, s1, control, cm)
	case *SubstInNode:
		return t.evalImplSubstInKind(expr, c, s0, s1, control, cm)
	case *APSubstInNode:
		return t.evalImplAPSubstInKind(expr, c, s0, s1, control, cm)
	case *NumeralNode:
		return expr.Value, nil
	case *DecimalNode:
		return nil, newTLCErrorCode(ECTLCCantHandleRealNumbers, SemanticString(expr))
	case *StringNode:
		return expr.Value, nil
	case *AtNode:
		if value, ok := c.LookupFunc(func(sym *SymbolNode) bool { return sym != nil && sym.Name != nil && sym.Name.String() == "@" }).(Value); ok {
			return value, nil
		}
		return ValUndef, newTLCError(ECGeneral, "attempted to evaluate @ outside an EXCEPT expression")
	case *OpArgNode:
		return t.evalImplOpArgKind(expr, c, s0, s1, control, cm)
	case *PossibleTrackNode:
		return t.evalPossibleTrackNode(expr, c, s0, s1, control, cm)
	case *PossibleCheckNode:
		return t.evalPossibleCheckNode(expr)
	default:
		if value := SemanticToolObject(expr); value != nil {
			if v, ok := value.(Value); ok {
				return v, nil
			}
		}
		return ValUndef, newTLCError(ECGeneral, "attempted to evaluate an expression that cannot be evaluated: %s", SemanticString(expr))
	}
}

func (t *Tool) evalImplLetInKind(expr *LetInNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	c1 := letDefinitionsContextWithCostModel(c, expr.Lets, cm, expr.Bindings...)
	return t.Eval(expr.Body, c1, s0, s1, control, cm)
}

func (t *Tool) evalImplSubstInKind(expr *SubstInNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	c1 := c
	for _, sub := range expr.Substs {
		subCM := cm
		if CoverageEnabled() {
			subCM = cm.GetSubst(sub)
		}
		val := t.GetVal(sub.Expr, c, true, subCM)
		c1 = c1.Cons(sub.Op, val)
	}
	return t.Eval(expr.Body, c1, s0, s1, control, cm)
}

func (t *Tool) evalImplAPSubstInKind(expr *APSubstInNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	c1 := c
	for _, sub := range expr.Substs {
		val := t.GetVal(sub.Expr, c, true, cm)
		c1 = c1.Cons(sub.Op, val)
	}
	return t.Eval(expr.Body, c1, s0, s1, control, cm)
}

func (t *Tool) evalImplOpArgKind(expr *OpArgNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	val := t.Lookup(expr.Op, c, s0, false)
	switch v := val.(type) {
	case *OpDefNode:
		return t.setValueSource(expr, NewOpLambdaValue(v, t, c, s0, s1, control, cm)), nil
	case Value:
		return v, nil
	default:
		return ValUndef, newTLCError(ECGeneral, "operator argument %s is undefined", expr.Op)
	}
}

func (t *Tool) EvalAppl(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	done := t.callStackEnter(expr)
	defer func() {
		if failure := recover(); failure != nil {
			switch failure := failure.(type) {
			case *TLCError, *EvalException, *FingerprintException:
				value, err = nil, failure.(error)
			default:
				done(nil)
				panic(failure)
			}
		}
		if err == nil {
			value = t.callStackToolValue(value)
		}
		done(err)
	}()
	return t.EvalApplImpl(expr, c, s0, s1, control, cm)
}

func (t *Tool) EvalApplImpl(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if CoverageEnabled() {
		cm = cm.GetAndIncrement(expr)
	}
	args := expr.Args
	opNode := expr.Operator
	opcode := GetOpCode(opNode.Name)

	if opcode == 0 {
		val := t.Lookup(opNode, c, s0, EvalIsPrimed(control))
		if lv := asLazyValue(val); lv != nil {
			if s1 == nil {
				evaluated, err := t.Eval(lv.Expr, lv.Con, s0, nil, control, lv.CM)
				if err != nil {
					return nil, err
				}
				val = evaluated
			} else {
				evaluated, err, ok := lazyValueGetValue(val, t, s0, s1, control)
				if ok {
					return evaluated, err
				}
			}
		}
		switch v := val.(type) {
		case *OpDefNode:
			opcode = GetOpCode(v.Name)
			if opcode == 0 {
				c1, err := t.GetOpContext(v, args, c, true, cm)
				if err != nil {
					return nil, err
				}
				return t.Eval(v.Body, c1, s0, s1, control, cm)
			}
		case Value:
			if isOperatorValue(v) {
				return EvalOperatorValueWithTool(v, t, args, c, s0, s1, control, cm)
			}
			return v, nil
		case *ThmOrAssumpDefNode:
			c1, err := t.GetThmOrAssumpContext(v, args, c, true)
			if err != nil {
				return nil, err
			}
			return t.Eval(v.Body, c1, s0, s1, control, cm)
		case nil:
			if !EvalIsEnabled(control) && EvalIsPrimed(control) && opNode.IsVariableDecl() {
				return nil, newTLCError(ECGeneral, "state is not completely specified: %s in %s", opNode.Name, SemanticString(expr))
			}
			return nil, NewTLCRuntimeException(ECTLCConfigUndefinedOrNoOperator, opNode.Name.String(), semanticNodeLocationString(expr))
		default:
			return ValUndef, newTLCError(ECGeneral, "cannot evaluate operator %s bound to %T", opNode, val)
		}
		if opcode == 0 {
			return ValUndef, nil
		}
	}

	switch opcode {
	case OpcodeBC:
		return t.evalBoundedChoose(expr, c, s0, s1, control, cm)
	case OpcodeBE:
		return t.evalBoundedExists(expr, c, s0, s1, control, cm)
	case OpcodeBF:
		return t.evalBoundedForall(expr, c, s0, s1, control, cm)
	case OpcodeCase:
		return t.evalCase(expr, c, s0, s1, control, cm)
	case OpcodeCP:
		values, err := t.evalArgs(args, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return t.setValueSource(expr, NewSetOfTuplesValue(values, cm)), nil
	case OpcodeCL:
		for _, arg := range args {
			value, err := t.Eval(arg, c, s0, s1, control, cm)
			if err != nil {
				return nil, err
			}
			bval, err := requireBoolValue(value, "A non-boolean expression (%s) was used as a formula in a conjunction.\n%s", valueKindString(value), SemanticString(arg))
			if err != nil {
				return nil, err
			}
			if !bval.Val {
				return BoolFalse, nil
			}
		}
		return BoolTrue, nil
	case OpcodeDL:
		for _, arg := range args {
			value, err := t.Eval(arg, c, s0, s1, control, cm)
			if err != nil {
				return nil, err
			}
			bval, err := requireBoolValue(value, "A non-boolean expression (%s) was used as a formula in a disjunction.\n%s", valueKindString(value), SemanticString(arg))
			if err != nil {
				return nil, err
			}
			if bval.Val {
				return BoolTrue, nil
			}
		}
		return BoolFalse, nil
	case OpcodeExc:
		return t.evalExcept(expr, c, s0, s1, control, cm)
	case OpcodeFA:
		return t.evalFcnApply(expr, c, s0, s1, control, cm)
	case OpcodeFC, OpcodeNRFS, OpcodeRFS:
		return t.evalFcnConstructor(expr, opcode, c, s0, s1, control, cm)
	case OpcodeITE:
		guardValue, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		guard, err := requireBoolValue(guardValue, "A non-boolean expression (%s) was used as the condition of an IF.\n%s", valueKindString(guardValue), SemanticString(expr))
		if err != nil {
			return nil, err
		}
		if guard.Val {
			return t.Eval(args[1], c, s0, s1, control, cm)
		}
		return t.Eval(args[2], c, s0, s1, control, cm)
	case OpcodeRC:
		return t.evalRecordConstructor(expr, c, s0, s1, control, cm)
	case OpcodeRS:
		return t.evalRecordSelect(expr, c, s0, s1, control, cm)
	case OpcodeSE:
		values, err := t.evalArgs(args, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return t.setValueSource(expr, NewSetEnumValue(values, false, cm)), nil
	case OpcodeSOA:
		return t.evalSetOfAll(expr, c, s0, s1, control, cm)
	case OpcodeSOR:
		return t.evalSetOfRecords(expr, c, s0, s1, control, cm)
	case OpcodeSOF:
		lhs, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		rhs, err := t.Eval(args[1], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return t.setValueSource(expr, NewSetOfFcnsValue(lhs, rhs, cm)), nil
	case OpcodeSSO:
		return t.evalSubsetOf(expr, c, s0, s1, control, cm)
	case OpcodeTup:
		values, err := t.evalArgs(args, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return t.setValueSource(expr, NewTupleValue(values, cm)), nil
	case OpcodeUC:
		return nil, newTLCError(ECGeneral, "TLC attempted to evaluate an unbounded CHOOSE.\nMake sure that the expression is of form CHOOSE x \\in S: P(x).\n%s", SemanticString(expr))
	case OpcodeUE:
		return nil, newTLCError(ECGeneral, "TLC attempted to evaluate an unbounded \\E.\nMake sure that the expression is of form \\E x \\in S: P(x).\n%s", SemanticString(expr))
	case OpcodeUF:
		return nil, newTLCError(ECGeneral, "TLC attempted to evaluate an unbounded \\A.\nMake sure that the expression is of form \\A x \\in S: P(x).\n%s", SemanticString(expr))
	case OpcodeLnot:
		value, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		arg, err := requireBoolValue(value, "Attempted to apply the operator ~ to a non-boolean\n(%s)\n%s", valueKindString(value), SemanticString(expr))
		if err != nil {
			return nil, err
		}
		return NewBoolValue(!arg.Val), nil
	case OpcodeSubset:
		arg, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return t.setValueSource(expr, NewSubsetValue(arg, cm)), nil
	case OpcodeUnion:
		arg, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		value, err := Union(arg)
		if err != nil {
			return nil, err
		}
		return t.setValueSource(expr, value), nil
	case OpcodeDomain:
		arg, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		value, err := domainValue(expr, arg)
		if err != nil {
			return nil, err
		}
		return t.setValueSource(expr, value), nil
	case OpcodeEnabled:
		sfun := NewFunctionalState()
		enabled, err := t.EnabledImpl(args[0], EmptyActionItemList, BranchContext(c), s0, sfun, cm)
		if err != nil {
			return nil, err
		}
		if enabled != nil {
			return BoolTrue, nil
		}
		return BoolFalse, nil
	case OpcodeEq:
		return t.evalEq(args[0], args[1], c, s0, s1, control, cm, false)
	case OpcodeLand:
		return t.evalBinaryConjunction(args, c, s0, s1, control, cm, expr)
	case OpcodeLor:
		return t.evalBinaryDisjunction(args, c, s0, s1, control, cm, expr)
	case OpcodeImplies:
		value1, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		arg1, err := requireBoolValue(value1, "Attempted to evaluate an expression of form P => Q when P was\n%s.\n%s", valueKindString(value1), SemanticString(expr))
		if err != nil {
			return nil, err
		}
		if !arg1.Val {
			return BoolTrue, nil
		}
		value2, err := t.Eval(args[1], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return requireBoolValue(value2, "Attempted to evaluate an expression of form P => Q when Q was\n%s.\n%s", valueKindString(value2), SemanticString(expr))
	case OpcodeEquiv:
		value1, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		value2, err := t.Eval(args[1], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		arg1, ok1 := value1.(*BoolValue)
		arg2, ok2 := value2.(*BoolValue)
		if !ok1 || !ok2 {
			return nil, newTLCError(ECGeneral, "Attempted to evaluate an expression of form P <=> Q when P or Q was not a boolean.\n%s", SemanticString(expr))
		}
		return NewBoolValue(arg1.Val == arg2.Val), nil
	case OpcodeNoteq:
		return t.evalEq(args[0], args[1], c, s0, s1, control, cm, true)
	case OpcodeSubseteq:
		arg1, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		arg2, err := t.Eval(args[1], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return subsetEq(expr, arg1, arg2)
	case OpcodeIn, OpcodeNotin:
		arg1, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		arg2, err := t.Eval(args[1], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		member, err := arg2.Member(arg1)
		if err != nil {
			return nil, err
		}
		if opcode == OpcodeNotin {
			member = !member
		}
		return NewBoolValue(member), nil
	case OpcodeSetdiff, OpcodeCap, OpcodeCup:
		value, err := t.evalSetOp(opcode, args[0], args[1], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return t.setValueSource(expr, value), nil
	case OpcodeNop:
		return t.Eval(args[0], c, s0, s1, control, cm)
	case OpcodePrime:
		return t.Eval(args[0], c, s1, nil, EvalSetPrimedIfEnabled(control), cm)
	case OpcodeUnchanged:
		v0, err := t.Eval(args[0], c, s0, EmptyState, control, cm)
		if err != nil {
			return nil, err
		}
		v1, err := t.Eval(args[0], c, s1, nil, EvalSetPrimedIfEnabled(control), cm)
		if err != nil {
			return nil, err
		}
		eq, err := v0.Equal(v1)
		if err != nil {
			return nil, err
		}
		return NewBoolValue(eq), nil
	case OpcodeAA, OpcodeSA:
		return t.evalActionSubscript(expr, opcode, args, c, s0, s1, control, cm)
	case OpcodeCdot:
		return t.evalActionComposition(args, c, s0, s1, control, cm)
	case OpcodeSF:
		return nil, newTLCErrorCode(ECTLCEncounteredFormulaInPredicate, "SF", SemanticString(expr))
	case OpcodeWF:
		return nil, newTLCErrorCode(ECTLCEncounteredFormulaInPredicate, "WF", SemanticString(expr))
	case OpcodeTE:
		return nil, newTLCErrorCode(ECTLCEncounteredFormulaInPredicate, "\\EE", SemanticString(expr))
	case OpcodeTF:
		return nil, newTLCErrorCode(ECTLCEncounteredFormulaInPredicate, "\\AA", SemanticString(expr))
	case OpcodeLeadsto:
		return nil, newTLCErrorCode(ECTLCEncounteredFormulaInPredicate, "a ~> b", SemanticString(expr))
	case OpcodeArrow:
		return nil, newTLCErrorCode(ECTLCEncounteredFormulaInPredicate, "a -+-> formula", SemanticString(expr))
	case OpcodeBox:
		return nil, newTLCErrorCode(ECTLCEncounteredFormulaInPredicate, "[]A", SemanticString(expr))
	case OpcodeDiamond:
		return nil, newTLCErrorCode(ECTLCEncounteredFormulaInPredicate, "<>A", SemanticString(expr))
	default:
		return nil, newTLCError(ECGeneral, "TLC BUG: could not evaluate this expression.\n%s", SemanticString(expr))
	}
}

func (t *Tool) GetVal(expr SemanticNode, c *Context, cachable bool, cm CostModel) any {
	if opArg, ok := expr.(*OpArgNode); ok {
		return t.Lookup(opArg.Op, c, nil, false)
	}
	return NewLazyValue(expr, c, cachable, cm)
}

func (t *Tool) GetOpContext(opDef *OpDefNode, args []SemanticNode, c *Context, cachable bool, cm CostModel) (*Context, error) {
	if opDef == nil {
		return c, newTLCError(ECGeneral, "attempted to apply nil operator definition")
	}
	if len(opDef.Params) != len(args) {
		return c, newTLCError(ECGeneral, "applying operator %s with wrong number of arguments", opDef)
	}
	c1 := c
	if c1 == nil {
		c1 = EmptyContext
	}
	for i, param := range opDef.Params {
		c1 = c1.Cons(param, t.GetVal(args[i], c, cachable, cm))
	}
	return c1, nil
}

func (t *Tool) GetThmOrAssumpContext(opDef *ThmOrAssumpDefNode, args []SemanticNode, c *Context, cachable bool) (*Context, error) {
	if opDef == nil {
		return c, newTLCError(ECGeneral, "attempted to apply nil theorem or assumption definition")
	}
	if len(opDef.Params) != len(args) {
		return c, newTLCError(ECGeneral, "applying theorem or assumption %s with wrong number of arguments", opDef)
	}
	c1 := c
	if c1 == nil {
		c1 = EmptyContext
	}
	for i, param := range opDef.Params {
		c1 = c1.Cons(param, t.GetVal(args[i], c, cachable, DoNotRecordCostModel))
	}
	return c1, nil
}

func (t *Tool) GetLevelBound(expr SemanticNode, c *Context) int {
	if c == nil {
		c = EmptyContext
	}
	switch expr := expr.(type) {
	case *OpApplNode:
		return t.GetLevelBoundAppl(expr, c)
	case *LetInNode:
		level := TLCLevelConstant
		c1 := c
		for _, opDef := range expr.Lets {
			if opDef == nil || opDef.Name == nil {
				continue
			}
			if bodyLevel := t.GetLevelBound(opDef.Body, c1); bodyLevel > level {
				level = bodyLevel
			}
			sym := opDef.Symbol
			if sym == nil {
				sym = &SymbolNode{Name: opDef.Name}
			}
			c1 = c1.Cons(sym, IntOne)
		}
		for _, binding := range expr.Bindings {
			if binding.Symbol == nil {
				continue
			}
			c1 = c1.Cons(binding.Symbol, binding.Value)
		}
		if bodyLevel := t.GetLevelBound(expr.Body, c1); bodyLevel > level {
			level = bodyLevel
		}
		return level
	case *SubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, true, DoNotRecordCostModel))
		}
		return t.GetLevelBound(expr.Body, c1)
	case *APSubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, true, DoNotRecordCostModel))
		}
		return t.GetLevelBound(expr.Body, c1)
	case *LabelNode:
		return t.GetLevelBound(expr.Body, c)
	default:
		return TLCLevelConstant
	}
}

func letDefinitionsContext(c *Context, lets []*OpDefNode, bindings ...LetBinding) *Context {
	return letDefinitionsContextWithCostModel(c, lets, DoNotRecordCostModel, bindings...)
}

func letDefinitionsContextWithCostModel(c *Context, lets []*OpDefNode, cm CostModel, bindings ...LetBinding) *Context {
	if c == nil {
		c = EmptyContext
	}
	c1 := c
	for _, opDef := range lets {
		if opDef == nil || opDef.Name == nil {
			continue
		}
		sym := opDef.Symbol
		if sym == nil {
			sym = &SymbolNode{Name: opDef.Name}
		}
		if opDef.Arity() == 0 {
			c1 = c1.Cons(sym, NewLazyValue(opDef.Body, c1, true, cm))
			continue
		}
		c1 = c1.Cons(sym, opDef)
	}
	for _, binding := range bindings {
		if binding.Symbol == nil {
			continue
		}
		c1 = c1.Cons(binding.Symbol, binding.Value)
	}
	return c1
}

func (t *Tool) GetLevelBoundAppl(expr *OpApplNode, c *Context) int {
	if expr == nil || expr.Operator == nil || expr.Operator.Name == nil {
		return TLCLevelConstant
	}
	opNode := expr.Operator
	opName := opNode.Name
	opcode := GetOpCode(opName)

	if IsTemporalOpcode(opcode) {
		return TLCLevelTemporal
	}
	if IsActionOpcode(opcode) {
		return TLCLevelAction
	}
	if opcode == OpcodeEnabled {
		return TLCLevelState
	}

	level := TLCLevelConstant
	for _, bound := range expr.BdedQuantBounds {
		if boundLevel := t.GetLevelBound(bound, c); boundLevel > level {
			level = boundLevel
		}
	}

	if opcode == OpcodeRFS && len(expr.UnbdedQuantSymbols) > 0 {
		c = c.Cons(expr.UnbdedQuantSymbols[0], IntOne)
	}

	for _, arg := range expr.Args {
		if arg == nil {
			continue
		}
		if argLevel := t.GetLevelBound(arg, c); argLevel > level {
			level = argLevel
		}
	}

	if opcode != 0 {
		return level
	}
	if opName.VarLoc() >= 0 {
		return TLCLevelState
	}
	val := t.LookupWithCutoff(opNode, c, false, EmptyState, false)
	switch v := val.(type) {
	case *OpDefNode:
		c1 := c.Cons(opNode, IntOne)
		if bodyLevel := t.GetLevelBound(v.Body, c1); bodyLevel > level {
			level = bodyLevel
		}
	case *LazyValue:
		if lazyLevel := t.GetLevelBound(v.Expr, v.Con); lazyLevel > level {
			level = lazyLevel
		}
	case *LazySupplierValue:
		lv := asLazyValue(v)
		if lazyLevel := t.GetLevelBound(lv.Expr, lv.Con); lazyLevel > level {
			level = lazyLevel
		}
	case *EvaluatingValue:
		if v.MinLevel > level {
			level = v.MinLevel
		}
	case *PriorityEvaluatingValue:
		if v.MinLevel > level {
			level = v.MinLevel
		}
	case *CallableValue:
		if v.MinLevel > level {
			level = v.MinLevel
		}
	case *MethodValue:
		if v.MinLevel > level {
			level = v.MinLevel
		}
	}
	return level
}

func (t *Tool) Contexts(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (*ContextEnumerator, error) {
	return t.contexts(expr, c, s0, s1, control, cm, false)
}

func (t *Tool) ContextsRandomized(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (*ContextEnumerator, error) {
	return t.contexts(expr, c, s0, s1, control, cm, true)
}

func (t *Tool) contexts(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel, randomized bool) (*ContextEnumerator, error) {
	alen := 0
	for i := range expr.BdedQuantBounds {
		if i < len(expr.BdedQuantATuple) && expr.BdedQuantATuple[i] {
			alen++
			continue
		}
		if i < len(expr.BdedQuantSymbolLists) {
			alen += len(expr.BdedQuantSymbolLists[i])
		}
	}
	vars := make([]any, 0, alen)
	enums := make([]ValueEnumeration, 0, alen)
	for i, bound := range expr.BdedQuantBounds {
		val, err := t.Eval(bound, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		enumerable, ok := asEnumerable(val)
		if !ok {
			return nil, newTLCError(ECGeneral, "%s", nonEnumerableErrorMsg(val, bound))
		}
		if i < len(expr.BdedQuantATuple) && expr.BdedQuantATuple[i] {
			if i >= len(expr.BdedQuantSymbolLists) || len(expr.BdedQuantSymbolLists[i]) == 0 {
				return nil, newTLCError(ECGeneral, "bounded tuple quantifier has no bound variables")
			}
			enum, err := t.boundValueEnumeration(enumerable, randomized)
			if err != nil {
				return nil, err
			}
			vars = append(vars, expr.BdedQuantSymbolLists[i])
			enums = append(enums, enum)
		} else if i < len(expr.BdedQuantSymbolLists) && len(expr.BdedQuantSymbolLists[i]) > 0 {
			for _, symbol := range expr.BdedQuantSymbolLists[i] {
				enum, err := t.boundValueEnumeration(enumerable, randomized)
				if err != nil {
					return nil, err
				}
				vars = append(vars, symbol)
				enums = append(enums, enum)
			}
		} else {
			return nil, newTLCError(ECGeneral, "bounded quantifier has no bound variables")
		}
	}
	return NewContextEnumerator(vars, enums, c), nil
}

func (t *Tool) boundValueEnumeration(enumerable Enumerable, randomized bool) (ValueEnumeration, error) {
	if randomized {
		return randomizedValueEnumeration(enumerable)
	}
	return enumerable.Elements(), nil
}

func (t *Tool) evalArgs(args []SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) ([]Value, error) {
	values := make([]Value, len(args))
	for i, arg := range args {
		value, err := t.Eval(arg, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

func (t *Tool) evalBool(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel, where string) (*BoolValue, error) {
	value, err := t.Eval(expr, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	bval, ok := value.(*BoolValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "expected boolean while evaluating %s, got %s", where, value.KindString())
	}
	return bval, nil
}

func requireBoolValue(value Value, format string, args ...any) (*BoolValue, error) {
	if bval, ok := value.(*BoolValue); ok {
		return bval, nil
	}
	return nil, newTLCError(ECGeneral, format, args...)
}

func valueKindString(value Value) string {
	if value == nil {
		return "<nil>"
	}
	return value.KindString()
}

func (t *Tool) evalBinaryConjunction(args []SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel, expr SemanticNode) (Value, error) {
	for i, arg := range args {
		value, err := t.Eval(arg, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		side := "Q"
		if i == 0 {
			side = "P"
		}
		bval, err := requireBoolValue(value, "Attempted to evaluate an expression of form P /\\ Q when %s was\n%s.\n%s", side, valueKindString(value), SemanticString(expr))
		if err != nil {
			return nil, err
		}
		if !bval.Val {
			return BoolFalse, nil
		}
	}
	return BoolTrue, nil
}

func (t *Tool) evalBinaryDisjunction(args []SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel, expr SemanticNode) (Value, error) {
	for i, arg := range args {
		value, err := t.Eval(arg, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		side := "Q"
		if i == 0 {
			side = "P"
		}
		bval, err := requireBoolValue(value, "Attempted to evaluate an expression of form P \\/ Q when %s was\n%s.\n%s", side, valueKindString(value), SemanticString(expr))
		if err != nil {
			return nil, err
		}
		if bval.Val {
			return BoolTrue, nil
		}
	}
	return BoolFalse, nil
}

func (t *Tool) evalBoundedChoose(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if len(expr.Args) == 0 || len(expr.BdedQuantBounds) == 0 {
		return nil, newTLCError(ECGeneral, "malformed bounded CHOOSE")
	}
	inVal, err := t.Eval(expr.BdedQuantBounds[0], c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	enumerable, ok := asEnumerable(inVal)
	if !ok {
		return nil, newTLCError(ECGeneral, "Attempted to compute the value of an expression of\nform CHOOSE x \\in S: P, but S was not enumerable.\n%s", SemanticString(expr))
	}
	inVal = inVal.Normalize()
	pred := expr.Args[0]
	bvars := expr.BdedQuantSymbolLists[0]
	isTuple := len(expr.BdedQuantATuple) > 0 && expr.BdedQuantATuple[0]
	enum, err := normalizedChooseEnumeration(inVal, enumerable)
	if err != nil {
		return nil, err
	}
	for val := enum.NextElement(); val != nil; val = enum.NextElement() {
		c1 := c
		if isTuple {
			tuple := asTupleValue(val)
			if tuple == nil || len(tuple.Elems) != len(bvars) {
				return nil, newTLCError(ECGeneral, "Attempted to compute the value of an expression of form\nCHOOSE <<x1, ... , xN>> \\in S: P, but S was not a set\nof N-tuples.\n%s", SemanticString(expr))
			}
			for i, variable := range bvars {
				c1 = c1.Cons(variable, tuple.Elems[i])
			}
		} else {
			c1 = c1.Cons(bvars[0], val)
		}
		value, err := t.Eval(pred, c1, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		bval, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCExpectedValue, "boolean", SemanticString(expr))
		}
		if bval.Val {
			return val, nil
		}
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	return nil, newTLCError(ECGeneral, "Attempted to compute the value of an expression of form\nCHOOSE x \\in S: P, but no element of S satisfied P.\n%s", SemanticString(expr))
}

func normalizedChooseEnumeration(value Value, enumerable Enumerable) (ValueEnumeration, error) {
	switch v := value.(type) {
	case *SetEnumValue:
		if _, err := v.normalizeSet(); err != nil {
			return nil, err
		}
		return v.Elements(), nil
	case *SubsetValue:
		v.Normalize()
		return v.Elements(), nil
	case *KSubsetValue:
		v.Normalize()
		return v.Elements(), nil
	case *IntervalValue:
		return v.Elements(), nil
	}
	// Java EnumerableValue.elements(Ordering.NORMALIZED) falls back to
	// enumerating into a SetEnumValue and normalizing that enumerated set.
	set, err := toSetEnumValue(value)
	if err != nil {
		set, err = setEnumFromEnumeration(enumerable.Elements(), false)
		if err != nil {
			return nil, err
		}
	}
	if _, err := set.normalizeSet(); err != nil {
		return nil, err
	}
	return set.Elements(), nil
}

func (t *Tool) evalBoundedExists(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	enum, err := t.Contexts(expr, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
		value, err := t.Eval(expr.Args[0], c1, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		bval, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCExpectedValue, "boolean", SemanticString(expr))
		}
		if bval.Val {
			return BoolTrue, nil
		}
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	return BoolFalse, nil
}

func (t *Tool) evalBoundedForall(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	enum, err := t.Contexts(expr, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
		value, err := t.Eval(expr.Args[0], c1, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		bval, ok := value.(*BoolValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCExpectedValue, "boolean", SemanticString(expr))
		}
		if !bval.Val {
			return BoolFalse, nil
		}
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func (t *Tool) evalCase(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	var other SemanticNode
	for _, arg := range expr.Args {
		pair, ok := arg.(*OpApplNode)
		if !ok || len(pair.Args) < 2 {
			return nil, newTLCError(ECGeneral, "malformed CASE arm")
		}
		if pair.Args[0] == nil {
			other = pair.Args[1]
			if CoverageEnabled() {
				cm = cm.Get(pair)
			}
			continue
		}
		armCM := cm
		if CoverageEnabled() {
			armCM = cm.Get(pair)
		}
		value, err := t.Eval(pair.Args[0], c, s0, s1, control, armCM)
		if err != nil {
			return nil, err
		}
		guard, err := requireBoolValue(value, "A non-boolean expression (%s) was used as a condition of a CASE. %s", valueKindString(value), SemanticString(pair.Args[0]))
		if err != nil {
			return nil, err
		}
		if guard.Val {
			return t.Eval(pair.Args[1], c, s0, s1, control, armCM)
		}
	}
	if other == nil {
		return nil, newTLCError(ECGeneral, "Attempted to evaluate a CASE with no conditions true.\n%s", SemanticString(expr))
	}
	return t.Eval(other, c, s0, s1, control, cm)
}

func (t *Tool) evalExcept(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if len(expr.Args) == 0 {
		return nil, newTLCError(ECGeneral, "malformed EXCEPT")
	}
	result, err := t.Eval(expr.Args[0], c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	for _, arg := range expr.Args[1:] {
		pair, ok := arg.(*OpApplNode)
		if !ok || len(pair.Args) < 2 {
			return nil, newTLCError(ECGeneral, "malformed EXCEPT pair")
		}
		pathNode, ok := pair.Args[0].(*OpApplNode)
		if !ok {
			return nil, newTLCError(ECGeneral, "malformed EXCEPT path")
		}
		pairCM := cm
		pathCM := cm
		if CoverageEnabled() {
			pairCM = cm.Get(pair)
			pathCM = pairCM.Get(pathNode)
		}
		lhs, err := t.evalArgs(pathNode.Args, c, s0, s1, control, pathCM)
		if err != nil {
			return nil, err
		}
		atVal, err := selectPath(result, lhs)
		if err != nil {
			return nil, err
		}
		if atVal == nil {
			PrintWarning(ECTLCExceptAppliedToUnknownField, SemanticString(expr.Args[0]))
			continue
		}
		atSym := NewSymbolNode("@")
		rhs, err := t.Eval(pair.Args[1], c.Cons(atSym, atVal), s0, s1, control, pairCM)
		if err != nil {
			return nil, err
		}
		result, err = result.TakeExcept(ValueExcept{Path: lhs, Value: rhs})
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (t *Tool) evalFcnApply(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	args := expr.Args
	fval, err := t.Eval(args[0], c, s0, s1, EvalSetKeepLazy(control), cm)
	if err != nil {
		return nil, err
	}
	return t.applyEvaluatedFunction(expr, fval, c, s0, s1, control, cm, true)
}

func (t *Tool) applyEvaluatedFunction(expr *OpApplNode, fval Value, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel, rejectTupleRecordMultiArg bool) (Value, error) {
	args := expr.Args
	switch f := fval.(type) {
	case *FcnRcdValue:
		argVal, err := t.evalFunctionApplicationArgument(expr, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return f.Apply(argVal)
	case *FcnLambdaValue:
		argVal, err := t.evalFunctionApplicationArgument(expr, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return f.ApplyWithControl(argVal, control)
	case *TupleValue:
		if rejectTupleRecordMultiArg && len(args) != 2 {
			return nil, newTLCError(ECGeneral, "Attempted to evaluate an expression of form f[e1, ... , eN]\nwith f a tuple or record and N > 1.\n%s", SemanticString(expr))
		}
		argVal, err := t.evalFunctionApplicationArgument(expr, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return f.Apply(argVal)
	case *RecordValue:
		if rejectTupleRecordMultiArg && len(args) != 2 {
			return nil, newTLCError(ECGeneral, "Attempted to evaluate an expression of form f[e1, ... , eN]\nwith f a tuple or record and N > 1.\n%s", SemanticString(expr))
		}
		argVal, err := t.evalFunctionApplicationArgument(expr, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return f.Apply(argVal)
	case *CounterExample:
		record := asRecordValue(f)
		if record == nil {
			return nil, newTLCError(ECGeneral, "A non-function (%s) was applied as a function.\n%s", valueKindString(fval), SemanticString(expr))
		}
		if rejectTupleRecordMultiArg && len(args) != 2 {
			return nil, newTLCError(ECGeneral, "Attempted to evaluate an expression of form f[e1, ... , eN]\nwith f a tuple or record and N > 1.\n%s", SemanticString(expr))
		}
		argVal, err := t.evalFunctionApplicationArgument(expr, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return record.Apply(argVal)
	default:
		return nil, newTLCError(ECGeneral, "A non-function (%s) was applied as a function.\n%s", valueKindString(fval), SemanticString(expr))
	}
}

func (t *Tool) evalFunctionApplicationArgument(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if len(expr.Args) < 2 {
		return nil, newTLCError(ECGeneral, "malformed function application: %s", SemanticString(expr))
	}
	return t.Eval(expr.Args[1], c, s0, s1, control, cm)
}

func (t *Tool) getFcnContext(fcn *FcnLambdaValue, expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (*Context, error) {
	fcon := fcn.Con
	plen := fcn.Params.Length()
	formals := fcn.Params.Formals
	domains := fcn.Params.Domains
	isTuples := fcn.Params.IsTuples
	argVal, err := t.evalFunctionApplicationArgument(expr, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	if plen == 1 {
		member, err := domains[0].Member(argVal)
		if err != nil {
			return nil, err
		}
		if !member {
			return nil, newTLCError(ECGeneral, "In applying the function\n%s,\nthe first argument is:\n%swhich is not in its domain.\n%s", ValuesPPR(fcn), ValuesPPR(argVal), SemanticString(expr.Args[0]))
		}
		if isTuples[0] {
			ids := formals[0]
			tuple := asTupleValue(argVal)
			matches := tuple != nil
			if matches {
				size, err := argVal.Size()
				if err != nil {
					return nil, err
				}
				matches = size == len(ids)
			}
			if !matches {
				// Java intentionally prints this.toString(), including subclass overrides.
				identity := fmt.Sprintf("%T@%p", t, t)
				if t.CallStack != nil {
					identity = t.CallStack.String()
				}
				return nil, newTLCError(ECGeneral, "In applying the function\n%s,\nthe argument is:\n%swhich does not match its formal parameter.\n%s", ValuesPPRString(identity), ValuesPPR(argVal), SemanticString(expr.Args[0]))
			}
			for i, id := range ids {
				fcon = fcon.Cons(id, tuple.Elems[i])
			}
		} else {
			fcon = fcon.Cons(formals[0][0], argVal)
		}
		return fcon, nil
	}
	tuple := asTupleValue(argVal)
	if tuple == nil {
		return nil, newTLCError(ECGeneral, "Attempted to apply a function to an argument not in its domain.\n%s", SemanticString(expr.Args[0]))
	}
	argn := 0
	elems := tuple.Elems
	for i, ids := range formals {
		domain := domains[i]
		if isTuples[i] {
			member, err := domain.Member(elems[argn])
			if err != nil {
				return nil, err
			}
			if !member {
				return nil, newTLCError(ECGeneral, "In applying the function\n%s,\nthe argument number %d is:\n%s\nwhich is not in its domain.\n%s", ValuesPPR(fcn), argn+1, ValuesPPR(elems[argn]), SemanticString(expr.Args[0]))
			}
			inner := asTupleValue(elems[argn])
			argn++
			if inner == nil || len(inner.Elems) != len(ids) {
				return nil, newTLCError(ECGeneral, "In applying the function\n%s,\nthe argument number %d is:\n%swhich does not match its formal parameter.\n%s", ValuesPPR(fcn), argn, ValuesPPR(elems[argn-1]), SemanticString(expr.Args[0]))
			}
			for j, id := range ids {
				fcon = fcon.Cons(id, inner.Elems[j])
			}
		} else {
			for _, id := range ids {
				member, err := domain.Member(elems[argn])
				if err != nil {
					return nil, err
				}
				if !member {
					return nil, newTLCError(ECGeneral, "In applying the function\n%s,\nthe argument number %d is:\n%s\nwhich is not in its domain.\n%s", ValuesPPR(fcn), argn+1, ValuesPPR(elems[argn]), SemanticString(expr.Args[0]))
				}
				fcon = fcon.Cons(id, elems[argn])
				argn++
			}
		}
	}
	return fcon, nil
}

func (t *Tool) evalFcnConstructor(expr *OpApplNode, opcode int, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	dvals := make([]Value, len(expr.BdedQuantBounds))
	isFcnRcd := true
	for i, domain := range expr.BdedQuantBounds {
		value, err := t.Eval(domain, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		dvals[i] = value
		isFcnRcd = isFcnRcd && isReducibleValue(value)
	}
	params := NewFcnParams(expr.BdedQuantSymbolLists, expr.BdedQuantATuple, dvals)
	fval := NewFcnLambdaValue(params, expr.Args[0], t, c, s0, s1, control, cm)
	t.setValueSource(expr, fval)
	if opcode == OpcodeRFS && len(expr.UnbdedQuantSymbols) > 0 {
		fval.MakeRecursive(expr.UnbdedQuantSymbols[0])
		isFcnRcd = false
	}
	if isFcnRcd && !EvalIsKeepLazy(control) {
		return fval.materializeFcnRcd()
	}
	return fval, nil
}

func isReducibleValue(value Value) bool {
	switch value.(type) {
	case *IntervalValue, *SetEnumValue:
		return true
	default:
		return false
	}
}

func (t *Tool) evalRecordConstructor(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	names := make([]*UniqueString, len(expr.Args))
	values := make([]Value, len(expr.Args))
	for i, arg := range expr.Args {
		pair, ok := arg.(*OpApplNode)
		if !ok || len(pair.Args) < 2 {
			return nil, newTLCError(ECGeneral, "malformed record constructor")
		}
		name, err := stringUniqueFromNode(pair.Args[0])
		if err != nil {
			return nil, err
		}
		pairCM := cm
		if CoverageEnabled() {
			pairCM = cm.Get(pair)
		}
		value, err := t.Eval(pair.Args[1], c, s0, s1, control, pairCM)
		if err != nil {
			return nil, err
		}
		names[i] = name
		values[i] = value
	}
	return t.setValueSource(expr, NewRecordValue(names, values, false, cm)), nil
}

func (t *Tool) evalRecordSelect(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	rval, err := t.Eval(expr.Args[0], c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	sval, ok := SemanticToolObject(expr.Args[1]).(Value)
	if !ok {
		sval, err = t.Eval(expr.Args[1], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
	}
	if record, ok := rval.(*RecordValue); ok {
		result, err := record.Select(sval)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, newTLCError(ECGeneral, "attempted to select nonexistent field %s from record %s", sval, record)
		}
		return result, nil
	}
	if fcn, ok := rval.(*FcnRcdValue); ok {
		return fcn.Apply(sval)
	}
	if fcn := asFcnRcdValue(rval); fcn != nil {
		return fcn.Apply(sval)
	}
	return nil, newTLCError(ECGeneral, "attempted to select field %s from non-record value %s", sval, rval)
}

func (t *Tool) evalSetOfAll(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	enum, err := t.Contexts(expr, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	values := NewValueVec(0)
	for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
		value, err := t.Eval(expr.Args[0], c1, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		values.Add(value)
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	return t.setValueSource(expr, NewSetEnumValueVec(values, false, cm)), nil
}

func (t *Tool) evalSetOfRecords(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	names := make([]*UniqueString, len(expr.Args))
	values := make([]Value, len(expr.Args))
	for i, arg := range expr.Args {
		pair, ok := arg.(*OpApplNode)
		if !ok || len(pair.Args) < 2 {
			return nil, newTLCError(ECGeneral, "malformed set-of-records field")
		}
		name, err := stringUniqueFromNode(pair.Args[0])
		if err != nil {
			return nil, err
		}
		pairCM := cm
		if CoverageEnabled() {
			pairCM = cm.Get(pair)
		}
		value, err := t.Eval(pair.Args[1], c, s0, s1, control, pairCM)
		if err != nil {
			return nil, err
		}
		names[i] = name
		values[i] = value
	}
	value, err := NewSetOfRcdsValue(names, values, false, cm)
	if err != nil {
		return nil, err
	}
	return t.setValueSource(expr, value), nil
}

func (t *Tool) evalSubsetOf(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	inVal, err := t.Eval(expr.BdedQuantBounds[0], c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	bvars := expr.BdedQuantSymbolLists[0]
	isTuple := len(expr.BdedQuantATuple) > 0 && expr.BdedQuantATuple[0]
	if isReducibleValue(inVal) {
		enum, _ := asEnumerable(inVal)
		values := NewValueVec(0)
		e := enum.Elements()
		for elem := e.NextElement(); elem != nil; elem = e.NextElement() {
			c1, err := bindBoundedValue(c, bvars, isTuple, elem)
			if err != nil {
				return nil, err
			}
			value, err := t.Eval(expr.Args[0], c1, s0, s1, control, cm)
			if err != nil {
				return nil, err
			}
			bval, err := requireBoolValue(value, "Attempted to evaluate an expression of form {x \\in S : P(x)} when P was %s.\n%s", valueKindString(value), SemanticString(expr.Args[0]))
			if err != nil {
				return nil, err
			}
			if bval.Val {
				values.Add(elem)
			}
		}
		if err := e.Err(); err != nil {
			return nil, err
		}
		return t.setValueSource(expr, NewSetEnumValueVec(values, inVal.IsNormalized(), cm)), nil
	}
	if isTuple {
		return t.setValueSource(expr, NewSetPredValue(bvars, inVal, expr.Args[0], t, c, s0, s1, control, cm)), nil
	}
	return t.setValueSource(expr, NewSetPredValue(bvars[0], inVal, expr.Args[0], t, c, s0, s1, control, cm)), nil
}

func (t *Tool) evalEq(left SemanticNode, right SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel, negate bool) (Value, error) {
	arg1, err := t.Eval(left, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	arg2, err := t.Eval(right, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	eq, err := arg1.Equal(arg2)
	if err != nil {
		return nil, err
	}
	if negate {
		eq = !eq
	}
	return NewBoolValue(eq), nil
}

func (t *Tool) evalSetOp(opcode int, left SemanticNode, right SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	arg1, err := t.Eval(left, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	arg2, err := t.Eval(right, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	switch opcode {
	case OpcodeSetdiff:
		switch left := arg1.(type) {
		case *IntervalValue:
			return left.Diff(arg2)
		case *SetEnumValue:
			return left.Diff(arg2)
		}
		return NewSetDiffValue(arg1, arg2), nil
	case OpcodeCap:
		switch left := arg1.(type) {
		case *IntervalValue:
			return left.Cap(arg2)
		case *SetEnumValue:
			return left.Cap(arg2)
		}
		switch right := arg2.(type) {
		case *IntervalValue:
			return right.Cap(arg1)
		case *SetEnumValue:
			return right.Cap(arg1)
		}
		return NewSetCapValue(arg1, arg2), nil
	case OpcodeCup:
		switch left := arg1.(type) {
		case *IntervalValue:
			return left.Cup(arg2)
		case *SetEnumValue:
			return left.Cup(arg2)
		}
		switch right := arg2.(type) {
		case *IntervalValue:
			return right.Cup(arg1)
		case *SetEnumValue:
			return right.Cup(arg1)
		}
		return NewSetCupValue(arg1, arg2, cm), nil
	default:
		return nil, newTLCError(ECGeneral, "unknown set operator opcode %d", opcode)
	}
}

func (t *Tool) evalActionSubscript(expr SemanticNode, opcode int, args []SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	value, err := t.Eval(args[0], c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	form := "<A>_e"
	if opcode == OpcodeSA {
		form = "[A]_e"
	}
	res, err := requireBoolValue(value, "Attempted to evaluate an expression of form %s, but A was not a boolean.\n%s", form, SemanticString(expr))
	if err != nil {
		return nil, err
	}
	if opcode == OpcodeAA && !res.Val {
		return BoolFalse, nil
	}
	if opcode == OpcodeSA && res.Val {
		return BoolTrue, nil
	}
	v0, err := t.Eval(args[1], c, s0, EmptyState, control, cm)
	if err != nil {
		return nil, err
	}
	v1, err := t.Eval(args[1], c, s1, nil, EvalSetPrimedIfEnabled(control), cm)
	if err != nil {
		return nil, err
	}
	eq, err := v0.Equal(v1)
	if err != nil {
		return nil, err
	}
	if opcode == OpcodeAA {
		return NewBoolValue(res.Val && !eq), nil
	}
	return NewBoolValue(res.Val || eq), nil
}

func (t *Tool) evalActionComposition(args []SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if !actionCompositionEnabled() {
		return nil, newTLCError(ECGeneral, actionCompositionUnsupportedMessage)
	}
	if len(args) < 2 {
		return nil, newTLCError(ECGeneral, "malformed action composition")
	}
	intermediate, _, err := t.actionCompositionIntermediateStates(UnknownAction, args[0], EmptyActionItemList, c, s0, NewEmptyState(), cm)
	if err != nil {
		return nil, err
	}
	for i := 0; i < intermediate.Size(); i++ {
		res, err := t.evalBool(args[1], c, intermediate.At(i), s1, control, cm, "action composition")
		if err != nil {
			return nil, err
		}
		if res.Val {
			return BoolTrue, nil
		}
	}
	return BoolFalse, nil
}

func isOperatorValue(value Value) bool {
	switch value.(type) {
	case *OpLambdaValue, *OpRcdValue, *MethodValue, *EvaluatingValue, *PriorityEvaluatingValue, *CallableValue:
		return true
	default:
		return false
	}
}

func bindBoundedValue(base *Context, vars []*SymbolNode, isTuple bool, value Value) (*Context, error) {
	if base == nil {
		base = EmptyContext
	}
	c1 := base
	if isTuple {
		tuple, ok := value.(*TupleValue)
		if !ok || len(tuple.Elems) != len(vars) {
			return nil, newTLCError(ECGeneral, "bounded tuple variable expected %d-tuple, got %s", len(vars), value)
		}
		for i, variable := range vars {
			c1 = c1.Cons(variable, tuple.Elems[i])
		}
		return c1, nil
	}
	if len(vars) == 0 {
		return nil, newTLCError(ECGeneral, "bounded quantifier has no variable")
	}
	return c1.Cons(vars[0], value), nil
}

func selectPath(root Value, path []Value) (Value, error) {
	cur := root
	for _, step := range path {
		switch v := cur.(type) {
		case *TupleValue:
			next, err := selectTupleValue(v, step)
			if err != nil {
				return nil, err
			}
			cur = next
		case *RecordValue:
			next, err := selectRecordValue(v, step)
			if err != nil {
				return nil, err
			}
			cur = next
		case *CounterExample:
			next, err := selectRecordValue(asRecordValue(v), step)
			if err != nil {
				return nil, err
			}
			cur = next
		case *FcnRcdValue:
			next, err := v.Select(step)
			if err != nil {
				return nil, err
			}
			cur = next
		case *FcnLambdaValue:
			next, err := v.Select(step)
			if err != nil {
				return nil, err
			}
			cur = next
		default:
			return nil, newTLCError(ECGeneral, "Attempted to apply EXCEPT construct to the value %s.", valueString(cur))
		}
		if cur == nil {
			return nil, nil
		}
	}
	return cur, nil
}

func selectTupleValue(v *TupleValue, arg Value) (Value, error) {
	return v.Select(arg)
}

func selectRecordValue(v *RecordValue, arg Value) (Value, error) {
	return v.Select(arg)
}

func stringUniqueFromNode(node SemanticNode) (*UniqueString, error) {
	switch n := node.(type) {
	case *StringNode:
		return n.Value.Val, nil
	case *ValueNode:
		if value, ok := n.Value.(*StringValue); ok {
			return value.Val, nil
		}
	case *StringValue:
		return n.Val, nil
	}
	if value, ok := SemanticToolObject(node).(*StringValue); ok {
		return value.Val, nil
	}
	return nil, newTLCError(ECGeneral, "record field name is not a string: %s", SemanticString(node))
}

func domainValue(expr SemanticNode, value Value) (Value, error) {
	switch v := value.(type) {
	case *TupleValue:
		return v.Domain(), nil
	case *RecordValue:
		return v.DomainValue(), nil
	case *CounterExample:
		record := asRecordValue(v)
		if record == nil {
			return nil, newTLCError(ECGeneral, "Attempted to apply the operator DOMAIN to a non-function\n(%s)\n%s", valueKindString(value), SemanticString(expr))
		}
		return record.DomainValue(), nil
	case *FcnRcdValue:
		return v.DomainValue(), nil
	case *FcnLambdaValue:
		return v.GetDomain()
	default:
		return nil, newTLCError(ECGeneral, "Attempted to apply the operator DOMAIN to a non-function\n(%s)\n%s", valueKindString(value), SemanticString(expr))
	}
}

func subsetEq(expr SemanticNode, left Value, right Value) (*BoolValue, error) {
	enum, ok := asEnumerable(left)
	if !ok {
		return nil, newTLCError(ECGeneral, "Attempted to evaluate an expression of form S \\subseteq T, but S was not enumerable.\n%s", SemanticString(expr))
	}
	return enumerableSubsetEq(enum, left, right)
}

func enumerableSubsetEq(enum Enumerable, left Value, right Value) (resultBool *BoolValue, err error) {
	if interval, ok := left.(*IntervalValue); ok {
		defer catchValueFailure(interval, &err)
		if other, ok := right.(*IntervalValue); ok && other.Low <= interval.Low && other.High >= interval.High {
			return BoolTrue, nil
		}
		return defaultEnumerableSubsetEq(enum, left, right)
	}
	if subset, ok := left.(*SubsetValue); ok {
		defer catchValueFailure(subset, &err)
		if other, ok := right.(*SubsetValue); ok {
			if setEnum, ok := asEnumerable(subset.Set); ok {
				return enumerableSubsetEq(setEnum, subset.Set, other.Set)
			}
		}
	}
	if _, ok := left.(*KSubsetValue); ok {
		// KSubsetValue inherits SubsetValue's catch but takes its naive fallback.
		defer catchValueFailure(left, &err)
	}
	return defaultEnumerableSubsetEq(enum, left, right)
}

func defaultEnumerableSubsetEq(enum Enumerable, left Value, right Value) (resultBool *BoolValue, err error) {
	defer catchValueFailure(left, &err)
	e := enum.Elements()
	for elem := e.NextElement(); elem != nil; elem = e.NextElement() {
		member, err := right.Member(elem)
		if err != nil {
			return nil, err
		}
		if !member {
			return BoolFalse, nil
		}
	}
	if err := e.Err(); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}
