package tlago

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"path/filepath"
	"strconv"
	"strings"
)

const apalacheJSONDescription = "https://apalache-mc.org/docs/adr/005adr-json.html"

type ApalacheIROptions struct {
	IncludeSource bool
}

func ApalacheIRJSONSource(file, source string, opts ApalacheIROptions) ([]byte, Diagnostics) {
	spec, diags := CheckSanySource(file, source)
	if diags.HasErrors() {
		return nil, diags
	}
	data, irDiags := ApalacheIRJSON(spec, opts)
	diags = append(diags, irDiags...)
	return data, diags
}

func ApalacheIRJSON(spec *Spec, opts ApalacheIROptions) ([]byte, Diagnostics) {
	root, diags := apalacheIRRoot(spec, opts)
	if diags.HasErrors() {
		return nil, diags
	}
	data, err := marshalCanonicalJSON(root)
	if err != nil {
		diags = append(diags, errorAt(Position{}, "E6000", "cannot serialize ApalacheIR JSON: %v", err))
		return nil, diags
	}
	return data, diags
}

type apalacheRootJSON struct {
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Description string               `json:"description"`
	Modules     []apalacheModuleJSON `json:"modules"`
}

type apalacheModuleJSON struct {
	Kind         string        `json:"kind"`
	Name         string        `json:"name"`
	Declarations []interface{} `json:"declarations"`
}

type apalacheDeclItem struct {
	Name   string
	Params []string
	Body   Expr
	JSON   interface{}
}

type apalacheSourcePositionJSON struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type apalacheSourceJSON struct {
	Filename string                     `json:"filename"`
	From     apalacheSourcePositionJSON `json:"from"`
	To       apalacheSourcePositionJSON `json:"to"`
}

type apalacheConstDeclJSON struct {
	Source *apalacheSourceJSON `json:"source,omitempty"`
	Type   string              `json:"type"`
	Kind   string              `json:"kind"`
	Name   string              `json:"name"`
}

type apalacheVarDeclJSON struct {
	Source *apalacheSourceJSON `json:"source,omitempty"`
	Type   string              `json:"type"`
	Kind   string              `json:"kind"`
	Name   string              `json:"name"`
}

type apalacheOperParamJSON struct {
	Kind  string `json:"kind"`
	Name  string `json:"name"`
	Arity int    `json:"arity"`
}

type apalacheOperDeclJSON struct {
	Source       interface{}             `json:"source,omitempty"`
	Type         string                  `json:"type"`
	Kind         string                  `json:"kind"`
	Name         string                  `json:"name"`
	FormalParams []apalacheOperParamJSON `json:"formalParams"`
	IsRecursive  bool                    `json:"isRecursive"`
	Body         interface{}             `json:"body"`
}

var apalacheUnknownSourcePosition = Position{File: "UNKNOWN"}

type apalacheAssumeDeclJSON struct {
	Source *apalacheSourceJSON `json:"source,omitempty"`
	Type   string              `json:"type"`
	Kind   string              `json:"kind"`
	Body   interface{}         `json:"body"`
	Name   string              `json:"name,omitempty"`
}

type apalacheTheoremDeclJSON struct {
	Source *apalacheSourceJSON `json:"source,omitempty"`
	Type   string              `json:"type"`
	Kind   string              `json:"kind"`
	Name   string              `json:"name"`
	Body   interface{}         `json:"body"`
}

type apalacheNameExJSON struct {
	Source *apalacheSourceJSON `json:"source,omitempty"`
	Type   string              `json:"type"`
	Kind   string              `json:"kind"`
	Name   string              `json:"name"`
}

type apalacheValExJSON struct {
	Source *apalacheSourceJSON `json:"source,omitempty"`
	Type   string              `json:"type"`
	Kind   string              `json:"kind"`
	Value  interface{}         `json:"value"`
}

type apalacheTlaValueJSON struct {
	Kind  string      `json:"kind"`
	Value interface{} `json:"value,omitempty"`
}

type apalacheOperExJSON struct {
	Source *apalacheSourceJSON `json:"source,omitempty"`
	Type   string              `json:"type"`
	Kind   string              `json:"kind"`
	Oper   string              `json:"oper"`
	Args   []interface{}       `json:"args"`
}

type apalacheLetInExJSON struct {
	Source *apalacheSourceJSON `json:"source,omitempty"`
	Type   string              `json:"type"`
	Kind   string              `json:"kind"`
	Body   interface{}         `json:"body"`
	Decls  []interface{}       `json:"decls"`
}

type apalacheExprContext struct {
	Module     *Module
	Spec       *Spec
	Operators  map[string]bool
	Recursives map[string]bool
	Locals     map[string]bool
	AtValue    Expr
}

func newApalacheExprContext(mod *Module, spec *Spec) apalacheExprContext {
	ctx := apalacheExprContext{
		Module:     mod,
		Spec:       spec,
		Operators:  map[string]bool{},
		Recursives: map[string]bool{},
		Locals:     map[string]bool{},
	}
	if mod == nil {
		return ctx
	}
	for _, def := range mod.Definitions {
		if !def.TheoremLike {
			ctx.Operators[def.Name] = true
		}
	}
	for _, decl := range mod.Recursives {
		for _, name := range decl.Names {
			ctx.Recursives[name] = true
		}
	}
	return ctx
}

func (ctx apalacheExprContext) withLocals(names ...string) apalacheExprContext {
	next := apalacheExprContext{
		Module:     ctx.Module,
		Spec:       ctx.Spec,
		Operators:  ctx.Operators,
		Recursives: ctx.Recursives,
		Locals:     copyBoolMap(ctx.Locals),
		AtValue:    ctx.AtValue,
	}
	for _, name := range names {
		next.Locals[name] = true
	}
	return next
}

func (ctx apalacheExprContext) withOperators(names ...string) apalacheExprContext {
	next := apalacheExprContext{
		Module:     ctx.Module,
		Spec:       ctx.Spec,
		Operators:  copyBoolMap(ctx.Operators),
		Recursives: ctx.Recursives,
		Locals:     ctx.Locals,
		AtValue:    ctx.AtValue,
	}
	for _, name := range names {
		next.Operators[name] = true
	}
	return next
}

func (ctx apalacheExprContext) withRecursiveOperators(names ...string) apalacheExprContext {
	next := apalacheExprContext{
		Module:     ctx.Module,
		Spec:       ctx.Spec,
		Operators:  copyBoolMap(ctx.Operators),
		Recursives: copyBoolMap(ctx.Recursives),
		Locals:     ctx.Locals,
		AtValue:    ctx.AtValue,
	}
	for _, name := range names {
		next.Operators[name] = true
		next.Recursives[name] = true
	}
	return next
}

func (ctx apalacheExprContext) resolvesToOperator(name string) bool {
	return ctx.Operators[name] && !ctx.Locals[name]
}

func (ctx apalacheExprContext) withAtValue(expr Expr) apalacheExprContext {
	ctx.AtValue = expr
	return ctx
}

func apalacheIRRoot(spec *Spec, opts ApalacheIROptions) (apalacheRootJSON, Diagnostics) {
	root := apalacheRootJSON{
		Name:        "ApalacheIR",
		Version:     "1.0",
		Description: apalacheJSONDescription,
	}
	if spec == nil || spec.Root == nil {
		return root, nil
	}
	mod, diags := apalacheModule(spec.Root, spec, opts)
	root.Modules = append(root.Modules, mod)
	return root, diags
}

func apalacheModule(mod *Module, spec *Spec, opts ApalacheIROptions) (apalacheModuleJSON, Diagnostics) {
	out := apalacheModuleJSON{Kind: "TlaModule", Name: mod.Name}
	var diags Diagnostics
	var items []apalacheDeclItem
	extendedModules := apalacheExtendedModules(mod, spec)
	ctx := newApalacheExprContext(mod, spec)
	for _, imported := range extendedModules {
		apalacheRegisterModuleOperators(ctx, imported, false)
	}
	instanceDefs := apalacheInstanceDefinitions(mod, spec)
	for _, def := range instanceDefs {
		ctx.Operators[def.Name] = true
	}
	for _, imported := range extendedModules {
		diags = append(diags, apalacheAppendModuleDeclarations(&items, imported, opts, ctx, false)...)
	}
	diags = append(diags, apalacheAppendModuleDeclarations(&items, mod, opts, ctx, true)...)
	for _, def := range instanceDefs {
		decl, defDiags := apalacheOperDecl(def, opts, ctx)
		diags = append(diags, defDiags...)
		if !defDiags.HasErrors() {
			items = append(items, apalacheDeclItem{
				Name:   def.Name,
				Params: append([]string(nil), def.Params...),
				Body:   def.Expr,
				JSON:   decl,
			})
		}
	}
	out.Declarations = sortedApalacheDeclarations(items)
	return out, diags
}

func apalacheAppendModuleDeclarations(items *[]apalacheDeclItem, mod *Module, opts ApalacheIROptions, ctx apalacheExprContext, includeLocal bool) Diagnostics {
	var diags Diagnostics
	for _, decl := range mod.Declarations {
		if decl.Kind != ConstantDecl {
			continue
		}
		for _, name := range decl.Names {
			*items = append(*items, apalacheDeclItem{Name: name, JSON: apalacheConstDeclJSON{
				Source: apalacheSource(declarationNamePosition(decl, name), opts),
				Type:   "Untyped",
				Kind:   "TlaConstDecl",
				Name:   name,
			}})
		}
	}
	for _, decl := range mod.Declarations {
		if decl.Kind != VariableDecl {
			continue
		}
		for _, name := range decl.Names {
			*items = append(*items, apalacheDeclItem{Name: name, JSON: apalacheVarDeclJSON{
				Source: apalacheSource(declarationNamePosition(decl, name), opts),
				Type:   "Untyped",
				Kind:   "TlaVarDecl",
				Name:   name,
			}})
		}
	}
	for _, def := range mod.Definitions {
		if def.TheoremLike || (!includeLocal && def.Local) {
			continue
		}
		if !includeLocal && apalacheImportedDefinitionIsBuiltinAlias(mod.Name, def.Name) {
			continue
		}
		decl, defDiags := apalacheOperDecl(def, opts, ctx)
		diags = append(diags, defDiags...)
		if !defDiags.HasErrors() {
			*items = append(*items, apalacheDeclItem{
				Name:   def.Name,
				Params: append([]string(nil), def.Params...),
				Body:   def.Expr,
				JSON:   decl,
			})
		}
	}
	for _, assume := range mod.Assumptions {
		ex, exprDiags := apalacheExprIn(assume.Expr, opts, ctx)
		diags = append(diags, exprDiags...)
		if exprDiags.HasErrors() {
			continue
		}
		decl := apalacheAssumeDeclJSON{
			Source: apalacheSource(assume.SourcePosition(), opts),
			Type:   "Untyped",
			Kind:   "TlaAssumeDecl",
			Body:   ex,
			Name:   assume.Name,
		}
		*items = append(*items, apalacheDeclItem{Name: assume.Name, Body: assume.Expr, JSON: decl})
	}
	return diags
}

