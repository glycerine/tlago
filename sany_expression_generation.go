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
	node     *sanySemOpDefNode
	name     string
	position Position
	arity    int
	level    int
	defined  bool
}

type sanyExpressionGeneration struct {
	labelsEnabled        bool
	labelScopes          []*sanyLabelScope
	labelExceptDepth     int
	labelAPDepth         int
	labelAPForbidden     bool
	labelGoalUnsupported bool
	functions            []sanyFunctionGeneration
	nodes                *sanyGeneratorNodes
	formalTable          *sanySymbolTable
	formals              map[string]localSymbol
	fact                 bool
	operatorArgument     bool
	symbolReferenceOnly  bool
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
	defer g.pushFormalContext(0)()
	letContext := g.formalSymbolTable().topContext()
	definitions := make([]sanySemSymbol, 0, len(expr.Definitions))
	completeGraph := len(expr.Instances) == 0
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
			if g.module.sum == 0 {
				g.module.section++
			}
			g.module.counts[g.level] += len(declaration.Names)
			g.module.sum += len(declaration.Names)
			for i, name := range declaration.Names {
				arity, _ := declarationArity(*declaration, name)
				node, generated := g.constructRecursiveDeclaration(*declaration, i, positions)
				diags = append(diags, generated...)
				binding := &sanyRecursiveBinding{node: node, name: name, position: declarationSymbolPosition(*declaration, name), arity: arity, level: g.level}
				g.declarations = append(g.declarations, binding)
				if node == nil {
					completeGraph = false
				} else if g.formalSymbolTable().resolveSymbol(node.semName()) != node {
					continue
				}
				if previous, exists := g.lookupSymbol(name, positions); exists {
					diags = append(diags, instanceSymbolConflict(semanticExportedSymbol{name: name, kind: OperatorDecl, arity: arity, source: binding.position}, previous)...)
				} else {
					g.bindings[name] = binding
					positions[name] = binding.position
					g.symbols[name] = localSymbol{opDefNode: node, kind: OperatorDecl, arity: arity, pos: binding.position}
					letLocals[name] = true
				}
			}
		case unit.definition != nil:
			diags = append(diags, g.generateLocalDefinition(unit.definition, positions, letLocals)...)
			if node := unit.definition.semanticNode; node != nil {
				definitions = append(definitions, node)
				if function, ok := unit.definition.Expr.(*FunctionExpr); ok && unit.definition.FunctionDef && function.semanticGraph == nil {
					completeGraph = false
				}
				if node.body == nil && sanyExpressionGenerationFailure(unit.definition.Expr) != sanyGenerationNullExpression {
					completeGraph = false
				}
				if !unit.definition.FunctionDef && g.currentModule != nil && g.currentModule.semanticNode != nil {
					g.currentModule.semanticNode.definitions = append(g.currentModule.semanticNode.definitions, node)
				}
			} else {
				completeGraph = false
			}
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
	if body := sanyGeneratedExpressionNode(expr.Body); completeGraph && body != nil {
		expr.semanticGraph = newSanySemLetInNode(expr.Syntax, definitions, make([]sanySemanticGraphNode, 0), body, letContext)
	}
	return diags
}

