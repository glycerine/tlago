package tlago

import "sort"

func (a *sanyLeibnizAnalyzer) substitutionLevel(expr Expr, module *Module, locals map[string]bool) tlaLevel {
	context := &sanyLeibnizContext{module: module, formals: map[string]sanyLeibnizBinding{}}
	for name := range locals {
		context.formals[name] = sanyLeibnizBinding{}
	}
	return a.levelInContext(expr, context)
}

func (a *sanyLeibnizAnalyzer) levelInContext(expr Expr, context *sanyLeibnizContext) tlaLevel {
	a.signatures = map[sanyLeibnizDefinitionKey]*sanyLeibnizSignature{}
	a.nextID = 0
	for {
		a.changed = false
		a.evaluated = map[sanyLeibnizDefinitionKey]bool{}
		a.expressions = map[sanyLeibnizExpressionKey]sanyLeibnizUse{}
		use := a.expression(expr, context)
		if !a.changed {
			return use.level
		}
	}
}

// ModuleNode combines expression constraints with ASSUME/PROVE constraints.
// LevelNode.addTemporalLevelConstraintToConstants limits constant levelParams
// to ActionLevel. AssumeNode exposes its expression constraints through its
// getter, rather than the separate constraints written on AssumeNode itself.
func moduleTemporalConstantConstraints(module *Module, spec *Spec) map[string]tlaLevel {
	names, use := moduleSubstitutionConstraints(module, spec)
	constraints := map[string]tlaLevel{}
	for id, maximum := range use.constraints {
		if id < len(names) {
			constraints[names[id]] = maximum
		}
	}
	return constraints
}

