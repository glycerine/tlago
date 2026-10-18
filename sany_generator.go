package tlago

import "github.com/glycerine/tlago/tlc"

// SanyGenerator adapts tla2sany.semantic.Generator's public module-generation
// boundary over the shared Go semantic engine.
type SanyGenerator struct {
	moduleTable *SanyExternalModuleTable
	errors      *SanyErrors
	symbolTable *SanySymbolTable
	sourceNodes map[tlc.SemanticNode]sanySemanticGraphNode
	// Java retains every semantic view reachable through its one growing
	// SymbolTable. Seed each short-lived bridge from the same canonical views.
	runtimeNodes        map[sanySemanticGraphNode]tlc.SemanticNode
	runtimeContexts     map[*sanyContext]*tlc.SemanticContext
	runtimeSymbolTables map[*sanySymbolTable]*SanySymbolTable
	runtimeModules      map[*Module]*tlc.ModuleNode
	nodes               *sanyGeneratorNodes
}

func NewSanyGenerator(moduleTable *SanyExternalModuleTable, errors *SanyErrors) *SanyGenerator {
	// Generator constructs its four error-recovery sentinels before its
	// SymbolTable. Besides being reused by every generate call, these nodes
	// consume public semantic UIDs in this exact constructor order.
	nodes := newSanyGeneratorNodes()
	symbolTable := NewSanySymbolTable(moduleTable, errors)
	generator := &SanyGenerator{
		moduleTable:         moduleTable,
		errors:              errors,
		symbolTable:         symbolTable,
		sourceNodes:         map[tlc.SemanticNode]sanySemanticGraphNode{},
		runtimeNodes:        map[sanySemanticGraphNode]tlc.SemanticNode{},
		runtimeContexts:     map[*sanyContext]*tlc.SemanticContext{},
		runtimeSymbolTables: map[*sanySymbolTable]*SanySymbolTable{},
		runtimeModules:      map[*Module]*tlc.ModuleNode{},
		nodes:               nodes,
	}
	initial := sanyGlobalInitialContext(false)
	for _, binding := range symbolTable.GetContext().SnapshotPairHistory() {
		if binding.Key.Name == nil {
			continue
		}
		if source, ok := initial.getSymbol(binding.Key.Name.String()).(sanySemanticGraphNode); ok {
			generator.sourceNodes[binding.Node] = source
		}
	}
	return generator
}

func (g *SanyGenerator) require() *SanyGenerator {
	if g == nil {
		panic(tlc.NewNullPointerException())
	}
	return g
}

func (g *SanyGenerator) GetSymbolTable() *SanySymbolTable {
	return g.require().symbolTable
}

func sanyGeneratorSourceNode(node tlc.SemanticNode) sanySemanticGraphNode {
	if node == nil {
		return nil
	}
	if module, ok := node.(*tlc.ModuleNode); ok {
		if source, ok := module.SANYSourceModule.(*Module); ok && source != nil {
			return source.semanticNode
		}
		return nil
	}
	data, ok := node.(interface{ GetCanonicalLevelData() any })
	if !ok {
		return nil
	}
	levels, ok := data.GetCanonicalLevelData().(*sanyLevelData)
	if !ok || levels == nil || levels.runtimeSource == nil {
		return nil
	}
	source, _ := levels.runtimeSource.(sanySemanticGraphNode)
	return source
}

// Recreate the private generator context from the live public context. Pair
// history is newest-first publicly, while add appends oldest-to-newest.
func (g *SanyGenerator) sourceContext(context *SanyContext) *sanyContext {
	if context == nil {
		return nil
	}
	result := newSanyContext()
	history := context.SnapshotPairHistory()
	for index := len(history) - 1; index >= 0; index-- {
		binding := history[index]
		source := sanyGeneratorSourceNode(binding.Node)
		if source == nil {
			source = g.sourceNodes[binding.Node]
		}
		if source == nil {
			panic(tlc.NewUnsupportedOperationException("Generator context symbol has no canonical SANY source"))
		}
		symbol, ok := source.(sanySemSymbol)
		if !ok {
			panic(tlc.NewClassCastException("Generator context entry is not a SymbolNode"))
		}
		result.add(sanyContextKey{module: binding.Key.Module, name: binding.Key.Name.String()}, symbol)
	}
	return result
}

