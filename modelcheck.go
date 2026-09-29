package tlago

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type State map[string]int

type ModelCheckOptions struct {
	MaxStates int
}

type ModelCheckResult struct {
	OK             bool
	StatesExplored int
	Error          string
	Trace          []State
}

func ModelCheck(spec *Spec, cfg ModelConfig, opts ModelCheckOptions) (ModelCheckResult, Diagnostics) {
	if opts.MaxStates == 0 {
		opts.MaxStates = 10000
	}
	if spec == nil || spec.Root == nil {
		return ModelCheckResult{}, Diagnostics{errorAt(Position{}, "E1500", "missing root module")}
	}
	mod := spec.Root
	defs := expandDefinitionsByName(applyModelValueConstants(definitionsByName(spec), cfg.ModelValues))
	initDef := defs[cfg.Init]
	nextDef := defs[cfg.Next]
	if (initDef == nil || nextDef == nil) && cfg.Specification != "" {
		specDef := defs[cfg.Specification]
		if specDef == nil {
			return ModelCheckResult{}, Diagnostics{errorAt(mod.Pos, "E1510", "SPECIFICATION operator %s not found", cfg.Specification)}
		}
		initExpr, nextExpr, ok := decomposeTemporalSpecification(specDef.Expr)
		if !ok {
			return ModelCheckResult{}, Diagnostics{errorAt(specDef.Pos, "E1511", "SPECIFICATION %s must have form Init /\\ [][Next]_vars", cfg.Specification)}
		}
		initDef = &Definition{Name: cfg.Specification + "!Init", Expr: initExpr, Pos: specDef.Pos}
		nextDef = &Definition{Name: cfg.Specification + "!Next", Expr: nextExpr, Pos: specDef.Pos}
	}
	if initDef == nil {
		return ModelCheckResult{}, Diagnostics{errorAt(mod.Pos, "E1501", "INIT operator %s not found", cfg.Init)}
	}
	if nextDef == nil {
		return ModelCheckResult{}, Diagnostics{errorAt(mod.Pos, "E1502", "NEXT operator %s not found", cfg.Next)}
	}
	constants := constantEnvironment(spec, cfg.Constants)
	check, diag := checkAssumptions(spec, defs, constants, cfg.ModelValues)
	if diag != nil {
		return ModelCheckResult{}, Diagnostics{*diag}
	}
	if check != "" {
		return ModelCheckResult{
			OK:    false,
			Error: check,
		}, nil
	}
	vars := moduleVariables(mod)
	keyOf := func(st State) (string, error) {
		return stateKeyWithView(st, vars, defs, cfg.Views, constants)
	}
	initStates, err := initialStates(initDef.Expr, vars, constants)
	if err != nil {
		return ModelCheckResult{}, Diagnostics{errorAt(initDef.Pos, "E1503", "%v", err)}
	}
	var queue []State
	seen := map[string]bool{}
	parent := map[string]string{}
	stateByKey := map[string]State{}
	for _, st := range initStates {
		ok, err := stateSatisfiesAll(defs, cfg.Constraints, st, constants)
		if err != nil {
			return ModelCheckResult{}, Diagnostics{errorAt(initDef.Pos, "E1507", "%v", err)}
		}
		if !ok {
			continue
		}
		key, err := keyOf(st)
		if err != nil {
			return ModelCheckResult{}, Diagnostics{errorAt(initDef.Pos, "E1512", "%v", err)}
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		stateByKey[key] = cloneState(st)
		queue = append(queue, st)
	}
	explored := 0
	for len(queue) > 0 {
		if explored >= opts.MaxStates {
			return ModelCheckResult{
				OK:             false,
				StatesExplored: explored,
				Error:          fmt.Sprintf("state limit %d reached", opts.MaxStates),
			}, nil
		}
		st := queue[0]
		queue = queue[1:]
		explored++
		check, diag := checkStatePredicates(defs, cfg.Invariants, "invariant", mod.Pos, st, constants)
		if diag != nil {
			return ModelCheckResult{}, Diagnostics{*diag}
		}
		if check != "" {
			key, err := keyOf(st)
			if err != nil {
				return ModelCheckResult{}, Diagnostics{errorAt(mod.Pos, "E1512", "%v", err)}
			}
			return ModelCheckResult{
				OK:             false,
				StatesExplored: explored,
				Error:          check,
				Trace:          reconstructTrace(key, parent, stateByKey),
			}, nil
		}
		check, diag = checkStatePredicates(defs, cfg.Properties, "property", mod.Pos, st, constants)
		if diag != nil {
			return ModelCheckResult{}, Diagnostics{*diag}
		}
		if check != "" {
			key, err := keyOf(st)
			if err != nil {
				return ModelCheckResult{}, Diagnostics{errorAt(mod.Pos, "E1512", "%v", err)}
			}
			return ModelCheckResult{
				OK:             false,
				StatesExplored: explored,
				Error:          check,
				Trace:          reconstructTrace(key, parent, stateByKey),
			}, nil
		}
		nextStates, err := nextStates(nextDef.Expr, st, vars, constants)
		if err != nil {
			return ModelCheckResult{}, Diagnostics{errorAt(nextDef.Pos, "E1506", "%v", err)}
		}
		constrainedNextStates := nextStates[:0]
		for _, ns := range nextStates {
			ok, err := transitionSatisfiesAll(defs, cfg.ActionConstraints, st, ns, constants)
			if err != nil {
				return ModelCheckResult{}, Diagnostics{errorAt(nextDef.Pos, "E1508", "%v", err)}
			}
			if ok {
				constrainedNextStates = append(constrainedNextStates, ns)
			}
		}
		if len(constrainedNextStates) == 0 && cfg.CheckDeadlock {
			key, err := keyOf(st)
			if err != nil {
				return ModelCheckResult{}, Diagnostics{errorAt(nextDef.Pos, "E1512", "%v", err)}
			}
			return ModelCheckResult{
				OK:             false,
				StatesExplored: explored,
				Error:          "deadlock reached",
				Trace:          reconstructTrace(key, parent, stateByKey),
			}, nil
		}
		fromKey, err := keyOf(st)
		if err != nil {
			return ModelCheckResult{}, Diagnostics{errorAt(nextDef.Pos, "E1512", "%v", err)}
		}
		for _, ns := range constrainedNextStates {
			ok, err := stateSatisfiesAll(defs, cfg.Constraints, ns, constants)
			if err != nil {
				return ModelCheckResult{}, Diagnostics{errorAt(nextDef.Pos, "E1507", "%v", err)}
			}
			if !ok {
				continue
			}
			key, err := keyOf(ns)
			if err != nil {
				return ModelCheckResult{}, Diagnostics{errorAt(nextDef.Pos, "E1512", "%v", err)}
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			parent[key] = fromKey
			stateByKey[key] = cloneState(ns)
			queue = append(queue, ns)
		}
	}
	check, diag = checkPostconditions(defs, cfg.Postconditions, mod.Pos, constants)
	if diag != nil {
		return ModelCheckResult{}, Diagnostics{*diag}
	}
	if check != "" {
		return ModelCheckResult{
			OK:             false,
			StatesExplored: explored,
			Error:          check,
		}, nil
	}
	return ModelCheckResult{OK: true, StatesExplored: explored}, nil
}

func decomposeTemporalSpecification(expr Expr) (Expr, Expr, bool) {
	switch e := expr.(type) {
	case *LabelExpr:
		return decomposeTemporalSpecification(e.Body)
	case *BinaryExpr:
		if e.Op != "/\\" {
			return nil, nil, false
		}
		if next, ok := temporalNextRelation(e.Right); ok {
			return e.Left, next, true
		}
		if next, ok := temporalNextRelation(e.Left); ok {
			return e.Right, next, true
		}
	}
	return nil, nil, false
}

func temporalNextRelation(expr Expr) (Expr, bool) {
	switch e := expr.(type) {
	case *LabelExpr:
		return temporalNextRelation(e.Body)
	case *UnaryExpr:
		if e.Op == "[]" {
			return temporalActionOperand(e.Expr)
		}
	}
	return nil, false
}

func temporalActionOperand(expr Expr) (Expr, bool) {
	switch e := expr.(type) {
	case *LabelExpr:
		return temporalActionOperand(e.Body)
	case *ActionExpr:
		return e, true
	}
	return nil, false
}

func constantEnvironment(spec *Spec, cfgConstants State) State {
	env := cloneState(cfgConstants)
	if spec == nil {
		return env
	}
	moduleNames := make([]string, 0, len(spec.Modules))
	for name := range spec.Modules {
		moduleNames = append(moduleNames, name)
	}
	sort.Strings(moduleNames)
	rootExtends := map[string]bool{}
	if spec.Root != nil {
		for _, name := range moduleImports(spec.Root) {
			rootExtends[name] = true
		}
	}
	for pass := 0; pass < 8; pass++ {
		changed := false
		for _, moduleName := range moduleNames {
			mod := spec.Modules[moduleName]
			if mod == nil {
				continue
			}
			for _, def := range mod.Definitions {
				if len(def.Params) != 0 || def.Expr == nil {
					continue
				}
				value, err := evalInt(def.Expr, env)
				if err != nil {
					continue
				}
				names := []string{moduleName + "!" + def.Name}
				if spec.Root != nil && (moduleName == spec.Root.Name || rootExtends[moduleName]) {
					names = append(names, def.Name)
				}
				for _, name := range names {
					if _, exists := cfgConstants[name]; exists {
						continue
					}
					if old, exists := env[name]; !exists || old != value {
						env[name] = value
						changed = true
					}
				}
			}
		}
		if !changed {
			break
		}
	}
	for name, value := range cfgConstants {
		env[name] = value
	}
	return env
}

func checkAssumptions(spec *Spec, defs map[string]*Definition, constants State, modelValues map[string]string) (string, *Diagnostic) {
	if spec == nil {
		return "", nil
	}
	replacements := modelValueReplacements(modelValues)
	moduleNames := make([]string, 0, len(spec.Modules))
	for name := range spec.Modules {
		moduleNames = append(moduleNames, name)
	}
	sort.Strings(moduleNames)
	for _, moduleName := range moduleNames {
		mod := spec.Modules[moduleName]
		if mod == nil {
			continue
		}
		for _, assumption := range mod.Assumptions {
			if assumption.Expr == nil {
				continue
			}
			expr := substituteParams(assumption.Expr, replacements)
			expr = inlineDefinitionCalls(expr, defs, map[string]bool{})
			ok, err := evalBool(expr, constants)
			if err != nil {
				diag := errorAt(assumption.Position(), "E1516", "%v", err)
				return "", &diag
			}
			if !ok {
				return fmt.Sprintf("assumption in module %s is violated", mod.Name), nil
			}
		}
	}
	return "", nil
}

func applyModelValueConstants(defs map[string]*Definition, modelValues map[string]string) map[string]*Definition {
	if len(modelValues) == 0 {
		return defs
	}
	replacements := modelValueReplacements(modelValues)
	out := map[string]*Definition{}
	for name, def := range defs {
		if def == nil {
			continue
		}
		next := *def
		defReplacements := map[string]Expr{}
		for replacementName, replacement := range replacements {
			defReplacements[replacementName] = replacement
		}
		for _, param := range def.Params {
			delete(defReplacements, param)
		}
		next.Expr = substituteParams(def.Expr, defReplacements)
		out[name] = &next
	}
	return out
}

func modelValueReplacements(modelValues map[string]string) map[string]Expr {
	replacements := map[string]Expr{}
	for name, value := range modelValues {
		replacements[name] = &LiteralExpr{Kind: "model", Value: value}
	}
	return replacements
}

func checkStatePredicates(defs map[string]*Definition, names []string, label string, pos Position, st State, constants State) (string, *Diagnostic) {
	for _, name := range names {
		def := defs[name]
		if def == nil {
			diag := errorAt(pos, "E1504", "%s operator %s not found", label, name)
			return "", &diag
		}
		ok, err := evalBool(def.Expr, mergeState(constants, st))
		if err != nil {
			diag := errorAt(def.Pos, "E1505", "%v", err)
			return "", &diag
		}
		if !ok {
			return fmt.Sprintf("%s %s is violated", label, name), nil
		}
	}
	return "", nil
}

func checkPostconditions(defs map[string]*Definition, names []string, pos Position, constants State) (string, *Diagnostic) {
	env := cloneState(constants)
	for _, name := range names {
		def := defs[name]
		if def == nil {
			diag := errorAt(pos, "E1513", "postcondition operator %s not found", name)
			return "", &diag
		}
		if len(def.Params) != 0 {
			diag := errorAt(def.Pos, "E1514", "postcondition operator %s must take no arguments", name)
			return "", &diag
		}
		ok, err := evalBool(def.Expr, env)
		if err != nil {
			diag := errorAt(def.Pos, "E1515", "%v", err)
			return "", &diag
		}
		if !ok {
			return fmt.Sprintf("postcondition %s is violated", name), nil
		}
	}
	return "", nil
}

func stateSatisfiesAll(defs map[string]*Definition, names []string, st State, constants State) (bool, error) {
	env := mergeState(constants, st)
	for _, name := range names {
		def := defs[name]
		if def == nil {
			return false, fmt.Errorf("constraint operator %s not found", name)
		}
		ok, err := evalBool(def.Expr, env)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func transitionSatisfiesAll(defs map[string]*Definition, names []string, current State, next State, constants State) (bool, error) {
	env := mergeState(constants, current)
	for name, val := range next {
		env[name+"'"] = val
	}
	for _, name := range names {
		def := defs[name]
		if def == nil {
			return false, fmt.Errorf("action constraint operator %s not found", name)
		}
		ok, err := evalBool(def.Expr, env)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func definitionsByName(spec *Spec) map[string]*Definition {
	defs := map[string]*Definition{}
	if spec == nil || spec.Root == nil {
		return defs
	}
	rootName := spec.Root.Name
	rootExtends := map[string]bool{}
	for _, name := range spec.Root.Extends {
		rootExtends[name] = true
	}
	moduleNames := make([]string, 0, len(spec.Modules))
	for name := range spec.Modules {
		moduleNames = append(moduleNames, name)
	}
	sort.Strings(moduleNames)
	for _, moduleName := range moduleNames {
		mod := spec.Modules[moduleName]
		if mod == nil {
			continue
		}
		for i := range mod.Definitions {
			def := &mod.Definitions[i]
			if moduleName != "" {
				defs[moduleName+"!"+def.Name] = def
			}
			if moduleName == rootName {
				defs[def.Name] = def
				continue
			}
			if rootExtends[moduleName] {
				if _, exists := defs[def.Name]; !exists {
					defs[def.Name] = def
				}
			}
		}
		if rootExtends[moduleName] {
			for _, inst := range mod.Instances {
				if inst.Local {
					continue
				}
				addInstanceDefinitions(defs, spec, inst)
			}
		}
	}
	for _, inst := range spec.Root.Instances {
		addInstanceDefinitions(defs, spec, inst)
	}
	return defs
}

func addInstanceDefinitions(defs map[string]*Definition, spec *Spec, inst Instance) {
	if spec == nil {
		return
	}
	mod := spec.Modules[inst.Module]
	if mod == nil {
		return
	}
	for i := range mod.Definitions {
		if mod.Definitions[i].Local {
			continue
		}
		def := instantiatedDefinition(&mod.Definitions[i], inst.Substitutions)
		if def == nil {
			continue
		}
		qualifier := inst.qualifier()
		if qualifier != "" {
			defs[qualifier+"!"+def.Name] = def
		}
		if inst.exportsUnqualified() {
			if _, exists := defs[def.Name]; exists {
				continue
			}
			defs[def.Name] = def
		}
	}
}

func instantiatedDefinition(def *Definition, substitutions map[string]Expr) *Definition {
	if def == nil {
		return nil
	}
	next := *def
	if len(substitutions) > 0 {
		next.Expr = substituteParams(def.Expr, substitutions)
	}
	return &next
}

func expandDefinitionsByName(defs map[string]*Definition) map[string]*Definition {
	expanded := map[string]*Definition{}
	for name, def := range defs {
		if def == nil {
			continue
		}
		next := *def
		next.Expr = inlineDefinitionCalls(def.Expr, defs, map[string]bool{name: true})
		expanded[name] = &next
	}
	return expanded
}

func inlineDefinitionCalls(expr Expr, defs map[string]*Definition, seen map[string]bool) Expr {
	switch e := expr.(type) {
	case *IdentExpr:
		def := defs[e.Name]
		if def == nil || len(def.Params) != 0 || seen[e.Name] {
			return e
		}
		nextSeen := copySeen(seen)
		nextSeen[e.Name] = true
		return inlineDefinitionCalls(def.Expr, defs, nextSeen)
	case *UnaryExpr:
		operand := inlineDefinitionCalls(e.Expr, defs, copySeen(seen))
		if def := defs[e.Op]; def != nil && len(def.Params) == 1 && !seen[e.Op] {
			replacements := map[string]Expr{def.Params[0]: operand}
			nextSeen := copySeen(seen)
			nextSeen[e.Op] = true
			return inlineDefinitionCalls(substituteParams(def.Expr, replacements), defs, nextSeen)
		}
		return &UnaryExpr{Op: e.Op, Expr: operand, Pos: e.Pos}
	case *BinaryExpr:
		left := inlineDefinitionCalls(e.Left, defs, copySeen(seen))
		right := inlineDefinitionCalls(e.Right, defs, copySeen(seen))
		if def := defs[e.Op]; def != nil && len(def.Params) == 2 && !seen[e.Op] {
			replacements := map[string]Expr{
				def.Params[0]: left,
				def.Params[1]: right,
			}
			nextSeen := copySeen(seen)
			nextSeen[e.Op] = true
			return inlineDefinitionCalls(substituteParams(def.Expr, replacements), defs, nextSeen)
		}
		return &BinaryExpr{
			Op:           e.Op,
			Left:         left,
			Right:        right,
			Pos:          e.Pos,
			JunctionList: e.JunctionList,
		}
	case *CallExpr:
		call := &CallExpr{Callee: inlineDefinitionCalls(e.Callee, defs, copySeen(seen)), Pos: e.Pos}
		for _, arg := range e.Args {
			call.Args = append(call.Args, inlineDefinitionCalls(arg, defs, copySeen(seen)))
		}
		callee, ok := e.Callee.(*IdentExpr)
		if !ok {
			return call
		}
		def := defs[callee.Name]
		if def == nil || len(def.Params) != len(e.Args) || seen[callee.Name] {
			return call
		}
		replacements := map[string]Expr{}
		for i, param := range def.Params {
			replacements[param] = call.Args[i]
		}
		nextSeen := copySeen(seen)
		nextSeen[callee.Name] = true
		return inlineDefinitionCalls(substituteParams(def.Expr, replacements), defs, nextSeen)
	case *IfExpr:
		return &IfExpr{
			Cond: inlineDefinitionCalls(e.Cond, defs, copySeen(seen)),
			Then: inlineDefinitionCalls(e.Then, defs, copySeen(seen)),
			Else: inlineDefinitionCalls(e.Else, defs, copySeen(seen)),
			Pos:  e.Pos,
		}
	case *LetExpr:
		let := &LetExpr{Recursives: append([]Declaration(nil), e.Recursives...), Instances: append([]Instance(nil), e.Instances...), Pos: e.Pos}
		for _, def := range e.Definitions {
			nextDef := def
			nextDef.Expr = inlineDefinitionCalls(def.Expr, defs, copySeen(seen))
			let.Definitions = append(let.Definitions, nextDef)
		}
		let.Body = inlineDefinitionCalls(e.Body, defs, copySeen(seen))
		return let
	case *QuantifierExpr:
		return &QuantifierExpr{
			Kind:       e.Kind,
			Var:        e.Var,
			VarPos:     e.VarPos,
			Set:        inlineDefinitionCalls(e.Set, defs, copySeen(seen)),
			Body:       inlineDefinitionCalls(e.Body, defs, copySeen(seen)),
			TupleBound: e.TupleBound,
			Pos:        e.Pos,
		}
	case *CaseExpr:
		caseExpr := &CaseExpr{Pos: e.Pos, OtherPos: e.OtherPos}
		for _, arm := range e.Arms {
			caseExpr.Arms = append(caseExpr.Arms, CaseArm{
				Test:  inlineDefinitionCalls(arm.Test, defs, copySeen(seen)),
				Value: inlineDefinitionCalls(arm.Value, defs, copySeen(seen)),
				Pos:   arm.Pos,
			})
		}
		if e.Other != nil {
			caseExpr.Other = inlineDefinitionCalls(e.Other, defs, copySeen(seen))
		}
		return caseExpr
	case *ChooseExpr:
		return &ChooseExpr{
			Var:    e.Var,
			VarPos: e.VarPos,
			Set:    inlineDefinitionCalls(e.Set, defs, copySeen(seen)),
			Body:   inlineDefinitionCalls(e.Body, defs, copySeen(seen)),
			Pos:    e.Pos,
		}
	case *TupleExpr:
		tuple := &TupleExpr{Pos: e.Pos}
		for _, elem := range e.Elems {
			tuple.Elems = append(tuple.Elems, inlineDefinitionCalls(elem, defs, copySeen(seen)))
		}
		return tuple
	case *SetExpr:
		set := &SetExpr{Pos: e.Pos}
		for _, elem := range e.Elems {
			set.Elems = append(set.Elems, inlineDefinitionCalls(elem, defs, copySeen(seen)))
		}
		return set
	case *RecordExpr:
		record := &RecordExpr{Pos: e.Pos}
		for _, field := range e.Fields {
			record.Fields = append(record.Fields, RecordField{
				Name:   field.Name,
				Value:  inlineDefinitionCalls(field.Value, defs, copySeen(seen)),
				Pos:    field.Pos,
				Source: field.Source,
			})
		}
		return record
	case *RecordComponentExpr:
		return &RecordComponentExpr{Record: inlineDefinitionCalls(e.Record, defs, copySeen(seen)), Field: e.Field, FieldPos: e.FieldPos, Pos: e.Pos}
	case *RecordSetExpr:
		record := &RecordSetExpr{Pos: e.Pos}
		for _, field := range e.Fields {
			record.Fields = append(record.Fields, RecordSetField{
				Name:   field.Name,
				Set:    inlineDefinitionCalls(field.Set, defs, copySeen(seen)),
				Pos:    field.Pos,
				Source: field.Source,
			})
		}
		return record
	case *FunctionExpr:
		fn := &FunctionExpr{Pos: e.Pos}
		for _, bound := range e.Bounds {
			fn.Bounds = append(fn.Bounds, BoundVar{
				Name: bound.Name,
				Set:  inlineDefinitionCalls(bound.Set, defs, copySeen(seen)),
				Pos:  bound.Pos,
			})
		}
		fn.Body = inlineDefinitionCalls(e.Body, defs, copySeen(seen))
		return fn
	case *FunctionAppExpr:
		app := &FunctionAppExpr{Function: inlineDefinitionCalls(e.Function, defs, copySeen(seen)), Pos: e.Pos}
		for _, arg := range e.Args {
			app.Args = append(app.Args, inlineDefinitionCalls(arg, defs, copySeen(seen)))
		}
		return app
	case *ExceptExpr:
		except := &ExceptExpr{Base: inlineDefinitionCalls(e.Base, defs, copySeen(seen)), Pos: e.Pos}
		for _, spec := range e.Specs {
			nextSpec := ExceptSpec{Value: inlineDefinitionCalls(spec.Value, defs, copySeen(seen)), Pos: spec.Pos}
			for _, component := range spec.Components {
				nextComponent := ExceptComponent{Field: component.Field, FieldPos: component.FieldPos, Pos: component.Pos}
				for _, index := range component.Indices {
					nextComponent.Indices = append(nextComponent.Indices, inlineDefinitionCalls(index, defs, copySeen(seen)))
				}
				nextSpec.Components = append(nextSpec.Components, nextComponent)
			}
			except.Specs = append(except.Specs, nextSpec)
		}
		return except
	case *LabelExpr:
		return &LabelExpr{Name: e.Name, Body: inlineDefinitionCalls(e.Body, defs, copySeen(seen)), Pos: e.Pos}
	case *ActionExpr:
		return &ActionExpr{
			Kind:      e.Kind,
			Action:    inlineDefinitionCalls(e.Action, defs, copySeen(seen)),
			Subscript: inlineDefinitionCalls(e.Subscript, defs, copySeen(seen)),
			Pos:       e.Pos,
		}
	case *FairnessExpr:
		return &FairnessExpr{
			Kind:      e.Kind,
			Subscript: inlineDefinitionCalls(e.Subscript, defs, copySeen(seen)),
			Action:    inlineDefinitionCalls(e.Action, defs, copySeen(seen)),
			Pos:       e.Pos,
		}
	case *FunctionSetExpr:
		return &FunctionSetExpr{
			Domain: inlineDefinitionCalls(e.Domain, defs, copySeen(seen)),
			Range:  inlineDefinitionCalls(e.Range, defs, copySeen(seen)),
			Pos:    e.Pos,
		}
	case *SetComprehensionExpr:
		comp := &SetComprehensionExpr{Pos: e.Pos}
		for _, bound := range e.Bounds {
			comp.Bounds = append(comp.Bounds, BoundVar{
				Name: bound.Name,
				Set:  inlineDefinitionCalls(bound.Set, defs, copySeen(seen)),
				Pos:  bound.Pos,
			})
		}
		comp.Element = inlineDefinitionCalls(e.Element, defs, copySeen(seen))
		if e.Predicate != nil {
			comp.Predicate = inlineDefinitionCalls(e.Predicate, defs, copySeen(seen))
		}
		return comp
	default:
		return expr
	}
}

func substituteParams(expr Expr, replacements map[string]Expr) Expr {
	switch e := expr.(type) {
	case *IdentExpr:
		if replacement, ok := replacements[e.Name]; ok {
			return replacement
		}
		return e
	case *UnaryExpr:
		return &UnaryExpr{Op: e.Op, Expr: substituteParams(e.Expr, replacements), Pos: e.Pos}
	case *BinaryExpr:
		return &BinaryExpr{Op: e.Op, Left: substituteParams(e.Left, replacements), Right: substituteParams(e.Right, replacements), Pos: e.Pos, JunctionList: e.JunctionList}
	case *CallExpr:
		call := &CallExpr{Callee: substituteParams(e.Callee, replacements), Pos: e.Pos}
		for _, arg := range e.Args {
			call.Args = append(call.Args, substituteParams(arg, replacements))
		}
		return call
	case *IfExpr:
		return &IfExpr{Cond: substituteParams(e.Cond, replacements), Then: substituteParams(e.Then, replacements), Else: substituteParams(e.Else, replacements), Pos: e.Pos}
	case *LetExpr:
		letReplacements := map[string]Expr{}
		for name, replacement := range replacements {
			letReplacements[name] = replacement
		}
		let := &LetExpr{Recursives: append([]Declaration(nil), e.Recursives...), Instances: append([]Instance(nil), e.Instances...), Pos: e.Pos}
		for _, def := range e.Definitions {
			delete(letReplacements, def.Name)
		}
		for _, def := range e.Definitions {
			nextDef := def
			defReplacements := map[string]Expr{}
			for name, replacement := range letReplacements {
				defReplacements[name] = replacement
			}
			for _, param := range def.Params {
				delete(defReplacements, param)
			}
			nextDef.Expr = substituteParams(def.Expr, defReplacements)
			let.Definitions = append(let.Definitions, nextDef)
		}
		let.Body = substituteParams(e.Body, letReplacements)
		return let
	case *QuantifierExpr:
		quantReplacements := map[string]Expr{}
		for name, replacement := range replacements {
			if name != e.Var {
				quantReplacements[name] = replacement
			}
		}
		return &QuantifierExpr{Kind: e.Kind, Var: e.Var, VarPos: e.VarPos, Set: substituteParams(e.Set, replacements), Body: substituteParams(e.Body, quantReplacements), TupleBound: e.TupleBound, Pos: e.Pos}
	case *CaseExpr:
		caseExpr := &CaseExpr{Pos: e.Pos, OtherPos: e.OtherPos}
		for _, arm := range e.Arms {
			caseExpr.Arms = append(caseExpr.Arms, CaseArm{
				Test:  substituteParams(arm.Test, replacements),
				Value: substituteParams(arm.Value, replacements),
				Pos:   arm.Pos,
			})
		}
		if e.Other != nil {
			caseExpr.Other = substituteParams(e.Other, replacements)
		}
		return caseExpr
	case *ChooseExpr:
		chooseReplacements := map[string]Expr{}
		for name, replacement := range replacements {
			if name != e.Var {
				chooseReplacements[name] = replacement
			}
		}
		return &ChooseExpr{Var: e.Var, VarPos: e.VarPos, Set: substituteParams(e.Set, replacements), Body: substituteParams(e.Body, chooseReplacements), Pos: e.Pos}
	case *TupleExpr:
		tuple := &TupleExpr{Pos: e.Pos}
		for _, elem := range e.Elems {
			tuple.Elems = append(tuple.Elems, substituteParams(elem, replacements))
		}
		return tuple
	case *SetExpr:
		set := &SetExpr{Pos: e.Pos}
		for _, elem := range e.Elems {
			set.Elems = append(set.Elems, substituteParams(elem, replacements))
		}
		return set
	case *RecordExpr:
		record := &RecordExpr{Pos: e.Pos}
		for _, field := range e.Fields {
			record.Fields = append(record.Fields, RecordField{
				Name:   field.Name,
				Value:  substituteParams(field.Value, replacements),
				Pos:    field.Pos,
				Source: field.Source,
			})
		}
		return record
	case *RecordComponentExpr:
		return &RecordComponentExpr{Record: substituteParams(e.Record, replacements), Field: e.Field, FieldPos: e.FieldPos, Pos: e.Pos}
	case *RecordSetExpr:
		record := &RecordSetExpr{Pos: e.Pos}
		for _, field := range e.Fields {
			record.Fields = append(record.Fields, RecordSetField{
				Name:   field.Name,
				Set:    substituteParams(field.Set, replacements),
				Pos:    field.Pos,
				Source: field.Source,
			})
		}
		return record
	case *FunctionExpr:
		fnReplacements := map[string]Expr{}
		for name, replacement := range replacements {
			fnReplacements[name] = replacement
		}
		fn := &FunctionExpr{Pos: e.Pos}
		for _, bound := range e.Bounds {
			delete(fnReplacements, bound.Name)
			fn.Bounds = append(fn.Bounds, BoundVar{
				Name: bound.Name,
				Set:  substituteParams(bound.Set, replacements),
				Pos:  bound.Pos,
			})
		}
		fn.Body = substituteParams(e.Body, fnReplacements)
		return fn
	case *FunctionAppExpr:
		app := &FunctionAppExpr{Function: substituteParams(e.Function, replacements), Pos: e.Pos}
		for _, arg := range e.Args {
			app.Args = append(app.Args, substituteParams(arg, replacements))
		}
		return app
	case *ExceptExpr:
		except := &ExceptExpr{Base: substituteParams(e.Base, replacements), Pos: e.Pos}
		for _, spec := range e.Specs {
			nextSpec := ExceptSpec{Value: substituteParams(spec.Value, replacements), Pos: spec.Pos}
			for _, component := range spec.Components {
				nextComponent := ExceptComponent{Field: component.Field, FieldPos: component.FieldPos, Pos: component.Pos}
				for _, index := range component.Indices {
					nextComponent.Indices = append(nextComponent.Indices, substituteParams(index, replacements))
				}
				nextSpec.Components = append(nextSpec.Components, nextComponent)
			}
			except.Specs = append(except.Specs, nextSpec)
		}
		return except
	case *LabelExpr:
		return &LabelExpr{Name: e.Name, Body: substituteParams(e.Body, replacements), Pos: e.Pos}
	case *ActionExpr:
		return &ActionExpr{
			Kind:      e.Kind,
			Action:    substituteParams(e.Action, replacements),
			Subscript: substituteParams(e.Subscript, replacements),
			Pos:       e.Pos,
		}
	case *FairnessExpr:
		return &FairnessExpr{
			Kind:      e.Kind,
			Subscript: substituteParams(e.Subscript, replacements),
			Action:    substituteParams(e.Action, replacements),
			Pos:       e.Pos,
		}
	case *FunctionSetExpr:
		return &FunctionSetExpr{
			Domain: substituteParams(e.Domain, replacements),
			Range:  substituteParams(e.Range, replacements),
			Pos:    e.Pos,
		}
	case *SetComprehensionExpr:
		compReplacements := map[string]Expr{}
		for name, replacement := range replacements {
			compReplacements[name] = replacement
		}
		comp := &SetComprehensionExpr{Pos: e.Pos}
		for _, bound := range e.Bounds {
			delete(compReplacements, bound.Name)
			comp.Bounds = append(comp.Bounds, BoundVar{
				Name: bound.Name,
				Set:  substituteParams(bound.Set, replacements),
				Pos:  bound.Pos,
			})
		}
		comp.Element = substituteParams(e.Element, compReplacements)
		if e.Predicate != nil {
			comp.Predicate = substituteParams(e.Predicate, compReplacements)
		}
		return comp
	default:
		return expr
	}
}

func moduleVariables(mod *Module) []string {
	var vars []string
	for _, decl := range mod.Declarations {
		if decl.Kind != VariableDecl {
			continue
		}
		vars = append(vars, decl.Names...)
	}
	sort.Strings(vars)
	return vars
}

func initialStates(expr Expr, vars []string, constants State) ([]State, error) {
	assignments, err := evalAssignments(expr, constants, []State{{}})
	if err != nil {
		return nil, err
	}
	if len(assignments) == 0 {
		return nil, fmt.Errorf("initial predicate is false")
	}
	var states []State
	for _, assigns := range assignments {
		st := State{}
		for _, v := range vars {
			val, exists := assigns[v]
			if !exists {
				return nil, fmt.Errorf("initial predicate does not assign %s", v)
			}
			st[v] = val
		}
		states = append(states, st)
	}
	return states, nil
}

func nextStates(expr Expr, st State, vars []string, constants State) ([]State, error) {
	assignments, err := evalAction(expr, st, constants)
	if err != nil {
		return nil, err
	}
	out := make([]State, 0, len(assignments))
	for _, asn := range assignments {
		ns := cloneState(st)
		for name, val := range asn {
			ns[name] = val
		}
		for _, v := range vars {
			if _, ok := ns[v]; !ok {
				ns[v] = st[v]
			}
		}
		out = append(out, ns)
	}
	return out, nil
}

func evalAction(expr Expr, st State, constants State) ([]State, error) {
	env := mergeState(constants, st)
	return evalAssignments(expr, env, []State{{}})
}

func evalAssignments(expr Expr, env State, existing []State) ([]State, error) {
	switch e := expr.(type) {
	case *LetExpr:
		return evalAssignments(expandLet(e), env, existing)
	case *LabelExpr:
		return evalAssignments(e.Body, env, existing)
	case *ActionExpr:
		assignments, err := evalAssignments(e.Action, env, existing)
		if err != nil {
			return nil, err
		}
		if e.Kind != "square" {
			return assignments, nil
		}
		out := make([]State, 0, len(assignments)+len(existing))
		out = append(out, assignments...)
		for _, partial := range existing {
			out = append(out, cloneState(partial))
		}
		return out, nil
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, env)
		if err != nil || !ok {
			return nil, err
		}
		return evalAssignments(selected, env, existing)
	case *QuantifierExpr:
		if e.Kind == "\\E" && e.Set != nil {
			return evalExistentialAssignments(e, env, existing)
		}
		return filterAssignmentsByBool(expr, env, existing)
	case *UnaryExpr:
		if isUnaryJunctionOp(e.Op) {
			return evalAssignments(e.Expr, env, existing)
		}
		if e.Op == "UNCHANGED" {
			out := make([]State, 0, len(existing))
			for _, partial := range existing {
				out = append(out, cloneState(partial))
			}
			return out, nil
		}
	case *BinaryExpr:
		if e.Op == "/\\" {
			left, err := evalAssignments(e.Left, env, existing)
			if err != nil || len(left) == 0 {
				return left, err
			}
			return evalAssignments(e.Right, env, left)
		}
		if e.Op == "\\/" {
			left, err := evalAssignments(e.Left, env, existing)
			if err != nil {
				return nil, err
			}
			right, err := evalAssignments(e.Right, env, existing)
			if err != nil {
				return nil, err
			}
			return append(left, right...), nil
		}
		if e.Op == "=" || e.Op == "\\in" {
			if name, ok := primedIdentifier(e.Left); ok {
				return assignFromExpression(existing, env, name, e.Right, e.Op)
			}
			if name, ok := unprimedIdentifier(e.Left); ok {
				if assigned, err := assignFromExpression(existing, env, name, e.Right, e.Op); err == nil {
					return assigned, nil
				}
			}
		}
	}
	var out []State
	for _, partial := range existing {
		ok, err := evalBool(expr, mergeState(env, partial))
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, cloneState(partial))
		}
	}
	return out, nil
}

func assignFromExpression(existing []State, env State, name string, expr Expr, op string) ([]State, error) {
	var out []State
	for _, partial := range existing {
		localEnv := mergeState(env, partial)
		switch op {
		case "=":
			val, err := evalInt(expr, localEnv)
			if err != nil {
				return nil, err
			}
			next := cloneState(partial)
			next[name] = val
			out = append(out, next)
		case "\\in":
			vals, err := evalIntSet(expr, localEnv)
			if err != nil {
				return nil, err
			}
			for _, val := range vals {
				next := cloneState(partial)
				next[name] = val
				out = append(out, next)
			}
		}
	}
	return out, nil
}

func evalIntSet(expr Expr, st State) ([]int, error) {
	switch e := expr.(type) {
	case *SetExpr:
		vals := make([]int, 0, len(e.Elems))
		for _, elem := range e.Elems {
			val, err := evalInt(elem, st)
			if err != nil {
				return nil, err
			}
			vals = append(vals, val)
		}
		return vals, nil
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return nil, err
		}
		if cond {
			return evalIntSet(e.Then, st)
		}
		return evalIntSet(e.Else, st)
	case *LetExpr:
		return evalIntSet(expandLet(e), st)
	case *LabelExpr:
		return evalIntSet(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("CASE expression has no true arm")
		}
		return evalIntSet(selected, st)
	case *SetComprehensionExpr:
		return evalIntSetComprehension(e, st)
	case *UnaryExpr:
		if e.Op == "DOMAIN" {
			if seq, err := evalIntSequence(e.Expr, st); err == nil {
				domain := make([]int, 0, len(seq))
				for i := range seq {
					domain = append(domain, i+1)
				}
				return domain, nil
			}
			return evalFiniteFunctionIntDomain(e.Expr, st)
		}
		return nil, fmt.Errorf("expression is not a finite integer set")
	case *BinaryExpr:
		switch e.Op {
		case "\\union", "\\intersect", "\\":
			left, err := evalIntSet(e.Left, st)
			if err != nil {
				return nil, err
			}
			right, err := evalIntSet(e.Right, st)
			if err != nil {
				return nil, err
			}
			return evalIntSetBinaryOp(e.Op, left, right), nil
		}
		if e.Op == ".." {
			lo, err := evalInt(e.Left, st)
			if err != nil {
				return nil, err
			}
			hi, err := evalInt(e.Right, st)
			if err != nil {
				return nil, err
			}
			if hi < lo {
				return nil, nil
			}
			vals := make([]int, 0, hi-lo+1)
			for i := lo; i <= hi; i++ {
				vals = append(vals, i)
			}
			return vals, nil
		}
		return nil, fmt.Errorf("expression is not a finite integer set")
	case *QuantifierExpr:
		return nil, fmt.Errorf("quantifier is not a finite integer set")
	default:
		return nil, fmt.Errorf("expression is not a finite integer set")
	}
}

func evalIsFiniteSet(expr Expr, st State) bool {
	if _, err := evalIntSet(expr, st); err == nil {
		return true
	}
	if _, err := evalStringSet(expr, st); err == nil {
		return true
	}
	if _, err := evalModelValueSet(expr, st); err == nil {
		return true
	}
	if _, err := evalIntSetSet(expr, st); err == nil {
		return true
	}
	if _, err := evalStringSetSet(expr, st); err == nil {
		return true
	}
	return false
}

func evalIntSetBinaryOp(op string, left, right []int) []int {
	outSet := map[int]bool{}
	rightSet := map[int]bool{}
	for _, value := range right {
		rightSet[value] = true
	}
	switch op {
	case "\\union":
		for _, value := range left {
			outSet[value] = true
		}
		for value := range rightSet {
			outSet[value] = true
		}
	case "\\intersect":
		for _, value := range left {
			if rightSet[value] {
				outSet[value] = true
			}
		}
	case "\\":
		for _, value := range left {
			if !rightSet[value] {
				outSet[value] = true
			}
		}
	}
	out := make([]int, 0, len(outSet))
	for value := range outSet {
		out = append(out, value)
	}
	sort.Ints(out)
	return out
}

func evalFiniteFunctionIntDomain(expr Expr, st State) ([]int, error) {
	switch e := expr.(type) {
	case *FunctionExpr:
		if len(e.Bounds) != 1 {
			return nil, fmt.Errorf("finite function domain requires one integer domain")
		}
		return evalIntSet(e.Bounds[0].Set, st)
	case *ExceptExpr:
		return evalFiniteFunctionIntDomain(e.Base, st)
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return nil, err
		}
		if cond {
			return evalFiniteFunctionIntDomain(e.Then, st)
		}
		return evalFiniteFunctionIntDomain(e.Else, st)
	case *LetExpr:
		return evalFiniteFunctionIntDomain(expandLet(e), st)
	case *LabelExpr:
		return evalFiniteFunctionIntDomain(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("CASE expression has no true arm")
		}
		return evalFiniteFunctionIntDomain(selected, st)
	default:
		return nil, fmt.Errorf("DOMAIN target is not a finite integer function")
	}
}

func evalIntSetComprehension(expr *SetComprehensionExpr, st State) ([]int, error) {
	var out []int
	seen := map[int]bool{}
	var visit func(int, State) error
	visit = func(index int, env State) error {
		if index == len(expr.Bounds) {
			if expr.Predicate != nil {
				ok, err := evalBool(expr.Predicate, env)
				if err != nil || !ok {
					return err
				}
			}
			value, err := evalInt(expr.Element, env)
			if err != nil {
				return err
			}
			if !seen[value] {
				seen[value] = true
				out = append(out, value)
			}
			return nil
		}
		bound := expr.Bounds[index]
		if bound.Set == nil {
			return fmt.Errorf("set comprehension bound %s has no finite domain", bound.Name)
		}
		values, err := evalIntSet(bound.Set, env)
		if err != nil {
			return err
		}
		for _, value := range values {
			next := cloneState(env)
			next[bound.Name] = value
			if err := visit(index+1, next); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(0, cloneState(st)); err != nil {
		return nil, err
	}
	return out, nil
}

func evalBool(expr Expr, st State) (bool, error) {
	switch e := expr.(type) {
	case *LiteralExpr:
		if e.Kind == "bool" {
			return e.Value == "TRUE", nil
		}
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return false, err
		}
		if cond {
			return evalBool(e.Then, st)
		}
		return evalBool(e.Else, st)
	case *LetExpr:
		return evalBool(expandLet(e), st)
	case *LabelExpr:
		return evalBool(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, fmt.Errorf("CASE expression has no true arm")
		}
		return evalBool(selected, st)
	case *QuantifierExpr:
		return evalQuantifierBool(e, st)
	case *CallExpr:
		if callee, ok := e.Callee.(*IdentExpr); ok {
			switch callee.Name {
			case "Assert":
				if len(e.Args) != 2 {
					return false, fmt.Errorf("Assert expects two arguments")
				}
				ok, err := evalBool(e.Args[0], st)
				if err != nil {
					return false, err
				}
				if !ok {
					message := "TLC Assert failed"
					if text, err := evalString(e.Args[1], st); err == nil && text != "" {
						message += ": " + text
					}
					return false, fmt.Errorf("%s", message)
				}
				return true, nil
			case "Print":
				if len(e.Args) != 2 {
					return false, fmt.Errorf("Print expects two arguments")
				}
				return evalBool(e.Args[1], st)
			case "IsFiniteSet":
				if len(e.Args) != 1 {
					return false, fmt.Errorf("IsFiniteSet expects one argument")
				}
				return evalIsFiniteSet(e.Args[0], st), nil
			}
		}
		return false, fmt.Errorf("unsupported boolean operator call")
	case *UnaryExpr:
		if isUnaryJunctionOp(e.Op) {
			return evalBool(e.Expr, st)
		}
		if e.Op == "\\lnot" || e.Op == "~" {
			v, err := evalBool(e.Expr, st)
			return !v, err
		}
		if e.Op == "ENABLED" {
			assignments, err := evalAction(e.Expr, st, State{})
			if err != nil {
				return false, err
			}
			return len(assignments) > 0, nil
		}
	case *BinaryExpr:
		switch e.Op {
		case "/\\":
			l, err := evalBool(e.Left, st)
			if err != nil || !l {
				return l, err
			}
			return evalBool(e.Right, st)
		case "\\/":
			l, err := evalBool(e.Left, st)
			if err != nil || l {
				return l, err
			}
			return evalBool(e.Right, st)
		case "=>":
			l, err := evalBool(e.Left, st)
			if err != nil || !l {
				return !l, err
			}
			return evalBool(e.Right, st)
		case "<=>", "\\equiv":
			l, err := evalBool(e.Left, st)
			if err != nil {
				return false, err
			}
			r, err := evalBool(e.Right, st)
			if err != nil {
				return false, err
			}
			return l == r, nil
		case "=":
			li, lerr := evalInt(e.Left, st)
			ri, rerr := evalInt(e.Right, st)
			if lerr == nil && rerr == nil {
				return li == ri, nil
			}
			lb, lerr := evalBool(e.Left, st)
			rb, rerr := evalBool(e.Right, st)
			if lerr == nil && rerr == nil {
				return lb == rb, nil
			}
			lstr, lerr := evalString(e.Left, st)
			rstr, rerr := evalString(e.Right, st)
			if lerr == nil && rerr == nil {
				return lstr == rstr, nil
			}
			lmodel, lerr := evalModelValue(e.Left, st)
			rmodel, rerr := evalModelValue(e.Right, st)
			if lerr == nil && rerr == nil {
				return lmodel == rmodel, nil
			}
			ls, lerr := evalIntSet(e.Left, st)
			rs, rerr := evalIntSet(e.Right, st)
			if lerr == nil && rerr == nil {
				return intSetsEqual(ls, rs), nil
			}
			lstrings, lerr := evalStringSet(e.Left, st)
			rstrings, rerr := evalStringSet(e.Right, st)
			if lerr == nil && rerr == nil {
				return stringSetsEqual(lstrings, rstrings), nil
			}
			lmodels, lerr := evalModelValueSet(e.Left, st)
			rmodels, rerr := evalModelValueSet(e.Right, st)
			if lerr == nil && rerr == nil {
				return stringSetsEqual(lmodels, rmodels), nil
			}
			lIntSets, lerr := evalIntSetSet(e.Left, st)
			rIntSets, rerr := evalIntSetSet(e.Right, st)
			if lerr == nil && rerr == nil {
				return intSetSetsEqual(lIntSets, rIntSets), nil
			}
			lStringSets, lerr := evalStringSetSet(e.Left, st)
			rStringSets, rerr := evalStringSetSet(e.Right, st)
			if lerr == nil && rerr == nil {
				return stringSetSetsEqual(lStringSets, rStringSets), nil
			}
			lseq, lerr := evalIntSequence(e.Left, st)
			rseq, rerr := evalIntSequence(e.Right, st)
			if lerr == nil && rerr == nil {
				return intSequencesEqual(lseq, rseq), nil
			}
			return false, fmt.Errorf("cannot compare values with =")
		case "/=":
			eq, err := evalBool(&BinaryExpr{Op: "=", Left: e.Left, Right: e.Right, Pos: e.Pos}, st)
			return !eq, err
		case "\\in":
			if val, err := evalInt(e.Left, st); err == nil {
				set, err := evalIntSet(e.Right, st)
				if err != nil {
					return false, err
				}
				for _, item := range set {
					if item == val {
						return true, nil
					}
				}
				return false, nil
			}
			if val, err := evalString(e.Left, st); err == nil {
				set, err := evalStringSet(e.Right, st)
				if err != nil {
					return false, err
				}
				for _, item := range set {
					if item == val {
						return true, nil
					}
				}
				return false, nil
			}
			if val, err := evalModelValue(e.Left, st); err == nil {
				set, err := evalModelValueSet(e.Right, st)
				if err != nil {
					return false, err
				}
				for _, item := range set {
					if item == val {
						return true, nil
					}
				}
				return false, nil
			}
			if val, err := evalIntSet(e.Left, st); err == nil {
				set, err := evalIntSetSet(e.Right, st)
				if err != nil {
					return false, err
				}
				return intSetSetContains(set, val), nil
			}
			if val, err := evalStringSet(e.Left, st); err == nil {
				set, err := evalStringSetSet(e.Right, st)
				if err != nil {
					return false, err
				}
				return stringSetSetContains(set, val), nil
			}
			return false, fmt.Errorf("cannot evaluate membership")
		case "\\notin":
			member, err := evalBool(&BinaryExpr{Op: "\\in", Left: e.Left, Right: e.Right, Pos: e.Pos}, st)
			return !member, err
		case "\\subseteq":
			if left, err := evalIntSet(e.Left, st); err == nil {
				right, err := evalIntSet(e.Right, st)
				if err != nil {
					return false, err
				}
				return intSetSubset(left, right), nil
			}
			if left, err := evalStringSet(e.Left, st); err == nil {
				right, err := evalStringSet(e.Right, st)
				if err != nil {
					return false, err
				}
				return stringSetSubset(left, right), nil
			}
			left, err := evalModelValueSet(e.Left, st)
			if err != nil {
				return false, err
			}
			right, err := evalModelValueSet(e.Right, st)
			if err != nil {
				return false, err
			}
			return stringSetSubset(left, right), nil
		case "\\subset":
			if left, err := evalIntSet(e.Left, st); err == nil {
				right, err := evalIntSet(e.Right, st)
				if err != nil {
					return false, err
				}
				return intSetSubset(left, right) && !intSetsEqual(left, right), nil
			}
			if left, err := evalStringSet(e.Left, st); err == nil {
				right, err := evalStringSet(e.Right, st)
				if err != nil {
					return false, err
				}
				return stringSetSubset(left, right) && !stringSetsEqual(left, right), nil
			}
			left, err := evalModelValueSet(e.Left, st)
			if err != nil {
				return false, err
			}
			right, err := evalModelValueSet(e.Right, st)
			if err != nil {
				return false, err
			}
			return stringSetSubset(left, right) && !stringSetsEqual(left, right), nil
		case "<", ">", "\\leq", "\\geq":
			l, err := evalInt(e.Left, st)
			if err != nil {
				return false, err
			}
			r, err := evalInt(e.Right, st)
			if err != nil {
				return false, err
			}
			switch e.Op {
			case "<":
				return l < r, nil
			case ">":
				return l > r, nil
			case "\\leq":
				return l <= r, nil
			case "\\geq":
				return l >= r, nil
			}
		}
	}
	return false, fmt.Errorf("expression is not boolean")
}

func evalString(expr Expr, st State) (string, error) {
	switch e := expr.(type) {
	case *LiteralExpr:
		if e.Kind == "string" {
			return e.Value, nil
		}
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return "", err
		}
		if cond {
			return evalString(e.Then, st)
		}
		return evalString(e.Else, st)
	case *LetExpr:
		return evalString(expandLet(e), st)
	case *LabelExpr:
		return evalString(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("CASE expression has no true arm")
		}
		return evalString(selected, st)
	case *CallExpr:
		if callee, ok := e.Callee.(*IdentExpr); ok && callee.Name == "Print" {
			if len(e.Args) != 2 {
				return "", fmt.Errorf("Print expects two arguments")
			}
			return evalString(e.Args[1], st)
		}
	}
	return "", fmt.Errorf("expression is not a string")
}

func evalModelValue(expr Expr, st State) (string, error) {
	switch e := expr.(type) {
	case *LiteralExpr:
		if e.Kind == "model" {
			return e.Value, nil
		}
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return "", err
		}
		if cond {
			return evalModelValue(e.Then, st)
		}
		return evalModelValue(e.Else, st)
	case *LetExpr:
		return evalModelValue(expandLet(e), st)
	case *LabelExpr:
		return evalModelValue(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("CASE expression has no true arm")
		}
		return evalModelValue(selected, st)
	case *CallExpr:
		if callee, ok := e.Callee.(*IdentExpr); ok && callee.Name == "Print" {
			if len(e.Args) != 2 {
				return "", fmt.Errorf("Print expects two arguments")
			}
			return evalModelValue(e.Args[1], st)
		}
	}
	return "", fmt.Errorf("expression is not a model value")
}

