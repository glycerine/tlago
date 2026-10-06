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
func (g *sanyExpressionGeneration) proofReference(reference ProofRef, module *Module, context map[string]Position, locals map[string]bool) Diagnostics {
	if reference.Module != "" {
		target := g.spec.Modules[reference.Module]
		if reference.Module == module.Name || (target != nil && !(sameSourceFile(target.Pos, reference.Pos) && positionBefore(reference.Pos, target.Pos))) {
			return nil
		}
		diagnostic := errorAt(reference.Pos, "E4005", "module %s is unavailable in proof", reference.Module)
		diagnostic.SANYMessage = fmt.Sprintf("Module `%s' used without being extended or instantiated.", reference.Module)
		diagnostic.SANYRange = SanyRange{Begin: reference.Pos, End: reference.Pos.SourceEnd()}
		return Diagnostics{diagnostic}
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
		name := reference.Name
		if len(diags) != 0 {
			// genIdToSelector has diagnosed the malformed selector.
		} else if selected := sanyExprSelection(reference.Expr); selected != nil {
			// DEF entries must end at a named operator, not an expression operand.
			if len(selected.lets) > 0 && reference.Expr != nil {
				if source := sanyExprSource(reference.Expr); source != nil && source.Selector != nil {
					last := source.Selector.Steps[len(source.Selector.Steps)-1]
					if last.Kind == SanySelectorName {
						return nil
					}
				}
			}
		} else if source := sanyExprSource(reference.Expr); source != nil && source.selectorFailure {
			diags = append(diags, g.checkExpr(reference.Expr, context, locals)...)
		} else if symbol, exists := g.lookupSymbol(name, context); exists {
			if symbol.proofStepKind == "DEFINE" {
				return nil
			}
			if (symbol.kind == OperatorDecl || symbol.kind == InstanceDecl || symbol.kind == semanticTheoremImportKind) && (len(name) == 0 || name[0] != '<') {
				return nil
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
		return append(diags, diagnostic)
	}
	if reference.Expr == nil {
		return nil
	}
	if identifier, ok := reference.Expr.(*IdentExpr); ok {
		if symbol, exists := g.lookupSymbol(identifier.Name, context); exists && symbol.proofStepKind == "DEFINE" {
			diagnostic := errorAt(reference.Pos, "E4004", "Step number of non-fact used as a fact")
			diagnostic.SANYMessage = diagnostic.Message
			diagnostic.SANYRange = SanyRange{Begin: reference.Pos, End: reference.Pos.SourceEnd()}
			return Diagnostics{diagnostic}
		}
	}
	previousFact := g.fact
	g.fact = sanyProofReferenceDirect(reference)
	diags := g.checkExpr(reference.Expr, context, locals)
	g.fact = previousFact
	arities, parameters := g.proofSignatures()
	diags = append(diags, checkCallArity(reference.Expr, arities, parameters, locals)...)
	return append(diags, checkOperatorArgumentKinds(reference.Expr, parameters, arities, locals)...)
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
	diags = append(diags, g.leafProofReferences(proof.LeafRefs, module, base)...)
	baseSymbols := g.symbols
	symbolScopes := map[int]map[string]localSymbol{}

	scopes := map[int]map[string]Position{}
	pendingSuffices := map[int]*AssumeProve{}
	pendingPicks := map[int][]BoundVar{}
	for _, step := range proof.Steps {
		for depth, bounds := range pendingPicks {
			if step.Depth <= depth {
				if scopes[depth] == nil {
					scopes[depth] = map[string]Position{}
				}
				for _, bound := range bounds {
					scopes[depth][bound.Name] = bound.Pos
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
		for _, definition := range step.Definitions {
			previous, exists := g.lookupSymbol(definition.Name, current)
			binding := g.bindings[definition.Name]
			failedFunction := definition.FunctionDef && ((exists && previous.kind != OperatorDecl && previous.kind != InstanceDecl) || (binding != nil && !binding.defined && binding.level != g.level))
			diags = append(diags, g.generateLocalDefinition(definition, current, map[string]bool{})...)
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
		}
		for _, instance := range step.Instances {
			for _, symbol := range semanticInstanceSymbols(instance, g.spec) {
				binding := localSymbol{kind: symbol.importKind(), arity: symbol.arity, pos: symbol.sourcePosition()}
				g.symbols[symbol.name] = binding
				symbolScopes[step.Depth][symbol.name] = binding
				current[symbol.name] = symbol.sourcePosition()
				scopes[step.Depth][symbol.name] = symbol.sourcePosition()
			}
			if instance.Name != "" {
				current[instance.Name] = instance.SourcePosition()
				scopes[step.Depth][instance.Name] = instance.SourcePosition()
				current[instanceNameSentinel(instance.Name)] = instance.SourcePosition()
				scopes[step.Depth][instanceNameSentinel(instance.Name)] = instance.SourcePosition()
			}
		}
		statementContext := copySanyExpressionContext(current)
		if step.AssumeProveBody != nil {
			diags = append(diags, checkAssumeProveBindings(step.AssumeProveBody, statementContext, nil, g)...)
		} else if step.Kind == "PICK" || step.Kind == "TAKE" {
			var previousDomain Expr
			for _, bound := range step.Bounds {
				if bound.Set != nil && bound.Set != previousDomain {
					diags = append(diags, g.proofExpression(bound.Set, current, nil)...)
					previousDomain = bound.Set
				}
			}
			for _, bound := range step.Bounds {
				if previous, exists := statementContext[bound.Name]; exists {
					diagnostic := errorAt(bound.Pos, "E4201", "bound symbol %s conflicts with existing symbol", bound.Name)
					diagnostic.SANYMessage = fmt.Sprintf("Multiply-defined symbol '%s': this definition or declaration conflicts \nwith the one at %s.", bound.Name, sanySymbolLocation(previous))
					diagnostic.SANYRange = SanyRange{Begin: bound.Pos, End: bound.Pos.SourceEnd()}
					diags = append(diags, diagnostic)
				} else {
					statementContext[bound.Name] = bound.Pos
				}
			}
			diags = append(diags, g.proofExpression(step.Expr, statementContext, nil)...)
		} else {
			diags = append(diags, g.proofExpression(step.Expr, statementContext, nil)...)
		}
		for _, expression := range step.Exprs {
			diags = append(diags, g.proofExpression(expression, current, nil)...)
		}
		for _, reference := range step.UseHideRefs {
			diags = append(diags, g.proofReference(reference, module, current, nil)...)
		}
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
		if step.AssumeProveBody != nil {
			diags = append(diags, checkAssumeProveLabels(step.AssumeProveBody, true)...)
		}
		if step.QualifiedName != "" {
			kind := step.Kind
			if step.Suffices {
				kind = "SUFFICES"
			}
			symbol := localSymbol{proofStepKind: kind, proofAssumeProve: step.AssumeProveBody != nil, kind: semanticTheoremImportKind, pos: step.Pos}
			g.symbols[step.QualifiedName] = symbol
			symbolScopes[step.Depth][step.QualifiedName] = symbol
			current[step.QualifiedName] = step.Pos
			scopes[step.Depth][step.QualifiedName] = step.Pos
		}
		if step.Kind == "TAKE" {
			for _, bound := range step.Bounds {
				current[bound.Name] = bound.Pos
				scopes[step.Depth][bound.Name] = bound.Pos
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
		diags = append(diags, g.leafProofReferences(step.LeafRefs, module, ownProof)...)
		g.symbols = ownSymbols
		if step.Kind == "PICK" {
			pendingPicks[step.Depth] = step.Bounds
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
	return diags
}

// Failed DEF and MODULE entries are omitted from the generated vectors. Failed
// expression facts still occupy a slot, as in generateUseOrHide's vec.addElement.
func (g *sanyExpressionGeneration) leafProofReferences(references []ProofRef, module *Module, context map[string]Position) Diagnostics {
	var diags Diagnostics
	entries := 0
	for _, reference := range references {
		generated := g.proofReference(reference, module, context, nil)
		diags = append(diags, generated...)
		if (!reference.Defs && reference.Module == "") || len(generated) == 0 {
			entries++
		}
	}
	if len(references) != 0 && entries == 0 {
		syntax := references[0].Syntax
		position := sanyNodePosition(syntax)
		diagnostic := errorAt(position, "E4003", "Empty BY")
		diagnostic.SANYMessage = diagnostic.Message
		diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
		diags = append(diags, diagnostic)
	}
	return diags
}

func sanyProofNewBindings(symbols map[string]localSymbol, body *AssumeProve) {
	if body == nil {
		return
	}
	for _, clause := range body.Assumptions {
		if symbol := clause.NewSymbol; symbol != nil {
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
