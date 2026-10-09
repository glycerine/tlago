// Copyright (c) 2003 Compaq Corporation.  All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation.  All rights reserved.
// Port of the represented SANY walkGraph/isDefinedWith/substituteFor paths.

package tlc

// Walk the represented SANY graph by semantic identity, without following
// evaluation tool objects. Java symbols that define operators are OpDefNodes;
// Go retains that relation in SymbolNode.Definition.
func walkSemanticGraph(roots []SemanticNode, preVisit func(SemanticNode)) {
	seen := map[semanticNodeKey]bool{}
	var walk func(SemanticNode)
	walk = func(node SemanticNode) {
		if node == nil {
			return
		}
		if symbol, ok := node.(*SymbolNode); ok {
			if symbol == nil {
				return
			}
			if symbol.Definition != nil {
				node = symbol.Definition
			} else if theorem, ok := symbol.Data.(*ThmOrAssumpDefNode); ok {
				node = theorem
			}
		}
		// AtNode has no visited-table registration in the source.
		if _, ok := node.(*AtNode); ok {
			preVisit(node)
			return
		}
		key := newSemanticNodeKey(node)
		if seen[key] {
			return
		}
		seen[key] = true
		preVisit(node)
		switch n := node.(type) {
		case *ModuleNode:
			if n == nil {
				return
			}
			if n.Context != nil {
				n.Context.WalkGraphNodes(walk)
			}
			for _, top := range n.graphStatements() {
				walk(top)
			}
		case *OpDefNode:
			if n == nil {
				return
			}
			for _, param := range n.Params {
				walk(param)
			}
			walk(n.Body)
			walk(n.StepNode)
		case *OpApplNode:
			if n == nil {
				return
			}
			walk(n.Operator)
			for _, symbol := range n.UnbdedQuantSymbols {
				walk(symbol)
			}
			for _, arg := range n.Args {
				walk(arg)
			}
			for _, bound := range n.BdedQuantBounds {
				walk(bound)
			}
			for _, symbols := range n.BdedQuantSymbolLists {
				for _, symbol := range symbols {
					walk(symbol)
				}
			}
		case *LetInNode:
			if n != nil {
				if n.Context != nil {
					n.Context.WalkGraphNodes(walk)
				} else {
					for _, def := range coverageLetDefinitions(n.Lets) {
						walk(def)
					}
				}
				walk(n.Body)
			}
		case *SubstInNode:
			if n != nil {
				for _, subst := range n.Substs {
					walk(subst.Op)
					walk(subst.Expr)
				}
				walk(n.Body)
			}
		case *APSubstInNode:
			if n != nil {
				for _, subst := range n.Substs {
					walk(subst.Op)
					walk(subst.Expr)
				}
				walk(n.Body)
			}
		case *OpArgNode:
			if n != nil {
				walk(n.Op)
			}
		case *LabelNode:
			walk(n.Body)
			for _, param := range n.Params {
				walk(param)
			}
		case *ThmOrAssumpDefNode:
			walk(n.Body)
		case *AssumeProveNode:
			for _, assume := range n.Assumes {
				walk(assume)
			}
			walk(n.Prove)
		case *NewSymbNode:
			walk(n.Set)
		case *AssumeNode:
			walk(n.Assume)
		case *TheoremNode:
			walk(n.Theorem)
			walk(n.Proof)
		case *LeafProofNode:
			for _, fact := range n.Facts {
				walk(fact)
			}
		case *NonLeafProofNode:
			for _, step := range n.Steps {
				walk(step)
			}
			if n.Context != nil {
				n.Context.WalkGraphNodes(walk)
			}
		case *DefStepNode:
			for _, def := range n.Defs {
				walk(def)
			}
		case *UseOrHideNode:
			for _, fact := range n.Facts {
				walk(fact)
			}
		}
	}
	for _, root := range roots {
		walk(root)
	}
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