func apalacheImportedDefinitionIsBuiltinAlias(moduleName, defName string) bool {
	switch moduleName {
	case "Naturals":
		switch defName {
		case "Nat", "+", "-", "*", "^", "<", ">", "\\leq", "\\geq", "%", "\\div", "..":
			return true
		}
	case "Sequences":
		switch defName {
		case "Seq", "Len", "\\o", "Append", "Head", "SubSeq", "Tail":
			return true
		}
	case "FiniteSets":
		switch defName {
		case "IsFiniteSet", "Cardinality":
			return true
		}
	case "__apalache_folds":
		switch defName {
		case "__ApalacheFoldSet", "__ApalacheFoldSeq", "__ApalacheMkSeq":
			return true
		}
	}
	return false
}

func apalacheRegisterModuleOperators(ctx apalacheExprContext, mod *Module, includeLocal bool) {
	if mod == nil {
		return
	}
	for _, def := range mod.Definitions {
		if def.TheoremLike || (!includeLocal && def.Local) {
			continue
		}
		ctx.Operators[def.Name] = true
	}
	for _, decl := range mod.Recursives {
		for _, name := range decl.Names {
			ctx.Operators[name] = true
			ctx.Recursives[name] = true
		}
	}
}

func apalacheExtendedModules(mod *Module, spec *Spec) []*Module {
	if mod == nil || spec == nil {
		return nil
	}
	seen := map[string]bool{}
	if mod.Name != "" {
		seen[mod.Name] = true
	}
	var out []*Module
	var visit func(string)
	visit = func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		dep := spec.Modules[name]
		if dep == nil || isEmbeddedStandardModule(dep) {
			return
		}
		for _, parent := range dep.Extends {
			visit(parent)
		}
		out = append(out, dep)
	}
	for _, name := range mod.Extends {
		visit(name)
	}
	return out
}

func apalacheInstanceDefinitions(mod *Module, spec *Spec) []Definition {
	if mod == nil || spec == nil {
		return nil
	}
	var out []Definition
	ownDefinitions := map[string]bool{}
	for _, def := range mod.Definitions {
		ownDefinitions[def.Name] = true
	}
	for _, inst := range mod.Instances {
		if inst.Local {
			continue
		}
		target := spec.Modules[inst.Module]
		if target == nil {
			continue
		}
		qualifier := inst.Name
		for i := range target.Definitions {
			if target.Definitions[i].Local || target.Definitions[i].TheoremLike {
				continue
			}
			def := apalacheInstantiatedDefinition(&target.Definitions[i], inst, mod, target, spec)
			if def == nil {
				continue
			}
			if qualifier != "" {
				def.Name = qualifier + "!" + def.Name
				out = append(out, *def)
				continue
			}
			if ownDefinitions[def.Name] {
				continue
			}
			out = append(out, *def)
		}
	}
	return out
}

func apalacheLetInstanceDefinitions(instances []Instance, ctx apalacheExprContext) []Definition {
	if len(instances) == 0 || ctx.Spec == nil {
		return nil
	}
	var out []Definition
	for _, inst := range instances {
		target := ctx.Spec.Modules[inst.Module]
		if target == nil {
			continue
		}
		qualifier := inst.qualifier()
		for _, source := range apalacheInstanceDefinitionModules(target, ctx.Spec) {
			if source == nil {
				continue
			}
			for i := range source.Definitions {
				if source.Definitions[i].Local || source.Definitions[i].TheoremLike {
					continue
				}
				def := apalacheInstantiatedDefinition(&source.Definitions[i], inst, ctx.Module, source, ctx.Spec)
				if def == nil {
					continue
				}
				if qualifier != "" {
					def.Name = qualifier + "!" + def.Name
				}
				out = append(out, *def)
			}
		}
	}
	return out
}

func apalacheLetInstanceSourceOperatorNames(instances []Instance, ctx apalacheExprContext) []string {
	if len(instances) == 0 || ctx.Spec == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, inst := range instances {
		target := ctx.Spec.Modules[inst.Module]
		if target == nil {
			continue
		}
		for _, source := range apalacheInstanceDefinitionModules(target, ctx.Spec) {
			if source == nil {
				continue
			}
			for _, def := range source.Definitions {
				if def.Local || def.TheoremLike || def.Name == "" || seen[def.Name] {
					continue
				}
				seen[def.Name] = true
				out = append(out, def.Name)
			}
		}
	}
	return out
}

func apalacheLetInstanceRecursiveNames(defs []Definition, instances []Instance, ctx apalacheExprContext) []string {
	if len(defs) == 0 || len(instances) == 0 || ctx.Spec == nil {
		return nil
	}
	recursiveSources := map[string]bool{}
	for _, inst := range instances {
		target := ctx.Spec.Modules[inst.Module]
		if target == nil {
			continue
		}
		for _, source := range apalacheInstanceDefinitionModules(target, ctx.Spec) {
			if source == nil {
				continue
			}
			for _, decl := range source.Recursives {
				for _, name := range decl.Names {
					recursiveSources[inst.qualifier()+"!"+name] = true
					recursiveSources[name] = true
				}
			}
		}
	}
	var out []string
	for _, def := range defs {
		if recursiveSources[def.Name] {
			out = append(out, def.Name)
		}
	}
	return out
}

func apalacheInstanceDefinitionModules(target *Module, spec *Spec) []*Module {
	if target == nil || spec == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []*Module
	var visit func(*Module)
	visit = func(mod *Module) {
		if mod == nil || seen[mod.Name] {
			return
		}
		seen[mod.Name] = true
		for _, parent := range mod.Extends {
			visit(spec.Modules[parent])
		}
		out = append(out, mod)
	}
	visit(target)
	return out
}

func apalacheInstantiatedDefinition(def *Definition, inst Instance, mod, target *Module, spec *Spec) *Definition {
	if def == nil {
		return nil
	}
	next := *def
	next.Source = inst.SourcePosition()
	replacements := apalacheScopedSubstitutions(apalacheInstanceReplacements(mod, target, spec, inst), def.Params)
	if len(replacements) > 0 {
		next.Expr = apalacheSubstituteParamsWithSharedUseSource(def.Expr, replacements)
	}
	return &next
}

type apalacheImplicitReplacement struct {
	Expr  Expr
	Arity int
}

func apalacheInstanceReplacements(mod, target *Module, spec *Spec, inst Instance) map[string]Expr {
	targets := moduleOwnSubstitutionTargets(target)
	if len(targets) == 0 {
		return nil
	}
	replacements := map[string]Expr{}
	for _, subst := range instanceSubstitutions(inst) {
		if subst.Name == "" || subst.Expr == nil {
			continue
		}
		replacements[subst.Name] = subst.Expr
	}
	implicit := apalacheImplicitInstanceReplacements(mod, spec)
	for name, target := range targets {
		if _, explicit := replacements[name]; explicit {
			continue
		}
		replacement, ok := implicit[name]
		if !ok || replacement.Arity != target.Arity {
			continue
		}
		replacements[name] = replacement.Expr
	}
	if len(replacements) == 0 {
		return nil
	}
	return replacements
}

func apalacheImplicitInstanceReplacements(mod *Module, spec *Spec) map[string]apalacheImplicitReplacement {
	replacements := map[string]apalacheImplicitReplacement{}
	var collect func(*Module)
	collect = func(cur *Module) {
		if cur == nil {
			return
		}
		for _, decl := range cur.Declarations {
			for _, name := range decl.Names {
				if _, exists := replacements[name]; exists {
					continue
				}
				arity := 0
				if decl.Kind == ConstantDecl {
					if declared, ok := declarationArity(decl, name); ok {
						arity = declared
					}
				}
				replacements[name] = apalacheImplicitReplacement{
					Expr:  &IdentExpr{Name: name, Pos: declarationNamePosition(decl, name)},
					Arity: arity,
				}
			}
		}
		for _, def := range cur.Definitions {
			if def.Local {
				continue
			}
			if _, exists := replacements[def.Name]; exists {
				continue
			}
			replacements[def.Name] = apalacheImplicitReplacement{
				Expr:  &IdentExpr{Name: def.Name, Pos: def.SourcePosition()},
				Arity: len(def.Params),
			}
		}
	}
	collect(mod)
	if spec != nil && mod != nil {
		for _, dep := range mod.Extends {
			collect(spec.Modules[dep])
		}
	}
	if len(replacements) == 0 {
		return nil
	}
	return replacements
}

func apalacheScopedSubstitutions(substitutions map[string]Expr, shadows []string) map[string]Expr {
	if len(substitutions) == 0 {
		return nil
	}
	out := make(map[string]Expr, len(substitutions))
	for name, replacement := range substitutions {
		out[name] = replacement
	}
	for _, shadow := range shadows {
		delete(out, shadow)
	}
	return out
}

func apalacheSubstituteParamsWithSharedUseSource(expr Expr, replacements map[string]Expr) Expr {
	if len(replacements) == 0 || expr == nil {
		return expr
	}
	finalPositions := map[string]Position{}
	apalacheCollectSubstitutionFinalPositions(expr, replacements, finalPositions)
	return apalacheSubstituteParamsAtSharedUseSource(expr, replacements, finalPositions)
}

