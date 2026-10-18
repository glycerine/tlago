package tlc

import (
	"fmt"
	"os"
)

const (
	actionCompositionProperty           = "tlc2.tool.impl.Tool.cdot"
	actionCompositionUnsupportedMessage = "The current version of TLC does not support action composition.  An incomplete implementation can be enabled via the tlc2.tool.impl.Tool.cdot=true java property."
	toolProbabilisticProperty           = "tlc2.tool.impl.Tool.probabilistic"
)

func (t *Tool) GetInitStatesImpl(functor *StateFunctor) error {
	if t == nil {
		panic(NewNullPointerException())
	}
	init := t.GetInitStateSpec()
	acts := emptyActionItemListExt
	for i := int32(init.Size()) - 1; i > 0; i-- {
		elem := init.ElementAt(int(i))
		if elem == nil {
			panic(NewNullPointerException())
		}
		acts = acts.ConsAction(elem, ActionItemPred)
	}
	if init.Size() == 0 {
		return nil
	}
	elem := init.ElementAt(0)
	if elem == nil {
		panic(NewNullPointerException())
	}
	ps := NewEmptyState()
	if acts.IsEmpty() {
		acts.SetAction(elem)
	}
	return t.GetInitStatesForPredicate(elem.Pred, acts, elem.Con, ps, functor, elem.CM)
}

func (t *Tool) MakeStateImpl(pred SemanticNode) (*TLCStateMut, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	vec := NewStateVec(0)
	functor := NewStateFunctor(func(state *TLCStateMut) (any, error) {
		vec.Add(state)
		return vec, nil
	})
	if err := t.GetInitStatesForPredicate(pred, EmptyActionItemList, EmptyContext, NewEmptyState(), functor, CostModel{}); err != nil {
		return nil, err
	}
	if vec.Size() != 1 {
		return nil, NewTLCDetailedRuntimeException(ECGeneral, "The predicate does not specify a unique state."+SemanticString(pred), pred, EmptyContext)
	}
	state := vec.At(0)
	if !t.IsGoodState(state) {
		return nil, NewTLCDetailedRuntimeException(ECGeneral, "The state specified by the predicate is not complete."+SemanticString(pred), pred, EmptyContext)
	}
	return state, nil
}

// ProcessUnsatisfiedState is the initial-state source overload of
// processUnsatisfied.
func (t *Tool) ProcessUnsatisfiedState(state *TLCStateMut, pred SemanticNode, context *Context, functor *StateFunctor, _ CostModel) *TLCStateMut {
	if t == nil || functor == nil {
		panic(NewNullPointerException())
	}
	if debugToolActive(t) {
		functor = debugWrapStateFunctor(t, functor)
	}
	return functor.AddUnsatisfiedState(state, pred, context)
}

// ProcessUnsatisfiedNextState is the successor source overload of
// processUnsatisfied.
func (t *Tool) ProcessUnsatisfiedNextState(state *TLCStateMut, action *Action, successor *TLCStateMut, pred SemanticNode, context *Context, functor *NextStateFunctor, _ CostModel) *TLCStateMut {
	if t == nil || functor == nil {
		panic(NewNullPointerException())
	}
	if debugToolActive(t) {
		functor = debugWrapNextStateFunctor(t, functor)
	}
	return functor.AddUnsatisfiedNextState(state, action, successor, pred, context)
}

func (t *Tool) GetInitStatesForPredicate(init SemanticNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) (err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	done := t.callStackEnter(init)
	defer func() { done(err) }()
	if debugToolActive(t) {
		if t.DebugEvalMode == DebugEvalDebugger {
			return t.NoDebug().GetInitStatesForPredicate(init, acts, c, ps, states, cm)
		}
		t.DebugEvalMode = DebugEvalState
		states = debugWrapStateFunctor(t, states)
		err = debugGetInitStatesImplWithReset(t, init, acts, c, ps, states, cm)
		if debugResetTargets(err, init) {
			if retryErr := t.GetInitStatesForPredicate(init, acts, c, ps, states, cm); retryErr != nil {
				return retryErr
			}
		}
		return err
	}
	return t.getInitStatesForPredicateImpl(init, acts, c, ps, states, cm)
}

func debugGetInitStatesImplWithReset(t *Tool, init SemanticNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if reset, ok := recovered.(*ResetEvalException); ok {
				err = reset
			} else {
				panic(recovered)
			}
		}
	}()
	return t.getInitStatesForPredicateImpl(init, acts, c, ps, states, cm)
}

func (t *Tool) getInitStatesForPredicateImpl(init SemanticNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	if isNil(init) {
		panic(NewNullPointerException())
	}
	switch init := init.(type) {
	case *OpApplNode:
		return t.GetInitStatesAppl(init, acts, c, ps, states, cm)
	case *LetInNode:
		return t.GetInitStatesForPredicate(init.Body, acts, c, ps, states, cm)
	case *SubstInNode:
		c1 := c
		for _, sub := range init.Substs {
			subCM := cm
			if specCoverageEnabled() {
				subCM = cm.GetSubst(sub)
			}
			c1 = c1.Cons(sub.GetOp(), t.GetVal(sub.GetExpr(), c, false, subCM))
		}
		return t.GetInitStatesForPredicate(init.GetBody(), acts, c1, ps, states, cm)
	case *APSubstInNode:
		c1 := c
		for _, sub := range init.Substs {
			c1 = c1.Cons(sub.GetOp(), t.GetVal(sub.GetExpr(), c, false, cm))
		}
		return t.GetInitStatesForPredicate(init.GetBody(), acts, c1, ps, states, cm)
	case *LabelNode:
		return t.GetInitStatesForPredicate(init.Body, acts, c, ps, states, cm)
	default:
		return NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("The init state relation is not a boolean expression.\n%s", SemanticString(init)), init, c)
	}
}

