// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2022, Oracle and/or its affiliates.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import (
	"fmt"
	"github.com/glycerine/tlago/tlc"
)

func sanyProofReferenceDirect(reference ProofRef) bool {
	if source, ok := reference.Expr.(interface{ GetSyntaxNode() *SanySyntaxNode }); ok && source.GetSyntaxNode() != nil {
		node := source.GetSyntaxNode()
		return node.Kind.JavaName() == "N_GeneralId" || (node.Token != nil && isSanyProofStepStartKind(node.Token.Kind))
	}
	_, identifier := reference.Expr.(*IdentExpr)
	return identifier
}

// generateUseOrHide supplies both leaf BY and USE/HIDE facts and definitions.
// Only a direct GeneralId is generated with isFact; its operands are expressions.
// The result flag records whether source generateUseOrHide appends a vector
// entry. Failed expression facts occupy a slot; unavailable modules do not.
func (g *sanyExpressionGeneration) generateProofReference(reference ProofRef, module *Module, context map[string]Position, locals map[string]bool) (Diagnostics, bool) {
	if reference.Module != "" {
		target := g.spec.Modules[reference.Module]
		if reference.Module == module.Name || (target != nil && !(sameSourceFile(target.Pos, reference.Pos) && positionBefore(reference.Pos, target.Pos))) {
			return nil, true
		}
		diagnostic := errorAt(reference.Pos, "E4005", "module %s is unavailable in proof", reference.Module)
		diagnostic.SANYMessage = fmt.Sprintf("Module `%s' used without being extended or instantiated.", reference.Module)
		diagnostic.SANYRange = SanyRange{Begin: reference.Pos, End: reference.Pos.SourceEnd()}
		return Diagnostics{diagnostic}, false
	}
	if reference.Defs {
		var diags Diagnostics
		if source := sanyExprSource(reference.Expr); source != nil && source.Syntax != nil && source.Syntax.Kind.JavaName() == "N_OpApplication" {
			diagnostic := errorAt(reference.Pos, "E4003", "Was expecting a GeneralId.")
			diagnostic.SANYMessage = diagnostic.Message
			diagnostic.SANYRange = source.Syntax.Range
			diags = append(diags, diagnostic)
			heirs := source.Syntax.GetHeirs()
			if len(heirs) > 1 {
				diagnostic := errorAt(sanyNodePosition(heirs[1]), "E4005", "Selector `!(...)' should not have argument(s).")
				diagnostic.SANYMessage = diagnostic.Message
				diagnostic.SANYRange = heirs[1].Range
				diags = append(diags, diagnostic)
			}
		}
		if source := sanyGenerationSource(reference.Expr); len(diags) == 0 && source != nil && source.Syntax != nil && source.Syntax.Kind.JavaName() == "N_GeneralId" {
			selector := sanyExprSource(reference.Expr)
			if selector == nil || selector.Selector == nil || len(selector.Selector.Steps) <= 1 {
				if symbol := g.formalSymbolTable().resolveSymbol(reference.Name); symbol != nil {
					valid := symbol.semKind() == sanyUserDefinedOpKind || symbol.semKind() == sanyModuleInstanceKind || symbol.semKind() == sanyThmOrAssumpDefKind && (len(symbol.semName()) == 0 || symbol.semName()[0] != '<')
					if definition, ok := symbol.(*sanySemOpDefNode); ok && definition.semKind() == sanyNumberedProofStepKind {
						if definition.stepNode == nil {
							panic(tlc.NewNullPointerException(""))
						}
						if definition.stepNode.Kind() == tlc.SemanticKind(sanyDefStepKind) {
							return nil, true
						}
						first := errorAt(reference.Pos, "E4004", "DEF clause entry refers to a non-definition step.")
						first.SANYMessage, first.SANYRange = first.Message, source.Syntax.Range
						second := errorAt(reference.Pos, "E4200", "DEF clause entry should describe a defined operator.")
						second.SANYMessage, second.SANYRange = second.Message, source.Syntax.Range
						return Diagnostics{first, second}, false
					}
					if valid {
						return nil, true
					}
					diagnostic := errorAt(reference.Pos, "E4200", "DEF clause entry should describe a defined operator.")
					diagnostic.SANYMessage = diagnostic.Message
					diagnostic.SANYRange = source.Syntax.Range
					// selectorToNode and generateUseOrHide each diagnose this.
					return Diagnostics{diagnostic, diagnostic}, false
				}
			}
		}
		name := reference.Name
		if len(diags) != 0 {
			// genIdToSelector has diagnosed the malformed selector.
		} else if selected := sanyExprSelection(reference.Expr); selected != nil {
			// DEF entries must end at a named operator, not an expression operand.
			if len(selected.lets) > 0 && reference.Expr != nil {
				if source := sanyExprSource(reference.Expr); source != nil && source.Selector != nil {
					last := source.Selector.Steps[len(source.Selector.Steps)-1]
					if last.Kind == SanySelectorName {
						return nil, true
					}
				}
			}
		} else if source := sanyExprSource(reference.Expr); source != nil && source.selectorFailure {
			diags = append(diags, g.checkExpr(reference.Expr, context, locals)...)
		} else if symbol, exists := g.lookupSymbol(name, context); exists {
			if symbol.proofStepKind == "DEFINE" {
				return nil, true
			}
			if symbol.proofStepKind != "" {
				diagnostic := errorAt(reference.Pos, "E4004", "DEF clause entry refers to a non-definition step.")
				diagnostic.SANYMessage = diagnostic.Message
				diagnostic.SANYRange = SanyRange{Begin: reference.Pos, End: reference.Pos.SourceEnd()}
				diags = append(diags, diagnostic)
			}
			if (symbol.kind == OperatorDecl || symbol.kind == InstanceDecl || symbol.kind == semanticTheoremImportKind) && (len(name) == 0 || name[0] != '<') {
				return nil, true
			}
		} else if !builtinIdentifiers[name] {
			if _, builtin := builtinOperatorArity(name); !builtin {
				identifier := &IdentExpr{Name: name, Pos: reference.Pos}
				if source := sanyExprSource(reference.Expr); source != nil {
					identifier.SanyExprSource = *source
				}
				diags = append(diags, sanyUndefinedIdentifierDiagnostic(identifier))
			}
		}
		diagnostic := errorAt(reference.Pos, "E4200", "DEF clause entry should describe a defined operator.")
		diagnostic.SANYRange = SanyRange{Begin: reference.Pos, End: reference.Pos.SourceEnd()}
		diagnostic.SANYMessage = diagnostic.Message
		return append(diags, diagnostic), false
	}
	if reference.Expr == nil {
		return nil, true
	}
	if identifier, ok := reference.Expr.(*IdentExpr); ok {
		if symbol, exists := g.lookupSymbol(identifier.Name, context); exists && symbol.proofStepKind == "DEFINE" {
			diagnostic := errorAt(reference.Pos, "E4004", "Step number of non-fact used as a fact")
			diagnostic.SANYMessage = diagnostic.Message
			diagnostic.SANYRange = SanyRange{Begin: reference.Pos, End: reference.Pos.SourceEnd()}
			return Diagnostics{diagnostic}, true
		}
	}
	previousFact := g.fact
	g.fact = sanyProofReferenceDirect(reference)
	diags := g.checkExpr(reference.Expr, context, locals)
	g.fact = previousFact
	arities, parameters := g.proofSignatures()
	diags = append(diags, checkCallArity(reference.Expr, arities, parameters, locals)...)
	return append(diags, checkOperatorArgumentKinds(reference.Expr, parameters, arities, locals)...), true
}

