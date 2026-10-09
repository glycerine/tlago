package tlc

import "testing"

func TestDefnsUsesExplicitDefinitionCount(t *testing.T) {
	UniqueStringInitialize()
	SetStateVariables([]string{"x", "y"})

	defns := NewDefns()
	defns.SetDefnCount(2)
	defns.Put("A", IntOne)

	if loc := UniqueStringOf("A").DefnLoc(); loc != 2 {
		t.Fatalf("definition loc = %d, want 2", loc)
	}
	if got := defns.Get("A"); got != IntOne {
		t.Fatalf("definition value = %v, want %v", got, IntOne)
	}
}

func TestDefnsSnapshotCopiesTable(t *testing.T) {
	UniqueStringInitialize()
	defns := NewDefns()
	defns.Put("A", IntOne)

	snapshot := defns.Snapshot()
	two := NewIntValue(2)
	defns.Put("A", two)

	if got := snapshot.Get("A"); got != IntOne {
		t.Fatalf("snapshot definition value = %v, want %v", got, IntOne)
	}
	if got := defns.Get("A"); got != two {
		t.Fatalf("definition value = %v, want %v", got, two)
	}
}

func TestSpecsGetLevelFollowsLazyValuesAndOperatorDefinitions(t *testing.T) {
	UniqueStringInitialize()
	param := NewSymbolNode("P")
	expr := NewNumeralNode(1)
	expr.AddLevelParams(param)
	expr.LevelCheckNext()

	lazyBody := NewNumeralNode(2)
	lazyBody.SetLevel(TLCLevelAction)
	lazyBody.LevelCheckNext()
	lazyCtx := EmptyContext.Cons(param, NewLazyValue(lazyBody, EmptyContext, true))
	if got := SpecsGetLevel(expr, lazyCtx); got != TLCLevelAction {
		t.Fatalf("lazy level = %d, want %d", got, TLCLevelAction)
	}

	op := NewOpDefNode("Op", nil, nil)
	op.SetLevel(TLCLevelTemporal)
	opCtx := EmptyContext.Cons(param, op)
	if got := SpecsGetLevel(expr, opCtx); got != TLCLevelTemporal {
		t.Fatalf("operator level = %d, want %d", got, TLCLevelTemporal)
	}
}

func TestSpecsAddSubstsWrapsInListOrder(t *testing.T) {
	UniqueStringInitialize()
	body := NewNumeralNode(1)
	a := NewSymbolNode("A")
	b := NewSymbolNode("B")
	subA := NewSubstInNode(nil, NewSubst(a, NewNumeralNode(2)))
	subB := NewSubstInNode(nil, NewSubst(b, NewNumeralNode(3)))

	got, ok := SpecsAddSubsts(body, NewList(subA, subB)).(*SubstInNode)
	if !ok {
		t.Fatalf("outer node = %T, want *SubstInNode", got)
	}
	if len(got.Substs) != 1 || got.Substs[0].Op != b {
		t.Fatalf("outer subst = %+v, want B", got.Substs)
	}
	inner, ok := got.Body.(*SubstInNode)
	if !ok {
		t.Fatalf("inner node = %T, want *SubstInNode", got.Body)
	}
	if len(inner.Substs) != 1 || inner.Substs[0].Op != a || inner.Body != body {
		t.Fatalf("inner subst/body = %+v/%p, want A/%p", inner.Substs, inner.Body, body)
	}
}