// processOperator and processFunction are shared by LET and proof DEFINE steps.
// The caller owns the lexical context and the Generator's current LET level.
func (g *sanyExpressionGeneration) generateLocalDefinition(definition *Definition, positions map[string]Position, letLocals map[string]bool) Diagnostics {
	definition.semanticNode = nil
	var diags Diagnostics
	diags = append(diags, checkDefinitionParams(*definition)...)
	collisionContext := copySanyExpressionContext(positions)
	for name, symbol := range g.formals {
		collisionContext[name] = symbol.pos
	}
	diags = append(diags, checkDefinitionParamCollisions(*definition, collisionContext, letLocals)...)
	bodyLocals := copyBoolMap(letLocals)
	for _, name := range definition.Params {
		bodyLocals[name] = true
	}
	// Domains precede resolving the function's own symbol.
	if definition.FunctionDef {
		diags = append(diags, checkDefinitionFunctionDomains(*definition, positions, bodyLocals, g)...)
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
		diags = append(diags, sanyRecursiveDefinitionDiagnostic(*definition, "E4293", fmt.Sprintf("Recursive %s %s defined at wrong LET/IN level.", kind, definition.Name)))
	} else if recursive {
		matches := len(definition.Params) == binding.arity
		if matches && len(definition.Params) > 0 {
			// Java overwrites paramsMatch in its loop; the final parameter
			// determines this check, rather than accumulating conjunctions.
			matches = definition.ParamArities[definition.Params[len(definition.Params)-1]] == 0
		}
		if !matches {
			message := fmt.Sprintf("Definition of %s has different arity than its RECURSIVE declaration.", definition.Name)
			if definition.FunctionDef {
				message = fmt.Sprintf("Function %s has operator arguments in its RECURSIVE declaration.", definition.Name)
			}
			diags = append(diags, sanyRecursiveDefinitionDiagnostic(*definition, "E4292", message))
		}
		if definition.FunctionDef && binding.arity == 0 {
			g.complete(binding, definition.SourcePosition())
		}
		if !definition.FunctionDef && binding.node == nil {
			binding.arity = len(definition.Params)
		}
	} else if _, exists := positions[definition.Name]; symbolExists || exists || letLocals[definition.Name] {
		message := fmt.Sprintf("Operator %s already defined or declared.", definition.Name)
		if definition.FunctionDef {
			message = fmt.Sprintf("Function name `%s' already defined or declared.", definition.Name)
		}
		diags = append(diags, sanyRecursiveDefinitionDiagnostic(*definition, "E4201", message))
	}
	if definition.FunctionDef {
		diags = append(diags, g.prepareNamedFunctionDefinition(definition)...)
		if wrongLevel || (symbolExists && previousSymbol.kind != OperatorDecl) {
			// processFunction still generates the body, but does not push its
			// quantifier context when resolving the symbol failed.
			function, _ := definition.Expr.(*FunctionExpr)
			if function != nil {
				diags = append(diags, g.checkRejectedNamedFunctionBody(*definition, positions, letLocals)...)
			}
		} else {
			if binding == nil && !symbolExists {
				letLocals[definition.Name] = true
				positions[definition.Name] = definition.SourcePosition()
				specs, _ := definitionOperatorParamSpecs(*definition)
				g.symbols[definition.Name] = localSymbol{opDefNode: definition.semanticNode, kind: OperatorDecl, arity: len(definition.Params), pos: definition.SourcePosition(), operatorParams: specs}
			}
			diags = append(diags, checkDefinitionFunctionBody(*definition, positions, bodyLocals, g)...)
		}
	} else {
		diags = append(diags, g.checkDefinitionBody(*definition, positions, bodyLocals)...)
		// processOperator constructs and registers only after its parameter
		// context has been popped. Do not reconstruct missing body symbols.
		if !symbolExists || previousSymbol.opDefNode != nil {
			diags = append(diags, g.constructOrdinaryDefinition(definition)...)
		}
		if recursive && !wrongLevel {
			g.complete(binding, definition.SourcePosition())
		}
		if wrongLevel && definition.semanticNode == nil {
			// The newly constructed OpDefNode calls SymbolTable.addSymbol; the
			// existing declaration remains the symbol table binding.
			diags = append(diags, instanceSymbolConflict(semanticExportedSymbol{name: definition.Name, kind: OperatorDecl, arity: len(definition.Params), source: definition.SourcePosition()}, localSymbol{kind: OperatorDecl, arity: binding.arity, pos: binding.position})...)
		} else if !recursive {
			if previous, exists := g.lookupSymbol(definition.Name, positions); exists {
				if definition.semanticNode == nil {
					diags = append(diags, instanceSymbolConflict(semanticExportedSymbol{name: definition.Name, kind: OperatorDecl, arity: len(definition.Params), source: definition.SourcePosition()}, previous)...)
				}
			} else {
				positions[definition.Name] = definition.SourcePosition()
				specs, _ := definitionOperatorParamSpecs(*definition)
				g.symbols[definition.Name] = localSymbol{opDefNode: definition.semanticNode, kind: OperatorDecl, arity: len(definition.Params), pos: definition.SourcePosition(), operatorParams: specs}
				letLocals[definition.Name] = true
			}
		}
	}
	return diags
}