func apalacheSubstituteParamsAtSharedUseSource(expr Expr, replacements map[string]Expr, finalPositions map[string]Position) Expr {
	if len(replacements) == 0 || expr == nil {
		return expr
	}
	switch e := expr.(type) {
	case *IdentExpr:
		if replacement, ok := replacements[e.Name]; ok {
			pos := finalPositions[e.Name]
			if pos.Line == 0 && pos.Column == 0 && pos.File == "" {
				pos = e.Pos
			}
			return apalacheExprAtPosition(replacement, pos)
		}
		return e
	case *LiteralExpr:
		return e
	case *UnaryExpr:
		return &UnaryExpr{Op: e.Op, Expr: apalacheSubstituteParamsAtSharedUseSource(e.Expr, replacements, finalPositions), Pos: e.Pos}
	case *BinaryExpr:
		return &BinaryExpr{Op: e.Op, Left: apalacheSubstituteParamsAtSharedUseSource(e.Left, replacements, finalPositions), Right: apalacheSubstituteParamsAtSharedUseSource(e.Right, replacements, finalPositions), Pos: e.Pos, JunctionList: e.JunctionList, SanyNary: e.SanyNary}
	case *CallExpr:
		call := &CallExpr{Callee: apalacheSubstituteParamsAtSharedUseSource(e.Callee, replacements, finalPositions), Pos: e.Pos}
		for _, arg := range e.Args {
			call.Args = append(call.Args, apalacheSubstituteParamsAtSharedUseSource(arg, replacements, finalPositions))
		}
		return call
	case *IfExpr:
		return &IfExpr{
			Cond: apalacheSubstituteParamsAtSharedUseSource(e.Cond, replacements, finalPositions),
			Then: apalacheSubstituteParamsAtSharedUseSource(e.Then, replacements, finalPositions),
			Else: apalacheSubstituteParamsAtSharedUseSource(e.Else, replacements, finalPositions),
			Pos:  e.Pos,
		}
	case *LetExpr:
		letReplacements := apalacheScopedSubstitutions(replacements, nil)
		let := &LetExpr{Recursives: append([]Declaration(nil), e.Recursives...), Instances: append([]Instance(nil), e.Instances...), Pos: e.Pos}
		for _, def := range e.Definitions {
			delete(letReplacements, def.Name)
		}
		for _, def := range e.Definitions {
			nextDef := def
			nextDef.Pos = apalacheUnknownSourcePosition
			nextDef.Source = apalacheUnknownSourcePosition
			defReplacements := apalacheScopedSubstitutions(letReplacements, def.Params)
			nextDef.Expr = apalacheSubstituteParamsAtSharedUseSource(def.Expr, defReplacements, finalPositions)
			let.Definitions = append(let.Definitions, nextDef)
		}
		let.Body = apalacheSubstituteParamsAtSharedUseSource(e.Body, letReplacements, finalPositions)
		return let
	case *QuantifierExpr:
		quantReplacements := apalacheScopedSubstitutions(replacements, []string{e.Var})
		return &QuantifierExpr{
			Kind:       e.Kind,
			Var:        e.Var,
			VarPos:     e.VarPos,
			Set:        apalacheSubstituteParamsAtSharedUseSource(e.Set, replacements, finalPositions),
			Body:       apalacheSubstituteParamsAtSharedUseSource(e.Body, quantReplacements, finalPositions),
			TupleBound: e.TupleBound,
			Pos:        e.Pos,
		}
	case *CaseExpr:
		caseExpr := &CaseExpr{Pos: e.Pos, OtherPos: e.OtherPos}
		for _, arm := range e.Arms {
			caseExpr.Arms = append(caseExpr.Arms, CaseArm{
				Test:  apalacheSubstituteParamsAtSharedUseSource(arm.Test, replacements, finalPositions),
				Value: apalacheSubstituteParamsAtSharedUseSource(arm.Value, replacements, finalPositions),
				Pos:   arm.Pos,
			})
		}
		if e.Other != nil {
			caseExpr.Other = apalacheSubstituteParamsAtSharedUseSource(e.Other, replacements, finalPositions)
		}
		return caseExpr
	case *ChooseExpr:
		chooseReplacements := apalacheScopedSubstitutions(replacements, e.boundNames())
		return &ChooseExpr{
			TupleVars: e.TupleVars,
			Var:       e.Var,
			VarPos:    e.VarPos,
			Set:       apalacheSubstituteParamsAtSharedUseSource(e.Set, replacements, finalPositions),
			Body:      apalacheSubstituteParamsAtSharedUseSource(e.Body, chooseReplacements, finalPositions),
			Pos:       e.Pos,
		}
	case *TupleExpr:
		tuple := &TupleExpr{Pos: e.Pos}
		for _, elem := range e.Elems {
			tuple.Elems = append(tuple.Elems, apalacheSubstituteParamsAtSharedUseSource(elem, replacements, finalPositions))
		}
		return tuple
	case *SetExpr:
		set := &SetExpr{Pos: e.Pos}
		for _, elem := range e.Elems {
			set.Elems = append(set.Elems, apalacheSubstituteParamsAtSharedUseSource(elem, replacements, finalPositions))
		}
		return set
	case *RecordExpr:
		record := &RecordExpr{Pos: e.Pos}
		for _, field := range e.Fields {
			record.Fields = append(record.Fields, RecordField{
				Name:   field.Name,
				Value:  apalacheSubstituteParamsAtSharedUseSource(field.Value, replacements, finalPositions),
				Pos:    field.Pos,
				Source: field.Source,
			})
		}
		return record
	case *RecordComponentExpr:
		return &RecordComponentExpr{Record: apalacheSubstituteParamsAtSharedUseSource(e.Record, replacements, finalPositions), Field: e.Field, FieldPos: e.FieldPos, Pos: e.Pos}
	case *RecordSetExpr:
		record := &RecordSetExpr{Pos: e.Pos}
		for _, field := range e.Fields {
			record.Fields = append(record.Fields, RecordSetField{
				Name:   field.Name,
				Set:    apalacheSubstituteParamsAtSharedUseSource(field.Set, replacements, finalPositions),
				Pos:    field.Pos,
				Source: field.Source,
			})
		}
		return record
	case *FunctionExpr:
		fnReplacements := apalacheScopedSubstitutions(replacements, nil)
		fn := &FunctionExpr{Pos: e.Pos}
		for _, bound := range e.Bounds {
			delete(fnReplacements, bound.Name)
			fn.Bounds = append(fn.Bounds, BoundVar{
				Name: bound.Name,
				Set:  apalacheSubstituteParamsAtSharedUseSource(bound.Set, replacements, finalPositions),
				Pos:  bound.Pos,
			})
		}
		fn.Body = apalacheSubstituteParamsAtSharedUseSource(e.Body, fnReplacements, finalPositions)
		return fn
	case *FunctionAppExpr:
		app := &FunctionAppExpr{Function: apalacheSubstituteParamsAtSharedUseSource(e.Function, replacements, finalPositions), Pos: e.Pos}
		for _, arg := range e.Args {
			app.Args = append(app.Args, apalacheSubstituteParamsAtSharedUseSource(arg, replacements, finalPositions))
		}
		return app
	case *ExceptExpr:
		except := &ExceptExpr{Base: apalacheSubstituteParamsAtSharedUseSource(e.Base, replacements, finalPositions), Pos: e.Pos}
		for _, spec := range e.Specs {
			nextSpec := ExceptSpec{Value: apalacheSubstituteParamsAtSharedUseSource(spec.Value, replacements, finalPositions), Pos: spec.Pos}
			for _, component := range spec.Components {
				nextComponent := ExceptComponent{Field: component.Field, FieldPos: component.FieldPos, Pos: component.Pos}
				for _, index := range component.Indices {
					nextComponent.Indices = append(nextComponent.Indices, apalacheSubstituteParamsAtSharedUseSource(index, replacements, finalPositions))
				}
				nextSpec.Components = append(nextSpec.Components, nextComponent)
			}
			except.Specs = append(except.Specs, nextSpec)
		}
		return except
	case *LabelExpr:
		return &LabelExpr{Name: e.Name, Params: append([]string(nil), e.Params...), Body: apalacheSubstituteParamsAtSharedUseSource(e.Body, replacements, finalPositions), Pos: e.Pos}
	case *ActionExpr:
		return &ActionExpr{
			Kind:      e.Kind,
			Action:    apalacheSubstituteParamsAtSharedUseSource(e.Action, replacements, finalPositions),
			Subscript: apalacheSubstituteParamsAtSharedUseSource(e.Subscript, replacements, finalPositions),
			Pos:       e.Pos,
		}
	case *FairnessExpr:
		return &FairnessExpr{
			Kind:      e.Kind,
			Subscript: apalacheSubstituteParamsAtSharedUseSource(e.Subscript, replacements, finalPositions),
			Action:    apalacheSubstituteParamsAtSharedUseSource(e.Action, replacements, finalPositions),
			Pos:       e.Pos,
		}
	case *FunctionSetExpr:
		return &FunctionSetExpr{
			Domain: apalacheSubstituteParamsAtSharedUseSource(e.Domain, replacements, finalPositions),
			Range:  apalacheSubstituteParamsAtSharedUseSource(e.Range, replacements, finalPositions),
			Pos:    e.Pos,
		}
	case *SetComprehensionExpr:
		compReplacements := apalacheScopedSubstitutions(replacements, nil)
		comp := &SetComprehensionExpr{Pos: e.Pos}
		for _, bound := range e.Bounds {
			delete(compReplacements, bound.Name)
			comp.Bounds = append(comp.Bounds, BoundVar{
				Name: bound.Name,
				Set:  apalacheSubstituteParamsAtSharedUseSource(bound.Set, replacements, finalPositions),
				Pos:  bound.Pos,
			})
		}
		comp.Element = apalacheSubstituteParamsAtSharedUseSource(e.Element, compReplacements, finalPositions)
		if e.Predicate != nil {
			comp.Predicate = apalacheSubstituteParamsAtSharedUseSource(e.Predicate, compReplacements, finalPositions)
		}
		return comp
	default:
		return expr
	}
}

