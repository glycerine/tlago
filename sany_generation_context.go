// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2022, Oracle and/or its affiliates.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import (
	"sort"
	"strings"
)

// An unresolved identifier yields Generator's nullOAN placeholder. An
// unresolved symbolic operator yields null instead; N_FcnAppl distinguishes
// the two before generating its arguments.
type sanyGenerationFailure uint8

const (
	sanyGenerationSucceeded sanyGenerationFailure = iota
	sanyGenerationNullOperator
	sanyGenerationNullExpression
)

func sanyExpressionGenerationFailure(expr Expr) sanyGenerationFailure {
	if source, ok := expr.(interface{ generationSource() *SanyExprSource }); ok {
		return source.generationSource().generationFailure
	}
	return sanyGenerationSucceeded
}

func setSanyExpressionGenerationFailure(expr Expr, failure sanyGenerationFailure) {
	if source, ok := expr.(interface{ generationSource() *SanyExprSource }); ok {
		source.generationSource().generationFailure = failure
	}
}

// Generator.generateModule visits the module body's heirs in order. An ordinary
// operator's body is generated before its name is entered in SymbolTable;
// RECURSIVE declarations enter their names when that declaration is visited.
// Keep these expression contexts separate from the completed module context
// used by the subsequent level-checking phase.
type sanyExpressionContexts struct {
	bindings         map[*SanySyntaxNode]map[string]Position
	definitions      map[*SanySyntaxNode]map[string]bool
	recursiveArities map[*SanySyntaxNode]map[string]int
}

func (contexts sanyExpressionContexts) at(syntax *SanySyntaxNode, completed map[string]Position) map[string]Position {
	if context, ok := contexts.bindings[syntax]; ok {
		return context
	}
	// Native semantic AST callers can supply nodes without parser syntax.
	return completed
}

func copySanyExpressionContext(context map[string]Position) map[string]Position {
	copy := make(map[string]Position, len(context))
	for name, position := range context {
		copy[name] = position
	}
	return copy
}

func sanyModuleExpressionContexts(mod *Module, spec *Spec, inherited map[string]Position) sanyExpressionContexts {
	contexts := sanyExpressionContexts{bindings: map[*SanySyntaxNode]map[string]Position{}, definitions: map[*SanySyntaxNode]map[string]bool{}, recursiveArities: map[*SanySyntaxNode]map[string]int{}}
	if mod.Syntax == nil {
		return contexts
	}
	heirs := mod.Syntax.GetHeirs()
	if len(heirs) < 3 {
		return contexts
	}
	context := copySanyExpressionContext(inherited)
	recursiveArities := map[string]int{}
	completed := make(map[string]bool, len(inherited))
	for name := range inherited {
		completed[name] = true
	}
	add := func(name string, position Position) {
		if _, exists := context[name]; !exists {
			context[name] = position
		}
	}
	addLocal := func(name string, position Position) {
		add(name, position)
	}
	for _, unit := range heirs[2].GetHeirs() {
		contexts.bindings[unit] = copySanyExpressionContext(context)
		contexts.definitions[unit] = copyBoolMap(completed)
		contexts.recursiveArities[unit] = copyIntMap(recursiveArities)
		switch unit.Kind.JavaName() {
		case "N_VariableDeclaration", "N_ParamDeclaration":
			for _, declaration := range mod.Declarations {
				if declaration.Syntax != unit {
					continue
				}
				for _, name := range declaration.Names {
					addLocal(name, declarationSymbolPosition(declaration, name))
				}
			}
		case "N_Recursive":
			for _, declaration := range mod.Recursives {
				if declaration.Syntax != unit {
					continue
				}
				for _, name := range declaration.Names {
					addLocal(name, declarationSymbolPosition(declaration, name))
					recursiveArities[name], _ = declarationArity(declaration, name)
				}
			}
		case "N_OperatorDefinition", "N_FunctionDefinition", "N_Theorem":
			for _, definition := range mod.Definitions {
				if definition.Syntax != unit {
					continue
				}
				addLocal(definition.Name, definition.DeclarationPosition())
				complete := true
				if want, recursive := recursiveArities[definition.Name]; recursive && definition.FunctionDef && want > 0 {
					complete = false
				}
				completed[definition.Name] = complete
				addSubexpressionReferenceNames(context, definition.Name, definition.Expr)
			}
		case "N_Assumption":
			for _, assumption := range mod.Assumptions {
				if assumption.Syntax == unit && assumption.Name != "" {
					addLocal(assumption.Name, assumption.SourcePosition())
					completed[assumption.Name] = true
					addSubexpressionReferenceNames(context, assumption.Name, assumption.Expr)
				}
			}
		case "N_ModuleDefinition", "N_Instance":
			for _, instance := range mod.Instances {
				if instance.Syntax != unit {
					continue
				}
				if instance.Name != "" {
					add(instance.Name, instance.SourcePosition())
					add(instanceNameSentinel(instance.Name), instance.SourcePosition())
				}
				for _, symbol := range semanticInstanceSymbols(instance, spec) {
					add(symbol.name, symbol.sourcePosition())
					completed[symbol.name] = true
				}
				if spec == nil {
					continue
				}
				instancee := spec.Modules[instance.Module]
				if instancee == nil {
					continue
				}
				for _, definition := range instancee.Definitions {
					if definition.Local {
						continue
					}
					if instance.exportsUnqualified() {
						addSubexpressionReferenceNames(context, definition.Name, definition.Expr)
					}
					if instance.Name != "" {
						addSubexpressionReferenceNames(context, instance.qualifier()+"!"+definition.Name, definition.Expr)
					}
				}
			}
		}
	}
	return contexts
}

