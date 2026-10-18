package tlc

import (
	"fmt"
	"strconv"
	"time"
	"unicode/utf16"
)

// Context.walkGraph enumerates keys and resolves the current binding before
// visiting each child; Pair history and declaration order are separate views.
func (c *SemanticContext) WalkGraphNodes(visit func(SemanticNode)) {
	entries := c.GetContextSymbolEnumeration()
	for entries.HasMoreElements() {
		entry := entries.nextEntry()
		if entry.key.Module {
			fmt.Printf("Bug in debugging caused by inner module %s\nSANY will throw a null pointer exception.\n", entry.key.Name)
		} else {
			visit(c.GetSymbol(entry.key))
		}
	}
}

// SemanticContext is tla2sany.semantic.Context. Context in context.go is the
// separate tlc2.util.Context used for runtime evaluation bindings.
type SemanticContext struct {
	ModuleTable *ExternalModuleTable
	lastPair    *semanticContextPair
	buckets     []*semanticContextEntry
	count       int
	threshold   int
}

type SemanticContextKey struct {
	Name   *UniqueString
	Module bool // SymbolTable.ModuleName has a separate equality namespace.
}

type semanticContextPair struct {
	link *semanticContextPair
	node SemanticNode
	key  SemanticContextKey
}

type semanticContextEntry struct {
	key  SemanticContextKey
	pair *semanticContextPair
	next *semanticContextEntry
}

func NewSemanticContext(table *ExternalModuleTable) *SemanticContext {
	return &SemanticContext{ModuleTable: table, buckets: make([]*semanticContextEntry, 11), threshold: 8}
}

func semanticContextHash(key SemanticContextKey, size int) int {
	if key.Name == nil {
		panic(NewNullPointerException())
	}
	var hash uint32
	for _, unit := range utf16.Encode([]rune(key.Name.String())) {
		hash = 31*hash + uint32(unit)
	}
	return int(hash&0x7fffffff) % size
}

func (c *SemanticContext) put(key SemanticContextKey, pair *semanticContextPair) {
	i := semanticContextHash(key, len(c.buckets))
	for entry := c.buckets[i]; entry != nil; entry = entry.next {
		if entry.key == key {
			entry.pair = pair
			return
		}
	}
	if c.count >= c.threshold {
		old := c.buckets
		c.buckets = make([]*semanticContextEntry, 2*len(old)+1)
		c.threshold = 3 * len(c.buckets) / 4
		for j := len(old) - 1; j >= 0; j-- {
			for entry := old[j]; entry != nil; {
				next := entry.next
				bucket := semanticContextHash(entry.key, len(c.buckets))
				entry.next = c.buckets[bucket]
				c.buckets[bucket] = entry
				entry = next
			}
		}
		i = semanticContextHash(key, len(c.buckets))
	}
	c.buckets[i] = &semanticContextEntry{key: key, pair: pair, next: c.buckets[i]}
	c.count++
}

func (c *SemanticContext) AddSymbolToContext(key SemanticContextKey, node SemanticNode) {
	if c == nil {
		panic(NewNullPointerException())
	}
	if semanticExploreNull(node) {
		node = nil
	}
	pair := &semanticContextPair{link: c.lastPair, node: node, key: key}
	c.lastPair = pair
	c.put(key, pair)
}

func (c *SemanticContext) GetSymbol(key SemanticContextKey) SemanticNode {
	if c == nil {
		panic(NewNullPointerException())
	}
	for entry := c.buckets[semanticContextHash(key, len(c.buckets))]; entry != nil; entry = entry.next {
		if entry.key == key {
			return entry.pair.node
		}
	}
	return nil
}

func (c *SemanticContext) OccurSymbol(key SemanticContextKey) bool {
	if c == nil {
		panic(NewNullPointerException())
	}
	for entry := c.buckets[semanticContextHash(key, len(c.buckets))]; entry != nil; entry = entry.next {
		if entry.key == key {
			return true
		}
	}
	return false
}