func (g *sanyExpressionGeneration) checkDefinitionBody(definition Definition, context map[string]Position, locals map[string]bool) Diagnostics {
	defer g.pushFormalContext(len(definition.Params))()
	var diags Diagnostics
	parameters := make([]*sanyFormalParamNode, 0, len(definition.Params))
	for index, parameter := range sanyDefinitionParams(&definition) {
		// The source allocates the node before SymbolTable.addSymbol decides
		// whether the declaration can replace an existing binding.
		syntax := sanyDefinitionFormalSyntax(&definition, index)
		position := parameter.Pos
		if syntax != nil {
			position = sanyNodePosition(syntax)
		}
		node := newSanyFormalParamNode(parameter.Name, parameter.OperatorArity, position, syntax, g.currentModule)
		parameters = append(parameters, node)
		if _, exists := g.lookupSymbol(parameter.Name, context); exists {
			continue
		}
		if _, builtin := builtinOperatorArity(parameter.Name); builtin {
			continue
		}
		accepted, registrationDiags := g.formalSymbolTable().registerSymbol(node)
		diags = append(diags, registrationDiags...)
		if accepted {
			g.formals[parameter.Name] = localSymbol{formalNode: node, kind: "FORMAL", arity: parameter.OperatorArity, pos: parameter.Pos}
		}
	}
	// processOperator resolves its own name while its parameters are visible.
	// An accepted parameter with that name triggers the redefinition report,
	// even though it disappears before the new definition is registered.
	for _, parameter := range parameters {
		if parameter.semName() == definition.Name && g.formals[definition.Name].formalNode == parameter {
			diags = append(diags, sanyRecursiveDefinitionDiagnostic(definition, "E4201", fmt.Sprintf("Operator %s already defined or declared.", definition.Name)))
			break
		}
	}
	if source, ok := definition.Expr.(interface{ generationSource() *SanyExprSource }); ok {
		source.generationSource().definitionFormals = parameters
	}
	if binding := g.bindings[definition.Name]; binding != nil && binding.node != nil && !binding.node.defined && binding.node.letInLevel == g.level {
		// Source setParams replaces only the array, preserving declared arity
		// and constructor-sized level/Leibniz arrays even after a mismatch.
		if g.level == 0 && len(binding.node.formalNodes) == len(parameters) && len(parameters) > 0 && parameters[len(parameters)-1].semArity() != 0 {
			// Match Java's overwrite in the parameter loop, including its
			// final-parameter behavior for earlier higher-order formals.
			message := fmt.Sprintf("Definition of %s has different arity than its RECURSIVE declaration.", definition.Name)
			diags = append(diags, sanyRecursiveDefinitionDiagnostic(definition, "E4292", message))
		}
		binding.node.formalNodes = parameters
	}
	finishLabels := g.pushLabelScope()
	defer func() {
		labels := finishLabels()
		if source := sanyGenerationSource(definition.Expr); source != nil {
			source.definitionLabels = labels
		}
	}()
	return append(diags, g.checkExpr(definition.Expr, context, locals)...)
}

