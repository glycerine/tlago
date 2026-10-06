// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2022, Oracle and/or its affiliates.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import (
	"fmt"
	"strings"
)

type sanySelectorInstance struct {
	owner *Module
	inst  Instance
}
type sanySelectorDefinition struct {
	module         *Module
	def            *Definition
	wrappers       []sanySelectorInstance
	params         []BoundVar
	suffices       bool
	recursiveLevel int
	undefined      bool // A RECURSIVE declaration whose body has not been generated.
}
type sanySelectorSelection struct {
	name        string
	definition  sanySelectorDefinition
	body        Expr
	params      []BoundVar
	args        []Expr
	lets        []*LetExpr
	operator    bool
	newSymbol   *NewSymbol
	assumeProve *AssumeProve
}
type sanySelectorResolver struct {
	spec        *Spec
	scopes      map[*Module]map[string]sanySelectorDefinition
	visiting    map[*Module]bool
	diags       Diagnostics
	inProof     map[*AssumeProve]bool
	fact        bool
	letContexts int
	letLevel    int
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
	ownDefinitions := map[string]bool{}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		if ownDefinitions[def.Name] {
			continue
		}
		ownDefinitions[def.Name] = true
		out[def.Name] = sanySelectorDefinition{module: mod, def: def, params: sanyDefinitionParams(def)}
	}
	for _, assumption := range mod.Assumptions {
		if assumption.Name != "" {
			definition := &Definition{Name: assumption.Name, Expr: assumption.Expr, Syntax: assumption.Syntax,
				AssumeProveBody: assumption.AssumeProveBody, TheoremLike: true, Pos: assumption.Pos}
			out[assumption.Name] = sanySelectorDefinition{module: mod, def: definition}
		}
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

type sanyModuleSelectorGenerator struct {
	resolver *sanySelectorResolver
	module   *Module
	scope    map[string]sanySelectorDefinition
	at       func(*SanySyntaxNode) map[string]sanySelectorDefinition
}

func (r *sanySelectorResolver) moduleGenerator(mod *Module) *sanyModuleSelectorGenerator {
	scope := r.scope(mod)
	// Prepare selectors in each body's generation context. Looking through a
	// later definition can otherwise report selector errors before Java even
	// has that operator in SymbolTable.
	inherited := map[string]Position{}
	for _, name := range mod.Extends {
		for key, ref := range r.scope(r.spec.Modules[name]) {
			if !ref.def.Local {
				inherited[key] = ref.def.DeclarationPosition()
			}
		}
	}
	contexts := sanyModuleExpressionContexts(mod, r.spec, inherited)
	at := func(syntax *SanySyntaxNode) map[string]sanySelectorDefinition {
		if context, ok := contexts.bindings[syntax]; ok {
			visible := make(map[string]sanySelectorDefinition, len(context))
			for name, ref := range scope {
				if _, exists := context[name]; exists {
					ref.undefined = !contexts.definitions[syntax][name]
					if arity, recursive := contexts.recursiveArities[syntax][name]; ref.undefined && recursive {
						ref.params = make([]BoundVar, arity)
					}
					visible[name] = ref
				}
			}
			return visible
		}
		return scope
	}
	return &sanyModuleSelectorGenerator{resolver: r, module: mod, scope: scope, at: at}
}

func (g *sanyModuleSelectorGenerator) prepare(run func()) Diagnostics {
	before := len(g.resolver.diags)
	run()
	return g.resolver.diags[before:]
}

func (g *sanyModuleSelectorGenerator) definition(def *Definition) Diagnostics {
	return g.prepare(func() { g.resolver.walkDefinition(def, g.module, g.at(def.Syntax)) })
}

// processFunction generates domains before checking the definition's symbol,
// and generates the function body afterward in its temporary formal context.
func (g *sanyModuleSelectorGenerator) functionDomains(def *Definition) Diagnostics {
	return g.prepare(func() {
		if function, ok := def.Expr.(*FunctionExpr); ok {
			for _, domain := range sanyFunctionDomainExpressions(function) {
				g.resolver.walk(domain, g.module, g.at(def.Syntax), 0)
			}
		}
	})
}

func (g *sanyModuleSelectorGenerator) functionBody(def *Definition) Diagnostics {
	return g.prepare(func() {
		if function, ok := def.Expr.(*FunctionExpr); ok {
			g.resolver.walk(function.Body, g.module, g.at(def.Syntax), 0)
		}
	})
}

func (g *sanyModuleSelectorGenerator) fact(fact NamedExpr) Diagnostics {
	return g.prepare(func() {
		if fact.AssumeProveBody != nil {
			g.resolver.walkAssumeProve(fact.AssumeProveBody, g.module, g.at(fact.Syntax))
		} else {
			g.resolver.walk(fact.Expr, g.module, g.at(fact.Syntax), 0)
		}
	})
}

func (g *sanyModuleSelectorGenerator) instance(instance Instance) Diagnostics {
	return g.prepare(func() { g.resolver.walkInstance(instance, g.module, g.at(instance.Syntax)) })
}

func (g *sanyModuleSelectorGenerator) reference(ref ProofRef) Diagnostics {
	return g.prepare(func() {
		if !ref.Defs && ref.Expr != nil {
			_, direct := ref.Expr.(*IdentExpr)
			g.resolver.walkProofFact(ProofFact{Expr: ref.Expr, Direct: direct}, g.module, g.at(ref.Syntax))
		}
	})
}

func (g *sanyModuleSelectorGenerator) proof(proof ProofSummary) Diagnostics {
	return g.prepare(func() {
		scope := g.at(proof.Syntax)
		// processTheorem registers the name after generating the statement and
		// before generating its proof. Later module units remain unavailable.
		for _, theorem := range g.module.Theorems {
			if theorem.Syntax == proof.Syntax && theorem.Name != "" {
				if ref, ok := g.scope[theorem.Name]; ok {
					scope[theorem.Name] = ref
				}
			}
		}
		g.resolver.walkProof(proof, g.module, scope)
	})
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
		source.selectorFailure = false
		source.selectorDiagnostic = nil
		selection, handled, err := r.selectExpr(expr, scope, expected)
		if handled {
			if err != nil {
				source.selectorFailure = true
				diagnostic := errorAt(expr.Position(), "E4340", "%s", err)
				if detail, ok := err.(*sanySelectorLocationError); ok {
					diagnostic.SANYRange = detail.location
					diagnostic.SANYMessage = detail.message
					diagnostic.SANYParameters = append([]any(nil), detail.parameters...)
					if detail.code != "" {
						diagnostic.Code = detail.code
					}
				}
				if r.letContexts > 0 {
					source.selectorDiagnostic = &diagnostic
				} else {
					r.diags = append(r.diags, diagnostic)
				}
				return
			}
			source.selection = selection
			previousFact := r.fact
			r.fact = false
			defer func() { r.fact = previousFact }()
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
		r.letContexts++
		r.letLevel++
		nested := make(map[string]sanySelectorDefinition, len(scope)+len(e.Definitions))
		for name, ref := range scope {
			nested[name] = ref
		}
		for _, unit := range sanyLetGenerationUnits(e) {
			switch {
			case unit.declaration != nil:
				for _, name := range unit.declaration.Names {
					arity, _ := declarationArity(*unit.declaration, name)
					def := &Definition{Name: name, Pos: unit.declaration.Pos, Syntax: unit.declaration.Syntax}
					if _, exists := nested[name]; !exists {
						nested[name] = sanySelectorDefinition{module: mod, def: def, params: make([]BoundVar, arity), undefined: true, recursiveLevel: r.letLevel}
					}
				}
			case unit.definition != nil:
				def := unit.definition
				previous, exists := nested[def.Name]
				r.walkDefinition(def, mod, nested)
				if !exists || (previous.undefined && previous.recursiveLevel == r.letLevel && (!def.FunctionDef || len(previous.params) == 0)) {
					nested[def.Name] = sanySelectorDefinition{module: mod, def: def, params: sanyDefinitionParams(def)}
				}
			case unit.instance != nil:
				r.walkInstance(*unit.instance, mod, nested)
				r.addInstance(nested, mod, *unit.instance)
			}
		}
		r.letLevel--
		r.walk(e.Body, mod, nested, 0)
		r.letContexts--
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
	// selectorToNode validates the declared operator's arity before it asks
	// whether a selected recursive body has been completed.
	if !result.operator && len(result.params) != len(rootArgs) {
		detail := &sanySelectorLocationError{code: "E4204", message: fmt.Sprintf("The operator %s requires %d arguments.", name, len(result.params)), parameters: []any{name, len(result.params)}}
		if selector.Syntax != nil {
			detail.location = selector.Syntax.Range
		}
		if arguments := selector.Steps[end-1].Arguments; arguments != nil {
			detail.location = arguments.Range
		}
		return nil, true, detail
	}
	if err := bind(result.params, rootArgs); err != nil {
		return nil, true, err
	}
	if ref.undefined {
		detail := &sanySelectorLocationError{code: "E4005", message: fmt.Sprintf("Subexpression of  `%s' used inside the operator's definition.", name)}
		if selector.Steps[end-1].Syntax != nil {
			detail.location = selector.Steps[end-1].Syntax.Range
		}
		return nil, true, detail
	}
	currentAP := ref.def.AssumeProveBody
	apGoal := currentAP
	inSuffices := false
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
			if currentAP != nil {
				var declarationScope bool
				label, declarationScope = sanyAssumeProveLabel(currentAP, step.Name)
				if label != nil && declarationScope && ref.suffices == r.inProof[apGoal] {
					return nil, true, sanySelectorErrorAt(step, fmt.Sprintf("Accessing subexpression labeled `%s' of ASSUME/PROVE clause within the scope of a declaration\n from outside that declaration's scope.", step.Name))
				}
			}
			if label == nil {
				return nil, true, fmt.Errorf("cannot find label %s", step.Name)
			}
			var params []BoundVar
			for _, name := range label.Params {
				params = append(params, BoundVar{Name: name, Pos: label.Pos})
			}
			if !result.operator && len(params) != len(args) {
				detail := &sanySelectorLocationError{code: "E4337", message: fmt.Sprintf("Label `%s' used with wrong number of arguments.", step.Name), parameters: []any{step.Name}}
				if step.Syntax != nil {
					detail.location = step.Syntax.Range
				}
				return nil, true, detail
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
			currentAP = nil
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
			if currentAP != nil {
				if ref.suffices && currentAP == apGoal && !inSuffices {
					if sanySelectorArgNumber(step, 1) != 1 {
						return nil, true, sanySelectorErrorAt(step, "Accessing non-existent subexpression of a SUFFICES")
					}
					inSuffices = true
					continue
				}
				inSuffices = false
				position := sanySelectorArgNumber(step, len(currentAP.Assumptions)+1)
				if position < 1 {
					return nil, true, sanyReportSelectorError(step)
				}
				declared := false
				for _, clause := range currentAP.Assumptions[:position-1] {
					declared = declared || clause.NewSymbol != nil
				}
				if declared && (currentAP != apGoal || ref.suffices == r.inProof[apGoal]) {
					return nil, true, sanySelectorErrorAt(step, "Accessing ASSUME/PROVE clause within the scope of a declaration\n from outside that declaration's scope.")
				}
				if position == len(currentAP.Assumptions)+1 {
					result.body = currentAP.Prove
					currentAP = nil
				} else {
					clause := currentAP.Assumptions[position-1]
					currentAP = clause.Nested
					if currentAP != nil {
						result.body = assumeProveExpr(currentAP, currentAP.Pos)
					} else if clause.NewSymbol != nil {
						if i+1 < len(selector.Steps) {
							return nil, true, sanySelectorErrorAt(step, "Selected a subexpression of a NEW clause of an ASSUME.")
						}
						if !r.fact {
							return nil, true, &sanySelectorLocationError{message: "Selected a NEW declaration as an expression or operator.", location: selector.Syntax.Range}
						}
						result.newSymbol = clause.NewSymbol
						result.body = nil
					} else {
						result.body = clause.Expr
					}
				}
				continue
			}
			if _, ok := result.body.(*CaseExpr); ok && (i+1 == len(selector.Steps) || selector.Steps[i+1].Kind == SanySelectorName) {
				return nil, true, fmt.Errorf("subexpression of CASE must have form !i!j")
			}
			if let, ok := result.body.(*LetExpr); ok {
				result.lets = append(result.lets, let)
			}
			if literal, ok := result.body.(*LiteralExpr); ok && (literal.Kind == "number" || literal.Kind == "string") {
				detail := &sanySelectorLocationError{code: "E4005", message: "Selecting subexpression of expression that has none."}
				if step.Syntax != nil {
					detail.location = step.Syntax.Range
				}
				return nil, true, detail
			}
			next, ok := sanySelectorOperand(result.body, step)
			if !ok {
				return nil, true, sanyReportSelectorError(step)
			}
			result.body = next
		}
	}
	if currentAP != nil && !r.fact {
		detail := &sanySelectorLocationError{message: "Selected ASSUME/PROVE instead of expression."}
		if selector.Syntax != nil {
			detail.location = selector.Syntax.Range
		}
		return nil, true, detail
	}
	result.assumeProve = currentAP
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

// Location-bearing Generator failures retain the individual selector token.
type sanySelectorLocationError struct {
	code       string
	parameters []any
	message    string
	location   SanyRange
}

func (e *sanySelectorLocationError) Error() string { return e.message }
func sanySelectorErrorAt(step SanySelectorStep, message string) error {
	detail := &sanySelectorLocationError{message: message}
	if step.Syntax != nil {
		detail.location = step.Syntax.Range
	}
	return detail
}

// sanyReportSelectorError mirrors Generator.reportSelectorError, including the
// selector description and the location of the individual selector node.
func sanyReportSelectorError(step SanySelectorStep) error {
	item := step.Name
	if step.Syntax != nil {
		switch step.Syntax.Kind.JavaName() {
		case "N_StructOp":
			item = "argument selector " + step.Name
		case "N_OpArgs":
			item = "!(...)"
		}
	}
	detail := &sanySelectorLocationError{
		code:    "E4005",
		message: fmt.Sprintf("Nonexistent operand specified by `%s'.", item),
	}
	if step.Syntax != nil {
		detail.location = step.Syntax.Range
	}
	return detail
}

func sanySelectorArgNumber(step SanySelectorStep, count int) int {
	switch step.Kind {
	case SanySelectorFirst:
		if count > 0 {
			return 1
		}
	case SanySelectorLast:
		if count == 2 {
			return 2
		}
	default:
		if step.Kind > 0 && step.Kind <= count {
			return step.Kind
		}
	}
	return -1
}
func sanyAssumeProveLabel(body *AssumeProve, name string) (*LabelExpr, bool) {
	declared := false
	for _, clause := range body.Assumptions {
		expr := clause.Expr
		if clause.NewSymbol != nil {
			expr = clause.NewSymbol.Domain
		}
		if clause.Nested != nil {
			if label, _ := sanyAssumeProveLabel(clause.Nested, name); label != nil {
				return label, declared
			}
		} else if label, _ := sanySelectorLabel(expr, name, nil); label != nil {
			return label, declared
		}
		declared = declared || clause.NewSymbol != nil
	}
	label, _ := sanySelectorLabel(body.Prove, name, nil)
	return label, declared
}
func (r *sanySelectorResolver) walkProof(proof ProofSummary, mod *Module, scope map[string]sanySelectorDefinition) {
	nested := make(map[string]sanySelectorDefinition, len(scope))
	for name, ref := range scope {
		nested[name] = ref
	}
	previous := r.inProof
	r.inProof = map[*AssumeProve]bool{}
	defer func() { r.inProof = previous }()
	for _, ref := range scope {
		if ref.def.Expr == proof.Goal && ref.def.AssumeProveBody != nil {
			r.inProof[ref.def.AssumeProveBody] = true
		}
	}
	type entry struct {
		depth int
		name  string
		body  *AssumeProve
	}
	var active []entry
	for _, fact := range proof.Facts {
		r.walkProofFact(fact, mod, nested)
	}
	for _, step := range proof.Steps {
		kept := active[:0]
		for _, item := range active {
			if item.depth > step.Depth {
				delete(nested, item.name)
			} else {
				kept = append(kept, item)
				if item.body != nil {
					r.inProof[item.body] = item.depth < step.Depth
				}
			}
		}
		active = kept
		if step.AssumeProveBody != nil {
			r.walkAssumeProve(step.AssumeProveBody, mod, nested)
		} else {
			r.walk(step.Expr, mod, nested, 0)
		}
		for _, expr := range step.Exprs {
			r.walk(expr, mod, nested, 0)
		}
		for _, bound := range step.Bounds {
			r.walk(bound.Set, mod, nested, 0)
		}
		if step.QualifiedName != "" && step.Expr != nil {
			def := &Definition{Name: step.QualifiedName, Expr: step.Expr, AssumeProveBody: step.AssumeProveBody, Pos: step.Pos, TheoremLike: true}
			nested[step.QualifiedName] = sanySelectorDefinition{module: mod, def: def, suffices: step.Suffices}
			active = append(active, entry{depth: step.Depth, name: step.QualifiedName, body: step.AssumeProveBody})
		}
		if step.AssumeProveBody != nil {
			r.inProof[step.AssumeProveBody] = true
		}
		for _, fact := range step.Facts {
			r.walkProofFact(fact, mod, nested)
		}
		if step.AssumeProveBody != nil {
			r.inProof[step.AssumeProveBody] = false
		}
	}
}

func (r *sanySelectorResolver) walkProofFact(fact ProofFact, mod *Module, scope map[string]sanySelectorDefinition) {
	previous := r.fact
	r.fact = fact.Direct
	r.walk(fact.Expr, mod, scope, 0)
	r.fact = previous
}

// Preserve NEW domains and nested clause structure during semantic generation;
// the ordinary expression projection omits domains and changes operand indices.
func (r *sanySelectorResolver) walkDefinition(def *Definition, mod *Module, scope map[string]sanySelectorDefinition) {
	if def.AssumeProveBody != nil {
		r.walkAssumeProve(def.AssumeProveBody, mod, scope)
	} else {
		r.walk(def.Expr, mod, scope, 0)
	}
}
func (r *sanySelectorResolver) walkAssumeProve(body *AssumeProve, mod *Module, scope map[string]sanySelectorDefinition) {
	for _, clause := range body.Assumptions {
		if clause.NewSymbol != nil {
			r.walk(clause.NewSymbol.Domain, mod, scope, 0)
		} else if clause.Nested != nil {
			r.walkAssumeProve(clause.Nested, mod, scope)
		} else {
			r.walk(clause.Expr, mod, scope, 0)
		}
	}
	r.walk(body.Prove, mod, scope, 0)
}
