package sany_tests

import (
	"reflect"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestErrors.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestErrors_testWarningMessages(t *testing.T) {
	t.Skip("tla2sany wip")

	loc1 := genSANYErrorLocation(1)
	warn1 := tlago.Diagnostic{Code: "W4800", Severity: tlago.SeverityWarning, Pos: loc1, Message: "This is a test warning message"}
	loc2 := genSANYErrorLocation(2)
	warn2 := tlago.Diagnostic{Code: "W4801", Severity: tlago.SeverityWarning, Pos: loc2, Message: "This is another test warning message"}
	loc3 := tlago.Position{}
	warn3 := tlago.Diagnostic{Code: "W4800", Severity: tlago.SeverityWarning, Pos: loc3, Message: "This is yet another test warning message"}
	log := tlago.Diagnostics{warn1, warn2, warn3}

	if got, want := log.Warnings(), (tlago.Diagnostics{warn1, warn2, warn3}); !reflect.DeepEqual(got, want) {
		t.Fatalf("warnings = %#v, want %#v", got, want)
	}
	if log.HasErrors() {
		t.Fatalf("warnings-only log should not be failure:\n%s", log.Error())
	}
	if !log.IsSuccess() {
		t.Fatal("warnings-only log should be successful")
	}
	if got, want := len(log), 3; got != want {
		t.Fatalf("message count = %d, want %d", got, want)
	}
	if got := len(log.Errors()); got != 0 {
		t.Fatalf("error count = %d, want 0", got)
	}
	for _, diag := range log {
		if !log.ContainsMessage(diag.Message) {
			t.Fatalf("summary missing %q\n%s", diag.Message, log.Error())
		}
	}
}

func TestTestErrors_testErrorMessages(t *testing.T) {
	t.Skip("tla2sany wip")

	loc1 := genSANYErrorLocation(4)
	err1 := tlago.Diagnostic{Code: "E1300", Severity: tlago.SeverityError, Pos: loc1, Message: "This is a test error message"}
	loc2 := genSANYErrorLocation(5)
	err2 := tlago.Diagnostic{Code: "E1300", Severity: tlago.SeverityError, Pos: loc2, Message: "This is another test error message"}
	loc3 := tlago.Position{}
	err3 := tlago.Diagnostic{Code: "E1300", Severity: tlago.SeverityError, Pos: loc3, Message: "This is yet another test error message"}
	log := tlago.Diagnostics{err1, err2, err3}

	if got, want := log.Errors(), (tlago.Diagnostics{err1, err2, err3}); !reflect.DeepEqual(got, want) {
		t.Fatalf("errors = %#v, want %#v", got, want)
	}
	if !log.HasErrors() {
		t.Fatal("error log should be failure")
	}
	if log.IsSuccess() {
		t.Fatal("error log should not be successful")
	}
	if got, want := len(log), 3; got != want {
		t.Fatalf("message count = %d, want %d", got, want)
	}
	if got := len(log.Warnings()); got != 0 {
		t.Fatalf("warning count = %d, want 0", got)
	}
	for _, diag := range log {
		if !log.ContainsMessage(diag.Message) {
			t.Fatalf("summary missing %q\n%s", diag.Message, log.Error())
		}
	}
}

func TestTestErrors_testMixedMessageLevels(t *testing.T) {
	t.Skip("tla2sany wip")

	loc1 := genSANYErrorLocation(7)
	warn := tlago.Diagnostic{Code: "W4800", Severity: tlago.SeverityWarning, Pos: loc1, Message: "This is a test warning message"}
	loc2 := genSANYErrorLocation(8)
	err := tlago.Diagnostic{Code: "E1300", Severity: tlago.SeverityError, Pos: loc2, Message: "This is a test error message"}
	log := tlago.Diagnostics{warn, err}

	if got, want := log.Warnings(), (tlago.Diagnostics{warn}); !reflect.DeepEqual(got, want) {
		t.Fatalf("warnings = %#v, want %#v", got, want)
	}
	if got, want := log.Errors(), (tlago.Diagnostics{err}); !reflect.DeepEqual(got, want) {
		t.Fatalf("errors = %#v, want %#v", got, want)
	}
	if !log.HasErrors() || log.IsSuccess() {
		t.Fatalf("mixed log should be failure:\n%s", log.Error())
	}
	for _, diag := range log {
		if !log.ContainsMessage(diag.Message) {
			t.Fatalf("summary missing %q\n%s", diag.Message, log.Error())
		}
	}
}

func TestTestErrors_testDuplicateErrorsIgnored(t *testing.T) {
	t.Skip("tla2sany wip")

	loc1 := genSANYErrorLocation(9)
	warn := tlago.Diagnostic{Code: "W4800", Severity: tlago.SeverityWarning, Pos: loc1, Message: "This is a test warning message"}
	loc2 := genSANYErrorLocation(10)
	err := tlago.Diagnostic{Code: "E1300", Severity: tlago.SeverityError, Pos: loc2, Message: "This is a test error message"}
	log := (tlago.Diagnostics{warn, warn, warn, err, err, err}).Deduplicated()

	if got, want := log.Warnings(), (tlago.Diagnostics{warn}); !reflect.DeepEqual(got, want) {
		t.Fatalf("deduplicated warnings = %#v, want %#v", got, want)
	}
	if got, want := log.Errors(), (tlago.Diagnostics{err}); !reflect.DeepEqual(got, want) {
		t.Fatalf("deduplicated errors = %#v, want %#v", got, want)
	}
	if !log.HasErrors() || log.IsSuccess() {
		t.Fatalf("deduplicated mixed log should be failure:\n%s", log.Error())
	}
	if got, want := len(log), 2; got != want {
		t.Fatalf("deduplicated message count = %d, want %d", got, want)
	}
}

func genSANYErrorLocation(seed int) tlago.Position {
	return tlago.Position{
		File:      "Test.tla",
		Line:      seed * 3,
		Column:    seed * 5,
		EndLine:   seed * 7,
		EndColumn: seed * 11,
	}
}
