package tlc

import "fmt"

type DebugEvalMode int

const (
	DebugEvalConst DebugEvalMode = iota
	DebugEvalState
	DebugEvalAction
	DebugEvalDebugger
)

func (m DebugEvalMode) String() string {
	switch m {
	case DebugEvalConst:
		return "Const"
	case DebugEvalState:
		return "State"
	case DebugEvalAction:
		return "Action"
	case DebugEvalDebugger:
		return "Debugger"
	default:
		return fmt.Sprintf("DebugTool.EvalMode(%d)", m)
	}
}

func (t *Tool) installDebugTool() *Tool {
	if t == nil || t.ownedDebugger() == nil {
		return t
	}
	if t.ownedDebugFastTool() == nil {
		fast := *t
		fast.Debugger = nil
		fast.DebugPort = -1
		fast.DebugFastTool = nil
		fast.debugOwnersCaptured = false
		fast.debuggerOwner = nil
		fast.debugFastToolOwner = nil
		fast.DebugEvalMode = DebugEvalConst
		t.DebugFastTool = &fast
	}
	t.captureDebugOwners()
	t.DebugEvalMode = DebugEvalConst
	t.EvalFunc = debugToolEval
	t.GetInitStatesFunc = debugToolGetInitStates
	t.GetNextStatesFunc = debugToolGetNextStates
	t.GetNextStatesWithFunctorFn = debugToolGetNextStatesWithFunctor
	t.GetNextStatesForActionFunc = debugToolGetNextStatesForAction
	t.IsValidExprFunc = debugToolIsValidExpr
	t.IsValidStateFunc = debugToolIsValidState
	t.NoDebugFunc = debugToolNoDebug
	t.IsDebuggerFunc = func(tool *Tool) bool { return tool != nil && tool.ownedDebugger() != nil }
	return t
}

func debugToolNoDebug(t *Tool) *Tool {
	if t == nil {
		return nil
	}
	if fast := t.ownedDebugFastTool(); fast != nil {
		return fast
	}
	copied := *t
	copied.Debugger = nil
	copied.DebugPort = -1
	copied.DebugFastTool = nil
	copied.debugOwnersCaptured = false
	copied.debuggerOwner = nil
	copied.debugFastToolOwner = nil
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
		panic(NewNullPointerException())
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
		t.DebugEvalMode = DebugEvalState
		return t.debugEvalImpl(expr, c, s0, s1, control, cm)
	}
	if EvalIsConst(control) {
		t.DebugEvalMode = DebugEvalConst
		return t.debugEvalImpl(expr, c, s0, s1, control, cm)
	}
	if t.DebugEvalMode == DebugEvalState && stateNoneAssigned(s1) {
		return t.debugEvalImpl(expr, c, s0, s1, control, cm)
	}
	if t.DebugEvalMode == DebugEvalConst && stateNoneAssigned(s0) && stateNoneAssigned(s1) {
		return t.debugEvalImpl(expr, c, s0, s1, control, cm)
	}
	t.DebugEvalMode = DebugEvalAction
	return t.debugEvalImpl(expr, c, s0, s1, control, cm)
}

func (t *Tool) debugEvalImpl(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	for {
		value, err := debugEvalImplWithReset(t, expr, c, s0, s1, control, cm)
		if debugResetTargets(err, expr) {
			continue
		}
		return value, err
	}
}

func debugEvalImplWithReset(t *Tool, expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if reset, ok := recovered.(*ResetEvalException); ok {
				err = reset
			} else {
				panic(recovered)
			}
		}
	}()
	return t.debugEvalImplOnce(expr, c, s0, s1, control, cm)
}

