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
	return moduleSubstitutionConstraintsInProgress(module, spec, map[*Module]bool{})
}

func moduleSubstitutionConstraintsInProgress(module *Module, spec *Spec, active map[*Module]bool) ([]string, sanyLeibnizUse) {
	var constraints sanyLeibnizUse
	if module == nil || spec == nil || active[module] || isEmbeddedStandardModule(module) {
		return nil, constraints
	}
	active[module] = true
	defer delete(active, module)
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
	add := func(evaluate func() sanyLeibnizUse, temporal bool, result *sanyLeibnizUse) {
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
		result.merge(use)
	}
	var expression func(Expr, *sanyLeibnizContext, *sanyLeibnizUse)
	var instanceConstraints func(Instance, *sanyLeibnizContext, *sanyLeibnizUse)
	var definition func(*Definition, *sanyLeibnizContext, *sanyLeibnizUse)
	definition = func(def *Definition, context *sanyLeibnizContext, result *sanyLeibnizUse) {
		nested := sanyLeibnizNestedContext(context)
		for _, name := range def.Params {
			nested.formals[name] = sanyLeibnizBinding{}
		}
		if def.AssumeProveBody != nil {
			add(func() sanyLeibnizUse { return a.assumeProveDependencies(def.AssumeProveBody, nested) }, true, result)
		}
		expression(def.Expr, nested, result)
	}
	expression = func(expr Expr, context *sanyLeibnizContext, result *sanyLeibnizUse) {
		add(func() sanyLeibnizUse { return a.expression(expr, context) }, false, result)
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
				definition(&let.Definitions[i], nested, result)
			}
			expression(let.Body, nested, result)
			for _, imported := range let.instanceDefinitions {
				add(func() sanyLeibnizUse { return a.substitutedDefinitionConstraints(imported, nested) }, false, result)
			}
			for _, instance := range let.Instances {
				var instanceUse sanyLeibnizUse
				instanceConstraints(instance, nested, &instanceUse)
				// LetInNode imports only InstanceNode's ArgLevelParams;
				// its scalar/argument constraints come from body/opDefs.
				for _, key := range instanceUse.argParamOrder {
					result.addArgumentParameter(key)
				}
			}
			return
		}
		for _, child := range sanySubexpressionChildren(expr) {
			expression(child, context, result)
		}
	}
	visiting := map[*Module]bool{}
	var collect func(*Module, *sanyLeibnizContext, bool, *sanyLeibnizUse)
	collect = func(current *Module, context *sanyLeibnizContext, imported bool, result *sanyLeibnizUse) {
		if current == nil || visiting[current] || isEmbeddedStandardModule(current) {
			return
		}
		visiting[current] = true
		defer delete(visiting, current)
		// A nonconstant ModuleNode initializes its constant constraints to
		// ConstantLevel. Translate those declaration parameters through the
		// current substitution context for each retained InstanceNode too.
		// EXTENDS merges exported nodes, not the extendee's aggregate bounds.
		if !imported && moduleRequiresSubstitutionLevelMatch(current, spec) {
			var constants []string
			for name, target := range moduleSubstitutionTargets(current, spec) {
				if target.Kind == ConstantDecl {
					constants = append(constants, name)
				}
			}
			sort.Strings(constants)
			for _, name := range constants {
				add(func() sanyLeibnizUse {
					use := a.expression(&IdentExpr{Name: name}, context)
					use.constrain(use.levelParams, constantLevel)
					return use
				}, false, result)
			}
		}
		for _, extended := range current.Extends {
			nested := sanyLeibnizNestedContext(context)
			nested.module = spec.Modules[extended]
			collect(nested.module, nested, true, result)
		}
		for _, assumption := range current.Assumptions {
			if assumption.AssumeProveBody != nil {
				add(func() sanyLeibnizUse { return a.assumeProveDependencies(assumption.AssumeProveBody, context) }, true, result)
			} else {
				// AssumeNode writes its own constraints, but getLevelConstraints
				// returns the expression constraints. Only constraints inside
				// the expression propagate to ModuleNode.
				expression(assumption.Expr, context, result)
			}
		}
		for i := range current.Definitions {
			// EXTENDS imports the Context's exported OpDefNodes. A LOCAL
			// definition contributes to its own module, not an importer's
			// aggregate unless an exported definition depends on its body.
			if imported && current.Definitions[i].Local {
				continue
			}
			definition(&current.Definitions[i], context, result)
		}
		for _, theorem := range current.Theorems {
			if theorem.AssumeProveBody != nil {
				add(func() sanyLeibnizUse { return a.assumeProveDependencies(theorem.AssumeProveBody, context) }, true, result)
			} else {
				expression(theorem.Expr, context, result)
			}
		}
		for _, proof := range current.Proofs {
			for _, step := range proof.Steps {
				if step.AssumeProveBody != nil {
					add(func() sanyLeibnizUse { return a.assumeProveDependencies(step.AssumeProveBody, context) }, true, result)
				}
			}
		}
		for _, instance := range current.Instances {
			instanceConstraints(instance, context, result)
		}
	}
	instanceConstraints = func(instance Instance, context *sanyLeibnizContext, result *sanyLeibnizUse) {
		target := spec.Modules[instance.Module]
		owner := sanyLeibnizNestedContext(context)
		for _, param := range instance.Params {
			owner.formals[param] = sanyLeibnizBinding{}
		}
		for _, subst := range instance.generatedSubstitutions {
			// InstanceNode merges every substitution expression's own
			// constraints, even when the target declaration is unused.
			// Keep the actual resolved default/WITH array and its order.
			expression(subst.expr, owner, result)
		}
		sourceNames, sourceUse := moduleSubstitutionConstraintsInProgress(target, spec, active)
		add(func() sanyLeibnizUse {
			return a.substitutedConstraints(sourceUse, sourceNames, moduleSubstitutionTargets(target, spec), instance, context)
		}, false, result)
	}
	collect(module, ctx, false, &constraints)
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

