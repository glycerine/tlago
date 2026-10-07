package tlago

import (
	"fmt"
	"strings"
)

// OpDefNode computes isLeibniz from the formal parameters in its body's
// nonLeibnizParams. OpApplNode propagates all/non-Leibniz dependencies through
// arguments, and SubstInNode replaces a source dependency with its substitution.
// Keep those dependencies separate from the occurrence of a prime anywhere in
// the expression: priming a free state variable does not prime an argument.
type sanyLeibnizUse struct {
	all         map[int]bool
	non         map[int]bool
	levelParams map[int]bool
	level       tlaLevel
	constraints map[int]tlaLevel
}

func (u *sanyLeibnizUse) merge(v sanyLeibnizUse) {
	u.level = maxTlaLevel(u.level, v.level)
	if u.all == nil {
		u.all = map[int]bool{}
	}
	if u.non == nil {
		u.non = map[int]bool{}
	}
	for id := range v.all {
		u.all[id] = true
	}
	for id := range v.non {
		u.non[id] = true
	}
	u.mergeLevelParams(v)
	for id, maximum := range v.constraints {
		u.constrainID(id, maximum)
	}
}

func (u *sanyLeibnizUse) constrain(params map[int]bool, maximum tlaLevel) {
	if maximum >= temporalLevel || len(params) == 0 {
		return
	}
	if u.constraints == nil {
		u.constraints = map[int]tlaLevel{}
	}
	for id := range params {
		u.constrainID(id, maximum)
	}
}

func (u *sanyLeibnizUse) constrainID(id int, maximum tlaLevel) {
	if maximum >= temporalLevel {
		return
	}
	if u.constraints == nil {
		u.constraints = map[int]tlaLevel{}
	}
	if previous, constrained := u.constraints[id]; !constrained || maximum < previous {
		u.constraints[id] = maximum
	}
}
func (u *sanyLeibnizUse) mergeLevelParams(v sanyLeibnizUse) {
	u.level = maxTlaLevel(u.level, v.level)
	if len(v.levelParams) > 0 && u.levelParams == nil {
		u.levelParams = map[int]bool{}
	}
	for id := range v.levelParams {
		u.levelParams[id] = true
	}
}
func (u *sanyLeibnizUse) restrict() {
	// Do not mutate an argument node's cached dependencies when the parent
	// operator places its allParams into the parent's nonLeibnizParams.
	non := make(map[int]bool, len(u.non)+len(u.all))
	for id := range u.non {
		non[id] = true
	}
	for id := range u.all {
		non[id] = true
	}
	u.non = non
}

type sanyLeibnizBinding struct {
	expr    Expr
	context *sanyLeibnizContext
	use     sanyLeibnizUse
}
type sanyLeibnizLocal struct {
	recursive bool
	ref       sanySelectorDefinition
	context   *sanyLeibnizContext
}
type sanyLeibnizContext struct {
	atUse     sanyLeibnizUse
	hasAt     bool
	module    *Module
	variables map[string]sanyLeibnizBinding
	formals   map[string]sanyLeibnizBinding
	locals    map[string]sanyLeibnizLocal
}
type sanyLeibnizSignature struct {
	ids       []int
	non       []bool
	weights   []bool
	maxLevels []tlaLevel
	free      sanyLeibnizUse
}
type sanyLeibnizDefinitionKey struct {
	definition *Definition
	body       Expr
	instances  string
	operators  string
}
type sanyLeibnizAnalyzer struct {
	resolver    *sanySelectorResolver
	active      map[sanyLeibnizDefinitionKey]bool
	evaluated   map[sanyLeibnizDefinitionKey]bool
	expressions map[sanyLeibnizExpressionKey]sanyLeibnizUse
	signatures  map[sanyLeibnizDefinitionKey]*sanyLeibnizSignature
	nextID      int
	changed     bool
	declKinds   map[*Module]map[string]DeclarationKind
}

type sanyLeibnizExpressionKey struct {
	expr    Expr
	context *sanyLeibnizContext
}

