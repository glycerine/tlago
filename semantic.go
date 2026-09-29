package tlago

import (
	"sort"
	"strings"
)

func CheckSpec(spec *Spec) Diagnostics {
	if spec == nil {
		return Diagnostics{errorAt(Position{}, "E1300", "nil spec")}
	}
	var diags Diagnostics
	for _, mod := range spec.Modules {
		diags = append(diags, checkModule(mod, spec)...)
	}
	return diags
}

func checkModule(mod *Module, spec *Spec) Diagnostics {
	var diags Diagnostics
	defined := map[string]Position{}
	declKinds := map[string]DeclarationKind{}
	arities := map[string]int{}
	functionArities := map[string]int{}
	operatorParamSpecs := map[string][]operatorParamSpec{}
	extendedSymbols := map[string]importedSymbol{}
	diags = append(diags, checkPlusCalChecksumWarnings(mod)...)
	diags = append(diags, checkNestedStandardModuleConflicts(mod)...)
	addName := func(name string, pos Position) {
		if builtinIdentifiers[name] && !isEmbeddedStandardModule(mod) {
			diags = append(diags, errorAt(pos, "E1301", "cannot redefine built-in symbol %s", name))
			return
		}
		if prev, ok := defined[name]; ok {
			diags = append(diags, errorAt(pos, "E1301", "duplicate declaration or definition %s; first declared at %s", name, prev))
			return
		}
		defined[name] = pos
	}
	importInheritedModule := func(depMod *Module) {
		if depMod == nil {
			return
		}
		for _, d := range depMod.Declarations {
			for _, name := range d.Names {
				diags = append(diags, checkImportedSymbolKind(name, d.Kind, d.Pos, declKinds)...)
				diags = append(diags, checkImportedSymbolAmbiguity(name, d.Kind, d.Pos, depMod.Name, extendedSymbols, "W4800")...)
				if _, exists := defined[name]; !exists {
					defined[name] = d.Pos
				}
				if _, exists := declKinds[name]; !exists {
					declKinds[name] = d.Kind
				}
				if d.Kind == ConstantDecl {
					if arity, ok := declarationArity(d, name); ok {
						if _, exists := arities[name]; !exists {
							arities[name] = arity
						}
					}
				}
			}
		}
		for _, def := range depMod.Definitions {
			if def.Local {
				continue
			}
			diags = append(diags, checkImportedSymbolKind(def.Name, OperatorDecl, def.Pos, declKinds)...)
			diags = append(diags, checkImportedSymbolAmbiguity(def.Name, OperatorDecl, def.Pos, depMod.Name, extendedSymbols, "W4800")...)
			if _, exists := defined[def.Name]; !exists {
				defined[def.Name] = def.Pos
			}
			if _, exists := arities[def.Name]; !exists {
				arities[def.Name] = len(def.Params)
			}
			if specs, ok := definitionOperatorParamSpecsForModule(depMod.Name, def); ok {
				if _, exists := operatorParamSpecs[def.Name]; !exists {
					operatorParamSpecs[def.Name] = specs
				}
			}
			if arity, ok := definitionFunctionArity(def); ok {
				if _, exists := functionArities[def.Name]; !exists {
					functionArities[def.Name] = arity
				}
			}
			if _, exists := declKinds[def.Name]; !exists {
				declKinds[def.Name] = OperatorDecl
			}
		}
		for _, assumption := range depMod.Assumptions {
			if assumption.Name == "" {
				continue
			}
			pos := assumption.SourcePosition()
			diags = append(diags, checkImportedSymbolKind(assumption.Name, OperatorDecl, pos, declKinds)...)
			diags = append(diags, checkImportedSymbolAmbiguity(assumption.Name, OperatorDecl, pos, depMod.Name, extendedSymbols, "W4800")...)
			if _, exists := defined[assumption.Name]; !exists {
				defined[assumption.Name] = pos
			}
			if _, exists := arities[assumption.Name]; !exists {
				arities[assumption.Name] = 0
			}
			if _, exists := declKinds[assumption.Name]; !exists {
				declKinds[assumption.Name] = OperatorDecl
			}
		}
		for _, export := range syntheticStandardExports(depMod.Name, depMod.Pos) {
			diags = append(diags, checkImportedSymbolKind(export.Name, export.Kind, export.Pos, declKinds)...)
			diags = append(diags, checkImportedSymbolAmbiguity(export.Name, export.Kind, export.Pos, depMod.Name, extendedSymbols, "W4800")...)
			if _, exists := defined[export.Name]; !exists {
				defined[export.Name] = export.Pos
			}
			if _, exists := declKinds[export.Name]; !exists {
				declKinds[export.Name] = export.Kind
			}
			if _, exists := arities[export.Name]; !exists {
				arities[export.Name] = export.Arity
			}
		}
	}
	for _, dep := range mod.Extends {
		if depMod := spec.Modules[dep]; depMod != nil {
			for _, inherited := range transitiveExtendedModules(spec, depMod, map[string]bool{depMod.Name: true}) {
				importInheritedModule(inherited)
			}
			for _, d := range depMod.Declarations {
				for _, name := range d.Names {
					diags = append(diags, checkImportedSymbolKind(name, d.Kind, d.Pos, declKinds)...)
					diags = append(diags, checkImportedSymbolAmbiguity(name, d.Kind, d.Pos, depMod.Name, extendedSymbols, "W4800")...)
					if _, exists := defined[name]; !exists {
						defined[name] = d.Pos
					}
					qualified := depMod.Name + "!" + name
					if depMod.Name != "" {
						if _, exists := defined[qualified]; !exists {
							defined[qualified] = d.Pos
						}
					}
					if _, exists := declKinds[name]; !exists {
						declKinds[name] = d.Kind
					}
					if depMod.Name != "" {
						if _, exists := declKinds[qualified]; !exists {
							declKinds[qualified] = d.Kind
						}
					}
					if d.Kind == ConstantDecl {
						if arity, ok := declarationArity(d, name); ok {
							if _, exists := arities[name]; !exists {
								arities[name] = arity
							}
							if depMod.Name != "" {
								if _, exists := arities[qualified]; !exists {
									arities[qualified] = arity
								}
							}
						}
					}
				}
			}
			for _, def := range depMod.Definitions {
				if def.Local {
					continue
				}
				diags = append(diags, checkImportedSymbolKind(def.Name, OperatorDecl, def.Pos, declKinds)...)
				diags = append(diags, checkImportedSymbolAmbiguity(def.Name, OperatorDecl, def.Pos, depMod.Name, extendedSymbols, "W4800")...)
				if _, exists := defined[def.Name]; !exists {
					defined[def.Name] = def.Pos
				}
				qualified := depMod.Name + "!" + def.Name
				if depMod.Name != "" {
					if _, exists := defined[qualified]; !exists {
						defined[qualified] = def.Pos
					}
				}
				if _, exists := arities[def.Name]; !exists {
					arities[def.Name] = len(def.Params)
				}
				if specs, ok := definitionOperatorParamSpecsForModule(depMod.Name, def); ok {
					if _, exists := operatorParamSpecs[def.Name]; !exists {
						operatorParamSpecs[def.Name] = specs
					}
				}
				if arity, ok := definitionFunctionArity(def); ok {
					if _, exists := functionArities[def.Name]; !exists {
						functionArities[def.Name] = arity
					}
				}
				if _, exists := declKinds[def.Name]; !exists {
					declKinds[def.Name] = OperatorDecl
				}
				if depMod.Name != "" {
					if _, exists := arities[qualified]; !exists {
						arities[qualified] = len(def.Params)
					}
				}
			}
			for _, assumption := range depMod.Assumptions {
				if assumption.Name == "" {
					continue
				}
				pos := assumption.SourcePosition()
				diags = append(diags, checkImportedSymbolKind(assumption.Name, OperatorDecl, pos, declKinds)...)
				diags = append(diags, checkImportedSymbolAmbiguity(assumption.Name, OperatorDecl, pos, depMod.Name, extendedSymbols, "W4800")...)
				if _, exists := defined[assumption.Name]; !exists {
					defined[assumption.Name] = pos
				}
				qualified := depMod.Name + "!" + assumption.Name
				if depMod.Name != "" {
					if _, exists := defined[qualified]; !exists {
						defined[qualified] = pos
					}
				}
				if _, exists := arities[assumption.Name]; !exists {
					arities[assumption.Name] = 0
				}
				if _, exists := declKinds[assumption.Name]; !exists {
					declKinds[assumption.Name] = OperatorDecl
				}
				if depMod.Name != "" {
					if _, exists := arities[qualified]; !exists {
						arities[qualified] = 0
					}
					if _, exists := declKinds[qualified]; !exists {
						declKinds[qualified] = OperatorDecl
					}
				}
			}
			for _, export := range syntheticStandardExports(depMod.Name, depMod.Pos) {
				diags = append(diags, checkImportedSymbolKind(export.Name, export.Kind, export.Pos, declKinds)...)
				diags = append(diags, checkImportedSymbolAmbiguity(export.Name, export.Kind, export.Pos, depMod.Name, extendedSymbols, "W4800")...)
				if _, exists := defined[export.Name]; !exists {
					defined[export.Name] = export.Pos
				}
				if _, exists := declKinds[export.Name]; !exists {
					declKinds[export.Name] = export.Kind
				}
				if _, exists := arities[export.Name]; !exists {
					arities[export.Name] = export.Arity
				}
			}
			for _, inst := range depMod.Instances {
				if inst.Local {
					continue
				}
				instMod := spec.Modules[inst.Module]
				if instMod == nil {
					continue
				}
				qualifier := inst.qualifier()
				exportUnqualified := inst.exportsUnqualified()
				for _, d := range instMod.Declarations {
					for _, name := range d.Names {
						if exportUnqualified {
							if _, exists := defined[name]; !exists {
								defined[name] = d.Pos
							}
							if _, exists := declKinds[name]; !exists {
								declKinds[name] = d.Kind
							}
						}
						qualified := qualifier + "!" + name
						if qualifier != "" {
							if _, exists := defined[qualified]; !exists {
								defined[qualified] = d.Pos
							}
							if _, exists := declKinds[qualified]; !exists {
								declKinds[qualified] = d.Kind
							}
						}
						if d.Kind == ConstantDecl {
							if arity, ok := declarationArity(d, name); ok {
								if exportUnqualified {
									if _, exists := arities[name]; !exists {
										arities[name] = arity
									}
								}
								if qualifier != "" {
									if _, exists := arities[qualified]; !exists {
										arities[qualified] = arity
									}
								}
							}
						}
					}
				}
				for _, def := range instMod.Definitions {
					if def.Local {
						continue
					}
					if exportUnqualified {
						if _, exists := defined[def.Name]; !exists {
							defined[def.Name] = def.Pos
						}
						if _, exists := arities[def.Name]; !exists {
							arities[def.Name] = len(def.Params)
						}
					}
					qualified := qualifier + "!" + def.Name
					if qualifier != "" {
						if _, exists := defined[qualified]; !exists {
							defined[qualified] = def.Pos
						}
						if _, exists := arities[qualified]; !exists {
							arities[qualified] = len(def.Params)
						}
						if specs, ok := definitionOperatorParamSpecsForModule(depMod.Name, def); ok {
							if _, exists := operatorParamSpecs[qualified]; !exists {
								operatorParamSpecs[qualified] = specs
							}
						}
						if arity, ok := definitionFunctionArity(def); ok {
							if _, exists := functionArities[qualified]; !exists {
								functionArities[qualified] = arity
							}
						}
					}
				}
			}
		}
	}
	instanceSymbols := map[string]importedSymbol{}
	localSymbols := moduleOwnSymbols(mod)
	for _, inst := range mod.Instances {
		diags = append(diags, addInstanceSymbols(inst, spec, defined, declKinds, arities, operatorParamSpecs, instanceSymbols, localSymbols)...)
	}
	for _, d := range mod.Declarations {
		seenInDecl := map[string]bool{}
		for _, name := range d.Names {
			if seenInDecl[name] {
				diags = append(diags, errorAt(d.Pos, "E1301", "duplicate declaration %s", name))
				continue
			}
			seenInDecl[name] = true
			addName(name, d.Pos)
			declKinds[name] = d.Kind
			if mod.Name != "" {
				qualified := mod.Name + "!" + name
				defined[qualified] = d.Pos
				declKinds[qualified] = d.Kind
			}
			if d.Kind == ConstantDecl {
				if arity, ok := declarationArity(d, name); ok {
					arities[name] = arity
					if mod.Name != "" {
						arities[mod.Name+"!"+name] = arity
					}
				}
			}
		}
	}
	recursiveArities := map[string]int{}
	recursivePositions := map[string]Position{}
	satisfiedRecursive := map[string]bool{}
	for _, d := range mod.Recursives {
		seenInDecl := map[string]bool{}
		for _, name := range d.Names {
			if seenInDecl[name] {
				diags = append(diags, errorAt(d.Pos, "E1301", "duplicate recursive declaration %s", name))
				continue
			}
			seenInDecl[name] = true
			if prev, ok := defined[name]; ok {
				diags = append(diags, errorAt(d.Pos, "E1301", "recursive declaration %s conflicts with declaration or definition at %s", name, prev))
				continue
			}
			defined[name] = d.Pos
			declKinds[name] = RecursiveDecl
			arity, ok := declarationArity(d, name)
			if !ok {
				arity = 0
			}
			recursiveArities[name] = arity
			recursivePositions[name] = d.Pos
			arities[name] = arity
			if mod.Name != "" {
				qualified := mod.Name + "!" + name
				defined[qualified] = d.Pos
				declKinds[qualified] = RecursiveDecl
				arities[qualified] = arity
			}
		}
	}
	for _, def := range mod.Definitions {
		diags = append(diags, checkDefinitionParams(def)...)
		diags = append(diags, checkDefinitionParamCollisions(def, defined, nil)...)
		if want, recursive := recursiveArities[def.Name]; recursive {
			satisfiedRecursive[def.Name] = true
			if got := len(def.Params); got != want {
				diags = append(diags, errorAt(def.Pos, "E1307", "recursive definition %s arity mismatch: got %d, want %d", def.Name, got, want))
			}
			if exprContainsPrime(def.Expr) {
				diags = append(diags, errorAt(def.Pos, "E1320", "recursive definition %s cannot contain prime", def.Name))
			}
		} else {
			if !definitionSatisfiesSymbolicConstantDeclaration(def, declKinds, arities) {
				addName(def.Name, def.Pos)
			}
		}
		arities[def.Name] = len(def.Params)
		if specs, ok := definitionOperatorParamSpecs(def); ok {
			operatorParamSpecs[def.Name] = specs
		}
		if arity, ok := definitionFunctionArity(def); ok {
			functionArities[def.Name] = arity
		}
		if mod.Name != "" {
			qualified := mod.Name + "!" + def.Name
			defined[qualified] = def.Pos
			arities[qualified] = len(def.Params)
			if specs, ok := definitionOperatorParamSpecs(def); ok {
				operatorParamSpecs[qualified] = specs
			}
			if arity, ok := definitionFunctionArity(def); ok {
				functionArities[qualified] = arity
			}
		}
	}
	for _, assumption := range mod.Assumptions {
		if assumption.Name == "" {
			continue
		}
		pos := assumption.SourcePosition()
		addName(assumption.Name, pos)
		arities[assumption.Name] = 0
		declKinds[assumption.Name] = OperatorDecl
		if mod.Name != "" {
			qualified := mod.Name + "!" + assumption.Name
			defined[qualified] = pos
			arities[qualified] = 0
			declKinds[qualified] = OperatorDecl
		}
	}
	diags = append(diags, checkModuleRecursiveSections(mod)...)
	for name, pos := range recursivePositions {
		if !satisfiedRecursive[name] {
			diags = append(diags, errorAt(pos, "E1308", "recursive declaration %s has no definition", name))
		}
	}
	assumeProveDefs := assumeProveDefinitionNames(mod.Definitions)
	theoremLikeDefs := theoremLikeDefinitionNames(mod.Definitions)
	addNamedAssumptions(theoremLikeDefs, mod.Assumptions)
	proofStepNames := proofStepNameSet(mod.Proofs)
	defExprPositions := definitionExpressionPositions(mod.Definitions)
	assumeProveExprPositions := assumeProveDefinitionExpressionPositions(mod.Definitions)
	for _, inst := range mod.Instances {
		diags = append(diags, checkInstanceSubstitutions(mod, inst, spec, defined, declKinds, arities, operatorParamSpecs)...)
	}
	for _, ref := range mod.ProofRefs {
		diags = append(diags, checkProofRef(ref, defined)...)
		diags = append(diags, checkHideRef(ref, theoremLikeDefs, proofStepNames)...)
	}
	for _, proof := range mod.Proofs {
		diags = append(diags, checkProofSummary(proof, declKinds)...)
	}
	for _, assumption := range mod.Assumptions {
		expr := assumption.Expr
		if expr == nil {
			continue
		}
		if !defExprPositions[positionKey(expr.Position())] {
			diags = append(diags, checkLabels(expr, labelCheckContext{})...)
		}
		diags = append(diags, checkExpr(expr, defined, nil)...)
		diags = append(diags, checkCallArity(expr, arities, operatorParamSpecs, nil)...)
		diags = append(diags, checkOperatorArgumentKinds(expr, operatorParamSpecs, arities, nil)...)
		diags = append(diags, checkFunctionArity(expr, functionArities, nil)...)
		if !assumeProveExprPositions[positionKey(expr.Position())] {
			diags = append(diags, checkLevelComposition(expr, declKinds, nil)...)
		}
		diags = append(diags, checkPrimedConstants(expr, declKinds, nil)...)
		diags = append(diags, checkAssumptionConstantLevel(expr, declKinds)...)
	}
	for _, theorem := range mod.Theorems {
		expr := theorem.Expr
		if expr == nil {
			continue
		}
		if !defExprPositions[positionKey(expr.Position())] {
			diags = append(diags, checkLabels(expr, labelCheckContext{})...)
		}
		diags = append(diags, checkExpr(expr, defined, nil)...)
		diags = append(diags, checkCallArity(expr, arities, operatorParamSpecs, nil)...)
		diags = append(diags, checkOperatorArgumentKinds(expr, operatorParamSpecs, arities, nil)...)
		diags = append(diags, checkFunctionArity(expr, functionArities, nil)...)
		if !assumeProveExprPositions[positionKey(expr.Position())] {
			diags = append(diags, checkLevelComposition(expr, declKinds, nil)...)
		}
		diags = append(diags, checkPrimedConstants(expr, declKinds, nil)...)
	}
	for _, def := range mod.Definitions {
		locals := map[string]bool{}
		for _, param := range def.Params {
			locals[param] = true
		}
		defArities := definitionBodyArities(arities, def)
		diags = append(diags, checkExpr(def.Expr, defined, locals)...)
		diags = append(diags, checkCallArity(def.Expr, defArities, operatorParamSpecs, locals)...)
		diags = append(diags, checkOperatorArgumentKinds(def.Expr, operatorParamSpecs, defArities, locals)...)
		diags = append(diags, checkFunctionArity(def.Expr, functionArities, locals)...)
		diags = append(diags, checkLabels(def.Expr, labelCheckContext{allowed: true})...)
		if !def.AssumeProve {
			diags = append(diags, checkAssumeProveDefinitionUse(def.Expr, assumeProveDefs, locals)...)
		}
		if !def.AssumeProve {
			diags = append(diags, checkLevelComposition(def.Expr, declKinds, locals)...)
		}
		diags = append(diags, checkPrimedConstants(def.Expr, declKinds, locals)...)
	}
	return diags
}

