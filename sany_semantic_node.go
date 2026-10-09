package tlago

import "github.com/glycerine/tlago/tlc"

// SemanticNode's constructor identity is shared with the evaluator's semantic
// nodes. Remaining graph constructors and allocation order are ported separately.
type sanySemanticNode struct {
	*tlc.SemanticNodeBase
	*sanyLevelData
}

func (n *sanySemanticNode) runtimeSemanticBase() *tlc.SemanticNodeBase {
	return n.SemanticNodeBase
}

func newSanySemanticNode(kind sanySemKind) sanySemanticNode {
	base := tlc.NewSemanticNodeBase(tlc.SemanticKind(kind), "")
	data := newSanyLevelData(&base)
	base.CanonicalLevelData = data
	return sanySemanticNode{SemanticNodeBase: &base, sanyLevelData: data}
}

func (n *sanySemanticNode) getUID() int32            { return n.GetUID() }
func (n *sanySemanticNode) getKind() sanySemKind     { return sanySemKind(n.Kind()) }
func (n *sanySemanticNode) setKind(kind sanySemKind) { n.KindValue = tlc.SemanticKind(kind) }
func (n *sanySemanticNode) hashCode() int32          { return n.JavaHashCode() }

func (n *sanySemanticNode) getToolObject(toolID int) any {
	return n.GetToolObjectAt(int32(toolID))
}

func (n *sanySemanticNode) setToolObject(toolID int, object any) {
	n.SetToolObjectAt(int32(toolID), object)
}