func (g *sanyExpressionGeneration) lookupSymbol(name string, context map[string]Position) (localSymbol, bool) {
	if symbol, exists := g.formals[name]; exists {
		return symbol, true
	}
	if binding := g.bindings[name]; binding != nil {
		return localSymbol{opDefNode: binding.node, kind: OperatorDecl, arity: binding.arity, pos: binding.position}, true
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
	setSanyExpressionGenerationFailure(argument, sanyGenerationSucceeded)
	if source := sanyGenerationSource(argument); source != nil {
		source.semanticGraph = nil
	}
	if expected <= 0 {
		diags := g.checkExpr(argument, context, locals)
		if sanyExpressionGenerationFailure(argument) == sanyGenerationNullOperator {
			g.retainNullOperatorOperand(argument, false)
		}
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
			g.retainNullOperatorOperand(argument, false)
			return Diagnostics{diagnostic}
		}
	}
	if function, ok := argument.(*FunctionExpr); ok && function.IsLambda {
		return g.generateLambdaOperand(owner, index, expected, function, context, locals)
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
				g.retainNullOperatorOperand(argument, false)
				return Diagnostics{diagnostic}
			}
		}
	}
	if identifier, ok := argument.(*IdentExpr); ok {
		if source := sanyGenerationSource(argument); source != nil && source.Syntax != nil {
			switch source.Syntax.Kind.JavaName() {
			case "N_GenInfixOp", "N_GenNonExpPrefixOp", "N_GenPostfixOp", "N_GenPrefixOp":
				// Qualified GenID prefixes may generate their own arguments;
				// that path still awaits canonical prefix graph integration.
				heirs := source.Syntax.GetHeirs()
				if len(heirs) > 0 && len(heirs[0].GetHeirs()) == 0 {
					return g.generateFixityOperatorOperand(owner, index, expected, identifier, context, locals)
				}
			}
		}
	}
	operator := argument
	// In the GeneralId operator-argument path selectorToNode does not
	// generate attached expression arguments; it resolves the operator.
	if call, ok := argument.(*CallExpr); ok && (call.Selector == nil || len(call.Selector.Steps) <= 1) {
		operator = call.Callee
	}
	// GeneralId selectors reject an incorrect expected arity before allocating
	// OpArg. The older fixity GenID path has different error/sentinel behavior.
	if source := sanyGenerationSource(argument); source != nil && source.Syntax != nil && source.Syntax.Kind.JavaName() == "N_GeneralId" && (source.Selector == nil || len(source.Selector.Steps) <= 1) {
		got, known := operatorArgumentArity(operator, g.moduleArities, locals)
		if identifier, ok := operator.(*IdentExpr); ok && !known {
			if symbol, exists := g.lookupSymbol(identifier.Name, context); exists && symbol.arity >= 0 {
				got, known = symbol.arity, true
			}
		}
		if literal, ok := operator.(*LiteralExpr); ok && literal.Kind == "bool" {
			got, known = 0, true
		}
		if known && got != expected {
			g.retainNullOperatorOperand(argument, false)
			diagnostic := sanyDiagnosticParameters(errorAt(position, "E4271", "operator argument arity mismatch: got %d, want %d", got, expected), expected, got)
			diagnostic.SANYRange = source.Syntax.Range
			diagnostic.SANYMessage = fmt.Sprintf("Expected arity %d but found operator of arity %d.", expected, got)
			return Diagnostics{diagnostic}
		}
	}
	previousOperatorArgument := g.operatorArgument
	g.operatorArgument = true
	diags := g.checkExpr(operator, context, locals)
	g.operatorArgument = previousOperatorArgument
	if sanyExpressionGenerationFailure(operator) != sanyGenerationSucceeded {
		g.retainNullOperatorOperand(argument, false)
		return diags
	}
	if operator != argument {
		if source := sanyGenerationSource(argument); source != nil {
			source.semanticGraph = sanyGeneratedExpressionNode(operator)
		}
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

	if source := sanyGenerationSource(argument); source != nil {
		source.semanticGraph = nil
		if source.Syntax != nil && source.Syntax.Kind.JavaName() == "N_GeneralId" {
			g.retainNullOperatorOperand(argument, false)
		}
	}
	return append(diags, diagnostic)
}

