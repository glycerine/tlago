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
		// Publish before adapting the operator: its definition can refer back
		// to this application through recursive or mutually recursive bodies.
		b.canonicalGraphs[source] = node
		if owner, ok := source.(interface{ runtimeSemanticBase() *tlc.SemanticNodeBase }); ok {
			base := owner.runtimeSemanticBase()
			if source.Kind() == tlc.SemanticKindOf(node) {
				switch node := node.(type) {
				case *tlc.OpApplNode:
					node.SemanticNodeBase = base
					if source, ok := source.(*sanySemOpApplNode); ok && source.operator != nil {
						node.Operator = b.canonicalSymbol(source.operator)
					}
				case *tlc.OpArgNode:
					node.SemanticNodeBase = base
					if source, ok := source.(*sanySemOpArgNode); ok && source.operator != nil {
						node.Op = b.canonicalSymbol(source.operator)
					}
				case *tlc.LabelNode:
					node.SemanticNodeBase = base
					if source, ok := source.(*sanySemLabelNode); ok {
						node.Params = b.canonicalParameters(source.formalNodes)
					}
				case *tlc.AtNode:
					node.SemanticNodeBase = base
				case *tlc.SubstInNode:
					node.SemanticNodeBase = base
				case *tlc.APSubstInNode:
					node.SemanticNodeBase = base
				}
			}
		}
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
	if sanyExploreNull(source) {
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
				owned, _ := tlcBridgeOwnedDeclaration(module, source.semName())
				if owned != source {
					break
				}
				node := b.declarationSymbol(module, source.semName())
				remember(node)
				return node
			}
		}
		// Proof-local NEW declarations are distinct from same-named module
		// declarations. They retain their own canonical declaration identity.
		node := tlc.NewSymbolNode(source.semName())
		node.SemanticBase, node.Arity = source.SemanticNodeBase, source.semArity()
		node.Location, node.TreeNode = source.Location, source.TreeNode
		if source.semKind() == sanyVariableDeclKind {
			node.MarkVariableDecl()
		} else if source.semKind() == sanyConstantDeclKind {
			node.Kind = tlc.SymbolConstantDecl
		}
		remember(node)
		return node
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
			node := b.canonicalBuiltinDefinition(source)
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
		// Action.getDeclaration uses the first child of the syntax node's
		// one array, including the name of a named INSTANCE declaration.
		if tree, ok := source.TreeNode.(*SanySyntaxNode); ok && tree != nil && len(tree.One) > 0 {
			node.SetDeclarationLocation(b.sourceLocationForPosition(sanyNodePosition(tree.One[0])))
		}
		node.Local, node.InRecursive = source.semLocal(), source.inRecursive
		node.CompoundID = source.compoundID
		if origin := source.getSource(); origin != source {
			node.SourceDefinition = b.canonicalGraph(origin).(*tlc.OpDefNode)
		}
		node.Body = b.canonicalGraph(source.body)
		node.StepNode = b.canonicalGraph(source.stepNode)
		node.OriginallyDefinedInModule = b.canonicalModuleOwner(source.module)
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
		node.OriginallyDefinedInModule = b.canonicalModuleOwner(source.module)
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
		node.Params = b.canonicalParameters(source.formalNodes)
		return node
	case *sanySemAtNode:
		node := &tlc.AtNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		return node
	case *sanySemAssumeProveNode:
		node := &tlc.AssumeProveNode{SemanticNodeBase: source.SemanticNodeBase,
			InScopeOfDecl: source.inScopeOfDecl, InProof: source.inProof, Suffices: source.suffices, IsBoxAssumeProve: source.isBoxAssumeProve}
		remember(node)
		node.Assumes = b.canonicalGraphArray(source.assumes)
		node.Prove, node.Goal = b.canonicalGraph(source.prove), b.canonicalGraph(source.goal)
		return node
	case *sanySemNewSymbNode:
		node := &tlc.NewSymbNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.OpDecl = b.canonicalSymbol(source.opDeclNode)
		node.Set = b.canonicalGraph(source.set)
		return node
	case *sanySemAssumeNode:
		node := &tlc.AssumeNode{SemanticNodeBase: source.SemanticNodeBase, Module: b.canonicalModuleOwner(source.module), IsAxiom: source.isAxiom}
		remember(node)
		node.Assume = b.canonicalGraph(source.assumeExpr)
		if source.def != nil {
			node.Def = b.canonicalGraph(source.def).(*tlc.ThmOrAssumpDefNode)
		}
		return node
	case *sanySemTheoremNode:
		node := &tlc.TheoremNode{SemanticNodeBase: source.SemanticNodeBase, Module: b.canonicalModuleOwner(source.module), Suffices: source.suffices}
		remember(node)
		if b.canonicalTheorems == nil {
			b.canonicalTheorems = map[*sanySemTheoremNode]*tlc.TheoremNode{}
		}
		b.canonicalTheorems[source] = node
		node.Theorem = b.canonicalGraph(source.theoremExprOrAssumeProve)
		if source.def != nil {
			node.Def = b.canonicalGraph(source.def).(*tlc.ThmOrAssumpDefNode)
		}
		node.Proof = b.canonicalGraph(source.proof)
		return node
	case *sanySemLeafProofNode:
		node := &tlc.LeafProofNode{SemanticNodeBase: source.SemanticNodeBase, Omitted: source.omitted, Only: source.isOnly}
		remember(node)
		node.Facts, node.Defs = b.canonicalGraphArray(source.facts), b.canonicalSymbols(source.defs)
		return node
	case *sanySemNonLeafProofNode:
		node := &tlc.NonLeafProofNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.Steps, node.Instances = b.canonicalGraphArray(source.steps), b.canonicalGraphArray(source.insts)
		node.Context = b.canonicalContext(source.context)
		return node
	case *sanySemDefStepNode:
		node := &tlc.DefStepNode{SemanticNodeBase: source.SemanticNodeBase, StepNumber: source.stepNumber}
		remember(node)
		if source.defs != nil {
			node.Defs = make([]*tlc.OpDefNode, len(source.defs))
			for i, def := range source.defs {
				if def != nil {
					node.Defs[i] = b.canonicalGraph(def).(*tlc.OpDefNode)
				}
			}
		}
		return node
	case *sanySemUseOrHideNode:
		node := &tlc.UseOrHideNode{SemanticNodeBase: source.SemanticNodeBase, Only: source.isOnly, StepName: source.stepName}
		remember(node)
		node.Facts, node.Defs = b.canonicalGraphArray(source.facts), b.canonicalSymbols(source.defs)
		return node
	case *sanySemInstanceNode:
		node := &tlc.InstanceNode{SemanticNodeBase: source.SemanticNodeBase, Name: source.name, StepName: source.stepName,
			Local: source.local, Module: b.canonicalModuleOwner(source.module)}
		remember(node)
		node.Params, node.Substs = b.canonicalParameters(source.params), b.canonicalSubstitutions(source.substs)
		return node
	case *sanySemModuleNode:
		if node := b.canonicalModuleOwner(source); node != nil {
			remember(node)
			return node
		}
		panic(tlc.NewUnsupportedOperationException("runtime module adapter has not been installed"))

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

