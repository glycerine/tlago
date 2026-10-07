// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2022, Oracle and/or its affiliates.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import (
	"fmt"

	"github.com/glycerine/tlago/tlc"
)

// Generator retains declaration identities across LET contexts. In particular,
// leaving a LET does not discard its entries from ModuleNode.recursiveDecls.
type sanyRecursiveBinding struct {
	name     string
	position Position
	arity    int
	level    int
	defined  bool
}

type sanyExpressionGeneration struct {
	nodes                *sanyGeneratorNodes
	formals              map[string]localSymbol
	fact                 bool
	operatorArgument     bool
	level                int
	spec                 *Spec
	currentModule        *Module
	module               *sanyModuleRecursiveGeneration
	declarations         []*sanyRecursiveBinding
	bindings             map[string]*sanyRecursiveBinding
	symbols              map[string]localSymbol
	moduleKinds          map[string]DeclarationKind
	moduleArities        map[string]int
	moduleOperatorParams map[string][]operatorParamSpec
	moduleSymbols        map[string]localSymbol
}

func sanyExpressionGenerator(generators []*sanyExpressionGeneration) *sanyExpressionGeneration {
	if len(generators) != 0 && generators[0] != nil {
		return generators[0]
	}
	return &sanyExpressionGeneration{module: &sanyModuleRecursiveGeneration{}, bindings: map[string]*sanyRecursiveBinding{}}
}

func (g *sanyExpressionGeneration) complete(binding *sanyRecursiveBinding, position Position) {
	binding.defined = true
	binding.position = position
	if g.level == 0 {
		g.module.count--
	} else {
		g.module.counts[g.level]--
	}
	g.module.sum--
	if g.module.sum < 0 {
		panic(tlc.NewWrongInvocationException("Defined more recursive operators than were declared in RECURSIVE statements."))
	}
}

func sanyRecursiveDefinitionDiagnostic(definition Definition, code, message string) Diagnostic {
	position := definition.SourcePosition()
	diagnostic := sanyDiagnosticParameters(errorAt(position, code, "%s", message), definition.Name)
	diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
	diagnostic.SANYMessage = message
	if code == "E4293" {
		diagnostic.Message = fmt.Sprintf("recursive declaration %s is defined in the wrong LET/IN level", definition.Name)
	}
	return diagnostic
}

func sanyUndefinedRecursiveDiagnostic(binding *sanyRecursiveBinding) Diagnostic {
	diagnostic := sanyDiagnosticParameters(errorAt(binding.position, "E4291", "recursive declaration %s has no definition", binding.name), binding.name)
	diagnostic.SANYRange = SanyRange{Begin: binding.position, End: binding.position.SourceEnd()}
	diagnostic.SANYMessage = fmt.Sprintf("Symbol %s declared in RECURSIVE statement but not defined.", binding.name)
	return diagnostic
}

