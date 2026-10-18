package tlago

import (
	"unsafe"

	"github.com/glycerine/tlago/tlc"
)

type tlcCanonicalSubstArray struct {
	first  **sanySemSubst
	length int
}

type sanyRuntimeSubst struct {
	source *sanySemSubst
	graph  func(sanySemanticGraphNode) tlc.SemanticNode
	symbol func(sanySemSymbol) *tlc.SymbolNode
}

type sanyRuntimeLabelTable struct {
	source *sanyLabelTable
	graph  func(sanySemanticGraphNode) tlc.SemanticNode
}

func sanyCanonicalLabelSource(label *tlc.LabelNode) (*sanySemLabelNode, bool) {
	if label == nil || label.SemanticNodeBase == nil {
		return nil, false
	}
	data, ok := label.CanonicalLevelData.(*sanyLevelData)
	if !ok || data == nil {
		return nil, false
	}
	source, ok := data.runtimeSource.(*sanySemLabelNode)
	return source, ok && source != nil
}

func (s *sanyRuntimeLabelTable) RuntimeLabelGet(name *tlc.UniqueString) (*tlc.LabelNode, bool) {
	if s == nil || s.source == nil || s.graph == nil {
		return nil, false
	}
	label := s.source.get(name.String())
	if label == nil {
		return nil, true
	}
	return s.graph(label).(*tlc.LabelNode), true
}

func (s *sanyRuntimeLabelTable) RuntimeLabelAdd(label *tlc.LabelNode) (bool, bool) {
	if s == nil || s.source == nil {
		return false, false
	}
	source, ok := sanyCanonicalLabelSource(label)
	if !ok {
		return false, false
	}
	return s.source.add(source), true
}

func (s *sanyRuntimeLabelTable) RuntimeLabelElements() ([]*tlc.LabelNode, bool) {
	if s == nil || s.source == nil || s.graph == nil {
		return nil, false
	}
	source := s.source.elements()
	labels := make([]*tlc.LabelNode, len(source))
	for i, label := range source {
		labels[i] = s.graph(label).(*tlc.LabelNode)
	}
	return labels, true
}

func sanyLabelTableFromPublic(table *tlc.SemanticLabelTable) (*sanyLabelTable, bool) {
	if table == nil {
		return nil, true
	}
	if runtime, ok := table.CanonicalRuntime().(*sanyRuntimeLabelTable); ok && runtime != nil && runtime.source != nil {
		return runtime.source, true
	}
	buckets, threshold := table.SnapshotBuckets()
	source := &sanyLabelTable{buckets: make([]*sanyLabelEntry, len(buckets)), entries: make(map[string]*sanyLabelEntry), threshold: threshold}
	for i, labels := range buckets {
		var tail **sanyLabelEntry = &source.buckets[i]
		for _, label := range labels {
			canonical, ok := sanyCanonicalLabelSource(label)
			if !ok {
				return nil, false
			}
			name := canonical.name
			entry := &sanyLabelEntry{name: name, hash: sanyLabelNameHash(name), value: canonical}
			*tail = entry
			tail = &entry.next
			source.entries[name] = entry
		}
	}
	return source, true
}

func (s *sanyRuntimeSubst) RuntimeSubstOp() (*tlc.SymbolNode, bool) {
	if s == nil || s.source == nil || s.symbol == nil {
		return nil, false
	}
	if s.source.op == nil {
		return nil, true
	}
	return s.symbol(s.source.op), true
}

func (s *sanyRuntimeSubst) RuntimeSetSubstOp(op *tlc.SymbolNode) bool {
	if s == nil || s.source == nil {
		return false
	}
	if op == nil {
		s.source.op = nil
		return true
	}
	if op.SemanticBase == nil {
		return false
	}
	data, ok := op.SemanticBase.CanonicalLevelData.(*sanyLevelData)
	if !ok || data == nil {
		return false
	}
	source, ok := data.runtimeSource.(*sanySemOpDeclNode)
	if !ok || source == nil {
		return false
	}
	s.source.op = source
	return true
}

