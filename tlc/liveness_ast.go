package tlc

import "strings"

//go:noinline
func ASTToLive(tool *Tool, expr SemanticNode, con *Context) (*LiveExprNode, error) {
	// Keep a real stack pointer throughout the recursive translation. Go adjusts
	// this pointer when growing the goroutine stack; it must never escape to heap.
	var stackAnchor byte
	return astToLive(tool, expr, con, &stackAnchor)
}

func astToLive(tool *Tool, expr SemanticNode, con *Context, stackAnchor *byte) (*LiveExprNode, error) {
	checkLivenessStack(stackAnchor)
	switch n := expr.(type) {
	case nil:
		panic(NewNullPointerException())
	case *LiveExprNode:
		if n == nil {
			panic(NewNullPointerException())
		}
		return n, nil
	case *LabelNode:
		if n == nil {
			panic(NewNullPointerException())
		}
		if !livenessSemanticExprNode(n.Body) {
			return nil, NewTLCRuntimeException(ECTLCLiveCannotHandleFormula, SemanticString(expr))
		}
		return astToLive(tool, n.Body, con, stackAnchor)
	case *LetInNode:
		if n == nil {
			panic(NewNullPointerException())
		}
		return astToLive(tool, n.Body, con, stackAnchor)
	case *SubstInNode:
		if n == nil {
			panic(NewNullPointerException())
		}
		con1 := con
		for _, subst := range n.Substs {
			con1 = con1.Cons(subst.GetOp(), tool.GetVal(subst.GetExpr(), con, false, DoNotRecordCostModel))
		}
		return astToLive(tool, n.GetBody(), con1, stackAnchor)
	case *APSubstInNode:
		if n == nil {
			panic(NewNullPointerException())
		}
		con1 := con
		for _, subst := range n.Substs {
			con1 = con1.Cons(subst.GetOp(), tool.GetVal(subst.GetExpr(), con, false, DoNotRecordCostModel))
		}
		return astToLive(tool, n.GetBody(), con1, stackAnchor)
	case *OpApplNode:
		return astToLiveAppl(tool, n, con, stackAnchor)
	default:
		return astToLiveLevel(tool, expr, con, SpecsGetLevel(expr, con))
	}
}

func livenessSemanticExprNode(node SemanticNode) bool {
	if isNil(node) {
		return false
	}
	switch node.(type) {
	case *OpApplNode, *LetInNode, *SubstInNode, *NumeralNode, *DecimalNode, *StringNode, *AtNode, *LabelNode, *ValueNode, *LiveExprNode:
		return true
	default:
		return false
	}
}

func astToLiveLevel(tool *Tool, expr SemanticNode, con *Context, level int) (*LiveExprNode, error) {
	if level > TLCLevelAction {
		return nil, NewTLCRuntimeException(ECTLCLiveCannotHandleFormula, SemanticString(expr))
	}
	if level == TLCLevelConstant {
		if tool == nil {
			panic(NewNullPointerException())
		}
		level = tool.GetLevelBound(expr, con)
	}
	if level == TLCLevelConstant {
		value, err := tool.Eval(expr, con, EmptyState)
		if err != nil {
			return nil, err
		}
		if boolValue, ok := value.(*BoolValue); ok {
			if boolValue.Val {
				return LNTrue, nil
			}
			return LNFalse, nil
		}
		return nil, NewTLCRuntimeException(ECTLCExpectedValue, "boolean", SemanticString(expr))
	}
	if level == TLCLevelState {
		state := NewLNState("", expr, con, nil)
		state.stringFunc = func() string { return SemanticString(expr) }
		state.dotVizFunc = func() string { return liveStateDotVizLabel(expr) }
		return state, nil
	}
	action := NewLNAction("", expr, con, nil)
	action.stringFunc = func() string { return liveActionBodyString(expr) }
	return action, nil
}

func astToLiveFallback(tool *Tool, expr SemanticNode, con *Context) (*LiveExprNode, error) {
	return astToLiveLevel(tool, expr, con, SpecsGetLevel(expr, con))
}