func (g *SanyGenerator) sourceModules() (map[string]*Module, *sanyExternalModuleTable) {
	modules := map[string]*Module{}
	table := newSanyExternalModuleTable()
	seen := map[*Module]bool{}
	retain := func(runtimeModule *tlc.ModuleNode, external bool) {
		if runtimeModule == nil {
			return
		}
		source, ok := runtimeModule.SANYSourceModule.(*Module)
		if !ok || source == nil || source.semanticNode == nil || seen[source] {
			return
		}
		seen[source] = true
		// ExternalModuleTable stores the public Context object, and callers can
		// mutate it between Generator instances. Nested modules retained in this
		// Generator's live Context have the same source-backed mutation boundary.
		// Refresh either view before resolving EXTENDS or INSTANCE.
		oldContext := source.semanticNode.context
		liveContext := g.sourceContext(runtimeModule.Context)
		source.semanticNode.context = liveContext
		if source.symbolTable != nil {
			for index, context := range source.symbolTable.stack {
				if context == oldContext {
					source.symbolTable.stack[index] = liveContext
				}
			}
		}
		modules[source.Name] = source
		if external {
			table.put(source.Name, liveContext, source.semanticNode)
		}
	}
	if g.moduleTable != nil {
		for _, runtimeModule := range g.moduleTable.GetModuleNodes() {
			retain(runtimeModule, true)
		}
	}
	// Inner modules live in SymbolTable's distinct module-name namespace and
	// are deliberately absent from ExternalModuleTable. A later generate call
	// can nevertheless EXTEND or INSTANCE them through that shared table.
	for _, binding := range g.symbolTable.GetContext().SnapshotPairHistory() {
		if binding.Key.Module {
			if runtimeModule, ok := binding.Node.(*tlc.ModuleNode); ok {
				retain(runtimeModule, false)
			}
		}
	}
	return modules, table
}

func sanyGeneratorSeedRuntimeGraph(bridge *tlcBridge, source *Module, runtimeModule *tlc.ModuleNode) {
	if source == nil || source.semanticNode == nil || runtimeModule == nil {
		return
	}
	bridge.moduleNodes[source] = runtimeModule
	bridge.canonicalContexts[source.semanticNode.context] = runtimeModule.Context
	visited := map[tlc.SemanticNode]bool{}
	var retain func(tlc.SemanticNode)
	retain = func(node tlc.SemanticNode) {
		if node == nil || visited[node] {
			return
		}
		visited[node] = true
		canonical := sanyGeneratorSourceNode(node)
		if canonical != nil {
			bridge.canonicalGraphs[canonical] = node
			switch node := node.(type) {
			case *tlc.OpDefNode:
				if definition, ok := canonical.(*sanySemOpDefNode); ok {
					bridge.canonicalDefinitions[definition] = node
					for index, formal := range definition.formalNodes {
						if index < len(node.Params) {
							bridge.canonicalFormals[formal] = node.Params[index]
							bridge.canonicalGraphs[formal] = node.Params[index]
						}
					}
				}
			case *tlc.ThmOrAssumpDefNode:
				if definition, ok := canonical.(*sanySemThmOrAssumpDefNode); ok {
					for index, formal := range definition.formalNodes {
						if index < len(node.Params) {
							bridge.canonicalFormals[formal] = node.Params[index]
							bridge.canonicalGraphs[formal] = node.Params[index]
						}
					}
				}
			}
		}
		for _, child := range tlc.SemanticListOfChildren(node) {
			retain(child)
		}
	}
	retain(runtimeModule)
	for _, binding := range runtimeModule.Context.SnapshotPairHistory() {
		retain(binding.Node)
	}
}

func (g *SanyGenerator) appendDiagnostics(diags Diagnostics) {
	if len(diags) == 0 {
		return
	}
	if g.errors == nil {
		panic(tlc.NewNullPointerException())
	}
	retained := append(Diagnostics(nil), diags...)
	for index := range retained {
		if retained[index].sanyDetails == nil {
			retained[index].sanyDetails = sanyErrorDetailsForDiagnostic(retained[index])
		}
	}
	g.errors.diagnostics = appendSanyDiagnostics(g.errors.diagnostics, retained...)
}