func (s *sanyRuntimeSubst) RuntimeSubstExpr() (tlc.SemanticNode, bool) {
	if s == nil || s.source == nil || s.graph == nil {
		return nil, false
	}
	return s.graph(s.source.expr), true
}

func (s *sanyRuntimeSubst) RuntimeSetSubstExpr(expr tlc.SemanticNode, implicit bool) bool {
	if s == nil || s.source == nil {
		return false
	}
	if sanyExploreNull(expr) {
		s.source.expr, s.source.implicit = nil, implicit
		return true
	}
	carrier, ok := expr.(sanyRuntimeLevelDataCarrier)
	if !ok {
		return false
	}
	data, ok := carrier.GetCanonicalLevelData().(*sanyLevelData)
	if !ok || data == nil {
		return false
	}
	source, ok := data.runtimeSource.(sanySemanticGraphNode)
	if !ok || sanyExploreNull(source) {
		return false
	}
	s.source.expr, s.source.implicit = source, implicit
	return true
}

func (s *sanyRuntimeSubst) RuntimeSubstExprSyntax() (any, bool) {
	if s == nil || s.source == nil {
		return nil, false
	}
	return s.source.exprSTN, true
}

func (s *sanyRuntimeSubst) RuntimeSetSubstExprSyntax(syntax any) bool {
	if s == nil || s.source == nil {
		return false
	}
	if syntax == nil {
		s.source.exprSTN = nil
		return true
	}
	tree, ok := syntax.(*SanySyntaxNode)
	if !ok {
		return false
	}
	s.source.exprSTN = tree
	return true
}

