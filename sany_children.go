// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"path/filepath"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

func (node *sanySemModuleNode) getChildren() []sanySemanticGraphNode {
	if node.children != nil {
		return node.children
	}
	definitions := node.getOpDefs()
	count := sanyGraphArrayLength(node.topLevel)
	node.children = make([]sanySemanticGraphNode, len(definitions)+count)
	for i, definition := range definitions {
		node.children[i] = definition
	}
	copy(node.children[len(definitions):], node.topLevel)
	return node.children
}
func (node *sanySemOpApplNode) getChildren() []sanySemanticGraphNode {
	children := make([]sanySemanticGraphNode, sanyGraphArrayLength(node.ranges)+sanyGraphArrayLength(node.operands))
	copy(children, node.ranges)
	copy(children[len(node.ranges):], node.operands)
	return children
}
func (node *sanySemOpApplNode) getQuantSymbolLists() []*sanyFormalParamNode {
	symbols := make([]*sanyFormalParamNode, 0)
	symbols = append(symbols, node.unboundedBoundSymbols...)
	if node.boundedBoundSymbols != nil {
		for _, group := range node.boundedBoundSymbols {
			if group == nil {
				panic(tlc.NewNullPointerException())
			}
			symbols = append(symbols, group...)
		}
	}
	return symbols
}

// getChildren is independent of walkGraph: it follows source containment,
// includes null entries, and defaults to null on source leaf classes.
func sanyChildren(node sanySemanticGraphNode) []sanySemanticGraphNode {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	if owner, ok := node.(interface {
		getChildren() []sanySemanticGraphNode
	}); ok {
		return owner.getChildren()
	}
	return nil
}
func sanyListOfChildren(node sanySemanticGraphNode) []sanySemanticGraphNode {
	children := sanyChildren(node)
	result := make([]sanySemanticGraphNode, 0, len(children))
	for _, child := range children {
		if !sanyExploreNull(child) {
			result = append(result, child)
		}
	}
	return result
}
func sanyHasChildren(node sanySemanticGraphNode) bool { return len(sanyListOfChildren(node)) > 0 }

type sanyChildrenVisitor[T any] struct {
	preVisit  func(sanySemanticGraphNode)
	preempt   func(sanySemanticGraphNode) bool
	postVisit func(sanySemanticGraphNode) *sanyChildrenVisitor[T]
	get       func() T
}

// walkChildren snapshots the filtered children after preVisit. A child's
// returned visitor is ignored; the root returns its own postVisit result.
func sanyWalkChildren[T any](node sanySemanticGraphNode, visitor *sanyChildrenVisitor[T]) *sanyChildrenVisitor[T] {
	if sanyExploreNull(node) || visitor == nil {
		panic(tlc.NewNullPointerException())
	}
	if visitor.preVisit != nil {
		visitor.preVisit(node)
	}
	for _, child := range sanyListOfChildren(node) {
		if visitor.preempt == nil || visitor.preempt(child) {
			continue
		}
		sanyWalkChildren(child, visitor)
	}
	if visitor.postVisit != nil {
		return visitor.postVisit(node)
	}
	return visitor
}
func (visitor *sanyChildrenVisitor[T]) result() T {
	if visitor == nil {
		panic(tlc.NewNullPointerException())
	}
	if visitor.get != nil {
		return visitor.get()
	}
	var zero T
	return zero
}
func sanyGraphLocation(node sanySemanticGraphNode) tlc.SourceLocation {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	owner, ok := node.(interface{ GetTreeNode() any })
	if !ok {
		panic(tlc.NewClassCastException())
	}
	tree := owner.GetTreeNode()
	if sanyExploreNull(tree) {
		return tlc.NullSourceLocation
	}
	if syntax, ok := tree.(*SanySyntaxNode); ok {
		return sanySyntaxLocation(syntax)
	}
	if tree == tlc.NullSemanticNodeInstance.GetTreeNode() {
		return tlc.NullSemanticNodeInstance.Location
	}
	panic(tlc.NewClassCastException())
}

func sanySyntaxLocation(syntax *SanySyntaxNode) tlc.SourceLocation {
	if syntax == nil {
		panic(tlc.NewNullPointerException())
	}
	// SyntaxTreeNode.getLocation reads all four current coordinates directly.
	// Preserve zero end coordinates instead of filling them from the beginning.
	source := syntax.FileName
	if source != "" {
		source = strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
	}
	return tlc.NewSourceLocation(source, syntax.Range.Begin.Line, syntax.Range.Begin.Column, syntax.Range.End.Line, syntax.Range.End.Column)
}

// Action.getDeclaration reads the first current child of syntax.one().
func (syntax *SanySyntaxNode) GetDeclarationLocation() tlc.SourceLocation {
	if syntax == nil || len(syntax.One) == 0 {
		return tlc.NullSourceLocation
	}
	return sanySyntaxLocation(syntax.One[0])
}

// Port of SemanticNode.pathTo. Its visitor searches syntax children and formal
// declarations, returning the first matching path with the innermost node first.
func sanyPathTo(node sanySemanticGraphNode, location tlc.SourceLocation, exact ...bool) []sanySemanticGraphNode {
	requireExact := len(exact) == 0 || exact[0]
	var path []sanySemanticGraphNode
	visitor := &sanyChildrenVisitor[[]sanySemanticGraphNode]{}
	visitor.get = func() []sanySemanticGraphNode {
		if path == nil {
			return make([]sanySemanticGraphNode, 0)
		}
		return path
	}
	visitor.preVisit = func(current sanySemanticGraphNode) {
		position := sanyGraphLocation(current)
		if location == position || (!requireExact && !sanyHasChildren(current) && position.Includes(location)) {
			path = make([]sanySemanticGraphNode, 0)
		} else {
			var parameters []*sanyFormalParamNode
			switch n := current.(type) {
			case *sanySemOpDefNode:
				if n.formalNodes == nil {
					panic(tlc.NewNullPointerException())
				}
				parameters = n.formalNodes
			case *sanySemOpApplNode:
				parameters = n.getQuantSymbolLists()
			}
			for _, parameter := range parameters {
				if location == sanyGraphLocation(parameter) {
					path = []sanySemanticGraphNode{parameter}
				}
			}
		}
	}
	visitor.preempt = func(current sanySemanticGraphNode) bool {
		return path != nil || !sanyGraphLocation(current).Includes(location)
	}
	visitor.postVisit = func(current sanySemanticGraphNode) *sanyChildrenVisitor[[]sanySemanticGraphNode] {
		if path != nil {
			path = append(path, current)
		}
		return visitor
	}
	return sanyWalkChildren(node, visitor).result()
}