// The enclosing theorem's NEW context is visible in its proof. A statement's
// NEW context remains visible in its subproof; SUFFICES retains it afterward.
func sanyProofNewSymbols(context map[string]Position, body *AssumeProve) {
	if body == nil {
		return
	}
	for _, clause := range body.Assumptions {
		if clause.NewSymbol != nil {
			context[clause.NewSymbol.Name] = clause.NewSymbol.Pos
		}
	}
}

func (g *sanyExpressionGeneration) proofReferences(proof ProofSummary, module *Module, context map[string]Position) Diagnostics {
	var diags Diagnostics
	previousGoalUnsupported := g.labelGoalUnsupported
	g.labelGoalUnsupported = true
	defer func() { g.labelGoalUnsupported = previousGoalUnsupported }()
	defer func() {
		if failure := recover(); failure != nil {
			g.spec.SemanticDiags = appendSanyDiagnostics(append(Diagnostics(nil), diags...), g.spec.SemanticDiags...)
			panic(failure)
		}
	}()
	previousSymbols := g.symbols
	defer func() { g.symbols = previousSymbols }()
	g.symbols = map[string]localSymbol{}
	for name, symbol := range previousSymbols {
		g.symbols[name] = symbol
	}
	base := copySanyExpressionContext(context)
	for name := range base {
		if symbol, exists := g.lookupSymbol(name, base); exists {
			base[name] = symbol.pos
		}
	}
	for _, theorem := range module.Theorems {
		if theorem.Syntax == proof.Syntax {
			sanyProofNewSymbols(base, theorem.AssumeProveBody)
			sanyProofNewBindings(g.symbols, theorem.AssumeProveBody)
		}
	}
	diags = append(diags, g.leafProofReferences(proof.LeafRefs, proof.Syntax, module, base)...)
	baseSymbols := g.symbols
	symbolScopes := map[int]map[string]localSymbol{}

	scopes := map[int]map[string]Position{}
	pendingSuffices := map[int]*AssumeProve{}
	pendingPicks := map[int]map[string]localSymbol{}
	previousInfixRHS := map[int]Expr{}
	graphs := newSanyProofGraphGeneration(g, proof.Syntax)
	for stepIndex, step := range proof.Steps {
		graphs.beforeStep(step.Syntax)
		if graphs != nil {
			diags = append(diags, graphs.diagnostics...)
			graphs.diagnostics = nil
		}
		for depth := range previousInfixRHS {
			if depth > step.Depth {
				delete(previousInfixRHS, depth)
			}
		}
		if step.Kind != "ASSERT" || step.AssumeProveBody != nil {
			delete(previousInfixRHS, step.Depth)
		}
		for depth, bounds := range pendingPicks {
			if step.Depth <= depth {
				if scopes[depth] == nil {
					scopes[depth] = map[string]Position{}
				}
				if symbolScopes[depth] == nil {
					symbolScopes[depth] = map[string]localSymbol{}
				}
				for name, symbol := range bounds {
					scopes[depth][name] = symbol.pos
					symbolScopes[depth][name] = symbol
				}
				delete(pendingPicks, depth)
			}
		}
		// A SUFFICES NEW context is installed only after its whole subproof.
		for depth, body := range pendingSuffices {
			if step.Depth <= depth {
				if scopes[depth] == nil {
					scopes[depth] = map[string]Position{}
				}
				if symbolScopes[depth] == nil {
					symbolScopes[depth] = map[string]localSymbol{}
				}
				sanyProofNewSymbols(scopes[depth], body)
				sanyProofNewBindings(symbolScopes[depth], body)
				delete(pendingSuffices, depth)
			}
		}
		for depth := range scopes {
			if depth > step.Depth {
				delete(scopes, depth)
				delete(symbolScopes, depth)
			}
		}
		g.symbols = map[string]localSymbol{}
		for name, symbol := range baseSymbols {
			g.symbols[name] = symbol
		}
		for depth := 0; depth <= step.Depth; depth++ {
			for name, symbol := range symbolScopes[depth] {
				g.symbols[name] = symbol
			}
		}
		if symbolScopes[step.Depth] == nil {
			symbolScopes[step.Depth] = map[string]localSymbol{}
		}
		current := copySanyExpressionContext(base)
		for depth := 0; depth <= step.Depth; depth++ {
			for name, pos := range scopes[depth] {
				current[name] = pos
			}
		}
		if scopes[step.Depth] == nil {
			scopes[step.Depth] = map[string]Position{}
		}
		if step.Implicit && step.Name != "" {
			diagnostic := errorAt(step.Pos, "E4350", "implicit proof step cannot have name %s", step.Name)
			diagnostic.SANYMessage = "<*> and <+> cannot be used for a named step."
			diagnostic.SANYRange = SanyRange{Begin: step.Pos, End: step.Pos.SourceEnd()}
			diags = append(diags, diagnostic)
		}
		for _, unit := range sanyProofDefinitionUnits(&step) {
			if unit.definition != nil {
				definition := *unit.definition
				previous, exists := g.lookupSymbol(definition.Name, current)
				binding := g.bindings[definition.Name]
				failedFunction := definition.FunctionDef && ((exists && previous.kind != OperatorDecl && previous.kind != InstanceDecl) || (binding != nil && !binding.defined && binding.level != g.level))
				diags = append(diags, g.generateLocalDefinition(unit.definition, current, map[string]bool{})...)
				if failedFunction {
					panic(tlc.NewArrayIndexOutOfBoundsExceptionNoMessage())
				}
				locals := map[string]bool{}
				for _, name := range definition.Params {
					locals[name] = true
				}
				arities, parameters := g.proofSignatures()
				arities = definitionBodyArities(arities, definition)
				diags = append(diags, checkCallArity(definition.Expr, arities, parameters, locals)...)
				diags = append(diags, checkOperatorArgumentKinds(definition.Expr, parameters, arities, locals)...)
				if symbol, exists := g.symbols[definition.Name]; exists {
					symbolScopes[step.Depth][definition.Name] = symbol
					scopes[step.Depth][definition.Name] = current[definition.Name]
				}
			} else if unit.instance != nil {
				instance := unit.instance
				diags = append(diags, g.generateInstanceSubstitutions(instance, module, current)...)
				diags = append(diags, g.generateUnnamedInstance(instance, false)...)
				registered, names := g.registerInstanceSymbols(*instance, current)
				diags = append(diags, registered...)
				for _, name := range names {
					symbolScopes[step.Depth][name] = g.symbols[name]
					scopes[step.Depth][name] = current[name]
				}
				if instance.Name != "" {
					sentinel := instanceNameSentinel(instance.Name)
					scopes[step.Depth][sentinel] = current[sentinel]
				}
			}
		}
		var introducedFormals map[string]localSymbol
		statementContext := copySanyExpressionContext(current)
		if step.AssumeProveBody != nil {
			if graphs != nil {
				diags = append(diags, graphs.assumeProveStatement(step.AssumeProveBody, statementContext)...)
			} else {
				g.labelGoalUnsupported = true
				diags = append(diags, checkAssumeProveBindings(step.AssumeProveBody, statementContext, nil, g)...)
			}
		} else if step.Kind == "PICK" || step.Kind == "TAKE" {
			var generated Diagnostics
			introducedFormals, generated = g.generateProofBinder(&proof.Steps[stepIndex], statementContext)
			diags = append(diags, generated...)
		} else {
			infix, isInfix := step.Expr.(*BinaryExpr)
			isInfix = isInfix && infix.Syntax != nil && infix.Syntax.Kind.JavaName() == "N_InfixExpr"
			if step.Kind == "ASSERT" && !step.Suffices && isInfix {
				if left, ok := infix.Left.(*IdentExpr); ok && left.Name == "@" && left.Syntax != nil && left.Syntax.Kind.JavaName() == "N_GeneralId" {
					heirs := left.Syntax.GetHeirs()
					if len(heirs) == 2 && len(heirs[0].GetHeirs()) == 0 && heirs[1].Kind.JavaName() == "IDENTIFIER" {
						left.proofAtTarget = previousInfixRHS[step.Depth]
					}
				}
			}
			if generated, handled := g.proofInfixExpression(step.Expr, statementContext); handled {
				diags = append(diags, generated...)
			} else {
				diags = append(diags, g.proofExpression(step.Expr, statementContext, nil)...)
			}
			if step.Kind == "ASSERT" && !step.Suffices {
				delete(previousInfixRHS, step.Depth)
				if isInfix && sanyExpressionGenerationFailure(step.Expr) == sanyGenerationSucceeded {
					previousInfixRHS[step.Depth] = infix.Right
				}
			}
		}
		for _, expression := range step.Exprs {
			diags = append(diags, g.proofExpression(expression, current, nil)...)
		}
		entries := 0
		builder := newSanyUseOrHideBuilder()
		for _, reference := range step.UseHideRefs {
			generated, appended := g.generateProofReference(reference, module, current, nil)
			diags = append(diags, generated...)
			builder.appendReference(g, reference, appended)
			if appended {
				entries++
			}
		}
		var useHide *sanySemUseOrHideNode
		if step.Kind == "USE" || step.Kind == "HIDE" {
			for _, body := range step.Syntax.GetHeirs() {
				if body.Kind.JavaName() == "N_UseOrHide" {
					useHide = builder.finish(body)
					diags = append(diags, sanyEmptyProofCommand(body, entries, "Empty USE or HIDE statement.")...)
				}
			}
		}
		if useHide != nil {
			diags = append(diags, useHide.factCheck()...)
		} else {
			for _, reference := range step.UseHideRefs {
				facts, steps := map[string]bool{}, map[string]bool{}
				for name := range current {
					if symbol, exists := g.lookupSymbol(name, current); exists {
						if symbol.kind == semanticTheoremImportKind {
							facts[name] = true
						}
						if symbol.proofStepKind != "" {
							steps[name] = true
						}
					}
				}
				diags = append(diags, checkHideRef(reference, facts, steps)...)
			}
		}
		diags = append(diags, graphs.statement(&proof.Steps[stepIndex], useHide, current)...)
		if step.AssumeProveBody != nil {
			diags = append(diags, checkAssumeProveLabels(step.AssumeProveBody, true)...)
		}
		if step.QualifiedName != "" {
			kind := step.Kind
			if step.Suffices {
				kind = "SUFFICES"
			}
			symbol := localSymbol{proofStepKind: kind, proofAssumeProve: step.AssumeProveBody != nil, kind: semanticTheoremImportKind, pos: step.Pos}
			if graphs != nil {
				switch actual := g.formalSymbolTable().resolveSymbol(step.QualifiedName).(type) {
				case *sanySemThmOrAssumpDefNode:
					symbol.theoremDefNode = actual
				case *sanySemOpDefNode:
					symbol.opDefNode = actual
				}
			}
			g.symbols[step.QualifiedName] = symbol
			symbolScopes[step.Depth][step.QualifiedName] = symbol
			current[step.QualifiedName] = step.Pos
			scopes[step.Depth][step.QualifiedName] = step.Pos
		}
		if step.Kind == "TAKE" {
			for name, symbol := range introducedFormals {
				current[name] = symbol.pos
				scopes[step.Depth][name] = symbol.pos
				g.symbols[name] = symbol
				symbolScopes[step.Depth][name] = symbol
			}
		}
		ownProof := copySanyExpressionContext(current)
		ownSymbols := g.symbols
		g.symbols = map[string]localSymbol{}
		for name, symbol := range ownSymbols {
			g.symbols[name] = symbol
		}
		if !step.Suffices {
			sanyProofNewSymbols(ownProof, step.AssumeProveBody)
			sanyProofNewBindings(g.symbols, step.AssumeProveBody)
		}
		diags = append(diags, g.leafProofReferences(step.LeafRefs, step.Syntax, module, ownProof)...)
		g.symbols = ownSymbols
		if step.Kind == "PICK" {
			pendingPicks[step.Depth] = introducedFormals
		}

		if step.AssumeProveBody != nil {
			depth := step.Depth + 1
			if step.Suffices {
				pendingSuffices[step.Depth] = step.AssumeProveBody
				continue
			}
			if scopes[depth] == nil {
				scopes[depth] = map[string]Position{}
			}
			sanyProofNewSymbols(scopes[depth], step.AssumeProveBody)
			if symbolScopes[depth] == nil {
				symbolScopes[depth] = map[string]localSymbol{}
			}
			sanyProofNewBindings(symbolScopes[depth], step.AssumeProveBody)
		}
	}
	graphs.finish()
	if graphs != nil {
		diags = append(diags, graphs.diagnostics...)
		graphs.diagnostics = nil
	}
	return diags
}

