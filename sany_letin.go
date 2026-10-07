// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// LetInNode retains the supplied arrays, body and definition context. getLets
// lazily filters user-defined operators once, returning that cached array.
type sanySemLetInNode struct {
	sanySemanticNode
	opDefs     []sanySemSymbol
	instances  []sanySemanticGraphNode
	body       sanySemanticGraphNode
	context    *sanyContext
	gottenLets []*sanySemOpDefNode
}

func newSanySemLetInNode(syntax *SanySyntaxNode, definitions []sanySemSymbol, instances []sanySemanticGraphNode, body sanySemanticGraphNode, context *sanyContext) *sanySemLetInNode {
	node := &sanySemLetInNode{sanySemanticNode: newSanySemanticNode(sanyLetInKind), opDefs: definitions, instances: instances, body: body, context: context}
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

func (n *sanySemLetInNode) getLets() []*sanySemOpDefNode {
	if n.gottenLets == nil {
		if n.opDefs == nil {
			panic(tlc.NewNullPointerException())
		}
		count := 0
		for _, definition := range n.opDefs {
			if definition == nil {
				panic(tlc.NewNullPointerException())
			}
			if definition.semKind() == sanyUserDefinedOpKind {
				count++
			}
		}
		n.gottenLets = make([]*sanySemOpDefNode, count)
		index := 0
		for _, definition := range n.opDefs {
			if definition.semKind() == sanyUserDefinedOpKind {
				n.gottenLets[index] = definition.(*sanySemOpDefNode)
				index++
			}
		}
	}
	return n.gottenLets
}

// Ident LHS parameters own their entire declaration syntax, including the
// placeholders of operator-valued parameters. Fixity LHS parameters are tokens.
func sanyDefinitionFormalSyntax(definition *Definition, index int) *SanySyntaxNode {
	if definition.Syntax == nil {
		return nil
	}
	children := definition.Syntax.One
	if len(children) == 0 {
		return nil
	}
	lhs := children[0]
	parts := lhs.GetHeirs()
	position := -1
	switch lhs.Kind.JavaName() {
	case "N_IdentLHS":
		position = 2 + 2*index
	case "N_PrefixLHS":
		position = 1
	case "N_InfixLHS":
		position = 2 * index
	case "N_PostfixLHS":
		position = 0
	}
	if position >= 0 && position < len(parts) {
		return parts[position]
	}
	return nil
}
