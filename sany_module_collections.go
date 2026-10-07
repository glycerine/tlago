// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

// Context declaration and module vectors use Hashtable enumeration. Definition
// vectors instead follow the Pair history chain, including replaced entries.
func (context *sanyContext) getConstantDecls() []*sanySemOpDeclNode {
	return context.getDeclarations(sanyConstantDeclKind)
}
func (context *sanyContext) getVariableDecls() []*sanySemOpDeclNode {
	return context.getDeclarations(sanyVariableDeclKind)
}
func (context *sanyContext) getDeclarations(kind sanySemKind) []*sanySemOpDeclNode {
	result := make([]*sanySemOpDeclNode, 0)
	for _, symbol := range context.contentSymbols() {
		if declaration, ok := symbol.(*sanySemOpDeclNode); ok && declaration.semKind() == kind {
			result = append(result, declaration)
		}
	}
	return result
}
func (context *sanyContext) getOpDefs() []*sanySemOpDefNode {
	if context == nil {
		panic(tlc.NewNullPointerException())
	}
	result := make([]*sanySemOpDefNode, 0)
	for i := len(context.order) - 1; i >= 0; i-- {
		if definition, ok := context.order[i].sym.(*sanySemOpDefNode); ok && definition.semKind() != sanyModuleInstanceKind && definition.semKind() != sanyBuiltInKind {
			result = append(result, definition)
		}
	}
	return result
}
func (context *sanyContext) getThmOrAssDefs() []*sanySemThmOrAssumpDefNode {
	if context == nil {
		panic(tlc.NewNullPointerException())
	}
	result := make([]*sanySemThmOrAssumpDefNode, 0)
	for i := len(context.order) - 1; i >= 0; i-- {
		if definition, ok := context.order[i].sym.(*sanySemThmOrAssumpDefNode); ok {
			result = append(result, definition)
		}
	}
	return result
}
func (context *sanyContext) getModDefs() []*sanySemModuleNode {
	result := make([]*sanySemModuleNode, 0)
	for _, symbol := range context.contentSymbols() {
		if module, ok := symbol.(*sanySemModuleNode); ok {
			result = append(result, module)
		}
	}
	return result
}

// Module arrays are lazily copied once. Later context mutations do not invalidate
// these arrays, and callers receive the same array on every subsequent access.
func (module *sanySemModuleNode) getConstantDecls() []*sanySemOpDeclNode {
	if module.constantDecls == nil {
		vector := module.context.getConstantDecls()
		module.constantDecls = make([]*sanySemOpDeclNode, len(vector))
		for i, declaration := range vector {
			module.constantDecls[len(vector)-1-i] = declaration
		}
	}
	return module.constantDecls
}
func (module *sanySemModuleNode) getVariableDecls() []*sanySemOpDeclNode {
	if module.variableDecls == nil {
		vector := module.context.getVariableDecls()
		module.variableDecls = make([]*sanySemOpDeclNode, len(vector))
		for i, declaration := range vector {
			module.variableDecls[len(vector)-1-i] = declaration
		}
	}
	return module.variableDecls
}
func (module *sanySemModuleNode) getOpDefs() []*sanySemOpDefNode {
	if module.opDefs == nil {
		vector := module.context.getOpDefs()
		module.opDefs = make([]*sanySemOpDefNode, len(vector))
		for i, definition := range vector {
			module.opDefs[len(vector)-1-i] = definition
		}
	}
	return module.opDefs
}
func (module *sanySemModuleNode) getThmOrAssDefs() []*sanySemThmOrAssumpDefNode {
	if module.thmOrAssDefs == nil {
		vector := module.context.getThmOrAssDefs()
		module.thmOrAssDefs = make([]*sanySemThmOrAssumpDefNode, len(vector))
		for i, definition := range vector {
			module.thmOrAssDefs[len(vector)-1-i] = definition
		}
	}
	return module.thmOrAssDefs
}
func (module *sanySemModuleNode) getInnerModules() []*sanySemModuleNode {
	if module.modDefs == nil {
		vector := module.context.getModDefs()
		module.modDefs = make([]*sanySemModuleNode, len(vector))
		copy(module.modDefs, vector)
	}
	return module.modDefs
}