func (t *Tool) GetInitStatesFromActionList(acts *ActionItemList, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	if t == nil {
		panic(NewNullPointerException())
	}
	if acts.IsEmpty() {
		if specCoverageEnabled() {
			cm.IncInvocations()
		}
		if CoverageActionEnabled() {
			cm.GetRoot().IncInvocations()
		}
		_, err := states.AddElement(ps.Copy().SetAction(acts.GetAction()))
		return err
	}
	if ps.AllAssigned() {
		for !acts.IsEmpty() {
			pred := acts.CarPred()
			value, err := t.Eval(pred, acts.CarContext(), ps, EmptyState, EvalInit, acts.ownedCostModel())
			if err != nil {
				return err
			}
			bval, ok := value.(*BoolValue)
			if !ok {
				return newTLCDetailedRuntimeExceptionCode(ECTLCExpectedExpressionInComputing, pred, acts.CarContext(), "initial states", "boolean", value.String(), SemanticString(pred))
			}
			if !bval.Val {
				if specCoverageEnabled() {
					cm.GetRoot().IncSecondary()
				}
				t.ProcessUnsatisfiedState(ps, acts.CarPred(), acts.CarContext(), states, cm)
				return nil
			}
			acts = acts.Cdr()
		}
		if specCoverageEnabled() {
			cm.IncInvocations()
			cm.GetRoot().IncInvocations()
		}
		_, err := states.AddElement(ps.Copy().SetAction(acts.GetAction()))
		return err
	}
	acts1 := acts.Cdr()
	return t.GetInitStatesForPredicate(acts.CarPred(), acts1, acts.CarContext(), ps, states, acts.ownedCostModel())
}

func (t *Tool) GetInitStatesAppl(init *OpApplNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) (err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	done := t.callStackEnter(init)
	defer func() { done(err) }()
	if debugToolActive(t) {
		debugger := t.ownedDebugger()
		if t.DebugEvalMode == DebugEvalDebugger {
			return t.NoDebug().GetInitStatesAppl(init, acts, c, ps, states, cm)
		}
		t.DebugEvalMode = DebugEvalState
		debugger.PushStateFrame(t, init, c, ps)
		defer debugger.PopFrame(t, init, c)
	}
	if specCoverageEnabled() {
		cm = cm.Get(init)
	}
	args := init.Args
	opNode := init.Operator
	opcode := GetOpCode(opNode.Name)
	if opcode == 0 {
		val := t.Lookup(opNode, c, ps, false)
		if lv := asLazyValue(val); lv != nil {
			cached := lv.GetCachedValue(t, ps, nil, EvalClear)
			if cached == nil {
				return t.GetInitStatesForPredicate(lv.ownedExpr(), acts, lv.ownedContext(), ps, states, cm)
			}
			val = cached
		}
		switch v := val.(type) {
		case *OpDefNode:
			opcode = GetOpCode(v.Name)
			if opcode == 0 {
				c1, err := t.GetOpContext(v, args, c, true, cm)
				if err != nil {
					return err
				}
				return t.GetInitStatesForPredicate(v.GetBody(), acts, c1, ps, states, cm)
			}
		case *ThmOrAssumpDefNode:
			c1, err := t.GetThmOrAssumpContext(v, args, c, true)
			if err != nil {
				return err
			}
			return t.GetInitStatesForPredicate(v.GetBody(), acts, c1, ps, states, cm)
		case Value:
			if isOperatorValue(v) {
				bval, err := EvalOperatorValueWithTool(v, t, args, c, ps, EmptyState, EvalInit, cm)
				if err != nil {
					return err
				}
				return t.continueInitIfBool(init, bval, c, acts, ps, states, cm)
			}
			return t.continueInitIfBool(init, v, c, acts, ps, states, cm)
		default:
			if isNil(val) {
				return NewNullPointerException()
			}
			return newTLCDetailedRuntimeExceptionCode(ECTLCExpectedExpressionInComputing, init, c, "initial states", "boolean", fmt.Sprint(val), SemanticString(init))
		}
	}

	switch opcode {
	case OpcodeDL, OpcodeLor:
		if len(args) == 0 {
			return nil
		}
		for _, arg := range args {
			if err := t.GetInitStatesForPredicate(arg, acts, c, ps, states, cm); err != nil {
				return err
			}
		}
		return nil
	case OpcodeCL, OpcodeLand:
		acts1 := acts
		for i := len(args) - 1; i > 0; i-- {
			acts1 = acts1.Cons(args[i], c, cm, i)
		}
		if len(args) == 0 {
			panic(NewArrayIndexOutOfBoundsException(0, 0))
		}
		return t.GetInitStatesForPredicate(args[0], acts1, c, ps, states, cm)
	case OpcodeBE:
		enum, err := t.Contexts(init, c, ps, EmptyState, EvalInit, cm)
		if err != nil {
			return err
		}
		for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
			if err := enum.Err(); err != nil {
				return err
			}
			if err := t.GetInitStatesForPredicate(args[0], acts, c1, ps, states, cm); err != nil {
				return err
			}
		}
		return enum.Err()
	case OpcodeBF:
		return t.initBoundedForall(init, acts, c, ps, states, cm)
	case OpcodeITE:
		guard, err := t.Eval(args[0], c, ps, EmptyState, EvalInit, cm)
		if err != nil {
			return err
		}
		bguard, ok := guard.(*BoolValue)
		if !ok {
			return NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("In computing initial states, a non-boolean expression (%s) was used as the condition of an IF.\n%s", ValueKindString(guard), SemanticString(init)), init, c)
		}
		idx := 2
		if bguard.Val {
			idx = 1
		}
		return t.GetInitStatesForPredicate(args[idx], acts, c, ps, states, cm)
	case OpcodeCase:
		return t.initCase(init, acts, c, ps, states, cm)
	case OpcodeFA:
		return t.initFcnApply(init, acts, c, ps, states, cm)
	case OpcodeEq:
		return t.initEquality(init, args[0], args[1], acts, c, ps, states, cm)
	case OpcodeSubseteq:
		return t.initSubsetEq(init, args[0], args[1], acts, c, ps, states, cm)
	case OpcodeIn:
		return t.initMembership(init, args[0], args[1], acts, c, ps, states, cm)
	case OpcodeImplies:
		value, err := t.Eval(args[0], c, ps, EmptyState, EvalInit, cm)
		if err != nil {
			return err
		}
		lval, ok := value.(*BoolValue)
		if !ok {
			return NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("In computing initial states of a predicate of form P => Q, P was %s\n.%s", valueKindString(value), SemanticString(init)), init, c)
		}
		if lval.Val {
			return t.GetInitStatesForPredicate(args[1], acts, c, ps, states, cm)
		}
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	case OpcodeNop:
		return t.GetInitStatesForPredicate(args[0], acts, c, ps, states, cm)
	default:
		bval, err := t.Eval(init, c, ps, EmptyState, EvalInit, cm)
		if err != nil {
			return err
		}
		value, ok := bval.(*BoolValue)
		if !ok {
			return NewTLCDetailedRuntimeException(ECGeneral, "In computing initial states, TLC expected a boolean expression,\nbut instead found "+bval.String()+".\n"+SemanticString(init), init, c)
		}
		if value.Val {
			return t.GetInitStatesFromActionList(acts, ps, states, cm)
		}
		return nil
	}
}

