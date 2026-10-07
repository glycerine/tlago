package tlago

import (
	"sync/atomic"

	"github.com/glycerine/tlago/tlc"
)

// SemanticNode's constructor, UID, kind, hash and tool-object storage. All
// semantic node constructors must share this counter; front-end initialization
// does not reset it. The remaining graph node constructors are ported separately.
var sanySemanticNodeUID atomic.Int32

type sanySemanticNode struct {
	myUID int32
	kind  sanySemKind
	tools []any
}

func newSanySemanticNode(kind sanySemKind) sanySemanticNode {
	// AtomicInteger.getAndIncrement returns the previous signed 32-bit value,
	// including Java int wraparound.
	return sanySemanticNode{myUID: sanySemanticNodeUID.Add(1) - 1, kind: kind}
}

func (n *sanySemanticNode) getUID() int32            { return n.myUID }
func (n *sanySemanticNode) getKind() sanySemKind     { return n.kind }
func (n *sanySemanticNode) setKind(kind sanySemKind) { n.kind = kind }

func (n *sanySemanticNode) hashCode() int32 {
	result := int32(1)
	result = 31*result + int32(n.kind)
	result = 31*result + n.myUID
	return result
}

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