func apalacheCollectSubstitutionFinalPositions(expr Expr, replacements map[string]Expr, finalPositions map[string]Position) {
	if len(replacements) == 0 || expr == nil {
		return
	}
	switch e := expr.(type) {
	case *IdentExpr:
		if _, ok := replacements[e.Name]; ok {
			finalPositions[e.Name] = e.Pos
		}
	case *LiteralExpr:
	case *UnaryExpr:
		apalacheCollectSubstitutionFinalPositions(e.Expr, replacements, finalPositions)
	case *BinaryExpr:
		apalacheCollectSubstitutionFinalPositions(e.Left, replacements, finalPositions)
		apalacheCollectSubstitutionFinalPositions(e.Right, replacements, finalPositions)
	case *CallExpr:
		apalacheCollectSubstitutionFinalPositions(e.Callee, replacements, finalPositions)
		for _, arg := range e.Args {
			apalacheCollectSubstitutionFinalPositions(arg, replacements, finalPositions)
		}
	case *IfExpr:
		apalacheCollectSubstitutionFinalPositions(e.Cond, replacements, finalPositions)
		apalacheCollectSubstitutionFinalPositions(e.Then, replacements, finalPositions)
		apalacheCollectSubstitutionFinalPositions(e.Else, replacements, finalPositions)
	case *LetExpr:
		letReplacements := apalacheScopedSubstitutions(replacements, nil)
		for _, def := range e.Definitions {
			delete(letReplacements, def.Name)
		}
		apalacheCollectSubstitutionFinalPositions(e.Body, letReplacements, finalPositions)
		for _, def := range e.Definitions {
			defReplacements := apalacheScopedSubstitutions(letReplacements, def.Params)
			apalacheCollectSubstitutionFinalPositions(def.Expr, defReplacements, finalPositions)
		}
	case *QuantifierExpr:
		quantReplacements := apalacheScopedSubstitutions(replacements, []string{e.Var})
		apalacheCollectSubstitutionFinalPositions(e.Set, replacements, finalPositions)
		apalacheCollectSubstitutionFinalPositions(e.Body, quantReplacements, finalPositions)
	case *CaseExpr:
		for _, arm := range e.Arms {
			apalacheCollectSubstitutionFinalPositions(arm.Test, replacements, finalPositions)
			apalacheCollectSubstitutionFinalPositions(arm.Value, replacements, finalPositions)
		}
		apalacheCollectSubstitutionFinalPositions(e.Other, replacements, finalPositions)
	case *ChooseExpr:
		chooseReplacements := apalacheScopedSubstitutions(replacements, e.boundNames())
		apalacheCollectSubstitutionFinalPositions(e.Set, replacements, finalPositions)
		apalacheCollectSubstitutionFinalPositions(e.Body, chooseReplacements, finalPositions)
	case *TupleExpr:
		for _, elem := range e.Elems {
			apalacheCollectSubstitutionFinalPositions(elem, replacements, finalPositions)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			apalacheCollectSubstitutionFinalPositions(elem, replacements, finalPositions)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			apalacheCollectSubstitutionFinalPositions(field.Value, replacements, finalPositions)
		}
	case *RecordComponentExpr:
		apalacheCollectSubstitutionFinalPositions(e.Record, replacements, finalPositions)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			apalacheCollectSubstitutionFinalPositions(field.Set, replacements, finalPositions)
		}
	case *FunctionExpr:
		fnReplacements := apalacheScopedSubstitutions(replacements, nil)
		for _, bound := range e.Bounds {
			delete(fnReplacements, bound.Name)
			apalacheCollectSubstitutionFinalPositions(bound.Set, replacements, finalPositions)
		}
		apalacheCollectSubstitutionFinalPositions(e.Body, fnReplacements, finalPositions)
	case *FunctionAppExpr:
		apalacheCollectSubstitutionFinalPositions(e.Function, replacements, finalPositions)
		for _, arg := range e.Args {
			apalacheCollectSubstitutionFinalPositions(arg, replacements, finalPositions)
		}
	case *ExceptExpr:
		apalacheCollectSubstitutionFinalPositions(e.Base, replacements, finalPositions)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					apalacheCollectSubstitutionFinalPositions(index, replacements, finalPositions)
				}
			}
			apalacheCollectSubstitutionFinalPositions(spec.Value, replacements, finalPositions)
		}
	case *LabelExpr:
		apalacheCollectSubstitutionFinalPositions(e.Body, replacements, finalPositions)
	case *ActionExpr:
		apalacheCollectSubstitutionFinalPositions(e.Action, replacements, finalPositions)
		apalacheCollectSubstitutionFinalPositions(e.Subscript, replacements, finalPositions)
	case *FairnessExpr:
		apalacheCollectSubstitutionFinalPositions(e.Subscript, replacements, finalPositions)
		apalacheCollectSubstitutionFinalPositions(e.Action, replacements, finalPositions)
	case *FunctionSetExpr:
		apalacheCollectSubstitutionFinalPositions(e.Domain, replacements, finalPositions)
		apalacheCollectSubstitutionFinalPositions(e.Range, replacements, finalPositions)
	case *SetComprehensionExpr:
		compReplacements := apalacheScopedSubstitutions(replacements, nil)
		for _, bound := range e.Bounds {
			delete(compReplacements, bound.Name)
			apalacheCollectSubstitutionFinalPositions(bound.Set, replacements, finalPositions)
		}
		apalacheCollectSubstitutionFinalPositions(e.Element, compReplacements, finalPositions)
		apalacheCollectSubstitutionFinalPositions(e.Predicate, compReplacements, finalPositions)
	}
}

func apalacheExprAtPosition(expr Expr, pos Position) Expr {
	switch e := expr.(type) {
	case *IdentExpr:
		return &IdentExpr{Name: e.Name, Pos: pos}
	case *LiteralExpr:
		return &LiteralExpr{Kind: e.Kind, Value: e.Value, Pos: pos}
	case *UnaryExpr:
		return &UnaryExpr{Op: e.Op, Expr: e.Expr, Pos: pos}
	case *BinaryExpr:
		return &BinaryExpr{Op: e.Op, Left: e.Left, Right: e.Right, Pos: pos, JunctionList: e.JunctionList, SanyNary: e.SanyNary}
	case *CallExpr:
		return &CallExpr{Callee: e.Callee, Args: append([]Expr(nil), e.Args...), Pos: pos}
	case *IfExpr:
		return &IfExpr{Cond: e.Cond, Then: e.Then, Else: e.Else, Pos: pos}
	case *LetExpr:
		return &LetExpr{Recursives: append([]Declaration(nil), e.Recursives...), Definitions: append([]Definition(nil), e.Definitions...), Instances: append([]Instance(nil), e.Instances...), Body: e.Body, Pos: pos}
	case *QuantifierExpr:
		return &QuantifierExpr{Kind: e.Kind, Var: e.Var, VarPos: e.VarPos, Set: e.Set, Body: e.Body, TupleBound: e.TupleBound, Pos: pos}
	case *CaseExpr:
		return &CaseExpr{Arms: append([]CaseArm(nil), e.Arms...), Other: e.Other, OtherPos: e.OtherPos, Pos: pos}
	case *ChooseExpr:
		return &ChooseExpr{TupleVars: e.TupleVars, Var: e.Var, VarPos: e.VarPos, Set: e.Set, Body: e.Body, Pos: pos}
	case *TupleExpr:
		return &TupleExpr{Elems: append([]Expr(nil), e.Elems...), Pos: pos}
	case *SetExpr:
		return &SetExpr{Elems: append([]Expr(nil), e.Elems...), Pos: pos}
	case *RecordExpr:
		return &RecordExpr{Fields: append([]RecordField(nil), e.Fields...), Pos: pos}
	case *RecordComponentExpr:
		return &RecordComponentExpr{Record: e.Record, Field: e.Field, FieldPos: e.FieldPos, Pos: pos}
	case *RecordSetExpr:
		return &RecordSetExpr{Fields: append([]RecordSetField(nil), e.Fields...), Pos: pos}
	case *FunctionExpr:
		return &FunctionExpr{Bounds: append([]BoundVar(nil), e.Bounds...), Body: e.Body, Pos: pos}
	case *FunctionAppExpr:
		return &FunctionAppExpr{Function: e.Function, Args: append([]Expr(nil), e.Args...), Pos: pos}
	case *ExceptExpr:
		return &ExceptExpr{Base: e.Base, Specs: append([]ExceptSpec(nil), e.Specs...), Pos: pos}
	case *LabelExpr:
		return &LabelExpr{Name: e.Name, Params: append([]string(nil), e.Params...), Body: e.Body, Pos: pos}
	case *ActionExpr:
		return &ActionExpr{Kind: e.Kind, Action: e.Action, Subscript: e.Subscript, Pos: pos}
	case *FairnessExpr:
		return &FairnessExpr{Kind: e.Kind, Subscript: e.Subscript, Action: e.Action, Pos: pos}
	case *FunctionSetExpr:
		return &FunctionSetExpr{Domain: e.Domain, Range: e.Range, Pos: pos}
	case *SetComprehensionExpr:
		return &SetComprehensionExpr{Element: e.Element, Bounds: append([]BoundVar(nil), e.Bounds...), Predicate: e.Predicate, Pos: pos}
	default:
		return expr
	}
}

