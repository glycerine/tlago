package tlago

import (
	"fmt"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaGetScopedIdentifiers(t *testing.T) {
	const wrapper = "---- MODULE Test ----\nVARIABLE x\nCONSTANT y\n%s\nTRUE%s\n===="
	location := tlc.NewSourceLocation("Test", 5, 1, 5, 2)
	cases := []struct {
		input, terminator string
		expected          []string
	}{
		{"op ≜", "", []string{}},
		{"op(i, j, k) ≜", "", []string{"j", "i", "k"}},
		{"op(i) ≜ ∀ j ∈ {} :", "", []string{"i", "j"}},
		{"op ≜ ∀ i, j ∈ {} :", "", []string{"i", "j"}},
		{"op ≜ ∀ ⟨i, j, k⟩ ∈ {} :", "", []string{"i", "j", "k"}},
		{"op(i, j) ≜ LET k == TRUE l == TRUE IN", "", []string{"i", "j", "k", "l"}},
		{"op ≜ LET i(j, k) == ", " IN TRUE", []string{"i(_,_)", "j", "k"}},
		{"op ≜ LET RECURSIVE i i == TRUE IN", "", []string{"i"}},
		{"op ≜ LET i == TRUE j ==", " IN TRUE", []string{"i", "j"}},
		{"op(i) ≜ [j, k ∈ {} ↦", "]", []string{"i", "j", "k"}},
		{"op ≜ [⟨i, j, k⟩ ∈ {} ↦", "]", []string{"i", "j", "k"}},
		{"op(i, j) ≜ {k ∈ {} :", "}", []string{"i", "j", "k"}},
		{"RECURSIVE op(_, _) op(f(_,_), i) ≜ op(LAMBDA j, k : ", ", TRUE)", []string{"f(_,_)", "i", "j", "k"}},
		{"op ≜ {⟨i, j, k⟩ ∈ {} : ", "}", []string{"i", "j", "k"}},
		{"op(i) ≜ {j ∈ {} : ", "}", []string{"i", "j"}},
		{"op ≜ {", ": i, j ∈ {}}", []string{"i", "j"}},
		{"op ≜ {", ": ⟨i, j, k⟩ ∈ {}}", []string{"i", "j", "k"}},
		// Preserve the upstream known infix-parameter bug and its expected signature.
		{"op(_+_) ≜", "", []string{"+(_,_)"}},
	}
	for index, test := range cases {
		t.Run(fmt.Sprintf("%d: %s TRUE %s", index, test.input, test.terminator), func(t *testing.T) {
			input := fmt.Sprintf(wrapper, test.input, test.terminator)
			spec, log := CheckSanySource("Test.tla", input)
			// The production SANY bridge constructs the TLC semantic graph used by the
			// debugger helper. This frontend test does not start a model checker.
			tool, bridgeLog := BuildTLCTool(spec, tlc.NewModelConfig("Test"), tlc.RuntimeParameters{})
			log = append(log, bridgeLog...)
			if tool == nil || tool.GetSpecProcessor().GetModuleTbl() == nil {
				t.Fatal("module table is nil")
			}
			modules := tool.GetSpecProcessor().GetModuleTbl()
			if log.HasErrors() {
				t.Fatalf("frontend failed: %v", log)
			}
			debuggerAssertEqual(t, 1, len(modules.GetRootModule().GetOpDefs()))
			actual := tlc.GetScopedIdentifiers(modules.GetRootModule(), location)
			actualSet := map[string]struct{}{}
			for name := range actual.All() {
				actualSet[name] = struct{}{}
			}
			expectedSet := map[string]struct{}{}
			for _, name := range test.expected {
				expectedSet[name] = struct{}{}
			}
			debuggerAssertEqual(t, expectedSet, actualSet)
		})
	}
}