func transitiveExtendedModules(spec *Spec, mod *Module, seen map[string]bool) []*Module {
	if spec == nil || mod == nil {
		return nil
	}
	var modules []*Module
	for _, dep := range mod.Extends {
		if seen[dep] {
			continue
		}
		seen[dep] = true
		depMod := spec.Modules[dep]
		if depMod == nil {
			continue
		}
		modules = append(modules, depMod)
		modules = append(modules, transitiveExtendedModules(spec, depMod, seen)...)
	}
	return modules
}

type recursiveSectionItem struct {
	kind string
	pos  Position
}

func checkModuleRecursiveSections(mod *Module) Diagnostics {
	var diags Diagnostics
	if mod == nil || len(mod.Recursives) == 0 {
		return nil
	}
	items := recursiveSectionItems(mod)
	for _, decl := range mod.Recursives {
		for _, name := range decl.Names {
			def, ok := firstDefinitionAfter(mod.Definitions, name, decl.Pos)
			if !ok {
				continue
			}
			for _, item := range items {
				if positionBetween(item.pos, decl.Pos, def.Pos) {
					diags = append(diags, errorAt(item.pos, "E1322", "%s may not appear within a recursive definition section", item.kind))
				}
			}
		}
	}
	return diags
}

func recursiveSectionItems(mod *Module) []recursiveSectionItem {
	var items []recursiveSectionItem
	for _, decl := range mod.Declarations {
		kind := "A declaration"
		if decl.Kind == VariableDecl {
			kind = "A VARIABLE declaration"
		}
		items = append(items, recursiveSectionItem{kind: kind, pos: decl.Pos})
	}
	for _, assumption := range mod.Assumptions {
		if assumption.Expr != nil {
			items = append(items, recursiveSectionItem{kind: "An ASSUME", pos: assumption.Position()})
		}
	}
	for _, theorem := range mod.Theorems {
		if theorem.Expr != nil {
			items = append(items, recursiveSectionItem{kind: "A THEOREM", pos: theorem.Position()})
		}
	}
	for _, ref := range mod.ProofRefs {
		items = append(items, recursiveSectionItem{kind: "A USE or HIDE", pos: ref.Pos})
	}
	for _, nested := range mod.Nested {
		if nested != nil {
			items = append(items, recursiveSectionItem{kind: "A MODULE", pos: nested.Pos})
		}
	}
	return items
}

func firstDefinitionAfter(defs []Definition, name string, after Position) (Definition, bool) {
	var out Definition
	found := false
	for _, def := range defs {
		if def.Name != name || !positionBefore(after, def.Pos) {
			continue
		}
		if !found || positionBefore(def.Pos, out.Pos) {
			out = def
			found = true
		}
	}
	return out, found
}

func positionBetween(pos, start, end Position) bool {
	return positionBefore(start, pos) && positionBefore(pos, end)
}

func checkLetRecursiveSections(expr *LetExpr) Diagnostics {
	var diags Diagnostics
	if expr == nil || len(expr.Recursives) == 0 {
		return nil
	}
	recursiveArities := map[string]int{}
	recursivePositions := map[string]Position{}
	for _, decl := range expr.Recursives {
		seenInDecl := map[string]bool{}
		for _, name := range decl.Names {
			if seenInDecl[name] {
				diags = append(diags, errorAt(decl.Pos, "E1301", "duplicate recursive declaration %s", name))
				continue
			}
			seenInDecl[name] = true
			arity, ok := declarationArity(decl, name)
			if !ok {
				arity = 0
			}
			recursiveArities[name] = arity
			recursivePositions[name] = decl.Pos
		}
	}
	for name, pos := range recursivePositions {
		def, ok := firstDefinitionAfter(expr.Definitions, name, pos)
		if !ok {
			diags = append(diags, errorAt(pos, "E1308", "recursive declaration %s has no definition", name))
			continue
		}
		if got, want := len(def.Params), recursiveArities[name]; got != want {
			diags = append(diags, errorAt(def.Pos, "E1307", "recursive definition %s arity mismatch: got %d, want %d", def.Name, got, want))
		}
		if exprContainsPrime(def.Expr) {
			diags = append(diags, errorAt(def.Pos, "E1320", "recursive definition %s cannot contain prime", def.Name))
		}
	}
	return diags
}

func letScopeLocals(locals map[string]bool, expr *LetExpr) map[string]bool {
	letLocals := copyBoolMap(locals)
	if expr == nil {
		return letLocals
	}
	for _, decl := range expr.Recursives {
		for _, name := range decl.Names {
			letLocals[name] = true
		}
	}
	for _, def := range expr.Definitions {
		letLocals[def.Name] = true
	}
	for _, inst := range expr.Instances {
		if inst.Name != "" {
			letLocals[inst.Name+"!"] = true
		}
	}
	return letLocals
}

func letDefinitionBodyLocals(letLocals map[string]bool, def Definition) map[string]bool {
	defLocals := copyBoolMap(letLocals)
	if _, isFunctionDefinition := definitionFunctionArity(def); !isFunctionDefinition {
		delete(defLocals, def.Name)
	}
	for _, param := range def.Params {
		defLocals[param] = true
	}
	return defLocals
}

func definitionExpressionPositions(defs []Definition) map[string]bool {
	positions := map[string]bool{}
	for _, def := range defs {
		if def.Expr != nil {
			positions[positionKey(def.Expr.Position())] = true
		}
	}
	return positions
}

func assumeProveDefinitionExpressionPositions(defs []Definition) map[string]bool {
	positions := map[string]bool{}
	for _, def := range defs {
		if def.AssumeProve && def.Expr != nil {
			positions[positionKey(def.Expr.Position())] = true
		}
	}
	return positions
}

func assumeProveDefinitionNames(defs []Definition) map[string]bool {
	names := map[string]bool{}
	for _, def := range defs {
		if def.AssumeProve && def.Name != "" {
			names[def.Name] = true
		}
	}
	return names
}

func theoremLikeDefinitionNames(defs []Definition) map[string]bool {
	names := map[string]bool{}
	for _, def := range defs {
		if def.TheoremLike && def.Name != "" {
			names[def.Name] = true
		}
	}
	return names
}

func addNamedAssumptions(names map[string]bool, assumptions []NamedExpr) {
	for _, assumption := range assumptions {
		if assumption.Name != "" {
			names[assumption.Name] = true
		}
	}
}

func proofStepNameSet(proofs []ProofSummary) map[string]bool {
	names := map[string]bool{}
	for _, proof := range proofs {
		for _, step := range proof.Steps {
			if step.Name != "" {
				names[step.Name] = true
			}
		}
	}
	return names
}

func positionKey(pos Position) string {
	return pos.String()
}

type labelCheckContext struct {
	allowed  bool
	inExcept bool
	bound    []string
}

func (ctx labelCheckContext) withBound(name string) labelCheckContext {
	if name == "" {
		return ctx
	}
	next := ctx
	next.bound = append(append([]string(nil), ctx.bound...), name)
	return next
}

func (ctx labelCheckContext) withBounds(bounds []BoundVar) labelCheckContext {
	next := ctx
	for _, bound := range bounds {
		next = next.withBound(bound.Name)
	}
	return next
}

