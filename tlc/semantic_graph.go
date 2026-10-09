// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlc

import (
	"fmt"
	"reflect"
)

// SemanticExplorerVisitor receives Context and substitution records as well as
// semantic nodes. Those records have no UID and are never registered as visited.
type SemanticExplorerVisitor struct {
	PreVisit  func(any)
	PostVisit func(any)
}

func (v *SemanticExplorerVisitor) pre(node any) {
	if v == nil {
		panic(NewNullPointerException())
	}
	if v.PreVisit != nil {
		v.PreVisit(node)
	}
}
func (v *SemanticExplorerVisitor) post(node any) {
	if v == nil {
		panic(NewNullPointerException())
	}
	if v.PostVisit != nil {
		v.PostVisit(node)
	}
}
func semanticExploreNull(node any) bool {
	return node == nil || (reflect.ValueOf(node).Kind() == reflect.Pointer && reflect.ValueOf(node).IsNil())
}
func semanticGraphArrayLength[T any](strict bool, values []T) int {
	if strict && values == nil {
		panic(NewNullPointerException())
	}
	return len(values)
}

// WalkSemanticGraph implements SANY walkGraph, including its callback order,
// UID table, mutation-visible indexed loops and required null boundaries.
// Substitutions are visited through their shared *SubstFields identity.
func WalkSemanticGraph(node any, table map[int32]any, visitor *SemanticExplorerVisitor) {
	register := func(node SemanticNode) bool {
		if table == nil {
			panic(NewNullPointerException())
		}
		var uid int32
		if symbol, ok := node.(*SymbolNode); ok {
			if symbol.SemanticBase == nil {
				panic(NewClassCastException("lookup alias has no semantic declaration"))
			}
			uid = symbol.SemanticBase.GetUID()
		} else if semantic, ok := node.(interface{ GetUID() int32 }); ok {
			uid = semantic.GetUID()
		} else {
			panic(NewClassCastException())
		}
		if !semanticExploreNull(table[uid]) {
			return true
		}
		table[uid] = node
		return false
	}
	walkSemanticExplore(node, register, visitor, true)
}

// Existing native callers also accept lookup aliases and AST-only LET/module
// fallbacks. Keep those memberships while sharing the actual traversal engine.
func walkSemanticGraph(roots []SemanticNode, preVisit func(SemanticNode)) {
	seen := map[semanticNodeKey]bool{}
	register := func(node SemanticNode) bool {
		key := newSemanticNodeKey(node)
		if seen[key] {
			return true
		}
		seen[key] = true
		return false
	}
	visitor := &SemanticExplorerVisitor{PreVisit: func(node any) {
		switch node.(type) {
		case *SemanticContext, *SubstFields:
			return
		}
		preVisit(node)
	}}
	for _, root := range roots {
		if !semanticExploreNull(root) {
			walkSemanticExplore(root, register, visitor, false)
		}
	}
}

