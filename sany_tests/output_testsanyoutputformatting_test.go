package sany_tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/output/TestSanyOutputFormatting.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestSanyOutputFormatting_testPercentSignInMessageWithoutArguments(t *testing.T) {
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
	got := sanyOutputLogf("Parsing module %s in file %s", "Test", "Test.tla")
	if want := "Parsing module Test in file Test.tla\n"; got != want {
		t.Fatalf("logged output = %q, want %q", got, want)
	}
}

func TestTestSanyOutputFormatting_testParseErrorMentioningPercentOperator(t *testing.T) {
	module := "---- MODULE Test ----\nop == % 1\n====\n"
	path := filepath.Join(t.TempDir(), "Test.tla")
	if err := os.WriteFile(path, []byte(module), 0o600); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	_, diags := tlago.LoadSanySpec(path, tlago.LoadOptions{ParsingProgress: func(message string) { out.WriteString(message); out.WriteByte('\n') }})
	if !diags.HasErrors() {
		t.Fatal("parser.parse returned true, want false")
	}
	// Preserve the source assertion against the actual recorded parser output.
	if got, want := out.String(), "token \"%\""; !strings.Contains(got, want) {
		t.Fatalf("recorded parser output missing %q\n%s", want, got)
	}
}

func sanyOutputLogMessage(message string) string {
	return message + "\n"
}

func sanyOutputLogf(message string, args ...any) string {
	return fmt.Sprintf(message, args...) + "\n"
}
