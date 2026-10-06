package tlago

type sanySemKind int

const (
	sanyModuleKind            sanySemKind = 1
	sanyConstantDeclKind      sanySemKind = 2
	sanyVariableDeclKind      sanySemKind = 3
	sanyBoundSymbolKind       sanySemKind = 4
	sanyUserDefinedOpKind     sanySemKind = 5
	sanyModuleInstanceKind    sanySemKind = 6
	sanyBuiltInKind           sanySemKind = 7
	sanyOpArgKind             sanySemKind = 8
	sanyOpApplKind            sanySemKind = 9
	sanyLetInKind             sanySemKind = 10
	sanyFormalParamKind       sanySemKind = 11
	sanyTheoremKind           sanySemKind = 12
	sanySubstInKind           sanySemKind = 13
	sanyAssumeProveKind       sanySemKind = 14
	sanyNumeralKind           sanySemKind = 16
	sanyDecimalKind           sanySemKind = 17
	sanyStringKind            sanySemKind = 18
	sanyAtNodeKind            sanySemKind = 19
	sanyAssumeKind            sanySemKind = 20
	sanyInstanceKind          sanySemKind = 21
	sanyNewSymbKind           sanySemKind = 22
	sanyThmOrAssumpDefKind    sanySemKind = 23
	sanyNewConstantKind       sanySemKind = 24
	sanyNewVariableKind       sanySemKind = 25
	sanyNewStateKind          sanySemKind = 26
	sanyNewActionKind         sanySemKind = 27
	sanyNewTemporalKind       sanySemKind = 28
	sanyLabelKind             sanySemKind = 29
	sanyAPSubstInKind         sanySemKind = 30
	sanyUseKind               sanySemKind = 31
	sanyHideKind              sanySemKind = 32
	sanyLeafProofKind         sanySemKind = 33
	sanyNonLeafProofKind      sanySemKind = 34
	sanyQEDStepKind           sanySemKind = 35
	sanyDefStepKind           sanySemKind = 36
	sanyNumberedProofStepKind sanySemKind = 37
)

type sanySemSymbol interface {
	semName() string
	semKind() sanySemKind
	semArity() int
	semLocal() bool
	semOriginalModuleName() string
	semPosition() Position
	semBase() *sanySemSymbolBase
}

type sanySemSymbolBase struct {
	name               string
	kind               sanySemKind
	arity              int
	local              bool
	originalModuleName string
	pos                Position
	source             sanySemSymbol
	parameterized      bool
}

func (s *sanySemSymbolBase) semBase() *sanySemSymbolBase   { return s }
func (s *sanySemSymbolBase) semName() string               { return s.name }
func (s *sanySemSymbolBase) semKind() sanySemKind          { return s.kind }
func (s *sanySemSymbolBase) semArity() int                 { return s.arity }
func (s *sanySemSymbolBase) semLocal() bool                { return s.local }
func (s *sanySemSymbolBase) semOriginalModuleName() string { return s.originalModuleName }
func (s *sanySemSymbolBase) semPosition() Position         { return s.pos }

type sanySemModuleNode struct {
	sanySemSymbolBase
	context *sanyContext
}

func newSanySemModuleNode(name string, ctx *sanyContext, pos Position) *sanySemModuleNode {
	return &sanySemModuleNode{
		sanySemSymbolBase: sanySemSymbolBase{
			name:               name,
			kind:               sanyModuleKind,
			arity:              0,
			originalModuleName: name,
			pos:                pos,
		},
		context: ctx,
	}
}

func newSanySemSymbol(name string, kind sanySemKind, arity int, module string, pos Position) *sanySemSymbolBase {
	return &sanySemSymbolBase{name: name, kind: kind, arity: arity, originalModuleName: module, pos: pos}
}

type sanyContextKey struct {
	module bool
	name   string
}

type sanyContextEntry struct {
	key sanyContextKey
	sym sanySemSymbol
}

type sanyContext struct {
	table map[sanyContextKey]*sanyContextEntry
	order []*sanyContextEntry
}

func newSanyContext() *sanyContext {
	return &sanyContext{table: map[sanyContextKey]*sanyContextEntry{}}
}

func (c *sanyContext) duplicate() *sanyContext {
	dup := newSanyContext()
	if c == nil {
		return dup
	}
	for _, entry := range c.order {
		dup.add(entry.key, entry.sym)
	}
	return dup
}

