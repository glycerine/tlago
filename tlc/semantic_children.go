package tlc

import "reflect"

// SemanticChildren follows SANY getChildren, rather than walkGraph. Operator
// definitions referenced by an application are not its syntax children.
func SemanticChildren(node SemanticNode) []SemanticNode {
	if semanticExploreNull(node) {
		panic(NewNullPointerException())
	}
	switch n := node.(type) {
	case *ModuleNode:
		return n.GetChildren()
	case *OpDefNode:
		return n.GetChildren()
	case *OpApplNode:
		return n.GetChildren()
	case *LetInNode:
		return n.GetChildren()
	case *SubstInNode:
		return n.GetChildren()
	case *APSubstInNode:
		return n.GetChildren()
	case *LabelNode:
		return n.GetChildren()
	case *AssumeNode:
		return n.GetChildren()
	case *TheoremNode:
		return n.GetChildren()
	case *ThmOrAssumpDefNode:
		return n.GetChildren()
	case *AssumeProveNode:
		return n.GetChildren()
	case *NewSymbNode:
		return n.GetChildren()
	case *LeafProofNode:
		return n.GetChildren()
	case *NonLeafProofNode:
		return n.GetChildren()
	case *DefStepNode:
		return n.GetChildren()
	case *UseOrHideNode:
		return n.GetChildren()
	case *InstanceNode:
		return n.GetChildren()
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

// SemanticHasChildren ports SemanticNode.hasChildren over the filtered child
// list, so null entries do not make a node non-leaf.
func SemanticHasChildren(node SemanticNode) bool {
	if semanticExploreNull(node) {
		panic(NewNullPointerException())
	}
	return len(SemanticListOfChildren(node)) != 0
}

// SemanticChildrenVisitor adapts SemanticNode.ChildrenVisitor. Nil callbacks
// retain the source defaults: no-op preVisit, preempt every child, return the
// same visitor from postVisit, and a zero result from get.
type SemanticChildrenVisitor[T any] struct {
	PreVisit  func(SemanticNode)
	Preempt   func(SemanticNode) bool
	PostVisit func(SemanticNode) *SemanticChildrenVisitor[T]
	Get       func() T
}

func (visitor *SemanticChildrenVisitor[T]) Result() T {
	if visitor == nil {
		panic(NewNullPointerException())
	}
	if visitor.Get != nil {
		return visitor.Get()
	}
	var zero T
	return zero
}

// WalkSemanticChildren ports SemanticNode.walkChildren. The filtered child
// slice is captured after PreVisit, each child's returned visitor is ignored,
// and the root returns its own PostVisit result.
func WalkSemanticChildren[T any](node SemanticNode, visitor *SemanticChildrenVisitor[T]) *SemanticChildrenVisitor[T] {
	if semanticExploreNull(node) || visitor == nil {
		panic(NewNullPointerException())
	}
	if visitor.PreVisit != nil {
		visitor.PreVisit(node)
	}
	for _, child := range SemanticListOfChildren(node) {
		preempt := true
		if visitor.Preempt != nil {
			preempt = visitor.Preempt(child)
		}
		if !preempt {
			WalkSemanticChildren(child, visitor)
		}
	}
	if visitor.PostVisit != nil {
		return visitor.PostVisit(node)
	}
	return visitor
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
				parameters = n.GetParams()
			case *OpApplNode:
				parameters = n.GetQuantSymbolLists()
			}
			for _, parameter := range parameters {
				if parameter.GetSourceLocation() == location {
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
	n.require()
	result := append(make([]*SymbolNode, 0), n.UnbdedQuantSymbols...)
	for _, symbols := range n.BdedQuantSymbolLists {
		if symbols == nil {
			panic(NewNullPointerException())
		}
		result = append(result, symbols...)
	}
	return result
}

func (n *OpApplNode) GetChildren() []SemanticNode {
	n.require()
	if n.BdedQuantBounds == nil || n.Args == nil {
		panic(NewNullPointerException())
	}
	children := make([]SemanticNode, 0, len(n.BdedQuantBounds)+len(n.Args))
	children = append(children, n.BdedQuantBounds...)
	children = append(children, n.Args...)
	return children
}

func (m *ModuleNode) PathTo(location SourceLocation, requireExact ...bool) []SemanticNode {
	m.require()
	if len(requireExact) > 1 {
		panic(NewIllegalArgumentException("PathTo accepts at most one exact-match flag"))
	}
	exact := len(requireExact) == 0 || requireExact[0]
	return SemanticPathTo(m, location, exact)
}
