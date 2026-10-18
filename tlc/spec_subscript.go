package tlc

// checkNextStateSubscript mirrors SpecProcessor's SubscriptCollector. Only
// variables reached through tuples, definitions and substitution contexts count;
// other built-in expressions do not contribute their operands.
func (p *SpecProcessor) checkNextStateSubscript(tool *Tool, expr SemanticNode, c *Context, subs *List) {
	var components []*SymbolNode
	collected := false
	func() {
		defer func() {
			if failure := recover(); failure != nil {
				err, ok := failure.(error)
				if !ok || isJavaError(err) {
					panic(failure)
				}
				// Java supplies a non-null empty array, which the elevated
				// exception retains instead of selecting the null-parameter path.
				PrintWarning(ECTLCCouldNotDetermineSubscript, []string{}...)
			}
		}()
		c1 := c
		for remaining := subs; remaining != nil && !remaining.IsEmpty(); remaining = remaining.Cdr() {
			item := remaining.Car()
			sn, ok := item.(*SubstInNode)
			if !ok {
				panic(valueStreamClassCast(item, "tla2sany.semantic.SubstInNode"))
			}
			if sn == nil {
				panic(NewNullPointerException())
			}
			for _, sub := range sn.Substs {
				c1 = c1.Cons(sub.GetOp(), tool.GetVal(sub.GetExpr(), c, false, DoNotRecordCostModel))
			}
		}
		var enter func(SemanticNode, *Context)
		enter = func(node SemanticNode, con *Context) {
			if variable := tool.GetVar(node, con, false); variable != nil {
				components = append(components, variable)
				return
			}
			switch n := node.(type) {
			case *OpApplNode:
				opcode := GetOpCode(n.Operator.GetName())
				if opcode == OpcodeTup {
					for _, arg := range n.Args {
						enter(arg, con)
					}
					return
				}
				if opcode != 0 {
					return
				}
				switch value := tool.LookupSymbolValue(n.Operator, con, false).(type) {
				case *OpDefNode:
					context, err := tool.GetOpContext(value, n.Args, con, false, DoNotRecordCostModel)
					if err != nil {
						panic(err)
					}
					enter(value.GetBody(), context)
				case *LazyValue:
					enter(value.ownedExpr(), value.ownedContext())
				case *LazySupplierValue:
					lazy := asLazyValue(value)
					enter(lazy.ownedExpr(), lazy.ownedContext())
				}
			case *SubstInNode:
				context := con
				for _, sub := range n.Substs {
					context = context.Cons(sub.GetOp(), tool.GetVal(sub.GetExpr(), con, false, DoNotRecordCostModel))
				}
				enter(n.GetBody(), context)
			case *LetInNode:
				enter(n.Body, con)
			case *LabelNode:
				enter(n.Body, con)
			default:
				panic(NewTLCRuntimeException(ECTLCCantHandleSubscript, SemanticString(node)))
			}
		}
		enter(expr, c1)
		collected = true
	}()
	if !collected {
		return
	}
	for _, variable := range p.VariablesNodes {
		found := false
		for _, component := range components {
			if component == variable {
				found = true
				break
			}
		}
		if !found {
			PrintWarning(ECTLCSubscriptContainNoStateVar, variable.GetName().String())
		}
	}
}
