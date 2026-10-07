// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"fmt"

	"github.com/glycerine/tlago/tlc"
)

// Java AnyDefNode is implemented by exactly these two semantic classes.
type sanyAnyDefinition interface {
	sanySemSymbol
	sanyCanonicalLevelNode
	getMaxLevel(int) tlaLevel
	getWeight(int) int
	getMinMaxLevel(int, int) tlaLevel
	getOpLevelCond(int, int, int) bool
	getIsLeibnizArg() []bool
	getIsLeibniz() bool
	getParams() []*sanyFormalParamNode
}

func sanyApplicationDefinition(symbol sanySemSymbol) sanyAnyDefinition {
	switch node := symbol.(type) {
	case *sanySemOpDefNode:
		if node != nil {
			return node
		}
	case *sanySemThmOrAssumpDefNode:
		if node != nil {
			return node
		}
	}
	return nil
}
func sanyApplicationArgDefinition(node sanySemanticGraphNode) sanyAnyDefinition {
	if arg, ok := node.(*sanySemOpArgNode); ok && arg != nil {
		return sanyApplicationDefinition(arg.operator)
	}
	return nil
}
func sanyApplicationArgParameter(node sanySemanticGraphNode) (sanySemSymbol, bool) {
	if arg, ok := node.(*sanySemOpArgNode); ok && arg != nil {
		op := sanySubstOperator(arg)
		return op, sanySubstOperatorIsParam(op)
	}
	return nil, false
}
func (n *sanySemOpApplNode) getArg(param sanySemSymbol) sanySemanticGraphNode {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	definition := sanyApplicationDefinition(n.operator)
	if definition == nil {
		if sanyLevelSymbolReference(n.operator) == nil {
			panic(tlc.NewNullPointerException())
		}
		panic(tlc.NewClassCastException("not an AnyDefNode"))
	}
	formals := definition.getParams()
	if n.operands == nil {
		panic(tlc.NewNullPointerException())
	}
	for i := 0; i < len(n.operands); i++ {
		if sanyLevelSymbolReference(sanyLevelArrayAt(formals, i)) == sanyLevelSymbolReference(param) {
			return n.operands[i]
		}
	}
	return nil
}
func (n *sanySemOpApplNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	n.levelCorrect = true
	operands := func() int {
		if n.operands == nil {
			panic(tlc.NewNullPointerException())
		}
		return len(n.operands)
	}
	ranges := func() int {
		if n.ranges == nil {
			panic(tlc.NewNullPointerException())
		}
		return len(n.ranges)
	}
	operand := func(i int) sanyCanonicalLevelNode {
		return sanyRequireCanonicalLevelNode(sanyLevelArrayAt(n.operands, i))
	}
	bound := func(i int) sanyCanonicalLevelNode {
		return sanyRequireCanonicalLevelNode(sanyLevelArrayAt(n.ranges, i))
	}
	message := func(code, text string, parameters ...any) {
		sanyAddFixedLevelMessage(errors, n.GetTreeNode(), code, text, parameters...)
	}
	for i := 0; i < operands(); i++ {
		if sanyGraphNodePresent(n.operands[i]) && !operand(i).levelCheck(iter, errors) {
			n.levelCorrect = false
		}
	}
	for i := 0; i < ranges(); i++ {
		if sanyGraphNodePresent(n.ranges[i]) && !bound(i).levelCheck(iter, errors) {
			n.levelCorrect = false
		}
	}
	if definition := sanyApplicationDefinition(n.operator); definition != nil {
		correct := definition.levelCheck(iter, errors)
		for i := 0; i < operands(); i++ {
			graph := n.operands[i]
			if sanyGraphNodePresent(graph) {
				node := sanyRequireCanonicalLevelNode(graph)
				if node.getLevel() > definition.getMaxLevel(i) {
					if correct && node.levelCheck(iter, errors) {
						message("E4205", fmt.Sprintf("Level error in applying operator %s:\nThe level of argument %d exceeds the maximum level allowed by the operator.", definition.semName(), i+1), definition.semName(), int32(i+1))
					}
					n.levelCorrect = false
				}
				if argument := sanyApplicationArgDefinition(graph); argument != nil {
					argument.levelCheck(iter, errors)
					arity := argument.semArity()
					for j := 0; j < arity; j++ {
						if argument.getMaxLevel(j) < definition.getMinMaxLevel(i, j) {
							if correct && node.levelCheck(iter, errors) {
								message("E4272", fmt.Sprintf("Level error in applying operator %s:\nThe permitted level of argument %d of the operator argument %d \nmust be at least %d.", definition.semName(), j+1, i+1, definition.getMinMaxLevel(i, j)), definition.semName(), int32(j+1), int32(i+1), int32(definition.getMinMaxLevel(i, j)))
							}
							n.levelCorrect = false
						}
					}
					for j := 0; j < operands(); j++ {
						for k := 0; k < arity; k++ {
							if definition.getOpLevelCond(i, j, k) && operand(j).getLevel() > argument.getMaxLevel(k) {
								if node.levelCheck(iter, errors) && operand(j).levelCheck(iter, errors) {
									message("E4273", fmt.Sprintf("Level error in applying operator %s:\nThe level of argument %d exceeds the maximum level allowed by the operator.", definition.semName(), j+1), definition.semName(), int32(j+1))
								}
								n.levelCorrect = false
							}
						}
					}
				}
			}
		}
		for i := 0; i < ranges(); i++ {
			graph := n.ranges[i]
			if sanyGraphNodePresent(graph) {
				node := sanyRequireCanonicalLevelNode(graph)
				correct := node.levelCheck(iter, errors)
				if node.getLevel() > actionLevel {
					if correct {
						symbol := sanyLevelArrayAt(sanyLevelArrayAt(n.boundedBoundSymbols, i), 0)
						name := "null"
						if symbol != nil {
							name = sanyLevelSymbolString(symbol)
						}
						message("E4315", fmt.Sprintf("Level error in applying operator %s:\nThe level of the range for the bounded variable %s \nexceeds the maximum level allowed by the operator.", definition.semName(), name), definition.semName(), symbol)
					}
					n.levelCorrect = false
				}
			}
		}
		*n.level = definition.getLevel()
		for i := 0; i < operands(); i++ {
			if sanyGraphNodePresent(n.operands[i]) && definition.getWeight(i) == 1 {
				*n.level = max(*n.level, operand(i).getLevel())
			}
		}
		for i := 0; i < ranges(); i++ {
			*n.level = max(*n.level, bound(i).getLevel())
		}
		n.levelParams.addAll(definition.getLevelParams())
		n.allParams.addAll(definition.getAllParams())
		n.nonLeibnizParams.addAll(definition.getNonLeibnizParams())
		arity := definition.semArity()
		for i := 0; i < operands(); i++ {
			if sanyGraphNodePresent(n.operands[i]) && definition.getWeight(i) == 1 {
				n.levelParams.addAll(operand(i).getLevelParams())
			}
			if sanyGraphNodePresent(n.operands[i]) {
				n.allParams.addAll(operand(i).getAllParams())
				n.nonLeibnizParams.addAll(operand(i).getNonLeibnizParams())
			}
			index := i
			if arity == -1 {
				index = 0
			}
			if !sanyLevelArrayAt(definition.getIsLeibnizArg(), index) {
				n.nonLeibnizParams.addAll(operand(i).getAllParams())
			}
		}
		for i := 0; i < ranges(); i++ {
			n.levelParams.addAll(bound(i).getLevelParams())
			n.allParams.addAll(bound(i).getAllParams())
			n.nonLeibnizParams.addAll(bound(i).getNonLeibnizParams())
		}
		allBounds := newSanyLevelSymbolSet()
		for _, symbol := range n.unboundedBoundSymbols {
			allBounds.add(symbol)
		}
		for _, group := range n.boundedBoundSymbols {
			for _, symbol := range group {
				allBounds.add(symbol)
			}
		}
		for symbol := range allBounds.all() {
			n.levelParams.remove(symbol)
			n.allParams.remove(symbol)
			n.nonLeibnizParams.remove(symbol)
		}
		n.levelConstraints.putAll(sanyLevelConstraintMap(definition.getLevelConstraints()))
		for i := 0; i < operands(); i++ {
			if sanyGraphNodePresent(n.operands[i]) {
				if allBounds.len() == 0 {
					n.levelConstraints.putAll(sanyLevelConstraintMap(operand(i).getLevelConstraints()))
				} else {
					constraints := operand(i).getLevelConstraints()
					for symbol := range sanyLevelConstraintMap(constraints).All() {
						if !allBounds.contains(symbol) {
							n.levelConstraints.put(symbol, constraints.get(symbol))
						}
					}
				}
			}
		}
		for i := 0; i < ranges(); i++ {
			n.levelConstraints.putAll(sanyLevelConstraintMap(bound(i).getLevelConstraints()))
		}
		for i := 0; i < operands(); i++ {
			maximum := int32(definition.getMaxLevel(i))
			if sanyGraphNodePresent(n.operands[i]) {
				for symbol := range operand(i).getLevelParams().all() {
					n.levelConstraints.put(symbol, &maximum)
				}
			}
		}
		for i := 0; i < operands(); i++ {
			if argument := sanyApplicationArgDefinition(n.operands[i]); argument != nil {
				argument.levelCheck(iter, errors)
				arity := argument.semArity()
				for j := 0; j < operands(); j++ {
					for k := 0; k < arity; k++ {
						if definition.getOpLevelCond(i, j, k) {
							maximum := int32(argument.getMaxLevel(k))
							for symbol := range operand(j).getLevelParams().all() {
								n.levelConstraints.put(symbol, &maximum)
							}
						}
					}
				}
				if !argument.getIsLeibniz() {
					for j := 0; j < operands(); j++ {
						for k := 0; k < arity; k++ {
							if definition.getOpLevelCond(i, j, k) && !sanyLevelArrayAt(argument.getIsLeibnizArg(), k) {
								n.nonLeibnizParams.addAll(operand(j).getAllParams())
							}
						}
					}
				}
			}
		}
		dependencies := definition.getArgLevelParams()
		for dependency := range dependencies.all() {
			if dependency == nil {
				panic(tlc.NewNullPointerException())
			}
			if argument := sanyApplicationArgDefinition(n.getArg(dependency.op)); argument != nil {
				argument.levelCheck(iter, errors)
				maximum := int32(argument.getMaxLevel(int(dependency.i)))
				n.levelConstraints.put(dependency.param, &maximum)
			}
		}
		n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(definition.getArgLevelConstraints()))
		for i := 0; i < operands(); i++ {
			if sanyGraphNodePresent(n.operands[i]) {
				n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(operand(i).getArgLevelConstraints()))
			}
		}
		for i := 0; i < ranges(); i++ {
			n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(bound(i).getArgLevelConstraints()))
		}
		for i := 0; i < operands(); i++ {
			if symbol, ok := sanyApplicationArgParameter(n.operands[i]); ok {
				arity := symbol.semArity()
				for j := 0; j < arity; j++ {
					value := int32(definition.getMinMaxLevel(i, j))
					n.argLevelConstraints.put(newSanyParamAndPosition(symbol, int32(j)), &value)
				}
				for j := 0; j < operands(); j++ {
					for k := 0; k < arity; k++ {
						if definition.getOpLevelCond(i, j, k) {
							value := int32(operand(j).getLevel())
							n.argLevelConstraints.put(newSanyParamAndPosition(symbol, int32(k)), &value)
						}
					}
				}
			}
		}
		for dependency := range dependencies.all() {
			if dependency == nil {
				panic(tlc.NewNullPointerException())
			}
			graph := n.getArg(dependency.op)
			if sanyGraphNodePresent(graph) {
				node := sanyRequireCanonicalLevelNode(graph)
				node.levelCheck(iter, errors)
				value := int32(node.getLevel())
				n.argLevelConstraints.put(newSanyParamAndPosition(dependency.op, dependency.i), &value)
			}
		}
		n.argLevelParams = newSanyArgLevelSet()
		for i := 0; i < operands(); i++ {
			if sanyGraphNodePresent(n.operands[i]) {
				if allBounds.len() == 0 {
					n.argLevelParams.addAll(operand(i).getArgLevelParams())
				} else {
					for dependency := range operand(i).getArgLevelParams().all() {
						if dependency == nil {
							panic(tlc.NewNullPointerException())
						}
						if !allBounds.contains(dependency.param) {
							n.argLevelParams.add(dependency)
						}
					}
				}
			}
		}
		for i := 0; i < ranges(); i++ {
			n.argLevelParams.addAll(bound(i).getArgLevelParams())
		}
		for dependency := range dependencies.all() {
			if dependency == nil {
				panic(tlc.NewNullPointerException())
			}
			graph := n.getArg(dependency.op)
			if !sanyGraphNodePresent(graph) {
				graph = n.getArg(dependency.param)
				if !sanyGraphNodePresent(graph) {
					n.argLevelParams.add(dependency)
				} else {
					node := sanyRequireCanonicalLevelNode(graph)
					node.levelCheck(iter, errors)
					for symbol := range node.getLevelParams().all() {
						n.argLevelParams.add(newSanyArgLevelParam(dependency.op, dependency.i, symbol))
					}
				}
			} else if symbol, ok := sanyApplicationArgParameter(graph); ok {
				n.argLevelParams.add(newSanyArgLevelParam(symbol, dependency.i, dependency.param))
			}
		}
		for i := 0; i < operands(); i++ {
			if symbol, ok := sanyApplicationArgParameter(n.operands[i]); ok {
				arity := symbol.semArity()
				for j := 0; j < operands(); j++ {
					for k := 0; k < arity; k++ {
						if definition.getOpLevelCond(i, j, k) {
							for param := range operand(j).getLevelParams().all() {
								n.argLevelParams.add(newSanyArgLevelParam(symbol, int32(k), param))
							}
						}
					}
				}
			}
		}
	} else {
		if sanyLevelSymbolReference(n.operator) == nil {
			panic(tlc.NewNullPointerException())
		}
		operator := func() sanyCanonicalLevelNode {
			return sanyRequireCanonicalLevelNode(n.operator.(sanySemanticGraphNode))
		}
		operator().levelCheck(iter, errors)
		*n.level = operator().getLevel()
		for i := 0; i < operands(); i++ {
			operand(i).levelCheck(iter, errors)
			*n.level = max(*n.level, operand(i).getLevel())
		}
		n.levelParams = newSanyLevelSymbolSet()
		n.levelParams.add(n.operator)
		n.allParams.add(n.operator)
		for i := 0; i < operands(); i++ {
			n.levelParams.addAll(operand(i).getLevelParams())
			n.allParams.addAll(operand(i).getAllParams())
			n.nonLeibnizParams.addAll(operand(i).getNonLeibnizParams())
		}
		n.levelConstraints = newSanySetOfLevelConstraints()
		for i := 0; i < operands(); i++ {
			n.levelConstraints.putAll(sanyLevelConstraintMap(operand(i).getLevelConstraints()))
		}
		n.argLevelConstraints = newSanySetOfArgLevelConstraints()
		for i := 0; i < operands(); i++ {
			n.argLevelConstraints.putArgument(n.operator, int32(i), int32(operand(i).getLevel()))
			n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(operand(i).getArgLevelConstraints()))
		}
		n.argLevelParams = newSanyArgLevelSet()
		for i := 0; i < operands(); i++ {
			for symbol := range operand(i).getLevelParams().all() {
				n.argLevelParams.add(newSanyArgLevelParam(n.operator, int32(i), symbol))
			}
			n.argLevelParams.addAll(operand(i).getArgLevelParams())
		}
	}
	n.levelCheckTemporalApplication(errors)
	return n.levelCorrect
}

