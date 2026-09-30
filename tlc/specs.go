package tlc

func SpecsGetLevel(expr SemanticNode, c *Context) int {
	if c == nil {
		c = EmptyContext
	}
	level := SemanticLevel(expr)
	params := SemanticLevelParams(expr)
	if len(params) == 0 {
		return level
	}
	for _, param := range params {
		res := c.LookupCutoff(param, true)
		switch value := res.(type) {
		case *LazyValue:
			plevel := SpecsGetLevel(value.Expr, value.Con)
			if plevel > level {
				level = plevel
			}
		case *OpDefNode:
			plevel := SpecsGetLevel(value, c)
			if plevel > level {
				level = plevel
			}
		}
	}
	return level
}

func SpecsAddSubsts(expr SemanticNode, subs *List) SemanticNode {
	res := expr
	for subs != nil && !subs.IsEmpty() {
		sn, ok := subs.Car().(*SubstInNode)
		if !ok || sn == nil {
			panic(newTLCError(ECGeneral, "SpecsAddSubsts expected *SubstInNode, got %T", subs.Car()))
		}
		res = NewSubstInNode(res, sn.Substs...)
		subs = subs.Cdr()
	}
	return res
}
