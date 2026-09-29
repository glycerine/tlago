package sany_tests

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestSubexpressionSelectors.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestSubexpressionSelectors_testUnresolvedCompoundOperatorName(t *testing.T) {
	t.Skip("tla2sany wip")

	diags := processSANYSubexpressionSelectorBody(t, "use == module!op")
	requireSANYSubexpressionSelectorUserError(t, diags, "module!op")
}

func TestTestSubexpressionSelectors_testConsecutiveTreeNavigationSelectors(t *testing.T) {
	t.Skip("tla2sany wip")

	diags := processSANYSubexpressionSelectorBody(t, "tree_nav == op!<<!>>")
	requireSANYSubexpressionSelectorUserError(t, diags, "op")
}

func TestTestSubexpressionSelectors_testAllTreeNavigationSelectors(t *testing.T) {
	t.Skip("tla2sany wip")

	diags := processSANYSubexpressionSelectorBody(t, "tree_nav == op(a, b)!<<!>>!3!(x, y)!:!@")
	requireSANYSubexpressionSelectorUserError(t, diags, "op")
}

func processSANYSubexpressionSelectorBody(t *testing.T, body string) tlago.Diagnostics {
	t.Helper()
	_, diags := tlago.CheckSanySource("Test.tla", wrapSANYTestModule(body))
	return diags
}

func requireSANYSubexpressionSelectorUserError(t *testing.T, diags tlago.Diagnostics, unresolved string) {
	t.Helper()
	if !diags.HasErrors() {
		t.Fatalf("expected unresolved subexpression selector %q to fail", unresolved)
	}
	got := diags.Error()
	if strings.Contains(strings.ToLower(got), "internal") {
		t.Fatalf("expected user-facing error, got internal diagnostic:\n%s", got)
	}
	if !strings.Contains(got, unresolved) {
		t.Fatalf("subexpression selector diagnostic missing %q\n%s", unresolved, got)
	}
}