func livenessArgAt(args []SemanticNode, index int) SemanticNode {
	index = int(int32(index))
	if index < 0 || index >= len(args) {
		panic(NewArrayIndexOutOfBoundsException(index, len(args)))
	}
	return args[index]
}

func livenessExprNode(node SemanticNode) SemanticNode {
	if !isNil(node) && !livenessSemanticExprNode(node) {
		panic(valueStreamClassCast(node, "tla2sany.semantic.ExprNode"))
	}
	return node
}

func livenessExprArgAt(args []SemanticNode, index int) SemanticNode {
	return livenessExprNode(livenessArgAt(args, index))
}

func astToLiveITEChoice(tool *Tool, args []SemanticNode, con *Context) (chosen SemanticNode) {
	defer func() {
		if failure := recover(); failure != nil {
			exception, translated := failure.(error)
			if !translated || isJavaError(exception) {
				panic(failure)
			}
			chosen = nil
		}
	}()
	if tool == nil {
		panic(NewNullPointerException())
	}
	guard, err := tool.Eval(livenessExprArgAt(args, 0), con, EmptyState)
	if err != nil {
		if isJavaError(err) {
			panic(err)
		}
		return nil
	}
	boolGuard, ok := guard.(*BoolValue)
	if !ok {
		return nil
	}
	if boolGuard.Val {
		return livenessExprArgAt(args, 1)
	}
	return livenessExprArgAt(args, 2)
}

func astToLiveOpDefAttempt(tool *Tool, expr *OpApplNode, opDef *OpDefNode, args []SemanticNode, con *Context, stackAnchor *byte, recursive bool) (live *LiveExprNode, caught error, ok bool) {
	defer func() {
		if failure := recover(); failure != nil {
			exception, translated := failure.(error)
			if !translated || isJavaError(exception) || recursive {
				panic(failure)
			}
			live, caught, ok = nil, exception, false
		}
	}()
	formals := opDef.GetParams()
	con1 := con
	for i, arg := range args {
		if formals == nil {
			panic(NewNullPointerException())
		}
		if i >= len(formals) {
			panic(NewArrayIndexOutOfBoundsException(i, len(formals)))
		}
		con1 = con1.Cons(formals[i], tool.GetVal(arg, con, false, DoNotRecordCostModel))
	}
	live, err := astToLive(tool, opDef.GetBody(), con1, stackAnchor)
	if err != nil {
		if isJavaError(err) {
			panic(err)
		}
		return nil, err, false
	}
	level := live.GetLevel()
	if level > LiveLevelAction {
		return live, nil, true
	}
	live, err = astToLiveLevel(tool, expr, con, level)
	if err != nil {
		if isJavaError(err) {
			panic(err)
		}
		return nil, err, false
	}
	return live, nil, true
}

func astToLiveBoundedAttempt(tool *Tool, expr *OpApplNode, body SemanticNode, con *Context, stackAnchor *byte, forall bool) (live *LiveExprNode, caught error, ok bool) {
	defer func() {
		if failure := recover(); failure != nil {
			exception, translated := failure.(error)
			if !translated || isJavaError(exception) {
				panic(failure)
			}
			live, caught, ok = nil, exception, false
		}
	}()
	if tool == nil {
		panic(NewNullPointerException())
	}
	enum, err := tool.Contexts(expr, con, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel)
	if err != nil {
		if isJavaError(err) {
			panic(err)
		}
		return nil, err, false
	}
	junction := NewLNDisj()
	if forall {
		junction = NewLNConj()
	}
	for c1 := enum.NextElement(); c1 != nil; c1 = enum.NextElement() {
		if err := enum.Err(); err != nil {
			if isJavaError(err) {
				panic(err)
			}
			return nil, err, false
		}
		kid, err := astToLive(tool, body, c1, stackAnchor)
		if err != nil {
			if isJavaError(err) {
				panic(err)
			}
			return nil, err, false
		}
		if forall {
			junction.AddConj(kid)
		} else {
			junction.AddDisj(kid)
		}
	}
	if err := enum.Err(); err != nil {
		if isJavaError(err) {
			panic(err)
		}
		return nil, err, false
	}
	if junction.Count() == 0 {
		if forall {
			return LNTrue, nil, true
		}
		return LNFalse, nil, true
	}
	if junction.GetLevel() > LiveLevelAction {
		return junction, nil, true
	}
	live, err = astToLiveLevel(tool, expr, con, junction.GetLevel())
	if err != nil {
		if isJavaError(err) {
			panic(err)
		}
		return nil, err, false
	}
	return live, nil, true
}