// Ordinary definitions return their label-scope completion for the caller to
// invoke after construction/registration. Named functions complete their own
// scopes after generating their body.
func checkDefinitionExpression(definition Definition, context map[string]Position, locals map[string]bool, generators ...*sanyExpressionGeneration) (Diagnostics, func() *sanyLabelTable) {
	_, ok := definition.Expr.(*FunctionExpr)
	if !definition.FunctionDef || !ok {
		return sanyExpressionGenerator(generators).checkDefinitionBody(definition, context, locals)
	}
	return append(checkDefinitionFunctionDomains(definition, context, locals, generators...), checkDefinitionFunctionBody(definition, context, locals, generators...)...), nil
}

func checkDefinitionFunctionDomains(definition Definition, context map[string]Position, locals map[string]bool, generators ...*sanyExpressionGeneration) Diagnostics {
	function, ok := definition.Expr.(*FunctionExpr)
	if !ok {
		return nil
	}
	var diags Diagnostics
	g := sanyExpressionGenerator(generators)
	closeContext := g.pushFormalContext(len(function.Bounds) + 1)
	for _, domain := range sanyFunctionDomainExpressions(function) {
		diags = append(diags, checkExpr(domain, context, locals, generators...)...)
	}
	boundLocals := copyBoolMap(locals)
	function.formalNodes = nil
	for _, bound := range function.Bounds {
		node := g.newFormalParameter(bound.Name, 0, bound.Pos, definition.Syntax)
		function.formalNodes = append(function.formalNodes, node)
		diags = append(diags, g.bindFormalParameter(node, context, boundLocals)...)
		boundLocals[bound.Name] = true
	}
	function.constructorSymbol, function.constructorSymbolExists = g.lookupSymbol(definition.Name, context)
	if !function.constructorSymbolExists {
		if builtin := g.initialBuiltin(definition.Name); builtin != nil {
			function.constructorSymbol = localSymbol{builtinNode: builtin, kind: OperatorDecl, arity: builtin.semArity(), pos: builtin.semPosition()}
			function.constructorSymbolExists = true
		}
	}
	function.functionSymbol = newSanyFormalParamNode(definition.Name, 0, definition.SourcePosition(), definition.Syntax, g.currentModule)
	if !function.constructorSymbolExists {
		node := function.functionSymbol
		_, generated := g.formalSymbolTable().registerSymbol(node)
		diags = append(diags, generated...)
		g.formals[definition.Name] = localSymbol{formalNode: node, kind: "FORMAL", arity: 0, pos: node.semPosition()}
	}
	// processFunction pops this context before constructing/resolving the
	// OpDefNode, then pushes the same context for an accepted body's generation.
	function.definitionFormalContext = g.formals
	function.definitionContext = g.formalSymbolTable().topContext()
	function.functionApplication = nil
	function.semanticGraph = nil
	closeContext()
	return diags
}

func checkDefinitionFunctionBody(definition Definition, context map[string]Position, locals map[string]bool, generators ...*sanyExpressionGeneration) Diagnostics {
	function, ok := definition.Expr.(*FunctionExpr)
	if !ok {
		return nil
	}
	var diags Diagnostics
	g := sanyExpressionGenerator(generators)
	previous := g.formals
	if function.definitionFormalContext != nil {
		g.formals = function.definitionFormalContext
	}
	if function.definitionContext != nil {
		g.formalSymbolTable().pushContext(function.definitionContext)
	}
	if function.functionApplication != nil {
		g.functions = append(g.functions, sanyFunctionGeneration{definition.Name, function.functionApplication})
	}
	finishLabels := g.pushLabelScope()
	popLabelFormals := g.pushLabelFormals(function.formalNodes)
	bodyLocals := copyBoolMap(locals)
	for _, bound := range function.Bounds {
		bodyLocals[bound.Name] = true
	}
	bodyLocals[definition.Name] = true
	diags = append(diags, g.checkExpr(function.Body, context, bodyLocals)...)
	popLabelFormals()
	labels := finishLabels()
	if definition.semanticNode != nil {
		definition.semanticNode.labels = labels
	}
	if function.functionApplication != nil {
		g.functions = g.functions[:len(g.functions)-1]
	}
	g.finishNamedFunction(function)
	if function.definitionContext != nil {
		g.formalSymbolTable().popContext()
	}
	g.formals = previous
	return diags
}

