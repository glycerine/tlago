// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

func (module *sanySemModuleNode) isParameterFree() bool {
	return len(module.getConstantDecls()) == 0 && len(module.getVariableDecls()) == 0
}
func (module *sanySemModuleNode) getInstances() []*sanySemInstanceNode {
	if module.instances == nil {
		module.instances = make([]*sanySemInstanceNode, len(module.instanceVec))
		copy(module.instances, module.instanceVec)
	}
	return module.instances
}
func (module *sanySemModuleNode) appendInstance(instance *sanySemInstanceNode) {
	module.instanceVec = append(module.instanceVec, instance)
	module.topLevelVec = append(module.topLevelVec, instance)
}

// generateInstance handles unnamed imports. Named module definitions have their
// own source constructor and parameter concatenation rules, ported separately.
func (g *sanyExpressionGeneration) generateUnnamedInstance(instance *Instance, topLevel bool) Diagnostics {
	instance.semanticNode = nil
	if instance.Name != "" || instance.substitutionNode == nil {
		return nil
	}
	target := g.spec.Modules[instance.Module]
	if target == nil || target.semanticNode == nil {
		return nil
	}
	module, origin := g.currentModule.semanticNode, target.semanticNode
	local := instance.Local
	if instance.Syntax != nil {
		local = instance.Syntax.Zero != nil
	}
	symbols := origin.context.contentSymbols()
	for _, symbol := range symbols {
		switch node := symbol.(type) {
		case *sanySemOpDefNode:
			if node.semKind() == sanyBuiltInKind || node.semKind() == sanyModuleInstanceKind || node.semLocal() {
				continue
			}
			if node.module == nil || node.body == nil {
				return nil
			}
		case *sanySemThmOrAssumpDefNode:
			if node.semLocal() {
				continue
			}
			if node.module == nil || node.body == nil {
				return nil
			}
		}
	}
	var diagnostics Diagnostics
	parameterFree := origin.isParameterFree()
	for _, symbol := range symbols {
		original, ok := symbol.(*sanySemOpDefNode)
		if !ok || original.semKind() == sanyBuiltInKind || original.semKind() == sanyModuleInstanceKind || original.semLocal() {
			continue
		}
		var imported *sanySemOpDefNode
		if !parameterFree && !original.module.isParameterFree() {
			body, generated := newSanySemSubstInNode(instance.Syntax, instance.substitutionNode, original.body, module, origin)
			diagnostics = append(diagnostics, generated...)
			imported, generated = newSanySemOpDefNode(original.semName(), sanyUserDefinedOpKind, original.formalNodes, local, body, module, g.formalSymbolTable(), instance.Syntax, true, sanyOriginalSource(original).(*sanySemOpDefNode))
			diagnostics = append(diagnostics, generated...)
			imported.labels = original.labels
		} else if local && (parameterFree || topLevel) {
			var generated Diagnostics
			imported, generated = newSanySemOpDefNode(original.semName(), sanyUserDefinedOpKind, original.formalNodes, true, original.body, original.module, g.formalSymbolTable(), instance.Syntax, true, sanyOriginalSource(original).(*sanySemOpDefNode))
			diagnostics = append(diagnostics, generated...)
			imported.labels = original.labels
		} else {
			imported = original
			diagnostics = append(diagnostics, g.formalSymbolTable().addSymbol(imported)...)
		}
		module.definitions = append(module.definitions, imported)
		g.setDefinitionRecursionFields(imported)
	}
	for _, symbol := range symbols {
		original, ok := symbol.(*sanySemThmOrAssumpDefNode)
		if !ok || original.semLocal() {
			continue
		}
		var imported *sanySemThmOrAssumpDefNode
		if !parameterFree && !original.module.isParameterFree() {
			body, generated := newSanySemAPSubstInNode(instance.Syntax, instance.substitutionNode, original.body, module, origin)
			diagnostics = append(diagnostics, generated...)
			imported, generated = newSanySemCompletedThmOrAssumpDefNode(original.semName(), original.theorem, body, module, g.formalSymbolTable(), instance.Syntax, original.formalNodes, origin, original.getSource())
			diagnostics = append(diagnostics, generated...)
			imported.local, imported.labels = local, original.labels
		} else if local && topLevel {
			originalModule := original.module
			if !parameterFree {
				originalModule = module
			}
			var generated Diagnostics
			imported, generated = newSanySemCompletedThmOrAssumpDefNode(original.semName(), original.theorem, original.body, originalModule, g.formalSymbolTable(), instance.Syntax, original.formalNodes, origin, original.getSource())
			diagnostics = append(diagnostics, generated...)
			imported.local = true
			if parameterFree {
				imported.labels = original.labels
			}
		} else {
			imported = original
			// Source omits registration in this specific parameterized-module
			// branch when the original theorem's module is parameter-free.
			if parameterFree {
				diagnostics = append(diagnostics, g.formalSymbolTable().addSymbol(imported)...)
			}
		}
		if topLevel {
			module.definitions = append(module.definitions, imported)
		}
	}
	instance.semanticNode = newSanySemInstanceNode(nil, local, nil, origin, instance.substitutionNode.substs, instance.Syntax)
	if topLevel {
		module.appendInstance(instance.semanticNode)
	}
	return diagnostics
}

func retainSanyInstanceSymbol(symbol localSymbol, actual sanySemSymbol) localSymbol {
	switch node := actual.(type) {
	case *sanySemOpDefNode:
		symbol.opDefNode = node
	case *sanySemThmOrAssumpDefNode:
		symbol.theoremDefNode = node
	}
	return symbol
}