func (t *Tool) debugEvalImplOnce(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (Value, error) {
	if t == nil {
		panic(NewNullPointerException())
	}
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
	if t == nil {
		panic(NewNullPointerException())
	}
	if debugger := t.ownedDebugger(); debugger != nil {
		defer func() {
			if isJavaEvalOrRuntimeException(err) && !debugErrorKnown(err) {
				debugger.PushExceptionFrame(t, expr, c, err)
				debugger.PopExceptionFrame(t, expr, c, value, err)
			}
			debugger.PopValueFrame(t, expr, c, value)
		}()
		debugger.PushFrame(t, expr, c)
	}
	return t.EvalImpl(expr, c, s0, s1, control, cm)
}

func (t *Tool) debugStateLevelEval(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	if debugger := t.ownedDebugger(); debugger != nil {
		defer func() {
			if isJavaEvalOrRuntimeException(err) && !debugErrorKnown(err) {
				debugger.PushStateExceptionFrame(t, expr, c, s0, err)
				debugger.PopExceptionFrame(t, expr, c, value, err)
			}
			debugger.PopValueFrame(t, expr, c, value)
		}()
		debugger.PushStateFrame(t, expr, c, s0)
	}
	return t.EvalImpl(expr, c, s0, s1, control, cm)
}

func (t *Tool) debugActionLevelEval(expr SemanticNode, c *Context, s0 *TLCStateMut, s1 *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	if t == nil {
		panic(NewNullPointerException())
	}
	action := actionFromDebugState(s1)
	if debugger := t.ownedDebugger(); debugger != nil {
		defer func() {
			if isJavaEvalOrRuntimeException(err) && !debugErrorKnown(err) {
				debugger.PushActionExceptionFrame(t, expr, c, s0, action, s1, err)
				debugger.PopExceptionFrame(t, expr, c, value, err)
			}
			debugger.PopValueFrame(t, expr, c, value)
		}()
		debugger.PushActionFrame(t, expr, c, s0, action, s1)
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
	if debugger := t.ownedDebugger(); debugger != nil {
		debugger.PushInitStatesFrame(t, functor)
		defer debugger.PopInitStatesFrame(t, functor)
	}
	wrapped := debugWrapStateFunctor(t, functor)
	if t.ownedSpecProcessor() != nil || t.GetInitStateSpec().Size() != 0 {
		return t.GetInitStatesImpl(wrapped)
	}
	for _, state := range t.InitStates {
		if _, err := wrapped.AddElement(state); err != nil {
			return err
		}
	}
	return nil
}

func debugToolGetNextStates(t *Tool, action *Action, state *TLCStateMut) (*StateVec, error) {
	if t == nil || action == nil || action.ownedPred() == nil {
		panic(NewNullPointerException())
	}
	if t.DebugEvalMode == DebugEvalDebugger {
		return t.NoDebug().GetNextStates(action, state)
	}
	t.DebugEvalMode = DebugEvalAction
	return t.getNextStateVec(action, state)
}

func debugToolGetNextStatesWithFunctor(t *Tool, functor *NextStateFunctor, state *TLCStateMut) (bool, error) {
	if t == nil {
		return false, nil
	}
	if t.DebugEvalMode == DebugEvalDebugger {
		for _, action := range t.GetActions() {
			if _, err := t.NoDebug().GetNextStatesForAction(functor, state, action); err != nil {
				return true, err
			}
		}
		return false, nil
	}
	if debugger := t.ownedDebugger(); debugger != nil {
		debugger.PushNextStatesFrame(t, functor, state)
		defer debugger.PopNextStatesFrame(t, functor, state)
	}
	for _, action := range t.GetActions() {
		if _, err := t.GetNextStatesForAction(functor, state, action); err != nil {
			return true, err
		}
	}
	return false, nil
}

func debugToolGetNextStatesForAction(t *Tool, functor *NextStateFunctor, state *TLCStateMut, action *Action) (bool, error) {
	if t == nil || action == nil || action.ownedPred() == nil {
		panic(NewNullPointerException())
	}
	if t.DebugEvalMode == DebugEvalDebugger {
		if _, err := t.NoDebug().GetNextStatesForAction(functor, state, action); err != nil {
			return true, err
		}
	}
	t.DebugEvalMode = DebugEvalAction
	functor = debugWrapNextStateFunctor(t, functor)
	defer func() {
		if t.GetMode() == ModeDebugger && state != nil {
			state.UnsetPredecessor()
		}
	}()
	for {
		_, err := debugActionEvalWithControl(t, func() (any, error) {
			s1 := NewEmptyState().SetPredecessor(state).SetAction(action)
			_, err := t.GetNextStatesForPredicate(action, action.ownedPred(), EmptyActionItemList, action.ownedContext(), state, s1, functor, action.CM)
			return nil, err
		})
		if reset, ok := err.(*ResetEvalException); ok && reset != nil {
			if !reset.IsTarget(action.GetOpDef()) {
				panic(NewAssertionError())
			}
			continue
		}
		if abort, ok := err.(*AbortEvalException); ok && abort != nil {
			return false, nil
		}
		if err != nil {
			return true, err
		}
		return false, nil
	}
}

func debugActionEvalWithControl(t *Tool, fn func() (any, error)) (value any, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			switch recovered := recovered.(type) {
			case *ResetEvalException:
				err = recovered
			case *AbortEvalException:
				err = recovered
			default:
				panic(recovered)
			}
		}
	}()
	return t.withDebugEvalModeAny(DebugEvalAction, fn)
}