func sortedApalacheDeclarations(items []apalacheDeclItem) []interface{} {
	if len(items) == 0 {
		return nil
	}
	nameToIndex := map[string]int{}
	for i, item := range items {
		if item.Name != "" {
			nameToIndex[item.Name] = i
		}
	}
	deps := make([]map[int]bool, len(items))
	for i, item := range items {
		uses := map[string]bool{}
		locals := map[string]bool{}
		for _, param := range item.Params {
			locals[param] = true
		}
		collectApalacheExprUses(item.Body, locals, uses)
		deps[i] = map[int]bool{}
		for name := range uses {
			dep, ok := nameToIndex[name]
			if ok && dep != i {
				deps[i][dep] = true
			}
		}
	}
	emitted := make([]bool, len(items))
	out := make([]interface{}, 0, len(items))
	for len(out) < len(items) {
		var layer []int
		for i := range items {
			if emitted[i] || !apalacheDepsEmitted(deps[i], emitted) {
				continue
			}
			layer = append(layer, i)
		}
		if len(layer) == 0 {
			for i, item := range items {
				if !emitted[i] {
					out = append(out, item.JSON)
					emitted[i] = true
				}
			}
			continue
		}
		for _, i := range layer {
			out = append(out, items[i].JSON)
			emitted[i] = true
		}
	}
	return out
}

func apalacheDepsEmitted(deps map[int]bool, emitted []bool) bool {
	for dep := range deps {
		if !emitted[dep] {
			return false
		}
	}
	return true
}

func declarationNamePosition(decl Declaration, name string) Position {
	if decl.NamePositions != nil {
		if pos, ok := decl.NamePositions[name]; ok {
			return pos
		}
	}
	return decl.Pos
}

func collectApalacheExprUses(expr Expr, locals map[string]bool, uses map[string]bool) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *IdentExpr:
		if e.Name != "" && !locals[e.Name] {
			uses[e.Name] = true
		}
	case *LiteralExpr:
	case *UnaryExpr:
		collectApalacheExprUses(e.Expr, locals, uses)
	case *BinaryExpr:
		collectApalacheExprUses(e.Left, locals, uses)
		collectApalacheExprUses(e.Right, locals, uses)
	case *CallExpr:
		collectApalacheExprUses(e.Callee, locals, uses)
		for _, arg := range e.Args {
			collectApalacheExprUses(arg, locals, uses)
		}
	case *IfExpr:
		collectApalacheExprUses(e.Cond, locals, uses)
		collectApalacheExprUses(e.Then, locals, uses)
		collectApalacheExprUses(e.Else, locals, uses)
	case *LetExpr:
		letLocals := copyBoolMap(locals)
		for _, def := range e.Definitions {
			letLocals[def.Name] = true
		}
		for _, def := range e.Definitions {
			defLocals := copyBoolMap(letLocals)
			for _, param := range def.Params {
				defLocals[param] = true
			}
			collectApalacheExprUses(def.Expr, defLocals, uses)
		}
		collectApalacheExprUses(e.Body, letLocals, uses)
	case *QuantifierExpr:
		collectApalacheExprUses(e.Set, locals, uses)
		bodyLocals := copyBoolMap(locals)
		bodyLocals[e.Var] = true
		collectApalacheExprUses(e.Body, bodyLocals, uses)
	case *CaseExpr:
		for _, arm := range e.Arms {
			collectApalacheExprUses(arm.Test, locals, uses)
			collectApalacheExprUses(arm.Value, locals, uses)
		}
		collectApalacheExprUses(e.Other, locals, uses)
	case *ChooseExpr:
		collectApalacheExprUses(e.Set, locals, uses)
		bodyLocals := copyBoolMap(locals)
		for _, name := range e.boundNames() {
			bodyLocals[name] = true
		}
		collectApalacheExprUses(e.Body, bodyLocals, uses)
	case *TupleExpr:
		for _, elem := range e.Elems {
			collectApalacheExprUses(elem, locals, uses)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			collectApalacheExprUses(elem, locals, uses)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			collectApalacheExprUses(field.Value, locals, uses)
		}
	case *RecordComponentExpr:
		collectApalacheExprUses(e.Record, locals, uses)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			collectApalacheExprUses(field.Set, locals, uses)
		}
	case *FunctionExpr:
		fnLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			collectApalacheExprUses(bound.Set, locals, uses)
			fnLocals[bound.Name] = true
		}
		collectApalacheExprUses(e.Body, fnLocals, uses)
	case *FunctionAppExpr:
		collectApalacheExprUses(e.Function, locals, uses)
		for _, arg := range e.Args {
			collectApalacheExprUses(arg, locals, uses)
		}
	case *ExceptExpr:
		collectApalacheExprUses(e.Base, locals, uses)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					collectApalacheExprUses(index, locals, uses)
				}
			}
			collectApalacheExprUses(spec.Value, locals, uses)
		}
	case *LabelExpr:
		collectApalacheExprUses(e.Body, locals, uses)
	case *ActionExpr:
		collectApalacheExprUses(e.Action, locals, uses)
		collectApalacheExprUses(e.Subscript, locals, uses)
	case *FairnessExpr:
		collectApalacheExprUses(e.Subscript, locals, uses)
		collectApalacheExprUses(e.Action, locals, uses)
	case *FunctionSetExpr:
		collectApalacheExprUses(e.Domain, locals, uses)
		collectApalacheExprUses(e.Range, locals, uses)
	case *SetComprehensionExpr:
		compLocals := copyBoolMap(locals)
		for _, bound := range e.Bounds {
			collectApalacheExprUses(bound.Set, locals, uses)
			compLocals[bound.Name] = true
		}
		collectApalacheExprUses(e.Element, compLocals, uses)
		collectApalacheExprUses(e.Predicate, compLocals, uses)
	}
}

func apalacheOperDecl(def Definition, opts ApalacheIROptions, ctx apalacheExprContext) (apalacheOperDeclJSON, Diagnostics) {
	bodyCtx := ctx.withLocals(def.Params...)
	var body interface{}
	var diags Diagnostics
	if fcn, ok := def.Expr.(*FunctionExpr); ok && ctx.Recursives[def.Name] {
		body, diags = apalacheFunctionOper(fcn, "FUN_REC_CTOR", opts, bodyCtx)
	} else {
		body, diags = apalacheExprIn(def.Expr, opts, bodyCtx)
	}
	params := make([]apalacheOperParamJSON, 0, len(def.Params))
	for _, param := range def.Params {
		params = append(params, apalacheOperParamJSON{
			Kind:  "OperParam",
			Name:  param,
			Arity: def.ParamArities[param],
		})
	}
	return apalacheOperDeclJSON{
		Source:       apalacheDeclSource(def.SourcePosition(), opts),
		Type:         "Untyped",
		Kind:         "TlaOperDecl",
		Name:         def.Name,
		FormalParams: params,
		IsRecursive:  ctx.Recursives[def.Name],
		Body:         body,
	}, diags
}

func apalacheExpr(expr Expr, opts ApalacheIROptions) (interface{}, Diagnostics) {
	return apalacheExprIn(expr, opts, apalacheExprContext{})
}

func apalacheExprIn(expr Expr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	if expr == nil {
		return nil, Diagnostics{errorAt(Position{}, "E6001", "cannot translate nil expression to ApalacheIR")}
	}
	switch e := expr.(type) {
	case *IdentExpr:
		switch e.Name {
		case "@":
			if ctx.AtValue != nil {
				return apalacheExprIn(ctx.AtValue, opts, ctx.withAtValue(nil))
			}
			return apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: e.Name}, nil
		case "TRUE":
			return apalacheBool(true, e.Pos, opts), nil
		case "FALSE":
			return apalacheBool(false, e.Pos, opts), nil
		case "BOOLEAN":
			return apalachePredefSet("TlaBoolSet", e.Pos, opts), nil
		case "Int":
			return apalachePredefSet("TlaIntSet", e.Pos, opts), nil
		case "Nat":
			return apalachePredefSet("TlaNatSet", e.Pos, opts), nil
		case "STRING":
			return apalachePredefSet("TlaStrSet", e.Pos, opts), nil
		default:
			name := apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: e.Name}
			if ctx.resolvesToOperator(e.Name) {
				return apalacheOper("OPER_APP", []interface{}{name}, e.Pos, opts), nil
			}
			return apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: e.Name}, nil
		}
	case *LiteralExpr:
		return apalacheLiteral(e, opts)
	case *UnaryExpr:
		arg, diags := apalacheExprIn(e.Expr, opts, ctx)
		if diags.HasErrors() {
			return nil, diags
		}
		if e.Op == "/\\" || e.Op == "\\/" {
			oper := "AND"
			if e.Op == "\\/" {
				oper = "OR"
			}
			return apalacheOper(oper, []interface{}{arg}, e.Pos, opts), nil
		}
		oper, ok := apalacheUnaryOperator(e.Op)
		if !ok {
			return nil, Diagnostics{errorAt(e.Pos, "E6002", "unsupported ApalacheIR unary operator %q", e.Op)}
		}
		return apalacheOper(oper, []interface{}{arg}, e.Pos, opts), nil
	case *BinaryExpr:
		oper, ok := apalacheBinaryOperator(e.Op)
		if !ok {
			if ctx.resolvesToOperator(e.Op) {
				left, leftDiags := apalacheExprIn(e.Left, opts, ctx)
				right, rightDiags := apalacheExprIn(e.Right, opts, ctx)
				diags := append(leftDiags, rightDiags...)
				if diags.HasErrors() {
					return nil, diags
				}
				name := apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: e.Op}
				return apalacheOper("OPER_APP", []interface{}{name, left, right}, e.Pos, opts), nil
			}
			return nil, Diagnostics{errorAt(e.Pos, "E6003", "unsupported ApalacheIR binary operator %q", e.Op)}
		}
		if oper == "SET_TIMES" || ((oper == "AND" || oper == "OR") && e.JunctionList) {
			args, diags := apalacheFlattenBinaryOper(e, oper, opts, ctx)
			if diags.HasErrors() {
				return nil, diags
			}
			return apalacheOper(oper, args, e.Pos, opts), nil
		}
		left, leftDiags := apalacheExprIn(e.Left, opts, ctx)
		right, rightDiags := apalacheExprIn(e.Right, opts, ctx)
		diags := append(leftDiags, rightDiags...)
		if diags.HasErrors() {
			return nil, diags
		}
		return apalacheOper(oper, []interface{}{left, right}, e.Pos, opts), nil
	case *CallExpr:
		return apalacheCall(e, opts, ctx)
	case *IfExpr:
		cond, condDiags := apalacheExprIn(e.Cond, opts, ctx)
		thenExpr, thenDiags := apalacheExprIn(e.Then, opts, ctx)
		elseExpr, elseDiags := apalacheExprIn(e.Else, opts, ctx)
		diags := append(append(condDiags, thenDiags...), elseDiags...)
		if diags.HasErrors() {
			return nil, diags
		}
		return apalacheOper("IF_THEN_ELSE", []interface{}{cond, thenExpr, elseExpr}, e.Pos, opts), nil
	case *LetExpr:
		return apalacheLetIn(e, opts, ctx)
	case *QuantifierExpr:
		return apalacheQuantifier(e, opts, ctx)
	case *CaseExpr:
		return apalacheCase(e, opts, ctx)
	case *ChooseExpr:
		return apalacheChoose(e, opts, ctx)
	case *TupleExpr:
		return apalacheExprListOper("TUPLE", e.Elems, e.Pos, opts, ctx)
	case *SetExpr:
		return apalacheExprListOper("SET_ENUM", e.Elems, e.Pos, opts, ctx)
	case *RecordExpr:
		return apalacheRecord(e, opts, ctx)
	case *RecordComponentExpr:
		record, diags := apalacheExprIn(e.Record, opts, ctx)
		if diags.HasErrors() {
			return nil, diags
		}
		field := apalacheStringValue(e.Field, apalacheRecordFieldPosition(e.FieldPos, e.Pos), opts)
		return apalacheOper("FUN_APP", []interface{}{record, field}, e.Pos, opts), nil
	case *RecordSetExpr:
		return apalacheRecordSet(e, opts, ctx)
	case *FunctionExpr:
		return apalacheFunction(e, opts, ctx)
	case *FunctionAppExpr:
		return apalacheFunctionApplication(e, opts, ctx)
	case *ExceptExpr:
		return apalacheExcept(e, opts, ctx)
	case *LabelExpr:
		return apalacheLabel(e, opts, ctx)
	case *ActionExpr:
		return apalacheAction(e, opts, ctx)
	case *FairnessExpr:
		return apalacheFairness(e, opts, ctx)
	case *FunctionSetExpr:
		domain, domainDiags := apalacheExprIn(e.Domain, opts, ctx)
		rng, rangeDiags := apalacheExprIn(e.Range, opts, ctx)
		diags := append(domainDiags, rangeDiags...)
		if diags.HasErrors() {
			return nil, diags
		}
		return apalacheOper("FUN_SET", []interface{}{domain, rng}, e.Pos, opts), nil
	case *SetComprehensionExpr:
		return apalacheSetComprehension(e, opts, ctx)
	default:
		return nil, Diagnostics{errorAt(expr.Position(), "E6004", "unsupported ApalacheIR expression %T", expr)}
	}
}

