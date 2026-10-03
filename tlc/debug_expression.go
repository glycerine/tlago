/*******************************************************************************
 * Copyright (c) 2024 Linux Foundation. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Andrew Helwer - code interfacing with SANY
 *   Markus Alexander Kuppe - code interfacing with TLC interpreter
 ******************************************************************************/
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
		if symbol.Arity != 0 {
			signature += "(" + strings.TrimSuffix(strings.Repeat("_,", symbol.Arity), ",") + ")"
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
			addParameters(node.Params)
		case *OpApplNode:
			addParameters(node.GetQuantSymbolLists())
		case *OpArgNode:
			if definition := node.Op.Definition; definition != nil {
				addParameters(definition.Params)
			}
		}
	}
	return identifiers
}
