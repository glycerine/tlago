// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// Ordinary OpDefNode construction retains the supplied body and parameter array.
// Generator must generate the body and pop its formal context before calling
// this constructor. Recursive completion reuses its previously declared node.
func newSanySemOpDefNode(name string, kind sanySemKind, parameters []*sanyFormalParamNode, local bool, body sanySemanticGraphNode, module *sanySemModuleNode, table *sanySymbolTable, syntax *SanySyntaxNode, defined bool, source *sanySemOpDefNode) (*sanySemOpDefNode, Diagnostics) {
	if parameters == nil {
		parameters = make([]*sanyFormalParamNode, 0)
	}
	node := &sanySemOpDefNode{
		sanySemSymbolBase: sanySemSymbolBase{sanySemanticNode: newSanySemanticNode(kind), name: name, arity: len(parameters), local: local},
		formalNodes:       parameters, body: body, module: module, table: table, defined: defined,
		argMaxLevels: make([]tlaLevel, len(parameters)), argWeights: make([]int, len(parameters)),
		leibniz: make([]bool, len(parameters)), isLeibniz: true,
		letInLevel: -1, recursiveSection: -1,
	}
	if source != nil {
		node.source = source
	}
	if module != nil {
		node.originalModuleName = module.semName()
	}
	if syntax == nil {
		node.TreeNode = nil
		node.Location = tlc.NullSourceLocation
	} else {
		node.TreeNode = syntax
		node.pos = sanyNodePosition(syntax)
		bridge := tlcBridge{convertingModule: node.originalModuleName}
		node.Location = bridge.sourceLocationForPosition(node.pos)
	}
	for i := range parameters {
		node.argMaxLevels[i] = temporalLevel
		node.leibniz[i] = true
	}
	if table != nil {
		return node, table.addSymbol(node)
	}
	return node, nil
}
