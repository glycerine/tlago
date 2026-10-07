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
	sanySemanticNode
	name               string
	arity              int
	local              bool
	originalModuleName string
	pos                Position
	source             sanySemSymbol
	parameterized      bool
}

func (s *sanySemSymbolBase) semBase() *sanySemSymbolBase   { return s }
func (s *sanySemSymbolBase) semName() string               { return s.name }
func (s *sanySemSymbolBase) semKind() sanySemKind          { return s.getKind() }
func (s *sanySemSymbolBase) semArity() int                 { return s.arity }
func (s *sanySemSymbolBase) semLocal() bool                { return s.local }
func (s *sanySemSymbolBase) semOriginalModuleName() string { return s.originalModuleName }
func (s *sanySemSymbolBase) semPosition() Position         { return s.pos }

type sanySemModuleNode struct {
	sanySemSymbolBase
	context                  *sanyContext
	isInstantiated           bool
	instanceVec              []*sanySemInstanceNode
	instances                []*sanySemInstanceNode
	nestingLevel             int
	definitions              []sanySemanticGraphNode
	records                  []*sanySemOpApplNode
	assumptionVec            []*sanySemAssumeNode
	theoremVec               []*sanySemTheoremNode
	topLevelVec              []sanySemanticGraphNode
	assumptions              []*sanySemAssumeNode
	theorems                 []*sanySemTheoremNode
	topLevel                 []sanySemanticGraphNode
	recursiveDecls           []*sanySemOpDefNode
	recursiveOpDefNodes      []*sanySemOpDefNode
	opDefsInRecursiveSection []*sanySemOpDefNode
}

func newSanySemModuleNode(name string, ctx *sanyContext, pos Position, syntax ...*SanySyntaxNode) *sanySemModuleNode {
	node := &sanySemModuleNode{
		sanySemSymbolBase: sanySemSymbolBase{
			sanySemanticNode:   newSanySemanticNode(sanyModuleKind),
			name:               name,
			arity:              -2,
			originalModuleName: name,
			pos:                pos,
		},
		context:     ctx,
		definitions: make([]sanySemanticGraphNode, 0),
	}
	if len(syntax) != 0 && syntax[0] != nil {
		node.TreeNode = syntax[0]
		bridge := tlcBridge{convertingModule: name}
		node.Location = bridge.sourceLocationForPosition(sanyNodePosition(syntax[0]))
	}
	return node
}

