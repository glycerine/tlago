// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import (
	"unsafe"

	"github.com/glycerine/tlago/tlc"
)

type tlcCanonicalSubstArray struct {
	first  **sanySemSubst
	length int
}

func (b *tlcBridge) retainCanonicalExpression(expr Expr, node tlc.SemanticNode) tlc.SemanticNode {
	node = b.withExprLocation(expr, node)
	if source := sanyGeneratedExpressionNode(expr); source != nil {
		if b.canonicalGraphs == nil {
			b.canonicalGraphs = map[sanySemanticGraphNode]tlc.SemanticNode{}
		}
		if owner, ok := source.(interface{ runtimeSemanticBase() *tlc.SemanticNodeBase }); ok {
			base := owner.runtimeSemanticBase()
			if source.Kind() == tlc.SemanticKindOf(node) {
				switch node := node.(type) {
				case *tlc.OpApplNode:
					node.SemanticNodeBase = base
				case *tlc.OpArgNode:
					node.SemanticNodeBase = base
				case *tlc.LabelNode:
					node.SemanticNodeBase = base
				case *tlc.AtNode:
					node.SemanticNodeBase = base
				case *tlc.SubstInNode:
					node.SemanticNodeBase = base
				case *tlc.APSubstInNode:
					node.SemanticNodeBase = base
				}
			}
		}
		b.canonicalGraphs[source] = node
	}
	return node
}

func (b *tlcBridge) canonicalContext(source *sanyContext) *tlc.SemanticContext {
	if source == nil {
		return nil
	}
	if context := b.canonicalContexts[source]; context != nil {
		return context
	}
	if b.canonicalContexts == nil {
		b.canonicalContexts = map[*sanyContext]*tlc.SemanticContext{}
	}
	context := tlc.NewSemanticContext(b.processor.ModuleTbl)
	b.canonicalContexts[source] = context
	context.ImportState(source.runtimeState(func(symbol sanySemSymbol) tlc.SemanticNode {
		return b.canonicalGraph(symbol.(sanySemanticGraphNode))
	}))
	return context
}

// Source contexts can contain symbols absent from the evaluator's filtered
// declaration lists. Adapt their actual graph identities, publishing shells
// before following children or source-definition links.
func (b *tlcBridge) canonicalGraph(source sanySemanticGraphNode) tlc.SemanticNode {
	if source == nil {
		return nil
	}
	if node := b.canonicalGraphs[source]; node != nil {
		return node
	}
	if b.canonicalGraphs == nil {
		b.canonicalGraphs = map[sanySemanticGraphNode]tlc.SemanticNode{}
	}
	remember := func(node tlc.SemanticNode) { b.canonicalGraphs[source] = node }
	switch source := source.(type) {
	case *tlc.NumeralNode, *tlc.DecimalNode, *tlc.StringNode:
		remember(source)
		return source
	case *sanyFormalParamNode:
		node := b.canonicalFormalParameter(source)
		remember(node)
		return node
	case *sanySemOpDeclNode:
		for _, module := range b.spec.Modules {
			if module.semanticNode == source.module {
				node := b.declarationSymbol(module, source.semName())
				remember(node)
				return node
			}
		}
		panic(tlc.NewUnsupportedOperationException("runtime adapter for declaration without a source module"))
	case *sanySemOpDefNode:
		if node := b.canonicalDefinitions[source]; node != nil {
			remember(node)
			return node
		}
		// Module definitions already have evaluator symbols used by native and
		// config overrides. Find the AST view by source identity, so importing
		// a LET Context cannot create a competing symbol for that definition.
		for definition, module := range b.definitionModules {
			if definition.semanticNode == source {
				node := b.convertSourceDefinitionAs(module+"!"+definition.Name, definition)
				remember(node)
				return node
			}
		}
		if source.semKind() == sanyBuiltInKind {
			node := b.builtinDefinition(source.semName())
			remember(node)
			return node
		}
		params := b.canonicalParameters(source.formalNodes)
		symbol := tlc.NewSymbolNode(source.semName())
		node := tlc.NewOpDefNodeForSymbolWithBase(symbol, params, nil, source.SemanticNodeBase)
		remember(node)
		if b.canonicalDefinitions == nil {
			b.canonicalDefinitions = map[*sanySemOpDefNode]*tlc.OpDefNode{}
		}
		b.canonicalDefinitions[source] = node
		node.Symbol.Data = node
		node.Local, node.InRecursive = source.semLocal(), source.inRecursive
		node.CompoundID = source.compoundID
		if origin := source.getSource(); origin != source {
			node.SourceDefinition = b.canonicalGraph(origin).(*tlc.OpDefNode)
		}
		node.Body = b.canonicalGraph(source.body)
		return node
	case *sanySemThmOrAssumpDefNode:
		node := &tlc.ThmOrAssumpDefNode{SemanticNodeBase: source.SemanticNodeBase,
			Name: tlc.UniqueStringOf(source.semName()), Params: b.canonicalParameters(source.formalNodes), Local: source.semLocal()}
		remember(node)
		node.Symbol = &tlc.SymbolNode{SemanticBase: source.SemanticNodeBase, Name: node.Name, Arity: source.semArity(), Data: node}
		if origin := source.getSource(); origin != source {
			node.SourceDefinition = b.canonicalGraph(origin).(*tlc.ThmOrAssumpDefNode)
		}
		node.Body = b.canonicalGraph(source.body)
		return node
	case *sanySemOpApplNode:
		node := &tlc.OpApplNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.Operator = b.canonicalSymbol(source.operator)
		node.Args = b.canonicalGraphArray(source.operands)
		node.BdedQuantBounds = b.canonicalGraphArray(source.ranges)
		node.BdedQuantATuple = source.tupleOrs
		node.UnbdedQuantSymbols = b.canonicalParameters(source.unboundedBoundSymbols)
		if source.boundedBoundSymbols != nil {
			node.BdedQuantSymbolLists = make([][]*tlc.SymbolNode, len(source.boundedBoundSymbols))
			for i, symbols := range source.boundedBoundSymbols {
				node.BdedQuantSymbolLists[i] = b.canonicalParameters(symbols)
			}
		}
		return node
	case *sanySemOpArgNode:
		node := &tlc.OpArgNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.Op = b.canonicalSymbol(source.operator)
		return node
	case *sanySemLetInNode:
		node := b.canonicalLets[source]
		if node != nil {
			remember(node)
			return node
		}
		node = tlc.NewLetInNodeWithBase(source.SemanticNodeBase, nil)
		remember(node)
		if b.canonicalLets == nil {
			b.canonicalLets = map[*sanySemLetInNode]*tlc.LetInNode{}
		}
		b.canonicalLets[source] = node
		for _, definition := range source.getLets() {
			node.Lets = append(node.Lets, b.canonicalGraph(definition).(*tlc.OpDefNode))
		}
		node.Context = b.canonicalContext(source.context)
		node.Body = b.canonicalGraph(source.body)
		return node
	case *sanySemSubstInNode:
		node := &tlc.SubstInNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.Substs = b.canonicalSubstitutions(source.substs)
		node.Body = b.canonicalGraph(source.body)
		return node
	case *sanySemAPSubstInNode:
		node := &tlc.APSubstInNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.Substs = b.canonicalSubstitutions(source.substs)
		node.Body = b.canonicalGraph(source.body)
		return node
	case *sanySemLabelNode:
		node := &tlc.LabelNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.Body = b.canonicalGraph(source.body)
		return node
	case *sanySemAtNode:
		node := &tlc.AtNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		return node
	default:
		panic(tlc.NewUnsupportedOperationException("runtime adapter for canonical context graph node"))
	}
}

