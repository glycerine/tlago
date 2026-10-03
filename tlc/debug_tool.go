package tlc

type DebugEvalMode int

const (
	DebugEvalConst DebugEvalMode = iota
	DebugEvalState
	DebugEvalAction
	DebugEvalDebugger
)

func (t *Tool) installDebugTool() *Tool {
	if t == nil || t.Debugger == nil {
		return t
	}
	if t.DebugFastTool == nil {
		fast := *t
		fast.Debugger = nil
		fast.DebugPort = -1
		fast.DebugFastTool = nil
		fast.DebugEvalMode = DebugEvalConst
		if fast.Mode == ModeDebugger {
			fast.Mode = ModeMC
		}
		t.DebugFastTool = &fast
	}
	t.DebugEvalMode = DebugEvalConst
	t.EvalFunc = debugToolEval
	t.GetInitStatesFunc = debugToolGetInitStates
	t.GetNextStatesFunc = debugToolGetNextStates
	t.GetNextStatesWithFunctorFn = debugToolGetNextStatesWithFunctor
	t.GetNextStatesForActionFunc = debugToolGetNextStatesForAction
	t.IsValidExprFunc = debugToolIsValidExpr
	t.IsValidStateFunc = debugToolIsValidState
	t.NoDebugFunc = debugToolNoDebug
	t.IsDebuggerFunc = func(tool *Tool) bool { return tool != nil && tool.Debugger != nil }
	return t
}

func debugToolNoDebug(t *Tool) *Tool {
	if t == nil {
		return nil
	}
	if t.DebugFastTool != nil {
		return t.DebugFastTool
	}
	copied := *t
	copied.Debugger = nil
	copied.DebugPort = -1
	copied.DebugFastTool = nil
	if copied.Mode == ModeDebugger {
		copied.Mode = ModeMC
	}
	return &copied
}

func debugToolEval(t *Tool, expr SemanticNode, args ...any) (Value, error) {
	c, s0, s1, control, cm := parseEvalArgs(args...)
	// Preserve DebugTool's distinct Java eval overloads. The overload accepting
	// only one state forces State mode, including VIEW fingerprint evaluation.
	states, explicitControl := 0, false
	for _, arg := range args {
		switch arg.(type) {
		case nil, *TLCStateMut:
			states++
		case int:
			explicitControl = true
		}
	}
	if !explicitControl && states < 2 {
		if states == 0 {
			t.DebugEvalMode = DebugEvalConst
			c = EmptyContext
		} else {
			t.DebugEvalMode = DebugEvalState
		}
		return t.debugEvalImpl(expr, c, s0, s1, control, cm)
	}
	for {
		value, err := debugEvalWithReset(t, expr, c, s0, s1, control, cm)
		if debugResetTargets(err, expr) {
			continue
		}
		return value, err
	}
}

func debugEvalWithReset(t *Tool, expr SemanticNode, c *Context, s0, s1 *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if reset, ok := recovered.(*ResetEvalException); ok {
				err = reset
			} else {
				panic(recovered)
			}
		}
	}()
	return t.debugEval(expr, c, s0, s1, control, cm)
}

func (t *Tool) debugEval(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if t == nil {
		return ValUndef, newTLCError(ECGeneral, "attempted to evaluate with nil debug tool")
	}
	fast := t.NoDebug()
	if t.DebugEvalMode == DebugEvalDebugger {
		return fast.EvalImpl(expr, c, s0, s1, control, cm)
	}
	if s1 == nil || EvalIsPrimed(control) || EvalIsEnabled(control) {
		return fast.EvalImpl(expr, c, s0, s1, control, cm)
	}
	if t.DebugEvalMode == DebugEvalAction && s1.GetAction() == nil {
		return fast.EvalImpl(expr, c, s0, s1, control, cm)
	}
	if EvalIsInit(control) {
		return t.withDebugEvalMode(DebugEvalState, func() (Value, error) {
			return t.debugEvalImpl(expr, c, s0, s1, control, cm)
		})
	}
	if EvalIsConst(control) {
		return t.withDebugEvalMode(DebugEvalConst, func() (Value, error) {
			return t.debugEvalImpl(expr, c, s0, s1, control, cm)
		})
	}
	if t.DebugEvalMode == DebugEvalState && stateNoneAssigned(s1) {
		return t.debugEvalImpl(expr, c, s0, s1, control, cm)
	}
	if t.DebugEvalMode == DebugEvalConst && stateNoneAssigned(s0) && stateNoneAssigned(s1) {
		return t.debugEvalImpl(expr, c, s0, s1, control, cm)
	}
	return t.withDebugEvalMode(DebugEvalAction, func() (Value, error) {
		return t.debugEvalImpl(expr, c, s0, s1, control, cm)
	})
}

