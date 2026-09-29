package sany_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestContext.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestContext_testDifferentSymbolClassesDiagnostic(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "Base.tla")
	if err := os.WriteFile(basePath, []byte("---- MODULE Base ----\nCONSTANT symbol\n====\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Join(dir, "Test.tla")
	if err := os.WriteFile(rootPath, []byte("---- MODULE Test ----\nEXTENDS Base\nsymbol == TRUE\n====\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, diags := tlago.LoadSanySpec(rootPath, tlago.LoadOptions{})
	requireNoSANYDiagnostics(t, "parse", diags)
	sem := tlago.CheckSpec(spec)
	if !sem.HasErrors() {
		t.Fatal("expected conflict between imported declaration and local definition")
	}
	got := sem.Error()
	for _, want := range []string{"symbol", "CONSTANT", "OPERATOR"} {
		if !strings.Contains(got, want) {
			t.Fatalf("semantic diagnostic missing %q\n%s", want, got)
		}
	}
}
