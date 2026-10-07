// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import (
	"fmt"
	"github.com/glycerine/tlago/tlc"
)

// SubstInNode keeps defaults in declaration enumeration order. WITH replaces
// an existing slot or appends a newly explicit substitution to that array.
type sanyGeneratedSubstitution struct {
	name        string
	target      substitutionTarget
	declaration *sanySemOpDeclNode
	expr        Expr
	implicit    bool
}

// processSubst constructs defaults first, generates each explicit RHS before
// duplicate detection, then checks remaining defaults and completeness.
func (g *sanyExpressionGeneration) generateInstanceSubstitutions(instance *Instance, module *Module, context map[string]Position) Diagnostics {
	previousLabelsEnabled := g.labelsEnabled
	g.labelsEnabled = true
	defer func() { g.labelsEnabled = previousLabelsEnabled }()
	instance.substitutionNode, instance.formalNodes = nil, nil
	target := g.spec.Modules[instance.Module]
	if target == nil {
		return nil
	}
	// getByClass filters Hashtable.elements(), not Context's insertion links.
	// Keep every context entry during rehashing; unrelated definitions and
	// builtin entries affect declaration enumeration too.
	targets := map[string]substitutionTarget{}
	var names []string
	entries := tlcBridgeContextContentOrder(tlcBridgeContextEntries(g.spec, target, map[*Module]bool{}))
	for _, entry := range entries {
		if entry.moduleKey || (entry.kind != ConstantDecl && entry.kind != VariableDecl) {
			continue
		}
		declaration, exists := moduleOwnSubstitutionTargets(entry.module)[entry.name]
		if entry.declaration != nil {
			declaration = substitutionTarget{Kind: entry.kind, Arity: entry.declaration.semArity(), Pos: entry.declaration.semPosition()}
			exists = true
		}
		if exists {
			targets[entry.name] = declaration
			names = append(names, entry.name)
		}
	}
	instance.generatedSubstitutions = make([]sanyGeneratedSubstitution, 0, len(names))
	declarations := map[string]*sanySemOpDeclNode{}
	for _, entry := range entries {
		declarations[entry.name] = entry.declaration
	}
	var diags Diagnostics
	context = copySanyExpressionContext(context)
	if len(instance.Params) > 0 {
		defer g.pushFormalContext(len(instance.Params))()
	}
	instance.formalNodes = make([]*sanyFormalParamNode, 0, len(instance.Params))
	for _, name := range instance.Params {
		position := instance.ParamPositions[name]
		node := g.newFormalParameter(name, instance.ParamArities[name], position, instance.Syntax)
		instance.formalNodes = append(instance.formalNodes, node)
		diags = append(diags, g.bindFormalParameter(node, context, nil)...)
	}
	base, _ := newSanySemSubstitutionNode(sanySubstInKind, instance.Syntax, make([]*sanySemSubst, 0), nil, module.semanticNode, target.semanticNode, false)
	template := &sanySemSubstInNode{base}
	complete := target.semanticNode != nil
	for _, name := range names {
		complete = complete && declarations[name] != nil
	}
	present := map[string]bool{}
	explicit := map[string]bool{}
	for _, name := range names {
		if symbol, exists := g.lookupSymbol(name, context); exists {
			present[name] = true
			arity := symbol.arity
			defaultExpr := &IdentExpr{Name: name, Pos: instance.SourcePosition(), formalNode: symbol.formalNode, declarationNode: symbol.declarationNode}
			defaultExpr.generationArity = &arity
			instance.generatedSubstitutions = append(instance.generatedSubstitutions, sanyGeneratedSubstitution{name: name, target: targets[name], declaration: declarations[name], expr: defaultExpr, implicit: true})
			actualSymbol := g.formalSymbolTable().resolveSymbol(name)
			if actualSymbol != nil && declarations[name] != nil {
				var node sanySemanticGraphNode
				if targets[name].Kind == VariableDecl || targets[name].Arity == 0 {
					application, generated, err := newSanySemOpApplNode(actualSymbol, make([]sanySemanticGraphNode, 0), instance.Syntax)
					if err != nil {
						panic(err)
					}
					node = application
					diags = append(diags, generated...)
				} else {
					node = newSanySemOpArgNode(actualSymbol, instance.Syntax, module.semanticNode)
				}
				defaultExpr.semanticGraph = node
				template.substs = append(template.substs, newSanySemSubst(declarations[name], node, nil, true))
			} else {
				complete = false
			}

			if actualSymbol == nil && targets[name].Arity == 0 && symbol.arity > 0 && symbol.formalNode == nil {
				position := instance.SourcePosition()
				if symbol.kind == ConstantDecl || symbol.kind == VariableDecl {
					diags = append(diags, sanyRegistrationDiagnostic(position, "E4004", "Operator used with the wrong number of arguments."))
				} else {
					diags = append(diags, sanyRegistrationDiagnostic(position, "E4004", "Wrong number of arguments (%d) given to operator '%s', \nwhich requires %d arguments.", 0, name, symbol.arity))
				}
			}
		}
	}
	for _, substitution := range instanceSubstitutions(*instance) {
		declaration, exists := targets[substitution.Name]
		if !exists {
			message := fmt.Sprintf("Identifier '%s' is not a legal target of a substitution. \nA legal target must be a declared CONSTANT or VARIABLE in the module being instantiated. \n(Also, check for warnings about multiple declarations of this same identifier.)", substitution.Name)
			diagnostic := sanyDiagnosticParameters(errorAt(substitution.Pos, "E4242", "%s", message), substitution.Name)
			diagnostic.SANYMessage = message
			diagnostic.SANYRange = SanyRange{Begin: substitution.Pos, End: substitution.Pos.SourceEnd()}
			diags = append(diags, diagnostic)
			continue
		}
		if declaration.Arity == 0 {
			diags = append(diags, g.proofExpression(substitution.Expr, context, nil)...)
		} else {
			owner := &IdentExpr{Name: substitution.Name, Pos: substitution.Pos}
			generated := g.generateOperatorOperand(owner, 0, declaration.Arity, substitution.Expr, context, nil)
			position := substitution.Expr.Position()
			failed := false
			lambda, isFunction := substitution.Expr.(*FunctionExpr)
			isLambda := isFunction && lambda.IsLambda
			for _, diagnostic := range generated {
				if isLambda && diagnostic.Code == "E4274" {
					continue
				}
				if diagnostic.Code == "E4270" {
					diagnostic.Code = "E4004"
					diagnostic.Message = fmt.Sprintf("Arity %d operator (not an expression) is expected \nto substitute for CONSTANT '%s'.", declaration.Arity, substitution.Name)
					diagnostic.SANYMessage = diagnostic.Message
					diagnostic.SANYParameters = []any{declaration.Arity, substitution.Name}
				}
				failed = failed || diagnostic.Severity == SeverityError
				diags = append(diags, diagnostic)
			}
			arities, _ := g.proofSignatures()
			operator := substitution.Expr
			if application, ok := operator.(*CallExpr); ok && (application.Selector == nil || len(application.Selector.Steps) <= 1) {
				operator = application.Callee
			}
			actual, known := operatorArgumentArity(operator, arities, nil)
			if isLambda {
				actual, known = len(lambda.Bounds), true
			}
			if failed && !isLambda {
				actual, known = 0, true // generateOpArg returns the nullOpArg sentinel.
				position = Position{}
			}
			if known && actual != declaration.Arity {
				message := fmt.Sprintf("An operator must be substituted for symbol '%s', and it must have arity %d.", substitution.Name, declaration.Arity)
				diagnostic := sanyDiagnosticParameters(errorAt(position, "E4243", "%s", message), substitution.Name, declaration.Arity)
				diagnostic.SANYMessage = message
				diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
				diags = append(diags, diagnostic)
			}
		}
		actual := sanyGeneratedExpressionNode(substitution.Expr)
		if actual == nil && sanyExpressionGenerationFailure(substitution.Expr) != sanyGenerationNullExpression {
			complete = false
		}
		canonicalMutation := declarations[substitution.Name] != nil && target.semanticNode != nil
		if canonicalMutation {
			var syntax *SanySyntaxNode
			if source := sanyGenerationSource(substitution.Expr); source != nil {
				syntax = source.Syntax
			}
			// Source mutation shares replaced Subst objects and isolates appended
			// arrays. It also owns duplicate reporting when its graph is available.
			diags = append(diags, template.addExplicitSubstitute(target.semanticNode.context, tlc.UniqueStringOf(substitution.Name), syntax, actual)...)
		} else {
			complete = false
		}
		if explicit[substitution.Name] {
			if !canonicalMutation {
				position := substitution.Expr.Position()
				diagnostic := sanyDiagnosticParameters(errorAt(position, "E4241", "Multiple substitutions for symbol '%s' in substitution.", substitution.Name), substitution.Name)
				diagnostic.SANYMessage = diagnostic.Message
				diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
				diags = append(diags, diagnostic)
			}
		} else {
			explicit[substitution.Name] = true
			present[substitution.Name] = true
			replacement := sanyGeneratedSubstitution{name: substitution.Name, target: declaration, declaration: declarations[substitution.Name], expr: substitution.Expr}
			replaced := false
			for i := range instance.generatedSubstitutions {
				if instance.generatedSubstitutions[i].name == substitution.Name {
					instance.generatedSubstitutions[i] = replacement
					replaced = true
					break
				}
			}
			if !replaced {
				instance.generatedSubstitutions = append(instance.generatedSubstitutions, replacement)
			}
		}
	}
	for _, name := range names {
		if present[name] && !explicit[name] && targets[name].Arity > 0 {
			symbol, _ := g.lookupSymbol(name, context)
			if symbol.arity != targets[name].Arity {
				message := fmt.Sprintf("An operator must be substituted for symbol '%s', and it must have arity %d.", name, targets[name].Arity)
				position := instance.SourcePosition()
				diagnostic := sanyDiagnosticParameters(errorAt(position, "E4243", "%s", message), name, targets[name].Arity)
				diagnostic.SANYMessage = message
				diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
				diags = append(diags, diagnostic)
			}
		}
	}
	for _, name := range names {
		if !present[name] {
			position := instance.SourcePosition()
			location := sanySymbolLocation(targets[name].Pos)
			message := fmt.Sprintf("Substitution missing for symbol %s declared at %s \nand instantiated in module %s.", name, location, module.Name)
			diagnostic := sanyDiagnosticParameters(errorAt(position, "E4240", "%s", message), name, location, module.Name)
			diagnostic.SANYMessage = message
			diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
			diags = append(diags, diagnostic)
		}
	}
	if complete {
		instance.substitutionNode = template
	}
	return diags
}

