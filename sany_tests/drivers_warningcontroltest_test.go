package sany_tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/drivers/WarningControlTest.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestWarningControlTest_testWarningAppearsWithDefaultSettings(t *testing.T) {
	t.Skip("tla2sany wip")

	diags := warningControlSemanticDiagnostics(t)
	if diags.HasErrors() {
		t.Fatalf("default warning-control spec should not fail:\n%s", diags.Error())
	}
	if !diags.ContainsMessage("Foo") && !strings.Contains(diags.Error(), "W4802") {
		t.Fatalf("expected field-name-clash warning:\n%s", diags.Error())
	}
}

func TestWarningControlTest_testSuppressMessagesViaSettings(t *testing.T) {
	t.Skip("tla2sany wip")

	diags := warningControlSemanticDiagnostics(t)
	filtered := tlago.DiagnosticOptions{
		SuppressedCodes: map[string]bool{"W4802": true},
	}.Apply(diags)
	if len(filtered.Warnings()) != 0 {
		t.Fatalf("expected suppressed warning to disappear:\n%s", filtered.Error())
	}
	if filtered.HasErrors() {
		t.Fatalf("suppressed warning should not fail:\n%s", filtered.Error())
	}
}

func TestWarningControlTest_testMessagesAsErrorsViaSettings(t *testing.T) {
	t.Skip("tla2sany wip")

	diags := warningControlSemanticDiagnostics(t)
	filtered := tlago.DiagnosticOptions{
		ElevatedCodes: map[string]bool{"W4802": true},
	}.Apply(diags)
	if !filtered.HasErrors() {
		t.Fatalf("expected elevated warning to become error:\n%s", filtered.Error())
	}
	if !filtered.ContainsMessage("Warning treated as error") {
		t.Fatalf("expected elevated warning message:\n%s", filtered.Error())
	}
}

func TestWarningControlTest_testCLISuppressMessagesSilencesWarning(t *testing.T) {
	t.Skip("tla2sany wip")

	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "4802", warningControlSpecPath()}, nil, &stderr)
	if code != tlago.ExitOK {
		t.Fatalf("check suppress exit = %d, want %d; stderr=%s", code, tlago.ExitOK, stderr.String())
	}
	if strings.Contains(stderr.String(), "W4802") || strings.Contains(stderr.String(), "field name") {
		t.Fatalf("suppressed warning stderr = %q, want no field-name warning", stderr.String())
	}
}

func TestWarningControlTest_testCLIMessagesAsErrorsCausesFailure(t *testing.T) {
	t.Skip("tla2sany wip")

	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-messagesAsErrors", "4802", warningControlSpecPath()}, nil, &stderr)
	if code != tlago.ExitSemanticFailure {
		t.Fatalf("check messages-as-errors exit = %d, want %d; stderr=%s", code, tlago.ExitSemanticFailure, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Warning treated as error") {
		t.Fatalf("elevated warning stderr = %q, want elevation message", stderr.String())
	}
}

func TestWarningControlTest_testCLIMultipleCodesSuppressed(t *testing.T) {
	t.Skip("tla2sany wip")

	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "4800,4802", warningControlSpecPath()}, nil, &stderr)
	if code != tlago.ExitOK {
		t.Fatalf("check multi-suppress exit = %d, want %d; stderr=%s", code, tlago.ExitOK, stderr.String())
	}
	if strings.Contains(stderr.String(), "W4802") || strings.Contains(stderr.String(), "field name") {
		t.Fatalf("multi-suppressed warning stderr = %q, want no field-name warning", stderr.String())
	}
}

func TestWarningControlTest_testCLIUnknownCodeInSuppressMessages(t *testing.T) {
	t.Skip("tla2sany wip")

	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "9999", warningControlSpecPath()}, nil, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("unknown suppress code exit = %d, want %d", code, tlago.ExitToolFailure)
	}
	if !strings.Contains(stderr.String(), "unknown message code") {
		t.Fatalf("unknown suppress code stderr = %q, want unknown message code", stderr.String())
	}
}

func TestWarningControlTest_testCLIUnknownCodeInMessagesAsErrors(t *testing.T) {
	t.Skip("tla2sany wip")

	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-messagesAsErrors", "9999", warningControlSpecPath()}, nil, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("unknown elevated code exit = %d, want %d", code, tlago.ExitToolFailure)
	}
	if !strings.Contains(stderr.String(), "unknown message code") {
		t.Fatalf("unknown elevated code stderr = %q, want unknown message code", stderr.String())
	}
}

func TestWarningControlTest_testCLISuppressMessagesMissingArgument(t *testing.T) {
	t.Skip("tla2sany wip")

	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages"}, nil, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("missing suppress arg exit = %d, want %d", code, tlago.ExitToolFailure)
	}
}

func TestWarningControlTest_testCLIMessagesAsErrorsMissingArgument(t *testing.T) {
	t.Skip("tla2sany wip")

	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-messagesAsErrors"}, nil, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("missing elevated arg exit = %d, want %d", code, tlago.ExitToolFailure)
	}
}

func TestWarningControlTest_testCLIOverlapBetweenSuppressMessagesAndMessagesAsErrors(t *testing.T) {
	t.Skip("tla2sany wip")

	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "4800", "-messagesAsErrors", "4800", warningControlSpecPath()}, nil, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("overlapping code exit = %d, want %d", code, tlago.ExitToolFailure)
	}
	if !strings.Contains(stderr.String(), "both -suppressMessages and -messagesAsErrors") {
		t.Fatalf("overlap stderr = %q, want overlap diagnostic", stderr.String())
	}
}

func TestWarningControlTest_testCLIErrorLevelCodeInSuppressMessages(t *testing.T) {
	t.Skip("tla2sany wip")

	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "4200", warningControlSpecPath()}, nil, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("error-level suppress exit = %d, want %d", code, tlago.ExitToolFailure)
	}
	if !strings.Contains(stderr.String(), "cannot be suppressed") {
		t.Fatalf("error-level suppress stderr = %q, want non-suppressable diagnostic", stderr.String())
	}
}

func warningControlSpecPath() string {
	return sanyTestVectorPath("tla2sany", "semantic", "error_corpus", "W4802_Pre_Test.tla")
}

func warningControlSemanticDiagnostics(t *testing.T) tlago.Diagnostics {
	t.Helper()
	spec, diags := tlago.LoadSanySpec(warningControlSpecPath(), tlago.LoadOptions{})
	requireNoSANYDiagnostics(t, "parse", diags)
	return tlago.CheckSpec(spec)
}