// Failed DEF and MODULE entries are omitted from the generated vectors. Failed
// expression facts still occupy a slot, as in generateUseOrHide's vec.addElement.
func (g *sanyExpressionGeneration) leafProofReferences(references []ProofRef, syntax *SanySyntaxNode, module *Module, context map[string]Position) Diagnostics {
	var diags Diagnostics
	entries := 0
	builder := newSanyUseOrHideBuilder()
	for _, reference := range references {
		generated, appended := g.generateProofReference(reference, module, context, nil)
		diags = append(diags, generated...)
		builder.appendReference(g, reference, appended)
		if appended {
			entries++
		}
	}
	if command := sanyLeafProofSyntax(syntax); command != nil {
		temporary := builder.finish(command)
		diags = append(diags, sanyEmptyProofCommand(command, entries, "Empty BY")...)
		if temporary != nil {
			proof := newSanySemLeafProofNode(command, temporary.facts, temporary.defs, false, temporary.isOnly)
			if g.leafProofGraphs == nil {
				g.leafProofGraphs = make(map[*SanySyntaxNode]*sanySemLeafProofNode)
			}
			g.leafProofGraphs[command] = proof
		}
	}
	return diags
}

func sanyEmptyProofCommand(syntax *SanySyntaxNode, entries int, emptyMessage string) Diagnostics {
	if entries != 0 || syntax == nil {
		return nil
	}
	heirs := syntax.GetHeirs()
	next := 1
	if len(heirs) > 0 && heirs[0].Token != nil && heirs[0].Token.Kind == SanyTokenProof {
		next++
	}
	var diags Diagnostics
	add := func(message string) {
		diagnostic := errorAt(sanyNodePosition(syntax), "E4003", "%s", message)
		diagnostic.SANYMessage = message
		diagnostic.SANYRange = syntax.Range
		diags = append(diags, diagnostic)
	}
	if next >= len(heirs) {
		add("Empty BY, USE, or HIDE")
	}
	add(emptyMessage)
	return diags
}

