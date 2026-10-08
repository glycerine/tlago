package tlago

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

func CheckSpec(spec *Spec) Diagnostics {
	diags := checkSpecWithProgress(spec, nil)
	if !diags.HasErrors() {
		diags = append(diags, lintSanySpec(spec, nil)...)
	}
	return diags
}

func checkSpecWithProgress(spec *Spec, progress func(string)) Diagnostics {
	return checkSpecWithModuleReport(spec, progress, nil)
}

// SANY shares one Errors instance across external modules and reports its
// accumulated contents after each module. Reprinting is not reinsertion.
func checkSpecWithModuleReport(spec *Spec, progress func(string), report func(Diagnostics)) Diagnostics {
	return generateSpecWithModuleReport(spec, progress, report, true)
}

// GenerateSanySpec runs Generator without ModuleNode.levelCheck, as the
// programmatic SANYFrontend.processSemantics phase does. The driver retains
// Java SANY's per-external-module generation/level-check sequence.
func GenerateSanySpec(spec *Spec) Diagnostics {
	if spec != nil {
		spec.initialContext = sanyGlobalInitialContext(true)
	}
	return generateSpecWithModuleReport(spec, nil, nil, false)
}

// CheckSanySpecLevels checks the generated root, retaining ModuleNode's cached
// levelCorrect result independently of the Errors supplied to each invocation.
func CheckSanySpecLevels(spec *Spec) (bool, Diagnostics) {
	if spec == nil || spec.Root == nil || spec.Root.semanticNode == nil {
		panic(tlc.NewNullPointerException())
	}
	var diags Diagnostics
	levelOK := sanyLevelCheckNext(spec.Root.semanticNode, &diags)
	return levelOK && !diags.HasErrors(), diags
}

func generateSpecWithModuleReport(spec *Spec, progress func(string), report func(Diagnostics), checkLevels bool) Diagnostics {
	if spec == nil {
		diags := Diagnostics{errorAt(Position{}, "E1300", "nil spec")}
		if report != nil {
			report(diags)
		}
		return diags
	}
	if spec.initialContext == nil {
		spec.initialContext = sanyGlobalInitialContext(false)
	}
	var diags Diagnostics
	spec.levelChecks = map[*Module]*sanyModuleLevelChecks{}
	spec.semanticModules = newSanyExternalModuleTable()
	spec.SemanticDiags = nil
	defer func() {
		spec.SemanticDiags = appendSanyDiagnostics(append(Diagnostics(nil), diags...), spec.SemanticDiags...)
	}()
	resolver := &sanySelectorResolver{spec: spec, scopes: map[*Module]map[string]sanySelectorDefinition{}, visiting: map[*Module]bool{}}
	enclosing := enclosingModules(spec)
	checked := make(map[*Module]bool, len(spec.Modules))
	var check func(*Module, *sanyModuleRecursiveGeneration) (*sanyModuleLevelChecks, Diagnostics)
	check = func(mod *Module, recursive *sanyModuleRecursiveGeneration) (*sanyModuleLevelChecks, Diagnostics) {
		if mod == nil || checked[mod] {
			return nil, nil
		}
		checked[mod] = true
		if parent := enclosing[mod]; parent != nil {
			mod.generatorNodes = parent.generatorNodes
		} else {
			mod.generatorNodes = newSanyGeneratorNodes()
		}
		checks := &sanyModuleLevelChecks{generator: resolver.moduleGenerator(mod), recursiveGeneration: recursive, operatorChecks: map[string]*sanyCachedLevelCheck{}}
		spec.levelChecks[mod] = checks
		// Nested graphs are generated at their module unit, sharing the external
		// module's reporting iteration. Their diagnostics remain in body order.
		checks.generateNested = func(nested *Module) Diagnostics {
			child, childDiags := check(nested, checks.recursiveGeneration)
			if child != nil {
				checks.nested = append(checks.nested, child)
			}
			return childDiags
		}
		generated := generateModuleWithEnclosing(mod, spec, enclosing[mod], checks)
		if !checkLevels {
			for _, entry := range tlcBridgeContextEntries(spec, mod, map[*Module]bool{}) {
				if entry.module != nil && entry.module != mod && entry.instance == nil {
					if parent := spec.levelChecks[entry.module]; parent != nil {
						if node := parent.operatorChecks[entry.name]; node != nil {
							checks.importedChecks = append(checks.importedChecks, node)
						}
					}
				}
			}
			// ModuleNode.copyTopLevel shares inherited assumptions, theorems
			// and instances. Definition nodes come from the imported Context.
			var inherited []sanyLevelCheck
			for _, name := range mod.Extends {
				if parent := spec.levelChecks[spec.Modules[name]]; parent != nil {
					inherited = append(inherited, parent.topLevel...)
				}
			}
			checks.topLevel = append(inherited, checks.topLevel...)
			checks.topLevelOrdered = true
		}
		return checks, generated
	}
	for index, name := range spec.SemanticOrder {
		mod := spec.Modules[name]
		if mod == nil || checked[mod] {
			continue
		}
		if progress != nil {
			progress("Semantic processing of module " + name)
		}
		_, generated := check(mod, nil)
		diags = appendSanyDiagnostics(diags, generated...)
		// SANY assigns this external module's standard provenance after
		// generation. The resolver call can itself throw during semantics.
		if spec.FilenameResolver != nil {
			mod.Library = spec.FilenameResolver.IsStandardModule(name)
		}
		if mod.semanticNode != nil {
			spec.semanticModules.put(name, mod.semanticNode.context, mod.semanticNode)
		}
		// Source tests raw Errors.isSuccess, before warning elevation.
		if checkLevels && mod.semanticNode != nil && !diags.HasErrors() {
			sanyLevelCheckNext(mod.semanticNode, &diags)
		}
		// SANY publishes the last analyzed external module as the root after
		// its level check, independently of the accumulated error result.
		if index == len(spec.SemanticOrder)-1 {
			spec.semanticModules.root = mod.semanticNode
		}
		if report != nil {
			report(diags)
		}
	}
	// Native callers without loader order still need deterministic checking.
	var remaining []string
	for name, mod := range spec.Modules {
		if !checked[mod] {
			remaining = append(remaining, name)
		}
	}
	sort.Strings(remaining)
	for _, name := range remaining {
		mod := spec.Modules[name]
		_, generated := check(mod, nil)
		diags = appendSanyDiagnostics(diags, generated...)
		if mod := spec.Modules[name]; enclosing[mod] == nil && mod.semanticNode != nil {
			spec.semanticModules.put(name, mod.semanticNode.context, mod.semanticNode)
		}
		if checkLevels && mod.semanticNode != nil && !diags.HasErrors() {
			sanyLevelCheckNext(mod.semanticNode, &diags)
		}
		if mod == spec.Root {
			spec.semanticModules.root = mod.semanticNode
		}
	}
	if report != nil && (len(remaining) > 0 || len(spec.SemanticOrder) == 0) {
		report(diags)
	}
	return diags
}

func checkModule(mod *Module, spec *Spec) Diagnostics {
	if spec.semanticModules == nil {
		spec.semanticModules = newSanyExternalModuleTable()
	}
	var diags Diagnostics
	// The single-module entry point still needs the preceding external
	// semantic graphs that SANY supplies to Generator in loader order.
	for _, name := range spec.SemanticOrder {
		if name == mod.Name {
			break
		}
		dependency := spec.Modules[name]
		if dependency == nil {
			continue
		}
		if dependency.semanticNode == nil {
			diags = appendSanyDiagnostics(diags, generateModuleWithEnclosing(dependency, spec, nil, &sanyModuleLevelChecks{})...)
		}
		if dependency.semanticNode != nil {
			spec.semanticModules.put(name, dependency.semanticNode.context, dependency.semanticNode)
		}
	}
	return appendSanyDiagnostics(diags, checkModuleWithEnclosing(mod, spec, nil)...)
}

func enclosingModules(spec *Spec) map[*Module]*Module {
	out := map[*Module]*Module{}
	if spec == nil {
		return out
	}
	var walk func(parent *Module)
	walk = func(parent *Module) {
		if parent == nil {
			return
		}
		for _, nested := range parent.Nested {
			if nested == nil {
				continue
			}
			out[nested] = parent
			walk(nested)
		}
	}
	for _, mod := range spec.Modules {
		walk(mod)
	}
	return out
}

func appendSanyDiagnostics(diags Diagnostics, added ...Diagnostic) Diagnostics {
	if len(added) == 0 {
		return diags
	}
	for _, diagnostic := range added {
		present := false
		for _, stored := range diags {
			if sameSanyLoggedDiagnostic(stored, diagnostic) {
				present = true
				break
			}
		}
		if !present {
			diags = append(diags, diagnostic)
		}
	}
	return diags
}

type sanyLevelCheck struct {
	position Position
	node     *sanyCachedLevelCheck
}

type sanyCachedLevelCheck struct {
	run              func() (bool, Diagnostics)
	checked, correct bool
	repeatSuccess    bool
}

func (node *sanyCachedLevelCheck) check() (bool, Diagnostics) {
	if node.checked {
		return node.correct || node.repeatSuccess, nil
	}
	node.checked, node.correct = true, true
	correct, diags := node.run()
	node.correct = correct
	return node.correct, diags
}

// Generator's module recursion shares curLevel zero, unresolvedCnt[0] and
// unresolvedSum with nested modules. checkForUndefinedRecursiveOps subtracts
// the count from the sum without clearing the count, including on invalid input.
type sanyModuleRecursiveGeneration struct {
	counts  [100]int
	section int
	count   int
	sum     int
}

func (state *sanyModuleRecursiveGeneration) complete() {
	state.count--
	state.sum--
	if state.sum < 0 {
		panic(tlc.NewWrongInvocationException("Defined more recursive operators than were declared in RECURSIVE statements."))
	}
}

type sanyModuleLevelChecks struct {
	levelChecked        bool
	levelCorrect        bool
	generator           *sanyModuleSelectorGenerator
	recursiveGeneration *sanyModuleRecursiveGeneration
	generateNested      func(*Module) Diagnostics
	recursive           []func() Diagnostics
	definitions         []*sanyCachedLevelCheck
	facts               []*sanyCachedLevelCheck
	topLevel            []sanyLevelCheck
	nested              []*sanyModuleLevelChecks
	operatorChecks      map[string]*sanyCachedLevelCheck
	importedChecks      []*sanyCachedLevelCheck
	topLevelOrdered     bool
}

func (checks *sanyModuleLevelChecks) check() Diagnostics {
	if checks.levelChecked {
		return nil
	}
	checks.levelChecked, checks.levelCorrect = true, true
	var diags Diagnostics

	for _, run := range checks.recursive {
		current := run()
		diags = appendSanyDiagnostics(diags, current...)
		checks.levelCorrect = checks.levelCorrect && !current.HasErrors()
	}
	for _, nested := range checks.nested {
		diags = appendSanyDiagnostics(diags, nested.check()...)
		checks.levelCorrect = checks.levelCorrect && nested.levelCorrect
	}
	for _, nodes := range [][]*sanyCachedLevelCheck{checks.importedChecks, checks.definitions, checks.facts} {
		for _, node := range nodes {
			correct, current := node.check()
			checks.levelCorrect = checks.levelCorrect && correct
			diags = appendSanyDiagnostics(diags, current...)
		}
	}
	// Legacy plans order retained proof summaries by their source position.
	// Programmatic generation already prepends copied inherited nodes.
	if !checks.topLevelOrdered {
		sort.SliceStable(checks.topLevel, func(i, j int) bool { return checks.topLevel[i].position.Compare(checks.topLevel[j].position) < 0 })
	}
	for _, node := range checks.topLevel {
		correct, current := node.node.check()
		checks.levelCorrect = checks.levelCorrect && correct
		diags = appendSanyDiagnostics(diags, current...)
	}
	return diags
}

func checkModuleWithEnclosing(mod *Module, spec *Spec, enclosing *Module) Diagnostics {
	checks := &sanyModuleLevelChecks{}
	diags := generateModuleWithEnclosing(mod, spec, enclosing, checks)
	if mod.semanticNode != nil && !diags.HasErrors() {
		sanyLevelCheckNext(mod.semanticNode, &diags)
	}
	return diags
}