func debugToolIsValidExpr(t *Tool, expr SemanticNode, ctxt *Context) (bool, error) {
	t.DebugEvalMode = DebugEvalConst
	return t.IsValidExprImpl(expr, ctxt)
}

// Java marks only a failed assumption, not every false constant expression.
func debugToolIsValidAssumption(t *Tool, expr SemanticNode) (bool, error) {
	debugger := t.ownedDebugger()
	for {
		valid, err := t.IsValidExpr(expr)
		if err != nil || valid {
			return valid, err
		}
		reset := false
		func() {
			defer func() {
				if failure := recover(); failure != nil {
					if _, ok := failure.(*ResetEvalException); !ok {
						panic(failure)
					}
					debugger.PopFrame(t, expr, EmptyContext)
					reset = true
				}
			}()
			debugger.MarkAssumptionViolatedFrame(t, expr, EmptyContext)
		}()
		if !reset {
			return false, nil
		}
	}
}

func debugToolIsValidState(t *Tool, action *Action, state *TLCStateMut) (bool, error) {
	if action.IsInternal() {
		t.DebugEvalMode = DebugEvalDebugger
		defer func() { t.DebugEvalMode = DebugEvalState }()
		return t.IsValidStateImpl(action, state)
	}
	t.DebugEvalMode = DebugEvalState
	valid, err := t.IsValidStateImpl(action, state)
	if _, reset := err.(*ResetEvalException); reset {
		return debugToolIsValidState(t, action, state)
	}
	return valid, err
}

func (t *Tool) SetDebugger() DebugEvalMode {
	if t == nil {
		panic(NewNullPointerException())
	}
	old := t.DebugEvalMode
	t.DebugEvalMode = DebugEvalDebugger
	return old
}

func (t *Tool) withDebugEvalModeAny(mode DebugEvalMode, fn func() (any, error)) (any, error) {
	if t == nil {
		return fn()
	}
	old := t.DebugEvalMode
	if mode == DebugEvalDebugger {
		old = t.SetDebugger()
	} else {
		t.DebugEvalMode = mode
	}
	if mode == DebugEvalDebugger && t.ownedDebugFastTool() != nil {
		oldStateTool := stateTool
		fastTool := t.NoDebug()
		oldTool := ResetExtendedTLCStateTool(fastTool)
		stateTool = fastTool
		defer func() {
			stateTool = oldStateTool
			ResetExtendedTLCStateTool(oldTool)
		}()
	}
	defer func() { t.DebugEvalMode = old }()
	return fn()
}

func debugWrapStateFunctor(tool *Tool, functor *StateFunctor) *StateFunctor {
	if tool == nil || tool.ownedDebugger() == nil || functor == nil || functor.debugWrapped {
		return functor
	}
	debugger := tool.ownedDebugger()
	wrapped := *functor
	wrapped.debugWrapped = true
	wrapped.AddElementFunc = func(state *TLCStateMut) (any, error) {
		defer debugger.PopGeneratedStateFrame(state)
		debugger.PushGeneratedStateFrame(state)
		return functor.AddElement(state)
	}
	wrapped.AddUnsatisfiedStateFunc = func(state *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
		debugger.PushUnsatisfiedFrame(tool, pred, con, state)
		debugger.PopFrame(tool, pred, con)
		return functor.AddUnsatisfiedState(state, pred, con)
	}
	return &wrapped
}

func debugWrapNextStateFunctor(tool *Tool, functor *NextStateFunctor) *NextStateFunctor {
	if tool == nil || tool.ownedDebugger() == nil || functor == nil || functor.debugWrapped {
		return functor
	}
	debugger := tool.ownedDebugger()
	wrapped := *functor
	wrapped.debugWrapped = true
	wrapped.AddUnsatisfiedNextStateFn = func(curState *TLCStateMut, action *Action, succState *TLCStateMut, pred SemanticNode, con *Context) *TLCStateMut {
		succState.SetAction(action).SetPredecessor(curState)
		debugger.PushUnsatisfiedActionFrame(tool, pred, con, curState, action, succState)
		debugger.PopFrame(tool, pred, con)
		return functor.AddUnsatisfiedNextState(curState, action, succState, pred, con)
	}
	return &wrapped
}

func debugToolActive(t *Tool) bool {
	return t != nil && t.ownedDebugger() != nil && t.ownedDebugFastTool() != nil
}

func (t *Tool) debugToolIsInitializing() bool {
	return t == nil || t.ownedDebugger() == nil || (MainChecker() == nil && CurrentSimulator() == nil)
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
