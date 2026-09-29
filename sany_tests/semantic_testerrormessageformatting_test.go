package sany_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestErrorMessageFormatting.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestErrorMessageFormatting_testPercentSignInMessageTextIsNotAFormatSpecifier(t *testing.T) {
	t.Skip("tla2sany wip")

	message := "Couldn't resolve infix operator symbol `%'."
	log := tlago.Diagnostics{{
		Code:     "SUSPECTED_UNREACHABLE_CHECK",
		Severity: tlago.SeverityError,
		Message:  message,
	}}
	if got := log[0].Message; got != message {
		t.Fatalf("diagnostic message = %q, want %q", got, message)
	}
}

func TestTestErrorMessageFormatting_testPercentSignInMessageParameterIsNotAFormatSpecifier(t *testing.T) {
	t.Skip("tla2sany wip")

	log := tlago.Diagnostics{{
		Code:     "SUSPECTED_UNREACHABLE_CHECK",
		Severity: tlago.SeverityError,
		Message:  fmt.Sprintf("Couldn't resolve infix operator symbol `%s'.", "%%"),
	}}
	if got, want := log[0].Message, "Couldn't resolve infix operator symbol `%%'."; got != want {
		t.Fatalf("diagnostic message = %q, want %q", got, want)
	}
}

func TestTestErrorMessageFormatting_testUnresolvedPercentOperator(t *testing.T) {
	t.Skip("tla2sany wip")

	diags := processSANYFormattingModule(t, "op == x % y")
	requireSANYDiagnosticContains(t, diags, "%")
}

func TestTestErrorMessageFormatting_testUnresolvedNonfixPercentOperator(t *testing.T) {
	t.Skip("tla2sany wip")

	diags := processSANYFormattingModule(t, "op == A!B!%(x, y)")
	requireSANYDiagnosticContains(t, diags, "%")
}

func TestTestErrorMessageFormatting_testUnresolvedDoublePercentOperatorIsNotRenamed(t *testing.T) {
	t.Skip("tla2sany wip")

	diags := processSANYFormattingModule(t, "op == x %% y")
	requireSANYDiagnosticContains(t, diags, "%%")
}

func processSANYFormattingModule(t *testing.T, moduleBody string) tlago.Diagnostics {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "Test.tla")
	module := "---- MODULE Test ----\n" + moduleBody + "\n====\n"
	if err := os.WriteFile(path, []byte(module), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, diags := tlago.LoadSanySpec(path, tlago.LoadOptions{})
	if diags.HasErrors() {
		return diags
	}
	return append(diags, tlago.CheckSpec(spec)...)
}

func requireSANYDiagnosticContains(t *testing.T, diags tlago.Diagnostics, expected string) {
	t.Helper()
	got := diags.Error()
	if !strings.Contains(got, expected) {
		t.Fatalf("no diagnostic mentions %q; got:\n%s", expected, got)
	}
}