func (c *SemanticContext) Duplicate(table *ExternalModuleTable) *SemanticContext {
	if c == nil {
		panic(NewNullPointerException())
	}
	dup := NewSemanticContext(table)
	var tail *semanticContextPair
	for pair := c.lastPair; pair != nil; pair = pair.link {
		if pair.node == nil {
			panic(NewNullPointerException())
		}
		// Java uses info.getName(), even when the source entry used the
		// distinct ModuleName namespace or another context lookup name.
		key := SemanticContextKey{Name: pair.key.Name}
		switch node := pair.node.(type) {
		case *SymbolNode:
			key.Name = node.GetName()
		case *OpDefNode:
			key.Name = node.Name
		case *ThmOrAssumpDefNode:
			key.Name = node.Name
		case *ModuleNode:
			key.Name = node.Name
		}
		copy := &semanticContextPair{node: pair.node, key: key}
		if tail == nil {
			dup.lastPair = copy
		} else {
			tail.link = copy
		}
		tail = copy
		dup.put(copy.key, copy)
	}
	return dup
}

// SemanticContextSymbolEnumeration retains the Hashtable bucket array and
// entry cursor, as Java's non-fail-fast Enumeration does. Rehashing relinks the
// same entries; replacement updates values without moving their bucket.
type SemanticContextSymbolEnumeration struct {
	buckets []*semanticContextEntry
	index   int
	entry   *semanticContextEntry
}

func (c *SemanticContext) GetContextSymbolEnumeration() *SemanticContextSymbolEnumeration {
	if c == nil {
		panic(NewNullPointerException())
	}
	return &SemanticContextSymbolEnumeration{buckets: c.buckets, index: len(c.buckets)}
}
func (e *SemanticContextSymbolEnumeration) HasMoreElements() bool {
	if e == nil {
		panic(NewNullPointerException())
	}
	for e.entry == nil && e.index > 0 {
		e.index--
		e.entry = e.buckets[e.index]
	}
	return e.entry != nil
}
func (e *SemanticContextSymbolEnumeration) nextEntry() *semanticContextEntry {
	if !e.HasMoreElements() {
		message := "Hashtable Enumerator"
		failure := NewNoSuchElementException()
		failure.Message = &message
		panic(failure)
	}
	entry := e.entry
	e.entry = entry.next
	return entry
}
func (e *SemanticContextSymbolEnumeration) NextElement() SemanticNode {
	return e.nextEntry().pair.node
}

// Content materializes Hashtable.elements() order for existing Go callers.
// Use GetContextSymbolEnumeration when mutations can occur during enumeration.
func (c *SemanticContext) Content() []SemanticNode {
	result := make([]SemanticNode, 0)
	entries := c.GetContextSymbolEnumeration()
	for entries.HasMoreElements() {
		result = append(result, entries.NextElement())
	}
	return result
}

type SemanticContextMergeConflict struct {
	Name      *UniqueString
	Incoming  SemanticNode
	Existing  SemanticNode
	SameClass bool
}

func semanticContextNodeLocal(node SemanticNode) bool {
	switch node := node.(type) {
	case *SymbolNode:
		return node.IsLocal()
	case *OpDefNode:
		return node.IsLocal()
	case *ThmOrAssumpDefNode:
		return node.IsLocal()
	case *ModuleNode:
		return node.IsLocal()
	default:
		panic(NewClassCastException("context entry is not a SymbolNode"))
	}
}

func semanticContextSameOriginalModule(left, right SemanticNode) bool {
	if leftDef, leftOK := semanticContextOpDef(left); leftOK {
		rightDef, rightOK := semanticContextOpDef(right)
		if !rightOK || leftDef.GetSource() != rightDef.GetSource() {
			return false
		}
		return semanticDefinitionModuleIsParameterFree(leftDef.GetSource().OriginallyDefinedInModule)
	}
	if leftDef, leftOK := semanticContextTheoremDef(left); leftOK {
		rightDef, rightOK := semanticContextTheoremDef(right)
		if !rightOK || leftDef.GetSource() != rightDef.GetSource() {
			return false
		}
		return semanticDefinitionModuleIsParameterFree(leftDef.GetSource().OriginallyDefinedInModule)
	}
	return false
}

func semanticContextOpDef(node SemanticNode) (*OpDefNode, bool) {
	switch node := node.(type) {
	case *OpDefNode:
		return node, node != nil
	case *SymbolNode:
		if node == nil {
			return nil, false
		}
		return node.Definition, node.Definition != nil
	default:
		return nil, false
	}
}

