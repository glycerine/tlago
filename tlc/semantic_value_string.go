package tlc

import (
	"sort"
	"strings"
)

// SemanticAllParams computes the parameter sets used by SANY's LevelNode.
// Its union/removal equations preserve symbol identity, bound parameters and
// substitution. Iteration handles recursive definitions without a depth limit.
func SemanticAllParams(root SemanticNode) []*SymbolNode {
	type parameterSet map[*SymbolNode]bool
	nodes := []SemanticNode{}
	indexes := map[semanticNodeKey]int{}
	var visit func(SemanticNode)
	visit = func(node SemanticNode) {
		if node == nil {
			return
		}
		key := newSemanticNodeKey(node)
		if _, ok := indexes[key]; ok {
			return
		}
		indexes[key] = len(nodes)
		nodes = append(nodes, node)
		switch n := node.(type) {
		case *OpApplNode:
			if n.Operator != nil && n.Operator.Definition != nil {
				visit(n.Operator.Definition)
			}
			for _, arg := range n.Args {
				visit(arg)
			}
			for _, bound := range n.BdedQuantBounds {
				visit(bound)
			}
		case *OpDefNode:
			visit(n.GetBody())
		case *OpArgNode:
			if n.Op != nil && n.Op.Definition != nil {
				visit(n.Op.Definition)
			}
		case *LetInNode:
			visit(n.Body)
		case *LabelNode:
			visit(n.Body)
		case *SubstInNode:
			visit(n.GetBody())
			for _, s := range n.Substs {
				visit(s.GetExpr())
			}
		case *APSubstInNode:
			visit(n.GetBody())
			for _, s := range n.Substs {
				visit(s.GetExpr())
			}
		case *AssumeNode:
			visit(n.Assume)
		}
	}
	visit(root)
	sets := make([]parameterSet, len(nodes))
	for i := range sets {
		sets[i] = parameterSet{}
	}
	changed := true
	for changed {
		changed = false
		for i, node := range nodes {
			out := parameterSet{}
			add := func(child SemanticNode) {
				if child == nil {
					return
				}
				if j, ok := indexes[newSemanticNodeKey(child)]; ok {
					for symbol := range sets[j] {
						out[symbol] = true
					}
				}
			}
			addOperator := func(symbol *SymbolNode) {
				if symbol == nil {
					return
				}
				if symbol.Definition != nil {
					add(symbol.Definition)
				} else if symbol.Kind == SymbolVariableDecl || symbol.Kind == SymbolConstantDecl || symbol.Kind == SymbolFormalParam {
					out[symbol] = true
				}
			}
			substitute := func(body SemanticNode, substitutions []Subst) {
				add(body)
				for _, s := range substitutions {
					op := s.GetOp()
					if out[op] {
						delete(out, op)
						add(s.GetExpr())
					}
				}
			}
			switch n := node.(type) {
			case *OpApplNode:
				addOperator(n.Operator)
				for _, arg := range n.Args {
					add(arg)
				}
				for _, bound := range n.BdedQuantBounds {
					add(bound)
				}
				for _, symbols := range n.BdedQuantSymbolLists {
					for _, symbol := range symbols {
						delete(out, symbol)
					}
				}
				for _, symbol := range n.UnbdedQuantSymbols {
					delete(out, symbol)
				}
			case *OpDefNode:
				add(n.GetBody())
				for _, symbol := range n.GetParams() {
					delete(out, symbol)
				}
			case *OpArgNode:
				addOperator(n.Op)
			case *LetInNode:
				add(n.Body)
			case *LabelNode:
				add(n.Body)
			case *SubstInNode:
				substitute(n.GetBody(), n.Substs)
			case *APSubstInNode:
				// APSubstInNode uses Subst.allParamSet on each original body
				// parameter, rather than SubstInNode's sequential replacement.
				if body := n.GetBody(); body != nil {
					if j, ok := indexes[newSemanticNodeKey(body)]; ok {
						for symbol := range sets[j] {
							replaced := false
							for _, s := range n.Substs {
								if s.GetOp() == symbol {
									add(s.GetExpr())
									replaced = true
									break
								}
							}
							if !replaced {
								out[symbol] = true
							}
						}
					}
				}
			case *AssumeNode:
				add(n.Assume)
			}
			equal := len(out) == len(sets[i])
			if equal {
				for symbol := range out {
					if !sets[i][symbol] {
						equal = false
						break
					}
				}
			}
			if !equal {
				sets[i] = out
				changed = true
			}
		}
	}
	if len(sets) == 0 {
		return nil
	}
	result := make([]*SymbolNode, 0, len(sets[0]))
	for symbol := range sets[0] {
		result = append(result, symbol)
	}
	return result
}

func semanticValueAllParams(node SemanticNode) []*SymbolNode {
	if carrier, ok := node.(interface{ GetCanonicalLevelData() any }); ok {
		if provider, ok := carrier.GetCanonicalLevelData().(interface{ RuntimeAllParams() []*SymbolNode }); ok {
			return provider.RuntimeAllParams()
		}
	}
	return SemanticAllParams(node)
}

// SemanticValueString ports SemanticNode.toString(IValue) and its OpApplNode
// override. VIEW tuples are labeled by allParams in variable-location order.
func SemanticValueString(node SemanticNode, value Value) string {
	if semanticExploreNull(node) || isNil(value) {
		panic(NewNullPointerException())
	}
	if _, ok := node.(*OpApplNode); ok {
		if tuple, ok := value.(*TupleValue); ok {
			params := semanticValueAllParams(node)
			if len(params) == len(tuple.GetElems()) {
				sort.Slice(params, func(i, j int) bool { return params[i].Name.VarLoc() < params[j].Name.VarLoc() })
				// Java's TreeSet comparator identifies equal variable locations.
				var result strings.Builder
				index := 0
				for i, param := range params {
					if i > 0 && params[i-1].Name.VarLoc() == param.Name.VarLoc() {
						continue
					}
					result.WriteString("/\\ ")
					result.WriteString(param.Name.String())
					result.WriteString(" = ")
					result.WriteString(ValuesPPR(tuple.GetElems()[index]))
					result.WriteByte('\n')
					index++
				}
				return result.String()
			}
		}
	}
	return ValuesPPR(value)
}
