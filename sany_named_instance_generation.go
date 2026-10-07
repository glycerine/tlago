// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// processModuleDefinition qualifies every eligible symbol and prepends instance
// parameters. LET callers own the returned definitions and instance vectors.
func (g *sanyExpressionGeneration) generateNamedInstance(instance *Instance, let, moduleInstance bool) Diagnostics {
	if instance.Name == "" || instance.substitutionNode == nil {
		return nil
	}
	instance.semanticNode, instance.definitionNode, instance.importedDefinitions = nil, nil, nil
	target := g.spec.Modules[instance.Module]
	if target == nil || target.semanticNode == nil {
		return nil
	}
	origin, module := target.semanticNode, g.currentModule.semanticNode
	symbols := origin.context.contentSymbols()
	for _, symbol := range symbols {
		switch node := symbol.(type) {
		case *sanySemOpDefNode:
			if node.semLocal() || (node.semKind() != sanyUserDefinedOpKind && node.semKind() != sanyModuleInstanceKind) {
				continue
			}
			if node.semKind() == sanyUserDefinedOpKind {
				if node.body == nil {
					return nil
				}
			}
		case *sanySemThmOrAssumpDefNode:
			if node.semLocal() {
				continue
			}
			if node.body == nil {
				return nil
			}
		}
	}
	var diagnostics Diagnostics
	prefix := instance.Name + "!"
	for _, symbol := range symbols {
		original, ok := symbol.(*sanySemOpDefNode)
		if !ok || original.semLocal() || (original.semKind() != sanyUserDefinedOpKind && original.semKind() != sanyModuleInstanceKind) {
			continue
		}
		parameters := make([]*sanyFormalParamNode, len(instance.formalNodes)+len(original.formalNodes))
		copy(parameters, instance.formalNodes)
		copy(parameters[len(instance.formalNodes):], original.formalNodes)
		var imported *sanySemOpDefNode
		var generated Diagnostics
		if original.semKind() == sanyModuleInstanceKind {
			imported, generated = newSanySemModuleInstanceOpDefNode(prefix+original.semName(), parameters, instance.Local, original.module, g.formalSymbolTable(), instance.Syntax, sanyOriginalSource(original).(*sanySemOpDefNode))
		} else {
			body := original.body
			if len(instance.substitutionNode.substs) > 0 {
				wrapper, ds := newSanySemSubstInNode(instance.Syntax, instance.substitutionNode, body, module, origin)
				diagnostics = append(diagnostics, ds...)
				body = wrapper
			}
			imported, generated = newSanySemOpDefNode(prefix+original.semName(), sanyUserDefinedOpKind, parameters, instance.Local, body, module, g.formalSymbolTable(), instance.Syntax, true, sanyOriginalSource(original).(*sanySemOpDefNode))
			if len(instance.substitutionNode.substs) > 0 {
				imported.compoundID = append([]*tlc.UniqueString{tlc.UniqueStringOf(instance.Name)}, original.getCompoundId()...)
			}
			g.setDefinitionRecursionFields(imported)
			imported.labels = original.labels
		}
		diagnostics = append(diagnostics, generated...)
		instance.importedDefinitions = append(instance.importedDefinitions, imported)
		if !let {
			module.definitions = append(module.definitions, imported)
		}
	}
	for _, symbol := range symbols {
		original, ok := symbol.(*sanySemThmOrAssumpDefNode)
		if !ok || original.semLocal() {
			continue
		}
		parameters := make([]*sanyFormalParamNode, len(instance.formalNodes)+len(original.formalNodes))
		copy(parameters, instance.formalNodes)
		copy(parameters[len(instance.formalNodes):], original.formalNodes)
		body := original.body
		if len(instance.substitutionNode.substs) > 0 {
			wrapper, ds := newSanySemAPSubstInNode(instance.Syntax, instance.substitutionNode, body, module, origin)
			diagnostics = append(diagnostics, ds...)
			body = wrapper
		}
		imported, ds := newSanySemCompletedThmOrAssumpDefNode(prefix+original.semName(), original.theorem, body, module, g.formalSymbolTable(), instance.Syntax, parameters, origin, original.getSource())
		diagnostics = append(diagnostics, ds...)
		imported.local, imported.labels = instance.Local, original.labels
		instance.importedDefinitions = append(instance.importedDefinitions, imported)
		if !let {
			module.definitions = append(module.definitions, imported)
		}
	}
	instance.semanticNode = newSanySemInstanceNode(tlc.UniqueStringOf(instance.Name), instance.Local, instance.formalNodes, origin, instance.substitutionNode.substs, instance.Syntax)
	if moduleInstance {
		module.appendInstance(instance.semanticNode)
	}
	definition, ds := newSanySemModuleInstanceOpDefNode(instance.Name, instance.formalNodes, instance.Local, module, g.formalSymbolTable(), instance.Syntax, nil)
	diagnostics = append(diagnostics, ds...)
	instance.definitionNode = definition
	return diagnostics
}
func (node *sanySemOpDefNode) getCompoundId() []*tlc.UniqueString {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if node.compoundID != nil {
		return node.compoundID
	}
	return []*tlc.UniqueString{tlc.UniqueStringOf(node.semName())}
}

