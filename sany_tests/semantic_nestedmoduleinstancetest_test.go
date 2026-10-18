package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/NestedModuleInstanceTest.java.
func TestNestedModuleInstanceTest_testTopLevelInstanceOfNestedModule(t *testing.T) {
	spec := checkedSANYSpecPath(t, sanyTestVectorPath("test-model", "sany", "NestedModuleTopLevelInstance.tla"))
	requireNoSANYDiagnostics(t, "semantic", tlago.CheckSpec(spec))
}

func TestNestedModuleInstanceTest_testLetInstanceOfNestedModule(t *testing.T) {
	// Preserve the original Java @Ignore reason and status.
	t.Skip("A nested module instantiated only from inside a LET is reported as multiply-defined")
	spec := checkedSANYSpecPath(t, sanyTestVectorPath("test-model", "sany", "NestedModuleLetInstance.tla"))
	requireNoSANYDiagnostics(t, "semantic", tlago.CheckSpec(spec))
}

func checkedSANYSpecPath(t *testing.T, path string) *tlago.Spec {
	t.Helper()
	spec, diags := tlago.LoadSanySpec(path, tlago.LoadOptions{})
	requireNoSANYDiagnostics(t, "parse", diags)
	return spec
}