// Generate returns nil for a non-module syntax node and otherwise creates the
// semantic module without running level checking, matching Generator.generate.
func (g *SanyGenerator) Generate(tree *SanySyntaxNode) *tlc.ModuleNode {
	g.require()
	if tree == nil {
		panic(tlc.NewNullPointerException())
	}
	if tree.Kind.JavaName() != "N_Module" {
		return nil
	}

	module, translated := sanyModuleFromSyntax(tree.FileName, tree)
	module.generatorNodes = g.nodes
	g.appendDiagnostics(translated)
	modules, sourceTable := g.sourceModules()
	modules[module.Name] = module
	spec := &Spec{
		initialContext:  g.sourceContext(g.symbolTable.GetContext()),
		semanticModules: sourceTable,
		Root:            module,
		Modules:         modules,
	}
	checks := &sanyModuleLevelChecks{symbolTable: newSanySymbolTable(spec.initialContext, sourceTable)}
	var diagnostics Diagnostics
	func() {
		defer func() {
			if failure := recover(); failure != nil {
				g.appendDiagnostics(spec.SemanticDiags)
				panic(failure)
			}
		}()
		diagnostics = generateModuleWithEnclosing(module, spec, nil, checks)
	}()
	g.appendDiagnostics(diagnostics)

	defns := tlc.NewDefns()
	processor := tlc.NewSpecProcessorWithSpecObj(module.Name, defns, tlc.NewModelConfig(module.Name), spec)
	processor.ModuleTbl = g.moduleTable
	bridge := &tlcBridge{
		processor:             processor,
		spec:                  spec,
		defns:                 defns,
		symbols:               map[string]*tlc.SymbolNode{},
		moduleDefinitionNames: moduleDefinitionNameIndex(spec),
		definitionModules:     map[*Definition]string{},
		sourceSymbols:         map[*Definition]*tlc.SymbolNode{},
		sourceDefinitions:     map[*Definition]*tlc.OpDefNode{},
		canonicalDefinitions:  map[*sanySemOpDefNode]*tlc.OpDefNode{},
		canonicalFormals:      map[*sanyFormalParamNode]*tlc.SymbolNode{},
		canonicalGraphs:       map[sanySemanticGraphNode]tlc.SemanticNode{},
		canonicalContexts:     map[*sanyContext]*tlc.SemanticContext{},
		canonicalSymbolTables: map[*sanySymbolTable]*SanySymbolTable{},
		moduleNodes:           map[*Module]*tlc.ModuleNode{},
		assumptionModules:     map[*Module]bool{},
		theoremModules:        map[*Module]bool{},
		lazyModuleAccessors:   true,
	}
	for source, node := range g.runtimeNodes {
		bridge.canonicalGraphs[source] = node
		switch source := source.(type) {
		case *sanySemOpDefNode:
			bridge.canonicalDefinitions[source] = node.(*tlc.OpDefNode)
		case *sanyFormalParamNode:
			bridge.canonicalFormals[source] = node.(*tlc.SymbolNode)
		}
	}
	for source, context := range g.runtimeContexts {
		bridge.canonicalContexts[source] = context
	}
	for source, table := range g.runtimeSymbolTables {
		bridge.canonicalSymbolTables[source] = table
	}
	for source, runtimeModule := range g.runtimeModules {
		bridge.moduleNodes[source] = runtimeModule
	}
	if g.moduleTable == nil {
		processor.ModuleTbl = tlc.NewExternalModuleTable()
	}
	for _, runtimeModule := range processor.ModuleTbl.GetModuleNodes() {
		if source, ok := runtimeModule.SANYSourceModule.(*Module); ok {
			sanyGeneratorSeedRuntimeGraph(bridge, source, runtimeModule)
			bridge.assumptionModules[source] = true
			bridge.theoremModules[source] = true
		}
	}
	// The root module and its definitions must retain the constructor-owned
	// public SymbolTable and Context identities.
	bridge.canonicalContexts[module.semanticNode.context] = g.symbolTable.GetContext()
	bridge.canonicalSymbolTables[module.symbolTable] = g.symbolTable
	bridge.extendModuleTable(false)
	bridge.bindCanonicalGraphModules()
	bridge.installModuleAssumptions(false)

	result := bridge.moduleNodes[module]
	for source, node := range bridge.canonicalGraphs {
		g.runtimeNodes[source] = node
		g.sourceNodes[node] = source
	}
	for source, context := range bridge.canonicalContexts {
		g.runtimeContexts[source] = context
	}
	for source, table := range bridge.canonicalSymbolTables {
		g.runtimeSymbolTables[source] = table
	}
	for source, runtimeModule := range bridge.moduleNodes {
		g.runtimeModules[source] = runtimeModule
	}
	result.SANYSourceModule = module
	result.Context.ModuleTable = g.moduleTable
	g.symbolTable.moduleTable = g.moduleTable
	g.symbolTable.SetModuleNode(result)
	return result
}
