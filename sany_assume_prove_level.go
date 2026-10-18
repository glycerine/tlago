// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

func (n *sanySemAssumeProveNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	n.levelCorrect = true
	count := func() int {
		if n.assumes == nil {
			panic(tlc.NewNullPointerException())
		}
		return len(n.assumes)
	}
	assume := func(i int) sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(n.assumes[i]) }
	prove := func() sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(n.prove) }
	for i := 0; i < count(); i++ {
		if sanyGraphNodePresent(n.assumes[i]) && !assume(i).levelCheck(iter, errors) {
			n.levelCorrect = false
		}
	}
	// The source intentionally ignores the prove expression's boolean result.
	prove().levelCheck(iter, errors)
	*n.level = prove().getLevel()
	for i := 0; i < count(); i++ {
		assume(i).levelCheck(iter, errors)
		if assume(i).getLevel() > *n.level {
			*n.level = assume(i).getLevel()
		}
	}
	n.levelParams.addAll(prove().getLevelParams())
	n.allParams.addAll(prove().getAllParams())
	for i := 0; i < count(); i++ {
		n.levelParams.addAll(assume(i).getLevelParams())
		n.allParams.addAll(assume(i).getAllParams())
	}
	n.levelConstraints.putAll(sanyLevelConstraintMap(prove().getLevelConstraints()))
	for i := 0; i < count(); i++ {
		n.levelConstraints.putAll(sanyLevelConstraintMap(assume(i).getLevelConstraints()))
	}
	n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(prove().getArgLevelConstraints()))
	for i := 0; i < count(); i++ {
		n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(assume(i).getArgLevelConstraints()))
	}
	n.argLevelParams.addAll(prove().getArgLevelParams())
	for i := 0; i < count(); i++ {
		n.argLevelParams.addAll(assume(i).getArgLevelParams())
	}
	// Source leaves nonLeibnizParams untouched and uses only assumption correctness.
	if n.levelCorrect {
		sanyAddTemporalLevelConstraintToConstants(n.levelParams, n.levelConstraints)
	}
	return n.levelCorrect
}
