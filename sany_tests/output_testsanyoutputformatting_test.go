package sany_tests

import (
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/output/TestSanyOutputFormatting.java.
func TestTestSanyOutputFormatting_testPercentSignInMessageWithoutArguments(t *testing.T) {
	percentMessage := "Couldn't resolve infix operator symbol `%'."
	if got, want := sanyOutputLogMessage(percentMessage), percentMessage+sanyOutputLineSeparator(); got != want {
		t.Fatalf("logged output = %q, want %q", got, want)
	}
	doublePercentMessage := "Couldn't resolve infix operator symbol `%%'."
	if got, want := sanyOutputLogMessage(doublePercentMessage), doublePercentMessage+sanyOutputLineSeparator(); got != want {
		t.Fatalf("logged output = %q, want %q", got, want)
	}
}

func TestTestSanyOutputFormatting_testArgumentsAreStillInterpolated(t *testing.T) {
	var out strings.Builder
	log := tlago.NewSimpleSanyOutput(&out, tlago.SanyLogInfo)
	log.Log(tlago.SanyLogError, "Parsing module %s in file %s", "Test", "Test.tla")
	got := out.String()
	if want := "Parsing module Test in file Test.tla" + sanyOutputLineSeparator(); got != want {
		t.Fatalf("logged output = %q, want %q", got, want)
	}
}

func TestTestSanyOutputFormatting_testParseErrorMentioningPercentOperator(t *testing.T) {
	module := "---- MODULE Test ----\nop == % 1\n====\n"
	var out strings.Builder
	log := tlago.NewSimpleSanyOutput(&out, tlago.SanyLogInfo)
	_, diags := tlago.ParseSanySyntaxWithOutput("Test.tla", module, log)
	if !diags.HasErrors() {
		t.Fatal("parser.parse returned true, want false")
	}
	// Preserve the source assertion against the actual recorded parser output.
	if got, want := out.String(), "token \"%\""; !strings.Contains(got, want) {
		t.Fatalf("recorded parser output missing %q\n%s", want, got)
	}
}

func sanyOutputLogMessage(message string) string {
	var out strings.Builder
	log := tlago.NewSimpleSanyOutput(&out, tlago.SanyLogInfo)
	log.Log(tlago.SanyLogError, message)
	return out.String()
}

// Original assertions use System.lineSeparator().
func sanyOutputLineSeparator() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}