func (s *sanyRuntimeSubst) RuntimeSubstImplicit() (bool, bool) {
	if s == nil || s.source == nil {
		return false, false
	}
	return s.source.implicit, true
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
					if source, ok := source.(*sanySemOpApplNode); ok {
						if source.operator != nil {
							node.Operator = b.canonicalSymbol(source.operator)
						}
						if source.subExpressionOf != nil {
							node.SubExpressionOf = b.canonicalSymbol(source.subExpressionOf)
						}
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
						node.IsAssumeProve = source.isAssumeProve
						node.GoalClause = source.goalClause
						if source.subExpressionOf != nil {
							node.SubExpressionOf = b.canonicalSymbol(source.subExpressionOf)
						}
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

func (b *tlcBridge) canonicalPublicSymbolTable(source *sanySymbolTable) *SanySymbolTable {
	if source == nil {
		return nil
	}
	if table := b.canonicalSymbolTables[source]; table != nil {
		return table
	}
	if b.canonicalSymbolTables == nil {
		b.canonicalSymbolTables = map[*sanySymbolTable]*SanySymbolTable{}
	}
	table := &SanySymbolTable{moduleTable: b.processor.ModuleTbl, errors: NewSanyErrors()}
	// Publish the shell before adapting contexts: their symbols refer back to
	// this same table while the canonical graph is being constructed.
	b.canonicalSymbolTables[source] = table
	stack := make([]*SanyContext, len(source.stack))
	for i, context := range source.stack {
		stack[i] = b.canonicalContext(context)
	}
	table.contextStack = stack
	if len(stack) != 0 {
		table.topContext, table.baseContext = stack[0], stack[len(stack)-1]
	}
	table.module = b.canonicalModuleOwner(source.module)
	return table
}

func (b *tlcBridge) bindCanonicalOpDefOrDeclTable(node tlc.SemanticNode, source *sanySymbolTable) {
	table := b.canonicalPublicSymbolTable(source)
	switch node := node.(type) {
	case *tlc.SymbolNode:
		node.SANYSymbolTable = table
	case *tlc.OpDefNode:
		node.SANYSymbolTable = table
		if node.Symbol != nil {
			node.Symbol.SANYSymbolTable = table
		}
	}
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
	remember := func(node tlc.SemanticNode) {
		b.canonicalGraphs[source] = node
		if levelNode, ok := source.(sanyCanonicalLevelNode); ok {
			bindSanyRuntimeLevelNode(levelNode.getLevelData(), levelNode, b.canonicalSymbol)
			levelNode.getLevelData().runtimeGraph = b.canonicalGraph
			levelNode.getLevelData().runtimeParameters = b.canonicalParameters
			levelNode.getLevelData().runtimeLabels = b.canonicalLabelTable
			levelNode.getLevelData().runtimeInstallLabels = b.installCanonicalLabelTable
		}
		if printer, ok := source.(interface{ levelDataToString() string }); ok {
			if owner, ok := source.(interface{ runtimeSemanticBase() *tlc.SemanticNodeBase }); ok {
				if data, ok := owner.runtimeSemanticBase().CanonicalLevelData.(*sanyLevelData); ok && data != nil {
					sourcePrinter := printer
					data.runtimeString = sourcePrinter.levelDataToString
				}
			}
		}
	}
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
				b.bindCanonicalOpDefOrDeclTable(node, source.table)
				remember(node)
				return node
			}
		}
		// Proof-local NEW declarations are distinct from same-named module
		// declarations. They retain their own canonical declaration identity.
		node := tlc.NewSymbolNode(source.semName())
		node.SemanticBase, node.Arity = source.SemanticNodeBase, source.semArity()
		node.Declaration = true
		node.OriginallyDefinedInModule = b.canonicalModuleOwner(source.module)
		node.Location, node.TreeNode = source.Location, source.TreeNode
		b.bindCanonicalOpDefOrDeclTable(node, source.table)
		if source.semKind() == sanyVariableDeclKind {
			node.MarkVariableDecl()
		} else if source.semKind() == sanyConstantDeclKind {
			node.Kind = tlc.SymbolConstantDecl
		}
		remember(node)
		return node
	case *sanySemOpDefNode:
		if node := b.canonicalDefinitions[source]; node != nil {
			b.bindCanonicalOpDefOrDeclTable(node, source.table)
			remember(node)
			return node
		}
		// Module definitions already have evaluator symbols used by native and
		// config overrides. Find the AST view by source identity, so importing
		// a LET Context cannot create a competing symbol for that definition.
		for definition, module := range b.definitionModules {
			if definition.semanticNode == source {
				node := b.convertSourceDefinitionAs(module+"!"+definition.Name, definition)
				b.bindCanonicalOpDefOrDeclTable(node, source.table)
				remember(node)
				return node
			}
		}
		if source.semKind() == sanyBuiltInKind {
			node := b.canonicalBuiltinDefinition(source)
			b.bindCanonicalOpDefOrDeclTable(node, source.table)
			remember(node)
			return node
		}
		params := b.canonicalParameters(source.formalNodes)
		symbol := tlc.NewSymbolNode(source.semName())
		node := tlc.NewOpDefNodeForSymbolWithBase(symbol, params, nil, source.SemanticNodeBase)
		node.SetNumberOfArgs(source.semArity())
		b.bindCanonicalOpDefOrDeclTable(node, source.table)
		transferCanonicalOpDefLevelMetadata(node, source)
		transferCanonicalOpDefRenderMetadata(node, source)
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
		node.Labels = b.canonicalLabelTable(source.labels)
		return node
	case *sanySemThmOrAssumpDefNode:
		node := &tlc.ThmOrAssumpDefNode{SemanticNodeBase: source.SemanticNodeBase,
			Name: tlc.UniqueStringOf(source.semName()), Params: b.canonicalParameters(source.formalNodes), Local: source.semLocal(),
			Theorem: source.theorem, Suffices: source.suffices, Arity: source.semArity(),
			IsLeibnizArg: source.leibniz, IsLeibnizValue: source.isLeibniz}
		transferCanonicalThmOrAssumpLevelMetadata(node, source)
		remember(node)
		node.Symbol = &tlc.SymbolNode{SemanticBase: source.SemanticNodeBase, Name: node.Name, Arity: source.semArity(), Data: node,
			Location: source.Location, TreeNode: source.TreeNode}
		if origin := source.getSource(); origin != source {
			node.SourceDefinition = b.canonicalGraph(origin).(*tlc.ThmOrAssumpDefNode)
		}
		node.Body = b.canonicalGraph(source.body)
		node.Proof = b.canonicalGraph(source.proof)
		node.OriginallyDefinedInModule = b.canonicalModuleOwner(source.module)
		node.InstantiatedFrom = b.canonicalModuleOwner(source.instantiatedFrom)
		node.Labels = b.canonicalLabelTable(source.labels)
		return node
	case *sanySemOpApplNode:
		node := &tlc.OpApplNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.Operator = b.canonicalSymbol(source.operator)
		if source.subExpressionOf != nil {
			node.SubExpressionOf = b.canonicalSymbol(source.subExpressionOf)
		}
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
		node.Name = tlc.UniqueStringOf(source.name)
		node.Arity = source.arity
		node.Module = b.canonicalModuleOwner(source.module)
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
		b.retainCanonicalLetChildren(node, source)
		node.Context = b.canonicalContext(source.context)
		node.Body = b.canonicalGraph(source.body)
		return node
	case *sanySemSubstInNode:
		node := &tlc.SubstInNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.Substs = b.canonicalSubstitutions(source.substs)
		node.Body = b.canonicalGraph(source.body)
		node.InstantiatingModule = b.canonicalModuleOwner(source.instantiatingModule)
		node.InstantiatedModule = b.canonicalModuleOwner(source.instantiatedModule)
		return node
	case *sanySemAPSubstInNode:
		node := &tlc.APSubstInNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.Substs = b.canonicalSubstitutions(source.substs)
		node.Body = b.canonicalGraph(source.body)
		node.InstantiatingModule = b.canonicalModuleOwner(source.instantiatingModule)
		node.InstantiatedModule = b.canonicalModuleOwner(source.instantiatedModule)
		return node
	case *sanySemLabelNode:
		node := &tlc.LabelNode{SemanticNodeBase: source.SemanticNodeBase,
			Name: tlc.UniqueStringOf(source.name), Arity: source.arity}
		remember(node)
		node.Body = b.canonicalGraph(source.body)
		node.Goal = b.canonicalGraph(source.goal)
		node.Params = b.canonicalParameters(source.formalNodes)
		node.Labels = b.canonicalLabelTable(source.labels)
		node.IsAssumeProve = source.isAssumeProve
		node.GoalClause = source.goalClause
		if source.subExpressionOf != nil {
			node.SubExpressionOf = b.canonicalSymbol(source.subExpressionOf)
		}
		return node
	case *sanySemAtNode:
		node := &tlc.AtNode{SemanticNodeBase: source.SemanticNodeBase}
		remember(node)
		node.ExceptRef = b.canonicalGraph(source.exceptRef).(*tlc.OpApplNode)
		node.ExceptComponentRef = b.canonicalGraph(source.exceptComponentRef).(*tlc.OpApplNode)
		return node
	case *sanySemAssumeProveNode:
		node := &tlc.AssumeProveNode{SemanticNodeBase: source.SemanticNodeBase,
			InScopeOfDecl: source.inScopeOfDecl, InProof: source.inProof, Suffices: source.suffices, IsBoxAssumeProve: source.isBoxAssumeProve}
		remember(node)
		node.Assumes = b.canonicalGraphArray(source.assumes)
		node.Prove = b.canonicalGraph(source.prove)
		if source.goal != nil {
			node.Goal = b.canonicalGraph(source.goal).(*tlc.ThmOrAssumpDefNode)
		}
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
		if source.assumeLevelChecked != 0 {
			node.Assume.(interface{ SetLevelParams(...*tlc.SymbolNode) }).SetLevelParams(b.canonicalLevelParameters(source.getLevelParams())...)
		}
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
		// A definition's public symbol table can expose an inner module before
		// installModuleTable creates the external module views. Java already has
		// one ModuleNode object at that point; publish a shell now and let module
		// installation fill this same identity later.
		node := tlc.NewModuleNodeWithBase(source.name, tlc.NewSemanticContext(nil), source.SemanticNodeBase)
		remember(node)
		return node

	default:
		panic(tlc.NewUnsupportedOperationException("runtime adapter for canonical context graph node"))
	}
}