func (g *sanyExpressionGeneration) checkLet(expr *LetExpr, context map[string]Position, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	expr.instanceDefinitions = nil
	instanceResolver := &sanySelectorResolver{spec: g.spec, scopes: map[*Module]map[string]sanySelectorDefinition{}, visiting: map[*Module]bool{}}
	defer func() {
		if failure := recover(); failure != nil {
			if g.spec != nil {
				g.spec.SemanticDiags = appendSanyDiagnostics(append(Diagnostics(nil), diags...), g.spec.SemanticDiags...)
			}
			panic(failure)
		}
	}()
	// processLetIn raises the level for the definitions, then lowers it before
	// generating IN while keeping the same symbol context on the stack.
	g.level++
	if g.level >= len(g.module.counts) {
		panic(tlc.NewArrayIndexOutOfBoundsException(g.level, len(g.module.counts)))
	}
	g.module.counts[g.level] = 0
	previousSymbols := g.symbols
	g.symbols = make(map[string]localSymbol, len(previousSymbols)+len(expr.Definitions))
	for name, symbol := range previousSymbols {
		g.symbols[name] = symbol
	}
	defer func() { g.symbols = previousSymbols }()
	previousBindings := g.bindings
	g.bindings = make(map[string]*sanyRecursiveBinding, len(previousBindings))
	for name, binding := range previousBindings {
		g.bindings[name] = binding
	}
	defer func() { g.bindings = previousBindings }()
	letLocals := copyBoolMap(locals)
	positions := copySanyExpressionContext(context)
	for _, unit := range sanyLetGenerationUnits(expr) {
		switch {
		case unit.declaration != nil:
			declaration := unit.declaration
			g.module.counts[g.level] += len(declaration.Names)
			g.module.sum += len(declaration.Names)
			for _, name := range declaration.Names {
				arity, _ := declarationArity(*declaration, name)
				binding := &sanyRecursiveBinding{name: name, position: declarationSymbolPosition(*declaration, name), arity: arity, level: g.level}
				g.declarations = append(g.declarations, binding)
				if previous, exists := g.lookupSymbol(name, positions); exists {
					diags = append(diags, instanceSymbolConflict(semanticExportedSymbol{name: name, kind: OperatorDecl, arity: arity, source: binding.position}, previous)...)
				} else {
					g.bindings[name] = binding
					positions[name] = binding.position
					g.symbols[name] = localSymbol{kind: OperatorDecl, arity: arity, pos: binding.position}
					letLocals[name] = true
				}
			}
		case unit.definition != nil:
			diags = append(diags, g.generateLocalDefinition(*unit.definition, positions, letLocals)...)
		case unit.instance != nil:
			instance := unit.instance
			diags = append(diags, g.generateInstanceSubstitutions(instance, g.currentModule, positions)...)
			registered, names := g.registerInstanceSymbols(*instance, positions)
			diags = append(diags, registered...)
			definitions := map[string]sanySelectorDefinition{}
			instanceResolver.addInstance(definitions, g.currentModule, *instance)
			for _, name := range names {
				if g.symbols[name].kind != InstanceDecl {
					if definition, exists := definitions[name]; exists {
						expr.instanceDefinitions = append(expr.instanceDefinitions, definition)
					}
					letLocals[name] = true
				}
			}
		}
	}
	if g.module.counts[g.level] > 0 {
		for _, binding := range g.declarations {
			if binding.level == g.level && !binding.defined {
				diags = append(diags, sanyUndefinedRecursiveDiagnostic(binding))
			}
		}
		g.module.sum -= g.module.counts[g.level]
	}
	g.level--
	diags = append(diags, g.checkExpr(expr.Body, positions, letLocals)...)
	return diags
}

