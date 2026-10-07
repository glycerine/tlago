// Copyright (c) 2007 Microsoft Corporation. All rights reserved.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"strconv"
)

func (n *sanySemThmOrAssumpDefNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	count := func() int {
		if n.formalNodes == nil {
			panic(tlc.NewNullPointerException())
		}
		return len(n.formalNodes)
	}
	n.argMaxLevels = make([]tlaLevel, count())
	n.argWeights = make([]int, count())
	for i := 0; i < count(); i++ {
		n.argMaxLevels[i] = temporalLevel
		n.argWeights[i] = 0
		n.isLeibniz = true
		// Source allocates this array inside the loop: only the last slot is true.
		n.leibniz = make([]bool, count())
		n.leibniz[i] = true
	}
	body := func() sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(n.body) }
	n.levelCorrect = body().levelCheck(iter, errors)
	*n.level = body().getLevel()
	lcSet := body().getLevelConstraints()
	for i := 0; i < count(); i++ {
		if level := lcSet.get(sanyLevelArrayAt(n.formalNodes, i)); level != nil {
			sanyLevelArrayAt(n.argMaxLevels, i)
			n.argMaxLevels[i] = tlaLevel(*level)
		}
	}
	// The source assignment leaves weights unchanged, even for occurring formals.
	for i := 0; i < count(); i++ {
		if body().getLevelParams().contains(sanyLevelArrayAt(n.formalNodes, i)) {
			old := sanyLevelArrayAt(n.argWeights, i)
			n.argWeights[i] = old
		}
	}
	n.minMaxLevel = make([][]tlaLevel, count())
	alcSet := body().getArgLevelConstraints()
	for i := 0; i < count(); i++ {
		formal := sanyLevelArrayAt(n.formalNodes, i)
		if formal == nil {
			panic(tlc.NewNullPointerException())
		}
		arity := formal.semArity()
		if arity < 0 {
			panic(tlc.NewNegativeArraySizeException(strconv.Itoa(arity)))
		}
		n.minMaxLevel[i] = make([]tlaLevel, arity)
		for j := 0; j < arity; j++ {
			if level := alcSet.get(newSanyParamAndPosition(formal, int32(j))); level != nil {
				n.minMaxLevel[i][j] = tlaLevel(*level)
			}
		}
	}
	width := count()
	n.opLevelCond = make([][][]bool, width)
	for i := range n.opLevelCond {
		n.opLevelCond[i] = make([][]bool, width)
	}
	alpSet := body().getArgLevelParams()
	for i := 0; i < count(); i++ {
		for j := 0; j < count(); j++ {
			formal := sanyLevelArrayAt(n.formalNodes, i)
			if formal == nil {
				panic(tlc.NewNullPointerException())
			}
			arity := formal.semArity()
			if arity < 0 {
				panic(tlc.NewNegativeArraySizeException(strconv.Itoa(arity)))
			}
			n.opLevelCond[i][j] = make([]bool, arity)
			for k := 0; k < formal.semArity(); k++ {
				n.opLevelCond[i][j][k] = alpSet.contains(newSanyArgLevelParam(formal, int32(k), sanyLevelArrayAt(n.formalNodes, j)))
			}
		}
	}
	n.levelParams.addAll(body().getLevelParams())
	n.allParams.addAll(body().getAllParams())
	n.nonLeibnizParams.addAll(body().getNonLeibnizParams())
	for i := 0; i < count(); i++ {
		formal := sanyLevelArrayAt(n.formalNodes, i)
		n.levelParams.remove(formal)
		n.allParams.remove(formal)
		if n.nonLeibnizParams.contains(formal) {
			n.nonLeibnizParams.remove(formal)
			sanyLevelArrayAt(n.leibniz, i)
			n.leibniz[i] = false
			n.isLeibniz = false
		}
	}
	n.levelConstraints = newSanySetOfLevelConstraintsFrom(sanyLevelConstraintMap(lcSet))
	for i := 0; i < count(); i++ {
		n.levelConstraints.entries.Remove(sanyLevelArrayAt(n.formalNodes, i))
	}
	n.argLevelConstraints = newSanySetOfArgLevelConstraintsFrom(sanyArgLevelConstraintMap(alcSet))
	for i := 0; i < count(); i++ {
		formal := sanyLevelArrayAt(n.formalNodes, i)
		if formal == nil {
			panic(tlc.NewNullPointerException())
		}
		for j := 0; j < formal.semArity(); j++ {
			n.argLevelConstraints.entries.Remove(newSanyParamAndPosition(formal, int32(j)))
		}
	}
	for param := range alpSet.all() {
		if param == nil {
			panic(tlc.NewNullPointerException())
		}
		if !sanyLevelSymbolOccurs(param.op, n.formalNodes) || !sanyLevelSymbolOccurs(param.param, n.formalNodes) {
			n.argLevelParams.add(param)
		}
	}
	return n.levelCorrect
}

func (n *sanySemThmOrAssumpDefNode) getMaxLevel(index int) tlaLevel {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getMaxLevel called before levelCheck")
	if n.semArity() == -1 {
		index = 0
	}
	return sanyLevelArrayAt(n.argMaxLevels, index)
}
func (n *sanySemThmOrAssumpDefNode) getWeight(index int) int {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getWeight called before levelCheck")
	if n.semArity() == -1 {
		index = 0
	}
	return sanyLevelArrayAt(n.argWeights, index)
}
func (n *sanySemThmOrAssumpDefNode) getMinMaxLevel(i, j int) tlaLevel {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getMinMaxLevel called before levelCheck")
	if n.minMaxLevel == nil {
		return constantLevel
	}
	return sanyLevelArrayAt(sanyLevelArrayAt(n.minMaxLevel, i), j)
}
func (n *sanySemThmOrAssumpDefNode) getOpLevelCond(i, j, k int) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getOpLevelCond called before levelCheck")
	if n.opLevelCond == nil {
		return false
	}
	return sanyLevelArrayAt(sanyLevelArrayAt(sanyLevelArrayAt(n.opLevelCond, i), j), k)
}

// Source exposes the actual mutable Leibniz array.
func (n *sanySemThmOrAssumpDefNode) getIsLeibnizArg() []bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.leibniz
}
func (n *sanySemThmOrAssumpDefNode) getIsLeibniz() bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.isLeibniz
}
