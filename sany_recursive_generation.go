// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// processRecursive builds an unregistered declaration, then unnamed dummy
// formals and the registered, undefined OpDef. Rejected nodes still enter all
// three source module vectors; declaration bookkeeping must not discard them.
func (g *sanyExpressionGeneration) constructRecursiveDeclaration(declaration Declaration, index int, context map[string]Position) (*sanySemOpDefNode, Diagnostics) {
	if declaration.Syntax == nil || g.currentModule == nil || g.currentModule.semanticNode == nil {
		return nil, nil
	}
	name := declaration.Names[index]
	if _, exists := g.lookupSymbol(name, context); exists && g.formalSymbolTable().resolveSymbol(name) == nil {
		// Earlier imported/native symbols lacking their actual graph are not
		// replaced by a fabricated registration target.
		return nil, nil
	}
	syntax := declaration.Syntax.GetHeirs()[2*index+1]
	arity, _ := declarationArity(declaration, name)
	module := g.currentModule.semanticNode
	transient := newSanySemOpDeclNode(sanyCanonicalOperatorImage(name), sanyConstantDeclKind, constantLevel, arity, module, syntax)
	parameters := make([]*sanyFormalParamNode, transient.semArity())
	for i := range parameters {
		parameters[i] = newSanyFormalParamNode("", 0, Position{}, nil, g.currentModule)
		parameters[i].nullName = true
	}
	node, diagnostics := newSanySemOpDefNode(transient.semName(), sanyUserDefinedOpKind, parameters, false, nil, module, g.formalSymbolTable(), syntax, false, nil)
	node.letInLevel = g.level
	node.inRecursive, node.inRecursiveSection = true, true
	node.recursiveSection = g.module.section
	module.recursiveDecls = append(module.recursiveDecls, node)
	module.opDefsInRecursiveSection = append(module.opDefsInRecursiveSection, node)
	module.recursiveOpDefNodes = append(module.recursiveOpDefNodes, node)
	return node, diagnostics
}

func (g *sanyExpressionGeneration) setDefinitionRecursionFields(node *sanySemOpDefNode) {
	node.letInLevel = g.level
	if g.module != nil && g.module.sum > 0 {
		node.recursiveSection = g.module.section
		count := g.module.count
		if g.level != 0 {
			if g.level < 0 || g.level >= len(g.module.counts) {
				panic(tlc.NewArrayIndexOutOfBoundsException(g.level, len(g.module.counts)))
			}
			count = g.module.counts[g.level]
		}
		node.inRecursiveSection = count > 0
		if g.currentModule != nil && g.currentModule.semanticNode != nil {
			module := g.currentModule.semanticNode
			module.opDefsInRecursiveSection = append(module.opDefsInRecursiveSection, node)
		}
	}
}

// endOpDefNode changes the declaration before decrementing its unresolved
// counters. A failure retains those changes and the caller's active scopes.
func (g *sanyExpressionGeneration) endRecursiveDefinition(node *sanySemOpDefNode, body sanySemanticGraphNode, syntax *SanySyntaxNode) {
	node.defined, node.body, node.TreeNode = true, body, syntax
	if syntax != nil {
		node.pos = sanyNodePosition(syntax)
		bridge := tlcBridge{convertingModule: node.originalModuleName}
		node.Location = bridge.sourceLocationForPosition(node.pos)
	}
	if binding := g.bindings[node.semName()]; binding != nil && binding.node == node {
		binding.defined = true
		binding.position = node.semPosition()
	}
	if node.inRecursive {
		if g.level < 0 || g.level >= len(g.module.counts) {
			panic(tlc.NewArrayIndexOutOfBoundsException(g.level, len(g.module.counts)))
		}
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
}