// processOperator and processFunction are shared by LET and proof DEFINE steps.
// The caller owns the lexical context and the Generator's current LET level.
func (g *sanyExpressionGeneration) generateLocalDefinition(definition Definition, positions map[string]Position, letLocals map[string]bool) Diagnostics {
	var diags Diagnostics
	diags = append(diags, checkDefinitionParams(definition)...)
	diags = append(diags, checkDefinitionParamCollisions(definition, positions, letLocals)...)
	bodyLocals := copyBoolMap(letLocals)
	for _, name := range definition.Params {
		bodyLocals[name] = true
	}
	// Domains precede resolving the function's own symbol.
	if definition.FunctionDef {
		diags = append(diags, checkDefinitionFunctionDomains(definition, positions, bodyLocals, g)...)
	}
	previousSymbol, symbolExists := g.lookupSymbol(definition.Name, positions)
	if function, ok := definition.Expr.(*FunctionExpr); ok && definition.FunctionDef {
		previousSymbol, symbolExists = function.constructorSymbol, function.constructorSymbolExists
	}
	binding := g.bindings[definition.Name]
	recursive := binding != nil && !binding.defined
	wrongLevel := recursive && binding.level != g.level
	if wrongLevel {
		kind := "operator"
		if definition.FunctionDef {
			kind = "function"
		}
		diags = append(diags, sanyRecursiveDefinitionDiagnostic(definition, "E4293", fmt.Sprintf("Recursive %s %s defined at wrong LET/IN level.", kind, definition.Name)))
	} else if recursive {
		matches := len(definition.Params) == binding.arity
		for _, arity := range definition.ParamArities {
			matches = matches && arity == 0
		}
		if !matches {
			message := fmt.Sprintf("Definition of %s has different arity than its RECURSIVE declaration.", definition.Name)
			if definition.FunctionDef {
				message = fmt.Sprintf("Function %s has operator arguments in its RECURSIVE declaration.", definition.Name)
			}
			diags = append(diags, sanyRecursiveDefinitionDiagnostic(definition, "E4292", message))
		}
		if definition.FunctionDef && binding.arity == 0 {
			g.complete(binding, definition.SourcePosition())
		}
		if !definition.FunctionDef {
			binding.arity = len(definition.Params)
		}
	} else if _, exists := positions[definition.Name]; symbolExists || exists || letLocals[definition.Name] {
		message := fmt.Sprintf("Operator %s already defined or declared.", definition.Name)
		if definition.FunctionDef {
			message = fmt.Sprintf("Function name `%s' already defined or declared.", definition.Name)
		}
		diags = append(diags, sanyRecursiveDefinitionDiagnostic(definition, "E4201", message))
	}
	if definition.FunctionDef {
		if wrongLevel || (symbolExists && previousSymbol.kind != OperatorDecl) {
			// processFunction still generates the body, but does not push its
			// quantifier context when resolving the symbol failed.
			function, _ := definition.Expr.(*FunctionExpr)
			if function != nil {
				diags = append(diags, g.checkExpr(function.Body, positions, letLocals)...)
			}
		} else {
			if binding == nil && !symbolExists {
				letLocals[definition.Name] = true
				positions[definition.Name] = definition.SourcePosition()
				specs, _ := definitionOperatorParamSpecs(definition)
				g.symbols[definition.Name] = localSymbol{kind: OperatorDecl, arity: len(definition.Params), pos: definition.SourcePosition(), operatorParams: specs}
			}
			diags = append(diags, checkDefinitionFunctionBody(definition, positions, bodyLocals, g)...)
		}
	} else {
		diags = append(diags, g.checkDefinitionBody(definition, positions, bodyLocals)...)
		if recursive && !wrongLevel {
			g.complete(binding, definition.SourcePosition())
		}
		if wrongLevel {
			// The newly constructed OpDefNode calls SymbolTable.addSymbol; the
			// existing declaration remains the symbol table binding.
			diags = append(diags, instanceSymbolConflict(semanticExportedSymbol{name: definition.Name, kind: OperatorDecl, arity: len(definition.Params), source: definition.SourcePosition()}, localSymbol{kind: OperatorDecl, arity: binding.arity, pos: binding.position})...)
		} else if !recursive {
			if previous, exists := g.lookupSymbol(definition.Name, positions); exists {
				diags = append(diags, instanceSymbolConflict(semanticExportedSymbol{name: definition.Name, kind: OperatorDecl, arity: len(definition.Params), source: definition.SourcePosition()}, previous)...)
			} else {
				positions[definition.Name] = definition.SourcePosition()
				specs, _ := definitionOperatorParamSpecs(definition)
				g.symbols[definition.Name] = localSymbol{kind: OperatorDecl, arity: len(definition.Params), pos: definition.SourcePosition(), operatorParams: specs}
				letLocals[definition.Name] = true
			}
		}
	}
	return diags
}

func (g *sanyExpressionGeneration) checkDefinitionBody(definition Definition, context map[string]Position, locals map[string]bool) Diagnostics {
	previous := g.formals
	g.formals = map[string]localSymbol{}
	for name, symbol := range previous {
		g.formals[name] = symbol
	}
	defer func() { g.formals = previous }()
	var parameters []*sanyFormalParamNode
	for _, parameter := range sanyDefinitionParams(&definition) {
		// The source allocates the node before SymbolTable.addSymbol decides
		// whether the declaration can replace an existing binding.
		node := g.newFormalParameter(parameter.Name, parameter.OperatorArity, parameter.Pos, definition.Syntax)
		parameters = append(parameters, node)
		if _, exists := g.lookupSymbol(parameter.Name, context); exists {
			continue
		}
		if _, builtin := builtinOperatorArity(parameter.Name); builtin {
			continue
		}
		g.formals[parameter.Name] = localSymbol{formalNode: node, kind: "FORMAL", arity: parameter.OperatorArity, pos: parameter.Pos}
	}
	if source, ok := definition.Expr.(interface{ generationSource() *SanyExprSource }); ok {
		source.generationSource().definitionFormals = parameters
	}
	return g.checkExpr(definition.Expr, context, locals)
}