func newSanySemSymbol(name string, kind sanySemKind, arity int, module string, pos Position) *sanySemSymbolBase {
	return &sanySemSymbolBase{sanySemanticNode: newSanySemanticNode(kind), name: name, arity: arity, originalModuleName: module, pos: pos}
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
	table            map[sanyContextKey]*sanyContextEntry
	order            []*sanyContextEntry
	contentBuckets   []*sanyContextBucketEntry
	contentThreshold int
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
	c.putContentEntry(entry)
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
	// Context compares concrete symbol classes, including kind-zero OpDefNodes.
	switch node := symbol.(type) {
	case *sanySemOpDefNode:
		return OperatorDecl
	case *sanySemOpDeclNode:
		if node.semKind() == sanyVariableDeclKind {
			return VariableDecl
		}
		return ConstantDecl
	case *sanyFormalParamNode:
		return semanticFormalParamImportKind
	case *sanySemModuleNode:
		return InstanceDecl
	}
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
	if definition, ok := a.(*sanySemThmOrAssumpDefNode); ok {
		other, sameClass := b.(*sanySemThmOrAssumpDefNode)
		if !sameClass || definition.getSource() != other.getSource() {
			return false
		}
		original := definition.getSource().module
		if original != nil && original.context != nil {
			for _, entry := range original.context.order {
				if entry.sym.semKind() == sanyConstantDeclKind || entry.sym.semKind() == sanyVariableDeclKind {
					return false
				}
			}
		}
		return true
	}

	// Source OpDefNodes must have the same concrete class and original source.
	// Parameter freedom comes from that source's module declarations, not its name.
	if definition, ok := a.(*sanySemOpDefNode); ok {
		other, sameClass := b.(*sanySemOpDefNode)
		if !sameClass {
			return false
		}
		source, otherSource := sanyOriginalSource(definition), sanyOriginalSource(other)
		if source != otherSource {
			return false
		}
		original, ok := source.(*sanySemOpDefNode)
		if !ok {
			return false
		}
		if original.module == nil {
			return true
		}
		for _, entry := range original.module.context.order {
			kind := entry.sym.semKind()
			if kind == sanyConstantDeclKind || kind == sanyVariableDeclKind {
				return false
			}
		}
		return true
	}
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

// addSymbol retains the diagnostic-only API used by native adapters. The
// registration result follows Java independently of warning severity.
func (st *sanySymbolTable) addSymbol(sym sanySemSymbol) Diagnostics {
	_, diagnostics := st.registerSymbol(sym)
	return diagnostics
}

func (st *sanySymbolTable) registerSymbol(sym sanySemSymbol) (bool, Diagnostics) {
	if st == nil || sym == nil {
		return false, nil
	}
	current := st.resolveSymbol(sym.semName())
	if current == sym {
		return true, nil
	}
	if current == nil {
		st.topContext().addSymbol(sym)
		return true, nil
	}
	name, position := sym.semName(), sym.semPosition()
	// Java tests the old syntax node's source, rather than its semantic kind.
	if current.semBase().Location.Source == "--TLA+ BUILTINS--" {
		message := "Symbol %s is a built-in operator, and cannot be redefined."
		diagnostic := sanyRegistrationDiagnostic(position, "E4202", message, name)
		return false, Diagnostics{diagnostic}
	}
	if sym.semKind() == sanyFormalParamKind || sym.semKind() == sanyBoundSymbolKind || current.semKind() != sym.semKind() || current.semArity() != sym.semArity() {
		message := "Multiply-defined symbol '%s': this definition or declaration conflicts \nwith the one at %s."
		diagnostic := sanyRegistrationDiagnostic(position, "E4201", message, name, sanyDiagnosticLocation{Position: current.semPosition()})
		return false, Diagnostics{diagnostic}
	}
	if sanySameOriginalModule(sym, current) {
		return true, nil
	}
	message := "Multiple declarations or definitions for symbol %s.  \nThis duplicates the one at %s."
	diagnostic := sanyRegistrationDiagnostic(position, "W4801", message, name, sanyDiagnosticLocation{Position: current.semPosition()})
	return true, Diagnostics{diagnostic}
}

func sanyRegistrationDiagnostic(position Position, code, message string, parameters ...any) Diagnostic {
	diagnostic := errorAt(position, code, message, parameters...)
	if code == "W4801" {
		diagnostic = warningAt(position, code, message, parameters...)
	}
	diagnostic.SANYMessage = diagnostic.Message
	diagnostic.SANYRange = SanyRange{Begin: position, End: position.SourceEnd()}
	diagnostic.SANYParameters = parameters
	return diagnostic
}

func (st *sanySymbolTable) addModule(mod *sanySemModuleNode) Diagnostics {
	_, diagnostics := st.registerModule(mod)
	return diagnostics
}

func (st *sanySymbolTable) registerModule(mod *sanySemModuleNode) (bool, Diagnostics) {
	if st == nil || mod == nil {
		return false, nil
	}
	current := st.resolveModule(mod.semName())
	if current == mod {
		return true, nil
	}
	if current == nil {
		st.topContext().addModule(mod)
		return true, nil
	}
	message := "Multiply-defined module '%s': this definition or declaration conflicts \nwith the one at %s."
	diagnostic := sanyRegistrationDiagnostic(mod.semPosition(), "E4223", message, mod.semName(), sanyDiagnosticLocation{Position: current.semPosition()})
	return false, Diagnostics{diagnostic}
}

// Context.isBuiltIn recognizes the actual current initial-context identities,
// rather than inferring builtin status from a node kind or spelling.
func sanyContextIsBuiltIn(node any) bool {
	for _, symbol := range sanyGlobalInitialContext(false).orderedSymbols() {
		if symbol == node {
			return true
		}
	}
	return false
}

func sanySymbolIsParam(symbol sanySemSymbol) bool {
	switch symbol.(type) {
	case *sanySemOpDeclNode, *sanyFormalParamNode:
		return true
	default:
		return false
	}
}

// ModuleNode records completed record constructors and record-set applications
// in generation order, retaining the actual application identities.
func (n *sanySemModuleNode) addRecord(record *sanySemOpApplNode) *sanySemOpApplNode {
	n.records = append(n.records, record)
	return record
}

func (n *sanySemModuleNode) getRecords() []*sanySemOpApplNode {
	result := make([]*sanySemOpApplNode, len(n.records))
	copy(result, n.records)
	return result
}