func newSanyLeibnizAnalyzer(spec *Spec) *sanyLeibnizAnalyzer {
	return &sanyLeibnizAnalyzer{
		resolver:  &sanySelectorResolver{spec: spec, scopes: map[*Module]map[string]sanySelectorDefinition{}, visiting: map[*Module]bool{}},
		active:    map[sanyLeibnizDefinitionKey]bool{},
		declKinds: map[*Module]map[string]DeclarationKind{},
	}
}
func (a *sanyLeibnizAnalyzer) operatorNonLeibniz(expr Expr, module *Module, arity int, locals map[string]bool) bool {
	ctx := &sanyLeibnizContext{module: module, formals: map[string]sanyLeibnizBinding{}}
	for name := range locals {
		ctx.formals[name] = sanyLeibnizBinding{}
	}
	return a.operatorNonLeibnizInContext(expr, arity, ctx)
}

func (a *sanyLeibnizAnalyzer) operatorNonLeibnizInContext(expr Expr, arity int, ctx *sanyLeibnizContext) bool {
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
		use = a.apply(expr, arguments, ctx)
		if !a.changed {
			break
		}
	}
	for id := range use.non {
		if id < arity {
			return true
		}
	}
	return false
}
func (a *sanyLeibnizAnalyzer) binding(b sanyLeibnizBinding) sanyLeibnizUse {
	if b.expr != nil {
		var use sanyLeibnizUse
		use.merge(b.use)
		use.merge(a.expression(b.expr, b.context))
		return use
	}
	return b.use
}
func (a *sanyLeibnizAnalyzer) argumentUses(arguments []sanyLeibnizBinding) sanyLeibnizUse {
	var use sanyLeibnizUse
	for _, argument := range arguments {
		use.merge(a.binding(argument))
	}
	return use
}
func (a *sanyLeibnizAnalyzer) apply(operator Expr, arguments []sanyLeibnizBinding, ctx *sanyLeibnizContext) sanyLeibnizUse {
	use := a.argumentUses(arguments)
	use.levelParams = nil
	use.level = constantLevel
	if selected := sanyExprSelection(operator); selected != nil && selected.operator {
		ref := selected.definition
		// The selected body is part of the signature key, without allocating
		// a fresh definition identity during each fixed-point iteration.
		use.merge(a.definitionBody(ref, selected.body, selected.params, arguments, ctx, nil))
		return use
	}
	switch op := operator.(type) {
	case *IdentExpr:
		if binding, ok := ctx.formals[op.Name]; ok {
			if binding.expr != nil {
				use.merge(binding.use)
				use.merge(a.apply(binding.expr, arguments, binding.context))
			} else {
				use.merge(binding.use)
				use.mergeLevelParams(a.argumentUses(arguments))
			}
			return use
		}
		if binding, ok := ctx.variables[op.Name]; ok {
			if binding.expr != nil {
				use.merge(binding.use)
				use.merge(a.apply(binding.expr, arguments, binding.context))
			} else {
				use.merge(binding.use)
				use.mergeLevelParams(a.argumentUses(arguments))
			}
			return use
		}
		if info, ok := sanyBuiltinOperatorInfo(op.Name); ok {
			use.level = info.level
			for i, argument := range arguments {
				if maximum, constrained := builtinArgMaxLevel(info, i); constrained {
					use.constrain(a.binding(argument).levelParams, maximum)
				}
				index := i
				if info.arity < 0 {
					index = 0
				}
				if index < len(info.argWeights) && info.argWeights[index] == 0 {
					argUse := a.binding(argument)
					argUse.restrict()
					argUse.levelParams = nil
					argUse.level = constantLevel
					use.merge(argUse)
				} else {
					use.mergeLevelParams(a.binding(argument))
				}
			}
			return use
		}
		if local, ok := ctx.locals[op.Name]; ok {
			use.merge(a.definition(local.ref, arguments, ctx, local.context))
			return use
		}
		if ref, ok := a.resolver.scope(ctx.module)[op.Name]; ok {
			use.merge(a.definition(ref, arguments, ctx, nil))
		} else {
			kinds := a.declKinds[ctx.module]
			if kinds == nil {
				kinds = moduleLevelDeclKinds(ctx.module, a.resolver.spec, nil)
				a.declKinds[ctx.module] = kinds
			}
			// OpApplNode's OpDecl branch preserves every operand's level and
			// level parameters; a declared operator has no argument weights.
			if kinds[op.Name] == ConstantDecl || kinds[op.Name] == VariableDecl {
				use.mergeLevelParams(a.argumentUses(arguments))
			}
			if kinds[op.Name] == VariableDecl {
				use.level = maxTlaLevel(use.level, variableLevel)
			}
		}
	case *FunctionExpr:
		if op.IsLambda {
			nested := sanyLeibnizNestedContext(ctx)
			for i, bound := range op.Bounds {
				if i < len(arguments) {
					nested.formals[bound.Name] = arguments[i]
				}
			}
			use.merge(a.expression(op.Body, nested))
		}
	}
	return use
}
func (a *sanyLeibnizAnalyzer) definition(ref sanySelectorDefinition, arguments []sanyLeibnizBinding, caller, lexical *sanyLeibnizContext) sanyLeibnizUse {
	return a.definitionBody(ref, ref.def.Expr, ref.params, arguments, caller, lexical)
}