func apalacheFlattenBinaryOper(expr Expr, oper string, opts ApalacheIROptions, ctx apalacheExprContext) ([]interface{}, Diagnostics) {
	if binary, ok := expr.(*BinaryExpr); ok && apalacheBinaryExprOper(binary) == oper && apalacheShouldFlattenBinaryOper(binary, oper) {
		left, leftDiags := apalacheFlattenBinaryOper(binary.Left, oper, opts, ctx)
		right, rightDiags := apalacheFlattenBinaryOper(binary.Right, oper, opts, ctx)
		return append(left, right...), append(leftDiags, rightDiags...)
	}
	one, diags := apalacheExprIn(expr, opts, ctx)
	if diags.HasErrors() {
		return nil, diags
	}
	return []interface{}{one}, nil
}

func apalacheBinaryExprOper(expr *BinaryExpr) string {
	if expr == nil {
		return ""
	}
	oper, _ := apalacheBinaryOperator(expr.Op)
	return oper
}

func apalacheShouldFlattenBinaryOper(expr *BinaryExpr, oper string) bool {
	if expr == nil {
		return false
	}
	if oper == "SET_TIMES" {
		return true
	}
	return expr.JunctionList
}

func apalacheLiteral(e *LiteralExpr, opts ApalacheIROptions) (interface{}, Diagnostics) {
	switch e.Kind {
	case "bool":
		switch strings.ToUpper(e.Value) {
		case "TRUE":
			return apalacheBool(true, e.Pos, opts), nil
		case "FALSE":
			return apalacheBool(false, e.Pos, opts), nil
		default:
			return nil, Diagnostics{errorAt(e.Pos, "E6005", "invalid boolean literal %q", e.Value)}
		}
	case "number":
		if strings.Contains(e.Value, ".") {
			return apalacheValExJSON{
				Source: apalacheSource(e.Pos, opts),
				Type:   "Untyped",
				Kind:   "ValEx",
				Value:  apalacheTlaValueJSON{Kind: "TlaDecimal", Value: e.Value},
			}, nil
		}
		n := new(big.Int)
		if _, ok := n.SetString(e.Value, 10); !ok {
			return nil, Diagnostics{errorAt(e.Pos, "E6006", "invalid integer literal %q", e.Value)}
		}
		var value interface{}
		if n.IsInt64() && n.Int64() >= math.MinInt32 && n.Int64() <= math.MaxInt32 {
			value = n.Int64()
		} else {
			value = map[string]string{"bigInt": n.String()}
		}
		return apalacheValExJSON{
			Source: apalacheSource(e.Pos, opts),
			Type:   "Untyped",
			Kind:   "ValEx",
			Value:  apalacheTlaValueJSON{Kind: "TlaInt", Value: value},
		}, nil
	case "string", "model":
		value := e.Value
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		return apalacheStringValue(value, e.Pos, opts), nil
	default:
		return nil, Diagnostics{errorAt(e.Pos, "E6007", "unsupported literal kind %q", e.Kind)}
	}
}

func apalacheCall(e *CallExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	args := make([]interface{}, 0, len(e.Args)+1)
	if ident, ok := e.Callee.(*IdentExpr); ok {
		if oper, ok := apalacheCallOperator(ident.Name); ok {
			for i, arg := range e.Args {
				argJSON, diags := apalacheExprIn(arg, opts, ctx)
				if apalacheBuiltinOperatorArgIsOperator(oper, i) {
					argJSON, diags = apalacheOperatorArgumentExpr(arg, opts, ctx)
				}
				if diags.HasErrors() {
					return nil, diags
				}
				args = append(args, argJSON)
			}
			return apalacheOper(oper, args, e.Pos, opts), nil
		}
		args = append(args, apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: ident.Name})
		for _, arg := range e.Args {
			argJSON, argDiags := apalacheExprIn(arg, opts, ctx)
			if argDiags.HasErrors() {
				return nil, argDiags
			}
			args = append(args, argJSON)
		}
		return apalacheOper("OPER_APP", args, e.Pos, opts), nil
	}
	callee, diags := apalacheExprIn(e.Callee, opts, ctx)
	if diags.HasErrors() {
		return nil, diags
	}
	args = append(args, callee)
	for _, arg := range e.Args {
		argJSON, argDiags := apalacheExprIn(arg, opts, ctx)
		diags = append(diags, argDiags...)
		args = append(args, argJSON)
	}
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper("OPER_APP", args, e.Pos, opts), nil
}

func apalacheBuiltinOperatorArgIsOperator(oper string, index int) bool {
	switch oper {
	case "Apalache!ApaFoldSet", "Apalache!ApaFoldSeqLeft":
		return index == 0
	case "Apalache!MkSeq":
		return index == 1
	default:
		return false
	}
}