func (t *Tool) debugEvalImpl(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	fast := t.NoDebug()
	if t.debugToolIsInitializing() || debugToolIsLiveness(control, s0, s1) || debugToolIsLeaf(expr) || t.DebugEvalMode == DebugEvalDebugger {
		return fast.EvalImpl(expr, c, s0, s1, control, cm)
	}
	switch t.DebugEvalMode {
	case DebugEvalConst:
		return t.debugConstLevelEval(expr, c, s0, s1, control, cm)
	case DebugEvalState:
		return t.debugStateLevelEval(expr, c, s0, s1, control, cm)
	default:
		return t.debugActionLevelEval(expr, c, s0, s1, control, cm)
	}
}

func (t *Tool) debugConstLevelEval(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	if t.Debugger != nil {
		defer func() {
			if isJavaEvalOrRuntimeException(err) && !debugErrorKnown(err) {
				t.Debugger.PushExceptionFrame(t, expr, c, err)
				t.Debugger.PopExceptionFrame(t, expr, c, value, err)
			}
			t.Debugger.PopValueFrame(t, expr, c, value)
		}()
		t.Debugger.PushFrame(t, expr, c)
	}
	return t.EvalImpl(expr, c, s0, s1, control, cm)
}

func (t *Tool) debugStateLevelEval(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	if t.Debugger != nil {
		defer func() {
			if isJavaEvalOrRuntimeException(err) && !debugErrorKnown(err) {
				t.Debugger.PushStateExceptionFrame(t, expr, c, s0, err)
				t.Debugger.PopExceptionFrame(t, expr, c, value, err)
			}
			t.Debugger.PopValueFrame(t, expr, c, value)
		}()
		t.Debugger.PushStateFrame(t, expr, c, s0)
	}
	return t.EvalImpl(expr, c, s0, s1, control, cm)
}

func (t *Tool) debugActionLevelEval(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	action := actionFromDebugState(s1)
	if t.Debugger != nil {
		defer func() {
			if isJavaEvalOrRuntimeException(err) && !debugErrorKnown(err) {
				t.Debugger.PushActionExceptionFrame(t, expr, c, s0, action, s1, err)
				t.Debugger.PopExceptionFrame(t, expr, c, value, err)
			}
			t.Debugger.PopValueFrame(t, expr, c, value)
		}()
		t.Debugger.PushActionFrame(t, expr, c, s0, action, s1)
	}
	return t.EvalImpl(expr, c, s0, s1, control, cm)
}