func (c *sanyContext) getSymbol(name string) sanySemSymbol {
	if c == nil {
		return nil
	}
	if entry := c.table[sanyContextKey{name: name}]; entry != nil {
		return entry.sym
	}
	return nil
}

func (c *sanyContext) getModule(name string) *sanySemModuleNode {
	if c == nil {
		return nil
	}
	entry := c.table[sanyContextKey{module: true, name: name}]
	if entry == nil {
		return nil
	}
	mod, _ := entry.sym.(*sanySemModuleNode)
	return mod
}

func (c *sanyContext) addSymbol(sym sanySemSymbol) {
	if sym == nil {
		return
	}
	c.add(sanyContextKey{name: sym.semName()}, sym)
}

func (c *sanyContext) addModule(mod *sanySemModuleNode) {
	if mod == nil {
		return
	}
	c.add(sanyContextKey{module: true, name: mod.semName()}, mod)
}

func (c *sanyContext) add(key sanyContextKey, sym sanySemSymbol) {
	if c.table == nil {
		c.table = map[sanyContextKey]*sanyContextEntry{}
	}
	entry := &sanyContextEntry{key: key, sym: sym}
	c.table[key] = entry
	c.order = append(c.order, entry)
}

func (c *sanyContext) orderedSymbols() []sanySemSymbol {
	if c == nil {
		return nil
	}
	out := make([]sanySemSymbol, 0, len(c.order))
	for _, entry := range c.order {
		out = append(out, entry.sym)
	}
	return out
}

func sanyContextImportKind(symbol sanySemSymbol) DeclarationKind {
	switch symbol.semKind() {
	case sanyUserDefinedOpKind, sanyBuiltInKind, sanyModuleInstanceKind:
		return OperatorDecl
	case sanyThmOrAssumpDefKind:
		return semanticTheoremImportKind
	case sanyConstantDeclKind:
		return ConstantDecl
	case sanyVariableDeclKind:
		return VariableDecl
	case sanyFormalParamKind:
		return semanticFormalParamImportKind
	case sanyModuleKind:
		return InstanceDecl
	default:
		return DeclarationKind("BOUND_SYMBOL")
	}
}

func (c *sanyContext) mergeExtendContext(imported *sanyContext) (bool, Diagnostics) {
	var diags Diagnostics
	if c == nil || imported == nil {
		return true, diags
	}
	success := true
	for _, entry := range imported.order {
		sym := entry.sym
		if sym == nil || sym.semLocal() {
			continue
		}
		current := c.table[entry.key]
		if current == nil {
			c.add(entry.key, sym)
			continue
		}
		if current.sym == sym || sanySameOriginalModule(current.sym, sym) {
			continue
		}
		diagnostic := sanyExtendConflict(sym.semName(), sanyContextImportKind(sym), sym.semPosition(), sanyContextImportKind(current.sym), current.sym.semPosition())
		diags = appendSanyDiagnostics(diags, diagnostic)
		if diagnostic.Severity == SeverityError {
			success = false
		}
	}
	return success, diags
}

func sanyOriginalSource(symbol sanySemSymbol) sanySemSymbol {
	if source := symbol.semBase().source; source != nil {
		return source
	}
	return symbol
}

func sanySameOriginalModule(a, b sanySemSymbol) bool {
	if a == nil || b == nil || semanticImportClass(sanyContextImportKind(a)) != semanticImportClass(sanyContextImportKind(b)) {
		return false
	}
	kind := a.semKind()
	if kind != sanyUserDefinedOpKind && kind != sanyBuiltInKind && kind != sanyModuleInstanceKind && kind != sanyThmOrAssumpDefKind {
		return false
	}
	as := sanyOriginalSource(a)
	bs := sanyOriginalSource(b)
	return as == bs && !as.semBase().parameterized
}

type sanyExternalModuleTable struct {
	entries map[string]sanyExternalModuleTableEntry
	order   []*sanySemModuleNode
	root    *sanySemModuleNode
}

type sanyExternalModuleTableEntry struct {
	context *sanyContext
	module  *sanySemModuleNode
}

func newSanyExternalModuleTable() *sanyExternalModuleTable {
	return &sanyExternalModuleTable{entries: map[string]sanyExternalModuleTableEntry{}}
}

