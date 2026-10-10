// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// The view allocates no semantic identity and never replaces the body node.
// Its collections belong to the actual literal's base; level and iteration
// cells alias that same base, including direct TLC LevelCheck/SetLevel calls.
type sanyLiteralLevelView struct {
	*sanyLevelData
	literal sanySemanticGraphNode
}

func sanyCanonicalLiteral(literal sanySemanticGraphNode, base *tlc.SemanticNodeBase) *sanyLiteralLevelView {
	base.SetSyntaxLocationResolver(sanyTreeLocation)
	if base.CanonicalLevelData == nil {
		base.CanonicalLevelData = newSanyLevelData(base)
	}
	data, ok := base.CanonicalLevelData.(*sanyLevelData)
	if !ok {
		panic(tlc.NewClassCastException("literal canonical level metadata"))
	}
	return &sanyLiteralLevelView{sanyLevelData: data, literal: literal}
}
func (n *sanyLiteralLevelView) getKind() sanySemKind { return sanySemKind(n.literal.Kind()) }
func (n *sanyLiteralLevelView) levelCheck(iter int32, errors *Diagnostics) bool {
	switch literal := n.literal.(type) {
	case *tlc.NumeralNode:
		return literal.LevelCheck(iter)
	case *tlc.DecimalNode:
		return literal.LevelCheck(iter)
	case *tlc.StringNode:
		return literal.LevelCheck(iter)
	}
	panic(tlc.NewClassCastException("not a literal node"))
}
