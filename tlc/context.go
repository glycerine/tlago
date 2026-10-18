package tlc

import (
	"fmt"
	"strings"
	"sync"
)

type SymbolKind int

const (
	SymbolUnknown SymbolKind = iota
	SymbolVariableDecl
	SymbolUserDefinedOp
	SymbolConstantDecl
	SymbolBuiltIn
	SymbolFormalParam
)

type SymbolNode struct {
	// The runtime declaration view retains the actual semantic owner when
	// supplied by SANY. Generic lookup aliases may have no semantic base yet.
	SemanticBase *SemanticNodeBase
	Name         *UniqueString
	Data         any
	// The evaluator's qualified lookup key can differ from SANY's declaration name.
	DeclarationName *UniqueString
	// SANY's operator symbol is the OpDefNode itself. Retain that semantic
	// identity separately from evaluation overrides and context bindings.
	Definition                *OpDefNode
	Kind                      SymbolKind
	Arity                     int
	Declaration               bool
	Module                    *ModuleNode
	OriginallyDefinedInModule *ModuleNode
	Location                  SourceLocation
	TreeNode                  any
	// SANYSymbolTable retains OpDefOrDeclNode's source SymbolTable adapter.
	// It is typed as any here to keep package tlc independent of package tlago.
	SANYSymbolTable any
	// Variable declarations own their coverage counters, as OpDeclNode does.
	countDistinct *CountDistinct
}

var semanticBuiltInResolver struct {
	sync.RWMutex
	resolve func(SemanticNode) bool
}

// SetSemanticBuiltInResolver connects the runtime semantic view to SANY's
// process-wide initial context without introducing a package import cycle.
func SetSemanticBuiltInResolver(resolve func(SemanticNode) bool) {
	semanticBuiltInResolver.Lock()
	semanticBuiltInResolver.resolve = resolve
	semanticBuiltInResolver.Unlock()
}

func semanticNodeIsBuiltIn(node SemanticNode, fallback bool) bool {
	semanticBuiltInResolver.RLock()
	resolve := semanticBuiltInResolver.resolve
	semanticBuiltInResolver.RUnlock()
	if resolve == nil {
		return fallback
	}
	return resolve(node)
}

func (s *SymbolNode) SetCountDistinct(counter *CountDistinct) *SymbolNode {
	s.require()
	s.countDistinct = counter
	return s
}

func (s *SymbolNode) GetCountDistinct() *CountDistinct {
	s.require()
	return s.countDistinct
}

func (s *SymbolNode) Count(value Value) *SymbolNode {
	s.require()
	if s.countDistinct == nil {
		panic(NewNullPointerException())
	}
	if value == nil && s.countDistinct.mode != CountDistinctNoop {
		panic(NewNullPointerException())
	}
	s.countDistinct.AddValue(value)
	return s
}

// Parser-backed symbol views share their declaration/definition's semantic
// owner. Native lookup aliases retain their separate adapter fields.
func (s *SymbolNode) hasSourceSyntax() bool {
	return s != nil && s.SemanticBase.hasSourceSyntax()
}

func (s *SymbolNode) GetSourceLocation() SourceLocation {
	s.require()
	if s.hasSourceSyntax() {
		return s.SemanticBase.GetSourceLocation()
	}
	return s.Location
}

func (s *SymbolNode) SetSourceLocation(location SourceLocation) {
	s.require()
	s.Location = location
}

func (s *SymbolNode) GetTreeNode() any {
	s.require()
	if s.hasSourceSyntax() {
		return s.SemanticBase.GetTreeNode()
	}
	return s.TreeNode
}

func (s *SymbolNode) SetTreeNode(tree any) {
	s.require()
	if s.hasSourceSyntax() {
		s.SemanticBase.SetTreeNode(tree)
	}
	s.TreeNode = tree
}

func NewFormalParamSymbolNode(name string, arity int) *SymbolNode {
	base := NewSemanticNodeBase(SemanticFormalParamKind, name)
	return &SymbolNode{SemanticBase: &base, Name: UniqueStringOf(name), Arity: int(int32(arity)), Kind: SymbolFormalParam}
}

func (s *SymbolNode) GetToolObjectAt(toolID int32) any {
	s.require()
	return s.SemanticBase.GetToolObjectAt(toolID)
}

func (s *SymbolNode) SetToolObjectAt(toolID int32, value any) {
	s.require()
	if s.SemanticBase == nil {
		panic(NewClassCastException("lookup alias has no semantic declaration"))
	}
	s.SemanticBase.SetToolObjectAt(toolID, value)
}

