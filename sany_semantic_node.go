package tlago

import "github.com/glycerine/tlago/tlc"

// SemanticNode's constructor identity is shared with the evaluator's semantic
// nodes. Remaining graph constructors and allocation order are ported separately.
type sanySemanticNode struct {
	tlc.SemanticNodeBase
	tools []any
}

func newSanySemanticNode(kind sanySemKind) sanySemanticNode {
	return sanySemanticNode{SemanticNodeBase: tlc.NewSemanticNodeBase(tlc.SemanticKind(kind), "")}
}

func (n *sanySemanticNode) getUID() int32            { return n.GetUID() }
func (n *sanySemanticNode) getKind() sanySemKind     { return sanySemKind(n.Kind()) }
func (n *sanySemanticNode) setKind(kind sanySemKind) { n.KindValue = tlc.SemanticKind(kind) }
func (n *sanySemanticNode) hashCode() int32          { return n.JavaHashCode() }

func (n *sanySemanticNode) getToolObject(toolID int) any {
	if len(n.tools) <= toolID {
		return nil
	}
	if toolID < 0 {
		panic(tlc.NewArrayIndexOutOfBoundsException(toolID, len(n.tools)))
	}
	return n.tools[toolID]
}

func (n *sanySemanticNode) setToolObject(toolID int, object any) {
	if len(n.tools) <= toolID {
		tools := make([]any, toolID+1)
		copy(tools, n.tools)
		n.tools = tools
	}
	if toolID < 0 {
		panic(tlc.NewArrayIndexOutOfBoundsException(toolID, len(n.tools)))
	}
	n.tools[toolID] = object
}
