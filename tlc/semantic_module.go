// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Port of SANY ExternalModuleTable and the module/context access used by TLC.

package tlc

import "unicode/utf16"

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

func (c *SemanticContext) GetOpDefs() []*OpDefNode {
	var result []*OpDefNode
	for pair := c.lastPair; pair != nil; pair = pair.link {
		if def, ok := pair.node.(*OpDefNode); ok && def.Kind() != SemanticBuiltInKind && def.Kind() != SemanticModuleInstanceKind {
			result = append(result, def)
		}
	}
	return result
}

func (c *SemanticContext) GetThmOrAssDefs() []*ThmOrAssumpDefNode {
	var result []*ThmOrAssumpDefNode
	for pair := c.lastPair; pair != nil; pair = pair.link {
		if def, ok := pair.node.(*ThmOrAssumpDefNode); ok {
			result = append(result, def)
		}
	}
	return result
}

func (c *SemanticContext) declarations(kind SymbolKind) []*SymbolNode {
	var result []*SymbolNode
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
	SemanticNodeBase
	Name         *UniqueString
	Context      *SemanticContext
	Extendees    []*ModuleNode
	Instantiated bool
	Standard     bool
	TopLevel     []SemanticNode

	constantDecls      []*SymbolNode
	variableDecls      []*SymbolNode
	opDefs             []*OpDefNode
	thmOrAssDefs       []*ThmOrAssumpDefNode
	innerModules       []*ModuleNode
	theoremVec         []*TheoremNode
	theorems           []*TheoremNode
	extendedModuleSets map[bool]*InsMap[*ModuleNode, struct{}]
}

func NewModuleNode(name string, context *SemanticContext) *ModuleNode {
	return &ModuleNode{SemanticNodeBase: NewSemanticNodeBase(SemanticModuleKind, name), Name: UniqueStringOf(name), Context: context}
}

func (m *ModuleNode) GetName() *UniqueString { return m.Name }
func (m *ModuleNode) GetArity() int          { return -2 }
func (m *ModuleNode) GetLevel() int {
	panic(NewWrongInvocationException("Internal Error: Should never call ModuleNode.getLevel()"))
}
func (m *ModuleNode) IsLocal() bool                { return false }
func (m *ModuleNode) GetContext() *SemanticContext { return m.Context }
func (m *ModuleNode) IsInstantiated() bool         { return m.Instantiated }
func (m *ModuleNode) SetInstantiated(value bool)   { m.Instantiated = value }
func (m *ModuleNode) IsStandard() bool             { return m.Standard }
func (m *ModuleNode) SetStandard(value bool)       { m.Standard = value }
func (m *ModuleNode) IsParameterFree() bool {
	return len(m.GetConstantDecls()) == 0 && len(m.GetVariableDecls()) == 0
}
func (m *ModuleNode) ProcessConstantDefns() bool { return !m.IsInstantiated() || m.IsParameterFree() }

// The bridge appends views of the actual retained theorem statements, including
// EXTENDS copies. Named definition enumeration has a different membership.
func (m *ModuleNode) AddTheoremStatement(node *TheoremNode) {
	m.theoremVec = append(m.theoremVec, node)
}

func (m *ModuleNode) GetTheorems() []*TheoremNode {
	if m.theorems == nil {
		m.theorems = make([]*TheoremNode, len(m.theoremVec))
		copy(m.theorems, m.theoremVec)
	}
	return m.theorems
}

func (m *ModuleNode) CreateExtendeeArray(extendees []*ModuleNode) {
	m.Extendees = append([]*ModuleNode(nil), extendees...)
}

func (m *ModuleNode) GetConstantDecls() []*SymbolNode {
	if m.constantDecls == nil {
		m.constantDecls = reverseSemanticSlice(m.Context.GetConstantDecls())
	}
	return m.constantDecls
}

