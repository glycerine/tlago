package tlago

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original TestSubexpressionSelectors.process: syntax/dependency processing,
// semantic generation without level checking, and retained errors on abort.
func processOriginalSubexpressionSelector(t *testing.T, body string) (log Diagnostics) {
	t.Helper()
	module := "---- MODULE Test ----\n" + body + "\n====\n"
	spec, parse := ParseSanySpecSource("Test.tla", module, LoadOptions{})
	if parse.HasErrors() {
		t.Fatal(parse)
	}
	defer func() {
		if failure := recover(); failure != nil {
			if _, ok := failure.(*SanySemanticAbort); !ok {
				panic(failure)
			}
			log = spec.SemanticDiags
		}
	}()
	return GenerateSanySpec(spec)
}

func assertOriginalSelectorUserFacingError(t *testing.T, log Diagnostics) {
	t.Helper()
	if !log.HasErrors() {
		t.Fatal("Expected the input to be rejected")
	}
	for _, error := range log {
		if error.Severity == SeverityError && error.Code == "E4003" {
			t.Fatalf("internal error details must be empty: %v", log)
		}
	}
}

func assertOriginalSelectorFirstError(t *testing.T, log Diagnostics, message, location string) {
	t.Helper()
	var details []Diagnostic
	for _, detail := range log {
		if detail.Severity == SeverityError {
			details = append(details, detail)
		}
	}
	error := details[0]
	if error.Code != "E4200" {
		t.Fatalf("error code = %s, want SYMBOL_UNDEFINED", error.Code)
	}
	if error.SANYMessage != message {
		t.Fatalf("error message = %q, want %q", error.SANYMessage, message)
	}
	span := error.SANYRange
	actual := tlc.SourceLocation{Source: strings.TrimSuffix(filepath.Base(span.Begin.File), ".tla"), BeginLine: span.Begin.Line, BeginColumn: span.Begin.Column, EndLine: span.End.Line, EndColumn: span.End.Column}.String()
	if actual != location {
		t.Fatalf("error location = %q, want %q", actual, location)
	}
}

func TestTestSubexpressionSelectors_testUnresolvedCompoundOperatorName(t *testing.T) {
	log := processOriginalSubexpressionSelector(t, "use == module!op")
	assertOriginalSelectorUserFacingError(t, log)
	assertOriginalSelectorFirstError(t, log, "Unknown operator: `module!op'.", "line 2, col 8 to line 2, col 16 of module Test")
}
func TestTestSubexpressionSelectors_testConsecutiveTreeNavigationSelectors(t *testing.T) {
	log := processOriginalSubexpressionSelector(t, "tree_nav == op!<<!>>")
	assertOriginalSelectorUserFacingError(t, log)
	assertOriginalSelectorFirstError(t, log, "Unknown operator: `op'.", "line 2, col 13 to line 2, col 14 of module Test")
}
func TestTestSubexpressionSelectors_testAllTreeNavigationSelectors(t *testing.T) {
	log := processOriginalSubexpressionSelector(t, "tree_nav == op(a, b)!<<!>>!3!(x, y)!:!@")
	assertOriginalSelectorUserFacingError(t, log)
	assertOriginalSelectorFirstError(t, log, "Unknown operator: `op'.", "line 2, col 13 to line 2, col 14 of module Test")
}