// selectorToNode resolves the named symbol before following body selectors.
// Keep only that named prefix when reporting an unavailable prepared selection.
func checkSanySelectedSymbol(expr Expr, selection *sanySelectorSelection, context map[string]Position, locals map[string]bool) Diagnostics {
	identifier := &IdentExpr{Name: selection.name, Pos: expr.Position()}
	if source := sanyExprSource(expr); source != nil && source.Selector != nil {
		selector := &SanySelector{Syntax: source.Selector.Syntax}
		var names []string
		for _, step := range source.Selector.Steps {
			if step.Kind != SanySelectorName {
				break
			}
			selector.Steps = append(selector.Steps, step)
			names = append(names, step.Name)
			if strings.Join(names, "!") == selection.name {
				break
			}
		}
		identifier.Selector = selector
	}
	return checkExpr(identifier, context, locals)
}

// processLetIn follows the LET body's syntax heirs just as generateModule does.
// Neither later operators nor later RECURSIVE declarations are predeclared.
type sanyLetGenerationUnit struct {
	syntax      *SanySyntaxNode
	position    Position
	declaration *Declaration
	definition  *Definition
	instance    *Instance
}

func sanyLetGenerationUnits(expr *LetExpr) []sanyLetGenerationUnit {
	var units []sanyLetGenerationUnit
	for i := range expr.Recursives {
		declaration := &expr.Recursives[i]
		units = append(units, sanyLetGenerationUnit{syntax: declaration.Syntax, position: declaration.Pos, declaration: declaration})
	}
	for i := range expr.Definitions {
		definition := &expr.Definitions[i]
		units = append(units, sanyLetGenerationUnit{syntax: definition.Syntax, position: definition.SourcePosition(), definition: definition})
	}
	for i := range expr.Instances {
		instance := &expr.Instances[i]
		units = append(units, sanyLetGenerationUnit{syntax: instance.Syntax, position: instance.SourcePosition(), instance: instance})
	}
	if expr.Syntax != nil {
		bySyntax := make(map[*SanySyntaxNode]sanyLetGenerationUnit, len(units))
		for _, unit := range units {
			bySyntax[unit.syntax] = unit
		}
		units = nil
		heirs := expr.Syntax.GetHeirs()
		if len(heirs) > 1 {
			for _, syntax := range heirs[1].GetHeirs() {
				if unit, ok := bySyntax[syntax]; ok {
					units = append(units, unit)
				}
			}
		}
	} else {
		// Hand-built native ASTs retain their source positions without syntax.
		sort.SliceStable(units, func(i, j int) bool { return units[i].position.Compare(units[j].position) < 0 })
	}
	return units
}

// generateProof processes every DEFINE heir in lexical order, including module
// definitions interleaved with operator and function definitions.
func sanyProofDefinitionUnits(step *ProofStep) []sanyLetGenerationUnit {
	var units []sanyLetGenerationUnit
	for i := range step.Definitions {
		definition := &step.Definitions[i]
		units = append(units, sanyLetGenerationUnit{syntax: definition.Syntax, position: definition.SourcePosition(), definition: definition})
	}
	for i := range step.Instances {
		instance := &step.Instances[i]
		units = append(units, sanyLetGenerationUnit{syntax: instance.Syntax, position: instance.SourcePosition(), instance: instance})
	}
	if step.Syntax == nil {
		sort.SliceStable(units, func(i, j int) bool { return units[i].position.Compare(units[j].position) < 0 })
		return units
	}
	bySyntax := make(map[*SanySyntaxNode]sanyLetGenerationUnit, len(units))
	for _, unit := range units {
		bySyntax[unit.syntax] = unit
	}
	units = nil
	for _, body := range step.Syntax.GetHeirs() {
		if body.Kind.JavaName() == "N_DefStep" {
			for _, syntax := range body.GetHeirs() {
				if unit, ok := bySyntax[syntax]; ok {
					units = append(units, unit)
				}
			}
		}
	}
	// A non-local INSTANCE step contains a single instance rather than DEFINE.
	if step.Kind == "INSTANCE" {
		for i := range step.Instances {
			instance := &step.Instances[i]
			units = append(units, sanyLetGenerationUnit{syntax: instance.Syntax, position: instance.SourcePosition(), instance: instance})
		}
	}
	return units
}

