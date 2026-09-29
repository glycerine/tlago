package sany_tests

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago"
)

func sanyTestVectorPath(parts ...string) string {
	all := append([]string{"test_vectors"}, parts...)
	return filepath.Join(all...)
}

func requireNoSANYDiagnostics(t *testing.T, phase string, diags tlago.Diagnostics) {
	t.Helper()
	if diags.HasErrors() {
		t.Fatalf("%s diagnostics:\n%s", phase, diags.Error())
	}
}
