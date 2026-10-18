package tlc

func (t *Tool) ensureActionsPrepared() error {
	if t == nil || t.actionsCaptured || t.actionsPrepared || len(t.Actions) != 0 {
		return nil
	}
	return t.PrepareActionsFromNextStateSpec()
}

func (t *Tool) PrepareActionsFromNextStateSpec() error {
	if t == nil {
		return nil
	}
	next := t.GetNextStateSpec()
	if next == nil {
		// Tool constructs an allocated, zero-length action array when the
		// specification has no next-state predicate (e.g. assumptions only).
		t.Actions = make([]*Action, 0)
		t.actionsPrepared = true
		t.captureActions(t.Actions)
		t.assignActionIDs(t.ownedActions())
		return nil
	}
	actions := make([]*Action, 0, 10)
	if err := t.collectActions(next.ownedPred(), next.ownedContext(), next.GetOpDef(), next.CM, &actions); err != nil {
		return err
	}
	t.Actions = actions
	t.actionsPrepared = true
	t.captureActions(t.Actions)
	t.assignActionIDs(t.ownedActions())
	return nil
}

func (t *Tool) collectActions(next SemanticNode, con *Context, opDef *OpDefNode, cm CostModel, actions *[]*Action) error {
	if isNil(next) {
		panic(NewNullPointerException())
	}
	switch node := next.(type) {
	case *OpApplNode:
		return t.collectActionsAppl(node, con, opDef, cm, actions)
	case *LetInNode:
		return t.collectActions(node.Body, con, opDef, cm, actions)
	case *SubstInNode:
		if len(node.Substs) == 0 {
			return t.collectActions(node.GetBody(), con, opDef, cm, actions)
		}
		t.appendSplitAction(actions, node, con, opDef, cm)
		return nil
	case *APSubstInNode:
		if len(node.Substs) == 0 {
			return t.collectActions(node.GetBody(), con, opDef, cm, actions)
		}
		t.appendSplitAction(actions, node, con, opDef, cm)
		return nil
	case *LabelNode:
		return t.collectActions(node.Body, con, opDef, cm, actions)
	default:
		return NewTLCDetailedRuntimeException(ECGeneral, "The next state relation is not a boolean expression.\n"+SemanticString(next), next, con)
	}
}

func (t *Tool) collectActionsAppl(next *OpApplNode, con *Context, opDef *OpDefNode, cm CostModel, actions *[]*Action) error {
	if next == nil || next.Operator == nil {
		panic(NewNullPointerException())
	}
	args := next.Args
	opcode := GetOpCode(next.Operator.Name)
	if opcode == 0 {
		val := t.Lookup(next.Operator, con, EmptyState, false)
		if def, ok := val.(*OpDefNode); ok {
			opcode = GetOpCode(def.Name)
			if opcode == 0 && t.collectConstantActions(def, args, con, cm, actions) {
				return nil
			}
		}
		if opcode == 0 {
			// Java uses the applied operator's declaration when state-level
			// arguments prevent decomposition, rather than the enclosing action.
			t.appendSplitAction(actions, next, con, next.Operator.Definition, cm)
			return nil
		}
	}

	switch opcode {
	case OpcodeBE:
		return t.collectBoundedExistsActions(next, con, opDef, cm, actions)
	case OpcodeDL, OpcodeLor:
		for _, arg := range args {
			if err := t.collectActions(arg, con, opDef, cm, actions); err != nil {
				return err
			}
		}
	default:
		t.appendSplitAction(actions, next, con, opDef, cm)
	}
	return nil
}

func (t *Tool) collectBoundedExistsActions(next *OpApplNode, con *Context, opDef *OpDefNode, cm CostModel, actions *[]*Action) (err error) {
	count := len(*actions)
	fallback := func() error {
		*actions = (*actions)[:count]
		t.appendSplitAction(actions, next, con, opDef, cm)
		return nil
	}
	defer func() {
		if recover() != nil {
			err = fallback()
		}
	}()
	enum, err := t.Contexts(next, con, EmptyState, EmptyState, EvalClear, cm)
	if err != nil || enum == nil || enum.Err() != nil || enum.IsDone() {
		return fallback()
	}
	for econ := enum.NextElement(); econ != nil; econ = enum.NextElement() {
		if err := enum.Err(); err != nil {
			return fallback()
		}
		if len(next.Args) == 0 {
			return fallback()
		}
		if err := t.collectActions(next.Args[0], econ, opDef, cm, actions); err != nil {
			return fallback()
		}
	}
	if enum.Err() != nil || len(*actions) == count {
		return fallback()
	}
	return nil
}

// Java catches Throwable while expanding constant operator arguments, then
// appends an unsplit action. Actions added before the failure remain in place.
func (t *Tool) collectConstantActions(def *OpDefNode, args []SemanticNode, con *Context, cm CostModel, actions *[]*Action) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	if !t.actionArgsAreConstant(args) {
		return false
	}
	c1, ok := t.actionConstantContext(def, args, con, cm)
	if !ok {
		return false
	}
	return t.collectActions(def.GetBody(), c1, def, cm, actions) == nil
}

func (t *Tool) actionArgsAreConstant(args []SemanticNode) bool {
	for _, arg := range args {
		if SemanticLevel(arg) != TLCLevelConstant {
			return false
		}
	}
	return true
}

func (t *Tool) actionConstantContext(def *OpDefNode, args []SemanticNode, con *Context, cm CostModel) (*Context, bool) {
	if def == nil {
		return con, false
	}
	params := def.GetParams()
	c1 := con
	for i, arg := range args {
		aval, err := t.Eval(arg, con, EmptyState, EmptyState, EvalClear, cm)
		if err != nil {
			return con, false
		}
		if params == nil {
			panic(NewNullPointerException())
		}
		if i >= len(params) {
			panic(NewArrayIndexOutOfBoundsException(i, len(params)))
		}
		c1 = c1.Cons(params[i], aval)
	}
	return c1, true
}

func (t *Tool) appendSplitAction(actions *[]*Action, pred SemanticNode, con *Context, opDef *OpDefNode, cm CostModel) {
	action := NewActionFromOpDef(pred, con, opDef, false, false)
	action.CM = cm.Get(pred)
	*actions = append(*actions, action)
}
