// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.

package tlago

import (
	"strings"

	"github.com/glycerine/tlago/tlc"
)

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

// Retain the external dependency order and full represented semantic contexts.
// Internal modules have graph identities too, but never become external table
// entries merely because the loader indexes them by name.
func (b *tlcBridge) installModuleTable() {
	table := tlc.NewExternalModuleTable()
	b.processor.ModuleTbl = table
	nodes := map[*Module]*tlc.ModuleNode{}
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
	b.moduleNodes = nodes
	b.localModuleDefinitions = map[string]*tlc.OpDefNode{}
	for definition, node := range b.sourceDefinitions {
		node.OriginallyDefinedInModule = nodes[b.spec.Modules[b.definitionModules[definition]]]
		node.Local = definition.Local
	}
	for _, mod := range modules {
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
		if outer[mod] == nil {
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
	table.SetRootModule(nodes[b.spec.Root])
	b.processor.RootModule = table.GetRootModule()
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
	if entry.instance != nil && entry.instance.Local {
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
		return b.convertDefinitionAs(key, def)
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
				params[i] = tlc.NewSymbolNode(param)
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
