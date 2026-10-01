package tlc

import "testing"

func TestSxPortedListPrintingAndAtomIdentity(t *testing.T) {
	a := SxAtom("a")
	if a != SxAtom("a") {
		t.Fatalf("SxAtom did not intern identical atoms")
	}
	list := SxList(a, SxFromInt(1), SxAtom("b"))
	if got := list.String(); got != "(a 1 b)" {
		t.Fatalf("list string = %q, want %q", got, "(a 1 b)")
	}
	dotted := SxCons(SxAtom("a"), SxAtom("b"))
	if got := dotted.String(); got != "(a . b)" {
		t.Fatalf("dotted pair string = %q, want %q", got, "(a . b)")
	}
	if !SxMemq(a, list) {
		t.Fatalf("SxMemq did not find interned atom")
	}
}

func TestSimpUtilPortedBuilders(t *testing.T) {
	if got := SimpTransOp(UniqueStringOf("\\land")).String(); got != "F_AND" {
		t.Fatalf("SimpTransOp(land) = %q, want F_AND", got)
	}
	if got := SimpTransOp(OpCP).String(); got != "cross2" {
		t.Fatalf("SimpTransOp(cartesian product) = %q, want cross2", got)
	}
	expr := SimpMkImplies(SxAtom("p"), SxAtom("q"))
	if got := expr.String(); got != "(F_IMPLIES p q)" {
		t.Fatalf("SimpMkImplies = %q, want (F_IMPLIES p q)", got)
	}
}