func (ctx labelCheckContext) insideExcept() labelCheckContext {
	next := ctx
	next.inExcept = true
	return next
}

func (ctx labelCheckContext) resetLabelBoundScope() labelCheckContext {
	next := ctx
	next.bound = nil
	return next
}

func checkLabels(expr Expr, ctx labelCheckContext) Diagnostics {
	if expr == nil {
		return nil
	}
	var diags Diagnostics
	switch e := expr.(type) {
	case *UnaryExpr:
		diags = append(diags, checkLabels(e.Expr, ctx)...)
	case *BinaryExpr:
		diags = append(diags, checkDuplicateSiblingLabels([]Expr{e.Left, e.Right})...)
		diags = append(diags, checkLabels(e.Left, ctx)...)
		diags = append(diags, checkLabels(e.Right, ctx)...)
	case *CallExpr:
		diags = append(diags, checkLabels(e.Callee, ctx)...)
		for _, arg := range e.Args {
			diags = append(diags, checkLabels(arg, ctx)...)
		}
	case *IfExpr:
		diags = append(diags, checkLabels(e.Cond, ctx)...)
		diags = append(diags, checkLabels(e.Then, ctx)...)
		diags = append(diags, checkLabels(e.Else, ctx)...)
	case *LetExpr:
		for _, def := range e.Definitions {
			diags = append(diags, checkLabels(def.Expr, labelCheckContext{allowed: true})...)
		}
		diags = append(diags, checkLabels(e.Body, ctx)...)
	case *QuantifierExpr:
		diags = append(diags, checkLabels(e.Set, ctx)...)
		diags = append(diags, checkLabels(e.Body, ctx.withBound(e.Var))...)
	case *CaseExpr:
		var values []Expr
		for _, arm := range e.Arms {
			diags = append(diags, checkLabels(arm.Test, ctx)...)
			values = append(values, arm.Value)
		}
		if e.Other != nil {
			values = append(values, e.Other)
		}
		diags = append(diags, checkDuplicateSiblingLabels(values)...)
		for _, value := range values {
			diags = append(diags, checkLabels(value, ctx)...)
		}
	case *ChooseExpr:
		diags = append(diags, checkLabels(e.Set, ctx)...)
		diags = append(diags, checkLabels(e.Body, ctx.withBound(e.Var))...)
	case *TupleExpr:
		diags = append(diags, checkDuplicateSiblingLabels(e.Elems)...)
		for _, elem := range e.Elems {
			diags = append(diags, checkLabels(elem, ctx)...)
		}
	case *SetExpr:
		diags = append(diags, checkDuplicateSiblingLabels(e.Elems)...)
		for _, elem := range e.Elems {
			diags = append(diags, checkLabels(elem, ctx)...)
		}
	case *RecordExpr:
		values := make([]Expr, 0, len(e.Fields))
		for _, field := range e.Fields {
			values = append(values, field.Value)
		}
		diags = append(diags, checkDuplicateSiblingLabels(values)...)
		for _, value := range values {
			diags = append(diags, checkLabels(value, ctx)...)
		}
	case *RecordComponentExpr:
		diags = append(diags, checkLabels(e.Record, ctx)...)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkLabels(field.Set, ctx)...)
		}
	case *FunctionExpr:
		for _, bound := range e.Bounds {
			diags = append(diags, checkLabels(bound.Set, ctx)...)
		}
		diags = append(diags, checkLabels(e.Body, ctx.withBounds(e.Bounds))...)
	case *FunctionAppExpr:
		diags = append(diags, checkLabels(e.Function, ctx)...)
		for _, arg := range e.Args {
			diags = append(diags, checkLabels(arg, ctx)...)
		}
	case *ExceptExpr:
		diags = append(diags, checkLabels(e.Base, ctx)...)
		exceptCtx := ctx.insideExcept()
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					diags = append(diags, checkLabels(index, exceptCtx)...)
				}
			}
			diags = append(diags, checkLabels(spec.Value, exceptCtx)...)
		}
	case *LabelExpr:
		if !ctx.allowed {
			diags = append(diags, errorAt(e.Pos, "E4333", "label %s is not in definition or proof step", e.Name))
		}
		if ctx.inExcept {
			diags = append(diags, errorAt(e.Pos, "E4335", "label %s is not allowed inside EXCEPT", e.Name))
		}
		diags = append(diags, checkLabelParameters(e, ctx.bound)...)
		diags = append(diags, checkLabels(e.Body, ctx.resetLabelBoundScope())...)
	case *ActionExpr:
		diags = append(diags, checkLabels(e.Action, ctx)...)
		diags = append(diags, checkLabels(e.Subscript, ctx)...)
	case *FairnessExpr:
		diags = append(diags, checkLabels(e.Subscript, ctx)...)
		diags = append(diags, checkLabels(e.Action, ctx)...)
	case *FunctionSetExpr:
		diags = append(diags, checkLabels(e.Domain, ctx)...)
		diags = append(diags, checkLabels(e.Range, ctx)...)
	case *SetComprehensionExpr:
		for _, bound := range e.Bounds {
			diags = append(diags, checkLabels(bound.Set, ctx)...)
		}
		bodyCtx := ctx.withBounds(e.Bounds)
		diags = append(diags, checkLabels(e.Element, bodyCtx)...)
		if e.Predicate != nil {
			diags = append(diags, checkLabels(e.Predicate, bodyCtx)...)
		}
	}
	return diags
}

func checkLabelParameters(label *LabelExpr, bound []string) Diagnostics {
	var diags Diagnostics
	seen := map[string]bool{}
	for _, param := range label.Params {
		if seen[param] {
			diags = append(diags, errorAt(label.Pos, "E4330", "repeated label parameter %s in label %s", param, label.Name))
			continue
		}
		seen[param] = true
	}
	required := map[string]bool{}
	for _, name := range bound {
		required[name] = true
		if !seen[name] {
			diags = append(diags, errorAt(label.Pos, "E4331", "label %s must contain bound parameter %s", label.Name, name))
		}
	}
	for _, param := range label.Params {
		if !required[param] {
			diags = append(diags, errorAt(label.Pos, "E4332", "unnecessary label parameter %s in label %s", param, label.Name))
		}
	}
	return diags
}

func checkDuplicateSiblingLabels(exprs []Expr) Diagnostics {
	var diags Diagnostics
	seen := map[string]Position{}
	for _, expr := range exprs {
		label, ok := expr.(*LabelExpr)
		if !ok || label.Name == "" {
			continue
		}
		if _, exists := seen[label.Name]; exists {
			diags = append(diags, errorAt(label.Pos, "E4336", "Duplicate label %s", label.Name))
			continue
		}
		seen[label.Name] = label.Pos
	}
	return diags
}

func isEmbeddedStandardModule(mod *Module) bool {
	if mod == nil {
		return false
	}
	source, ok := standardModules[mod.Name]
	return ok && mod.SourcePath == mod.Name+".tla" && strings.TrimSpace(mod.Source) == strings.TrimSpace(source)
}

func checkNestedStandardModuleConflicts(mod *Module) Diagnostics {
	if mod == nil {
		return nil
	}
	importsStandard := map[string]bool{}
	for _, name := range mod.Extends {
		if _, ok := standardModules[name]; ok {
			importsStandard[name] = true
		}
	}
	for _, inst := range mod.Instances {
		if _, ok := standardModules[inst.Module]; ok {
			importsStandard[inst.Module] = true
		}
	}
	var diags Diagnostics
	for _, nested := range mod.Nested {
		if nested != nil && importsStandard[nested.Name] {
			diags = append(diags, errorAt(nested.Pos, "E1315", "distinct modules with name %s are imported into module %s", nested.Name, mod.Name))
		}
	}
	return diags
}

func checkImportedSymbolKind(name string, kind DeclarationKind, pos Position, declKinds map[string]DeclarationKind) Diagnostics {
	if name == "" {
		return nil
	}
	if prev, ok := declKinds[name]; ok && prev != kind {
		return Diagnostics{errorAt(pos, "E1316", "conflicting imported symbol %s has kinds %s and %s", name, prev, kind)}
	}
	return nil
}

type importedSymbol struct {
	kind   DeclarationKind
	pos    Position
	source string
}

func checkImportedSymbolAmbiguity(name string, kind DeclarationKind, pos Position, source string, seen map[string]importedSymbol, code string) Diagnostics {
	if name == "" || source == "" || seen == nil {
		return nil
	}
	if prev, ok := seen[name]; ok {
		if prev.kind == kind && prev.source != source {
			return Diagnostics{warningAt(pos, code, "the %s symbol %s from module %s conflicts with the same kind of imported symbol from module %s at %s; the first import is used", kind, name, source, prev.source, prev.pos)}
		}
		return nil
	}
	seen[name] = importedSymbol{kind: kind, pos: pos, source: source}
	return nil
}

type localSymbol struct {
	kind DeclarationKind
	pos  Position
}

func moduleOwnSymbols(mod *Module) map[string]localSymbol {
	symbols := map[string]localSymbol{}
	if mod == nil {
		return symbols
	}
	for _, decl := range mod.Declarations {
		for _, name := range decl.Names {
			if name != "" {
				symbols[name] = localSymbol{kind: decl.Kind, pos: decl.Pos}
			}
		}
	}
	for _, decl := range mod.Recursives {
		for _, name := range decl.Names {
			if name != "" {
				symbols[name] = localSymbol{kind: RecursiveDecl, pos: decl.Pos}
			}
		}
	}
	for _, def := range mod.Definitions {
		if def.Name != "" {
			symbols[def.Name] = localSymbol{kind: OperatorDecl, pos: def.Pos}
		}
	}
	for _, assumption := range mod.Assumptions {
		if assumption.Name != "" {
			symbols[assumption.Name] = localSymbol{kind: OperatorDecl, pos: assumption.SourcePosition()}
		}
	}
	return symbols
}

func checkInstanceSymbolAmbiguity(name string, kind DeclarationKind, pos Position, source string, seen map[string]importedSymbol) Diagnostics {
	if name == "" || source == "" || seen == nil {
		return nil
	}
	if prev, ok := seen[name]; ok {
		if prev.source != source {
			return Diagnostics{warningAt(pos, "W4801", "the INSTANCE export %s from module %s conflicts with an INSTANCE export from module %s at %s; the first import is used", name, source, prev.source, prev.pos)}
		}
		return nil
	}
	seen[name] = importedSymbol{kind: kind, pos: pos, source: source}
	return nil
}

func checkInstanceLocalShadow(name string, pos Position, source string, localSymbols map[string]localSymbol) Diagnostics {
	if name == "" || source == "" || localSymbols == nil {
		return nil
	}
	if local, ok := localSymbols[name]; ok {
		return Diagnostics{warningAt(pos, "W4801", "the INSTANCE export %s from module %s conflicts with a local symbol at %s; the local symbol is used", name, source, local.pos)}
	}
	return nil
}

type syntheticExport struct {
	Name  string
	Kind  DeclarationKind
	Arity int
	Pos   Position
}

func syntheticStandardExports(moduleName string, pos Position) []syntheticExport {
	switch moduleName {
	case "Naturals":
		return []syntheticExport{
			{Name: "+", Kind: OperatorDecl, Arity: 2, Pos: pos},
		}
	default:
		return nil
	}
}

func definitionSatisfiesSymbolicConstantDeclaration(def Definition, declKinds map[string]DeclarationKind, arities map[string]int) bool {
	if def.Name == "" || isIdentifierName(def.Name) {
		return false
	}
	if declKinds[def.Name] != ConstantDecl {
		return false
	}
	return arities[def.Name] == len(def.Params)
}

func isIdentifierName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				continue
			}
			return false
		}
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

type semanticExportedSymbol struct {
	name              string
	kind              DeclarationKind
	pos               Position
	arity             int
	hasArity          bool
	operatorParams    []operatorParamSpec
	hasOperatorParams bool
	unqualified       bool
}

func semanticInstanceSymbols(inst Instance, spec *Spec) []semanticExportedSymbol {
	return semanticInstanceSymbolsWithVisiting(inst, spec, map[string]bool{})
}

func semanticInstanceSymbolsWithVisiting(inst Instance, spec *Spec, visiting map[string]bool) []semanticExportedSymbol {
	if spec == nil {
		return nil
	}
	instMod := spec.Modules[inst.Module]
	if instMod == nil {
		return nil
	}
	exports := semanticModuleExports(instMod, spec, visiting)
	qualifier := inst.qualifier()
	exportUnqualified := inst.exportsUnqualified()
	out := make([]semanticExportedSymbol, 0, len(exports))
	for _, symbol := range exports {
		if exportUnqualified {
			unqualified := symbol
			unqualified.unqualified = true
			out = append(out, unqualified)
		}
		if qualifier == "" || strings.Contains(symbol.name, "!") {
			continue
		}
		qualified := symbol
		qualified.name = qualifier + "!" + symbol.name
		qualified.unqualified = false
		out = append(out, qualified)
	}
	return out
}