// substitutedDefinitionConstraints follows Subst.getSubLCSet/getSubALCSet/
// getSubALPSet for a retained imported OpDef. Evaluate the source body with
// declaration identities before translating them; evaluating WITH inline loses
// the original declaration kind needed by nonconstant-module level matching.
func (a *sanyLeibnizAnalyzer) substitutedDefinitionConstraints(ref sanySelectorDefinition, caller *sanyLeibnizContext) sanyLeibnizUse {
	if len(ref.wrappers) == 0 {
		return a.definition(ref, nil, caller, nil)
	}
	wrapper := ref.wrappers[0]
	target := a.resolver.spec.Modules[wrapper.inst.Module]
	targets := moduleSubstitutionTargets(target, a.resolver.spec)
	names := make([]string, 0, len(targets))
	for name := range targets {
		names = append(names, name)
	}
	sort.Strings(names)
	symbolic := &sanyLeibnizContext{module: target, variables: map[string]sanyLeibnizBinding{}, formals: map[string]sanyLeibnizBinding{}}
	for id, name := range names {
		binding := sanyLeibnizBinding{use: sanyLeibnizUse{all: map[int]bool{id: true}, levelParams: map[int]bool{id: true}}}
		if targets[name].Kind == VariableDecl {
			binding.use.level = variableLevel
		}
		if targets[name].Arity > 0 {
			operatorID := id
			binding.operatorID = &operatorID
		}
		symbolic.variables[name] = binding
	}
	body := ref
	body.wrappers = ref.wrappers[1:]
	body.params = ref.params[len(wrapper.inst.Params):]
	source := newSanyLeibnizAnalyzer(a.resolver.spec)
	source.signatures = map[sanyLeibnizDefinitionKey]*sanyLeibnizSignature{}
	source.nextID = len(names)
	var use sanyLeibnizUse
	for {
		source.changed = false
		source.evaluated = map[sanyLeibnizDefinitionKey]bool{}
		source.expressions = map[sanyLeibnizExpressionKey]sanyLeibnizUse{}
		use = source.substitutedDefinitionConstraints(body, symbolic)
		if !source.changed {
			break
		}
	}
	return a.substitutedConstraints(use, names, targets, wrapper.inst, caller)
}

