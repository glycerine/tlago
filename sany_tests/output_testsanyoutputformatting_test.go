package sany_tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/output/TestSanyOutputFormatting.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestSanyOutputFormatting_testPercentSignInMessageWithoutArguments(t *testing.T) {
	t.Skip("tla2sany wip")

	percentMessage := "Couldn't resolve infix operator symbol `%'."
	if got, want := sanyOutputLogMessage(percentMessage), percentMessage+"\n"; got != want {
		t.Fatalf("logged output = %q, want %q", got, want)
	}
	doublePercentMessage := "Couldn't resolve infix operator symbol `%%'."
	if got, want := sanyOutputLogMessage(doublePercentMessage), doublePercentMessage+"\n"; got != want {
		t.Fatalf("logged output = %q, want %q", got, want)
	}
}

func TestTestSanyOutputFormatting_testArgumentsAreStillInterpolated(t *testing.T) {
	t.Skip("tla2sany wip")

	got := sanyOutputLogf("Parsing module %s in file %s", "Test", "Test.tla")
	if want := "Parsing module Test in file Test.tla\n"; got != want {
		t.Fatalf("logged output = %q, want %q", got, want)
	}
}

func TestTestSanyOutputFormatting_testParseErrorMentioningPercentOperator(t *testing.T) {
	t.Skip("tla2sany wip")

	module := "---- MODULE Test ----\nop == % 1\n====\n"
	_, diags := tlago.ParseSanySyntax("Test.tla", module)
	if !diags.HasErrors() {
		t.Fatalf("parse succeeded; want diagnostics mentioning %%")
	}
	if got, want := diags.Error(), "%"; !strings.Contains(got, want) {
		t.Fatalf("parse diagnostics missing %q\n%s", want, got)
	}
}

func sanyOutputLogMessage(message string) string {
	return message + "\n"
}

func sanyOutputLogf(message string, args ...any) string {
	return fmt.Sprintf(message, args...) + "\n"
}