func (m *ModuleNode) GetVariableDecls() []*SymbolNode {
	if m.variableDecls == nil {
		m.variableDecls = reverseSemanticSlice(m.Context.GetVariableDecls())
	}
	return m.variableDecls
}

func (m *ModuleNode) GetOpDefs() []*OpDefNode {
	if m.opDefs == nil {
		m.opDefs = reverseSemanticSlice(m.Context.GetOpDefs())
	}
	return m.opDefs
}

func (m *ModuleNode) GetOpDef(name *UniqueString) *OpDefNode {
	for _, def := range m.GetOpDefs() {
		if def.Name == name {
			return def
		}
	}
	return nil
}

func (m *ModuleNode) GetThmOrAssDefs() []*ThmOrAssumpDefNode {
	if m.thmOrAssDefs == nil {
		m.thmOrAssDefs = reverseSemanticSlice(m.Context.GetThmOrAssDefs())
	}
	return m.thmOrAssDefs
}

func (m *ModuleNode) GetInnerModules() []*ModuleNode {
	if m.innerModules == nil {
		m.innerModules = m.Context.GetModDefs()
	}
	return m.innerModules
}

func (m *ModuleNode) GetExtendedModuleSet(recursive ...bool) *InsMap[*ModuleNode, struct{}] {
	recurse := len(recursive) == 0 || recursive[0]
	if cached := m.extendedModuleSets[recurse]; cached != nil {
		return cached
	}
	result := NewInsMap[*ModuleNode, struct{}]()
	var visit func(*ModuleNode)
	visit = func(module *ModuleNode) {
		for _, extendee := range module.Extendees {
			if _, exists := result.Get2(extendee); exists {
				continue
			}
			result.Set(extendee, struct{}{})
			if recurse {
				visit(extendee)
			}
		}
	}
	visit(m)
	if m.extendedModuleSets == nil {
		m.extendedModuleSets = map[bool]*InsMap[*ModuleNode, struct{}]{}
	}
	m.extendedModuleSets[recurse] = result
	return result
}

func (m *ModuleNode) ExtendsModule(module *ModuleNode) bool {
	_, exists := m.GetExtendedModuleSet().Get2(module)
	return exists
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

func (e *ExternalModuleTableEntry) GetModuleNode() *ModuleNode { return e.ModuleNode }

type ExternalModuleTable struct {
	ModuleHashTable  *InsMap[*UniqueString, *ExternalModuleTableEntry]
	ModuleNodeVector []*ModuleNode
	RootModule       *ModuleNode
}

func NewExternalModuleTable() *ExternalModuleTable {
	return &ExternalModuleTable{ModuleHashTable: NewInsMap[*UniqueString, *ExternalModuleTableEntry]()}
}

func (t *ExternalModuleTable) GetRootModule() *ModuleNode     { return t.RootModule }
func (t *ExternalModuleTable) SetRootModule(root *ModuleNode) { t.RootModule = root }

func (t *ExternalModuleTable) GetContext(name *UniqueString) *SemanticContext {
	if entry := t.ModuleHashTable.Get(name); entry != nil {
		return entry.Context
	}
	return nil
}

func (t *ExternalModuleTable) GetContextForRootModule() *SemanticContext {
	return t.GetContext(t.RootModule.Name)
}

func (t *ExternalModuleTable) GetModuleNodes() []*ModuleNode {
	return append([]*ModuleNode(nil), t.ModuleNodeVector...)
}

func (t *ExternalModuleTable) GetModuleNode(name *UniqueString) *ModuleNode {
	if entry := t.ModuleHashTable.Get(name); entry != nil {
		return entry.ModuleNode
	}
	return nil
}

func (t *ExternalModuleTable) Put(name *UniqueString, context *SemanticContext, module *ModuleNode) {
	if t.ModuleHashTable.Get(name) == nil {
		t.ModuleHashTable.Set(name, &ExternalModuleTableEntry{ModuleNode: module, Context: context})
		t.ModuleNodeVector = append(t.ModuleNodeVector, module)
	}
}
