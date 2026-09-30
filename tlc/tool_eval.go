package tlc

type toolEvalArgs struct {
	con     *Context
	s0      *TLCStateMut
	s1      *TLCStateMut
	control int
	cm      CostModel
}

func parseEvalArgs(args ...any) (*Context, *TLCStateMut, *TLCStateMut, int, CostModel) {
	parsed := toolEvalArgs{con: EmptyContext, s0: EmptyState, s1: EmptyState, control: EvalClear}
	if parsed.s0 == nil {
		parsed.s0 = NewEmptyState()
	}
	if parsed.s1 == nil {
		parsed.s1 = NewEmptyState()
	}
	for _, arg := range args {
		switch v := arg.(type) {
		case nil:
		case *Context:
			if v == nil {
				parsed.con = EmptyContext
			} else {
				parsed.con = v
			}
		case *TLCStateMut:
			if parsed.s0 == EmptyState {
				parsed.s0 = v
			} else {
				parsed.s1 = v
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
		if state != nil {
			if val := state.Lookup(sym.Name); val != nil {
				return val
			}
		}
		if t != nil {
			if val := t.Definitions[sym]; val != nil {
				return muxToolObject(val, state)
			}
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
		return v.ValueForWorker(workerIDFromState(state))
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
		return expr.Value, nil
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
		return NewOpLambdaValue(v, t, c, s0, s1, control), nil
	case Value:
		return v, nil
	default:
		return ValUndef, newTLCError(ECGeneral, "operator argument %s is undefined", expr.Op)
	}
}

func (t *Tool) EvalAppl(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	done := t.callStackEnter(expr)
	defer func() {
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
		if lv, ok := val.(*LazyValue); ok {
			if s1 == nil {
				return t.Eval(lv.Expr, lv.Con, s0, nil, control, cm)
			}
			return lv.GetValue(t, s0, s1, control)
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
			return t.Eval(v.Body, c, s0, s1, control, cm)
		case nil:
			if EvalIsPrimed(control) && opNode != nil && opNode.Name != nil && opNode.Name.VarLoc() >= 0 {
				return nil, newTLCError(ECGeneral, "state is not completely specified: %s in %s", opNode.Name, SemanticString(expr))
			}
			return nil, newTLCError(ECGeneral, "undefined operator %s in %s", opNode, SemanticString(expr))
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
		return NewSetOfTuplesValue(values), nil
	case OpcodeCL:
		for _, arg := range args {
			bval, err := t.evalBool(arg, c, s0, s1, control, cm, "conjunction")
			if err != nil || !bval.Val {
				return bval, err
			}
		}
		return BoolTrue, nil
	case OpcodeDL:
		for _, arg := range args {
			bval, err := t.evalBool(arg, c, s0, s1, control, cm, "disjunction")
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
		guard, err := t.evalBool(args[0], c, s0, s1, control, cm, "IF")
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
		return NewSetEnumValue(values, false), nil
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
		return NewSetOfFcnsValue(lhs, rhs), nil
	case OpcodeSSO:
		return t.evalSubsetOf(expr, c, s0, s1, control, cm)
	case OpcodeTup:
		values, err := t.evalArgs(args, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return NewTupleValue(values), nil
	case OpcodeUC:
		return nil, newTLCError(ECGeneral, "TLC attempted to evaluate an unbounded CHOOSE: %s", SemanticString(expr))
	case OpcodeUE:
		return nil, newTLCError(ECGeneral, "TLC attempted to evaluate an unbounded \\E: %s", SemanticString(expr))
	case OpcodeUF:
		return nil, newTLCError(ECGeneral, "TLC attempted to evaluate an unbounded \\A: %s", SemanticString(expr))
	case OpcodeLnot:
		arg, err := t.evalBool(args[0], c, s0, s1, control, cm, "~")
		if err != nil {
			return nil, err
		}
		return NewBoolValue(!arg.Val), nil
	case OpcodeSubset:
		arg, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return NewSubsetValue(arg), nil
	case OpcodeUnion:
		arg, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return NewUnionValue(arg), nil
	case OpcodeDomain:
		arg, err := t.Eval(args[0], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		return domainValue(arg)
	case OpcodeEnabled:
		sfun := NewEmptyState()
		if t.Enabled(args[0], BranchContext(c), s0, sfun) != nil {
			return BoolTrue, nil
		}
		return BoolFalse, nil
	case OpcodeEq:
		return t.evalEq(args[0], args[1], c, s0, s1, control, cm, false)
	case OpcodeLand:
		for _, arg := range args {
			bval, err := t.evalBool(arg, c, s0, s1, control, cm, "/\\")
			if err != nil || !bval.Val {
				return bval, err
			}
		}
		return BoolTrue, nil
	case OpcodeLor:
		for _, arg := range args {
			bval, err := t.evalBool(arg, c, s0, s1, control, cm, "\\/")
			if err != nil {
				return nil, err
			}
			if bval.Val {
				return BoolTrue, nil
			}
		}
		return BoolFalse, nil
	case OpcodeImplies:
		arg1, err := t.evalBool(args[0], c, s0, s1, control, cm, "=>")
		if err != nil {
			return nil, err
		}
		if !arg1.Val {
			return BoolTrue, nil
		}
		return t.evalBool(args[1], c, s0, s1, control, cm, "=>")
	case OpcodeEquiv:
		arg1, err := t.evalBool(args[0], c, s0, s1, control, cm, "<=>")
		if err != nil {
			return nil, err
		}
		arg2, err := t.evalBool(args[1], c, s0, s1, control, cm, "<=>")
		if err != nil {
			return nil, err
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
		return subsetEq(arg1, arg2)
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
		return t.evalSetOp(opcode, args[0], args[1], c, s0, s1, control, cm)
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
		return t.evalActionSubscript(opcode, args, c, s0, s1, control, cm)
	case OpcodeSF, OpcodeWF, OpcodeTE, OpcodeTF, OpcodeLeadsto, OpcodeArrow, OpcodeBox, OpcodeDiamond:
		return nil, newTLCError(ECGeneral, "TLC encountered temporal formula in a predicate: %s", SemanticString(expr))
	default:
		return nil, newTLCError(ECGeneral, "TLC BUG: could not evaluate expression %s", SemanticString(expr))
	}
}

func (t *Tool) GetVal(expr SemanticNode, c *Context, lazy bool, cm CostModel) any {
	if lazy {
		return NewLazyValue(expr, c, true, cm)
	}
	val, err := t.Eval(expr, c, EmptyState, EmptyState, EvalClear, cm)
	if err != nil {
		return ValUndef
	}
	return val
}

func (t *Tool) GetOpContext(opDef *OpDefNode, args []SemanticNode, c *Context, lazy bool, cm CostModel) (*Context, error) {
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
		if lazy {
			c1 = c1.Cons(param, NewLazyValue(args[i], c, true, cm))
			continue
		}
		val, err := t.Eval(args[i], c, EmptyState, EmptyState, EvalClear, cm)
		if err != nil {
			return c1, err
		}
		c1 = c1.Cons(param, val)
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
	case *EvaluatingValue:
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
	vars := make([]any, len(expr.BdedQuantBounds))
	enums := make([]ValueEnumeration, len(expr.BdedQuantBounds))
	for i, bound := range expr.BdedQuantBounds {
		val, err := t.Eval(bound, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		enumerable, ok := asEnumerable(val)
		if !ok {
			return nil, newTLCError(ECGeneral, "bounded quantifier domain is not enumerable: %s", val)
		}
		if i < len(expr.BdedQuantATuple) && expr.BdedQuantATuple[i] {
			vars[i] = expr.BdedQuantSymbolLists[i]
		} else if i < len(expr.BdedQuantSymbolLists) && len(expr.BdedQuantSymbolLists[i]) > 0 {
			vars[i] = expr.BdedQuantSymbolLists[i][0]
		} else {
			return nil, newTLCError(ECGeneral, "bounded quantifier has no bound variables")
		}
		enums[i] = enumerable.Elements()
	}
	return NewContextEnumerator(vars, enums, c), nil
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

func (t *Tool) evalBoundedChoose(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if len(expr.Args) == 0 || len(expr.BdedQuantBounds) == 0 {
		return nil, newTLCError(ECGeneral, "malformed bounded CHOOSE")
	}
	inVal, err := t.Eval(expr.BdedQuantBounds[0], c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	enumerable, ok := asEnumerable(inVal.Normalize())
	if !ok {
		return nil, newTLCError(ECGeneral, "CHOOSE domain is not enumerable: %s", inVal)
	}
	pred := expr.Args[0]
	bvars := expr.BdedQuantSymbolLists[0]
	isTuple := len(expr.BdedQuantATuple) > 0 && expr.BdedQuantATuple[0]
	enum := enumerable.Elements()
	for val := enum.NextElement(); val != nil; val = enum.NextElement() {
		c1, err := bindBoundedValue(c, bvars, isTuple, val)
		if err != nil {
			return nil, err
		}
		bval, err := t.evalBool(pred, c1, s0, s1, control, cm, "CHOOSE")
		if err != nil {
			return nil, err
		}
		if bval.Val {
			return val, nil
		}
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	return nil, newTLCError(ECGeneral, "CHOOSE had no satisfying element: %s", SemanticString(expr))
}

func (t *Tool) evalBoundedExists(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	enum, err := t.Contexts(expr, c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
		bval, err := t.evalBool(expr.Args[0], c1, s0, s1, control, cm, "\\E")
		if err != nil {
			return nil, err
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
		bval, err := t.evalBool(expr.Args[0], c1, s0, s1, control, cm, "\\A")
		if err != nil {
			return nil, err
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
			continue
		}
		guard, err := t.evalBool(pair.Args[0], c, s0, s1, control, cm, "CASE")
		if err != nil {
			return nil, err
		}
		if guard.Val {
			return t.Eval(pair.Args[1], c, s0, s1, control, cm)
		}
	}
	if other == nil {
		return nil, newTLCError(ECGeneral, "CASE had no true condition: %s", SemanticString(expr))
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
		lhs, err := t.evalArgs(pathNode.Args, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		atVal, err := selectPath(result, lhs)
		if err != nil {
			return nil, err
		}
		if atVal == nil {
			continue
		}
		atSym := NewSymbolNode("@")
		rhs, err := t.Eval(pair.Args[1], c.Cons(atSym, atVal), s0, s1, control, cm)
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
	argVal, err := t.Eval(args[1], c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	switch f := fval.(type) {
	case *FcnRcdValue:
		return f.Apply(argVal)
	case *FcnLambdaValue:
		return f.ApplyWithControl(argVal, control)
	case *TupleValue:
		return f.Apply(argVal)
	case *RecordValue:
		return f.Apply(argVal)
	default:
		return nil, newTLCError(ECGeneral, "a non-function (%s) was applied as a function", fval.KindString())
	}
}

func (t *Tool) evalFcnConstructor(expr *OpApplNode, opcode int, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	dvals := make([]Value, len(expr.BdedQuantBounds))
	for i, domain := range expr.BdedQuantBounds {
		value, err := t.Eval(domain, c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		dvals[i] = value
	}
	params := NewFcnParams(expr.BdedQuantSymbolLists, expr.BdedQuantATuple, dvals)
	fval := NewFcnLambdaValue(params, expr.Args[0], t, c, s0, s1, control)
	if opcode == OpcodeRFS && len(expr.UnbdedQuantSymbols) > 0 {
		fval.MakeRecursive(expr.UnbdedQuantSymbols[0])
	}
	if !EvalIsKeepLazy(control) {
		if fcn := fval.ToFcnRcd(); fcn != nil {
			return fcn, nil
		}
	}
	return fval, nil
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
		value, err := t.Eval(pair.Args[1], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		names[i] = name
		values[i] = value
	}
	return NewRecordValue(names, values, false), nil
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
		result := record.Select(sval)
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
	return NewSetEnumValueVec(values, false), nil
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
		value, err := t.Eval(pair.Args[1], c, s0, s1, control, cm)
		if err != nil {
			return nil, err
		}
		names[i] = name
		values[i] = value
	}
	return NewSetOfRcdsValue(names, values, false)
}

func (t *Tool) evalSubsetOf(expr *OpApplNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	inVal, err := t.Eval(expr.BdedQuantBounds[0], c, s0, s1, control, cm)
	if err != nil {
		return nil, err
	}
	bvars := expr.BdedQuantSymbolLists[0]
	isTuple := len(expr.BdedQuantATuple) > 0 && expr.BdedQuantATuple[0]
	if enum, ok := asEnumerable(inVal); ok {
		values := NewValueVec(0)
		e := enum.Elements()
		for elem := e.NextElement(); elem != nil; elem = e.NextElement() {
			c1, err := bindBoundedValue(c, bvars, isTuple, elem)
			if err != nil {
				return nil, err
			}
			bval, err := t.evalBool(expr.Args[0], c1, s0, s1, control, cm, "set predicate")
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
		return NewSetEnumValueVec(values, inVal.IsNormalized()), nil
	}
	if isTuple {
		return NewSetPredValue(bvars, inVal, expr.Args[0], t, c, s0, s1, control), nil
	}
	return NewSetPredValue(bvars[0], inVal, expr.Args[0], t, c, s0, s1, control), nil
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
		return NewSetDiffValue(arg1, arg2), nil
	case OpcodeCap:
		return NewSetCapValue(arg1, arg2), nil
	case OpcodeCup:
		return NewSetCupValue(arg1, arg2), nil
	default:
		return nil, newTLCError(ECGeneral, "unknown set operator opcode %d", opcode)
	}
}

func (t *Tool) evalActionSubscript(opcode int, args []SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	res, err := t.evalBool(args[0], c, s0, s1, control, cm, "action subscript")
	if err != nil {
		return nil, err
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
			cur = v.Select(step)
		case *RecordValue:
			cur = v.Select(step)
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
			return nil, newTLCError(ECGeneral, "cannot select %s from %s", step, cur)
		}
		if cur == nil {
			return nil, nil
		}
	}
	return cur, nil
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

func domainValue(value Value) (Value, error) {
	switch v := value.(type) {
	case *TupleValue:
		return v.Domain(), nil
	case *RecordValue:
		return v.DomainValue(), nil
	case *FcnRcdValue:
		return v.DomainValue(), nil
	case *FcnLambdaValue:
		fcn := v.ToFcnRcd()
		if fcn == nil {
			return nil, newTLCError(ECGeneral, "could not materialize function domain for %s", value)
		}
		return fcn.DomainValue(), nil
	default:
		return nil, newTLCError(ECGeneral, "attempted to apply DOMAIN to non-function %s", value)
	}
}

func subsetEq(left Value, right Value) (*BoolValue, error) {
	enum, ok := asEnumerable(left)
	if !ok {
		return nil, newTLCError(ECGeneral, "left side of \\subseteq is not enumerable: %s", left)
	}
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