func sanyProofNewBindings(symbols map[string]localSymbol, body *AssumeProve) {
	if body == nil {
		return
	}
	for _, clause := range body.Assumptions {
		if symbol := clause.NewSymbol; symbol != nil {
			if symbol.bindingSymbol != nil {
				symbols[symbol.Name] = *symbol.bindingSymbol
				continue
			}
			kind := ConstantDecl
			if symbol.Kind == 25 {
				kind = VariableDecl
			}
			symbols[symbol.Name] = localSymbol{kind: kind, arity: symbol.Arity, pos: symbol.Pos}
		}
	}
}

func (g *sanyExpressionGeneration) proofExpression(expr Expr, context map[string]Position, locals map[string]bool) Diagnostics {
	diags := g.checkExpr(expr, context, locals)
	arities, parameters := g.proofSignatures()
	diags = append(diags, checkCallArity(expr, arities, parameters, locals)...)
	return append(diags, checkOperatorArgumentKinds(expr, parameters, arities, locals)...)
}

func (g *sanyExpressionGeneration) proofSignatures() (map[string]int, map[string][]operatorParamSpec) {
	arities := copyIntMap(g.moduleArities)
	parameters := map[string][]operatorParamSpec{}
	for name, specs := range g.moduleOperatorParams {
		parameters[name] = specs
	}
	for name, symbol := range g.symbols {
		if symbol.arity >= 0 {
			arities[name] = symbol.arity
		}
		parameters[name] = symbol.operatorParams
	}
	for name, symbol := range g.formals {
		arities[name] = symbol.arity
		arities[localOperatorArityKey(name)] = symbol.arity
		parameters[name] = nil
	}
	return arities, parameters
}

