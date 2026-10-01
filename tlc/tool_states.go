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
	init := t.GetInitStateSpec()
	acts := EmptyActionItemList
	for i := len(init) - 1; i > 0; i-- {
		elem := init[i]
		acts = acts.ConsAction(elem, ActionItemPred)
	}
	if len(init) == 0 {
		return nil
	}
	elem := init[0]
	ps := NewEmptyState()
	if acts.IsEmpty() {
		acts = &ActionItemList{Next: EmptyActionItemList, act: elem}
	}
	return t.GetInitStatesForPredicate(elem.Pred, acts, elem.Con, ps, functor, elem.CM)
}

func (t *Tool) MakeStateImpl(pred SemanticNode) (*TLCStateMut, error) {
	vec := NewStateVec(0)
	functor := NewStateFunctor(func(state *TLCStateMut) (any, error) {
		vec.Add(state)
		return nil, nil
	})
	if err := t.GetInitStatesForPredicate(pred, EmptyActionItemList, EmptyContext, NewEmptyState(), functor, CostModel{}); err != nil {
		return nil, err
	}
	if vec.Size() != 1 {
		return nil, newTLCError(ECGeneral, "the predicate does not specify a unique state: %s", SemanticString(pred))
	}
	state := vec.At(0)
	if !t.IsGoodState(state) {
		return nil, newTLCError(ECGeneral, "the state specified by the predicate is not complete: %s", SemanticString(pred))
	}
	return state, nil
}

func (t *Tool) GetInitStatesForPredicate(init SemanticNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) (err error) {
	done := t.callStackEnter(init)
	defer func() { done(err) }()
	if c == nil {
		c = EmptyContext
	}
	if acts == nil {
		acts = EmptyActionItemList
	}
	if ps == nil {
		ps = NewEmptyState()
	}
	switch init := init.(type) {
	case *OpApplNode:
		return t.GetInitStatesAppl(init, acts, c, ps, states, cm)
	case *LetInNode:
		return t.GetInitStatesForPredicate(init.Body, acts, letDefinitionsContextWithCostModel(c, init.Lets, cm, init.Bindings...), ps, states, cm)
	case *SubstInNode:
		c1 := c
		for _, sub := range init.Substs {
			subCM := cm
			if CoverageEnabled() {
				subCM = cm.GetSubst(sub)
			}
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, subCM))
		}
		return t.GetInitStatesForPredicate(init.Body, acts, c1, ps, states, cm)
	case *APSubstInNode:
		c1 := c
		for _, sub := range init.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, cm))
		}
		return t.GetInitStatesForPredicate(init.Body, acts, c1, ps, states, cm)
	case *LabelNode:
		return t.GetInitStatesForPredicate(init.Body, acts, c, ps, states, cm)
	default:
		return newTLCError(ECGeneral, "the init state relation is not a boolean expression: %s", SemanticString(init))
	}
}

func (t *Tool) GetInitStatesFromActionList(acts *ActionItemList, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	rootAction := UnknownAction
	if acts != nil && acts.GetAction() != nil {
		rootAction = acts.GetAction()
	}
	if acts == nil || acts.IsEmpty() {
		if CoverageActionEnabled() {
			cm.IncInvocations()
			cm.GetRoot().IncInvocations()
		}
		_, err := states.AddElement(ps.Copy().SetAction(rootAction))
		return err
	}
	if ps.AllAssigned() {
		for !acts.IsEmpty() {
			bval, err := t.evalBool(acts.CarPred(), acts.CarContext(), ps, EmptyState, EvalInit, acts.CM, "initial states")
			if err != nil {
				return err
			}
			if !bval.Val {
				if CoverageActionEnabled() {
					cm.GetRoot().IncSecondary()
				}
				states.AddUnsatisfiedState(ps, acts.CarPred(), acts.CarContext())
				return nil
			}
			acts = acts.Cdr()
		}
		if CoverageActionEnabled() {
			cm.IncInvocations()
			cm.GetRoot().IncInvocations()
		}
		_, err := states.AddElement(ps.Copy().SetAction(rootAction))
		return err
	}
	acts1 := acts.Cdr()
	return t.GetInitStatesForPredicate(acts.CarPred(), acts1, acts.CarContext(), ps, states, acts.CM)
}