func sanyDefinitionIsRecursive(ref sanySelectorDefinition, lexical *sanyLeibnizContext) bool {
	if lexical != nil {
		if local, ok := lexical.locals[ref.def.Name]; ok && local.ref.def == ref.def {
			return local.recursive
		}
	}
	if ref.module != nil {
		for _, decl := range ref.module.Declarations {
			if decl.Kind == RecursiveDecl {
				for _, name := range decl.Names {
					if name == ref.def.Name {
						return true
					}
				}
			}
		}
	}
	return false
}

// OpDefNode's formal dependencies are computed independently of the actual
// scalar values supplied at a use site. Recursive summaries grow monotonically,
// as in its levelCheck iterations, before mapping formals to those actuals.
func (a *sanyLeibnizAnalyzer) definitionBody(ref sanySelectorDefinition, body Expr, params []BoundVar, actuals []sanyLeibnizBinding, caller, lexical *sanyLeibnizContext) sanyLeibnizUse {
	key := sanyLeibnizDefinitionKey{definition: ref.def, body: body}
	var instances, operators []string
	for _, wrapper := range ref.wrappers {
		instances = append(instances, wrapper.owner.Name+"!"+wrapper.inst.SourcePosition().String()+"!"+wrapper.inst.Name)
	}
	for i, param := range params {
		if param.OperatorArity > 0 && i < len(actuals) {
			operators = append(operators, a.operatorIdentity(actuals[i]))
		}
	}
	key.instances = strings.Join(instances, "/")
	key.operators = strings.Join(operators, "/")
	signature := a.signatures[key]
	if signature == nil {
		signature = &sanyLeibnizSignature{ids: make([]int, len(params)), non: make([]bool, len(params)), weights: make([]bool, len(params)), maxLevels: make([]tlaLevel, len(params))}
		recursive := body == ref.def.Expr && sanyDefinitionIsRecursive(ref, lexical)
		firstFormal := 0
		for _, wrapper := range ref.wrappers {
			firstFormal += len(wrapper.inst.Params)
		}
		for i := range params {
			signature.maxLevels[i] = temporalLevel
			if recursive && i >= firstFormal && i < firstFormal+len(ref.def.Params) {
				// ModuleNode initializes recursive formals to ActionLevel and weight 1.
				signature.maxLevels[i] = actionLevel
				signature.weights[i] = true
			}
			signature.ids[i] = a.nextID
			a.nextID++
		}
		a.signatures[key] = signature
		a.changed = true
	}
	if a.active[key] || a.evaluated[key] {
		return a.signatureUse(signature, actuals)
	}
	a.evaluated[key] = true
	a.active[key] = true
	defer delete(a.active, key)
	arguments := make([]sanyLeibnizBinding, len(params))
	for i, param := range params {
		if param.OperatorArity > 0 && i < len(actuals) {
			arguments[i] = actuals[i]
		}
		arguments[i].use = sanyLeibnizUse{all: map[int]bool{signature.ids[i]: true}, levelParams: map[int]bool{signature.ids[i]: true}}
	}
	ctx := &sanyLeibnizContext{module: ref.module, variables: caller.variables, formals: map[string]sanyLeibnizBinding{}}
	if lexical != nil {
		ctx = sanyLeibnizNestedContext(lexical)
	}
	offset := 0
	for _, wrapper := range ref.wrappers {
		owner := &sanyLeibnizContext{module: wrapper.owner, variables: ctx.variables, formals: map[string]sanyLeibnizBinding{}, locals: ctx.locals}
		// A LET instance can capture its enclosing formals. A module instance
		// has only the formals explicitly prepended to its exported definitions.
		moduleInstance := false
		for _, inst := range wrapper.owner.Instances {
			if inst.SourcePosition() == wrapper.inst.SourcePosition() {
				moduleInstance = true
				break
			}
		}
		if !moduleInstance {
			owner.formals = copySanyLeibnizBindings(caller.formals)
		}
		for _, name := range wrapper.inst.Params {
			if offset < len(arguments) {
				owner.formals[name] = arguments[offset]
			}
			offset++
		}
		variables := map[string]sanyLeibnizBinding{}
		target := a.resolver.spec.Modules[wrapper.inst.Module]
		for name := range moduleSubstitutionTargets(target, a.resolver.spec) {
			variables[name] = sanyLeibnizBinding{expr: &IdentExpr{Name: name}, context: owner}
		}
		for _, subst := range instanceSubstitutions(wrapper.inst) {
			variables[subst.Name] = sanyLeibnizBinding{expr: subst.Expr, context: owner}
		}
		ctx = &sanyLeibnizContext{module: target, variables: variables, formals: map[string]sanyLeibnizBinding{}}
	}
	ctx.module = ref.module
	for _, name := range ref.def.Params {
		if offset < len(arguments) {
			ctx.formals[name] = arguments[offset]
		}
		offset++
	}
	// Label/bound parameters added by an operator selector are also formals.
	for offset < len(params) && offset < len(arguments) {
		ctx.formals[params[offset].Name] = arguments[offset]
		offset++
	}
	use := a.expression(body, ctx)
	if body == ref.def.Expr && ref.def.AssumeProveBody != nil {
		use = a.assumeProveDependencies(ref.def.AssumeProveBody, ctx)
	}
	own := map[int]bool{}
	for i, id := range signature.ids {
		own[id] = true
		if maximum, constrained := use.constraints[id]; constrained && maximum < signature.maxLevels[i] {
			signature.maxLevels[i] = maximum
			a.changed = true
		}
		if use.levelParams[id] && !signature.weights[i] {
			signature.weights[i] = true
			a.changed = true
		}
		if use.non[id] && !signature.non[i] {
			signature.non[i] = true
			a.changed = true
		}
	}
	for id := range use.all {
		if !own[id] && !signature.free.all[id] {
			if signature.free.all == nil {
				signature.free.all = map[int]bool{}
			}
			signature.free.all[id] = true
			a.changed = true
		}
	}
	for id := range use.non {
		if !own[id] && !signature.free.non[id] {
			if signature.free.non == nil {
				signature.free.non = map[int]bool{}
			}
			signature.free.non[id] = true
			a.changed = true
		}
	}
	if use.level > signature.free.level {
		signature.free.level = use.level
		a.changed = true
	}
	for id := range use.levelParams {
		if !own[id] && !signature.free.levelParams[id] {
			if signature.free.levelParams == nil {
				signature.free.levelParams = map[int]bool{}
			}
			signature.free.levelParams[id] = true
			a.changed = true
		}
	}
	for id, maximum := range use.constraints {
		if !own[id] {
			if previous, constrained := signature.free.constraints[id]; !constrained || maximum < previous {
				signature.free.constrainID(id, maximum)
				a.changed = true
			}
		}
	}
	return a.signatureUse(signature, actuals)
}
func copySanyLeibnizBindings(source map[string]sanyLeibnizBinding) map[string]sanyLeibnizBinding {
	result := make(map[string]sanyLeibnizBinding, len(source))
	for name, binding := range source {
		result[name] = binding
	}
	return result
}
func sanyLeibnizNestedContext(ctx *sanyLeibnizContext) *sanyLeibnizContext {
	return &sanyLeibnizContext{module: ctx.module, variables: ctx.variables, formals: copySanyLeibnizBindings(ctx.formals), locals: ctx.locals, atUse: ctx.atUse, hasAt: ctx.hasAt}
}
func (a *sanyLeibnizAnalyzer) boundExpression(bounds []BoundVar, expressions []Expr, ctx *sanyLeibnizContext) sanyLeibnizUse {
	var use sanyLeibnizUse
	nested := sanyLeibnizNestedContext(ctx)
	for _, bound := range bounds {
		domain := a.expression(bound.Set, ctx)
		use.merge(domain)
		use.constrain(domain.levelParams, actionLevel)
		binding := sanyLeibnizBinding{}
		if bound.LevelKnown {
			binding.use.level = tlaLevel(bound.Level)
			use.merge(binding.use)
		}
		nested.formals[bound.Name] = binding
	}
	for _, expr := range expressions {
		use.merge(a.expression(expr, nested))
	}
	return use
}
func (a *sanyLeibnizAnalyzer) expression(expr Expr, ctx *sanyLeibnizContext) sanyLeibnizUse {
	if expr == nil {
		return sanyLeibnizUse{}
	}
	key := sanyLeibnizExpressionKey{expr: expr, context: ctx}
	if use, evaluated := a.expressions[key]; evaluated {
		return use
	}
	use := a.expressionUncached(expr, ctx)
	a.expressions[key] = use
	return use
}

