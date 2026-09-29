package sany_tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterHelpText.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestXMLExporterHelpText_testPrintHelpText(t *testing.T) {
	t.Skip("tla2sany wip")

	var stdout, stderr bytes.Buffer
	code := tlago.RunCLI([]string{"sany-xml", "-help"}, &stdout, &stderr)
	if code != tlago.ExitOK {
		t.Fatalf("sany-xml -help exit = %d, want %d; stderr=%s", code, tlago.ExitOK, stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "sany-xml") || !strings.Contains(got, "FILE") {
		t.Fatalf("help text = %q, want XML exporter usage", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestTestXMLExporterHelpText_testPrintHelpTextOnNoArgs(t *testing.T) {
	t.Skip("tla2sany wip")

	var stdout, stderr bytes.Buffer
	code := tlago.RunCLI([]string{"sany-xml"}, &stdout, &stderr)
	if code != tlago.ExitToolFailure {
		t.Fatalf("sany-xml with no args exit = %d, want %d", code, tlago.ExitToolFailure)
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); !strings.Contains(got, "sany-xml") && !strings.Contains(got, "file") {
		t.Fatalf("stderr = %q, want help/error usage", got)
	}
}
