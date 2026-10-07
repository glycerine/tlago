// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

type sanySemLeafProofNode struct {
	sanySemanticNode
	facts   []sanySemanticGraphNode
	defs    []sanySemSymbol
	omitted bool
	isOnly  bool
}

func newSanySemLeafProofNode(syntax *SanySyntaxNode, facts []sanySemanticGraphNode, definitions []sanySemSymbol, omitted, only bool) *sanySemLeafProofNode {
	node := &sanySemLeafProofNode{sanySemanticNode: newSanySemanticNode(sanyLeafProofKind), facts: facts, defs: definitions, omitted: omitted, isOnly: only}
	sanyAssertionSyntax(&node.sanySemanticNode, syntax)
	return node
}
func (node *sanySemLeafProofNode) getFacts() []sanySemanticGraphNode { return node.facts }
func (node *sanySemLeafProofNode) getDefs() []sanySemSymbol          { return node.defs }
func (node *sanySemLeafProofNode) getOmitted() bool                  { return node.omitted }
func (node *sanySemLeafProofNode) getOnlyFlag() bool                 { return node.isOnly }
func (node *sanySemLeafProofNode) getChildren() []sanySemanticGraphNode {
	if len(node.facts) == 0 {
		return nil
	}
	children := make([]sanySemanticGraphNode, len(node.facts))
	copy(children, node.facts)
	return children
}
