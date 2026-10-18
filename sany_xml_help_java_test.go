package tlago

import (
	"bytes"
	"strings"
	"testing"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterHelpText.java.
// Compare complete independently captured production usage output.
func TestTestXMLExporterHelpText_testPrintHelpText(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunCLI([]string{"sany-xml", "-help"}, &stdout, &stderr)
	if code != int(XMLExporterOK) {
		t.Fatalf("sany-xml -help exit = %d, want %d; stderr=%s", code, int(XMLExporterOK), stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, originalXMLHelpText()) {
		t.Fatalf("help text = %q, want XML exporter usage", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestTestXMLExporterHelpText_testPrintHelpTextOnNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunCLI([]string{"sany-xml"}, &stdout, &stderr)
	if code != int(XMLArgsParsingFailure) {
		t.Fatalf("sany-xml with no args exit = %d, want %d", code, int(XMLArgsParsingFailure))
	}
	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got := stderr.String(); !strings.Contains(got, originalXMLHelpText()) {
		t.Fatalf("stderr = %q, want help/error usage", got)
	}
}

func originalXMLHelpText() string {
	var output bytes.Buffer
	printSanyXMLUsage(&output)
	return output.String()
}
