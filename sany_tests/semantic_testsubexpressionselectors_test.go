package sany_tests

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Supplementary native API checks. The complete original diagnostic assertions
// and generation-only helper live in root sany_subexpression_selectors_java_test.go.
func TestSubexpressionSelectorNativeAPI_testUnresolvedCompoundOperatorName(t *testing.T) {
	diags := processSANYSubexpressionSelectorBody(t, "use == module!op")
	requireSANYSubexpressionSelectorUserError(t, diags, "module!op")
}

func TestSubexpressionSelectorNativeAPI_testConsecutiveTreeNavigationSelectors(t *testing.T) {
	diags := processSANYSubexpressionSelectorBody(t, "tree_nav == op!<<!>>")
	requireSANYSubexpressionSelectorUserError(t, diags, "op")
}

func TestSubexpressionSelectorNativeAPI_testAllTreeNavigationSelectors(t *testing.T) {
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