func (t *Tool) continueInitIfBool(init SemanticNode, value Value, c *Context, acts *ActionItemList, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	bval, ok := value.(*BoolValue)
	if !ok {
		return newTLCDetailedRuntimeExceptionCode(ECTLCExpectedExpressionInComputing, init, c, "initial states", "boolean", value.String(), SemanticString(init))
	}
	if bval.Val {
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	return nil
}

func (t *Tool) initBoundedForall(init *OpApplNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	enum, err := t.Contexts(init, c, ps, EmptyState, EvalInit, cm)
	if err != nil {
		return err
	}
	c1 := enum.NextElement()
	if err := enum.Err(); err != nil {
		return err
	}
	if c1 == nil {
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	acts1 := acts
	for c2 := enum.NextElement(); c2 != nil; c2 = enum.NextElement() {
		if err := enum.Err(); err != nil {
			return err
		}
		acts1 = acts1.Cons(init.Args[0], c2, cm, ActionItemPred)
	}
	if err := enum.Err(); err != nil {
		return err
	}
	return t.GetInitStatesForPredicate(init.Args[0], acts1, c1, ps, states, cm)
}

func (t *Tool) initCase(init *OpApplNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	var other SemanticNode
	for _, arg := range init.Args {
		pair, ok := arg.(*OpApplNode)
		if !ok || len(pair.Args) < 2 {
			return newTLCError(ECGeneral, "malformed CASE in initial predicate")
		}
		if pair.Args[0] == nil {
			other = pair.Args[1]
			continue
		}
		value, err := t.Eval(pair.Args[0], c, ps, EmptyState, EvalInit, cm)
		if err != nil {
			return err
		}
		bval, ok := value.(*BoolValue)
		if !ok {
			return NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("In computing initial states, a non-boolean expression (%s) was used as a guard condition of a CASE.\n%s", ValueKindString(value), SemanticString(pair.Args[1])), pair.Args[1], c)
		}
		if bval.Val {
			return t.GetInitStatesForPredicate(pair.Args[1], acts, c, ps, states, cm)
		}
	}
	if other == nil {
		return NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("In computing initial states, TLC encountered a CASE with no conditions true.\n%s", SemanticString(init)), init, c)
	}
	return t.GetInitStatesForPredicate(other, acts, c, ps, states, cm)
}

