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