func semanticContextTheoremDef(node SemanticNode) (*ThmOrAssumpDefNode, bool) {
	switch node := node.(type) {
	case *ThmOrAssumpDefNode:
		return node, node != nil
	case *SymbolNode:
		if node == nil {
			return nil, false
		}
		definition, ok := node.Data.(*ThmOrAssumpDefNode)
		return definition, ok && definition != nil
	default:
		return nil, false
	}
}

// Java compares concrete SymbolNode classes. Go uses SymbolNode for both
// FormalParamNode and OpDeclNode, and also exposes definition-backed symbol
// adapters, so reflect.TypeOf alone collapses distinct source classes.
func semanticJavaClass(node any) string {
	switch node := node.(type) {
	case *OpDefNode:
		return "OpDefNode"
	case *ThmOrAssumpDefNode:
		return "ThmOrAssumpDefNode"
	case *ModuleNode:
		return "ModuleNode"
	case *SymbolNode:
		if node != nil {
			if node.Definition != nil {
				return "OpDefNode"
			}
			if definition, ok := node.Data.(*ThmOrAssumpDefNode); ok && definition != nil {
				return "ThmOrAssumpDefNode"
			}
			switch node.Kind {
			case SymbolFormalParam:
				return "FormalParamNode"
			case SymbolVariableDecl, SymbolConstantDecl:
				return "OpDeclNode"
			case SymbolUserDefinedOp, SymbolBuiltIn:
				return "OpDefNode"
			}
			if node.Declaration {
				return "OpDeclNode"
			}
		}
		return "SymbolNode"
	default:
		return fmt.Sprintf("%T", node)
	}
}

// MergeExtendContext merges oldest-to-newest Pair history, skipping local
// symbols and retaining the existing binding on conflicts.
func (c *SemanticContext) MergeExtendContext(imported *SemanticContext) (bool, []SemanticContextMergeConflict) {
	if c == nil || imported == nil {
		panic(NewNullPointerException())
	}
	pairs := make([]*semanticContextPair, 0)
	for pair := imported.lastPair; pair != nil; pair = pair.link {
		pairs = append(pairs, pair)
	}
	success := true
	var conflicts []SemanticContextMergeConflict
	for i := len(pairs) - 1; i >= 0; i-- {
		node := pairs[i].node
		if node == nil {
			panic(NewNullPointerException())
		}
		if semanticContextNodeLocal(node) {
			continue
		}
		name := semanticContextNodeName(node)
		key := SemanticContextKey{Name: name}
		if _, module := node.(*ModuleNode); module {
			key.Module = true
		}
		if !c.OccurSymbol(key) {
			c.AddSymbolToContext(key, node)
			continue
		}
		existing := c.GetSymbol(key)
		if existing == nil {
			panic(NewNullPointerException())
		}
		if existing == node {
			continue
		}
		sameClass := semanticJavaClass(existing) == semanticJavaClass(node)
		if sameClass && semanticContextSameOriginalModule(existing, node) {
			continue
		}
		conflicts = append(conflicts, SemanticContextMergeConflict{Name: name, Incoming: node, Existing: existing, SameClass: sameClass})
		if !sameClass {
			success = false
		}
	}
	return success, conflicts
}

func (c *SemanticContext) GetOpDefs() []*OpDefNode {
	if c == nil {
		panic(NewNullPointerException())
	}
	result := make([]*OpDefNode, 0)
	for pair := c.lastPair; pair != nil; pair = pair.link {
		if def, ok := pair.node.(*OpDefNode); ok && def.Kind() != SemanticBuiltInKind && def.Kind() != SemanticModuleInstanceKind {
			result = append(result, def)
		}
	}
	return result
}

func (c *SemanticContext) GetThmOrAssDefs() []*ThmOrAssumpDefNode {
	if c == nil {
		panic(NewNullPointerException())
	}
	result := make([]*ThmOrAssumpDefNode, 0)
	for pair := c.lastPair; pair != nil; pair = pair.link {
		if def, ok := pair.node.(*ThmOrAssumpDefNode); ok {
			result = append(result, def)
		}
	}
	return result
}

func (c *SemanticContext) declarations(kind SymbolKind) []*SymbolNode {
	result := make([]*SymbolNode, 0)
	for _, node := range c.Content() {
		if declaration, ok := node.(*SymbolNode); ok && declaration.Kind == kind {
			result = append(result, declaration)
		}
	}
	return result
}

