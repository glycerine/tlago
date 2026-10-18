package tlc

import "strings"

// GetScopedIdentifiers ports TLCDebuggerExpression.getScopedIdentifiers. The
// signatures retain operator arity so generated debugger parameters can accept
// operator arguments as well as values.
func GetScopedIdentifiers(root *ModuleNode, location SourceLocation) *InsMap[string, struct{}] {
	if root == nil {
		panic(NewNullPointerException())
	}
	symbols := GetScopedSymbols(root, root.PathTo(location, false))
	identifiers := NewInsMap[string, struct{}]()
	for symbol := range symbols.All() {
		name := symbol.GetName()
		if symbol.Definition != nil {
			name = symbol.Definition.Name
		}
		signature := name.String()
		if arity := symbol.GetArity(); arity != 0 {
			signature += "(" + strings.TrimSuffix(strings.Repeat("_,", arity), ",") + ")"
		}
		identifiers.Set(signature, struct{}{})
	}
	return identifiers
}

// GetScopedSymbols ports TLCDebuggerExpression.getScopedSymbols. Identity is
// retained until the caller formats names or removes LET definitions by name.
func GetScopedSymbols(root *ModuleNode, path []SemanticNode) *InsMap[*SymbolNode, struct{}] {
	identifiers := NewInsMap[*SymbolNode, struct{}]()
	addParameters := func(parameters []*SymbolNode) {
		for _, parameter := range parameters {
			identifiers.Set(parameter, struct{}{})
		}
	}
	for _, current := range path {
		switch node := current.(type) {
		case *LetInNode:
			for _, definition := range node.Lets {
				identifiers.Set(definition.Symbol, struct{}{})
			}
		case *OpDefNode:
			if root.GetOpDef(node.Name) == nil {
				identifiers.Set(node.Symbol, struct{}{})
			}
			addParameters(node.GetParams())
		case *OpApplNode:
			addParameters(node.GetQuantSymbolLists())
		case *OpArgNode:
			if definition := node.Op.Definition; definition != nil {
				addParameters(definition.GetParams())
			}
		}
	}
	return identifiers
}