func (node *sanySemOpDefNode) getLocalName() *tlc.UniqueString {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if node.compoundID != nil {
		if len(node.compoundID) == 0 {
			panic(tlc.NewArrayIndexOutOfBoundsException(-1, 0))
		}
		return node.compoundID[len(node.compoundID)-1]
	}
	return tlc.UniqueStringOf(node.semName())
}

func (node *sanySemOpDefNode) hasPath() bool {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	return node.compoundID != nil && len(node.compoundID) > 1
}

func (node *sanySemOpDefNode) getPathName() *tlc.UniqueString {
	if node == nil {
		panic(tlc.NewNullPointerException())
	}
	if node.compoundID != nil {
		return tlc.UniqueStringJoinN("!", len(node.compoundID)-1, node.compoundID)
	}
	return tlc.UniqueStringOf("")
}

// A selector consisting entirely of module-instance prefixes and a final name
// applies the actual qualified definition. Operand and label selectors follow
// different source rules and are not represented by reconstructed definitions.
func (g *sanyExpressionGeneration) retainCanonicalInstanceSelection(expr Expr, operatorArgument, symbolReferenceOnly bool) Diagnostics {
	source, selected := sanyExprSource(expr), sanyExprSelection(expr)
	if source == nil || selected == nil || source.Selector == nil || len(source.Selector.Steps) < 2 {
		return nil
	}
	name := ""
	for i, step := range source.Selector.Steps {
		if step.Kind != SanySelectorName {
			return nil
		}
		if name != "" {
			name += "!"
		}
		name += step.Name
		if i < len(source.Selector.Steps)-1 {
			prefix := g.formalSymbolTable().resolveSymbol(name)
			if prefix == nil || prefix.semKind() != sanyModuleInstanceKind {
				return nil
			}
		}
	}
	symbol := g.formalSymbolTable().resolveSymbol(name)
	if symbol == nil {
		return nil
	}
	args := make([]sanySemanticGraphNode, len(selected.args))
	for i, arg := range selected.args {
		args[i] = sanyGeneratedExpressionNode(arg)
		if args[i] == nil && sanyExpressionGenerationFailure(arg) != sanyGenerationNullExpression {
			return nil
		}
	}
	if operatorArgument {
		source.semanticGraph = newSanySemOpArgNode(symbol, source.Syntax, g.currentModule.semanticNode)
		return nil
	}
	if symbolReferenceOnly {
		return nil
	}
	node, diagnostics, err := newSanySemOpApplNode(symbol, args, source.Syntax)
	if err != nil {
		panic(err)
	}
	if symbol.semKind() == sanyUserDefinedOpKind || symbol.semKind() == sanyThmOrAssumpDefKind {
		node.subExpressionOf = symbol
	}
	source.semanticGraph = node
	return diagnostics
}

func (g *sanyExpressionGeneration) canonicalInstanceSelectorSymbol(expr Expr) sanySemSymbol {
	source := sanyExprSource(expr)
	if source == nil || source.Selector == nil || len(source.Selector.Steps) < 2 {
		return nil
	}
	name := ""
	for i, step := range source.Selector.Steps {
		if step.Kind != SanySelectorName {
			return nil
		}
		if name != "" {
			name += "!"
		}
		name += step.Name
		if i < len(source.Selector.Steps)-1 {
			symbol := g.formalSymbolTable().resolveSymbol(name)
			if symbol == nil || symbol.semKind() != sanyModuleInstanceKind {
				return nil
			}
		}
	}
	return g.formalSymbolTable().resolveSymbol(name)
}

// DEF selectors name a definition without applying any instance prefix.
func (g *sanyExpressionGeneration) canonicalInstanceDefinitionSymbol(expr Expr) sanySemSymbol {
	source := sanyExprSource(expr)
	if source == nil || source.Selector == nil {
		return nil
	}
	for _, step := range source.Selector.Steps {
		if step.Arguments != nil {
			return nil
		}
	}
	return g.canonicalInstanceSelectorSymbol(expr)
}
