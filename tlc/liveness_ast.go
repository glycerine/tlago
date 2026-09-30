package tlc

func ASTToLive(tool *Tool, expr SemanticNode, con *Context) (*LiveExprNode, error) {
	if con == nil {
		con = EmptyContext
	}
	switch n := expr.(type) {
	case nil:
		return LNTrue, nil
	case *LiveExprNode:
		return n, nil
	case *LabelNode:
		return ASTToLive(tool, n.Body, con)
	case *LetInNode:
		return ASTToLive(tool, n.Body, con)
	case *SubstInNode:
		con1 := con
		for _, subst := range n.Substs {
			if subst.Op == nil {
				continue
			}
			if tool == nil {
				con1 = con1.Cons(subst.Op, subst.Expr)
			} else {
				con1 = con1.Cons(subst.Op, tool.GetVal(subst.Expr, con, false, DoNotRecordCostModel))
			}
		}
		return ASTToLive(tool, n.Body, con1)
	case *APSubstInNode:
		con1 := con
		for _, subst := range n.Substs {
			if subst.Op == nil {
				continue
			}
			if tool == nil {
				con1 = con1.Cons(subst.Op, subst.Expr)
			} else {
				con1 = con1.Cons(subst.Op, tool.GetVal(subst.Expr, con, false, DoNotRecordCostModel))
			}
		}
		return ASTToLive(tool, n.Body, con1)
	case *OpApplNode:
		return astToLiveAppl(tool, n, con)
	default:
		level := SemanticLevel(expr)
		if tool != nil && level == TLCLevelConstant {
			level = tool.GetLevelBound(expr, con)
		}
		return astToLiveLevel(tool, expr, con, level)
	}
}

func astToLiveLevel(tool *Tool, expr SemanticNode, con *Context, level int) (*LiveExprNode, error) {
	if level == TLCLevelConstant && tool != nil {
		value, err := tool.Eval(expr, con, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
		if err == nil {
			if boolValue, ok := value.(*BoolValue); ok {
				if boolValue.Val {
					return LNTrue, nil
				}
				return LNFalse, nil
			}
		}
	}
	label := SemanticString(expr)
	if level <= TLCLevelState {
		return NewLNState(label, expr, con, nil), nil
	}
	return NewLNAction(label, expr, con, nil), nil
}

func astToLiveAppl(tool *Tool, expr *OpApplNode, con *Context) (*LiveExprNode, error) {
	if expr == nil {
		return LNTrue, nil
	}
	args := expr.Args
	opNode := expr.Operator
	opcode := 0
	if opNode != nil {
		opcode = GetOpCode(opNode.Name)
	}

	if opcode == 0 && tool != nil && opNode != nil {
		if val := tool.Lookup(opNode, con, EmptyState, false); val != nil {
			switch typed := val.(type) {
			case *OpDefNode:
				if typed != nil && typed.Name != nil {
					opcode = GetOpCode(typed.Name)
				}
				if opcode == 0 && typed != nil && typed.Body != nil && len(typed.Params) == len(args) {
					con1 := con
					for i, param := range typed.Params {
						if param != nil {
							con1 = con1.Cons(param, tool.GetVal(args[i], con, false, DoNotRecordCostModel))
						}
					}
					live, err := ASTToLive(tool, typed.Body, con1)
					if err == nil && live.GetLevel() > LiveLevelAction {
						return live, nil
					}
				}
			case *BoolValue:
				if typed.Val {
					return LNTrue, nil
				}
				return LNFalse, nil
			}
		}
	}

	switch opcode {
	case OpcodeCL, OpcodeLand:
		out := NewLNConj()
		for _, arg := range args {
			kid, err := ASTToLive(tool, arg, con)
			if err != nil {
				return nil, err
			}
			out.AddConj(kid)
		}
		if out.GetLevel() > LiveLevelAction {
			return out, nil
		}
	case OpcodeDL, OpcodeLor:
		out := NewLNDisj()
		for _, arg := range args {
			kid, err := ASTToLive(tool, arg, con)
			if err != nil {
				return nil, err
			}
			out.AddDisj(kid)
		}
		if out.GetLevel() > LiveLevelAction {
			return out, nil
		}
	case OpcodeITE:
		if len(args) >= 3 && tool != nil {
			if guard, err := tool.Eval(args[0], con, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel); err == nil {
				if boolGuard, ok := guard.(*BoolValue); ok {
					if boolGuard.Val {
						return ASTToLive(tool, args[1], con)
					}
					return ASTToLive(tool, args[2], con)
				}
			}
		}
		if len(args) >= 3 {
			guard, err := ASTToLive(tool, args[0], con)
			if err != nil {
				return nil, err
			}
			thenExpr, err := ASTToLive(tool, args[1], con)
			if err != nil {
				return nil, err
			}
			elseExpr, err := ASTToLive(tool, args[2], con)
			if err != nil {
				return nil, err
			}
			out := NewLNDisj(NewLNConj(guard, thenExpr), NewLNConj(NewLNNeg(guard), elseExpr))
			if out.GetLevel() > LiveLevelAction {
				return out, nil
			}
		}
	case OpcodeLnot:
		if len(args) >= 1 {
			arg, err := ASTToLive(tool, args[0], con)
			if err != nil {
				return nil, err
			}
			if arg.GetLevel() > LiveLevelAction {
				return NewLNNeg(arg), nil
			}
		}
	case OpcodeImplies:
		if len(args) >= 2 {
			left, err := ASTToLive(tool, args[0], con)
			if err != nil {
				return nil, err
			}
			right, err := ASTToLive(tool, args[1], con)
			if err != nil {
				return nil, err
			}
			if max(left.GetLevel(), right.GetLevel()) > LiveLevelAction {
				return NewLNDisj(NewLNNeg(left), right), nil
			}
		}
	case OpcodePrime, OpcodeAA:
		return NewLNAction(SemanticString(expr), expr, con, nil), nil
	case OpcodeLeadsto:
		if len(args) >= 2 {
			left, err := ASTToLive(tool, args[0], con)
			if err != nil {
				return nil, err
			}
			right, err := ASTToLive(tool, args[1], con)
			if err != nil {
				return nil, err
			}
			return NewLNAll(NewLNDisj(NewLNNeg(left), NewLNEven(right))), nil
		}
	case OpcodeBox:
		if len(args) >= 1 {
			arg, err := ASTToLive(tool, args[0], con)
			if err != nil {
				return nil, err
			}
			return NewLNAll(arg), nil
		}
	case OpcodeDiamond:
		if len(args) >= 1 {
			arg, err := ASTToLive(tool, args[0], con)
			if err != nil {
				return nil, err
			}
			return NewLNEven(arg), nil
		}
	case OpcodeNop:
		if len(args) >= 1 {
			return ASTToLive(tool, args[0], con)
		}
	}

	level := SemanticLevel(expr)
	if tool != nil {
		level = tool.GetLevelBound(expr, con)
	}
	return astToLiveLevel(tool, expr, con, level)
}