func (a *sanyLeibnizAnalyzer) expressionUncached(expr Expr, ctx *sanyLeibnizContext) sanyLeibnizUse {
	argument := func(expr Expr) sanyLeibnizBinding { return sanyLeibnizBinding{expr: expr, context: ctx} }
	if selected := sanyExprSelection(expr); selected != nil && !selected.operator && selected.body != nil {
		arguments := make([]sanyLeibnizBinding, len(selected.args))
		for i, argumentExpr := range selected.args {
			arguments[i] = argument(argumentExpr)
		}
		return a.definitionBody(selected.definition, selected.body, selected.params, arguments, ctx, nil)
	}
	switch e := expr.(type) {
	case *IdentExpr:
		if e.Name == "@" && ctx.hasAt {
			return ctx.atUse
		}
		if binding, ok := ctx.formals[e.Name]; ok {
			return a.binding(binding)
		}
		if binding, ok := ctx.variables[e.Name]; ok {
			return a.binding(binding)
		}
		return a.apply(e, nil, ctx)
	case *UnaryExpr:
		return a.apply(&IdentExpr{Name: e.Op}, []sanyLeibnizBinding{argument(e.Expr)}, ctx)
	case *BinaryExpr:
		return a.apply(&IdentExpr{Name: e.Op}, []sanyLeibnizBinding{argument(e.Left), argument(e.Right)}, ctx)
	case *CallExpr:
		arguments := make([]sanyLeibnizBinding, len(e.Args))
		for i, expr := range e.Args {
			arguments[i] = argument(expr)
		}
		return a.apply(e.Callee, arguments, ctx)
	case *QuantifierExpr:
		use := a.boundExpression([]BoundVar{{Name: e.Var, Set: e.Set, LevelKnown: e.LevelKnown, Level: e.Level}}, []Expr{e.Body}, ctx)
		if e.Kind == "\\AA" || e.Kind == "\\EE" || e.Kind == "TEMPORAL_FORALL" || e.Kind == "TEMPORAL_EXISTS" {
			use.levelParams = nil
			use.level = temporalLevel
		}
		return use
	case *ActionExpr:
		name := "$SquareAct"
		if actionExprIsAngle(e) {
			name = "$AngleAct"
		}
		return a.apply(&IdentExpr{Name: name}, []sanyLeibnizBinding{argument(e.Action), argument(e.Subscript)}, ctx)
	case *FairnessExpr:
		return a.apply(&IdentExpr{Name: "$" + e.Kind}, []sanyLeibnizBinding{argument(e.Subscript), argument(e.Action)}, ctx)
	case *ChooseExpr:
		bounds := e.boundVars()
		if len(bounds) > 0 {
			bounds[0].Set = e.Set
		}
		use := a.boundExpression(bounds, []Expr{e.Body}, ctx)
		use.constrain(use.levelParams, actionLevel)
		return use
	case *FunctionExpr:
		return a.boundExpression(e.Bounds, []Expr{e.Body}, ctx)
	case *SetComprehensionExpr:
		use := a.boundExpression(e.Bounds, []Expr{e.Element, e.Predicate}, ctx)
		use.constrain(use.levelParams, actionLevel)
		return use
	case *TupleExpr:
		arguments := make([]sanyLeibnizBinding, len(e.Elems))
		for i, element := range e.Elems {
			arguments[i] = argument(element)
		}
		return a.apply(&IdentExpr{Name: "$Tuple"}, arguments, ctx)
	case *SetExpr:
		arguments := make([]sanyLeibnizBinding, len(e.Elems))
		for i, element := range e.Elems {
			arguments[i] = argument(element)
		}
		return a.apply(&IdentExpr{Name: "$SetEnumerate"}, arguments, ctx)
	case *FunctionSetExpr:
		return a.apply(&IdentExpr{Name: "$SetOfFcns"}, []sanyLeibnizBinding{argument(e.Domain), argument(e.Range)}, ctx)
	case *RecordSetExpr:
		arguments := make([]sanyLeibnizBinding, len(e.Fields))
		for i, field := range e.Fields {
			arguments[i] = argument(field.Set)
		}
		return a.apply(&IdentExpr{Name: "$SetOfRcds"}, arguments, ctx)
	case *RecordComponentExpr:
		return a.apply(&IdentExpr{Name: "$RcdSelect"}, []sanyLeibnizBinding{argument(e.Record)}, ctx)
	case *FunctionAppExpr:
		var index sanyLeibnizUse
		for _, expr := range e.Args {
			index.merge(a.expression(expr, ctx))
		}
		index.constrain(index.levelParams, actionLevel)
		return a.apply(&IdentExpr{Name: "$FcnApply"}, []sanyLeibnizBinding{argument(e.Function), {use: index}}, ctx)
	case *ExceptExpr:
		base := a.expression(e.Base, ctx)
		var use, prefix sanyLeibnizUse
		use.merge(base)
		use.constrain(base.levelParams, actionLevel)
		prefix.merge(base)
		for _, spec := range e.Specs {
			nested := sanyLeibnizNestedContext(ctx)
			nested.hasAt = true
			nested.atUse = prefix
			// AtNode copies level/all parameters and constraints from the base
			// and preceding EXCEPT components, but not nonLeibnizParams.
			nested.atUse.non = nil
			var component sanyLeibnizUse
			for _, path := range spec.Components {
				for _, expr := range path.Indices {
					component.merge(a.expression(expr, ctx))
				}
			}
			component.constrain(component.levelParams, actionLevel)
			component.merge(a.expression(spec.Value, nested))
			use.merge(component)
			use.constrain(component.levelParams, actionLevel)
			var next sanyLeibnizUse
			next.merge(prefix)
			next.merge(component)
			prefix = next
		}
		return use
	case *LetExpr:
		nested := sanyLeibnizNestedContext(ctx)
		nested.locals = make(map[string]sanyLeibnizLocal, len(ctx.locals)+len(e.Definitions))
		for name, local := range ctx.locals {
			nested.locals[name] = local
		}
		for i := range e.Definitions {
			def := &e.Definitions[i]
			nested.locals[def.Name] = sanyLeibnizLocal{recursive: letRecursiveNames(e)[def.Name], ref: sanySelectorDefinition{module: ctx.module, def: def, params: sanyDefinitionParams(def)}, context: nested}
		}
		use := a.expression(e.Body, nested)
		// LetInNode.levelCheck copies allParams from the body, but does not
		// copy nonLeibnizParams. Preserve the source's propagation rule.
		use.non = nil
		return use
	default:
		var use sanyLeibnizUse
		for _, child := range sanySubexpressionChildren(expr) {
			use.merge(a.expression(child, ctx))
		}
		return use
	}
}