func (g *sanyExpressionGeneration) lookupSymbol(name string, context map[string]Position) (localSymbol, bool) {
	if symbol, exists := g.formals[name]; exists {
		return symbol, true
	}
	if binding := g.bindings[name]; binding != nil {
		return localSymbol{kind: OperatorDecl, arity: binding.arity, pos: binding.position}, true
	}
	if symbol, ok := g.symbols[name]; ok {
		return symbol, true
	}
	position, ok := context[name]
	if !ok {
		return localSymbol{}, false
	}
	if symbol, exists := g.moduleSymbols[name]; exists {
		return symbol, true
	}
	kind := g.moduleKinds[name]
	if kind == RecursiveDecl {
		kind = OperatorDecl
	}
	arity, known := g.moduleArities[name]
	if !known {
		arity = -1
	}
	return localSymbol{kind: kind, arity: arity, pos: position}, true
}

// generateExprOrOpArg selects the expression or operator-argument path from
// the receiving formal parameter's arity, before incomplete-name validation.
func (g *sanyExpressionGeneration) generateOperatorOperand(owner *IdentExpr, index, expected int, argument Expr, context map[string]Position, locals map[string]bool) Diagnostics {
	if expected <= 0 {
		diags := g.checkExpr(argument, context, locals)
		arities, parameters := g.proofSignatures()
		diags = append(diags, checkCallArity(argument, arities, parameters, locals)...)
		return append(diags, checkOperatorArgumentKinds(argument, parameters, arities, locals)...)
	}
	position := argument.Position()
	if source, ok := argument.(interface{ GetSyntaxNode() *SanySyntaxNode }); ok && source.GetSyntaxNode() != nil {
		rangeOfArgument := source.GetSyntaxNode().Range
		position = rangeOfArgument.Begin
		position.EndLine, position.EndColumn = rangeOfArgument.End.Line, rangeOfArgument.End.Column
		switch source.GetSyntaxNode().Kind.JavaName() {
		case "N_GeneralId", "N_GenInfixOp", "N_GenNonExpPrefixOp", "N_GenPostfixOp", "N_GenPrefixOp", "N_Lambda":
		default:
			diagnostic := sanyDiagnosticParameters(errorAt(position, "E4270", "operator parameter requires an operator argument of arity %d", expected), index+1, owner.Name)
			diagnostic.SANYRange = rangeOfArgument
			diagnostic.SANYMessage = fmt.Sprintf("An expression appears as argument number %d (counting from 1) to operator '%s', in a position an operator is required.", index+1, owner.Name)
			return Diagnostics{diagnostic}
		}
	}
	if source := sanyExprSource(argument); source != nil && source.Selector != nil {
		for i, step := range source.Selector.Steps {
			// Final application arguments are outside the GeneralId operand;
			// selectorToNode examines the GeneralId's prefix arguments.
			if i == len(source.Selector.Steps)-1 {
				break
			}
			if step.Kind == SanySelectorName && step.Arguments != nil {
				position := sanyNodePosition(step.Syntax)
				message := fmt.Sprintf("Selector `%s' should not have argument(s).", step.Name)
				diagnostic := errorAt(position, "E4005", "%s", message)
				diagnostic.SANYMessage = message
				diagnostic.SANYRange = step.Syntax.Range
				setSanyExpressionGenerationFailure(argument, sanyGenerationNullOperator)
				return Diagnostics{diagnostic}
			}
		}
	}
	operator := argument
	// In the GeneralId operator-argument path selectorToNode does not
	// generate attached expression arguments; it resolves the operator.
	if call, ok := argument.(*CallExpr); ok && (call.Selector == nil || len(call.Selector.Steps) <= 1) {
		operator = call.Callee
	}
	previousOperatorArgument := g.operatorArgument
	g.operatorArgument = true
	diags := g.checkExpr(operator, context, locals)
	g.operatorArgument = previousOperatorArgument
	if sanyExpressionGenerationFailure(operator) != sanyGenerationSucceeded {
		return diags
	}
	got, ok := operatorArgumentArity(operator, g.moduleArities, locals)
	if identifier, isIdentifier := operator.(*IdentExpr); !ok && isIdentifier {
		if symbol, exists := g.lookupSymbol(identifier.Name, context); exists {
			got, ok = symbol.arity, true
		}
	}
	if literal, isLiteral := argument.(*LiteralExpr); isLiteral && literal.Kind == "bool" {
		got, ok = 0, true
	}
	if !ok || got == expected {
		return diags
	}
	diagnostic := sanyDiagnosticParameters(errorAt(position, "E4271", "operator argument arity mismatch: got %d, want %d", got, expected), expected, got)
	diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
	diagnostic.SANYMessage = fmt.Sprintf("Expected arity %d but found operator of arity %d.", expected, got)
	if function, isFunction := argument.(*FunctionExpr); isFunction && function.IsLambda {
		diagnostic = sanyDiagnosticParameters(errorAt(owner.Pos, "E4274", "operator argument arity mismatch: got %d, want %d", got, expected), got, index+1, owner.Name, expected)
		diagnostic.SANYRange = SanyRange{Begin: owner.Pos, End: owner.Pos.SourceEnd()}
		diagnostic.SANYMessage = fmt.Sprintf("Lambda expression with arity %d used as argument %d of operator `%s', \nbut an operator of arity %d is required.", got, index+1, owner.Name, expected)
	}
	return append(diags, diagnostic)
}

