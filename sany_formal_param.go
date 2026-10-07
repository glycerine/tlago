// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// FormalParamNode owns its SemanticNode identity. The enclosing native module
// and retained parser node preserve declaration ownership for graph generation.
// Its LevelNode data and graph visitors are ported separately.
type sanyFormalParamNode struct {
	sanySemSymbolBase
	module *Module
}

func newSanyFormalParamNode(name string, arity int, position Position, syntax *SanySyntaxNode, module *Module) *sanyFormalParamNode {
	moduleName := ""
	if module != nil {
		moduleName = module.Name
	}
	node := &sanyFormalParamNode{
		sanySemSymbolBase: sanySemSymbolBase{
			sanySemanticNode: newSanySemanticNode(sanyFormalParamKind),
			name:             name, arity: arity, local: true,
			originalModuleName: moduleName, pos: position,
		},
		module: module,
	}
	if syntax == nil {
		node.TreeNode = tlc.NullSemanticNodeInstance.GetTreeNode()
		node.Location = tlc.NullSemanticNodeInstance.Location
	} else {
		node.TreeNode = syntax
		bridge := tlcBridge{convertingModule: moduleName}
		node.Location = bridge.sourceLocationForPosition(position)
	}
	return node
}

// SemanticNode.equals checks the concrete class, kind and UID, rather than the
// formal's name or declaration location.
func (n *sanyFormalParamNode) equals(other any) bool {
	o, ok := other.(*sanyFormalParamNode)
	if !ok || o == nil {
		return false
	}
	return n == o || n.getKind() == o.getKind() && n.getUID() == o.getUID()
}

func (g *sanyExpressionGeneration) newFormalParameter(name string, arity int, position Position, syntax *SanySyntaxNode) *sanyFormalParamNode {
	return newSanyFormalParamNode(name, arity, position, sanySyntaxAtPosition(syntax, position), g.currentModule)
}
