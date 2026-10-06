// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Copyright (c) 2022, Oracle and/or its affiliates.
// Copyright (c) 2026 NVIDIA Corporation. All rights reserved.

package tlago

import (
	"sort"
	"strings"
)

// An unresolved identifier yields Generator's nullOAN placeholder. An
// unresolved symbolic operator yields null instead; N_FcnAppl distinguishes
// the two before generating its arguments.
type sanyGenerationFailure uint8

const (
	sanyGenerationSucceeded sanyGenerationFailure = iota
	sanyGenerationNullOperator
	sanyGenerationNullExpression
)

func sanyExpressionGenerationFailure(expr Expr) sanyGenerationFailure {
	if source, ok := expr.(interface{ generationSource() *SanyExprSource }); ok {
		return source.generationSource().generationFailure
	}
	return sanyGenerationSucceeded
}

func setSanyExpressionGenerationFailure(expr Expr, failure sanyGenerationFailure) {
	if source, ok := expr.(interface{ generationSource() *SanyExprSource }); ok {
		source.generationSource().generationFailure = failure
	}
}

// Generator.generateModule visits the module body's heirs in order. An ordinary
// operator's body is generated before its name is entered in SymbolTable;
// RECURSIVE declarations enter their names when that declaration is visited.
// Keep these expression contexts separate from the completed module context
// used by the subsequent level-checking phase.
type sanyExpressionContexts struct {
	bindings    map[*SanySyntaxNode]map[string]Position
	definitions map[*SanySyntaxNode]map[string]bool
}

func (contexts sanyExpressionContexts) at(syntax *SanySyntaxNode, completed map[string]Position) map[string]Position {
	if context, ok := contexts.bindings[syntax]; ok {
		return context
	}
	// Native semantic AST callers can supply nodes without parser syntax.
	return completed
}

func copySanyExpressionContext(context map[string]Position) map[string]Position {
	copy := make(map[string]Position, len(context))
	for name, position := range context {
		copy[name] = position
	}
	return copy
}

func sanyModuleExpressionContexts(mod *Module, spec *Spec, inherited map[string]Position) sanyExpressionContexts {
	contexts := sanyExpressionContexts{bindings: map[*SanySyntaxNode]map[string]Position{}, definitions: map[*SanySyntaxNode]map[string]bool{}}
	if mod.Syntax == nil {
		return contexts
	}
	heirs := mod.Syntax.GetHeirs()
	if len(heirs) < 3 {
		return contexts
	}
	context := copySanyExpressionContext(inherited)
	completed := make(map[string]bool, len(inherited))
	for name := range inherited {
		completed[name] = true
	}
	add := func(name string, position Position) {
		if _, exists := context[name]; !exists {
			context[name] = position
		}
	}
	addLocal := func(name string, position Position) {
		add(name, position)
		if mod.Name != "" {
			add(mod.Name+"!"+name, position)
		}
	}
	for _, unit := range heirs[2].GetHeirs() {
		contexts.bindings[unit] = copySanyExpressionContext(context)
		contexts.definitions[unit] = copyBoolMap(completed)
		switch unit.Kind.JavaName() {
		case "N_VariableDeclaration", "N_ParamDeclaration":
			for _, declaration := range mod.Declarations {
				if declaration.Syntax != unit {
					continue
				}
				for _, name := range declaration.Names {
					addLocal(name, declarationSymbolPosition(declaration, name))
				}
			}
		case "N_Recursive":
			for _, declaration := range mod.Recursives {
				if declaration.Syntax != unit {
					continue
				}
				for _, name := range declaration.Names {
					addLocal(name, declarationSymbolPosition(declaration, name))
				}
			}
		case "N_OperatorDefinition", "N_FunctionDefinition", "N_Theorem":
			for _, definition := range mod.Definitions {
				if definition.Syntax != unit {
					continue
				}
				addLocal(definition.Name, definition.DeclarationPosition())
				completed[definition.Name] = true
				completed[mod.Name+"!"+definition.Name] = true
				addSubexpressionReferenceNames(context, definition.Name, definition.Expr)
				if mod.Name != "" {
					addSubexpressionReferenceNames(context, mod.Name+"!"+definition.Name, definition.Expr)
				}
			}
		case "N_Assumption":
			for _, assumption := range mod.Assumptions {
				if assumption.Syntax == unit && assumption.Name != "" {
					addLocal(assumption.Name, assumption.SourcePosition())
					completed[assumption.Name] = true
					completed[mod.Name+"!"+assumption.Name] = true
					addSubexpressionReferenceNames(context, assumption.Name, assumption.Expr)
				}
			}
		case "N_ModuleDefinition", "N_Instance":
			for _, instance := range mod.Instances {
				if instance.Syntax != unit {
					continue
				}
				if instance.Name != "" {
					add(instance.Name, instance.SourcePosition())
					add(instanceNameSentinel(instance.Name), instance.SourcePosition())
				}
				for _, symbol := range semanticInstanceSymbols(instance, spec) {
					add(symbol.name, symbol.sourcePosition())
					completed[symbol.name] = true
				}
				if spec == nil {
					continue
				}
				instancee := spec.Modules[instance.Module]
				if instancee == nil {
					continue
				}
				for _, definition := range instancee.Definitions {
					if definition.Local {
						continue
					}
					if instance.exportsUnqualified() {
						addSubexpressionReferenceNames(context, definition.Name, definition.Expr)
					}
					if instance.Name != "" {
						addSubexpressionReferenceNames(context, instance.qualifier()+"!"+definition.Name, definition.Expr)
					}
				}
			}
		}
	}
	return contexts
}

