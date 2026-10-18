package tlc

func SpecsGetLevel(expr SemanticNode, c *Context) int {
	if isNil(expr) {
		panic(NewNullPointerException())
	}
	params := SemanticLevelParams(expr)
	if len(params) == 0 {
		return SemanticLevel(expr)
	}
	level := SemanticLevel(expr)
	for _, param := range params {
		if c == nil {
			panic(NewNullPointerException())
		}
		res := c.LookupCutoff(param, true)
		switch value := res.(type) {
		case *LazyValue:
			if value == nil {
				panic(NewNullPointerException())
			}
			plevel := SpecsGetLevel(value.ownedExpr(), value.ownedContext())
			if plevel > level {
				level = plevel
			}
		case *LazySupplierValue:
			lv := asLazyValue(value)
			plevel := SpecsGetLevel(lv.ownedExpr(), lv.ownedContext())
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
	if subs == nil {
		panic(NewNullPointerException())
	}
	res := expr
	for !subs.IsEmpty() {
		item := subs.Car()
		if isNil(item) {
			panic(NewNullPointerException())
		}
		sn, ok := item.(*SubstInNode)
		if !ok {
			panic(valueStreamClassCast(item, "tla2sany.semantic.SubstInNode"))
		}
		res = NewSubstInNodeFromSource(sn, res)
		subs = subs.Cdr()
	}
	return res
}