func apalacheOperatorArgumentExpr(expr Expr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	if ident, ok := expr.(*IdentExpr); ok && ctx.resolvesToOperator(ident.Name) {
		return apalacheNameExJSON{Source: apalacheSource(ident.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: ident.Name}, nil
	}
	return apalacheExprIn(expr, opts, ctx)
}

func apalacheLetIn(e *LetExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	instanceDefs := apalacheLetInstanceDefinitions(e.Instances, ctx)
	localNames := make([]string, 0, len(e.Definitions)+len(instanceDefs))
	for _, def := range e.Definitions {
		localNames = append(localNames, def.Name)
	}
	for _, def := range instanceDefs {
		localNames = append(localNames, def.Name)
	}
	letCtx := ctx.withOperators(localNames...)
	var recursiveNames []string
	for _, decl := range e.Recursives {
		recursiveNames = append(recursiveNames, decl.Names...)
	}
	recursiveNames = append(recursiveNames, apalacheLetInstanceRecursiveNames(instanceDefs, e.Instances, ctx)...)
	letCtx = letCtx.withRecursiveOperators(recursiveNames...)
	body, diags := apalacheExprIn(e.Body, opts, letCtx)
	decls := make([]interface{}, 0, len(e.Definitions)+len(instanceDefs))
	for _, def := range e.Definitions {
		declCtx := letCtx
		if apalacheExprBindsName(def.Expr, def.Name) {
			declCtx = declCtx.withRecursiveOperators(def.Name)
		}
		decl, defDiags := apalacheOperDecl(def, opts, declCtx)
		diags = append(diags, defDiags...)
		decls = append(decls, decl)
	}
	instanceDeclCtx := letCtx.withOperators(apalacheLetInstanceSourceOperatorNames(e.Instances, ctx)...)
	var instanceItems []apalacheDeclItem
	for _, def := range instanceDefs {
		declCtx := instanceDeclCtx
		if apalacheExprBindsName(def.Expr, def.Name) {
			declCtx = declCtx.withRecursiveOperators(def.Name)
		}
		decl, defDiags := apalacheOperDecl(def, opts, declCtx)
		diags = append(diags, defDiags...)
		if !defDiags.HasErrors() {
			instanceItems = append(instanceItems, apalacheDeclItem{
				Name:   def.Name,
				Params: append([]string(nil), def.Params...),
				Body:   def.Expr,
				JSON:   decl,
			})
		}
	}
	decls = append(decls, sortedApalacheDeclarations(instanceItems)...)
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheLetInExJSON{
		Source: apalacheSource(e.Pos, opts),
		Type:   "Untyped",
		Kind:   "LetInEx",
		Body:   body,
		Decls:  decls,
	}, nil
}

func apalacheExprBindsName(expr Expr, name string) bool {
	if expr == nil || name == "" {
		return false
	}
	switch e := expr.(type) {
	case *UnaryExpr:
		return apalacheExprBindsName(e.Expr, name)
	case *BinaryExpr:
		return apalacheExprBindsName(e.Left, name) || apalacheExprBindsName(e.Right, name)
	case *CallExpr:
		if apalacheExprBindsName(e.Callee, name) {
			return true
		}
		for _, arg := range e.Args {
			if apalacheExprBindsName(arg, name) {
				return true
			}
		}
	case *IfExpr:
		return apalacheExprBindsName(e.Cond, name) || apalacheExprBindsName(e.Then, name) || apalacheExprBindsName(e.Else, name)
	case *LetExpr:
		for _, def := range e.Definitions {
			if apalacheExprBindsName(def.Expr, name) {
				return true
			}
		}
		return apalacheExprBindsName(e.Body, name)
	case *QuantifierExpr:
		return e.Var == name || apalacheExprBindsName(e.Set, name) || apalacheExprBindsName(e.Body, name)
	case *CaseExpr:
		for _, arm := range e.Arms {
			if apalacheExprBindsName(arm.Test, name) || apalacheExprBindsName(arm.Value, name) {
				return true
			}
		}
		return apalacheExprBindsName(e.Other, name)
	case *ChooseExpr:
		return e.bindsName(name) || apalacheExprBindsName(e.Set, name) || apalacheExprBindsName(e.Body, name)
	case *TupleExpr:
		for _, elem := range e.Elems {
			if apalacheExprBindsName(elem, name) {
				return true
			}
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			if apalacheExprBindsName(elem, name) {
				return true
			}
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			if apalacheExprBindsName(field.Value, name) {
				return true
			}
		}
	case *RecordComponentExpr:
		return apalacheExprBindsName(e.Record, name)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			if apalacheExprBindsName(field.Set, name) {
				return true
			}
		}
	case *FunctionExpr:
		for _, bound := range e.Bounds {
			if bound.Name == name || apalacheExprBindsName(bound.Set, name) {
				return true
			}
		}
		return apalacheExprBindsName(e.Body, name)
	case *FunctionAppExpr:
		if apalacheExprBindsName(e.Function, name) {
			return true
		}
		for _, arg := range e.Args {
			if apalacheExprBindsName(arg, name) {
				return true
			}
		}
	case *ExceptExpr:
		if apalacheExprBindsName(e.Base, name) {
			return true
		}
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					if apalacheExprBindsName(index, name) {
						return true
					}
				}
			}
			if apalacheExprBindsName(spec.Value, name) {
				return true
			}
		}
	case *LabelExpr:
		return apalacheExprBindsName(e.Body, name)
	case *ActionExpr:
		return apalacheExprBindsName(e.Action, name) || apalacheExprBindsName(e.Subscript, name)
	case *FairnessExpr:
		return apalacheExprBindsName(e.Subscript, name) || apalacheExprBindsName(e.Action, name)
	case *FunctionSetExpr:
		return apalacheExprBindsName(e.Domain, name) || apalacheExprBindsName(e.Range, name)
	case *SetComprehensionExpr:
		for _, bound := range e.Bounds {
			if bound.Name == name || apalacheExprBindsName(bound.Set, name) {
				return true
			}
		}
		return apalacheExprBindsName(e.Element, name) || apalacheExprBindsName(e.Predicate, name)
	}
	return false
}

func apalacheQuantifier(e *QuantifierExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	body, bodyDiags := apalacheExprIn(e.Body, opts, ctx.withLocals(e.Var))
	name := apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: e.Var}
	if e.Set == nil {
		if bodyDiags.HasErrors() {
			return nil, bodyDiags
		}
		oper := "FORALL2"
		switch e.Kind {
		case "\\E", "EXISTS":
			oper = "EXISTS2"
		case "\\EE", "TEMPORAL_EXISTS":
			oper = "TEMPORAL_EXISTS"
		case "\\AA", "TEMPORAL_FORALL":
			oper = "TEMPORAL_FORALL"
		}
		return apalacheOper(oper, []interface{}{name, body}, e.Pos, opts), nil
	}
	set, setDiags := apalacheExprIn(e.Set, opts, ctx)
	diags := append(setDiags, bodyDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	oper := "FORALL3"
	if e.Kind == "\\E" || e.Kind == "EXISTS" {
		oper = "EXISTS3"
	}
	return apalacheOper(oper, []interface{}{name, set, body}, e.Pos, opts), nil
}

func apalacheCase(e *CaseExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	var args []interface{}
	if e.Other != nil {
		other, diags := apalacheExprIn(e.Other, opts, ctx)
		if diags.HasErrors() {
			return nil, diags
		}
		args = append(args, other)
	}
	var diags Diagnostics
	for _, arm := range e.Arms {
		test, testDiags := apalacheExprIn(arm.Test, opts, ctx)
		value, valueDiags := apalacheExprIn(arm.Value, opts, ctx)
		diags = append(append(diags, testDiags...), valueDiags...)
		args = append(args, test, value)
	}
	if diags.HasErrors() {
		return nil, diags
	}
	if e.Other != nil {
		return apalacheOper("CASE_OTHER", args, e.Pos, opts), nil
	}
	return apalacheOper("CASE", args, e.Pos, opts), nil
}

func apalacheChoose(e *ChooseExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	if e.TupleVars != nil {
		return nil, Diagnostics{errorAt(e.Pos, "E6004", "unsupported ApalacheIR tuple CHOOSE")}
	}
	body, bodyDiags := apalacheExprIn(e.Body, opts, ctx.withLocals(e.Var))
	name := apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: e.Var}
	if e.Set == nil {
		if bodyDiags.HasErrors() {
			return nil, bodyDiags
		}
		return apalacheOper("CHOOSE2", []interface{}{name, body}, e.Pos, opts), nil
	}
	set, setDiags := apalacheExprIn(e.Set, opts, ctx)
	diags := append(setDiags, bodyDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper("CHOOSE3", []interface{}{name, set, body}, e.Pos, opts), nil
}

func apalacheExprListOper(oper string, exprs []Expr, pos Position, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	args := make([]interface{}, 0, len(exprs))
	var diags Diagnostics
	for _, expr := range exprs {
		arg, argDiags := apalacheExprIn(expr, opts, ctx)
		diags = append(diags, argDiags...)
		args = append(args, arg)
	}
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper(oper, args, pos, opts), nil
}

func apalacheRecord(e *RecordExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	args := make([]interface{}, 0, len(e.Fields)*2)
	var diags Diagnostics
	for _, field := range e.Fields {
		value, fieldDiags := apalacheExprIn(field.Value, opts, ctx)
		diags = append(diags, fieldDiags...)
		args = append(args, apalacheStringValue(field.Name, field.Pos, opts), value)
	}
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper("RECORD", args, e.Pos, opts), nil
}

func apalacheRecordSet(e *RecordSetExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	args := make([]interface{}, 0, len(e.Fields)*2)
	var diags Diagnostics
	for _, field := range e.Fields {
		set, fieldDiags := apalacheExprIn(field.Set, opts, ctx)
		diags = append(diags, fieldDiags...)
		args = append(args, apalacheStringValue(field.Name, field.Pos, opts), set)
	}
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper("RECORD_SET", args, e.Pos, opts), nil
}

func apalacheFunction(e *FunctionExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	return apalacheFunctionOper(e, "FUN_CTOR", opts, ctx)
}

func apalacheFunctionOper(e *FunctionExpr, oper string, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	localNames := make([]string, 0, len(e.Bounds))
	for _, bound := range e.Bounds {
		localNames = append(localNames, bound.Name)
	}
	body, diags := apalacheExprIn(e.Body, opts, ctx.withLocals(localNames...))
	args := []interface{}{body}
	for _, bound := range e.Bounds {
		set, setDiags := apalacheExprIn(bound.Set, opts, ctx)
		diags = append(diags, setDiags...)
		args = append(args, apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: bound.Name}, set)
	}
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper(oper, args, e.Pos, opts), nil
}

func apalacheFunctionApplication(e *FunctionAppExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	fn, diags := apalacheExprIn(e.Function, opts, ctx)
	var index interface{}
	switch len(e.Args) {
	case 0:
		return nil, Diagnostics{errorAt(e.Pos, "E6008", "function application has no arguments")}
	case 1:
		index, diags = apalacheExprIn(e.Args[0], opts, ctx)
	default:
		tuple, tupleDiags := apalacheExprListOper("TUPLE", e.Args, e.Pos, opts, ctx)
		diags = append(diags, tupleDiags...)
		index = tuple
	}
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper("FUN_APP", []interface{}{fn, index}, e.Pos, opts), nil
}

func apalacheExcept(e *ExceptExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	base, diags := apalacheExprIn(e.Base, opts, ctx)
	args := []interface{}{base}
	for _, spec := range e.Specs {
		var indexExprs []Expr
		for _, component := range spec.Components {
			indexExprs = append(indexExprs, apalacheExceptComponentPathExprs(component)...)
		}
		index, indexDiags := apalacheExprListOper("TUPLE", indexExprs, spec.Pos, opts, ctx)
		valueCtx := ctx.withAtValue(apalacheExceptOldValue(e.Base, spec))
		value, valueDiags := apalacheExprIn(spec.Value, opts, valueCtx)
		diags = append(append(diags, indexDiags...), valueDiags...)
		args = append(args, index, value)
	}
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper("EXCEPT", args, e.Pos, opts), nil
}

func apalacheExceptOldValue(base Expr, spec ExceptSpec) Expr {
	current := base
	for _, component := range spec.Components {
		for _, pathExpr := range apalacheExceptComponentPathExprs(component) {
			current = &FunctionAppExpr{Function: current, Args: []Expr{pathExpr}, Pos: spec.Pos}
		}
	}
	return current
}

