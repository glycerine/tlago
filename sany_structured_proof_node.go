// Portions Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// NonLeafProofNode keeps the proof's steps and module-definition instances in
// separate arrays. Only the steps are children; instances are level-check inputs.
type sanySemNonLeafProofNode struct {
	sanySemanticNode
	steps   []sanySemanticGraphNode
	insts   []sanySemanticGraphNode
	context *sanyContext
}

func newSanySemNonLeafProofNode(syntax *SanySyntaxNode, steps, instances []sanySemanticGraphNode, context *sanyContext) *sanySemNonLeafProofNode {
	node := &sanySemNonLeafProofNode{sanySemanticNode: newSanySemanticNode(sanyNonLeafProofKind), steps: steps, insts: instances, context: context}
	sanyAssertionSyntax(&node.sanySemanticNode, syntax)
	return node
}
func (node *sanySemNonLeafProofNode) getSteps() []sanySemanticGraphNode { return node.steps }
func (node *sanySemNonLeafProofNode) getContext() *sanyContext          { return node.context }
func (node *sanySemNonLeafProofNode) getChildren() []sanySemanticGraphNode {
	if len(node.steps) == 0 {
		return nil
	}
	children := make([]sanySemanticGraphNode, len(node.steps))
	copy(children, node.steps)
	return children
}

// DefStepNode retains the actual definitions and nullable interned step number.
type sanySemDefStepNode struct {
	sanySemanticNode
	stepNumber *tlc.UniqueString
	defs       []*sanySemOpDefNode
}

func newSanySemDefStepNode(syntax *SanySyntaxNode, stepNumber *tlc.UniqueString, definitions []*sanySemOpDefNode) *sanySemDefStepNode {
	node := &sanySemDefStepNode{sanySemanticNode: newSanySemanticNode(sanyDefStepKind), stepNumber: stepNumber, defs: definitions}
	sanyAssertionSyntax(&node.sanySemanticNode, syntax)
	return node
}
func (node *sanySemDefStepNode) getStepNumber() *tlc.UniqueString { return node.stepNumber }
func (node *sanySemDefStepNode) getDefs() []*sanySemOpDefNode     { return node.defs }
func (node *sanySemDefStepNode) getChildren() []sanySemanticGraphNode {
	if node.defs == nil {
		panic(tlc.NewNullPointerException(""))
	}
	children := make([]sanySemanticGraphNode, len(node.defs))
	for i, definition := range node.defs {
		if definition != nil {
			children[i] = definition
		}
	}
	return children
}