// Subst's constraint transformations preserve symbolic declaration identity
// until each source constraint is translated through the actual substitution.
func (a *sanyLeibnizAnalyzer) substitutedConstraints(use sanyLeibnizUse, names []string, targets map[string]substitutionTarget, instance Instance, caller *sanyLeibnizContext) sanyLeibnizUse {
	ids := map[string]int{}
	for id, name := range names {
		ids[name] = id
	}
	owner := sanyLeibnizNestedContext(caller)
	for _, name := range instance.Params {
		owner.formals[name] = sanyLeibnizBinding{}
	}
	actuals := map[int]sanyLeibnizUse{}
	operators := map[int]*int{}
	expressions := map[int]Expr{}
	for _, sub := range instance.generatedSubstitutions {
		id, exists := ids[sub.name]
		if !exists {
			continue
		}
		actuals[id] = a.expression(sub.expr, owner)
		operators[id] = a.substitutionParameterOperator(sub.expr, owner)
		expressions[id] = sub.expr
	}
	params := func(id int) map[int]bool {
		if actual, exists := actuals[id]; exists {
			return actual.levelParams
		}
		return nil
	}
	var result sanyLeibnizUse
	nonconstant := moduleRequiresSubstitutionLevelMatch(a.resolver.spec.Modules[instance.Module], a.resolver.spec)
	for id, maximum := range use.constraints {
		if id >= len(names) {
			continue // OpDef formals are not parameters of this instantiation.
		}
		if nonconstant {
			switch targets[names[id]].Kind {
			case ConstantDecl:
				maximum = constantLevel
			case VariableDecl:
				maximum = variableLevel
			}
		}
		result.constrain(params(id), maximum)
	}
	for key, minimum := range use.argConstraints {
		if operator := operators[key.operator]; operator != nil {
			result.requireArgument(sanyArgumentPosition{*operator, key.position}, minimum)
		}
	}
	for _, key := range use.argParamOrder {
		if operator := operators[key.operator]; operator != nil {
			if actual, exists := actuals[key.parameter]; exists {
				result.requireArgument(sanyArgumentPosition{*operator, key.position}, actual.level)
			}
			for parameter := range params(key.parameter) {
				result.addArgumentParameter(sanyArgumentParameter{*operator, key.position, parameter})
			}
		} else if expr := expressions[key.operator]; expr != nil && key.operator < len(names) {
			// Subst.getSubLCSet translates a substituted OpDef's maximum
			// into scalar constraints on its co-parameter's substitution.
			maximumChecker := newSanyLeibnizAnalyzer(a.resolver.spec)
			maximums := maximumChecker.applicationMaximums(expr, nil, targets[names[key.operator]].Arity, owner)
			if key.position < len(maximums) {
				result.constrain(params(key.parameter), maximums[key.position])
			}
		}
	}
	return result
}

func (a *sanyLeibnizAnalyzer) substitutionParameterOperator(expr Expr, context *sanyLeibnizContext) *int {
	ident, ok := expr.(*IdentExpr)
	if !ok || sanyExprSelection(expr) != nil {
		return nil
	}
	binding, bound := context.formals[ident.Name]
	if !bound {
		binding, bound = context.variables[ident.Name]
	}
	if !bound {
		return nil
	}
	if binding.operatorID != nil {
		return binding.operatorID
	}
	if binding.expr != nil {
		return a.substitutionParameterOperator(binding.expr, binding.context)
	}
	return nil
}

func (a *sanyLeibnizAnalyzer) instanceConstraintUse(instance Instance, context *sanyLeibnizContext) sanyLeibnizUse {
	target := a.resolver.spec.Modules[instance.Module]
	names, source := moduleSubstitutionConstraints(target, a.resolver.spec)
	result := a.substitutedConstraints(source, names, moduleSubstitutionTargets(target, a.resolver.spec), instance, context)
	owner := sanyLeibnizNestedContext(context)
	for _, name := range instance.Params {
		owner.formals[name] = sanyLeibnizBinding{}
	}
	for _, sub := range instance.generatedSubstitutions {
		result.mergeConstraints(a.expression(sub.expr, owner))
	}
	return result
}
