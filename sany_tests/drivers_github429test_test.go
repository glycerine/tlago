package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/drivers/Github429Test.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestGithub429Test_testForFailedParse(t *testing.T) {
	spec, diags := tlago.LoadSanySpec(sanyTestVectorPath("test-model", "Github429.tla"), tlago.LoadOptions{})
	requireNoSANYDiagnostics(t, "parse", diags)
	requireNoSANYDiagnostics(t, "semantic", tlago.CheckSpec(spec))
}
