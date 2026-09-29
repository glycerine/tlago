package tlago

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math/big"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func SanyXMLSource(file, source string) ([]byte, Diagnostics) {
	spec, diags := CheckSanySource(file, source)
	if diags.HasErrors() {
		return nil, diags
	}
	data, xmlDiags := SanyXML(spec)
	diags = append(diags, xmlDiags...)
	return data, diags
}

func SanyXML(spec *Spec) ([]byte, Diagnostics) {
	x := newSanyXMLExporter(spec)
	return x.xml()
}

type sanyXMLExporter struct {
	spec *Spec

	nextUID int
	entries []sanyXMLEntry

	builtins map[string]*sanyXMLSymbol
	modules  map[string]*sanyXMLSymbol
	decls    map[string]*sanyXMLSymbol
	defs     map[string]*sanyXMLSymbol
	newDecls map[string]*sanyXMLSymbol
	bounds   map[string]*sanyXMLSymbol
	lambdas  map[string]*sanyXMLSymbol

	assumes       map[string]*sanyXMLSymbol
	theorems      map[string]*sanyXMLSymbol
	proofTheorems map[*SanySyntaxNode]*sanyXMLSymbol
	proofDefs     map[*SanySyntaxNode]*sanyXMLSymbol
	emitted       map[string]bool

	localCounter int
}

type sanyXMLEntry struct {
	key  string
	uid  int
	body string
}

type sanyXMLSymbol struct {
	UID         int
	Key         string
	Kind        string
	Name        string
	Arity       int
	Level       tlaLevel
	Pos         Position
	DeclKind    DeclarationKind
	XMLDeclKind int
	Params      []*sanyXMLSymbol
}

type sanyXMLExprContext struct {
	module             *Module
	scope              sanyXMLScope
	formals            map[string]*sanyXMLSymbol
	defs               map[string]*sanyXMLSymbol
	proofDefs          map[string]*sanyXMLSymbol
	exceptAtBase       string
	exceptAtComponents string
	exceptAtPos        Position
	exceptAtLevel      tlaLevel
}

type sanyXMLScope struct {
	decls     map[string]*sanyXMLSymbol
	defs      map[string]*sanyXMLSymbol
	declKinds map[string]DeclarationKind
}

type sanyXMLBuiltinInfo struct {
	name    string
	arity   int
	level   tlaLevel
	leibniz []bool
}

func newSanyXMLExporter(spec *Spec) *sanyXMLExporter {
	x := &sanyXMLExporter{
		spec:          spec,
		nextUID:       154,
		builtins:      map[string]*sanyXMLSymbol{},
		modules:       map[string]*sanyXMLSymbol{},
		decls:         map[string]*sanyXMLSymbol{},
		defs:          map[string]*sanyXMLSymbol{},
		newDecls:      map[string]*sanyXMLSymbol{},
		bounds:        map[string]*sanyXMLSymbol{},
		lambdas:       map[string]*sanyXMLSymbol{},
		assumes:       map[string]*sanyXMLSymbol{},
		theorems:      map[string]*sanyXMLSymbol{},
		proofTheorems: map[*SanySyntaxNode]*sanyXMLSymbol{},
		proofDefs:     map[*SanySyntaxNode]*sanyXMLSymbol{},
		emitted:       map[string]bool{},
	}
	for _, mod := range x.sortedModules() {
		x.allocateModule(mod)
	}
	return x
}

