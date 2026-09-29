package sany_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/ParseErrorTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestParseErrorTests_testAll(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "SanyTest.tla")
	input := "---- MODULE SanyTest ----\nx = 0\n===="
	if err := os.WriteFile(modulePath, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}

	_, diags := tlago.LoadSanySpec(modulePath, tlago.LoadOptions{})
	if !diags.HasErrors() {
		t.Fatalf("expected parse failure for input:\n%s", input)
	}
	want := "Was expecting \"==== or more Module body\"\nEncountered \"x\" at line"
	if got := diags.Error(); !strings.Contains(got, want) {
		t.Fatalf("parse diagnostics missing %q\n%s", want, got)
	}
}