func walkSemanticExplore(node any, register func(SemanticNode) bool, visitor *SemanticExplorerVisitor, strict bool) {
	if semanticExploreNull(node) {
		if !strict {
			return
		}
		panic(NewNullPointerException())
	}
	walk := func(child any) { walkSemanticExplore(child, register, visitor, strict) }
	optional := func(child any) {
		if !semanticExploreNull(child) {
			walk(child)
		}
	}
	switch n := node.(type) {
	case *SemanticContext:
		visitor.pre(n)
		entries := n.GetContextSymbolEnumeration()
		for entries.HasMoreElements() {
			entry := entries.nextEntry()
			if entry.key.Module {
				fmt.Printf("Bug in debugging caused by inner module %s\nSANY will throw a null pointer exception.\n", entry.key.Name)
			} else {
				walk(n.GetSymbol(entry.key))
			}
			visitor.post(n)
		}
		return
	case Subst:
		walk(n.SubstFields)
		return
	case *SubstFields:
		visitor.pre(n)
		optional(n.Op)
		optional(n.Expr)
		visitor.post(n)
		return
	case *AtNode:
		// AtNode intentionally neither registers itself nor follows its EXCEPT
		// references. Repeated occurrences receive repeated visitor callbacks.
		visitor.pre(n)
		visitor.post(n)
		return
	}
	if symbol, ok := node.(*SymbolNode); ok {
		if symbol.Definition != nil {
			node = symbol.Definition
		} else if theorem, ok := symbol.Data.(*ThmOrAssumpDefNode); ok {
			node = theorem
		}
	}
	if register(node) {
		return
	}
	visitor.pre(node)
	switch n := node.(type) {
	case *ModuleNode:
		optional(n.Context)
		if strict {
			for i := 0; i < len(n.TopLevel); i++ {
				walk(n.TopLevel[i])
			}
		} else {
			for _, statement := range n.graphStatements() {
				walk(statement)
			}
		}
	case *OpDefNode:
		if n.Params != nil {
			for i := 0; i < semanticGraphArrayLength(strict, n.Params); i++ {
				optional(n.Params[i])
			}
		}
		optional(n.Body)
		optional(n.StepNode)
	case *OpApplNode:
		optional(n.Operator)
		if n.UnbdedQuantSymbols != nil {
			for i := 0; i < semanticGraphArrayLength(strict, n.UnbdedQuantSymbols); i++ {
				optional(n.UnbdedQuantSymbols[i])
			}
		}
		if n.Args != nil {
			for i := 0; i < semanticGraphArrayLength(strict, n.Args); i++ {
				optional(n.Args[i])
			}
		}
		for i := 0; i < semanticGraphArrayLength(strict, n.BdedQuantBounds); i++ {
			optional(n.BdedQuantBounds[i])
		}
		if n.BdedQuantSymbolLists != nil {
			for i := 0; i < semanticGraphArrayLength(strict, n.BdedQuantSymbolLists); i++ {
				if n.BdedQuantSymbolLists[i] != nil {
					for j := 0; j < semanticGraphArrayLength(strict, n.BdedQuantSymbolLists[i]); j++ {
						optional(n.BdedQuantSymbolLists[i][j])
					}
				}
			}
		}
	case *OpArgNode:
		optional(n.Op)
	case *LetInNode:
		if !strict && n.Context == nil {
			for _, def := range coverageLetDefinitions(n.Lets) {
				optional(def)
			}
		} else {
			optional(n.Context)
		}
		optional(n.Body)
	case *SubstInNode:
		if n.Substs != nil {
			for i := 0; i < semanticGraphArrayLength(strict, n.Substs); i++ {
				optional(n.Substs[i].SubstFields)
			}
		}
		optional(n.Body)
	case *APSubstInNode:
		if n.Substs != nil {
			for i := 0; i < semanticGraphArrayLength(strict, n.Substs); i++ {
				optional(n.Substs[i].SubstFields)
			}
		}
		optional(n.Body)
	case *LabelNode:
		optional(n.Body)
		for i := 0; i < semanticGraphArrayLength(strict, n.Params); i++ {
			walk(n.Params[i])
		}
	case *AssumeProveNode:
		for i := 0; i < semanticGraphArrayLength(strict, n.Assumes); i++ {
			walk(n.Assumes[i])
		}
		walk(n.Prove)
	case *NewSymbNode:
		optional(n.Set)
	case *AssumeNode:
		optional(n.Assume)
	case *TheoremNode:
		optional(n.Theorem)
		optional(n.Proof)
	case *ThmOrAssumpDefNode:
		optional(n.Body)
	case *LeafProofNode:
		for i := 0; i < semanticGraphArrayLength(strict, n.Facts); i++ {
			walk(n.Facts[i])
		}
	case *NonLeafProofNode:
		for i := 0; i < semanticGraphArrayLength(strict, n.Steps); i++ {
			walk(n.Steps[i])
		}
		walk(n.Context)
	case *DefStepNode:
		for i := 0; i < semanticGraphArrayLength(strict, n.Defs); i++ {
			walk(n.Defs[i])
		}
	case *UseOrHideNode:
		for i := 0; i < semanticGraphArrayLength(strict, n.Facts); i++ {
			walk(n.Facts[i])
		}
		// FormalParamNode, OpDeclNode, InstanceNode and literals use the source
		// leaf traversal, even where getChildren exposes additional references.
	}
	visitor.post(node)
}

// SemanticNode.isDefinedWith deliberately rejects a reachable SubstInNode,
// even when the target also occurs. Preserve that source restriction.
func SemanticIsDefinedWith(node, target SemanticNode) bool {
	cycle, substIn := false, false
	walkSemanticGraph([]SemanticNode{node}, func(next SemanticNode) {
		if _, ok := next.(*SubstInNode); ok {
			substIn = true
		}
		if sameSemanticNode(next, target) {
			cycle = true
		}
	})
	return cycle && !substIn
}

// SemanticNode.substituteFor collects applications before mutating their
// operators and excludes applications belonging to the replacement's graph.
func SemanticSubstituteFor(roots []SemanticNode, replacement, original *OpDefNode) {
	var matches []*OpApplNode
	walkSemanticGraph(roots, func(node SemanticNode) {
		if application, ok := node.(*OpApplNode); ok && application != nil && application.Operator == original.Symbol {
			if !SemanticIsDefinedWith(replacement, application) {
				matches = append(matches, application)
			}
		}
	})
	for _, application := range matches {
		application.Operator = replacement.Symbol
	}
}