func semanticModuleExports(mod *Module, spec *Spec, visiting map[string]bool) []semanticExportedSymbol {
	if mod == nil || visiting[mod.Name] {
		return nil
	}
	visiting[mod.Name] = true
	defer func() {
		visiting[mod.Name] = false
	}()

	byName := map[string]semanticExportedSymbol{}
	for _, ext := range mod.Extends {
		for _, symbol := range semanticModuleExports(spec.Modules[ext], spec, visiting) {
			byName[symbol.name] = symbol
		}
	}
	for _, inst := range mod.Instances {
		if inst.Local {
			continue
		}
		for _, symbol := range semanticInstanceSymbolsWithVisiting(inst, spec, visiting) {
			symbol.unqualified = false
			byName[symbol.name] = symbol
		}
	}
	for _, decl := range mod.Declarations {
		for _, name := range decl.Names {
			pos := decl.Pos
			if decl.NamePositions != nil {
				if namePos := decl.NamePositions[name]; namePos.Line > 0 || namePos.Column > 0 || namePos.File != "" {
					pos = namePos
				}
			}
			symbol := semanticExportedSymbol{name: name, kind: decl.Kind, pos: pos}
			if decl.Kind == ConstantDecl {
				if arity, ok := declarationArity(decl, name); ok {
					symbol.arity = arity
					symbol.hasArity = true
				}
			}
			byName[name] = symbol
		}
	}
	for i := range mod.Definitions {
		def := mod.Definitions[i]
		if def.Local {
			continue
		}
		symbol := semanticExportedSymbol{
			name:     def.Name,
			kind:     OperatorDecl,
			pos:      def.Pos,
			arity:    len(def.Params),
			hasArity: true,
		}
		if specs, ok := definitionOperatorParamSpecsForModule(mod.Name, def); ok {
			symbol.operatorParams = specs
			symbol.hasOperatorParams = true
		}
		byName[def.Name] = symbol
	}
	for _, assumption := range mod.Assumptions {
		if assumption.Name == "" {
			continue
		}
		byName[assumption.Name] = semanticExportedSymbol{
			name:     assumption.Name,
			kind:     OperatorDecl,
			pos:      assumption.SourcePosition(),
			arity:    0,
			hasArity: true,
		}
	}
	for _, export := range syntheticStandardExports(mod.Name, mod.Pos) {
		byName[export.Name] = semanticExportedSymbol{
			name:     export.Name,
			kind:     export.Kind,
			pos:      export.Pos,
			arity:    export.Arity,
			hasArity: true,
		}
	}

	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]semanticExportedSymbol, 0, len(names))
	for _, name := range names {
		out = append(out, byName[name])
	}
	return out
}

func addSemanticSymbol(symbol semanticExportedSymbol, defined map[string]Position, declKinds map[string]DeclarationKind, arities map[string]int, operatorParamSpecs map[string][]operatorParamSpec) {
	if _, exists := defined[symbol.name]; !exists {
		defined[symbol.name] = symbol.pos
	}
	if _, exists := declKinds[symbol.name]; !exists {
		declKinds[symbol.name] = symbol.kind
	}
	if symbol.hasArity {
		if _, exists := arities[symbol.name]; !exists {
			arities[symbol.name] = symbol.arity
		}
	}
	if symbol.hasOperatorParams {
		if _, exists := operatorParamSpecs[symbol.name]; !exists {
			operatorParamSpecs[symbol.name] = symbol.operatorParams
		}
	}
}

func addInstanceSymbols(inst Instance, spec *Spec, defined map[string]Position, declKinds map[string]DeclarationKind, arities map[string]int, operatorParamSpecs map[string][]operatorParamSpec, instanceSymbols map[string]importedSymbol, localSymbols map[string]localSymbol) Diagnostics {
	if spec == nil {
		return nil
	}
	var diags Diagnostics
	qualifier := inst.qualifier()
	exportUnqualified := inst.exportsUnqualified()
	for _, symbol := range semanticInstanceSymbols(inst, spec) {
		if exportUnqualified && symbol.unqualified {
			if _, shadowsLocal := localSymbols[symbol.name]; shadowsLocal {
				diags = append(diags, checkInstanceLocalShadow(symbol.name, symbol.pos, qualifier, localSymbols)...)
			} else {
				diags = append(diags, checkInstanceSymbolAmbiguity(symbol.name, symbol.kind, symbol.pos, qualifier, instanceSymbols)...)
				addSemanticSymbol(symbol, defined, declKinds, arities, operatorParamSpecs)
			}
			continue
		}
		addSemanticSymbol(symbol, defined, declKinds, arities, operatorParamSpecs)
	}
	return diags
}

func checkDefinitionParams(def Definition) Diagnostics {
	var diags Diagnostics
	seen := map[string]bool{}
	for _, param := range def.Params {
		if seen[param] {
			diags = append(diags, errorAt(def.Pos, "E1309", "duplicate parameter %s in definition %s", param, def.Name))
			continue
		}
		seen[param] = true
	}
	return diags
}

func checkDefinitionParamCollisions(def Definition, defined map[string]Position, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	for _, param := range def.Params {
		if _, operatorParam := def.ParamArities[param]; operatorParam && !isIdentifierName(param) {
			continue
		}
		diags = append(diags, checkBindingName("parameter", param, def.Pos, defined, locals)...)
	}
	return diags
}

func checkBindingName(kind, name string, pos Position, defined map[string]Position, locals map[string]bool) Diagnostics {
	if name == "" || builtinIdentifiers[name] {
		return nil
	}
	if locals != nil && locals[name] {
		return Diagnostics{errorAt(pos, "E1301", "%s %s conflicts with an existing local symbol", kind, name)}
	}
	if prev, ok := defined[name]; ok {
		return Diagnostics{errorAt(pos, "E1301", "%s %s conflicts with existing symbol declared at %s", kind, name, prev)}
	}
	return nil
}

func checkProofRef(ref ProofRef, defined map[string]Position) Diagnostics {
	if ref.Name == "" || builtinIdentifiers[ref.Name] {
		return nil
	}
	if _, ok := defined[ref.Name]; !ok {
		return Diagnostics{errorAt(ref.Pos, "E1302", "undefined identifier %s", ref.Name)}
	}
	return nil
}

func checkHideRef(ref ProofRef, theoremLikeDefs, proofStepNames map[string]bool) Diagnostics {
	if ref.Mode != "HIDE" || ref.Name == "" || builtinIdentifiers[ref.Name] {
		return nil
	}
	if theoremLikeDefs[ref.Name] || proofStepNames[ref.Name] {
		return nil
	}
	return Diagnostics{errorAt(ref.Pos, "E4357", "HIDE can only refer to theorems, assumptions, or proof steps; %s is not a proof fact", ref.Name)}
}

func checkProofSummary(proof ProofSummary, declKinds map[string]DeclarationKind) Diagnostics {
	var diags Diagnostics
	goalLevel := exprLevel(proof.Goal, declKinds, nil)
	nonExpressionSteps := map[string]ProofStep{}
	for _, step := range proof.Steps {
		if step.Name == "" || step.Kind == "ASSERT" {
			continue
		}
		nonExpressionSteps[step.Name] = step
	}
	for _, step := range proof.Steps {
		if step.Implicit && step.Name != "" {
			diags = append(diags, errorAt(step.Pos, "E4350", "implicit proof step cannot have name %s", step.Name))
		}
		for _, ref := range step.Refs {
			if bad, ok := nonExpressionSteps[ref]; ok {
				diags = append(diags, errorAt(step.Pos, "E4351", "proof step %s of kind %s cannot be used as an expression", ref, bad.Kind))
			}
		}
		if goalLevel == temporalLevel && step.Depth == 0 {
			switch step.Kind {
			case "HAVE":
				if exprLevel(step.Expr, declKinds, nil) != constantLevel {
					diags = append(diags, errorAt(step.Pos, "E4352", "temporal proof goal requires constant-level HAVE step"))
				}
			case "TAKE":
				for _, bound := range step.Bounds {
					if exprLevel(bound.Set, declKinds, nil) != constantLevel {
						diags = append(diags, errorAt(bound.Pos, "E4352", "temporal proof goal requires constant-level TAKE bound"))
					}
				}
			case "WITNESS":
				for _, expr := range step.Exprs {
					if exprLevel(expr, declKinds, nil) != constantLevel {
						diags = append(diags, errorAt(expr.Position(), "E4352", "temporal proof goal requires constant-level WITNESS expression"))
					}
				}
			case "CASE":
				if exprLevel(step.Expr, declKinds, nil) != constantLevel {
					diags = append(diags, errorAt(step.Pos, "E4353", "temporal proof goal requires constant-level CASE step"))
				}
			}
		}
		if step.Kind == "PICK" && exprLevel(step.Expr, declKinds, nil) == temporalLevel {
			for _, bound := range step.Bounds {
				if exprLevel(bound.Set, declKinds, nil) != constantLevel {
					diags = append(diags, errorAt(bound.Pos, "E4354", "temporal PICK formula requires constant-level bound"))
				}
			}
		}
	}
	return diags
}

func checkAssumeProveDefinitionUse(expr Expr, assumeProveDefs map[string]bool, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	switch e := expr.(type) {
	case *IdentExpr:
		if !locals[e.Name] && assumeProveDefs[e.Name] {
			diags = append(diags, errorAt(e.Pos, "E4355", "ASSUME/PROVE definition %s cannot be used where an ordinary expression is required", e.Name))
		}
	case *UnaryExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Expr, assumeProveDefs, locals)...)
	case *BinaryExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Left, assumeProveDefs, locals)...)
		diags = append(diags, checkAssumeProveDefinitionUse(e.Right, assumeProveDefs, locals)...)
	case *CallExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Callee, assumeProveDefs, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkAssumeProveDefinitionUse(arg, assumeProveDefs, locals)...)
		}
	case *IfExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Cond, assumeProveDefs, locals)...)
		diags = append(diags, checkAssumeProveDefinitionUse(e.Then, assumeProveDefs, locals)...)
		diags = append(diags, checkAssumeProveDefinitionUse(e.Else, assumeProveDefs, locals)...)
	case *LetExpr:
		letLocals := letScopeLocals(locals, e)
		for _, def := range e.Definitions {
			defLocals := letDefinitionBodyLocals(letLocals, def)
			diags = append(diags, checkAssumeProveDefinitionUse(def.Expr, assumeProveDefs, defLocals)...)
		}
		diags = append(diags, checkAssumeProveDefinitionUse(e.Body, assumeProveDefs, letLocals)...)
	case *QuantifierExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Set, assumeProveDefs, locals)...)
		diags = append(diags, checkAssumeProveDefinitionUse(e.Body, assumeProveDefs, withLocal(locals, e.Var))...)
	case *CaseExpr:
		for _, arm := range e.Arms {
			diags = append(diags, checkAssumeProveDefinitionUse(arm.Test, assumeProveDefs, locals)...)
			diags = append(diags, checkAssumeProveDefinitionUse(arm.Value, assumeProveDefs, locals)...)
		}
		if e.Other != nil {
			diags = append(diags, checkAssumeProveDefinitionUse(e.Other, assumeProveDefs, locals)...)
		}
	case *ChooseExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Set, assumeProveDefs, locals)...)
		diags = append(diags, checkAssumeProveDefinitionUse(e.Body, assumeProveDefs, withLocal(locals, e.Var))...)
	case *TupleExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkAssumeProveDefinitionUse(elem, assumeProveDefs, locals)...)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkAssumeProveDefinitionUse(elem, assumeProveDefs, locals)...)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkAssumeProveDefinitionUse(field.Value, assumeProveDefs, locals)...)
		}
	case *RecordComponentExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Record, assumeProveDefs, locals)...)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkAssumeProveDefinitionUse(field.Set, assumeProveDefs, locals)...)
		}
	case *FunctionExpr:
		fnLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, checkAssumeProveDefinitionUse(bound.Set, assumeProveDefs, locals)...)
			fnLocals[bound.Name] = true
		}
		diags = append(diags, checkAssumeProveDefinitionUse(e.Body, assumeProveDefs, fnLocals)...)
	case *FunctionAppExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Function, assumeProveDefs, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkAssumeProveDefinitionUse(arg, assumeProveDefs, locals)...)
		}
	case *ExceptExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Base, assumeProveDefs, locals)...)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					diags = append(diags, checkAssumeProveDefinitionUse(index, assumeProveDefs, locals)...)
				}
			}
			diags = append(diags, checkAssumeProveDefinitionUse(spec.Value, assumeProveDefs, locals)...)
		}
	case *LabelExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Body, assumeProveDefs, locals)...)
	case *ActionExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Action, assumeProveDefs, locals)...)
		diags = append(diags, checkAssumeProveDefinitionUse(e.Subscript, assumeProveDefs, locals)...)
	case *FairnessExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Subscript, assumeProveDefs, locals)...)
		diags = append(diags, checkAssumeProveDefinitionUse(e.Action, assumeProveDefs, locals)...)
	case *FunctionSetExpr:
		diags = append(diags, checkAssumeProveDefinitionUse(e.Domain, assumeProveDefs, locals)...)
		diags = append(diags, checkAssumeProveDefinitionUse(e.Range, assumeProveDefs, locals)...)
	case *SetComprehensionExpr:
		compLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, checkAssumeProveDefinitionUse(bound.Set, assumeProveDefs, locals)...)
			compLocals[bound.Name] = true
		}
		diags = append(diags, checkAssumeProveDefinitionUse(e.Element, assumeProveDefs, compLocals)...)
		if e.Predicate != nil {
			diags = append(diags, checkAssumeProveDefinitionUse(e.Predicate, assumeProveDefs, compLocals)...)
		}
	}
	return diags
}

