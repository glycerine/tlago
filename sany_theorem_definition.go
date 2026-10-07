// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

type sanySemThmOrAssumpDefNode struct {
	sanySemSymbolBase
	body             sanySemanticGraphNode
	thmOrAssump      sanySemanticGraphNode
	module           *sanySemModuleNode
	instantiatedFrom *sanySemModuleNode
	theorem          bool
	suffices         bool
	formalNodes      []*sanyFormalParamNode
	proof            sanySemanticGraphNode
	argMaxLevels     []tlaLevel
	argWeights       []int
	minMaxLevel      [][]tlaLevel
	opLevelCond      [][][]bool
	leibniz          []bool
	isLeibniz        bool
	labels           *sanyLabelTable
}

// The provisional constructor creates the goal identity before its body.
func newSanySemThmOrAssumpDefNode(name string, syntax *SanySyntaxNode) *sanySemThmOrAssumpDefNode {
	node := &sanySemThmOrAssumpDefNode{sanySemSymbolBase: sanySemSymbolBase{sanySemanticNode: newSanySemanticNode(sanyThmOrAssumpDefKind), name: name}, theorem: true, formalNodes: make([]*sanyFormalParamNode, 0)}
	if syntax == nil {
		node.TreeNode = nil
		node.Location = tlc.NullSourceLocation
	} else {
		node.TreeNode = syntax
		node.pos = sanyNodePosition(syntax)
		bridge := tlcBridge{}
		node.Location = bridge.sourceLocationForPosition(node.pos)
	}
	return node
}

// The full constructor initializes source/instantiation before registration,
// but installs parameters afterward. Registration therefore observes arity zero.
func newSanySemCompletedThmOrAssumpDefNode(name string, theorem bool, body sanySemanticGraphNode, module *sanySemModuleNode, table *sanySymbolTable, syntax *SanySyntaxNode, parameters []*sanyFormalParamNode, from *sanySemModuleNode, source *sanySemThmOrAssumpDefNode) (*sanySemThmOrAssumpDefNode, Diagnostics) {
	node := newSanySemThmOrAssumpDefNode(name, syntax)
	node.instantiatedFrom = from
	if source != nil {
		node.source = source
	}
	diagnostics := node.construct(theorem, body, module, table, parameters)
	return node, diagnostics
}
func (node *sanySemThmOrAssumpDefNode) construct(theorem bool, body sanySemanticGraphNode, module *sanySemModuleNode, table *sanySymbolTable, parameters []*sanyFormalParamNode) Diagnostics {
	node.theorem, node.body, node.module = theorem, body, module
	node.originalModuleName = ""
	if module != nil {
		node.originalModuleName = module.semName()
	}
	var diagnostics Diagnostics
	if table != nil {
		diagnostics = table.addSymbol(node)
	}
	if parameters != nil {
		node.formalNodes = parameters
		node.arity = len(parameters)
	}
	return diagnostics
}
func (node *sanySemThmOrAssumpDefNode) getSource() *sanySemThmOrAssumpDefNode {
	if node.source == nil {
		return node
	}
	return node.source.(*sanySemThmOrAssumpDefNode)
}
func (node *sanySemThmOrAssumpDefNode) getBody() sanySemanticGraphNode { return node.body }
func (node *sanySemThmOrAssumpDefNode) getOriginallyDefinedInModuleNode() *sanySemModuleNode {
	return node.module
}
func (node *sanySemThmOrAssumpDefNode) getInstantiatedFrom() *sanySemModuleNode {
	return node.instantiatedFrom
}
func (node *sanySemThmOrAssumpDefNode) isTheorem() bool                   { return node.theorem }
func (node *sanySemThmOrAssumpDefNode) isSuffices() bool                  { return node.suffices }
func (node *sanySemThmOrAssumpDefNode) setSuffices()                      { node.suffices = true }
func (node *sanySemThmOrAssumpDefNode) getProof() sanySemanticGraphNode   { return node.proof }
func (node *sanySemThmOrAssumpDefNode) getParams() []*sanyFormalParamNode { return node.formalNodes }
func (node *sanySemThmOrAssumpDefNode) setLocal(local bool)               { node.local = local }
func (node *sanySemThmOrAssumpDefNode) getChildren() []sanySemanticGraphNode {
	return []sanySemanticGraphNode{node.body}
}
func (node *sanySemThmOrAssumpDefNode) setLabels(labels *sanyLabelTable) { node.labels = labels }
func (node *sanySemThmOrAssumpDefNode) getLabelsHT() *sanyLabelTable     { return node.labels }
func (node *sanySemThmOrAssumpDefNode) getLabel(name string) *sanySemLabelNode {
	return node.labels.get(name)
}
func (node *sanySemThmOrAssumpDefNode) getLabels() []*sanySemLabelNode { return node.labels.elements() }
func (node *sanySemThmOrAssumpDefNode) addLabel(label *sanySemLabelNode) bool {
	if node.labels == nil {
		node.labels = newSanyLabelTable()
	}
	return node.labels.add(label)
}
func (node *sanySemThmOrAssumpDefNode) match(application *sanySemOpApplNode) bool {
	if application == nil || application.operator == nil {
		panic(tlc.NewNullPointerException(""))
	}
	return application.operator.semArity() == 0
}

func (node *sanySemThmOrAssumpDefNode) getArity() int { return node.arity }
func (node *sanySemThmOrAssumpDefNode) isExpr() bool {
	switch node.body.(type) {
	case *sanySemOpApplNode, *sanySemLetInNode, *sanySemLabelNode, *sanySemAtNode, *tlc.NumeralNode, *tlc.DecimalNode, *tlc.StringNode, *tlc.SubstInNode, *tlc.OpApplNode, *tlc.LetInNode, *tlc.LabelNode, *tlc.AtNode:
		return true
	}
	return false
}
