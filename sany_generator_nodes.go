// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"sync"
)

// Semantic graph links retain actual SANY and TLC nodes through their shared
// semantic identity methods; literal bodies need no adapter or duplicate node.
type sanySemanticGraphNode interface {
	Kind() tlc.SemanticKind
	GetUID() int32
}

type sanySemOpDefNode struct {
	sanySemSymbolBase
	labels             *sanyLabelTable
	formalNodes        []*sanyFormalParamNode
	body               sanySemanticGraphNode
	module             *sanySemModuleNode
	defined            bool
	level              tlaLevel
	levelChecked       int32
	argMaxLevels       []tlaLevel
	argWeights         []int
	leibniz            []bool
	isLeibniz          bool
	table              *sanySymbolTable
	letInLevel         int
	inRecursive        bool
	inRecursiveSection bool
	recursiveSection   int
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
	module   *sanySemModuleNode
}

type sanySemLabelNode struct {
	sanySemanticNode
	labels          *sanyLabelTable
	goal            sanySemanticGraphNode
	goalClause      int
	isAssumeProve   bool
	subExpressionOf sanySemSymbol
	name            string
	arity           int
	formalNodes     []*sanyFormalParamNode
	body            sanySemanticGraphNode
}

// Generator's class-wide ordinary ASSUME marker is an OpDeclNode, not
// nullSN or a formal. Its declaration is checked at constant level once.
type sanySemOpDeclNode struct {
	sanySemSymbolBase
	level        tlaLevel
	levelChecked int
	table        *sanySymbolTable
	module       *sanySemModuleNode
	levelParams  map[*sanySemOpDeclNode]struct{}
	allParams    map[*sanySemOpDeclNode]struct{}
}

// OpDeclNode initializes its level data before registration. Registration is
// performed by the generator's current scope; rejected nodes keep their identity.
func newSanySemOpDeclNode(name string, kind sanySemKind, level tlaLevel, arity int, module *sanySemModuleNode, syntax any) *sanySemOpDeclNode {
	n := &sanySemOpDeclNode{
		sanySemSymbolBase: sanySemSymbolBase{sanySemanticNode: newSanySemanticNode(kind), name: name, arity: arity},
		level:             level, levelChecked: 1, module: module,
		levelParams: make(map[*sanySemOpDeclNode]struct{}), allParams: make(map[*sanySemOpDeclNode]struct{}),
	}
	if module != nil {
		n.originalModuleName = module.semName()
	}
	switch tree := syntax.(type) {
	case *SanySyntaxNode:
		if tree == nil {
			n.TreeNode = nil
			n.Location = tlc.NullSourceLocation
		} else {
			n.TreeNode = tree
			n.pos = sanyNodePosition(tree)
			bridge := tlcBridge{convertingModule: n.originalModuleName}
			n.Location = bridge.sourceLocationForPosition(n.pos)
		}
	case nil:
		n.TreeNode = nil
		n.Location = tlc.NullSourceLocation
	default:
		// nullSTN is shared with the evaluator bridge; it is distinct from nil.
		if syntax != tlc.NullSemanticNodeInstance.GetTreeNode() {
			panic("unsupported declaration syntax node")
		}
		n.TreeNode = syntax
		n.Location = tlc.NullSemanticNodeInstance.Location
		n.pos = Position{File: n.Location.Source}
	}
	if kind == sanyConstantDeclKind {
		n.levelParams[n] = struct{}{}
		n.allParams[n] = struct{}{}
	}
	return n
}

var sanyInAssumeDummyNode = sync.OnceValue(func() *sanySemOpDeclNode {
	return newSanySemOpDeclNode("$$InAssume", 0, constantLevel, 0, nil, nil)
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

// OpDefNode(UniqueString) is also used directly by the original context test.
func newSanySemNullOpDefNode(name string) *sanySemOpDefNode {
	return &sanySemOpDefNode{letInLevel: -1, recursiveSection: -1, sanySemSymbolBase: sanySemSymbolBase{sanySemanticNode: sanyNullSyntaxNode(0), name: name, arity: -2, pos: Position{File: "--TLA+ BUILTINS--"}}}
}

// Generator constructs these four nodes in order. nullODN has kind zero,
// whereas nullOAN has OpApplKind so processing can continue after errors.
func newSanyGeneratorNodes() *sanyGeneratorNodes {
	n := &sanyGeneratorNodes{inAssumeDummy: sanyInAssumeDummyNode()}
	n.nullODN = newSanySemNullOpDefNode("nullODN")
	n.nullOAN = &sanySemOpApplNode{sanySemanticNode: sanyNullSyntaxNode(sanyOpApplKind), operator: n.nullODN, operands: make([]sanySemanticGraphNode, 0), ranges: make([]sanySemanticGraphNode, 0)}
	n.nullOpArg = &sanySemOpArgNode{sanySemanticNode: sanyNullSyntaxNode(sanyOpArgKind), name: "nullOpArg", arity: -2}
	n.nullLabelNode = &sanySemLabelNode{sanySemanticNode: sanyNullSyntaxNode(sanyLabelKind), name: "nullLabelNode", formalNodes: make([]*sanyFormalParamNode, 0), body: n.nullOAN}
	return n
}