func (s *SymbolNode) GetName() *UniqueString {
	s.require()
	if s.DeclarationName != nil {
		return s.DeclarationName
	}
	return s.Name
}

func (s *SymbolNode) GetArity() int {
	s.require()
	if s.SemanticBase != nil {
		if provider, ok := s.SemanticBase.CanonicalLevelData.(interface{ RuntimeSymbolArity() (int, bool) }); ok {
			if arity, present := provider.RuntimeSymbolArity(); present {
				return arity
			}
		}
	}
	return s.Arity
}

func (s *SymbolNode) GetCanonicalLevelData() any {
	s.require()
	if s.SemanticBase == nil {
		return nil
	}
	return s.SemanticBase.GetCanonicalLevelData()
}

func (s *SymbolNode) IsLocal() bool {
	s.require()
	if s.Definition != nil {
		return s.Definition.IsLocal()
	}
	if definition, ok := s.Data.(*ThmOrAssumpDefNode); ok && definition != nil {
		return definition.IsLocal()
	}
	return s.Kind == SymbolFormalParam
}

func (s *SymbolNode) Occur(params []*SymbolNode) bool {
	s.require()
	if params == nil {
		panic(NewNullPointerException())
	}
	for _, param := range params {
		if s == param {
			return true
		}
	}
	return false
}

func (s *SymbolNode) IsParam() bool {
	s.require()
	return s.Declaration || s.Kind == SymbolFormalParam || s.Kind == SymbolVariableDecl || s.Kind == SymbolConstantDecl
}

func (s *SymbolNode) GetModuleNode() *ModuleNode {
	s.require()
	return s.Module
}

func (s *SymbolNode) GetNumberOfArgs() int {
	return s.GetArity()
}

func (s *SymbolNode) SetNumberOfArgs(arity int) {
	s.require()
	arity = int(int32(arity))
	if s.SemanticBase != nil {
		if provider, ok := s.SemanticBase.CanonicalLevelData.(interface{ RuntimeSetSymbolArity(int) bool }); ok {
			provider.RuntimeSetSymbolArity(arity)
		}
	}
	s.Arity = arity
	if s.Definition != nil {
		s.Definition.arityValue = arity
		s.Definition.aritySet = true
	}
}

func (s *SymbolNode) GetOriginallyDefinedInModuleNode() *ModuleNode {
	s.require()
	if s.Definition != nil {
		return s.Definition.GetOriginallyDefinedInModuleNode()
	}
	return s.OriginallyDefinedInModule
}

func (s *SymbolNode) SameOriginallyDefinedInModule(other *SymbolNode) bool {
	s.require()
	if other == nil {
		panic(NewNullPointerException())
	}
	if s.Definition != nil && other.Definition != nil {
		source := s.Definition.GetSource()
		if source != other.Definition.GetSource() {
			return false
		}
		return semanticDefinitionModuleIsParameterFree(source.OriginallyDefinedInModule)
	}
	thisTheorem, thisOK := s.Data.(*ThmOrAssumpDefNode)
	otherTheorem, otherOK := other.Data.(*ThmOrAssumpDefNode)
	if thisOK && otherOK && thisTheorem != nil && otherTheorem != nil {
		source := thisTheorem.GetSource()
		if source != otherTheorem.GetSource() {
			return false
		}
		return semanticDefinitionModuleIsParameterFree(source.OriginallyDefinedInModule)
	}
	return false
}

func semanticDefinitionModuleIsParameterFree(module *ModuleNode) bool {
	return module == nil || len(module.GetConstantDecls()) == 0 && len(module.GetVariableDecls()) == 0
}

func (s *SymbolNode) GetSignature() string {
	s.require()
	if s.Definition != nil {
		return s.Definition.GetSignature()
	}
	return s.GetName().String()
}

// GetComment corresponds to OpDefOrDeclNode.getComment for declaration
// symbols. Definition-backed symbol views delegate to their OpDefNode.
func (s *SymbolNode) GetComment() string {
	s.require()
	if s.Definition != nil {
		return s.Definition.GetComment()
	}
	return semanticOpDefOrDeclComment(s.hasSourceSyntax(), s.GetTreeNode())
}

// IsBuiltIn corresponds to SemanticNode.isBuiltIn for the split Go symbol
// representation. The SANY resolver tests the current global context by source
// identity; standalone TLC package use retains the native kind fallback.
func (s *SymbolNode) IsBuiltIn() bool {
	s.require()
	return semanticNodeIsBuiltIn(s, s.Kind == SymbolBuiltIn)
}