func (t *Tool) GetInitStatesAppl(init *OpApplNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) (err error) {
	done := t.callStackEnter(init)
	defer func() { done(err) }()
	if CoverageEnabled() {
		cm = cm.Get(init)
	}
	args := init.Args
	opNode := init.Operator
	opcode := GetOpCode(opNode.Name)
	if opcode == 0 {
		val := t.Lookup(opNode, c, ps, false)
		if lv, ok := val.(*LazyValue); ok {
			cached := lv.GetCachedValue(t, ps, nil, EvalClear)
			if cached == nil {
				return t.GetInitStatesForPredicate(lv.Expr, acts, lv.Con, ps, states, cm)
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
				return t.GetInitStatesForPredicate(v.Body, acts, c1, ps, states, cm)
			}
		case *ThmOrAssumpDefNode:
			c1, err := t.GetThmOrAssumpContext(v, args, c, true)
			if err != nil {
				return err
			}
			return t.GetInitStatesForPredicate(v.Body, acts, c1, ps, states, cm)
		case Value:
			if isOperatorValue(v) {
				bval, err := EvalOperatorValueWithTool(v, t, args, c, ps, EmptyState, EvalInit, cm)
				if err != nil {
					return err
				}
				return t.continueInitIfBool(init, bval, acts, ps, states, cm)
			}
			return t.continueInitIfBool(init, v, acts, ps, states, cm)
		default:
			if val == nil {
				return newTLCError(ECGeneral, "undefined operator in initial predicate: %s", opNode)
			}
			return newTLCError(ECGeneral, "initial predicate operator %s evaluated to %T", opNode, val)
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
		if len(args) == 0 {
			return t.GetInitStatesFromActionList(acts, ps, states, cm)
		}
		acts1 := acts
		for i := len(args) - 1; i > 0; i-- {
			acts1 = acts1.Cons(args[i], c, cm, i)
		}
		return t.GetInitStatesForPredicate(args[0], acts1, c, ps, states, cm)
	case OpcodeBE:
		enum, err := t.Contexts(init, c, ps, EmptyState, EvalInit, cm)
		if err != nil {
			return err
		}
		for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
			if err := t.GetInitStatesForPredicate(args[0], acts, c1, ps, states, cm); err != nil {
				return err
			}
		}
		return enum.Err()
	case OpcodeBF:
		return t.initBoundedForall(init, acts, c, ps, states, cm)
	case OpcodeITE:
		guard, err := t.evalBool(args[0], c, ps, EmptyState, EvalInit, cm, "initial IF")
		if err != nil {
			return err
		}
		idx := 2
		if guard.Val {
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
		lval, err := t.evalBool(args[0], c, ps, EmptyState, EvalInit, cm, "initial implication")
		if err != nil {
			return err
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
		return t.continueInitIfBool(init, bval, acts, ps, states, cm)
	}
}

func (t *Tool) continueInitIfBool(init SemanticNode, value Value, acts *ActionItemList, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	bval, ok := value.(*BoolValue)
	if !ok {
		return newTLCError(ECGeneral, "in computing initial states, expected boolean but found %s in %s", value, SemanticString(init))
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
	if c1 == nil {
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	acts1 := acts
	for c2 := enum.NextElement(); c2 != nil; c2 = enum.NextElement() {
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
		bval, err := t.evalBool(pair.Args[0], c, ps, EmptyState, EvalInit, cm, "initial CASE")
		if err != nil {
			return err
		}
		if bval.Val {
			return t.GetInitStatesForPredicate(pair.Args[1], acts, c, ps, states, cm)
		}
	}
	if other == nil {
		return newTLCError(ECGeneral, "CASE has no true condition in initial predicate: %s", SemanticString(init))
	}
	return t.GetInitStatesForPredicate(other, acts, c, ps, states, cm)
}

func (t *Tool) initEquality(init SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	varNode := t.GetVar(left, c, false)
	if varNode == nil || varNode.Name.VarLoc() < 0 {
		bval, err := t.evalBool(init, c, ps, EmptyState, EvalInit, cm, "initial equality")
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
		ps.Unbind(varName)
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
		bval, err := t.evalBool(init, c, ps, EmptyState, EvalInit, cm, "initial membership")
		if err != nil || !bval.Val {
			return err
		}
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	rval, err := t.Eval(right, c, ps, EmptyState, EvalInit, cm)
	if err != nil {
		return err
	}
	return t.enumerateInitAssignment(varNode.Name, rval, acts, ps, states, cm)
}

func (t *Tool) initSubsetEq(init SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	varNode := t.GetVar(left, c, false)
	if varNode == nil || varNode.Name.VarLoc() < 0 {
		bval, err := t.evalBool(init, c, ps, EmptyState, EvalInit, cm, "initial subset")
		if err != nil || !bval.Val {
			return err
		}
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	rset, err := t.Eval(right, c, ps, EmptyState, EvalInit, cm)
	if err != nil {
		return err
	}
	return t.enumerateInitAssignment(varNode.Name, NewSubsetValue(rset), acts, ps, states, cm)
}

func (t *Tool) enumerateInitAssignment(varName *UniqueString, domain Value, acts *ActionItemList, ps *TLCStateMut, states *StateFunctor, cm CostModel) error {
	lval := ps.Lookup(varName)
	if lval != nil {
		member, err := domain.Member(lval)
		if err != nil || !member {
			return err
		}
		return t.GetInitStatesFromActionList(acts, ps, states, cm)
	}
	enumerable, ok := asEnumerable(domain)
	if !ok {
		return newTLCError(ECGeneral, "right side of \\in is not enumerable while assigning %s", varName)
	}
	enum := enumerable.Elements()
	for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
		ps.Bind(varName, elem)
		if err := t.GetInitStatesFromActionList(acts, ps, states, cm); err != nil {
			ps.Unbind(varName)
			return err
		}
		ps.Unbind(varName)
	}
	return enum.Err()
}

func (t *Tool) GetNextStatesImpl(action *Action, state *TLCStateMut) (*StateVec, error) {
	nss := NewStateVec(0)
	functor := NewNextStateFunctor(func(_ *TLCStateMut, _ *Action, succ *TLCStateMut) (any, error) {
		nss.Add(succ)
		return nil, nil
	})
	functor.HasStatesFunc = nss.HasStates
	s1 := NewEmptyState().SetPredecessor(state).SetAction(action)
	_, err := t.GetNextStatesForPredicate(action, action.Pred, EmptyActionItemList, action.Con, state, s1, functor, action.CM)
	if err == nil && action != nil && CoverageActionEnabled() {
		action.CM.IncInvocations(int64(nss.Size()))
	}
	if err == nil && toolProbabilisticEnabled() && nss.Size() > 1 {
		fmt.Fprintln(os.Stderr, "Simulator generated more than one next state")
	}
	return nss, err
}

func (t *Tool) GetNextStatesForPredicate(action *Action, pred SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (state *TLCStateMut, err error) {
	done := t.callStackEnter(pred)
	defer func() { done(err) }()
	if c == nil {
		c = EmptyContext
	}
	if acts == nil {
		acts = EmptyActionItemList
	}
	switch pred := pred.(type) {
	case *OpApplNode:
		if CoverageEnabled() {
			cm = cm.Get(pred)
		}
		return t.GetNextStatesAppl(action, pred, acts, c, s0, s1, nss, cm)
	case *LetInNode:
		return t.GetNextStatesForPredicate(action, pred.Body, acts, letDefinitionsContextWithCostModel(c, pred.Lets, cm, pred.Bindings...), s0, s1, nss, cm)
	case *SubstInNode:
		c1 := c
		for _, sub := range pred.Substs {
			subCM := cm
			if CoverageEnabled() {
				subCM = cm.GetSubst(sub)
			}
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, subCM))
		}
		return t.GetNextStatesForPredicate(action, pred.Body, acts, c1, s0, s1, nss, cm)
	case *APSubstInNode:
		c1 := c
		for _, sub := range pred.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, cm))
		}
		return t.GetNextStatesForPredicate(action, pred.Body, acts, c1, s0, s1, nss, cm)
	case *LabelNode:
		return t.GetNextStatesForPredicate(action, pred.Body, acts, c, s0, s1, nss, cm)
	default:
		return s1, newTLCError(ECGeneral, "the next-state relation is not a boolean expression: %s", SemanticString(pred))
	}
}

func (t *Tool) GetNextStatesFromActionList(action *Action, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	copyState, err := t.getNextStates0(action, acts, s0, s1, nss, cm)
	if err != nil {
		return copyState, err
	}
	if CoverageEnabled() && copyState != s1 {
		cm.IncInvocations()
	}
	return copyState, nil
}

func (t *Tool) getNextStates0(action *Action, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	if acts == nil || acts.IsEmpty() {
		if _, err := nss.AddNextElement(s0, action, s1); err != nil {
			return s1, err
		}
		return s1.Copy(), nil
	}
	if Globals.Warn && s1 != nil && s1.AllAssigned() {
		return t.getNextStatesAllAssigned(action, acts, s0, s1, nss, cm)
	}
	kind := acts.CarKind()
	pred := acts.CarPred()
	c := acts.CarContext()
	acts1 := acts.Cdr()
	switch {
	case kind > ActionItemConjunct || kind == ActionItemPred:
		return t.GetNextStatesForPredicate(action, pred, acts1, c, s0, s1, nss, acts.CM)
	case kind == ActionItemUnchanged:
		return t.ProcessUnchanged(action, pred, acts1, c, s0, s1, nss, acts.CM)
	case kind == ActionItemChanged:
		v1, err := t.Eval(pred, c, s0, EmptyState, EvalClear, acts.CM)
		if err != nil {
			return s1, err
		}
		v2, err := t.Eval(pred, c, s1, EmptyState, EvalClear, acts.CM)
		if err != nil {
			return s1, err
		}
		eq, err := v1.Equal(v2)
		if err != nil || eq {
			return s1, err
		}
		if CoverageEnabled() {
			return t.GetNextStatesFromActionList(action, acts1, s0, s1, nss, acts.CM)
		}
		return t.getNextStates0(action, acts1, s0, s1, nss, acts.CM)
	default:
		return t.GetNextStatesForPredicate(action, pred, acts1, c, s0, s1, nss, acts.CM)
	}
}

func (t *Tool) getNextStatesAllAssigned(action *Action, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	for !acts.IsEmpty() {
		kind := acts.CarKind()
		pred := acts.CarPred()
		c := acts.CarContext()
		cm2 := acts.CM
		switch {
		case kind > ActionItemConjunct || kind == ActionItemPred:
			bval, err := t.evalBool(pred, c, s0, s1, EvalClear, cm2, "next states")
			if err != nil {
				return s1, err
			}
			if !bval.Val {
				return nss.AddUnsatisfiedNextState(s0, action, s1, pred, c), nil
			}
		case kind == ActionItemUnchanged:
			return t.ProcessUnchanged(action, pred, acts.Cdr(), c, s0, s1, nss, cm2)
		case kind == ActionItemChanged:
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
		default:
			bval, err := t.evalBool(pred, c, s0, s1, EvalClear, cm2, "next states")
			if err != nil {
				return s1, err
			}
			if !bval.Val {
				return nss.AddUnsatisfiedNextState(s0, action, s1, pred, c), nil
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
	done := t.callStackEnter(pred)
	defer func() { done(err) }()
	args := pred.Args
	opNode := pred.Operator
	opcode := GetOpCode(opNode.Name)
	if opcode == 0 {
		val := t.Lookup(opNode, c, s0, false)
		if lv, ok := val.(*LazyValue); ok {
			cached := lv.GetCachedValue(t, s0, s1, EvalClear)
			if cached == nil {
				return t.GetNextStatesForPredicate(action, lv.Expr, acts, lv.Con, s0, s1, nss, lv.CM)
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
				return t.GetNextStatesForPredicate(action, v.Body, acts, c1, s0, s1, nss, cm)
			}
		case *ThmOrAssumpDefNode:
			c1, err := t.GetThmOrAssumpContext(v, args, c, true)
			if err != nil {
				return s1, err
			}
			return t.GetNextStatesForPredicate(action, v.Body, acts, c1, s0, s1, nss, cm)
		case Value:
			if isOperatorValue(v) {
				bval, err := EvalOperatorValueWithTool(v, t, args, c, s0, s1, EvalClear, cm)
				if err != nil {
					return s1, err
				}
				return t.continueNextIfBool(action, pred, bval, acts, s0, s1, nss, cm)
			}
			return t.continueNextIfBool(action, pred, v, acts, s0, s1, nss, cm)
		default:
			if val == nil {
				return s1, newTLCError(ECGeneral, "undefined operator in next-state predicate: %s", opNode)
			}
			return s1, newTLCError(ECGeneral, "next-state predicate operator %s evaluated to %T", opNode, val)
		}
	}
	switch opcode {
	case OpcodeCL, OpcodeLand:
		if len(args) == 0 {
			return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
		}
		acts1 := acts
		for i := len(args) - 1; i > 0; i-- {
			acts1 = acts1.Cons(args[i], c, cm, i)
		}
		return t.GetNextStatesForPredicate(action, args[0], acts1, c, s0, s1, nss, cm)
	case OpcodeDL, OpcodeLor:
		if len(args) == 0 {
			return s1, nil
		}
		res := s1
		if toolProbabilisticEnabled() {
			rng := probabilisticRandomGenerator()
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
		guard, err := t.evalBool(args[0], c, s0, s1, EvalClear, cm, "next IF")
		if err != nil {
			return s1, err
		}
		idx := 2
		if guard.Val {
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
		bval, err := t.evalBool(args[0], c, s0, s1, EvalClear, cm, "next implication")
		if err != nil {
			return s1, err
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
		return t.nextActionComposition(action, args, acts, c, s0, s1, nss, cm)
	case OpcodeFA:
		return t.nextFcnApply(action, pred, acts, c, s0, s1, nss, cm)
	default:
		bval, err := t.Eval(pred, c, s0, s1, EvalClear, cm)
		if err != nil {
			return s1, err
		}
		return t.continueNextIfBool(action, pred, bval, acts, s0, s1, nss, cm)
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
		return t.GetInitStatesForPredicate(fcn.Body, acts, c1, ps, states, cm)
	}
	bval, err := t.applyEvaluatedFunction(init, fval, c, ps, EmptyState, EvalInit, cm, false)
	if err != nil {
		return err
	}
	return t.continueInitIfBool(init, bval, acts, ps, states, cm)
}

func (t *Tool) continueNextIfBool(action *Action, pred SemanticNode, value Value, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	bval, ok := value.(*BoolValue)
	if !ok {
		return s1, newTLCError(ECGeneral, "in computing next states, expected boolean but found %s in %s", value, SemanticString(pred))
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

func probabilisticRandomGenerator() *JavaRandom {
	if simulator := CurrentSimulator(); simulator != nil && simulator.Rand != nil {
		return simulator.Rand
	}
	return RandomEnumerableGenerator()
}

func (t *Tool) nextActionComposition(action *Action, args []SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	if !actionCompositionEnabled() {
		return s1, newTLCError(ECGeneral, actionCompositionUnsupportedMessage)
	}
	if len(args) < 2 {
		return s1, newTLCError(ECGeneral, "malformed action composition")
	}
	tState := s0.CopyWith(s1)
	intermediate, res, err := t.actionCompositionIntermediateStates(action, args[0], acts, c, s0, tState, cm)
	if err != nil {
		return res, err
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
			return res, err
		}
	}
	return res, nil
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
		return t.GetNextStatesForPredicate(action, fcn.Body, acts, c1, s0, s1, nss, fcn.CM)
	}
	bval, err := t.applyEvaluatedFunction(pred, fval, c, s0, s1, EvalClear, cm, false)
	if err != nil {
		return s1, err
	}
	return t.continueNextIfBool(action, pred, bval, acts, s0, s1, nss, cm)
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
	if c1 == nil {
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	acts1 := acts
	for c2 := enum.NextElement(); c2 != nil; c2 = enum.NextElement() {
		acts1 = acts1.Cons(pred.Args[0], c2, cm, ActionItemPred)
	}
	if err := enum.Err(); err != nil {
		return s1, err
	}
	return t.GetNextStatesForPredicate(action, pred.Args[0], acts1, c1, s0, s1, nss, cm)
}

func (t *Tool) nextCase(action *Action, pred *OpApplNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	if toolProbabilisticEnabled() {
		return s1, newTLCError(ECGeneral, "Probabilistic evaluation of next-state relation not implemented for CASE yet.")
	}
	var other SemanticNode
	otherCM := cm
	for _, arg := range pred.Args {
		armCM := cm
		if CoverageEnabled() {
			armCM = cm.Get(arg)
		}
		pair, ok := arg.(*OpApplNode)
		if !ok || len(pair.Args) < 2 {
			return s1, newTLCError(ECGeneral, "malformed CASE in next-state predicate")
		}
		if pair.Args[0] == nil {
			other = pair.Args[1]
			otherCM = armCM
			continue
		}
		bval, err := t.evalBool(pair.Args[0], c, s0, s1, EvalClear, armCM, "next CASE")
		if err != nil {
			return s1, err
		}
		if bval.Val {
			return t.GetNextStatesForPredicate(action, pair.Args[1], acts, c, s0, s1, nss, armCM)
		}
	}
	if other == nil {
		return s1, newTLCError(ECGeneral, "CASE has no true condition in next-state predicate: %s", SemanticString(pred))
	}
	return t.GetNextStatesForPredicate(action, other, acts, c, s0, s1, nss, otherCM)
}

func (t *Tool) nextEquality(action *Action, pred SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	varNode := t.GetPrimedVar(left, c, false)
	if varNode == nil {
		bval, err := t.evalBool(pred, c, s0, s1, EvalClear, cm, "next equality")
		if err != nil || !bval.Val {
			return s1, err
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	rval, err := t.Eval(right, c, s0, s1, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	lval := s1.Lookup(varNode.Name)
	if lval == nil {
		s1.Bind(varNode.Name, rval)
		res, err := t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
		s1.Unbind(varNode.Name)
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
		bval, err := t.evalBool(pred, c, s0, s1, EvalClear, cm, "next membership")
		if err != nil || !bval.Val {
			return s1, err
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	rval, err := t.Eval(right, c, s0, s1, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	return t.enumerateNextAssignment(action, varNode.Name, rval, acts, s0, s1, nss, cm)
}

func (t *Tool) nextSubsetEq(action *Action, pred SemanticNode, left SemanticNode, right SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	varNode := t.GetPrimedVar(left, c, false)
	if varNode == nil {
		bval, err := t.evalBool(pred, c, s0, s1, EvalClear, cm, "next subset")
		if err != nil || !bval.Val {
			return s1, err
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	rset, err := t.Eval(right, c, s0, s1, EvalClear, cm)
	if err != nil {
		return s1, err
	}
	return t.enumerateNextAssignment(action, varNode.Name, NewSubsetValue(rset), acts, s0, s1, nss, cm)
}

func (t *Tool) enumerateNextAssignment(action *Action, varName *UniqueString, domain Value, acts *ActionItemList, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (*TLCStateMut, error) {
	lval := s1.Lookup(varName)
	if lval != nil {
		member, err := domain.Member(lval)
		if err != nil || !member {
			return s1, err
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	enumerable, ok := asEnumerable(domain)
	if !ok {
		return s1, newTLCError(ECGeneral, "right side of \\in is not enumerable while assigning %s'", varName)
	}
	res := s1
	if toolProbabilisticEnabled() {
		enum, err := randomizedValueEnumeration(enumerable)
		if err != nil {
			return res, err
		}
		for elem := enum.NextElement(); elem != nil; elem = enum.NextElement() {
			res.Bind(varName, elem)
			next, err := t.GetNextStatesFromActionList(action, acts, s0, res, nss, cm)
			res.Unbind(varName)
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
		res.Bind(varName, elem)
		next, err := t.GetNextStatesFromActionList(action, acts, s0, res, nss, cm)
		res.Unbind(varName)
		if err != nil {
			return res, err
		}
		res = next
	}
	return res, enum.Err()
}

func (t *Tool) ProcessUnchanged(action *Action, expr SemanticNode, acts *ActionItemList, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, nss *NextStateFunctor, cm CostModel) (state *TLCStateMut, err error) {
	done := t.callStackEnter(expr)
	defer func() { done(err) }()
	if CoverageEnabled() {
		cm = cm.Get(expr)
	}
	if varNode := t.GetVar(expr, c, false); varNode != nil {
		varName := varNode.Name
		val0 := s0.Lookup(varName)
		val1 := s1.Lookup(varName)
		if val1 == nil {
			s1.Bind(varName, val0)
			res, err := t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
			s1.Unbind(varName)
			return res, err
		}
		eq, err := val0.Equal(val1)
		if err != nil || !eq {
			return s1, err
		}
		return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
	}
	if appl, ok := expr.(*OpApplNode); ok && GetOpCode(appl.Operator.Name) == OpcodeTup {
		acts1 := acts
		for i := len(appl.Args) - 1; i > 0; i-- {
			acts1 = acts1.Cons(appl.Args[i], c, cm, ActionItemUnchanged)
		}
		if len(appl.Args) == 0 {
			return t.GetNextStatesFromActionList(action, acts, s0, s1, nss, cm)
		}
		return t.ProcessUnchanged(action, appl.Args[0], acts1, c, s0, s1, nss, cm)
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
	if c == nil {
		c = EmptyContext
	}
	switch expr := expr.(type) {
	case *SubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, DoNotRecordCostModel))
		}
		return t.GetVar(expr.Body, c1, cutoff)
	case *APSubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, DoNotRecordCostModel))
		}
		return t.GetVar(expr.Body, c1, cutoff)
	case *LetInNode:
		return t.GetVar(expr.Body, letDefinitionsContext(c, expr.Lets, expr.Bindings...), cutoff)
	case *LabelNode:
		return t.GetVar(expr.Body, c, cutoff)
	case *OpApplNode:
		if len(expr.Args) != 0 || expr.Operator == nil || expr.Operator.Name == nil {
			return nil
		}
		if GetOpCode(expr.Operator.Name) != 0 {
			return nil
		}
		isVarDecl := expr.Operator.Name.VarLoc() >= 0
		val := t.LookupWithCutoff(expr.Operator, c, cutoff && isVarDecl, EmptyState, false)
		switch v := val.(type) {
		case *LazyValue:
			return t.GetVar(v.Expr, v.Con, cutoff)
		case *OpDefNode:
			return t.GetVar(v.Body, c, cutoff)
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
	if c == nil {
		c = EmptyContext
	}
	switch expr := expr.(type) {
	case *SubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, DoNotRecordCostModel))
		}
		return t.GetPrimedVar(expr.Body, c1, cutoff)
	case *APSubstInNode:
		c1 := c
		for _, sub := range expr.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, false, DoNotRecordCostModel))
		}
		return t.GetPrimedVar(expr.Body, c1, cutoff)
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
		if len(expr.Args) != 0 {
			return nil
		}
		isVarDecl := expr.Operator.Name.VarLoc() >= 0
		val := t.LookupWithCutoff(expr.Operator, c, cutoff && isVarDecl, EmptyState, false)
		switch v := val.(type) {
		case *LazyValue:
			return t.GetPrimedVar(v.Expr, v.Con, cutoff)
		case *OpDefNode:
			return t.GetPrimedVar(v.Body, c, cutoff)
		default:
			return nil
		}
	default:
		return nil
	}
}
