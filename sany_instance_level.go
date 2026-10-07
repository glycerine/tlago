// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"fmt"
	"github.com/glycerine/tlago/tlc"
	"strings"
)

func (n *sanySemInstanceNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	n.levelParams = newSanyLevelSymbolSet()
	n.levelCorrect = true
	if !n.module.levelCheck(iter, errors) {
		n.levelCorrect = false
	}
	count := func() int {
		if n.substs == nil {
			panic(tlc.NewNullPointerException())
		}
		return len(n.substs)
	}
	sub := func(i int) *sanySemSubst { return sanyLevelArrayAt(n.substs, i) }
	expr := func(i int) sanyCanonicalLevelNode { return sanyRequireCanonicalLevelNode(sub(i).getExpr()) }
	message := func(code, text string, parameters ...any) {
		sanyAddFixedLevelMessage(errors, n.GetTreeNode(), code, text, parameters...)
	}
	levelMessage := func(param sanySemSymbol, bound tlaLevel) {
		message("E4245", fmt.Sprintf("Level error in instantiating module '%s':\nThe level of the expression or operator substituted for '%s' \nmust be at most %d.", n.module.semName(), sanyLevelSymbolName(param), bound), n.module.semName(), sanyLevelSymbolName(param), int32(bound))
	}
	for i := 0; i < count(); i++ {
		if !expr(i).levelCheck(iter, errors) {
			n.levelCorrect = false
		}
	}
	for i := 0; i < count(); i++ {
		param := sub(i).getOp()
		replacement := sanyRequireCanonicalLevelNode(sub(i).getExpr())
		replacement.levelCheck(iter, errors)
		if param == nil {
			panic(tlc.NewNullPointerException())
		}
		param.levelCheck(iter, errors)
		if !n.module.isConstant(errors) && replacement.getLevel() > param.getLevel() {
			if replacement.levelCheck(iter, errors) && param.levelCheck(iter, errors) {
				levelMessage(param, param.getLevel())
			}
			n.levelCorrect = false
		}
		if replacement.getKind() == sanyOpArgKind {
			op := sanySubstOperator(sanySubstAsOpArg(replacement.(sanySemanticGraphNode)))
			if op == nil {
				panic(tlc.NewNullPointerException())
			}
			if op.semKind() == sanyUserDefinedOpKind || op.semKind() == sanyBuiltInKind {
				definition, ok := op.(*sanySemOpDefNode)
				if !ok {
					panic(tlc.NewClassCastException("not an OpDefNode"))
				}
				if !definition.isLeibniz {
					message("E4244", fmt.Sprintf("Error in instantiating module '%s':\n A non-Leibniz operator substituted for '%s'.", n.module.semName(), param.semName()), n.module.semName(), param.semName())
				}
			}
		}
	}
	lc := n.module.getLevelConstraints()
	alc := n.module.getArgLevelConstraints()
	for i := 0; i < count(); i++ {
		param := sub(i).getOp()
		replacement := sanyRequireCanonicalLevelNode(sub(i).getExpr())
		bound := lc.get(param)
		if bound != nil && replacement.getLevel() > tlaLevel(*bound) {
			if replacement.levelCheck(iter, errors) {
				levelMessage(param, tlaLevel(*bound))
			}
			n.levelCorrect = false
		}
		if param == nil {
			panic(tlc.NewNullPointerException())
		}
		arity := param.semArity()
		if arity > 0 {
			if definition, ok := sanySubstOperator(sanySubstAsOpArg(replacement.(sanySemanticGraphNode))).(*sanySemOpDefNode); ok && definition != nil {
				for j := 0; j < arity; j++ {
					bound := alc.get(newSanyParamAndPosition(param, int32(j)))
					correct := definition.levelCheck(iter, errors)
					if bound != nil && definition.getMaxLevel(j) < tlaLevel(*bound) {
						if correct {
							message("E4246", fmt.Sprintf("Level error in instantiating module '%s':\nThe level of the argument %d of the operator %s \nmust be at least %d.", n.module.semName(), j+1, definition.semName(), *bound), n.module.semName(), int32(j+1), definition.semName(), *bound)
						}
						n.levelCorrect = false
					}
				}
			}
		}
	}
	for dependency := range n.module.getArgLevelParams().all() {
		for i := 0; i < count(); i++ {
			param := sub(i).getOp()
			for j := 0; j < count(); j++ {
				if dependency == nil {
					panic(tlc.NewNullPointerException())
				}
				if sanyLevelSymbolReference(dependency.op) == sanyLevelSymbolReference(param) && sanyLevelSymbolReference(dependency.param) == sanyLevelSymbolReference(sub(j).getOp()) {
					op := sanySubstOperator(sanySubstAsOpArg(sub(i).getExpr()))
					if op == nil {
						panic(tlc.NewNullPointerException())
					}
					correct := sanyRequireCanonicalLevelNode(op.(sanySemanticGraphNode)).levelCheck(iter, errors)
					if definition, ok := op.(*sanySemOpDefNode); ok && definition != nil && expr(j).getLevel() > definition.getMaxLevel(int(dependency.i)) {
						if correct && expr(j).levelCheck(iter, errors) {
							message("E4247", fmt.Sprintf("Level error when instantiating module '%s':\nThe level of the argument %d of the operator %s' \nmust be at most %d.", n.module.semName(), dependency.i, param.semName(), definition.getMaxLevel(int(dependency.i))), n.module.semName(), dependency.i, param.semName(), int32(definition.getMaxLevel(int(dependency.i))))
						}
						n.levelCorrect = false
					}
				}
			}
		}
	}
	mergeLC := func(set *sanySetOfLevelConstraints) {
		for param := range sanyLevelConstraintMap(set).All() {
			if !sanyInstanceSymbolOccur(param, n.params) {
				n.levelConstraints.put(param, set.get(param))
			}
		}
	}
	mergeALC := func(set *sanySetOfArgLevelConstraints) {
		for pap := range sanyArgLevelConstraintMap(set).All() {
			if pap == nil {
				panic(tlc.NewNullPointerException())
			}
			if !sanyInstanceSymbolOccur(pap.param, n.params) {
				n.argLevelConstraints.put(pap, set.get(pap))
			}
		}
	}
	mergeALP := func(set *sanyLevelSet[*sanyArgLevelParam]) {
		for dependency := range set.all() {
			if dependency == nil {
				panic(tlc.NewNullPointerException())
			}
			if !dependency.occur(sanyModuleFormalSymbols(n.params)) {
				n.argLevelParams.add(dependency)
			}
		}
	}
	constant := n.module.isConstant(errors)
	mergeLC(sanySubstGetSubLCSet(n.module, n.substs, constant, iter, errors))
	for i := 0; i < count(); i++ {
		mergeLC(expr(i).getLevelConstraints())
	}
	mergeALC(sanySubstGetSubALCSet(n.module, n.substs, iter, errors))
	for i := 0; i < count(); i++ {
		mergeALC(expr(i).getArgLevelConstraints())
	}
	mergeALP(sanySubstGetSubALPSet(n.module, n.substs))
	for i := 0; i < count(); i++ {
		mergeALP(expr(i).getArgLevelParams())
	}
	return n.levelCorrect
}
func sanyInstanceSymbolOccur(symbol sanySemSymbol, params []*sanyFormalParamNode) bool {
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
func (n *sanySemInstanceNode) getLevelParams() *sanyLevelSet[sanySemSymbol] {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.levelParams
}
func (n *sanySemInstanceNode) getLevelConstraints() *sanySetOfLevelConstraints {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.levelConstraints
}
func (n *sanySemInstanceNode) levelDataToString() string {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	lc, alc, alp := "null", "null", "null"
	if n.levelConstraints != nil {
		lc = n.levelConstraints.String()
	}
	if n.argLevelConstraints != nil {
		alc = n.argLevelConstraints.String()
	}
	if n.argLevelParams != nil {
		items := []string{}
		for value := range n.argLevelParams.all() {
			text := "null"
			if value != nil {
				text = value.String()
			}
			items = append(items, text)
		}
		alp = "[" + strings.Join(items, ", ") + "]"
	}
	return "LevelConstraints: " + lc + "\nArgLevelConstraints: " + alc + "\nArgLevelParams: " + alp + "\n"
}
