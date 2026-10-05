// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2022, Oracle and/or its affiliates.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import (
	"fmt"
	"sort"
	"strings"
)

type sanySelectorInstance struct {
	owner *Module
	inst  Instance
}
type sanySelectorDefinition struct {
	module   *Module
	def      *Definition
	wrappers []sanySelectorInstance
	params   []BoundVar
}
type sanySelectorSelection struct {
	name       string
	definition sanySelectorDefinition
	body       Expr
	params     []BoundVar
	args       []Expr
	lets       []*LetExpr
	operator   bool
}
type sanySelectorResolver struct {
	spec     *Spec
	scopes   map[*Module]map[string]sanySelectorDefinition
	visiting map[*Module]bool
	diags    Diagnostics
}

func sanyExprSource(expr Expr) *SanyExprSource {
	switch e := expr.(type) {
	case *IdentExpr:
		return &e.SanyExprSource
	case *CallExpr:
		return &e.SanyExprSource
	}
	return nil
}
func sanyExprSelection(expr Expr) *sanySelectorSelection {
	if source := sanyExprSource(expr); source != nil {
		return source.selection
	}
	return nil
}
func sanyDefinitionParams(def *Definition) []BoundVar {
	params := make([]BoundVar, len(def.Params))
	for i, name := range def.Params {
		params[i] = BoundVar{Name: name, Pos: def.ParamPositions[name], OperatorArity: def.ParamArities[name], HasOperatorArity: true}
	}
	return params
}
func (r *sanySelectorResolver) scope(mod *Module) map[string]sanySelectorDefinition {
	if scope := r.scopes[mod]; scope != nil {
		return scope
	}
	out := map[string]sanySelectorDefinition{}
	if mod == nil || r.visiting[mod] {
		return out
	}
	r.visiting[mod] = true
	for _, name := range mod.Extends {
		for key, ref := range r.scope(r.spec.Modules[name]) {
			if !ref.def.Local {
				out[key] = ref
			}
		}
	}
	for _, inst := range mod.Instances {
		r.addInstance(out, mod, inst)
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		out[def.Name] = sanySelectorDefinition{module: mod, def: def, params: sanyDefinitionParams(def)}
	}
	keys := make([]string, 0, len(out))
	for key := range out {
		keys = append(keys, key)
	}
	for _, key := range keys {
		ref := out[key]
		out[mod.Name+"!"+key] = ref
	}
	delete(r.visiting, mod)
	r.scopes[mod] = out
	return out
}
func (r *sanySelectorResolver) addInstance(scope map[string]sanySelectorDefinition, owner *Module, inst Instance) {
	target := r.spec.Modules[inst.Module]
	if target == nil {
		return
	}
	for name, ref := range r.scope(target) {
		if ref.def.Local || strings.HasPrefix(name, target.Name+"!") {
			continue
		}
		next := ref
		next.wrappers = append([]sanySelectorInstance{{owner: owner, inst: inst}}, ref.wrappers...)
		next.params = nil
		for _, name := range inst.Params {
			next.params = append(next.params, BoundVar{Name: name, Pos: inst.ParamPositions[name], OperatorArity: inst.ParamArities[name], HasOperatorArity: true})
		}
		next.params = append(next.params, ref.params...)
		if inst.Name != "" {
			scope[inst.Name+"!"+name] = next
		} else {
			scope[name] = next
		}
	}
}
func resolveSanySelectors(spec *Spec) Diagnostics {
	r := &sanySelectorResolver{spec: spec, scopes: map[*Module]map[string]sanySelectorDefinition{}, visiting: map[*Module]bool{}}
	names := make([]string, 0, len(spec.Modules))
	for name := range spec.Modules {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		mod := spec.Modules[name]
		if mod == nil {
			continue
		}
		scope := r.scope(mod)
		for i := range mod.Definitions {
			r.walk(mod.Definitions[i].Expr, mod, scope, 0)
		}
		for _, a := range mod.Assumptions {
			r.walk(a.Expr, mod, scope, 0)
		}
		for _, a := range mod.Theorems {
			r.walk(a.Expr, mod, scope, 0)
		}
		for _, inst := range mod.Instances {
			r.walkInstance(inst, mod, scope)
		}
	}
	return r.diags
}
func (r *sanySelectorResolver) walkInstance(inst Instance, mod *Module, scope map[string]sanySelectorDefinition) {
	arities := map[string]int{}
	if target := r.spec.Modules[inst.Module]; target != nil {
		for _, decl := range target.Declarations {
			for _, name := range decl.Names {
				arities[name] = decl.Arities[name]
			}
		}
	}
	for _, subst := range inst.SubstitutionList {
		r.walk(subst.Expr, mod, scope, arities[subst.Name])
	}
}
func (r *sanySelectorResolver) walk(expr Expr, mod *Module, scope map[string]sanySelectorDefinition, expected int) {
	if expr == nil {
		return
	}
	if source := sanyExprSource(expr); source != nil && source.Selector != nil {
		source.selection = nil
		selection, handled, err := r.selectExpr(expr, scope, expected)
		if handled {
			if err != nil {
				r.diags = append(r.diags, errorAt(expr.Position(), "E4340", "%s", err))
				return
			}
			source.selection = selection
			for i, arg := range selection.args {
				arity := 0
				if i < len(selection.params) {
					arity = selection.params[i].OperatorArity
				}
				r.walk(arg, mod, scope, arity)
			}
			return
		}
	}
	switch e := expr.(type) {
	case *CallExpr:
		var params []BoundVar
		if id, ok := e.Callee.(*IdentExpr); ok {
			if ref, ok := scope[id.Name]; ok {
				params = ref.params
			}
		} else {
			r.walk(e.Callee, mod, scope, 0)
		}
		for i, arg := range e.Args {
			arity := 0
			if i < len(params) {
				arity = params[i].OperatorArity
			}
			r.walk(arg, mod, scope, arity)
		}
	case *LetExpr:
		nested := make(map[string]sanySelectorDefinition, len(scope)+len(e.Definitions))
		for name, ref := range scope {
			nested[name] = ref
		}
		for i := range e.Definitions {
			def := &e.Definitions[i]
			nested[def.Name] = sanySelectorDefinition{module: mod, def: def, params: sanyDefinitionParams(def)}
		}
		for _, inst := range e.Instances {
			r.addInstance(nested, mod, inst)
			r.walkInstance(inst, mod, nested)
		}
		for i := range e.Definitions {
			r.walk(e.Definitions[i].Expr, mod, nested, 0)
		}
		r.walk(e.Body, mod, nested, 0)
	default:
		for _, child := range sanySubexpressionChildren(expr) {
			r.walk(child, mod, scope, 0)
		}
	}
}
func (r *sanySelectorResolver) selectExpr(expr Expr, scope map[string]sanySelectorDefinition, expected int) (*sanySelectorSelection, bool, error) {
	selector := sanyExprSource(expr).Selector
	if len(selector.Steps) < 2 {
		return nil, false, nil
	}
	// FindingOpName accumulates unresolved compound name components and their
	// arguments until an actual definition is found, including INSTANCE exports.
	var ref sanySelectorDefinition
	var found bool
	end := 0
	name := ""
	var flat []Expr
	if call, ok := expr.(*CallExpr); ok {
		flat = call.Args
	}
	argIndex := 0
	take := func(step SanySelectorStep) ([]Expr, error) {
		count := 0
		if step.Arguments != nil {
			count = len(expressionChildren(step.Arguments))
		}
		if argIndex+count > len(flat) {
			return nil, fmt.Errorf("selector argument syntax does not match generated arguments")
		}
		args := flat[argIndex : argIndex+count]
		argIndex += count
		return args, nil
	}
	var rootArgs []Expr
	for i, step := range selector.Steps {
		if step.Kind != SanySelectorName {
			break
		}
		if name != "" {
			name += "!"
		}
		name += sanyCanonicalOperatorImage(step.Name)
		args, err := take(step)
		if err != nil {
			return nil, true, err
		}
		rootArgs = append(rootArgs, args...)
		if ref, found = scope[name]; found {
			end = i + 1
			break
		}
	}
	if !found || end == len(selector.Steps) {
		return nil, false, nil
	}
	result := &sanySelectorSelection{name: name, definition: ref, body: ref.def.Expr, params: append([]BoundVar(nil), ref.params...), operator: expected > 0}
	bind := func(params []BoundVar, args []Expr) error {
		if result.operator {
			if len(args) != 0 {
				return fmt.Errorf("operator selector should not have arguments")
			}
		} else if len(args) != len(params) {
			return fmt.Errorf("selector requires %d arguments, got %d", len(params), len(args))
		}
		result.args = append(result.args, args...)
		return nil
	}
	if err := bind(result.params, rootArgs); err != nil {
		return nil, true, err
	}
	modeLabels := selector.Steps[end].Kind == SanySelectorName
	for i := end; i < len(selector.Steps); i++ {
		step := selector.Steps[i]
		args, err := take(step)
		if err != nil {
			return nil, true, err
		}
		if step.Kind != SanySelectorName && step.Kind != SanySelectorNull && len(args) != 0 {
			return nil, true, fmt.Errorf("selector %s should not have arguments", step.Name)
		}
		if modeLabels {
			label, lets := sanySelectorLabel(result.body, step.Name, nil)
			if label == nil {
				return nil, true, fmt.Errorf("cannot find label %s", step.Name)
			}
			var params []BoundVar
			for _, name := range label.Params {
				params = append(params, BoundVar{Name: name, Pos: label.Pos})
			}
			if err := bind(params, args); err != nil {
				return nil, true, err
			}
			result.params = append(result.params, params...)
			result.lets = append(result.lets, lets...)
			if i+1 == len(selector.Steps) {
				result.body = label
			} else {
				result.body = label.Body
			}
			modeLabels = i+1 < len(selector.Steps) && selector.Steps[i+1].Kind == SanySelectorName
			continue
		}
		for {
			label, ok := result.body.(*LabelExpr)
			if !ok {
				break
			}
			result.body = label.Body
		}
		switch step.Kind {
		case SanySelectorColon:
			if i != end && selector.Steps[i-1].Kind != SanySelectorName {
				return nil, true, fmt.Errorf("!: should not follow an operand selector")
			}
			if i+1 < len(selector.Steps) && selector.Steps[i+1].Kind != SanySelectorName {
				return nil, true, fmt.Errorf("!: must precede a LET operator name")
			}
		case SanySelectorName:
			let, ok := result.body.(*LetExpr)
			if !ok {
				return nil, true, fmt.Errorf("a name selector here must be from a LET clause")
			}
			var def *Definition
			for j := range let.Definitions {
				if let.Definitions[j].Name == step.Name {
					def = &let.Definitions[j]
					break
				}
			}
			if def == nil {
				return nil, true, fmt.Errorf("unknown LET operator %s", step.Name)
			}
			params := sanyDefinitionParams(def)
			if err := bind(params, args); err != nil {
				return nil, true, err
			}
			result.params = append(result.params, params...)
			result.lets = append(result.lets, let)
			result.body = def.Expr
			modeLabels = i+1 < len(selector.Steps) && selector.Steps[i+1].Kind == SanySelectorName
		case SanySelectorNull, SanySelectorAt:
			params, body := sanySelectorBoundBody(result.body)
			if body == nil || len(params) == 0 {
				return nil, true, fmt.Errorf("selector %s requires an expression with bound identifiers", step.Name)
			}
			if step.Kind == SanySelectorAt && !result.operator {
				return nil, true, fmt.Errorf("!@ requires an operator argument context")
			}
			if err := bind(params, args); err != nil {
				return nil, true, err
			}
			result.params = append(result.params, params...)
			result.body = body
		default:
			if _, ok := result.body.(*CaseExpr); ok && (i+1 == len(selector.Steps) || selector.Steps[i+1].Kind == SanySelectorName) {
				return nil, true, fmt.Errorf("subexpression of CASE must have form !i!j")
			}
			if let, ok := result.body.(*LetExpr); ok {
				result.lets = append(result.lets, let)
			}
			next, ok := sanySelectorOperand(result.body, step)
			if !ok {
				return nil, true, fmt.Errorf("non-existent operand selected by %s", step.Name)
			}
			result.body = next
		}
	}
	if result.operator && len(result.params) != expected {
		return nil, true, fmt.Errorf("expected operator arity %d, found %d", expected, len(result.params))
	}
	return result, true, nil
}
func sanySelectorLabel(expr Expr, name string, lets []*LetExpr) (*LabelExpr, []*LetExpr) {
	if label, ok := expr.(*LabelExpr); ok {
		if label.Name == name {
			return label, lets
		}
		return nil, nil
	}
	if let, ok := expr.(*LetExpr); ok {
		lets = append(append([]*LetExpr(nil), lets...), let)
	}
	for _, child := range sanySubexpressionChildren(expr) {
		if label, path := sanySelectorLabel(child, name, lets); label != nil {
			return label, path
		}
	}
	return nil, nil
}
func sanySelectorBoundBody(expr Expr) ([]BoundVar, Expr) {
	switch e := expr.(type) {
	case *QuantifierExpr:
		var params []BoundVar
		body := Expr(e)
		for {
			q, ok := body.(*QuantifierExpr)
			if !ok || q.Syntax != e.Syntax {
				break
			}
			params = append(params, BoundVar{Name: q.Var, Pos: q.VarPos, OperatorArity: q.OperatorArity, HasOperatorArity: q.HasOperatorArity})
			body = q.Body
			if e.Syntax == nil {
				break
			}
		}
		return params, body
	case *ChooseExpr:
		return e.boundVars(), e.Body
	case *FunctionExpr:
		return e.Bounds, e.Body
	case *SetComprehensionExpr:
		if e.Predicate != nil {
			return e.Bounds, e.Predicate
		}
		return e.Bounds, e.Element
	}
	return nil, nil
}
