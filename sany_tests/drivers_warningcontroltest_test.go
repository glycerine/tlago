package sany_tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/drivers/WarningControlTest.java.
// SANY ERROR is mapped to tlago.ExitToolFailure; OK and semantic failure retain 0 and 4.
func TestWarningControlTest_testCLISuppressMessagesSilencesWarning(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "4802", "-error-codes", warningControlSpecPath()}, &stderr, &stderr)
	if code != tlago.ExitOK {
		t.Fatalf("check suppress exit = %d, want %d; stderr=%s", code, tlago.ExitOK, stderr.String())
	}
	if strings.Contains(stderr.String(), "field name") {
		t.Fatalf("suppressed warning stderr = %q, want no field-name warning", stderr.String())
	}
}

func TestWarningControlTest_testCLIMessagesAsErrorsCausesFailure(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-messagesAsErrors", "4802", "-error-codes", warningControlSpecPath()}, &stderr, &stderr)
	if code != tlago.ExitSemanticFailure {
		t.Fatalf("check messages-as-errors exit = %d, want %d; stderr=%s", code, tlago.ExitSemanticFailure, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Warning treated as error") {
		t.Fatalf("elevated warning stderr = %q, want elevation message", stderr.String())
	}
}

func TestWarningControlTest_testCLIMultipleCodesSuppressed(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "4800,4802", "-error-codes", warningControlSpecPath()}, &stderr, &stderr)
	if code != tlago.ExitOK {
		t.Fatalf("check multi-suppress exit = %d, want %d; stderr=%s", code, tlago.ExitOK, stderr.String())
	}
	if strings.Contains(stderr.String(), "field name") {
		t.Fatalf("multi-suppressed warning stderr = %q, want no field-name warning", stderr.String())
	}
}

func TestWarningControlTest_testCLIUnknownCodeInSuppressMessages(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "9999", warningControlSpecPath()}, &stderr, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("unknown suppress code exit = %d, want %d", code, tlago.ExitToolFailure)
	}
	if !strings.Contains(stderr.String(), "unknown message code") {
		t.Fatalf("unknown suppress code stderr = %q, want unknown message code", stderr.String())
	}
}

func TestWarningControlTest_testCLIUnknownCodeInMessagesAsErrors(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-messagesAsErrors", "9999", warningControlSpecPath()}, &stderr, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("unknown elevated code exit = %d, want %d", code, tlago.ExitToolFailure)
	}
	if !strings.Contains(stderr.String(), "unknown message code") {
		t.Fatalf("unknown elevated code stderr = %q, want unknown message code", stderr.String())
	}
}

func TestWarningControlTest_testCLISuppressMessagesMissingArgument(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages"}, &stderr, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("missing suppress arg exit = %d, want %d", code, tlago.ExitToolFailure)
	}
}

func TestWarningControlTest_testCLIMessagesAsErrorsMissingArgument(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-messagesAsErrors"}, &stderr, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("missing elevated arg exit = %d, want %d", code, tlago.ExitToolFailure)
	}
}

func TestWarningControlTest_testCLIOverlapBetweenSuppressMessagesAndMessagesAsErrors(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "4800", "-messagesAsErrors", "4800", warningControlSpecPath()}, &stderr, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("overlapping code exit = %d, want %d", code, tlago.ExitToolFailure)
	}
	if !strings.Contains(stderr.String(), "codes were set to both -suppressMessages and -messagesAsErrors") {
		t.Fatalf("overlap stderr = %q, want overlap diagnostic", stderr.String())
	}
}

func TestWarningControlTest_testCLIErrorLevelCodeInSuppressMessages(t *testing.T) {
	var stderr bytes.Buffer
	code := tlago.RunCLI([]string{"check", "-suppressMessages", "4200", warningControlSpecPath()}, &stderr, &stderr)
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
