// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"fmt"
	"reflect"

	"github.com/glycerine/tlago/tlc"
)

// Context and Subst are ExploreNodes, but have no semantic UID. Callbacks
// therefore accept those actual objects as well as semantic graph nodes.
type sanyExplorerVisitor struct {
	preVisit  func(any)
	postVisit func(any)
}

func (v *sanyExplorerVisitor) pre(node any) {
	if v == nil {
		panic(tlc.NewNullPointerException())
	}
	if v.preVisit != nil {
		v.preVisit(node)
	}
}
func (v *sanyExplorerVisitor) post(node any) {
	if v == nil {
		panic(tlc.NewNullPointerException())
	}
	if v.postVisit != nil {
		v.postVisit(node)
	}
}
func sanyExploreNull(node any) bool {
	return node == nil || (reflect.ValueOf(node).Kind() == reflect.Pointer && reflect.ValueOf(node).IsNil())
}

// Indexed source loops read the current array after each visitor callback.
// A callback may replace it; a null replacement fails at the next length read.
func sanyGraphArrayLength[T any](values []T) int {
	if values == nil {
		panic(tlc.NewNullPointerException())
	}
	return len(values)
}

// Port of the canonical classes' walkGraph methods. Do not replace this with
// getChildren: the two source APIs deliberately follow different edges.
func sanyWalkGraph(node any, table map[int32]any, visitor *sanyExplorerVisitor) {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	walk := func(child any) { sanyWalkGraph(child, table, visitor) }
	optional := func(child any) {
		if !sanyExploreNull(child) {
			walk(child)
		}
	}
	switch n := node.(type) {
	case *sanyContext:
		visitor.pre(n)
		// Java enumerates Hashtable keys, not the Context Pair history. Its
		// postVisit is inside the enumeration loop, including ModuleName entries.
		buckets := n.contentBuckets
		for i := len(buckets) - 1; i >= 0; i-- {
			for entry := buckets[i]; entry != nil; {
				next := entry.next
				if entry.entry.key.module {
					fmt.Println("Bug in debugging caused by inner module " + entry.entry.key.name)
					fmt.Println("SANY will throw a null pointer exception.")
				} else {
					walk(n.getSymbol(entry.entry.key.name))
				}
				visitor.post(n)
				entry = next
			}
		}
		return
	case *sanySemSubst:
		visitor.pre(n)
		optional(n.op)
		optional(n.expr)
		visitor.post(n)
		return
	case *sanySemAtNode:
		// AtNode intentionally neither registers itself nor follows its EXCEPT
		// references. Repeated occurrences receive repeated visitor callbacks.
		visitor.pre(n)
		visitor.post(n)
		return
	}
	semantic, ok := node.(sanySemanticGraphNode)
	if !ok {
		panic(tlc.NewClassCastException())
	}
	if table == nil {
		panic(tlc.NewNullPointerException())
	}
	uid := semantic.GetUID()
	if !sanyExploreNull(table[uid]) {
		return
	}
	table[uid] = node
	visitor.pre(node)
	switch n := node.(type) {
	case *sanySemModuleNode:
		optional(n.context)
		for i := 0; i < len(n.topLevelVec); i++ {
			walk(n.topLevelVec[i])
		}
	case *sanySemOpDefNode:
		if n.formalNodes != nil {
			for i := 0; i < sanyGraphArrayLength(n.formalNodes); i++ {
				optional(n.formalNodes[i])
			}
		}
		optional(n.body)
		optional(n.stepNode)
	case *sanySemOpApplNode:
		optional(n.operator)
		if n.unboundedBoundSymbols != nil {
			for i := 0; i < sanyGraphArrayLength(n.unboundedBoundSymbols); i++ {
				optional(n.unboundedBoundSymbols[i])
			}
		}
		if n.operands != nil {
			for i := 0; i < sanyGraphArrayLength(n.operands); i++ {
				optional(n.operands[i])
			}
		}
		for i := 0; i < sanyGraphArrayLength(n.ranges); i++ {
			optional(n.ranges[i])
		}
		if n.boundedBoundSymbols != nil {
			for i := 0; i < sanyGraphArrayLength(n.boundedBoundSymbols); i++ {
				if n.boundedBoundSymbols[i] != nil {
					for j := 0; j < sanyGraphArrayLength(n.boundedBoundSymbols[i]); j++ {
						optional(n.boundedBoundSymbols[i][j])
					}
				}
			}
		}
	case *sanySemOpArgNode:
		optional(n.operator)
	case *sanySemLetInNode:
		optional(n.context)
		optional(n.body)
	case *sanySemSubstInNode:
		if n.substs != nil {
			for i := 0; i < sanyGraphArrayLength(n.substs); i++ {
				optional(n.substs[i])
			}
		}
		optional(n.body)
	case *sanySemAPSubstInNode:
		if n.substs != nil {
			for i := 0; i < sanyGraphArrayLength(n.substs); i++ {
				optional(n.substs[i])
			}
		}
		optional(n.body)
	case *sanySemLabelNode:
		optional(n.body)
		for i := 0; i < sanyGraphArrayLength(n.formalNodes); i++ {
			walk(n.formalNodes[i])
		}
	case *sanySemAssumeProveNode:
		for i := 0; i < sanyGraphArrayLength(n.assumes); i++ {
			walk(n.assumes[i])
		}
		walk(n.prove)
	case *sanySemNewSymbNode:
		optional(n.set)
	case *sanySemAssumeNode:
		optional(n.assumeExpr)
	case *sanySemTheoremNode:
		optional(n.theoremExprOrAssumeProve)
		optional(n.proof)
	case *sanySemThmOrAssumpDefNode:
		optional(n.body)
	case *sanySemLeafProofNode:
		for i := 0; i < sanyGraphArrayLength(n.facts); i++ {
			walk(n.facts[i])
		}
	case *sanySemNonLeafProofNode:
		for i := 0; i < sanyGraphArrayLength(n.steps); i++ {
			walk(n.steps[i])
		}
		walk(n.context)
	case *sanySemDefStepNode:
		for i := 0; i < sanyGraphArrayLength(n.defs); i++ {
			walk(n.defs[i])
		}
	case *sanySemUseOrHideNode:
		for i := 0; i < sanyGraphArrayLength(n.facts); i++ {
			walk(n.facts[i])
		}
		// FormalParamNode, OpDeclNode, InstanceNode and literals use the source
		// leaf traversal, even where getChildren exposes additional references.
	}
	visitor.post(node)
}

func sanyIsDefinedWith(node, target sanySemanticGraphNode) bool {
	cycle, substitution := false, false
	visitor := &sanyExplorerVisitor{preVisit: func(next any) {
		if _, ok := next.(*sanySemSubstInNode); ok {
			substitution = true
		}
		if next == target {
			cycle = true
		}
	}}
	sanyWalkGraph(node, make(map[int32]any), visitor)
	return cycle && !substitution
}

// Collect matches before mutation, excluding applications reachable from the
// replacement definition according to the source isDefinedWith restriction.
func sanySubstituteFor(node sanySemanticGraphNode, replacement, original *sanySemOpDefNode) {
	matches := make(map[*sanySemOpApplNode]struct{})
	visitor := &sanyExplorerVisitor{preVisit: func(next any) {
		if application, ok := next.(*sanySemOpApplNode); ok && application.operator == original {
			if !sanyIsDefinedWith(replacement, application) {
				matches[application] = struct{}{}
			}
		}
	}}
	sanyWalkGraph(node, make(map[int32]any), visitor)
	for application := range matches {
		application.operator = replacement
	}
}
