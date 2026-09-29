package tlago

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
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

	builtins    map[string]*sanyXMLSymbol
	modules     map[string]*sanyXMLSymbol
	decls       map[string]*sanyXMLSymbol
	defs        map[string]*sanyXMLSymbol
	instDefs    map[string]*sanyXMLSymbol
	instParams  map[string][]*sanyXMLSymbol
	letDefs     map[*LetExpr][]*sanyXMLSymbol
	letInsts    map[string]*sanyXMLSymbol
	letInstDefs map[string]*sanyXMLSymbol
	newDecls    map[string]*sanyXMLSymbol
	bounds      map[string]*sanyXMLSymbol
	lambdas     map[string]*sanyXMLSymbol

	assumes        map[string]*sanyXMLSymbol
	assumeDefs     map[string]*sanyXMLSymbol
	theorems       map[string]*sanyXMLSymbol
	proofTheorems  map[*SanySyntaxNode]*sanyXMLSymbol
	proofDefs      map[*SanySyntaxNode]*sanyXMLSymbol
	proofLocalDefs map[*SanySyntaxNode][]*sanyXMLSymbol
	emitted        map[string]bool

	localCounter int
}

type sanyXMLEntry struct {
	key  string
	uid  int
	body string
}

type sanyXMLInstanceNode struct {
	owner *Module
	inst  Instance
}

type sanyXMLDefinitionSource struct {
	name        string
	module      *Module
	def         *Definition
	wrappers    []sanyXMLInstanceWrapper
	fromExtends bool
}

type sanyXMLInstanceDefinitionSource struct {
	keyName     string
	cloneName   string
	module      *Module
	def         *Definition
	wrappers    []sanyXMLInstanceWrapper
	fromExtends bool
}

type sanyXMLInstanceWrapper struct {
	owner  *Module
	inst   Instance
	target *Module
}

type sanyXMLSymbol struct {
	UID         int
	Key         string
	Kind        string
	Name        string
	Arity       int
	Level       tlaLevel
	LevelKnown  bool
	Pos         Position
	DeclKind    DeclarationKind
	XMLDeclKind int
	Params      []*sanyXMLSymbol
	Def         *Definition
	Leibniz     []bool
	ArgWeights  []int
	LevelParams map[string]bool
	leveling    bool
	leveled     bool
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
	recursiveSection   int
	localRecursiveDefs map[string]int
	suppressLetDefs    bool
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
	weights []int
	leibniz []bool
}

type sanyXMLLevelData struct {
	level  tlaLevel
	params map[string]bool
}

type sanyXMLParamUse struct {
	all        map[string]bool
	nonLeibniz map[string]bool
}

type prettyXMLNode struct {
	Name     string
	Attrs    []xml.Attr
	Text     string
	Children []*prettyXMLNode
}

func newSanyXMLExporter(spec *Spec) *sanyXMLExporter {
	x := &sanyXMLExporter{
		spec:           spec,
		nextUID:        154,
		builtins:       map[string]*sanyXMLSymbol{},
		modules:        map[string]*sanyXMLSymbol{},
		decls:          map[string]*sanyXMLSymbol{},
		defs:           map[string]*sanyXMLSymbol{},
		instDefs:       map[string]*sanyXMLSymbol{},
		instParams:     map[string][]*sanyXMLSymbol{},
		letDefs:        map[*LetExpr][]*sanyXMLSymbol{},
		letInsts:       map[string]*sanyXMLSymbol{},
		letInstDefs:    map[string]*sanyXMLSymbol{},
		newDecls:       map[string]*sanyXMLSymbol{},
		bounds:         map[string]*sanyXMLSymbol{},
		lambdas:        map[string]*sanyXMLSymbol{},
		assumes:        map[string]*sanyXMLSymbol{},
		assumeDefs:     map[string]*sanyXMLSymbol{},
		theorems:       map[string]*sanyXMLSymbol{},
		proofTheorems:  map[*SanySyntaxNode]*sanyXMLSymbol{},
		proofDefs:      map[*SanySyntaxNode]*sanyXMLSymbol{},
		proofLocalDefs: map[*SanySyntaxNode][]*sanyXMLSymbol{},
		emitted:        map[string]bool{},
	}
	for _, mod := range x.semanticModules() {
		x.allocateModule(mod)
	}
	x.rebindInstanceDefinitionParams()
	return x
}

