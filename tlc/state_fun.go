package tlc

import (
	"fmt"
	"strings"
)

type TLCStateFun struct {
	Name  *SymbolNode
	Value Value
	Next  *TLCStateFun
}

var TLCStateFunEmpty = &TLCStateFun{}

func NewTLCStateFun(name *SymbolNode, value Value, state *TLCStateFun) *TLCStateFun {
	if state == nil {
		state = TLCStateFunEmpty
	}
	return &TLCStateFun{Name: name, Value: value, Next: state}
}

func (s *TLCStateFun) CreateEmpty() *TLCStateFun {
	return TLCStateFunEmpty
}

func (s *TLCStateFun) Bind(name *UniqueString, value Value) *TLCStateFun {
	panic(newTLCError(ECGeneral, "TLCStateFun.Bind: This is a TLC bug."))
}

func (s *TLCStateFun) BindSymbol(id *SymbolNode, value Value) *TLCStateFun {
	if s == nil {
		s = TLCStateFunEmpty
	}
	return NewTLCStateFun(id, value, s)
}

func (s *TLCStateFun) Unbind(name *UniqueString) *TLCStateFun {
	panic(newTLCError(ECGeneral, "TLCStateFun.Unbind: This is a TLC bug."))
}

func (s *TLCStateFun) Lookup(name *UniqueString) Value {
	for cur := s; cur != nil && cur != TLCStateFunEmpty; cur = cur.Next {
		if cur.Name != nil && cur.Name.Name == name {
			return cur.Value
		}
	}
	return nil
}

func (s *TLCStateFun) ContainsKey(name *UniqueString) bool {
	return s.Lookup(name) != nil
}

func (s *TLCStateFun) Copy() *TLCStateFun {
	return s
}

func (s *TLCStateFun) DeepCopy() *TLCStateFun {
	panic(newTLCError(ECGeneral, "TLCStateFun.DeepCopy: This is a TLC bug."))
}

func (s *TLCStateFun) DeepNormalize() {
	panic(newTLCError(ECGeneral, "TLCStateFun.DeepNormalize: This is a TLC bug."))
}

func (s *TLCStateFun) FingerPrint() uint64 {
	panic(newTLCError(ECGeneral, "TLCStateFun.FingerPrint: This is a TLC bug."))
}

func (s *TLCStateFun) AllAssigned() bool {
	return true
}

func (s *TLCStateFun) Unassigned() []StateVariable {
	return nil
}

func (s *TLCStateFun) AddToContext(c *Context) *Context {
	if c == nil {
		c = EmptyContext
	}
	c1 := c
	for cur := s; cur != nil && cur != TLCStateFunEmpty; cur = cur.Next {
		c1 = c1.Cons(cur.Name, cur.Value)
	}
	return c1
}

func (s *TLCStateFun) Read(in *ValueInputStream) error {
	return newTLCError(ECGeneral, "TLCStateFun.Read: This is a TLC bug.")
}

func (s *TLCStateFun) Write(out *ValueOutputStream) error {
	return newTLCError(ECGeneral, "TLCStateFun.Write: This is a TLC bug.")
}

func (s *TLCStateFun) String() string {
	if s == nil || s == TLCStateFunEmpty {
		return "[]"
	}
	var b strings.Builder
	b.WriteByte('[')
	b.WriteString(s.Name.Name.String())
	b.WriteString(" -> ")
	b.WriteString(fmt.Sprint(s.Value))
	for cur := s.Next; cur != nil && cur != TLCStateFunEmpty; cur = cur.Next {
		b.WriteString(", ")
		b.WriteString(cur.Name.Name.String())
		b.WriteString("->")
		b.WriteString(fmt.Sprint(cur.Value))
	}
	b.WriteByte(']')
	return b.String()
}