func (c *SemanticContext) GetConstantDecls() []*SymbolNode { return c.declarations(SymbolConstantDecl) }
func (c *SemanticContext) GetVariableDecls() []*SymbolNode { return c.declarations(SymbolVariableDecl) }

func (c *SemanticContext) GetModDefs() []*ModuleNode {
	result := make([]*ModuleNode, 0)
	for _, node := range c.Content() {
		if module, ok := node.(*ModuleNode); ok {
			result = append(result, module)
		}
	}
	return result
}

type ModuleNode struct {
	*SemanticNodeBase
	Name         *UniqueString
	Context      *SemanticContext
	Extendees    []*ModuleNode
	Instantiated bool
	Standard     bool
	TopLevel     []SemanticNode
	Definitions  []SemanticNode
	InstanceVec  []*InstanceNode
	Records      []*OpApplNode
	// SANYSourceModule retains the frontend module behind this public semantic
	// view. The root package owns its concrete type; tlc keeps it opaque to
	// avoid a package cycle while incremental Generator calls preserve source
	// declarations, instances and graph identity.
	SANYSourceModule any

	constantDecls      []*SymbolNode
	variableDecls      []*SymbolNode
	opDefs             []*OpDefNode
	thmOrAssDefs       []*ThmOrAssumpDefNode
	innerModules       []*ModuleNode
	assumptionVec      []*AssumeNode
	assumptions        []*AssumeNode
	theoremVec         []*TheoremNode
	theorems           []*TheoremNode
	instances          []*InstanceNode
	topLevel           []SemanticNode
	extendedModuleSets map[bool]*InsMap[*ModuleNode, struct{}]
	children           []SemanticNode
}

func NewModuleNode(name string, context *SemanticContext) *ModuleNode {
	return NewModuleNodeWithBase(name, context, nil)
}

func NewModuleNodeWithBase(name string, context *SemanticContext, base *SemanticNodeBase) *ModuleNode {
	if base == nil {
		base = newSemanticNodeBasePointer(SemanticModuleKind, name)
	}
	return &ModuleNode{SemanticNodeBase: base, Name: UniqueStringOf(name), Context: context, Definitions: make([]SemanticNode, 0)}
}

func (m *ModuleNode) GetName() *UniqueString {
	m.require()
	return m.Name
}
func (m *ModuleNode) GetSignature() string {
	m.require()
	return m.GetName().String()
}
func (m *ModuleNode) GetArity() int {
	m.require()
	return -2
}
func (m *ModuleNode) GetLevel() int {
	m.require()
	panic(NewWrongInvocationException("Internal Error: Should never call ModuleNode.getLevel()"))
}
func (m *ModuleNode) IsLocal() bool {
	m.require()
	return false
}
func (m *ModuleNode) GetContext() *SemanticContext {
	m.require()
	return m.Context
}

func (m *ModuleNode) GetDefinitions() []SemanticNode {
	m.require()
	return m.Definitions
}

func (m *ModuleNode) AppendDef(node SemanticNode) {
	m.require()
	m.Definitions = append(m.Definitions, node)
}

func (m *ModuleNode) GetInstances() []*InstanceNode {
	m.require()
	if m.instances == nil {
		m.instances = make([]*InstanceNode, len(m.InstanceVec))
		copy(m.instances, m.InstanceVec)
	}
	return m.instances
}

func (m *ModuleNode) AppendInstance(node *InstanceNode) {
	m.require()
	m.InstanceVec = append(m.InstanceVec, node)
	m.TopLevel = append(m.TopLevel, node)
}

func (m *ModuleNode) GetRecords() []*OpApplNode {
	m.require()
	result := make([]*OpApplNode, len(m.Records))
	copy(result, m.Records)
	return result
}

func (m *ModuleNode) GetTopLevel() []SemanticNode {
	m.require()
	if m.topLevel == nil {
		m.topLevel = make([]SemanticNode, len(m.TopLevel))
		copy(m.topLevel, m.TopLevel)
	}
	return m.topLevel
}