func (b *tlcBridge) retainCanonicalLetChildren(node *tlc.LetInNode, source *sanySemLetInNode) {
	if source.opDefs == nil {
		node.OpDefs = nil
	} else {
		node.OpDefs = make([]tlc.SemanticNode, len(source.opDefs))
		for i, definition := range source.opDefs {
			if !sanyExploreNull(definition) {
				node.OpDefs[i] = b.canonicalGraph(definition.(sanySemanticGraphNode))
			}
		}
	}
	node.Instances = b.canonicalGraphArray(source.instances)
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

func (b *tlcBridge) canonicalLevelParameters(source *sanyLevelSet[sanySemSymbol]) []*tlc.SymbolNode {
	if source == nil {
		return nil
	}
	result := make([]*tlc.SymbolNode, 0, source.len())
	for symbol := range source.all() {
		result = append(result, b.canonicalSymbol(symbol))
	}
	return result
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
			node.BindRuntime(&sanyRuntimeSubst{source: subst, graph: b.canonicalGraph, symbol: b.canonicalSymbol})
			node.Op = b.canonicalSymbol(subst.op)
			node.Expr = b.canonicalGraph(subst.expr)
			node.ExprSyntax = subst.exprSTN
			node.Implicit = subst.implicit
		}
		nodes[i] = node
	}
	return nodes
}