func moduleSubstitutionConstraints(module *Module, spec *Spec) ([]string, sanyLeibnizUse) {
	var constraints sanyLeibnizUse
	if module == nil || spec == nil || isEmbeddedStandardModule(module) {
		return nil, constraints
	}
	var names []string
	targets := moduleSubstitutionTargets(module, spec)
	for name := range targets {
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil, constraints
	}
	sort.Strings(names)
	ctx := &sanyLeibnizContext{module: module, variables: map[string]sanyLeibnizBinding{}, formals: map[string]sanyLeibnizBinding{}}
	for id, name := range names {
		binding := sanyLeibnizBinding{use: sanyLeibnizUse{all: map[int]bool{id: true}, levelParams: map[int]bool{id: true}}}
		if targets[name].Kind == VariableDecl {
			binding.use.level = variableLevel
		}
		if targets[name].Arity > 0 {
			symbolID := id
			binding.operatorID = &symbolID
		}
		ctx.variables[name] = binding
	}
	a := newSanyLeibnizAnalyzer(spec)
	add := func(evaluate func() sanyLeibnizUse, temporal bool) {
		a.signatures = map[sanyLeibnizDefinitionKey]*sanyLeibnizSignature{}
		a.nextID = len(names)
		var use sanyLeibnizUse
		for {
			a.changed = false
			a.evaluated = map[sanyLeibnizDefinitionKey]bool{}
			a.expressions = map[sanyLeibnizExpressionKey]sanyLeibnizUse{}
			use = evaluate()
			if !a.changed {
				break
			}
		}
		if temporal {
			for id := range use.levelParams {
				use.constrainID(id, actionLevel)
			}
		}
		constraints.merge(use)
	}
	var expression func(Expr, *sanyLeibnizContext)
	var definition func(*Definition, *sanyLeibnizContext)
	definition = func(def *Definition, context *sanyLeibnizContext) {
		nested := sanyLeibnizNestedContext(context)
		for _, name := range def.Params {
			nested.formals[name] = sanyLeibnizBinding{}
		}
		if def.AssumeProveBody != nil {
			add(func() sanyLeibnizUse { return a.assumeProveDependencies(def.AssumeProveBody, nested) }, true)
		}
		expression(def.Expr, nested)
	}
	expression = func(expr Expr, context *sanyLeibnizContext) {
		add(func() sanyLeibnizUse { return a.expression(expr, context) }, false)
		if let, ok := expr.(*LetExpr); ok {
			nested := sanyLeibnizNestedContext(context)
			nested.locals = make(map[string]sanyLeibnizLocal, len(context.locals)+len(let.Definitions))
			for name, local := range context.locals {
				nested.locals[name] = local
			}
			for i := range let.Definitions {
				def := &let.Definitions[i]
				nested.locals[def.Name] = sanyLeibnizLocal{ref: sanySelectorDefinition{module: context.module, def: def, params: sanyDefinitionParams(def)}, context: nested}
			}
			for i := range let.Definitions {
				definition(&let.Definitions[i], nested)
			}
			expression(let.Body, nested)
			return
		}
		for _, child := range sanySubexpressionChildren(expr) {
			expression(child, context)
		}
	}
	visiting := map[*Module]bool{}
	var collect func(*Module, *sanyLeibnizContext)
	collect = func(current *Module, context *sanyLeibnizContext) {
		if current == nil || visiting[current] || isEmbeddedStandardModule(current) {
			return
		}
		visiting[current] = true
		defer delete(visiting, current)
		for _, extended := range current.Extends {
			nested := sanyLeibnizNestedContext(context)
			nested.module = spec.Modules[extended]
			collect(nested.module, nested)
		}
		for _, assumption := range current.Assumptions {
			if assumption.AssumeProveBody != nil {
				add(func() sanyLeibnizUse { return a.assumeProveDependencies(assumption.AssumeProveBody, context) }, true)
			} else {
				// AssumeNode writes its own constraints, but getLevelConstraints
				// returns the expression constraints. Only constraints inside
				// the expression propagate to ModuleNode.
				expression(assumption.Expr, context)
			}
		}
		for i := range current.Definitions {
			definition(&current.Definitions[i], context)
		}
		for _, theorem := range current.Theorems {
			if theorem.AssumeProveBody != nil {
				add(func() sanyLeibnizUse { return a.assumeProveDependencies(theorem.AssumeProveBody, context) }, true)
			} else {
				expression(theorem.Expr, context)
			}
		}
		for _, proof := range current.Proofs {
			for _, step := range proof.Steps {
				if step.AssumeProveBody != nil {
					add(func() sanyLeibnizUse { return a.assumeProveDependencies(step.AssumeProveBody, context) }, true)
				}
			}
		}
		for _, instance := range current.Instances {
			target := spec.Modules[instance.Module]
			owner := sanyLeibnizNestedContext(context)
			for _, param := range instance.Params {
				owner.formals[param] = sanyLeibnizBinding{}
			}
			variables := map[string]sanyLeibnizBinding{}
			for name := range moduleSubstitutionTargets(target, spec) {
				variables[name] = sanyLeibnizBinding{expr: &IdentExpr{Name: name}, context: owner}
			}
			for _, subst := range instanceSubstitutions(instance) {
				variables[subst.Name] = sanyLeibnizBinding{expr: subst.Expr, context: owner}
			}
			collect(target, &sanyLeibnizContext{module: target, variables: variables, formals: map[string]sanyLeibnizBinding{}})
		}
	}
	collect(module, ctx)
	// ModuleNode starts a nonconstant module's constant declarations at
	// ConstantLevel before unioning expression constraints (union takes min).
	if moduleRequiresSubstitutionLevelMatch(module, spec) {
		for id, name := range names {
			if targets[name].Kind == ConstantDecl {
				constraints.constrainID(id, constantLevel)
			}
		}
	}
	return names, constraints
}

func (a *sanyLeibnizAnalyzer) assumeProveDependencies(body *AssumeProve, context *sanyLeibnizContext) sanyLeibnizUse {
	var use sanyLeibnizUse
	if body == nil {
		return use
	}
	nested := sanyLeibnizNestedContext(context)
	for _, item := range body.Assumptions {
		switch {
		case item.NewSymbol != nil:
			use.merge(a.expression(item.NewSymbol.Domain, nested))
			binding := sanyLeibnizBinding{use: sanyLeibnizUse{level: item.NewSymbol.Level}}
			use.merge(binding.use)
			nested.formals[item.NewSymbol.Name] = binding
		case item.Nested != nil:
			use.merge(a.assumeProveDependencies(item.Nested, nested))
		default:
			use.merge(a.expression(item.Expr, nested))
		}
	}
	use.merge(a.expression(body.Prove, nested))
	return use
}