// generateExprOrOpArg produces all operands before OpDefNode.match validates
// the resulting operator arguments. A failed operand is the source nullOAN.
func (g *sanyExpressionGeneration) generateApplicationOperands(call *CallExpr, identifier *IdentExpr, specs []operatorParamSpec, context map[string]Position, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	owner := *identifier
	owner.Syntax = call.Syntax
	var invalid []int
	arities, parameters := g.proofSignatures()
	for i, argument := range call.Args {
		generated := g.generateOperatorOperand(&owner, i, specs[i].Arity, argument, context, locals)
		for j := range generated {
			if generated[j].Code == "E4270" {
				generated[j].Message = fmt.Sprintf("operator parameter %s requires an operator argument of arity %d", specs[i].Name, specs[i].Arity)
			}
		}
		diags = append(diags, generated...)
		if specs[i].Arity > 0 {
			operator := argument
			if application, ok := argument.(*CallExpr); ok && (application.Selector == nil || len(application.Selector.Steps) <= 1) {
				operator = application.Callee
			}
			got, known := operatorArgumentArity(operator, arities, locals)
			if name, ok := operator.(*IdentExpr); ok && !known {
				if symbol, exists := g.lookupSymbol(name.Name, context); exists && symbol.arity >= 0 {
					got, known = symbol.arity, true
				}
			}
			if !known || got != specs[i].Arity || sanyExpressionGenerationFailure(operator) != sanyGenerationSucceeded || operatorArgumentRequiresOperatorParam(operator, parameters, locals) {
				invalid = append(invalid, i)
			}
		}
	}
	for _, i := range invalid {
		message := fmt.Sprintf("Argument number %d to operator '%s' \nshould be a %d-parameter operator.", i+1, identifier.Name, specs[i].Arity)
		diagnostic := sanyDiagnosticParameters(errorAt(call.Pos, "E4271", "%s", message), i+1, identifier.Name, specs[i].Arity)
		diagnostic.SANYMessage = message
		if call.Syntax != nil {
			diagnostic.SANYRange = call.Syntax.Range
		} else {
			diagnostic.SANYRange = SanyRange{Begin: call.Pos, End: call.Pos.SourceEnd()}
		}
		diags = append(diags, diagnostic)
	}
	call.operatorArgumentsGenerated = true
	return diags
}