// One syntactic module unit can carry a named theorem definition, its theorem
// statement, and its proof. They belong to one Generator dispatch iteration.
type sanyModuleGenerationUnit struct {
	syntax      *SanySyntaxNode
	position    Position
	declaration *Declaration
	recursive   *Declaration
	instance    *Instance
	nested      *Module
	definition  *Definition
	assumption  *NamedExpr
	theorem     *NamedExpr
	references  []ProofRef
	proofs      []ProofSummary
}

func sanyModuleGenerationUnits(mod *Module) []*sanyModuleGenerationUnit {
	var units []*sanyModuleGenerationUnit
	bySyntax := map[*SanySyntaxNode]*sanyModuleGenerationUnit{}
	unit := func(syntax *SanySyntaxNode, position Position) *sanyModuleGenerationUnit {
		if syntax != nil {
			if existing := bySyntax[syntax]; existing != nil {
				return existing
			}
		}
		result := &sanyModuleGenerationUnit{syntax: syntax, position: position}
		units = append(units, result)
		if syntax != nil {
			bySyntax[syntax] = result
		}
		return result
	}
	for i := range mod.Declarations {
		declaration := &mod.Declarations[i]
		unit(declaration.Syntax, declaration.Pos).declaration = declaration
	}
	for i := range mod.Recursives {
		declaration := &mod.Recursives[i]
		unit(declaration.Syntax, declaration.Pos).recursive = declaration
	}
	for i := range mod.Instances {
		instance := &mod.Instances[i]
		unit(instance.Syntax, instance.SourcePosition()).instance = instance
	}
	for i := range mod.Definitions {
		definition := &mod.Definitions[i]
		unit(definition.Syntax, definition.SourcePosition()).definition = definition
	}
	for i := range mod.Assumptions {
		assumption := &mod.Assumptions[i]
		unit(assumption.Syntax, assumption.SourcePosition()).assumption = assumption
	}
	for i := range mod.Theorems {
		theorem := &mod.Theorems[i]
		unit(theorem.Syntax, theorem.SourcePosition()).theorem = theorem
	}
	for _, nested := range mod.Nested {
		unit(nested.Syntax, nested.Pos).nested = nested
	}
	for _, syntax := range mod.ProofRefNodes {
		unit(syntax, sanyNodePosition(syntax))
	}
	for _, reference := range mod.ProofRefs {
		current := unit(reference.Syntax, reference.Pos)
		current.references = append(current.references, reference)
	}
	for _, proof := range mod.Proofs {
		current := unit(proof.Syntax, proof.Pos)
		current.proofs = append(current.proofs, proof)
	}
	if mod.Syntax != nil {
		heirs := mod.Syntax.GetHeirs()
		ordered := make([]*sanyModuleGenerationUnit, 0, len(units))
		if len(heirs) > 2 {
			for _, syntax := range heirs[2].GetHeirs() {
				if unit := bySyntax[syntax]; unit != nil {
					ordered = append(ordered, unit)
				}
			}
		}
		return ordered
	}
	// Native callers can construct ASTs directly without parser nodes.
	sort.SliceStable(units, func(i, j int) bool { return units[i].position.Compare(units[j].position) < 0 })
	return units
}

// checkIfInRecursiveSection runs at the unit itself, before its body.
// INSTANCE and ordinary definitions are permitted inside recursive sections.
func sanyModuleRecursiveSectionType(unit *sanyModuleGenerationUnit) string {
	if unit.syntax != nil {
		switch unit.syntax.Kind.JavaName() {
		case "N_VariableDeclaration":
			return "A VARIABLE declaration"
		case "N_ParamDeclaration":
			return "A declaration"
		case "N_Theorem":
			return "A THEOREM"
		case "N_Assumption":
			return "An ASSUME"
		case "N_UseOrHide":
			return "A USE or HIDE"
		case "N_Module":
			return "A MODULE "
		}
		return ""
	}
	switch {
	case unit.declaration != nil && unit.declaration.Kind == VariableDecl:
		return "A VARIABLE declaration"
	case unit.declaration != nil:
		return "A declaration"
	case unit.assumption != nil:
		return "An ASSUME"
	case unit.theorem != nil:
		return "A THEOREM"
	case len(unit.references) > 0:
		return "A USE or HIDE"
	}
	return ""
}

// processQuantBoundArgs generates each syntactic domain once before creating
// any formal names. A multi-name or tuple bound shares one projected domain.
func sanyFunctionDomainExpressions(function *FunctionExpr) []Expr {
	var domains []Expr
	for _, bound := range function.Bounds {
		if len(domains) == 0 || domains[len(domains)-1] != bound.Set {
			domains = append(domains, bound.Set)
		}
	}
	return domains
}