func evalStringSet(expr Expr, st State) ([]string, error) {
	switch e := expr.(type) {
	case *SetExpr:
		vals := make([]string, 0, len(e.Elems))
		for _, elem := range e.Elems {
			val, err := evalString(elem, st)
			if err != nil {
				return nil, err
			}
			vals = append(vals, val)
		}
		return vals, nil
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return nil, err
		}
		if cond {
			return evalStringSet(e.Then, st)
		}
		return evalStringSet(e.Else, st)
	case *LetExpr:
		return evalStringSet(expandLet(e), st)
	case *LabelExpr:
		return evalStringSet(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("CASE expression has no true arm")
		}
		return evalStringSet(selected, st)
	case *BinaryExpr:
		switch e.Op {
		case "\\union", "\\intersect", "\\":
			left, err := evalStringSet(e.Left, st)
			if err != nil {
				return nil, err
			}
			right, err := evalStringSet(e.Right, st)
			if err != nil {
				return nil, err
			}
			return evalStringSetBinaryOp(e.Op, left, right), nil
		}
	}
	return nil, fmt.Errorf("expression is not a finite string set")
}

func evalModelValueSet(expr Expr, st State) ([]string, error) {
	switch e := expr.(type) {
	case *SetExpr:
		vals := make([]string, 0, len(e.Elems))
		for _, elem := range e.Elems {
			val, err := evalModelValue(elem, st)
			if err != nil {
				return nil, err
			}
			vals = append(vals, val)
		}
		return vals, nil
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return nil, err
		}
		if cond {
			return evalModelValueSet(e.Then, st)
		}
		return evalModelValueSet(e.Else, st)
	case *LetExpr:
		return evalModelValueSet(expandLet(e), st)
	case *LabelExpr:
		return evalModelValueSet(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("CASE expression has no true arm")
		}
		return evalModelValueSet(selected, st)
	case *BinaryExpr:
		switch e.Op {
		case "\\union", "\\intersect", "\\":
			left, err := evalModelValueSet(e.Left, st)
			if err != nil {
				return nil, err
			}
			right, err := evalModelValueSet(e.Right, st)
			if err != nil {
				return nil, err
			}
			return evalStringSetBinaryOp(e.Op, left, right), nil
		}
	}
	return nil, fmt.Errorf("expression is not a finite model-value set")
}

