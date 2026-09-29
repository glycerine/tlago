package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestLevelChecking.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestLevelChecking_testAll(t *testing.T) {
	for _, tc := range []struct {
		expr string
		ok   bool
	}{
		{expr: `\neg c`, ok: true},
		{expr: "c.foo", ok: true},
		{expr: "([]c).foo", ok: false},
		{expr: "[]c", ok: true},
		{expr: "□c", ok: true},
		{expr: "<>c", ok: true},
		{expr: "◇c", ok: true},
		{expr: "[]x", ok: true},
		{expr: "□x", ok: true},
		{expr: "<>x", ok: true},
		{expr: "◇x", ok: true},
		{expr: "[](x')", ok: false},
		{expr: "□(x')", ok: false},
		{expr: "<>(x')", ok: false},
		{expr: "◇(x')", ok: false},
		{expr: "[][c]_x", ok: true},
		{expr: "□[c]_x", ok: true},
		{expr: "[]<<c>>_x", ok: false},
		{expr: "□⟨c⟩_x", ok: false},
		{expr: "<><<c>>_x", ok: true},
		{expr: "◇⟨c⟩_x", ok: true},
		{expr: "<>[c]_x", ok: false},
		{expr: "◇[c]_x", ok: false},
		{expr: "c ~> c", ok: true},
		{expr: "x' ~> c", ok: false},
		{expr: "c ~> x'", ok: false},
		{expr: "c ↝ c", ok: true},
		{expr: "x' ↝ c", ok: false},
		{expr: "c ↝ x'", ok: false},
		{expr: "c -+-> c", ok: true},
		{expr: "x' -+-> c", ok: false},
		{expr: "c -+-> x'", ok: false},
		{expr: "c ⇸ c", ok: true},
		{expr: "x' ⇸ c", ok: false},
		{expr: "c ⇸ x'", ok: false},
		{expr: `c /\ c`, ok: true},
		{expr: "c ∧ c", ok: true},
		{expr: `x /\ c`, ok: true},
		{expr: "x ∧ c", ok: true},
		{expr: `[]c /\ c`, ok: true},
		{expr: "□c ∧ c", ok: true},
		{expr: `c /\ x'`, ok: true},
		{expr: "c ∧ x'", ok: true},
		{expr: `[]c /\ x'`, ok: false},
		{expr: "□c ∧ x'", ok: false},
		{expr: "∀ v ∈ c  : v", ok: true},
		{expr: "∃ v ∈ x  : v", ok: true},
		{expr: "∀ v ∈ x' : v", ok: true},
		{expr: "∃ v ∈ x  : □v", ok: true},
		{expr: "∀ v ∈ x' : □v", ok: false},
		{expr: "∃ v ∈ □x : v", ok: false},
	} {
		t.Run(tc.expr, func(t *testing.T) {
			source := "---- MODULE Test ----\nCONSTANT c\nVARIABLE x\nop == " + tc.expr + "\n===="
			_, diags := tlago.CheckSanySource("Test.tla", source)
			if tc.ok && diags.HasErrors() {
				t.Fatalf("level check failed unexpectedly:\n%s", diags.Error())
			}
			if !tc.ok && !diags.HasErrors() {
				t.Fatal("level check succeeded unexpectedly")
			}
		})
	}
}
