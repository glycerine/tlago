package tlago

import "fmt"

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
		context.locals[def.Name] = sanyLeibnizLocal{recursive: letRecursiveNames(expr)[def.Name], ref: sanySelectorDefinition{module: context.module, def: def, params: sanyDefinitionParams(def)}, context: context}
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

// OpDefNode takes each formal's maximum from its body's levelConstraints;
// OpApplNode compares the actual argument level with that maximum. Analyze the
// source definition with symbolic arguments, including INSTANCE-prepended ones.
func (c *sanyLevelCompositionChecker) checkApplicationLevels(expr Expr, locals map[string]bool) Diagnostics {
	var operator Expr
	var arguments []Expr
	var selected *sanySelectorSelection
	name := "LAMBDA"
	if selection := sanyExprSelection(expr); selection != nil && !selection.operator {
		selected = selection
		arguments = selection.args
		name = selection.name
	} else if call, ok := expr.(*CallExpr); ok {
		operator, arguments = call.Callee, call.Args
		if id, ok := operator.(*IdentExpr); ok {
			name = id.Name
			if _, builtin := sanyBuiltinOperatorInfo(name); builtin {
				return nil
			}
		}
	} else {
		return nil
	}
	if len(arguments) == 0 {
		return nil
	}
	context := c.contextWithLocals(locals)
	maximums := c.dependencies.applicationMaximums(operator, selected, len(arguments), context)
	var diags Diagnostics
	for i, argument := range arguments {
		if c.dependencies.levelInContext(argument, context) > maximums[i] {
			diagnostic := sanyDiagnosticParameters(errorAt(expr.Position(), "E4205", "operator %s argument %d exceeds maximum level %d", name, i+1, maximums[i]), name, i+1)
			message := fmt.Sprintf("Level error in applying operator %s:\nThe level of argument %d exceeds the maximum level allowed by the operator.", name, i+1)
			diags = append(diags, sanyLevelDiagnostic(diagnostic, expr, message))
		}
	}
	return diags
}

func (a *sanyLeibnizAnalyzer) applicationMaximums(operator Expr, selected *sanySelectorSelection, arity int, context *sanyLeibnizContext) []tlaLevel {
	arguments := make([]sanyLeibnizBinding, arity)
	for i := range arguments {
		arguments[i].use = sanyLeibnizUse{all: map[int]bool{i: true}, levelParams: map[int]bool{i: true}}
	}
	a.signatures = map[sanyLeibnizDefinitionKey]*sanyLeibnizSignature{}
	a.nextID = arity
	var use sanyLeibnizUse
	for {
		a.changed = false
		a.evaluated = map[sanyLeibnizDefinitionKey]bool{}
		a.expressions = map[sanyLeibnizExpressionKey]sanyLeibnizUse{}
		if selected != nil {
			use = a.definitionBody(selected.definition, selected.body, selected.params, arguments, context, nil)
		} else {
			use = a.apply(operator, arguments, context)
		}
		if !a.changed {
			break
		}
	}
	maximums := make([]tlaLevel, arity)
	for i := range maximums {
		maximums[i] = temporalLevel
		if maximum, constrained := use.constraints[i]; constrained {
			maximums[i] = maximum
		}
	}
	return maximums
}

// ModuleNode forbids priming recursive formal arguments, not free variables in
// recursive bodies. The body's propagated maximum captures indirect priming.
func (c *sanyLevelCompositionChecker) checkRecursiveParameters(def Definition, locals map[string]bool) Diagnostics {
	maximums := c.dependencies.applicationMaximums(&IdentExpr{Name: def.Name}, nil, len(def.Params), c.contextWithLocals(locals))
	var diags Diagnostics
	for i, maximum := range maximums {
		if maximum >= actionLevel {
			continue
		}
		message := fmt.Sprintf("Argument %d of recursive operator %s is primed", i+1, def.Name)
		diagnostic := sanyDiagnosticParameters(errorAt(def.Pos, "E4290", "%s", message), i+1, def.Name)
		position := def.SourcePosition()
		if def.Syntax != nil {
			position = sanyNodePosition(def.Syntax)
		}
		diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
		diagnostic.SANYMessage = message
		diags = append(diags, diagnostic)
	}
	return diags
}