func checkInstanceSubstitutions(mod *Module, inst Instance, spec *Spec, defined map[string]Position, declKinds map[string]DeclarationKind, arities map[string]int, operatorParams map[string][]operatorParamSpec) Diagnostics {
	var diags Diagnostics
	if spec == nil {
		return nil
	}
	target := spec.Modules[inst.Module]
	if target == nil {
		return nil
	}
	targets := moduleOwnSubstitutionTargets(target)
	matchLevels := moduleRequiresSubstitutionLevelMatch(target)
	implicit := moduleImplicitSubstitutions(mod, spec)
	seen := map[string]Position{}
	for _, subst := range instanceSubstitutions(inst) {
		name := subst.Name
		expr := subst.Expr
		if name == "" || expr == nil {
			continue
		}
		if prev, ok := seen[name]; ok {
			diags = append(diags, errorAt(subst.Pos, "E1312", "duplicate INSTANCE substitution for %s; first substitution at %s", name, prev))
		} else {
			seen[name] = subst.Pos
		}
		target, ok := targets[name]
		if !ok {
			diags = append(diags, errorAt(subst.Pos, "E1305", "INSTANCE substitution target %s is not a CONSTANT or VARIABLE of module %s", name, inst.Module))
			continue
		}
		want := target.Arity
		got := substitutionExprArity(expr, arities)
		if got != want {
			diags = append(diags, errorAt(subst.Pos, "E1306", "INSTANCE substitution %s arity mismatch: got %d, want %d", name, got, want))
		}
		if matchLevels {
			level := exprLevel(expr, declKinds, nil)
			switch target.Kind {
			case ConstantDecl:
				if level != constantLevel {
					diags = append(diags, errorAt(subst.Pos, "E1314", "INSTANCE substitution %s must be constant-level", name))
				}
			case VariableDecl:
				if level != variableLevel {
					diags = append(diags, errorAt(subst.Pos, "E1314", "INSTANCE substitution %s must be variable-level", name))
				}
			}
		}
		diags = append(diags, checkExpr(expr, defined, nil)...)
		diags = append(diags, checkCallArity(expr, arities, operatorParams, nil)...)
		diags = append(diags, checkPrimedConstants(expr, declKinds, nil)...)
	}
	for name, target := range targets {
		if _, ok := seen[name]; ok {
			continue
		}
		want := target.Arity
		got, ok := implicit[name]
		if !ok {
			diags = append(diags, errorAt(inst.Pos, "E1313", "INSTANCE %s requires substitution for %s", inst.Module, name))
			continue
		}
		if got != want {
			diags = append(diags, errorAt(inst.Pos, "E1313", "implicit INSTANCE substitution %s arity mismatch: got %d, want %d", name, got, want))
		}
	}
	return diags
}

func instanceSubstitutions(inst Instance) []Substitution {
	if len(inst.SubstitutionList) > 0 {
		return inst.SubstitutionList
	}
	substitutions := make([]Substitution, 0, len(inst.Substitutions))
	for name, expr := range inst.Substitutions {
		substitutions = append(substitutions, Substitution{Name: name, Expr: expr, Pos: inst.Pos})
	}
	return substitutions
}

type substitutionTarget struct {
	Kind  DeclarationKind
	Arity int
}

func moduleOwnSubstitutionTargets(mod *Module) map[string]substitutionTarget {
	targets := map[string]substitutionTarget{}
	if mod == nil {
		return targets
	}
	if isEmbeddedStandardModule(mod) {
		return targets
	}
	for _, decl := range mod.Declarations {
		if decl.Kind != ConstantDecl && decl.Kind != VariableDecl {
			continue
		}
		for _, name := range decl.Names {
			arity := 0
			if decl.Kind == ConstantDecl {
				if declared, ok := declarationArity(decl, name); ok {
					arity = declared
				}
			}
			targets[name] = substitutionTarget{Kind: decl.Kind, Arity: arity}
		}
	}
	return targets
}

func moduleRequiresSubstitutionLevelMatch(mod *Module) bool {
	if mod == nil {
		return false
	}
	for _, decl := range mod.Declarations {
		if decl.Kind == VariableDecl {
			return true
		}
	}
	return false
}

func moduleImplicitSubstitutions(mod *Module, spec *Spec) map[string]int {
	arities := map[string]int{}
	var collect func(*Module)
	collect = func(cur *Module) {
		if cur == nil {
			return
		}
		for _, decl := range cur.Declarations {
			for _, name := range decl.Names {
				arity := 0
				if decl.Kind == ConstantDecl {
					if declared, ok := declarationArity(decl, name); ok {
						arity = declared
					}
				}
				if _, exists := arities[name]; !exists {
					arities[name] = arity
				}
			}
		}
		for _, def := range cur.Definitions {
			if def.Local {
				continue
			}
			if _, exists := arities[def.Name]; !exists {
				arities[def.Name] = len(def.Params)
			}
		}
	}
	collect(mod)
	if spec != nil && mod != nil {
		for _, dep := range mod.Extends {
			collect(spec.Modules[dep])
		}
	}
	return arities
}

func substitutionExprArity(expr Expr, arities map[string]int) int {
	if ident, ok := expr.(*IdentExpr); ok {
		if arity, exists := arities[ident.Name]; exists {
			return arity
		}
	}
	return 0
}

func declarationArity(decl Declaration, name string) (int, bool) {
	if decl.Arities == nil {
		return 0, false
	}
	arity, ok := decl.Arities[name]
	return arity, ok
}

func definitionFunctionArity(def Definition) (int, bool) {
	fn, ok := def.Expr.(*FunctionExpr)
	if !ok {
		return 0, false
	}
	return len(fn.Bounds), true
}

type operatorParamSpec struct {
	Name  string
	Arity int
}

func definitionOperatorParamSpecsForModule(moduleName string, def Definition) ([]operatorParamSpec, bool) {
	if moduleName == "Sequences" && def.Name == "SelectSeq" && len(def.Params) == 2 {
		return []operatorParamSpec{
			{Name: def.Params[0], Arity: -1},
			{Name: def.Params[1], Arity: 1},
		}, true
	}
	return definitionOperatorParamSpecs(def)
}

func definitionOperatorParamSpecs(def Definition) ([]operatorParamSpec, bool) {
	if len(def.Params) == 0 {
		return nil, false
	}
	specs := make([]operatorParamSpec, len(def.Params))
	for i, param := range def.Params {
		specs[i] = operatorParamSpec{Name: param, Arity: -1}
		if arity, ok := def.ParamArities[param]; ok {
			specs[i].Arity = arity
		}
	}
	return specs, true
}

func definitionBodyArities(base map[string]int, def Definition) map[string]int {
	out := copyIntMap(base)
	for _, param := range def.Params {
		delete(out, localOperatorArityKey(param))
	}
	for name, arity := range def.ParamArities {
		out[localOperatorArityKey(name)] = arity
	}
	return out
}

func quantifierBodyArities(base map[string]int, expr *QuantifierExpr) map[string]int {
	if expr == nil || !expr.HasOperatorArity {
		return base
	}
	out := copyIntMap(base)
	out[localOperatorArityKey(expr.Var)] = expr.OperatorArity
	return out
}

func localOperatorArityKey(name string) string {
	return "\x00local-operator-arity:" + name
}

func checkExpr(expr Expr, defined map[string]Position, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	switch e := expr.(type) {
	case *IdentExpr:
		if e.Name == "" || localIdentifierInScope(locals, e.Name) || builtinIdentifiers[e.Name] {
			return nil
		}
		if _, ok := builtinOperatorArity(e.Name); ok {
			return nil
		}
		if _, ok := defined[e.Name]; !ok {
			diags = append(diags, errorAt(e.Pos, "E1302", "undefined identifier %s", e.Name))
		}
	case *LiteralExpr:
	case *UnaryExpr:
		diags = append(diags, checkExpr(e.Expr, defined, locals)...)
	case *BinaryExpr:
		diags = append(diags, checkExpr(e.Left, defined, locals)...)
		diags = append(diags, checkExpr(e.Right, defined, locals)...)
	case *CallExpr:
		diags = append(diags, checkExpr(e.Callee, defined, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkExpr(arg, defined, locals)...)
		}
	case *IfExpr:
		diags = append(diags, checkExpr(e.Cond, defined, locals)...)
		diags = append(diags, checkExpr(e.Then, defined, locals)...)
		diags = append(diags, checkExpr(e.Else, defined, locals)...)
	case *LetExpr:
		diags = append(diags, checkLetRecursiveSections(e)...)
		letLocals := letScopeLocals(locals, e)
		for _, def := range e.Definitions {
			diags = append(diags, checkDefinitionParams(def)...)
			diags = append(diags, checkDefinitionParamCollisions(def, defined, letLocals)...)
			defLocals := letDefinitionBodyLocals(letLocals, def)
			diags = append(diags, checkExpr(def.Expr, defined, defLocals)...)
		}
		diags = append(diags, checkExpr(e.Body, defined, letLocals)...)
	case *QuantifierExpr:
		diags = append(diags, checkExpr(e.Set, defined, locals)...)
		diags = append(diags, checkBindingName("bound symbol", e.Var, e.Pos, defined, locals)...)
		quantLocals := map[string]bool{}
		for name, ok := range locals {
			quantLocals[name] = ok
		}
		quantLocals[e.Var] = true
		diags = append(diags, checkExpr(e.Body, defined, quantLocals)...)
	case *CaseExpr:
		for _, arm := range e.Arms {
			diags = append(diags, checkExpr(arm.Test, defined, locals)...)
			diags = append(diags, checkExpr(arm.Value, defined, locals)...)
		}
		if e.Other != nil {
			diags = append(diags, checkExpr(e.Other, defined, locals)...)
		}
	case *ChooseExpr:
		diags = append(diags, checkExpr(e.Set, defined, locals)...)
		diags = append(diags, checkBindingName("bound symbol", e.Var, e.Pos, defined, locals)...)
		chooseLocals := map[string]bool{}
		for name, ok := range locals {
			chooseLocals[name] = ok
		}
		chooseLocals[e.Var] = true
		diags = append(diags, checkExpr(e.Body, defined, chooseLocals)...)
	case *TupleExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkExpr(elem, defined, locals)...)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkExpr(elem, defined, locals)...)
		}
	case *RecordExpr:
		seenFields := map[string]Position{}
		for _, field := range e.Fields {
			if prev, ok := seenFields[field.Name]; ok {
				diags = append(diags, errorAt(field.Pos, "E1318", "duplicate record field %s; first field at %s", field.Name, prev))
			} else {
				seenFields[field.Name] = field.Pos
			}
			if field.Name != "" && (locals == nil || !locals[field.Name]) {
				if prev, ok := defined[field.Name]; ok {
					diags = append(diags, warningAt(field.Pos, "W4802", "record field %s has the same name as an existing symbol declared at %s; the symbol value is not used as the field name", field.Name, prev))
				}
			}
			diags = append(diags, checkExpr(field.Value, defined, locals)...)
		}
	case *RecordComponentExpr:
		diags = append(diags, checkExpr(e.Record, defined, locals)...)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkExpr(field.Set, defined, locals)...)
		}
	case *FunctionExpr:
		fnLocals := map[string]bool{}
		for name, ok := range locals {
			fnLocals[name] = ok
		}
		for _, bound := range e.Bounds {
			diags = append(diags, checkBindingName("bound symbol", bound.Name, bound.Pos, defined, fnLocals)...)
			diags = append(diags, checkExpr(bound.Set, defined, locals)...)
			fnLocals[bound.Name] = true
		}
		diags = append(diags, checkExpr(e.Body, defined, fnLocals)...)
	case *FunctionAppExpr:
		diags = append(diags, checkExpr(e.Function, defined, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkExpr(arg, defined, locals)...)
		}
	case *ExceptExpr:
		diags = append(diags, checkExpr(e.Base, defined, locals)...)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					diags = append(diags, checkExpr(index, defined, locals)...)
				}
			}
			exceptLocals := copyBoolMap(locals)
			exceptLocals["@"] = true
			diags = append(diags, checkExpr(spec.Value, defined, exceptLocals)...)
		}
	case *LabelExpr:
		diags = append(diags, checkExpr(e.Body, defined, locals)...)
	case *ActionExpr:
		diags = append(diags, checkExpr(e.Action, defined, locals)...)
		diags = append(diags, checkExpr(e.Subscript, defined, locals)...)
	case *FairnessExpr:
		diags = append(diags, checkExpr(e.Subscript, defined, locals)...)
		diags = append(diags, checkExpr(e.Action, defined, locals)...)
	case *FunctionSetExpr:
		diags = append(diags, checkExpr(e.Domain, defined, locals)...)
		diags = append(diags, checkExpr(e.Range, defined, locals)...)
	case *SetComprehensionExpr:
		compLocals := map[string]bool{}
		for name, ok := range locals {
			compLocals[name] = ok
		}
		for _, bound := range e.Bounds {
			diags = append(diags, checkBindingName("bound symbol", bound.Name, bound.Pos, defined, compLocals)...)
			diags = append(diags, checkExpr(bound.Set, defined, locals)...)
			compLocals[bound.Name] = true
		}
		diags = append(diags, checkExpr(e.Element, defined, compLocals)...)
		if e.Predicate != nil {
			diags = append(diags, checkExpr(e.Predicate, defined, compLocals)...)
		}
	}
	return diags
}

func localIdentifierInScope(locals map[string]bool, name string) bool {
	if locals[name] {
		return true
	}
	for local := range locals {
		if strings.HasSuffix(local, "!") && strings.HasPrefix(name, local) {
			return true
		}
	}
	return false
}

