// Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

type sanyUseOrHideBuilder struct {
	facts    []sanySemanticGraphNode
	defs     []sanySemSymbol
	complete bool
}

func newSanyUseOrHideBuilder() *sanyUseOrHideBuilder {
	return &sanyUseOrHideBuilder{facts: make([]sanySemanticGraphNode, 0), defs: make([]sanySemSymbol, 0), complete: true}
}

// Consume each reference immediately after generation so later bindings cannot
// change the identity that occupies its source vector slot.
func (builder *sanyUseOrHideBuilder) appendReference(g *sanyExpressionGeneration, reference ProofRef, appended bool) {
	if !appended {
		return
	}
	if reference.Module != "" {
		module := g.formalSymbolTable().resolveModule(reference.Module)
		if module == nil && g.currentModule != nil && reference.Module == g.currentModule.Name {
			module = g.currentModule.semanticNode
		}
		if module == nil {
			builder.complete = false
			return
		}
		if reference.Defs {
			builder.defs = append(builder.defs, module)
		} else {
			builder.facts = append(builder.facts, module)
		}
		return
	}
	if reference.Defs {
		symbol := g.formalSymbolTable().resolveSymbol(reference.Name)
		switch definition := symbol.(type) {
		case *sanySemOpDefNode:
			if definition.semKind() != sanyUserDefinedOpKind && definition.semKind() != sanyModuleInstanceKind && definition.semKind() != sanyNumberedProofStepKind {
				builder.complete = false
				return
			}
		case *sanySemThmOrAssumpDefNode:
			if len(definition.semName()) > 0 && definition.semName()[0] == '<' {
				builder.complete = false
				return
			}
		default:
			builder.complete = false
			return
		}
		// Qualified selectors still need their actual selected symbol identity.
		if source := sanyExprSource(reference.Expr); source != nil && source.Selector != nil && len(source.Selector.Steps) > 1 {
			builder.complete = false
			return
		}
		builder.defs = append(builder.defs, symbol)
		return
	}
	graph := sanyGeneratedExpressionNode(reference.Expr)
	if graph == nil && sanyExpressionGenerationFailure(reference.Expr) != sanyGenerationNullExpression {
		builder.complete = false
		return
	}
	builder.facts = append(builder.facts, graph)
}
func (builder *sanyUseOrHideBuilder) finish(syntax *SanySyntaxNode) *sanySemUseOrHideNode {
	if !builder.complete || syntax == nil {
		return nil
	}
	kind, only := sanyUseKind, false
	heirs := syntax.GetHeirs()
	if len(heirs) > 0 && heirs[0].Image == "HIDE" {
		kind = sanyHideKind
	}
	next := 1
	if len(heirs) > 0 && heirs[0].Image == "PROOF" {
		next++
	}
	if next < len(heirs) && heirs[next].Image == "ONLY" {
		only = true
	}
	return newSanySemUseOrHideNode(kind, syntax, builder.facts, builder.defs, only)
}