func (b *tlcBridge) canonicalLabelTable(source *sanyLabelTable) *tlc.SemanticLabelTable {
	if source == nil {
		return nil
	}
	if table := b.canonicalLabelTables[source]; table != nil {
		return table
	}
	if b.canonicalLabelTables == nil {
		b.canonicalLabelTables = make(map[*sanyLabelTable]*tlc.SemanticLabelTable)
	}
	sourceBuckets, threshold := source.snapshotBuckets()
	buckets := make([][]*tlc.LabelNode, len(sourceBuckets))
	for i, labels := range sourceBuckets {
		for _, label := range labels {
			buckets[i] = append(buckets[i], b.canonicalGraph(label).(*tlc.LabelNode))
		}
	}
	table := tlc.NewSemanticLabelTableFromBuckets(buckets, threshold)
	b.installCanonicalLabelTable(source, table)
	return table
}

func (b *tlcBridge) installCanonicalLabelTable(source *sanyLabelTable, table *tlc.SemanticLabelTable) {
	if source == nil || table == nil {
		return
	}
	if b.canonicalLabelTables == nil {
		b.canonicalLabelTables = make(map[*sanyLabelTable]*tlc.SemanticLabelTable)
	}
	table.BindRuntime(&sanyRuntimeLabelTable{source: source, graph: b.canonicalGraph})
	b.canonicalLabelTables[source] = table
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
	for source, table := range b.canonicalSymbolTables {
		table.moduleTable = b.processor.ModuleTbl
		table.module = b.canonicalModuleOwner(source.module)
	}
	for source, node := range b.canonicalGraphs {
		switch source := source.(type) {
		case *sanySemOpDefNode:
			target := node.(*tlc.OpDefNode)
			target.OriginallyDefinedInModule = b.canonicalModuleOwner(source.module)
			b.bindCanonicalOpDefOrDeclTable(target, source.table)
		case *sanySemOpDeclNode:
			symbol := node.(*tlc.SymbolNode)
			symbol.Declaration = true
			symbol.OriginallyDefinedInModule = b.canonicalModuleOwner(source.module)
			b.bindCanonicalOpDefOrDeclTable(symbol, source.table)
		case *sanyFormalParamNode:
			node.(*tlc.SymbolNode).Module = b.canonicalModuleOwner(source.semanticModule)
		case *sanySemThmOrAssumpDefNode:
			node.(*tlc.ThmOrAssumpDefNode).OriginallyDefinedInModule = b.canonicalModuleOwner(source.module)
		case *sanySemTheoremNode:
			node.(*tlc.TheoremNode).Module = b.canonicalModuleOwner(source.module)
		case *sanySemAssumeNode:
			node.(*tlc.AssumeNode).Module = b.canonicalModuleOwner(source.module)
		case *sanySemInstanceNode:
			node.(*tlc.InstanceNode).Module = b.canonicalModuleOwner(source.module)
		case *sanySemSubstInNode:
			target := node.(*tlc.SubstInNode)
			target.InstantiatingModule = b.canonicalModuleOwner(source.instantiatingModule)
			target.InstantiatedModule = b.canonicalModuleOwner(source.instantiatedModule)
		case *sanySemAPSubstInNode:
			target := node.(*tlc.APSubstInNode)
			target.InstantiatingModule = b.canonicalModuleOwner(source.instantiatingModule)
			target.InstantiatedModule = b.canonicalModuleOwner(source.instantiatedModule)
		}
		levelNode, hasLevels := source.(sanyCanonicalLevelNode)
		levelTarget, storesLevels := node.(interface{ SetLevelParams(...*tlc.SymbolNode) })
		if hasLevels {
			data := levelNode.getLevelData()
			bindSanyRuntimeLevelNode(data, levelNode, b.canonicalSymbol)
			sourceLevelNode := levelNode
			data.runtimeString = func() string { return sanyLevelDataToString(sourceLevelNode) }
		}
		if hasLevels && storesLevels && *levelNode.getLevelData().levelChecked != 0 {
			levelTarget.SetLevelParams(b.canonicalLevelParameters(levelNode.getLevelParams())...)
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
	node.SetNumberOfArgs(source.semArity())
	transferCanonicalOpDefLevelMetadata(node, source)
	transferCanonicalOpDefRenderMetadata(node, source)
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

func transferCanonicalOpDefLevelMetadata(node *tlc.OpDefNode, source *sanySemOpDefNode) {
	levels := func(values []tlaLevel) []int {
		if values == nil {
			return nil
		}
		result := make([]int, len(values))
		for i, value := range values {
			result[i] = int(value)
		}
		return result
	}
	levelMatrix := func(values [][]tlaLevel) [][]int {
		if values == nil {
			return nil
		}
		result := make([][]int, len(values))
		for i, value := range values {
			result[i] = levels(value)
		}
		return result
	}
	node.ArgMaxLevels = levels(source.argMaxLevels)
	node.ArgWeights = source.argWeights
	node.MinMaxLevel = levelMatrix(source.minMaxLevel)
	node.OpLevelCond = source.opLevelCond
	node.IsLeibnizArg = source.leibniz
	node.IsLeibnizValue = source.isLeibniz
}

func transferCanonicalOpDefRenderMetadata(node *tlc.OpDefNode, source *sanySemOpDefNode) {
	node.LetInLevel = source.letInLevel
	node.InRecursiveSection = source.inRecursiveSection
	node.RecursiveSection = source.recursiveSection
	node.SymbolTablePresent = source.table != nil
	node.ParamsNull = source.formalNodes == nil
	if node.ParamsNull {
		node.Params = nil
	}
}

func transferCanonicalThmOrAssumpLevelMetadata(node *tlc.ThmOrAssumpDefNode, source *sanySemThmOrAssumpDefNode) {
	levels := func(values []tlaLevel) []int {
		if values == nil {
			return nil
		}
		result := make([]int, len(values))
		for i, value := range values {
			result[i] = int(value)
		}
		return result
	}
	levelMatrix := func(values [][]tlaLevel) [][]int {
		if values == nil {
			return nil
		}
		result := make([][]int, len(values))
		for i, value := range values {
			result[i] = levels(value)
		}
		return result
	}
	node.ArgMaxLevels = levels(source.argMaxLevels)
	node.ArgWeights = source.argWeights
	node.MinMaxLevel = levelMatrix(source.minMaxLevel)
	node.OpLevelCond = source.opLevelCond
}