func checkCallArity(expr Expr, arities map[string]int, operatorParams map[string][]operatorParamSpec, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	recur := func(expr Expr, arities map[string]int, locals map[string]bool) Diagnostics {
		return checkCallArity(expr, arities, operatorParams, locals)
	}
	switch e := expr.(type) {
	case *IdentExpr:
		if e.Name != "" && !locals[e.Name] {
			if _, ok := builtinOperatorArity(e.Name); ok {
				return nil
			}
			if want, ok := arities[e.Name]; ok && want != 0 {
				diags = append(diags, errorAt(e.Pos, "E1304", "operator %s arity mismatch: got 0 args, want %d", e.Name, want))
			}
		}
	case *UnaryExpr:
		diags = append(diags, recur(e.Expr, arities, locals)...)
	case *BinaryExpr:
		diags = append(diags, recur(e.Left, arities, locals)...)
		diags = append(diags, recur(e.Right, arities, locals)...)
	case *CallExpr:
		var specs []operatorParamSpec
		if ident, ok := e.Callee.(*IdentExpr); ok && !locals[ident.Name] {
			want, ok := arities[ident.Name]
			if !ok {
				want, ok = builtinOperatorArity(ident.Name)
			}
			if ok && want != len(e.Args) {
				diags = append(diags, errorAt(e.Pos, "E1304", "operator %s arity mismatch: got %d args, want %d", ident.Name, len(e.Args), want))
			}
			specs = operatorParams[ident.Name]
		}
		if _, ok := e.Callee.(*IdentExpr); !ok {
			diags = append(diags, recur(e.Callee, arities, locals)...)
		}
		for i, arg := range e.Args {
			if callArgumentIsOperatorArgument(i, arg, specs, arities, locals) {
				continue
			}
			diags = append(diags, recur(arg, arities, locals)...)
		}
	case *IfExpr:
		diags = append(diags, recur(e.Cond, arities, locals)...)
		diags = append(diags, recur(e.Then, arities, locals)...)
		diags = append(diags, recur(e.Else, arities, locals)...)
	case *LetExpr:
		letArities := map[string]int{}
		for name, arity := range arities {
			letArities[name] = arity
		}
		letOperatorParams := map[string][]operatorParamSpec{}
		for name, specs := range operatorParams {
			letOperatorParams[name] = specs
		}
		for _, decl := range e.Recursives {
			for _, name := range decl.Names {
				arity, ok := declarationArity(decl, name)
				if !ok {
					arity = 0
				}
				letArities[name] = arity
			}
		}
		for _, def := range e.Definitions {
			if _, recursive := letArities[def.Name]; !recursive {
				letArities[def.Name] = len(def.Params)
			}
			if specs, ok := definitionOperatorParamSpecs(def); ok {
				letOperatorParams[def.Name] = specs
			}
		}
		for _, def := range e.Definitions {
			defLocals := copyBoolMap(locals)
			for _, param := range def.Params {
				defLocals[param] = true
			}
			defArities := definitionBodyArities(letArities, def)
			diags = append(diags, checkCallArity(def.Expr, defArities, letOperatorParams, defLocals)...)
		}
		diags = append(diags, checkCallArity(e.Body, letArities, letOperatorParams, locals)...)
	case *QuantifierExpr:
		diags = append(diags, recur(e.Set, arities, locals)...)
		quantLocals := copyBoolMap(locals)
		quantLocals[e.Var] = true
		diags = append(diags, recur(e.Body, quantifierBodyArities(arities, e), quantLocals)...)
	case *CaseExpr:
		for _, arm := range e.Arms {
			diags = append(diags, recur(arm.Test, arities, locals)...)
			diags = append(diags, recur(arm.Value, arities, locals)...)
		}
		if e.Other != nil {
			diags = append(diags, recur(e.Other, arities, locals)...)
		}
	case *ChooseExpr:
		diags = append(diags, recur(e.Set, arities, locals)...)
		chooseLocals := copyBoolMap(locals)
		chooseLocals[e.Var] = true
		diags = append(diags, recur(e.Body, arities, chooseLocals)...)
	case *TupleExpr:
		for _, elem := range e.Elems {
			diags = append(diags, recur(elem, arities, locals)...)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			diags = append(diags, recur(elem, arities, locals)...)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			diags = append(diags, recur(field.Value, arities, locals)...)
		}
	case *RecordComponentExpr:
		diags = append(diags, recur(e.Record, arities, locals)...)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			diags = append(diags, recur(field.Set, arities, locals)...)
		}
	case *FunctionExpr:
		fnLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, recur(bound.Set, arities, locals)...)
			fnLocals[bound.Name] = true
		}
		diags = append(diags, recur(e.Body, arities, fnLocals)...)
	case *FunctionAppExpr:
		diags = append(diags, recur(e.Function, arities, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, recur(arg, arities, locals)...)
		}
	case *ExceptExpr:
		diags = append(diags, recur(e.Base, arities, locals)...)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					diags = append(diags, recur(index, arities, locals)...)
				}
			}
			diags = append(diags, recur(spec.Value, arities, locals)...)
		}
	case *LabelExpr:
		diags = append(diags, recur(e.Body, arities, locals)...)
	case *ActionExpr:
		diags = append(diags, recur(e.Action, arities, locals)...)
		diags = append(diags, recur(e.Subscript, arities, locals)...)
	case *FairnessExpr:
		diags = append(diags, recur(e.Subscript, arities, locals)...)
		diags = append(diags, recur(e.Action, arities, locals)...)
	case *FunctionSetExpr:
		diags = append(diags, recur(e.Domain, arities, locals)...)
		diags = append(diags, recur(e.Range, arities, locals)...)
	case *SetComprehensionExpr:
		compLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, recur(bound.Set, arities, locals)...)
			compLocals[bound.Name] = true
		}
		diags = append(diags, recur(e.Element, arities, compLocals)...)
		if e.Predicate != nil {
			diags = append(diags, recur(e.Predicate, arities, compLocals)...)
		}
	}
	return diags
}

func callArgumentIsOperatorArgument(index int, arg Expr, specs []operatorParamSpec, arities map[string]int, locals map[string]bool) bool {
	if index >= len(specs) || specs[index].Arity < 0 {
		return false
	}
	_, ok := operatorArgumentArity(arg, arities, locals)
	return ok
}

func checkOperatorArgumentKinds(expr Expr, operatorParams map[string][]operatorParamSpec, arities map[string]int, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	switch e := expr.(type) {
	case *UnaryExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Expr, operatorParams, arities, locals)...)
	case *BinaryExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Left, operatorParams, arities, locals)...)
		diags = append(diags, checkOperatorArgumentKinds(e.Right, operatorParams, arities, locals)...)
	case *CallExpr:
		if ident, ok := e.Callee.(*IdentExpr); ok && !locals[ident.Name] {
			if specs, exists := operatorParams[ident.Name]; exists {
				limit := len(e.Args)
				if len(specs) < limit {
					limit = len(specs)
				}
				for i := 0; i < limit; i++ {
					spec := specs[i]
					arg := e.Args[i]
					if spec.Arity >= 0 {
						got, ok := operatorArgumentArity(arg, arities, locals)
						if !ok {
							diags = append(diags, errorAt(arg.Position(), "E1319", "operator parameter %s requires an operator argument of arity %d", spec.Name, spec.Arity))
							continue
						}
						if got != spec.Arity {
							diags = append(diags, errorAt(arg.Position(), "E1319", "operator argument arity mismatch for parameter %s: got %d, want %d", spec.Name, got, spec.Arity))
						}
					} else if fn, ok := arg.(*FunctionExpr); ok && fn.IsLambda {
						diags = append(diags, errorAt(arg.Position(), "E1319", "expression parameter %s cannot accept a LAMBDA operator argument", spec.Name))
					}
				}
			}
		} else {
			diags = append(diags, checkOperatorArgumentKinds(e.Callee, operatorParams, arities, locals)...)
		}
		for _, arg := range e.Args {
			diags = append(diags, checkOperatorArgumentKinds(arg, operatorParams, arities, locals)...)
		}
	case *IfExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Cond, operatorParams, arities, locals)...)
		diags = append(diags, checkOperatorArgumentKinds(e.Then, operatorParams, arities, locals)...)
		diags = append(diags, checkOperatorArgumentKinds(e.Else, operatorParams, arities, locals)...)
	case *LetExpr:
		letOperatorParams := map[string][]operatorParamSpec{}
		for name, specs := range operatorParams {
			letOperatorParams[name] = specs
		}
		letArities := map[string]int{}
		for name, arity := range arities {
			letArities[name] = arity
		}
		for _, decl := range e.Recursives {
			for _, name := range decl.Names {
				arity, ok := declarationArity(decl, name)
				if !ok {
					arity = 0
				}
				letArities[name] = arity
			}
		}
		for _, def := range e.Definitions {
			if _, recursive := letArities[def.Name]; !recursive {
				letArities[def.Name] = len(def.Params)
			}
			if specs, ok := definitionOperatorParamSpecs(def); ok {
				letOperatorParams[def.Name] = specs
			}
		}
		for _, def := range e.Definitions {
			defLocals := copyBoolMap(locals)
			for _, param := range def.Params {
				defLocals[param] = true
			}
			defArities := definitionBodyArities(letArities, def)
			diags = append(diags, checkOperatorArgumentKinds(def.Expr, letOperatorParams, defArities, defLocals)...)
		}
		diags = append(diags, checkOperatorArgumentKinds(e.Body, letOperatorParams, letArities, locals)...)
	case *QuantifierExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Set, operatorParams, arities, locals)...)
		quantLocals := copyBoolMap(locals)
		quantLocals[e.Var] = true
		diags = append(diags, checkOperatorArgumentKinds(e.Body, operatorParams, quantifierBodyArities(arities, e), quantLocals)...)
	case *CaseExpr:
		for _, arm := range e.Arms {
			diags = append(diags, checkOperatorArgumentKinds(arm.Test, operatorParams, arities, locals)...)
			diags = append(diags, checkOperatorArgumentKinds(arm.Value, operatorParams, arities, locals)...)
		}
		if e.Other != nil {
			diags = append(diags, checkOperatorArgumentKinds(e.Other, operatorParams, arities, locals)...)
		}
	case *ChooseExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Set, operatorParams, arities, locals)...)
		chooseLocals := copyBoolMap(locals)
		chooseLocals[e.Var] = true
		diags = append(diags, checkOperatorArgumentKinds(e.Body, operatorParams, arities, chooseLocals)...)
	case *TupleExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkOperatorArgumentKinds(elem, operatorParams, arities, locals)...)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkOperatorArgumentKinds(elem, operatorParams, arities, locals)...)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkOperatorArgumentKinds(field.Value, operatorParams, arities, locals)...)
		}
	case *RecordComponentExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Record, operatorParams, arities, locals)...)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkOperatorArgumentKinds(field.Set, operatorParams, arities, locals)...)
		}
	case *FunctionExpr:
		fnLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, checkOperatorArgumentKinds(bound.Set, operatorParams, arities, locals)...)
			fnLocals[bound.Name] = true
		}
		diags = append(diags, checkOperatorArgumentKinds(e.Body, operatorParams, arities, fnLocals)...)
	case *FunctionAppExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Function, operatorParams, arities, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkOperatorArgumentKinds(arg, operatorParams, arities, locals)...)
		}
	case *ExceptExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Base, operatorParams, arities, locals)...)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					diags = append(diags, checkOperatorArgumentKinds(index, operatorParams, arities, locals)...)
				}
			}
			diags = append(diags, checkOperatorArgumentKinds(spec.Value, operatorParams, arities, locals)...)
		}
	case *LabelExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Body, operatorParams, arities, locals)...)
	case *ActionExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Action, operatorParams, arities, locals)...)
		diags = append(diags, checkOperatorArgumentKinds(e.Subscript, operatorParams, arities, locals)...)
	case *FairnessExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Subscript, operatorParams, arities, locals)...)
		diags = append(diags, checkOperatorArgumentKinds(e.Action, operatorParams, arities, locals)...)
	case *FunctionSetExpr:
		diags = append(diags, checkOperatorArgumentKinds(e.Domain, operatorParams, arities, locals)...)
		diags = append(diags, checkOperatorArgumentKinds(e.Range, operatorParams, arities, locals)...)
	case *SetComprehensionExpr:
		compLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, checkOperatorArgumentKinds(bound.Set, operatorParams, arities, locals)...)
			compLocals[bound.Name] = true
		}
		diags = append(diags, checkOperatorArgumentKinds(e.Element, operatorParams, arities, compLocals)...)
		if e.Predicate != nil {
			diags = append(diags, checkOperatorArgumentKinds(e.Predicate, operatorParams, arities, compLocals)...)
		}
	}
	return diags
}

func operatorArgumentArity(expr Expr, arities map[string]int, locals map[string]bool) (int, bool) {
	switch e := expr.(type) {
	case *IdentExpr:
		if locals != nil && locals[e.Name] {
			arity, ok := arities[localOperatorArityKey(e.Name)]
			return arity, ok
		}
		if arity, ok := arities[e.Name]; ok {
			return arity, true
		}
		return builtinOperatorArity(e.Name)
	case *FunctionExpr:
		return len(e.Bounds), true
	default:
		return 0, false
	}
}

