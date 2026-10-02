package tlc

func (t *Tool) ensureActionsPrepared() error {
	if t == nil || t.actionsPrepared || t.NextStateSpec == nil || len(t.Actions) != 0 {
		return nil
	}
	return t.PrepareActionsFromNextStateSpec()
}

func (t *Tool) PrepareActionsFromNextStateSpec() error {
	if t == nil {
		return nil
	}
	next := t.NextStateSpec
	if next == nil {
		t.Actions = nil
		t.actionsPrepared = true
		return nil
	}
	actions := make([]*Action, 0, 10)
	if err := t.collectActions(next.Pred, next.Con, next.GetOpDef(), next.CM, &actions); err != nil {
		return err
	}
	t.Actions = actions
	t.actionsPrepared = true
	t.assignActionIDs(t.Actions)
	return nil
}

func (t *Tool) collectActions(next SemanticNode, con *Context, opDef *OpDefNode, cm CostModel, actions *[]*Action) error {
	if con == nil {
		con = EmptyContext
	}
	switch node := next.(type) {
	case *OpApplNode:
		return t.collectActionsAppl(node, con, opDef, cm, actions)
	case *LetInNode:
		return t.collectActions(node.Body, con, opDef, cm, actions)
	case *SubstInNode:
		if len(node.Substs) == 0 {
			return t.collectActions(node.Body, con, opDef, cm, actions)
		}
		t.appendSplitAction(actions, node, con, opDef, cm)
		return nil
	case *APSubstInNode:
		if len(node.Substs) == 0 {
			return t.collectActions(node.Body, con, opDef, cm, actions)
		}
		t.appendSplitAction(actions, node, con, opDef, cm)
		return nil
	case *LabelNode:
		return t.collectActions(node.Body, con, opDef, cm, actions)
	default:
		return newTLCError(ECGeneral, "The next state relation is not a boolean expression.\n%s", SemanticString(next))
	}
}

func (t *Tool) collectActionsAppl(next *OpApplNode, con *Context, opDef *OpDefNode, cm CostModel, actions *[]*Action) error {
	if next == nil || next.Operator == nil {
		t.appendSplitAction(actions, next, con, opDef, cm)
		return nil
	}
	args := next.Args
	opcode := GetOpCode(next.Operator.Name)
	if opcode == 0 {
		val := t.Lookup(next.Operator, con, EmptyState, false)
		if def, ok := val.(*OpDefNode); ok {
			opcode = GetOpCode(def.Name)
			if opcode == 0 && t.actionArgsAreConstant(args) {
				if c1, ok := t.actionConstantContext(def, args, con, cm); ok {
					return t.collectActions(def.Body, c1, def, cm, actions)
				}
			}
		}
		if opcode == 0 {
			t.appendSplitAction(actions, next, con, opDef, cm)
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

func (t *Tool) collectBoundedExistsActions(next *OpApplNode, con *Context, opDef *OpDefNode, cm CostModel, actions *[]*Action) error {
	count := len(*actions)
	fallback := func() error {
		*actions = (*actions)[:count]
		t.appendSplitAction(actions, next, con, opDef, cm)
		return nil
	}
	enum, err := t.Contexts(next, con, EmptyState, EmptyState, EvalClear, cm)
	if err != nil || enum == nil || enum.Err() != nil || enum.IsDone() {
		return fallback()
	}
	for econ := enum.NextElement(); econ != nil; econ = enum.NextElement() {
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

func (t *Tool) actionArgsAreConstant(args []SemanticNode) bool {
	for _, arg := range args {
		if SemanticLevel(arg) != TLCLevelConstant {
			return false
		}
	}
	return true
}

func (t *Tool) actionConstantContext(def *OpDefNode, args []SemanticNode, con *Context, cm CostModel) (*Context, bool) {
	if def == nil || len(def.Params) != len(args) {
		return con, false
	}
	c1 := con
	if c1 == nil {
		c1 = EmptyContext
	}
	for i, arg := range args {
		aval, err := t.Eval(arg, con, EmptyState, EmptyState, EvalClear, cm)
		if err != nil {
			return con, false
		}
		c1 = c1.Cons(def.Params[i], aval)
	}
	return c1, true
}

func (t *Tool) appendSplitAction(actions *[]*Action, pred SemanticNode, con *Context, opDef *OpDefNode, cm CostModel) {
	action := NewActionFromOpDef(pred, con, opDef, false, false)
	action.CM = cm.Get(pred)
	*actions = append(*actions, action)
}