func (a *sanyLeibnizAnalyzer) signatureUse(signature *sanyLeibnizSignature, arguments []sanyLeibnizBinding) sanyLeibnizUse {
	use := a.argumentUses(arguments)
	use.levelParams = nil
	use.level = constantLevel
	use.merge(signature.free)
	for i, argument := range arguments {
		if i < len(signature.maxLevels) {
			use.constrain(a.binding(argument).levelParams, signature.maxLevels[i])
		}
		if i < len(signature.weights) && signature.weights[i] {
			use.mergeLevelParams(a.binding(argument))
		}
		if i < len(signature.non) && signature.non[i] {
			argUse := a.binding(argument)
			argUse.restrict()
			argUse.levelParams = nil
			argUse.level = constantLevel
			use.merge(argUse)
		}
	}
	return use
}

// Follow formal aliases to the actual operator. The finite source operator/body
// identities keep recursive operator-argument summaries distinct from scalars.
func (a *sanyLeibnizAnalyzer) operatorIdentity(binding sanyLeibnizBinding) string {
	for binding.expr != nil {
		if id, ok := binding.expr.(*IdentExpr); ok && binding.context != nil {
			if next, ok := binding.context.formals[id.Name]; ok {
				binding = next
				continue
			}
			if next, ok := binding.context.variables[id.Name]; ok {
				binding = next
				continue
			}
			return binding.context.module.Name + "!" + id.Name
		}
		return fmt.Sprintf("%p", binding.expr)
	}
	return "formal"
}