func (x *sanyXMLExporter) xml() ([]byte, Diagnostics) {
	if x.spec == nil || x.spec.Root == nil {
		return nil, Diagnostics{errorAt(Position{}, "E7000", "cannot export nil SANY spec to XML")}
	}
	var diags Diagnostics
	for _, mod := range x.semanticModules() {
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
	for _, mod := range x.semanticModules() {
		if sym := x.modules[mod.Name]; sym != nil {
			x.writeRef(&b, sym)
		}
	}
	b.WriteString("</modules>")
	pretty, err := prettySanyXML(b.Bytes())
	if err != nil {
		return nil, Diagnostics{errorAt(Position{}, "E7000", "cannot pretty-print SANY XML: %v", err)}
	}
	return pretty, nil
}

func (x *sanyXMLExporter) semanticModules() []*Module {
	if x.spec == nil {
		return nil
	}
	seen := map[string]bool{}
	var mods []*Module
	for _, name := range x.spec.SemanticOrder {
		mod := x.spec.Modules[name]
		if mod == nil || mod.Name == "" || seen[mod.Name] {
			continue
		}
		mods = append(mods, mod)
		seen[mod.Name] = true
	}
	var remaining []*Module
	for _, mod := range x.spec.Modules {
		if mod == nil || mod.Name == "" || seen[mod.Name] {
			continue
		}
		remaining = append(remaining, mod)
		seen[mod.Name] = true
	}
	sort.SliceStable(remaining, func(i, j int) bool { return remaining[i].Name < remaining[j].Name })
	mods = append(mods, remaining...)
	return mods
}

func prettySanyXML(raw []byte) ([]byte, error) {
	const cdataPrefix = "__TLAGO_CDATA_"
	text := string(raw)
	declaration := ""
	if strings.HasPrefix(text, "<?xml ") {
		end := strings.Index(text, "?>")
		if end < 0 {
			return nil, fmt.Errorf("unterminated XML declaration")
		}
		declaration = text[:end+2]
		text = text[end+2:]
	}
	body, cdataSections := replaceCDATAWithPlaceholders(text, cdataPrefix)
	root, err := parsePrettyXMLTree(body)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	if declaration != "" {
		b.WriteString(declaration)
		b.WriteByte('\n')
	}
	writePrettyXMLNode(&b, root, 0, cdataSections)
	return b.Bytes(), nil
}

func replaceCDATAWithPlaceholders(text, prefix string) (string, map[string]string) {
	cdata := map[string]string{}
	var b strings.Builder
	for {
		start := strings.Index(text, "<![CDATA[")
		if start < 0 {
			b.WriteString(text)
			return b.String(), cdata
		}
		end := strings.Index(text[start:], "]]>")
		if end < 0 {
			b.WriteString(text)
			return b.String(), cdata
		}
		end += start + len("]]>")
		placeholder := fmt.Sprintf("%s%d__", prefix, len(cdata))
		b.WriteString(text[:start])
		b.WriteString(placeholder)
		cdata[placeholder] = text[start:end]
		text = text[end:]
	}
}

func parsePrettyXMLTree(text string) (*prettyXMLNode, error) {
	decoder := xml.NewDecoder(strings.NewReader(text))
	var stack []*prettyXMLNode
	var root *prettyXMLNode
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch tok := token.(type) {
		case xml.StartElement:
			node := &prettyXMLNode{Name: tok.Name.Local, Attrs: append([]xml.Attr(nil), tok.Attr...)}
			if len(stack) == 0 {
				if root != nil {
					return nil, fmt.Errorf("multiple XML root elements")
				}
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
		case xml.EndElement:
			if len(stack) == 0 || stack[len(stack)-1].Name != tok.Name.Local {
				return nil, fmt.Errorf("unexpected XML end element %s", tok.Name.Local)
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) == 0 {
				if strings.TrimSpace(string(tok)) != "" {
					return nil, fmt.Errorf("text outside XML root")
				}
				continue
			}
			stack[len(stack)-1].Text += string(tok)
		}
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("unclosed XML element %s", stack[len(stack)-1].Name)
	}
	if root == nil {
		return nil, fmt.Errorf("empty XML document")
	}
	return root, nil
}

func writePrettyXMLNode(b *bytes.Buffer, node *prettyXMLNode, depth int, cdata map[string]string) {
	indent := strings.Repeat("  ", depth)
	b.WriteString(indent)
	writePrettyXMLStart(b, node, len(node.Children) == 0 && node.Text == "")
	if len(node.Children) == 0 {
		if node.Text != "" {
			writePrettyXMLText(b, node.Text, cdata)
			writePrettyXMLEnd(b, node)
		}
		b.WriteByte('\n')
		return
	}
	b.WriteByte('\n')
	for _, child := range node.Children {
		writePrettyXMLNode(b, child, depth+1, cdata)
	}
	b.WriteString(indent)
	writePrettyXMLEnd(b, node)
	b.WriteByte('\n')
}

func writePrettyXMLStart(b *bytes.Buffer, node *prettyXMLNode, empty bool) {
	b.WriteByte('<')
	b.WriteString(node.Name)
	for _, attr := range node.Attrs {
		b.WriteByte(' ')
		if attr.Name.Space != "" {
			b.WriteString(attr.Name.Space)
			b.WriteByte(':')
		}
		b.WriteString(attr.Name.Local)
		b.WriteString(`="`)
		xmlText(b, attr.Value)
		b.WriteByte('"')
	}
	if empty {
		b.WriteString("/>")
		return
	}
	b.WriteByte('>')
}

func writePrettyXMLEnd(b *bytes.Buffer, node *prettyXMLNode) {
	b.WriteString("</")
	b.WriteString(node.Name)
	b.WriteByte('>')
}

func writePrettyXMLText(b *bytes.Buffer, text string, cdata map[string]string) {
	if raw, ok := cdata[text]; ok {
		b.WriteString(raw)
		return
	}
	xmlText(b, text)
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
	for instIndex, inst := range mod.Instances {
		x.allocateInstanceParams(mod, inst)
		for _, source := range x.instanceDefinitionSources(inst) {
			if x.skipInstanceDefinitionClone(mod, inst, source) {
				continue
			}
			if x.duplicateInstanceDefinitionClone(mod, instIndex, inst, source) {
				continue
			}
			key := x.instanceDefKey(mod.Name, instIndex, inst, source.keyName)
			if x.instDefs[key] != nil {
				continue
			}
			sym := x.newInstanceDefinitionSymbol(key, source, inst.SourcePosition(), x.instanceParamSymbols(mod, inst))
			x.instDefs[key] = sym
		}
	}
	for i, assume := range mod.Assumptions {
		key := fmt.Sprintf("assume:%s:%d:%s", mod.Name, i, assume.Name)
		if x.assumes[key] == nil {
			x.assumes[key] = x.newSymbol("AssumeNode", key, assume.Name, 0, constantLevel, assume.SourcePosition())
		}
		if assume.Name != "" && x.assumeDefs[key] == nil {
			x.assumeDefs[key] = x.newSymbol("AssumeDef", key+":def", assume.Name, 0, constantLevel, assume.SourcePosition())
		}
		if assume.AssumeProveBody != nil {
			x.allocateAssumeProveNewSymbols(assume.AssumeProveBody)
		}
	}
	for i, theorem := range mod.Theorems {
		key := fmt.Sprintf("theorem:%s:%d:%s", mod.Name, i, theorem.Name)
		if x.theorems[key] == nil {
			x.theorems[key] = x.newSymbol("TheoremNode", key, theorem.Name, 0, constantLevel, theorem.SourcePosition())
		}
		if theorem.AssumeProveBody != nil {
			x.allocateAssumeProveNewSymbols(theorem.AssumeProveBody)
		}
		x.allocateProofSteps(mod, theorem.Syntax)
	}
}

func (x *sanyXMLExporter) rebindInstanceDefinitionParams() {
	for _, mod := range x.semanticModules() {
		for instIndex, inst := range mod.Instances {
			instanceParams := x.instanceParamSymbols(mod, inst)
			for _, source := range x.instanceDefinitionSources(inst) {
				sym := x.instDefs[x.instanceDefKey(mod.Name, instIndex, inst, source.keyName)]
				if sym == nil || source.module == nil || source.def == nil {
					continue
				}
				original := x.defs[x.defKey(source.module.Name, source.def.Name)]
				if original == nil {
					continue
				}
				sym.Params = append(append([]*sanyXMLSymbol(nil), instanceParams...), original.Params...)
				sym.Arity = len(sym.Params)
			}
		}
	}
}

func (x *sanyXMLExporter) allocateInstanceParams(owner *Module, inst Instance) {
	if len(inst.Params) == 0 {
		return
	}
	key := x.instanceParamKey(owner, inst)
	if x.instParams[key] != nil {
		return
	}
	for _, name := range inst.Params {
		pos := inst.SourcePosition()
		if inst.ParamPositions != nil {
			if paramPos := inst.ParamPositions[name]; paramPos.Line > 0 || paramPos.Column > 0 || paramPos.File != "" {
				pos = paramPos
			}
		}
		paramKey := fmt.Sprintf("%s:param:%d:%s", key, len(x.instParams[key]), name)
		x.instParams[key] = append(x.instParams[key], x.newSymbol("FormalParamNode", paramKey, name, 0, constantLevel, pos))
	}
}

func (x *sanyXMLExporter) instanceParamSymbols(owner *Module, inst Instance) []*sanyXMLSymbol {
	if len(inst.Params) == 0 {
		return nil
	}
	return x.instParams[x.instanceParamKey(owner, inst)]
}

func (x *sanyXMLExporter) instanceParamKey(owner *Module, inst Instance) string {
	ownerName := ""
	if owner != nil {
		ownerName = owner.Name
	}
	return "instparams:" + ownerName + ":" + x.instanceNodeKey(inst)
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
		if body != nil && body.Kind.JavaName() == "N_DefStep" {
			if defs, diags := sanyXMLDefStepDefinitions(body); !diags.HasErrors() && len(defs) > 0 && x.proofLocalDefs[step] == nil {
				for i := range defs {
					def := &defs[i]
					key := fmt.Sprintf("prooflocaldef:%s:%d:%d:%d:%s", mod.Name, def.SourcePosition().Line, def.SourcePosition().Column, i, def.Name)
					x.proofLocalDefs[step] = append(x.proofLocalDefs[step], x.newDefinitionSymbol(key, def))
				}
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
	sym.Def = def
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
		paramKey := fmt.Sprintf("%s:param:%d:%s", key, len(sym.Params), param)
		paramSym := x.newSymbol("FormalParamNode", paramKey, param, arity, constantLevel, pos)
		paramSym.LevelKnown = exprReferencesName(def.Expr, param, nil)
		sym.Params = append(sym.Params, paramSym)
	}
	return sym
}

func (x *sanyXMLExporter) newInstanceDefinitionSymbol(key string, source sanyXMLInstanceDefinitionSource, pos Position, instanceParams []*sanyXMLSymbol) *sanyXMLSymbol {
	def := source.def
	kind := "UserDefinedOpKind"
	if def != nil && def.FactKind == "theorem" {
		kind = "TheoremDefNode"
	}
	sym := x.newSymbol(kind, key, source.cloneName, 0, constantLevel, pos)
	sym.Def = def
	sym.Params = append(sym.Params, instanceParams...)
	if source.module != nil && def != nil {
		original := x.defs[x.defKey(source.module.Name, def.Name)]
		if original != nil {
			sym.Params = append(append([]*sanyXMLSymbol(nil), instanceParams...), original.Params...)
		}
	}
	if len(sym.Params) == len(instanceParams) && def != nil {
		for _, param := range def.Params {
			pos := def.SourcePosition()
			if def.ParamPositions != nil {
				if paramPos := def.ParamPositions[param]; paramPos.Line > 0 || paramPos.Column > 0 || paramPos.File != "" {
					pos = paramPos
				}
			}
			arity := 0
			if def.ParamArities != nil {
				arity = def.ParamArities[param]
			}
			paramKey := fmt.Sprintf("%s:param:%d:%s", key, len(sym.Params), param)
			sym.Params = append(sym.Params, x.newSymbol("FormalParamNode", paramKey, param, arity, constantLevel, pos))
		}
	}
	sym.Arity = len(sym.Params)
	return sym
}

func (x *sanyXMLExporter) newSymbol(kind, key, name string, arity int, level tlaLevel, pos Position) *sanyXMLSymbol {
	uid := x.nextGeneratedUID()
	sym := &sanyXMLSymbol{
		UID:        uid,
		Key:        key,
		Kind:       kind,
		Name:       name,
		Arity:      arity,
		Level:      level,
		LevelKnown: level >= 0,
		Pos:        pos,
	}
	return sym
}

func (x *sanyXMLExporter) nextGeneratedUID() int {
	for sanyXMLReservedBuiltinUID(x.nextUID) {
		x.nextUID++
	}
	uid := x.nextUID
	x.nextUID++
	return uid
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
	for instIndex, inst := range mod.Instances {
		targetMod := x.spec.Modules[inst.Module]
		sourceContexts := map[string]sanyXMLExprContext{}
		for _, source := range x.instanceDefinitionSources(inst) {
			if x.skipInstanceDefinitionClone(mod, inst, source) {
				continue
			}
			if x.duplicateInstanceDefinitionClone(mod, instIndex, inst, source) {
				continue
			}
			if source.module == nil || source.def == nil {
				continue
			}
			sourceCtx, ok := sourceContexts[source.module.Name]
			if !ok {
				sourceCtx = sanyXMLExprContext{module: source.module, scope: x.scopeForModule(source.module, map[string]bool{}), formals: map[string]*sanyXMLSymbol{}, defs: map[string]*sanyXMLSymbol{}, proofDefs: map[string]*sanyXMLSymbol{}}
				sourceContexts[source.module.Name] = sourceCtx
			}
			sym := x.instDefs[x.instanceDefKey(mod.Name, instIndex, inst, source.keyName)]
			original := x.defs[x.defKey(source.module.Name, source.def.Name)]
			diags = append(diags, x.emitInstanceDefinitionEntry(sym, original, mod, inst, source.module, targetMod, source.def, sourceCtx, source.wrappers)...)
		}
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		diags = append(diags, x.emitDefinitionEntry(x.defs[x.defKey(mod.Name, def.Name)], def, ctx)...)
	}
	for i, assume := range mod.Assumptions {
		key := fmt.Sprintf("assume:%s:%d:%s", mod.Name, i, assume.Name)
		diags = append(diags, x.emitAssumeEntry(x.assumes[key], x.assumeDefs[key], assume, ctx)...)
	}
	for i, theorem := range mod.Theorems {
		key := fmt.Sprintf("theorem:%s:%d:%s", mod.Name, i, theorem.Name)
		diags = append(diags, x.emitTheoremEntry(x.theorems[key], theorem, ctx)...)
	}
	diags = append(diags, x.emitModuleEntry(mod)...)
	return diags
}

func (x *sanyXMLExporter) emitModuleEntry(mod *Module) Diagnostics {
	sym := x.modules[mod.Name]
	if sym == nil || x.emitted[sym.Key] {
		return nil
	}
	x.emitted[sym.Key] = true
	var b bytes.Buffer
	b.WriteString("<ModuleNode>")
	x.writeLocation(&b, mod.Pos)
	b.WriteString("<uniquename>")
	xmlText(&b, mod.Name)
	b.WriteString("</uniquename><extends>")
	extends := append([]string(nil), mod.Extends...)
	sort.Strings(extends)
	for _, ext := range extends {
		b.WriteString("<uniquename>")
		xmlText(&b, ext)
		b.WriteString("</uniquename>")
	}
	b.WriteString("</extends>")
	for _, ref := range x.moduleMemberRefs(mod) {
		x.writeRef(&b, ref)
	}
	var diags Diagnostics
	for _, node := range x.moduleInstanceNodes(mod) {
		instDiags := x.writeInstanceNode(&b, node.owner, node.inst)
		diags = append(diags, instDiags...)
	}
	b.WriteString("</ModuleNode>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return diags
}

func (x *sanyXMLExporter) moduleInstanceNodes(mod *Module) []sanyXMLInstanceNode {
	if mod == nil {
		return nil
	}
	var out []sanyXMLInstanceNode
	seen := map[string]bool{}
	for _, ext := range mod.Extends {
		x.addImportedModuleInstanceNodes(x.spec.Modules[ext], &out, seen, map[string]bool{})
	}
	for _, inst := range mod.Instances {
		key := x.instanceNodeKey(inst)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, sanyXMLInstanceNode{owner: mod, inst: inst})
	}
	return out
}

func (x *sanyXMLExporter) addImportedModuleInstanceNodes(mod *Module, out *[]sanyXMLInstanceNode, seen map[string]bool, visiting map[string]bool) {
	if mod == nil || visiting[mod.Name] {
		return
	}
	visiting[mod.Name] = true
	for _, ext := range mod.Extends {
		x.addImportedModuleInstanceNodes(x.spec.Modules[ext], out, seen, visiting)
	}
	for _, inst := range mod.Instances {
		*out = append(*out, sanyXMLInstanceNode{owner: mod, inst: inst})
	}
	visiting[mod.Name] = false
}

func (x *sanyXMLExporter) instanceNodeKey(inst Instance) string {
	pos := inst.SourcePosition()
	return fmt.Sprintf("%s:%s:%t:%d:%d:%d:%d", inst.Module, inst.qualifier(), inst.Local, pos.Line, pos.Column, pos.EndLine, pos.EndColumn)
}

func (x *sanyXMLExporter) writeInstanceNode(b *bytes.Buffer, owner *Module, inst Instance) Diagnostics {
	b.WriteString("<InstanceNode>")
	x.writeNode(b, inst.SourcePosition(), constantLevel)
	if inst.Name != "" {
		b.WriteString("<uniquename>")
		xmlText(b, inst.Name)
		b.WriteString("</uniquename>")
	}
	b.WriteString("<module>")
	xmlText(b, inst.Module)
	b.WriteString("</module>")
	diags := x.writeInstanceSubstitutions(b, owner, inst)
	params := x.instanceParamSymbols(owner, inst)
	if len(params) == 0 {
		b.WriteString("<params/>")
	} else {
		b.WriteString("<params>")
		for _, param := range params {
			x.emitFormalEntry(param)
			x.writeRef(b, param)
		}
		b.WriteString("</params>")
	}
	if inst.Local {
		b.WriteString("<local/>")
	}
	b.WriteString("</InstanceNode>")
	return diags
}

func (x *sanyXMLExporter) writeInstanceSubstitutions(b *bytes.Buffer, owner *Module, inst Instance) Diagnostics {
	substs, _, diags := x.instanceSubstitutionsXML(owner, inst)
	b.WriteString(substs)
	return diags
}

func (x *sanyXMLExporter) instanceSubstitutionsXML(owner *Module, inst Instance) (string, bool, Diagnostics) {
	var b bytes.Buffer
	b.WriteString("<substs>")
	var diags Diagnostics
	hasSubsts := false
	explicit := map[string]bool{}
	if len(inst.SubstitutionList) > 0 || len(inst.Substitutions) > 0 {
		ctx := sanyXMLExprContext{module: owner, scope: x.scopeForModule(owner, map[string]bool{}), formals: map[string]*sanyXMLSymbol{}, defs: map[string]*sanyXMLSymbol{}, proofDefs: map[string]*sanyXMLSymbol{}}
		for _, subst := range instanceSubstitutions(inst) {
			if subst.Name != "" {
				explicit[subst.Name] = true
			}
			target := x.substitutionTargetSymbol(inst.Module, subst.Name)
			if target == nil || subst.Expr == nil {
				continue
			}
			exprXML, exprDiags := x.substitutionReplacementXML(target, subst.Expr, ctx)
			diags = append(diags, exprDiags...)
			if exprDiags.HasErrors() {
				continue
			}
			hasSubsts = true
			b.WriteString("<Subst>")
			x.writeRef(&b, target)
			b.WriteString(exprXML)
			b.WriteString("</Subst>")
		}
	}

	instMod := x.spec.Modules[inst.Module]
	if owner == nil || instMod == nil {
		b.WriteString("</substs>")
		return b.String(), hasSubsts, diags
	}
	ownerScope := x.scopeForModule(owner, map[string]bool{})
	instanceParams := map[string]*sanyXMLSymbol{}
	for _, sym := range x.instanceParamSymbols(owner, inst) {
		instanceParams[sym.Name] = sym
	}
	for _, target := range x.substitutionTargetSymbols(instMod, map[string]bool{}) {
		name := target.Name
		if explicit[name] {
			continue
		}
		replacement := instanceParams[name]
		if replacement == nil {
			replacement = ownerScope.defs[name]
		}
		if replacement == nil {
			replacement = ownerScope.decls[name]
		}
		if replacement == nil {
			continue
		}
		x.ensureOperatorLevelData(replacement, sanyXMLExprContext{
			module:    owner,
			scope:     ownerScope,
			formals:   instanceParams,
			defs:      map[string]*sanyXMLSymbol{},
			proofDefs: map[string]*sanyXMLSymbol{},
		})
		replacementLevel := replacement.Level
		hasSubsts = true
		b.WriteString("<Subst>")
		x.writeRef(&b, target)
		if target.Arity > 0 && replacement.Arity > 0 {
			b.WriteString(x.opArgXML(inst.SourcePosition(), replacementLevel, replacement))
		} else {
			b.WriteString(x.opApplXML(inst.SourcePosition(), replacementLevel, replacement, nil, ""))
		}
		b.WriteString("</Subst>")
	}
	b.WriteString("</substs>")
	return b.String(), hasSubsts, diags
}

func (x *sanyXMLExporter) substitutionReplacementXML(target *sanyXMLSymbol, expr Expr, ctx sanyXMLExprContext) (string, Diagnostics) {
	if target != nil && target.Arity > 0 {
		if ident, ok := expr.(*IdentExpr); ok {
			if sym := x.operatorSymbol(ident.Name, ctx); sym != nil && sym.Arity > 0 {
				return x.opArgXML(ident.Pos, sym.Level, sym), nil
			}
		}
	}
	return x.exprXML(expr, ctx)
}

func (x *sanyXMLExporter) substInXML(pos Position, level tlaLevel, substs, body string, from *Module, to *Module) string {
	return x.substInXMLWithTag("SubstInNode", pos, level, substs, body, from, to)
}

func (x *sanyXMLExporter) substInXMLWithTag(tag string, pos Position, level tlaLevel, substs, body string, from *Module, to *Module) string {
	var b bytes.Buffer
	b.WriteByte('<')
	b.WriteString(tag)
	b.WriteByte('>')
	x.writeNode(&b, pos, level)
	b.WriteString(substs)
	b.WriteString("<body>")
	b.WriteString(body)
	b.WriteString("</body>")
	if from != nil {
		b.WriteString("<instFrom>")
		x.writeRef(&b, x.modules[from.Name])
		b.WriteString("</instFrom>")
	}
	if to != nil {
		b.WriteString("<instTo>")
		x.writeRef(&b, x.modules[to.Name])
		b.WriteString("</instTo>")
	}
	b.WriteString("</")
	b.WriteString(tag)
	b.WriteByte('>')
	return b.String()
}

func (x *sanyXMLExporter) substitutionTargetSymbol(module, name string) *sanyXMLSymbol {
	return x.substitutionTargetSymbolInModule(x.spec.Modules[module], name, map[string]bool{})
}

func (x *sanyXMLExporter) substitutionTargetSymbolInModule(mod *Module, name string, visiting map[string]bool) *sanyXMLSymbol {
	if mod == nil || visiting[mod.Name] {
		return nil
	}
	visiting[mod.Name] = true
	defer func() {
		visiting[mod.Name] = false
	}()
	if sym := x.decls[x.declKey(mod.Name, name)]; sym != nil {
		return sym
	}
	for _, ext := range mod.Extends {
		if sym := x.substitutionTargetSymbolInModule(x.spec.Modules[ext], name, visiting); sym != nil {
			return sym
		}
	}
	return nil
}

func (x *sanyXMLExporter) substitutionTargetSymbols(mod *Module, visiting map[string]bool) []*sanyXMLSymbol {
	if mod == nil || visiting[mod.Name] || isEmbeddedStandardModule(mod) {
		return nil
	}
	visiting[mod.Name] = true
	defer func() {
		visiting[mod.Name] = false
	}()
	var out []*sanyXMLSymbol
	for _, ext := range mod.Extends {
		out = append(out, x.substitutionTargetSymbols(x.spec.Modules[ext], visiting)...)
	}
	seen := map[string]bool{}
	for _, sym := range out {
		seen[sym.Name] = true
	}
	for _, decl := range mod.Declarations {
		if decl.Kind != ConstantDecl && decl.Kind != VariableDecl {
			continue
		}
		for _, name := range decl.Names {
			sym := x.decls[x.declKey(mod.Name, name)]
			if sym == nil {
				continue
			}
			if seen[name] {
				for i, existing := range out {
					if existing.Name == name {
						out[i] = sym
						break
					}
				}
				continue
			}
			seen[name] = true
			out = append(out, sym)
		}
	}
	return out
}

func (x *sanyXMLExporter) moduleMemberRefs(mod *Module) []*sanyXMLSymbol {
	var refs []*sanyXMLSymbol
	seen := map[string]bool{}
	add := func(sym *sanyXMLSymbol) {
		if sym == nil {
			return
		}
		if sym.Kind == "AssumeNode" || sym.Kind == "TheoremNode" {
			refs = append(refs, sym)
			return
		}
		if seen[sym.Key] {
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
	for instIndex, inst := range mod.Instances {
		clonedOrSkippedSources := map[string]bool{}
		for _, source := range x.instanceDefinitionSources(inst) {
			clonedOrSkippedSources[source.keyName] = true
			if x.skipInstanceDefinitionClone(mod, inst, source) {
				if source.fromExtends {
					add(x.instanceDefinitionSourceOriginalSymbol(source))
				}
				continue
			}
			if x.duplicateInstanceDefinitionClone(mod, instIndex, inst, source) {
				continue
			}
			sym := x.instDefs[x.instanceDefKey(mod.Name, instIndex, inst, source.keyName)]
			if sym != nil && sym.Kind == "TheoremDefNode" {
				continue
			}
			add(sym)
		}
		if inst.Local {
			continue
		}
		for _, source := range x.exportedDefinitionSources(x.spec.Modules[inst.Module], map[string]bool{}) {
			if !source.fromExtends || clonedOrSkippedSources[source.name] {
				continue
			}
			add(x.instanceDefinitionSourceOriginalSymbol(sanyXMLInstanceDefinitionSource{
				module: source.module,
				def:    source.def,
			}))
		}
	}
	for i, assume := range mod.Assumptions {
		add(x.assumes[fmt.Sprintf("assume:%s:%d:%s", mod.Name, i, assume.Name)])
	}
	for i, theorem := range mod.Theorems {
		add(x.theorems[fmt.Sprintf("theorem:%s:%d:%s", mod.Name, i, theorem.Name)])
	}
	return refs
}

func (x *sanyXMLExporter) moduleHasDefinition(mod *Module, name string) bool {
	if mod == nil {
		return false
	}
	for i := range mod.Definitions {
		if mod.Definitions[i].Name == name && !mod.Definitions[i].TheoremLike {
			return true
		}
	}
	return false
}

func (x *sanyXMLExporter) skipInstanceDefinitionClone(owner *Module, inst Instance, source sanyXMLInstanceDefinitionSource) bool {
	if !inst.exportsUnqualified() {
		return false
	}
	name := source.keyName
	return x.moduleHasDefinition(owner, name) || x.moduleExtendsDefinition(owner, name, map[string]bool{})
}

func (x *sanyXMLExporter) instanceDefinitionSourceOriginalSymbol(source sanyXMLInstanceDefinitionSource) *sanyXMLSymbol {
	if source.module == nil || source.def == nil {
		return nil
	}
	return x.defs[x.defKey(source.module.Name, source.def.Name)]
}

func (x *sanyXMLExporter) duplicateInstanceDefinitionClone(owner *Module, instIndex int, inst Instance, source sanyXMLInstanceDefinitionSource) bool {
	if !inst.exportsUnqualified() {
		return false
	}
	return x.earlierInstanceDefinitionSymbol(owner, instIndex, source.keyName) != nil
}

func (x *sanyXMLExporter) earlierInstanceDefinitionSymbol(owner *Module, instIndex int, name string) *sanyXMLSymbol {
	if owner == nil {
		return nil
	}
	if instIndex > len(owner.Instances) {
		instIndex = len(owner.Instances)
	}
	for i := 0; i < instIndex; i++ {
		inst := owner.Instances[i]
		if sym := x.instDefs[x.instanceDefKey(owner.Name, i, inst, name)]; sym != nil {
			return sym
		}
	}
	return nil
}

func (x *sanyXMLExporter) moduleExtendsDefinition(mod *Module, name string, visiting map[string]bool) bool {
	if mod == nil || visiting[mod.Name] {
		return false
	}
	visiting[mod.Name] = true
	defer func() {
		visiting[mod.Name] = false
	}()
	for _, ext := range mod.Extends {
		dep := x.spec.Modules[ext]
		for _, source := range x.exportedDefinitionSources(dep, map[string]bool{}) {
			if source.name == name {
				return true
			}
		}
		if x.moduleExtendsDefinition(dep, name, visiting) {
			return true
		}
	}
	return false
}

func (x *sanyXMLExporter) instanceDefinitionSources(inst Instance) []sanyXMLInstanceDefinitionSource {
	instMod := x.spec.Modules[inst.Module]
	if instMod == nil {
		return nil
	}
	sources := x.exportedDefinitionSources(instMod, map[string]bool{})
	if inst.exportsUnqualified() {
		sources = x.instanceExportedDefinitionSources(instMod, map[string]bool{}, inst.Local, len(inst.Params) > 0)
	} else if len(inst.Params) > 0 {
		sources = append(sources, x.directTheoremDefinitionSources(instMod)...)
	}
	out := make([]sanyXMLInstanceDefinitionSource, 0, len(sources))
	for _, source := range sources {
		if source.def == nil || source.module == nil {
			continue
		}
		cloneName := source.name
		if !inst.exportsUnqualified() {
			cloneName = inst.qualifier() + "!" + source.name
		}
		out = append(out, sanyXMLInstanceDefinitionSource{
			keyName:     source.name,
			cloneName:   cloneName,
			module:      source.module,
			def:         source.def,
			wrappers:    append([]sanyXMLInstanceWrapper(nil), source.wrappers...),
			fromExtends: source.fromExtends,
		})
	}
	return out
}

func (x *sanyXMLExporter) directTheoremDefinitionSources(mod *Module) []sanyXMLDefinitionSource {
	if mod == nil {
		return nil
	}
	var out []sanyXMLDefinitionSource
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		if def.Local || !def.TheoremLike {
			continue
		}
		out = append(out, sanyXMLDefinitionSource{name: def.Name, module: mod, def: def})
	}
	return out
}

func (x *sanyXMLExporter) instanceExportedDefinitionSources(mod *Module, visiting map[string]bool, includeLibraryExtends bool, includeTheoremDefs bool) []sanyXMLDefinitionSource {
	if mod == nil || visiting[mod.Name] {
		return nil
	}
	visiting[mod.Name] = true
	defer func() {
		visiting[mod.Name] = false
	}()

	byName := map[string]sanyXMLDefinitionSource{}
	for _, ext := range mod.Extends {
		dep := x.spec.Modules[ext]
		if dep == nil {
			continue
		}
		if dep.Library && !includeLibraryExtends {
			continue
		}
		for i := range dep.Definitions {
			def := &dep.Definitions[i]
			if def.Local || def.TheoremLike {
				continue
			}
			byName[def.Name] = sanyXMLDefinitionSource{name: def.Name, module: dep, def: def, fromExtends: true}
		}
	}
	for _, inst := range mod.Instances {
		if inst.Local {
			continue
		}
		target := x.spec.Modules[inst.Module]
		for _, source := range x.exportedDefinitionSources(target, visiting) {
			if source.def == nil || source.module == nil {
				continue
			}
			wrappers := append([]sanyXMLInstanceWrapper(nil), source.wrappers...)
			wrappers = append(wrappers, sanyXMLInstanceWrapper{owner: mod, inst: inst, target: target})
			if inst.exportsUnqualified() {
				source.wrappers = wrappers
				source.fromExtends = false
				byName[source.name] = source
			}
			if qualifier := inst.qualifier(); qualifier != "" {
				byName[qualifier+"!"+source.name] = sanyXMLDefinitionSource{
					name:        qualifier + "!" + source.name,
					module:      source.module,
					def:         source.def,
					wrappers:    wrappers,
					fromExtends: false,
				}
			}
		}
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		if def.Local || (def.TheoremLike && !includeTheoremDefs) {
			continue
		}
		byName[def.Name] = sanyXMLDefinitionSource{name: def.Name, module: mod, def: def}
	}
	return sortedDefinitionSources(byName)
}

func (x *sanyXMLExporter) exportedDefinitionSources(mod *Module, visiting map[string]bool) []sanyXMLDefinitionSource {
	if mod == nil || visiting[mod.Name] {
		return nil
	}
	visiting[mod.Name] = true
	defer func() {
		visiting[mod.Name] = false
	}()

	byName := map[string]sanyXMLDefinitionSource{}
	for _, ext := range mod.Extends {
		for _, source := range x.exportedDefinitionSources(x.spec.Modules[ext], visiting) {
			source.fromExtends = true
			byName[source.name] = source
		}
	}
	for _, inst := range mod.Instances {
		if inst.Local {
			continue
		}
		target := x.spec.Modules[inst.Module]
		for _, source := range x.exportedDefinitionSources(target, visiting) {
			if source.def == nil || source.module == nil {
				continue
			}
			wrappers := append([]sanyXMLInstanceWrapper(nil), source.wrappers...)
			wrappers = append(wrappers, sanyXMLInstanceWrapper{owner: mod, inst: inst, target: target})
			if inst.exportsUnqualified() {
				source.wrappers = wrappers
				source.fromExtends = false
				byName[source.name] = source
			}
			if qualifier := inst.qualifier(); qualifier != "" {
				byName[qualifier+"!"+source.name] = sanyXMLDefinitionSource{
					name:        qualifier + "!" + source.name,
					module:      source.module,
					def:         source.def,
					wrappers:    wrappers,
					fromExtends: false,
				}
			}
		}
	}
	for i := range mod.Definitions {
		def := &mod.Definitions[i]
		if def.Local || def.TheoremLike {
			continue
		}
		byName[def.Name] = sanyXMLDefinitionSource{name: def.Name, module: mod, def: def}
	}

	return sortedDefinitionSources(byName)
}

func sortedDefinitionSources(byName map[string]sanyXMLDefinitionSource) []sanyXMLDefinitionSource {
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]sanyXMLDefinitionSource, 0, len(names))
	for _, name := range names {
		out = append(out, byName[name])
	}
	return out
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
	for instIndex, inst := range mod.Instances {
		if inst.Local {
			continue
		}
		for _, source := range x.instanceDefinitionSources(inst) {
			if x.skipInstanceDefinitionClone(mod, inst, source) {
				if source.fromExtends {
					add(x.instanceDefinitionSourceOriginalSymbol(source))
				}
				continue
			}
			if x.duplicateInstanceDefinitionClone(mod, instIndex, inst, source) {
				continue
			}
			sym := x.instDefs[x.instanceDefKey(mod.Name, instIndex, inst, source.keyName)]
			if sym != nil && sym.Kind == "TheoremDefNode" {
				continue
			}
			add(sym)
		}
	}
	for i, assume := range mod.Assumptions {
		add(x.assumes[fmt.Sprintf("assume:%s:%d:%s", mod.Name, i, assume.Name)])
	}
	for i, theorem := range mod.Theorems {
		add(x.theorems[fmt.Sprintf("theorem:%s:%d:%s", mod.Name, i, theorem.Name)])
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

func (x *sanyXMLExporter) emitModuleInstanceKindEntry(sym *sanyXMLSymbol) {
	if sym == nil || x.emitted[sym.Key] {
		return
	}
	x.emitted[sym.Key] = true
	var b bytes.Buffer
	b.WriteString("<ModuleInstanceKind>")
	x.writeLocation(&b, sym.Pos)
	b.WriteString("<uniquename>")
	xmlText(&b, sym.Name)
	b.WriteString("</uniquename></ModuleInstanceKind>")
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
	ownRecursiveSection := x.moduleRecursiveDefinitionSection(ctx.module, sym)
	if ownRecursiveSection == 0 && ctx.localRecursiveDefs != nil {
		ownRecursiveSection = ctx.localRecursiveDefs[sym.Name]
	}
	if ownRecursiveSection > 0 {
		defCtx.recursiveSection = ownRecursiveSection
	}
	levelData := x.exprLevelData(def.Expr, defCtx, nil)
	level := levelData.level
	x.setOperatorLevelData(sym, def, levelData)
	var body string
	var diags Diagnostics
	if fcn, ok := def.Expr.(*FunctionExpr); ok && def.FunctionDef {
		if x.isRecursiveFunctionDefinition(def) {
			body, diags = x.recursiveFunctionSpecXML(def, fcn, defCtx, level)
		} else {
			body, diags = x.nonRecursiveFunctionSpecXML(def, fcn, defCtx, level)
		}
	} else {
		body, diags = x.exprXML(def.Expr, defCtx)
	}
	if diags.HasErrors() {
		return diags
	}
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
	b.WriteString("</body>")
	sym.Leibniz = x.definitionLeibnizArgs(sym, def, defCtx)
	x.writeLeibnizParams(&b, sym.Params, sym.Leibniz)
	x.writePreComments(&b, def.PreComments)
	if def.Local {
		b.WriteString("<local/>")
	}
	if ownRecursiveSection > 0 {
		b.WriteString("<recursive/>")
		b.WriteString("<recursiveSection>")
		xmlInt(&b, ownRecursiveSection)
		b.WriteString("</recursiveSection>")
	} else if ctx.recursiveSection > 0 {
		b.WriteString("<recursiveSection>")
		xmlInt(&b, ctx.recursiveSection)
		b.WriteString("</recursiveSection>")
	}
	b.WriteString("</UserDefinedOpKind>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return diags
}

func (x *sanyXMLExporter) emitInstanceDefinitionEntry(sym *sanyXMLSymbol, original *sanyXMLSymbol, owner *Module, inst Instance, sourceMod *Module, targetMod *Module, def *Definition, ctx sanyXMLExprContext, wrappers []sanyXMLInstanceWrapper) Diagnostics {
	if sym == nil || original == nil || def == nil || x.emitted[sym.Key] {
		return nil
	}
	x.emitted[sym.Key] = true
	defCtx := ctx
	defCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	defCtx.suppressLetDefs = true
	for i, param := range def.Params {
		if i < len(original.Params) {
			defCtx.formals[param] = original.Params[i]
		}
	}
	levelData := x.exprLevelData(def.Expr, defCtx, nil)
	level := levelData.level
	var body string
	var diags Diagnostics
	if def.AssumeProveBody != nil {
		body, diags = x.assumeProveXML(def.AssumeProveBody, defCtx)
		level = x.assumeProveLevel(def.AssumeProveBody, defCtx)
		levelData.level = level
	} else if fcn, ok := def.Expr.(*FunctionExpr); ok && def.FunctionDef {
		if x.isRecursiveFunctionDefinition(def) {
			body, diags = x.recursiveFunctionSpecXML(def, fcn, defCtx, level)
		} else {
			body, diags = x.nonRecursiveFunctionSpecXML(def, fcn, defCtx, level)
		}
	} else {
		body, diags = x.exprXML(def.Expr, defCtx)
	}
	if diags.HasErrors() {
		return diags
	}
	substTag := "SubstInNode"
	if sym.Kind == "TheoremDefNode" {
		substTag = "APSubstInNode"
	}
	for _, wrapper := range wrappers {
		substs, hasSubsts, substDiags := x.instanceSubstitutionsXML(wrapper.owner, wrapper.inst)
		diags = append(diags, substDiags...)
		if diags.HasErrors() {
			return diags
		}
		if hasSubsts {
			body = x.substInXMLWithTag(substTag, wrapper.inst.SourcePosition(), level, substs, body, wrapper.owner, wrapper.target)
			sourceMod = wrapper.owner
		}
	}
	substs, hasSubsts, substDiags := x.instanceSubstitutionsXML(owner, inst)
	diags = append(diags, substDiags...)
	if diags.HasErrors() {
		return diags
	}
	if hasSubsts {
		body = x.substInXMLWithTag(substTag, inst.SourcePosition(), level, substs, body, owner, targetMod)
	}
	instanceParamCount := len(x.instanceParamSymbols(owner, inst))
	x.setInstanceOperatorLevelData(sym, def, levelData, instanceParamCount)
	for _, param := range sym.Params {
		x.emitFormalEntry(param)
	}
	if sym.Kind == "TheoremDefNode" {
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
	var b bytes.Buffer
	b.WriteString("<UserDefinedOpKind>")
	x.writeNode(&b, sym.Pos, level)
	b.WriteString("<uniquename>")
	xmlText(&b, sym.Name)
	b.WriteString("</uniquename><arity>")
	xmlInt(&b, sym.Arity)
	b.WriteString("</arity>")
	originModule := sourceMod
	if hasSubsts {
		originModule = owner
	}
	x.writeDefinitionOriginFor(&b, original, originModule)
	b.WriteString("<body>")
	b.WriteString(body)
	b.WriteString("</body>")
	sym.Leibniz = x.definitionLeibnizArgsWithOffset(sym, def, defCtx, instanceParamCount)
	x.writeLeibnizParams(&b, sym.Params, sym.Leibniz)
	x.writePreComments(&b, inst.PreComments)
	if inst.Local {
		b.WriteString("<local/>")
	}
	b.WriteString("</UserDefinedOpKind>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return diags
}

func (x *sanyXMLExporter) writeLeibnizParams(b *bytes.Buffer, params []*sanyXMLSymbol, leibniz []bool) {
	b.WriteString("<params>")
	for i, param := range params {
		b.WriteString("<leibnizparam>")
		x.writeRef(b, param)
		if i >= len(leibniz) || leibniz[i] {
			b.WriteString("<leibniz/>")
		}
		b.WriteString("</leibnizparam>")
	}
	b.WriteString("</params>")
}

func (x *sanyXMLExporter) definitionLeibnizArgs(sym *sanyXMLSymbol, def *Definition, ctx sanyXMLExprContext) []bool {
	return x.definitionLeibnizArgsWithOffset(sym, def, ctx, 0)
}

func (x *sanyXMLExporter) definitionLeibnizArgsWithOffset(sym *sanyXMLSymbol, def *Definition, ctx sanyXMLExprContext, paramOffset int) []bool {
	count := len(sym.Params)
	leibniz := make([]bool, count)
	for i := range leibniz {
		leibniz[i] = true
	}
	if def == nil || def.Expr == nil {
		return leibniz
	}
	use := x.exprParamUse(def.Expr, ctx, nil)
	for i, param := range def.Params {
		idx := paramOffset + i
		if idx >= count {
			break
		}
		if use.nonLeibniz[param] {
			leibniz[idx] = false
		}
	}
	return leibniz
}

func (x *sanyXMLExporter) lambdaLeibnizArgs(sym *sanyXMLSymbol, body Expr, ctx sanyXMLExprContext) []bool {
	leibniz := make([]bool, len(sym.Params))
	for i := range leibniz {
		leibniz[i] = true
	}
	use := x.exprParamUse(body, ctx, nil)
	for i, param := range sym.Params {
		if use.nonLeibniz[param.Name] {
			leibniz[i] = false
		}
	}
	return leibniz
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

func (x *sanyXMLExporter) recursiveDefinitionSection(mod *Module, name string) int {
	if mod == nil {
		return 0
	}
	return recursiveDeclarationSections(mod.Recursives)[name]
}

func letRecursiveDefinitionSections(expr *LetExpr) map[string]int {
	if expr == nil || len(expr.Recursives) == 0 {
		return nil
	}
	return recursiveDeclarationSections(expr.Recursives)
}

func recursiveDeclarationSections(decls []Declaration) map[string]int {
	sections := map[string]int{}
	section := 0
	var prev Declaration
	for i, decl := range decls {
		if i == 0 || !recursiveDeclarationsAreContiguous(prev, decl) {
			section++
		}
		for _, name := range decl.Names {
			if name != "" {
				sections[name] = section
			}
		}
		prev = decl
	}
	return sections
}

func recursiveDeclarationsAreContiguous(prev, next Declaration) bool {
	prevEnd := prev.Pos.EndLine
	if prevEnd == 0 {
		prevEnd = prev.Pos.Line
	}
	return prevEnd > 0 && next.Pos.Line == prevEnd+1
}

func (x *sanyXMLExporter) moduleRecursiveDefinitionSection(mod *Module, sym *sanyXMLSymbol) int {
	if mod == nil || sym == nil || x.defs[x.defKey(mod.Name, sym.Name)] != sym {
		return 0
	}
	return x.recursiveDefinitionSection(mod, sym.Name)
}

func (x *sanyXMLExporter) isRecursiveFunctionDefinition(def *Definition) bool {
	if def == nil {
		return false
	}
	if _, ok := def.Expr.(*FunctionExpr); !ok {
		return false
	}
	return exprReferencesName(def.Expr, def.Name, nil)
}

func (x *sanyXMLExporter) emitAssumeEntry(sym *sanyXMLSymbol, defSym *sanyXMLSymbol, assume NamedExpr, ctx sanyXMLExprContext) Diagnostics {
	if sym == nil || x.emitted[sym.Key] {
		return nil
	}
	var body string
	var diags Diagnostics
	if assume.AssumeProveBody != nil {
		body, diags = x.assumeProveXML(assume.AssumeProveBody, ctx)
	} else {
		body, diags = x.exprXML(assume.Expr, ctx)
	}
	if diags.HasErrors() {
		return diags
	}
	level := x.exprLevel(assume.Expr, ctx)
	if assume.AssumeProveBody != nil {
		level = x.assumeProveLevel(assume.AssumeProveBody, ctx)
	}
	if defSym != nil && !x.emitted[defSym.Key] {
		x.emitted[defSym.Key] = true
		defSym.Level = level
		var def bytes.Buffer
		def.WriteString("<AssumeDef>")
		x.writeNode(&def, assume.SourcePosition(), level)
		def.WriteString("<uniquename>")
		xmlText(&def, assume.Name)
		def.WriteString("</uniquename>")
		def.WriteString(body)
		def.WriteString("</AssumeDef>")
		x.entries = append(x.entries, sanyXMLEntry{key: defSym.Key, uid: defSym.UID, body: def.String()})
	}
	x.emitted[sym.Key] = true
	sym.Level = level
	var b bytes.Buffer
	if defSym != nil {
		b.WriteString("<AssumeNode>")
		x.writeNode(&b, assume.SourcePosition(), level)
		b.WriteString("<definition>")
		x.writeRef(&b, defSym)
		b.WriteString("</definition><body>")
		b.WriteString(body)
		b.WriteString("</body></AssumeNode>")
		x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
		return diags
	}
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
	assumeProveBody := theorem.AssumeProveBody
	if def != nil && def.AssumeProveBody != nil {
		assumeProveBody = def.AssumeProveBody
	}
	var body string
	var diags Diagnostics
	if assumeProveBody != nil {
		body, diags = x.assumeProveXML(assumeProveBody, ctx)
	} else {
		body, diags = x.exprXML(theorem.Expr, ctx)
	}
	if diags.HasErrors() {
		return diags
	}
	level := x.exprLevel(theorem.Expr, ctx)
	if assumeProveBody != nil {
		level = x.assumeProveLevel(assumeProveBody, ctx)
	}
	proofCtx := ctx
	if assumeProveBody != nil {
		proofCtx = x.withAssumeProveNewSymbols(proofCtx, assumeProveBody)
	}
	proofCtx.proofDefs = x.proofDefinitionMap(theorem.Syntax)
	proofNode := sanyXMLTheoremProofNode(theorem.Syntax)
	proof, proofDiags := x.proofXML(proofNode, proofCtx)
	diags = append(diags, proofDiags...)
	if diags.HasErrors() {
		return diags
	}
	level = maxTlaLevel(level, x.proofNodeLevel(proofNode, proofCtx))
	sym.Level = level
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
	stepCtx.proofDefs = copySanyXMLSymbolMap(ctx.proofDefs)
	for _, step := range steps {
		if defSym := x.proofDefs[step]; defSym != nil {
			if name := sanyXMLProofStepName(step); name != "" {
				stepCtx.proofDefs[name] = defSym
			}
		}
	}
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
			item, itemLevel, itemDiags := x.useOrHideXML(body, stepCtx, sanyNodePosition(step))
			diags = append(diags, itemDiags...)
			level = maxTlaLevel(level, itemLevel)
			stepXML = append(stepXML, item)
		case body != nil && body.Kind.JavaName() == "N_DefStep":
			item, itemLevel, itemDiags := x.defStepXML(step, body, stepCtx)
			diags = append(diags, itemDiags...)
			level = maxTlaLevel(level, itemLevel)
			stepXML = append(stepXML, item)
			if syms := x.proofLocalDefs[step]; len(syms) > 0 {
				stepCtx.defs = copySanyXMLSymbolMap(stepCtx.defs)
				for _, sym := range syms {
					stepCtx.defs[sym.Name] = sym
				}
			}
		}
		if ap, ok := sanyXMLProofStepSufficesAssumeProveBody(body); ok {
			stepCtx = x.withAssumeProveNewSymbols(stepCtx, ap)
		}
		if body != nil && body.Kind.JavaName() == "N_PickStep" {
			stepCtx = x.withProofStepBounds(stepCtx, sanyProofStepBounds(body))
		}
	}
	if diags.HasErrors() {
		return "", diags
	}
	pos := sanyNodePosition(proof)
	if pos.Line == 0 && pos.Column == 0 {
		pos = sanyXMLSpanPosition(steps)
	}
	var b bytes.Buffer
	b.WriteString("<steps>")
	x.writeNode(&b, pos, level)
	if proofLevel := sanyXMLProofLevel(proof); proofLevel > 0 {
		b.WriteString("<proofLevel>")
		xmlInt(&b, proofLevel)
		b.WriteString("</proofLevel>")
	}
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
	if defSym := x.proofDefs[step]; defSym != nil {
		proofCtx.proofDefs = copySanyXMLSymbolMap(proofCtx.proofDefs)
		proofCtx.proofDefs[sanyXMLProofStepName(step)] = defSym
	}
	proofLevel := x.proofNodeLevel(proof, proofCtx)
	level := maxTlaLevel(bodyLevel, proofLevel)
	sym.Level = level
	proofBody, proofDiags := x.proofXML(proof, proofCtx)
	diags = append(diags, proofDiags...)
	if diags.HasErrors() {
		return diags
	}
	proofLevel = maxTlaLevel(proofLevel, x.proofNodeLevel(proof, proofCtx))
	level = maxTlaLevel(bodyLevel, proofLevel)
	sym.Level = level

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
			xml, diags := x.assumeProveXMLWithSuffices(ap, ctx, sanyXMLProofStepSuffices(bodyNode))
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
		level := x.exprLevel(expr, ctx)
		if sanyXMLProofStepSuffices(bodyNode) {
			xml = x.opApplXML(sanyNodePosition(bodyNode), level, x.builtin("$Suffices"), []string{xml}, "")
		}
		return xml, level, nil
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
		level := x.exprLevel(expr, ctx)
		return x.opApplXML(sanyNodePosition(bodyNode), level, x.builtin("$Pfcase"), []string{xml}, ""), level, nil
	case "N_PickStep":
		exprNode := lastSanyExpression(bodyNode)
		expr, exprDiags := sanyExpr(exprNode)
		if exprDiags.HasErrors() {
			return "", constantLevel, exprDiags
		}
		bounds := sanyProofStepBounds(bodyNode)
		xml, xmlDiags := x.boundOpXML("$Pick", sanyNodePosition(bodyNode), bounds, expr, ctx, true)
		if xmlDiags.HasErrors() {
			return "", constantLevel, xmlDiags
		}
		return xml, x.boundOpLevel(bounds, expr, ctx), nil
	case "N_WitnessStep":
		exprNodes := expressionChildren(bodyNode)
		operands := make([]string, 0, len(exprNodes))
		level := constantLevel
		for _, exprNode := range exprNodes {
			expr, exprDiags := sanyExpr(exprNode)
			if exprDiags.HasErrors() {
				return "", constantLevel, exprDiags
			}
			operand, xmlDiags := x.exprXML(expr, ctx)
			if xmlDiags.HasErrors() {
				return "", constantLevel, xmlDiags
			}
			operands = append(operands, operand)
			level = maxTlaLevel(level, x.exprLevel(expr, ctx))
		}
		return x.opApplXML(sanyNodePosition(bodyNode), level, x.builtin("$Witness"), operands, ""), level, nil
	default:
		return "", constantLevel, Diagnostics{errorAt(sanyNodePosition(bodyNode), "E7010", "unsupported proof step XML body %s", bodyNode.Kind.JavaName())}
	}
}

func (x *sanyXMLExporter) defStepXML(step, bodyNode *SanySyntaxNode, ctx sanyXMLExprContext) (string, tlaLevel, Diagnostics) {
	defs, diags := sanyXMLDefStepDefinitions(bodyNode)
	if len(defs) == 0 {
		return "", constantLevel, diags
	}
	syms := x.proofLocalDefs[step]
	if len(syms) != len(defs) {
		return "", constantLevel, Diagnostics{errorAt(sanyNodePosition(bodyNode), "E7010", "proof DEFINE step has no allocated symbol")}
	}
	defCtx := ctx
	defCtx.defs = copySanyXMLSymbolMap(ctx.defs)
	level := constantLevel
	for i := range defs {
		sym := syms[i]
		diags = append(diags, x.emitDefinitionEntry(sym, &defs[i], defCtx)...)
		if diags.HasErrors() {
			return "", constantLevel, diags
		}
		level = maxTlaLevel(level, sym.Level)
		defCtx.defs[sym.Name] = sym
	}
	var b bytes.Buffer
	b.WriteString("<DefStepNode>")
	x.writeNode(&b, sanyNodePosition(bodyNode), level)
	for _, sym := range syms {
		x.writeRef(&b, sym)
	}
	b.WriteString("</DefStepNode>")
	return b.String(), level, nil
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
			if body != nil && body.Kind.JavaName() == "N_DefStep" {
				var defLevel tlaLevel
				defLevel, stepCtx = x.proofDefStepLevelAndContext(step, body, stepCtx)
				level = maxTlaLevel(level, defLevel)
			} else {
				level = maxTlaLevel(level, x.proofStepLevel(step, stepCtx))
			}
			if ap, ok := sanyXMLProofStepSufficesAssumeProveBody(body); ok {
				stepCtx = x.withAssumeProveNewSymbols(stepCtx, ap)
			}
		}
		return level
	default:
		return constantLevel
	}
}

func (x *sanyXMLExporter) proofDefStepLevelAndContext(step, bodyNode *SanySyntaxNode, ctx sanyXMLExprContext) (tlaLevel, sanyXMLExprContext) {
	defs, diags := sanyXMLDefStepDefinitions(bodyNode)
	syms := x.proofLocalDefs[step]
	if diags.HasErrors() || len(defs) == 0 || len(syms) != len(defs) {
		return constantLevel, ctx
	}
	defCtx := ctx
	defCtx.defs = copySanyXMLSymbolMap(ctx.defs)
	level := constantLevel
	for i := range defs {
		sym := syms[i]
		defLevel := x.exprLevel(defs[i].Expr, defCtx)
		sym.Level = defLevel
		level = maxTlaLevel(level, defLevel)
		defCtx.defs[sym.Name] = sym
	}
	return level, defCtx
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
			x.writeNode(&b, sanyNodePosition(proof), constantLevel)
			b.WriteString("</obvious>")
			return b.String(), nil
		case SanyTokenOmitted:
			var b bytes.Buffer
			b.WriteString("<omitted>")
			x.writeNode(&b, sanyNodePosition(proof), constantLevel)
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
	facts, defs, only, diags := x.proofCommandRefs(proof, ctx)
	if diags.HasErrors() {
		return "", diags
	}
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
	return x.proofCommandLevel(proof, ctx)
}

func (x *sanyXMLExporter) proofCommandLevel(node *SanySyntaxNode, ctx sanyXMLExprContext) tlaLevel {
	level := constantLevel
	inDefs := false
	heirs := node.GetHeirs()
	for i := 0; i < len(heirs); i++ {
		child := heirs[i]
		if child == nil {
			continue
		}
		if child.Token == nil {
			if inDefs {
				continue
			}
			expr, diags := sanyExpr(child)
			if !diags.HasErrors() {
				level = maxTlaLevel(level, x.exprLevel(expr, ctx))
			}
			continue
		}
		switch child.Token.Kind {
		case SanyTokenDF:
			inDefs = true
		case SanyTokenBy, SanyTokenUse, SanyTokenHide, SanyTokenComma, SanyTokenOnly:
			continue
		case SanyTokenIdentifier:
			if inDefs {
				continue
			}
			if call, next, ok, diags := sanyXMLProofCommandCallExpr(heirs, i); ok {
				if !diags.HasErrors() {
					level = maxTlaLevel(level, x.exprLevel(call, ctx))
				}
				i = next
				continue
			}
			if sym := x.proofReferenceSymbol(child.Image, ctx); sym != nil {
				level = maxTlaLevel(level, sym.Level)
			}
		default:
			if !inDefs && isSanyProofStepStartKind(child.Token.Kind) {
				if sym := ctx.proofDefs[sanyXMLProofStepNameImage(child.Image)]; sym != nil {
					level = maxTlaLevel(level, sym.Level)
				}
			}
		}
	}
	return level
}

func (x *sanyXMLExporter) proofCommandRefs(node *SanySyntaxNode, ctx sanyXMLExprContext) ([]string, []string, bool, Diagnostics) {
	var facts []string
	var defs []string
	inDefs := false
	only := false
	var diags Diagnostics
	heirs := node.GetHeirs()
	for i := 0; i < len(heirs); i++ {
		child := heirs[i]
		if child == nil {
			continue
		}
		if child.Token == nil {
			if inDefs {
				if sym := x.proofDefinitionRefSymbol(child, ctx); sym != nil {
					var b bytes.Buffer
					x.writeRef(&b, sym)
					defs = append(defs, b.String())
				}
				continue
			}
			expr, exprDiags := sanyExpr(child)
			diags = append(diags, exprDiags...)
			if exprDiags.HasErrors() {
				continue
			}
			fact, factDiags := x.exprXML(expr, ctx)
			diags = append(diags, factDiags...)
			if !factDiags.HasErrors() {
				facts = append(facts, fact)
			}
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
				if !inDefs {
					if sym := ctx.proofDefs[sanyXMLProofStepNameImage(child.Image)]; sym != nil {
						facts = append(facts, x.proofFactXML(sanyNodePosition(child), sym))
					}
				}
				continue
			}
			if child.Token.Kind != SanyTokenIdentifier {
				continue
			}
			if inDefs {
				if sym := x.proofDefinitionSymbol(child.Image, ctx); sym != nil {
					var b bytes.Buffer
					x.writeRef(&b, sym)
					defs = append(defs, b.String())
				}
			} else {
				if call, next, ok, callDiags := sanyXMLProofCommandCallExpr(heirs, i); ok {
					diags = append(diags, callDiags...)
					if !callDiags.HasErrors() {
						fact, factDiags := x.callXML(call, ctx)
						diags = append(diags, factDiags...)
						if !factDiags.HasErrors() {
							facts = append(facts, fact)
						}
					}
					i = next
					continue
				}
				if sym := x.proofReferenceSymbol(child.Image, ctx); sym != nil {
					facts = append(facts, x.proofFactXML(sanyNodePosition(child), sym))
				}
			}
		}
	}
	return facts, defs, only, diags
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

func (x *sanyXMLExporter) proofDefinitionRefSymbol(node *SanySyntaxNode, ctx sanyXMLExprContext) *sanyXMLSymbol {
	if expr, diags := sanyExpr(node); !diags.HasErrors() {
		switch e := expr.(type) {
		case *IdentExpr:
			return x.proofDefinitionSymbol(e.Name, ctx)
		case *CallExpr:
			if ident, ok := e.Callee.(*IdentExpr); ok {
				return x.proofDefinitionSymbol(ident.Name, ctx)
			}
		}
	}
	if id := firstSanyIdentifier(node); id != nil {
		return x.proofDefinitionSymbol(id.Image, ctx)
	}
	return nil
}

func (x *sanyXMLExporter) proofDefinitionSymbol(name string, ctx sanyXMLExprContext) *sanyXMLSymbol {
	if sym := ctx.proofDefs[name]; sym != nil {
		return sym
	}
	if sym := ctx.defs[name]; sym != nil {
		return sym
	}
	if sym := ctx.scope.defs[name]; sym != nil {
		return sym
	}
	return nil
}

func sanyXMLProofCommandCallExpr(heirs []*SanySyntaxNode, start int) (*CallExpr, int, bool, Diagnostics) {
	if start < 0 || start+1 >= len(heirs) {
		return nil, start, false, nil
	}
	nameNode := heirs[start]
	if nameNode == nil || nameNode.Token == nil || nameNode.Token.Kind != SanyTokenIdentifier {
		return nil, start, false, nil
	}
	open := heirs[start+1]
	if open == nil || open.Token == nil || open.Token.Kind != SanyTokenLbr {
		return nil, start, false, nil
	}
	var args []Expr
	var diags Diagnostics
	end := start + 1
	for i := start + 2; i < len(heirs); i++ {
		child := heirs[i]
		if child == nil {
			continue
		}
		if child.Token != nil {
			switch child.Token.Kind {
			case SanyTokenRbr:
				end = i
				pos := sanyXMLSpanPosition(heirs[start : end+1])
				return &CallExpr{
					Callee: &IdentExpr{Name: nameNode.Image, Pos: sanyNodePosition(nameNode)},
					Args:   args,
					Pos:    pos,
				}, end, true, diags
			case SanyTokenComma:
				continue
			case SanyTokenNumberLiteral:
				args = append(args, &LiteralExpr{Kind: "number", Value: child.Image, Pos: sanyNodePosition(child)})
				continue
			case SanyTokenStringLiteral:
				args = append(args, &LiteralExpr{Kind: "string", Value: child.Image, Pos: sanyNodePosition(child)})
				continue
			case SanyTokenIdentifier:
				args = append(args, &IdentExpr{Name: child.Image, Pos: sanyNodePosition(child)})
				continue
			default:
				diags = append(diags, errorAt(sanyNodePosition(child), "E7011", "unsupported proof command call argument token %s", child.Token.Kind.JavaName()))
				continue
			}
		}
		expr, exprDiags := sanyExpr(child)
		diags = append(diags, exprDiags...)
		if !exprDiags.HasErrors() {
			args = append(args, expr)
		}
	}
	return nil, start, false, nil
}

func (x *sanyXMLExporter) proofFactXML(pos Position, sym *sanyXMLSymbol) string {
	return x.opApplXML(pos, sym.Level, sym, nil, "")
}

func (x *sanyXMLExporter) useOrHideXML(node *SanySyntaxNode, ctx sanyXMLExprContext, pos Position) (string, tlaLevel, Diagnostics) {
	level := x.useOrHideLevel(node, ctx)
	facts, defs, only, diags := x.proofCommandRefs(node, ctx)
	if diags.HasErrors() {
		return "", constantLevel, diags
	}
	if pos.Line == 0 && pos.Column == 0 && pos.File == "" {
		pos = sanyNodePosition(node)
	}
	var b bytes.Buffer
	b.WriteString("<UseOrHideNode>")
	x.writeNode(&b, pos, level)
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
	return x.proofCommandLevel(node, ctx)
}

func (x *sanyXMLExporter) assumeProveXML(body *AssumeProve, ctx sanyXMLExprContext) (string, Diagnostics) {
	return x.assumeProveXMLWithSuffices(body, ctx, false)
}

func (x *sanyXMLExporter) assumeProveXMLWithSuffices(body *AssumeProve, ctx sanyXMLExprContext, suffices bool) (string, Diagnostics) {
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
			level := item.NewSymbol.Level
			if item.NewSymbol.Domain != nil {
				level = maxTlaLevel(level, x.exprLevel(item.NewSymbol.Domain, apCtx))
			}
			x.writeNode(&b, item.NewSymbol.Source, level)
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
	b.WriteString("</prove>")
	if suffices {
		b.WriteString("<suffices></suffices>")
	}
	b.WriteString("</AssumeProveNode>")
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

func (x *sanyXMLExporter) withProofStepBounds(ctx sanyXMLExprContext, bounds []BoundVar) sanyXMLExprContext {
	if len(bounds) == 0 {
		return ctx
	}
	next := ctx
	next.formals = copySanyXMLSymbolMap(ctx.formals)
	for _, bound := range bounds {
		next.formals[bound.Name] = x.newBoundFormal("expr", bound.Name, bound.Pos)
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
		if sym.LevelKnown {
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
		UID:        uid,
		Key:        "builtin:" + info.name,
		Kind:       "BuiltInKind",
		Name:       info.name,
		Arity:      info.arity,
		Level:      info.level,
		ArgWeights: append([]int(nil), info.weights...),
		leveled:    true,
		Pos:        Position{File: "--TLA+ BUILTINS--", Line: 0, Column: 0, EndLine: 0, EndColumn: 0},
	}
	x.builtins[info.name] = sym
	for i := 0; i < info.arity; i++ {
		param := &sanyXMLSymbol{
			UID:   1_000_000_000 + uid*100 + i + 1,
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
		if e.Op == "/\\" || e.Op == "\\/" {
			oper := "$ConjList"
			if e.Op == "\\/" {
				oper = "$DisjList"
			}
			return x.opApplXML(e.Pos, x.exprLevel(e, ctx), x.builtin(oper), []string{operand}, ""), nil
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
		if e.IsLambda {
			return x.lambdaExprXML(e, ctx)
		}
		return x.boundOpXML("$FcnConstructor", e.Pos, e.Bounds, e.Body, ctx, false)
	case *FunctionAppExpr:
		fn, diags := x.functionApplicationFunctionXML(e, ctx)
		if diags.HasErrors() {
			return "", diags
		}
		args := []string{fn}
		if len(e.Args) > 1 {
			tuple, tupleDiags := x.exprListOpXML("$Tuple", e.Args, e.Pos, ctx)
			diags = append(diags, tupleDiags...)
			args = append(args, tuple)
		} else {
			for _, arg := range e.Args {
				argXML, argDiags := x.exprXML(arg, ctx)
				diags = append(diags, argDiags...)
				args = append(args, argXML)
			}
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
		return x.boundOpXML(oper, e.Pos, e.Bounds, body, ctx, false)
	default:
		return "", Diagnostics{errorAt(expr.Position(), "E7002", "unsupported SANY XML expression %T", expr)}
	}
}

func (x *sanyXMLExporter) identXML(e *IdentExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	if e.Name == "@" && ctx.exceptAtBase != "" && ctx.exceptAtComponents != "" {
		return x.atXML(ctx.exceptAtPos, ctx.exceptAtLevel, ctx.exceptAtBase, ctx.exceptAtComponents), nil
	}
	if xmlText, ok, diags := x.subexpressionReferenceXML(e, ctx); ok {
		return xmlText, diags
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

func (x *sanyXMLExporter) subexpressionReferenceXML(e *IdentExpr, ctx sanyXMLExprContext) (string, bool, Diagnostics) {
	selected, ok, diags := x.subexpressionReferenceExpr(e.Name, e.Pos, ctx)
	if !ok || diags.HasErrors() {
		return "", ok, diags
	}
	operand, operandDiags := x.exprXML(selected, ctx)
	if operandDiags.HasErrors() {
		return "", true, operandDiags
	}
	return x.opApplXML(e.Pos, x.exprLevel(selected, ctx), x.builtin("$Nop"), []string{operand}, ""), true, nil
}

func (x *sanyXMLExporter) subexpressionReferenceExpr(name string, pos Position, ctx sanyXMLExprContext) (Expr, bool, Diagnostics) {
	if base, ok := sanyBodySelectorBase(name); ok {
		defSym := x.definitionSymbol(base, ctx)
		if defSym == nil {
			return nil, false, nil
		}
		def := x.definitionForSymbol(defSym)
		if def == nil || def.Expr == nil {
			return nil, true, Diagnostics{errorAt(pos, "E7003", "cannot resolve base definition %s for subexpression %s", base, name)}
		}
		return def.Expr, true, nil
	}
	base, selectors, ok := sanySubexpressionPath(name)
	if !ok {
		return nil, false, nil
	}
	defSym := x.definitionSymbol(base, ctx)
	if defSym == nil {
		return nil, false, nil
	}
	def := x.definitionForSymbol(defSym)
	if def == nil || def.Expr == nil {
		return nil, true, Diagnostics{errorAt(pos, "E7003", "cannot resolve base definition %s for subexpression %s", base, name)}
	}
	selected := sanySelectSubexpression(def.Expr, selectors)
	if selected == nil {
		return nil, true, Diagnostics{errorAt(pos, "E7003", "cannot resolve subexpression %s", name)}
	}
	return selected, true, nil
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
		if nested, ok := expr.(*BinaryExpr); ok && sanySameJunctionFrame(e, nested) {
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
	for i, arg := range e.Args {
		argXML, argDiags := x.callArgumentXML(arg, i, operator, ctx)
		diags = append(diags, argDiags...)
		args = append(args, argXML)
	}
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(e.Pos, x.exprLevel(e, ctx), operator, args, ""), nil
}

func (x *sanyXMLExporter) callArgumentXML(arg Expr, index int, operator *sanyXMLSymbol, ctx sanyXMLExprContext) (string, Diagnostics) {
	if operator == nil || index >= len(operator.Params) || operator.Params[index].Arity <= 0 {
		return x.exprXML(arg, ctx)
	}
	if ident, ok := arg.(*IdentExpr); ok {
		if sym := x.operatorSymbol(ident.Name, ctx); sym != nil {
			return x.opArgXML(ident.Pos, x.exprLevel(arg, ctx), sym), nil
		}
	}
	return x.exprXML(arg, ctx)
}

func (x *sanyXMLExporter) lambdaForQuantifiedDefinitionCall(e *CallExpr, name string, ctx sanyXMLExprContext) (*sanyXMLSymbol, Diagnostics) {
	if len(e.Args) == 0 {
		return nil, nil
	}
	defSym := x.definitionSymbol(name, ctx)
	body := Expr(nil)
	lambdaKeyName := name
	if defSym == nil && strings.Contains(name, "!") {
		base, selectors, ok := sanySubexpressionPath(name)
		if ok {
			defSym = x.definitionSymbol(base, ctx)
			if def := x.definitionForSymbol(defSym); def != nil {
				body = sanySelectSubexpression(def.Expr, selectors)
				lambdaKeyName = name
			}
		}
	}
	if defSym == nil || defSym.Arity != 0 {
		return nil, nil
	}
	if body == nil {
		def := x.definitionForSymbol(defSym)
		if def == nil {
			return nil, nil
		}
		body = def.Expr
	}
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
	key := fmt.Sprintf("lambda:%s:%s:%d:%d:%d:%d:%d", ctx.module.Name, lambdaKeyName, e.Pos.Line, e.Pos.Column, e.Pos.EndLine, e.Pos.EndColumn, len(params))
	if sym := x.lambdas[key]; sym != nil {
		return sym, nil
	}
	sym := x.newSymbol("UserDefinedOpKind", key, "LAMBDA", len(params), constantLevel, e.Pos)
	sym.Params = params
	x.lambdas[key] = sym
	diags := x.emitLambdaEntry(sym, body, lambdaCtx, nil)
	if diags.HasErrors() {
		return nil, diags
	}
	return sym, nil
}

func sanySameJunctionFrame(parent, nested *BinaryExpr) bool {
	return parent != nil &&
		nested != nil &&
		parent.JunctionList &&
		nested.JunctionList &&
		parent.Op == nested.Op &&
		parent.Pos.Line == nested.Pos.Line &&
		parent.Pos.Column == nested.Pos.Column
}

func (x *sanyXMLExporter) emitLambdaEntry(sym *sanyXMLSymbol, body Expr, ctx sanyXMLExprContext, preComments []string) Diagnostics {
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
	b.WriteString("</body>")
	sym.Leibniz = x.lambdaLeibnizArgs(sym, body, ctx)
	x.writeLeibnizParams(&b, sym.Params, sym.Leibniz)
	x.writePreComments(&b, preComments)
	b.WriteString("</UserDefinedOpKind>")
	x.entries = append(x.entries, sanyXMLEntry{key: sym.Key, uid: sym.UID, body: b.String()})
	return nil
}

func (x *sanyXMLExporter) lambdaExprXML(fcn *FunctionExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	lambdaCtx := ctx
	lambdaCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	params := make([]*sanyXMLSymbol, 0, len(fcn.Bounds))
	for _, bound := range fcn.Bounds {
		formal := x.newBoundFormal("lambda", bound.Name, bound.Pos)
		formal.LevelKnown = exprReferencesName(fcn.Body, bound.Name, nil)
		x.emitFormalEntry(formal)
		lambdaCtx.formals[bound.Name] = formal
		params = append(params, formal)
	}
	key := fmt.Sprintf("lambdaexpr:%s:%d:%d:%d:%d:%d", ctx.module.Name, fcn.Pos.Line, fcn.Pos.Column, fcn.Pos.EndLine, fcn.Pos.EndColumn, len(params))
	if sym := x.lambdas[key]; sym != nil {
		return x.opArgXML(fcn.Pos, sym.Level, sym), nil
	}
	sym := x.newSymbol("UserDefinedOpKind", key, "LAMBDA", len(params), constantLevel, fcn.Pos)
	sym.Params = params
	x.lambdas[key] = sym
	diags := x.emitLambdaEntry(sym, fcn.Body, lambdaCtx, fcn.PreComments)
	if diags.HasErrors() {
		return "", diags
	}
	return x.opArgXML(fcn.Pos, sym.Level, sym), nil
}

func (x *sanyXMLExporter) letXML(e *LetExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	letCtx := ctx
	letCtx.defs = copySanyXMLSymbolMap(ctx.defs)
	letCtx.scope = copySanyXMLScope(ctx.scope)
	letCtx.localRecursiveDefs = letRecursiveDefinitionSections(e)
	localDefs := x.letDefs[e]
	if len(localDefs) != len(e.Definitions) {
		localDefs = make([]*sanyXMLSymbol, 0, len(e.Definitions))
		for i := range e.Definitions {
			def := &e.Definitions[i]
			sym := x.newLocalDefinitionSymbol("let", def)
			localDefs = append(localDefs, sym)
		}
		x.letDefs[e] = localDefs
	}
	for i := range e.Definitions {
		if i < len(localDefs) {
			letCtx.defs[e.Definitions[i].Name] = localDefs[i]
		}
	}
	localInsts := make([]*sanyXMLSymbol, 0, len(e.Instances))
	localInstDefs := make([]*sanyXMLSymbol, 0)
	type localInstanceSource struct {
		inst   Instance
		source sanyXMLInstanceDefinitionSource
		sym    *sanyXMLSymbol
	}
	var localSources []localInstanceSource
	for _, inst := range e.Instances {
		x.allocateInstanceParams(ctx.module, inst)
		instKey := x.letInstanceKindKey(ctx.module.Name, e, inst)
		instSym := x.letInsts[instKey]
		if instSym == nil {
			instSym = x.newSymbol("ModuleInstanceKind", instKey, inst.Name, 0, constantLevel, inst.SourcePosition())
			x.letInsts[instKey] = instSym
		}
		localInsts = append(localInsts, instSym)
		for _, source := range x.instanceDefinitionSources(inst) {
			defKey := x.letInstanceDefKey(ctx.module.Name, e, inst, source.keyName)
			sym := x.letInstDefs[defKey]
			if sym == nil {
				sym = x.newInstanceDefinitionSymbol(defKey, source, inst.SourcePosition(), x.instanceParamSymbols(ctx.module, inst))
				x.letInstDefs[defKey] = sym
			}
			letCtx.scope.defs[source.cloneName] = sym
			letCtx.scope.declKinds[source.cloneName] = OperatorDecl
			localInstDefs = append(localInstDefs, sym)
			localSources = append(localSources, localInstanceSource{inst: inst, source: source, sym: sym})
		}
	}
	var diags Diagnostics
	if !ctx.suppressLetDefs {
		for i := range e.Definitions {
			if i < len(localDefs) {
				diags = append(diags, x.emitDefinitionEntry(localDefs[i], &e.Definitions[i], letCtx)...)
			}
		}
		for _, instSym := range localInsts {
			x.emitModuleInstanceKindEntry(instSym)
		}
		sourceContexts := map[string]sanyXMLExprContext{}
		for _, item := range localSources {
			if item.source.module == nil || item.source.def == nil {
				continue
			}
			sourceCtx, ok := sourceContexts[item.source.module.Name]
			if !ok {
				sourceCtx = sanyXMLExprContext{module: item.source.module, scope: x.scopeForModule(item.source.module, map[string]bool{}), formals: map[string]*sanyXMLSymbol{}, defs: map[string]*sanyXMLSymbol{}, proofDefs: map[string]*sanyXMLSymbol{}}
				sourceContexts[item.source.module.Name] = sourceCtx
			}
			targetMod := x.spec.Modules[item.inst.Module]
			original := x.defs[x.defKey(item.source.module.Name, item.source.def.Name)]
			diags = append(diags, x.emitInstanceDefinitionEntry(item.sym, original, ctx.module, item.inst, item.source.module, targetMod, item.source.def, sourceCtx, item.source.wrappers)...)
		}
	}
	body, bodyDiags := x.exprXML(e.Body, letCtx)
	diags = append(diags, bodyDiags...)
	if diags.HasErrors() {
		return "", diags
	}
	var b bytes.Buffer
	b.WriteString("<LetInNode>")
	x.writeNode(&b, e.Pos, x.exprLevel(e.Body, letCtx))
	b.WriteString("<body>")
	b.WriteString(body)
	b.WriteString("</body><opDefs>")
	for _, sym := range localDefs {
		x.writeRef(&b, sym)
	}
	for _, sym := range localInstDefs {
		x.writeRef(&b, sym)
	}
	for _, sym := range localInsts {
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

	body := e.Body
	bounds := []BoundVar{{Name: e.Var, Set: e.Set, Pos: quantifierVarPosition(e)}}
	for {
		next, ok := body.(*QuantifierExpr)
		if !ok || next.Kind != e.Kind || !samePosition(next.Pos, e.Pos) {
			break
		}
		bounds = append(bounds, BoundVar{Name: next.Var, Set: next.Set, Pos: quantifierVarPosition(next)})
		body = next.Body
	}
	return x.boundOpXML(oper, e.Pos, bounds, body, ctx, false)
}

func quantifierVarPosition(e *QuantifierExpr) Position {
	if e.VarPos.Line > 0 || e.VarPos.Column > 0 || e.VarPos.File != "" {
		return e.VarPos
	}
	return e.Pos
}

func samePosition(left, right Position) bool {
	return left.File == right.File &&
		left.Line == right.Line &&
		left.Column == right.Column &&
		left.EndLine == right.EndLine &&
		left.EndColumn == right.EndColumn
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
	return x.boundOpXML(oper, e.Pos, bounds, e.Body, ctx, false)
}

func (x *sanyXMLExporter) boundOpXML(oper string, pos Position, bounds []BoundVar, body Expr, ctx sanyXMLExprContext, forceFormalLevel bool) (string, Diagnostics) {
	boundCtx := ctx
	boundCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	var boundSymbols bytes.Buffer
	boundSymbols.WriteString("<boundSymbols>")
	var diags Diagnostics
	level := constantLevel
	for i := 0; i < len(bounds); {
		set := bounds[i].Set
		if set == nil {
			bound := bounds[i]
			formal := x.newBoundFormal("expr", bound.Name, bound.Pos)
			formal.LevelKnown = forceFormalLevel || exprReferencesName(body, bound.Name, nil)
			x.emitFormalEntry(formal)
			boundCtx.formals[bound.Name] = formal
			boundSymbols.WriteString("<unbound>")
			x.writeRef(&boundSymbols, formal)
			boundSymbols.WriteString("</unbound>")
			i++
			continue
		}
		setXML := ""
		var setDiags Diagnostics
		setXML, setDiags = x.exprXML(set, ctx)
		diags = append(diags, setDiags...)
		level = maxTlaLevel(level, x.exprLevel(set, ctx))
		var refs bytes.Buffer
		j := i
		for j < len(bounds) && bounds[j].Set == set {
			bound := bounds[j]
			formal := x.newBoundFormal("expr", bound.Name, bound.Pos)
			formal.LevelKnown = forceFormalLevel || exprReferencesName(body, bound.Name, nil)
			x.emitFormalEntry(formal)
			boundCtx.formals[bound.Name] = formal
			x.writeRef(&refs, formal)
			j++
		}
		boundSymbols.WriteString("<bound>")
		boundSymbols.WriteString(refs.String())
		if sanyXMLBoundsHaveTuple(bounds[i:j]) {
			boundSymbols.WriteString("<tuple/>")
		}
		boundSymbols.WriteString(setXML)
		boundSymbols.WriteString("</bound>")
		i = j
	}
	boundSymbols.WriteString("</boundSymbols>")
	bodyXML, bodyDiags := x.exprXML(body, boundCtx)
	diags = append(diags, bodyDiags...)
	if diags.HasErrors() {
		return "", diags
	}
	level = maxTlaLevel(level, x.exprLevel(body, boundCtx))
	return x.opApplXML(pos, level, x.builtin(oper), []string{bodyXML}, boundSymbols.String()), nil
}

func (x *sanyXMLExporter) boundOpLevel(bounds []BoundVar, body Expr, ctx sanyXMLExprContext) tlaLevel {
	boundCtx := ctx
	boundCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	level := constantLevel
	for _, bound := range bounds {
		if bound.Set != nil {
			level = maxTlaLevel(level, x.exprLevel(bound.Set, ctx))
		}
		formal := &sanyXMLSymbol{Name: bound.Name, Kind: "FormalParamNode", Level: constantLevel}
		boundCtx.formals[bound.Name] = formal
	}
	return maxTlaLevel(level, x.exprLevel(body, boundCtx))
}

func sanyXMLBoundsHaveTuple(bounds []BoundVar) bool {
	for _, bound := range bounds {
		if bound.TupleBound {
			return true
		}
	}
	return false
}

func (x *sanyXMLExporter) recursiveFunctionSpecXML(def *Definition, fcn *FunctionExpr, ctx sanyXMLExprContext, level tlaLevel) (string, Diagnostics) {
	boundCtx := ctx
	boundCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	self := x.newBoundFormal("recfcn", def.Name, def.SourcePosition())
	x.emitFormalEntry(self)
	boundCtx.formals[def.Name] = self

	var boundSymbols bytes.Buffer
	boundSymbols.WriteString("<boundSymbols><unbound>")
	x.writeRef(&boundSymbols, self)
	boundSymbols.WriteString("</unbound>")

	var diags Diagnostics
	for i := 0; i < len(fcn.Bounds); {
		set := fcn.Bounds[i].Set
		if set == nil {
			bound := fcn.Bounds[i]
			formal := x.newBoundFormal("expr", bound.Name, bound.Pos)
			formal.LevelKnown = exprReferencesName(fcn.Body, bound.Name, nil)
			x.emitFormalEntry(formal)
			boundCtx.formals[bound.Name] = formal
			boundSymbols.WriteString("<unbound>")
			x.writeRef(&boundSymbols, formal)
			boundSymbols.WriteString("</unbound>")
			i++
			continue
		}

		setXML, setDiags := x.exprXML(set, ctx)
		diags = append(diags, setDiags...)
		var refs bytes.Buffer
		j := i
		for j < len(fcn.Bounds) && fcn.Bounds[j].Set == set {
			bound := fcn.Bounds[j]
			formal := x.newBoundFormal("expr", bound.Name, bound.Pos)
			formal.LevelKnown = exprReferencesName(fcn.Body, bound.Name, nil)
			x.emitFormalEntry(formal)
			boundCtx.formals[bound.Name] = formal
			x.writeRef(&refs, formal)
			j++
		}
		boundSymbols.WriteString("<bound>")
		boundSymbols.WriteString(refs.String())
		if sanyXMLBoundsHaveTuple(fcn.Bounds[i:j]) {
			boundSymbols.WriteString("<tuple/>")
		}
		boundSymbols.WriteString(setXML)
		boundSymbols.WriteString("</bound>")
		i = j
	}
	boundSymbols.WriteString("</boundSymbols>")

	bodyXML, bodyDiags := x.exprXML(fcn.Body, boundCtx)
	diags = append(diags, bodyDiags...)
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(def.SourcePosition(), level, x.builtin("$RecursiveFcnSpec"), []string{bodyXML}, boundSymbols.String()), nil
}

func (x *sanyXMLExporter) nonRecursiveFunctionSpecXML(def *Definition, fcn *FunctionExpr, ctx sanyXMLExprContext, level tlaLevel) (string, Diagnostics) {
	boundCtx := ctx
	boundCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	var boundSymbols bytes.Buffer
	boundSymbols.WriteString("<boundSymbols>")

	var diags Diagnostics
	for i := 0; i < len(fcn.Bounds); {
		set := fcn.Bounds[i].Set
		if set == nil {
			bound := fcn.Bounds[i]
			formal := x.newBoundFormal("expr", bound.Name, bound.Pos)
			formal.LevelKnown = exprReferencesName(fcn.Body, bound.Name, nil)
			x.emitFormalEntry(formal)
			boundCtx.formals[bound.Name] = formal
			boundSymbols.WriteString("<unbound>")
			x.writeRef(&boundSymbols, formal)
			boundSymbols.WriteString("</unbound>")
			i++
			continue
		}

		setXML, setDiags := x.exprXML(set, ctx)
		diags = append(diags, setDiags...)
		var refs bytes.Buffer
		j := i
		for j < len(fcn.Bounds) && fcn.Bounds[j].Set == set {
			bound := fcn.Bounds[j]
			formal := x.newBoundFormal("expr", bound.Name, bound.Pos)
			formal.LevelKnown = exprReferencesName(fcn.Body, bound.Name, nil)
			x.emitFormalEntry(formal)
			boundCtx.formals[bound.Name] = formal
			x.writeRef(&refs, formal)
			j++
		}
		boundSymbols.WriteString("<bound>")
		boundSymbols.WriteString(refs.String())
		if sanyXMLBoundsHaveTuple(fcn.Bounds[i:j]) {
			boundSymbols.WriteString("<tuple/>")
		}
		boundSymbols.WriteString(setXML)
		boundSymbols.WriteString("</bound>")
		i = j
	}
	boundSymbols.WriteString("</boundSymbols>")

	bodyXML, bodyDiags := x.exprXML(fcn.Body, boundCtx)
	diags = append(diags, bodyDiags...)
	if diags.HasErrors() {
		return "", diags
	}
	return x.opApplXML(def.SourcePosition(), level, x.builtin("$NonRecursiveFcnSpec"), []string{bodyXML}, boundSymbols.String()), nil
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

func (x *sanyXMLExporter) functionApplicationFunctionXML(e *FunctionAppExpr, ctx sanyXMLExprContext) (string, Diagnostics) {
	if ident, ok := e.Function.(*IdentExpr); ok {
		if ident.Name == "@" {
			return x.exprXML(e.Function, ctx)
		}
		return x.opApplXML(ident.Pos, x.exprLevel(ident, ctx), x.operatorSymbol(ident.Name, ctx), nil, ""), nil
	}
	return x.exprXML(e.Function, ctx)
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
			if len(component.Indices) > 1 {
				tupleXML, tupleDiags := x.exprListOpXML("$Tuple", component.Indices, component.Pos, ctx)
				diags = append(diags, tupleDiags...)
				componentArgs = append(componentArgs, tupleXML)
				componentLevel = maxTlaLevel(componentLevel, x.exprLevel(&TupleExpr{Elems: component.Indices, Pos: component.Pos}, ctx))
			} else {
				for _, index := range component.Indices {
					indexXML, indexDiags := x.exprXML(index, ctx)
					diags = append(diags, indexDiags...)
					componentArgs = append(componentArgs, indexXML)
					componentLevel = maxTlaLevel(componentLevel, x.exprLevel(index, ctx))
				}
			}
			if component.Field != "" {
				field := x.stringXML(component.Field, apalacheRecordFieldPosition(component.FieldPos, component.Pos))
				componentArgs = append(componentArgs, field)
			}
		}
		components := x.opApplXML(spec.Pos, componentLevel, x.builtin("$Seq"), componentArgs, "")
		valueCtx := ctx
		valueCtx.exceptAtBase = base
		valueCtx.exceptAtComponents = components
		valueCtx.exceptAtPos = spec.Pos
		valueCtx.exceptAtLevel = maxTlaLevel(baseLevel, componentLevel)
		value, valueDiags := x.exprXML(spec.Value, valueCtx)
		diags = append(diags, valueDiags...)
		valueLevel := x.exprLevel(spec.Value, valueCtx)
		args = append(args, x.opApplXML(spec.Pos, maxTlaLevel(componentLevel, valueLevel), x.builtin("$Pair"), []string{components, value}, ""))
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

func (x *sanyXMLExporter) opArgXML(pos Position, level tlaLevel, argument *sanyXMLSymbol) string {
	var b bytes.Buffer
	b.WriteString("<OpArgNode>")
	x.writeNode(&b, pos, level)
	b.WriteString("<argument>")
	x.writeRef(&b, argument)
	b.WriteString("</argument></OpArgNode>")
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

func (x *sanyXMLExporter) operatorSymbolForLeibniz(name string, ctx sanyXMLExprContext) *sanyXMLSymbol {
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
	return nil
}

func (x *sanyXMLExporter) operatorLeibnizArg(name string, index int, ctx sanyXMLExprContext) bool {
	if sym := x.operatorSymbolForLeibniz(name, ctx); sym != nil {
		return x.symbolLeibnizArg(sym, index, ctx)
	}
	name = sanyXMLBuiltinName(name)
	if !sanyXMLKnownBuiltin(name) {
		return true
	}
	info := sanyXMLBuiltin(name)
	return index >= len(info.leibniz) || info.leibniz[index]
}

func sanyXMLKnownBuiltin(name string) bool {
	if builtinIdentifiers[name] {
		return true
	}
	if _, ok := GetSanyOperator(name); ok {
		return true
	}
	switch name {
	case "'", "\\prime", "ENABLED", "UNCHANGED", "[]", "<>", "\\lnot", "SUBSET", "UNION", "DOMAIN", "\\cdot", "~>", "-+->",
		"$AngleAct", "$BoundedChoose", "$BoundedExists", "$BoundedForall", "$Case", "$ConjList", "$DisjList", "$Except", "$FcnApply", "$FcnConstructor", "$IfThenElse", "$NonRecursiveFcnSpec", "$Nop", "$Pair", "$Pfcase", "$Pick", "$Qed", "$RecursiveFcnSpec", "$RcdConstructor", "$RcdSelect", "$Seq", "$SetEnumerate", "$SetOfAll", "$SetOfFcns", "$SetOfRcds", "$SquareAct", "$SubsetOf", "$Suffices", "$TemporalExists", "$TemporalForall", "$Tuple", "$UnboundedChoose", "$UnboundedExists", "$UnboundedForall", "$WF", "$SF":
		return true
	default:
		return false
	}
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
	if sym.Def != nil {
		return sym.Def
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

func (x *sanyXMLExporter) exprParamUse(expr Expr, ctx sanyXMLExprContext, shadowed map[string]bool) sanyXMLParamUse {
	if expr == nil {
		return sanyXMLParamUse{}
	}
	switch e := expr.(type) {
	case *IdentExpr:
		return x.identParamUse(e.Name, ctx, shadowed)
	case *LiteralExpr:
		return sanyXMLParamUse{}
	case *UnaryExpr:
		use := x.exprParamUse(e.Expr, ctx, shadowed)
		if !x.operatorLeibnizArg(e.Op, 0, ctx) {
			use.addNonLeibniz(use.all)
		}
		return use
	case *BinaryExpr:
		left := x.exprParamUse(e.Left, ctx, shadowed)
		right := x.exprParamUse(e.Right, ctx, shadowed)
		use := mergeSanyXMLParamUse(left, right, x.identParamUse(e.Op, ctx, shadowed))
		if !x.operatorLeibnizArg(e.Op, 0, ctx) {
			use.addNonLeibniz(left.all)
		}
		if !x.operatorLeibnizArg(e.Op, 1, ctx) {
			use.addNonLeibniz(right.all)
		}
		return use
	case *CallExpr:
		use := x.exprParamUse(e.Callee, ctx, shadowed)
		operatorName := ""
		if ident, ok := e.Callee.(*IdentExpr); ok {
			operatorName = ident.Name
		}
		for i, arg := range e.Args {
			argUse := x.exprParamUse(arg, ctx, shadowed)
			use.merge(argUse)
			if operatorName != "" && !x.operatorLeibnizArg(operatorName, i, ctx) {
				use.addNonLeibniz(argUse.all)
			}
		}
		return use
	case *IfExpr:
		return mergeSanyXMLParamUse(
			x.exprParamUse(e.Cond, ctx, shadowed),
			x.exprParamUse(e.Then, ctx, shadowed),
			x.exprParamUse(e.Else, ctx, shadowed),
		)
	case *LetExpr:
		use := sanyXMLParamUse{}
		for _, def := range e.Definitions {
			defShadowed := copyBoolMap(shadowed)
			for _, param := range def.Params {
				defShadowed[param] = true
			}
			use.merge(x.exprParamUse(def.Expr, ctx, defShadowed))
		}
		for _, inst := range e.Instances {
			for _, expr := range inst.Substitutions {
				use.merge(x.exprParamUse(expr, ctx, shadowed))
			}
		}
		bodyShadowed := copyBoolMap(shadowed)
		for _, def := range e.Definitions {
			bodyShadowed[def.Name] = true
		}
		use.merge(x.exprParamUse(e.Body, ctx, bodyShadowed))
		use.nonLeibniz = nil
		return use
	case *QuantifierExpr:
		use := x.exprParamUse(e.Set, ctx, shadowed)
		bodyShadowed := copyBoolMap(shadowed)
		bodyShadowed[e.Var] = true
		use.merge(x.exprParamUse(e.Body, ctx, bodyShadowed))
		return use
	case *CaseExpr:
		use := sanyXMLParamUse{}
		for _, arm := range e.Arms {
			use.merge(x.exprParamUse(arm.Test, ctx, shadowed))
			use.merge(x.exprParamUse(arm.Value, ctx, shadowed))
		}
		use.merge(x.exprParamUse(e.Other, ctx, shadowed))
		return use
	case *ChooseExpr:
		use := x.exprParamUse(e.Set, ctx, shadowed)
		bodyShadowed := copyBoolMap(shadowed)
		bodyShadowed[e.Var] = true
		use.merge(x.exprParamUse(e.Body, ctx, bodyShadowed))
		return use
	case *TupleExpr:
		return x.exprListParamUse(e.Elems, ctx, shadowed)
	case *SetExpr:
		return x.exprListParamUse(e.Elems, ctx, shadowed)
	case *RecordExpr:
		use := sanyXMLParamUse{}
		for _, field := range e.Fields {
			use.merge(x.exprParamUse(field.Value, ctx, shadowed))
		}
		return use
	case *RecordComponentExpr:
		return x.exprParamUse(e.Record, ctx, shadowed)
	case *RecordSetExpr:
		use := sanyXMLParamUse{}
		for _, field := range e.Fields {
			use.merge(x.exprParamUse(field.Set, ctx, shadowed))
		}
		return use
	case *FunctionExpr:
		use := sanyXMLParamUse{}
		bodyShadowed := copyBoolMap(shadowed)
		for _, bound := range e.Bounds {
			use.merge(x.exprParamUse(bound.Set, ctx, shadowed))
			bodyShadowed[bound.Name] = true
		}
		use.merge(x.exprParamUse(e.Body, ctx, bodyShadowed))
		return use
	case *FunctionAppExpr:
		use := x.exprParamUse(e.Function, ctx, shadowed)
		use.merge(x.exprListParamUse(e.Args, ctx, shadowed))
		return use
	case *ExceptExpr:
		use := x.exprParamUse(e.Base, ctx, shadowed)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				use.merge(x.exprListParamUse(component.Indices, ctx, shadowed))
			}
			use.merge(x.exprParamUse(spec.Value, ctx, shadowed))
		}
		return use
	case *LabelExpr:
		return x.exprParamUse(e.Body, ctx, shadowed)
	case *ActionExpr:
		actionUse := x.exprParamUse(e.Action, ctx, shadowed)
		subscriptUse := x.exprParamUse(e.Subscript, ctx, shadowed)
		use := mergeSanyXMLParamUse(actionUse, subscriptUse)
		use.addNonLeibniz(actionUse.all)
		use.addNonLeibniz(subscriptUse.all)
		return use
	case *FairnessExpr:
		subscriptUse := x.exprParamUse(e.Subscript, ctx, shadowed)
		actionUse := x.exprParamUse(e.Action, ctx, shadowed)
		use := mergeSanyXMLParamUse(subscriptUse, actionUse)
		use.addNonLeibniz(subscriptUse.all)
		use.addNonLeibniz(actionUse.all)
		return use
	case *FunctionSetExpr:
		return mergeSanyXMLParamUse(x.exprParamUse(e.Domain, ctx, shadowed), x.exprParamUse(e.Range, ctx, shadowed))
	case *SetComprehensionExpr:
		use := sanyXMLParamUse{}
		bodyShadowed := copyBoolMap(shadowed)
		for _, bound := range e.Bounds {
			use.merge(x.exprParamUse(bound.Set, ctx, shadowed))
			bodyShadowed[bound.Name] = true
		}
		use.merge(x.exprParamUse(e.Element, ctx, bodyShadowed))
		use.merge(x.exprParamUse(e.Predicate, ctx, bodyShadowed))
		return use
	}
	return sanyXMLParamUse{}
}

func (x *sanyXMLExporter) identParamUse(name string, ctx sanyXMLExprContext, shadowed map[string]bool) sanyXMLParamUse {
	if name == "" || shadowed[name] || ctx.formals[name] == nil {
		return sanyXMLParamUse{}
	}
	use := sanyXMLParamUse{}
	use.addAll(name)
	return use
}

func (x *sanyXMLExporter) exprListParamUse(exprs []Expr, ctx sanyXMLExprContext, shadowed map[string]bool) sanyXMLParamUse {
	use := sanyXMLParamUse{}
	for _, expr := range exprs {
		use.merge(x.exprParamUse(expr, ctx, shadowed))
	}
	return use
}

func (x *sanyXMLExporter) symbolLeibnizArg(sym *sanyXMLSymbol, index int, ctx sanyXMLExprContext) bool {
	if sym == nil || index < 0 {
		return true
	}
	if sym.Kind == "BuiltInKind" {
		info := sanyXMLBuiltin(sym.Name)
		return index >= len(info.leibniz) || info.leibniz[index]
	}
	if index < len(sym.Leibniz) {
		return sym.Leibniz[index]
	}
	if def := x.definitionForSymbol(sym); def != nil && def != sym.Def {
		sym.Leibniz = x.definitionLeibnizArgs(sym, def, ctx)
		if index < len(sym.Leibniz) {
			return sym.Leibniz[index]
		}
	}
	return true
}

func mergeSanyXMLParamUse(uses ...sanyXMLParamUse) sanyXMLParamUse {
	var out sanyXMLParamUse
	for _, use := range uses {
		out.merge(use)
	}
	return out
}

func (u *sanyXMLParamUse) merge(other sanyXMLParamUse) {
	for name := range other.all {
		u.addAll(name)
	}
	u.addNonLeibniz(other.nonLeibniz)
}

func (u *sanyXMLParamUse) addAll(name string) {
	if name == "" {
		return
	}
	if u.all == nil {
		u.all = map[string]bool{}
	}
	u.all[name] = true
}

func (u *sanyXMLParamUse) addNonLeibniz(names map[string]bool) {
	for name := range names {
		if name == "" {
			continue
		}
		u.addAll(name)
		if u.nonLeibniz == nil {
			u.nonLeibniz = map[string]bool{}
		}
		u.nonLeibniz[name] = true
	}
}

func (x *sanyXMLExporter) writeRef(b *bytes.Buffer, sym *sanyXMLSymbol) {
	if sym == nil {
		return
	}
	ref := "BuiltInKindRef"
	switch sym.Kind {
	case "ModuleNode":
		ref = "ModuleNodeRef"
	case "ModuleInstanceKind":
		ref = "ModuleInstanceKindRef"
	case "OpDeclNode":
		ref = "OpDeclNodeRef"
	case "UserDefinedOpKind":
		ref = "UserDefinedOpKindRef"
	case "FormalParamNode":
		ref = "FormalParamNodeRef"
	case "AssumeNode":
		ref = "AssumeNodeRef"
	case "AssumeDef":
		ref = "AssumeDefRef"
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
	x.writeDefinitionOriginFor(b, sym, mod)
}

func (x *sanyXMLExporter) writeDefinitionOriginFor(b *bytes.Buffer, sym *sanyXMLSymbol, mod *Module) {
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
		comment = normalizeSanyNestedBlockPreComment(comment)
		comment = strings.TrimRight(comment, " \t\r\n")
		if comment == "" {
			continue
		}
		parts = append(parts, comment)
	}
	return strings.Join(parts, "\n")
}

func normalizeSanyNestedBlockPreComment(comment string) string {
	if strings.Count(comment, "(*") <= 1 {
		return comment
	}
	lines := strings.Split(comment, "\n")
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		open := strings.Index(line, "(*")
		if open < 0 {
			continue
		}
		prefix := line[:open]
		rest := line[open+2:]
		lines[i] = prefix + "(*\n" + rest
		if i+1 < len(lines) && lines[i+1] != "" {
			lines[i] += "\n"
		}
	}
	return strings.Join(lines, "\n")
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
	return x.exprLevelData(expr, ctx, nil).level
}

func (x *sanyXMLExporter) exprLevelData(expr Expr, ctx sanyXMLExprContext, shadowed map[string]bool) sanyXMLLevelData {
	switch e := expr.(type) {
	case *IdentExpr:
		if e.Name == "@" && ctx.exceptAtBase != "" && ctx.exceptAtComponents != "" {
			return sanyXMLLevelData{level: ctx.exceptAtLevel}
		}
		if shadowed[e.Name] {
			return sanyXMLLevelData{level: constantLevel}
		}
		if selected, ok, diags := x.subexpressionReferenceExpr(e.Name, e.Pos, ctx); ok && !diags.HasErrors() {
			return x.exprLevelData(selected, ctx, shadowed)
		}
		data := sanyXMLLevelData{level: x.operatorLevel(e.Name, ctx)}
		if ctx.formals[e.Name] != nil {
			data.addParam(e.Name)
		}
		return data
	case *LiteralExpr:
		return sanyXMLLevelData{level: constantLevel}
	case *UnaryExpr:
		return x.operatorApplicationLevelData(e.Op, x.resolvedOperatorSymbol(e.Op, ctx), []Expr{e.Expr}, ctx, shadowed)
	case *BinaryExpr:
		return x.binaryExprLevelData(e, ctx, shadowed)
	case *CallExpr:
		return x.callExprLevelData(e, ctx, shadowed)
	case *IfExpr:
		return x.operatorApplicationLevelData("$IfThenElse", nil, []Expr{e.Cond, e.Then, e.Else}, ctx, shadowed)
	case *LetExpr:
		data := x.exprLevelData(e.Body, ctx, shadowed)
		for i := range e.Definitions {
			data.merge(x.exprLevelData(e.Definitions[i].Expr, ctx, shadowed))
		}
		return data
	case *QuantifierExpr:
		bodyShadowed := copyBoolMap(shadowed)
		bodyShadowed[e.Var] = true
		return mergeSanyXMLLevelData(x.exprLevelData(e.Set, ctx, shadowed), x.exprLevelData(e.Body, ctx, bodyShadowed))
	case *CaseExpr:
		data := sanyXMLLevelData{level: constantLevel}
		for _, arm := range e.Arms {
			data.merge(x.exprLevelData(arm.Test, ctx, shadowed))
			data.merge(x.exprLevelData(arm.Value, ctx, shadowed))
		}
		if e.Other != nil {
			data.merge(x.exprLevelData(e.Other, ctx, shadowed))
		}
		return data
	case *ChooseExpr:
		bodyShadowed := copyBoolMap(shadowed)
		bodyShadowed[e.Var] = true
		return mergeSanyXMLLevelData(x.exprLevelData(e.Set, ctx, shadowed), x.exprLevelData(e.Body, ctx, bodyShadowed))
	case *TupleExpr:
		return x.operatorApplicationLevelData("$Tuple", nil, e.Elems, ctx, shadowed)
	case *SetExpr:
		return x.operatorApplicationLevelData("$SetEnumerate", nil, e.Elems, ctx, shadowed)
	case *RecordExpr:
		data := sanyXMLLevelData{level: constantLevel}
		for _, field := range e.Fields {
			data.merge(x.exprLevelData(field.Value, ctx, shadowed))
		}
		return data
	case *RecordComponentExpr:
		return x.exprLevelData(e.Record, ctx, shadowed)
	case *RecordSetExpr:
		data := sanyXMLLevelData{level: constantLevel}
		for _, field := range e.Fields {
			data.merge(x.exprLevelData(field.Set, ctx, shadowed))
		}
		return data
	case *FunctionExpr:
		data := sanyXMLLevelData{level: constantLevel}
		bodyShadowed := copyBoolMap(shadowed)
		for _, bound := range e.Bounds {
			data.merge(x.exprLevelData(bound.Set, ctx, shadowed))
			bodyShadowed[bound.Name] = true
		}
		data.merge(x.exprLevelData(e.Body, ctx, bodyShadowed))
		return data
	case *FunctionAppExpr:
		data := x.exprLevelData(e.Function, ctx, shadowed)
		for _, arg := range e.Args {
			data.merge(x.exprLevelData(arg, ctx, shadowed))
		}
		return data
	case *ExceptExpr:
		data := x.exprLevelData(e.Base, ctx, shadowed)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					data.merge(x.exprLevelData(index, ctx, shadowed))
				}
			}
			data.merge(x.exprLevelData(spec.Value, ctx, shadowed))
		}
		return data
	case *LabelExpr:
		return x.exprLevelData(e.Body, ctx, shadowed)
	case *ActionExpr:
		return sanyXMLLevelData{level: maxTlaLevel(actionLevel, maxTlaLevel(x.exprLevelData(e.Action, ctx, shadowed).level, x.exprLevelData(e.Subscript, ctx, shadowed).level))}
	case *FairnessExpr:
		return sanyXMLLevelData{level: temporalLevel}
	case *FunctionSetExpr:
		return mergeSanyXMLLevelData(x.exprLevelData(e.Domain, ctx, shadowed), x.exprLevelData(e.Range, ctx, shadowed))
	case *SetComprehensionExpr:
		data := sanyXMLLevelData{level: constantLevel}
		bodyShadowed := copyBoolMap(shadowed)
		for _, bound := range e.Bounds {
			data.merge(x.exprLevelData(bound.Set, ctx, shadowed))
			bodyShadowed[bound.Name] = true
		}
		data.merge(x.exprLevelData(e.Element, ctx, bodyShadowed))
		if e.Predicate != nil {
			data.merge(x.exprLevelData(e.Predicate, ctx, bodyShadowed))
		}
		return data
	default:
		return sanyXMLLevelData{level: constantLevel}
	}
}

func (x *sanyXMLExporter) binaryExprLevel(e *BinaryExpr, ctx sanyXMLExprContext) tlaLevel {
	return x.binaryExprLevelData(e, ctx, nil).level
}

func (x *sanyXMLExporter) callExprLevel(e *CallExpr, ctx sanyXMLExprContext) tlaLevel {
	return x.callExprLevelData(e, ctx, nil).level
}

func (x *sanyXMLExporter) binaryExprLevelData(e *BinaryExpr, ctx sanyXMLExprContext, shadowed map[string]bool) sanyXMLLevelData {
	if e == nil {
		return sanyXMLLevelData{level: constantLevel}
	}
	return x.operatorApplicationLevelData(e.Op, x.resolvedOperatorSymbol(e.Op, ctx), []Expr{e.Left, e.Right}, ctx, shadowed)
}

func (x *sanyXMLExporter) callExprLevelData(e *CallExpr, ctx sanyXMLExprContext, shadowed map[string]bool) sanyXMLLevelData {
	if e == nil {
		return sanyXMLLevelData{level: constantLevel}
	}
	operator := x.callExprOperator(e, ctx)
	name := ""
	if ident, ok := e.Callee.(*IdentExpr); ok {
		name = ident.Name
	}
	if operator != nil {
		name = operator.Name
	}
	if name == "" {
		data := x.exprLevelData(e.Callee, ctx, shadowed)
		for _, arg := range e.Args {
			data.merge(x.exprLevelData(arg, ctx, shadowed))
		}
		return data
	}
	return x.operatorApplicationLevelData(name, operator, e.Args, ctx, shadowed)
}

func (x *sanyXMLExporter) callExprOperator(e *CallExpr, ctx sanyXMLExprContext) *sanyXMLSymbol {
	if e == nil {
		return nil
	}
	ident, ok := e.Callee.(*IdentExpr)
	if !ok {
		return nil
	}
	if lambda, diags := x.lambdaForQuantifiedDefinitionCall(e, ident.Name, ctx); lambda != nil && !diags.HasErrors() {
		return lambda
	}
	return x.resolvedOperatorSymbol(ident.Name, ctx)
}

func (x *sanyXMLExporter) resolvedOperatorSymbol(name string, ctx sanyXMLExprContext) *sanyXMLSymbol {
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
	return nil
}

func (x *sanyXMLExporter) operatorApplicationLevelData(name string, operator *sanyXMLSymbol, args []Expr, ctx sanyXMLExprContext, shadowed map[string]bool) sanyXMLLevelData {
	data := x.operatorBaseLevelData(name, operator, ctx)
	if operator != nil && operator.Kind == "FormalParamNode" && !shadowed[operator.Name] {
		data.addParam(operator.Name)
	}
	for i, arg := range args {
		argData := x.exprLevelData(arg, ctx, shadowed)
		if operator == nil {
			if !sanyXMLKnownBuiltin(name) || sanyXMLBuiltinArgWeight(sanyXMLBuiltin(name), i) > 0 {
				data.merge(argData)
			}
			continue
		}
		if x.operatorArgWeight(operator, i, ctx) > 0 {
			data.merge(argData)
		}
	}
	return data
}

func (x *sanyXMLExporter) operatorBaseLevelData(name string, operator *sanyXMLSymbol, ctx sanyXMLExprContext) sanyXMLLevelData {
	if operator == nil || operator.Kind == "BuiltInKind" {
		info := sanyXMLBuiltin(name)
		return sanyXMLLevelData{level: info.level}
	}
	x.ensureOperatorLevelData(operator, ctx)
	data := sanyXMLLevelData{level: operator.Level}
	for param := range operator.LevelParams {
		data.addParam(param)
	}
	return data
}

func (x *sanyXMLExporter) operatorArgWeight(operator *sanyXMLSymbol, index int, ctx sanyXMLExprContext) int {
	if index < 0 {
		return 1
	}
	if operator == nil {
		return 1
	}
	if operator.Kind == "BuiltInKind" {
		return sanyXMLBuiltinArgWeight(sanyXMLBuiltin(operator.Name), index)
	}
	if operator.Kind == "OpDeclNode" || operator.Kind == "FormalParamNode" {
		return 1
	}
	x.ensureOperatorLevelData(operator, ctx)
	if index < len(operator.ArgWeights) {
		return operator.ArgWeights[index]
	}
	return 1
}

func (x *sanyXMLExporter) ensureOperatorLevelData(sym *sanyXMLSymbol, ctx sanyXMLExprContext) {
	if sym == nil || sym.leveled || sym.leveling {
		return
	}
	sym.leveling = true
	defer func() {
		sym.leveling = false
		sym.leveled = true
	}()
	def := x.definitionForSymbol(sym)
	if def == nil || def.Expr == nil {
		return
	}
	defCtx := ctx
	defCtx.formals = copySanyXMLSymbolMap(ctx.formals)
	for i, paramName := range def.Params {
		if i < len(sym.Params) {
			defCtx.formals[paramName] = sym.Params[i]
		}
	}
	data := x.exprLevelData(def.Expr, defCtx, nil)
	x.setOperatorLevelData(sym, def, data)
}

func (x *sanyXMLExporter) setOperatorLevelData(sym *sanyXMLSymbol, def *Definition, data sanyXMLLevelData) {
	if sym == nil {
		return
	}
	sym.Level = data.level
	sym.LevelParams = copyBoolMap(data.params)
	if def != nil {
		for _, paramName := range def.Params {
			delete(sym.LevelParams, paramName)
		}
		sym.ArgWeights = make([]int, len(sym.Params))
		for i, paramName := range def.Params {
			if i >= len(sym.ArgWeights) {
				break
			}
			if data.params[paramName] {
				sym.ArgWeights[i] = 1
			}
		}
	}
	sym.leveled = true
}

func (x *sanyXMLExporter) setInstanceOperatorLevelData(sym *sanyXMLSymbol, def *Definition, data sanyXMLLevelData, paramOffset int) {
	if sym == nil {
		return
	}
	sym.Level = data.level
	sym.LevelParams = copyBoolMap(data.params)
	for i, param := range sym.Params {
		delete(sym.LevelParams, param.Name)
		if i < paramOffset && sym.ArgWeights == nil {
			sym.ArgWeights = make([]int, len(sym.Params))
		}
	}
	if def != nil {
		if sym.ArgWeights == nil {
			sym.ArgWeights = make([]int, len(sym.Params))
		}
		for i := 0; i < paramOffset && i < len(sym.ArgWeights); i++ {
			sym.ArgWeights[i] = 1
		}
		for i, paramName := range def.Params {
			idx := paramOffset + i
			if idx >= len(sym.ArgWeights) {
				break
			}
			if data.params[paramName] {
				sym.ArgWeights[idx] = 1
			}
		}
	}
	sym.leveled = true
}

func (x *sanyXMLExporter) operatorLevel(name string, ctx sanyXMLExprContext) tlaLevel {
	if sym := ctx.formals[name]; sym != nil {
		x.ensureOperatorLevelData(sym, ctx)
		return sym.Level
	}
	if sym := ctx.defs[name]; sym != nil {
		x.ensureOperatorLevelData(sym, ctx)
		return sym.Level
	}
	if sym := ctx.scope.decls[name]; sym != nil {
		x.ensureOperatorLevelData(sym, ctx)
		return sym.Level
	}
	if sym := ctx.scope.defs[name]; sym != nil {
		x.ensureOperatorLevelData(sym, ctx)
		return sym.Level
	}
	info := sanyXMLBuiltin(name)
	return info.level
}

func mergeSanyXMLLevelData(items ...sanyXMLLevelData) sanyXMLLevelData {
	out := sanyXMLLevelData{level: constantLevel}
	for _, item := range items {
		out.merge(item)
	}
	return out
}

func (d *sanyXMLLevelData) merge(other sanyXMLLevelData) {
	d.level = maxTlaLevel(d.level, other.level)
	for param := range other.params {
		d.addParam(param)
	}
}

func (d *sanyXMLLevelData) addParam(param string) {
	if param == "" {
		return
	}
	if d.params == nil {
		d.params = map[string]bool{}
	}
	d.params[param] = true
}

func (x *sanyXMLExporter) scopeForModule(mod *Module, visiting map[string]bool) sanyXMLScope {
	return x.scopeForModuleMode(mod, visiting, true)
}

func (x *sanyXMLExporter) exportedScopeForModule(mod *Module, visiting map[string]bool) sanyXMLScope {
	return x.scopeForModuleMode(mod, visiting, false)
}

func (x *sanyXMLExporter) scopeForModuleMode(mod *Module, visiting map[string]bool, includeLocal bool) sanyXMLScope {
	scope := sanyXMLScope{
		decls:     map[string]*sanyXMLSymbol{},
		defs:      map[string]*sanyXMLSymbol{},
		declKinds: map[string]DeclarationKind{},
	}
	if mod == nil {
		return scope
	}
	visitKey := fmt.Sprintf("%s:%t", mod.Name, includeLocal)
	if visiting[visitKey] {
		return scope
	}
	visiting[visitKey] = true
	for _, ext := range mod.Extends {
		dep := x.spec.Modules[ext]
		depScope := x.exportedScopeForModule(dep, visiting)
		mergeSanyXMLScope(scope, depScope)
		if dep != nil {
			x.addModuleLocalScope(scope, dep, true, false)
		}
	}
	for instIndex, inst := range mod.Instances {
		x.addInstanceScope(scope, mod, instIndex, inst, includeLocal, visiting)
	}
	x.addModuleLocalScope(scope, mod, false, includeLocal)
	visiting[visitKey] = false
	return scope
}

func (x *sanyXMLExporter) addInstanceScope(scope sanyXMLScope, owner *Module, instIndex int, inst Instance, includeLocal bool, visiting map[string]bool) {
	if inst.Local && !includeLocal {
		return
	}
	instMod := x.spec.Modules[inst.Module]
	if instMod == nil {
		return
	}
	instScope := x.exportedScopeForModule(instMod, visiting)
	if inst.exportsUnqualified() {
		mergeSanyXMLScope(scope, instScope)
	}
	qualifier := inst.qualifier()
	if qualifier == "" {
		return
	}
	for name, sym := range instScope.decls {
		if strings.Contains(name, "!") {
			continue
		}
		scope.decls[qualifier+"!"+name] = sym
		if kind, ok := instScope.declKinds[name]; ok {
			scope.declKinds[qualifier+"!"+name] = kind
		}
	}
	for name, sym := range instScope.defs {
		scope.defs[qualifier+"!"+name] = x.instanceDefinitionScopeSymbol(owner, instIndex, inst, name, sym)
		scope.declKinds[qualifier+"!"+name] = OperatorDecl
	}
	if inst.exportsUnqualified() {
		for name, sym := range instScope.defs {
			scope.defs[name] = x.instanceDefinitionScopeSymbol(owner, instIndex, inst, name, sym)
			scope.declKinds[name] = OperatorDecl
		}
	}
}

func (x *sanyXMLExporter) instanceDefinitionScopeSymbol(owner *Module, instIndex int, inst Instance, name string, fallback *sanyXMLSymbol) *sanyXMLSymbol {
	if owner == nil {
		return fallback
	}
	if sym := x.instDefs[x.instanceDefKey(owner.Name, instIndex, inst, name)]; sym != nil {
		return sym
	}
	if inst.exportsUnqualified() {
		if sym := x.earlierInstanceDefinitionSymbol(owner, instIndex, name); sym != nil {
			return sym
		}
	}
	return fallback
}

func (x *sanyXMLExporter) addModuleLocalScope(scope sanyXMLScope, mod *Module, qualifiedOnly bool, includeLocal bool) {
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
		if def.Local && (qualifiedOnly || !includeLocal) {
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
	for i, assumption := range mod.Assumptions {
		if assumption.Name == "" {
			continue
		}
		key := fmt.Sprintf("assume:%s:%d:%s", mod.Name, i, assumption.Name)
		sym := x.assumeDefs[key]
		if sym == nil {
			continue
		}
		if !qualifiedOnly {
			scope.defs[assumption.Name] = sym
			scope.declKinds[assumption.Name] = OperatorDecl
		}
		scope.defs[mod.Name+"!"+assumption.Name] = sym
		scope.declKinds[mod.Name+"!"+assumption.Name] = OperatorDecl
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

func sanyXMLProofLevel(proof *SanySyntaxNode) int {
	for _, step := range sanyXMLDirectProofSteps(proof) {
		start := sanyXMLProofStepStartNode(step)
		if start == nil || start.Token == nil {
			continue
		}
		level, ok := sanyProofStepLevel(start.Token)
		if ok && level > 0 {
			return level
		}
	}
	return 0
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

func sanyXMLDefStepDefinitions(body *SanySyntaxNode) ([]Definition, Diagnostics) {
	if body == nil {
		return nil, nil
	}
	var defs []Definition
	var diags Diagnostics
	for _, child := range body.GetHeirs() {
		if child == nil {
			continue
		}
		switch child.Kind.JavaName() {
		case "N_OperatorDefinition":
			def, defDiags := sanyDefinition(child)
			diags = append(diags, defDiags...)
			if !defDiags.HasErrors() {
				defs = append(defs, def)
			}
		case "N_FunctionDefinition":
			def, defDiags := sanyFunctionDefinition(child)
			diags = append(diags, defDiags...)
			if !defDiags.HasErrors() {
				defs = append(defs, def)
			}
		}
	}
	if len(defs) == 0 && !diags.HasErrors() {
		diags = append(diags, errorAt(sanyNodePosition(body), "E7010", "proof DEFINE step has no definition"))
	}
	return defs, diags
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

func (x *sanyXMLExporter) instanceDefKey(module string, index int, inst Instance, name string) string {
	return fmt.Sprintf("instdef:%s:%d:%s:%s:%s", module, index, inst.qualifier(), inst.Module, name)
}

func (x *sanyXMLExporter) letInstanceKindKey(module string, let *LetExpr, inst Instance) string {
	pos := let.Position()
	instPos := inst.SourcePosition()
	return fmt.Sprintf("letinst:%s:%s:%s:%d:%d:%d:%d:%d:%d", module, inst.Name, inst.Module, pos.Line, pos.Column, pos.EndLine, pos.EndColumn, instPos.Line, instPos.Column)
}

func (x *sanyXMLExporter) letInstanceDefKey(module string, let *LetExpr, inst Instance, name string) string {
	return x.letInstanceKindKey(module, let, inst) + ":def:" + name
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
	case "\\X", "\\times":
		return "$CartesianProd"
	default:
		return name
	}
}

func sanyXMLBuiltin(name string) sanyXMLBuiltinInfo {
	name = sanyXMLBuiltinName(name)
	if info, ok := sanyBuiltinOperatorInfo(name); ok {
		return sanyXMLBuiltinInfo{
			name:    info.name,
			arity:   info.arity,
			level:   info.level,
			weights: append([]int(nil), info.argWeights...),
			leibniz: sanyBuiltinLeibniz(info),
		}
	}
	level := constantLevel
	arity := 0
	weights := []int{}
	unary := func(weight int) {
		arity = 1
		weights = []int{weight}
	}
	binary := func(left, right int) {
		arity = 2
		weights = []int{left, right}
	}
	variadic := func(weight int) {
		arity = -1
		weights = []int{weight}
	}
	switch name {
	case "TRUE", "FALSE", "BOOLEAN", "STRING", "$Qed":
		arity = 0
		weights = nil
	case "~>", "-+->":
		binary(0, 0)
		level = temporalLevel
	case "$IfThenElse":
		arity = 3
		weights = []int{1, 1, 1}
	case "\\lnot", "$Nop", "$Pfcase", "$Pick", "$Suffices", "$NonRecursiveFcnSpec", "$RecursiveFcnSpec", "SUBSET", "UNION", "DOMAIN":
		unary(1)
	case "'", "\\prime", "ENABLED", "UNCHANGED", "[]", "<>":
		unary(0)
		if name == "'" || name == "\\prime" || name == "UNCHANGED" {
			level = actionLevel
		}
		if name == "ENABLED" {
			level = variableLevel
		}
		if name == "[]" || name == "<>" {
			level = temporalLevel
		}
	case "$Pair", "$RcdSelect", "$FcnApply":
		binary(1, 1)
	case "$SquareAct", "$AngleAct":
		binary(0, 0)
		if name == "$SquareAct" || name == "$AngleAct" {
			level = actionLevel
		}
	case "\\cdot":
		binary(0, 0)
		level = actionLevel
	case "$SubsetOf":
		unary(1)
	case "$Case", "$ConjList", "$DisjList", "$Tuple", "$Seq", "$SetEnumerate", "$RcdConstructor", "$SetOfAll", "$SetOfFcns", "$FcnConstructor", "$BoundedForall", "$BoundedExists", "$BoundedChoose", "$SetOfRcds", "$Except", "$CartesianProd":
		variadic(1)
	case "$UnboundedForall", "$UnboundedExists", "$UnboundedChoose", "$TemporalExists", "$TemporalForall":
		unary(1)
		if name == "$TemporalExists" || name == "$TemporalForall" {
			level = temporalLevel
			weights[0] = 0
		}
	case "$WF", "$SF":
		binary(0, 0)
		level = temporalLevel
	default:
		binary(1, 1)
	}
	paramCount := arity
	if paramCount < 0 {
		paramCount = 0
	}
	leibniz := make([]bool, paramCount)
	for i := range leibniz {
		weight := 1
		if i < len(weights) {
			weight = weights[i]
		}
		leibniz[i] = weight > 0
	}
	return sanyXMLBuiltinInfo{name: name, arity: arity, level: level, weights: weights, leibniz: leibniz}
}

func sanyXMLBuiltinArgWeight(info sanyXMLBuiltinInfo, index int) int {
	if index < 0 {
		return 1
	}
	if len(info.weights) == 0 {
		return 0
	}
	if info.arity == -1 {
		return info.weights[0]
	}
	if index < len(info.weights) {
		return info.weights[index]
	}
	return 1
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
	case "$Pick":
		return 294
	case "$Suffices":
		return 297
	case "$RcdSelect":
		return 250
	case "$NonRecursiveFcnSpec":
		return 244
	case "$RecursiveFcnSpec":
		return 253
	case "$CartesianProd":
		return 231
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

func sanyXMLReservedBuiltinUID(uid int) bool {
	switch uid {
	case 1, 2, 4, 13, 15, 19, 22, 31, 40, 61, 63, 78, 79, 80, 82, 83, 84, 85, 89, 90, 96, 99, 105, 106, 107, 108, 113, 116, 125, 126, 137, 138, 231, 244, 250, 253, 294, 297:
		return true
	default:
		return false
	}
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
