package tlc

import (
	"fmt"
	"strings"
)

type SxKind int

const (
	SxKindNil SxKind = iota
	SxKindAtom
	SxKindInt
	SxKindPair
)

type Sx struct {
	Kind SxKind
	Atom string
	Int  int
	Car  *Sx
	Cdr  *Sx
}

var (
	SxNil      = &Sx{Kind: SxKindNil}
	sxAtoms    = NewInsMap[string, *Sx]()
	sxSymCount int
)

func SxCons(a *Sx, b *Sx) *Sx {
	return &Sx{Kind: SxKindPair, Car: a, Cdr: b}
}

func SxCar(a *Sx) *Sx {
	if a == nil || a.Kind != SxKindPair {
		panic("Car must be applied to cons.")
	}
	return a.Car
}

func SxCdr(a *Sx) *Sx {
	if a == nil || a.Kind != SxKindPair {
		panic("Cdr must be applied to cons.")
	}
	return a.Cdr
}

func SxList(values ...*Sx) *Sx {
	out := SxNil
	for i := len(values) - 1; i >= 0; i-- {
		out = SxCons(values[i], out)
	}
	return out
}

func SxAppend(a *Sx, b *Sx) *Sx {
	if a == SxNil || a == nil {
		return b
	}
	return SxCons(SxCar(a), SxAppend(SxCdr(a), b))
}

func SxGenSym() *Sx {
	name := fmt.Sprintf("F__%d", sxSymCount)
	sxSymCount++
	return SxAtom(name)
}

func SxMemq(a *Sx, p *Sx) bool {
	for p != nil && p != SxNil {
		if a == SxCar(p) {
			return true
		}
		p = SxCdr(p)
	}
	return false
}

func SxAtom(st string) *Sx {
	if atom, ok := sxAtoms.Get2(st); ok {
		return atom
	}
	atom := &Sx{Kind: SxKindAtom, Atom: st}
	sxAtoms.Set(st, atom)
	return atom
}

func SxFromInt(k int) *Sx {
	return &Sx{Kind: SxKindInt, Int: k}
}

func (s *Sx) String() string {
	if s == nil {
		return "nil"
	}
	switch s.Kind {
	case SxKindNil:
		return "nil"
	case SxKindAtom:
		return s.Atom
	case SxKindInt:
		return fmt.Sprint(s.Int)
	case SxKindPair:
		var b strings.Builder
		b.WriteByte('(')
		b.WriteString(s.Car.String())
		next := s.Cdr
		for next != nil && next.Kind == SxKindPair {
			b.WriteByte(' ')
			b.WriteString(next.Car.String())
			next = next.Cdr
		}
		if next != nil && next != SxNil {
			b.WriteString(" . ")
			b.WriteString(next.String())
		}
		b.WriteByte(')')
		return b.String()
	default:
		return "nil"
	}
}

var (
	SimpStore    = SxAtom("store")
	SimpSelect   = SxAtom("select")
	SimpTrue     = SxAtom("|@true|")
	SimpFalse    = SxAtom("|@false|")
	SimpForall   = SxAtom("FORALL")
	SimpExists   = SxAtom("EXISTS")
	SimpFEq      = SxAtom("F_EQ")
	SimpImplies  = SxAtom("IMPLIES")
	SimpFImplies = SxAtom("F_IMPLIES")
	SimpAnd      = SxAtom("AND")
	SimpFAnd     = SxAtom("F_AND")
	SimpOr       = SxAtom("OR")
	SimpFOr      = SxAtom("F_OR")
	SimpIff      = SxAtom("IFF")
	SimpEq       = SxAtom("EQ")
	SimpIn       = SxAtom("in")
	SimpEmpty    = SxAtom("EMPTY")
	SimpMkdom    = SxAtom("mkdom")
	SimpIntv     = SxAtom("intv")
	SimpCond     = SxAtom("cond")
	SimpCross2   = SxAtom("cross2")
	SimpBgPush   = SxAtom("BG_PUSH")
	SimpPats     = SxAtom("PATS")

	SimpOpDefns  = SxNil
	SimpFcnDefns = SxNil
	SimpThms     = SxNil

	simpTrop = map[string]string{
		"\\lnot": "F_NOT",
		"=":      "F_EQ",
		"/=":     "F_NEQ",
		"=>":     "F_IMPLIES",
		"\\land": "F_AND",
		"\\lor":  "F_OR",
		"\\in":   "in",
		"..":     "intv",
		"\\leq":  "F_LE",
		"\\geq":  "F_GE",
		">":      "F_GT",
	}
	SimpDefns = NewInsMap[string, *Sx]()
)

func SimpTransOp(name *UniqueString) *UniqueString {
	if name == nil {
		return nil
	}
	if op, ok := simpTrop[name.String()]; ok {
		return UniqueStringOf(op)
	}
	return name
}

func SimpMkForall(vars *Sx, body *Sx) *Sx {
	return SxList(SimpForall, vars, body)
}

func SimpMkForallPats(vars *Sx, pats *Sx, body *Sx) *Sx {
	return SxList(SimpForall, vars, pats, body)
}

func SimpMkImplies(p *Sx, q *Sx) *Sx {
	return SxList(SimpFImplies, p, q)
}

func SimpMkAnd(p *Sx, q *Sx) *Sx {
	return SxList(SimpFAnd, p, q)
}

func SimpMkEquality(p *Sx, q *Sx) *Sx {
	return SxList(SimpEq, p, q)
}

func SimpMkInterval(i *Sx, j *Sx) *Sx {
	return SxList(SimpIntv, i, j)
}

func SimpMkIn(x *Sx, s *Sx) *Sx {
	return SxList(SimpIn, x, s)
}

func SimpMkIff(x *Sx, s *Sx) *Sx {
	return SxList(SimpIff, x, s)
}

func SimpMkBgPush(x *Sx) *Sx {
	return SxList(SimpBgPush, x)
}
