package sany_tests

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestInstanceNode.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestInstanceNode_testOperatorArgumentMinimumLevelDiagnostic(t *testing.T) {
	module := "---- MODULE Test ----\n" +
		"---- MODULE Inner ----\n" +
		"CONSTANT F(_, _)\n" +
		"op == F([]TRUE, 0)\n" +
		"====\n" +
		"INSTANCE Inner WITH F <- =\n" +
		"====\n"
	_, diags := tlago.CheckSanySource("Test.tla", module)
	if !diags.HasErrors() {
		t.Fatal("expected INSTANCE substitution level diagnostic")
	}
	got := diags.Error()
	for _, want := range []string{"F", "1", "level"} {
		if !strings.Contains(got, want) {
			t.Fatalf("INSTANCE substitution diagnostic missing %q\n%s", want, got)
		}
	}
}
