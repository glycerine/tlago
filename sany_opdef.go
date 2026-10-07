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

// Numbered non-theorem steps are OpDefNodes, not operator definitions. Keep
// their ordinary body/parameter/level arrays null and register the step backlink
// only after construction, as in OpDefNode's numbered-proof-step overload.
func newSanySemNumberedProofStepNode(name string, step sanySemanticGraphNode, module *sanySemModuleNode, table *sanySymbolTable, syntax *SanySyntaxNode) (*sanySemOpDefNode, Diagnostics) {
	node := &sanySemOpDefNode{
		sanySemSymbolBase: sanySemSymbolBase{sanySemanticNode: newSanySemanticNode(sanyNumberedProofStepKind), name: name, arity: 0},
		stepNode:          step, module: module, table: table,
		letInLevel: -1, recursiveSection: -1,
	}
	if module != nil {
		node.originalModuleName = module.semName()
	}
	sanyAssertionSyntax(&node.sanySemanticNode, syntax)
	if syntax != nil {
		node.pos = sanyNodePosition(syntax)
	}
	if table == nil {
		panic(tlc.NewNullPointerException(""))
	}
	return node, table.addSymbol(node)
}

func (node *sanySemOpDefNode) getStepNode() sanySemanticGraphNode { return node.stepNode }

// The body remains the only child even for a numbered step. walkGraph also
// visits stepNode; getChildren deliberately does not, matching Java.
func (node *sanySemOpDefNode) getChildren() []sanySemanticGraphNode {
	return []sanySemanticGraphNode{node.body}
}
