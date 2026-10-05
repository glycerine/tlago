package tlago

import "sort"

func (a *sanyLeibnizAnalyzer) substitutionLevel(expr Expr, module *Module, locals map[string]bool) tlaLevel {
	context := &sanyLeibnizContext{module: module, formals: map[string]sanyLeibnizBinding{}}
	for name := range locals {
		context.formals[name] = sanyLeibnizBinding{}
	}
	a.signatures = map[sanyLeibnizDefinitionKey]*sanyLeibnizSignature{}
	a.nextID = 0
	for {
		a.changed = false
		use := a.expression(expr, context)
		if !a.changed {
			return use.level
		}
	}
}

// LevelNode.addTemporalLevelConstraintToConstants constrains ConstantDecl
// symbols in an ASSUME or ASSUME/PROVE's levelParams to ActionLevel. These
// constraints also apply when InstanceNode's instancee is a constant module.
func moduleTemporalConstantConstraints(module *Module, spec *Spec) map[string]tlaLevel {
	constraints := map[string]tlaLevel{}
	if module == nil || spec == nil || isEmbeddedStandardModule(module) {
		return constraints
	}
	var names []string
	for name, target := range moduleSubstitutionTargets(module, spec) {
		if target.Kind == ConstantDecl {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return constraints
	}
	sort.Strings(names)
	ctx := &sanyLeibnizContext{module: module, variables: map[string]sanyLeibnizBinding{}, formals: map[string]sanyLeibnizBinding{}}
	for id, name := range names {
		ctx.variables[name] = sanyLeibnizBinding{use: sanyLeibnizUse{all: map[int]bool{id: true}, levelParams: map[int]bool{id: true}}}
	}
	a := newSanyLeibnizAnalyzer(spec)
	add := func(evaluate func() sanyLeibnizUse) {
		a.signatures = map[sanyLeibnizDefinitionKey]*sanyLeibnizSignature{}
		a.nextID = len(names)
		var use sanyLeibnizUse
		for {
			a.changed = false
			use = evaluate()
			if !a.changed {
				break
			}
		}
		for id := range use.levelParams {
			if id < len(names) {
				constraints[names[id]] = actionLevel
			}
		}
	}
	var expression func(Expr, *sanyLeibnizContext)
	var definition func(*Definition, *sanyLeibnizContext)
	definition = func(def *Definition, context *sanyLeibnizContext) {
		nested := sanyLeibnizNestedContext(context)
		for _, name := range def.Params {
			nested.formals[name] = sanyLeibnizBinding{}
		}
		if def.AssumeProveBody != nil {
			add(func() sanyLeibnizUse { return a.assumeProveDependencies(def.AssumeProveBody, nested) })
		}
		expression(def.Expr, nested)
	}
	expression = func(expr Expr, context *sanyLeibnizContext) {
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
				add(func() sanyLeibnizUse { return a.assumeProveDependencies(assumption.AssumeProveBody, context) })
			} else {
				add(func() sanyLeibnizUse { return a.expression(assumption.Expr, context) })
			}
		}
		for i := range current.Definitions {
			definition(&current.Definitions[i], context)
		}
		for _, theorem := range current.Theorems {
			if theorem.AssumeProveBody != nil {
				add(func() sanyLeibnizUse { return a.assumeProveDependencies(theorem.AssumeProveBody, context) })
			}
		}
		for _, proof := range current.Proofs {
			for _, step := range proof.Steps {
				if step.AssumeProveBody != nil {
					add(func() sanyLeibnizUse { return a.assumeProveDependencies(step.AssumeProveBody, context) })
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
	return constraints
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
