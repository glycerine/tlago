// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"fmt"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

func (n *sanySemModuleNode) levelCheck(iter int32, errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if *n.levelChecked >= iter {
		return n.levelCorrect
	}
	*n.levelChecked = iter
	// Each contiguous recursive section gets the source's initialization and
	// two checking passes, including nonrecursive definitions in the section.
	for first := 0; first < len(n.opDefsInRecursiveSection); {
		current := first
		node := sanyModuleRecursiveNode(n, current)
		section := node.recursiveSection
		for {
			if node.inRecursive {
				*node.levelChecked = 1
				for i := 0; i < node.semArity(); i++ {
					sanyLevelArrayAt(node.argMaxLevels, i)
					node.argMaxLevels[i] = actionLevel
					sanyLevelArrayAt(node.argWeights, i)
					node.argWeights[i] = 1
				}
			} else {
				*node.levelChecked = 0
			}
			current++
			if current >= len(n.opDefsInRecursiveSection) {
				break
			}
			node = sanyModuleRecursiveNode(n, current)
			if node.recursiveSection != section {
				break
			}
		}
		maximum := constantLevel
		parameters, all := newSanyLevelSymbolSet(), newSanyLevelSymbolSet()
		for i := first; i < current; i++ {
			node = sanyModuleRecursiveNode(n, i)
			if node.inRecursive {
				*node.levelChecked = 0
			}
			node.levelCheck(1, errors)
			if node.inRecursive {
				for j := 0; j < node.semArity(); j++ {
					if sanyLevelArrayAt(node.argMaxLevels, j) < actionLevel {
						sanyAddFixedLevelMessage(errors, node.GetTreeNode(), "E4290", fmt.Sprintf("Argument %d of recursive operator %s is primed", j+1, node.semName()), int32(j+1), node.semName())
					}
				}
				maximum = max(maximum, *node.level)
				parameters.addAll(node.levelParams)
				all.addAll(node.allParams)
			}
		}
		for i := first; i < current; i++ {
			node = sanyModuleRecursiveNode(n, i)
			if node.inRecursive {
				*node.levelChecked = 2
			}
			*node.level = max(*node.level, maximum)
			node.levelParams.addAll(parameters)
			node.allParams.addAll(all)
		}
		for i := first; i < current; i++ {
			node = sanyModuleRecursiveNode(n, i)
			if node.inRecursive {
				*node.levelChecked = 1
			}
			node.levelCheck(2, errors)
		}
		first = current
	}
	n.levelCorrect = true
	for _, module := range n.getInnerModules() {
		if !module.levelCheck(1, errors) {
			n.levelCorrect = false
		}
	}
	definitions := n.getOpDefs()
	for _, definition := range definitions {
		if !definition.levelCheck(1, errors) {
			n.levelCorrect = false
		}
	}
	n.thmOrAssDefs = n.getThmOrAssDefs()
	for i := 0; i < len(n.thmOrAssDefs); i++ {
		definition := n.thmOrAssDefs[i]
		if !definition.levelCheck(1, errors) {
			n.levelCorrect = false
		}
	}
	top := n.getTopLevel()
	for _, child := range top {
		if !sanyRequireCanonicalLevelNode(child).levelCheck(1, errors) {
			n.levelCorrect = false
		}
	}
	declarations := n.getConstantDecls()
	for _, declaration := range declarations {
		n.levelParams.add(declaration)
		n.allParams.add(declaration)
	}
	if !n.isConstant(errors) {
		for _, declaration := range declarations {
			level := int32(constantLevel)
			n.levelConstraints.put(declaration, &level)
		}
	}
	for _, definition := range definitions {
		if definition == nil {
			panic(tlc.NewNullPointerException())
		}
		n.levelConstraints.putAll(sanyLevelConstraintMap(definition.getLevelConstraints()))
		n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(definition.getArgLevelConstraints()))
		for dependency := range definition.getArgLevelParams().all() {
			if dependency == nil {
				panic(tlc.NewNullPointerException())
			}
			if !dependency.occur(sanyModuleFormalSymbols(definition.getParams())) {
				n.argLevelParams.add(dependency)
			}
		}
	}
	for _, child := range top {
		node := sanyRequireCanonicalLevelNode(child)
		n.levelConstraints.putAll(sanyLevelConstraintMap(node.getLevelConstraints()))
		n.argLevelConstraints.putAll(sanyArgLevelConstraintMap(node.getArgLevelConstraints()))
		n.argLevelParams.addAll(node.getArgLevelParams())
	}
	return n.levelCorrect
}
func sanyModuleRecursiveNode(module *sanySemModuleNode, index int) *sanySemOpDefNode {
	node := sanyLevelArrayAt(module.opDefsInRecursiveSection, index)
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	return node
}
func (n *sanySemModuleNode) isConstant(errors *Diagnostics) bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if len(n.getVariableDecls()) > 0 {
		return false
	}
	n.levelCheck(1, errors)
	for _, definition := range n.getOpDefs() {
		if definition == nil {
			panic(tlc.NewNullPointerException())
		}
		if definition.getKind() != sanyModuleInstanceKind && sanyRequireCanonicalLevelNode(definition.getBody()).getLevel() != constantLevel {
			return false
		}
	}
	for i := 0; i < len(n.theoremVec); i++ {
		theorem := n.theoremVec[i]
		if theorem == nil {
			panic(tlc.NewNullPointerException())
		}
		if theorem.getLevel() != constantLevel {
			return false
		}
	}
	return true
}
func (n *sanySemModuleNode) getLevel() tlaLevel {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	panic(tlc.NewWrongInvocationException("Internal Error: Should never call ModuleNode.getLevel()"))
}

func sanyModuleFormalSymbols(formals []*sanyFormalParamNode) []sanySemSymbol {
	if formals == nil {
		return nil
	}
	result := make([]sanySemSymbol, len(formals))
	for i, formal := range formals {
		result[i] = formal
	}
	return result
}

// ModuleNode uses Java collection formatting, rather than the base formatter
// (which would call the deliberately forbidden module getLevel method).
func (n *sanySemModuleNode) levelDataToString() string {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	parameters := n.getLevelParams()
	p := "null"
	if parameters != nil {
		p = sanyLevelSymbolSetString(parameters)
	}
	out := "LevelParams: " + p + "\n"
	lc := n.getLevelConstraints()
	l := "null"
	if lc != nil {
		l = lc.String()
	}
	out += "LevelConstraints: " + l + "\n"
	alc := n.getArgLevelConstraints()
	a := "null"
	if alc != nil {
		a = alc.String()
	}
	out += "ArgLevelConstraints: " + a + "\n"
	alp := n.getArgLevelParams()
	deps := "null"
	if alp != nil {
		parts := []string{}
		for value := range alp.all() {
			item := "null"
			if value != nil {
				item = value.String()
			}
			parts = append(parts, item)
		}
		deps = "[" + strings.Join(parts, ", ") + "]"
	}
	return out + "ArgLevelParams: " + deps + "\n"
}