// GenerateUnusedName returns a formatted name that is absent from the
// context visible at this module root. Java uses the current millisecond as a
// string argument and retries while the resulting symbol already exists.
func (m *ModuleNode) GenerateUnusedName(pattern string) string {
	m.require()
	definedNames := m.GetContext()
	for {
		suffix := strconv.FormatInt(time.Now().UnixMilli(), 10)
		name, err := JavaFormatStrings(pattern, suffix)
		if err != nil {
			panic(err)
		}
		if !definedNames.OccurSymbol(SemanticContextKey{Name: UniqueStringOf(name)}) {
			return name
		}
	}
}

func (m *ModuleNode) GetThms() []*TheoremNode {
	m.require()
	return nil
}

func (m *ModuleNode) GetAssumes() []*AssumeProveNode {
	m.require()
	return nil
}
func (m *ModuleNode) IsInstantiated() bool {
	m.require()
	if provider, ok := m.CanonicalLevelData.(interface{ RuntimeModuleInstantiated() (bool, bool) }); ok {
		if instantiated, present := provider.RuntimeModuleInstantiated(); present {
			return instantiated
		}
	}
	return m.Instantiated
}
func (m *ModuleNode) SetInstantiated(value bool) {
	m.require()
	if provider, ok := m.CanonicalLevelData.(interface{ RuntimeSetModuleInstantiated(bool) bool }); ok {
		provider.RuntimeSetModuleInstantiated(value)
	}
	m.Instantiated = value
}
func (m *ModuleNode) IsStandard() bool {
	m.require()
	return m.Standard
}
func (m *ModuleNode) SetStandard(value bool) {
	m.require()
	m.Standard = value
}
func (m *ModuleNode) IsParameterFree() bool {
	m.require()
	return len(m.GetConstantDecls()) == 0 && len(m.GetVariableDecls()) == 0
}
func (m *ModuleNode) ProcessConstantDefns() bool {
	m.require()
	return !m.IsInstantiated() || m.IsParameterFree()
}

// Constant processing uses the assumption vector, independently of the
// interleaved top-level graph vector. Native AST statements can still lack a
// generated SANY graph; they remain required runtime assumptions.
func (m *ModuleNode) AddAssumptionStatement(node *AssumeNode) {
	m.require()
	m.assumptionVec = append(m.assumptionVec, node)
	m.TopLevel = append(m.TopLevel, node)
}

func (m *ModuleNode) GetAssumptions() []*AssumeNode {
	m.require()
	if m.assumptions == nil {
		if m.assumptionVec != nil {
			m.assumptions = make([]*AssumeNode, len(m.assumptionVec))
			copy(m.assumptions, m.assumptionVec)
		} else {
			m.assumptions = make([]*AssumeNode, 0)
			for _, statement := range m.TopLevel {
				if assume, ok := statement.(*AssumeNode); ok {
					m.assumptions = append(m.assumptions, assume)
				}
			}
		}
	}
	return m.assumptions
}

// Retain AST fallback assumptions not yet represented by the generated SANY
// vector. Canonical statements keep their source order and identity.
func (m *ModuleNode) graphStatements() []SemanticNode {
	statements := append([]SemanticNode(nil), m.TopLevel...)
	present := map[semanticNodeKey]bool{}
	for _, statement := range statements {
		present[newSemanticNodeKey(statement)] = true
	}
	for _, assume := range m.GetAssumptions() {
		if !present[newSemanticNodeKey(assume)] {
			statements = append(statements, assume)
		}
	}
	return statements
}

// The bridge appends views of the actual retained theorem statements, including
// EXTENDS copies. Named definition enumeration has a different membership.
func (m *ModuleNode) AddTheoremStatement(node *TheoremNode) {
	m.require()
	m.theoremVec = append(m.theoremVec, node)
}

func (m *ModuleNode) GetTheorems() []*TheoremNode {
	m.require()
	if m.theorems == nil {
		m.theorems = make([]*TheoremNode, len(m.theoremVec))
		copy(m.theorems, m.theoremVec)
	}
	return m.theorems
}

func (m *ModuleNode) GetChildren() []SemanticNode {
	m.require()
	if m.children == nil {
		definitions := m.GetOpDefs()
		if m.topLevel == nil {
			panic(NewNullPointerException())
		}
		statements := m.topLevel
		m.children = make([]SemanticNode, 0, len(definitions)+len(statements))
		for _, definition := range definitions {
			m.children = append(m.children, definition)
		}
		m.children = append(m.children, statements...)
	}
	return m.children
}

