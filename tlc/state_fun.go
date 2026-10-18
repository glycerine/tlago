package tlc

import (
	"strings"
)

type TLCStateFun struct {
	Name  *SymbolNode
	Value Value
	Next  *TLCStateFun

	ownersCaptured bool
	nameOwner      *SymbolNode
	valueOwner     Value
	nextOwner      *TLCStateFun
}

var TLCStateFunEmpty = newTLCStateFun(nil, nil, nil)

func NewTLCStateFun(name *SymbolNode, value Value, state *TLCStateFun) *TLCStateFun {
	if state == nil {
		state = TLCStateFunEmpty
	}
	return newTLCStateFun(name, value, state)
}

func newTLCStateFun(name *SymbolNode, value Value, next *TLCStateFun) *TLCStateFun {
	return &TLCStateFun{
		Name: name, Value: value, Next: next,
		ownersCaptured: true, nameOwner: name, valueOwner: value, nextOwner: next,
	}
}

func (s *TLCStateFun) ownedName() *SymbolNode {
	if s.ownersCaptured {
		return s.nameOwner
	}
	return s.Name
}

func (s *TLCStateFun) ownedValue() Value {
	if s.ownersCaptured {
		return s.valueOwner
	}
	return s.Value
}

func (s *TLCStateFun) ownedNext() *TLCStateFun {
	if s.ownersCaptured {
		return s.nextOwner
	}
	return s.Next
}

func (s *TLCStateFun) CreateEmpty() *TLCStateFun {
	s.require()
	return TLCStateFunEmpty
}

func (s *TLCStateFun) Bind(name *UniqueString, value Value) *TLCStateFun {
	s.require()
	panic(NewWrongInvocationException("TLCStateFun.bind: This is a TLC bug."))
}

func (s *TLCStateFun) BindSymbol(id *SymbolNode, value Value) *TLCStateFun {
	if s == nil {
		panic(NewNullPointerException())
	}
	return NewTLCStateFun(id, value, s)
}

func (s *TLCStateFun) Unbind(name *UniqueString) *TLCStateFun {
	s.require()
	panic(NewWrongInvocationException("TLCStateFun.unbind: This is a TLC bug."))
}

func (s *TLCStateFun) Lookup(name *UniqueString) Value {
	for cur := s; cur != TLCStateFunEmpty; cur = cur.ownedNext() {
		if cur == nil || cur.ownedName() == nil {
			panic(NewNullPointerException())
		}
		if cur.ownedName().GetName() == name {
			return cur.ownedValue()
		}
	}
	return nil
}

func (s *TLCStateFun) ContainsKey(name *UniqueString) bool {
	return !isNil(s.Lookup(name))
}

func (s *TLCStateFun) Copy() *TLCStateFun {
	s.require()
	return s
}

func (s *TLCStateFun) DeepCopy() *TLCStateFun {
	s.require()
	panic(NewWrongInvocationException("TLCStateFun.deepCopy: This is a TLC bug."))
}

func (s *TLCStateFun) DeepNormalize() {
	s.require()
	panic(NewWrongInvocationException("TLCStateFun.normalizeFcns: This is a TLC bug."))
}

func (s *TLCStateFun) FingerPrint() uint64 {
	s.require()
	panic(NewWrongInvocationException("TLCStateFun.fingerPrint: This is a TLC bug."))
}

func (s *TLCStateFun) AllAssigned() bool {
	s.require()
	return true
}

func (s *TLCStateFun) Unassigned() []StateVariable {
	s.require()
	return []StateVariable{}
}

func (s *TLCStateFun) AddToContext(c *Context) *Context {
	c1 := c
	for cur := s; cur != TLCStateFunEmpty; cur = cur.ownedNext() {
		if cur == nil {
			panic(NewNullPointerException())
		}
		c1 = c1.Cons(cur.ownedName(), cur.ownedValue())
	}
	return c1
}

func (s *TLCStateFun) Read(in *ValueInputStream) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	return NewWrongInvocationException("TLCStateFun.read: This is a TLC bug.")
}

func (s *TLCStateFun) Write(out *ValueOutputStream) error {
	if s == nil {
		panic(NewNullPointerException())
	}
	return NewWrongInvocationException("TLCStateFun.write: This is a TLC bug.")
}

func (s *TLCStateFun) String() string {
	if s == nil {
		panic(NewNullPointerException())
	}
	var b strings.Builder
	b.WriteByte('[')
	if s != TLCStateFunEmpty {
		b.WriteString(functionalSymbolName(s.ownedName()))
		b.WriteString(" -> ")
		if isNil(s.ownedValue()) {
			panic(NewNullPointerException())
		}
		b.WriteString(s.ownedValue().String())
		for cur := s.ownedNext(); cur != TLCStateFunEmpty; cur = cur.ownedNext() {
			if cur == nil {
				panic(NewNullPointerException())
			}
			b.WriteString(", ")
			b.WriteString(functionalSymbolName(cur.ownedName()))
			b.WriteString("->")
			if isNil(cur.ownedValue()) {
				b.WriteString("null")
			} else {
				b.WriteString(cur.ownedValue().String())
			}
		}
	}
	b.WriteByte(']')
	return b.String()
}

func functionalSymbolName(symbol *SymbolNode) string {
	if symbol == nil || symbol.GetName() == nil {
		panic(NewNullPointerException())
	}
	return symbol.GetName().String()
}

func (s *TLCStateFun) require() {
	if s == nil {
		panic(NewNullPointerException())
	}
}
