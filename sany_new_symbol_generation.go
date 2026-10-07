// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

type sanySemNewSymbNode struct {
	sanySemanticNode
	opDeclNode *sanySemOpDeclNode
	set        sanySemanticGraphNode
}

func newSanySemNewSymbNode(declaration *sanySemOpDeclNode, set sanySemanticGraphNode, syntax *SanySyntaxNode) *sanySemNewSymbNode {
	node := &sanySemNewSymbNode{sanySemanticNode: newSanySemanticNode(sanyNewSymbKind), opDeclNode: declaration, set: set}
	if syntax == nil {
		node.TreeNode = nil
		node.Location = tlc.NullSourceLocation
	} else {
		node.TreeNode = syntax
		bridge := tlcBridge{}
		node.Location = bridge.sourceLocationForPosition(sanyNodePosition(syntax))
	}
	return node
}
func (node *sanySemNewSymbNode) getSet() sanySemanticGraphNode     { return node.set }
func (node *sanySemNewSymbNode) getOpDeclNode() *sanySemOpDeclNode { return node.opDeclNode }

func (node *sanySemNewSymbNode) getChildren() []sanySemanticGraphNode {
	if node.set == nil {
		return nil
	}
	return []sanySemanticGraphNode{node.set}
}

// Generator.generateNewSymb generates the set before buildParameter declares
// the symbol, and allocates NewSymbNode after that declaration's registration.
func (g *sanyExpressionGeneration) generateNewSymbol(symbol *NewSymbol, context map[string]Position, locals map[string]bool) Diagnostics {
	symbol.semanticNode = nil
	symbol.bindingSymbol = nil
	var diagnostics Diagnostics
	if symbol.Domain != nil {
		diagnostics = append(diagnostics, g.proofExpression(symbol.Domain, context, locals)...)
	}
	var module *sanySemModuleNode
	if g.currentModule != nil {
		module = g.currentModule.semanticNode
	}
	name := ResolveSanyOperatorSynonym(symbol.Name)
	declaration := newSanySemOpDeclNode(name, sanySemKind(symbol.Kind), symbol.Level, symbol.Arity, module, symbol.declarationSyntax)
	table := g.formalSymbolTable()
	declaration.table = table
	// Missing earlier imported/native identities cannot be replaced by a newly
	// accepted declaration. Retain the existing native binding in that case.
	previous, exists := g.lookupSymbol(name, context)
	if table.resolveSymbol(name) == nil && exists {
		diagnostics = append(diagnostics, checkBindingName("NEW symbol", name, symbol.Pos, context, locals)...)
	} else {
		_, current := table.registerSymbol(declaration)
		diagnostics = append(diagnostics, current...)
		if resolved, ok := table.resolveSymbol(name).(*sanySemOpDeclNode); ok {
			kind := ConstantDecl
			if resolved.semKind() == sanyNewVariableKind || resolved.semKind() == sanyVariableDeclKind {
				kind = VariableDecl
			}
			previous = localSymbol{kind: kind, arity: resolved.semArity(), pos: resolved.semPosition(), declarationNode: resolved}
			exists = true
			g.symbols[name] = previous
		}
	}
	if exists {
		binding := previous
		symbol.bindingSymbol = &binding
	}
	set := sanyGeneratedExpressionNode(symbol.Domain)
	node := newSanySemNewSymbNode(declaration, set, symbol.Syntax)
	if symbol.Domain == nil || set != nil || sanyExpressionGenerationFailure(symbol.Domain) == sanyGenerationNullExpression {
		symbol.semanticNode = node
	}
	return diagnostics
}
