// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2022, Oracle and/or its affiliates.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import (
	"fmt"

	"github.com/glycerine/tlago/tlc"
)

// Generator retains declaration identities across LET contexts. In particular,
// leaving a LET does not discard its entries from ModuleNode.recursiveDecls.
type sanyRecursiveBinding struct {
	name     string
	position Position
	arity    int
	level    int
	defined  bool
}

type sanyExpressionGeneration struct {
	level         int
	spec          *Spec
	module        *sanyModuleRecursiveGeneration
	declarations  []*sanyRecursiveBinding
	bindings      map[string]*sanyRecursiveBinding
	symbols       map[string]localSymbol
	moduleKinds   map[string]DeclarationKind
	moduleArities map[string]int
	moduleSymbols map[string]localSymbol
}

func sanyExpressionGenerator(generators []*sanyExpressionGeneration) *sanyExpressionGeneration {
	if len(generators) != 0 && generators[0] != nil {
		return generators[0]
	}
	return &sanyExpressionGeneration{module: &sanyModuleRecursiveGeneration{}, bindings: map[string]*sanyRecursiveBinding{}}
}

func (g *sanyExpressionGeneration) complete(binding *sanyRecursiveBinding, position Position) {
	binding.defined = true
	binding.position = position
	if g.level == 0 {
		g.module.count--
	} else {
		g.module.counts[g.level]--
	}
	g.module.sum--
	if g.module.sum < 0 {
		panic(tlc.NewWrongInvocationException("Defined more recursive operators than were declared in RECURSIVE statements."))
	}
}

func sanyRecursiveDefinitionDiagnostic(definition Definition, code, message string) Diagnostic {
	position := definition.SourcePosition()
	diagnostic := sanyDiagnosticParameters(errorAt(position, code, "%s", message), definition.Name)
	diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
	diagnostic.SANYMessage = message
	if code == "E4293" {
		diagnostic.Message = fmt.Sprintf("recursive declaration %s is defined in the wrong LET/IN level", definition.Name)
	}
	return diagnostic
}

func sanyUndefinedRecursiveDiagnostic(binding *sanyRecursiveBinding) Diagnostic {
	diagnostic := sanyDiagnosticParameters(errorAt(binding.position, "E4291", "recursive declaration %s has no definition", binding.name), binding.name)
	diagnostic.SANYRange = SanyRange{Begin: binding.position, End: binding.position.SourceEnd()}
	diagnostic.SANYMessage = fmt.Sprintf("Symbol %s declared in RECURSIVE statement but not defined.", binding.name)
	return diagnostic
}