func (s *SymbolNode) GetHumanReadableImage() string {
	s.require()
	return semanticNodeHoverImage(s)
}

func NewSymbolNode(name string) *SymbolNode {
	return &SymbolNode{Name: UniqueStringOf(name)}
}

func NewVariableSymbolNode(name string) *SymbolNode {
	return &SymbolNode{Name: UniqueStringOf(name), Kind: SymbolVariableDecl, Declaration: true}
}

func (s *SymbolNode) MarkVariableDecl() {
	s.require()
	s.Kind = SymbolVariableDecl
	s.Declaration = true
}

func (s *SymbolNode) MarkUserDefinedOp() {
	s.require()
	s.Kind = SymbolUserDefinedOp
}

func (s *SymbolNode) IsUserDefinedOp() bool {
	s.require()
	return s.Kind == SymbolUserDefinedOp
}

func (s *SymbolNode) IsVariableDecl() bool {
	s.require()
	return s.Kind == SymbolVariableDecl
}

func (s *SymbolNode) String() string {
	s.require()
	if s.Name == nil {
		return ""
	}
	return s.Name.String()
}

func (s *SymbolNode) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}

type Context struct {
	name  *SymbolNode
	value any
	next  *Context
}

func (c *Context) require() {
	if c == nil {
		panic(NewNullPointerException())
	}
}

var (
	EmptyContext      = &Context{}
	baseBranchContext = &Context{next: EmptyContext}
)

func BranchContext(base *Context) *Context {
	if base == EmptyContext {
		return baseBranchContext
	}
	return &Context{next: base}
}

func (c *Context) Cons(name *SymbolNode, value any) *Context {
	if c == nil {
		panic(NewNullPointerException())
	}
	return &Context{name: name, value: value, next: c}
}

func (c *Context) Lookup(sym *SymbolNode) any {
	c.require()
	for cur := c; cur != EmptyContext; cur = cur.next {
		cur.require()
		if cur.name == sym {
			return cur.value
		}
	}
	return nil
}

func (c *Context) LookupFunc(pred func(*SymbolNode) bool) any {
	c.require()
	for cur := c; cur != EmptyContext; cur = cur.next {
		cur.require()
		if pred == nil {
			panic(NewNullPointerException())
		}
		if pred(cur.name) {
			return cur.value
		}
	}
	return nil
}

func (c *Context) LookupName(pred func(*SymbolNode) bool) *SymbolNode {
	c.require()
	for cur := c; cur != EmptyContext; cur = cur.next {
		cur.require()
		if pred == nil {
			panic(NewNullPointerException())
		}
		if pred(cur.name) {
			return cur.name
		}
	}
	return nil
}

func (c *Context) LookupCutoff(sym *SymbolNode, cutoff bool) any {
	c.require()
	for cur := c; cur != EmptyContext; cur = cur.next {
		cur.require()
		if cur.name != nil {
			if cur.name == sym {
				return cur.value
			}
		} else if cutoff {
			return nil
		}
	}
	return nil
}

func (c *Context) ToMap() *InsMap[*UniqueString, Value] {
	c.require()
	res := NewInsMap[*UniqueString, Value]()
	c.fillMap(res)
	return res
}

func (c *Context) fillMap(out *InsMap[*UniqueString, Value]) {
	c.require()
	if c == EmptyContext {
		return
	}
	if c.name == nil {
		c.next.fillMap(out)
		return
	}
	for cur := c; cur != nil && cur != EmptyContext && cur.name != nil; cur = cur.next {
		out.Set(cur.name.GetName(), contextValueAsValue(cur.value))
		if cur.next == nil || cur.next == EmptyContext || cur.next.name == nil {
			if cur.next != nil {
				cur.next.fillMap(out)
			}
			return
		}
	}
}

func contextValueAsValue(value any) Value {
	if v, ok := value.(Value); ok {
		if isNil(v) {
			panic(NewNullPointerException())
		}
		return v
	}
	if isNil(value) {
		panic(NewNullPointerException())
	}
	return NewStringValue(fmt.Sprint(value))
}

func (c *Context) String() string {
	c.require()
	var out strings.Builder
	out.WriteByte('[')
	c.appendString(&out)
	out.WriteByte(']')
	return out.String()
}