// generateExprOrOpArg produces all operands before OpDefNode.match validates
// the resulting operator arguments. A failed operand is the source nullOAN.
func (g *sanyExpressionGeneration) generateApplicationOperands(call *CallExpr, identifier *IdentExpr, operator sanySemSymbol, specs []operatorParamSpec, context map[string]Position, locals map[string]bool) Diagnostics {
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
			if generated[j].Code == "E4275" {
				generated[j].Message = fmt.Sprintf("expression parameter %s cannot accept a LAMBDA operator argument", specs[i].Name)
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
	if operator != nil {
		complete := true
		for _, argument := range call.Args {
			complete = complete && (sanyGeneratedExpressionNode(argument) != nil || sanyExpressionGenerationFailure(argument) == sanyGenerationNullExpression)
		}
		if complete {
			diags = append(diags, retainSanyMatchedApplication(call, operator, call.Args)...)
			call.operatorArgumentsGenerated = true
			return diags
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
	restore := g.pushFormalContext(len(parameters))
	defer func() {
		if restore != nil {
			restore()
		}
	}()
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
	popLabelFormals := g.pushLabelFormals(root.quantifierFormals)
	diags = append(diags, g.checkExpr(body, context, bodyLocals)...)
	popLabelFormals()
	restore()
	restore = nil
	operand := sanyGeneratedExpressionNode(body)
	if operand != nil {
		if root.Set == nil {
			operator := "$UnboundedForall"
			switch root.Kind {
			case "\\E":
				operator = "$UnboundedExists"
			case "\\EE":
				operator = "$TemporalExists"
			case "\\AA":
				operator = "$TemporalForall"
			}
			root.semanticGraph = newSanySemUnboundedOpApplNode(operator, []sanySemanticGraphNode{operand}, root.quantifierFormals, root.Syntax)
		} else {
			operator := "$BoundedForall"
			if root.Kind == "\\E" {
				operator = "$BoundedExists"
			}
			bounds := make([]BoundVar, len(parameters))
			for i, parameter := range parameters {
				bounds[i] = BoundVar{Name: parameter.Var, Set: parameter.Set, TupleBound: parameter.TupleBound}
			}
			retainSanyBoundApplication(root, operator, bounds, root.quantifierFormals, body)
		}
	}
	return diags
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
	popLabelFormals := g.pushLabelFormals(expr.formalNodes)
	diags = append(diags, g.checkExpr(expr.Body, context, bodyLocals)...)
	popLabelFormals()
	operand := sanyGeneratedExpressionNode(expr.Body)
	if operand != nil {
		if expr.Set == nil {
			expr.semanticGraph = newSanySemUnboundedOpApplNode("$UnboundedChoose", []sanySemanticGraphNode{operand}, expr.formalNodes, expr.Syntax)
		} else {
			// CHOOSE always has one bound group, even for tuple parameters.
			domain := sanyGeneratedExpressionNode(expr.Set)
			if domain != nil {
				expr.semanticGraph = newSanySemBoundedOpApplNode("$BoundedChoose", nil, []sanySemanticGraphNode{operand}, [][]*sanyFormalParamNode{expr.formalNodes}, []bool{expr.TupleVars != nil}, []sanySemanticGraphNode{domain}, expr.Syntax)
			}
		}
	}
	return diags
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
	defer g.pushLabelFormals(nodes)()
	return nodes, append(diags, g.checkExpr(body, context, bodyLocals)...)
}

// processRcdForms creates the field string before checking each earlier label,
// generates the value, then creates that field's pair before the next field.
type sanyRecordGenerationField struct {
	name     string
	value    Expr
	position Position
}

func (g *sanyExpressionGeneration) checkRecordForm(expr Expr, operator string, fields []sanyRecordGenerationField, context map[string]Position, locals map[string]bool) Diagnostics {
	var diagnostics Diagnostics
	source := sanyGenerationSource(expr)
	var children []*SanySyntaxNode
	if source != nil && source.Syntax != nil {
		children = source.Syntax.GetHeirs()
	}
	pairs := make([]sanySemanticGraphNode, len(fields))
	// Errors.addMessage deduplicates equal code/location/format/parameters.
	// This diagnostic has a fixed code/format and no parameters, so only
	// its actual source location distinguishes repeated reports here.
	redefinitions := make(map[SanyRange]bool)
	for i, field := range fields {
		var syntax *SanySyntaxNode
		var label *tlc.StringNode
		if 2*i+1 < len(children) {
			syntax = children[2*i+1]
			parts := syntax.GetHeirs()
			if len(parts) > 0 {
				label = tlc.NewStringNode(field.name)
				bridge := tlcBridge{}
				bridge.withSyntaxNode(parts[0], label)
				bridge.withPositionLocation(sanyNodePosition(parts[0]), label)
			}
		}
		for j := 0; j < i; j++ {
			if field.name == fields[j].name {
				diagnostic := sanyDuplicateRecordFieldDiagnostic(field.name, field.position, fields[j].position)
				if !redefinitions[diagnostic.SANYRange] {
					diagnostics = append(diagnostics, diagnostic)
					redefinitions[diagnostic.SANYRange] = true
				}
			}
		}
		diagnostics = append(diagnostics, g.checkExpr(field.value, context, locals)...)
		if value := sanyGeneratedExpressionNode(field.value); label != nil && value != nil {
			pairs[i] = newSanySemBuiltInOpApplNode("$Pair", []sanySemanticGraphNode{label, value}, syntax)
		}
	}
	retainSanyGeneratedOperands(expr, operator, pairs)
	if node, ok := sanyGeneratedExpressionNode(expr).(*sanySemOpApplNode); ok && g.currentModule != nil && g.currentModule.semanticNode != nil {
		g.currentModule.semanticNode.addRecord(node)
	}
	return diagnostics
}

// The actual body/formals are retained by generation. Construction happens only
// after the parameter scope has been popped, and registers in the current table.
func (g *sanyExpressionGeneration) constructOrdinaryDefinition(definition *Definition) Diagnostics {
	if sanyExpressionGenerationFailure(definition.Expr) == sanyGenerationNullOperator {
		g.retainNullOperatorOperand(definition.Expr, false)
	}
	source := sanyGenerationSource(definition.Expr)
	if source == nil || (source.semanticGraph == nil && sanyExpressionGenerationFailure(definition.Expr) != sanyGenerationNullExpression) {
		if binding := g.bindings[definition.Name]; binding != nil && binding.node != nil && !binding.node.defined && binding.node.letInLevel == g.level {
			// Body generation completed on the native path, but its canonical
			// graph is still unported. Preserve source completion/syntax on
			// the actual declaration without fabricating a body node.
			g.endRecursiveDefinition(binding.node, nil, definition.Syntax)
			if source != nil {
				binding.node.labels = source.definitionLabels
			}
			definition.semanticNode = binding.node
		}
		return nil
	}
	var module *sanySemModuleNode
	if g.currentModule != nil {
		module = g.currentModule.semanticNode
	}
	if binding := g.bindings[definition.Name]; binding != nil && binding.node != nil && !binding.node.defined && binding.node.letInLevel == g.level {
		g.endRecursiveDefinition(binding.node, source.semanticGraph, definition.Syntax)
		binding.node.labels = source.definitionLabels
		definition.semanticNode = binding.node
		return nil
	}
	node, diagnostics := newSanySemOpDefNode(definition.Name, sanyUserDefinedOpKind, source.definitionFormals, definition.Local, source.semanticGraph, module, g.formalSymbolTable(), definition.Syntax, true, nil)
	g.setDefinitionRecursionFields(node)
	node.labels = source.definitionLabels
	definition.semanticNode = node
	return diagnostics
}

// Failed operands use this Generator's source sentinel, not a fresh node.
func (g *sanyExpressionGeneration) retainNullOperatorOperand(expr Expr, operatorArgument bool) {
	if source := sanyGenerationSource(expr); source != nil && g.nodes != nil {
		if operatorArgument {
			source.semanticGraph = g.nodes.nullOpArg
		} else {
			source.semanticGraph = g.nodes.nullOAN
		}
	}
}

func (g *sanyExpressionGeneration) generateLambdaOperand(owner *IdentExpr, index, expected int, function *FunctionExpr, context map[string]Position, locals map[string]bool) Diagnostics {
	function.lambdaNode = nil
	parameters, diagnostics := g.checkBoundExpression(function.Bounds, function.Syntax, context, locals, function.Body)
	function.formalNodes = parameters
	if sanyExpressionGenerationFailure(function.Body) == sanyGenerationNullOperator {
		g.retainNullOperatorOperand(function.Body, false)
	}
	body := sanyGeneratedExpressionNode(function.Body)
	if body == nil && sanyExpressionGenerationFailure(function.Body) != sanyGenerationNullExpression {
		return diagnostics
	}
	var module *sanySemModuleNode
	if g.currentModule != nil {
		module = g.currentModule.semanticNode
	}
	// generateLambda pops its formal context before constructing an unregistered
	// OpDef. It does not set ordinary definition recursion/LET fields.
	node, generated := newSanySemOpDefNode("LAMBDA", sanyUserDefinedOpKind, parameters, false, body, module, nil, function.Syntax, true, nil)
	function.lambdaNode = node
	diagnostics = append(diagnostics, generated...)
	if node.semArity() == expected {
		function.semanticGraph = newSanySemOpArgNode(node, function.Syntax, module)
		return diagnostics
	}
	g.retainNullOperatorOperand(function, true)
	diagnostic := sanyDiagnosticParameters(errorAt(owner.Pos, "E4274", "operator argument arity mismatch: got %d, want %d", node.semArity(), expected), node.semArity(), index+1, owner.Name, expected)
	diagnostic.SANYRange = SanyRange{Begin: owner.Pos, End: owner.Pos.SourceEnd()}
	diagnostic.SANYMessage = fmt.Sprintf("Lambda expression with arity %d used as argument %d of operator `%s', \nbut an operator of arity %d is required.", node.semArity(), index+1, owner.Name, expected)
	return append(diagnostics, diagnostic)
}

// The SANY1 GenID branch resolves before allocating OpArg. Unlike GeneralId,
// wrong arity reports at the enclosing application and returns nullOpArg.
func (g *sanyExpressionGeneration) generateFixityOperatorOperand(owner *IdentExpr, index, expected int, argument *IdentExpr, context map[string]Position, locals map[string]bool) Diagnostics {
	name := argument.Name
	argument.Name = ResolveSanyOperatorSynonym(sanyOperatorGenIDName(argument.Syntax))
	previousOperatorArgument, previousReferenceOnly := g.operatorArgument, g.symbolReferenceOnly
	g.operatorArgument, g.symbolReferenceOnly = true, true
	diagnostics := g.checkExpr(argument, context, locals)
	g.operatorArgument, g.symbolReferenceOnly = previousOperatorArgument, previousReferenceOnly
	operator := g.applicationOperator(argument.Name, nil, context)
	argument.Name = name
	if sanyExpressionGenerationFailure(argument) != sanyGenerationSucceeded {
		// GenID.finalAppend logs a declaration lookup failure rather than
		// selectorToNode's unknown-operator diagnostic.
		for i := range diagnostics {
			if diagnostics[i].Code == "E4200" {
				rawName := sanyOperatorGenIDName(argument.Syntax)
				diagnostic := sanyDiagnosticParameters(errorAt(argument.Pos, "E4004", "undefined operator %s", rawName), rawName)
				diagnostic.SANYRange = argument.Syntax.Range
				diagnostic.SANYMessage = fmt.Sprintf("Could not find declaration or definition of symbol '%s'.", rawName)
				diagnostics[i] = diagnostic
			}
		}
		g.retainNullOperatorOperand(argument, false)
		return diagnostics
	}
	if operator == nil {
		return diagnostics
	}
	if operator.semArity() == expected {
		return append(diagnostics, retainSanySymbolReference(argument, operator, true, g.currentModule)...)
	}
	g.retainNullOperatorOperand(argument, true)
	mainOperator := g.applicationOperator(owner.Name, nil, context)
	if mainOperator == nil {
		// Its source location/string requires an actual resolved SymbolNode.
		return diagnostics
	}
	format := "Operator with incorrect arity passed as argument. \nOperator '%s' of arity %s is argument number %s (counting from 1) to operator `%s', \nbut an operator of arity %s was expected."
	parameters := []any{operator.semName(), operator.semArity(), index + 1, mainOperator, expected}
	diagnostic := sanyDiagnosticParameters(errorAt(owner.Pos, "E4004", "operator argument arity mismatch: got %d, want %d", operator.semArity(), expected), parameters...)
	diagnostic.SANYMessage = fmt.Sprintf(format, operator.semName(), fmt.Sprint(operator.semArity()), fmt.Sprint(index+1), mainOperator.semBase().Location.String(), fmt.Sprint(expected))
	diagnostic.SANYRange = SanyRange{Begin: owner.Pos, End: owner.Pos.SourceEnd()}
	return append(diagnostics, diagnostic)
}