// Generator.processFunction evaluates all domains before introducing the
// function's temporary recursion symbol. That symbol and the bound variables
// are available only while generating the function body.
func checkDefinitionExpression(definition Definition, context map[string]Position, locals map[string]bool) Diagnostics {
	function, ok := definition.Expr.(*FunctionExpr)
	if !definition.FunctionDef || !ok {
		return checkExpr(definition.Expr, context, locals)
	}
	var diags Diagnostics
	for _, bound := range function.Bounds {
		diags = append(diags, checkExpr(bound.Set, context, locals)...)
	}
	bodyLocals := copyBoolMap(locals)
	for _, bound := range function.Bounds {
		diags = append(diags, checkBoundName(bound.Name, bound.Pos, context, bodyLocals)...)
		bodyLocals[bound.Name] = true
	}
	bodyLocals[definition.Name] = true
	diags = append(diags, checkExpr(function.Body, context, bodyLocals)...)
	return diags
}

// selectorToNode resolves the named symbol before following body selectors.
// Keep only that named prefix when reporting an unavailable prepared selection.
func checkSanySelectedSymbol(expr Expr, selection *sanySelectorSelection, context map[string]Position, locals map[string]bool) Diagnostics {
	identifier := &IdentExpr{Name: selection.name, Pos: expr.Position()}
	if source := sanyExprSource(expr); source != nil && source.Selector != nil {
		selector := &SanySelector{Syntax: source.Selector.Syntax}
		var names []string
		for _, step := range source.Selector.Steps {
			if step.Kind != SanySelectorName {
				break
			}
			selector.Steps = append(selector.Steps, step)
			names = append(names, step.Name)
			if strings.Join(names, "!") == selection.name {
				break
			}
		}
		identifier.Selector = selector
	}
	return checkExpr(identifier, context, locals)
}

// processLetIn follows the LET body's syntax heirs just as generateModule does.
// Neither later operators nor later RECURSIVE declarations are predeclared.
type sanyLetGenerationUnit struct {
	syntax      *SanySyntaxNode
	position    Position
	declaration *Declaration
	definition  *Definition
	instance    *Instance
}

func sanyLetGenerationUnits(expr *LetExpr) []sanyLetGenerationUnit {
	var units []sanyLetGenerationUnit
	for i := range expr.Recursives {
		declaration := &expr.Recursives[i]
		units = append(units, sanyLetGenerationUnit{syntax: declaration.Syntax, position: declaration.Pos, declaration: declaration})
	}
	for i := range expr.Definitions {
		definition := &expr.Definitions[i]
		units = append(units, sanyLetGenerationUnit{syntax: definition.Syntax, position: definition.SourcePosition(), definition: definition})
	}
	for i := range expr.Instances {
		instance := &expr.Instances[i]
		units = append(units, sanyLetGenerationUnit{syntax: instance.Syntax, position: instance.SourcePosition(), instance: instance})
	}
	if expr.Syntax != nil {
		bySyntax := make(map[*SanySyntaxNode]sanyLetGenerationUnit, len(units))
		for _, unit := range units {
			bySyntax[unit.syntax] = unit
		}
		units = nil
		heirs := expr.Syntax.GetHeirs()
		if len(heirs) > 1 {
			for _, syntax := range heirs[1].GetHeirs() {
				if unit, ok := bySyntax[syntax]; ok {
					units = append(units, unit)
				}
			}
		}
	} else {
		// Hand-built native ASTs retain their source positions without syntax.
		sort.SliceStable(units, func(i, j int) bool { return units[i].position.Compare(units[j].position) < 0 })
	}
	return units
}

func checkLetExpression(expr *LetExpr, context map[string]Position, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	diags = append(diags, checkLetRecursiveSections(expr)...)
	units := sanyLetGenerationUnits(expr)
	letLocals := copyBoolMap(locals)
	for _, unit := range units {
		switch {
		case unit.declaration != nil:
			for _, name := range unit.declaration.Names {
				letLocals[name] = true
			}
		case unit.definition != nil:
			definition := *unit.definition
			diags = append(diags, checkDefinitionParams(definition)...)
			diags = append(diags, checkDefinitionParamCollisions(definition, context, letLocals)...)
			bodyLocals := copyBoolMap(letLocals)
			for _, name := range definition.Params {
				bodyLocals[name] = true
			}
			diags = append(diags, checkDefinitionExpression(definition, context, bodyLocals)...)
			letLocals[definition.Name] = true
		case unit.instance != nil:
			instance := unit.instance
			bodyLocals := copyBoolMap(letLocals)
			for _, name := range instance.Params {
				bodyLocals[name] = true
			}
			for _, substitution := range instance.SubstitutionList {
				diags = append(diags, checkExpr(substitution.Expr, context, bodyLocals)...)
			}
			if instance.Name != "" {
				letLocals[instance.Name+"!"] = true
			}
		}
	}
	diags = append(diags, checkExpr(expr.Body, context, letLocals)...)
	return diags
}
