package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestLevelChecking.java.
// Retains the complete original parameter matrix and separate phase assertions.
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
			spec, parseDiags := tlago.ParseSanySpecSource("Test.tla", source, tlago.LoadOptions{})
			if parseDiags.HasErrors() {
				t.Fatalf("Input:\n%s\nLog:\n%s", source, parseDiags.Error())
			}
			semanticLog := tlago.GenerateSanySpec(spec)
			if semanticLog.HasErrors() {
				t.Fatalf("Input:\n%s\nLog:\n%s", source, semanticLog.Error())
			}
			actualLevelCheckingResult, levelCheckingLog := tlago.CheckSanySpecLevels(spec)
			if semanticLog.HasErrors() {
				t.Fatalf("Input:\n%s\nLog:\n%s", source, semanticLog.Error())
			}
			if logSuccess := !levelCheckingLog.HasErrors(); logSuccess != actualLevelCheckingResult {
				t.Fatalf("Input:\n%s\nLog:\n%s\nlog success = %v, level result = %v", source, levelCheckingLog.Error(), logSuccess, actualLevelCheckingResult)
			}
			if tc.ok != actualLevelCheckingResult {
				t.Fatalf("Input:\n%s\nLog:\n%s\nlevel result = %v, want %v", source, levelCheckingLog.Error(), actualLevelCheckingResult, tc.ok)
			}
		})
	}
}
