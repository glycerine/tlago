// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import "fmt"

// processSubst constructs defaults first, generates each explicit RHS before
// duplicate detection, then checks remaining defaults and completeness.
func (g *sanyExpressionGeneration) generateProofInstanceSubstitutions(instance Instance, module *Module, context map[string]Position) Diagnostics {
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
		if declaration, exists := moduleOwnSubstitutionTargets(entry.module)[entry.name]; exists {
			targets[entry.name] = declaration
			names = append(names, entry.name)
		}
	}
	var diags Diagnostics
	previousFormals := g.formals
	g.formals = map[string]localSymbol{}
	for name, symbol := range previousFormals {
		g.formals[name] = symbol
	}
	defer func() { g.formals = previousFormals }()
	context = copySanyExpressionContext(context)
	for _, name := range instance.Params {
		position := instance.ParamPositions[name]
		node := g.newFormalParameter(name, instance.ParamArities[name], position, instance.Syntax)
		if previous, exists := g.lookupSymbol(name, context); exists {
			diagnostic := errorAt(position, "E4201", "Multiply-defined symbol '%s': this definition or declaration conflicts \nwith the one at %s.", name, sanySymbolLocation(previous.pos))
			diagnostic.SANYMessage = diagnostic.Message
			diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
			diags = append(diags, diagnostic)
		} else {
			g.formals[name] = localSymbol{formalNode: node, kind: "FORMAL", arity: instance.ParamArities[name], pos: position}
			context[name] = position
		}
	}
	present := map[string]bool{}
	explicit := map[string]bool{}
	for _, name := range names {
		if symbol, exists := g.lookupSymbol(name, context); exists {
			present[name] = true
			if targets[name].Arity == 0 {
				identifier := &IdentExpr{Name: name, Pos: instance.SourcePosition()}
				arity := symbol.arity
				identifier.generationArity = &arity
				arities, parameters := g.proofSignatures()
				diags = append(diags, checkCallArity(identifier, arities, parameters, nil)...)
			}
		}
	}
	for _, substitution := range instanceSubstitutions(instance) {
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
		if explicit[substitution.Name] {
			position := substitution.Expr.Position()
			diagnostic := sanyDiagnosticParameters(errorAt(position, "E4241", "Multiple substitutions for symbol '%s' in substitution.", substitution.Name), substitution.Name)
			diagnostic.SANYMessage = diagnostic.Message
			diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
			diags = append(diags, diagnostic)
		} else {
			explicit[substitution.Name] = true
			present[substitution.Name] = true
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
	return diags
}

// generateInstance and processModuleDefinition import OpDefNodes first, then
// ThmOrAssumpDefNodes, with Hashtable.elements() order within each class.
func (g *sanyExpressionGeneration) proofInstanceSymbols(instance Instance) []semanticExportedSymbol {
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