// processModuleDefinition registers imported definitions before its own
// ModuleInstanceKind symbol. Rejected registrations keep the earlier binding.
func (g *sanyExpressionGeneration) registerInstanceSymbols(instance Instance, context map[string]Position) (Diagnostics, []string) {
	var diags Diagnostics
	var added []string
	for _, symbol := range g.instanceSymbols(instance) {
		// Named instantiation constructs a new qualified definition at the
		// module definition syntax, while retaining its original source.
		if instance.Name != "" {
			symbol.source = instance.SourcePosition()
		}
		if previous, exists := g.lookupSymbol(symbol.name, context); exists {
			// SymbolTable compares kind and arity before allowing the same
			// original definition from a parameter-free source module.
			if previous.kind == symbol.importKind() && previous.arity == symbol.arity &&
				previous.instanceOrigin == symbol.origin && symbol.originSyntax != nil &&
				previous.instanceSyntax == symbol.originSyntax &&
				semanticModuleParameterFree(symbol.origin, g.spec, map[*Module]bool{}) {
				continue
			}
			conflict := symbol
			if previous.pos != symbol.sourcePosition() && !symbol.theoremLike {
				conflict.source = instance.SourcePosition()
			}
			diags = append(diags, instanceSymbolConflict(conflict, previous)...)
			continue
		}
		g.symbols[symbol.name] = localSymbol{instanceOrigin: symbol.origin, instanceSyntax: symbol.originSyntax, kind: symbol.importKind(), arity: symbol.arity, pos: symbol.sourcePosition(), operatorParams: symbol.operatorParams}
		context[symbol.name] = symbol.sourcePosition()
		added = append(added, symbol.name)
	}
	if instance.Name != "" {
		if previous, exists := g.lookupSymbol(instance.Name, context); exists {
			symbol := semanticExportedSymbol{name: instance.Name, kind: InstanceDecl, arity: len(instance.Params), source: instance.SourcePosition()}
			diags = append(diags, instanceSymbolConflict(symbol, previous)...)
		} else {
			g.symbols[instance.Name] = localSymbol{kind: InstanceDecl, arity: len(instance.Params), pos: instance.SourcePosition()}
			context[instance.Name] = instance.SourcePosition()
			added = append(added, instance.Name)
		}
		context[instanceNameSentinel(instance.Name)] = context[instance.Name]
	}
	return diags, added
}

// generateInstance and processModuleDefinition import OpDefNodes first, then
// ThmOrAssumpDefNodes, with Hashtable.elements() order within each class.
func (g *sanyExpressionGeneration) instanceSymbols(instance Instance) []semanticExportedSymbol {
	byName := map[string]semanticExportedSymbol{}
	for _, symbol := range semanticInstanceSymbols(instance, g.spec) {
		byName[symbol.name] = symbol
	}
	entries := tlcBridgeContextContentOrder(tlcBridgeContextEntries(g.spec, g.spec.Modules[instance.Module], map[*Module]bool{}))
	var symbols []semanticExportedSymbol
	for _, kind := range []DeclarationKind{OperatorDecl, ""} {
		for _, entry := range entries {
			if entry.kind != kind || entry.module == nil || entry.moduleKey || entry.local {
				continue
			}
			name := entry.name
			if instance.Name != "" {
				name = instance.Name + "!" + name
			}
			if symbol, exists := byName[name]; exists {
				symbols = append(symbols, symbol)
			}
		}
	}
	return symbols
}
