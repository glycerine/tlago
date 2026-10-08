// Copyright (c) 2024 Linux Foundation. All rights reserved.
// SPDX-License-Identifier: MIT
package tlago

import (
	"fmt"
	"maps"
	"os"

	"github.com/glycerine/tlago/tlc"
)

// TLCDebuggerExpression.resolveDependencies uses the default resolver, not the
// resolver that originally loaded the spec. Successful dependencies remain in
// the external table even when a subsequent expression fails to compile.
func (b *tlcBridge) resolveDebuggerDependencies(dependencies []string) (bool, error) {
	classpath, err := tlcApplicationClasspath(nil)
	if err != nil {
		return false, err
	}
	resolver := tlc.NewSimpleFilenameToStream(nil, tlc.FilenameResolverOptions{Classpath: classpath})
	failed := false
	loading := map[string]bool{}
	var resolve func([]string) error
	resolve = func(dependencies []string) error {
		for _, name := range dependencies {
			if b.processor.ModuleTbl.GetModuleNode(tlc.UniqueStringOf(name)) != nil {
				continue
			}
			if loading[name] {
				// Java's recursive resolver also overflows on a dependency cycle.
				panic(tlc.NewStackOverflowError())
			}
			file := resolver.Resolve(name+".tla", false)
			source, readError := os.ReadFile(file.GetPath())
			if readError != nil {
				failed = true // Java adds an IO error to the shared semantic log.
				continue
			}
			dependency, transitive, diagnostics := parseSanyModuleSourceWithDependencies(name+".tla", tlc.DecodeUTF8Replacing(source))
			if diagnostics.HasErrors() || dependency == nil {
				return fmt.Errorf("Syntax error while parsing breakpoint expression's dependency \"%s\"", name)
			}
			loading[name] = true
			if err := resolve(transitive); err != nil {
				return err
			}
			delete(loading, name)
			if failed {
				continue
			}
			candidate := b.debuggerSpec(dependency)
			diagnostics = CheckSpec(candidate)
			semanticFailure, levelFailure := debuggerDiagnosticFailures(diagnostics)
			if semanticFailure {
				failed = true
				continue
			}
			convert := b.debuggerBridge(candidate)
			convert.prepareInstanceDefinitions()
			convert.extendModuleTable(false)
			convert.installModuleAssumptions(false)
			if convert.diags.HasErrors() {
				failed = true
				continue
			}
			node := convert.moduleNodes[dependency]
			b.processor.ProcessConstantsDynamicExtendee(node)
			b.processor.ModuleTbl.Put(tlc.UniqueStringOf(name), node.Context, node)
			// Retain the original AST root and semantic root while keeping the
			// newly compiled dependency and its source/instance identities.
			candidate.Root = b.spec.Root
			*b = *convert
			failed = levelFailure
		}
		return nil
	}
	err = resolve(dependencies)
	return failed, err
}

func debuggerDiagnosticFailures(diagnostics Diagnostics) (semantic, level bool) {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity != SeverityError {
			continue
		}
		if debuggerLevelDiagnostic(diagnostic.Code) {
			level = true
		} else {
			semantic = true
		}
	}
	return
}

func (b *tlcBridge) debuggerSpec(module *Module) *Spec {
	spec := *b.spec
	spec.Root = module
	spec.Modules = maps.Clone(b.spec.Modules)
	var add func(*Module)
	add = func(mod *Module) {
		spec.Modules[mod.Name] = mod
		for _, nested := range mod.Nested {
			add(nested)
		}
	}
	add(module)
	spec.SemanticOrder = append(append([]string(nil), b.spec.SemanticOrder...), module.Name)
	return &spec
}

// A temporary wrapper gets its own conversion indexes, while existing semantic
// nodes keep the config/native/lazy objects installed by the running tool.
func (b *tlcBridge) debuggerBridge(spec *Spec) *tlcBridge {
	convert := *b
	convert.spec = spec
	convert.defs = maps.Clone(b.defs)
	for name, definition := range tlcBridgeDefinitionsByName(spec) {
		if convert.defs[name] == nil {
			convert.defs[name] = definition
		}
	}
	for i := range spec.Root.Definitions {
		definition := &spec.Root.Definitions[i]
		convert.defs[spec.Root.Name+"!"+definition.Name] = definition
	}
	convert.symbols = maps.Clone(b.symbols)
	convert.definitionModules = maps.Clone(b.definitionModules)
	convert.sourceSymbols = maps.Clone(b.sourceSymbols)
	convert.sourceDefinitions = maps.Clone(b.sourceDefinitions)
	convert.canonicalLets = maps.Clone(b.canonicalLets)
	convert.canonicalFormals = maps.Clone(b.canonicalFormals)
	convert.conversionContexts = maps.Clone(b.conversionContexts)
	convert.instanceDefinitions = maps.Clone(b.instanceDefinitions)
	convert.moduleNodes = maps.Clone(b.moduleNodes)
	convert.localModuleDefinitions = maps.Clone(b.localModuleDefinitions)
	convert.builtinDefinitions = maps.Clone(b.builtinDefinitions)
	convert.indexedModules = maps.Clone(b.indexedModules)
	convert.assumptionModules = maps.Clone(b.assumptionModules)
	convert.moduleDefinitionNames = moduleDefinitionNameIndex(spec)
	convert.convertBoundNames = map[string]int{}
	convert.diags = nil
	return &convert
}