func astToLiveFcnApply(tool *Tool, expr *OpApplNode, args []SemanticNode, con *Context, stackAnchor *byte) (live *LiveExprNode, ok bool) {
	defer func() {
		if failure := recover(); failure != nil {
			exception, translated := failure.(error)
			if !translated || isJavaError(exception) {
				panic(failure)
			}
			live, ok = nil, false
		}
	}()
	if tool == nil {
		panic(NewNullPointerException())
	}
	fval, err := tool.Eval(livenessArgAt(args, 0), con, EmptyState)
	if err != nil {
		if isJavaError(err) {
			panic(err)
		}
		return nil, false
	}
	fcn, ok := fval.(*FcnLambdaValue)
	if !ok || fcn.HasRcd() {
		return nil, false
	}
	if _, err := tool.getFcnContext(fcn, expr, con, EmptyState, EmptyState, EvalClear, DoNotRecordCostModel); err != nil {
		if isJavaError(err) {
			panic(err)
		}
		return nil, false
	}
	body := fcn.GetBody()
	if !isNil(body) && !livenessSemanticExprNode(body) {
		panic(valueStreamClassCast(body, "tla2sany.semantic.ExprNode"))
	}
	live, err = astToLive(tool, body, con, stackAnchor)
	if err != nil {
		if isJavaError(err) {
			panic(err)
		}
		return nil, false
	}
	return live, true
}

func newLiveStateEnabled(body SemanticNode, con *Context, subscript SemanticNode, isBox bool) *LiveExprNode {
	enabled := NewLNState("", body, con, func(tool *Tool, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
		if isBox && subscript != nil {
			return true, nil
		}
		if tool == nil {
			panic(NewNullPointerException())
		}
		sfun := NewFunctionalState()
		c1 := BranchContext(con)
		var state *TLCStateMut
		var err error
		if subscript != nil {
			state, err = tool.EnabledWithActionItem(body, c1, s1, sfun, subscript, ActionItemChanged)
		} else {
			state, err = tool.Enabled(body, c1, s1, sfun)
		}
		if err != nil {
			return false, err
		}
		return state != nil, nil
	})
	enabled.stringFunc = func() string { return "ENABLED " + liveActionLabel(body, subscript, isBox) }
	enabled.dotVizFunc = func() string { return liveStateDotVizLabel(body) }
	return enabled
}

func liveStateDotVizLabel(expr SemanticNode) string {
	pred, ok := expr.(*OpApplNode)
	if !ok {
		return ""
	}
	tree := pred.GetTreeNode()
	if tree == nil {
		panic(NewNullPointerException())
	}
	syntax, ok := tree.(interface {
		GetZeroHumanReadableImages() []string
		GetOneHumanReadableImages() []string
	})
	if !ok {
		panic(valueStreamClassCast(tree, "tla2sany.parser.SyntaxTreeNode"))
	}
	zero := syntax.GetZeroHumanReadableImages()
	if zero == nil {
		panic(NewNullPointerException())
	}
	return "(" + strings.Join(zero, "") +
		strings.Join(syntax.GetOneHumanReadableImages(), "") + ")"
}

