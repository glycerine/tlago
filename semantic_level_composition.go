package tlago

// OpApplNode checks the levels of the actual arguments, including operator
// definitions and lexical LET bindings, rather than just their identifier kinds.
type sanyLevelCompositionChecker struct {
	dependencies *sanyLeibnizAnalyzer
	context      *sanyLeibnizContext
}

func newSanyLevelCompositionChecker(module *Module, spec *Spec) *sanyLevelCompositionChecker {
	dependencies := newSanyLeibnizAnalyzer(spec)
	parents := enclosingModules(spec)
	scope := dependencies.resolver.scope(module)
	for parent := parents[module]; parent != nil; parent = parents[parent] {
		for name, ref := range dependencies.resolver.scope(parent) {
			if _, local := scope[name]; !local {
				scope[name] = ref
			}
		}
	}
	return &sanyLevelCompositionChecker{dependencies: dependencies, context: &sanyLeibnizContext{module: module, formals: map[string]sanyLeibnizBinding{}}}
}

func (c *sanyLevelCompositionChecker) contextWithLocals(locals map[string]bool) *sanyLeibnizContext {
	context := sanyLeibnizNestedContext(c.context)
	for name := range locals {
		if _, definition := context.locals[name]; !definition {
			if _, bound := context.formals[name]; !bound {
				context.formals[name] = sanyLeibnizBinding{}
			}
		}
	}
	return context
}

func (c *sanyLevelCompositionChecker) level(expr Expr, locals map[string]bool) tlaLevel {
	return c.dependencies.levelInContext(expr, c.contextWithLocals(locals))
}

func (c *sanyLevelCompositionChecker) withLet(expr *LetExpr, locals map[string]bool) *sanyLevelCompositionChecker {
	context := c.contextWithLocals(locals)
	context.locals = make(map[string]sanyLeibnizLocal, len(c.context.locals)+len(expr.Definitions))
	for name, local := range c.context.locals {
		context.locals[name] = local
	}
	for i := range expr.Definitions {
		def := &expr.Definitions[i]
		context.locals[def.Name] = sanyLeibnizLocal{ref: sanySelectorDefinition{module: context.module, def: def, params: sanyDefinitionParams(def)}, context: context}
	}
	return &sanyLevelCompositionChecker{dependencies: c.dependencies, context: context}
}

func sanyOperatorApplicationKind(expr Expr) bool {
	if selected := sanyExprSelection(expr); selected != nil && !selected.operator {
		if selected.assumeProve != nil || selected.newSymbol != nil {
			return false
		}
		expr = selected.body
	}
	switch e := expr.(type) {
	case nil, *LetExpr, *LabelExpr:
		return false
	case *LiteralExpr:
		return e.Kind == "bool"
	default:
		return true
	}
}

func sanyLevelDiagnostic(diagnostic Diagnostic, expr Expr, message string) Diagnostic {
	position := expr.Position()
	if source, ok := expr.(interface{ GetSyntaxNode() *SanySyntaxNode }); ok && source.GetSyntaxNode() != nil {
		position = sanyNodePosition(source.GetSyntaxNode())
	}
	diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
	diagnostic.SANYMessage = message
	return diagnostic
}