func (m *ModuleNode) CreateExtendeeArray(extendees []*ModuleNode) {
	m.require()
	if extendees == nil {
		panic(NewNullPointerException())
	}
	m.Extendees = make([]*ModuleNode, len(extendees))
	copy(m.Extendees, extendees)
}

func (m *ModuleNode) GetConstantDecls() []*SymbolNode {
	m.require()
	if m.constantDecls == nil {
		m.constantDecls = reverseSemanticSlice(m.Context.GetConstantDecls())
	}
	return m.constantDecls
}

func (m *ModuleNode) GetVariableDecls() []*SymbolNode {
	m.require()
	if m.variableDecls == nil {
		m.variableDecls = reverseSemanticSlice(m.Context.GetVariableDecls())
	}
	return m.variableDecls
}

func (m *ModuleNode) GetOpDefs() []*OpDefNode {
	m.require()
	if m.opDefs == nil {
		m.opDefs = reverseSemanticSlice(m.Context.GetOpDefs())
	}
	return m.opDefs
}

func (m *ModuleNode) GetOpDef(name *UniqueString) *OpDefNode {
	m.require()
	for _, def := range m.GetOpDefs() {
		if def.Name == name {
			return def
		}
	}
	return nil
}

// GetOpDefByName adapts ModuleNode.getOpDef(String); the source interns the
// string and delegates to the UniqueString overload.
func (m *ModuleNode) GetOpDefByName(name string) *OpDefNode {
	m.require()
	return m.GetOpDef(UniqueStringOf(name))
}

func (m *ModuleNode) GetThmOrAssDefs() []*ThmOrAssumpDefNode {
	m.require()
	if m.thmOrAssDefs == nil {
		m.thmOrAssDefs = reverseSemanticSlice(m.Context.GetThmOrAssDefs())
	}
	return m.thmOrAssDefs
}

func (m *ModuleNode) GetInnerModules() []*ModuleNode {
	m.require()
	if m.innerModules == nil {
		m.innerModules = m.Context.GetModDefs()
	}
	return m.innerModules
}

func (m *ModuleNode) GetExtendedModuleSet(recursive ...bool) *InsMap[*ModuleNode, struct{}] {
	m.require()
	if len(recursive) > 1 {
		panic(NewIllegalArgumentException("GetExtendedModuleSet accepts at most one recursive flag"))
	}
	recurse := len(recursive) == 0 || recursive[0]
	if cached := m.extendedModuleSets[recurse]; cached != nil {
		return cached
	}
	if m.Extendees == nil {
		panic(NewNullPointerException())
	}
	result := newInsMapAllowNilKeys[*ModuleNode, struct{}]()
	for _, extendee := range m.Extendees {
		result.Set(extendee, struct{}{})
		if recurse {
			for nested := range extendee.GetExtendedModuleSet(true).All() {
				result.Set(nested, struct{}{})
			}
		}
	}
	if m.extendedModuleSets == nil {
		m.extendedModuleSets = map[bool]*InsMap[*ModuleNode, struct{}]{}
	}
	m.extendedModuleSets[recurse] = result
	return result
}

func (m *ModuleNode) ExtendsModule(module *ModuleNode) bool {
	m.require()
	_, exists := m.GetExtendedModuleSet().Get2(module)
	return exists
}

func (m *ModuleNode) require() {
	if m == nil {
		panic(NewNullPointerException())
	}
}

func reverseSemanticSlice[T any](values []T) []T {
	if values == nil {
		return make([]T, 0)
	}
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	return values
}

type ExternalModuleTableEntry struct {
	ModuleNode *ModuleNode
	Context    *SemanticContext
}

func (e *ExternalModuleTableEntry) require() *ExternalModuleTableEntry {
	if e == nil {
		panic(NewNullPointerException())
	}
	return e
}

func (e *ExternalModuleTableEntry) GetModuleNode() *ModuleNode { return e.require().ModuleNode }
func (e *ExternalModuleTableEntry) LevelDataToString() string {
	e.require()
	return "Dummy level string"
}

type ExternalModuleTable struct {
	ModuleHashTable  *InsMap[*UniqueString, *ExternalModuleTableEntry]
	ModuleNodeVector []*ModuleNode
	RootModule       *ModuleNode
}

