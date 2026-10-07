// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"reflect"
	"strconv"

	"github.com/glycerine/tlago/tlc"
)

func (n *sanySemOpArgNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	operator := func() sanyCanonicalLevelNode {
		return sanyRequireCanonicalLevelNode(n.operator.(sanySemanticGraphNode))
	}
	// Nil symbols must fail after recording the iteration, as in the source.
	if sanyLevelSymbolReference(n.operator) == nil {
		panic(tlc.NewNullPointerException())
	}
	n.levelCorrect = operator().levelCheck(iter, errors)
	*n.level = operator().getLevel()
	n.levelParams = operator().getLevelParams()
	n.allParams = operator().getAllParams()
	n.levelConstraints = operator().getLevelConstraints()
	n.argLevelConstraints = operator().getArgLevelConstraints()
	n.argLevelParams = operator().getArgLevelParams()
	// Source does not copy nonLeibnizParams here.
	return n.levelCorrect
}

func (n *sanySemOpDefNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter || (!n.inRecursiveSection && *n.levelChecked > 0) {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	if n.getKind() == sanyNumberedProofStepKind {
		step := sanyRequireCanonicalLevelNode(n.stepNode)
		n.levelCorrect = step.levelCheck(iter, errors)
		// This call intentionally sees the iteration just set above and returns
		// without aggregating. Preserve the source's numbered-step behavior.
		return n.levelCheckSubnodes(iter, []sanyCanonicalLevelNode{step}, errors)
	}
	body := func() sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(n.body) }
	n.levelCorrect = body().levelCheck(iter, errors)
	*n.level = max(*n.level, body().getLevel())
	lcSet := body().getLevelConstraints()
	count := func() int {
		if n.formalNodes == nil {
			panic(tlc.NewNullPointerException())
		}
		return len(n.formalNodes)
	}
	for i := 0; i < count(); i++ {
		if level := lcSet.get(sanyLevelArrayAt(n.formalNodes, i)); level != nil {
			old := sanyLevelArrayAt(n.argMaxLevels, i)
			n.argMaxLevels[i] = min(old, tlaLevel(*level))
		}
	}
	for i := 0; i < count(); i++ {
		if body().getLevelParams().contains(sanyLevelArrayAt(n.formalNodes, i)) {
			old := sanyLevelArrayAt(n.argWeights, i)
			n.argWeights[i] = max(old, 1)
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

func sanyRequireCanonicalLevelNode(node sanySemanticGraphNode) sanyCanonicalLevelNode {
	if node == nil || reflect.ValueOf(node).IsNil() {
		panic(tlc.NewNullPointerException())
	}
	if levelNode, ok := node.(sanyCanonicalLevelNode); ok {
		return levelNode
	}
	switch literal := node.(type) {
	case *tlc.NumeralNode:
		return sanyCanonicalLiteral(literal, &literal.SemanticNodeBase)
	case *tlc.DecimalNode:
		return sanyCanonicalLiteral(literal, &literal.SemanticNodeBase)
	case *tlc.StringNode:
		return sanyCanonicalLiteral(literal, &literal.SemanticNodeBase)
	}
	panic(tlc.NewUnsupportedOperationException("Canonical level metadata is not yet integrated for this semantic node"))
}
func sanyLevelSymbolOccurs(symbol sanySemSymbol, params []*sanyFormalParamNode) bool {
	symbol = sanyLevelSymbolReference(symbol)
	if symbol == nil || params == nil {
		panic(tlc.NewNullPointerException())
	}
	for _, param := range params {
		if symbol == sanyLevelSymbolReference(param) {
			return true
		}
	}
	return false
}
func sanyLevelConstraintMap(set *sanySetOfLevelConstraints) *tlc.JavaSemanticMap[sanySemSymbol, *int32] {
	if set == nil {
		return nil
	}
	return set.entries
}
func sanyArgLevelConstraintMap(set *sanySetOfArgLevelConstraints) *tlc.JavaSemanticMap[*sanyParamAndPosition, *int32] {
	if set == nil {
		return nil
	}
	return set.entries
}
func sanyLevelArrayAt[T any](values []T, index int) T {
	if values == nil {
		panic(tlc.NewNullPointerException())
	}
	if index < 0 || index >= len(values) {
		panic(tlc.NewArrayIndexOutOfBoundsException(index, len(values)))
	}
	return values[index]
}
func (n *sanySemOpDefNode) getMaxLevel(index int) tlaLevel {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getMaxLevel called before levelCheck")
	if n.semArity() == -1 {
		index = 0
	}
	return sanyLevelArrayAt(n.argMaxLevels, index)
}
func (n *sanySemOpDefNode) getWeight(index int) int {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getWeight called before levelCheck")
	if n.semArity() == -1 {
		index = 0
	}
	return sanyLevelArrayAt(n.argWeights, index)
}
func (n *sanySemOpDefNode) getMinMaxLevel(i, j int) tlaLevel {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getMinMaxLevel called before levelCheck")
	if n.minMaxLevel == nil {
		return constantLevel
	}
	return sanyLevelArrayAt(sanyLevelArrayAt(n.minMaxLevel, i), j)
}
func (n *sanySemOpDefNode) getOpLevelCond(i, j, k int) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.requireChecked("getOpLevelCond called before levelCheck")
	if n.opLevelCond == nil {
		return false
	}
	return sanyLevelArrayAt(sanyLevelArrayAt(sanyLevelArrayAt(n.opLevelCond, i), j), k)
}
func (n *sanySemOpDefNode) getArgMaxLevels() []tlaLevel {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.argMaxLevels == nil {
		return nil
	}
	result := make([]tlaLevel, len(n.argMaxLevels))
	copy(result, n.argMaxLevels)
	return result
}
func (n *sanySemOpDefNode) getArgWeights() []int {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.argWeights == nil {
		return nil
	}
	result := make([]int, len(n.argWeights))
	copy(result, n.argWeights)
	return result
}

// Source exposes the actual mutable Leibniz array, unlike max levels/weights.
func (n *sanySemOpDefNode) getIsLeibnizArg() []bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.leibniz
}
func (n *sanySemOpDefNode) getIsLeibniz() bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.isLeibniz
}