func (t *Tool) initEquality(init SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	varNode := t.GetVar(left, c, false)
	if varNode == nil || varNode.Name.VarLoc() < 0 {
		bval, err := t.evalBool(init, c, ps, EmptyState, EvalInit, cm)
		if err != nil || !bval.Val {
			return err
		}
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	varName := varNode.Name
	lval := ps.Lookup(varName)
	rval, err := t.Eval(right, c, ps, EmptyState, EvalInit, cm)
	if err != nil {
		return err
	}
	if lval == nil {
		ps.Bind(varName, rval)
		err = t.GetInitStatesFromActionList(acts, ps, states, cm)
		if err == nil {
			ps.Unbind(varName)
		}
		return err
	}
	eq, err := lval.Equal(rval)
	if err != nil || !eq {
		return err
	}
	return t.GetInitStatesFromActionList(acts, ps, states, cm)
}

func (t *Tool) initMembership(init SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	varNode := t.GetVar(left, c, false)
	if varNode == nil || varNode.Name.VarLoc() < 0 {
		bval, err := t.evalBool(init, c, ps, EmptyState, EvalInit, cm)
		if err != nil || !bval.Val {
			return err
		}
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	lval := ps.Lookup(varNode.Name)
	rval, err := t.Eval(right, c, ps, EmptyState, EvalInit, cm)
	if err != nil {
		return err
	}
	return t.enumerateInitAssignment(varNode.Name, lval, rval, init, c, acts, ps, states, cm)
}

func (t *Tool) initSubsetEq(init SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	varNode := t.GetVar(left, c, false)
	if varNode == nil || varNode.Name.VarLoc() < 0 {
		bval, err := t.evalBool(init, c, ps, EmptyState, EvalInit, cm)
		if err != nil || !bval.Val {
			return err
		}
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	lval := ps.Lookup(varNode.Name)
	rset, err := t.Eval(right, c, ps, EmptyState, EvalInit, cm)
	if err != nil {
		return err
	}
	return t.enumerateInitAssignment(varNode.Name, lval, t.SetSource(right, NewSubsetValue(rset, cm)), init, c, acts, ps, states, cm)
}

func (t *Tool) enumerateInitAssignment(varName *UniqueString, lval Value, domain Value, pred SemanticNode, c *Context, acts *ActionItemList, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	if lval != nil {
		member, err := domain.Member(lval)
		if err != nil || !member {
			return err
		}
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	enumerable, ok := asEnumerable(domain)
	if !ok {
		return NewTLCDetailedRuntimeException(ECGeneral, "In computing initial states, the right side of \\IN is not enumerable.\n"+SemanticString(pred), pred, c)
	}
	enum := enumerable.Elements()
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		if err := enum.Err(); err != nil {
			return err
		}
		ps.Bind(varName, elem)
		if err := t.GetInitStatesFromActionList(acts, ps, states, cm); err != nil {
			return err
		}
		ps.Unbind(varName)
	}
	return enum.Err()
}

func (t *Tool) getNextStateVec(action *Action, state *TLCStateMut) (*StateVec, error) {
	return t.getNextStatesWithContext(action, action.ownedContext(), state)
}

func (t *Tool) getNextStatesWithContext(action *Action, context *Context, state *TLCStateMut) (*StateVec, error) {
	nss := NewStateVec(0)
	functor := NewNextStateFunctor(func(_ *TLCStateMut, _ *Action, succ *TLCStateMut) (any, error) {
		nss.Add(succ)
		return nss, nil
	})
	functor.HasStatesFunc = nss.HasStates
	s1 := NewEmptyState().SetPredecessor(state).SetAction(action)
	_, err := t.GetNextStatesForPredicate(action, action.ownedPred(), EmptyActionItemList, context, state, s1, functor, action.CM)
	if err == nil && action != nil && specCoverageEnabled() {
		action.CM.IncInvocations(int64(nss.Size()))
	}
	if err == nil && toolProbabilisticEnabled() && nss.Size() > 1 {
		fmt.Fprintln(os.Stderr, "Simulator generated more than one next state")
	}
	return nss, err
}

func (t *Tool) GetNextStatesForPredicate(action *Action, pred SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (state *TLCStateMut, err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	done := t.callStackEnter(pred)
	defer func() { done(err) }()
	if debugToolActive(t) {
		if t.DebugEvalMode == DebugEvalDebugger {
			return t.NoDebug().GetNextStatesForPredicate(action, pred, acts, c, s0, s1, nss, cm)
		}
		t.DebugEvalMode = DebugEvalAction
		s1 = s1.SetPredecessor(s0).SetAction(action)
		state, err = debugGetNextStatesImplWithReset(t, action, pred, acts, c, s0, s1, nss, cm)
		if debugResetTargets(err, pred) {
			return t.GetNextStatesForPredicate(action, pred, acts, c, s0, s1, nss, cm)
		}
		return state, err
	}
	return t.GetNextStatesImpl(action, pred, acts, c, s0, s1, nss, cm)
}

func debugGetNextStatesImplWithReset(t *Tool, action *Action, pred SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (state *TLCStateMut, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if reset, ok := recovered.(*ResetEvalException); ok {
				err = reset
			} else {
				panic(recovered)
			}
		}
	}()
	return t.GetNextStatesImpl(action, pred, acts, c, s0, s1, nss, cm)
}

func (t *Tool) GetNextStatesImpl(action *Action, pred SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (state *TLCStateMut, err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if pred == nil {
		panic(NewNullPointerException())
	}
	switch pred := pred.(type) {
	case *OpApplNode:
		if specCoverageEnabled() {
			cm = cm.Get(pred)
		}
		return t.GetNextStatesAppl(action, pred, acts, c, s0, s1, nss, cm)
	case *LetInNode:
		return t.GetNextStatesForPredicate(action, pred.Body, acts, c, s0, s1, nss, cm)
	case *SubstInNode:
		c1 := c
		for _, sub := range pred.Substs {
			subCM := cm
			if specCoverageEnabled() {
				subCM = cm.GetSubst(sub)
			}
			c1 = c1.Cons(sub.GetOp(), t.GetVal(sub.GetExpr(), c, false, subCM))
		}
		return t.GetNextStatesForPredicate(action, pred.GetBody(), acts, c1, s0, s1, nss, cm)
	case *APSubstInNode:
		c1 := c
		for _, sub := range pred.Substs {
			c1 = c1.Cons(sub.GetOp(), t.GetVal(sub.GetExpr(), c, false, cm))
		}
		return t.GetNextStatesForPredicate(action, pred.GetBody(), acts, c1, s0, s1, nss, cm)
	case *LabelNode:
		return t.GetNextStatesForPredicate(action, pred.Body, acts, c, s0, s1, nss, cm)
	default:
		return s1, NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("The next state relation is not a boolean expression.\n%s", SemanticString(pred)), pred, c)
	}
}

func (t *Tool) GetNextStatesFromActionList(action *Action, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	copyState, err := t.getNextStates0(action, acts, s0, s1, nss, cm)
	if err != nil {
		return copyState, err
	}
	if specCoverageEnabled() && copyState != s1 {
		cm.IncInvocations()
	}
	return copyState, nil
}

func (t *Tool) getNextStates0(action *Action, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	if acts.IsEmpty() {
		if _, err := nss.AddNextElement(s0, action, s1); err != nil {
			return s1, err
		}
		return s1.Copy(), nil
	}
	if Globals.Warn && s1.AllAssigned() {
		return t.getNextStatesAllAssigned(action, acts, s0, s1, nss, cm)
	}
	kind := acts.CarKind()
	pred := acts.CarPred()
	c := acts.CarContext()
	acts1 := acts.Cdr()
	switch {
	case kind > ActionItemConjunct || kind == ActionItemPred:
		return t.GetNextStatesForPredicate(action, pred, acts1, c, s0, s1, nss, acts.ownedCostModel())
	case kind == ActionItemUnchanged:
		return t.ProcessUnchanged(action, pred, acts1, c, s0, s1, nss, acts.ownedCostModel())
	default:
		v1, err := t.Eval(pred, c, s0, EmptyState, EvalClear, acts.ownedCostModel())
		if err != nil {
			return s1, err
		}
		v2, err := t.Eval(pred, c, s1, EmptyState, EvalClear, acts.ownedCostModel())
		if err != nil {
			return s1, err
		}
		eq, err := v1.Equal(v2)
		if err != nil || eq {
			return s1, err
		}
		if specCoverageEnabled() {
			return t.GetNextStatesFromActionList(action, acts1, s0, s1, nss, acts.ownedCostModel())
		}
		return t.getNextStates0(action, acts1, s0, s1, nss, acts.ownedCostModel())
	}
}

func (t *Tool) getNextStatesAllAssigned(action *Action, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	for !acts.IsEmpty() {
		kind := acts.CarKind()
		pred := acts.CarPred()
		c := acts.CarContext()
		cm2 := acts.ownedCostModel()
		switch {
		case kind > ActionItemConjunct || kind == ActionItemPred:
			value, err := t.Eval(pred, c, s0, s1, EvalClear, cm2)
			if err != nil {
				return s1, err
			}
			bval, ok := value.(*BoolValue)
			if !ok {
				return s1, newTLCDetailedRuntimeExceptionCode(ECTLCExpectedExpressionInComputing, pred, c, "next states", "boolean", value.String(), SemanticString(pred))
			}
			if !bval.Val {
				return t.ProcessUnsatisfiedNextState(s0, action, s1, pred, c, nss, cm), nil
			}
		case kind == ActionItemUnchanged:
			return t.ProcessUnchanged(action, pred, acts.Cdr(), c, s0, s1, nss, cm2)
		default:
			v1, err := t.Eval(pred, c, s0, EmptyState, EvalClear, cm2)
			if err != nil {
				return s1, err
			}
			v2, err := t.Eval(pred, c, s1, EmptyState, EvalClear, cm2)
			if err != nil {
				return s1, err
			}
			eq, err := v1.Equal(v2)
			if err != nil || eq {
				return s1, err
			}
		}
		acts = acts.Cdr()
	}
	if _, err := nss.AddNextElement(s0, action, s1); err != nil {
		return s1, err
	}
	return s1.Copy(), nil
}

func (t *Tool) GetNextStatesAppl(action *Action, pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (state *TLCStateMut, err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	done := t.callStackEnter(pred)
	defer func() { done(err) }()
	if debugToolActive(t) {
		debugger := t.ownedDebugger()
		if t.DebugEvalMode == DebugEvalDebugger {
			return t.NoDebug().GetNextStatesAppl(action, pred, acts, c, s0, s1, nss, cm)
		}
		t.DebugEvalMode = DebugEvalAction
		debugger.PushActionFrame(t, pred, c, s0, action, s1)
		defer func() {
			if !debugErrorKnown(err) {
				if IsInvariantViolatedException(err) {
					debugger.MarkInvariantViolatedFrame(t, pred, c, s0, action, s1, err)
					debugger.PopExceptionFrame(t, pred, c, nil, err)
				} else if isJavaEvalOrRuntimeException(err) {
					// This source catch uses the base-frame overload.
					debugger.PushExceptionFrame(t, pred, c, err)
					debugger.PopExceptionFrame(t, pred, c, nil, err)
				}
			}
			debugger.PopFrame(t, pred, c)
		}()
	}
	return t.GetNextStatesApplImpl(action, pred, acts, c, s0, s1, nss, cm)
}

func (t *Tool) GetNextStatesApplImpl(action *Action, pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (state *TLCStateMut, err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	args := pred.Args
	opNode := pred.Operator
	opcode := GetOpCode(opNode.Name)
	if opcode == 0 {
		val := t.Lookup(opNode, c, s0, false)
		if lv := asLazyValue(val); lv != nil {
			cached := lv.GetCachedValue(t, s0, s1, EvalClear)
			if cached == nil {
				return t.GetNextStatesForPredicate(action, lv.ownedExpr(), acts, lv.ownedContext(), s0, s1, nss, lv.CM)
			}
			val = cached
		}
		switch v := val.(type) {
		case *OpDefNode:
			opcode = GetOpCode(v.Name)
			if opcode == 0 {
				c1, err := t.GetOpContext(v, args, c, true, cm)
				if err != nil {
					return s1, err
				}
				return t.GetNextStatesForPredicate(action, v.GetBody(), acts, c1, s0, s1, nss, cm)
			}
		case *ThmOrAssumpDefNode:
			c1, err := t.GetThmOrAssumpContext(v, args, c, true)
			if err != nil {
				return s1, err
			}
			return t.GetNextStatesForPredicate(action, v.GetBody(), acts, c1, s0, s1, nss, cm)
		case Value:
			if isOperatorValue(v) {
				bval, err := EvalOperatorValueWithTool(v, t, args, c, s0, s1, EvalClear, cm)
				if err != nil {
					return s1, err
				}
				return t.continueNextIfBool(action, pred, bval, EmptyContext, acts, s0, s1, nss, cm)
			}
			return t.continueNextIfBool(action, pred, v, EmptyContext, acts, s0, s1, nss, cm)
		default:
			if isNil(val) {
				return s1, NewNullPointerException()
			}
			return s1, newTLCDetailedRuntimeExceptionCode(ECTLCExpectedExpressionInComputing, pred, EmptyContext, "next states", "boolean", fmt.Sprint(val), SemanticString(pred))
		}
	}
	switch opcode {
	case OpcodeCL, OpcodeLand:
		acts1 := acts
		for i := len(args) - 1; i > 0; i-- {
			acts1 = acts1.Cons(args[i], c, cm, i)
		}
		if len(args) == 0 {
			panic(NewArrayIndexOutOfBoundsException(0, 0))
		}
		return t.GetNextStatesForPredicate(action, args[0], acts1, c, s0, s1, nss, cm)
	case OpcodeDL, OpcodeLor:
		res := s1
		if toolProbabilisticEnabled() {
			rng := probabilisticRandomGenerator(t)
			if rng == nil {
				panic(NewNullPointerException())
			}
			index := int(rng.NextDouble() * float64(len(args)))
			stride := rng.NextPrime()
			for i := 0; i < len(args); i++ {
				next, err := t.GetNextStatesForPredicate(action, args[index], acts, c, s0, res, nss, cm)
				if err != nil {
					return res, err
				}
				res = next
				if nss.HasStates() {
					return res, nil
				}
				index = (index + stride) % len(args)
			}
			return res, nil
		}
		for _, arg := range args {
			next, err := t.GetNextStatesForPredicate(action, arg, acts, c, s0, res, nss, cm)
			if err != nil {
				return res, err
			}
			res = next
		}
		return res, nil
	case OpcodeBE:
		return t.nextBoundedExists(action, pred, acts, c, s0, s1, nss, cm)
	case OpcodeBF:
		return t.nextBoundedForall(action, pred, acts, c, s0, s1, nss, cm)
	case OpcodeAA:
		acts1 := acts.Cons(args[1], c, cm, ActionItemChanged)
		return t.GetNextStatesForPredicate(action, args[0], acts1, c, s0, s1, nss, cm)
	case OpcodeSA:
		res, err := t.GetNextStatesForPredicate(action, args[0], acts, c, s0, s1, nss, cm)
		if err != nil {
			return res, err
		}
		return t.ProcessUnchanged(action, args[1], acts, c, s0, res, nss, cm)
	case OpcodeITE:
		guard, err := t.Eval(args[0], c, s0, s1, EvalClear, cm)
		if err != nil {
			return s1, err
		}
		bguard, ok := guard.(*BoolValue)
		if !ok {
			return s1, NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("In computing next states, a non-boolean expression (%s) was used as the condition of an IF.%s", ValueKindString(guard), SemanticString(pred)), pred, c)
		}
		idx := 2
		if bguard.Val {
			idx = 1
		}
		return t.GetNextStatesForPredicate(action, args[idx], acts, c, s0, s1, nss, cm)
	case OpcodeCase:
		return t.nextCase(action, pred, acts, c, s0, s1, nss, cm)
	case OpcodeEq:
		return t.nextEquality(action, pred, args[0], args[1], acts, c, s0, s1, nss, cm)
	case OpcodeSubseteq:
		return t.nextSubsetEq(action, pred, args[0], args[1], acts, c, s0, s1, nss, cm)
	case OpcodeIn:
		return t.nextMembership(action, pred, args[0], args[1], acts, c, s0, s1, nss, cm)
	case OpcodeImplies:
		value, err := t.Eval(args[0], c, s0, s1, EvalClear, cm)
		if err != nil {
			return s1, err
		}
		bval, ok := value.(*BoolValue)
		if !ok {
			return s1, NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("In computing next states of a predicate of the form P => Q, P was\n%s.\n%s", valueKindString(value), SemanticString(pred)), pred, c)
		}
		if bval.Val {
			return t.GetNextStatesForPredicate(action, args[1], acts, c, s0, s1, nss, cm)
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	case OpcodeUnchanged:
		return t.ProcessUnchanged(action, args[0], acts, c, s0, s1, nss, cm)
	case OpcodeNop:
		return t.GetNextStatesForPredicate(action, args[0], acts, c, s0, s1, nss, cm)
	case OpcodeCdot:
		return t.nextActionComposition(action, pred, args, acts, c, s0, s1, nss, cm)
	case OpcodeFA:
		return t.nextFcnApply(action, pred, acts, c, s0, s1, nss, cm)
	default:
		bval, err := t.Eval(pred, c, s0, s1, EvalClear, cm)
		if err != nil {
			return s1, err
		}
		return t.continueNextIfBool(action, pred, bval, c, acts, s0, s1, nss, cm)
	}
}

func (t *Tool) initFcnApply(init *OpApplNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	if len(init.Args) < 2 {
		return newTLCError(ECGeneral, "malformed function application in initial-state predicate: %s", SemanticString(init))
	}
	fval, err := t.Eval(init.Args[0], c, ps, EmptyState, EvalInit, cm)
	if err != nil {
		return err
	}
	if fcn, ok := fval.(*FcnLambdaValue); ok && fcn.FcnRcd == nil {
		c1, err := t.getFcnContext(fcn, init, c, ps, EmptyState, EvalInit, cm)
		if err != nil {
			return err
		}
		return t.GetInitStatesForPredicate(fcn.ownedBody(), acts, c1, ps, states, cm)
	}
	bval, err := t.applyPredicateFunction("initial states", init, fval, c, ps, EmptyState, EvalInit, cm)
	if err != nil {
		return err
	}
	if bval.Val {
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	return nil
}

func (t *Tool) continueNextIfBool(action *Action, pred SemanticNode, value Value, c *Context, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	bval, ok := value.(*BoolValue)
	if !ok {
		// User-defined values supply Context.Empty; builtin evaluation retains c.
		return s1, newTLCDetailedRuntimeExceptionCode(ECTLCExpectedExpressionInComputing, pred, c, "next states", "boolean", value.String(), SemanticString(pred))
	}
	if bval.Val {
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	return s1, nil
}

func (t *Tool) actionCompositionIntermediateStates(action *Action, pred SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, cm CostModel) (*StateVec, *TLCStateMut, error) {
	intermediate := NewStateVec(0)
	collector := NewNextStateFunctor(func(predecessor *TLCStateMut, act *Action, state *TLCStateMut) (any, error) {
		intermediate.Add(state.SetPredecessor(predecessor).SetAction(act))
		return intermediate, nil
	})
	collector.HasStatesFunc = intermediate.HasStates
	res, err := t.GetNextStatesForPredicate(action, pred, acts, c, s0, s1, collector, cm)
	return intermediate, res, err
}

func probabilisticRandomGenerator(tool *Tool) *JavaRandom {
	if simulator := CurrentSimulator(); simulator != nil {
		return simulator.GetRNG()
	}
	if tool != nil && tool.ownedSpecProcessor() != nil {
		panic(NewNullPointerException())
	}
	return RandomEnumerableGenerator()
}

func (t *Tool) nextActionComposition(action *Action, pred SemanticNode, args []SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	if !actionCompositionEnabled() {
		return s1, NewTLCDetailedRuntimeException(ECGeneral, actionCompositionUnsupportedMessage, pred, c)
	}
	if len(args) < 2 {
		return s1, newTLCError(ECGeneral, "malformed action composition")
	}
	tState := s0.CopyWith(s1)
	intermediate, _, err := t.actionCompositionIntermediateStates(action, args[0], acts, c, s0, tState, cm)
	if err != nil {
		return s1, err
	}
	nss.IncrementStatesGenerated(int64(intermediate.Size()))
	for i := 0; i < intermediate.Size(); i++ {
		mid := intermediate.At(i)
		u := s1.Copy()
		wrapper := &NextStateFunctor{
			AddNextElementFunc: func(_ *TLCStateMut, _ *Action, succ *TLCStateMut) (any, error) {
				return nss.AddNextElement(s0, action, succ.SetPredecessor(s0))
			},
			AddUnsatisfiedNextStateFn: func(_ *TLCStateMut, _ *Action, succ *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
				return nss.AddUnsatisfiedNextState(s0, action, succ.SetPredecessor(s0), pred, con)
			},
			HaltFunc:       nss.Halt,
			ShouldHaltFunc: nss.ShouldHalt,
		}
		if _, err := t.GetNextStatesForPredicate(action, args[1], acts, c, mid, u, wrapper, cm); err != nil {
			return s1, err
		}
	}
	return s1, nil
}

func (t *Tool) nextFcnApply(action *Action, pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	if len(pred.Args) < 2 {
		return s1, newTLCError(ECGeneral, "malformed function application in next-state predicate: %s", SemanticString(pred))
	}
	fval, err := t.Eval(pred.Args[0], c, s0, s1, EvalKeepLazy, cm)
	if err != nil {
		return s1, err
	}
	if fcn, ok := fval.(*FcnLambdaValue); ok && fcn.FcnRcd == nil {
		c1, err := t.getFcnContext(fcn, pred, c, s0, s1, EvalClear, cm)
		if err != nil {
			return s1, err
		}
		return t.GetNextStatesForPredicate(action, fcn.ownedBody(), acts, c1, s0, s1, nss, fcn.CM)
	}
	bval, err := t.applyPredicateFunction("next states", pred, fval, c, s0, s1, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	if bval.Val {
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	return s1, nil
}

func (t *Tool) nextBoundedExists(action *Action, pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	var enum *ContextEnumerator
	var err error
	if toolProbabilisticEnabled() {
		enum, err = t.ContextsRandomized(pred, c, s0, s1, EvalClear, cm)
	} else {
		enum, err = t.Contexts(pred, c, s0, s1, EvalClear, cm)
	}
	if err != nil {
		return s1, err
	}
	res := s1
	for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
		if err := enum.Err(); err != nil {
			return res, err
		}
		next, err := t.GetNextStatesForPredicate(action, pred.Args[0], acts, c1, s0, res, nss, cm)
		if err != nil {
			return res, err
		}
		res = next
		if toolProbabilisticEnabled() && nss.HasStates() {
			return res, nil
		}
	}
	return res, enum.Err()
}

func (t *Tool) nextBoundedForall(action *Action, pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	enum, err := t.Contexts(pred, c, s0, s1, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	c1 := enum.NextElement()
	if err := enum.Err(); err != nil {
		return s1, err
	}
	if c1 == nil {
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	acts1 := acts
	for c2 := enum.NextElement(); c2 != nil; c2 = enum.NextElement() {
		if err := enum.Err(); err != nil {
			return s1, err
		}
		acts1 = acts1.Cons(pred.Args[0], c2, cm, ActionItemPred)
	}
	if err := enum.Err(); err != nil {
		return s1, err
	}
	return t.GetNextStatesForPredicate(action, pred.Args[0], acts1, c1, s0, s1, nss, cm)
}

func (t *Tool) nextCase(action *Action, pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	if toolProbabilisticEnabled() {
		return s1, NewUnsupportedOperationException("Probabilistic evaluation of next-state relation not implemented for CASE yet.")
	}
	var other SemanticNode
	for _, arg := range pred.Args {
		pair, ok := arg.(*OpApplNode)
		if !ok || len(pair.Args) < 2 {
			return s1, newTLCError(ECGeneral, "malformed CASE in next-state predicate")
		}
		if pair.Args[0] == nil {
			other = pair.Args[1]
			continue
		}
		armCM := cm
		if specCoverageEnabled() {
			armCM = cm.Get(arg)
		}
		value, err := t.Eval(pair.Args[0], c, s0, s1, EvalClear, armCM)
		if err != nil {
			return s1, err
		}
		bval, ok := value.(*BoolValue)
		if !ok {
			return s1, NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("In computing next states, a non-boolean expression (%s) was used as a guard condition of a CASE.\n%s", ValueKindString(value), SemanticString(pair.Args[1])), pair.Args[1], c)
		}
		if bval.Val {
			return t.GetNextStatesForPredicate(action, pair.Args[1], acts, c, s0, s1, nss, armCM)
		}
	}
	if other == nil {
		return s1, NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("In computing next states, TLC encountered a CASE with no conditions true.\n%s", SemanticString(pred)), pred, c)
	}
	otherCM := cm
	if specCoverageEnabled() && len(pred.Args) > 0 {
		otherCM = cm.Get(pred.Args[len(pred.Args)-1])
	}
	return t.GetNextStatesForPredicate(action, other, acts, c, s0, s1, nss, otherCM)
}

func (t *Tool) applyPredicateFunction(where string, expr *OpApplNode, fval Value, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (*BoolValue, error) {
	if fval == nil {
		return nil, newTLCError(ECGeneral, "In computing %s, a non-function (<nil>) was applied as a function.\n%s", where, SemanticString(expr))
	}
	switch f := fval.(type) {
	case *FcnLambdaValue:
		if f.FcnRcd != nil {
			fval = f.FcnRcd
		}
	case *FcnRcdValue, *TupleValue, *RecordValue, *CounterExample:
	default:
		return nil, NewTLCDetailedRuntimeException(ECGeneral, fmt.Sprintf("In computing %s, a non-function (%s) was applied as a function.\n%s", where, ValueKindString(fval), SemanticString(expr)), expr, c)
	}
	value, err := t.applyEvaluatedFunction(expr, fval, c, s0, s1, control, cm, false)
	if err != nil {
		return nil, err
	}
	bval, ok := value.(*BoolValue)
	if !ok {
		return nil, newTLCDetailedRuntimeExceptionCode(ECTLCExpectedExpressionInComputing2, expr.Args[1], c, where, "boolean", SemanticString(expr))
	}
	return bval, nil
}

func (t *Tool) nextEquality(action *Action, pred SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	varNode := t.GetPrimedVar(left, c, false)
	if varNode == nil {
		bval, err := t.evalBool(pred, c, s0, s1, EvalClear, cm)
		if err != nil || !bval.Val {
			return s1, err
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	lval := s1.Lookup(varNode.Name)
	rval, err := t.Eval(right, c, s0, s1, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	if lval == nil {
		resState := s1.Bind(varNode.Name, rval)
		res, err := t.GetNextStatesFromActionList(action, acts, s0, resState, nss, cm)
		if err == nil {
			res.Unbind(varNode.Name)
		}
		return res, err
	}
	eq, err := lval.Equal(rval)
	if err != nil || !eq {
		return s1, err
	}
	return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
}

func (t *Tool) nextMembership(action *Action, pred SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	varNode := t.GetPrimedVar(left, c, false)
	if varNode == nil {
		bval, err := t.evalBool(pred, c, s0, s1, EvalClear, cm)
		if err != nil || !bval.Val {
			return s1, err
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	lval := s1.Lookup(varNode.Name)
	rval, err := t.Eval(right, c, s0, s1, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	return t.enumerateNextAssignment(action, varNode.Name, lval, rval, pred, c, acts, s0, s1, nss, cm)
}

func (t *Tool) nextSubsetEq(action *Action, pred SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	varNode := t.GetPrimedVar(left, c, false)
	if varNode == nil {
		bval, err := t.evalBool(pred, c, s0, s1, EvalClear, cm)
		if err != nil || !bval.Val {
			return s1, err
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	lval := s1.Lookup(varNode.Name)
	rset, err := t.Eval(right, c, s0, s1, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	return t.enumerateNextAssignment(action, varNode.Name, lval, t.SetSource(right, NewSubsetValue(rset, cm)), pred, c, acts, s0, s1, nss, cm)
}

func (t *Tool) enumerateNextAssignment(action *Action, varName *UniqueString, lval Value, domain Value, pred SemanticNode, c *Context, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	if lval != nil {
		member, err := domain.Member(lval)
		if err != nil || !member {
			return s1, err
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	enumerable, ok := asEnumerable(domain)
	if !ok {
		return s1, NewTLCDetailedRuntimeException(ECGeneral, "In computing next states, the right side of \\IN is not enumerable.\n"+SemanticString(pred), pred, c)
	}
	res := s1
	if toolProbabilisticEnabled() {
		enum, err := randomizedValueEnumeration(enumerable)
		if err != nil {
			return res, err
		}
		for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
			if err := enum.Err(); err != nil {
				return res, err
			}
			resState := res.Bind(varName, elem)
			next, err := t.GetNextStatesFromActionList(action, acts, s0, resState, nss, cm)
			if err == nil {
				next.Unbind(varName)
			}
			if err != nil {
				return res, err
			}
			res = next
			if nss.HasStates() {
				return res, nil
			}
		}
		if err := enum.Err(); err != nil {
			return res, err
		}
	}
	enum := enumerable.Elements()
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		if err := enum.Err(); err != nil {
			return res, err
		}
		resState := res.Bind(varName, elem)
		next, err := t.GetNextStatesFromActionList(action, acts, s0, resState, nss, cm)
		if err == nil {
			next.Unbind(varName)
		}
		if err != nil {
			return res, err
		}
		res = next
	}
	return res, enum.Err()
}

func (t *Tool) ProcessUnchanged(action *Action, expr SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (state *TLCStateMut, err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	done := t.callStackEnter(expr)
	defer func() { done(err) }()
	if debugToolActive(t) {
		debugger := t.ownedDebugger()
		if t.DebugEvalMode == DebugEvalDebugger {
			return t.NoDebug().ProcessUnchanged(action, expr, acts, c, s0, s1, nss, cm)
		}
		defer func() {
			if !debugErrorKnown(err) {
				if IsInvariantViolatedException(err) {
					debugger.MarkInvariantViolatedFrame(t, expr, c, s0, action, s1, err)
					debugger.PopExceptionFrame(t, expr, c, nil, err)
				} else if isJavaEvalOrRuntimeException(err) {
					debugger.PushActionExceptionFrame(t, expr, c, s0, action, s1, err)
					debugger.PopExceptionFrame(t, expr, c, nil, err)
				}
			}
		}()
	}
	return t.ProcessUnchangedImpl(action, expr, acts, c, s0, s1, nss, cm)
}

func (t *Tool) ProcessUnchangedImpl(action *Action, expr SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (state *TLCStateMut, err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if specCoverageEnabled() {
		cm = cm.Get(expr)
	}
	if varNode := t.GetVar(expr, c, false); varNode != nil {
		varName := varNode.Name
		val0 := s0.Lookup(varName)
		val1 := s1.Lookup(varName)
		if val1 == nil {
			resState := s1.Bind(varName, val0)
			res, err := t.GetNextStatesFromActionList(action, acts, s0, resState, nss, cm)
			if err == nil {
				res.Unbind(varName)
			}
			return res, err
		}
		eq, err := val0.Equal(val1)
		if err != nil {
			return s1, err
		}
		if !eq {
			PrintWarning(ECTLCUnchangedVariableChanged, varName.String(), SemanticString(expr))
			return s1, nil
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	if appl, ok := expr.(*OpApplNode); ok {
		opName := appl.Operator.Name
		opcode := GetOpCode(opName)
		if opcode == OpcodeTup {
			nestedCM := cm
			if specCoverageEnabled() {
				nestedCM = cm.Get(appl)
			}
			acts1 := acts
			for i := len(appl.Args) - 1; i > 0; i-- {
				acts1 = acts1.Cons(appl.Args[i], c, nestedCM, ActionItemUnchanged)
			}
			if len(appl.Args) == 0 {
				return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
			}
			return t.ProcessUnchanged(action, appl.Args[0], acts1, c, s0, s1, nss, nestedCM)
		}
		if opcode == 0 && len(appl.Args) == 0 {
			val := t.Lookup(appl.Operator, c, EmptyState, false)
			switch v := val.(type) {
			case *OpDefNode:
				return t.ProcessUnchanged(action, v.GetBody(), acts, c, s0, s1, nss, cm)
			case *LazyValue:
				return t.ProcessUnchanged(action, v.ownedExpr(), acts, v.ownedContext(), s0, s1, nss, cm)
			case *LazySupplierValue:
				lv := asLazyValue(v)
				return t.ProcessUnchanged(action, lv.ownedExpr(), acts, lv.ownedContext(), s0, s1, nss, cm)
			}
		}
	}
	v0, err := t.Eval(expr, c, s0, EmptyState, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	v1, err := t.Eval(expr, c, s1, nil, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	eq, err := v0.Equal(v1)
	if err != nil || !eq {
		return s1, err
	}
	return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
}

func (t *Tool) GetVar(expr SemanticNode, c *Context, cutoff bool) *SymbolNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	switch expr := expr.(type) {
	case *SubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.GetOp(), t.GetVal(sub.GetExpr(), c, false, DoNotRecordCostModel))
		}
		return t.GetVar(expr.GetBody(), c1, cutoff)
	case *APSubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.GetOp(), t.GetVal(sub.GetExpr(), c, false, DoNotRecordCostModel))
		}
		return t.GetVar(expr.GetBody(), c1, cutoff)
	case *LetInNode:
		return t.GetVar(expr.Body, letDefinitionsContext(c, expr.Lets, expr.Bindings...), cutoff)
	case *LabelNode:
		return t.GetVar(expr.Body, c, cutoff)
	case *OpApplNode:
		if expr.Operator == nil || expr.Operator.Name == nil || expr.Operator.GetArity() != 0 {
			return nil
		}
		isVarDecl := expr.Operator.IsVariableDecl()
		val := t.LookupWithCutoff(expr.Operator, c, cutoff && isVarDecl, EmptyState, false)
		switch v := val.(type) {
		case *LazyValue:
			return t.GetVar(v.ownedExpr(), v.ownedContext(), cutoff)
		case *LazySupplierValue:
			lv := asLazyValue(v)
			return t.GetVar(lv.ownedExpr(), lv.ownedContext(), cutoff)
		case *OpDefNode:
			return t.GetVar(v.GetBody(), c, cutoff)
		default:
			if isVarDecl {
				return expr.Operator
			}
			return nil
		}
	default:
		return nil
	}
}

func (t *Tool) GetPrimedVar(expr SemanticNode, c *Context, cutoff bool) *SymbolNode {
	if t == nil {
		panic(NewNullPointerException())
	}
	switch expr := expr.(type) {
	case *SubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.GetOp(), t.GetVal(sub.GetExpr(), c, false, DoNotRecordCostModel))
		}
		return t.GetPrimedVar(expr.GetBody(), c1, cutoff)
	case *APSubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.GetOp(), t.GetVal(sub.GetExpr(), c, false, DoNotRecordCostModel))
		}
		return t.GetPrimedVar(expr.GetBody(), c1, cutoff)
	case *LetInNode:
		return t.GetPrimedVar(expr.Body, letDefinitionsContext(c, expr.Lets, expr.Bindings...), cutoff)
	case *LabelNode:
		return t.GetPrimedVar(expr.Body, c, cutoff)
	case *OpApplNode:
		if expr.Operator == nil || expr.Operator.Name == nil {
			return nil
		}
		if GetOpCode(expr.Operator.Name) == OpcodePrime && len(expr.Args) == 1 {
			return t.GetVar(expr.Args[0], c, cutoff)
		}
		if expr.Operator.GetArity() != 0 {
			return nil
		}
		isVarDecl := expr.Operator.IsVariableDecl()
		val := t.LookupWithCutoff(expr.Operator, c, cutoff && isVarDecl, EmptyState, false)
		switch v := val.(type) {
		case *LazyValue:
			return t.GetPrimedVar(v.ownedExpr(), v.ownedContext(), cutoff)
		case *LazySupplierValue:
			lv := asLazyValue(v)
			return t.GetPrimedVar(lv.ownedExpr(), lv.ownedContext(), cutoff)
		case *OpDefNode:
			return t.GetPrimedVar(v.GetBody(), c, cutoff)
		default:
			return nil
		}
	default:
		return nil
	}
}