func (c *Context) appendString(out *strings.Builder) {
	c.require()
	if c.name == nil {
		if c == EmptyContext {
			return
		}
		c.next.appendString(out)
		return
	}
	out.WriteString(c.name.GetName().String())
	out.WriteString("->")
	out.WriteString(toContextString(c.value))
	cur := c.next
	for {
		cur.require()
		if cur.name == nil {
			break
		}
		out.WriteString(", ")
		out.WriteString(cur.name.GetName().String())
		out.WriteString("->")
		out.WriteString(toContextString(cur.value))
		cur = cur.next
	}
	cur.appendString(out)
}

func toContextString(value any) string {
	if isNil(value) {
		return "null"
	}
	if s, ok := value.(interface{ String() string }); ok {
		return s.String()
	}
	return fmt.Sprint(value)
}

func (c *Context) Name() *SymbolNode {
	return c.GetName()
}

func (c *Context) GetName() *SymbolNode {
	c.require()
	return c.name
}

func (c *Context) Value() any {
	return c.GetValue()
}

func (c *Context) GetValue() any {
	c.require()
	return c.value
}

func (c *Context) Next() *Context {
	c.require()
	return c.next
}

func (c *Context) HasNext() bool {
	c.require()
	return c.next != nil
}

func (c *Context) IsEmpty() bool {
	c.require()
	return c == EmptyContext
}

func (c *Context) IsDeepEmpty() bool {
	c.require()
	return c == EmptyContext || c == baseBranchContext
}

func (c *Context) Depth() int {
	c.require()
	depth := 1
	for child := c.Next(); child.HasNext(); child = child.Next() {
		depth++
	}
	return depth
}

func (c *Context) DeepCopy() *Context {
	c.require()
	if c == EmptyContext {
		return EmptyContext
	}
	return &Context{name: c.name, value: c.value, next: c.next.DeepCopy()}
}

type ContextEnumerator struct {
	con          *Context
	vars         []any
	enums        []ValueEnumeration
	currentElems []Value
	done         bool
	err          error
}

func NewContextEnumerator(vars []any, enums []ValueEnumeration, con *Context) *ContextEnumerator {
	if enums == nil {
		panic(NewNullPointerException())
	}
	out := &ContextEnumerator{
		con:          con,
		vars:         vars,
		enums:        enums,
		currentElems: make([]Value, len(enums)),
	}
	for i, enum := range enums {
		if isNil(enum) {
			panic(NewNullPointerException())
		}
		out.currentElems[i] = enum.NextElement()
		if err := enum.Err(); err != nil {
			out.err = err
			out.done = true
			break
		}
		if out.currentElems[i] == nil {
			out.done = true
			break
		}
	}
	return out
}

func (e *ContextEnumerator) NextElement() *Context {
	e.require()
	if e.done || e.err != nil {
		return nil
	}
	con := e.con
	for i := range e.enums {
		if i >= len(e.vars) {
			panic(NewArrayIndexOutOfBoundsException(i, len(e.vars)))
		}
		switch vars := e.vars[i].(type) {
		case *SymbolNode:
			con = con.Cons(vars, e.currentElems[i])
		case []*SymbolNode:
			tuple, ok := e.currentElems[i].(*TupleValue)
			if !ok || tuple == nil || len(vars) != len(tuple.GetElems()) {
				// ContextEnumerator calls Assert.fail with the first tuple formal.
				if len(vars) == 0 {
					panic(NewArrayIndexOutOfBoundsException(0, 0))
				}
				e.err = NewTLCRuntimeException(ECTLCArgumentMismatch, vars[0].String())
				e.done = true
				return nil
			}
			for j, variable := range vars {
				con = con.Cons(variable, tuple.GetElems()[j])
			}
		default:
			e.err = valueStreamClassCast(e.vars[i], "[Ltla2sany.semantic.SymbolNode;")
			e.done = true
			return nil
		}
	}
	e.advance()
	return con
}

func (e *ContextEnumerator) advance() {
	for i := range e.enums {
		e.currentElems[i] = e.enums[i].NextElement()
		if err := e.enums[i].Err(); err != nil {
			e.err = err
			e.done = true
			return
		}
		if e.currentElems[i] != nil {
			return
		}
		if i == len(e.enums)-1 {
			e.done = true
			return
		}
		e.enums[i].Reset()
		e.currentElems[i] = e.enums[i].NextElement()
		if err := e.enums[i].Err(); err != nil {
			e.err = err
			e.done = true
			return
		}
	}
}

func (e *ContextEnumerator) IsDone() bool {
	e.require()
	return e.done
}

func (e *ContextEnumerator) Err() error {
	e.require()
	return e.err
}

func (e *ContextEnumerator) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}