func (b *tlcBridge) canonicalSymbols(source []sanySemSymbol) []*tlc.SymbolNode {
	if source == nil {
		return nil
	}
	nodes := make([]*tlc.SymbolNode, len(source))
	for i, symbol := range source {
		if !sanyExploreNull(symbol) {
			nodes[i] = b.canonicalSymbol(symbol)
		}
	}
	return nodes
}

func (b *tlcBridge) canonicalModuleOwner(source *sanySemModuleNode) *tlc.ModuleNode {
	if source == nil {
		return nil
	}
	for module, node := range b.moduleNodes {
		if module.semanticNode == source {
			return node
		}
	}
	return nil
}

// Definitions can be adapted before runtime module shells are installed.
func (b *tlcBridge) bindCanonicalGraphModules() {
	for _, context := range b.canonicalContexts {
		context.ModuleTable = b.processor.ModuleTbl
	}
	for source, node := range b.canonicalGraphs {
		switch source := source.(type) {
		case *sanySemOpDefNode:
			node.(*tlc.OpDefNode).OriginallyDefinedInModule = b.canonicalModuleOwner(source.module)
		case *sanySemThmOrAssumpDefNode:
			node.(*tlc.ThmOrAssumpDefNode).OriginallyDefinedInModule = b.canonicalModuleOwner(source.module)
		case *sanySemTheoremNode:
			node.(*tlc.TheoremNode).Module = b.canonicalModuleOwner(source.module)
		case *sanySemAssumeNode:
			node.(*tlc.AssumeNode).Module = b.canonicalModuleOwner(source.module)
		case *sanySemInstanceNode:
			node.(*tlc.InstanceNode).Module = b.canonicalModuleOwner(source.module)
		}
	}
}

// Builtins are actual SANY OpDefNodes, including their phony formal parameters.
// Cache by source identity: a later frontend reInit creates different symbols
// with the same names while existing running tools retain the earlier context.
func (b *tlcBridge) canonicalBuiltinDefinition(source *sanySemOpDefNode) *tlc.OpDefNode {
	if node := b.canonicalDefinitions[source]; node != nil {
		return node
	}
	symbol := tlc.NewSymbolNode(source.semName())
	node := tlc.NewOpDefNodeForSymbolWithBase(symbol, nil, nil, source.SemanticNodeBase)
	symbol.Kind, symbol.Arity = tlc.SymbolBuiltIn, source.semArity()
	if b.canonicalDefinitions == nil {
		b.canonicalDefinitions = map[*sanySemOpDefNode]*tlc.OpDefNode{}
	}
	if b.canonicalGraphs == nil {
		b.canonicalGraphs = map[sanySemanticGraphNode]tlc.SemanticNode{}
	}
	b.canonicalDefinitions[source], b.canonicalGraphs[source] = node, node
	node.Params = b.canonicalParameters(source.formalNodes)
	return node
}
