package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/NestedModuleInstanceTest.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestNestedModuleInstanceTest_testTopLevelInstanceOfNestedModule(t *testing.T) {
	t.Skip("tla2sany wip")

	spec := checkedSANYSpecPath(t, sanyTestVectorPath("test-model", "sany", "NestedModuleTopLevelInstance.tla"))
	requireNoSANYDiagnostics(t, "semantic", tlago.CheckSpec(spec))
}

func TestNestedModuleInstanceTest_testLetInstanceOfNestedModule(t *testing.T) {
	t.Skip("tla2sany wip")

	spec := checkedSANYSpecPath(t, sanyTestVectorPath("test-model", "sany", "NestedModuleLetInstance.tla"))
	requireNoSANYDiagnostics(t, "semantic", tlago.CheckSpec(spec))
}

func checkedSANYSpecPath(t *testing.T, path string) *tlago.Spec {
	t.Helper()
	spec, diags := tlago.LoadSanySpec(path, tlago.LoadOptions{})
	requireNoSANYDiagnostics(t, "parse", diags)
	return spec
}
