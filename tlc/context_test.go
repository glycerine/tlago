package tlc

import "testing"

func TestContextLookupEmptyAndBranch(t *testing.T) {
	UniqueStringInitialize()
	name := NewSymbolNode("missing")
	if got := EmptyContext.Lookup(name); got != nil {
		t.Fatalf("empty Lookup = %v, want nil", got)
	}
	if got := EmptyContext.LookupCutoff(name, true); got != nil {
		t.Fatalf("empty LookupCutoff(true) = %v, want nil", got)
	}
	if got := EmptyContext.LookupCutoff(name, false); got != nil {
		t.Fatalf("empty LookupCutoff(false) = %v, want nil", got)
	}
	branch := BranchContext(EmptyContext)
	if got := branch.Lookup(name); got != nil {
		t.Fatalf("branch Lookup = %v, want nil", got)
	}
	if got := branch.LookupCutoff(name, true); got != nil {
		t.Fatalf("branch LookupCutoff(true) = %v, want nil", got)
	}
	if got := branch.LookupCutoff(name, false); got != nil {
		t.Fatalf("branch LookupCutoff(false) = %v, want nil", got)
	}
}

func TestContextLookupUsesSymbolIdentityAcrossBranch(t *testing.T) {
	UniqueStringInitialize()
	name := NewSymbolNode("ctx1")
	sameTextDifferentNode := NewSymbolNode("ctx1")
	value := "value1"

	ctx1 := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx1)
	ctx2 := branch.Cons(NewSymbolNode("ctx2"), "value2")
	ctx3 := ctx2.Cons(NewSymbolNode("ctx3"), "value3")

	if got := ctx3.Lookup(name); got != value {
		t.Fatalf("Lookup(name) = %v, want %v", got, value)
	}
	if got := ctx3.Lookup(sameTextDifferentNode); got != nil {
		t.Fatalf("Lookup(same text different node) = %v, want nil", got)
	}
	if got := ctx3.LookupCutoff(name, false); got != value {
		t.Fatalf("LookupCutoff(false) = %v, want %v", got, value)
	}
	if got := ctx3.LookupCutoff(name, true); got != nil {
		t.Fatalf("LookupCutoff(true) = %v, want nil after branch cutoff", got)
	}
}

func TestContextLookupAtBranchHonorsCutoff(t *testing.T) {
	UniqueStringInitialize()
	name := NewSymbolNode("ctx1")
	value := "value1"

	ctx := EmptyContext.Cons(name, value)
	branch := BranchContext(ctx)

	if got := branch.Lookup(name); got != value {
		t.Fatalf("branch Lookup = %v, want %v", got, value)
	}
	if got := branch.LookupCutoff(name, false); got != value {
		t.Fatalf("branch LookupCutoff(false) = %v, want %v", got, value)
	}
	if got := branch.LookupCutoff(name, true); got != nil {
		t.Fatalf("branch LookupCutoff(true) = %v, want nil", got)
	}
}

func TestContextDeepEmptyDepthAndCopy(t *testing.T) {
	UniqueStringInitialize()
	if !EmptyContext.IsEmpty() || !EmptyContext.IsDeepEmpty() {
		t.Fatalf("empty context should be empty and deep-empty")
	}
	branch := BranchContext(EmptyContext)
	if branch.IsEmpty() || !branch.IsDeepEmpty() {
		t.Fatalf("base branch should be deep-empty but not empty")
	}

	name := NewSymbolNode("ctx")
	ctx := branch.Cons(name, "value")
	if ctx.IsDeepEmpty() {
		t.Fatalf("non-empty context reported deep-empty")
	}
	if ctx.Depth() != 2 {
		t.Fatalf("context depth = %d, want 2", ctx.Depth())
	}
	copy := ctx.DeepCopy()
	if copy == ctx || copy.Next() == ctx.Next() {
		t.Fatalf("DeepCopy reused context nodes")
	}
	if got := copy.Lookup(name); got != "value" {
		t.Fatalf("copy Lookup = %v, want value", got)
	}
}
