// Portions Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

func (n *sanySemUseOrHideNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	// DEF references are already checked; only facts participate here.
	return n.levelCheckGraphSubnodes(iter, n.facts, errors)
}
func (n *sanySemLeafProofNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	return n.levelCheckGraphSubnodes(iter, n.facts, errors)
}
func (n *sanySemDefStepNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	defs := n.defs
	return n.levelCheckSubnodesAccess(iter, defs == nil, len(defs), func(i int) sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(defs[i]) }, errors)
}
func (n *sanySemNonLeafProofNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	if n.steps == nil || n.insts == nil {
		panic(tlc.NewNullPointerException())
	}
	sub := make([]sanySemanticGraphNode, len(n.steps)+len(n.insts))
	copy(sub, n.steps)
	copy(sub[len(n.steps):], n.insts)
	return n.levelCheckGraphSubnodes(iter, sub, errors)
}