func debugToolGetInitStates(t *Tool, functor *StateFunctor) error {
	if t == nil {
		return nil
	}
	if t.DebugEvalMode == DebugEvalDebugger {
		return t.NoDebug().GetInitStates(functor)
	}
	if t.Debugger != nil {
		t.Debugger.PushInitStatesFrame(t, functor)
		defer t.Debugger.PopInitStatesFrame(t, functor)
	}
	wrapped := debugWrapStateFunctor(t, functor)
	_, err := t.withDebugEvalModeAny(DebugEvalState, func() (any, error) {
		if len(t.InitStateSpec) != 0 {
			return nil, t.GetInitStatesImpl(wrapped)
		}
		for _, state := range t.InitStates {
			if _, err := wrapped.AddElement(state); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func debugToolGetNextStates(t *Tool, action *Action, state *TLCStateMut) (*StateVec, error) {
	if t == nil || action == nil || action.Pred == nil {
		return NewStateVec(0), nil
	}
	if t.DebugEvalMode == DebugEvalDebugger {
		return t.NoDebug().GetNextStates(action, state)
	}
	var out *StateVec
	_, err := t.withDebugEvalModeAny(DebugEvalAction, func() (any, error) {
		vec, err := t.GetNextStatesImpl(action, state)
		out = vec
		return vec, err
	})
	if out == nil {
		out = NewStateVec(0)
	}
	return out, err
}

func debugToolGetNextStatesWithFunctor(t *Tool, functor *NextStateFunctor, state *TLCStateMut) (bool, error) {
	if t == nil {
		return false, nil
	}
	if t.DebugEvalMode == DebugEvalDebugger {
		for _, action := range t.GetActions() {
			if halt, err := t.NoDebug().GetNextStatesForAction(functor, state, action); halt || err != nil {
				return halt, err
			}
		}
		return false, nil
	}
	if t.Debugger != nil {
		t.Debugger.PushNextStatesFrame(t, functor, state)
		defer t.Debugger.PopNextStatesFrame(t, functor, state)
	}
	for _, action := range t.GetActions() {
		halt, err := t.GetNextStatesForAction(functor, state, action)
		if halt || err != nil {
			return halt, err
		}
	}
	return false, nil
}

func debugToolGetNextStatesForAction(t *Tool, functor *NextStateFunctor, state *TLCStateMut, action *Action) (bool, error) {
	if t == nil || action == nil || action.Pred == nil {
		return false, nil
	}
	if t.DebugEvalMode == DebugEvalDebugger {
		return t.NoDebug().GetNextStatesForAction(functor, state, action)
	}
	functor = debugWrapNextStateFunctor(t, functor)
	defer func() {
		if t.GetMode() == ModeDebugger && state != nil {
			state.UnsetPredecessor()
		}
	}()
	for {
		_, err := t.withDebugEvalModeAny(DebugEvalAction, func() (any, error) {
			s1 := NewEmptyState().SetPredecessor(state).SetAction(action)
			_, err := t.GetNextStatesForPredicate(action, action.Pred, EmptyActionItemList, action.Con, state, s1, functor, action.CM)
			return nil, err
		})
		if debugResetTargets(err, action.GetOpDef()) {
			continue
		}
		if abort, ok := err.(*AbortEvalException); ok && abort != nil {
			return false, nil
		}
		if err != nil {
			return true, err
		}
		return functor.ShouldHalt(), nil
	}
}

func debugToolIsValidExpr(t *Tool, expr SemanticNode, ctxt *Context) (bool, error) {
	var ok bool
	_, err := t.withDebugEvalModeAny(DebugEvalConst, func() (any, error) {
		valid, err := t.IsValidExprImpl(expr, ctxt)
		ok = valid
		return nil, err
	})
	if err == nil && !ok && t != nil && t.Debugger != nil {
		t.Debugger.MarkAssumptionViolatedFrame(t, expr, EmptyContext)
	}
	return ok, err
}

func debugToolIsValidState(t *Tool, action *Action, state *TLCStateMut) (bool, error) {
	if action != nil && action.IsInternal() {
		return debugToolWithModeBool(t, DebugEvalDebugger, func() (bool, error) {
			return t.IsValidStateImpl(action, state)
		})
	}
	return debugToolWithModeBool(t, DebugEvalState, func() (bool, error) {
		return t.IsValidStateImpl(action, state)
	})
}

func (t *Tool) withDebugEvalMode(mode DebugEvalMode, fn func() (Value, error)) (Value, error) {
	if t == nil {
		return fn()
	}
	old := t.DebugEvalMode
	t.DebugEvalMode = mode
	defer func() { t.DebugEvalMode = old }()
	return fn()
}

func (t *Tool) withDebugEvalModeAny(mode DebugEvalMode, fn func() (any, error)) (any, error) {
	if t == nil {
		return fn()
	}
	old := t.DebugEvalMode
	t.DebugEvalMode = mode
	if mode == DebugEvalDebugger && t.DebugFastTool != nil {
		oldTool := stateTool
		stateTool = t.NoDebug()
		defer func() { stateTool = oldTool }()
	}
	defer func() { t.DebugEvalMode = old }()
	return fn()
}

func debugToolWithModeBool(t *Tool, mode DebugEvalMode, fn func() (bool, error)) (bool, error) {
	if t == nil {
		return fn()
	}
	old := t.DebugEvalMode
	t.DebugEvalMode = mode
	defer func() { t.DebugEvalMode = old }()
	return fn()
}

func debugWrapStateFunctor(tool *Tool, functor *StateFunctor) *StateFunctor {
	if tool == nil || tool.Debugger == nil || functor == nil {
		return functor
	}
	wrapped := *functor
	wrapped.AddElementFunc = func(state *TLCStateMut) (any, error) {
		defer tool.Debugger.PopGeneratedStateFrame(state)
		tool.Debugger.PushGeneratedStateFrame(state)
		return functor.AddElement(state)
	}
	wrapped.AddUnsatisfiedStateFunc = func(state *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
		tool.Debugger.PushUnsatisfiedFrame(tool, pred, con, state)
		defer tool.Debugger.PopFrame(tool, pred, con)
		return functor.AddUnsatisfiedState(state, pred, con)
	}
	return &wrapped
}

func debugWrapNextStateFunctor(tool *Tool, functor *NextStateFunctor) *NextStateFunctor {
	if tool == nil || tool.Debugger == nil || functor == nil {
		return functor
	}
	wrapped := *functor
	wrapped.AddUnsatisfiedNextStateFn = func(curState *TLCStateMut, action *Action, succState *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
		tool.Debugger.PushUnsatisfiedActionFrame(tool, pred, con, curState, action, succState)
		defer tool.Debugger.PopFrame(tool, pred, con)
		return functor.AddUnsatisfiedNextState(curState, action, succState, pred, con)
	}
	return &wrapped
}

func debugToolActive(t *Tool) bool {
	return t != nil && t.Debugger != nil && t.DebugFastTool != nil
}

func (t *Tool) debugToolIsInitializing() bool {
	return t == nil || t.Debugger == nil
}

func debugToolIsLiveness(control int, s0 *TLCStateMut, s1 *TLCStateMut) bool {
	return EvalIsEnabled(control) || EvalIsPrimed(control) || stateIsFunctional(s0) || stateIsFunctional(s1)
}

func debugToolIsLeaf(expr SemanticNode) bool {
	switch SemanticKindOf(expr) {
	case SemanticNumeralKind, SemanticDecimalKind, SemanticStringKind:
		return true
	default:
		return false
	}
}

func stateIsFunctional(state *TLCStateMut) bool {
	return state != nil && state.functional
}

func stateNoneAssigned(state *TLCStateMut) bool {
	return state == nil || state.NoneAssigned()
}

func actionFromDebugState(state *TLCStateMut) *Action {
	if state == nil {
		return nil
	}
	return state.GetAction()
}

func debugResetTargets(err error, expr SemanticNode) bool {
	reset, ok := err.(*ResetEvalException)
	return ok && reset != nil && reset.IsTarget(expr)
}

func debugErrorKnown(err error) bool {
	// Both Java exception families inherit StatefulRuntimeException's flag.
	// Match the thrown object, leaving known flags on nested causes alone.
	if failure, ok := err.(interface{ IsKnown() bool }); ok {
		return failure.IsKnown()
	}
	return false
}

func debugExceptionNotYetHandled(err error) bool {
	if failure, ok := err.(interface{ SetKnown() bool }); ok {
		return !failure.SetKnown()
	}
	return true
}
