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
	if level > TLCLevelAction {
		return nil, newTLCError(ECTLCLiveCannotHandleFormula, "cannot handle liveness formula %s", SemanticString(expr))
	}
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

func astToLiveFallback(tool *Tool, expr SemanticNode, con *Context) (*LiveExprNode, error) {
	level := SemanticLevel(expr)
	if tool != nil {
		level = tool.GetLevelBound(expr, con)
	}
	return astToLiveLevel(tool, expr, con, level)
}

func newLiveStateEnabled(body SemanticNode, con *Context, subscript SemanticNode, isBox bool) *LiveExprNode {
	label := "ENABLED " + SemanticString(body)
	return NewLNState(label, body, con, func(tool *Tool, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
		if tool == nil {
			return true, nil
		}
		if isBox && subscript != nil {
			return true, nil
		}
		sfun := NewEmptyState()
		acts := EmptyActionItemList
		if subscript != nil {
			acts = acts.Cons(subscript, con, DoNotRecordCostModel, ActionItemChanged)
		}
		state, err := tool.EnabledImpl(body, acts, BranchContext(con), s1, sfun, DoNotRecordCostModel)
		if err != nil {
			return false, err
		}
		return state != nil, nil
	})
}

func newLiveAction(body SemanticNode, con *Context, subscript SemanticNode, isBox bool) *LiveExprNode {
	label := SemanticString(body)
	return NewLNAction(label, body, con, func(tool *Tool, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
		if tool == nil {
			return true, nil
		}
		if subscript != nil {
			v1, err := tool.Eval(subscript, con, s1, EmptyState, EvalClear, DoNotRecordCostModel)
			if err != nil {
				return false, err
			}
			v2, err := tool.Eval(subscript, con, s2, nil, EvalClear, DoNotRecordCostModel)
			if err != nil {
				return false, err
			}
			eq, err := v1.Equal(v2)
			if err != nil {
				return false, err
			}
			if isBox && eq {
				return true, nil
			}
			if !isBox && eq {
				return false, nil
			}
		}
		value, err := tool.Eval(body, con, s1, s2, EvalClear, DoNotRecordCostModel)
		if err != nil {
			return false, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return false, newTLCError(ECGeneral, "liveness predicate %s evaluated to non-boolean %s", label, value)
		}
		return boolValue.Val, nil
	})
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
	case OpcodeBE:
		if len(args) >= 1 && tool != nil {
			enum, err := tool.Contexts(expr, con, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
			if err == nil {
				out := NewLNDisj()
				ok := true
				for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
					kid, err := ASTToLive(tool, args[0], c1)
					if err != nil {
						ok = false
						break
					}
					out.AddDisj(kid)
				}
				if err := enum.Err(); err != nil {
					ok = false
				}
				if ok {
					if out.Count() == 0 {
						return LNFalse, nil
					}
					if out.GetLevel() > LiveLevelAction {
						return out, nil
					}
					return astToLiveFallback(tool, expr, con)
				}
			}
		}
	case OpcodeBF:
		if len(args) >= 1 && tool != nil {
			enum, err := tool.Contexts(expr, con, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
			if err == nil {
				out := NewLNConj()
				ok := true
				for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
					kid, err := ASTToLive(tool, args[0], c1)
					if err != nil {
						ok = false
						break
					}
					out.AddConj(kid)
				}
				if err := enum.Err(); err != nil {
					ok = false
				}
				if ok {
					if out.Count() == 0 {
						return LNTrue, nil
					}
					if out.GetLevel() > LiveLevelAction {
						return out, nil
					}
					return astToLiveFallback(tool, expr, con)
				}
			}
		}
	case OpcodeFA:
		if len(args) >= 2 && tool != nil {
			fval, err := tool.Eval(args[0], con, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
			if err == nil {
				if fcn, ok := fval.(*FcnLambdaValue); ok && fcn.FcnRcd == nil {
					if _, err := tool.getFcnContext(fcn, expr, con, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel); err == nil {
						return ASTToLive(tool, fcn.Body, con)
					}
				}
			}
		}
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
	case OpcodePrime:
		return NewLNAction(SemanticString(expr), expr, con, nil), nil
	case OpcodeAA:
		if len(args) >= 2 {
			return newLiveAction(args[0], con, args[1], false), nil
		}
	case OpcodeSF:
		if len(args) >= 2 {
			subscript := args[0]
			body := args[1]
			enabled := NewLNNeg(newLiveStateEnabled(body, con, subscript, false))
			action := newLiveAction(body, con, subscript, false)
			return NewLNDisj(NewLNEven(NewLNAll(enabled)), NewLNAll(NewLNEven(action))), nil
		}
	case OpcodeWF:
		if len(args) >= 2 {
			subscript := args[0]
			body := args[1]
			enabled := NewLNNeg(newLiveStateEnabled(body, con, subscript, false))
			action := newLiveAction(body, con, subscript, false)
			return NewLNAll(NewLNEven(NewLNDisj(enabled, action))), nil
		}
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

	return astToLiveFallback(tool, expr, con)
}
