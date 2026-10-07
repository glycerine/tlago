// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlc

import "reflect"

// SemanticChildren follows SANY getChildren, rather than walkGraph. Operator
// definitions referenced by an application are not its syntax children.
func SemanticChildren(node SemanticNode) []SemanticNode {
	switch n := node.(type) {
	case *ModuleNode:
		result := make([]SemanticNode, 0, len(n.GetOpDefs()))
		for _, op := range n.GetOpDefs() {
			result = append(result, op)
		}
		return append(result, n.TopLevel...)
	case *OpDefNode:
		return []SemanticNode{n.Body}
	case *OpApplNode:
		return append(append([]SemanticNode(nil), n.BdedQuantBounds...), n.Args...)
	case *LetInNode:
		result := make([]SemanticNode, 0, len(n.Lets)+1)
		for _, op := range n.Lets {
			result = append(result, op)
		}
		return append(result, n.Body)
	case *SubstInNode:
		result := []SemanticNode{n.Body}
		for _, subst := range n.Substs {
			result = append(result, subst.Expr)
		}
		return result
	case *APSubstInNode:
		result := []SemanticNode{n.Body}
		for _, subst := range n.Substs {
			result = append(result, subst.Expr)
		}
		return result
	case *LabelNode:
		return []SemanticNode{n.Body}
	case *AssumeNode:
		return []SemanticNode{n.Assume}
	default:
		return nil
	}
}

// SemanticListOfChildren follows SemanticNode.getListOfChildren: null entries
// are removed without changing the raw getChildren result.
func SemanticListOfChildren(node SemanticNode) []SemanticNode {
	children := SemanticChildren(node)
	result := make([]SemanticNode, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}
		value := reflect.ValueOf(child)
		if value.Kind() == reflect.Pointer && value.IsNil() {
			continue
		}
		result = append(result, child)
	}
	return result
}

// SemanticPathTo ports SemanticNode.pathTo: the first syntax path is returned
// innermost first, and formal declarations are considered explicitly.
func SemanticPathTo(root SemanticNode, location SourceLocation, requireExact bool) []SemanticNode {
	var path []SemanticNode
	var visit func(SemanticNode)
	visit = func(node SemanticNode) {
		loc, ok := semanticNodeSourceLocation(node)
		children := SemanticListOfChildren(node)
		if ok && (loc == location || !requireExact && len(children) == 0 && loc.Includes(location)) {
			path = make([]SemanticNode, 0)
		} else {
			var parameters []*SymbolNode
			switch n := node.(type) {
			case *OpDefNode:
				parameters = n.Params
			case *OpApplNode:
				parameters = n.GetQuantSymbolLists()
			}
			for _, parameter := range parameters {
				if parameter.Location == location {
					path = []SemanticNode{parameter}
				}
			}
		}
		for _, child := range children {
			childLocation, hasLocation := semanticNodeSourceLocation(child)
			if path == nil && hasLocation && childLocation.Includes(location) {
				visit(child)
			}
		}
		if path != nil {
			path = append(path, node)
		}
	}
	visit(root)
	if path == nil {
		return make([]SemanticNode, 0)
	}
	return path
}

func (n *OpApplNode) GetQuantSymbolLists() []*SymbolNode {
	result := append([]*SymbolNode(nil), n.UnbdedQuantSymbols...)
	for _, symbols := range n.BdedQuantSymbolLists {
		result = append(result, symbols...)
	}
	return result
}

func (m *ModuleNode) PathTo(location SourceLocation, requireExact ...bool) []SemanticNode {
	exact := len(requireExact) == 0 || requireExact[0]
	return SemanticPathTo(m, location, exact)
}