func (x *sanyXMLExporter) xml() ([]byte, Diagnostics) {
	if x.spec == nil || x.spec.Root == nil {
		return nil, Diagnostics{errorAt(Position{}, "E7000", "cannot export nil SANY spec to XML")}
	}
	var diags Diagnostics
	for _, mod := range x.sortedModules() {
		diags = append(diags, x.emitModuleEntries(mod)...)
	}
	if diags.HasErrors() {
		return nil, diags
	}

	sort.SliceStable(x.entries, func(i, j int) bool {
		if x.entries[i].key == x.entries[j].key {
			return x.entries[i].uid < x.entries[j].uid
		}
		return x.entries[i].key < x.entries[j].key
	})

	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`)
	b.WriteString("<modules><RootModule>")
	xmlText(&b, x.spec.Root.Name)
	b.WriteString("</RootModule><context>")
	for _, entry := range x.entries {
		b.WriteString("<entry><UID>")
		xmlInt(&b, entry.uid)
		b.WriteString("</UID>")
		b.WriteString(entry.body)
		b.WriteString("</entry>")
	}
	b.WriteString("</context>")
	for _, mod := range x.sortedModules() {
		if sym := x.modules[mod.Name]; sym != nil {
			x.writeRef(&b, sym)
		}
	}
	b.WriteString("</modules>")
	return b.Bytes(), nil
}

func (x *sanyXMLExporter) sortedModules() []*Module {
	if x.spec == nil {
		return nil
	}
	seen := map[string]bool{}
	var mods []*Module
	if x.spec.Root != nil && x.spec.Root.Name != "" {
		mods = append(mods, x.spec.Root)
		seen[x.spec.Root.Name] = true
	}
	for _, mod := range x.spec.Modules {
		if mod == nil || mod.Name == "" || seen[mod.Name] {
			continue
		}
		mods = append(mods, mod)
		seen[mod.Name] = true
	}
	sort.SliceStable(mods, func(i, j int) bool {
		if mods[i] == x.spec.Root {
			return false
		}
		if mods[j] == x.spec.Root {
			return true
		}
		return mods[i].Name < mods[j].Name
	})
	return mods
}

func (x *sanyXMLExporter) allocateModule(mod *Module) {
	if mod == nil || mod.Name == "" {
		return
	}
	if x.modules[mod.Name] == nil {
		x.modules[mod.Name] = x.newSymbol("ModuleNode", "module:"+mod.Name, mod.Name, 0, constantLevel, mod.Pos)
	}
	for _, decl := range mod.Declarations {
		for _, name := range decl.Names {
			key := x.declKey(mod.Name, name)
			if x.decls[key] != nil {
				continue
			}
			arity := 0
			if a, ok := declarationArity(decl, name); ok {
				arity = a
			}
			level := constantLevel
			if decl.Kind == VariableDecl {
				level = variableLevel
			}
			pos := decl.Pos
			if decl.NamePositions != nil {
				if namePos := decl.NamePositions[name]; namePos.Line > 0 || namePos.Column > 0 || namePos.File != "" {
					pos = namePos
				}
			}
			sym := x.newSymbol("OpDeclNode", key, name, arity, level, pos)
			sym.DeclKind = decl.Kind
			x.decls[key] = sym
		}
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		key := x.defKey(mod.Name, def.Name)
		if x.defs[key] == nil {
			x.defs[key] = x.newDefinitionSymbol(key, def)
		}
		if def.AssumeProveBody != nil {
			x.allocateAssumeProveNewSymbols(def.AssumeProveBody)
		}
	}
	for i, assume := range mod.Assumptions {
		key := fmt.Sprintf("assume:%s:%d:%s", mod.Name, i, assume.Name)
		if x.assumes[key] == nil {
			x.assumes[key] = x.newSymbol("AssumeNode", key, assume.Name, 0, constantLevel, assume.SourcePosition())
		}
	}
	for i, theorem := range mod.Theorems {
		key := fmt.Sprintf("theorem:%s:%d:%s", mod.Name, i, theorem.Name)
		if x.theorems[key] == nil {
			x.theorems[key] = x.newSymbol("TheoremNode", key, theorem.Name, 0, constantLevel, theorem.SourcePosition())
		}
		x.allocateProofSteps(mod, theorem.Syntax)
	}
}

func (x *sanyXMLExporter) allocateProofSteps(mod *Module, theoremSyntax *SanySyntaxNode) {
	x.allocateProofNodeSteps(mod, sanyXMLTheoremProofNode(theoremSyntax))
}

func (x *sanyXMLExporter) allocateProofNodeSteps(mod *Module, proof *SanySyntaxNode) {
	if proof == nil || proof.Kind.JavaName() != "N_Proof" {
		return
	}
	for _, step := range sanyXMLDirectProofSteps(proof) {
		body := sanyXMLProofStepBodyNode(step)
		if sanyXMLProofStepIsTheoremLike(body) {
			key := x.proofStepKey(mod.Name, step)
			if x.proofTheorems[step] == nil {
				x.proofTheorems[step] = x.newSymbol("TheoremNode", key, sanyXMLProofStepName(step), 0, constantLevel, sanyNodePosition(step))
			}
			if name := sanyXMLProofStepName(step); name != "" && x.proofDefs[step] == nil {
				x.proofDefs[step] = x.newSymbol("TheoremDefNode", x.proofStepDefKey(mod.Name, step), name, 0, constantLevel, sanyNodePosition(sanyXMLProofStepStartNode(step)))
			}
			if ap, ok := sanyXMLProofStepAssumeProveBody(body); ok {
				x.allocateAssumeProveNewSymbols(ap)
			}
		}
		x.allocateProofNodeSteps(mod, sanyXMLNestedProofNode(step))
	}
}

func (x *sanyXMLExporter) newDefinitionSymbol(key string, def *Definition) *sanyXMLSymbol {
	kind := "UserDefinedOpKind"
	if def != nil && def.FactKind == "theorem" {
		kind = "TheoremDefNode"
	}
	sym := x.newSymbol(kind, key, def.Name, len(def.Params), constantLevel, def.SourcePosition())
	for _, param := range def.Params {
		arity := 0
		if def.ParamArities != nil {
			arity = def.ParamArities[param]
		}
		pos := def.SourcePosition()
		if def.ParamPositions != nil {
			if paramPos := def.ParamPositions[param]; paramPos.Line > 0 || paramPos.Column > 0 || paramPos.File != "" {
				pos = paramPos
			}
		}
		level := tlaLevel(-1)
		if exprReferencesName(def.Expr, param, nil) {
			level = constantLevel
		}
		paramKey := fmt.Sprintf("%s:param:%d:%s", key, len(sym.Params), param)
		paramSym := x.newSymbol("FormalParamNode", paramKey, param, arity, level, pos)
		sym.Params = append(sym.Params, paramSym)
	}
	return sym
}

func (x *sanyXMLExporter) newSymbol(kind, key, name string, arity int, level tlaLevel, pos Position) *sanyXMLSymbol {
	sym := &sanyXMLSymbol{
		UID:   x.nextUID,
		Key:   key,
		Kind:  kind,
		Name:  name,
		Arity: arity,
		Level: level,
		Pos:   pos,
	}
	x.nextUID++
	return sym
}

func (x *sanyXMLExporter) newLocalDefinitionSymbol(prefix string, def *Definition) *sanyXMLSymbol {
	x.localCounter++
	key := fmt.Sprintf("%s:localdef:%d:%s", prefix, x.localCounter, def.Name)
	return x.newDefinitionSymbol(key, def)
}

func (x *sanyXMLExporter) newBoundFormal(prefix, name string, pos Position) *sanyXMLSymbol {
	key := fmt.Sprintf("%s:bound:%s:%s:%d:%d:%d:%d", prefix, name, pos.File, pos.Line, pos.Column, pos.EndLine, pos.EndColumn)
	if sym := x.bounds[key]; sym != nil {
		return sym
	}
	sym := x.newSymbol("FormalParamNode", key, name, 0, constantLevel, pos)
	x.bounds[key] = sym
	return sym
}

func (x *sanyXMLExporter) allocateAssumeProveNewSymbols(body *AssumeProve) {
	for _, item := range body.Assumptions {
		if item.NewSymbol != nil {
			key := x.newSymbolDeclKey(*item.NewSymbol)
			if x.newDecls[key] == nil {
				sym := x.newSymbol("OpDeclNode", key, item.NewSymbol.Name, item.NewSymbol.Arity, item.NewSymbol.Level, item.NewSymbol.Pos)
				sym.XMLDeclKind = item.NewSymbol.Kind
				x.newDecls[key] = sym
			}
		}
		if item.Nested != nil {
			x.allocateAssumeProveNewSymbols(item.Nested)
		}
	}
}

func (x *sanyXMLExporter) emitModuleEntries(mod *Module) Diagnostics {
	if mod == nil {
		return nil
	}
	scope := x.scopeForModule(mod, map[string]bool{})
	ctx := sanyXMLExprContext{module: mod, scope: scope, formals: map[string]*sanyXMLSymbol{}, defs: map[string]*sanyXMLSymbol{}, proofDefs: map[string]*sanyXMLSymbol{}}
	var diags Diagnostics

	for _, decl := range mod.Declarations {
		for _, name := range decl.Names {
			x.emitDeclEntry(x.decls[x.declKey(mod.Name, name)])
		}
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		diags = append(diags, x.emitDefinitionEntry(x.defs[x.defKey(mod.Name, def.Name)], def, ctx)...)
	}
	for i, assume := range mod.Assumptions {
		key := fmt.Sprintf("assume:%s:%d:%s", mod.Name, i, assume.Name)
		diags = append(diags, x.emitAssumeEntry(x.assumes[key], assume, ctx)...)
	}
	for i, theorem := range mod.Theorems {
		key := fmt.Sprintf("theorem:%s:%d:%s", mod.Name, i, theorem.Name)
		diags = append(diags, x.emitTheoremEntry(x.theorems[key], theorem, ctx)...)
	}
	x.emitModuleEntry(mod)
	return diags
}

func (x *sanyXMLExporter) emitModuleEntry(mod *Module) {
	sym := x.modules[mod.Name]
	if sym == nil || x.emitted[sym.Key] {
		return
	}
	x.emitted[sym.Key] = true
	var b bytes.Buffer
	b.WriteString("<ModuleNode>")
	x.writeLocation(&b, mod.Pos)
	b.WriteString("<uniquename>")
	xmlText(&b, mod.Name)
	b.WriteString("</uniquename><extends>")
	for _, ext := range mod.Extends {
		b.WriteString("<uniquename>")
		xmlText(&b, ext)
		b.WriteString("</uniquename>")
	}
	b.WriteString("</extends>")
	for _, ref := range x.moduleMemberRefs(mod) {
		x.writeRef(&b, ref)
	}
	b.WriteString("</ModuleNode>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
}

func (x *sanyXMLExporter) moduleMemberRefs(mod *Module) []*sanyXMLSymbol {
	var refs []*sanyXMLSymbol
	seen := map[string]bool{}
	add := func(sym *sanyXMLSymbol) {
		if sym == nil || seen[sym.Key] {
			return
		}
		seen[sym.Key] = true
		refs = append(refs, sym)
	}
	for _, decl := range mod.Declarations {
		for _, name := range decl.Names {
			add(x.decls[x.declKey(mod.Name, name)])
		}
	}
	for _, ext := range mod.Extends {
		if dep := x.spec.Modules[ext]; dep != nil {
			x.addImportedModuleMemberRefs(dep, add, map[string]bool{})
		}
	}
	for i := range mod.Definitions {
		if mod.Definitions[i].TheoremLike {
			continue
		}
		add(x.defs[x.defKey(mod.Name, mod.Definitions[i].Name)])
	}
	for i, assume := range mod.Assumptions {
		add(x.assumes[fmt.Sprintf("assume:%s:%d:%s", mod.Name, i, assume.Name)])
	}
	for i, theorem := range mod.Theorems {
		add(x.theorems[fmt.Sprintf("theorem:%s:%d:%s", mod.Name, i, theorem.Name)])
	}
	return refs
}

func (x *sanyXMLExporter) addImportedModuleMemberRefs(mod *Module, add func(*sanyXMLSymbol), visiting map[string]bool) {
	if mod == nil || visiting[mod.Name] {
		return
	}
	visiting[mod.Name] = true
	for _, ext := range mod.Extends {
		x.addImportedModuleMemberRefs(x.spec.Modules[ext], add, visiting)
	}
	for _, decl := range mod.Declarations {
		for _, name := range decl.Names {
			add(x.decls[x.declKey(mod.Name, name)])
		}
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		if def.Local || def.TheoremLike {
			continue
		}
		add(x.defs[x.defKey(mod.Name, def.Name)])
	}
	visiting[mod.Name] = false
}

func sortedSymbolMap(m map[string]*sanyXMLSymbol) []*sanyXMLSymbol {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]*sanyXMLSymbol, 0, len(keys))
	for _, key := range keys {
		out = append(out, m[key])
	}
	return out
}

func (x *sanyXMLExporter) emitDeclEntry(sym *sanyXMLSymbol) {
	if sym == nil || x.emitted[sym.Key] {
		return
	}
	x.emitted[sym.Key] = true
	kind := sym.XMLDeclKind
	if kind == 0 {
		kind = 2
		if sym.DeclKind == VariableDecl {
			kind = 3
		}
	}
	var b bytes.Buffer
	b.WriteString("<OpDeclNode>")
	x.writeNode(&b, sym.Pos, sym.Level)
	b.WriteString("<uniquename>")
	xmlText(&b, sym.Name)
	b.WriteString("</uniquename><arity>")
	xmlInt(&b, sym.Arity)
	b.WriteString("</arity><kind>")
	xmlInt(&b, kind)
	b.WriteString("</kind></OpDeclNode>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
}

func (x *sanyXMLExporter) emitDefinitionEntry(sym *sanyXMLSymbol, def *Definition, ctx sanyXMLExprContext) Diagnostics {
	if sym == nil || def == nil || x.emitted[sym.Key] {
		return nil
	}
	x.emitted[sym.Key] = true
	if sym.Kind == "TheoremDefNode" {
		return x.emitTheoremDefEntry(sym, def, ctx)
	}
	for _, param := range sym.Params {
		x.emitFormalEntry(param)
	}
	defCtx := ctx
	defCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	for i, param := range def.Params {
		if i < len(sym.Params) {
			defCtx.formals[param] = sym.Params[i]
		}
	}
	body, diags := x.exprXML(def.Expr, defCtx)
	if diags.HasErrors() {
		return diags
	}
	level := x.exprLevel(def.Expr, defCtx)
	sym.Level = level
	var b bytes.Buffer
	b.WriteString("<UserDefinedOpKind>")
	x.writeNode(&b, sym.Pos, level)
	b.WriteString("<uniquename>")
	xmlText(&b, sym.Name)
	b.WriteString("</uniquename><arity>")
	xmlInt(&b, sym.Arity)
	b.WriteString("</arity>")
	x.writeDefinitionOrigin(&b, sym, ctx.module)
	b.WriteString("<body>")
	b.WriteString(body)
	b.WriteString("</body><params>")
	for _, param := range sym.Params {
		b.WriteString("<leibnizparam>")
		x.writeRef(&b, param)
		b.WriteString("<leibniz/></leibnizparam>")
	}
	b.WriteString("</params>")
	x.writePreComments(&b, def.PreComments)
	if def.Local {
		b.WriteString("<local/>")
	}
	if ctx.scope.defs != nil {
		if _, ok := ctx.scope.defs[sym.Name]; ok && x.isRecursiveDefinition(ctx.module, sym.Name) {
			b.WriteString("<recursive/>")
		}
	}
	b.WriteString("</UserDefinedOpKind>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return diags
}

func (x *sanyXMLExporter) emitTheoremDefEntry(sym *sanyXMLSymbol, def *Definition, ctx sanyXMLExprContext) Diagnostics {
	if def.AssumeProveBody == nil {
		body, diags := x.exprXML(def.Expr, ctx)
		if diags.HasErrors() {
			return diags
		}
		level := x.exprLevel(def.Expr, ctx)
		sym.Level = level
		var b bytes.Buffer
		b.WriteString("<TheoremDefNode>")
		x.writeNode(&b, sym.Pos, level)
		b.WriteString("<uniquename>")
		xmlText(&b, sym.Name)
		b.WriteString("</uniquename>")
		b.WriteString(body)
		b.WriteString("</TheoremDefNode>")
		x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
		return diags
	}

	body, diags := x.assumeProveXML(def.AssumeProveBody, ctx)
	if diags.HasErrors() {
		return diags
	}
	level := x.assumeProveLevel(def.AssumeProveBody, ctx)
	sym.Level = level
	var b bytes.Buffer
	b.WriteString("<TheoremDefNode>")
	x.writeNode(&b, sym.Pos, level)
	b.WriteString("<uniquename>")
	xmlText(&b, sym.Name)
	b.WriteString("</uniquename>")
	b.WriteString(body)
	b.WriteString("</TheoremDefNode>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return diags
}

func (x *sanyXMLExporter) isRecursiveDefinition(mod *Module, name string) bool {
	if mod == nil {
		return false
	}
	for _, decl := range mod.Recursives {
		for _, recursive := range decl.Names {
			if recursive == name {
				return true
			}
		}
	}
	return false
}

func (x *sanyXMLExporter) emitAssumeEntry(sym *sanyXMLSymbol, assume NamedExpr, ctx sanyXMLExprContext) Diagnostics {
	if sym == nil || x.emitted[sym.Key] {
		return nil
	}
	x.emitted[sym.Key] = true
	body, diags := x.exprXML(assume.Expr, ctx)
	if diags.HasErrors() {
		return diags
	}
	level := x.exprLevel(assume.Expr, ctx)
	var b bytes.Buffer
	b.WriteString("<AssumeNode>")
	x.writeNode(&b, assume.SourcePosition(), level)
	b.WriteString("<body>")
	b.WriteString(body)
	b.WriteString("</body></AssumeNode>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return diags
}

func (x *sanyXMLExporter) emitTheoremEntry(sym *sanyXMLSymbol, theorem NamedExpr, ctx sanyXMLExprContext) Diagnostics {
	if sym == nil || x.emitted[sym.Key] {
		return nil
	}
	x.emitted[sym.Key] = true
	defSym, def := x.theoremDefinition(ctx.module, theorem.Name)
	var body string
	var diags Diagnostics
	if def != nil && def.AssumeProveBody != nil {
		body, diags = x.assumeProveXML(def.AssumeProveBody, ctx)
	} else {
		body, diags = x.exprXML(theorem.Expr, ctx)
	}
	if diags.HasErrors() {
		return diags
	}
	level := x.exprLevel(theorem.Expr, ctx)
	if def != nil && def.AssumeProveBody != nil {
		level = x.assumeProveLevel(def.AssumeProveBody, ctx)
	}
	proofCtx := ctx
	if def != nil && def.AssumeProveBody != nil {
		proofCtx = x.withAssumeProveNewSymbols(proofCtx, def.AssumeProveBody)
	}
	proofCtx.proofDefs = x.proofDefinitionMap(theorem.Syntax)
	proof, proofDiags := x.proofXML(sanyXMLTheoremProofNode(theorem.Syntax), proofCtx)
	diags = append(diags, proofDiags...)
	if diags.HasErrors() {
		return diags
	}
	var b bytes.Buffer
	b.WriteString("<TheoremNode>")
	x.writeNode(&b, theorem.SourcePosition(), level)
	if defSym != nil {
		b.WriteString("<definition><TheoremDefRef><UID>")
		xmlInt(&b, defSym.UID)
		b.WriteString("</UID></TheoremDefRef></definition>")
	}
	b.WriteString("<body>")
	b.WriteString(body)
	b.WriteString("</body>")
	b.WriteString(proof)
	b.WriteString("</TheoremNode>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return diags
}

func (x *sanyXMLExporter) theoremDefinition(mod *Module, name string) (*sanyXMLSymbol, *Definition) {
	if mod == nil || name == "" {
		return nil, nil
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		if def.Name != name || def.FactKind != "theorem" {
			continue
		}
		sym := x.defs[x.defKey(mod.Name, def.Name)]
		if sym == nil || sym.Kind != "TheoremDefNode" {
			return nil, def
		}
		return sym, def
	}
	return nil, nil
}

func (x *sanyXMLExporter) proofDefinitionMap(theoremSyntax *SanySyntaxNode) map[string]*sanyXMLSymbol {
	out := map[string]*sanyXMLSymbol{}
	var walk func(*SanySyntaxNode)
	walk = func(proof *SanySyntaxNode) {
		if proof == nil || proof.Kind.JavaName() != "N_Proof" {
			return
		}
		for _, step := range sanyXMLDirectProofSteps(proof) {
			if sym := x.proofDefs[step]; sym != nil && sym.Name != "" {
				out[sym.Name] = sym
			}
			walk(sanyXMLNestedProofNode(step))
		}
	}
	walk(sanyXMLTheoremProofNode(theoremSyntax))
	return out
}

func (x *sanyXMLExporter) proofXML(proof *SanySyntaxNode, ctx sanyXMLExprContext) (string, Diagnostics) {
	if proof == nil {
		return "", nil
	}
	switch proof.Kind.JavaName() {
	case "N_Proof":
		return x.proofStepsXML(proof, ctx)
	case "N_TerminalProof":
		return x.terminalProofXML(proof, ctx)
	default:
		return "", nil
	}
}

func (x *sanyXMLExporter) proofStepsXML(proof *SanySyntaxNode, ctx sanyXMLExprContext) (string, Diagnostics) {
	steps := sanyXMLDirectProofSteps(proof)
	if len(steps) == 0 {
		return "", nil
	}
	level := constantLevel
	var stepXML []string
	var diags Diagnostics
	stepCtx := ctx
	for _, step := range steps {
		body := sanyXMLProofStepBodyNode(step)
		switch {
		case sanyXMLProofStepIsTheoremLike(body):
			sym := x.proofTheorems[step]
			diags = append(diags, x.emitProofStepTheoremEntry(sym, step, stepCtx)...)
			if sym != nil {
				level = maxTlaLevel(level, sym.Level)
			}
			var ref bytes.Buffer
			x.writeRef(&ref, sym)
			stepXML = append(stepXML, ref.String())
		case body != nil && body.Kind.JavaName() == "N_UseOrHide":
			item, itemLevel, itemDiags := x.useOrHideXML(body, stepCtx)
			diags = append(diags, itemDiags...)
			level = maxTlaLevel(level, itemLevel)
			stepXML = append(stepXML, item)
		}
		if ap, ok := sanyXMLProofStepSufficesAssumeProveBody(body); ok {
			stepCtx = x.withAssumeProveNewSymbols(stepCtx, ap)
		}
	}
	if diags.HasErrors() {
		return "", diags
	}
	pos := sanyXMLSpanPosition(steps)
	var b bytes.Buffer
	b.WriteString("<steps>")
	x.writeNode(&b, pos, level)
	for _, item := range stepXML {
		b.WriteString(item)
	}
	b.WriteString("</steps>")
	return b.String(), nil
}

func (x *sanyXMLExporter) emitProofStepTheoremEntry(sym *sanyXMLSymbol, step *SanySyntaxNode, ctx sanyXMLExprContext) Diagnostics {
	if sym == nil || x.emitted[sym.Key] {
		return nil
	}
	bodyNode := sanyXMLProofStepBodyNode(step)
	body, bodyLevel, diags := x.proofStepBodyXML(bodyNode, ctx)
	if diags.HasErrors() {
		return diags
	}
	if defSym := x.proofDefs[step]; defSym != nil {
		diags = append(diags, x.emitProofStepDefEntry(defSym, bodyNode, ctx)...)
		if diags.HasErrors() {
			return diags
		}
	}
	proof := sanyXMLNestedProofNode(step)
	proofCtx := ctx
	if ap, ok := sanyXMLProofStepAssumeProveBody(bodyNode); ok {
		proofCtx = x.withAssumeProveNewSymbols(proofCtx, ap)
	}
	proofLevel := x.proofNodeLevel(proof, proofCtx)
	level := maxTlaLevel(bodyLevel, proofLevel)
	sym.Level = level
	proofBody, proofDiags := x.proofXML(proof, proofCtx)
	diags = append(diags, proofDiags...)
	if diags.HasErrors() {
		return diags
	}

	x.emitted[sym.Key] = true
	var b bytes.Buffer
	b.WriteString("<TheoremNode>")
	x.writeNode(&b, sym.Pos, level)
	if defSym := x.proofDefs[step]; defSym != nil {
		b.WriteString("<definition>")
		x.writeRef(&b, defSym)
		b.WriteString("</definition>")
	}
	b.WriteString("<body>")
	b.WriteString(body)
	b.WriteString("</body>")
	b.WriteString(proofBody)
	if sanyXMLProofStepSuffices(bodyNode) {
		b.WriteString("<suffices></suffices>")
	}
	b.WriteString("</TheoremNode>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return diags
}

func (x *sanyXMLExporter) emitProofStepDefEntry(sym *sanyXMLSymbol, bodyNode *SanySyntaxNode, ctx sanyXMLExprContext) Diagnostics {
	if sym == nil || x.emitted[sym.Key] {
		return nil
	}
	body, level, diags := x.proofStepBodyXML(bodyNode, ctx)
	if diags.HasErrors() {
		return diags
	}
	sym.Level = level
	x.emitted[sym.Key] = true
	var b bytes.Buffer
	b.WriteString("<TheoremDefNode>")
	x.writeNode(&b, sym.Pos, level)
	b.WriteString("<uniquename>")
	xmlText(&b, sym.Name)
	b.WriteString("</uniquename>")
	b.WriteString(body)
	b.WriteString("</TheoremDefNode>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return diags
}

func (x *sanyXMLExporter) proofStepBodyXML(bodyNode *SanySyntaxNode, ctx sanyXMLExprContext) (string, tlaLevel, Diagnostics) {
	if bodyNode == nil {
		return "", constantLevel, nil
	}
	switch bodyNode.Kind.JavaName() {
	case "N_QEDStep":
		pos := sanyNodePosition(firstSanyChildKind(bodyNode, "QED"))
		xml := x.opApplXML(pos, constantLevel, x.builtin("$Qed"), nil, "")
		return xml, constantLevel, nil
	case "N_AssertStep", "N_HaveStep":
		if ap, ok := sanyXMLProofStepAssumeProveBody(bodyNode); ok {
			xml, diags := x.assumeProveXML(ap, ctx)
			return xml, x.assumeProveLevel(ap, ctx), diags
		}
		exprNode := lastSanyExpression(bodyNode)
		expr, exprDiags := sanyExpr(exprNode)
		if exprDiags.HasErrors() {
			return "", constantLevel, exprDiags
		}
		xml, xmlDiags := x.exprXML(expr, ctx)
		if xmlDiags.HasErrors() {
			return "", constantLevel, xmlDiags
		}
		return xml, x.exprLevel(expr, ctx), nil
	case "N_CaseStep":
		exprNode := lastSanyExpression(bodyNode)
		expr, exprDiags := sanyExpr(exprNode)
		if exprDiags.HasErrors() {
			return "", constantLevel, exprDiags
		}
		xml, xmlDiags := x.exprXML(expr, ctx)
		if xmlDiags.HasErrors() {
			return "", constantLevel, xmlDiags
		}
		x.builtin("$Pair")
		level := x.exprLevel(expr, ctx)
		return x.opApplXML(sanyNodePosition(bodyNode), level, x.builtin("$Pfcase"), []string{xml}, ""), level, nil
	default:
		return "", constantLevel, Diagnostics{errorAt(sanyNodePosition(bodyNode), "E7010", "unsupported proof step XML body %s", bodyNode.Kind.JavaName())}
	}
}

func (x *sanyXMLExporter) proofNodeLevel(proof *SanySyntaxNode, ctx sanyXMLExprContext) tlaLevel {
	if proof == nil {
		return constantLevel
	}
	switch proof.Kind.JavaName() {
	case "N_TerminalProof":
		return x.terminalProofLevel(proof, ctx)
	case "N_Proof":
		level := constantLevel
		stepCtx := ctx
		for _, step := range sanyXMLDirectProofSteps(proof) {
			body := sanyXMLProofStepBodyNode(step)
			level = maxTlaLevel(level, x.proofStepLevel(step, stepCtx))
			if ap, ok := sanyXMLProofStepSufficesAssumeProveBody(body); ok {
				stepCtx = x.withAssumeProveNewSymbols(stepCtx, ap)
			}
		}
		return level
	default:
		return constantLevel
	}
}

func (x *sanyXMLExporter) proofStepLevel(step *SanySyntaxNode, ctx sanyXMLExprContext) tlaLevel {
	body := sanyXMLProofStepBodyNode(step)
	switch {
	case sanyXMLProofStepIsTheoremLike(body):
		if sym := x.proofTheorems[step]; sym != nil && sym.Level != constantLevel {
			return sym.Level
		}
		_, bodyLevel, diags := x.proofStepBodyXML(body, ctx)
		if diags.HasErrors() {
			return constantLevel
		}
		proofCtx := ctx
		if ap, ok := sanyXMLProofStepAssumeProveBody(body); ok {
			proofCtx = x.withAssumeProveNewSymbols(proofCtx, ap)
		}
		return maxTlaLevel(bodyLevel, x.proofNodeLevel(sanyXMLNestedProofNode(step), proofCtx))
	case body != nil && body.Kind.JavaName() == "N_UseOrHide":
		return x.useOrHideLevel(body, ctx)
	default:
		return constantLevel
	}
}

func (x *sanyXMLExporter) terminalProofXML(proof *SanySyntaxNode, ctx sanyXMLExprContext) (string, Diagnostics) {
	if proof == nil {
		return "", nil
	}
	for _, child := range proof.GetHeirs() {
		if child == nil || child.Token == nil {
			continue
		}
		switch child.Token.Kind {
		case SanyTokenObvious:
			var b bytes.Buffer
			b.WriteString("<obvious>")
			x.writeNode(&b, sanyNodePosition(child), constantLevel)
			b.WriteString("</obvious>")
			return b.String(), nil
		case SanyTokenOmitted:
			var b bytes.Buffer
			b.WriteString("<omitted>")
			x.writeNode(&b, sanyNodePosition(child), constantLevel)
			b.WriteString("</omitted>")
			return b.String(), nil
		case SanyTokenBy:
			return x.byProofXML(proof, ctx)
		}
	}
	return "", nil
}

func (x *sanyXMLExporter) byProofXML(proof *SanySyntaxNode, ctx sanyXMLExprContext) (string, Diagnostics) {
	level := x.terminalProofLevel(proof, ctx)
	facts, defs, only := x.proofCommandRefs(proof, ctx)
	var b bytes.Buffer
	b.WriteString("<by>")
	x.writeNode(&b, sanyNodePosition(proof), level)
	b.WriteString("<facts>")
	for _, fact := range facts {
		b.WriteString(fact)
	}
	b.WriteString("</facts><defs>")
	for _, def := range defs {
		b.WriteString(def)
	}
	b.WriteString("</defs>")
	if only {
		b.WriteString("<only></only>")
	}
	b.WriteString("</by>")
	return b.String(), nil
}

func (x *sanyXMLExporter) terminalProofLevel(proof *SanySyntaxNode, ctx sanyXMLExprContext) tlaLevel {
	level := constantLevel
	inDefs := false
	for _, child := range proof.GetHeirs() {
		if child == nil || child.Token == nil {
			continue
		}
		switch child.Token.Kind {
		case SanyTokenDF:
			inDefs = true
		case SanyTokenIdentifier:
			if sym := x.proofReferenceSymbol(child.Image, ctx); sym != nil {
				level = maxTlaLevel(level, sym.Level)
			}
		default:
			if isSanyProofStepStartKind(child.Token.Kind) {
				if sym := ctx.proofDefs[sanyXMLProofStepNameImage(child.Image)]; sym != nil {
					level = maxTlaLevel(level, sym.Level)
				}
			}
			_ = inDefs
		}
	}
	return level
}

func (x *sanyXMLExporter) proofCommandRefs(node *SanySyntaxNode, ctx sanyXMLExprContext) ([]string, []string, bool) {
	var facts []string
	var defs []string
	inDefs := false
	only := false
	for _, child := range node.GetHeirs() {
		if child == nil || child.Token == nil {
			continue
		}
		switch child.Token.Kind {
		case SanyTokenBy, SanyTokenUse, SanyTokenHide, SanyTokenComma:
			continue
		case SanyTokenOnly:
			only = true
		case SanyTokenDF:
			inDefs = true
		default:
			if isSanyProofStepStartKind(child.Token.Kind) {
				if sym := ctx.proofDefs[sanyXMLProofStepNameImage(child.Image)]; sym != nil {
					facts = append(facts, x.proofFactXML(sanyNodePosition(child), sym))
				}
				continue
			}
			if child.Token.Kind != SanyTokenIdentifier {
				continue
			}
			sym := x.proofReferenceSymbol(child.Image, ctx)
			if sym == nil {
				continue
			}
			if inDefs {
				var b bytes.Buffer
				x.writeRef(&b, sym)
				defs = append(defs, b.String())
			} else {
				facts = append(facts, x.proofFactXML(sanyNodePosition(child), sym))
			}
		}
	}
	return facts, defs, only
}

func (x *sanyXMLExporter) proofReferenceSymbol(name string, ctx sanyXMLExprContext) *sanyXMLSymbol {
	if sym := ctx.proofDefs[name]; sym != nil {
		return sym
	}
	if sym := ctx.defs[name]; sym != nil {
		return sym
	}
	if sym := ctx.scope.defs[name]; sym != nil {
		return sym
	}
	if sym := ctx.scope.decls[name]; sym != nil {
		return sym
	}
	return nil
}

func (x *sanyXMLExporter) proofFactXML(pos Position, sym *sanyXMLSymbol) string {
	return x.opApplXML(pos, sym.Level, sym, nil, "")
}

func (x *sanyXMLExporter) useOrHideXML(node *SanySyntaxNode, ctx sanyXMLExprContext) (string, tlaLevel, Diagnostics) {
	level := x.useOrHideLevel(node, ctx)
	facts, defs, only := x.proofCommandRefs(node, ctx)
	var b bytes.Buffer
	b.WriteString("<UseOrHideNode>")
	x.writeNode(&b, sanyNodePosition(node), level)
	b.WriteString("<facts>")
	for _, fact := range facts {
		b.WriteString(fact)
	}
	b.WriteString("</facts><defs>")
	for _, def := range defs {
		b.WriteString(def)
	}
	b.WriteString("</defs>")
	if only {
		b.WriteString("<only></only>")
	}
	if sanyXMLUseOrHideIsHide(node) {
		b.WriteString("<hide></hide>")
	}
	b.WriteString("</UseOrHideNode>")
	return b.String(), level, nil
}

func (x *sanyXMLExporter) useOrHideLevel(node *SanySyntaxNode, ctx sanyXMLExprContext) tlaLevel {
	level := constantLevel
	for _, child := range node.GetHeirs() {
		if child == nil || child.Token == nil {
			continue
		}
		if child.Token.Kind == SanyTokenIdentifier {
			if sym := x.proofReferenceSymbol(child.Image, ctx); sym != nil {
				level = maxTlaLevel(level, sym.Level)
			}
		}
	}
	return level
}

func (x *sanyXMLExporter) assumeProveXML(body *AssumeProve, ctx sanyXMLExprContext) (string, Diagnostics) {
	if body == nil {
		return "", Diagnostics{errorAt(Position{}, "E7003", "cannot export nil ASSUME/PROVE to SANY XML")}
	}
	apCtx := ctx
	apCtx.scope = copySanyXMLScope(ctx.scope)
	for _, sym := range x.assumeProveNewSymbolMap(body) {
		apCtx.scope.decls[sym.Name] = sym
	}

	var b bytes.Buffer
	b.WriteString("<AssumeProveNode>")
	x.writeNode(&b, body.Pos, x.assumeProveLevel(body, apCtx))
	b.WriteString("<assumes>")
	var diags Diagnostics
	for _, item := range body.Assumptions {
		switch {
		case item.NewSymbol != nil:
			sym := x.newDecls[x.newSymbolDeclKey(*item.NewSymbol)]
			x.emitDeclEntry(sym)
			b.WriteString("<NewSymbNode>")
			x.writeNode(&b, item.NewSymbol.Source, item.NewSymbol.Level)
			x.writeRef(&b, sym)
			if item.NewSymbol.Domain != nil {
				domain, domainDiags := x.exprXML(item.NewSymbol.Domain, apCtx)
				diags = append(diags, domainDiags...)
				b.WriteString(domain)
			}
			b.WriteString("</NewSymbNode>")
		case item.Nested != nil:
			nested, nestedDiags := x.assumeProveXML(item.Nested, apCtx)
			diags = append(diags, nestedDiags...)
			b.WriteString(nested)
		case item.Expr != nil:
			expr, exprDiags := x.exprXML(item.Expr, apCtx)
			diags = append(diags, exprDiags...)
			b.WriteString(expr)
		}
	}
	b.WriteString("</assumes><prove>")
	if body.Prove != nil {
		prove, proveDiags := x.exprXML(body.Prove, apCtx)
		diags = append(diags, proveDiags...)
		b.WriteString(prove)
	}
	b.WriteString("</prove></AssumeProveNode>")
	if diags.HasErrors() {
		return "", diags
	}
	return b.String(), nil
}

func (x *sanyXMLExporter) assumeProveNewSymbolMap(body *AssumeProve) map[string]*sanyXMLSymbol {
	out := map[string]*sanyXMLSymbol{}
	var walk func(*AssumeProve)
	walk = func(ap *AssumeProve) {
		if ap == nil {
			return
		}
		for _, item := range ap.Assumptions {
			if item.NewSymbol != nil {
				if sym := x.newDecls[x.newSymbolDeclKey(*item.NewSymbol)]; sym != nil {
					out[sym.Name] = sym
				}
			}
			if item.Nested != nil {
				walk(item.Nested)
			}
		}
	}
	walk(body)
	return out
}

func (x *sanyXMLExporter) withAssumeProveNewSymbols(ctx sanyXMLExprContext, body *AssumeProve) sanyXMLExprContext {
	if body == nil {
		return ctx
	}
	next := ctx
	next.scope = copySanyXMLScope(ctx.scope)
	for _, sym := range x.assumeProveNewSymbolMap(body) {
		next.scope.decls[sym.Name] = sym
	}
	return next
}

func (x *sanyXMLExporter) assumeProveLevel(body *AssumeProve, ctx sanyXMLExprContext) tlaLevel {
	if body == nil {
		return constantLevel
	}
	apCtx := ctx
	apCtx.scope = copySanyXMLScope(ctx.scope)
	for _, sym := range x.assumeProveNewSymbolMap(body) {
		apCtx.scope.decls[sym.Name] = sym
	}
	level := constantLevel
	for _, item := range body.Assumptions {
		switch {
		case item.NewSymbol != nil:
			level = maxTlaLevel(level, item.NewSymbol.Level)
			if item.NewSymbol.Domain != nil {
				level = maxTlaLevel(level, x.exprLevel(item.NewSymbol.Domain, apCtx))
			}
		case item.Nested != nil:
			level = maxTlaLevel(level, x.assumeProveLevel(item.Nested, apCtx))
		case item.Expr != nil:
			level = maxTlaLevel(level, x.exprLevel(item.Expr, apCtx))
		}
	}
	if body.Prove != nil {
		level = maxTlaLevel(level, x.exprLevel(body.Prove, apCtx))
	}
	return level
}

func (x *sanyXMLExporter) emitFormalEntry(sym *sanyXMLSymbol) {
	if sym == nil || x.emitted[sym.Key] {
		return
	}
	x.emitted[sym.Key] = true
	var b bytes.Buffer
	b.WriteString("<FormalParamNode>")
	if sym.Pos.Line > 0 || sym.Pos.Column > 0 || sym.Pos.File != "" {
		x.writeLocation(&b, sym.Pos)
		if sym.Level >= 0 {
			b.WriteString("<level>")
			xmlInt(&b, int(sym.Level))
			b.WriteString("</level>")
		}
	}
	b.WriteString("<uniquename>")
	xmlText(&b, sym.Name)
	b.WriteString("</uniquename><arity>")
	xmlInt(&b, sym.Arity)
	b.WriteString("</arity></FormalParamNode>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
}

func (x *sanyXMLExporter) builtin(name string) *sanyXMLSymbol {
	info := sanyXMLBuiltin(name)
	if existing := x.builtins[info.name]; existing != nil {
		return existing
	}
	uid := sanyXMLStableBuiltinUID(info.name)
	sym := &sanyXMLSymbol{
		UID:   uid,
		Key:   "builtin:" + info.name,
		Kind:  "BuiltInKind",
		Name:  info.name,
		Arity: info.arity,
		Level: info.level,
		Pos:   Position{File: "--TLA+ BUILTINS--", Line: 0, Column: 0, EndLine: 0, EndColumn: 0},
	}
	x.builtins[info.name] = sym
	for i := 0; i < info.arity; i++ {
		param := &sanyXMLSymbol{
			UID:   uid*100 + i + 1,
			Key:   fmt.Sprintf("builtin:%s:param:%d", info.name, i),
			Kind:  "FormalParamNode",
			Name:  fmt.Sprintf("Formal_%d", i),
			Arity: 0,
			Level: constantLevel,
		}
		sym.Params = append(sym.Params, param)
		x.emitFormalEntry(param)
	}
	x.emitBuiltinEntry(sym, info)
	return sym
}

func (x *sanyXMLExporter) emitBuiltinEntry(sym *sanyXMLSymbol, info sanyXMLBuiltinInfo) {
	if sym == nil || x.emitted[sym.Key] {
		return
	}
	x.emitted[sym.Key] = true
	var b bytes.Buffer
	b.WriteString("<BuiltInKind>")
	x.writeNode(&b, sym.Pos, sym.Level)
	b.WriteString("<uniquename>")
	xmlText(&b, sym.Name)
	b.WriteString("</uniquename><arity>")
	xmlInt(&b, sym.Arity)
	b.WriteString("</arity>")
	if sym.Arity >= 0 {
		b.WriteString("<params>")
		for i, param := range sym.Params {
			b.WriteString("<leibnizparam>")
			x.writeRef(&b, param)
			if i >= len(info.leibniz) || info.leibniz[i] {
				b.WriteString("<leibniz/>")
			}
			b.WriteString("</leibnizparam>")
		}
		b.WriteString("</params>")
	}
	b.WriteString("</BuiltInKind>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
}

func (x *sanyXMLExporter) exprXML(expr Expr, ctx sanyXMLExprContext) (string, Diagnostics) {
	if expr == nil {
		return "", Diagnostics{errorAt(Position{}, "E7001", "cannot export nil expression to SANY XML")}
	}
	switch e := expr.(type) {
	case *IdentExpr:
		return x.identXML(e, ctx)
	case *LiteralExpr:
		return x.literalXML(e, ctx)
	case *UnaryExpr:
		operand, diags := x.exprXML(e.Expr, ctx)
		if diags.HasErrors() {
			return "", diags
		}
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.operatorSymbol(e.Op, ctx), []string{operand}, ""), nil
	case *BinaryExpr:
		if e.JunctionList {
			return x.junctionListXML(e, ctx)
		}
		left, leftDiags := x.exprXML(e.Left, ctx)
		right, rightDiags := x.exprXML(e.Right, ctx)
		diags := append(leftDiags, rightDiags...)
		if diags.HasErrors() {
			return "", diags
		}
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.operatorSymbol(e.Op, ctx), []string{left, right}, ""), nil
	case *CallExpr:
		return x.callXML(e, ctx)
	case *IfExpr:
		cond, condDiags := x.exprXML(e.Cond, ctx)
		thenExpr, thenDiags := x.exprXML(e.Then, ctx)
		elseExpr, elseDiags := x.exprXML(e.Else, ctx)
		diags := append(append(condDiags, thenDiags...), elseDiags...)
		if diags.HasErrors() {
			return "", diags
		}
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin("$IfThenElse"), []string{cond, thenExpr, elseExpr}, ""), nil
	case *LetExpr:
		return x.letXML(e, ctx)
	case *QuantifierExpr:
		return x.quantifierXML(e, ctx)
	case *CaseExpr:
		return x.caseXML(e, ctx)
	case *ChooseExpr:
		return x.chooseXML(e, ctx)
	case *TupleExpr:
		return x.exprListOpXML("$Tuple", e.Elems, e.Pos, ctx)
	case *SetExpr:
		return x.exprListOpXML("$SetEnumerate", e.Elems, e.Pos, ctx)
	case *RecordExpr:
		return x.recordXML(e, ctx)
	case *RecordComponentExpr:
		record, diags := x.exprXML(e.Record, ctx)
		if diags.HasErrors() {
			return "", diags
		}
		field := x.stringXML(e.Field, apalacheRecordFieldPosition(e.FieldPos, e.Pos))
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin("$RcdSelect"), []string{record, field}, ""), nil
	case *RecordSetExpr:
		return x.recordSetXML(e, ctx)
	case *FunctionExpr:
		return x.boundOpXML("$FcnConstructor", e.Pos, e.Bounds, e.Body, ctx)
	case *FunctionAppExpr:
		fn, diags := x.exprXML(e.Function, ctx)
		if diags.HasErrors() {
			return "", diags
		}
		args := []string{fn}
		for _, arg := range e.Args {
			argXML, argDiags := x.exprXML(arg, ctx)
			diags = append(diags, argDiags...)
			args = append(args, argXML)
		}
		if diags.HasErrors() {
			return "", diags
		}
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin("$FcnApply"), args, ""), nil
	case *ExceptExpr:
		return x.exceptXML(e, ctx)
	case *LabelExpr:
		body, diags := x.exprXML(e.Body, ctx)
		if diags.HasErrors() {
			return "", diags
		}
		var b bytes.Buffer
		b.WriteString("<LabelNode>")
		x.writeNode(&b, e.Pos, x.exprLevel(e, ctx))
		b.WriteString("<uniquename>")
		xmlText(&b, e.Name)
		b.WriteString("</uniquename><arity>")
		xmlInt(&b, len(e.Params))
		b.WriteString("</arity><body>")
		b.WriteString(body)
		b.WriteString("</body><params/></LabelNode>")
		return b.String(), nil
	case *ActionExpr:
		action, actionDiags := x.exprXML(e.Action, ctx)
		subscript, subDiags := x.exprXML(e.Subscript, ctx)
		diags := append(actionDiags, subDiags...)
		if diags.HasErrors() {
			return "", diags
		}
		oper := "$SquareAct"
		if e.Kind == "angle" || e.Kind == "<>" || e.Kind == "NO_STUTTER" {
			oper = "$AngleAct"
		}
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin(oper), []string{action, subscript}, ""), nil
	case *FairnessExpr:
		subscript, subDiags := x.exprXML(e.Subscript, ctx)
		action, actionDiags := x.exprXML(e.Action, ctx)
		diags := append(subDiags, actionDiags...)
		if diags.HasErrors() {
			return "", diags
		}
		oper := "$WF"
		if e.Kind == "SF" || e.Kind == "SF_" {
			oper = "$SF"
		}
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin(oper), []string{subscript, action}, ""), nil
	case *FunctionSetExpr:
		domain, domainDiags := x.exprXML(e.Domain, ctx)
		rng, rangeDiags := x.exprXML(e.Range, ctx)
		diags := append(domainDiags, rangeDiags...)
		if diags.HasErrors() {
			return "", diags
		}
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin("$SetOfFcns"), []string{domain, rng}, ""), nil
	case *SetComprehensionExpr:
		oper := "$SetOfAll"
		if e.Predicate != nil {
			oper = "$SubsetOf"
		}
		body := e.Element
		if e.Predicate != nil {
			body = e.Predicate
		}
		return x.boundOpXML(oper, e.Pos, e.Bounds, body, ctx)
	default:
		return "", Diagnostics{errorAt(expr.Position(), "E7002", "unsupported SANY XML expression %T", expr)}
	}
}

func (x *sanyXMLExporter) identXML(e *IdentExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	if e.Name == "@" && ctx.exceptAtBase != "" && ctx.exceptAtComponents != "" {
		return x.atXML(ctx.exceptAtPos, ctx.exceptAtLevel, ctx.exceptAtBase, ctx.exceptAtComponents), nil
	}
	if sym := ctx.formals[e.Name]; sym != nil {
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), sym, nil, ""), nil
	}
	if sym := ctx.defs[e.Name]; sym != nil {
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), sym, nil, ""), nil
	}
	if sym := ctx.scope.decls[e.Name]; sym != nil {
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), sym, nil, ""), nil
	}
	if sym := ctx.scope.defs[e.Name]; sym != nil {
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), sym, nil, ""), nil
	}
	if e.Name == "TRUE" || e.Name == "FALSE" || builtinIdentifiers[e.Name] {
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin(e.Name), nil, ""), nil
	}
	return "", Diagnostics{errorAt(e.Pos, "E7003", "cannot resolve %s for SANY XML export", e.Name)}
}

func (x *sanyXMLExporter) literalXML(e *LiteralExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	switch e.Kind {
	case "bool":
		return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin(strings.ToUpper(e.Value)), nil, ""), nil
	case "number":
		if strings.Contains(e.Value, ".") {
			var b bytes.Buffer
			b.WriteString("<DecimalNode><mantissa>")
			xmlText(&b, strings.ReplaceAll(e.Value, ".", ""))
			b.WriteString("</mantissa><exponent>")
			parts := strings.SplitN(e.Value, ".", 2)
			xmlInt(&b, -len(parts[1]))
			b.WriteString("</exponent></DecimalNode>")
			return b.String(), nil
		}
		n := new(big.Int)
		if _, ok := n.SetString(e.Value, 0); !ok {
			return "", Diagnostics{errorAt(e.Pos, "E7004", "invalid integer literal %q for SANY XML export", e.Value)}
		}
		var b bytes.Buffer
		b.WriteString("<NumeralNode>")
		x.writeNode(&b, e.Pos, constantLevel)
		b.WriteString("<IntValue>")
		xmlText(&b, n.String())
		b.WriteString("</IntValue></NumeralNode>")
		return b.String(), nil
	case "string", "model":
		return x.stringXML(e.Value, e.Pos), nil
	default:
		return "", Diagnostics{errorAt(e.Pos, "E7005", "unsupported SANY XML literal kind %q", e.Kind)}
	}
}

func (x *sanyXMLExporter) stringXML(value string, pos Position) string {
	if unquoted, err := strconv.Unquote(value); err == nil {
		value = unquoted
	}
	var b bytes.Buffer
	b.WriteString("<StringNode>")
	x.writeNode(&b, pos, constantLevel)
	b.WriteString("<StringValue>")
	xmlText(&b, value)
	b.WriteString("</StringValue></StringNode>")
	return b.String()
}

func (x *sanyXMLExporter) junctionListXML(e *BinaryExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	oper := "$ConjList"
	if e.Op == "\\/" {
		oper = "$DisjList"
	}
	var args []string
	var diags Diagnostics
	var collect func(Expr)
	collect = func(expr Expr) {
		if diags.HasErrors() {
			return
		}
		if nested, ok := expr.(*BinaryExpr); ok && nested.JunctionList && nested.Op == e.Op {
			collect(nested.Left)
			collect(nested.Right)
			return
		}
		xml, xmlDiags := x.exprXML(expr, ctx)
		diags = append(diags, xmlDiags...)
		args = append(args, xml)
	}
	collect(e.Left)
	collect(e.Right)
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin(oper), args, ""), nil
}

func (x *sanyXMLExporter) callXML(e *CallExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	var diags Diagnostics
	var operator *sanyXMLSymbol
	if ident, ok := e.Callee.(*IdentExpr); ok {
		lambda, lambdaDiags := x.lambdaForQuantifiedDefinitionCall(e, ident.Name, ctx)
		diags = append(diags, lambdaDiags...)
		if lambda != nil {
			operator = lambda
		} else {
			operator = x.operatorSymbol(ident.Name, ctx)
		}
	} else {
		return "", Diagnostics{errorAt(e.Pos, "E7006", "SANY XML export supports named operator calls, got %T", e.Callee)}
	}
	args := make([]string, 0, len(e.Args))
	for _, arg := range e.Args {
		argXML, argDiags := x.exprXML(arg, ctx)
		diags = append(diags, argDiags...)
		args = append(args, argXML)
	}
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(e.Pos, x.exprLevel(e, ctx), operator, args, ""), nil
}

func (x *sanyXMLExporter) lambdaForQuantifiedDefinitionCall(e *CallExpr, name string, ctx sanyXMLExprContext) (*sanyXMLSymbol, Diagnostics) {
	if len(e.Args) == 0 {
		return nil, nil
	}
	defSym := x.definitionSymbol(name, ctx)
	if defSym == nil || defSym.Arity != 0 {
		return nil, nil
	}
	def := x.definitionForSymbol(defSym)
	if def == nil {
		return nil, nil
	}
	body := def.Expr
	lambdaCtx := ctx
	lambdaCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	params := make([]*sanyXMLSymbol, 0, len(e.Args))
	for range e.Args {
		quant, ok := body.(*QuantifierExpr)
		if !ok || quant.Set == nil {
			return nil, nil
		}
		pos := quant.VarPos
		if pos.Line == 0 && pos.Column == 0 && pos.File == "" {
			pos = quant.Pos
		}
		param := x.newBoundFormal("expr", quant.Var, pos)
		x.emitFormalEntry(param)
		lambdaCtx.formals[quant.Var] = param
		params = append(params, param)
		body = quant.Body
	}
	key := fmt.Sprintf("lambda:%s:%s:%d:%d:%d:%d:%d", ctx.module.Name, name, e.Pos.Line, e.Pos.Column, e.Pos.EndLine, e.Pos.EndColumn, len(params))
	if sym := x.lambdas[key]; sym != nil {
		return sym, nil
	}
	sym := x.newSymbol("UserDefinedOpKind", key, "LAMBDA", len(params), constantLevel, e.Pos)
	sym.Params = params
	x.lambdas[key] = sym
	diags := x.emitLambdaEntry(sym, body, lambdaCtx)
	if diags.HasErrors() {
		return nil, diags
	}
	return sym, nil
}

func (x *sanyXMLExporter) emitLambdaEntry(sym *sanyXMLSymbol, body Expr, ctx sanyXMLExprContext) Diagnostics {
	if sym == nil || x.emitted[sym.Key] {
		return nil
	}
	x.emitted[sym.Key] = true
	bodyXML, diags := x.exprXML(body, ctx)
	if diags.HasErrors() {
		return diags
	}
	level := x.exprLevel(body, ctx)
	sym.Level = level
	var b bytes.Buffer
	b.WriteString("<UserDefinedOpKind>")
	x.writeNode(&b, sym.Pos, level)
	b.WriteString("<uniquename>")
	xmlText(&b, sym.Name)
	b.WriteString("</uniquename><arity>")
	xmlInt(&b, sym.Arity)
	b.WriteString("</arity>")
	x.writeDefinitionOrigin(&b, sym, ctx.module)
	b.WriteString("<body>")
	b.WriteString(bodyXML)
	b.WriteString("</body><params>")
	for _, param := range sym.Params {
		b.WriteString("<leibnizparam>")
		x.writeRef(&b, param)
		b.WriteString("<leibniz/></leibnizparam>")
	}
	b.WriteString("</params></UserDefinedOpKind>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return nil
}

func (x *sanyXMLExporter) letXML(e *LetExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	letCtx := ctx
	letCtx.defs = copySanyXMLSymbolMap(ctx.defs)
	localDefs := make([]*sanyXMLSymbol, 0, len(e.Definitions))
	for i := range e.Definitions {
		def := &e.Definitions[i]
		sym := x.newLocalDefinitionSymbol("let", def)
		letCtx.defs[def.Name] = sym
		localDefs = append(localDefs, sym)
	}
	var diags Diagnostics
	for i := range e.Definitions {
		diags = append(diags, x.emitDefinitionEntry(localDefs[i], &e.Definitions[i], letCtx)...)
	}
	body, bodyDiags := x.exprXML(e.Body, letCtx)
	diags = append(diags, bodyDiags...)
	if diags.HasErrors() {
		return "", diags
	}
	var b bytes.Buffer
	b.WriteString("<LetInNode>")
	x.writeNode(&b, e.Pos, x.exprLevel(e, ctx))
	b.WriteString("<body>")
	b.WriteString(body)
	b.WriteString("</body><opDefs>")
	for _, sym := range localDefs {
		x.writeRef(&b, sym)
	}
	b.WriteString("</opDefs></LetInNode>")
	return b.String(), nil
}

func (x *sanyXMLExporter) quantifierXML(e *QuantifierExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	oper := "$BoundedForall"
	if e.Set == nil {
		oper = "$UnboundedForall"
	}
	if e.Kind == "\\E" || e.Kind == "EXISTS" {
		if e.Set == nil {
			oper = "$UnboundedExists"
		} else {
			oper = "$BoundedExists"
		}
	}
	if e.Kind == "\\EE" || e.Kind == "TEMPORAL_EXISTS" {
		oper = "$TemporalExists"
	}
	if e.Kind == "\\AA" || e.Kind == "TEMPORAL_FORALL" {
		oper = "$TemporalForall"
	}
	pos := e.VarPos
	if pos.Line == 0 && pos.Column == 0 && pos.File == "" {
		pos = e.Pos
	}
	bounds := []BoundVar{{Name: e.Var, Set: e.Set, Pos: pos}}
	return x.boundOpXML(oper, e.Pos, bounds, e.Body, ctx)
}

func (x *sanyXMLExporter) chooseXML(e *ChooseExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	oper := "$UnboundedChoose"
	if e.Set != nil {
		oper = "$BoundedChoose"
	}
	pos := e.VarPos
	if pos.Line == 0 && pos.Column == 0 && pos.File == "" {
		pos = e.Pos
	}
	bounds := []BoundVar{{Name: e.Var, Set: e.Set, Pos: pos}}
	return x.boundOpXML(oper, e.Pos, bounds, e.Body, ctx)
}

func (x *sanyXMLExporter) boundOpXML(oper string, pos Position, bounds []BoundVar, body Expr, ctx sanyXMLExprContext) (string, Diagnostics) {
	boundCtx := ctx
	boundCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	var boundSymbols bytes.Buffer
	boundSymbols.WriteString("<boundSymbols>")
	var diags Diagnostics
	for _, bound := range bounds {
		formal := x.newBoundFormal("expr", bound.Name, bound.Pos)
		x.emitFormalEntry(formal)
		boundCtx.formals[bound.Name] = formal
		if bound.Set == nil {
			boundSymbols.WriteString("<unbound>")
			x.writeRef(&boundSymbols, formal)
			boundSymbols.WriteString("</unbound>")
			continue
		}
		setXML, setDiags := x.exprXML(bound.Set, ctx)
		diags = append(diags, setDiags...)
		boundSymbols.WriteString("<bound>")
		x.writeRef(&boundSymbols, formal)
		boundSymbols.WriteString(setXML)
		boundSymbols.WriteString("</bound>")
	}
	boundSymbols.WriteString("</boundSymbols>")
	bodyXML, bodyDiags := x.exprXML(body, boundCtx)
	diags = append(diags, bodyDiags...)
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(pos, x.exprLevel(body, boundCtx), x.builtin(oper), []string{bodyXML}, boundSymbols.String()), nil
}

func (x *sanyXMLExporter) caseXML(e *CaseExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	var args []string
	var diags Diagnostics
	for _, arm := range e.Arms {
		test, testDiags := x.exprXML(arm.Test, ctx)
		value, valueDiags := x.exprXML(arm.Value, ctx)
		diags = append(append(diags, testDiags...), valueDiags...)
		level := maxTlaLevel(x.exprLevel(arm.Test, ctx), x.exprLevel(arm.Value, ctx))
		args = append(args, x.opApplXML(arm.Pos, level, x.builtin("$Pair"), []string{test, value}, ""))
	}
	if e.Other != nil {
		other, otherDiags := x.exprXML(e.Other, ctx)
		diags = append(diags, otherDiags...)
		pos := e.OtherPos
		if pos.Line == 0 && pos.Column == 0 && pos.File == "" {
			pos = e.Other.Position()
		}
		args = append(args, x.opApplXML(pos, x.exprLevel(e.Other, ctx), x.builtin("$Pair"), []string{"<StringNode><StringValue>$Other</StringValue></StringNode>", other}, ""))
	}
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin("$Case"), args, ""), nil
}

func (x *sanyXMLExporter) exprListOpXML(oper string, exprs []Expr, pos Position, ctx sanyXMLExprContext) (string, Diagnostics) {
	var args []string
	var diags Diagnostics
	for _, expr := range exprs {
		arg, argDiags := x.exprXML(expr, ctx)
		diags = append(diags, argDiags...)
		args = append(args, arg)
	}
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(pos, x.exprLevel(&TupleExpr{Elems: exprs, Pos: pos}, ctx), x.builtin(oper), args, ""), nil
}

func (x *sanyXMLExporter) recordXML(e *RecordExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	var args []string
	var diags Diagnostics
	for _, field := range e.Fields {
		value, valueDiags := x.exprXML(field.Value, ctx)
		diags = append(diags, valueDiags...)
		pos := fieldSourcePosition(field.Source, field.Pos)
		pair := x.opApplXML(pos, x.exprLevel(field.Value, ctx), x.builtin("$Pair"), []string{x.stringXML(field.Name, field.Pos), value}, "")
		args = append(args, pair)
	}
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin("$RcdConstructor"), args, ""), nil
}

func (x *sanyXMLExporter) recordSetXML(e *RecordSetExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	var args []string
	var diags Diagnostics
	for _, field := range e.Fields {
		set, setDiags := x.exprXML(field.Set, ctx)
		diags = append(diags, setDiags...)
		pos := fieldSourcePosition(field.Source, field.Pos)
		pair := x.opApplXML(pos, x.exprLevel(field.Set, ctx), x.builtin("$Pair"), []string{x.stringXML(field.Name, field.Pos), set}, "")
		args = append(args, pair)
	}
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin("$SetOfRcds"), args, ""), nil
}

func fieldSourcePosition(source, fallback Position) Position {
	if source.Line > 0 || source.Column > 0 || source.File != "" {
		return source
	}
	return fallback
}

func (x *sanyXMLExporter) exceptXML(e *ExceptExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	base, diags := x.exprXML(e.Base, ctx)
	baseLevel := x.exprLevel(e.Base, ctx)
	args := []string{base}
	for _, spec := range e.Specs {
		componentArgs := []string{}
		componentLevel := constantLevel
		for _, component := range spec.Components {
			for _, index := range component.Indices {
				indexXML, indexDiags := x.exprXML(index, ctx)
				diags = append(diags, indexDiags...)
				args = append(args, indexXML)
				componentArgs = append(componentArgs, indexXML)
				componentLevel = maxTlaLevel(componentLevel, x.exprLevel(index, ctx))
			}
			if component.Field != "" {
				field := x.stringXML(component.Field, apalacheRecordFieldPosition(component.FieldPos, component.Pos))
				args = append(args, field)
				componentArgs = append(componentArgs, field)
			}
		}
		valuePos := spec.Pos
		if spec.Value != nil {
			valuePos = spec.Value.Position()
		}
		valueCtx := ctx
		valueCtx.exceptAtBase = base
		valueCtx.exceptAtComponents = x.opApplXML(valuePos, componentLevel, x.builtin("$Seq"), componentArgs, "")
		valueCtx.exceptAtPos = valuePos
		valueCtx.exceptAtLevel = maxTlaLevel(baseLevel, componentLevel)
		value, valueDiags := x.exprXML(spec.Value, valueCtx)
		diags = append(diags, valueDiags...)
		args = append(args, value)
	}
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin("$Except"), args, ""), nil
}

func (x *sanyXMLExporter) atXML(pos Position, level tlaLevel, base, components string) string {
	var b bytes.Buffer
	b.WriteString("<AtNode>")
	x.writeNode(&b, pos, level)
	b.WriteString(base)
	b.WriteString(components)
	b.WriteString("</AtNode>")
	return b.String()
}

func (x *sanyXMLExporter) opApplXML(pos Position, level tlaLevel, operator *sanyXMLSymbol, operands []string, boundSymbols string) string {
	var b bytes.Buffer
	b.WriteString("<OpApplNode>")
	x.writeNode(&b, pos, level)
	b.WriteString("<operator>")
	x.writeRef(&b, operator)
	b.WriteString("</operator><operands>")
	for _, operand := range operands {
		b.WriteString(operand)
	}
	b.WriteString("</operands>")
	b.WriteString(boundSymbols)
	b.WriteString("</OpApplNode>")
	return b.String()
}

func (x *sanyXMLExporter) operatorSymbol(name string, ctx sanyXMLExprContext) *sanyXMLSymbol {
	if sym := ctx.formals[name]; sym != nil {
		return sym
	}
	if sym := ctx.defs[name]; sym != nil {
		return sym
	}
	if sym := ctx.scope.decls[name]; sym != nil {
		return sym
	}
	if sym := ctx.scope.defs[name]; sym != nil {
		return sym
	}
	return x.builtin(sanyXMLBuiltinName(name))
}

func (x *sanyXMLExporter) definitionSymbol(name string, ctx sanyXMLExprContext) *sanyXMLSymbol {
	if sym := ctx.defs[name]; sym != nil {
		return sym
	}
	if sym := ctx.scope.defs[name]; sym != nil {
		return sym
	}
	return nil
}

func (x *sanyXMLExporter) definitionForSymbol(sym *sanyXMLSymbol) *Definition {
	if sym == nil {
		return nil
	}
	for _, mod := range x.spec.Modules {
		for i := range mod.Definitions {
			def := &mod.Definitions[i]
			if x.defs[x.defKey(mod.Name, def.Name)] == sym {
				return def
			}
		}
	}
	return nil
}

func (x *sanyXMLExporter) writeRef(b *bytes.Buffer, sym *sanyXMLSymbol) {
	if sym == nil {
		return
	}
	ref := "BuiltInKindRef"
	switch sym.Kind {
	case "ModuleNode":
		ref = "ModuleNodeRef"
	case "OpDeclNode":
		ref = "OpDeclNodeRef"
	case "UserDefinedOpKind":
		ref = "UserDefinedOpKindRef"
	case "FormalParamNode":
		ref = "FormalParamNodeRef"
	case "AssumeNode":
		ref = "AssumeNodeRef"
	case "TheoremNode":
		ref = "TheoremNodeRef"
	case "TheoremDefNode":
		ref = "TheoremDefRef"
	}
	b.WriteByte('<')
	b.WriteString(ref)
	b.WriteString("><UID>")
	xmlInt(b, sym.UID)
	b.WriteString("</UID></")
	b.WriteString(ref)
	b.WriteByte('>')
}

func (x *sanyXMLExporter) writeDefinitionOrigin(b *bytes.Buffer, sym *sanyXMLSymbol, mod *Module) {
	if sym == nil || mod == nil {
		return
	}
	b.WriteString("<originalOperator>")
	x.writeRef(b, sym)
	b.WriteString("</originalOperator><originallyDefinedInModule>")
	x.writeRef(b, x.modules[mod.Name])
	b.WriteString("</originallyDefinedInModule>")
}

func (x *sanyXMLExporter) writePreComments(b *bytes.Buffer, comments []string) {
	normalized := normalizedSanyPreComments(comments)
	if normalized == "" {
		return
	}
	b.WriteString("<pre-comments><![CDATA[")
	b.WriteString(strings.ReplaceAll(normalized, "]]>", "]]]]><![CDATA[>"))
	b.WriteString("]]></pre-comments>")
}

func normalizedSanyPreComments(comments []string) string {
	if len(comments) == 0 {
		return ""
	}
	parts := make([]string, 0, len(comments))
	for _, comment := range comments {
		comment = strings.TrimRight(comment, "\r\n")
		if comment == "" {
			continue
		}
		parts = append(parts, comment)
	}
	return strings.Join(parts, "\n")
}

func (x *sanyXMLExporter) writeNode(b *bytes.Buffer, pos Position, level tlaLevel) {
	x.writeLocation(b, pos)
	b.WriteString("<level>")
	xmlInt(b, int(level))
	b.WriteString("</level>")
}

func (x *sanyXMLExporter) writeLocation(b *bytes.Buffer, pos Position) {
	b.WriteString("<location><column><begin>")
	xmlInt(b, pos.Column)
	b.WriteString("</begin><end>")
	end := pos.SourceEnd()
	xmlInt(b, end.Column)
	b.WriteString("</end></column><line><begin>")
	xmlInt(b, pos.Line)
	b.WriteString("</begin><end>")
	xmlInt(b, end.Line)
	b.WriteString("</end></line><filename>")
	xmlText(b, sanyXMLSourceFilename(pos.File))
	b.WriteString("</filename></location>")
}

func (x *sanyXMLExporter) exprLevel(expr Expr, ctx sanyXMLExprContext) tlaLevel {
	switch e := expr.(type) {
	case *IdentExpr:
		if e.Name == "@" && ctx.exceptAtBase != "" && ctx.exceptAtComponents != "" {
			return ctx.exceptAtLevel
		}
		return x.operatorLevel(e.Name, ctx)
	case *LiteralExpr:
		return constantLevel
	case *UnaryExpr:
		if e.Op == "'" || e.Op == "UNCHANGED" {
			return maxTlaLevel(actionLevel, x.exprLevel(e.Expr, ctx))
		}
		if e.Op == "[]" || e.Op == "<>" {
			return temporalLevel
		}
		return x.exprLevel(e.Expr, ctx)
	case *BinaryExpr:
		level := maxTlaLevel(x.exprLevel(e.Left, ctx), x.exprLevel(e.Right, ctx))
		if e.Op == "~>" || e.Op == "-+->" {
			return maxTlaLevel(temporalLevel, level)
		}
		return level
	case *CallExpr:
		level := x.exprLevel(e.Callee, ctx)
		for _, arg := range e.Args {
			level = maxTlaLevel(level, x.exprLevel(arg, ctx))
		}
		return level
	case *IfExpr:
		return maxTlaLevel(x.exprLevel(e.Cond, ctx), maxTlaLevel(x.exprLevel(e.Then, ctx), x.exprLevel(e.Else, ctx)))
	case *LetExpr:
		return x.exprLevel(e.Body, ctx)
	case *QuantifierExpr:
		return maxTlaLevel(x.exprLevel(e.Set, ctx), x.exprLevel(e.Body, ctx))
	case *CaseExpr:
		level := constantLevel
		for _, arm := range e.Arms {
			level = maxTlaLevel(level, x.exprLevel(arm.Test, ctx))
			level = maxTlaLevel(level, x.exprLevel(arm.Value, ctx))
		}
		if e.Other != nil {
			level = maxTlaLevel(level, x.exprLevel(e.Other, ctx))
		}
		return level
	case *ChooseExpr:
		return maxTlaLevel(x.exprLevel(e.Set, ctx), x.exprLevel(e.Body, ctx))
	case *TupleExpr:
		level := constantLevel
		for _, elem := range e.Elems {
			level = maxTlaLevel(level, x.exprLevel(elem, ctx))
		}
		return level
	case *SetExpr:
		level := constantLevel
		for _, elem := range e.Elems {
			level = maxTlaLevel(level, x.exprLevel(elem, ctx))
		}
		return level
	case *RecordExpr:
		level := constantLevel
		for _, field := range e.Fields {
			level = maxTlaLevel(level, x.exprLevel(field.Value, ctx))
		}
		return level
	case *RecordComponentExpr:
		return x.exprLevel(e.Record, ctx)
	case *RecordSetExpr:
		level := constantLevel
		for _, field := range e.Fields {
			level = maxTlaLevel(level, x.exprLevel(field.Set, ctx))
		}
		return level
	case *FunctionExpr:
		level := constantLevel
		for _, bound := range e.Bounds {
			level = maxTlaLevel(level, x.exprLevel(bound.Set, ctx))
		}
		return maxTlaLevel(level, x.exprLevel(e.Body, ctx))
	case *FunctionAppExpr:
		level := x.exprLevel(e.Function, ctx)
		for _, arg := range e.Args {
			level = maxTlaLevel(level, x.exprLevel(arg, ctx))
		}
		return level
	case *ExceptExpr:
		level := x.exprLevel(e.Base, ctx)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					level = maxTlaLevel(level, x.exprLevel(index, ctx))
				}
			}
			level = maxTlaLevel(level, x.exprLevel(spec.Value, ctx))
		}
		return level
	case *LabelExpr:
		return x.exprLevel(e.Body, ctx)
	case *ActionExpr:
		return maxTlaLevel(actionLevel, maxTlaLevel(x.exprLevel(e.Action, ctx), x.exprLevel(e.Subscript, ctx)))
	case *FairnessExpr:
		return temporalLevel
	case *FunctionSetExpr:
		return maxTlaLevel(x.exprLevel(e.Domain, ctx), x.exprLevel(e.Range, ctx))
	case *SetComprehensionExpr:
		level := constantLevel
		for _, bound := range e.Bounds {
			level = maxTlaLevel(level, x.exprLevel(bound.Set, ctx))
		}
		level = maxTlaLevel(level, x.exprLevel(e.Element, ctx))
		if e.Predicate != nil {
			level = maxTlaLevel(level, x.exprLevel(e.Predicate, ctx))
		}
		return level
	default:
		return constantLevel
	}
}

func (x *sanyXMLExporter) operatorLevel(name string, ctx sanyXMLExprContext) tlaLevel {
	if sym := ctx.formals[name]; sym != nil {
		return sym.Level
	}
	if sym := ctx.defs[name]; sym != nil {
		return sym.Level
	}
	if sym := ctx.scope.decls[name]; sym != nil {
		return sym.Level
	}
	if sym := ctx.scope.defs[name]; sym != nil {
		return sym.Level
	}
	info := sanyXMLBuiltin(name)
	return info.level
}

func (x *sanyXMLExporter) scopeForModule(mod *Module, visiting map[string]bool) sanyXMLScope {
	scope := sanyXMLScope{
		decls:     map[string]*sanyXMLSymbol{},
		defs:      map[string]*sanyXMLSymbol{},
		declKinds: map[string]DeclarationKind{},
	}
	if mod == nil {
		return scope
	}
	if visiting[mod.Name] {
		return scope
	}
	visiting[mod.Name] = true
	for _, ext := range mod.Extends {
		dep := x.spec.Modules[ext]
		depScope := x.scopeForModule(dep, visiting)
		mergeSanyXMLScope(scope, depScope)
		if dep != nil {
			x.addModuleLocalScope(scope, dep, true)
		}
	}
	x.addModuleLocalScope(scope, mod, false)
	visiting[mod.Name] = false
	return scope
}

func (x *sanyXMLExporter) addModuleLocalScope(scope sanyXMLScope, mod *Module, qualifiedOnly bool) {
	if mod == nil {
		return
	}
	for _, decl := range mod.Declarations {
		for _, name := range decl.Names {
			sym := x.decls[x.declKey(mod.Name, name)]
			if !qualifiedOnly {
				scope.decls[name] = sym
				scope.declKinds[name] = decl.Kind
			}
			scope.decls[mod.Name+"!"+name] = sym
			scope.declKinds[mod.Name+"!"+name] = decl.Kind
		}
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		if def.Local && qualifiedOnly {
			continue
		}
		sym := x.defs[x.defKey(mod.Name, def.Name)]
		if !qualifiedOnly {
			scope.defs[def.Name] = sym
			scope.declKinds[def.Name] = OperatorDecl
		}
		scope.defs[mod.Name+"!"+def.Name] = sym
		scope.declKinds[mod.Name+"!"+def.Name] = OperatorDecl
	}
}

func mergeSanyXMLScope(dst, src sanyXMLScope) {
	for name, sym := range src.decls {
		if dst.decls[name] == nil {
			dst.decls[name] = sym
		}
	}
	for name, sym := range src.defs {
		if dst.defs[name] == nil {
			dst.defs[name] = sym
		}
	}
	for name, kind := range src.declKinds {
		if _, ok := dst.declKinds[name]; !ok {
			dst.declKinds[name] = kind
		}
	}
}

func copySanyXMLScope(src sanyXMLScope) sanyXMLScope {
	return sanyXMLScope{
		decls:     copySanyXMLSymbolMap(src.decls),
		defs:      copySanyXMLSymbolMap(src.defs),
		declKinds: copyDeclarationKindMap(src.declKinds),
	}
}

func copySanyXMLSymbolMap(src map[string]*sanyXMLSymbol) map[string]*sanyXMLSymbol {
	out := make(map[string]*sanyXMLSymbol, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func copyDeclarationKindMap(src map[string]DeclarationKind) map[string]DeclarationKind {
	out := make(map[string]DeclarationKind, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func (x *sanyXMLExporter) newSymbolDeclKey(sym NewSymbol) string {
	return fmt.Sprintf("new:%s:%d:%s:%d:%d:%d:%d:%d", sym.Pos.File, sym.Kind, sym.Name, sym.Pos.Line, sym.Pos.Column, sym.Pos.EndLine, sym.Pos.EndColumn, sym.Arity)
}

func sanyXMLTheoremProofNode(theoremSyntax *SanySyntaxNode) *SanySyntaxNode {
	for _, child := range theoremSyntax.GetHeirs() {
		if child != nil && (child.Kind.JavaName() == "N_Proof" || child.Kind.JavaName() == "N_TerminalProof") {
			return child
		}
	}
	return nil
}

func sanyXMLDirectProofSteps(proof *SanySyntaxNode) []*SanySyntaxNode {
	var out []*SanySyntaxNode
	if proof == nil || proof.Kind.JavaName() != "N_Proof" {
		return out
	}
	for _, child := range proof.GetHeirs() {
		if child != nil && child.Kind.JavaName() == "N_ProofStep" {
			out = append(out, child)
		}
	}
	return out
}

func sanyXMLNestedProofNode(step *SanySyntaxNode) *SanySyntaxNode {
	for _, child := range step.GetHeirs() {
		if child != nil && (child.Kind.JavaName() == "N_Proof" || child.Kind.JavaName() == "N_TerminalProof") {
			return child
		}
	}
	return nil
}

func sanyXMLProofStepStartNode(step *SanySyntaxNode) *SanySyntaxNode {
	for _, child := range step.GetHeirs() {
		if child != nil && child.Token != nil && isSanyProofStepStartKind(child.Token.Kind) {
			return child
		}
	}
	return nil
}

func sanyXMLProofStepBodyNode(step *SanySyntaxNode) *SanySyntaxNode {
	for _, child := range step.GetHeirs() {
		if child == nil {
			continue
		}
		switch child.Kind.JavaName() {
		case "N_AssertStep", "N_CaseStep", "N_HaveStep", "N_PickStep", "N_QEDStep", "N_TakeStep", "N_WitnessStep", "N_UseOrHide", "N_DefStep", "N_Instance":
			return child
		}
	}
	return nil
}

func sanyXMLProofStepIsTheoremLike(body *SanySyntaxNode) bool {
	if body == nil {
		return false
	}
	switch body.Kind.JavaName() {
	case "N_AssertStep", "N_CaseStep", "N_HaveStep", "N_PickStep", "N_QEDStep", "N_TakeStep", "N_WitnessStep":
		return true
	default:
		return false
	}
}

func sanyXMLProofStepName(step *SanySyntaxNode) string {
	start := sanyXMLProofStepStartNode(step)
	if start == nil {
		return ""
	}
	return sanyXMLProofStepNameImage(start.Image)
}

func sanyXMLProofStepNameImage(image string) string {
	end := strings.Index(image, ">")
	if end < 0 || end+1 >= len(image) {
		return ""
	}
	suffix := strings.TrimRight(image[end+1:], ".")
	if suffix == "" {
		return ""
	}
	return image[:end+1] + suffix
}

func sanyXMLProofStepAssumeProveBody(body *SanySyntaxNode) (*AssumeProve, bool) {
	if body == nil {
		return nil, false
	}
	for _, child := range body.GetHeirs() {
		if child != nil && child.Kind.JavaName() == "N_AssumeProve" {
			ap, _ := sanyAssumeProveBody(child)
			return ap, ap != nil
		}
	}
	return nil, false
}

func sanyXMLProofStepSuffices(body *SanySyntaxNode) bool {
	if body == nil {
		return false
	}
	for _, child := range body.GetHeirs() {
		if child != nil && child.Token != nil && child.Token.Kind == SanyTokenSuffices {
			return true
		}
	}
	return false
}

func sanyXMLProofStepSufficesAssumeProveBody(body *SanySyntaxNode) (*AssumeProve, bool) {
	if !sanyXMLProofStepSuffices(body) {
		return nil, false
	}
	return sanyXMLProofStepAssumeProveBody(body)
}

func sanyXMLUseOrHideIsHide(node *SanySyntaxNode) bool {
	for _, child := range node.GetHeirs() {
		if child != nil && child.Token != nil && child.Token.Kind == SanyTokenHide {
			return true
		}
	}
	return false
}

func sanyXMLSpanPosition(nodes []*SanySyntaxNode) Position {
	if len(nodes) == 0 {
		return Position{}
	}
	pos := sanyNodePosition(nodes[0])
	last := sanyNodePosition(nodes[len(nodes)-1])
	if last.Line > 0 {
		pos.EndLine = last.EndLine
		if pos.EndLine == 0 {
			pos.EndLine = last.Line
		}
		pos.EndColumn = last.EndColumn
		if pos.EndColumn == 0 {
			pos.EndColumn = last.Column
		}
	}
	return pos
}

func formalsToLocals(formals map[string]*sanyXMLSymbol) map[string]bool {
	out := make(map[string]bool, len(formals))
	for name := range formals {
		out[name] = true
	}
	return out
}

func exprReferencesName(expr Expr, name string, shadowed map[string]bool) bool {
	if expr == nil || name == "" {
		return false
	}
	switch e := expr.(type) {
	case *IdentExpr:
		return e.Name == name && !shadowed[name]
	case *LiteralExpr:
		return false
	case *UnaryExpr:
		return exprReferencesName(e.Expr, name, shadowed)
	case *BinaryExpr:
		if e.Op == name && !shadowed[name] {
			return true
		}
		return exprReferencesName(e.Left, name, shadowed) || exprReferencesName(e.Right, name, shadowed)
	case *CallExpr:
		if exprReferencesName(e.Callee, name, shadowed) {
			return true
		}
		for _, arg := range e.Args {
			if exprReferencesName(arg, name, shadowed) {
				return true
			}
		}
	case *IfExpr:
		return exprReferencesName(e.Cond, name, shadowed) || exprReferencesName(e.Then, name, shadowed) || exprReferencesName(e.Else, name, shadowed)
	case *LetExpr:
		letShadowed := copyBoolMap(shadowed)
		for _, def := range e.Definitions {
			letShadowed[def.Name] = true
		}
		for _, def := range e.Definitions {
			if exprReferencesName(def.Expr, name, shadowed) {
				return true
			}
		}
		return exprReferencesName(e.Body, name, letShadowed)
	case *QuantifierExpr:
		if exprReferencesName(e.Set, name, shadowed) {
			return true
		}
		next := copyBoolMap(shadowed)
		next[e.Var] = true
		return exprReferencesName(e.Body, name, next)
	case *CaseExpr:
		for _, arm := range e.Arms {
			if exprReferencesName(arm.Test, name, shadowed) || exprReferencesName(arm.Value, name, shadowed) {
				return true
			}
		}
		return exprReferencesName(e.Other, name, shadowed)
	case *ChooseExpr:
		if exprReferencesName(e.Set, name, shadowed) {
			return true
		}
		next := copyBoolMap(shadowed)
		next[e.Var] = true
		return exprReferencesName(e.Body, name, next)
	case *TupleExpr:
		for _, elem := range e.Elems {
			if exprReferencesName(elem, name, shadowed) {
				return true
			}
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			if exprReferencesName(elem, name, shadowed) {
				return true
			}
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			if exprReferencesName(field.Value, name, shadowed) {
				return true
			}
		}
	case *RecordComponentExpr:
		return exprReferencesName(e.Record, name, shadowed)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			if exprReferencesName(field.Set, name, shadowed) {
				return true
			}
		}
	case *FunctionExpr:
		next := copyBoolMap(shadowed)
		for _, bound := range e.Bounds {
			if exprReferencesName(bound.Set, name, shadowed) {
				return true
			}
			next[bound.Name] = true
		}
		return exprReferencesName(e.Body, name, next)
	case *FunctionAppExpr:
		if exprReferencesName(e.Function, name, shadowed) {
			return true
		}
		for _, arg := range e.Args {
			if exprReferencesName(arg, name, shadowed) {
				return true
			}
		}
	case *ExceptExpr:
		if exprReferencesName(e.Base, name, shadowed) {
			return true
		}
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					if exprReferencesName(index, name, shadowed) {
						return true
					}
				}
			}
			if exprReferencesName(spec.Value, name, shadowed) {
				return true
			}
		}
	case *LabelExpr:
		return exprReferencesName(e.Body, name, shadowed)
	case *ActionExpr:
		return exprReferencesName(e.Action, name, shadowed) || exprReferencesName(e.Subscript, name, shadowed)
	case *FairnessExpr:
		return exprReferencesName(e.Subscript, name, shadowed) || exprReferencesName(e.Action, name, shadowed)
	case *FunctionSetExpr:
		return exprReferencesName(e.Domain, name, shadowed) || exprReferencesName(e.Range, name, shadowed)
	case *SetComprehensionExpr:
		next := copyBoolMap(shadowed)
		for _, bound := range e.Bounds {
			if exprReferencesName(bound.Set, name, shadowed) {
				return true
			}
			next[bound.Name] = true
		}
		return exprReferencesName(e.Element, name, next) || exprReferencesName(e.Predicate, name, next)
	}
	return false
}

func (x *sanyXMLExporter) declKey(module, name string) string {
	return "decl:" + module + ":" + name
}

func (x *sanyXMLExporter) defKey(module, name string) string {
	return "def:" + module + ":" + name
}

func (x *sanyXMLExporter) proofStepKey(module string, step *SanySyntaxNode) string {
	pos := sanyNodePosition(step)
	name := sanyXMLProofStepName(step)
	return fmt.Sprintf("proof:%s:%s:%d:%d:%d:%d", module, name, pos.Line, pos.Column, pos.EndLine, pos.EndColumn)
}

func (x *sanyXMLExporter) proofStepDefKey(module string, step *SanySyntaxNode) string {
	pos := sanyNodePosition(sanyXMLProofStepStartNode(step))
	name := sanyXMLProofStepName(step)
	return fmt.Sprintf("proofdef:%s:%s:%d:%d:%d:%d", module, name, pos.Line, pos.Column, pos.EndLine, pos.EndColumn)
}

func sanyXMLBuiltinName(name string) string {
	switch name {
	case "~":
		return "\\lnot"
	case "/\\":
		return "\\land"
	case "\\/":
		return "\\lor"
	default:
		return name
	}
}

func sanyXMLBuiltin(name string) sanyXMLBuiltinInfo {
	name = sanyXMLBuiltinName(name)
	level := constantLevel
	arity := 0
	switch name {
	case "TRUE", "FALSE", "BOOLEAN", "STRING", "$Qed":
		arity = 0
	case "~>", "-+->":
		arity = 2
		level = temporalLevel
	case "$IfThenElse":
		arity = 3
	case "\\lnot", "'", "\\prime", "ENABLED", "UNCHANGED", "[]", "<>", "SUBSET", "UNION", "DOMAIN", "$Pfcase":
		arity = 1
		if name == "'" || name == "\\prime" || name == "UNCHANGED" {
			level = actionLevel
		}
		if name == "ENABLED" {
			level = variableLevel
		}
		if name == "[]" || name == "<>" {
			level = temporalLevel
		}
	case "$Pair", "$SquareAct", "$AngleAct", "$RcdSelect", "$FcnApply":
		arity = 2
		if name == "$SquareAct" || name == "$AngleAct" {
			level = actionLevel
		}
	case "$SubsetOf":
		arity = 1
	case "$Case", "$ConjList", "$DisjList", "$Tuple", "$Seq", "$SetEnumerate", "$RcdConstructor", "$SetOfAll", "$SetOfFcns", "$FcnConstructor", "$BoundedForall", "$BoundedExists", "$BoundedChoose", "$SetOfRcds", "$Except":
		arity = -1
	case "$UnboundedForall", "$UnboundedExists", "$UnboundedChoose", "$TemporalExists", "$TemporalForall":
		arity = 1
		if name == "$TemporalExists" || name == "$TemporalForall" {
			level = temporalLevel
		}
	case "$WF", "$SF":
		arity = 2
		level = temporalLevel
	default:
		arity = 2
	}
	paramCount := arity
	if paramCount < 0 {
		paramCount = 0
	}
	leibniz := make([]bool, paramCount)
	for i := range leibniz {
		leibniz[i] = true
	}
	if name == "'" || name == "\\prime" || name == "ENABLED" || name == "UNCHANGED" || name == "[]" || name == "<>" || name == "$SquareAct" || name == "$AngleAct" || name == "$WF" || name == "$SF" || name == "~>" || name == "-+->" {
		for i := range leibniz {
			leibniz[i] = false
		}
	}
	return sanyXMLBuiltinInfo{name: name, arity: arity, level: level, leibniz: leibniz}
}

func sanyXMLStableBuiltinUID(name string) int {
	switch name {
	case "FALSE":
		return 1
	case "TRUE":
		return 2
	case "=":
		return 4
	case "'":
		return 13
	case "\\lnot":
		return 15
	case "\\land":
		return 19
	case "\\lor":
		return 22
	case "SUBSET":
		return 31
	case "\\in":
		return 40
	case "[]":
		return 61
	case "<>":
		return 63
	case "$BoundedChoose":
		return 78
	case "$BoundedExists":
		return 79
	case "$BoundedForall":
		return 80
	case "$Case":
		return 82
	case "$ConjList":
		return 83
	case "$DisjList":
		return 84
	case "$Except":
		return 85
	case "$FcnConstructor":
		return 89
	case "$IfThenElse":
		return 90
	case "$Pair":
		return 96
	case "$RcdConstructor":
		return 99
	case "$SetEnumerate":
		return 106
	case "$Seq":
		return 105
	case "$SetOfAll":
		return 107
	case "$SetOfFcns":
		return 108
	case "$SubsetOf":
		return 116
	case "$SquareAct":
		return 113
	case "$Tuple":
		return 125
	case "$UnboundedChoose":
		return 126
	case "$Qed":
		return 137
	case "$Pfcase":
		return 138
	}
	hash := 5381
	for _, r := range name {
		hash = ((hash << 5) + hash) + int(r)
	}
	if hash < 0 {
		hash = -hash
	}
	return 10000 + hash%1000000
}

func sanyXMLSourceFilename(file string) string {
	if file == "" {
		return ""
	}
	if file == "--TLA+ BUILTINS--" {
		return file
	}
	base := filepath.Base(file)
	if ext := filepath.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return base
}

func xmlText(b *bytes.Buffer, text string) {
	_ = xml.EscapeText(b, []byte(text))
}

func xmlInt(b *bytes.Buffer, n int) {
	fmt.Fprintf(b, "%d", n)
}