func (t *sanyExternalModuleTable) put(name string, ctx *sanyContext, mod *sanySemModuleNode) {
	if t == nil || name == "" || mod == nil {
		return
	}
	if t.entries == nil {
		t.entries = map[string]sanyExternalModuleTableEntry{}
	}
	if _, exists := t.entries[name]; exists {
		return
	}
	t.entries[name] = sanyExternalModuleTableEntry{context: ctx, module: mod}
	t.order = append(t.order, mod)
}

func (t *sanyExternalModuleTable) getContext(name string) *sanyContext {
	if t == nil {
		return nil
	}
	return t.entries[name].context
}

func (t *sanyExternalModuleTable) getModule(name string) *sanySemModuleNode {
	if t == nil {
		return nil
	}
	return t.entries[name].module
}

type sanySymbolTable struct {
	stack       []*sanyContext
	moduleTable *sanyExternalModuleTable
	module      *sanySemModuleNode
}

func newSanySymbolTable(base *sanyContext, modules *sanyExternalModuleTable) *sanySymbolTable {
	if base == nil {
		base = newSanyContext()
	}
	return &sanySymbolTable{stack: []*sanyContext{base}, moduleTable: modules}
}

func (st *sanySymbolTable) duplicateForInnerModule() *sanySymbolTable {
	if st == nil {
		return newSanySymbolTable(nil, nil)
	}
	next := &sanySymbolTable{
		stack:       append([]*sanyContext(nil), st.stack...),
		moduleTable: st.moduleTable,
		module:      st.module,
	}
	return next
}

func (st *sanySymbolTable) topContext() *sanyContext {
	if st == nil || len(st.stack) == 0 {
		return nil
	}
	return st.stack[0]
}

func (st *sanySymbolTable) externalContext() *sanyContext {
	if st == nil || len(st.stack) == 0 {
		return nil
	}
	return st.stack[len(st.stack)-1]
}

func (st *sanySymbolTable) pushContext(ctx *sanyContext) {
	if st == nil {
		return
	}
	if ctx == nil {
		ctx = newSanyContext()
	}
	st.stack = append([]*sanyContext{ctx}, st.stack...)
}

func (st *sanySymbolTable) popContext() {
	if st == nil || len(st.stack) == 0 {
		return
	}
	st.stack = st.stack[1:]
}

func (st *sanySymbolTable) resolveSymbol(name string) sanySemSymbol {
	if st == nil {
		return nil
	}
	for _, ctx := range st.stack {
		if sym := ctx.getSymbol(name); sym != nil {
			return sym
		}
	}
	return nil
}

func (st *sanySymbolTable) resolveModule(name string) *sanySemModuleNode {
	if st == nil {
		return nil
	}
	for _, ctx := range st.stack {
		if mod := ctx.getModule(name); mod != nil {
			return mod
		}
	}
	if st.moduleTable != nil {
		return st.moduleTable.getModule(name)
	}
	return nil
}

func (st *sanySymbolTable) addSymbol(sym sanySemSymbol) Diagnostics {
	var diags Diagnostics
	if st == nil || sym == nil {
		return diags
	}
	current := st.resolveSymbol(sym.semName())
	if current == nil {
		st.topContext().addSymbol(sym)
		return diags
	}
	if current == sym || sanySameOriginalModule(current, sym) {
		return diags
	}
	if current.semKind() != sym.semKind() || current.semArity() != sym.semArity() ||
		sym.semKind() == sanyFormalParamKind || sym.semKind() == sanyBoundSymbolKind {
		return Diagnostics{errorAt(sym.semPosition(), "E4802", "multiply-defined symbol %s", sym.semName())}
	}
	return Diagnostics{warningAt(sym.semPosition(), "W4802", "multiple declarations or definitions for symbol %s", sym.semName())}
}

func (st *sanySymbolTable) addModule(mod *sanySemModuleNode) Diagnostics {
	if st == nil || mod == nil {
		return nil
	}
	if current := st.resolveModule(mod.semName()); current != nil && current != mod {
		return Diagnostics{errorAt(mod.semPosition(), "E4803", "multiply-defined module %s", mod.semName())}
	}
	st.topContext().addModule(mod)
	return nil
}