func evalStringSetBinaryOp(op string, left, right []string) []string {
	outSet := map[string]bool{}
	rightSet := map[string]bool{}
	for _, value := range right {
		rightSet[value] = true
	}
	switch op {
	case "\\union":
		for _, value := range left {
			outSet[value] = true
		}
		for value := range rightSet {
			outSet[value] = true
		}
	case "\\intersect":
		for _, value := range left {
			if rightSet[value] {
				outSet[value] = true
			}
		}
	case "\\":
		for _, value := range left {
			if !rightSet[value] {
				outSet[value] = true
			}
		}
	}
	out := make([]string, 0, len(outSet))
	for value := range outSet {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func evalIntSetSet(expr Expr, st State) ([][]int, error) {
	switch e := expr.(type) {
	case *SetExpr:
		values := make([][]int, 0, len(e.Elems))
		for _, elem := range e.Elems {
			value, err := evalIntSet(elem, st)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	case *UnaryExpr:
		if e.Op == "SUBSET" {
			values, err := evalIntSet(e.Expr, st)
			if err != nil {
				return nil, err
			}
			return intPowerset(values), nil
		}
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return nil, err
		}
		if cond {
			return evalIntSetSet(e.Then, st)
		}
		return evalIntSetSet(e.Else, st)
	case *LetExpr:
		return evalIntSetSet(expandLet(e), st)
	case *LabelExpr:
		return evalIntSetSet(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("CASE expression has no true arm")
		}
		return evalIntSetSet(selected, st)
	}
	return nil, fmt.Errorf("expression is not a finite integer powerset")
}

func evalStringSetSet(expr Expr, st State) ([][]string, error) {
	switch e := expr.(type) {
	case *SetExpr:
		values := make([][]string, 0, len(e.Elems))
		for _, elem := range e.Elems {
			value, err := evalStringSet(elem, st)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	case *UnaryExpr:
		if e.Op == "SUBSET" {
			values, err := evalStringSet(e.Expr, st)
			if err != nil {
				return nil, err
			}
			return stringPowerset(values), nil
		}
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return nil, err
		}
		if cond {
			return evalStringSetSet(e.Then, st)
		}
		return evalStringSetSet(e.Else, st)
	case *LetExpr:
		return evalStringSetSet(expandLet(e), st)
	case *LabelExpr:
		return evalStringSetSet(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("CASE expression has no true arm")
		}
		return evalStringSetSet(selected, st)
	}
	return nil, fmt.Errorf("expression is not a finite string powerset")
}

func intPowerset(values []int) [][]int {
	base := normalizedIntSet(values)
	out := make([][]int, 0, 1<<len(base))
	for mask := 0; mask < 1<<len(base); mask++ {
		subset := make([]int, 0, len(base))
		for i, value := range base {
			if mask&(1<<i) != 0 {
				subset = append(subset, value)
			}
		}
		out = append(out, subset)
	}
	return out
}

func stringPowerset(values []string) [][]string {
	base := normalizedStringSet(values)
	out := make([][]string, 0, 1<<len(base))
	for mask := 0; mask < 1<<len(base); mask++ {
		subset := make([]string, 0, len(base))
		for i, value := range base {
			if mask&(1<<i) != 0 {
				subset = append(subset, value)
			}
		}
		out = append(out, subset)
	}
	return out
}

func normalizedIntSet(values []int) []int {
	seen := map[int]bool{}
	for _, value := range values {
		seen[value] = true
	}
	out := make([]int, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Ints(out)
	return out
}

func normalizedStringSet(values []string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	out := make([]string, 0, len(seen))
	for value := range seen {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func intSetKey(values []int) string {
	normalized := normalizedIntSet(values)
	parts := make([]string, 0, len(normalized))
	for _, value := range normalized {
		parts = append(parts, strconv.Itoa(value))
	}
	return strings.Join(parts, ",")
}

func stringSetKey(values []string) string {
	normalized := normalizedStringSet(values)
	parts := make([]string, 0, len(normalized))
	for _, value := range normalized {
		parts = append(parts, strconv.Quote(value))
	}
	return strings.Join(parts, ",")
}

func intSetSetContains(sets [][]int, target []int) bool {
	key := intSetKey(target)
	for _, set := range sets {
		if intSetKey(set) == key {
			return true
		}
	}
	return false
}

func stringSetSetContains(sets [][]string, target []string) bool {
	key := stringSetKey(target)
	for _, set := range sets {
		if stringSetKey(set) == key {
			return true
		}
	}
	return false
}

func intSetsEqual(left, right []int) bool {
	return intSetSubset(left, right) && intSetSubset(right, left)
}

func stringSetsEqual(left, right []string) bool {
	return stringSetSubset(left, right) && stringSetSubset(right, left)
}

func intSetSetsEqual(left, right [][]int) bool {
	if intSetSetCardinality(left) != intSetSetCardinality(right) {
		return false
	}
	for _, value := range left {
		if !intSetSetContains(right, value) {
			return false
		}
	}
	return true
}

func stringSetSetsEqual(left, right [][]string) bool {
	if stringSetSetCardinality(left) != stringSetSetCardinality(right) {
		return false
	}
	for _, value := range left {
		if !stringSetSetContains(right, value) {
			return false
		}
	}
	return true
}

func intSequencesEqual(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func intSetSubset(left, right []int) bool {
	rightSet := map[int]bool{}
	for _, value := range right {
		rightSet[value] = true
	}
	for _, value := range left {
		if !rightSet[value] {
			return false
		}
	}
	return true
}

func stringSetSubset(left, right []string) bool {
	rightSet := map[string]bool{}
	for _, value := range right {
		rightSet[value] = true
	}
	for _, value := range left {
		if !rightSet[value] {
			return false
		}
	}
	return true
}

func evalInt(expr Expr, st State) (int, error) {
	switch e := expr.(type) {
	case *LiteralExpr:
		if e.Kind == "number" {
			return parseTLAIntegerLiteral(e.Value)
		}
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return 0, err
		}
		if cond {
			return evalInt(e.Then, st)
		}
		return evalInt(e.Else, st)
	case *LetExpr:
		return evalInt(expandLet(e), st)
	case *LabelExpr:
		return evalInt(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, fmt.Errorf("CASE expression has no true arm")
		}
		return evalInt(selected, st)
	case *ChooseExpr:
		return evalChooseInt(e, st)
	case *FunctionAppExpr:
		return evalFunctionAppInt(e, st)
	case *RecordComponentExpr:
		return evalRecordComponentInt(e, st)
	case *QuantifierExpr:
		return 0, fmt.Errorf("quantifier is not an integer")
	case *CallExpr:
		if callee, ok := e.Callee.(*IdentExpr); ok {
			switch callee.Name {
			case "Cardinality":
				if len(e.Args) != 1 {
					return 0, fmt.Errorf("Cardinality expects one argument")
				}
				if values, err := evalIntSet(e.Args[0], st); err == nil {
					return intSetCardinality(values), nil
				}
				if values, err := evalStringSet(e.Args[0], st); err == nil {
					return stringSetCardinality(values), nil
				}
				if values, err := evalModelValueSet(e.Args[0], st); err == nil {
					return stringSetCardinality(values), nil
				}
				if values, err := evalIntSetSet(e.Args[0], st); err == nil {
					return intSetSetCardinality(values), nil
				}
				values, err := evalStringSetSet(e.Args[0], st)
				if err != nil {
					return 0, err
				}
				return stringSetSetCardinality(values), nil
			case "Len":
				if len(e.Args) != 1 {
					return 0, fmt.Errorf("Len expects one argument")
				}
				values, err := evalIntSequence(e.Args[0], st)
				if err != nil {
					return 0, err
				}
				return len(values), nil
			case "Head":
				if len(e.Args) != 1 {
					return 0, fmt.Errorf("Head expects one argument")
				}
				values, err := evalIntSequence(e.Args[0], st)
				if err != nil {
					return 0, err
				}
				if len(values) == 0 {
					return 0, fmt.Errorf("Head expects a non-empty sequence")
				}
				return values[0], nil
			case "Print":
				if len(e.Args) != 2 {
					return 0, fmt.Errorf("Print expects two arguments")
				}
				return evalInt(e.Args[1], st)
			}
		}
		return 0, fmt.Errorf("unsupported integer operator call")
	case *IdentExpr:
		if v, ok := st[e.Name]; ok {
			return v, nil
		}
		return 0, fmt.Errorf("unknown integer identifier %s", e.Name)
	case *UnaryExpr:
		if e.Op == "-" {
			v, err := evalInt(e.Expr, st)
			return -v, err
		}
		if e.Op == "'" {
			if name, ok := unprimedIdentifier(e.Expr); ok {
				if v, exists := st[name+"'"]; exists {
					return v, nil
				}
				return 0, fmt.Errorf("unknown primed integer identifier %s'", name)
			}
		}
	case *BinaryExpr:
		l, err := evalInt(e.Left, st)
		if err != nil {
			return 0, err
		}
		r, err := evalInt(e.Right, st)
		if err != nil {
			return 0, err
		}
		switch e.Op {
		case "+":
			return l + r, nil
		case "-":
			return l - r, nil
		case "*":
			return l * r, nil
		case "/":
			if r == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return l / r, nil
		case "\\div":
			if r == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return l / r, nil
		case "%":
			if r == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return l % r, nil
		}
	}
	return 0, fmt.Errorf("expression is not an integer")
}

func parseTLAIntegerLiteral(text string) (int, error) {
	if len(text) >= 3 && text[0] == '\\' {
		base := 0
		switch text[1] {
		case 'b', 'B':
			base = 2
		case 'o', 'O':
			base = 8
		case 'h', 'H':
			base = 16
		}
		if base != 0 {
			value, err := strconv.ParseInt(text[2:], base, 0)
			if err != nil {
				return 0, err
			}
			return int(value), nil
		}
	}
	return strconv.Atoi(text)
}

func evalIntSequence(expr Expr, st State) ([]int, error) {
	switch e := expr.(type) {
	case *TupleExpr:
		values := make([]int, 0, len(e.Elems))
		for _, elem := range e.Elems {
			value, err := evalInt(elem, st)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return nil, err
		}
		if cond {
			return evalIntSequence(e.Then, st)
		}
		return evalIntSequence(e.Else, st)
	case *LetExpr:
		return evalIntSequence(expandLet(e), st)
	case *LabelExpr:
		return evalIntSequence(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("CASE expression has no true arm")
		}
		return evalIntSequence(selected, st)
	case *CallExpr:
		callee, ok := e.Callee.(*IdentExpr)
		if !ok {
			break
		}
		switch callee.Name {
		case "Append":
			if len(e.Args) != 2 {
				return nil, fmt.Errorf("Append expects two arguments")
			}
			values, err := evalIntSequence(e.Args[0], st)
			if err != nil {
				return nil, err
			}
			value, err := evalInt(e.Args[1], st)
			if err != nil {
				return nil, err
			}
			out := append([]int{}, values...)
			out = append(out, value)
			return out, nil
		case "Tail":
			if len(e.Args) != 1 {
				return nil, fmt.Errorf("Tail expects one argument")
			}
			values, err := evalIntSequence(e.Args[0], st)
			if err != nil {
				return nil, err
			}
			if len(values) == 0 {
				return nil, fmt.Errorf("Tail expects a non-empty sequence")
			}
			return append([]int{}, values[1:]...), nil
		case "SubSeq":
			if len(e.Args) != 3 {
				return nil, fmt.Errorf("SubSeq expects three arguments")
			}
			values, err := evalIntSequence(e.Args[0], st)
			if err != nil {
				return nil, err
			}
			lo, err := evalInt(e.Args[1], st)
			if err != nil {
				return nil, err
			}
			hi, err := evalInt(e.Args[2], st)
			if err != nil {
				return nil, err
			}
			if hi < lo {
				return nil, nil
			}
			if lo < 1 || hi > len(values) {
				return nil, fmt.Errorf("SubSeq bounds %d..%d outside sequence domain", lo, hi)
			}
			return append([]int{}, values[lo-1:hi]...), nil
		case "SelectSeq":
			if len(e.Args) != 2 {
				return nil, fmt.Errorf("SelectSeq expects two arguments")
			}
			values, err := evalIntSequence(e.Args[0], st)
			if err != nil {
				return nil, err
			}
			pred, ok := e.Args[1].(*FunctionExpr)
			if !ok || len(pred.Bounds) != 1 {
				return nil, fmt.Errorf("SelectSeq expects a unary predicate")
			}
			out := make([]int, 0, len(values))
			for _, value := range values {
				env := cloneState(st)
				env[pred.Bounds[0].Name] = value
				keep, err := evalBool(pred.Body, env)
				if err != nil {
					return nil, err
				}
				if keep {
					out = append(out, value)
				}
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("expression is not a finite integer sequence")
}

func intSetCardinality(values []int) int {
	seen := map[int]bool{}
	for _, value := range values {
		seen[value] = true
	}
	return len(seen)
}

func stringSetCardinality(values []string) int {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	return len(seen)
}

func intSetSetCardinality(values [][]int) int {
	seen := map[string]bool{}
	for _, value := range values {
		seen[intSetKey(value)] = true
	}
	return len(seen)
}

func stringSetSetCardinality(values [][]string) int {
	seen := map[string]bool{}
	for _, value := range values {
		seen[stringSetKey(value)] = true
	}
	return len(seen)
}

func evalRecordComponentInt(expr *RecordComponentExpr, st State) (int, error) {
	record, err := evalRecordIntFields(expr.Record, st)
	if err != nil {
		return 0, err
	}
	value, ok := record[expr.Field]
	if !ok {
		return 0, fmt.Errorf("record field %s not found", expr.Field)
	}
	return value, nil
}

func evalRecordIntFields(expr Expr, st State) (map[string]int, error) {
	switch e := expr.(type) {
	case *RecordExpr:
		record := map[string]int{}
		for _, field := range e.Fields {
			value, err := evalInt(field.Value, st)
			if err != nil {
				return nil, err
			}
			record[field.Name] = value
		}
		return record, nil
	case *ExceptExpr:
		record, err := evalRecordIntFields(e.Base, st)
		if err != nil {
			return nil, err
		}
		for _, spec := range e.Specs {
			if len(spec.Components) != 1 {
				return nil, fmt.Errorf("nested record EXCEPT paths are not supported")
			}
			component := spec.Components[0]
			if component.Field == "" || len(component.Indices) != 0 {
				return nil, fmt.Errorf("function-style EXCEPT updates are not supported for records")
			}
			old, ok := record[component.Field]
			if !ok {
				return nil, fmt.Errorf("record field %s not found", component.Field)
			}
			env := cloneState(st)
			env["@"] = old
			value, err := evalInt(spec.Value, env)
			if err != nil {
				return nil, err
			}
			record[component.Field] = value
		}
		return record, nil
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return nil, err
		}
		if cond {
			return evalRecordIntFields(e.Then, st)
		}
		return evalRecordIntFields(e.Else, st)
	case *LetExpr:
		return evalRecordIntFields(expandLet(e), st)
	case *LabelExpr:
		return evalRecordIntFields(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("CASE expression has no true arm")
		}
		return evalRecordIntFields(selected, st)
	default:
		return nil, fmt.Errorf("record component target is not a record")
	}
}

func evalFunctionAppInt(expr *FunctionAppExpr, st State) (int, error) {
	fn, ok := expr.Function.(*FunctionExpr)
	if !ok {
		if len(expr.Args) != 1 {
			return 0, fmt.Errorf("function application target is not a finite function")
		}
		arg, err := evalInt(expr.Args[0], st)
		if err != nil {
			return 0, err
		}
		seq, seqErr := evalIntSequence(expr.Function, st)
		if seqErr == nil {
			if arg < 1 || arg > len(seq) {
				return 0, fmt.Errorf("sequence index %d out of domain", arg)
			}
			return seq[arg-1], nil
		}
		values, err := evalFiniteFunctionIntValues(expr.Function, st)
		if err != nil {
			return 0, err
		}
		value, exists := values[arg]
		if !exists {
			return 0, fmt.Errorf("function argument %d not in finite domain", arg)
		}
		return value, nil
	}
	if len(fn.Bounds) != len(expr.Args) {
		return 0, fmt.Errorf("function application arity mismatch: got %d args, want %d", len(expr.Args), len(fn.Bounds))
	}
	env := cloneState(st)
	for i, bound := range fn.Bounds {
		arg, err := evalInt(expr.Args[i], st)
		if err != nil {
			return 0, err
		}
		if bound.Set != nil {
			values, err := evalIntSet(bound.Set, st)
			if err != nil {
				return 0, err
			}
			if !intInSet(arg, values) {
				return 0, fmt.Errorf("function argument %d not in finite domain", arg)
			}
		}
		env[bound.Name] = arg
	}
	return evalInt(fn.Body, env)
}

func evalFiniteFunctionIntValues(expr Expr, st State) (map[int]int, error) {
	switch e := expr.(type) {
	case *FunctionExpr:
		if len(e.Bounds) != 1 {
			return nil, fmt.Errorf("finite function value requires one integer domain")
		}
		bound := e.Bounds[0]
		values, err := evalIntSet(bound.Set, st)
		if err != nil {
			return nil, err
		}
		out := map[int]int{}
		for _, arg := range values {
			env := cloneState(st)
			env[bound.Name] = arg
			value, err := evalInt(e.Body, env)
			if err != nil {
				return nil, err
			}
			out[arg] = value
		}
		return out, nil
	case *ExceptExpr:
		values, err := evalFiniteFunctionIntValues(e.Base, st)
		if err != nil {
			return nil, err
		}
		for _, spec := range e.Specs {
			if len(spec.Components) != 1 {
				return nil, fmt.Errorf("nested function EXCEPT paths are not supported")
			}
			component := spec.Components[0]
			if component.Field != "" || len(component.Indices) != 1 {
				return nil, fmt.Errorf("record-style EXCEPT updates are not supported for functions")
			}
			index, err := evalInt(component.Indices[0], st)
			if err != nil {
				return nil, err
			}
			old, ok := values[index]
			if !ok {
				return nil, fmt.Errorf("function argument %d not in finite domain", index)
			}
			env := cloneState(st)
			env["@"] = old
			value, err := evalInt(spec.Value, env)
			if err != nil {
				return nil, err
			}
			values[index] = value
		}
		return values, nil
	case *IfExpr:
		cond, err := evalBool(e.Cond, st)
		if err != nil {
			return nil, err
		}
		if cond {
			return evalFiniteFunctionIntValues(e.Then, st)
		}
		return evalFiniteFunctionIntValues(e.Else, st)
	case *LetExpr:
		return evalFiniteFunctionIntValues(expandLet(e), st)
	case *LabelExpr:
		return evalFiniteFunctionIntValues(e.Body, st)
	case *CaseExpr:
		selected, ok, err := selectCaseExpr(e, st)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("CASE expression has no true arm")
		}
		return evalFiniteFunctionIntValues(selected, st)
	default:
		return nil, fmt.Errorf("function application target is not a finite function")
	}
}

func intInSet(value int, values []int) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func filterAssignmentsByBool(expr Expr, env State, existing []State) ([]State, error) {
	var out []State
	for _, partial := range existing {
		ok, err := evalBool(expr, mergeState(env, partial))
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, cloneState(partial))
		}
	}
	return out, nil
}

func evalExistentialAssignments(expr *QuantifierExpr, env State, existing []State) ([]State, error) {
	var out []State
	for _, partial := range existing {
		localEnv := mergeState(env, partial)
		values, err := evalIntSet(expr.Set, localEnv)
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			boundEnv := cloneState(localEnv)
			boundEnv[expr.Var] = value
			assignments, err := evalAssignments(expr.Body, boundEnv, []State{cloneState(partial)})
			if err != nil {
				return nil, err
			}
			out = append(out, assignments...)
		}
	}
	return out, nil
}

func evalQuantifierBool(expr *QuantifierExpr, st State) (bool, error) {
	values, err := evalIntSet(expr.Set, st)
	if err != nil {
		return false, err
	}
	switch expr.Kind {
	case "\\A":
		for _, value := range values {
			env := cloneState(st)
			env[expr.Var] = value
			ok, err := evalBool(expr.Body, env)
			if err != nil || !ok {
				return ok, err
			}
		}
		return true, nil
	case "\\E":
		for _, value := range values {
			env := cloneState(st)
			env[expr.Var] = value
			ok, err := evalBool(expr.Body, env)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported quantifier %s", expr.Kind)
	}
}

func selectCaseExpr(expr *CaseExpr, st State) (Expr, bool, error) {
	for _, arm := range expr.Arms {
		ok, err := evalBool(arm.Test, st)
		if err != nil {
			return nil, false, err
		}
		if ok {
			return arm.Value, true, nil
		}
	}
	if expr.Other != nil {
		return expr.Other, true, nil
	}
	return nil, false, nil
}

func evalChooseInt(expr *ChooseExpr, st State) (int, error) {
	if expr.Set == nil {
		return 0, fmt.Errorf("unbounded CHOOSE is not an integer")
	}
	values, err := evalIntSet(expr.Set, st)
	if err != nil {
		return 0, err
	}
	for _, value := range values {
		env := cloneState(st)
		env[expr.Var] = value
		ok, err := evalBool(expr.Body, env)
		if err != nil {
			return 0, err
		}
		if ok {
			return value, nil
		}
	}
	return 0, fmt.Errorf("CHOOSE predicate has no satisfying integer")
}

func expandLet(let *LetExpr) Expr {
	bindings := map[string]Expr{}
	for _, def := range let.Definitions {
		if len(def.Params) == 0 && def.Expr != nil {
			bindings[def.Name] = substituteLets(def.Expr, bindings, map[string]bool{})
		}
	}
	return substituteLets(let.Body, bindings, map[string]bool{})
}

func substituteLets(expr Expr, bindings map[string]Expr, seen map[string]bool) Expr {
	switch e := expr.(type) {
	case *IdentExpr:
		bound, ok := bindings[e.Name]
		if !ok || seen[e.Name] {
			return e
		}
		nextSeen := copySeen(seen)
		nextSeen[e.Name] = true
		return substituteLets(bound, bindings, nextSeen)
	case *UnaryExpr:
		return &UnaryExpr{Op: e.Op, Expr: substituteLets(e.Expr, bindings, copySeen(seen)), Pos: e.Pos}
	case *BinaryExpr:
		return &BinaryExpr{
			Op:           e.Op,
			Left:         substituteLets(e.Left, bindings, copySeen(seen)),
			Right:        substituteLets(e.Right, bindings, copySeen(seen)),
			Pos:          e.Pos,
			JunctionList: e.JunctionList,
		}
	case *CallExpr:
		call := &CallExpr{Callee: substituteLets(e.Callee, bindings, copySeen(seen)), Pos: e.Pos}
		for _, arg := range e.Args {
			call.Args = append(call.Args, substituteLets(arg, bindings, copySeen(seen)))
		}
		return call
	case *IfExpr:
		return &IfExpr{
			Cond: substituteLets(e.Cond, bindings, copySeen(seen)),
			Then: substituteLets(e.Then, bindings, copySeen(seen)),
			Else: substituteLets(e.Else, bindings, copySeen(seen)),
			Pos:  e.Pos,
		}
	case *LetExpr:
		nested := map[string]Expr{}
		for name, value := range bindings {
			nested[name] = value
		}
		for _, def := range e.Definitions {
			if len(def.Params) == 0 && def.Expr != nil {
				nested[def.Name] = substituteLets(def.Expr, nested, map[string]bool{})
			}
		}
		return substituteLets(e.Body, nested, map[string]bool{})
	case *QuantifierExpr:
		quantBindings := map[string]Expr{}
		for name, value := range bindings {
			if name != e.Var {
				quantBindings[name] = value
			}
		}
		return &QuantifierExpr{
			Kind:       e.Kind,
			Var:        e.Var,
			VarPos:     e.VarPos,
			Set:        substituteLets(e.Set, bindings, copySeen(seen)),
			Body:       substituteLets(e.Body, quantBindings, copySeen(seen)),
			TupleBound: e.TupleBound,
			Pos:        e.Pos,
		}
	case *CaseExpr:
		caseExpr := &CaseExpr{Pos: e.Pos, OtherPos: e.OtherPos}
		for _, arm := range e.Arms {
			caseExpr.Arms = append(caseExpr.Arms, CaseArm{
				Test:  substituteLets(arm.Test, bindings, copySeen(seen)),
				Value: substituteLets(arm.Value, bindings, copySeen(seen)),
				Pos:   arm.Pos,
			})
		}
		if e.Other != nil {
			caseExpr.Other = substituteLets(e.Other, bindings, copySeen(seen))
		}
		return caseExpr
	case *ChooseExpr:
		chooseBindings := map[string]Expr{}
		for name, binding := range bindings {
			if name != e.Var {
				chooseBindings[name] = binding
			}
		}
		return &ChooseExpr{Var: e.Var, VarPos: e.VarPos, Set: substituteLets(e.Set, bindings, copySeen(seen)), Body: substituteLets(e.Body, chooseBindings, copySeen(seen)), Pos: e.Pos}
	case *TupleExpr:
		tuple := &TupleExpr{Pos: e.Pos}
		for _, elem := range e.Elems {
			tuple.Elems = append(tuple.Elems, substituteLets(elem, bindings, copySeen(seen)))
		}
		return tuple
	case *SetExpr:
		set := &SetExpr{Pos: e.Pos}
		for _, elem := range e.Elems {
			set.Elems = append(set.Elems, substituteLets(elem, bindings, copySeen(seen)))
		}
		return set
	case *RecordExpr:
		record := &RecordExpr{Pos: e.Pos}
		for _, field := range e.Fields {
			record.Fields = append(record.Fields, RecordField{
				Name:   field.Name,
				Value:  substituteLets(field.Value, bindings, copySeen(seen)),
				Pos:    field.Pos,
				Source: field.Source,
			})
		}
		return record
	case *RecordComponentExpr:
		return &RecordComponentExpr{Record: substituteLets(e.Record, bindings, copySeen(seen)), Field: e.Field, FieldPos: e.FieldPos, Pos: e.Pos}
	case *RecordSetExpr:
		record := &RecordSetExpr{Pos: e.Pos}
		for _, field := range e.Fields {
			record.Fields = append(record.Fields, RecordSetField{
				Name:   field.Name,
				Set:    substituteLets(field.Set, bindings, copySeen(seen)),
				Pos:    field.Pos,
				Source: field.Source,
			})
		}
		return record
	case *FunctionExpr:
		fnBindings := map[string]Expr{}
		for name, binding := range bindings {
			fnBindings[name] = binding
		}
		fn := &FunctionExpr{Pos: e.Pos}
		for _, bound := range e.Bounds {
			delete(fnBindings, bound.Name)
			fn.Bounds = append(fn.Bounds, BoundVar{
				Name: bound.Name,
				Set:  substituteLets(bound.Set, bindings, copySeen(seen)),
				Pos:  bound.Pos,
			})
		}
		fn.Body = substituteLets(e.Body, fnBindings, copySeen(seen))
		return fn
	case *FunctionAppExpr:
		app := &FunctionAppExpr{Function: substituteLets(e.Function, bindings, copySeen(seen)), Pos: e.Pos}
		for _, arg := range e.Args {
			app.Args = append(app.Args, substituteLets(arg, bindings, copySeen(seen)))
		}
		return app
	case *ExceptExpr:
		except := &ExceptExpr{Base: substituteLets(e.Base, bindings, copySeen(seen)), Pos: e.Pos}
		for _, spec := range e.Specs {
			nextSpec := ExceptSpec{Value: substituteLets(spec.Value, bindings, copySeen(seen)), Pos: spec.Pos}
			for _, component := range spec.Components {
				nextComponent := ExceptComponent{Field: component.Field, FieldPos: component.FieldPos, Pos: component.Pos}
				for _, index := range component.Indices {
					nextComponent.Indices = append(nextComponent.Indices, substituteLets(index, bindings, copySeen(seen)))
				}
				nextSpec.Components = append(nextSpec.Components, nextComponent)
			}
			except.Specs = append(except.Specs, nextSpec)
		}
		return except
	case *LabelExpr:
		return &LabelExpr{Name: e.Name, Body: substituteLets(e.Body, bindings, copySeen(seen)), Pos: e.Pos}
	case *ActionExpr:
		return &ActionExpr{
			Kind:      e.Kind,
			Action:    substituteLets(e.Action, bindings, copySeen(seen)),
			Subscript: substituteLets(e.Subscript, bindings, copySeen(seen)),
			Pos:       e.Pos,
		}
	case *FairnessExpr:
		return &FairnessExpr{
			Kind:      e.Kind,
			Subscript: substituteLets(e.Subscript, bindings, copySeen(seen)),
			Action:    substituteLets(e.Action, bindings, copySeen(seen)),
			Pos:       e.Pos,
		}
	case *FunctionSetExpr:
		return &FunctionSetExpr{
			Domain: substituteLets(e.Domain, bindings, copySeen(seen)),
			Range:  substituteLets(e.Range, bindings, copySeen(seen)),
			Pos:    e.Pos,
		}
	case *SetComprehensionExpr:
		compBindings := map[string]Expr{}
		for name, binding := range bindings {
			compBindings[name] = binding
		}
		comp := &SetComprehensionExpr{Pos: e.Pos}
		for _, bound := range e.Bounds {
			delete(compBindings, bound.Name)
			comp.Bounds = append(comp.Bounds, BoundVar{
				Name: bound.Name,
				Set:  substituteLets(bound.Set, bindings, copySeen(seen)),
				Pos:  bound.Pos,
			})
		}
		comp.Element = substituteLets(e.Element, compBindings, copySeen(seen))
		if e.Predicate != nil {
			comp.Predicate = substituteLets(e.Predicate, compBindings, copySeen(seen))
		}
		return comp
	default:
		return expr
	}
}

func copySeen(seen map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k, v := range seen {
		out[k] = v
	}
	return out
}

func primedIdentifier(expr Expr) (string, bool) {
	unary, ok := expr.(*UnaryExpr)
	if !ok || unary.Op != "'" {
		return "", false
	}
	return unprimedIdentifier(unary.Expr)
}

func unprimedIdentifier(expr Expr) (string, bool) {
	ident, ok := expr.(*IdentExpr)
	if !ok {
		return "", false
	}
	return ident.Name, true
}

func isUnaryJunctionOp(op string) bool {
	return op == "/\\" || op == "\\/"
}

func cloneState(st State) State {
	out := State{}
	for k, v := range st {
		out[k] = v
	}
	return out
}

func mergeState(base State, overlay State) State {
	out := cloneState(base)
	for k, v := range overlay {
		out[k] = v
	}
	return out
}

func stateKey(st State, vars []string) string {
	var b strings.Builder
	for i, v := range vars {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(v)
		b.WriteByte('=')
		b.WriteString(strconv.Itoa(st[v]))
	}
	return b.String()
}

func stateKeyWithView(st State, vars []string, defs map[string]*Definition, views []string, constants State) (string, error) {
	if len(views) == 0 {
		return stateKey(st, vars), nil
	}
	env := mergeState(constants, st)
	parts := make([]string, 0, len(views))
	for _, name := range views {
		def := defs[name]
		if def == nil {
			return "", fmt.Errorf("VIEW operator %s not found", name)
		}
		value, err := evalInt(def.Expr, env)
		if err != nil {
			return "", fmt.Errorf("VIEW %s: %w", name, err)
		}
		parts = append(parts, fmt.Sprintf("%s=%d", name, value))
	}
	sort.Strings(parts)
	return strings.Join(parts, ";"), nil
}

func reconstructTrace(key string, parent map[string]string, states map[string]State) []State {
	var rev []State
	for key != "" {
		rev = append(rev, cloneState(states[key]))
		key = parent[key]
	}
	trace := make([]State, len(rev))
	for i := range rev {
		trace[len(rev)-1-i] = rev[i]
	}
	return trace
}
