package sany_tests

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/BelchDefTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestBelchDefTests_runTestCase(t *testing.T) {
	t.Skip("tla2sany wip")

	for _, tc := range []struct {
		name   string
		inputs []string
	}{
		{name: "Empty module"},
		{name: "Single definition", inputs: []string{"op == 0"}},
		{name: "Multiple definitions", inputs: []string{"op == 0", "op2 == 1"}},
		{name: "Named theorem", inputs: []string{"op == 0 THEOREM T == TRUE"}},
		{name: "Named theorem after submodule", inputs: []string{"op == 0 ---- MODULE Inner ---- ==== THEOREM T == TRUE"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, diags := tlago.ParseSanySyntax("Test.tla", wrapSANYTestModule(tc.inputs...))
			requireNoSANYDiagnostics(t, "parse", diags)
			for _, input := range tc.inputs {
				if strings.Contains(input, "==") && findSANYNodeByJavaKind(root, "N_OperatorDefinition") == nil {
					t.Fatalf("expected operator definition in parsed module:\n%s", input)
				}
			}
		})
	}
}