// TAKE binds in the current proof context. PICK temporarily binds for its
// predicate, then installs only accepted nodes after its own proof finishes.
func (g *sanyExpressionGeneration) generateProofBinder(step *ProofStep, context map[string]Position) (map[string]localSymbol, Diagnostics) {
	step.binderNode, step.pickContext, step.formalNodes = nil, nil, nil
	var diagnostics Diagnostics
	var closeContext func()
	if step.Kind == "PICK" {
		closeContext = g.pushFormalContext(len(step.Bounds))
	}
	var previousDomain Expr
	for _, bound := range step.Bounds {
		if bound.Set != nil && bound.Set != previousDomain {
			diagnostics = append(diagnostics, g.proofExpression(bound.Set, context, nil)...)
			previousDomain = bound.Set
		}
	}
	introduced := map[string]localSymbol{}
	step.formalNodes = make([]*sanyFormalParamNode, 0, len(step.Bounds))
	for _, bound := range step.Bounds {
		node := g.newFormalParameter(bound.Name, 0, bound.Pos, step.Syntax)
		step.formalNodes = append(step.formalNodes, node)
		generated := g.bindFormalParameter(node, context, nil)
		diagnostics = append(diagnostics, generated...)
		if len(generated) == 0 {
			introduced[bound.Name] = g.formals[bound.Name]
			context[bound.Name] = bound.Pos
		}
	}
	operands := make([]sanySemanticGraphNode, 0)
	complete := true
	if step.Kind == "PICK" {
		step.pickContext = g.formalSymbolTable().topContext()
		popLabelFormals := g.pushLabelFormals(step.formalNodes)
		diagnostics = append(diagnostics, g.proofExpression(step.Expr, context, nil)...)
		popLabelFormals()
		body := sanyGeneratedExpressionNode(step.Expr)
		complete = body != nil || sanyExpressionGenerationFailure(step.Expr) == sanyGenerationNullExpression
		operands = append(operands, body)
		closeContext()
	}
	var syntax *SanySyntaxNode
	if step.Syntax != nil {
		heirs := step.Syntax.GetHeirs()
		if len(heirs) > 1 {
			syntax = heirs[1]
		}
	}
	operator := "$Take"
	if step.Kind == "PICK" {
		operator = "$Pick"
	}
	if len(step.Bounds) > 0 && step.Bounds[0].Set != nil {
		groups := make([][]*sanyFormalParamNode, 0)
		tuples := make([]bool, 0)
		ranges := make([]sanySemanticGraphNode, 0)
		for i := 0; i < len(step.Bounds); {
			domain := step.Bounds[i].Set
			end := i + 1
			for end < len(step.Bounds) && step.Bounds[end].Set == domain {
				end++
			}
			bound := sanyGeneratedExpressionNode(domain)
			if bound == nil && sanyExpressionGenerationFailure(domain) != sanyGenerationNullExpression {
				complete = false
			}
			groups = append(groups, step.formalNodes[i:end])
			tuples = append(tuples, step.Bounds[i].TupleBound)
			ranges = append(ranges, bound)
			i = end
		}
		if complete {
			step.binderNode = newSanySemBoundedOpApplNode(operator, nil, operands, groups, tuples, ranges, syntax)
		}
	} else if complete {
		step.binderNode = newSanySemUnboundedOpApplNode(operator, operands, step.formalNodes, syntax)
	}
	return introduced, diagnostics
}
