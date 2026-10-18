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
	base.SetSyntaxLocationResolver(sanyTreeLocation)
	data := newSanyLevelData(&base)
	base.CanonicalLevelData = data
	return sanySemanticNode{SemanticNodeBase: &base, sanyLevelData: data}
}

func (n *sanySemanticNode) getUID() int32            { return n.GetUid() }
func (n *sanySemanticNode) getKind() sanySemKind     { return sanySemKind(n.Kind()) }
func (n *sanySemanticNode) setKind(kind sanySemKind) { n.KindValue = tlc.SemanticKind(kind) }
func (n *sanySemanticNode) hashCode() int32          { return n.JavaHashCode() }

func (n *sanySemanticNode) getToolObject(toolID int) any {
	return n.GetToolObjectAt(int32(toolID))
}

func (n *sanySemanticNode) setToolObject(toolID int, object any) {
	n.SetToolObjectAt(int32(toolID), object)
}

// SanySemanticNodeWalkGraph adapts
// SemanticNode.walkGraph(Hashtable, ExplorerVisitor).
func SanySemanticNodeWalkGraph(node tlc.SemanticNode, nodes map[int32]any, visitor *SanyExplorerVisitor) {
	tlc.WalkSemanticGraph(node, nodes, visitor)
}

// SanySemanticNodePathTo adapts both SemanticNode.pathTo overloads. The
// no-flag form requires an exact source-location match.
func SanySemanticNodePathTo(node tlc.SemanticNode, location tlc.SourceLocation, requireExact ...bool) []tlc.SemanticNode {
	if len(requireExact) > 1 {
		panic(tlc.NewIllegalArgumentException("SemanticNode.pathTo accepts at most one exact-match flag"))
	}
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	exact := len(requireExact) == 0 || requireExact[0]
	return tlc.SemanticPathTo(node, location, exact)
}

// SanySemanticNodeIsDefinedWith adapts SemanticNode.isDefinedWith.
func SanySemanticNodeIsDefinedWith(node, target tlc.SemanticNode) bool {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	return tlc.SemanticIsDefinedWith(node, target)
}

// SanySemanticNodeSubstituteFor adapts SemanticNode.substituteFor.
func SanySemanticNodeSubstituteFor(node tlc.SemanticNode, replacement, original *tlc.OpDefNode) {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	tlc.SemanticSubstituteFor([]tlc.SemanticNode{node}, replacement, original)
}
