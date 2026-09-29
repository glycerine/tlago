package sany_tests

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/drivers/IllegalOperatorTest.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestIllegalOperatorTest_test(t *testing.T) {
	specPath := sanyTestVectorPath("test-model", "IllegalOperatorTest.tla")
	spec, parseDiags := tlago.LoadSanySpec(specPath, tlago.LoadOptions{})
	requireNoSANYDiagnostics(t, "parse", parseDiags)
	diags := tlago.CheckSpec(spec).Errors().Deduplicated()
	if got := len(diags); got != 1 {
		t.Fatalf("semantic error count = %d, want 1\n%s", got, diags.Error())
	}
	diag := diags[0]
	if diag.Pos.Line != 3 || diag.Pos.Column != 8 || diag.Pos.EndLine != 3 || diag.Pos.EndColumn != 11 {
		t.Fatalf("diagnostic location = %+v, want line 3, col 8 to line 3, col 11", diag.Pos)
	}
	for _, want := range []string{"Argument number 1", "operator 'D'", "should be a 1-parameter operator"} {
		if !strings.Contains(diag.String(), want) {
			t.Fatalf("diagnostic missing %q\n%s", want, diag.String())
		}
	}
}