func builtinOperatorArity(name string) (int, bool) {
	op, ok := GetSanyOperator(name)
	if !ok {
		return 0, false
	}
	switch {
	case op.IsPrefix(), op.IsPostfix():
		return 1, true
	case op.IsInfix():
		return 2, true
	default:
		return 0, false
	}
}

func checkFunctionArity(expr Expr, functionArities map[string]int, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	switch e := expr.(type) {
	case *UnaryExpr:
		diags = append(diags, checkFunctionArity(e.Expr, functionArities, locals)...)
	case *BinaryExpr:
		diags = append(diags, checkFunctionArity(e.Left, functionArities, locals)...)
		diags = append(diags, checkFunctionArity(e.Right, functionArities, locals)...)
	case *CallExpr:
		diags = append(diags, checkFunctionArity(e.Callee, functionArities, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkFunctionArity(arg, functionArities, locals)...)
		}
	case *IfExpr:
		diags = append(diags, checkFunctionArity(e.Cond, functionArities, locals)...)
		diags = append(diags, checkFunctionArity(e.Then, functionArities, locals)...)
		diags = append(diags, checkFunctionArity(e.Else, functionArities, locals)...)
	case *LetExpr:
		letArities := map[string]int{}
		for name, arity := range functionArities {
			letArities[name] = arity
		}
		for _, def := range e.Definitions {
			if arity, ok := definitionFunctionArity(def); ok {
				letArities[def.Name] = arity
			}
		}
		for _, def := range e.Definitions {
			defLocals := copyBoolMap(locals)
			for _, param := range def.Params {
				defLocals[param] = true
			}
			diags = append(diags, checkFunctionArity(def.Expr, letArities, defLocals)...)
		}
		diags = append(diags, checkFunctionArity(e.Body, letArities, locals)...)
	case *QuantifierExpr:
		diags = append(diags, checkFunctionArity(e.Set, functionArities, locals)...)
		quantLocals := copyBoolMap(locals)
		quantLocals[e.Var] = true
		diags = append(diags, checkFunctionArity(e.Body, functionArities, quantLocals)...)
	case *CaseExpr:
		for _, arm := range e.Arms {
			diags = append(diags, checkFunctionArity(arm.Test, functionArities, locals)...)
			diags = append(diags, checkFunctionArity(arm.Value, functionArities, locals)...)
		}
		if e.Other != nil {
			diags = append(diags, checkFunctionArity(e.Other, functionArities, locals)...)
		}
	case *ChooseExpr:
		diags = append(diags, checkFunctionArity(e.Set, functionArities, locals)...)
		chooseLocals := copyBoolMap(locals)
		chooseLocals[e.Var] = true
		diags = append(diags, checkFunctionArity(e.Body, functionArities, chooseLocals)...)
	case *TupleExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkFunctionArity(elem, functionArities, locals)...)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkFunctionArity(elem, functionArities, locals)...)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkFunctionArity(field.Value, functionArities, locals)...)
		}
	case *RecordComponentExpr:
		diags = append(diags, checkFunctionArity(e.Record, functionArities, locals)...)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkFunctionArity(field.Set, functionArities, locals)...)
		}
	case *FunctionExpr:
		fnLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, checkFunctionArity(bound.Set, functionArities, locals)...)
			fnLocals[bound.Name] = true
		}
		diags = append(diags, checkFunctionArity(e.Body, functionArities, fnLocals)...)
	case *FunctionAppExpr:
		if ident, ok := e.Function.(*IdentExpr); ok && !locals[ident.Name] {
			if want, exists := functionArities[ident.Name]; exists && !validFunctionApplicationArity(want, len(e.Args)) {
				diags = append(diags, errorAt(e.Pos, "E1317", "function %s arity mismatch: got %d args, want %d", ident.Name, len(e.Args), want))
			}
		} else {
			diags = append(diags, checkFunctionArity(e.Function, functionArities, locals)...)
		}
		for _, arg := range e.Args {
			diags = append(diags, checkFunctionArity(arg, functionArities, locals)...)
		}
	case *ExceptExpr:
		diags = append(diags, checkFunctionArity(e.Base, functionArities, locals)...)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					diags = append(diags, checkFunctionArity(index, functionArities, locals)...)
				}
			}
			diags = append(diags, checkFunctionArity(spec.Value, functionArities, locals)...)
		}
	case *LabelExpr:
		diags = append(diags, checkFunctionArity(e.Body, functionArities, locals)...)
	case *ActionExpr:
		diags = append(diags, checkFunctionArity(e.Action, functionArities, locals)...)
		diags = append(diags, checkFunctionArity(e.Subscript, functionArities, locals)...)
	case *FairnessExpr:
		diags = append(diags, checkFunctionArity(e.Subscript, functionArities, locals)...)
		diags = append(diags, checkFunctionArity(e.Action, functionArities, locals)...)
	case *FunctionSetExpr:
		diags = append(diags, checkFunctionArity(e.Domain, functionArities, locals)...)
		diags = append(diags, checkFunctionArity(e.Range, functionArities, locals)...)
	case *SetComprehensionExpr:
		compLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, checkFunctionArity(bound.Set, functionArities, locals)...)
			compLocals[bound.Name] = true
		}
		diags = append(diags, checkFunctionArity(e.Element, functionArities, compLocals)...)
		if e.Predicate != nil {
			diags = append(diags, checkFunctionArity(e.Predicate, functionArities, compLocals)...)
		}
	}
	return diags
}

func validFunctionApplicationArity(want, got int) bool {
	return got == want || (want > 1 && got == 1)
}

func exprContainsPrime(expr Expr) bool {
	switch e := expr.(type) {
	case *UnaryExpr:
		return e.Op == "'" || exprContainsPrime(e.Expr)
	case *BinaryExpr:
		return exprContainsPrime(e.Left) || exprContainsPrime(e.Right)
	case *CallExpr:
		if exprContainsPrime(e.Callee) {
			return true
		}
		for _, arg := range e.Args {
			if exprContainsPrime(arg) {
				return true
			}
		}
	case *IfExpr:
		return exprContainsPrime(e.Cond) || exprContainsPrime(e.Then) || exprContainsPrime(e.Else)
	case *LetExpr:
		for _, def := range e.Definitions {
			if exprContainsPrime(def.Expr) {
				return true
			}
		}
		return exprContainsPrime(e.Body)
	case *QuantifierExpr:
		return exprContainsPrime(e.Set) || exprContainsPrime(e.Body)
	case *CaseExpr:
		for _, arm := range e.Arms {
			if exprContainsPrime(arm.Test) || exprContainsPrime(arm.Value) {
				return true
			}
		}
		return exprContainsPrime(e.Other)
	case *ChooseExpr:
		return exprContainsPrime(e.Set) || exprContainsPrime(e.Body)
	case *TupleExpr:
		for _, elem := range e.Elems {
			if exprContainsPrime(elem) {
				return true
			}
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			if exprContainsPrime(elem) {
				return true
			}
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			if exprContainsPrime(field.Value) {
				return true
			}
		}
	case *RecordComponentExpr:
		return exprContainsPrime(e.Record)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			if exprContainsPrime(field.Set) {
				return true
			}
		}
	case *FunctionExpr:
		for _, bound := range e.Bounds {
			if exprContainsPrime(bound.Set) {
				return true
			}
		}
		return exprContainsPrime(e.Body)
	case *FunctionAppExpr:
		if exprContainsPrime(e.Function) {
			return true
		}
		for _, arg := range e.Args {
			if exprContainsPrime(arg) {
				return true
			}
		}
	case *ExceptExpr:
		if exprContainsPrime(e.Base) {
			return true
		}
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					if exprContainsPrime(index) {
						return true
					}
				}
			}
			if exprContainsPrime(spec.Value) {
				return true
			}
		}
	case *LabelExpr:
		return exprContainsPrime(e.Body)
	case *ActionExpr:
		return exprContainsPrime(e.Action) || exprContainsPrime(e.Subscript)
	case *FairnessExpr:
		return exprContainsPrime(e.Subscript) || exprContainsPrime(e.Action)
	case *FunctionSetExpr:
		return exprContainsPrime(e.Domain) || exprContainsPrime(e.Range)
	case *SetComprehensionExpr:
		for _, bound := range e.Bounds {
			if exprContainsPrime(bound.Set) {
				return true
			}
		}
		return exprContainsPrime(e.Element) || exprContainsPrime(e.Predicate)
	}
	return false
}

func copyBoolMap(in map[string]bool) map[string]bool {
	out := map[string]bool{}
	for name, ok := range in {
		out[name] = ok
	}
	return out
}

func copyIntMap(in map[string]int) map[string]int {
	out := map[string]int{}
	for name, value := range in {
		out[name] = value
	}
	return out
}

func checkPrimedConstants(expr Expr, declKinds map[string]DeclarationKind, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	switch e := expr.(type) {
	case *UnaryExpr:
		if e.Op == "'" {
			if inner, ok := e.Expr.(*UnaryExpr); ok && inner.Op == "'" {
				diags = append(diags, errorAt(e.Pos, "E1310", "cannot prime an already primed expression"))
			}
			if ident, ok := e.Expr.(*IdentExpr); ok && !locals[ident.Name] && declKinds[ident.Name] == ConstantDecl {
				diags = append(diags, errorAt(e.Pos, "E1303", "cannot prime constant %s", ident.Name))
			}
		}
		diags = append(diags, checkPrimedConstants(e.Expr, declKinds, locals)...)
	case *BinaryExpr:
		diags = append(diags, checkPrimedConstants(e.Left, declKinds, locals)...)
		diags = append(diags, checkPrimedConstants(e.Right, declKinds, locals)...)
	case *CallExpr:
		diags = append(diags, checkPrimedConstants(e.Callee, declKinds, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkPrimedConstants(arg, declKinds, locals)...)
		}
	case *IfExpr:
		diags = append(diags, checkPrimedConstants(e.Cond, declKinds, locals)...)
		diags = append(diags, checkPrimedConstants(e.Then, declKinds, locals)...)
		diags = append(diags, checkPrimedConstants(e.Else, declKinds, locals)...)
	case *LetExpr:
		letLocals := letScopeLocals(locals, e)
		for _, def := range e.Definitions {
			defLocals := letDefinitionBodyLocals(letLocals, def)
			diags = append(diags, checkPrimedConstants(def.Expr, declKinds, defLocals)...)
		}
		diags = append(diags, checkPrimedConstants(e.Body, declKinds, letLocals)...)
	case *QuantifierExpr:
		diags = append(diags, checkPrimedConstants(e.Set, declKinds, locals)...)
		quantLocals := map[string]bool{}
		for name, ok := range locals {
			quantLocals[name] = ok
		}
		quantLocals[e.Var] = true
		diags = append(diags, checkPrimedConstants(e.Body, declKinds, quantLocals)...)
	case *CaseExpr:
		for _, arm := range e.Arms {
			diags = append(diags, checkPrimedConstants(arm.Test, declKinds, locals)...)
			diags = append(diags, checkPrimedConstants(arm.Value, declKinds, locals)...)
		}
		if e.Other != nil {
			diags = append(diags, checkPrimedConstants(e.Other, declKinds, locals)...)
		}
	case *ChooseExpr:
		diags = append(diags, checkPrimedConstants(e.Set, declKinds, locals)...)
		chooseLocals := map[string]bool{}
		for name, ok := range locals {
			chooseLocals[name] = ok
		}
		chooseLocals[e.Var] = true
		diags = append(diags, checkPrimedConstants(e.Body, declKinds, chooseLocals)...)
	case *TupleExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkPrimedConstants(elem, declKinds, locals)...)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkPrimedConstants(elem, declKinds, locals)...)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkPrimedConstants(field.Value, declKinds, locals)...)
		}
	case *RecordComponentExpr:
		diags = append(diags, checkPrimedConstants(e.Record, declKinds, locals)...)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkPrimedConstants(field.Set, declKinds, locals)...)
		}
	case *FunctionExpr:
		fnLocals := map[string]bool{}
		for name, ok := range locals {
			fnLocals[name] = ok
		}
		for _, bound := range e.Bounds {
			diags = append(diags, checkPrimedConstants(bound.Set, declKinds, locals)...)
			fnLocals[bound.Name] = true
		}
		diags = append(diags, checkPrimedConstants(e.Body, declKinds, fnLocals)...)
	case *FunctionAppExpr:
		diags = append(diags, checkPrimedConstants(e.Function, declKinds, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkPrimedConstants(arg, declKinds, locals)...)
		}
	case *ExceptExpr:
		diags = append(diags, checkPrimedConstants(e.Base, declKinds, locals)...)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					diags = append(diags, checkPrimedConstants(index, declKinds, locals)...)
				}
			}
			diags = append(diags, checkPrimedConstants(spec.Value, declKinds, locals)...)
		}
	case *LabelExpr:
		diags = append(diags, checkPrimedConstants(e.Body, declKinds, locals)...)
	case *ActionExpr:
		diags = append(diags, checkPrimedConstants(e.Action, declKinds, locals)...)
		diags = append(diags, checkPrimedConstants(e.Subscript, declKinds, locals)...)
	case *FairnessExpr:
		diags = append(diags, checkPrimedConstants(e.Subscript, declKinds, locals)...)
		diags = append(diags, checkPrimedConstants(e.Action, declKinds, locals)...)
	case *FunctionSetExpr:
		diags = append(diags, checkPrimedConstants(e.Domain, declKinds, locals)...)
		diags = append(diags, checkPrimedConstants(e.Range, declKinds, locals)...)
	case *SetComprehensionExpr:
		compLocals := map[string]bool{}
		for name, ok := range locals {
			compLocals[name] = ok
		}
		for _, bound := range e.Bounds {
			diags = append(diags, checkPrimedConstants(bound.Set, declKinds, locals)...)
			compLocals[bound.Name] = true
		}
		diags = append(diags, checkPrimedConstants(e.Element, declKinds, compLocals)...)
		if e.Predicate != nil {
			diags = append(diags, checkPrimedConstants(e.Predicate, declKinds, compLocals)...)
		}
	}
	return diags
}