func apalacheExceptComponentPathExprs(component ExceptComponent) []Expr {
	if component.Field != "" {
		return []Expr{&LiteralExpr{Kind: "string", Value: component.Field, Pos: apalacheRecordFieldPosition(component.FieldPos, component.Pos)}}
	}
	switch len(component.Indices) {
	case 0:
		return nil
	case 1:
		return []Expr{component.Indices[0]}
	default:
		return []Expr{&TupleExpr{Elems: append([]Expr(nil), component.Indices...), Pos: component.Pos}}
	}
}

func apalacheRecordFieldPosition(fieldPos, fallback Position) Position {
	if fieldPos.Line > 0 && fieldPos.Column > 0 {
		return fieldPos
	}
	return fallback
}

func apalacheLabel(e *LabelExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	body, diags := apalacheExprIn(e.Body, opts, ctx)
	args := []interface{}{body, apalacheStringValue(e.Name, e.Pos, opts)}
	for _, param := range e.Params {
		args = append(args, apalacheStringValue(param, e.Pos, opts))
	}
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper("LABEL", args, e.Pos, opts), nil
}

func apalacheAction(e *ActionExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	action, actionDiags := apalacheExprIn(e.Action, opts, ctx)
	subscript, subDiags := apalacheExprIn(e.Subscript, opts, ctx)
	diags := append(actionDiags, subDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	oper := "STUTTER"
	if e.Kind == "angle" || e.Kind == "<>" || e.Kind == "NO_STUTTER" {
		oper = "NO_STUTTER"
	}
	return apalacheOper(oper, []interface{}{action, subscript}, e.Pos, opts), nil
}

func apalacheFairness(e *FairnessExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	subscript, subDiags := apalacheExprIn(e.Subscript, opts, ctx)
	action, actionDiags := apalacheExprIn(e.Action, opts, ctx)
	diags := append(subDiags, actionDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	oper := "WEAK_FAIRNESS"
	if e.Kind == "SF" || e.Kind == "SF_" {
		oper = "STRONG_FAIRNESS"
	}
	return apalacheOper(oper, []interface{}{subscript, action}, e.Pos, opts), nil
}

func apalacheSetComprehension(e *SetComprehensionExpr, opts ApalacheIROptions, ctx apalacheExprContext) (interface{}, Diagnostics) {
	localNames := make([]string, 0, len(e.Bounds))
	for _, bound := range e.Bounds {
		localNames = append(localNames, bound.Name)
	}
	bodyCtx := ctx.withLocals(localNames...)
	if e.Predicate != nil && len(e.Bounds) == 1 {
		if ident, ok := e.Element.(*IdentExpr); ok && ident.Name == e.Bounds[0].Name {
			set, setDiags := apalacheExprIn(e.Bounds[0].Set, opts, ctx)
			pred, predDiags := apalacheExprIn(e.Predicate, opts, bodyCtx)
			diags := append(setDiags, predDiags...)
			if diags.HasErrors() {
				return nil, diags
			}
			name := apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: e.Bounds[0].Name}
			return apalacheOper("SET_FILTER", []interface{}{name, set, pred}, e.Pos, opts), nil
		}
	}
	elem, diags := apalacheExprIn(e.Element, opts, bodyCtx)
	args := []interface{}{elem}
	for _, bound := range e.Bounds {
		set, setDiags := apalacheExprIn(bound.Set, opts, ctx)
		diags = append(diags, setDiags...)
		args = append(args, apalacheNameExJSON{Source: apalacheSource(e.Pos, opts), Type: "Untyped", Kind: "NameEx", Name: bound.Name}, set)
	}
	if e.Predicate != nil {
		return nil, Diagnostics{errorAt(e.Pos, "E6009", "ApalacheIR set comprehension with mapped element and predicate is not implemented")}
	}
	if diags.HasErrors() {
		return nil, diags
	}
	return apalacheOper("SET_MAP", args, e.Pos, opts), nil
}

func apalacheOper(oper string, args []interface{}, pos Position, opts ApalacheIROptions) apalacheOperExJSON {
	if args == nil {
		args = []interface{}{}
	}
	return apalacheOperExJSON{
		Source: apalacheSource(pos, opts),
		Type:   "Untyped",
		Kind:   "OperEx",
		Oper:   oper,
		Args:   args,
	}
}

func apalacheBool(value bool, pos Position, opts ApalacheIROptions) apalacheValExJSON {
	return apalacheValExJSON{
		Source: apalacheSource(pos, opts),
		Type:   "Untyped",
		Kind:   "ValEx",
		Value:  apalacheTlaValueJSON{Kind: "TlaBool", Value: value},
	}
}

func apalacheStringValue(value string, pos Position, opts ApalacheIROptions) apalacheValExJSON {
	return apalacheValExJSON{
		Source: apalacheSource(pos, opts),
		Type:   "Untyped",
		Kind:   "ValEx",
		Value:  apalacheTlaValueJSON{Kind: "TlaStr", Value: value},
	}
}

func apalachePredefSet(kind string, pos Position, opts ApalacheIROptions) apalacheValExJSON {
	return apalacheValExJSON{
		Source: apalacheSource(pos, opts),
		Type:   "Untyped",
		Kind:   "ValEx",
		Value:  apalacheTlaValueJSON{Kind: kind},
	}
}

func apalacheSource(pos Position, opts ApalacheIROptions) *apalacheSourceJSON {
	if !opts.IncludeSource || pos.Line <= 0 || pos.Column <= 0 {
		return nil
	}
	end := pos.SourceEnd()
	return &apalacheSourceJSON{
		Filename: apalacheSourceFilename(pos.File),
		From:     apalacheSourcePositionJSON{Line: pos.Line, Column: pos.Column},
		To:       apalacheSourcePositionJSON{Line: end.Line, Column: end.Column},
	}
}

func apalacheDeclSource(pos Position, opts ApalacheIROptions) interface{} {
	if !opts.IncludeSource {
		return nil
	}
	if pos.File == apalacheUnknownSourcePosition.File && pos.Line <= 0 && pos.Column <= 0 {
		return "UNKNOWN"
	}
	return apalacheSource(pos, opts)
}

func apalacheSourceFilename(file string) string {
	if file == "" {
		return ""
	}
	base := filepath.Base(file)
	if ext := filepath.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return base
}

func apalacheUnaryOperator(op string) (string, bool) {
	switch op {
	case "~", "\\lnot", "\\neg", "¬":
		return "NOT", true
	case "-", "-.":
		return "UNARY_MINUS", true
	case "SUBSET":
		return "SET_POWERSET", true
	case "UNION":
		return "SET_UNARY_UNION", true
	case "DOMAIN":
		return "DOMAIN", true
	case "ENABLED":
		return "ENABLED", true
	case "UNCHANGED":
		return "UNCHANGED", true
	case "'", "\\prime":
		return "PRIME", true
	case "[]":
		return "GLOBALLY", true
	case "<>":
		return "EVENTUALLY", true
	default:
		return "", false
	}
}

func apalacheBinaryOperator(op string) (string, bool) {
	switch op {
	case "=":
		return "EQ", true
	case "/=", "#", "≠":
		return "NE", true
	case "+":
		return "PLUS", true
	case "-":
		return "MINUS", true
	case "*":
		return "MULT", true
	case "\\div", "÷":
		return "DIV", true
	case "%":
		return "MOD", true
	case "/":
		return "REAL_DIV", true
	case "^":
		return "POW", true
	case "..", "‥":
		return "INT_RANGE", true
	case "<":
		return "LT", true
	case ">":
		return "GT", true
	case "<=", "=<", "\\leq", "≤":
		return "LE", true
	case ">=", "\\geq", "≥":
		return "GE", true
	case "/\\", "\\land", "∧":
		return "AND", true
	case "\\/", "\\lor", "∨":
		return "OR", true
	case "=>", "⇒":
		return "IMPLIES", true
	case "<=>", "\\equiv", "⇔":
		return "EQUIV", true
	case "\\in", "∈":
		return "SET_IN", true
	case "\\notin", "∉":
		return "SET_NOT_IN", true
	case "\\union", "\\cup", "∪":
		return "SET_UNION2", true
	case "\\intersect", "\\cap", "∩":
		return "SET_INTERSECT", true
	case "\\subseteq", "⊆":
		return "SET_SUBSET_EQ", true
	case "\\", "\\setminus":
		return "SET_MINUS", true
	case "\\X", "\\times", "×":
		return "SET_TIMES", true
	case "~>", "↝":
		return "LEADS_TO", true
	case "-+->":
		return "GUARANTEES", true
	case "\\cdot":
		return "COMPOSE", true
	case "\\o":
		return "Sequences!Concat", true
	default:
		return "", false
	}
}

func apalacheCallOperator(name string) (string, bool) {
	if oper, ok := apalacheUnaryOperator(name); ok {
		return oper, true
	}
	if oper, ok := apalacheBinaryOperator(name); ok {
		return oper, true
	}
	switch name {
	case "Seq", "Sequences!Seq":
		return "Sequences!Seq", true
	case "Len", "Sequences!Len":
		return "Sequences!Len", true
	case "Head", "Sequences!Head":
		return "Sequences!Head", true
	case "Tail", "Sequences!Tail":
		return "Sequences!Tail", true
	case "Append", "Sequences!Append":
		return "Sequences!Append", true
	case "SubSeq", "Sequences!SubSeq":
		return "Sequences!SubSeq", true
	case "IsFiniteSet", "FiniteSets!IsFiniteSet":
		return "FiniteSets!IsFiniteSet", true
	case "Cardinality", "FiniteSets!Cardinality":
		return "FiniteSets!Cardinality", true
	case "__ApalacheFoldSet", "__apalache_folds!__ApalacheFoldSet":
		return "Apalache!ApaFoldSet", true
	case "__ApalacheFoldSeq", "__apalache_folds!__ApalacheFoldSeq":
		return "Apalache!ApaFoldSeqLeft", true
	case "__ApalacheMkSeq", "__apalache_folds!__ApalacheMkSeq":
		return "Apalache!MkSeq", true
	default:
		return "", false
	}
}

func apalacheDebugJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<json error: %v>", err)
	}
	return string(data)
}
