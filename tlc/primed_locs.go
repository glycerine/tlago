package tlc

func (t *Tool) GetPrimedLocs() *SemanticNodeLongTable {
	tbl := NewSemanticNodeLongTable(10)
	if t == nil {
		return tbl
	}
	action := t.GetNextStateSpec()
	if action == nil {
		return tbl
	}
	t.collectPrimedLocs(action.Pred, action.Con, tbl, make(map[*OpDefNode]bool))
	return tbl
}

func (t *Tool) CollectPrimedLocs(pred SemanticNode, c *Context, tbl *SemanticNodeLongTable) {
	t.collectPrimedLocs(pred, c, tbl, make(map[*OpDefNode]bool))
}

func (t *Tool) collectPrimedLocs(pred SemanticNode, c *Context, tbl *SemanticNodeLongTable, visiting map[*OpDefNode]bool) {
	if tbl == nil {
		return
	}
	switch pred := pred.(type) {
	case *OpApplNode:
		t.collectPrimedLocsAppl(pred, c, tbl, visiting)
	case *LetInNode:
		t.collectPrimedLocs(pred.Body, c, tbl, visiting)
	case *SubstInNode:
		c1 := c
		for _, sub := range pred.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, true, DoNotRecordCostModel))
		}
		_ = c1
		t.collectPrimedLocs(pred.Body, c, tbl, visiting)
	case *APSubstInNode:
		c1 := c
		for _, sub := range pred.Substs {
			c1 = c1.Cons(sub.Op, t.GetVal(sub.Expr, c, true, DoNotRecordCostModel))
		}
		_ = c1
		t.collectPrimedLocs(pred.Body, c, tbl, visiting)
	case *LabelNode:
		t.collectPrimedLocs(pred.Body, c, tbl, visiting)
	}
}

func (t *Tool) collectPrimedLocsAppl(pred *OpApplNode, c *Context, tbl *SemanticNodeLongTable, visiting map[*OpDefNode]bool) {
	if pred == nil || pred.Operator == nil || pred.Operator.Name == nil {
		return
	}
	args := pred.Args
	opNode := pred.Operator
	opcode := GetOpCode(opNode.Name)

	switch opcode {
	case OpcodeFA:
		if len(args) > 0 {
			t.collectPrimedLocs(args[0], c, tbl, visiting)
		}
	case OpcodeITE:
		if len(args) > 1 {
			t.collectPrimedLocs(args[1], c, tbl, visiting)
		}
		if len(args) > 2 {
			t.collectPrimedLocs(args[2], c, tbl, visiting)
		}
	case OpcodeCase:
		for _, arg := range args {
			pair, ok := arg.(*OpApplNode)
			if ok && len(pair.Args) > 1 {
				t.collectPrimedLocs(pair.Args[1], c, tbl, visiting)
			}
		}
	case OpcodeEq, OpcodeIn:
		if len(args) > 0 {
			varNode := t.GetPrimedVar(args[0], c, false)
			if varNode != nil && varNode.IsVariableDecl() && varNode.Name != nil && varNode.Name.VarLoc() != -1 {
				tbl.Put(pred, 0)
			}
		}
	case OpcodeCL, OpcodeDL, OpcodeBE, OpcodeBF, OpcodeLand, OpcodeLor, OpcodeImplies, OpcodeNop:
		for _, arg := range args {
			t.collectPrimedLocs(arg, c, tbl, visiting)
		}
	case OpcodeUnchanged:
		if len(args) > 0 {
			t.collectUnchangedLocs(args[0], c, tbl, visiting)
		}
	case OpcodeAA:
		if len(args) > 0 {
			t.collectPrimedLocs(args[0], c, tbl, visiting)
		}
	case OpcodeSA:
		if len(args) > 0 {
			t.collectPrimedLocs(args[0], c, tbl, visiting)
		}
		if len(args) > 1 {
			tbl.Put(args[1], 0)
		}
	default:
		if opcode != 0 {
			return
		}
		val := t.Lookup(opNode, c, EmptyState, false)
		switch v := val.(type) {
		case *OpDefNode:
			if visiting[v] {
				return
			}
			visiting[v] = true
			c1, err := t.GetOpContext(v, args, c, true, DoNotRecordCostModel)
			if err == nil {
				t.collectPrimedLocs(v.Body, c1, tbl, visiting)
			}
			delete(visiting, v)
		case *LazyValue:
			t.collectPrimedLocs(v.Expr, v.Con, tbl, visiting)
		}
	}
}

func (t *Tool) collectUnchangedLocs(expr SemanticNode, c *Context, tbl *SemanticNodeLongTable, visiting map[*OpDefNode]bool) {
	expr1, ok := expr.(*OpApplNode)
	if !ok || expr1.Operator == nil || expr1.Operator.Name == nil {
		return
	}
	opNode := expr1.Operator
	opName := opNode.Name
	opcode := GetOpCode(opName)
	if opNode.IsVariableDecl() && opName.VarLoc() >= 0 {
		tbl.Put(expr, 0)
		return
	}
	args := expr1.Args
	if opcode == OpcodeTup {
		for _, arg := range args {
			t.collectUnchangedLocs(arg, c, tbl, visiting)
		}
		return
	}
	if opcode == 0 && len(args) == 0 {
		val := t.Lookup(opNode, c, EmptyState, false)
		if opDef, ok := val.(*OpDefNode); ok {
			if visiting[opDef] {
				return
			}
			visiting[opDef] = true
			t.collectUnchangedLocs(opDef.Body, c, tbl, visiting)
			delete(visiting, opDef)
		}
	}
}
