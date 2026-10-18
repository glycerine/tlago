package tlago

import (
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// SanySymbolTable adapts the public stack and resolution core of
// tla2sany.semantic.SymbolTable over the shared runtime semantic contexts.
type SanySymbolTable struct {
	contextStack []*SanyContext
	topContext   *SanyContext
	baseContext  *SanyContext
	moduleTable  *SanyExternalModuleTable
	module       *tlc.ModuleNode
	errors       *SanyErrors
}

// NewSanySymbolTable constructs the external-module form with a duplicate of
// the current global built-in context.
func NewSanySymbolTable(table *SanyExternalModuleTable, errors *SanyErrors) *SanySymbolTable {
	top := SanyContextGetGlobalContext().Duplicate(table)
	return &SanySymbolTable{
		contextStack: []*SanyContext{top}, topContext: top, baseContext: top,
		moduleTable: table, errors: errors,
	}
}

// NewSanyInnerSymbolTable copies the parent's context stack. Java's inner
// constructor does not initialize topContext; a later push/pop does so.
func NewSanyInnerSymbolTable(table *SanyExternalModuleTable, errors *SanyErrors, parent *SanySymbolTable) *SanySymbolTable {
	if parent == nil {
		panic(tlc.NewNullPointerException())
	}
	stack := append([]*SanyContext(nil), parent.contextStack...)
	var base *SanyContext
	if len(stack) != 0 {
		base = stack[len(stack)-1]
	}
	return &SanySymbolTable{contextStack: stack, baseContext: base, moduleTable: table, module: parent.module, errors: errors}
}

func (s *SanySymbolTable) require() *SanySymbolTable {
	if s == nil {
		panic(tlc.NewNullPointerException())
	}
	return s
}

func (s *SanySymbolTable) GetExternalContext() *SanyContext { return s.require().baseContext }
func (s *SanySymbolTable) GetContext() *SanyContext         { return s.require().topContext }

func (s *SanySymbolTable) PushContext(context *SanyContext) {
	s.require()
	if context == nil {
		panic(tlc.NewNullPointerException())
	}
	s.contextStack = append([]*SanyContext{context}, s.contextStack...)
	s.topContext = context
}

func (s *SanySymbolTable) PopContext() {
	s.require()
	if len(s.contextStack) == 0 {
		panic(tlc.NewNoSuchElementException())
	}
	s.contextStack = s.contextStack[1:]
	if len(s.contextStack) == 0 {
		s.topContext = nil
	} else {
		s.topContext = s.contextStack[0]
	}
}

func (s *SanySymbolTable) SetModuleNode(module *tlc.ModuleNode) {
	s.require().module = module
}

func (s *SanySymbolTable) GetModuleNode() *tlc.ModuleNode { return s.require().module }

func (s *SanySymbolTable) ResolveSymbol(name *tlc.UniqueString) tlc.SemanticNode {
	s.require()
	for _, context := range s.contextStack {
		if node := context.GetSymbol(SanyContextKey{Name: name}); node != nil {
			return node
		}
	}
	return nil
}

func (s *SanySymbolTable) ResolveModule(name *tlc.UniqueString) *tlc.ModuleNode {
	s.require()
	for _, context := range s.contextStack {
		if node := context.GetSymbol(SanyContextKey{Name: name, Module: true}); node != nil {
			module, ok := node.(*tlc.ModuleNode)
			if !ok {
				panic(tlc.NewClassCastException())
			}
			return module
		}
	}
	return s.moduleTable.GetModuleNode(name)
}

func sanySymbolTableCode(value int) SanyErrorCode {
	return SanyErrorCodeFromStandardValue(value)
}

func sanySymbolTableArity(node tlc.SemanticNode) int {
	withArity, ok := node.(interface{ GetArity() int })
	if !ok {
		panic(tlc.NewClassCastException("context entry is not a SymbolNode"))
	}
	return withArity.GetArity()
}

func sanySymbolTableSameOriginalModule(incoming, existing tlc.SemanticNode) bool {
	switch incoming := incoming.(type) {
	case *tlc.SymbolNode:
		existing, ok := existing.(*tlc.SymbolNode)
		return ok && incoming.SameOriginallyDefinedInModule(existing)
	case *tlc.OpDefNode:
		existing, ok := existing.(*tlc.OpDefNode)
		if !ok || incoming.GetSource() != existing.GetSource() {
			return false
		}
		module := incoming.GetSource().GetOriginallyDefinedInModuleNode()
		return module == nil || len(module.GetConstantDecls()) == 0 && len(module.GetVariableDecls()) == 0
	case *tlc.ThmOrAssumpDefNode:
		existing, ok := existing.(*tlc.ThmOrAssumpDefNode)
		if !ok || incoming.GetSource() != existing.GetSource() {
			return false
		}
		module := incoming.GetSource().GetOriginallyDefinedInModuleNode()
		return module == nil || len(module.GetConstantDecls()) == 0 && len(module.GetVariableDecls()) == 0
	default:
		return false
	}
}

// AddSymbol registers a visible symbol or records the source diagnostic for a
// conflicting binding. Warning 4801 still reports successful registration.
func (s *SanySymbolTable) AddSymbol(name *tlc.UniqueString, symbol tlc.SemanticNode) bool {
	current := s.ResolveSymbol(name)
	symbolIsNull := sanyExploreNull(symbol)
	if current == nil && symbolIsNull || !symbolIsNull && current == symbol {
		return true
	}
	if current == nil {
		s.topContext.AddSymbolToContext(SanyContextKey{Name: name}, symbol)
		return true
	}
	if symbolIsNull {
		panic(tlc.NewNullPointerException())
	}
	location := sanyRuntimeContextNodeLocation(symbol)
	if sanyRuntimeContextNodeLocation(current).Source == "--TLA+ BUILTINS--" {
		s.errors.AddMessage(sanySymbolTableCode(4202), location,
			"Symbol %s is a built-in operator, and cannot be redefined.", name.String())
		return false
	}
	newKind, oldKind := tlc.SemanticKindOf(symbol), tlc.SemanticKindOf(current)
	if newKind == tlc.SemanticFormalParamKind || newKind == tlc.SemanticBoundSymbolKind ||
		oldKind != newKind || sanySymbolTableArity(current) != sanySymbolTableArity(symbol) {
		s.errors.AddMessage(sanySymbolTableCode(4201), location,
			"Multiply-defined symbol '%s': this definition or declaration conflicts \nwith the one at %s.",
			name.String(), sanyRuntimeContextNodeLocation(current).String())
		return false
	}
	if sanySymbolTableSameOriginalModule(symbol, current) {
		return true
	}
	s.errors.AddMessage(sanySymbolTableCode(4801), location,
		"Multiple declarations or definitions for symbol %s.  \nThis duplicates the one at %s.",
		name.String(), sanyRuntimeContextNodeLocation(current).String())
	return true
}

// AddModule registers a module in the separate module-name namespace.
func (s *SanySymbolTable) AddModule(name *tlc.UniqueString, module *tlc.ModuleNode) bool {
	current := s.ResolveModule(name)
	if current == module {
		return true
	}
	if current == nil {
		s.topContext.AddSymbolToContext(SanyContextKey{Name: name, Module: true}, module)
		return true
	}
	if module == nil {
		panic(tlc.NewNullPointerException())
	}
	s.errors.AddMessage(sanySymbolTableCode(4223), sanyRuntimeContextNodeLocation(module),
		"Multiply-defined module '%s': this definition or declaration conflicts \nwith the one at %s.",
		name.String(), sanyRuntimeContextNodeLocation(current).String())
	return false
}

func sanySymbolTableNodeName(node tlc.SemanticNode) *tlc.UniqueString {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	switch node := node.(type) {
	case *tlc.SymbolNode:
		return node.GetName()
	case *tlc.OpDefNode:
		return node.Name
	case *tlc.ThmOrAssumpDefNode:
		return node.Name
	case *tlc.ModuleNode:
		return node.GetName()
	default:
		panic(tlc.NewClassCastException("context entry is not a SymbolNode"))
	}
}

func sanySymbolTableContextEntries(context *SanyContext) []string {
	// Context.getContextEntryStringVector obtains this context before checking
	// depth or includeInitial. SymbolTable always supplies depth=1 and true.
	_ = context.ModuleTable.GetContext(tlc.UniqueStringOf("Naturals"))
	history := context.SnapshotPairHistory()
	entries := make([]string, 0, len(history))
	for _, pair := range history {
		name := sanySymbolTableNodeName(pair.Node)
		current := context.GetSymbol(SanyContextKey{Name: name})
		withUID, ok := current.(interface{ GetUID() int32 })
		if !ok {
			if current == nil {
				panic(tlc.NewNullPointerException())
			}
			panic(tlc.NewClassCastException("context entry is not a SemanticNode"))
		}
		entries = append(entries, "\nContext Entry: "+name.String()+"  "+strconv.FormatInt(int64(withUID.GetUID()), 10)+" ")
	}
	for left, right := 0, len(entries)-1; left < right; left, right = left+1, right-1 {
		entries[left], entries[right] = entries[right], entries[left]
	}
	return entries
}

// String preserves SymbolTable.toString's context headers, Pair-history order,
// current-binding lookup and trailing separators.
func (s *SanySymbolTable) String() string {
	s.require()
	var result strings.Builder
	result.WriteString("\n\n***SymbolTable\n\n*** top context")
	for _, context := range s.contextStack {
		for _, entry := range sanySymbolTableContextEntries(context) {
			result.WriteString(entry)
		}
		result.WriteString("\n\n*** next context\n")
	}
	return result.String()
}