func (n *sanySemOpApplNode) levelCheckTemporalApplication(errors *Diagnostics) {
	if sanyLevelSymbolReference(n.operator) == nil {
		panic(tlc.NewNullPointerException())
	}
	name := sanyApplicationOperatorName(n.operator)
	arg := func(i int) sanyCanonicalLevelNode {
		return sanyRequireCanonicalLevelNode(sanyLevelArrayAt(n.operands, i))
	}
	message := func(code, text string, parameters ...any) {
		sanyAddFixedLevelMessage(errors, n.GetTreeNode(), code, text, parameters...)
	}
	if name == "[]" || name == "<>" {
		graph := sanyLevelArrayAt(n.operands, 0)
		if _, ok := graph.(*sanySemOpArgNode); ok && sanyGraphNodePresent(graph) {
			panic(tlc.NewClassCastException("not an ExprNode"))
		}
		node := sanyRequireCanonicalLevelNode(graph)
		if node.getLevel() == actionLevel && node.getKind() == sanyOpApplKind {
			application, ok := graph.(*sanySemOpApplNode)
			if !ok {
				panic(tlc.NewClassCastException("not an OpApplNode"))
			}
			if application.operator == nil {
				panic(tlc.NewNullPointerException())
			}
			expected, code, text := "$SquareAct", "E4310", "[] followed by action not of form [A]_v."
			if name == "<>" {
				expected, code, text = "$AngleAct", "E4311", "<> followed by action not of form <<A>>_v."
			}
			if sanyApplicationOperatorName(application.operator) != expected {
				message(code, text)
				n.levelCorrect = false
			}
		}
	}
	if name == "~>" || name == "-+->" {
		if arg(0).getLevel() == actionLevel || arg(1).getLevel() == actionLevel {
			message("E4312", "Action used where only temporal formula or state predicate allowed.")
			n.levelCorrect = false
		}
	}
	if name == "\\land" || name == "\\lor" || name == "=>" || name == "\\equiv" || name == "$ConjList" || name == "$DisjList" {
		temporal, action := false, false
		if n.operands == nil {
			panic(tlc.NewNullPointerException())
		}
		for i := 0; i < len(n.operands); i++ {
			temporal = temporal || arg(i).getLevel() == temporalLevel
			action = action || arg(i).getLevel() == actionLevel
		}
		if temporal && action {
			printed := name
			if printed == "$ConjList" {
				printed = "Conjunction list"
			}
			if printed == "$DisjList" {
				printed = "Disjunction list"
			}
			message("E4313", printed+" has both temporal formula and action as arguments.", printed)
			n.levelCorrect = false
		}
	}
	if *n.level == temporalLevel && (name == "$BoundedExists" || name == "$BoundedForall") {
		if n.ranges == nil {
			panic(tlc.NewNullPointerException())
		}
		for i := 0; i < len(n.ranges); i++ {
			graph := n.ranges[i]
			if sanyRequireCanonicalLevelNode(graph).getLevel() == actionLevel {
				sanyAddFixedLevelMessage(errors, sanyGraphTreeNode(graph), "E4314", "Action-level bound of quantified temporal formula.")
				n.levelCorrect = false
			}
		}
	}
}

// OpAppl's final name tests call getName().toString(), not String.valueOf.
func sanyApplicationOperatorName(symbol sanySemSymbol) string {
	symbol = sanyLevelSymbolReference(symbol)
	if symbol == nil {
		panic(tlc.NewNullPointerException())
	}
	if formal, ok := symbol.(*sanyFormalParamNode); ok && formal.nullName {
		panic(tlc.NewNullPointerException())
	}
	return symbol.semName()
}
