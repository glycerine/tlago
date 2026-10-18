package tlc

import "testing"

func TestMVPermBehaviorMatchesJavaModelValuePermutation(t *testing.T) {
	ModelValueInit()
	a := AddModelValue("A")
	b := AddModelValue("B")
	c := AddModelValue("C")

	perm := NewMVPerm()
	if perm.Size() != 0 {
		t.Fatalf("new permutation size = %d, want 0", perm.Size())
	}
	if got := a.Permute(perm); got != a {
		t.Fatalf("unmapped model value permuted to %v, want original %v", got, a)
	}

	perm.Put(a, a)
	if perm.Size() != 0 {
		t.Fatalf("identity mapping changed size to %d, want 0", perm.Size())
	}

	perm.Put(a, b)
	if perm.Size() != 1 {
		t.Fatalf("non-identity mapping size = %d, want 1", perm.Size())
	}
	if got := a.Permute(perm); got != b {
		t.Fatalf("A permuted to %v, want B", got)
	}
	if got := c.Permute(perm); got != c {
		t.Fatalf("unmapped C permuted to %v, want original C", got)
	}

	perm.Put(a, c)
	if got := a.Permute(perm); got != b {
		t.Fatalf("second Put overwrote A mapping to %v, want original B", got)
	}
	if got := perm.String(); got != "[A -> B]" {
		t.Fatalf("permutation string = %q, want %q", got, "[A -> B]")
	}
}

func TestMVPermComposeMatchesJavaArrayByIndexSemantics(t *testing.T) {
	ModelValueInit()
	a := AddModelValue("A")
	b := AddModelValue("B")
	c := AddModelValue("C")

	first := NewMVPerm()
	first.Put(a, b)
	first.Put(b, c)

	second := NewMVPerm()
	second.Put(b, a)
	second.Put(c, b)

	composed := first.Compose(second)
	if got := a.Permute(composed); got != a {
		t.Fatalf("A composed to %v, want A", got)
	}
	if got := b.Permute(composed); got != b {
		t.Fatalf("B composed to %v, want B", got)
	}
	if got := c.Permute(composed); got != b {
		t.Fatalf("C composed to %v, want B", got)
	}
	if composed.Size() != 1 {
		t.Fatalf("composed size = %d, want 1 after identity mappings are elided", composed.Size())
	}
}

func TestPermutationsMatchesJavaOrdering(t *testing.T) {
	ModelValueInit()
	a := AddModelValue("A")
	b := AddModelValue("B")
	c := AddModelValue("C")

	perms, err := Permutations(NewSetEnumValue([]Value{a, b, c}, true))
	if err != nil {
		t.Fatalf("Permutations returned error: %v", err)
	}
	if perms.Elems.Len() != 6 {
		t.Fatalf("Permutations size = %d, want 6", perms.Elems.Len())
	}

	got := make([]string, perms.Elems.Len())
	for i := range got {
		got[i] = perms.Elems.At(i).String()
	}
	want := []string{
		"(A :> A @@ B :> B @@ C :> C)",
		"(A :> A @@ B :> C @@ C :> B)",
		"(A :> B @@ B :> A @@ C :> C)",
		"(A :> B @@ B :> C @@ C :> A)",
		"(A :> C @@ B :> A @@ C :> B)",
		"(A :> C @@ B :> B @@ C :> A)",
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("permutation %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPermutationSubgroupMatchesJavaClosureBehavior(t *testing.T) {
	ModelValueInit()
	a := AddModelValue("A")
	b := AddModelValue("B")
	c := AddModelValue("C")

	swapAB := NewFcnRcdValue([]Value{a, b, c}, []Value{b, a, c}, true)
	swapBC := NewFcnRcdValue([]Value{a, b, c}, []Value{a, c, b}, true)
	generators := NewSetEnumValue([]Value{swapAB, swapBC}, true)

	subgroup, err := PermutationSubgroup(generators)
	if err != nil {
		t.Fatalf("PermutationSubgroup returned error: %v", err)
	}
	if len(subgroup) != 5 {
		t.Fatalf("subgroup size = %d, want 5 non-identity permutations", len(subgroup))
	}

	got := map[string]bool{}
	for _, perm := range subgroup {
		got[perm.String()] = true
	}
	for _, want := range []string{
		"[A -> B, B -> A]",
		"[B -> C, C -> B]",
		"[A -> B, B -> C, C -> A]",
		"[A -> C, B -> A, C -> B]",
		"[A -> C, C -> A]",
	} {
		if !got[want] {
			t.Fatalf("subgroup missing %s; got %v", want, got)
		}
	}
}

func TestPermutationSubgroupRejectsNonModelValuePairs(t *testing.T) {
	ModelValueInit()
	a := AddModelValue("A")
	fcn := NewFcnRcdValue([]Value{a}, []Value{NewIntValue(1)}, true)

	if _, err := PermutationSubgroup(NewSetEnumValue([]Value{fcn}, true)); err == nil {
		t.Fatal("PermutationSubgroup accepted a function whose range is not a model value")
	}
}
