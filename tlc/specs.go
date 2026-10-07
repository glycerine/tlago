package tlc

func SpecsGetLevel(expr SemanticNode, c *Context) int {
	if c == nil {
		c = EmptyContext
	}
	params := SemanticLevelParams(expr)
	if len(params) == 0 {
		return SemanticLevel(expr)
	}
	level := SemanticLevel(expr)
	for _, param := range params {
		res := c.LookupCutoff(param, true)
		switch value := res.(type) {
		case *LazyValue:
			plevel := SpecsGetLevel(value.Expr, value.Con)
			if plevel > level {
				level = plevel
			}
		case *LazySupplierValue:
			lv := asLazyValue(value)
			plevel := SpecsGetLevel(lv.Expr, lv.Con)
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
		res = NewSubstInNodeFromSource(sn, res)
		subs = subs.Cdr()
	}
	return res
}