func (b *tlcBridge) canonicalParameters(source []*sanyFormalParamNode) []*tlc.SymbolNode {
	if source == nil {
		return nil
	}
	nodes := make([]*tlc.SymbolNode, len(source))
	for i, parameter := range source {
		nodes[i] = b.canonicalFormalParameter(parameter)
	}
	return nodes
}

func (b *tlcBridge) canonicalGraphArray(source []sanySemanticGraphNode) []tlc.SemanticNode {
	if source == nil {
		return nil
	}
	nodes := make([]tlc.SemanticNode, len(source))
	for i, node := range source {
		nodes[i] = b.canonicalGraph(node)
	}
	return nodes
}

func (b *tlcBridge) canonicalSymbol(source sanySemSymbol) *tlc.SymbolNode {
	switch node := b.canonicalGraph(source.(sanySemanticGraphNode)).(type) {
	case *tlc.SymbolNode:
		return node
	case *tlc.OpDefNode:
		return node.Symbol
	case *tlc.ThmOrAssumpDefNode:
		return node.Symbol
	default:
		panic(tlc.NewUnsupportedOperationException("runtime adapter for canonical context symbol"))
	}
}

func (b *tlcBridge) canonicalSubstitutions(source []*sanySemSubst) []tlc.Subst {
	if source == nil {
		return nil
	}
	key := tlcCanonicalSubstArray{first: unsafe.SliceData(source), length: len(source)}
	if nodes, found := b.canonicalSubstArrays[key]; found {
		return nodes
	}
	if b.canonicalSubstArrays == nil {
		b.canonicalSubstArrays = map[tlcCanonicalSubstArray][]tlc.Subst{}
	}
	if b.canonicalSubsts == nil {
		b.canonicalSubsts = map[*sanySemSubst]tlc.Subst{}
	}
	nodes := make([]tlc.Subst, len(source))
	b.canonicalSubstArrays[key] = nodes
	for i, subst := range source {
		node, found := b.canonicalSubsts[subst]
		if !found {
			node = tlc.NewSubst(nil, nil)
			b.canonicalSubsts[subst] = node
			node.Op = b.canonicalSymbol(subst.op)
			node.Expr = b.canonicalGraph(subst.expr)
		}
		nodes[i] = node
	}
	return nodes
}