type externalModuleTableHashEntry struct {
	hash  int32
	value *ExternalModuleTableEntry
	next  *externalModuleTableHashEntry
}

// externalModuleTableEntriesInJavaOrder reconstructs Hashtable.elements()
// from the retained insertion-ordered map. Java's ExternalModuleTable uses
// default capacity 11, load factor .75, head insertion and 2*n+1 rehashes.
func externalModuleTableEntriesInJavaOrder(table *ExternalModuleTable) []*ExternalModuleTableEntry {
	table.require()
	if table.ModuleHashTable == nil {
		panic(NewNullPointerException())
	}
	buckets := make([]*externalModuleTableHashEntry, 11)
	threshold := 8
	count := 0
	for name, value := range table.ModuleHashTable.All() {
		if count >= threshold {
			old := buckets
			buckets = make([]*externalModuleTableHashEntry, 2*len(old)+1)
			threshold = int(float64(len(buckets)) * 0.75)
			for index := len(old) - 1; index >= 0; index-- {
				for entry := old[index]; entry != nil; {
					next := entry.next
					bucket := int(uint32(entry.hash)&0x7fffffff) % len(buckets)
					entry.next = buckets[bucket]
					buckets[bucket] = entry
					entry = next
				}
			}
		}
		if name == nil {
			panic(NewNullPointerException())
		}
		hash := name.HashCode()
		bucket := int(uint32(hash)&0x7fffffff) % len(buckets)
		buckets[bucket] = &externalModuleTableHashEntry{hash: hash, value: value, next: buckets[bucket]}
		count++
	}
	result := make([]*ExternalModuleTableEntry, 0, count)
	for index := len(buckets) - 1; index >= 0; index-- {
		for entry := buckets[index]; entry != nil; entry = entry.next {
			result = append(result, entry.value)
		}
	}
	return result
}

func NewExternalModuleTable() *ExternalModuleTable {
	return &ExternalModuleTable{
		ModuleHashTable:  NewInsMap[*UniqueString, *ExternalModuleTableEntry](),
		ModuleNodeVector: make([]*ModuleNode, 0),
	}
}

func (t *ExternalModuleTable) require() *ExternalModuleTable {
	if t == nil {
		panic(NewNullPointerException())
	}
	return t
}

func (t *ExternalModuleTable) GetRootModule() *ModuleNode { return t.require().RootModule }
func (t *ExternalModuleTable) SetRootModule(root *ModuleNode) {
	t.require().RootModule = root
}

func (t *ExternalModuleTable) GetContext(name *UniqueString) *SemanticContext {
	t.require()
	if name == nil {
		panic(NewNullPointerException())
	}
	if entry := t.ModuleHashTable.Get(name); entry != nil {
		return entry.Context
	}
	return nil
}

func (t *ExternalModuleTable) GetContextForRootModule() *SemanticContext {
	return t.GetContext(t.GetRootModule().GetName())
}

func (t *ExternalModuleTable) GetModuleNodes() []*ModuleNode {
	t.require()
	result := make([]*ModuleNode, len(t.ModuleNodeVector))
	copy(result, t.ModuleNodeVector)
	return result
}

func (t *ExternalModuleTable) GetModuleNode(name *UniqueString) *ModuleNode {
	t.require()
	if name == nil {
		panic(NewNullPointerException())
	}
	if entry := t.ModuleHashTable.Get(name); entry != nil {
		return entry.ModuleNode
	}
	return nil
}

func (t *ExternalModuleTable) GetModuleNodeByName(name string) *ModuleNode {
	return t.GetModuleNode(UniqueStringOf(name))
}

func (t *ExternalModuleTable) Put(name *UniqueString, context *SemanticContext, module *ModuleNode) {
	t.require()
	if name == nil {
		panic(NewNullPointerException())
	}
	if t.ModuleHashTable.Get(name) == nil {
		t.ModuleHashTable.Set(name, &ExternalModuleTableEntry{ModuleNode: module, Context: context})
		t.ModuleNodeVector = append(t.ModuleNodeVector, module)
	}
}

func (t *ExternalModuleTable) LevelDataToString() string {
	t.require()
	return "Dummy level string"
}
