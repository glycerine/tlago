/*******************************************************************************
 * Copyright (c) 2025 NVIDIA Corp. All rights reserved.
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
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
package tlago

import (
	"fmt"
	"sort"

	"github.com/glycerine/tlago/tlc"
)

// The binding machinery is shared with semantic level checking. Constructing
// it does not export XML. Records retain the lexical context in which SANY
// generated them; field-name lookup, in contrast, uses the module context.
type sanyRecordLinter struct {
	symbols  *sanyXMLExporter
	records  map[*Module][]sanyLintRecord
	contexts map[*Definition]sanyXMLExprContext
}

type sanyLintRecord struct {
	expr *RecordExpr
	ctx  sanyXMLExprContext
}

func lintSanySpec(spec *Spec, progress func(string)) Diagnostics {
	if spec == nil || spec.Root == nil {
		return nil
	}
	l := &sanyRecordLinter{symbols: newSanyXMLExporter(spec, SanyXMLOptions{}),
		records: map[*Module][]sanyLintRecord{}, contexts: map[*Definition]sanyXMLExprContext{}}
	modules := l.symbols.semanticModules()
	for _, mod := range modules {
		l.collectModule(mod)
	}
	// ExternalModuleTable's insertion vector excludes nested modules. Native
	// source callers without a loader order still have their root module.
	order := spec.SemanticOrder
	if len(order) == 0 {
		order = []string{spec.Root.Name}
	}
	var diags Diagnostics
	for _, name := range order {
		mod := spec.Modules[name]
		if mod == nil || mod.Library {
			continue
		}
		if progress != nil {
			progress("Linting of module " + mod.Name)
		}
		all := append([]sanyLintRecord(nil), l.records[mod]...)
		visited := map[*Module]bool{mod: true}
		var extend func(*Module)
		extend = func(m *Module) {
			for _, name := range m.Extends {
				dep := spec.Modules[name]
				if dep == nil || visited[dep] {
					continue
				}
				visited[dep] = true
				all = append(all, l.records[dep]...)
				extend(dep)
			}
		}
		extend(mod)
		scope := l.moduleContext(mod)
		for _, record := range l.records[mod] {
			for _, field := range record.expr.Fields {
				symbol := l.symbols.resolvedOperatorSymbol(field.Name, scope)
				if symbol == nil && sanyXMLKnownBuiltin(field.Name) {
					symbol = l.symbols.builtin(field.Name)
				}
				if symbol == nil {
					continue
				}
				if l.builtFromDeclarations(field.Value, record.ctx, map[*Definition]bool{}) {
					continue
				}
				suppress := false
				for _, other := range all {
					if !sanyRecordsSameDomain(other.expr, record.expr) {
						continue
					}
					for _, pair := range other.expr.Fields {
						if l.builtFromDeclarations(pair.Value, other.ctx, map[*Definition]bool{}) {
							suppress = true
							break
						}
					}
					if suppress {
						break
					}
				}
				if suppress {
					continue
				}
				message := fmt.Sprintf("The field name %q in the record constructor is identical to the existing definition or declaration\n"+
					"named %s, located at %s.\n"+
					"The field in the record will not take the value of the %s definition or declaration.\n"+
					"In TLA+, field names in records are strings, regardless of any similarly named declarations or definitions.\n"+
					"Therefore, DOMAIN [%s |-> ...] = {%q} holds true.", field.Name, field.Name,
					tlaPositionToTLCSourceLocation(mod, symbol.Pos), field.Name, field.Name, field.Name)
				pos := field.Pos
				pos.EndLine, pos.EndColumn = pos.Line, pos.Column+len([]rune(field.Name))-1
				diagnostic := warningAt(pos, "W4802", "%s", message)
				diagnostic.SANYRange = SanyRange{Begin: pos, End: pos.SourceEnd()}
				diagnostic.SANYMessage = message
				diags = append(diags, diagnostic)
			}
		}
	}
	return diags
}

func sanyRecordsSameDomain(a, b *RecordExpr) bool {
	if a == b {
		return true
	}
	if len(a.Fields) != len(b.Fields) {
		return false
	}
	for _, left := range a.Fields {
		found := false
		for _, right := range b.Fields {
			if left.Name == right.Name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (l *sanyRecordLinter) moduleContext(mod *Module) sanyXMLExprContext {
	return sanyXMLExprContext{module: mod, scope: l.symbols.scopeForModule(mod, map[string]bool{}),
		formals: map[string]*sanyXMLSymbol{}, defs: map[string]*sanyXMLSymbol{}}
}

func (l *sanyRecordLinter) definitionContext(def *Definition, symbol *sanyXMLSymbol, ctx sanyXMLExprContext) sanyXMLExprContext {
	next := ctx
	next.formals = copySanyXMLSymbolMap(ctx.formals)
	for i, param := range def.Params {
		if symbol != nil && i < len(symbol.Params) {
			next.formals[param] = symbol.Params[i]
		}
	}
	l.contexts[def] = next
	return next
}

func (l *sanyRecordLinter) collectModule(mod *Module) {
	ctx := l.moduleContext(mod)
	type item struct {
		pos  Position
		walk func()
	}
	var items []item
	collect := func(expr Expr, context sanyXMLExprContext) {
		l.walk(expr, context, func(e Expr, c sanyXMLExprContext) bool {
			if record, ok := e.(*RecordExpr); ok {
				l.records[mod] = append(l.records[mod], sanyLintRecord{record, c})
			}
			return false
		}, false, nil)
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		sym := l.symbols.defs[l.symbols.defKey(mod.Name, def.Name)]
		defCtx := l.definitionContext(def, sym, ctx)
		items = append(items, item{def.SourcePosition(), func() {
			collect(def.Expr, defCtx)
			l.walkAssumeProve(def.AssumeProveBody, defCtx, collect)
		}})
	}
	for _, named := range append(append([]NamedExpr(nil), mod.Assumptions...), mod.Theorems...) {
		items = append(items, item{named.SourcePosition(), func() {
			collect(named.Expr, ctx)
			l.walkAssumeProve(named.AssumeProveBody, ctx, collect)
			l.walkProof(sanyXMLTheoremProofNode(named.Syntax), l.symbols.withAssumeProveNewSymbols(ctx, named.AssumeProveBody), collect)
		}})
	}
	for _, inst := range mod.Instances {
		items = append(items, item{inst.SourcePosition(), func() {
			for _, subst := range inst.SubstitutionList {
				collect(subst.Expr, ctx)
			}
		}})
	}
	for _, node := range mod.ProofRefNodes {
		items = append(items, item{sanyNodePosition(node), func() {
			for _, ref := range sanyUseOrHideRefs(node) {
				if !ref.Defs {
					collect(ref.Expr, ctx)
				}
			}
		}})
	}
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i].pos, items[j].pos
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Column < b.Column
	})
	for _, item := range items {
		item.walk()
	}
	// Proof summaries can reference a syntax expression already collected from
	// a theorem. Java genRecord registers the actual record node once.
	seen := map[*RecordExpr]bool{}
	seenSyntax := map[*SanySyntaxNode]bool{}
	unique := l.records[mod][:0]
	for _, record := range l.records[mod] {
		if !seen[record.expr] && (record.expr.Syntax == nil || !seenSyntax[record.expr.Syntax]) {
			unique = append(unique, record)
			seen[record.expr] = true
			if record.expr.Syntax != nil {
				seenSyntax[record.expr.Syntax] = true
			}
		}
	}
	l.records[mod] = unique
}

func (l *sanyRecordLinter) walkAssumeProve(body *AssumeProve, ctx sanyXMLExprContext, visit func(Expr, sanyXMLExprContext)) {
	if body == nil {
		return
	}
	ctx = l.symbols.withAssumeProveNewSymbols(ctx, body)
	for _, item := range body.Assumptions {
		if item.NewSymbol != nil {
			visit(item.NewSymbol.Domain, ctx)
		}
		visit(item.Expr, ctx)
		l.walkAssumeProve(item.Nested, ctx, visit)
	}
	visit(body.Prove, ctx)
}

func (l *sanyRecordLinter) builtFromDeclarations(expr Expr, ctx sanyXMLExprContext, active map[*Definition]bool) bool {
	return l.walk(expr, ctx, nil, true, active)
}

// walk follows SemanticNode.walkChildren. Collection is postorder, as in
// Generator.genRecord; the declaration visitor is preorder and stops as soon
// as it finds an OpAppl whose operator is a declaration or formal parameter.
func (l *sanyRecordLinter) walk(expr Expr, ctx sanyXMLExprContext, collect func(Expr, sanyXMLExprContext) bool, follow bool, active map[*Definition]bool) bool {
	if expr == nil {
		return false
	}
	walk := func(e Expr, c sanyXMLExprContext) bool { return l.walk(e, c, collect, follow, active) }
	operator := func(name string, opArg bool) bool {
		// BOOLEAN is a built-in OpDefNode in Java's global context.
		if follow && !opArg && name == "BOOLEAN" {
			return true
		}
		sym := l.symbols.resolvedOperatorSymbol(name, ctx)
		if sym == nil || opArg || !follow {
			return false
		}
		if sym.Kind == "OpDeclNode" || sym.Kind == "FormalParamNode" {
			return true
		}
		if sym.Kind != "UserDefinedOpKind" {
			return false
		}
		origin := l.symbols.definitionModuleForSymbol(sym)
		if sym.Name == "BOOLEAN" || origin != nil &&
			(sym.Name == "Int" && origin.Name == "Integers" || sym.Name == "Nat" && origin.Name == "Naturals" || sym.Name == "Reals" && origin.Name == "Reals") {
			return true
		}
		def := l.symbols.definitionForSymbol(sym)
		if def == nil {
			return false
		}
		// Java has no visited-node guard: an unproductive recursive traversal
		// throws StackOverflowError. Preserve that catchable Error in Go.
		if active[def] {
			panic(tlc.NewStackOverflowError())
		}
		active[def] = true
		defer delete(active, def)
		defCtx, ok := l.contexts[def]
		if !ok {
			defCtx = l.symbols.subexpressionSelectedContext(sym, ctx)
			defCtx = l.definitionContext(def, sym, defCtx)
		}
		if walk(def.Expr, defCtx) {
			return true
		}
		if meta, ok := l.symbols.instDefMeta[sym.Key]; ok {
			// SubstInNode visits the original body before replacement expressions.
			for _, wrapper := range meta.source.wrappers {
				if l.walkInstanceSubstitutions(wrapper.owner, wrapper.inst, walk) {
					return true
				}
			}
			if l.walkInstanceSubstitutions(meta.owner, meta.inst, walk) {
				return true
			}
		}
		return false
	}
	if selected := sanyExprSelection(expr); follow && selected != nil {
		// OpArgNode does not expose its operator through walkChildren.
		if selected.operator {
			return false
		}
		if selected.newSymbol != nil {
			return true
		}
		selectedCtx := l.moduleContext(selected.definition.module)
		selectedCtx = l.withBounds(selectedCtx, selected.params)
		for _, let := range selected.lets {
			selectedCtx = l.symbols.prepareLetContext(let, selectedCtx).ctx
		}
		if walk(selected.body, selectedCtx) {
			return true
		}
		for _, wrapper := range selected.definition.wrappers {
			if l.walkInstanceSubstitutions(wrapper.owner, wrapper.inst, walk) {
				return true
			}
		}
		for _, arg := range selected.args {
			if walk(arg, ctx) {
				return true
			}
		}
		return false
	}
	switch e := expr.(type) {
	case *IdentExpr:
		sym := l.symbols.resolvedOperatorSymbol(e.Name, ctx)
		if operator(e.Name, sym != nil && sym.Arity > 0) {
			return true
		}
	case *CallExpr:
		if id, ok := e.Callee.(*IdentExpr); ok {
			if operator(id.Name, false) {
				return true
			}
		} else if walk(e.Callee, ctx) {
			return true
		}
		for _, arg := range e.Args {
			if walk(arg, ctx) {
				return true
			}
		}
	case *UnaryExpr:
		if operator(e.Op, false) || walk(e.Expr, ctx) {
			return true
		}
	case *BinaryExpr:
		if operator(e.Op, false) || walk(e.Left, ctx) || walk(e.Right, ctx) {
			return true
		}
	case *LetExpr:
		prep := l.symbols.prepareLetContext(e, ctx)
		for _, source := range prep.localSources {
			l.symbols.instDefMeta[source.sym.Key] = sanyXMLInstanceDefinitionMeta{
				owner: ctx.module, inst: source.inst, targetMod: l.symbols.spec.Modules[source.inst.Module], source: source.source}
		}
		for i := range e.Definitions {
			def := &e.Definitions[i]
			defCtx := l.definitionContext(def, prep.localDefs[i], prep.ctx)
			if walk(def.Expr, defCtx) {
				return true
			}
		}
		for _, inst := range e.Instances {
			for _, subst := range inst.SubstitutionList {
				if walk(subst.Expr, prep.ctx) {
					return true
				}
			}
		}
		if walk(e.Body, prep.ctx) {
			return true
		}
	case *QuantifierExpr:
		if walk(e.Set, ctx) {
			return true
		}
		bounds := []BoundVar{{Name: e.Var, Pos: e.VarPos, OperatorArity: e.OperatorArity, HasOperatorArity: e.HasOperatorArity}}
		if walk(e.Body, l.withBounds(ctx, bounds)) {
			return true
		}
	case *ChooseExpr:
		if walk(e.Set, ctx) || walk(e.Body, l.withBounds(ctx, e.boundVars())) {
			return true
		}
	case *FunctionExpr:
		// A LAMBDA supplied as an operator argument is an OpArgNode, which has
		// no walkChildren children, unlike an ordinary function constructor.
		if follow && e.IsLambda {
			return false
		}
		for _, bound := range e.Bounds {
			if walk(bound.Set, ctx) {
				return true
			}
		}
		if walk(e.Body, l.withBounds(ctx, e.Bounds)) {
			return true
		}
	case *SetComprehensionExpr:
		for _, bound := range e.Bounds {
			if walk(bound.Set, ctx) {
				return true
			}
		}
		boundCtx := l.withBounds(ctx, e.Bounds)
		if walk(e.Element, boundCtx) || walk(e.Predicate, boundCtx) {
			return true
		}
	default:
		for _, child := range sanySubexpressionChildren(expr) {
			if walk(child, ctx) {
				return true
			}
		}
	}
	if collect != nil {
		return collect(expr, ctx)
	}
	return false
}

// SubstInNode and InstanceNode include implicit substitutions as well as WITH
// clauses. An unused replacement can therefore establish declaration use.
func (l *sanyRecordLinter) walkInstanceSubstitutions(owner *Module, inst Instance, walk func(Expr, sanyXMLExprContext) bool) bool {
	ctx := l.moduleContext(owner)
	for _, sym := range l.symbols.instanceParamSymbols(owner, inst) {
		ctx.formals[sym.Name] = sym
	}
	explicit := map[string]bool{}
	for _, subst := range instanceSubstitutions(inst) {
		explicit[subst.Name] = true
		if walk(subst.Expr, ctx) {
			return true
		}
	}
	target := l.symbols.spec.Modules[inst.Module]
	if target == nil {
		return false
	}
	for _, sym := range l.symbols.substitutionTargetSymbols(target, map[string]bool{}) {
		if explicit[sym.Name] {
			continue
		}
		if walk(&IdentExpr{Name: sym.Name, Pos: inst.SourcePosition()}, ctx) {
			return true
		}
	}
	return false
}

func (l *sanyRecordLinter) withBounds(ctx sanyXMLExprContext, bounds []BoundVar) sanyXMLExprContext {
	next := l.symbols.withProofStepBounds(ctx, bounds)
	for _, bound := range bounds {
		if bound.HasOperatorArity {
			next.formals[bound.Name].Arity = bound.OperatorArity
		}
	}
	return next
}

// ProofSummary deliberately summarizes proof commands and omits DEFINE steps.
// Walk their retained syntax with the existing semantic proof translators so
// records in local definitions and nested proof scopes are included as in SANY.
func (l *sanyRecordLinter) walkProof(proof *SanySyntaxNode, ctx sanyXMLExprContext, collect func(Expr, sanyXMLExprContext)) {
	if proof == nil {
		return
	}
	if proof.Kind.JavaName() == "N_TerminalProof" {
		for _, fact := range sanyLeafProofFacts(proof) {
			collect(fact.Expr, ctx)
		}
		return
	}
	if proof.Kind.JavaName() != "N_Proof" {
		return
	}
	current := ctx
	for _, step := range sanyXMLDirectProofSteps(proof) {
		body := sanyXMLProofStepBodyNode(step)
		stepCtx := current
		if body != nil {
			switch body.Kind.JavaName() {
			case "N_DefStep":
				current.defs = copySanyXMLSymbolMap(current.defs)
				for _, sym := range l.symbols.proofLocalDefs[step] {
					def := l.symbols.definitionForSymbol(sym)
					if def == nil {
						continue
					}
					defCtx := l.definitionContext(def, sym, current)
					collect(def.Expr, defCtx)
					l.walkAssumeProve(def.AssumeProveBody, defCtx, collect)
					current.defs[sym.Name] = sym
				}
				stepCtx = current
			case "N_UseOrHide":
				for _, ref := range sanyUseOrHideRefs(body) {
					if !ref.Defs {
						collect(ref.Expr, current)
					}
				}
			case "N_TakeStep", "N_PickStep":
				bounds := sanyProofStepBounds(body)
				for _, bound := range bounds {
					collect(bound.Set, current)
				}
				stepCtx = l.withBounds(current, bounds)
				if body.Kind.JavaName() == "N_PickStep" {
					expr, _ := sanyExpr(lastSanyExpression(body))
					collect(expr, stepCtx)
				}
				current = stepCtx
			case "N_WitnessStep":
				for _, expr := range sanyProofStepExprs(body) {
					collect(expr, current)
				}
			default:
				expr, ap, _ := sanyExprWithAssumeProveBody(lastSanyExpression(body))
				collect(expr, current)
				l.walkAssumeProve(ap, current, collect)
				stepCtx = l.symbols.withAssumeProveNewSymbols(current, ap)
				if _, ok := sanyXMLProofStepSufficesAssumeProveBody(body); ok {
					current = stepCtx
				}
			}
		}
		l.walkProof(sanyXMLNestedProofNode(step), stepCtx, collect)
	}
}