func checkAssumptionConstantLevel(expr Expr, declKinds map[string]DeclarationKind) Diagnostics {
	if ident, ok := expr.(*IdentExpr); ok && declKinds[ident.Name] == VariableDecl {
		return Diagnostics{errorAt(ident.Pos, "E1311", "assumption must be constant-level; %s is variable-level", ident.Name)}
	}
	return nil
}

func checkLevelComposition(expr Expr, declKinds map[string]DeclarationKind, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	switch e := expr.(type) {
	case *UnaryExpr:
		if (e.Op == "[]" || e.Op == "<>") && exprLevel(e.Expr, declKinds, locals) == actionLevel {
			if _, wrapped := e.Expr.(*ActionExpr); !wrapped {
				diags = append(diags, errorAt(e.Pos, "E1321", "temporal operator %s cannot be applied directly to an action-level formula", e.Op))
			}
		}
		diags = append(diags, checkLevelComposition(e.Expr, declKinds, locals)...)
	case *BinaryExpr:
		leftLevel := exprLevel(e.Left, declKinds, locals)
		rightLevel := exprLevel(e.Right, declKinds, locals)
		if (e.Op == "~>" || e.Op == "-+->") && (leftLevel == actionLevel || rightLevel == actionLevel) {
			diags = append(diags, errorAt(e.Pos, "E1321", "leads-to operator %s cannot have an action-level operand", e.Op))
		}
		leftLogicalLevel := logicalOperandLevel(e.Left, declKinds, locals)
		rightLogicalLevel := logicalOperandLevel(e.Right, declKinds, locals)
		if isLogicalLevelMixingOperator(e.Op) && levelsMixActionAndTemporal(leftLogicalLevel, rightLogicalLevel) {
			diags = append(diags, errorAt(e.Pos, "E1321", "operator %s cannot mix action and temporal operands", e.Op))
		}
		diags = append(diags, checkLevelComposition(e.Left, declKinds, locals)...)
		diags = append(diags, checkLevelComposition(e.Right, declKinds, locals)...)
	case *CallExpr:
		diags = append(diags, checkLevelComposition(e.Callee, declKinds, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkLevelComposition(arg, declKinds, locals)...)
		}
	case *IfExpr:
		diags = append(diags, checkLevelComposition(e.Cond, declKinds, locals)...)
		diags = append(diags, checkLevelComposition(e.Then, declKinds, locals)...)
		diags = append(diags, checkLevelComposition(e.Else, declKinds, locals)...)
	case *LetExpr:
		letLocals := letScopeLocals(locals, e)
		for _, def := range e.Definitions {
			defLocals := letDefinitionBodyLocals(letLocals, def)
			diags = append(diags, checkLevelComposition(def.Expr, declKinds, defLocals)...)
		}
		diags = append(diags, checkLevelComposition(e.Body, declKinds, letLocals)...)
	case *QuantifierExpr:
		setLevel := exprLevel(e.Set, declKinds, locals)
		bodyLevel := exprLevel(e.Body, declKinds, withLocal(locals, e.Var))
		if setLevel == temporalLevel {
			diags = append(diags, errorAt(e.Pos, "E1321", "quantifier cannot have a temporal-level bound"))
		}
		if setLevel == actionLevel && bodyLevel == temporalLevel {
			diags = append(diags, errorAt(e.Pos, "E1321", "quantifier with a temporal-level body cannot have an action-level bound"))
		}
		diags = append(diags, checkLevelComposition(e.Set, declKinds, locals)...)
		diags = append(diags, checkLevelComposition(e.Body, declKinds, withLocal(locals, e.Var))...)
	case *CaseExpr:
		for _, arm := range e.Arms {
			diags = append(diags, checkLevelComposition(arm.Test, declKinds, locals)...)
			diags = append(diags, checkLevelComposition(arm.Value, declKinds, locals)...)
		}
		if e.Other != nil {
			diags = append(diags, checkLevelComposition(e.Other, declKinds, locals)...)
		}
	case *ChooseExpr:
		diags = append(diags, checkLevelComposition(e.Set, declKinds, locals)...)
		diags = append(diags, checkLevelComposition(e.Body, declKinds, withLocal(locals, e.Var))...)
	case *TupleExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkLevelComposition(elem, declKinds, locals)...)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			diags = append(diags, checkLevelComposition(elem, declKinds, locals)...)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkLevelComposition(field.Value, declKinds, locals)...)
		}
	case *RecordComponentExpr:
		diags = append(diags, checkLevelComposition(e.Record, declKinds, locals)...)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			diags = append(diags, checkLevelComposition(field.Set, declKinds, locals)...)
		}
	case *FunctionExpr:
		fnLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, checkLevelComposition(bound.Set, declKinds, locals)...)
			fnLocals[bound.Name] = true
		}
		diags = append(diags, checkLevelComposition(e.Body, declKinds, fnLocals)...)
	case *FunctionAppExpr:
		diags = append(diags, checkLevelComposition(e.Function, declKinds, locals)...)
		for _, arg := range e.Args {
			diags = append(diags, checkLevelComposition(arg, declKinds, locals)...)
		}
	case *ExceptExpr:
		diags = append(diags, checkLevelComposition(e.Base, declKinds, locals)...)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					diags = append(diags, checkLevelComposition(index, declKinds, locals)...)
				}
			}
			diags = append(diags, checkLevelComposition(spec.Value, declKinds, locals)...)
		}
	case *LabelExpr:
		diags = append(diags, checkLevelComposition(e.Body, declKinds, locals)...)
	case *ActionExpr:
		diags = append(diags, checkLevelComposition(e.Action, declKinds, locals)...)
		diags = append(diags, checkLevelComposition(e.Subscript, declKinds, locals)...)
	case *FairnessExpr:
		diags = append(diags, checkLevelComposition(e.Subscript, declKinds, locals)...)
		diags = append(diags, checkLevelComposition(e.Action, declKinds, locals)...)
	case *FunctionSetExpr:
		diags = append(diags, checkLevelComposition(e.Domain, declKinds, locals)...)
		diags = append(diags, checkLevelComposition(e.Range, declKinds, locals)...)
	case *SetComprehensionExpr:
		compLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			diags = append(diags, checkLevelComposition(bound.Set, declKinds, locals)...)
			compLocals[bound.Name] = true
		}
		diags = append(diags, checkLevelComposition(e.Element, declKinds, compLocals)...)
		if e.Predicate != nil {
			diags = append(diags, checkLevelComposition(e.Predicate, declKinds, compLocals)...)
		}
	}
	return diags
}

func logicalOperandLevel(expr Expr, declKinds map[string]DeclarationKind, locals map[string]bool) tlaLevel {
	if call, ok := expr.(*CallExpr); ok {
		return exprLevel(call.Callee, declKinds, locals)
	}
	return exprLevel(expr, declKinds, locals)
}

func isLogicalLevelMixingOperator(op string) bool {
	switch op {
	case "/\\", "\\/", "=>", "<=>":
		return true
	default:
		return false
	}
}

func levelsMixActionAndTemporal(left, right tlaLevel) bool {
	return (left == actionLevel && right == temporalLevel) || (left == temporalLevel && right == actionLevel)
}

func withLocal(locals map[string]bool, name string) map[string]bool {
	out := copyBoolMap(locals)
	out[name] = true
	return out
}

type tlaLevel int

const (
	constantLevel tlaLevel = iota
	variableLevel
	actionLevel
	temporalLevel
)

func exprLevel(expr Expr, declKinds map[string]DeclarationKind, locals map[string]bool) tlaLevel {
	switch e := expr.(type) {
	case *IdentExpr:
		if locals != nil && locals[e.Name] {
			return constantLevel
		}
		if declKinds[e.Name] == VariableDecl {
			return variableLevel
		}
		return constantLevel
	case *LiteralExpr:
		return constantLevel
	case *UnaryExpr:
		if e.Op == "'" || e.Op == "UNCHANGED" {
			return maxTlaLevel(actionLevel, exprLevel(e.Expr, declKinds, locals))
		}
		if e.Op == "ENABLED" {
			return variableLevel
		}
		if e.Op == "[]" || e.Op == "<>" {
			return temporalLevel
		}
		return exprLevel(e.Expr, declKinds, locals)
	case *BinaryExpr:
		level := maxTlaLevel(exprLevel(e.Left, declKinds, locals), exprLevel(e.Right, declKinds, locals))
		if e.Op == "~>" || e.Op == "-+->" {
			return maxTlaLevel(temporalLevel, level)
		}
		return level
	case *CallExpr:
		level := exprLevel(e.Callee, declKinds, locals)
		for _, arg := range e.Args {
			level = maxTlaLevel(level, exprLevel(arg, declKinds, locals))
		}
		return level
	case *IfExpr:
		return maxTlaLevel(exprLevel(e.Cond, declKinds, locals), maxTlaLevel(exprLevel(e.Then, declKinds, locals), exprLevel(e.Else, declKinds, locals)))
	case *LetExpr:
		letLocals := letScopeLocals(locals, e)
		return exprLevel(e.Body, declKinds, letLocals)
	case *QuantifierExpr:
		quantLocals := copyBoolMap(locals)
		quantLocals[e.Var] = true
		return maxTlaLevel(exprLevel(e.Set, declKinds, locals), exprLevel(e.Body, declKinds, quantLocals))
	case *CaseExpr:
		level := constantLevel
		for _, arm := range e.Arms {
			level = maxTlaLevel(level, exprLevel(arm.Test, declKinds, locals))
			level = maxTlaLevel(level, exprLevel(arm.Value, declKinds, locals))
		}
		if e.Other != nil {
			level = maxTlaLevel(level, exprLevel(e.Other, declKinds, locals))
		}
		return level
	case *ChooseExpr:
		chooseLocals := copyBoolMap(locals)
		chooseLocals[e.Var] = true
		return maxTlaLevel(exprLevel(e.Set, declKinds, locals), exprLevel(e.Body, declKinds, chooseLocals))
	case *TupleExpr:
		level := constantLevel
		for _, elem := range e.Elems {
			level = maxTlaLevel(level, exprLevel(elem, declKinds, locals))
		}
		return level
	case *SetExpr:
		level := constantLevel
		for _, elem := range e.Elems {
			level = maxTlaLevel(level, exprLevel(elem, declKinds, locals))
		}
		return level
	case *RecordExpr:
		level := constantLevel
		for _, field := range e.Fields {
			level = maxTlaLevel(level, exprLevel(field.Value, declKinds, locals))
		}
		return level
	case *RecordComponentExpr:
		return exprLevel(e.Record, declKinds, locals)
	case *RecordSetExpr:
		level := constantLevel
		for _, field := range e.Fields {
			level = maxTlaLevel(level, exprLevel(field.Set, declKinds, locals))
		}
		return level
	case *FunctionExpr:
		fnLocals := copyBoolMap(locals)
		level := constantLevel
		for _, bound := range e.Bounds {
			level = maxTlaLevel(level, exprLevel(bound.Set, declKinds, locals))
			fnLocals[bound.Name] = true
		}
		return maxTlaLevel(level, exprLevel(e.Body, declKinds, fnLocals))
	case *FunctionAppExpr:
		level := exprLevel(e.Function, declKinds, locals)
		for _, arg := range e.Args {
			level = maxTlaLevel(level, exprLevel(arg, declKinds, locals))
		}
		return level
	case *ExceptExpr:
		level := exprLevel(e.Base, declKinds, locals)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					level = maxTlaLevel(level, exprLevel(index, declKinds, locals))
				}
			}
			level = maxTlaLevel(level, exprLevel(spec.Value, declKinds, locals))
		}
		return level
	case *LabelExpr:
		return exprLevel(e.Body, declKinds, locals)
	case *ActionExpr:
		return maxTlaLevel(actionLevel, maxTlaLevel(exprLevel(e.Action, declKinds, locals), exprLevel(e.Subscript, declKinds, locals)))
	case *FairnessExpr:
		return temporalLevel
	case *FunctionSetExpr:
		return maxTlaLevel(exprLevel(e.Domain, declKinds, locals), exprLevel(e.Range, declKinds, locals))
	case *SetComprehensionExpr:
		compLocals := copyBoolMap(locals)
		level := constantLevel
		for _, bound := range e.Bounds {
			level = maxTlaLevel(level, exprLevel(bound.Set, declKinds, locals))
			compLocals[bound.Name] = true
		}
		level = maxTlaLevel(level, exprLevel(e.Element, declKinds, compLocals))
		if e.Predicate != nil {
			level = maxTlaLevel(level, exprLevel(e.Predicate, declKinds, compLocals))
		}
		return level
	default:
		return constantLevel
	}
}

func maxTlaLevel(a, b tlaLevel) tlaLevel {
	if a > b {
		return a
	}
	return b
}

var builtinIdentifiers = map[string]bool{
	"TRUE":    true,
	"FALSE":   true,
	"BOOLEAN": true,
	"STRING":  true,
}