func (g *sanyExpressionGeneration) checkLet(expr *LetExpr, context map[string]Position, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	defer func() {
		if failure := recover(); failure != nil {
			if g.spec != nil {
				g.spec.SemanticDiags = appendSanyDiagnostics(append(Diagnostics(nil), diags...), g.spec.SemanticDiags...)
			}
			panic(failure)
		}
	}()
	// processLetIn raises the level for the definitions, then lowers it before
	// generating IN while keeping the same symbol context on the stack.
	g.level++
	if g.level >= len(g.module.counts) {
		panic(tlc.NewArrayIndexOutOfBoundsException(g.level, len(g.module.counts)))
	}
	g.module.counts[g.level] = 0
	previousSymbols := g.symbols
	g.symbols = make(map[string]localSymbol, len(previousSymbols)+len(expr.Definitions))
	for name, symbol := range previousSymbols {
		g.symbols[name] = symbol
	}
	defer func() { g.symbols = previousSymbols }()
	previousBindings := g.bindings
	g.bindings = make(map[string]*sanyRecursiveBinding, len(previousBindings))
	for name, binding := range previousBindings {
		g.bindings[name] = binding
	}
	defer func() { g.bindings = previousBindings }()
	letLocals := copyBoolMap(locals)
	positions := copySanyExpressionContext(context)
	for _, unit := range sanyLetGenerationUnits(expr) {
		switch {
		case unit.declaration != nil:
			declaration := unit.declaration
			g.module.counts[g.level] += len(declaration.Names)
			g.module.sum += len(declaration.Names)
			for _, name := range declaration.Names {
				arity, _ := declarationArity(*declaration, name)
				binding := &sanyRecursiveBinding{name: name, position: declarationSymbolPosition(*declaration, name), arity: arity, level: g.level}
				g.declarations = append(g.declarations, binding)
				if previous, exists := g.lookupSymbol(name, positions); exists {
					diags = append(diags, instanceSymbolConflict(semanticExportedSymbol{name: name, kind: OperatorDecl, arity: arity, source: binding.position}, previous)...)
				} else {
					g.bindings[name] = binding
					positions[name] = binding.position
					g.symbols[name] = localSymbol{kind: OperatorDecl, arity: arity, pos: binding.position}
					letLocals[name] = true
				}
			}
		case unit.definition != nil:
			definition := *unit.definition
			diags = append(diags, checkDefinitionParams(definition)...)
			diags = append(diags, checkDefinitionParamCollisions(definition, positions, letLocals)...)
			bodyLocals := copyBoolMap(letLocals)
			for _, name := range definition.Params {
				bodyLocals[name] = true
			}
			// Domains precede resolving the function's own symbol.
			if definition.FunctionDef {
				diags = append(diags, checkDefinitionFunctionDomains(definition, positions, bodyLocals, g)...)
			}
			previousSymbol, symbolExists := g.lookupSymbol(definition.Name, positions)
			binding := g.bindings[definition.Name]
			recursive := binding != nil && !binding.defined
			wrongLevel := recursive && binding.level != g.level
			if wrongLevel {
				kind := "operator"
				if definition.FunctionDef {
					kind = "function"
				}
				diags = append(diags, sanyRecursiveDefinitionDiagnostic(definition, "E4293", fmt.Sprintf("Recursive %s %s defined at wrong LET/IN level.", kind, definition.Name)))
			} else if recursive {
				matches := len(definition.Params) == binding.arity
				for _, arity := range definition.ParamArities {
					matches = matches && arity == 0
				}
				if !matches {
					message := fmt.Sprintf("Definition of %s has different arity than its RECURSIVE declaration.", definition.Name)
					if definition.FunctionDef {
						message = fmt.Sprintf("Function %s has operator arguments in its RECURSIVE declaration.", definition.Name)
					}
					diags = append(diags, sanyRecursiveDefinitionDiagnostic(definition, "E4292", message))
				}
				if definition.FunctionDef && binding.arity == 0 {
					g.complete(binding, definition.SourcePosition())
				}
				if !definition.FunctionDef {
					binding.arity = len(definition.Params)
				}
			} else if _, exists := positions[definition.Name]; exists || locals[definition.Name] {
				message := fmt.Sprintf("Operator %s already defined or declared.", definition.Name)
				if definition.FunctionDef {
					message = fmt.Sprintf("Function name `%s' already defined or declared.", definition.Name)
				}
				diags = append(diags, sanyRecursiveDefinitionDiagnostic(definition, "E4201", message))
			}
			if definition.FunctionDef {
				if wrongLevel || (symbolExists && previousSymbol.kind != OperatorDecl) {
					// processFunction still generates the body, but does not push its
					// quantifier context when resolving the symbol failed.
					function, _ := definition.Expr.(*FunctionExpr)
					if function != nil {
						diags = append(diags, g.checkExpr(function.Body, positions, letLocals)...)
					}
				} else {
					if binding == nil && !symbolExists {
						letLocals[definition.Name] = true
						positions[definition.Name] = definition.SourcePosition()
						g.symbols[definition.Name] = localSymbol{kind: OperatorDecl, arity: len(definition.Params), pos: definition.SourcePosition()}
					}
					diags = append(diags, checkDefinitionFunctionBody(definition, positions, bodyLocals, g)...)
				}
			} else {
				diags = append(diags, g.checkDefinitionBody(definition, positions, bodyLocals)...)
				if recursive && !wrongLevel {
					g.complete(binding, definition.SourcePosition())
				}
				if wrongLevel {
					// The newly constructed OpDefNode calls SymbolTable.addSymbol; the
					// existing declaration remains the symbol table binding.
					diags = append(diags, instanceSymbolConflict(semanticExportedSymbol{name: definition.Name, kind: OperatorDecl, arity: len(definition.Params), source: definition.SourcePosition()}, localSymbol{kind: OperatorDecl, arity: binding.arity, pos: binding.position})...)
				} else if !recursive {
					if previous, exists := g.lookupSymbol(definition.Name, positions); exists {
						diags = append(diags, instanceSymbolConflict(semanticExportedSymbol{name: definition.Name, kind: OperatorDecl, arity: len(definition.Params), source: definition.SourcePosition()}, previous)...)
					} else {
						positions[definition.Name] = definition.SourcePosition()
						g.symbols[definition.Name] = localSymbol{kind: OperatorDecl, arity: len(definition.Params), pos: definition.SourcePosition()}
						letLocals[definition.Name] = true
					}
				}
			}
		case unit.instance != nil:
			instance := unit.instance
			bodyLocals := copyBoolMap(letLocals)
			for _, name := range instance.Params {
				bodyLocals[name] = true
			}
			for _, substitution := range instance.SubstitutionList {
				diags = append(diags, g.checkExpr(substitution.Expr, positions, bodyLocals)...)
			}
			if instance.Name != "" {
				letLocals[instance.Name+"!"] = true
			}
		}
	}
	if g.module.counts[g.level] > 0 {
		for _, binding := range g.declarations {
			if binding.level == g.level && !binding.defined {
				diags = append(diags, sanyUndefinedRecursiveDiagnostic(binding))
			}
		}
		g.module.sum -= g.module.counts[g.level]
	}
	g.level--
	diags = append(diags, g.checkExpr(expr.Body, positions, letLocals)...)
	return diags
}

func (g *sanyExpressionGeneration) checkDefinitionBody(definition Definition, context map[string]Position, locals map[string]bool) Diagnostics {
	return g.checkExpr(definition.Expr, context, locals)
}

func (g *sanyExpressionGeneration) lookupSymbol(name string, context map[string]Position) (localSymbol, bool) {
	if binding := g.bindings[name]; binding != nil {
		return localSymbol{kind: OperatorDecl, arity: binding.arity, pos: binding.position}, true
	}
	if symbol, ok := g.symbols[name]; ok {
		return symbol, true
	}
	position, ok := context[name]
	if !ok {
		return localSymbol{}, false
	}
	if symbol, exists := g.moduleSymbols[name]; exists {
		return symbol, true
	}
	kind := g.moduleKinds[name]
	if kind == RecursiveDecl {
		kind = OperatorDecl
	}
	return localSymbol{kind: kind, arity: g.moduleArities[name], pos: position}, true
}
