// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.

package tlago

import (
	"maps"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// Keep Integers' override available when a debugger expression first loads it
// after the running spec was compiled with just Naturals.
func (b *tlcBridge) retainIntegerNativeOverride() {
	tool := *b.tool
	tool.Definitions = maps.Clone(tool.Definitions)
	tool.DefnsByName = maps.Clone(tool.DefnsByName)
	tool.InstallIntegerDefinitions()
	for _, name := range []string{"GEQ", "\\geq"} {
		b.nativeDefinitions[tlc.UniqueStringOf("Integers!"+name)] = tool.DefnsByName[tlc.UniqueStringOf(name)]
	}
}

func (b *tlcBridge) builtinDefinition(name string) *tlc.OpDefNode {
	if b.builtinDefinitions == nil {
		b.builtinDefinitions = map[string]*tlc.OpDefNode{}
	}
	name = tlcSymbolName(name)
	if def := b.builtinDefinitions[name]; def != nil {
		return def
	}
	def := tlc.NewOpDefNodeForSymbol(b.symbol(name), nil, nil)
	def.KindValue = tlc.SemanticBuiltInKind
	def.Symbol.Kind = tlc.SymbolBuiltIn
	if info, found := sanyBuiltinOperatorInfo(name); found {
		def.Symbol.Arity = info.arity
		def.SetLevel(int(info.level))
	}
	b.builtinDefinitions[name] = def
	return def
}

func (b *tlcBridge) builtinNode(op *tlc.UniqueString, args ...tlc.SemanticNode) *tlc.OpApplNode {
	return tlc.NewOpApplNode(b.builtinDefinition(op.String()).Symbol, args...)
}

// SpecProcessor.processModuleOverrides visits each module's actual OpDefs.
// Inherited operators therefore receive the overriding module's methods too,
// with the global unqualified entry following the same dependency order.
func (b *tlcBridge) installModuleNativeOverrides() {
	for _, module := range b.processor.ModuleTbl.GetModuleNodes() {
		for _, def := range module.GetOpDefs() {
			moduleName, name := module.Name.String(), def.Name.String()
			native := bridgeNativeOverrideModuleMembers[moduleName][name]
			if moduleName == "Integers" {
				native = native || bridgeNativeOverrideModuleMembers["Naturals"][name]
			}
			if !native {
				continue
			}
			value := b.nativeDefinitions[tlc.UniqueStringOf(module.Name.String()+"!"+def.Name.String())]
			if value == nil {
				value = b.nativeDefinitions[def.Name]
			}
			if value == nil || def.Body == nil || !bridgeNativeMethodArityMatches(moduleName, value, def.Arity()) {
				continue
			}
			value = tlc.WithEvaluatingOpDef(value, def)
			if body, ok := def.Body.(interface{ SetToolObject(any) }); ok {
				body.SetToolObject(value)
			}
			b.defns.Put(def.Name, value)
			b.tool.Define(def.Symbol, value)
			b.tool.DefnsByName[def.Name] = value
		}
	}
	b.processor.ProcessModuleOverrides(b.tool, tlc.NewNativeClassLoader("tlc2.module", b.spec.FilenameResolver))
}

func bridgeNativeMethodArityMatches(module string, value any, arity int) bool {
	// The legacy built-in module classes install their public static methods
	// without the annotated override's method-parameter-count check.
	switch module {
	case "Naturals", "Integers", "Sequences", "FiniteSets", "Bags", "TLC", "Randomization":
		return true
	}
	if method, ok := value.(*tlc.MethodValue); ok && method.ParameterCount >= 0 {
		return method.ParameterCount == arity
	}
	// @Evaluation methods receive unevaluated arguments and use a different
	// signature; their source registration does not apply this arity check.
	return true
}

// Retain the external dependency order and full represented semantic contexts.
// Internal modules have graph identities too, but never become external table
// entries merely because the loader indexes them by name.
func (b *tlcBridge) installModuleTable() {
	b.processor.ModuleTbl = tlc.NewExternalModuleTable()
	b.moduleNodes = map[*Module]*tlc.ModuleNode{}
	b.localModuleDefinitions = map[string]*tlc.OpDefNode{}
	b.extendModuleTable(true)
}

// Add newly compiled modules without replacing the live modules, contexts,
// source symbols or previously applied configuration/native bindings.
func (b *tlcBridge) extendModuleTable(publishRoot bool) {
	table := b.processor.ModuleTbl
	nodes := b.moduleNodes
	added := map[*Module]bool{}
	var modules []*Module
	seen := map[*Module]bool{}
	var visit func(*Module)
	visit = func(mod *Module) {
		if mod == nil || seen[mod] {
			return
		}
		seen[mod] = true
		for _, dependency := range moduleSemanticImports(mod) {
			visit(b.spec.Modules[dependency])
		}
		for _, nested := range mod.Nested {
			visit(nested)
		}
		modules = append(modules, mod)
	}
	for _, name := range b.spec.SemanticOrder {
		visit(b.spec.Modules[name])
	}
	visit(b.spec.Root)

	global := tlc.NewSemanticContext(nil)
	for _, builtin := range sanyBuiltinOperators {
		global.AddSymbolToContext(tlc.SemanticContextKey{Name: tlc.UniqueStringOf(builtin.name)}, b.builtinDefinition(builtin.name))
	}
	outer := enclosingModules(b.spec)
	for _, mod := range modules {
		if nodes[mod] != nil {
			continue
		}
		added[mod] = true
		context := tlc.NewSemanticContext(table)
		if outer[mod] == nil {
			context = global.Duplicate(table)
		}
		node := tlc.NewModuleNode(mod.Name, context)
		position := mod.Pos
		position.File = mod.Name
		node.SetSourceLocation(b.sourceLocationForPosition(position))
		node.SetStandard(outer[mod] == nil && mod.Library)
		nodes[mod] = node
	}
	for definition, node := range b.sourceDefinitions {
		node.OriginallyDefinedInModule = nodes[b.spec.Modules[b.definitionModules[definition]]]
		node.Local = definition.Local
	}
	for _, mod := range modules {
		if !added[mod] {
			continue
		}
		node := nodes[mod]
		for _, ext := range mod.Extends {
			node.Extendees = append(node.Extendees, nodes[b.spec.Modules[ext]])
		}
		for _, entry := range tlcBridgeContextEntries(b.spec, mod, map[*Module]bool{}) {
			if entry.initial {
				continue // Already duplicated, including the original Pair links.
			}
			key := tlc.SemanticContextKey{Name: tlc.UniqueStringOf(entry.name), Module: entry.moduleKey}
			var symbol tlc.SemanticNode
			switch {
			case entry.moduleKey:
				symbol = nodes[entry.module]
			case entry.kind == VariableDecl || entry.kind == ConstantDecl:
				declaration := b.declarationSymbol(entry.module, entry.name)
				if entry.kind == VariableDecl {
					declaration.MarkVariableDecl()
				} else {
					declaration.Kind = tlc.SymbolConstantDecl
					for _, decl := range entry.module.Declarations {
						if decl.Kind == ConstantDecl {
							if arity, exists := decl.Arities[entry.name]; exists {
								declaration.Arity = arity
							}
						}
					}
				}
				position := entry.position
				position.File = entry.module.Name
				declaration.Location = b.sourceLocationForPosition(position)
				symbol = declaration
			case entry.kind == OperatorDecl:
				if def := b.moduleContextDefinition(mod, entry); def != nil {
					symbol = def
				}
			default:
				// Named theorem/assumption definitions share their evaluated
				// source body with the represented runtime operator.
				if def := b.moduleContextDefinition(mod, entry); def != nil {
					symbol = tlc.NewThmOrAssumpDefNode(entry.name, def.Body, def.Params...)
				}
			}
			if symbol != nil {
				node.Context.AddSymbolToContext(key, symbol)
			}
		}
		if outer[mod] == nil && (publishRoot || mod != b.spec.Root) {
			table.Put(node.Name, node.Context, node)
		}
	}
	for _, mod := range modules {
		markInstance := func(name string) {
			if instancee := nodes[b.spec.Modules[name]]; instancee != nil {
				instancee.SetInstantiated(true)
			}
		}
		for _, instance := range mod.Instances {
			moduleImportsFromInstance(instance, markInstance)
		}
		for _, definition := range mod.Definitions {
			moduleImportsFromExpr(definition.Expr, markInstance)
		}
		for _, assumption := range mod.Assumptions {
			moduleImportsFromExpr(assumption.Expr, markInstance)
		}
		for _, theorem := range mod.Theorems {
			moduleImportsFromExpr(theorem.Expr, markInstance)
		}
	}
	// Some hidden source definitions are converted while contexts are built.
	for definition, node := range b.sourceDefinitions {
		node.OriginallyDefinedInModule = nodes[b.spec.Modules[b.definitionModules[definition]]]
		node.Local = definition.Local
	}
	for _, binding := range b.instanceDefinitions {
		for definition, clone := range binding.defs {
			source := b.sourceDefinitions[definition]
			if source == nil || clone == source {
				continue
			}
			clone.SourceDefinition = source.GetSource()
			clone.Local = binding.inst.Local
			if binding.inst.Name == "" && source.OriginallyDefinedInModule != nil && source.OriginallyDefinedInModule.IsParameterFree() {
				clone.OriginallyDefinedInModule = source.OriginallyDefinedInModule
			} else {
				clone.OriginallyDefinedInModule = nodes[binding.owner]
			}
		}
	}
	if publishRoot {
		table.SetRootModule(nodes[b.spec.Root])
		b.processor.RootModule = table.GetRootModule()
	}
	// Publish fully populated cached accessors before worker threads use them.
	for _, mod := range modules {
		node := nodes[mod]
		node.GetConstantDecls()
		node.GetVariableDecls()
		node.GetOpDefs()
		node.GetThmOrAssDefs()
		node.GetInnerModules()
	}
}

func (b *tlcBridge) moduleContextDefinition(mod *Module, entry tlcBridgeContextEntry) *tlc.OpDefNode {
	if entry.instance == nil && entry.module != nil {
		// A source module's qualified lookup key can also be a named INSTANCE
		// export in the root. Its context must still retain the original node.
		for i := range entry.module.Definitions {
			def := &entry.module.Definitions[i]
			if def.Name == entry.name {
				return b.convertSourceDefinitionAs(entry.module.Name+"!"+def.Name, def)
			}
		}
	}
	if entry.instance != nil && entry.instance.Name == "" && entry.instance.Local {
		instancee := b.moduleNodes[b.spec.Modules[entry.instance.Module]]
		if instancee != nil {
			if source, ok := instancee.Context.GetSymbol(tlc.SemanticContextKey{Name: tlc.UniqueStringOf(entry.instanceSourceName)}).(*tlc.OpDefNode); ok && source != nil && source.OriginallyDefinedInModule != nil && source.OriginallyDefinedInModule.IsParameterFree() {
				key := entry.instanceOwner.Name + "!" + entry.name
				if clone := b.localModuleDefinitions[key]; clone != nil {
					return clone
				}
				clone := tlc.NewOpDefNodeForSymbol(tlc.NewSymbolNode(key), source.Params, source.Body)
				clone.Name = source.Name
				clone.Local = true
				clone.OriginallyDefinedInModule = source.OriginallyDefinedInModule
				clone.SourceDefinition = source.GetSource()
				b.withPositionLocation(entry.instance.SourcePosition(), clone)
				b.localModuleDefinitions[key] = clone
				b.define(clone.Symbol, clone)
				// Bind references in this module's own definitions to the LOCAL
				// instance symbol. Shared imported bodies retain their own context.
				var roots []tlc.SemanticNode
				for i := range entry.instanceOwner.Definitions {
					if node := b.sourceDefinitions[&entry.instanceOwner.Definitions[i]]; node != nil {
						roots = append(roots, node)
					}
				}
				tlc.SemanticSubstituteFor(roots, clone, source)
				return clone
			}
		}
	}
	key := mod.Name + "!" + entry.name
	if mod == b.spec.Root {
		key = entry.name
	}
	if def := b.defs[key]; def != nil {
		binding := b.instanceDefinitions[key]
		if entry.instance == nil || binding != nil && binding.owner == entry.instanceOwner && binding.inst.SourcePosition() == entry.instance.SourcePosition() {
			return b.convertDefinitionAs(key, def)
		}
	}
	if entry.instance != nil && entry.instance.Name != "" {
		// Native definitions inherited by an instancee are semantic OpDefs too,
		// even when the runtime export index contains only their native values.
		instancee := b.moduleNodes[b.spec.Modules[entry.instance.Module]]
		if instancee != nil {
			if source, ok := instancee.Context.GetSymbol(tlc.SemanticContextKey{Name: tlc.UniqueStringOf(entry.instanceSourceName)}).(*tlc.OpDefNode); ok && source != nil {
				cacheKey := entry.instanceOwner.Name + "!" + entry.name
				if clone := b.localModuleDefinitions[cacheKey]; clone != nil {
					return clone
				}
				binding := b.instanceBinding(entry.instanceOwner, *entry.instance)
				b.prepareInstanceBinding(binding)
				params := append(append([]*tlc.SymbolNode(nil), binding.params...), source.Params...)
				body := source.Body
				if len(binding.substs) > 0 {
					body = b.withPositionLocation(entry.instance.SourcePosition(), tlc.NewSubstInNode(body, binding.substs...))
				}
				// The enclosing module's lookup key may also name an inner
				// INSTANCE export. This clone owns the full name in its context.
				symbol := tlc.NewSymbolNode(entry.name)
				clone := tlc.NewOpDefNodeForSymbol(symbol, params, body)
				clone.Local = entry.instance.Local
				clone.SourceDefinition = source.GetSource()
				clone.OriginallyDefinedInModule = b.moduleNodes[entry.instanceOwner]
				if len(binding.substs) > 0 {
					clone.CompoundID = append([]*tlc.UniqueString{tlc.UniqueStringOf(entry.instance.Name)}, source.GetCompoundID()...)
				}
				b.withPositionLocation(entry.instance.SourcePosition(), clone)
				b.define(clone.Symbol, clone)
				b.localModuleDefinitions[cacheKey] = clone
				return clone
			}
		}
	}
	if entry.module != nil {
		for i := range entry.module.Definitions {
			def := &entry.module.Definitions[i]
			if def.Name == entry.name {
				return b.convertSourceDefinitionAs(entry.module.Name+"!"+def.Name, def)
			}
		}
	}
	// A named INSTANCE is a ModuleInstanceKind OpDef with no body. It is
	// visible in the context but excluded by Context.getOpDefs.
	for _, instance := range mod.Instances {
		if instance.Name == entry.name {
			params := make([]*tlc.SymbolNode, len(instance.Params))
			for i, param := range instance.Params {
				params[i] = b.formalParameter(param, instance.ParamArities[param], instance.ParamPositions[param], nil)
			}
			def := tlc.NewOpDefNode(entry.name, params, nil)
			def.KindValue = tlc.SemanticModuleInstanceKind
			return def
		}
	}
	if entry.module != nil && !strings.Contains(entry.name, "!") {
		for _, assumption := range entry.module.Assumptions {
			if assumption.Name == entry.name {
				previous := b.convertingModule
				b.convertingModule = entry.module.Name
				body := b.convertExpr(assumption.Expr)
				b.convertingModule = previous
				return tlc.NewOpDefNode(entry.name, nil, body)
			}
		}
	}
	return nil
}
