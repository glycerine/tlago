// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// Non-Leibniz parameters retain the inherited getter and the label's own set;
// Java delegates only the six getters implemented below.
func (n *sanySemLabelNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	// Source returns true on a cached check, even after a false body result.
	if *n.levelChecked >= iter {
		return true
	}
	*n.levelChecked = iter
	if n.formalNodes == nil {
		panic(tlc.NewNullPointerException())
	}
	for i := 0; i < len(n.formalNodes); i++ {
		if n.formalNodes[i] != nil {
			n.formalNodes[i].levelCheck(iter, errors)
		}
	}
	return sanyRequireCanonicalLevelNode(n.body).levelCheck(iter, errors)
}

func (n *sanySemLabelNode) getLevel() tlaLevel {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getLevel called for TheoremNode before levelCheck")
	return sanyRequireCanonicalLevelNode(n.body).getLevel()
}

func (n *sanySemLabelNode) getLevelParams() *sanyLevelSet[sanySemSymbol] {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getLevelParams called for ThmNode before levelCheck")
	return sanyRequireCanonicalLevelNode(n.body).getLevelParams()
}

func (n *sanySemLabelNode) getAllParams() *sanyLevelSet[sanySemSymbol] {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getAllParams called for ThmNode before levelCheck")
	return sanyRequireCanonicalLevelNode(n.body).getAllParams()
}

func (n *sanySemLabelNode) getLevelConstraints() *sanySetOfLevelConstraints {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getLevelConstraints called for ThmNode before levelCheck")
	return sanyRequireCanonicalLevelNode(n.body).getLevelConstraints()
}

func (n *sanySemLabelNode) getArgLevelConstraints() *sanySetOfArgLevelConstraints {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getArgLevelConstraints called for ThmNode before levelCheck")
	return sanyRequireCanonicalLevelNode(n.body).getArgLevelConstraints()
}

func (n *sanySemLabelNode) getArgLevelParams() *sanyLevelSet[*sanyArgLevelParam] {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getArgLevelParams called for ThmNode before levelCheck")
	return sanyRequireCanonicalLevelNode(n.body).getArgLevelParams()
}