func newLiveAction(body SemanticNode, con *Context, subscript SemanticNode, isBox bool) *LiveExprNode {
	action := NewLNAction("", body, con, func(tool *Tool, s1 *TLCStateMut, s2 *TLCStateMut) (bool, error) {
		if tool == nil {
			panic(NewNullPointerException())
		}
		if subscript != nil {
			v1, err := tool.Eval(subscript, con, s1, EmptyState, EvalClear)
			if err != nil {
				return false, err
			}
			v2, err := tool.Eval(subscript, con, s2, nil, EvalClear)
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
		value, err := tool.Eval(body, con, s1, s2, EvalClear)
		if err != nil {
			return false, err
		}
		boolValue, ok := value.(*BoolValue)
		if !ok {
			return false, NewTLCRuntimeException(ECTLCLiveEncounteredNonboolPredicate)
		}
		return boolValue.Val, nil
	})
	action.stringFunc = func() string { return liveActionLabel(body, subscript, isBox) }
	return action
}

func liveActionLabel(body, subscript SemanticNode, isBox bool) string {
	label := liveActionBodyString(body)
	if subscript == nil {
		return label
	}
	if isBox {
		return "[" + label + "]_" + SemanticString(subscript)
	}
	return "<" + label + ">_" + SemanticString(subscript)
}

func liveActionBodyString(body SemanticNode) string {
	switch body.(type) {
	case *ValueNode, *LiveExprNode:
		return SemanticString(body)
	}
	var label strings.Builder
	SemanticNodeAppendString(body, &label, "")
	return label.String()
}

func astToLiveAppl(tool *Tool, expr *OpApplNode, con *Context, stackAnchor *byte) (*LiveExprNode, error) {
	if expr == nil {
		panic(NewNullPointerException())
	}
	args := expr.Args
	opNode := expr.Operator
	if opNode == nil {
		panic(NewNullPointerException())
	}
	opcode := GetOpCode(opNode.GetName())

	if opcode == 0 {
		if tool == nil {
			panic(NewNullPointerException())
		}
		if val := tool.Lookup(opNode, con, EmptyState, false); val != nil {
			switch typed := val.(type) {
			case *OpDefNode:
				if typed != nil {
					opcode = GetOpCode(typed.Name)
				}
				if opcode == 0 && typed != nil {
					expand := true
					recursive := false
					if typed.GetInRecursive() {
						recursive = true
						expand = tool.GetLevelBound(expr, con) > TLCLevelAction
					}
					if expand {
						live, caught, ok := astToLiveOpDefAttempt(tool, expr, typed, args, con, stackAnchor, recursive)
						if ok {
							return live, nil
						}
						if recursive {
							return nil, caught
						}
					}
				}
			case *BoolValue:
				if typed.Val {
					return LNTrue, nil
				}
				return LNFalse, nil
			}
		}
		if opcode == 0 {
			level := SpecsGetLevel(expr, con)
			if level > TLCLevelAction {
				return nil, NewTLCRuntimeException(ECTLCLiveCannotHandleFormula, SemanticString(expr))
			}
			return astToLiveLevel(tool, expr, con, level)
		}
	}

	switch opcode {
	case OpcodeBE:
		body := livenessExprArgAt(args, 0)
		if live, _, ok := astToLiveBoundedAttempt(tool, expr, body, con, stackAnchor, false); ok {
			return live, nil
		}
		level := SpecsGetLevel(expr, con)
		if level > TLCLevelAction {
			return nil, NewTLCRuntimeException(ECTLCLiveCannotHandleFormula, SemanticString(expr))
		}
		return astToLiveLevel(tool, expr, con, level)
	case OpcodeBF:
		body := livenessExprArgAt(args, 0)
		if live, caught, ok := astToLiveBoundedAttempt(tool, expr, body, con, stackAnchor, true); ok {
			return live, nil
		} else {
			level := SpecsGetLevel(expr, con)
			if level > TLCLevelAction {
				if failure := javaRuntimeException(caught); failure != nil {
					return nil, NewTLCRuntimeException(ECTLCLiveCannotHandleFormula, SemanticString(expr), failure.Error())
				}
				return nil, NewTLCRuntimeException(ECTLCLiveCannotHandleFormula, SemanticString(expr))
			}
			return astToLiveLevel(tool, expr, con, level)
		}
	case OpcodeFA:
		if live, ok := astToLiveFcnApply(tool, expr, args, con, stackAnchor); ok {
			return live, nil
		}
		return astToLiveLevel(tool, expr, con, SemanticLevel(expr))
	case OpcodeCL, OpcodeLand:
		out := NewLNConj()
		for i := range args {
			kid, err := astToLive(tool, livenessExprArgAt(args, i), con, stackAnchor)
			if err != nil {
				return nil, err
			}
			out.AddConj(kid)
		}
		if out.GetLevel() > LiveLevelAction {
			return out, nil
		}
		return astToLiveLevel(tool, expr, con, out.GetLevel())
	case OpcodeDL, OpcodeLor:
		out := NewLNDisj()
		for i := range args {
			kid, err := astToLive(tool, livenessExprArgAt(args, i), con, stackAnchor)
			if err != nil {
				return nil, err
			}
			out.AddDisj(kid)
		}
		if out.GetLevel() > LiveLevelAction {
			return out, nil
		}
		return astToLiveLevel(tool, expr, con, out.GetLevel())
	case OpcodeITE:
		if chosen := astToLiveITEChoice(tool, args, con); !isNil(chosen) {
			return astToLive(tool, chosen, con, stackAnchor)
		}
		guard, err := astToLive(tool, livenessExprArgAt(args, 0), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		thenExpr, err := astToLive(tool, livenessExprArgAt(args, 1), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		elseExpr, err := astToLive(tool, livenessExprArgAt(args, 2), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		out := NewLNDisj(NewLNConj(guard, thenExpr), NewLNConj(NewLNNeg(guard), elseExpr))
		if out.GetLevel() > LiveLevelAction {
			return out, nil
		}
		return astToLiveLevel(tool, expr, con, out.GetLevel())
	case OpcodeLnot:
		arg, err := astToLive(tool, livenessExprArgAt(args, 0), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		if arg.GetLevel() > LiveLevelAction {
			return NewLNNeg(arg), nil
		}
		return astToLiveLevel(tool, expr, con, arg.GetLevel())
	case OpcodeImplies:
		left, err := astToLive(tool, livenessExprArgAt(args, 0), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		right, err := astToLive(tool, livenessExprArgAt(args, 1), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		level := max(left.GetLevel(), right.GetLevel())
		if level > LiveLevelAction {
			return NewLNDisj(NewLNNeg(left), right), nil
		}
		return astToLiveLevel(tool, expr, con, level)
	case OpcodePrime:
		action := NewLNAction("", expr, con, nil)
		action.stringFunc = func() string { return liveActionBodyString(expr) }
		return action, nil
	case OpcodeAA:
		return newLiveAction(livenessExprArgAt(args, 0), con, livenessExprArgAt(args, 1), false), nil
	case OpcodeSF:
		subscript := livenessExprArgAt(args, 0)
		body := livenessExprArgAt(args, 1)
		enabled := NewLNNeg(newLiveStateEnabled(body, con, subscript, false))
		action := newLiveAction(body, con, subscript, false)
		return NewLNDisj(NewLNEven(NewLNAll(enabled)), NewLNAll(NewLNEven(action))), nil
	case OpcodeWF:
		subscript := livenessExprArgAt(args, 0)
		body := livenessExprArgAt(args, 1)
		enabled := NewLNNeg(newLiveStateEnabled(body, con, subscript, false))
		action := newLiveAction(body, con, subscript, false)
		return NewLNAll(NewLNEven(NewLNDisj(enabled, action))), nil
	case OpcodeLeadsto:
		left, err := astToLive(tool, livenessExprArgAt(args, 0), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		right, err := astToLive(tool, livenessExprArgAt(args, 1), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		return NewLNAll(NewLNDisj(NewLNNeg(left), NewLNEven(right))), nil
	case OpcodeBox:
		arg, err := astToLive(tool, livenessExprArgAt(args, 0), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		return NewLNAll(arg), nil
	case OpcodeDiamond:
		arg, err := astToLive(tool, livenessExprArgAt(args, 0), con, stackAnchor)
		if err != nil {
			return nil, err
		}
		return NewLNEven(arg), nil
	case OpcodeNop:
		return astToLive(tool, livenessExprArgAt(args, 0), con, stackAnchor)
	}

	return astToLiveFallback(tool, expr, con)
}