func generateModuleWithEnclosing(mod *Module, spec *Spec, enclosing *Module, checks *sanyModuleLevelChecks) Diagnostics {
	var diags Diagnostics
	// Errors belong to SpecObj, independently of a completed ModuleNode. On an
	// exception, preserve the parent's earlier units before the child's errors.
	defer func() {
		if failure := recover(); failure != nil {
			spec.SemanticDiags = appendSanyDiagnostics(append(Diagnostics(nil), diags...), spec.SemanticDiags...)
			panic(failure)
		}
	}()
	if mod.generatorNodes == nil {
		if enclosing != nil {
			mod.generatorNodes = enclosing.generatorNodes
		}
		if mod.generatorNodes == nil {
			mod.generatorNodes = newSanyGeneratorNodes()
		}
	}
	if checks.generator == nil {
		resolver := &sanySelectorResolver{spec: spec, scopes: map[*Module]map[string]sanySelectorDefinition{}, visiting: map[*Module]bool{}}
		checks.generator = resolver.moduleGenerator(mod)
	}
	if checks.recursiveGeneration == nil {
		checks.recursiveGeneration = &sanyModuleRecursiveGeneration{}
	}
	if checks.generateNested == nil {
		checks.generateNested = func(nested *Module) Diagnostics {
			child := &sanyModuleLevelChecks{generator: checks.generator.resolver.moduleGenerator(nested), recursiveGeneration: checks.recursiveGeneration}
			checks.nested = append(checks.nested, child)
			return generateModuleWithEnclosing(nested, spec, mod, child)
		}
	}
	context := newSanyContext()
	if enclosing == nil {
		if spec.initialContext == nil {
			spec.initialContext = sanyGlobalInitialContext(false)
		}
		context = spec.initialContext.duplicate()
	}
	mod.semanticNode = newSanySemModuleNode(mod.Name, context, mod.Pos, mod.Syntax)
	if enclosing != nil && enclosing.symbolTable != nil {
		mod.symbolTable = enclosing.symbolTable.duplicateForInnerModule()
		mod.symbolTable.pushContext(context)
	} else {
		if spec.semanticModules == nil {
			spec.semanticModules = newSanyExternalModuleTable()
		}
		mod.symbolTable = newSanySymbolTable(context, spec.semanticModules)
	}
	mod.symbolTable.module = mod.semanticNode
	mod.semanticNode.nestingLevel = 0
	if enclosing != nil && enclosing.semanticNode != nil {
		mod.semanticNode.nestingLevel = enclosing.semanticNode.nestingLevel + 1
		enclosing.semanticNode.definitions = append(enclosing.semanticNode.definitions, mod.semanticNode)
	}
	expressionGeneration := &sanyExpressionGeneration{nodes: mod.generatorNodes, spec: spec, currentModule: mod, module: checks.recursiveGeneration, bindings: map[string]*sanyRecursiveBinding{}}
	checkExpr := func(expr Expr, context map[string]Position, locals map[string]bool) Diagnostics {
		return expressionGeneration.checkExpr(expr, context, locals)
	}
	defined := map[string]Position{}
	declKinds := map[string]DeclarationKind{}
	arities := map[string]int{}
	expressionGeneration.moduleKinds = declKinds
	expressionGeneration.moduleArities = arities
	expressionGeneration.moduleSymbols = map[string]localSymbol{}
	functionArities := map[string]int{}
	operatorParamSpecs := map[string][]operatorParamSpec{}
	expressionGeneration.moduleOperatorParams = operatorParamSpecs
	extendedSymbols := map[string]importedSymbol{}
	enclosingBindings := map[string]importedSymbol{}
	diags = append(diags, checkPlusCalChecksumWarnings(mod)...)
	addName := func(name string, pos Position, kind DeclarationKind) {
		if builtinIdentifiers[name] && !isEmbeddedStandardModule(mod) {
			diags = append(diags, sanyDiagnosticParameters(errorAt(pos, "E4202", "cannot redefine built-in symbol %s", name), name))
			return
		}
		if prev, ok := defined[name]; ok {
			if prevKind, ok := declKinds[name]; ok && prevKind != "" && kind != "" && prevKind != kind {
				diags = append(diags, errorAt(pos, "E4201", "duplicate declaration or definition %s; existing symbol class %s conflicts with %s at %s", name, prevKind, kind, prev))
				return
			}
			diags = append(diags, errorAt(pos, "E4201", "duplicate declaration or definition %s; first declared at %s", name, prev))
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
				recordImportedSymbol(name, d.Kind, declarationSymbolPosition(d, name), depMod.Name, extendedSymbols)
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
			recordImportedSymbol(def.Name, semanticDefinitionImportKind(def), def.SourcePosition(), depMod.Name, extendedSymbols)
			if _, exists := defined[def.Name]; !exists {
				defined[def.Name] = def.Pos
			}
			if _, exists := arities[def.Name]; !exists {
				arities[def.Name] = len(def.Params)
			}
			addSubexpressionReferenceNames(defined, def.Name, def.Expr)
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
			recordImportedSymbol(assumption.Name, semanticTheoremImportKind, pos, depMod.Name, extendedSymbols)
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
	}
	if enclosing != nil {
		for _, symbol := range semanticModuleExports(enclosing, spec, map[string]bool{}) {
			if positionInsideModule(symbol.pos, mod) {
				continue
			}
			if !enclosingSymbolVisibleBeforeNested(symbol, enclosing, mod, spec) {
				continue
			}
			addSemanticSymbol(symbol, defined, declKinds, arities, operatorParamSpecs)
			binding := importedSymbol{kind: symbol.importKind(), pos: symbol.sourcePosition(), arity: symbol.arity}
			if symbol.kind == ConstantDecl || symbol.kind == VariableDecl {
				binding.declarationNode = sanyModuleDeclarationNode(symbol.origin, symbol.name)
				if binding.declarationNode != nil {
					binding.pos = binding.declarationNode.semPosition()
					binding.arity = binding.declarationNode.semArity()
				}
			}
			enclosingBindings[symbol.name] = binding
		}
	}
	extendees := make([]*sanySemModuleNode, 0, len(mod.Extends))
	for depIndex, dep := range mod.Extends {
		extendee := mod.symbolTable.resolveModule(dep)
		if extendee == nil {
			diagnostic := sanyRegistrationDiagnostic(sanyExtendeePosition(mod, depIndex), "E4003", "Could not find module %s", tlc.UniqueStringOf(dep))
			diags = appendSanyDiagnostics(diags, diagnostic)
			panic(newSanySemanticAbort(diagnostic, nil, &spec.SemanticDiags))
		}
		var mergeDiagnostics Diagnostics
		if extendee.context != nil {
			_, mergeDiagnostics = mod.semanticNode.context.mergeExtendContext(extendee.context)
		} else {
			position := sanyExtendeePosition(mod, depIndex)
			mergeDiagnostics = Diagnostics{sanyRegistrationDiagnostic(position, "E4003", "Couldn't find context for module `%s'.", tlc.UniqueStringOf(dep))}
		}
		for i := range mergeDiagnostics {
			if mergeDiagnostics[i].Code == "E4224" {
				// Keep the native API's explanatory prefix; ErrorDetails use
				// the exact canonical message and parameters from Context.
				name := mergeDiagnostics[i].SANYParameters[1].(string)
				mergeDiagnostics[i].Message = "conflicting imported symbol " + name + ": " + mergeDiagnostics[i].SANYMessage
			}
		}
		diags = appendSanyDiagnostics(diags, mergeDiagnostics...)
		extendees = append(extendees, extendee)
		mod.semanticNode.copyAssumes(extendee)
		mod.semanticNode.copyTheorems(extendee)
		mod.semanticNode.copyTopLevel(extendee)
		if depMod := spec.Modules[dep]; depMod != nil {
			for _, inherited := range transitiveExtendedModules(spec, depMod, map[string]bool{depMod.Name: true}) {
				importInheritedModule(inherited)
			}
			for _, d := range depMod.Declarations {
				for _, name := range d.Names {
					recordImportedSymbol(name, d.Kind, declarationSymbolPosition(d, name), depMod.Name, extendedSymbols)
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
				recordImportedSymbol(def.Name, semanticDefinitionImportKind(def), def.SourcePosition(), depMod.Name, extendedSymbols)
				if _, exists := defined[def.Name]; !exists {
					defined[def.Name] = def.Pos
				}
				if _, exists := arities[def.Name]; !exists {
					arities[def.Name] = len(def.Params)
				}
				addSubexpressionReferenceNames(defined, def.Name, def.Expr)
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
				recordImportedSymbol(assumption.Name, semanticTheoremImportKind, pos, depMod.Name, extendedSymbols)
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
			for _, inst := range depMod.Instances {
				if inst.Local {
					continue
				}
				for _, symbol := range semanticInstanceSymbols(inst, spec) {
					recordImportedSymbol(symbol.name, symbol.importKind(), symbol.sourcePosition(), depMod.Name, extendedSymbols)
					addSemanticSymbol(symbol, defined, declKinds, arities, operatorParamSpecs)
				}
			}
		}
	}
	mod.semanticNode.createExtendeeArray(extendees)
	// Local symbols enter the generation context in module-body order.
	expressionContexts := sanyModuleExpressionContexts(mod, spec, defined)
	instanceSymbols := enclosingBindings
	// SymbolTable.resolveSymbol sees the already merged EXTENDS context.
	for name, symbol := range extendedSymbols {
		symbol.arity = arities[name]
		if symbol.kind == ConstantDecl || symbol.kind == VariableDecl {
			symbol.declarationNode = sanyModuleDeclarationNode(mod, name)
			if symbol.declarationNode != nil {
				symbol.pos = symbol.declarationNode.semPosition()
				symbol.arity = symbol.declarationNode.semArity()
			}
		}
		instanceSymbols[name] = symbol
	}
	for name, symbol := range instanceSymbols {
		expressionGeneration.moduleSymbols[name] = retainSanyInstanceSymbol(localSymbol{declarationNode: symbol.declarationNode, kind: symbol.kind, arity: symbol.arity, pos: symbol.pos}, mod.symbolTable.resolveSymbol(name))
	}
	localSymbols := moduleOwnSymbols(mod)
	registerInstance := func(inst Instance) {
		if inst.Name != "" {
			before := len(diags)
			addName(inst.Name, inst.SourcePosition(), InstanceDecl)
			if inst.definitionNode != nil {
				// The actual constructor has already reported any rejected
				// registration. Keep native metadata without a second error.
				diags = diags[:before]
			}
			defined[instanceNameSentinel(inst.Name)] = inst.SourcePosition()
			declKinds[inst.Name] = InstanceDecl
			arities[inst.Name] = len(inst.Params)
			specs := make([]operatorParamSpec, len(inst.Params))
			for i, name := range inst.Params {
				specs[i] = operatorParamSpec{Name: name, Arity: -1}
				if arity := inst.ParamArities[name]; arity > 0 {
					specs[i].Arity = arity
				}
			}
			operatorParamSpecs[inst.Name] = specs
		}
		registered := addInstanceSymbols(inst, spec, defined, declKinds, arities, operatorParamSpecs, instanceSymbols, localSymbols)
		if inst.semanticNode == nil {
			diags = append(diags, registered...)
		}
		for name, symbol := range instanceSymbols {
			if _, exists := expressionGeneration.moduleSymbols[name]; !exists {
				expressionGeneration.moduleSymbols[name] = localSymbol{declarationNode: symbol.declarationNode, kind: symbol.kind, arity: symbol.arity, pos: symbol.pos}
			}
		}
		if inst.semanticNode != nil {
			if inst.definitionNode != nil {
				expressionGeneration.moduleSymbols[inst.Name] = retainSanyInstanceSymbol(expressionGeneration.moduleSymbols[inst.Name], mod.symbolTable.resolveSymbol(inst.Name))
			}
			for _, symbol := range expressionGeneration.instanceSymbols(inst) {
				actual := mod.symbolTable.resolveSymbol(symbol.name)
				expressionGeneration.moduleSymbols[symbol.name] = retainSanyInstanceSymbol(expressionGeneration.moduleSymbols[symbol.name], actual)
			}
		}
	}
	mod.declarationNodes = nil
	registerDeclaration := func(d Declaration) {
		seenInDecl := map[string]bool{}
		for i, name := range d.Names {
			var syntax *SanySyntaxNode
			if d.Syntax != nil {
				syntax = d.Syntax.GetHeirs()[2*i+1]
			}
			kind, level := sanyConstantDeclKind, constantLevel
			if d.Kind == VariableDecl {
				kind, level = sanyVariableDeclKind, variableLevel
			}
			arity, _ := declarationArity(d, name)
			if syntax != nil {
				arity = 0
				if d.Kind == ConstantDecl {
					switch syntax.Kind.JavaName() {
					case "N_IdentDecl":
						arity = (len(syntax.GetHeirs()) - 1) / 2
					case "N_PrefixDecl", "N_PostfixDecl":
						arity = 1
					case "N_InfixDecl":
						arity = 2
					}
				}
			}
			node := newSanySemOpDeclNode(sanyCanonicalOperatorImage(name), kind, level, arity, mod.semanticNode, syntax)
			node.table = mod.symbolTable
			mod.declarationNodes = append(mod.declarationNodes, node)
			if syntax != nil {
				if _, exists := mod.symbolTable.resolveSymbol(node.semName()).(*sanySemOpDeclNode); exists {
					// Declaration registration keeps the existing binding even
					// when a same-kind/arity duplicate is only a warning.
					diags = append(diags, mod.symbolTable.addSymbol(node)...)
					continue
				}
			}
			if seenInDecl[name] {
				diags = append(diags, errorAt(d.Pos, "E4201", "duplicate declaration %s", name))
				continue
			}
			seenInDecl[name] = true
			position := d.Pos
			if syntax != nil {
				position = node.semPosition()
			}
			if _, exists := defined[name]; !exists {
				expressionGeneration.moduleSymbols[name] = localSymbol{declarationNode: node, kind: d.Kind, arity: arity, pos: position}
				if mod.semanticNode.context.getSymbol(node.semName()) == nil {
					mod.semanticNode.context.addSymbol(node)
				}
			}
			addName(name, position, d.Kind)
			declKinds[name] = d.Kind
			if d.Kind == ConstantDecl {
				if _, ok := declarationArity(d, name); ok {
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
	var recursiveOrder []string
	completedRecursive := map[string]bool{}
	satisfiedRecursive := map[string]bool{}
	registerRecursive := func(d Declaration) {
		if checks.recursiveGeneration.sum == 0 {
			checks.recursiveGeneration.section++
		}
		checks.recursiveGeneration.count += len(d.Names)
		checks.recursiveGeneration.sum += len(d.Names)
		seenInDecl := map[string]bool{}
		for i, name := range d.Names {
			node, generated := expressionGeneration.constructRecursiveDeclaration(d, i, defined)
			diags = append(diags, generated...)
			if node != nil {
				expressionGeneration.declarations = append(expressionGeneration.declarations, &sanyRecursiveBinding{node: node, name: name, position: node.semPosition(), arity: node.semArity(), level: 0})
				if expressionGeneration.formalSymbolTable().resolveSymbol(node.semName()) != node {
					continue
				}
			}
			if seenInDecl[name] {
				diags = append(diags, sanyDiagnosticParameters(errorAt(d.Pos, "E4291", "duplicate recursive declaration %s", name), name))
				continue
			}
			seenInDecl[name] = true
			if prev, ok := defined[name]; ok {
				diags = append(diags, errorAt(d.Pos, "E4294", "recursive declaration %s conflicts with declaration or definition at %s", name, prev))
				continue
			}
			defined[name] = d.Pos
			declKinds[name] = RecursiveDecl
			arity, ok := declarationArity(d, name)
			if !ok {
				arity = 0
			}
			recursiveArities[name] = arity
			recursivePositions[name] = declarationSymbolPosition(d, name)
			recursiveOrder = append(recursiveOrder, name)
			expressionGeneration.bindings[name] = &sanyRecursiveBinding{node: node, name: name, position: recursivePositions[name], arity: arity, level: 0}
			if node == nil {
				expressionGeneration.declarations = append(expressionGeneration.declarations, expressionGeneration.bindings[name])
			}
			arities[name] = arity
		}
	}
	constructorConflicts := map[string]localSymbol{}
	registerDefinition := func(def Definition) {
		recursiveArity, declaredRecursive := recursiveArities[def.Name]
		binding := expressionGeneration.bindings[def.Name]
		recursive := declaredRecursive && !completedRecursive[def.Name] && (binding == nil || !binding.defined)
		rejectedRecursiveFunction := recursive && def.FunctionDef && recursiveArity != 0
		diags = append(diags, checkDefinitionParams(def)...)
		diags = append(diags, checkDefinitionParamCollisions(def, defined, nil)...)
		if recursive {
			want := recursiveArity
			// processFunction only completes a RECURSIVE declaration of arity
			// zero. A rejected declaration remains undefined at module end.
			satisfiedRecursive[def.Name] = !def.FunctionDef || want == 0
			if got := len(def.Params); got != want {
				diagnostic := sanyDiagnosticParameters(errorAt(def.Pos, "E4292", "Definition of %s has different arity than its RECURSIVE declaration. The operator %s requires %d arguments.", def.Name, def.Name, want), def.Name)
				diagnostic.SANYMessage = fmt.Sprintf("Definition of %s has different arity than its RECURSIVE declaration.", def.Name)
				if def.Syntax != nil {
					diagnostic.SANYRange = def.Syntax.Range
				}
				if def.FunctionDef {
					diagnostic.SANYMessage = fmt.Sprintf("Function %s has operator arguments in its RECURSIVE declaration.", def.Name)
				}
				diags = append(diags, diagnostic)
			}
		} else if !definitionSatisfiesSymbolicConstantDeclaration(def, declKinds, arities) {
			previous, exists := expressionGeneration.lookupSymbol(def.Name, defined)
			if function, ok := def.Expr.(*FunctionExpr); ok && def.FunctionDef {
				previous, exists = function.constructorSymbol, function.constructorSymbolExists
			}
			if exists {
				position := def.SourcePosition()
				diagnostic := errorAt(def.Pos, "E4201", "duplicate declaration or definition %s; first declared at %s", def.Name, previous.pos)
				diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
				diagnostic.SANYMessage = fmt.Sprintf("Operator %s already defined or declared.", def.Name)
				diagnostic.SANYParameters = []any{def.Name}
				if def.FunctionDef {
					diagnostic.SANYMessage = fmt.Sprintf("Function name `%s' already defined or declared.", def.Name)
				}
				diags = append(diags, diagnostic)
				constructorConflicts[positionKey(position)] = previous
				return
			} else {
				addName(def.Name, def.Pos, OperatorDecl)
			}
		}
		if binding := expressionGeneration.bindings[def.Name]; binding != nil {
			if !def.FunctionDef && binding.node == nil {
				binding.arity = len(def.Params)
			}
			if def.FunctionDef && binding.arity == 0 && binding.node == nil && !completedRecursive[def.Name] {
				checks.recursiveGeneration.complete()
				completedRecursive[def.Name] = true
				binding.defined = true
				binding.position = def.SourcePosition()
			}
		}
		// A rejected recursive function leaves the original declaration in the
		// context, including its operator arity and undefined body.
		if rejectedRecursiveFunction {
			return
		}
		expressionGeneration.moduleSymbols[def.Name] = localSymbol{kind: semanticDefinitionImportKind(def), arity: len(def.Params), pos: def.SourcePosition()}
		arities[def.Name] = len(def.Params)
		if binding := expressionGeneration.bindings[def.Name]; binding != nil && binding.node != nil {
			arities[def.Name] = binding.node.semArity()
			expressionGeneration.moduleSymbols[def.Name] = localSymbol{opDefNode: binding.node, kind: OperatorDecl, arity: binding.node.semArity(), pos: def.SourcePosition()}
		}
		addSubexpressionReferenceNames(defined, def.Name, def.Expr)
		if specs, ok := definitionOperatorParamSpecs(def); ok {
			operatorParamSpecs[def.Name] = specs
		}
		if arity, ok := definitionFunctionArity(def); ok {
			functionArities[def.Name] = arity
		}
	}
	finishDefinition := func(def Definition) {
		if def.semanticNode != nil {
			return
		}
		if previous, exists := constructorConflicts[positionKey(def.SourcePosition())]; exists && !def.FunctionDef {
			symbol := semanticExportedSymbol{name: def.Name, kind: OperatorDecl, theoremLike: def.TheoremLike, arity: len(def.Params), source: def.SourcePosition()}
			diags = append(diags, instanceSymbolConflict(symbol, previous)...)
		}
	}
	registerAssumption := func(assumption NamedExpr) {
		if node := assumption.definitionNode; node != nil {
			if mod.semanticNode.context.getSymbol(node.semName()) == node {
				pos := assumption.SourcePosition()
				defined[assumption.Name] = pos
				expressionGeneration.moduleSymbols[assumption.Name] = localSymbol{theoremDefNode: node, kind: semanticTheoremImportKind, arity: 0, pos: pos}
				arities[assumption.Name] = 0
				declKinds[assumption.Name] = OperatorDecl
			}
			return
		}
		if assumption.Name == "" {
			return
		}
		pos := assumption.SourcePosition()
		if _, exists := defined[assumption.Name]; !exists {
			expressionGeneration.moduleSymbols[assumption.Name] = localSymbol{kind: semanticTheoremImportKind, pos: pos}
		}
		addName(assumption.Name, pos, OperatorDecl)
		arities[assumption.Name] = 0
		declKinds[assumption.Name] = OperatorDecl
	}
	assumeProveDefs := assumeProveDefinitionNames(mod.Definitions)
	theoremLikeDefs := theoremLikeDefinitionNames(mod.Definitions)
	addNamedAssumptions(theoremLikeDefs, mod.Assumptions)
	for name, ref := range newSanyLeibnizAnalyzer(spec).resolver.scope(mod) {
		if ref.def.TheoremLike {
			theoremLikeDefs[name] = true
		}
	}
	proofStepNames := proofStepNameSet(mod.Proofs)
	defExprPositions := definitionExpressionPositions(mod.Definitions)
	assumeProveExprPositions := assumeProveDefinitionExpressionPositions(mod.Definitions)
	labelArities := moduleLabelArities(mod.Definitions)
	levelChecker := newSanyLevelCompositionChecker(mod, spec)
	levelChecker.dependencies.declKinds[mod] = declKinds
	generateInstance := func(inst *Instance) {
		diags = append(diags, checks.generator.instance(*inst)...)
		diags = append(diags, expressionGeneration.generateInstanceSubstitutions(inst, mod, expressionContexts.at(inst.Syntax, defined))...)
		diags = append(diags, expressionGeneration.generateUnnamedInstance(inst, true)...)
		diags = append(diags, expressionGeneration.generateNamedInstance(inst, false, true)...)
		checks.topLevel = append(checks.topLevel, sanyLevelCheck{position: inst.SourcePosition(), node: &sanyCachedLevelCheck{run: func() (bool, Diagnostics) {
			return levelChecker.checkInstanceSubstitutionLevelResult(*inst)
		}}})
	}
	generateProofRef := func(ref ProofRef) bool {
		diags = append(diags, checks.generator.reference(ref)...)
		generated, appended := expressionGeneration.generateProofReference(ref, mod, defined, nil)
		diags = append(diags, generated...)
		return appended
	}
	generateProof := func(proof ProofSummary) {
		diags = append(diags, checks.generator.proof(proof)...)
		diags = append(diags, expressionGeneration.proofReferences(proof, mod, defined)...)
		checks.topLevel = append(checks.topLevel, sanyLevelCheck{position: proof.Pos, node: &sanyCachedLevelCheck{run: func() (bool, Diagnostics) {
			d := checkProofSummary(proof, declKinds, mod, spec, true)
			return !d.HasErrors(), d
		}}})
	}
	generateAssumption := func(assumptionNode *NamedExpr) {
		assumptionNode.semanticNode, assumptionNode.definitionNode = nil, nil
		assumption := *assumptionNode
		diags = append(diags, checks.generator.fact(assumption)...)
		expr := assumption.Expr
		if expr == nil {
			return
		}
		if !assumeProveExprPositions[positionKey(expr.Position())] {
			if assumption.AssumeProve && assumption.AssumeProveBody != nil {
				// The native quantifier-shaped view is for legacy consumers;
				// it must not generate false FormalParam/quantifier graphs for NEW.
				if source := sanyGenerationSource(expr); source != nil {
					source.semanticGraph = nil
				}
				diags = append(diags, expressionGeneration.checkAssumeProveBody(assumption.AssumeProveBody, expressionContexts.at(assumption.Syntax, defined), nil, assumption.Name != "")...)
			} else {
				diags = append(diags, expressionGeneration.generateAssumptionExpression(assumptionNode, expressionContexts.at(assumption.Syntax, defined))...)
			}
		}
		if !assumption.AssumeProve && !defExprPositions[positionKey(expr.Position())] {
			diags = append(diags, checkLabels(expr, labelCheckContext{allowed: assumption.Name != ""})...)
		}
		diags = append(diags, checkCallArity(expr, arities, operatorParamSpecs, nil)...)
		diags = append(diags, checkOperatorArgumentKinds(expr, operatorParamSpecs, arities, nil)...)
		diags = append(diags, checkFunctionArity(expr, functionArities, nil)...)
		diags = append(diags, checkLabelReferenceArities(expr, labelArities)...)
		if assumption.AssumeProve && assumption.AssumeProveBody != nil {
			diags = append(diags, checkAssumeProveLabels(assumption.AssumeProveBody, true)...)
		}
		checks.topLevel = append(checks.topLevel, sanyLevelCheck{position: assumption.SourcePosition(), node: &sanyCachedLevelCheck{repeatSuccess: true, run: func() (bool, Diagnostics) {
			var diags Diagnostics
			correct := true
			if assumption.AssumeProve && assumption.AssumeProveBody != nil {
				expressionCorrect, current := levelChecker.checkAssumeProveResult(assumption.AssumeProveBody, nil)
				correct = correct && expressionCorrect
				diags = append(diags, current...)
			}
			if !assumption.AssumeProve && !assumeProveExprPositions[positionKey(expr.Position())] {
				expressionCorrect, current := levelChecker.checkResult(expr, nil)
				correct = correct && expressionCorrect
				diags = append(diags, current...)
			}
			diags = append(diags, checkAssumptionConstantLevel(assumption, levelChecker)...)
			return correct, diags
		}}})
	}
	var finishTheoremGeneration func()
	generateTheorem := func(theorem *NamedExpr) {
		diags = append(diags, checks.generator.fact(*theorem)...)
		expr := theorem.Expr
		if expr == nil {
			return
		}
		if finishTheoremGeneration == nil && !assumeProveExprPositions[positionKey(expr.Position())] {
			if theorem.Syntax != nil {
				var current Diagnostics
				current, finishTheoremGeneration = expressionGeneration.generateTheoremStatement(theorem, expressionContexts.at(theorem.Syntax, defined))
				diags = append(diags, current...)
			} else if theorem.AssumeProve && theorem.AssumeProveBody != nil {
				// The native quantifier-shaped view is for legacy consumers;
				// it must not generate false FormalParam/quantifier graphs for NEW.
				if source := sanyGenerationSource(expr); source != nil {
					source.semanticGraph = nil
				}
				diags = append(diags, expressionGeneration.checkAssumeProveBody(theorem.AssumeProveBody, expressionContexts.at(theorem.Syntax, defined), nil, theorem.Name != "")...)
			} else {
				diags = append(diags, checkExpr(expr, expressionContexts.at(theorem.Syntax, defined), nil)...)
			}
		}
		if !theorem.AssumeProve && !defExprPositions[positionKey(expr.Position())] {
			diags = append(diags, checkLabels(expr, labelCheckContext{})...)
		}
		diags = append(diags, checkCallArity(expr, arities, operatorParamSpecs, nil)...)
		diags = append(diags, checkOperatorArgumentKinds(expr, operatorParamSpecs, arities, nil)...)
		diags = append(diags, checkFunctionArity(expr, functionArities, nil)...)
		diags = append(diags, checkLabelReferenceArities(expr, labelArities)...)
		if theorem.AssumeProve && theorem.AssumeProveBody != nil {
			diags = append(diags, checkAssumeProveLabels(theorem.AssumeProveBody, true)...)
		}
		checks.topLevel = append(checks.topLevel, sanyLevelCheck{position: theorem.SourcePosition(), node: &sanyCachedLevelCheck{repeatSuccess: true, run: func() (bool, Diagnostics) {
			var diags Diagnostics
			correct := true
			if theorem.AssumeProve && theorem.AssumeProveBody != nil {
				expressionCorrect, current := levelChecker.checkAssumeProveResult(theorem.AssumeProveBody, nil)
				correct = correct && expressionCorrect
				diags = append(diags, current...)
			}
			if !theorem.AssumeProve && !assumeProveExprPositions[positionKey(expr.Position())] {
				expressionCorrect, current := levelChecker.checkResult(expr, nil)
				correct = correct && expressionCorrect
				diags = append(diags, current...)
			}
			return correct, diags
		}}})
	}
	generateFunctionDomains := func(def Definition) {
		diags = append(diags, checks.generator.functionDomains(&def)...)
		diags = append(diags, checkDefinitionFunctionDomains(def, expressionContexts.at(def.Syntax, defined), nil, expressionGeneration)...)
	}
	generateDefinition := func(definition *Definition, theorem *NamedExpr) {
		definition.semanticNode = nil
		def := *definition
		if def.FunctionDef {
			diags = append(diags, checks.generator.functionBody(&def)...)
		} else {
			diags = append(diags, checks.generator.definition(&def)...)
		}
		if _, recursive := recursiveArities[def.Name]; recursive {
			checks.recursive = append(checks.recursive, func() Diagnostics { return levelChecker.checkRecursiveParameters(def, nil) })
		}
		locals := map[string]bool{}
		for _, param := range def.Params {
			locals[param] = true
		}
		defArities := definitionBodyArities(arities, def)
		bodyContext := expressionContexts.at(def.Syntax, defined)
		var finishDefinitionLabels func() *sanyLabelTable
		if def.TheoremLike && theorem != nil && theorem.Syntax != nil {
			var current Diagnostics
			current, finishTheoremGeneration = expressionGeneration.generateTheoremStatement(theorem, bodyContext)
			diags = append(diags, current...)
		} else if def.AssumeProve && def.AssumeProveBody != nil {
			diags = append(diags, expressionGeneration.checkAssumeProveBody(def.AssumeProveBody, bodyContext, locals, true)...)
		} else if def.FunctionDef {
			diags = append(diags, expressionGeneration.prepareNamedFunctionDefinition(definition)...)
			def.semanticNode = definition.semanticNode
			if node := definition.semanticNode; node != nil && expressionGeneration.formalSymbolTable().resolveSymbol(def.Name) == node {
				symbol := expressionGeneration.moduleSymbols[def.Name]
				symbol.opDefNode = node
				expressionGeneration.moduleSymbols[def.Name] = symbol
			}
			previous, conflict := constructorConflicts[positionKey(def.SourcePosition())]
			if conflict && previous.kind != OperatorDecl && previous.kind != InstanceDecl {
				if _, ok := def.Expr.(*FunctionExpr); ok {
					diags = append(diags, expressionGeneration.checkRejectedNamedFunctionBody(def, bodyContext, locals)...)
				}
			} else {
				diags = append(diags, checkDefinitionFunctionBody(def, bodyContext, locals, expressionGeneration)...)
			}
		} else {
			var current Diagnostics
			current, finishDefinitionLabels = checkDefinitionExpression(def, bodyContext, locals, expressionGeneration)
			diags = append(diags, current...)
		}
		// Java processOperator constructs/registers after generating the body
		// and popping its parameter scope. Keep missing imported/recursive
		// identities incomplete instead of installing a synthetic definition.
		_, recursive := recursiveArities[def.Name]
		_, conflict := constructorConflicts[positionKey(def.SourcePosition())]
		previous := expressionGeneration.formalSymbolTable().resolveSymbol(def.Name)
		if !def.FunctionDef && !def.TheoremLike && !def.AssumeProve && (!recursive || (expressionGeneration.bindings[def.Name] != nil && expressionGeneration.bindings[def.Name].node != nil)) && (!conflict || previous != nil) && !definitionSatisfiesSymbolicConstantDeclaration(def, declKinds, arities) {
			diags = append(diags, expressionGeneration.constructOrdinaryDefinition(definition, finishDefinitionLabels)...)
			def.semanticNode = definition.semanticNode
			if node := def.semanticNode; node != nil {
				mod.semanticNode.definitions = append(mod.semanticNode.definitions, node)
				if expressionGeneration.formalSymbolTable().resolveSymbol(def.Name) == node {
					symbol := expressionGeneration.moduleSymbols[def.Name]
					symbol.opDefNode = node
					expressionGeneration.moduleSymbols[def.Name] = symbol
				}
			}
		} else if finishDefinitionLabels != nil {
			finishDefinitionLabels()
		}
		diags = append(diags, checkCallArity(def.Expr, defArities, operatorParamSpecs, locals)...)
		diags = append(diags, checkOperatorArgumentKinds(def.Expr, operatorParamSpecs, defArities, locals)...)
		diags = append(diags, checkFunctionArity(def.Expr, functionArities, locals)...)
		diags = append(diags, checkLabelReferenceArities(def.Expr, labelArities)...)
		if def.AssumeProve && def.AssumeProveBody != nil {
			diags = append(diags, checkAssumeProveLabels(def.AssumeProveBody, true)...)
		} else {
			diags = append(diags, checkLabels(def.Expr, labelCheckContext{allowed: true})...)
		}
		if !def.AssumeProve {
			diags = append(diags, checkAssumeProveDefinitionUse(def.Expr, assumeProveDefs, locals)...)
		}
		levelCheck := func() (bool, Diagnostics) {
			var diags Diagnostics
			correct := true
			if def.AssumeProve && def.AssumeProveBody != nil {
				expressionCorrect, current := levelChecker.checkAssumeProveResult(def.AssumeProveBody, locals)
				correct = correct && expressionCorrect
				diags = append(diags, current...)
			}
			if !def.AssumeProve {
				expressionCorrect, current := levelChecker.checkResult(def.Expr, locals)
				correct = correct && expressionCorrect
				diags = append(diags, current...)
			}
			return correct, diags
		}
		if _, recursive := recursiveArities[def.Name]; recursive && satisfiedRecursive[def.Name] && !completedRecursive[def.Name] && (expressionGeneration.bindings[def.Name] == nil || !expressionGeneration.bindings[def.Name].defined) {
			checks.recursiveGeneration.complete()
			if binding := expressionGeneration.bindings[def.Name]; binding != nil {
				binding.defined = true
				binding.position = def.SourcePosition()
			}
			completedRecursive[def.Name] = true
		}
		node := &sanyCachedLevelCheck{run: levelCheck}
		if checks.operatorChecks == nil {
			checks.operatorChecks = map[string]*sanyCachedLevelCheck{}
		}
		checks.operatorChecks[def.Name] = node
		if def.TheoremLike {
			checks.facts = append(checks.facts, node)
		} else {
			checks.definitions = append(checks.definitions, node)
		}
	}
	// Generator.generateModule dispatches the actual body's heirs. Complete
	// each unit before moving to the next; diagnostic order comes from this
	// traversal rather than sorting the resulting messages.
	for _, unit := range sanyModuleGenerationUnits(mod) {
		if kind := sanyModuleRecursiveSectionType(unit); kind != "" && checks.recursiveGeneration.sum > 0 {
			diagnostic := sanyDiagnosticParameters(errorAt(unit.position, "E4294", "%s may not appear within a recursive definition section.", kind), kind)
			if unit.syntax != nil {
				diagnostic.SANYRange = unit.syntax.Range
			}
			diagnostic.SANYMessage = diagnostic.Message
			diags = append(diags, diagnostic)
		}
		switch {
		case unit.declaration != nil:
			registerDeclaration(*unit.declaration)
		case unit.recursive != nil:
			registerRecursive(*unit.recursive)
		case unit.nested != nil:
			diags = append(diags, checks.generateNested(unit.nested)...)
			if unit.nested.semanticNode != nil {
				for _, diagnostic := range mod.symbolTable.addModule(unit.nested.semanticNode) {
					// Keep the native API's description alongside the actual
					// constructor diagnostic retained in SANYMessage.
					if diagnostic.Code == "E4223" {
						diagnostic.Message = fmt.Sprintf("distinct modules with name %s: this definition or declaration conflicts with the one at %v", unit.nested.Name, diagnostic.SANYParameters[1])
					}
					diags = append(diags, diagnostic)
				}
			}
		case unit.instance != nil:
			generateInstance(unit.instance)
			registerInstance(*unit.instance)
		case unit.assumption != nil:
			generateAssumption(unit.assumption)
			registerAssumption(*unit.assumption)
		case unit.definition != nil:
			definition := *unit.definition
			if definition.TheoremLike {
				generateDefinition(unit.definition, unit.theorem)
				if unit.theorem != nil && unit.theorem.definitionNode != nil {
					registerAssumption(*unit.theorem)
					if mod.semanticNode.context.getSymbol(definition.Name) == unit.theorem.definitionNode {
						addSubexpressionReferenceNames(defined, definition.Name, definition.Expr)
					}
				} else {
					registerDefinition(definition)
				}
			} else {
				if definition.FunctionDef {
					generateFunctionDomains(definition)
				}
				registerDefinition(definition)
				generateDefinition(unit.definition, nil)
				finishDefinition(*unit.definition)
			}
		}
		if unit.theorem != nil {
			generateTheorem(unit.theorem)
		}
		entries := 0
		builder := newSanyUseOrHideBuilder()
		for _, ref := range unit.references {
			appended := generateProofRef(ref)
			builder.appendReference(expressionGeneration, ref, appended)
			if appended {
				entries++
			}
		}
		if unit.syntax != nil && unit.syntax.Kind.JavaName() == "N_UseOrHide" {
			if node := builder.finish(unit.syntax); node != nil {
				diags = append(diags, node.factCheck()...)
				mod.semanticNode.addTopLevel(node)
			} else {
				for _, ref := range unit.references {
					diags = append(diags, checkHideRef(ref, theoremLikeDefs, proofStepNames)...)
				}
			}
			diags = append(diags, sanyEmptyProofCommand(unit.syntax, entries, "Empty USE or HIDE statement.")...)
		}
		for _, proof := range unit.proofs {
			generateProof(proof)
		}
		if finishTheoremGeneration != nil {
			finishTheoremGeneration()
			finishTheoremGeneration = nil
		}
		if unit.theorem != nil && unit.theorem.AssumeProveBody != nil && unit.theorem.AssumeProveBody.semanticNode != nil {
			unit.theorem.AssumeProveBody.semanticNode.inProof = false
		}
		if unit.definition != nil && unit.definition.TheoremLike && unit.definition.AssumeProveBody != nil && unit.definition.AssumeProveBody.semanticNode != nil {
			unit.definition.AssumeProveBody.semanticNode.inProof = false
		}
	}
	// checkForUndefinedRecursiveOps visits the declaration vector, not a map.
	if checks.recursiveGeneration.count > 0 {
		for _, node := range mod.semanticNode.recursiveDecls {
			if node.letInLevel == 0 && !node.defined {
				diags = append(diags, sanyUndefinedRecursiveDiagnostic(&sanyRecursiveBinding{name: node.semName(), position: node.semPosition()}))
			}
		}
		for _, name := range recursiveOrder {
			if binding := expressionGeneration.bindings[name]; binding != nil && binding.node != nil {
				continue
			}
			binding := expressionGeneration.bindings[name]
			if !satisfiedRecursive[name] && (binding == nil || !binding.defined) {
				pos := recursivePositions[name]
				diagnostic := sanyDiagnosticParameters(errorAt(pos, "E4291", "recursive declaration %s has no definition", name), name)
				diagnostic.SANYRange = SanyRange{Begin: pos, End: pos.SourceEnd()}
				diagnostic.SANYMessage = fmt.Sprintf("Symbol %s declared in RECURSIVE statement but not defined.", name)
				diags = append(diags, diagnostic)
			}
		}
		checks.recursiveGeneration.sum -= checks.recursiveGeneration.count
	}
	return diags
}

func checkAssumeProveBindings(body *AssumeProve, defined map[string]Position, locals map[string]bool, generators ...*sanyExpressionGeneration) Diagnostics {
	if body == nil {
		return nil
	}
	var diags Diagnostics
	if len(generators) == 0 || generators[0] == nil {
		apLocals := copyBoolMap(locals)
		for _, item := range body.Assumptions {
			switch {
			case item.NewSymbol != nil:
				symbol := item.NewSymbol
				diags = append(diags, checkBindingName("NEW symbol", symbol.Name, symbol.Pos, defined, apLocals)...)
				if symbol.Domain != nil {
					diags = append(diags, checkExpr(symbol.Domain, defined, apLocals)...)
				}
				apLocals[symbol.Name] = true
			case item.Nested != nil:
				diags = append(diags, checkAssumeProveBindings(item.Nested, defined, apLocals)...)
			case item.Expr != nil:
				diags = append(diags, checkExpr(item.Expr, defined, apLocals)...)
			}
		}
		if body.Prove != nil {
			diags = append(diags, checkExpr(body.Prove, defined, apLocals)...)
		}
		return diags
	}
	var generation *sanyExpressionGeneration
	body.semanticNode = nil
	var goal sanySemanticGraphNode
	if len(generators) != 0 && generators[0] != nil && generators[0].labelAPDepth == 0 && generators[0].currentGoal != nil {
		goal = generators[0].currentGoal
	}
	generation = generators[0]
	generation.labelAPDepth++
	node := newSanySemAssumeProveNode(body.Syntax, goal)
	node.assumes = make([]sanySemanticGraphNode, len(body.Assumptions))
	node.inScopeOfDecl = make([]bool, len(body.Assumptions)+1)
	if body.Syntax != nil {
		heirs := body.Syntax.GetHeirs()
		if len(heirs)%2 != 0 {
			panic(tlc.NewWrongInvocationException("AssumeProve has odd number of children"))
		}
		if len(heirs) > 0 {
			node.isBoxAssumeProve = heirs[0].Image == "[]ASSUME"
			if len(heirs) > 1 {
				prove := heirs[len(heirs)-2].Image
				message := ""
				if node.isBoxAssumeProve && prove != "[]PROVE" {
					message = "[]ASSUME matched by PROVE instead of []PROVE"
				}
				if !node.isBoxAssumeProve && prove != "PROVE" {
					message = "ASSUME matched by []PROVE instead of PROVE"
				}
				if message != "" {
					diagnostic := errorAt(sanyNodePosition(heirs[0]), "E4005", "%s", message)
					diagnostic.SANYMessage = message
					diagnostic.SANYRange = heirs[0].Range
					diags = append(diags, diagnostic)
				}
			}
		}
	}
	complete := true
	var closeContext func()
	var previousSymbols map[string]localSymbol
	var previousLabelsEnabled, previousGoalUnsupported bool
	if len(generators) != 0 && generators[0] != nil {
		previousLabelsEnabled = generation.labelsEnabled
		previousGoalUnsupported = generation.labelGoalUnsupported
		if generation.labelAPDepth == 1 && generation.apGoalUnavailable {
			generation.labelGoalUnsupported = true
		}
		if generation.labelAPDepth == 1 {
			generation.currentGoalClause = 0
		}
		if generation.labelAPDepth < 0 || generation.labelAPDepth >= int32(len(generation.inScopeOfAPDecl)) {
			panic(tlc.NewArrayIndexOutOfBoundsException(int(generation.labelAPDepth), len(generation.inScopeOfAPDecl)))
		}
		generation.inScopeOfAPDecl[generation.labelAPDepth] = false
		generation.labelsEnabled = true
		previousSymbols = generation.symbols
		generation.symbols = make(map[string]localSymbol, len(previousSymbols))
		for name, symbol := range previousSymbols {
			generation.symbols[name] = symbol
		}
		ownedContext := generation.labelAPDepth == 1 && generation.outerAPContextOwned
		if !ownedContext {
			closeContext = generation.pushFormalContext(0)
		}
		if generation.nodes != nil {
			table := generation.formalSymbolTable()
			if node.isBoxAssumeProve {
				if table.resolveSymbol("$$InAssume") != nil && body.Syntax != nil {
					token := body.Syntax.GetHeirs()[0]
					message := "[]ASSUME used within the scope of an ordinary ASSUME's assumptions"
					diagnostic := errorAt(sanyNodePosition(token), "E4005", "%s", message)
					diagnostic.SANYMessage = message
					diagnostic.SANYRange = token.Range
					diags = append(diags, diagnostic)
				}
			} else if table.resolveSymbol("$$InAssume") == nil {
				diags = append(diags, table.addSymbol(generation.nodes.inAssumeDummy)...)
			}
		}
	}
	generate := func(expr Expr, locals map[string]bool) Diagnostics {
		if len(generators) != 0 && generators[0] != nil {
			return generators[0].proofExpression(expr, defined, locals)
		}
		return checkExpr(expr, defined, locals)
	}
	apLocals := copyBoolMap(locals)
	for i, item := range body.Assumptions {
		node.inScopeOfDecl[i+1] = node.inScopeOfDecl[i]
		switch {
		case item.NewSymbol != nil:
			sym := item.NewSymbol
			if generation != nil {
				diags = append(diags, generation.generateNewSymbol(sym, defined, apLocals)...)
			} else {
				diags = append(diags, checkBindingName("NEW symbol", sym.Name, sym.Pos, defined, apLocals)...)
				if sym.Domain != nil {
					diags = append(diags, generate(sym.Domain, apLocals)...)
				}
			}
			if sym.semanticNode != nil {
				node.assumes[i] = sym.semanticNode
			}
			node.inScopeOfDecl[i+1] = true
			apLocals[sym.Name] = true
			if generation != nil {
				generation.inScopeOfAPDecl[generation.labelAPDepth] = true
			}
		case item.Nested != nil:
			diags = append(diags, checkAssumeProveBindings(item.Nested, defined, apLocals, generators...)...)
			if item.Nested.semanticNode != nil {
				node.assumes[i] = item.Nested.semanticNode
			}
		case item.Expr != nil:
			previousAllowed := false
			if generation != nil {
				previousAllowed = generation.allowLabeledAP
				generation.allowLabeledAP = true
			}
			diags = append(diags, generate(item.Expr, apLocals)...)
			if generation != nil {
				generation.allowLabeledAP = previousAllowed
			}
			node.assumes[i] = sanyGeneratedExpressionNode(item.Expr)
		}
		complete = complete && (node.assumes[i] != nil || (item.Expr != nil && sanyExpressionGenerationFailure(item.Expr) == sanyGenerationNullExpression))
		if generation != nil && generation.labelAPDepth == 1 {
			generation.currentGoalClause++
		}
	}
	if body.Prove != nil {
		diags = append(diags, generate(body.Prove, apLocals)...)
		node.prove = sanyGeneratedExpressionNode(body.Prove)
		complete = complete && (node.prove != nil || sanyExpressionGenerationFailure(body.Prove) == sanyGenerationNullExpression)
	}
	if generation != nil && complete && !(generation.labelAPDepth == 1 && generation.apGoalUnavailable) {
		body.semanticNode = node
	}
	if closeContext != nil {
		closeContext()
		generation.symbols = previousSymbols
	}
	generation.labelAPDepth--
	generation.labelGoalUnsupported = previousGoalUnsupported
	generation.labelsEnabled = previousLabelsEnabled
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

func positionInsideModule(pos Position, mod *Module) bool {
	if mod == nil || pos.Line == 0 || mod.Pos.Line == 0 || mod.Pos.EndLine == 0 {
		return false
	}
	if pos.File != "" && mod.Pos.File != "" && pos.File != mod.Pos.File {
		return false
	}
	if pos.Line < mod.Pos.Line || pos.Line > mod.Pos.EndLine {
		return false
	}
	if pos.Line == mod.Pos.Line && mod.Pos.Column > 0 && pos.Column < mod.Pos.Column {
		return false
	}
	if pos.Line == mod.Pos.EndLine && mod.Pos.EndColumn > 0 && pos.Column > mod.Pos.EndColumn {
		return false
	}
	return true
}

func enclosingSymbolVisibleBeforeNested(symbol semanticExportedSymbol, enclosing, nested *Module, spec *Spec) bool {
	if sameSourceFile(symbol.pos, nested.Pos) && positionBefore(symbol.pos, nested.Pos) {
		return true
	}
	if base, _, ok := strings.Cut(symbol.name, "!"); ok && enclosingInstanceBeforeNested(enclosing, base, nested) {
		return true
	}
	for _, inst := range enclosing.Instances {
		if !positionBefore(inst.SourcePosition(), nested.Pos) {
			continue
		}
		for _, imported := range semanticInstanceSymbols(inst, spec) {
			if imported.name == symbol.name {
				return true
			}
		}
	}
	if enclosingExtendedExportNames(enclosing, spec)[symbol.name] {
		return true
	}
	return false
}

func enclosingInstanceBeforeNested(enclosing *Module, name string, nested *Module) bool {
	if enclosing == nil || nested == nil || name == "" {
		return false
	}
	for _, inst := range enclosing.Instances {
		if inst.Name != name {
			continue
		}
		pos := inst.SourcePosition()
		if sameSourceFile(pos, nested.Pos) && positionBefore(pos, nested.Pos) {
			return true
		}
	}
	return false
}

func enclosingExtendedExportNames(enclosing *Module, spec *Spec) map[string]bool {
	names := map[string]bool{}
	if enclosing == nil || spec == nil {
		return names
	}
	for _, ext := range enclosing.Extends {
		for _, symbol := range semanticModuleExports(spec.Modules[ext], spec, map[string]bool{}) {
			names[symbol.name] = true
		}
	}
	return names
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

func letScopeLocalsBeforeDefinition(locals map[string]bool, expr *LetExpr, defIndex int) map[string]bool {
	letLocals := copyBoolMap(locals)
	if expr == nil {
		return letLocals
	}
	for _, decl := range expr.Recursives {
		for _, name := range decl.Names {
			letLocals[name] = true
		}
	}
	if defIndex > len(expr.Definitions) {
		defIndex = len(expr.Definitions)
	}
	for i := 0; i < defIndex; i++ {
		letLocals[expr.Definitions[i].Name] = true
	}
	for _, inst := range expr.Instances {
		if inst.Name != "" {
			letLocals[inst.Name+"!"] = true
		}
	}
	return letLocals
}

func letRecursiveNames(expr *LetExpr) map[string]bool {
	names := map[string]bool{}
	if expr == nil {
		return names
	}
	for _, decl := range expr.Recursives {
		for _, name := range decl.Names {
			names[name] = true
		}
	}
	return names
}

func letDefinitionBodyLocals(letLocals map[string]bool, def Definition, recursive bool) map[string]bool {
	defLocals := copyBoolMap(letLocals)
	if _, isFunctionDefinition := definitionFunctionArity(def); !isFunctionDefinition && !recursive {
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
	allowed          bool
	noLabels         bool
	inExcept         bool
	bound            []string
	formalGroups     [][]*sanyFormalParamNode
	unresolvedBounds bool
}

// Each group corresponds to one pushFormalParams on the current LS frame.
// The value context restores the previous sequence when a traversal returns.
func (ctx labelCheckContext) withBounds(bounds []BoundVar, nodes []*sanyFormalParamNode) labelCheckContext {
	next := ctx
	next.bound = append([]string(nil), ctx.bound...)
	for _, bound := range bounds {
		if bound.Name != "" {
			next.bound = append(next.bound, bound.Name)
		}
	}
	next.formalGroups = append(append([][]*sanyFormalParamNode(nil), ctx.formalGroups...), nodes)
	next.unresolvedBounds = ctx.unresolvedBounds || len(nodes) != len(bounds)
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
	next.formalGroups = nil
	next.unresolvedBounds = false
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
		definitionCtx := ctx.resetLabelBoundScope()
		definitionCtx.allowed = true
		for _, def := range e.Definitions {
			if def.AssumeProve && def.AssumeProveBody != nil {
				diags = append(diags, checkAssumeProveLabelsWithContext(def.AssumeProveBody, true, definitionCtx)...)
			} else {
				diags = append(diags, checkLabels(def.Expr, definitionCtx)...)
			}
		}
		diags = append(diags, checkLabels(e.Body, ctx)...)
	case *QuantifierExpr:
		parameters, body := sanyQuantifierGroup(e)
		var bounds []BoundVar
		seenDomains := map[Expr]bool{}
		for _, parameter := range parameters {
			if parameter.Set != nil && !seenDomains[parameter.Set] {
				diags = append(diags, checkLabels(parameter.Set, ctx)...)
				seenDomains[parameter.Set] = true
			}
			bounds = append(bounds, BoundVar{Name: parameter.Var})
		}
		diags = append(diags, checkLabels(body, ctx.withBounds(bounds, e.quantifierFormals))...)
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
		diags = append(diags, checkLabels(e.Body, ctx.withBounds(e.boundVars(), e.formalNodes))...)
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
		diags = append(diags, checkLabels(e.Body, ctx.withBounds(e.Bounds, e.formalNodes))...)
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
		if e.labelGenerated {
			return nil
		}
		// Generator.generateLabel returns nullLabelNode at the first failed
		// guard, without generating the body or checking its parameters.
		if !ctx.allowed {
			diagnostic := errorAt(e.Pos, "E4333", "label %s is not in definition or proof step", e.Name)
			diagnostic.SANYMessage = "Label not in definition or proof step."
			return Diagnostics{diagnostic}
		}
		if ctx.noLabels {
			diagnostic := errorAt(e.Pos, "E4334", "label %s is not allowed in a nested ASSUME/PROVE block with NEW", e.Name)
			diagnostic.SANYMessage = "Label not allowed within scope of declaration in nested ASSUME/PROVE."
			if e.Syntax != nil {
				diagnostic.SANYRange = e.Syntax.Range
			}
			return Diagnostics{diagnostic}
		}
		if ctx.inExcept {
			diagnostic := errorAt(e.Pos, "E4335", "label %s is not allowed inside EXCEPT", e.Name)
			diagnostic.SANYMessage = "Labels inside EXCEPT clauses are not yet implemented."
			return Diagnostics{diagnostic}
		}
		// The source pushes a new label scope and generates the body before
		// resolving parameters and running formalParamsEqual on this label.
		diags = append(diags, checkLabels(e.Body, ctx.resetLabelBoundScope())...)
		diags = append(diags, checkLabelParameters(e, ctx)...)

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
		body := e.Element
		if e.Predicate != nil {
			body = e.Predicate
		}
		diags = append(diags, checkLabels(body, ctx.withBounds(e.Bounds, e.formalNodes))...)
	}
	return diags
}

func checkLabelParameters(label *LabelExpr, ctx labelCheckContext) Diagnostics {
	var diags Diagnostics
	// generateLabel resolves every argument before formalParamsEqual. Each
	// non-formal occurrence reports at its own argument syntax, even when
	// another occurrence has the same spelling.
	for i, syntax := range label.illegalParameterSyntax {
		if syntax == nil {
			continue
		}
		param := label.Params[i]
		diagnostic := sanyDiagnosticParameters(errorAt(sanyNodePosition(syntax), "E4332", "Illegal parameter %s of label `%s'.", param, label.Name), param, label.Name)
		diagnostic.SANYMessage = diagnostic.Message
		diagnostic.SANYRange = syntax.Range
		diags = append(diags, diagnostic)
	}
	seen := map[string]bool{}
	seenFormals := tlc.NewJavaSemanticUIDSet(int32(sanyFormalParamKind))
	formalByUID := map[int32]*sanyFormalParamNode{}
	for i, param := range label.Params {
		repeated := seen[param]
		if label.formalNodes != nil {
			// All entries have the same concrete FormalParamNode class/kind;
			// SemanticNode.equals therefore compares their retained UIDs.
			uid := label.formalNodes[i].getUID()
			repeated = !seenFormals.Add(uid)
			if !repeated {
				formalByUID[uid] = label.formalNodes[i]
			}
		}
		if repeated {
			diagnostic := sanyDiagnosticParameters(errorAt(label.Pos, "E4330", "repeated label parameter %s in label %s", param, label.Name), param, label.Name)
			diagnostic.SANYMessage = fmt.Sprintf("Repeated formal parameter %s \nin label `%s'.", param, label.Name)
			diags = append(diags, diagnostic)
			continue
		}
		seen[param] = true
	}
	required := map[string]bool{}
	missing := func(name string) {
		diagnostic := sanyDiagnosticParameters(errorAt(label.Pos, "E4331", "label %s must contain bound parameter %s", label.Name, name), label.Name, name)
		diagnostic.SANYMessage = fmt.Sprintf("Label %s must contain formal parameter `%s'.", label.Name, name)
		diags = append(diags, diagnostic)
	}
	if label.formalNodes != nil && !ctx.unresolvedBounds {
		// formalParamsEqual removes each required identity in LS sequence
		// order. A rejected same-named declaration remains a distinct node.
		for _, group := range ctx.formalGroups {
			for _, node := range group {
				name := node.semName()
				required[name] = true
				uid := node.getUID()
				if !seenFormals.Remove(uid) {
					missing(name)
				}
			}
		}
	} else {
		for _, name := range ctx.bound {
			required[name] = true
			if !seen[name] {
				missing(name)
			}
		}
	}
	if label.formalNodes != nil && !ctx.unresolvedBounds {
		if seenFormals.Len() != 0 {
			message := "Label " + label.Name + " declares extra parameter(s)  "
			for uid := range seenFormals.All() {
				message += formalByUID[uid].semName() + "  "
			}
			diagnostic := errorAt(label.Pos, "E4332", "unnecessary label parameter(s) in label %s", label.Name)
			diagnostic.SANYMessage = message
			diags = append(diags, diagnostic)
		}
	} else {
		for _, param := range label.Params {
			if !required[param] {
				diags = append(diags, errorAt(label.Pos, "E4332", "unnecessary label parameter %s in label %s", param, label.Name))
			}
		}
	}
	return diags
}

func checkDuplicateSiblingLabels(exprs []Expr) Diagnostics {
	var diags Diagnostics
	seen := map[string]Position{}
	for _, expr := range exprs {
		label, ok := expr.(*LabelExpr)
		if !ok || label.Name == "" || label.labelGenerated {
			continue
		}
		if _, exists := seen[label.Name]; exists {
			diags = append(diags, sanyDiagnosticParameters(errorAt(label.Pos, "E4336", "Duplicate label %s", label.Name), label.Name))
			continue
		}
		seen[label.Name] = label.Pos
	}
	return diags
}

func moduleLabelArities(defs []Definition) map[string]int {
	out := map[string]int{}
	for _, def := range defs {
		if def.Name == "" {
			continue
		}
		collectLabelArities(def.Name, def.Expr, out)
	}
	return out
}

func collectLabelArities(base string, expr Expr, out map[string]int) {
	if expr == nil {
		return
	}
	if label, ok := expr.(*LabelExpr); ok && label.Name != "" {
		out[base+"!"+label.Name] = len(label.Params)
	}
	if let, ok := expr.(*LetExpr); ok {
		for _, def := range let.Definitions {
			collectLabelArities(base, def.Expr, out)
		}
	}
	for _, child := range sanySubexpressionChildren(expr) {
		collectLabelArities(base, child, out)
	}
}

func checkLabelReferenceArities(expr Expr, labelArities map[string]int) Diagnostics {
	if sanyExpressionGenerationFailure(expr) != sanyGenerationSucceeded {
		return nil
	}
	if expr == nil || len(labelArities) == 0 {
		return nil
	}
	var diags Diagnostics
	if selected := sanyExprSelection(expr); selected != nil {
		// selectorToNode has already checked each argument group against its
		// own operator/label. The flattened call includes outer parameters.
		for _, arg := range selected.args {
			diags = append(diags, checkLabelReferenceArities(arg, labelArities)...)
		}
		return diags
	}
	if call, ok := expr.(*CallExpr); ok {
		if ident, ok := call.Callee.(*IdentExpr); ok {
			if want, exists := labelArities[ident.Name]; exists && len(call.Args) != want {
				diags = append(diags, sanyDiagnosticParameters(errorAt(call.Pos, "E4337", "label %s arity mismatch: got %d args, want %d", ident.Name, len(call.Args), want), ident.Name))
			}
		}
	} else if ident, ok := expr.(*IdentExpr); ok {
		if want, exists := labelArities[ident.Name]; exists && want != 0 {
			diags = append(diags, sanyDiagnosticParameters(errorAt(ident.Pos, "E4337", "label %s arity mismatch: got 0 args, want %d", ident.Name, want), ident.Name))
		}
	}
	if let, ok := expr.(*LetExpr); ok {
		for _, def := range let.Definitions {
			diags = append(diags, checkLabelReferenceArities(def.Expr, labelArities)...)
		}
	}
	for _, child := range sanySubexpressionChildren(expr) {
		diags = append(diags, checkLabelReferenceArities(child, labelArities)...)
	}
	return diags
}

func checkAssumeProveLabels(body *AssumeProve, topLevel bool) Diagnostics {
	return checkAssumeProveLabelsInScope(body, topLevel, false)
}

// Generator tracks NEW declaration scope independently of its formal-parameter
// stack. A NEW declaration is an OpDeclNode, not a quantified FormalParamNode;
// it does not become a required label parameter. In nested ASSUME/PROVE blocks,
// labels are forbidden only after a NEW declaration enters scope.
func checkAssumeProveLabelsInScope(body *AssumeProve, topLevel, declarationScope bool) Diagnostics {
	return checkAssumeProveLabelsWithContext(body, topLevel, labelCheckContext{allowed: true, noLabels: declarationScope})
}

func checkAssumeProveLabelsWithContext(body *AssumeProve, topLevel bool, ctx labelCheckContext) Diagnostics {
	if body == nil {
		return nil
	}
	var diags Diagnostics
	check := func(expr Expr) {
		diags = append(diags, checkLabels(expr, ctx)...)
	}
	for _, item := range body.Assumptions {
		switch {
		case item.NewSymbol != nil:
			check(item.NewSymbol.Domain)
			if !topLevel {
				ctx.noLabels = true
			}
		case item.Nested != nil:
			diags = append(diags, checkAssumeProveLabelsWithContext(item.Nested, false, ctx)...)
		case item.Expr != nil:
			check(item.Expr)
		}
	}
	check(body.Prove)
	return diags
}

func isEmbeddedStandardModule(mod *Module) bool {
	if mod == nil {
		return false
	}
	source, ok := standardModules[mod.Name]
	return ok && mod.SourcePath == mod.Name+".tla" && strings.TrimSpace(mod.Source) == strings.TrimSpace(source)
}

type importedSymbol struct {
	declarationNode *sanySemOpDeclNode
	arity           int
	kind            DeclarationKind
	pos             Position
	source          string
}

// Context compares Java semantic node classes, not TLC declaration levels.
const (
	semanticTheoremImportKind     DeclarationKind = "THM_OR_ASSUMP"
	semanticFormalParamImportKind DeclarationKind = "FORMAL_PARAM"
)

func semanticDefinitionImportKind(definition Definition) DeclarationKind {
	if definition.TheoremLike {
		return semanticTheoremImportKind
	}
	return OperatorDecl
}

func semanticImportDescription(kind DeclarationKind) string {
	if kind == OperatorDecl || kind == semanticFormalParamImportKind {
		return "definition"
	}
	return "declaration"
}

// EXTENDS and the enclosing module's symbol stack reuse the original
// declaration object. They do not construct a new declaration at the reference.
func sanyModuleDeclarationNode(module *Module, name string) *sanySemOpDeclNode {
	if module == nil || module.semanticNode == nil {
		return nil
	}
	node, _ := module.semanticNode.context.getSymbol(sanyCanonicalOperatorImage(name)).(*sanySemOpDeclNode)
	return node
}

func declarationSymbolPosition(declaration Declaration, name string) Position {
	if position, ok := declaration.NamePositions[name]; ok {
		return position
	}
	return declaration.Pos
}

func sanySymbolLocation(position Position) string {
	if position == (Position{}) {
		return "Unknown location"
	}
	end := position.SourceEnd()
	if position.Line == 0 && position.Column == 0 && end.Line == 0 && end.Column == 0 && position.File != "" {
		return "In module " + moduleNameForSourcePosition(position)
	}
	return fmt.Sprintf("line %d, col %d to line %d, col %d of module %s", position.Line, position.Column, end.Line, end.Column, moduleNameForSourcePosition(position))
}

type sanyDiagnosticLocation struct {
	Position Position
}

func (location sanyDiagnosticLocation) String() string {
	return sanySymbolLocation(location.Position)
}

// Context.mergeExtendContext's ErrorDetails retain structured parameters as
// well as their rendered message. Both the parser and direct context use this.
func sanyExtendConflictForClasses(name string, incomingKind DeclarationKind, incoming Position, existingKind DeclarationKind, existing Position, sameClass bool) Diagnostic {
	incomingDescription := semanticImportDescription(incomingKind)
	existingDescription := semanticImportDescription(existingKind)
	parameters := []any{incomingDescription, name, existingDescription, sanyDiagnosticLocation{Position: existing}}
	diagnostic := errorAt(incoming, "E4224", "The %s of '%s' conflicts with \nits %s at %s.", parameters...)
	if sameClass {
		parameters[0] = existingDescription
		diagnostic = warningAt(incoming, "W4800", "Warning: the %s of '%s' conflicts with \nits %s at %s.", parameters...)
	}
	diagnostic.SANYRange = SanyRange{Begin: incoming, End: incoming.SourceEnd()}
	diagnostic.SANYMessage = diagnostic.Message
	diagnostic.SANYParameters = parameters
	return diagnostic
}

func sanyExtendeePosition(module *Module, index int) Position {
	if module.Syntax != nil {
		heirs := module.Syntax.GetHeirs()
		if len(heirs) > 1 && heirs[1] != nil {
			for _, syntax := range heirs[1].GetHeirs() {
				if syntax != nil && syntax.Kind.JavaName() == "IDENTIFIER" {
					if index == 0 {
						return sanyNodePosition(syntax)
					}
					index--
				}
			}
		}
	}
	return module.Pos // Native implicit EXTENDS has no source token.
}

// Retain native expression metadata without duplicating Context's diagnostics.
func recordImportedSymbol(name string, kind DeclarationKind, pos Position, source string, seen map[string]importedSymbol) {
	if name == "" || source == "" || seen == nil {
		return
	}
	if _, exists := seen[name]; !exists {
		seen[name] = importedSymbol{kind: kind, pos: pos, source: source}
	}
}

type localSymbol struct {
	instanceOrigin   *Module
	instanceSyntax   *SanySyntaxNode
	builtinNode      *sanySemOpDefNode
	opDefNode        *sanySemOpDefNode
	theoremDefNode   *sanySemThmOrAssumpDefNode
	formalNode       *sanyFormalParamNode
	declarationNode  *sanySemOpDeclNode
	proofStepKind    string
	proofAssumeProve bool
	operatorParams   []operatorParamSpec
	arity            int
	kind             DeclarationKind
	pos              Position
}

func moduleOwnSymbols(mod *Module) map[string]localSymbol {
	symbols := map[string]localSymbol{}
	if mod == nil {
		return symbols
	}
	for _, decl := range mod.Declarations {
		for _, name := range decl.Names {
			if name != "" {
				symbols[name] = localSymbol{kind: decl.Kind, pos: declarationSymbolPosition(decl, name), arity: decl.Arities[name]}
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
			symbols[def.Name] = localSymbol{kind: semanticDefinitionImportKind(def), pos: def.SourcePosition(), arity: len(def.Params)}
		}
	}
	for _, assumption := range mod.Assumptions {
		if assumption.Name != "" {
			symbols[assumption.Name] = localSymbol{kind: semanticTheoremImportKind, pos: assumption.SourcePosition()}
		}
	}
	return symbols
}

func instanceSymbolConflict(symbol semanticExportedSymbol, previous localSymbol) Diagnostics {
	position := symbol.sourcePosition()
	if previous.pos == position {
		return nil
	}
	diagnostic := warningAt(position, "W4801", "the INSTANCE export %s conflicts with an existing symbol at %s; the first binding is used", symbol.name, previous.pos)
	diagnostic.SANYMessage = fmt.Sprintf("Multiple declarations or definitions for symbol %s.  \nThis duplicates the one at %s.", symbol.name, sanySymbolLocation(previous.pos))
	diagnostic.SANYParameters = []any{symbol.name, sanySymbolLocation(previous.pos)}
	if previous.kind != symbol.importKind() || previous.arity != symbol.arity {
		diagnostic = errorAt(position, "E4201", "INSTANCE export %s has kind/arity %s/%d, conflicting with %s/%d at %s", symbol.name, symbol.importKind(), symbol.arity, previous.kind, previous.arity, previous.pos)
		diagnostic.SANYMessage = fmt.Sprintf("Multiply-defined symbol '%s': this definition or declaration conflicts \nwith the one at %s.", symbol.name, sanySymbolLocation(previous.pos))
	}
	diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
	return Diagnostics{diagnostic}
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
	originSyntax      *SanySyntaxNode
	theoremLike       bool
	source            Position
	origin            *Module
	name              string
	kind              DeclarationKind
	pos               Position
	arity             int
	hasArity          bool
	operatorParams    []operatorParamSpec
	hasOperatorParams bool
	unqualified       bool
}

func (symbol semanticExportedSymbol) importKind() DeclarationKind {
	if symbol.theoremLike {
		return semanticTheoremImportKind
	}
	return symbol.kind
}

func (symbol semanticExportedSymbol) sourcePosition() Position {
	if symbol.source.Line > 0 {
		return symbol.source
	}
	return symbol.pos
}

func semanticModuleParameterFree(mod *Module, spec *Spec, visiting map[*Module]bool) bool {
	if mod == nil || visiting[mod] {
		return true
	}
	visiting[mod] = true
	defer delete(visiting, mod)
	if len(mod.Declarations) > 0 {
		return false
	}
	for _, ext := range mod.Extends {
		if !semanticModuleParameterFree(spec.Modules[ext], spec, visiting) {
			return false
		}
	}
	return true
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
	qualifier := inst.Name
	exportUnqualified := inst.exportsUnqualified()
	out := make([]semanticExportedSymbol, 0, len(exports))
	for _, symbol := range exports {
		// Generator instantiates OpDefNode and ThmOrAssumpDefNode entries,
		// never the instancee's substituted declaration nodes.
		if symbol.kind != OperatorDecl {
			continue
		}
		if !semanticModuleParameterFree(instMod, spec, map[*Module]bool{}) && !semanticModuleParameterFree(symbol.origin, spec, map[*Module]bool{}) {
			symbol.source = inst.SourcePosition()
		}
		if exportUnqualified {
			unqualified := symbol
			applyInstanceParamArity(&unqualified, inst)
			unqualified.unqualified = true
			out = append(out, unqualified)
		}
		if qualifier == "" {
			continue
		}
		qualified := symbol
		qualified.name = qualifier + "!" + symbol.name
		applyInstanceParamArity(&qualified, inst)
		qualified.unqualified = false
		out = append(out, qualified)
	}
	return out
}

func applyInstanceParamArity(symbol *semanticExportedSymbol, inst Instance) {
	if symbol == nil || len(inst.Params) == 0 {
		return
	}
	symbol.arity += len(inst.Params)
	symbol.hasArity = true
	// An instantiated operator receives the module-definition parameters
	// before its own parameters. Preserve those positions in the argument
	// specifications as well as in the total arity, as Generator does when
	// constructing the instantiated OpDefNode's formal parameter list.
	prefix, _ := definitionOperatorParamSpecs(Definition{
		Params: inst.Params, ParamArities: inst.ParamArities,
	})
	symbol.operatorParams = append(prefix, symbol.operatorParams...)
	symbol.hasOperatorParams = true
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
			// Context.mergeExtendContext retains the existing binding.
			if _, exists := byName[symbol.name]; !exists {
				byName[symbol.name] = symbol
			}
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
			symbol := semanticExportedSymbol{name: name, kind: decl.Kind, pos: pos, source: pos, origin: mod}
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
			originSyntax: def.Syntax,
			theoremLike:  def.TheoremLike,
			name:         def.Name,
			source:       def.SourcePosition(),
			origin:       mod,
			kind:         OperatorDecl,
			pos:          def.Pos,
			arity:        len(def.Params),
			hasArity:     true,
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
			originSyntax: assumption.Syntax,
			theoremLike:  true,
			name:         assumption.Name,
			source:       assumption.SourcePosition(),
			origin:       mod,
			kind:         OperatorDecl,
			pos:          assumption.SourcePosition(),
			arity:        0,
			hasArity:     true,
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
	instMod := spec.Modules[inst.Module]
	for _, symbol := range semanticInstanceSymbols(inst, spec) {
		if exportUnqualified && symbol.unqualified {
			if local, exists := localSymbols[symbol.name]; exists && local.pos.Compare(inst.SourcePosition()) < 0 {
				diags = append(diags, instanceSymbolConflict(symbol, local)...)
				continue
			}
			if previous, exists := instanceSymbols[symbol.name]; exists {
				diags = append(diags, instanceSymbolConflict(symbol, localSymbol{kind: previous.kind, pos: previous.pos, arity: previous.arity})...)
				continue
			}
			instanceSymbols[symbol.name] = importedSymbol{kind: symbol.importKind(), pos: symbol.sourcePosition(), arity: symbol.arity, source: qualifier}
			addSemanticSymbol(symbol, defined, declKinds, arities, operatorParamSpecs)
			continue
		}
		addSemanticSymbol(symbol, defined, declKinds, arities, operatorParamSpecs)
	}
	if instMod != nil {
		for i := range instMod.Definitions {
			def := instMod.Definitions[i]
			if def.Local {
				continue
			}
			if exportUnqualified {
				addSubexpressionReferenceNames(defined, def.Name, def.Expr)
			}
			if inst.Name != "" {
				addSubexpressionReferenceNames(defined, qualifier+"!"+def.Name, def.Expr)
			}
		}
	}
	return diags
}

func checkDefinitionParams(def Definition) Diagnostics {
	var diags Diagnostics
	seen := map[string]Position{}
	for _, param := range sanyDefinitionParams(&def) {
		if previous, exists := seen[param.Name]; exists {
			diagnostic := errorAt(param.Pos, "E4201", "duplicate parameter %s in definition %s", param.Name, def.Name)
			diagnostic.SANYRange = SanyRange{Begin: param.Pos, End: param.Pos.SourceEnd()}
			diagnostic.SANYMessage = fmt.Sprintf("Multiply-defined symbol '%s': this definition or declaration conflicts \nwith the one at %s.", param.Name, sanySymbolLocation(previous))
			diags = append(diags, diagnostic)
		} else {
			seen[param.Name] = param.Pos
		}
	}
	return diags
}

func checkDefinitionParamCollisions(def Definition, defined map[string]Position, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	for _, param := range sanyDefinitionParams(&def) {
		if param.OperatorArity > 0 && !isIdentifierName(param.Name) {
			continue
		}
		generated := checkBindingName("parameter", param.Name, param.Pos, defined, locals)
		if previous, exists := defined[param.Name]; exists {
			for i := range generated {
				generated[i].SANYMessage = fmt.Sprintf("Multiply-defined symbol '%s': this definition or declaration conflicts \nwith the one at %s.", param.Name, sanySymbolLocation(previous))
				generated[i].SANYRange = SanyRange{Begin: param.Pos, End: param.Pos.SourceEnd()}
			}
		}
		diags = append(diags, generated...)
	}
	return diags
}

func checkBindingName(kind, name string, pos Position, defined map[string]Position, locals map[string]bool) Diagnostics {
	if name == "" || builtinIdentifiers[name] {
		return nil
	}
	if locals != nil && locals[name] {
		return Diagnostics{errorAt(pos, "E4201", "%s %s conflicts with an existing local symbol", kind, name)}
	}
	if prev, ok := defined[name]; ok {
		if sameSourceFile(prev, pos) && positionBefore(pos, prev) {
			return nil
		}
		return Diagnostics{errorAt(pos, "E4201", "%s %s conflicts with existing symbol declared at %s", kind, name, prev)}
	}
	return nil
}

func checkBoundName(name string, pos Position, defined map[string]Position, locals map[string]bool) Diagnostics {
	if name == "" || builtinIdentifiers[name] {
		return nil
	}
	if locals != nil && locals[name] {
		return Diagnostics{errorAt(pos, "E4201", "bound symbol %s conflicts with an existing local symbol", name)}
	}
	if prev, ok := defined[name]; ok {
		if sameSourceFile(prev, pos) && prev.Line == pos.Line && positionBefore(prev, pos) {
			return nil
		}
		if sameSourceFile(prev, pos) && positionBefore(pos, prev) {
			return nil
		}
		diagnostic := errorAt(pos, "E4201", "bound symbol %s conflicts with existing symbol declared at %s", name, prev)
		diagnostic.SANYRange = SanyRange{Begin: pos, End: pos.SourceEnd()}
		diagnostic.SANYMessage = fmt.Sprintf("Multiply-defined symbol '%s': this definition or declaration conflicts \nwith the one at %s.", name, sanySymbolLocation(prev))
		return Diagnostics{diagnostic}
	}
	return nil
}

func sameSourceFile(a, b Position) bool {
	return a.File != "" && b.File != "" && a.File == b.File
}

func checkProofRef(ref ProofRef, defined map[string]Position) Diagnostics {
	if ref.Name == "" || builtinIdentifiers[ref.Name] {
		return nil
	}
	if _, ok := defined[ref.Name]; !ok {
		return Diagnostics{errorAt(ref.Pos, "E4200", "undefined identifier %s", ref.Name)}
	}
	return nil
}

func checkHideRef(ref ProofRef, theoremLikeDefs, proofStepNames map[string]bool) Diagnostics {
	if ref.Mode != "HIDE" || ref.Defs || (ref.Name == "" && ref.Expr == nil) {
		return nil
	}
	expr := ref.Expr
	if selected := sanyExprSelection(expr); selected != nil {
		if selected.newSymbol != nil || selected.assumeProve != nil {
			return nil
		}
		expr = selected.body
	}
	name := ref.Name
	switch e := expr.(type) {
	case *LiteralExpr:
		// UseOrHideNode.factCheck checks OpApplKind only. Java represents
		// booleans as builtin applications, and numbers/strings separately.
		if e.Kind != "bool" {
			return nil
		}
	case *LetExpr, *LabelExpr:
		return nil
	case *IdentExpr:
		name = e.Name
	case *CallExpr:
		if id, ok := e.Callee.(*IdentExpr); ok {
			name = id.Name
		}
	}
	if theoremLikeDefs[name] || proofStepNames[name] {
		return nil
	}
	diagnostic := errorAt(ref.Pos, "E4357", "HIDE can only refer to theorems, assumptions, or proof steps; %s is not a proof fact", ref.Name)
	diagnostic.SANYRange = SanyRange{Begin: ref.Pos, End: ref.Pos.SourceEnd()}
	diagnostic.SANYMessage = "The only expression allowed as a fact in a HIDE is \nthe name of a theorem, assumption, or step."
	return Diagnostics{diagnostic}
}

func checkProofSummary(proof ProofSummary, declKinds map[string]DeclarationKind, module *Module, spec *Spec, levelChecking bool) Diagnostics {
	var diags Diagnostics
	dependencies := newSanyLeibnizAnalyzer(spec)
	level := func(expr Expr, locals map[string]bool) tlaLevel {
		if !levelChecking {
			return constantLevel
		}
		return dependencies.substitutionLevel(expr, module, locals)
	}
	goalLevels := []tlaLevel{level(proof.Goal, nil)}
	var nonExprScopes []proofNameScope
	var boundScopes []proofNameScope
	var factScopes []proofNameScope
	factDefinitions := map[string]bool{}
	for name, ref := range dependencies.resolver.scope(module) {
		if ref.def.TheoremLike {
			factDefinitions[name] = true
		}
	}
	for _, step := range proof.Steps {
		if !levelChecking && step.AssumeProveBody != nil {
			diags = append(diags, checkAssumeProveLabels(step.AssumeProveBody, true)...)
		}
		nonExprScopes = pruneProofNameScopes(nonExprScopes, step.Depth)
		boundScopes = pruneProofNameScopes(boundScopes, step.Depth)
		factScopes = pruneProofNameScopes(factScopes, step.Depth)
		if !levelChecking {
			for _, ref := range step.UseHideRefs {
				diags = append(diags, checkHideRef(ref, factDefinitions, activeProofNames(factScopes))...)
			}
		}
		nonExprSteps := activeProofKinds(nonExprScopes)
		boundNames := activeProofNames(boundScopes)
		if !levelChecking && step.Implicit && step.Name != "" {
			diags = append(diags, errorAt(step.Pos, "E4350", "implicit proof step cannot have name %s", step.Name))
		}
		// TheoremNode.LevelCheckTemporal follows CASE and QED subproofs
		// with the enclosing goal. Other assertions start their own goals.
		for len(goalLevels) <= step.Depth {
			goalLevels = append(goalLevels, constantLevel)
		}
		goalLevel := goalLevels[step.Depth]
		goalLevels = goalLevels[:step.Depth+1]
		stepLevel := level(step.Expr, boundNames)
		for _, expr := range step.Exprs {
			stepLevel = maxTlaLevel(stepLevel, level(expr, boundNames))
		}
		for _, bound := range step.Bounds {
			stepLevel = maxTlaLevel(stepLevel, level(bound.Set, boundNames))
		}
		if goalLevel == temporalLevel && stepLevel != constantLevel {
			var diagnostic Diagnostic
			switch step.Kind {
			case "HAVE", "TAKE", "WITNESS":
				diagnostic = errorAt(step.Pos, "E4352", "temporal proof goal requires constant-level %s step", step.Kind)
				diagnostic.SANYMessage = "Non-constant TAKE, WITNESS, or HAVE for temporal goal."
			case "CASE":
				diagnostic = errorAt(step.Pos, "E4353", "temporal proof goal requires constant-level CASE step")
				diagnostic.SANYMessage = "Non-constant CASE for temporal goal."
			}
			if diagnostic.SANYMessage != "" {
				position := step.Statement
				if position.Line == 0 {
					position = step.Pos
				}
				diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
				diags = append(diags, diagnostic)
			}
		}
		childGoal := stepLevel
		if step.Kind == "CASE" || step.Kind == "QED" {
			childGoal = goalLevel
		}
		goalLevels = append(goalLevels, childGoal)

		if levelChecking && step.Kind == "PICK" && exprLevel(step.Expr, declKinds, nil) == temporalLevel {
			for _, bound := range step.Bounds {
				if exprLevel(bound.Set, declKinds, nil) != constantLevel {
					diags = append(diags, errorAt(bound.Pos, "E4354", "temporal PICK formula requires constant-level bound"))
				}
			}
		}
		if !levelChecking && step.Kind == "ASSERT" && step.Expr != nil {
			for _, ref := range step.Refs {
				if nonExprSteps[ref] != "" && declKinds[ref] == "" {
					diags = append(diags, sanyDiagnosticParameters(errorAt(step.Pos, "E4351", "proof step %s is not an expression and cannot be used as one", ref), nonExprSteps[ref]))
				}
			}
		}
		exprBoundNames := boundNames
		if step.Kind == "PICK" {
			exprBoundNames = proofNamesWithBounds(exprBoundNames, step.Bounds)
		}
		if !levelChecking {
			diags = append(diags, checkProofStepExpressionRefs(step.Expr, nonExprSteps, declKinds, exprBoundNames)...)
			for _, expr := range step.Exprs {
				diags = append(diags, checkProofStepExpressionRefs(expr, nonExprSteps, declKinds, boundNames)...)
			}
		}
		if step.Name != "" && step.Kind != "ASSERT" {
			nonExprScopes = append(nonExprScopes, proofNameScope{Depth: step.Depth, Names: map[string]bool{step.Name: true}, Kinds: map[string]string{step.Name: step.Kind}})
		}
		if step.Name != "" {
			factScopes = append(factScopes, proofNameScope{Depth: step.Depth, Names: map[string]bool{step.Name: true, step.QualifiedName: true}})
		}
		if len(step.Bounds) > 0 && (step.Kind == "PICK" || step.Kind == "TAKE") {
			boundScopes = append(boundScopes, proofNameScope{Depth: step.Depth, Names: proofBoundNames(step.Bounds)})
		}
	}
	return diags
}

type proofNameScope struct {
	Kinds map[string]string
	Depth int
	Names map[string]bool
}

func pruneProofNameScopes(scopes []proofNameScope, depth int) []proofNameScope {
	out := scopes[:0]
	for _, scope := range scopes {
		if scope.Depth <= depth {
			out = append(out, scope)
		}
	}
	return out
}

func activeProofKinds(scopes []proofNameScope) map[string]string {
	kinds := map[string]string{}
	for _, scope := range scopes {
		for name, kind := range scope.Kinds {
			kinds[name] = kind
		}
	}
	return kinds
}

func activeProofNames(scopes []proofNameScope) map[string]bool {
	names := map[string]bool{}
	for _, scope := range scopes {
		for name := range scope.Names {
			names[name] = true
		}
	}
	return names
}

func proofBoundNames(bounds []BoundVar) map[string]bool {
	names := map[string]bool{}
	for _, bound := range bounds {
		if bound.Name != "" {
			names[bound.Name] = true
		}
	}
	return names
}

func proofNamesWithBounds(names map[string]bool, bounds []BoundVar) map[string]bool {
	if len(bounds) == 0 {
		return names
	}
	out := make(map[string]bool, len(names)+len(bounds))
	for name := range names {
		out[name] = true
	}
	for _, bound := range bounds {
		if bound.Name != "" {
			out[bound.Name] = true
		}
	}
	return out
}

func proofNamesWithName(names map[string]bool, name string) map[string]bool {
	if name == "" {
		return names
	}
	out := make(map[string]bool, len(names)+1)
	for existing := range names {
		out[existing] = true
	}
	out[name] = true
	return out
}

func checkProofStepExpressionRefs(expr Expr, nonExprSteps map[string]string, declKinds map[string]DeclarationKind, boundNames map[string]bool) Diagnostics {
	if expr == nil || len(nonExprSteps) == 0 {
		return nil
	}
	var diags Diagnostics
	switch e := expr.(type) {
	case *IdentExpr:
		if !boundNames[e.Name] && nonExprSteps[e.Name] != "" && declKinds[e.Name] == "" {
			diags = append(diags, sanyDiagnosticParameters(errorAt(e.Pos, "E4351", "proof step %s is not an expression and cannot be used as one", e.Name), nonExprSteps[e.Name]))
		}
		return diags
	case *QuantifierExpr:
		diags = append(diags, checkProofStepExpressionRefs(e.Set, nonExprSteps, declKinds, boundNames)...)
		diags = append(diags, checkProofStepExpressionRefs(e.Body, nonExprSteps, declKinds, proofNamesWithName(boundNames, e.Var))...)
		return diags
	case *ChooseExpr:
		diags = append(diags, checkProofStepExpressionRefs(e.Set, nonExprSteps, declKinds, boundNames)...)
		diags = append(diags, checkProofStepExpressionRefs(e.Body, nonExprSteps, declKinds, proofNamesWithBounds(boundNames, e.boundVars()))...)
		return diags
	case *FunctionExpr:
		for _, bound := range e.Bounds {
			diags = append(diags, checkProofStepExpressionRefs(bound.Set, nonExprSteps, declKinds, boundNames)...)
		}
		diags = append(diags, checkProofStepExpressionRefs(e.Body, nonExprSteps, declKinds, proofNamesWithBounds(boundNames, e.Bounds))...)
		return diags
	case *SetComprehensionExpr:
		for _, bound := range e.Bounds {
			diags = append(diags, checkProofStepExpressionRefs(bound.Set, nonExprSteps, declKinds, boundNames)...)
		}
		bodyNames := proofNamesWithBounds(boundNames, e.Bounds)
		diags = append(diags, checkProofStepExpressionRefs(e.Element, nonExprSteps, declKinds, bodyNames)...)
		diags = append(diags, checkProofStepExpressionRefs(e.Predicate, nonExprSteps, declKinds, bodyNames)...)
		return diags
	}
	for _, child := range sanySubexpressionChildren(expr) {
		diags = append(diags, checkProofStepExpressionRefs(child, nonExprSteps, declKinds, boundNames)...)
	}
	return diags
}

func checkAssumeProveDefinitionUse(expr Expr, assumeProveDefs map[string]bool, locals map[string]bool) Diagnostics {
	var diags Diagnostics
	switch e := expr.(type) {
	case *IdentExpr:
		if sanyExpressionGenerationFailure(e) != sanyGenerationSucceeded {
			return nil
		}
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
		recursiveNames := letRecursiveNames(e)
		for _, def := range e.Definitions {
			defLocals := letDefinitionBodyLocals(letLocals, def, recursiveNames[def.Name])
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
		diags = append(diags, checkAssumeProveDefinitionUse(e.Body, assumeProveDefs, withLocal(locals, e.boundNames()...))...)
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

func moduleLevelDeclKinds(mod *Module, spec *Spec, base map[string]DeclarationKind) map[string]DeclarationKind {
	kinds := copyDeclKindMap(base)
	visited := map[string]bool{}
	var collect func(*Module)
	collect = func(cur *Module) {
		if cur == nil || visited[cur.Name] || isEmbeddedStandardModule(cur) {
			return
		}
		visited[cur.Name] = true
		for _, ext := range cur.Extends {
			if spec != nil {
				collect(spec.Modules[ext])
			}
		}
		for _, decl := range cur.Declarations {
			for _, name := range decl.Names {
				kinds[name] = decl.Kind
				if cur.Name != "" {
					kinds[cur.Name+"!"+name] = decl.Kind
				}
			}
		}
		for _, decl := range cur.Recursives {
			for _, name := range decl.Names {
				kinds[name] = RecursiveDecl
				if cur.Name != "" {
					kinds[cur.Name+"!"+name] = RecursiveDecl
				}
			}
		}
		for _, def := range cur.Definitions {
			kinds[def.Name] = OperatorDecl
			if cur.Name != "" {
				kinds[cur.Name+"!"+def.Name] = OperatorDecl
			}
		}
		for _, assumption := range cur.Assumptions {
			if assumption.Name == "" {
				continue
			}
			kinds[assumption.Name] = OperatorDecl
			if cur.Name != "" {
				kinds[cur.Name+"!"+assumption.Name] = OperatorDecl
			}
		}
	}
	collect(mod)
	return kinds
}

func builtinArgMaxLevel(info sanyBuiltinOperator, index int) (tlaLevel, bool) {
	if index < 0 || len(info.argMaxLevels) == 0 {
		return constantLevel, false
	}
	if info.arity == -1 {
		if index < len(info.argMaxLevels) {
			return info.argMaxLevels[index], true
		}
		return info.argMaxLevels[len(info.argMaxLevels)-1], true
	}
	if index >= len(info.argMaxLevels) {
		return constantLevel, false
	}
	return info.argMaxLevels[index], true
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
	Pos   Position
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
			targets[name] = substitutionTarget{Kind: decl.Kind, Arity: arity, Pos: declarationSymbolPosition(decl, name)}
		}
	}
	return targets
}

func moduleSubstitutionTargets(mod *Module, spec *Spec) map[string]substitutionTarget {
	targets := map[string]substitutionTarget{}
	var collect func(*Module, map[string]bool)
	collect = func(cur *Module, visiting map[string]bool) {
		if cur == nil || visiting[cur.Name] || isEmbeddedStandardModule(cur) {
			return
		}
		visiting[cur.Name] = true
		for _, ext := range cur.Extends {
			if spec != nil {
				collect(spec.Modules[ext], visiting)
			}
		}
		for name, target := range moduleOwnSubstitutionTargets(cur) {
			targets[name] = target
		}
		visiting[cur.Name] = false
	}
	collect(mod, map[string]bool{})
	return targets
}

// InstanceNode applies declaration-level matching when ModuleNode.isConstant
// is false. That predicate includes operator bodies and theorems, not just
// the presence of VARIABLE declarations.
func moduleRequiresSubstitutionLevelMatch(mod *Module, spec *Spec) bool {
	if mod == nil {
		return false
	}
	for _, symbol := range semanticModuleExports(mod, spec, map[string]bool{}) {
		if symbol.kind == VariableDecl {
			return true
		}
	}
	analyzer := newSanyLeibnizAnalyzer(spec)
	for name, ref := range analyzer.resolver.scope(mod) {
		// getOpDefs excludes ThmOrAssumpDefNode and module-instance placeholders.
		if ref.def.TheoremLike {
			continue
		}
		if analyzer.substitutionLevel(&IdentExpr{Name: name}, mod, nil) != constantLevel {
			return true
		}
	}
	// copyTheorems copies EXTENDS theorems into the module's theorem vector.
	// INSTANCE creates context definitions, not entries in that vector.
	seen := map[*Module]bool{}
	var nonconstantTheorem func(*Module) bool
	nonconstantTheorem = func(current *Module) bool {
		if current == nil || seen[current] {
			return false
		}
		seen[current] = true
		for _, theorem := range current.Theorems {
			if analyzer.substitutionLevel(theorem.Expr, current, nil) != constantLevel {
				return true
			}
		}
		for _, name := range current.Extends {
			if spec != nil && nonconstantTheorem(spec.Modules[name]) {
				return true
			}
		}
		return false
	}
	return nonconstantTheorem(mod)
}

func moduleImplicitSubstitutions(mod *Module, spec *Spec) map[string]int {
	arities := map[string]int{}
	var collect func(*Module, map[string]bool)
	collect = func(cur *Module, visiting map[string]bool) {
		if cur == nil || visiting[cur.Name] {
			return
		}
		visiting[cur.Name] = true
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
		for _, dep := range cur.Extends {
			if spec != nil {
				collect(spec.Modules[dep], visiting)
			}
		}
		visiting[cur.Name] = false
	}
	collect(mod, map[string]bool{})
	return arities
}

func substitutionExprArity(expr Expr, arities map[string]int) int {
	if selected := sanyExprSelection(expr); selected != nil && selected.operator {
		return len(selected.params)
	}
	if ident, ok := expr.(*IdentExpr); ok {
		if arity, exists := arities[ident.Name]; exists {
			return arity
		}
		if arity, exists := builtinOperatorArity(ident.Name); exists {
			return arity
		}
	}
	if lambda, ok := expr.(*FunctionExpr); ok && lambda.IsLambda {
		return len(lambda.Bounds)
	}
	return 0
}

func substitutionBuiltinOperatorInfo(expr Expr) (sanyBuiltinOperator, bool) {
	ident, ok := expr.(*IdentExpr)
	if !ok {
		return sanyBuiltinOperator{}, false
	}
	return sanyBuiltinOperatorInfo(ident.Name)
}

func substitutionExprIsOperatorArgument(expr Expr, targetArity int, arities map[string]int) bool {
	if targetArity <= 0 {
		return false
	}
	ident, ok := expr.(*IdentExpr)
	if !ok {
		return false
	}
	_, exists := arities[ident.Name]
	if !exists {
		_, exists = builtinOperatorArity(ident.Name)
	}
	// A mismatched operator is still an OpArgNode; do not regenerate it as
	// a zero-argument expression after reporting its substitution arity.
	return exists
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
	// OpApplNode.getNumberOfBoundedBoundSymbols counts a tuple binder once.
	count := 0
	for i := 0; i < len(fn.Bounds); {
		bound := fn.Bounds[i]
		count++
		i++
		if bound.TupleBound {
			for i < len(fn.Bounds) && fn.Bounds[i].TupleBound && fn.Bounds[i].Set == bound.Set {
				i++
			}
		}
	}
	return count, true
}

type operatorParamSpec struct {
	Name          string
	Arity         int
	ArgMinLevels  []tlaLevel
	ArgParamNames []string
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
	enrichOperatorParamSpecsFromExpr(specs, def.Expr)
	return specs, true
}

func enrichOperatorParamSpecsFromExpr(specs []operatorParamSpec, expr Expr) {
	if len(specs) == 0 || expr == nil {
		return
	}
	byName := map[string]int{}
	for i, spec := range specs {
		byName[spec.Name] = i
	}
	var walk func(Expr)
	walk = func(cur Expr) {
		if cur == nil {
			return
		}
		if call, ok := cur.(*CallExpr); ok {
			if ident, ok := call.Callee.(*IdentExpr); ok {
				if specIndex, exists := byName[ident.Name]; exists && specs[specIndex].Arity >= 0 {
					ensureOperatorParamArgMetadata(&specs[specIndex], len(call.Args))
					for i, arg := range call.Args {
						if argIdent, ok := arg.(*IdentExpr); ok {
							if _, isParam := byName[argIdent.Name]; isParam {
								specs[specIndex].ArgParamNames[i] = argIdent.Name
								continue
							}
						}
						level := exprLevel(arg, nil, nil)
						if level > specs[specIndex].ArgMinLevels[i] {
							specs[specIndex].ArgMinLevels[i] = level
						}
					}
				}
			}
		}
		if let, ok := cur.(*LetExpr); ok {
			for _, def := range let.Definitions {
				walk(def.Expr)
			}
		}
		for _, child := range sanySubexpressionChildren(cur) {
			walk(child)
		}
	}
	walk(expr)
}

func ensureOperatorParamArgMetadata(spec *operatorParamSpec, n int) {
	for len(spec.ArgMinLevels) < n {
		spec.ArgMinLevels = append(spec.ArgMinLevels, constantLevel)
	}
	for len(spec.ArgParamNames) < n {
		spec.ArgParamNames = append(spec.ArgParamNames, "")
	}
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

// The source represents a quantified variable list in one node. Native wrappers
// sharing that source node must generate their domains in the enclosing scope.
func sanyQuantifierGroup(root *QuantifierExpr) ([]*QuantifierExpr, Expr) {
	parameters := []*QuantifierExpr{root}
	body := root.Body
	for root.Syntax != nil {
		next, ok := body.(*QuantifierExpr)
		if !ok || next.Syntax != root.Syntax || next.Kind != root.Kind {
			break
		}
		parameters = append(parameters, next)
		body = next.Body
	}
	return parameters, body
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

func instanceNameSentinel(name string) string {
	return "\x00instance-name:" + name
}

func checkExpr(expr Expr, defined map[string]Position, locals map[string]bool, generators ...*sanyExpressionGeneration) Diagnostics {
	generation := sanyExpressionGenerator(generators)
	return generation.checkExpr(expr, defined, locals)
}

func (generation *sanyExpressionGeneration) checkExpr(expr Expr, defined map[string]Position, locals map[string]bool, selectorOrigins ...*SanySelector) (result Diagnostics) {
	var diags Diagnostics
	allowLabeledAP := generation.allowLabeledAP
	generation.allowLabeledAP = false
	defer func() { generation.allowLabeledAP = allowLabeledAP }()
	fact := generation.fact
	operatorArgument := generation.operatorArgument
	symbolReferenceOnly := generation.symbolReferenceOnly
	generation.fact = false
	generation.operatorArgument = false
	generation.symbolReferenceOnly = false
	defer func() {
		generation.fact = fact
		generation.operatorArgument = operatorArgument
		generation.symbolReferenceOnly = symbolReferenceOnly
	}()
	defer func() {
		// generateExpression's GeneralId branch returns this Generator's
		// nullOAN on selector failure; enclosing constructors still use it.
		if source := sanyGenerationSource(expr); source != nil && source.Syntax != nil && source.Syntax.Kind.JavaName() == "N_GeneralId" && source.semanticGraph == nil && sanyExpressionGenerationFailure(expr) == sanyGenerationNullOperator {
			generation.retainNullOperatorOperand(expr, false)
		}
	}()
	setSanyExpressionGenerationFailure(expr, sanyGenerationSucceeded)
	if source := sanyGenerationSource(expr); source != nil {
		source.semanticGraph = nil
	}
	if source := sanyExprSource(expr); source != nil {
		source.operatorArgumentsGenerated = false
	}
	if source := sanyExprSource(expr); source != nil && source.Selector != nil {
		// Selector.finish reports every unrecognized syntax kind before
		// selectorToNode starts resolving the first name.
		var constructorDiags Diagnostics
		for _, step := range source.Selector.Steps {
			if step.Kind == 0 && step.Syntax != nil && step.Syntax.Kind.JavaName() != "N_StructOp" {
				diagnostic := errorAt(sanyNodePosition(step.Syntax), "E4003", "Unexpected token found.")
				diagnostic.SANYRange = step.Syntax.Range
				diagnostic.SANYMessage = diagnostic.Message
				constructorDiags = append(constructorDiags, diagnostic)
			}
		}
		defer func() { result = append(constructorDiags, result...) }()
		if len(source.Selector.Steps) > 0 && source.Selector.Steps[0].Kind != SanySelectorName {
			step := source.Selector.Steps[0]
			message := fmt.Sprintf("Need name or step number here, not `%s'.", step.Name)
			diagnostic := errorAt(sanyNodePosition(step.Syntax), "E4005", "%s", message)
			diagnostic.SANYRange = step.Syntax.Range
			diagnostic.SANYMessage = message
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
			return Diagnostics{diagnostic}
		}
		name := ""
		opDefArityFound := 0
		for i, step := range source.Selector.Steps {
			if i == len(source.Selector.Steps)-1 || step.Kind != SanySelectorName {
				break
			}
			if name != "" {
				name += "!"
			}
			name += step.Name
			// An unapplied INSTANCE prefix contributes its formal parameters
			// to a selected operator argument, rather than requiring arguments.
			if operatorArgument && step.Arguments == nil {
				continue
			}
			symbol, exists := generation.lookupSymbol(name, defined)
			nodeArity := opDefArityFound + symbol.arity
			if actual := generation.formalSymbolTable().resolveSymbol(name); actual != nil && actual.semKind() == sanyModuleInstanceKind {
				symbol.kind, symbol.arity, exists = InstanceDecl, actual.semArity(), true
				nodeArity = actual.semArity()
			}
			if exists && symbol.kind == InstanceDecl && symbol.arity >= 0 {
				// Imported INSTANCE signatures include earlier name components'
				// parameters. Java checks only the remaining arity here.
				remaining := nodeArity - opDefArityFound
				count := 0
				if step.Arguments != nil {
					count = len(expressionChildren(step.Arguments))
				}
				if count != remaining {
					position := expr.Position()
					location := SanyRange{Begin: position, End: position.SourceEnd()}
					if source.Syntax != nil {
						location = source.Syntax.Range
					}
					if step.Arguments != nil {
						location = step.Arguments.Range
					}
					diagnostic := sanyDiagnosticParameters(errorAt(position, "E4204", "The operator %s requires %d arguments.", name, remaining), name, remaining)
					diagnostic.SANYMessage = diagnostic.Message
					diagnostic.SANYRange = location
					setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
					return Diagnostics{diagnostic}
				}
				opDefArityFound = nodeArity
			}
		}
	}
	if source := sanyExprSource(expr); source != nil && source.selectorFailure {
		// selectorToNode has already reported the error and returned nullOAN.
		setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
		if source.selectorDiagnostic != nil {
			return Diagnostics{*source.selectorDiagnostic}
		}
		return nil
	}
	if selected := sanyExprSelection(expr); selected != nil {
		// Selector preparation cannot make a later declaration visible. Resolve
		// the selected symbol in this body's actual generation context first.
		if missing := checkSanySelectedSymbol(expr, selected, defined, locals); len(missing) != 0 {
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
			return missing
		}
		// The selected body is checked in its declaration's lexical scope.
		// Only the actual arguments originate in this use site's scope.
		owner := &IdentExpr{Name: selected.name, Pos: expr.Position()}
		for i, arg := range selected.args {
			expected := 0
			if i < len(selected.params) {
				expected = selected.params[i].OperatorArity
			}
			diags = append(diags, generation.generateOperatorOperand(owner, i, expected, arg, defined, locals)...)
		}
		diags = append(diags, generation.retainCanonicalInstanceSelection(expr, operatorArgument, symbolReferenceOnly)...)
		if sanyGeneratedExpressionNode(expr) == nil {
			if !operatorArgument && !symbolReferenceOnly {
				diags = append(diags, generation.retainCanonicalLabelSelection(expr)...)
			}
		}
		if sanyGeneratedExpressionNode(expr) == nil {
			diags = append(diags, generation.retainCanonicalSubexpression(expr, operatorArgument, symbolReferenceOnly, fact)...)
		}
		return diags
	}
	switch e := expr.(type) {
	case *IdentExpr:
		if e.Name == "@" && e.proofAtTarget == nil && len(generation.excepts) > 0 && len(generation.exceptSpecs) > 0 {
			e.semanticGraph = newSanySemAtNode(generation.excepts[len(generation.excepts)-1], generation.exceptSpecs[len(generation.exceptSpecs)-1])
			return nil
		}
		if e.proofAtTarget != nil {
			// generateProof constructs $Nop with the already-generated previous
			// RHS. Do not regenerate it or repeat its diagnostics in this scope.
			arity := 0
			e.generationArity = &arity
			return nil
		}
		if symbol, exists := generation.lookupSymbol(e.Name, defined); exists && symbol.proofStepKind != "" {
			if symbol.proofStepKind == "DEFINE" || symbol.proofStepKind == "USE" || symbol.proofStepKind == "HIDE" || symbol.proofStepKind == "INSTANCE" {
				setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
				message := "Step number of non-expression step used as an expression."
				if fact {
					message = "Step number of non-fact used as a fact"
				}
				diagnostic := errorAt(e.Pos, "E4004", "%s", message)
				diagnostic.SANYMessage = message
				diagnostic.SANYRange = SanyRange{Begin: e.Pos, End: e.Pos.SourceEnd()}
				return Diagnostics{diagnostic}
			}
			if !fact && symbol.proofAssumeProve {
				setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
				diagnostic := errorAt(e.Pos, "E4355", "ASSUME/PROVE used where an expression is required.")
				diagnostic.SANYMessage = diagnostic.Message
				diagnostic.SANYRange = SanyRange{Begin: e.Pos, End: e.Pos.SourceEnd()}
				return Diagnostics{diagnostic}
			}
			if !fact && symbol.proofStepKind != "ASSERT" {
				kind := symbol.proofStepKind
				if kind == "QED" {
					kind = "QED step"
				}
				setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
				diagnostic := sanyDiagnosticParameters(errorAt(e.Pos, "E4351", "%s proof step selected instead of expression.", kind), kind)
				diagnostic.SANYMessage = diagnostic.Message
				diagnostic.SANYRange = SanyRange{Begin: e.Pos, End: e.Pos.SourceEnd()}
				return Diagnostics{diagnostic}
			}
		}
		e.formalNode = nil
		e.declarationNode = nil
		if symbol, exists := generation.lookupSymbol(e.Name, defined); exists {
			e.formalNode = symbol.formalNode
			e.declarationNode = symbol.declarationNode
		}
		e.generationArity = nil
		if generation.symbols != nil {
			if symbol, exists := generation.lookupSymbol(e.Name, defined); exists && symbol.arity >= 0 {
				arity := symbol.arity
				e.generationArity = &arity
			}
		}
		var graphSymbol sanySemSymbol
		if e.formalNode != nil {
			graphSymbol = e.formalNode
		} else if e.declarationNode != nil {
			graphSymbol = e.declarationNode
		} else if symbol, exists := generation.lookupSymbol(e.Name, defined); exists && symbol.theoremDefNode != nil {
			graphSymbol = symbol.theoremDefNode
		} else if symbol, exists := generation.lookupSymbol(e.Name, defined); exists && symbol.opDefNode != nil {
			graphSymbol = symbol.opDefNode
		} else if symbol := generation.canonicalInstanceSelectorSymbol(e); symbol != nil {
			graphSymbol = symbol
		} else {
			graphSymbol = sanyGlobalInitialContext(false).getSymbol(e.Name)
		}
		if definition, ok := graphSymbol.(*sanySemThmOrAssumpDefNode); ok && !fact && definition.body != nil && definition.body.Kind() == tlc.SemanticKind(sanyAssumeProveKind) {
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
			diagnostic := errorAt(e.Pos, "E4355", "ASSUME/PROVE used where an expression is required.")
			diagnostic.SANYMessage = diagnostic.Message
			diagnostic.SANYRange = SanyRange{Begin: e.Pos, End: e.Pos.SourceEnd()}
			if e.Syntax != nil {
				diagnostic.SANYRange = e.Syntax.Range
			}
			return Diagnostics{diagnostic}
		}
		if !symbolReferenceOnly && !operatorArgument && graphSymbol != nil && graphSymbol.semArity() > 0 {
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
			diagnostic := sanyDiagnosticParameters(errorAt(e.Pos, "E4204", "operator %s arity mismatch: got 0 args, want %d", e.Name, graphSymbol.semArity()), e.Name, graphSymbol.semArity())
			diagnostic.SANYMessage = fmt.Sprintf("The operator %s requires %d arguments.", e.Name, graphSymbol.semArity())
			if e.Syntax != nil {
				diagnostic.SANYRange = e.Syntax.Range
			}
			return Diagnostics{diagnostic}
		}
		if graphSymbol != nil && graphSymbol.semKind() == sanyModuleInstanceKind {
			if !fact && !symbolReferenceOnly {
				setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
				return Diagnostics{sanyIncompleteOperatorDiagnostic(e)}
			}
			e.semanticGraph = graphSymbol.(sanySemanticGraphNode)
			return nil
		}
		if !symbolReferenceOnly {
			diags = append(diags, retainSanySymbolReference(e, graphSymbol, operatorArgument, generation.currentModule)...)
		}
		if e.Syntax != nil && e.Syntax.Kind.JavaName() == "N_GeneralId" {
			if application, ok := sanyGeneratedExpressionNode(e).(*sanySemOpApplNode); ok {
				generation.checkFunctionRecursion(application.operator.semName())
			}
		}

		if e.Name == "" || e.formalNode != nil || localIdentifierInScope(locals, e.Name) || builtinIdentifiers[e.Name] {
			return nil
		}
		if _, ok := builtinOperatorArity(e.Name); ok {
			return nil
		}
		if base, ok := theoremStatementReferenceBase(e.Name); ok {
			if _, ok := defined[base]; ok {
				return nil
			}
		}
		if subexpressionReferenceNameDefined(e.Name, defined) {
			if base, _, ok := strings.Cut(e.Name, "!"); ok {
				_, exact := defined[e.Name]
				if _, isInstance := defined[instanceNameSentinel(base)]; isInstance && !exact {
					setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
					diags = append(diags, sanyUndefinedIdentifierDiagnostic(e, selectorOrigins...))
					return diags
				}
			}
			return nil
		}
		if _, ok := defined[e.Name]; !ok {
			if e.Name == "@" {
				diagnostic := errorAt(e.Pos, "E4261", "@ may only be used inside a function EXCEPT replacement")
				diagnostic.SANYMessage = "@ used where its meaning is not defined."
				diagnostic.SANYRange = SanyRange{Begin: e.Pos, End: e.Pos.SourceEnd()}
				setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
				diags = append(diags, diagnostic)
				return diags
			}
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
			diags = append(diags, sanyUndefinedIdentifierDiagnostic(e, selectorOrigins...))
		} else if _, ok := defined[instanceNameSentinel(e.Name)]; ok {
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
			if want := generation.moduleArities[e.Name]; want != 0 {
				diagnostic := sanyDiagnosticParameters(errorAt(e.Pos, "E4204", "operator %s arity mismatch: got 0 args, want %d", e.Name, want), e.Name, want)
				diagnostic.SANYMessage = fmt.Sprintf("The operator %s requires %d arguments.", e.Name, want)
				diags = append(diags, diagnostic)
			} else if !fact {
				diags = append(diags, sanyIncompleteOperatorDiagnostic(e))
			} else {
				setSanyExpressionGenerationFailure(expr, sanyGenerationSucceeded)
			}
		}
	case *LiteralExpr:
		// Generator constructs literal nodes before the evaluator
		// bridge processes constants. A fresh generation replaces prior nodes.
		e.numeralNode, e.decimalNode, e.stringNode = nil, nil, nil
		if e.Kind == "number" && strings.Contains(e.Value, ".") {
			parts := strings.SplitN(e.Value, ".", 2)
			node := tlc.NewDecimalNodeFromParts(parts[0], parts[1])
			bridge := tlcBridge{}
			bridge.withExprLocation(e, node)
			sanyCanonicalLiteral(node, &node.SemanticNodeBase)
			e.decimalNode = node
			e.semanticGraph = node
		} else if e.Kind == "number" {
			node, err := tlc.NewNumeralNodeFromString(e.Value)
			if err != nil {
				panic(err)
			}
			bridge := tlcBridge{}
			bridge.withExprLocation(e, node)
			sanyCanonicalLiteral(node, &node.SemanticNodeBase)
			e.numeralNode = node
			e.semanticGraph = node
		} else if e.Kind == "string" {
			node := tlc.NewStringNode(e.Value)
			bridge := tlcBridge{}
			bridge.withExprLocation(e, node)
			sanyCanonicalLiteral(node, &node.SemanticNodeBase)
			e.stringNode = node
			e.semanticGraph = node
		} else if e.Kind == "bool" {
			diags = append(diags, retainSanySymbolReference(e, sanyGlobalInitialContext(false).getSymbol(e.Value), operatorArgument, generation.currentModule)...)
		}
	case *UnaryExpr:
		if unresolved := checkSanyOperatorSymbolDefined(e.Op, e.Pos, e.Syntax, defined, locals); len(unresolved) != 0 {
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullExpression)
			return unresolved
		}
		operator := generation.applicationOperator(e.Op, e.Syntax, defined)
		diags = append(diags, generation.checkExpr(e.Expr, defined, locals)...)
		if e.Syntax != nil && (e.Syntax.Kind.JavaName() == "N_ConjList" || e.Syntax.Kind.JavaName() == "N_DisjList") {
			name := "$ConjList"
			if e.Syntax.Kind.JavaName() == "N_DisjList" {
				name = "$DisjList"
			}
			retainSanyBuiltInApplication(e, name, []Expr{e.Expr})
		} else {
			diags = append(diags, retainSanyMatchedApplication(e, operator, []Expr{e.Expr})...)
		}
	case *BinaryExpr:
		if unresolved := checkSanyOperatorSymbolDefined(e.Op, e.Pos, e.Syntax, defined, locals); len(unresolved) != 0 {
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullExpression)
			return unresolved
		}
		operator := generation.applicationOperator(e.Op, e.Syntax, defined)
		diags = append(diags, generation.checkExpr(e.Left, defined, locals)...)
		diags = append(diags, generation.checkExpr(e.Right, defined, locals)...)
		if e.Syntax != nil {
			switch e.Syntax.Kind.JavaName() {
			case "N_ConjList", "N_DisjList":
				name := "$ConjList"
				if e.Syntax.Kind.JavaName() == "N_DisjList" {
					name = "$DisjList"
				}
				retainSanyBuiltInApplication(e, name, sanySourceNaryOperands(e))
			case "N_Times":
				retainSanyBuiltInApplication(e, "$CartesianProd", sanySourceNaryOperands(e))
			default:
				diags = append(diags, retainSanyMatchedApplication(e, operator, []Expr{e.Left, e.Right})...)
			}
		} else if !e.JunctionList && !e.SanyNary {
			diags = append(diags, retainSanyMatchedApplication(e, operator, []Expr{e.Left, e.Right})...)
		}
	case *CallExpr:
		if ident, ok := e.Callee.(*IdentExpr); ok {
			_, instance := defined[instanceNameSentinel(ident.Name)]
			_, visible := defined[ident.Name]
			if instance && visible && !locals[ident.Name] {
				setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
				if want := generation.moduleArities[ident.Name]; want != len(e.Args) {
					return Diagnostics{sanyCallArityDiagnostic(e, ident.Name, want)}
				}
				specs := generation.moduleOperatorParams[ident.Name]
				for i, arg := range e.Args {
					expected := 0
					if i < len(specs) {
						expected = specs[i].Arity
					}
					diags = append(diags, generation.generateOperatorOperand(ident, i, expected, arg, defined, locals)...)
				}
				if fact {
					setSanyExpressionGenerationFailure(expr, sanyGenerationSucceeded)
					return diags
				}
				return append(diags, sanyIncompleteOperatorDiagnostic(ident))
			}
		}
		// selectorToNode resolves the symbol and checks its supplied arity before
		// generating operands. The callee is a name, not a separate expression.
		var operator sanySemSymbol
		if identifier, ok := e.Callee.(*IdentExpr); ok && (e.Selector == nil || len(e.Selector.Steps) <= 1 || generation.canonicalInstanceSelectorSymbol(e) != nil) {
			operator = generation.canonicalInstanceSelectorSymbol(e)
			if operator == nil {
				operator = generation.applicationOperator(identifier.Name, nil, defined)
			}
			if operator != nil && operator.semArity() >= 0 && operator.semArity() != len(e.Args) {
				setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
				return Diagnostics{sanyCallArityDiagnostic(e, identifier.Name, operator.semArity())}
			}
		}
		generation.symbolReferenceOnly = true
		// Native calls separate a callee image from the source selector. Java
		// resolves using that selector directly; retain it for name and location
		// accumulation instead of diagnosing the flattened navigation image.
		diags = append(diags, generation.checkExpr(e.Callee, defined, locals, e.Selector)...)
		generation.symbolReferenceOnly = false
		if sanyExpressionGenerationFailure(e.Callee) != sanyGenerationSucceeded {
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullOperator)
			return diags
		}
		_, signatures := generation.proofSignatures()
		var specs []operatorParamSpec
		if identifier, ok := e.Callee.(*IdentExpr); ok {
			specs = signatures[identifier.Name]
			if definition, ok := operator.(*sanySemOpDefNode); ok && (definition.semKind() == sanyUserDefinedOpKind || definition.semKind() == sanyModuleInstanceKind) {
				specs = make([]operatorParamSpec, len(definition.formalNodes))
				for i, parameter := range definition.formalNodes {
					specs[i] = operatorParamSpec{Name: parameter.semName(), Arity: parameter.semArity()}
				}
			}
			if len(specs) == len(e.Args) {
				higherOrder := false
				for _, spec := range specs {
					higherOrder = higherOrder || spec.Arity > 0
				}
				if higherOrder {
					return append(diags, generation.generateApplicationOperands(e, identifier, operator, specs, defined, locals)...)
				}
			}
		}
		for i, arg := range e.Args {
			if definition, ok := operator.(*sanySemOpDefNode); ok && definition.semKind() == sanyUserDefinedOpKind && i >= len(definition.formalNodes) {
				generation.retainNullOperatorOperand(arg, false)
				continue
			}
			var generated Diagnostics
			if identifier, ok := e.Callee.(*IdentExpr); ok && i < len(specs) && specs[i].Arity > 0 {
				owner := *identifier
				owner.Pos, owner.Syntax = e.Pos, e.Syntax
				if e.Syntax != nil {
					owner.Pos = sanyNodePosition(e.Syntax)
				}
				generated = generation.generateOperatorOperand(&owner, i, specs[i].Arity, arg, defined, locals)
			} else {
				generated = generation.checkExpr(arg, defined, locals)
			}
			for j := range generated {
				if generated[j].Code == "E4275" && i < len(specs) {
					generated[j].Message = fmt.Sprintf("expression parameter %s cannot accept a LAMBDA operator argument", specs[i].Name)
				}
			}
			diags = append(diags, generated...)
		}
		diags = append(diags, retainSanyMatchedApplication(e, operator, e.Args)...)
		if _, complete := sanyGeneratedExpressionNode(e).(*sanySemOpApplNode); complete {
			e.operatorArgumentsGenerated = true
		}
	case *IfExpr:
		diags = append(diags, generation.checkExpr(e.Cond, defined, locals)...)
		diags = append(diags, generation.checkExpr(e.Then, defined, locals)...)
		diags = append(diags, generation.checkExpr(e.Else, defined, locals)...)
		retainSanyBuiltInApplication(e, "$IfThenElse", []Expr{e.Cond, e.Then, e.Else})
	case *LetExpr:
		diags = append(diags, generation.checkLet(e, defined, locals)...)
	case *QuantifierExpr:
		diags = append(diags, generation.checkQuantifier(e, defined, locals)...)
	case *CaseExpr:
		pairs := make([]sanySemanticGraphNode, 0, len(e.Arms)+1)
		var children []*SanySyntaxNode
		if e.Syntax != nil {
			children = e.Syntax.GetHeirs()
		}
		for i, arm := range e.Arms {
			diags = append(diags, generation.checkExpr(arm.Test, defined, locals)...)
			diags = append(diags, generation.checkExpr(arm.Value, defined, locals)...)
			var syntax *SanySyntaxNode
			if 2*i+1 < len(children) {
				syntax = children[2*i+1]
			}
			pairs = append(pairs, sanyGeneratedCasePair(arm.Test, arm.Value, syntax, false))
		}
		if e.Other != nil {
			diags = append(diags, generation.checkExpr(e.Other, defined, locals)...)
			var syntax *SanySyntaxNode
			if 2*len(e.Arms)+1 < len(children) {
				syntax = children[2*len(e.Arms)+1]
			}
			pairs = append(pairs, sanyGeneratedCasePair(nil, e.Other, syntax, true))
		}
		retainSanyGeneratedOperands(e, "$Case", pairs)
	case *ChooseExpr:
		diags = append(diags, generation.checkChoose(e, defined, locals)...)
	case *TupleExpr:
		for _, elem := range e.Elems {
			diags = append(diags, generation.checkExpr(elem, defined, locals)...)
		}
		retainSanyBuiltInApplication(e, "$Tuple", e.Elems)
	case *SetExpr:
		for _, elem := range e.Elems {
			diags = append(diags, generation.checkExpr(elem, defined, locals)...)
		}
		retainSanyBuiltInApplication(e, "$SetEnumerate", e.Elems)
	case *RecordExpr:
		fields := make([]sanyRecordGenerationField, len(e.Fields))
		for i, field := range e.Fields {
			fields[i] = sanyRecordGenerationField{field.Name, field.Value, field.Pos}
		}
		diags = append(diags, generation.checkRecordForm(e, "$RcdConstructor", fields, defined, locals)...)
	case *RecordComponentExpr:
		diags = append(diags, generation.checkExpr(e.Record, defined, locals)...)
		if e.Syntax != nil {
			children := e.Syntax.GetHeirs()
			if len(children) > 2 {
				field := tlc.NewStringNode(e.Field)
				bridge := tlcBridge{}
				bridge.withSyntaxNode(children[2], field)
				bridge.withPositionLocation(sanyNodePosition(children[2]), field)
				retainSanyGeneratedOperands(e, "$RcdSelect", []sanySemanticGraphNode{sanyGeneratedExpressionNode(e.Record), field})
			}
		}
	case *RecordSetExpr:
		fields := make([]sanyRecordGenerationField, len(e.Fields))
		for i, field := range e.Fields {
			fields[i] = sanyRecordGenerationField{field.Name, field.Set, field.Pos}
		}
		diags = append(diags, generation.checkRecordForm(e, "$SetOfRcds", fields, defined, locals)...)
	case *FunctionExpr:
		if e.IsLambda {
			e.lambdaNode, e.formalNodes = nil, nil
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullExpression)
			diagnostic := errorAt(e.Pos, "E4275", "LAMBDA expression used where an expression is required.")
			diagnostic.SANYMessage = diagnostic.Message
			if e.Syntax != nil {
				diagnostic.SANYRange = e.Syntax.Range
			}
			return Diagnostics{diagnostic}
		}
		var generated Diagnostics
		e.formalNodes, generated = generation.checkBoundExpression(e.Bounds, e.Syntax, defined, locals, e.Body)
		diags = append(diags, generated...)
		if !e.IsLambda {
			retainSanyBoundApplication(e, "$FcnConstructor", e.Bounds, e.formalNodes, e.Body)
		}
	case *FunctionAppExpr:
		diags = append(diags, generation.checkExpr(e.Function, defined, locals)...)
		if sanyExpressionGenerationFailure(e.Function) == sanyGenerationNullExpression {
			setSanyExpressionGenerationFailure(expr, sanyGenerationNullExpression)
			return diags
		}
		for _, arg := range e.Args {
			diags = append(diags, generation.checkExpr(arg, defined, locals)...)
		}
		retainSanyFunctionApplication(e)
	case *ExceptExpr:
		diags = append(diags, generation.generateExcept(e, defined, locals)...)
	case *LabelExpr:
		e.labelGenerated = false
		if generation.labelsEnabled && !generation.labelGoalUnsupported {
			diags = append(diags, generation.generateLabel(e, defined, locals)...)
			if label, ok := e.semanticGraph.(*sanySemLabelNode); ok && label.isAssumeProve && !allowLabeledAP {
				diagnostic := errorAt(e.Pos, "E4004", "Labeled ASSUME/PROVE used where an expression is required.")
				diagnostic.SANYMessage = diagnostic.Message
				if e.Syntax != nil {
					diagnostic.SANYRange = e.Syntax.Range
				}
				diags = append(diags, diagnostic)
			}
		} else {
			diags = append(diags, generation.checkExpr(e.Body, defined, locals)...)
			generation.resolveLabelFormals(e, defined)
		}
	case *ActionExpr:
		diags = append(diags, generation.checkExpr(e.Action, defined, locals)...)
		diags = append(diags, generation.checkExpr(e.Subscript, defined, locals)...)
		operator := "$SquareAct"
		if e.Kind == "angle" || e.Kind == "<>" || e.Kind == "NO_STUTTER" {
			operator = "$AngleAct"
		}
		retainSanyBuiltInApplication(e, operator, []Expr{e.Action, e.Subscript})
	case *FairnessExpr:
		diags = append(diags, generation.checkExpr(e.Subscript, defined, locals)...)
		diags = append(diags, generation.checkExpr(e.Action, defined, locals)...)
		operator := "$WF"
		if e.Kind == "SF" || e.Kind == "SF_" {
			operator = "$SF"
		}
		retainSanyBuiltInApplication(e, operator, []Expr{e.Subscript, e.Action})
	case *FunctionSetExpr:
		diags = append(diags, generation.checkExpr(e.Domain, defined, locals)...)
		diags = append(diags, generation.checkExpr(e.Range, defined, locals)...)
		retainSanyBuiltInApplication(e, "$SetOfFcns", []Expr{e.Domain, e.Range})
	case *SetComprehensionExpr:
		body := e.Element
		if e.Predicate != nil {
			// N_SubsetOf has only a predicate operand. The native Element
			// is derived syntax and is not separately generated by Java.
			body = e.Predicate
		}
		var generated Diagnostics
		e.formalNodes, generated = generation.checkBoundExpression(e.Bounds, e.Syntax, defined, locals, body)
		diags = append(diags, generated...)
		operator := "$SetOfAll"
		if e.Predicate != nil {
			operator = "$SubsetOf"
		}
		retainSanyBoundApplication(e, operator, e.Bounds, e.formalNodes, body)
	}
	return diags
}

// GenID appends each raw prefix identifier and the final raw operator token.
// Alias canonicalization belongs to symbol lookup, not the reported name.
func sanyOperatorGenIDName(operator *SanySyntaxNode) string {
	heirs := operator.GetHeirs()
	if len(heirs) != 2 {
		return ""
	}
	var name strings.Builder
	for _, element := range heirs[0].GetHeirs() {
		parts := element.GetHeirs()
		if len(parts) < 2 {
			continue
		}
		name.WriteString(sanyFirstTokenImage(parts[0]))
		name.WriteByte('!')
	}
	name.WriteString(sanyFirstTokenImage(heirs[1]))
	return name.String()
}

// Generator resolves the operator before generating any operand. Parser
// precedence metadata does not declare a symbol in the semantic context.
func checkSanyOperatorSymbolDefined(op string, pos Position, syntax *SanySyntaxNode, defined map[string]Position, locals map[string]bool) Diagnostics {
	name := op
	form := "infix"
	operatorRange := SanyRange{Begin: pos, End: pos.SourceEnd()}
	expressionRange := operatorRange
	if syntax != nil {
		expressionRange = syntax.Range
		heirs := syntax.GetHeirs()
		index := -1
		switch syntax.Kind.JavaName() {
		case "N_InfixExpr":
			index = 1
		case "N_PrefixExpr":
			form, index = "prefix", 0
		case "N_PostfixExpr":
			form, index = "postfix", 1
		}
		if index >= 0 && index < len(heirs) {
			operator := heirs[index]
			operatorRange = operator.Range
			if rawName := sanyOperatorGenIDName(operator); rawName != "" {
				name = rawName
				// GenID.finalAppend changes only the final raw unary '-'.
				if form == "prefix" && (name == "-" || strings.HasSuffix(name, "!-")) {
					name += "."
				}
			}
		}
	}
	resolvedName := ResolveSanyOperatorSynonym(name)
	if resolvedName == "" || localIdentifierInScope(locals, resolvedName) {
		return nil
	}
	if _, ok := sanyInitialBuiltinOperatorInfo(resolvedName); ok {
		return nil
	}
	if _, ok := defined[resolvedName]; ok {
		return nil
	}
	missing := errorAt(operatorRange.Begin, "E4004", "undefined operator %s", resolvedName)
	missing.SANYRange = operatorRange
	missing.SANYMessage = fmt.Sprintf("Could not find declaration or definition of symbol '%s'.", name)
	unresolved := errorAt(pos, "E4004", "could not resolve %s operator %s", form, resolvedName)
	unresolved.SANYRange = expressionRange
	unresolved.SANYMessage = fmt.Sprintf("Couldn't resolve %s operator symbol `%s'.", form, name)
	return Diagnostics{missing, unresolved}
}

func theoremStatementReferenceBase(name string) (string, bool) {
	base, ok := strings.CutSuffix(name, "!:")
	if !ok || base == "" {
		return "", false
	}
	return base, true
}

func subexpressionReferenceNameDefined(name string, defined map[string]Position) bool {
	if name == "" || !strings.Contains(name, "!") {
		return false
	}
	if base, ok := sanyBodySelectorBase(name); ok {
		_, exists := defined[base]
		return exists
	}
	parts := strings.Split(name, "!")
	if len(parts) < 2 {
		return false
	}
	for cut := len(parts) - 1; cut >= 1; cut-- {
		candidate := strings.Join(parts[:cut], "!")
		if candidate == "" {
			continue
		}
		if _, exists := defined[candidate]; !exists {
			continue
		}
		_, valid := sanyParseSubexpressionSelectors(parts[cut:])
		return valid
	}
	return false
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

func addSubexpressionReferenceNames(defined map[string]Position, base string, expr Expr) {
	if defined == nil || base == "" || expr == nil {
		return
	}
	colonName := base + "!:"
	if _, exists := defined[colonName]; !exists {
		defined[colonName] = expr.Position()
	}
	var walk func(string, Expr)
	walk = func(prefix string, current Expr) {
		for i, child := range sanySubexpressionChildren(current) {
			if child == nil {
				continue
			}
			name := prefix + "!" + strconv.Itoa(i+1)
			if _, exists := defined[name]; !exists {
				defined[name] = child.Position()
			}
			walk(name, child)
		}
	}
	walk(base, expr)
}

func checkCallArity(expr Expr, arities map[string]int, operatorParams map[string][]operatorParamSpec, locals map[string]bool) Diagnostics {
	if sanyExpressionGenerationFailure(expr) != sanyGenerationSucceeded {
		return nil
	}
	if source := sanyExprSource(expr); source != nil && source.operatorArgumentsGenerated {
		return nil
	}

	var diags Diagnostics
	if selected := sanyExprSelection(expr); selected != nil {
		for i, arg := range selected.args {
			if i < len(selected.params) && selected.params[i].OperatorArity > 0 {
				continue
			}
			diags = append(diags, checkCallArity(arg, arities, operatorParams, locals)...)
		}
		return diags
	}
	recur := func(expr Expr, arities map[string]int, locals map[string]bool) Diagnostics {
		return checkCallArity(expr, arities, operatorParams, locals)
	}
	switch e := expr.(type) {
	case *IdentExpr:
		_, formalArity := arities[localOperatorArityKey(e.Name)]
		if e.Name != "" && (!locals[e.Name] || e.generationArity != nil || formalArity) {
			if _, ok := builtinOperatorArity(e.Name); ok {
				return nil
			}
			want, ok := arities[e.Name]
			if locals[e.Name] {
				want, ok = arities[localOperatorArityKey(e.Name)]
			}
			if e.generationArity != nil {
				want, ok = *e.generationArity, true
			}
			if ok && want != 0 {
				diagnostic := sanyDiagnosticParameters(errorAt(e.Pos, "E4204", "operator %s arity mismatch: got 0 args, want %d", e.Name, want), e.Name, want)
				diagnostic.SANYMessage = fmt.Sprintf("The operator %s requires %d arguments.", e.Name, want)
				diags = append(diags, diagnostic)
			}
		}
	case *UnaryExpr:
		diags = append(diags, recur(e.Expr, arities, locals)...)
	case *BinaryExpr:
		diags = append(diags, recur(e.Left, arities, locals)...)
		diags = append(diags, recur(e.Right, arities, locals)...)
	case *CallExpr:
		var specs []operatorParamSpec
		if ident, isIdentifier := e.Callee.(*IdentExpr); isIdentifier {
			_, formalArity := arities[localOperatorArityKey(ident.Name)]
			if !locals[ident.Name] || ident.generationArity != nil || formalArity {
				want, ok := arities[ident.Name]
				if locals[ident.Name] {
					want, ok = arities[localOperatorArityKey(ident.Name)]
				}
				if ident.generationArity != nil {
					want, ok = *ident.generationArity, true
				}
				if !ok {
					want, ok = builtinOperatorArity(ident.Name)
				}
				if ok && want != len(e.Args) {
					diags = append(diags, sanyCallArityDiagnostic(e, ident.Name, want))
				}
				specs = operatorParams[ident.Name]
			}
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
		recursiveNames := letRecursiveNames(e)
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
			if !recursiveNames[def.Name] {
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
		parameters, body := sanyQuantifierGroup(e)
		quantLocals := copyBoolMap(locals)
		bodyArities := arities
		seenDomains := map[Expr]bool{}
		for _, parameter := range parameters {
			if parameter.Set != nil && !seenDomains[parameter.Set] {
				diags = append(diags, recur(parameter.Set, arities, locals)...)
				seenDomains[parameter.Set] = true
			}
			quantLocals[parameter.Var] = true
			bodyArities = quantifierBodyArities(bodyArities, parameter)
		}
		diags = append(diags, recur(body, bodyArities, quantLocals)...)
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
		for _, name := range e.boundNames() {
			chooseLocals[name] = true
		}
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
	if sanyExpressionGenerationFailure(expr) != sanyGenerationSucceeded {
		return nil
	}
	if source := sanyExprSource(expr); source != nil && source.operatorArgumentsGenerated {
		call, ok := expr.(*CallExpr)
		if !ok {
			return nil
		}
		identifier, ok := call.Callee.(*IdentExpr)
		if !ok {
			return nil
		}
		var diags Diagnostics
		specs := operatorParams[identifier.Name]
		for i, spec := range specs {
			if i >= len(call.Args) {
				break
			}
			if spec.Arity > 0 {
				got, known := operatorArgumentArity(call.Args[i], arities, locals)
				if known && got == spec.Arity {
					diags = append(diags, checkHigherOrderArgumentLevelConstraints(call.Args[i], spec, call.Args, specs, locals, identifier.Name, i)...)
				}
			}
		}
		return diags
	}

	var diags Diagnostics
	if selected := sanyExprSelection(expr); selected != nil {
		for i, arg := range selected.args {
			if i < len(selected.params) && selected.params[i].OperatorArity > 0 {
				want := selected.params[i].OperatorArity
				got, ok := operatorArgumentArity(arg, arities, locals)
				if !ok || got != want {
					diags = append(diags, errorAt(arg.Position(), "E4271", "operator argument arity mismatch: got %d, want %d", got, want))
				}
			}
			diags = append(diags, checkOperatorArgumentKinds(arg, operatorParams, arities, locals)...)
		}
		return diags
	}
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
							diags = append(diags, sanyDiagnosticParameters(errorAt(arg.Position(), "E4270", "operator parameter %s requires an operator argument of arity %d", spec.Name, spec.Arity), i+1, ident.Name))
							continue
						}
						if got != spec.Arity {
							code := "E4271"
							if fn, ok := arg.(*FunctionExpr); ok && fn.IsLambda {
								code = "E4274"
							}
							diags = append(diags, sanyDiagnosticParameters(errorAt(arg.Position(), code, "operator argument arity mismatch for parameter %s: got %d, want %d", spec.Name, got, spec.Arity), got, i+1, ident.Name, spec.Arity))
							continue
						}
						diags = append(diags, checkHigherOrderArgumentLevelConstraints(arg, spec, e.Args, specs, locals, ident.Name, i)...)
						if operatorArgumentRequiresOperatorParam(arg, operatorParams, locals) {
							diags = append(diags, errorAt(e.Pos, "E4271", "Argument number %d to operator '%s' should be a %d-parameter operator.", i+1, ident.Name, spec.Arity))
						}
					} else if fn, ok := arg.(*FunctionExpr); ok && fn.IsLambda && sanyExpressionGenerationFailure(arg) != sanyGenerationNullExpression {
						diags = append(diags, errorAt(arg.Position(), "E4275", "expression parameter %s cannot accept a LAMBDA operator argument", spec.Name))
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
		recursiveNames := letRecursiveNames(e)
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
			if !recursiveNames[def.Name] {
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
		parameters, body := sanyQuantifierGroup(e)
		quantLocals := copyBoolMap(locals)
		bodyArities := arities
		seenDomains := map[Expr]bool{}
		for _, parameter := range parameters {
			if parameter.Set != nil && !seenDomains[parameter.Set] {
				diags = append(diags, checkOperatorArgumentKinds(parameter.Set, operatorParams, arities, locals)...)
				seenDomains[parameter.Set] = true
			}
			quantLocals[parameter.Var] = true
			bodyArities = quantifierBodyArities(bodyArities, parameter)
		}
		diags = append(diags, checkOperatorArgumentKinds(body, operatorParams, bodyArities, quantLocals)...)
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
		for _, name := range e.boundNames() {
			chooseLocals[name] = true
		}
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

func operatorArgumentRequiresOperatorParam(arg Expr, operatorParams map[string][]operatorParamSpec, locals map[string]bool) bool {
	ident, ok := arg.(*IdentExpr)
	if !ok || localIdentifierInScope(locals, ident.Name) {
		return false
	}
	for _, spec := range operatorParams[ident.Name] {
		if spec.Arity >= 0 {
			return true
		}
	}
	return false
}

func checkHigherOrderArgumentLevelConstraints(arg Expr, spec operatorParamSpec, callArgs []Expr, specs []operatorParamSpec, locals map[string]bool, operatorName string, operatorArgIndex int) Diagnostics {
	var diags Diagnostics
	for i, minLevel := range spec.ArgMinLevels {
		if minLevel == constantLevel {
			continue
		}
		maxLevel, ok := operatorArgumentMaxLevel(arg, i)
		if !ok || maxLevel >= minLevel {
			continue
		}
		diags = append(diags, sanyDiagnosticParameters(errorAt(arg.Position(), "E4272", "operator argument %s cannot accept required level %d at argument %d", spec.Name, minLevel, i+1), operatorName, i+1, operatorArgIndex+1, minLevel))
	}
	for i, paramName := range spec.ArgParamNames {
		if paramName == "" {
			continue
		}
		maxLevel, ok := operatorArgumentMaxLevel(arg, i)
		if !ok {
			continue
		}
		callIndex := operatorParamSpecIndex(specs, paramName)
		if callIndex < 0 || callIndex >= len(callArgs) {
			continue
		}
		level := exprLevel(callArgs[callIndex], nil, locals)
		if level <= maxLevel {
			continue
		}
		diags = append(diags, sanyDiagnosticParameters(errorAt(callArgs[callIndex].Position(), "E4273", "operator argument %s co-parameter %s exceeds argument %d level constraint", spec.Name, paramName, i+1), operatorName, callIndex+1))
	}
	return diags
}

func operatorArgumentMaxLevel(arg Expr, index int) (tlaLevel, bool) {
	ident, ok := arg.(*IdentExpr)
	if !ok {
		return constantLevel, false
	}
	info, ok := sanyBuiltinOperatorInfo(ident.Name)
	if !ok {
		return constantLevel, false
	}
	return builtinArgMaxLevel(info, index)
}

func operatorParamSpecIndex(specs []operatorParamSpec, name string) int {
	for i, spec := range specs {
		if spec.Name == name {
			return i
		}
	}
	return -1
}

func operatorArgumentArity(expr Expr, arities map[string]int, locals map[string]bool) (int, bool) {
	if selected := sanyExprSelection(expr); selected != nil && selected.operator {
		return len(selected.params), true
	}
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
	info, ok := sanyInitialBuiltinOperatorInfo(name)
	if !ok {
		return 0, false
	}
	return info.arity, true
}

func checkFunctionArity(expr Expr, functionArities map[string]int, locals map[string]bool) Diagnostics {
	if sanyExpressionGenerationFailure(expr) != sanyGenerationSucceeded {
		return nil
	}
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
		for _, name := range e.boundNames() {
			chooseLocals[name] = true
		}
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
		if ident, ok := e.Function.(*IdentExpr); ok && !locals[ident.Name] && sanyExpressionGenerationFailure(ident) == sanyGenerationSucceeded {
			if want, exists := functionArities[ident.Name]; exists && !validFunctionApplicationArity(want, len(e.Args)) {
				diags = append(diags, sanyDiagnosticParameters(errorAt(e.Pos, "E4260", "function %s arity mismatch: got %d args, want %d", ident.Name, len(e.Args), want), ident.Name, want, len(e.Args)))
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

func copyDeclKindMap(in map[string]DeclarationKind) map[string]DeclarationKind {
	out := map[string]DeclarationKind{}
	for name, value := range in {
		out[name] = value
	}
	return out
}

func checkAssumptionConstantLevel(assumption NamedExpr, checker *sanyLevelCompositionChecker) Diagnostics {
	level := checker.level(assumption.Expr, nil)
	if assumption.AssumeProveBody != nil {
		level = checker.dependencies.assumeProveDependencies(assumption.AssumeProveBody, checker.context).level
	}
	if level == constantLevel {
		return nil
	}
	position := assumption.SourcePosition()
	diagnostic := sanyDiagnosticParameters(errorAt(position, "E4206", "assumption must be constant-level; expression has level %d", level), level)
	diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
	diagnostic.SANYMessage = fmt.Sprintf("Level error: assumptions must be level 0 (Constant), \n"+"but this one has level %d.", level)
	return Diagnostics{diagnostic}
}

func (levelChecker *sanyLevelCompositionChecker) check(expr Expr, locals map[string]bool) Diagnostics {
	_, diags := levelChecker.checkResult(expr, locals)
	return diags
}

// ExprNode.levelCorrect follows child check results. An Errors entry does not
// imply that its source node returned false (InstanceNode is one such case).
func (levelChecker *sanyLevelCompositionChecker) checkResult(expr Expr, locals map[string]bool) (bool, Diagnostics) {
	var diags Diagnostics
	correct := true
	child := func(checker *sanyLevelCompositionChecker, expression Expr, context map[string]bool) {
		childCorrect, childDiags := checker.checkResult(expression, context)
		correct = correct && childCorrect
		diags = append(diags, childDiags...)
	}
	constraints := func(current Diagnostics) {
		correct = correct && !current.HasErrors()
		diags = append(diags, current...)
	}
	switch expr.(type) {
	case *UnaryExpr, *ActionExpr, *FairnessExpr:
		// OpApplNode checks operands before applying these builtin maxima.
		// An invalid operand suppresses a redundant enclosing level error.
	default:
		constraints(levelChecker.checkApplicationLevels(expr, locals))
	}
	switch e := expr.(type) {
	case *LiteralExpr:
		if e.numeralNode != nil {
			correct = e.numeralNode.LevelCheck(1) && correct
		}
		if e.decimalNode != nil {
			correct = e.decimalNode.LevelCheck(1) && correct
		}
		if e.stringNode != nil {
			correct = e.stringNode.LevelCheck(1) && correct
		}
	case *UnaryExpr:
		operandCorrect, operandDiags := levelChecker.checkResult(e.Expr, locals)
		correct = correct && operandCorrect
		diags = append(diags, operandDiags...)
		constraints(levelChecker.checkApplicationLevelsWithValidity(e, locals, []bool{operandCorrect}))
		if (e.Op == "[]" || e.Op == "<>") && levelChecker.level(e.Expr, locals) == actionLevel && sanyOperatorApplicationKind(e.Expr) {
			if action, wrapped := e.Expr.(*ActionExpr); wrapped {
				if e.Op == "[]" && actionExprIsAngle(action) {
					correct = false
					diags = append(diags, sanyLevelDiagnostic(errorAt(e.Pos, "E4310", "temporal operator %s cannot be applied to an angle action", e.Op), e, "[] followed by action not of form [A]_v."))
				}
				if e.Op == "<>" && !actionExprIsAngle(action) {
					correct = false
					diags = append(diags, sanyLevelDiagnostic(errorAt(e.Pos, "E4311", "temporal operator %s cannot be applied to a square action", e.Op), e, "<> followed by action not of form <<A>>_v."))
				}
			} else {
				code := "E4310"
				if e.Op == "<>" {
					code = "E4311"
				}
				message := "[] followed by action not of form [A]_v."
				if e.Op == "<>" {
					message = "<> followed by action not of form <<A>>_v."
				}
				correct = false
				diags = append(diags, sanyLevelDiagnostic(errorAt(e.Pos, code, "temporal operator %s cannot be applied directly to an action-level formula", e.Op), e, message))
			}
		}
	case *BinaryExpr:
		leftLevel := levelChecker.level(e.Left, locals)
		rightLevel := levelChecker.level(e.Right, locals)
		if (e.Op == "~>" || e.Op == "-+->") && (leftLevel == actionLevel || rightLevel == actionLevel) {
			correct = false
			diags = append(diags, sanyLevelDiagnostic(errorAt(e.Pos, "E4312", "leads-to operator %s cannot have an action-level operand", e.Op), e, "Action used where only temporal formula or state predicate allowed."))
		}
		leftLogicalLevel := leftLevel
		rightLogicalLevel := rightLevel
		if isLogicalLevelMixingOperator(e.Op) && levelsMixActionAndTemporal(leftLogicalLevel, rightLogicalLevel) {
			correct = false
			diags = append(diags, sanyDiagnosticParameters(errorAt(e.Pos, "E4313", "operator %s cannot mix action and temporal operands", e.Op), sanyLogicalOperatorDiagnosticName(e)))
		}
		child(levelChecker, e.Left, locals)
		child(levelChecker, e.Right, locals)
	case *CallExpr:
		child(levelChecker, e.Callee, locals)
		for _, arg := range e.Args {
			child(levelChecker, arg, locals)
		}
	case *IfExpr:
		child(levelChecker, e.Cond, locals)
		child(levelChecker, e.Then, locals)
		child(levelChecker, e.Else, locals)
	case *LetExpr:
		levelChecker := levelChecker.withLet(e, locals)
		letLocals := letScopeLocals(locals, e)
		recursiveNames := letRecursiveNames(e)
		for _, def := range e.Definitions {
			if recursiveNames[def.Name] {
				constraints(levelChecker.checkRecursiveParameters(def, letLocals))
			}
			defLocals := letDefinitionBodyLocals(letLocals, def, recursiveNames[def.Name])
			child(levelChecker, def.Expr, defLocals)
		}
		child(levelChecker, e.Body, letLocals)
		// LetInNode checks its retained InstanceNodes after definitions/body,
		// even when no exported operator appears in IN.
		for _, instance := range e.Instances {
			instanceCorrect, instanceDiags := levelChecker.checkInstanceSubstitutionLevelResult(instance)
			correct = correct && instanceCorrect
			diags = append(diags, instanceDiags...)
		}
	case *QuantifierExpr:
		setLevel := levelChecker.level(e.Set, locals)
		bodyLevel := levelChecker.level(e.Body, withLocal(locals, e.Var))
		if setLevel == temporalLevel {
			correct = false
			diags = append(diags, sanyDiagnosticParameters(errorAt(e.Pos, "E4315", "quantifier cannot have a temporal-level bound"), sanyQuantifierOperatorName(e), e.Var))
		}
		if setLevel == actionLevel && bodyLevel == temporalLevel {
			correct = false
			diags = append(diags, sanyLevelDiagnostic(errorAt(e.Pos, "E4314", "quantifier with a temporal-level body cannot have an action-level bound"), e.Set, "Action-level bound of quantified temporal formula."))
		}
		child(levelChecker, e.Set, locals)
		child(levelChecker, e.Body, withLocal(locals, e.Var))
	case *CaseExpr:
		for _, arm := range e.Arms {
			child(levelChecker, arm.Test, locals)
			child(levelChecker, arm.Value, locals)
		}
		if e.Other != nil {
			child(levelChecker, e.Other, locals)
		}
	case *ChooseExpr:
		child(levelChecker, e.Set, locals)
		child(levelChecker, e.Body, withLocal(locals, e.boundNames()...))
	case *TupleExpr:
		for _, elem := range e.Elems {
			child(levelChecker, elem, locals)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			child(levelChecker, elem, locals)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			child(levelChecker, field.Value, locals)
		}
	case *RecordComponentExpr:
		if levelChecker.level(e.Record, locals) > actionLevel {
			correct = false
			diags = append(diags, errorAt(e.Pos, "E4205", "record selection cannot be applied to a temporal-level expression"))
		}
		child(levelChecker, e.Record, locals)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			child(levelChecker, field.Set, locals)
		}
	case *FunctionExpr:
		fnLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			child(levelChecker, bound.Set, locals)
			fnLocals[bound.Name] = true
		}
		child(levelChecker, e.Body, fnLocals)
	case *FunctionAppExpr:
		child(levelChecker, e.Function, locals)
		for _, arg := range e.Args {
			child(levelChecker, arg, locals)
		}
	case *ExceptExpr:
		child(levelChecker, e.Base, locals)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					child(levelChecker, index, locals)
				}
			}
			child(levelChecker, spec.Value, locals)
		}
	case *LabelExpr:
		child(levelChecker, e.Body, locals)
	case *ActionExpr:
		actionCorrect, actionDiags := levelChecker.checkResult(e.Action, locals)
		subscriptCorrect, subscriptDiags := levelChecker.checkResult(e.Subscript, locals)
		correct = correct && actionCorrect && subscriptCorrect
		diags = append(diags, actionDiags...)
		diags = append(diags, subscriptDiags...)
		constraints(levelChecker.checkApplicationLevelsWithValidity(e, locals, []bool{actionCorrect, subscriptCorrect}))
	case *FairnessExpr:
		subscriptCorrect, subscriptDiags := levelChecker.checkResult(e.Subscript, locals)
		actionCorrect, actionDiags := levelChecker.checkResult(e.Action, locals)
		diags = append(diags, subscriptDiags...)
		correct = correct && actionCorrect && subscriptCorrect
		diags = append(diags, actionDiags...)
		constraints(levelChecker.checkApplicationLevelsWithValidity(e, locals, []bool{subscriptCorrect, actionCorrect}))
	case *FunctionSetExpr:
		child(levelChecker, e.Domain, locals)
		child(levelChecker, e.Range, locals)
	case *SetComprehensionExpr:
		compLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			child(levelChecker, bound.Set, locals)
			compLocals[bound.Name] = true
		}
		child(levelChecker, e.Element, compLocals)
		if e.Predicate != nil {
			child(levelChecker, e.Predicate, compLocals)
		}
	}
	return correct, diags
}

func actionExprIsAngle(expr *ActionExpr) bool {
	if expr == nil {
		return false
	}
	return expr.Kind == "angle" || expr.Kind == "<>" || expr.Kind == "NO_STUTTER"
}

func isLogicalLevelMixingOperator(op string) bool {
	switch op {
	case "/\\", "\\/", "=>", "<=>", "\\equiv":
		return true
	default:
		return false
	}
}

func levelsMixActionAndTemporal(left, right tlaLevel) bool {
	return (left == actionLevel && right == temporalLevel) || (left == temporalLevel && right == actionLevel)
}

func withLocal(locals map[string]bool, names ...string) map[string]bool {
	out := copyBoolMap(locals)
	for _, name := range names {
		out[name] = true
	}
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
		for _, name := range e.boundNames() {
			chooseLocals[name] = true
		}
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

// Generator.selectorToNode reports the final argument-list location and the
// remaining signature after arguments attached to earlier name components.
func sanyCallArityDiagnostic(call *CallExpr, name string, want int) Diagnostic {
	diagnostic := errorAt(call.Pos, "E4204", "operator %s arity mismatch: got %d args, want %d", name, len(call.Args), want)
	remaining := want
	if selector := call.Selector; selector != nil && len(selector.Steps) > 0 {
		if selector.Syntax != nil {
			diagnostic.SANYRange = selector.Syntax.Range
		}
		for i, step := range selector.Steps {
			if i == len(selector.Steps)-1 {
				if step.Arguments != nil {
					diagnostic.SANYRange = step.Arguments.Range
				}
			} else if step.Arguments != nil {
				remaining -= len(expressionChildren(step.Arguments))
			}
		}
	}
	diagnostic.SANYMessage = fmt.Sprintf("The operator %s requires %d arguments.", name, remaining)
	diagnostic.SANYParameters = []any{name, remaining}
	return diagnostic
}

// Generator.selectorToNode extends curNameLoc across the unresolved compound
// name, excluding supplied arguments, before reporting SYMBOL_UNDEFINED.
func sanyUndefinedIdentifierDiagnostic(identifier *IdentExpr, origins ...*SanySelector) Diagnostic {
	diagnostic := errorAt(identifier.Pos, "E4200", "undefined identifier %s", identifier.Name)
	name := identifier.Name
	selector := identifier.Selector
	if selector == nil && len(origins) > 0 {
		selector = origins[0]
	}
	if selector != nil {
		var names []string
		for _, step := range selector.Steps {
			if step.Syntax == nil || step.Kind != SanySelectorName {
				continue
			}
			names = append(names, sanyCanonicalOperatorImage(step.Name))
			if diagnostic.SANYRange.Begin.Line == 0 {
				diagnostic.SANYRange = step.Syntax.Range
			} else {
				diagnostic.SANYRange.End = step.Syntax.Range.End
			}
		}
		if len(names) > 0 {
			name = strings.Join(names, "!")
		}
	} else if identifier.Syntax != nil {
		diagnostic.SANYRange = identifier.Syntax.Range
	}
	diagnostic.SANYMessage = fmt.Sprintf("Unknown operator: `%s'.", name)
	diagnostic.SANYParameters = []any{tlc.UniqueStringOf(name)}
	return diagnostic
}

// Generator.processRcdForms checks both record constructors and sets of records
// and reports the repeated field token, rather than the whole field/value pair.
func sanyDuplicateRecordFieldDiagnostic(name string, position, previous Position) Diagnostic {
	diagnostic := errorAt(position, "E4262", "duplicate record field %s; first field at %s", name, previous)
	diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
	diagnostic.SANYMessage = "Non-unique fields in constructor."
	return diagnostic
}

func sanyQuantifierOperatorName(expr *QuantifierExpr) string {
	if expr.Kind == "\\A" {
		return "$BoundedForall"
	}
	return "$BoundedExists"
}

func sanyLogicalOperatorDiagnosticName(expr *BinaryExpr) string {
	if expr.Syntax != nil {
		switch expr.Syntax.Kind.JavaName() {
		case "N_ConjList":
			return "Conjunction list"
		case "N_DisjList":
			return "Disjunction list"
		}
	}
	return sanyCanonicalOperatorImage(expr.Op)
}

func sanyInstanceLevelDiagnostic(instance Instance, name string, maximum tlaLevel, diagnostic Diagnostic) Diagnostic {
	position := instance.SourcePosition()
	diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
	diagnostic.SANYMessage = fmt.Sprintf("Level error in instantiating module '%s':\nThe level of the expression or operator substituted for '%s' \nmust be at most %d.", instance.Module, name, maximum)
	return diagnostic
}

func sanyIncompleteOperatorDiagnostic(identifier *IdentExpr) Diagnostic {
	diagnostic := sanyDiagnosticParameters(errorAt(identifier.Pos, "E4203", "operator name %s is incomplete", identifier.Name), identifier.Name)
	diagnostic.SANYMessage = fmt.Sprintf("Operator name %s is incomplete.", identifier.Name)
	if identifier.Syntax != nil {
		diagnostic.SANYRange = identifier.Syntax.Range
	}
	return diagnostic
}
