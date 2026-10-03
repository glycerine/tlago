// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

func (b *tlcBridge) formalParameter(name string, arity int, position Position, tree *SanySyntaxNode) *tlc.SymbolNode {
	parameter := tlc.NewFormalParamSymbolNode(name, arity)
	parameter.Location = b.sourceLocationForPosition(position)
	parameter.TreeNode = sanySyntaxAtPosition(tree, position)
	return parameter
}

// Use the actual retained parser node; do not rebuild a declaration's syntax
// from its name. The central AST positions originate from these exact trees.
func sanySyntaxAtPosition(tree *SanySyntaxNode, position Position) *SanySyntaxNode {
	if tree == nil || position.Line == 0 {
		return nil
	}
	r := tree.Range
	if r.Begin.Line == position.Line && r.Begin.Column == position.Column && r.End.Line == position.EndLine && r.End.Column == position.EndColumn {
		return tree
	}
	for _, child := range tree.Heirs {
		if node := sanySyntaxAtPosition(child, position); node != nil {
			return node
		}
	}
	return nil
}

func (b *tlcBridge) pushFormalParameters(parameters []*tlc.SymbolNode) func() {
	names := make([]string, len(parameters))
	previous := make([]*tlc.SymbolNode, len(parameters))
	for i, parameter := range parameters {
		name := parameter.Name.String()
		names[i] = name
		previous[i] = b.symbols[name]
		b.symbols[name] = parameter
	}
	restoreNames := b.pushConvertBoundNames(names...)
	return func() {
		restoreNames()
		for i := len(parameters) - 1; i >= 0; i-- {
			if previous[i] == nil {
				delete(b.symbols, names[i])
			} else {
				b.symbols[names[i]] = previous[i]
			}
		}
	}
}

func (b *tlcBridge) boundParameters(bounds []BoundVar, tree *SanySyntaxNode) []*tlc.SymbolNode {
	parameters := make([]*tlc.SymbolNode, len(bounds))
	for i, bound := range bounds {
		parameters[i] = b.formalParameter(bound.Name, bound.OperatorArity, bound.Pos, tree)
	}
	return parameters
}
