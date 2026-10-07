// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

type sanySemAssumeProveNode struct {
	sanySemanticNode
	assumes          []sanySemanticGraphNode
	prove            sanySemanticGraphNode
	goal             sanySemanticGraphNode
	inScopeOfDecl    []bool
	inProof          bool
	suffices         bool
	isBoxAssumeProve bool
}

func newSanySemAssumeProveNode(syntax *SanySyntaxNode, goal sanySemanticGraphNode) *sanySemAssumeProveNode {
	node := &sanySemAssumeProveNode{sanySemanticNode: newSanySemanticNode(sanyAssumeProveKind), goal: goal, inProof: true}
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
func (node *sanySemAssumeProveNode) getAssumes() []sanySemanticGraphNode { return node.assumes }
func (node *sanySemAssumeProveNode) getProve() sanySemanticGraphNode     { return node.prove }
func (node *sanySemAssumeProveNode) getGoal() sanySemanticGraphNode      { return node.goal }
func (node *sanySemAssumeProveNode) getSuffices() bool                   { return node.suffices }
func (node *sanySemAssumeProveNode) setSuffices()                        { node.suffices = true }
func (node *sanySemAssumeProveNode) getIsBoxAssumeProve() bool           { return node.isBoxAssumeProve }
func (node *sanySemAssumeProveNode) setIsBoxAssumeProve(value bool)      { node.isBoxAssumeProve = value }
func (node *sanySemAssumeProveNode) getChildren() []sanySemanticGraphNode {
	if node.assumes == nil {
		panic(tlc.NewNullPointerException(""))
	}
	children := make([]sanySemanticGraphNode, len(node.assumes)+1)
	copy(children, node.assumes)
	children[len(node.assumes)] = node.prove
	return children
}

func (g *sanyExpressionGeneration) checkAssumeProveBody(body *AssumeProve, context map[string]Position, locals map[string]bool, goalUnavailable bool) Diagnostics {
	previous := g.apGoalUnavailable
	g.apGoalUnavailable = goalUnavailable
	defer func() { g.apGoalUnavailable = previous }()
	return checkAssumeProveBindings(body, context, locals, g)
}