// Source processQuantBoundArgs generates every domain before introducing any
// quantified formal. Unbounded quantifiers use the same parameter scope without
// domains. Flattened wrappers from one source node retain its identity so an
// explicitly nested quantifier remains a separate scope.
func (g *sanyExpressionGeneration) checkQuantifier(root *QuantifierExpr, context map[string]Position, locals map[string]bool) Diagnostics {
	parameters, body := sanyQuantifierGroup(root)
	var diags Diagnostics
	seenDomains := map[Expr]bool{}
	for _, parameter := range parameters {
		if parameter.Set != nil && !seenDomains[parameter.Set] {
			diags = append(diags, g.proofExpression(parameter.Set, context, locals)...)
			seenDomains[parameter.Set] = true
		}
	}
	defer g.pushFormalContext(len(parameters))()
	bodyLocals := copyBoolMap(locals)
	root.quantifierFormals = nil
	for _, parameter := range parameters {
		position := parameter.VarPos
		if position.Line == 0 {
			position = parameter.Pos
		}
		node := g.newFormalParameter(parameter.Var, 0, position, parameter.Syntax)
		parameter.formalNode = node
		root.quantifierFormals = append(root.quantifierFormals, node)
		diags = append(diags, g.bindFormalParameter(node, context, bodyLocals)...)
		bodyLocals[parameter.Var] = true
	}
	return append(diags, g.checkExpr(body, context, bodyLocals)...)
}

func (g *sanyExpressionGeneration) checkChoose(expr *ChooseExpr, context map[string]Position, locals map[string]bool) Diagnostics {
	// processChoose generates the domain before allocating its formals.
	diags := g.checkExpr(expr.Set, context, locals)
	bounds := expr.boundVars()
	defer g.pushFormalContext(len(bounds))()
	bodyLocals := copyBoolMap(locals)
	expr.formalNodes = nil
	for _, bound := range bounds {
		position := bound.Pos
		if expr.Set == nil && expr.TupleVars == nil && expr.Syntax != nil {
			// The source unbounded scalar constructor receives
			// children[0] (the CHOOSE token), rather than the identifier.
			if heirs := expr.Syntax.GetHeirs(); len(heirs) != 0 {
				position = sanyNodePosition(heirs[0])
			}
		}
		node := g.newFormalParameter(bound.Name, 0, position, expr.Syntax)
		expr.formalNodes = append(expr.formalNodes, node)
		diags = append(diags, g.bindFormalParameter(node, context, bodyLocals)...)
		bodyLocals[bound.Name] = true
	}
	return append(diags, g.checkExpr(expr.Body, context, bodyLocals)...)
}

// processFcnConst/processSetOfAll/processSubsetOf generate domains before
// allocating formals, then generate the body in the resulting symbol context.
// generateLambda uses the same scope and constructor order without domains.
func (g *sanyExpressionGeneration) checkBoundExpression(bounds []BoundVar, syntax *SanySyntaxNode, context map[string]Position, locals map[string]bool, body Expr) ([]*sanyFormalParamNode, Diagnostics) {
	var diags Diagnostics
	seenDomains := map[Expr]bool{}
	for _, bound := range bounds {
		if bound.Set != nil && !seenDomains[bound.Set] {
			diags = append(diags, g.checkExpr(bound.Set, context, locals)...)
			seenDomains[bound.Set] = true
		}
	}
	defer g.pushFormalContext(len(bounds))()
	bodyLocals := copyBoolMap(locals)
	nodes := make([]*sanyFormalParamNode, 0, len(bounds))
	for _, bound := range bounds {
		node := g.newFormalParameter(bound.Name, 0, bound.Pos, syntax)
		nodes = append(nodes, node)
		diags = append(diags, g.bindFormalParameter(node, context, bodyLocals)...)
		bodyLocals[bound.Name] = true
	}
	return nodes, append(diags, g.checkExpr(body, context, bodyLocals)...)
}
