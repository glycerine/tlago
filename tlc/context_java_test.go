package tlc

import "testing"

func requireJavaContextLookup(t *testing.T, actual, expected any) {
	t.Helper()
	if actual != expected {
		t.Fatalf("Context.lookup = %v, want %v", actual, expected)
	}
}

func TestJavaContextLookupEmpty(t *testing.T) {

	requireJavaContextLookup(t, EmptyContext.Lookup(NewSymbolNode("Dummy")), nil)
	requireJavaContextLookup(t, EmptyContext.LookupCutoff(NewSymbolNode("Dummy"), true), nil)
	requireJavaContextLookup(t, EmptyContext.LookupCutoff(NewSymbolNode("Dummy"), false), nil)
	requireJavaContextLookup(t, EmptyContext.Lookup(nil), nil)
	requireJavaContextLookup(t, EmptyContext.LookupCutoff(nil, true), nil)
	requireJavaContextLookup(t, EmptyContext.LookupCutoff(nil, false), nil)
}

func TestJavaContextLookupBranch(t *testing.T) {

	ctx := BranchContext(EmptyContext)
	requireJavaContextLookup(t, ctx.Lookup(NewSymbolNode("Dummy")), nil)
	requireJavaContextLookup(t, ctx.LookupCutoff(NewSymbolNode("Dummy"), true), nil)
	requireJavaContextLookup(t, ctx.LookupCutoff(NewSymbolNode("Dummy"), false), nil)
	requireJavaContextLookup(t, ctx.Lookup(nil), nil)
	requireJavaContextLookup(t, ctx.LookupCutoff(nil, false), nil)
	requireJavaContextLookup(t, ctx.LookupCutoff(nil, true), nil)
}

func TestJavaContextLookupSymbolNodeNull(t *testing.T) {

	ctx := BranchContext(EmptyContext)
	requireJavaContextLookup(t, ctx.Lookup(nil), nil)
}

func TestJavaContextLookup(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)
	ctx2 := branch.Cons(NewSymbolNode("ctx2"), "value2")
	ctx3 := ctx2.Cons(NewSymbolNode("ctx3"), "value3")

	requireJavaContextLookup(t, ctx3.Lookup(name), value)
}

func TestJavaContextLookupCutOffFalse(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)
	ctx2 := branch.Cons(NewSymbolNode("ctx2"), "value2")
	ctx3 := ctx2.Cons(NewSymbolNode("ctx3"), "value3")

	requireJavaContextLookup(t, ctx3.LookupCutoff(name, false), value)
}

func TestJavaContextLookupCutOffTrue(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)
	ctx2 := branch.Cons(NewSymbolNode("ctx2"), "value2")
	ctx3 := ctx2.Cons(NewSymbolNode("ctx3"), "value3")

	requireJavaContextLookup(t, ctx3.LookupCutoff(name, true), nil)
}

func TestJavaContextLookupWithAtBranching(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)

	requireJavaContextLookup(t, branch.Lookup(name), value)
}

func TestJavaContextLookupWithCutOffFalseAtBranching(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)

	requireJavaContextLookup(t, branch.LookupCutoff(name, false), value)
}

func TestJavaContextLookupWithCutOffTrueAtBranching(t *testing.T) {

	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)

	requireJavaContextLookup(t, branch.LookupCutoff(name, true), nil)
}

func TestJavaContextLookupSymbolNode(t *testing.T) {

	name := NewSymbolNode("Dummy")
	value := new(int)

	ctx := BranchContext(EmptyContext)
	cons := ctx.Cons(name, value)

	lookup := cons.Lookup(name)
	requireJavaContextLookup(t, lookup, value)
}
