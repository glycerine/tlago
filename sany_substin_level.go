// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

func (n *sanySemSubstInNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.sanySemSubstitutionNode.levelCheckSubstitution(iter, errors, false)
}
func (n *sanySemAPSubstInNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.sanySemSubstitutionNode.levelCheckSubstitution(iter, errors, true)
}
func (n *sanySemSubstitutionNode) levelCheckSubstitution(iter int32, errors *Diagnostics, ap bool) bool {
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	n.levelCorrect = true
	body := func() sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(n.body) }
	count := func() int {
		if n.substs == nil {
			panic(tlc.NewNullPointerException())
		}
		return len(n.substs)
	}
	replacement := func(i int) sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(n.getSubWith(i)) }
	if !body().levelCheck(iter, errors) {
		n.levelCorrect = false
	}
	for i := 0; i < count(); i++ {
		if !replacement(i).levelCheck(iter, errors) {
			n.levelCorrect = false
		}
	}
	*n.level = body().getLevel()
	parameters := body().getLevelParams()
	for i := 0; i < count(); i++ {
		if parameters.contains(n.getSubFor(i)) {
			*n.level = max(*n.level, replacement(i).getLevel())
		}
	}
	for parameter := range parameters.all() {
		n.levelParams.addAll(sanySubstParamSet(parameter, n.substs))
	}
	if ap {
		// APSubstIn retains allParams and leaves nonLeibnizParams untouched.
		for parameter := range body().getAllParams().all() {
			n.allParams.addAll(sanySubstAllParamSet(parameter, n.substs))
		}
	} else {
		// SubstIn copies both body sets, then rewrites the current sets in order.
		// Later substitutions see changes made by earlier substitutions.
		n.allParams = newSanyLevelSymbolSetFrom(body().getAllParams())
		n.nonLeibnizParams = newSanyLevelSymbolSetFrom(body().getNonLeibnizParams())
		for i := 0; i < count(); i++ {
			parameter := n.substitutionAt(i).getOp()
			if n.allParams.contains(parameter) {
				n.allParams.remove(parameter)
				n.allParams.addAll(sanyRequireCanonicalLevelNode(n.substitutionAt(i).getExpr()).getAllParams())
				n.nonLeibnizParams.addAll(sanyRequireCanonicalLevelNode(n.substitutionAt(i).getExpr()).getNonLeibnizParams())
				if n.nonLeibnizParams.contains(parameter) {
					n.nonLeibnizParams.remove(parameter)
					n.nonLeibnizParams.addAll(sanyRequireCanonicalLevelNode(n.substitutionAt(i).getExpr()).getAllParams())
				}
			}
		}
	}
	constant := n.instantiatedModule.isConstant(errors)
	n.levelConstraints = sanySubstGetSubLCSet(body(), n.substs, constant, iter, errors)
	n.argLevelConstraints = sanySubstGetSubALCSet(body(), n.substs, iter, errors)
	n.argLevelParams = sanySubstGetSubALPSet(body(), n.substs)
	return n.levelCorrect
}
