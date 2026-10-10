package tlc

import (
	"fmt"
	"strings"
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
	Definition *OpDefNode
	Kind       SymbolKind
	Arity      int
	Location   SourceLocation
	TreeNode   any
}

// Parser-backed symbol views share their declaration/definition's semantic
// owner. Native lookup aliases retain their separate adapter fields.
func (s *SymbolNode) hasSourceSyntax() bool {
	return s != nil && s.SemanticBase.hasSourceSyntax()
}

func (s *SymbolNode) GetSourceLocation() SourceLocation {
	if s.hasSourceSyntax() {
		return s.SemanticBase.GetSourceLocation()
	}
	return s.Location
}

func (s *SymbolNode) SetSourceLocation(location SourceLocation) { s.Location = location }

func (s *SymbolNode) GetTreeNode() any {
	if s.hasSourceSyntax() {
		return s.SemanticBase.GetTreeNode()
	}
	return s.TreeNode
}

func (s *SymbolNode) SetTreeNode(tree any) {
	if s.hasSourceSyntax() {
		s.SemanticBase.SetTreeNode(tree)
	}
	s.TreeNode = tree
}

func NewFormalParamSymbolNode(name string, arity int) *SymbolNode {
	base := NewSemanticNodeBase(SemanticFormalParamKind, name)
	return &SymbolNode{SemanticBase: &base, Name: UniqueStringOf(name), Arity: arity, Kind: SymbolFormalParam}
}

func (s *SymbolNode) GetToolObjectAt(toolID int32) any {
	return s.SemanticBase.GetToolObjectAt(toolID)
}

func (s *SymbolNode) SetToolObjectAt(toolID int32, value any) {
	if s.SemanticBase == nil {
		panic(NewClassCastException("lookup alias has no semantic declaration"))
	}
	s.SemanticBase.SetToolObjectAt(toolID, value)
}

func (s *SymbolNode) GetName() *UniqueString {
	if s.DeclarationName != nil {
		return s.DeclarationName
	}
	return s.Name
}

func (s *SymbolNode) GetSignature() string {
	if s.Definition != nil {
		return s.Definition.GetSignature()
	}
	return s.GetName().String()
}

func (s *SymbolNode) GetHumanReadableImage() string {
	return semanticNodeHoverImage(s)
}

func NewSymbolNode(name string) *SymbolNode {
	return &SymbolNode{Name: UniqueStringOf(name)}
}

func NewVariableSymbolNode(name string) *SymbolNode {
	return &SymbolNode{Name: UniqueStringOf(name), Kind: SymbolVariableDecl}
}

func (s *SymbolNode) MarkVariableDecl() {
	if s != nil {
		s.Kind = SymbolVariableDecl
	}
}

func (s *SymbolNode) MarkUserDefinedOp() {
	if s != nil {
		s.Kind = SymbolUserDefinedOp
	}
}

func (s *SymbolNode) IsUserDefinedOp() bool {
	return s != nil && s.Kind == SymbolUserDefinedOp
}

func (s *SymbolNode) IsVariableDecl() bool {
	return s != nil && s.Kind == SymbolVariableDecl
}

func (s *SymbolNode) String() string {
	if s == nil || s.Name == nil {
		return ""
	}
	return s.Name.String()
}

type Context struct {
	name  *SymbolNode
	value any
	next  *Context
}

var (
	EmptyContext      = &Context{}
	baseBranchContext = &Context{next: EmptyContext}
)

func BranchContext(base *Context) *Context {
	if base == nil || base == EmptyContext {
		return baseBranchContext
	}
	return &Context{next: base}
}

func (c *Context) Cons(name *SymbolNode, value any) *Context {
	if c == nil {
		c = EmptyContext
	}
	return &Context{name: name, value: value, next: c}
}

func (c *Context) Lookup(sym *SymbolNode) any {
	for cur := c; cur != nil && cur != EmptyContext; cur = cur.next {
		if cur.name == sym {
			return cur.value
		}
	}
	return nil
}

func (c *Context) LookupFunc(pred func(*SymbolNode) bool) any {
	for cur := c; cur != nil && cur != EmptyContext; cur = cur.next {
		if pred(cur.name) {
			return cur.value
		}
	}
	return nil
}

func (c *Context) LookupName(pred func(*SymbolNode) bool) *SymbolNode {
	for cur := c; cur != nil && cur != EmptyContext; cur = cur.next {
		if pred(cur.name) {
			return cur.name
		}
	}
	return nil
}

func (c *Context) LookupCutoff(sym *SymbolNode, cutoff bool) any {
	for cur := c; cur != nil && cur != EmptyContext; cur = cur.next {
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
	res := NewInsMap[*UniqueString, Value]()
	c.fillMap(res)
	return res
}

func (c *Context) fillMap(out *InsMap[*UniqueString, Value]) {
	if c == nil || c == EmptyContext {
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
		return v
	}
	if value == nil {
		return ValUndef
	}
	return NewStringValue(fmt.Sprint(value))
}

func (c *Context) String() string {
	var parts []string
	for cur := c; cur != nil && cur != EmptyContext; cur = cur.next {
		if cur.name == nil {
			continue
		}
		parts = append(parts, cur.name.String()+"->"+toContextString(cur.value))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func toContextString(value any) string {
	if value == nil {
		return "<nil>"
	}
	if s, ok := value.(interface{ String() string }); ok {
		return s.String()
	}
	return fmt.Sprint(value)
}

func (c *Context) Name() *SymbolNode {
	if c == nil {
		return nil
	}
	return c.name
}

func (c *Context) Value() any {
	if c == nil {
		return nil
	}
	return c.value
}

func (c *Context) Next() *Context {
	if c == nil {
		return nil
	}
	return c.next
}

func (c *Context) IsEmpty() bool {
	return c == nil || c == EmptyContext
}

func (c *Context) IsDeepEmpty() bool {
	return c == nil || c == EmptyContext || c == baseBranchContext
}

func (c *Context) Depth() int {
	if c == nil {
		return 0
	}
	depth := 1
	for child := c.Next(); child != nil && child.Next() != nil; child = child.Next() {
		depth++
	}
	return depth
}

func (c *Context) DeepCopy() *Context {
	if c == nil || c == EmptyContext {
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
	out := &ContextEnumerator{
		con:          con,
		vars:         vars,
		enums:        enums,
		currentElems: make([]Value, len(enums)),
	}
	if out.con == nil {
		out.con = EmptyContext
	}
	for i, enum := range enums {
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
	if e.done || e.err != nil {
		return nil
	}
	con := e.con
	for i := range e.enums {
		switch vars := e.vars[i].(type) {
		case *SymbolNode:
			con = con.Cons(vars, e.currentElems[i])
		case []*SymbolNode:
			tuple, ok := e.currentElems[i].(*TupleValue)
			if !ok || len(vars) != len(tuple.Elems) {
				// ContextEnumerator calls Assert.fail with the first tuple formal.
				if len(vars) == 0 {
					panic(NewArrayIndexOutOfBoundsException(0, 0))
				}
				e.err = NewTLCRuntimeException(ECTLCArgumentMismatch, vars[0].String())
				e.done = true
				return nil
			}
			for j, variable := range vars {
				con = con.Cons(variable, tuple.Elems[j])
			}
		default:
			e.err = newTLCError(ECGeneral, "unsupported context variable binding %T", e.vars[i])
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
	return e.done
}

func (e *ContextEnumerator) Err() error {
	return e.err
}
