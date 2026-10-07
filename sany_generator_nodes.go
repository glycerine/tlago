// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"sync"
)

// Semantic graph links retain node identity independently of the native AST.
// Remaining ordinary constructors and LevelNode data are ported separately.
type sanySemanticGraphNode interface {
	getKind() sanySemKind
	getUID() int32
}

type sanySemOpDefNode struct {
	sanySemSymbolBase
	formalNodes []*sanyFormalParamNode
	body        sanySemanticGraphNode
}

type sanySemOpApplNode struct {
	sanySemanticNode
	operator              sanySemSymbol
	operands              []sanySemanticGraphNode
	ranges                []sanySemanticGraphNode
	unboundedBoundSymbols []*sanyFormalParamNode
	boundedBoundSymbols   [][]*sanyFormalParamNode
	tupleOrs              []bool
}

type sanySemOpArgNode struct {
	sanySemanticNode
	name     string
	arity    int
	operator sanySemSymbol
	module   *Module
}

type sanySemLabelNode struct {
	sanySemanticNode
	name        string
	arity       int
	formalNodes []*sanyFormalParamNode
	body        sanySemanticGraphNode
}

// Generator's class-wide ordinary ASSUME marker is an OpDeclNode, not
// nullSN or a formal. Its declaration is checked at constant level once.
type sanySemOpDeclNode struct {
	sanySemSymbolBase
	level        tlaLevel
	levelChecked int
}

var sanyInAssumeDummyNode = sync.OnceValue(func() *sanySemOpDeclNode {
	return &sanySemOpDeclNode{
		sanySemSymbolBase: sanySemSymbolBase{sanySemanticNode: newSanySemanticNode(0), name: "$$InAssume", arity: 0},
		level:             constantLevel,
		levelChecked:      1,
	}
})

type sanyGeneratorNodes struct {
	inAssumeDummy *sanySemOpDeclNode

	nullODN       *sanySemOpDefNode
	nullOAN       *sanySemOpApplNode
	nullOpArg     *sanySemOpArgNode
	nullLabelNode *sanySemLabelNode
}

func sanyNullSyntaxNode(kind sanySemKind) sanySemanticNode {
	node := newSanySemanticNode(kind)
	node.TreeNode = tlc.NullSemanticNodeInstance.GetTreeNode()
	node.Location = tlc.NullSemanticNodeInstance.Location
	return node
}

// Generator constructs these four nodes in order. nullODN has kind zero,
// whereas nullOAN has OpApplKind so processing can continue after errors.
func newSanyGeneratorNodes() *sanyGeneratorNodes {
	n := &sanyGeneratorNodes{inAssumeDummy: sanyInAssumeDummyNode()}
	n.nullODN = &sanySemOpDefNode{sanySemSymbolBase: sanySemSymbolBase{sanySemanticNode: sanyNullSyntaxNode(0), name: "nullODN", arity: -2, pos: Position{File: "--TLA+ BUILTINS--"}}}
	n.nullOAN = &sanySemOpApplNode{sanySemanticNode: sanyNullSyntaxNode(sanyOpApplKind), operator: n.nullODN, operands: make([]sanySemanticGraphNode, 0), ranges: make([]sanySemanticGraphNode, 0)}
	n.nullOpArg = &sanySemOpArgNode{sanySemanticNode: sanyNullSyntaxNode(sanyOpArgKind), name: "nullOpArg", arity: -2}
	n.nullLabelNode = &sanySemLabelNode{sanySemanticNode: sanyNullSyntaxNode(sanyLabelKind), name: "nullLabelNode", formalNodes: make([]*sanyFormalParamNode, 0), body: n.nullOAN}
	return n
}
