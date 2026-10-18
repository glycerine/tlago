package sany_tests

import (
	"bytes"
	"github.com/glycerine/tlago"
	"testing"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterStringEscapes.java.
// Preserve original library/CLI result categories and complete XML parsing.
func TestTestXMLExporterStringEscapes_testSupportedStringEscapes(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tlago.RunCLI([]string{"sany-xml", sanyTestVectorPath("test-model", "sany", "StringEscapes.tla")}, &stdout, &stderr)
	if code != int(tlago.XMLExporterOK) {
		t.Fatalf("run = %d, want OK: %s", code, stderr.String())
	}
	got := soleStringValueFromXML(t, stdout.String())
	if want := "\\ \n \r \t \""; got != want {
		t.Fatalf("StringValue = %q, want %q", got, want)
	}
}

func TestTestXMLExporterStringEscapes_testFormFeedStringEscapeIsRejected(t *testing.T) {
	requireXMLExporterFailure(t, []string{sanyTestVectorPath("test-model", "sany", "StringFormFeedEscape.tla")}, tlago.XMLUnrepresentableCharacter, "U+000C", false)
}
func soleStringValueFromXML(t *testing.T, xmlText string) string {
	t.Helper()
	root := parseSANYXMLTestDocument(t, xmlText)
	values := xmlDescendants(root, "StringValue")
	if len(values) != 1 {
		t.Fatalf("StringValue count = %d, want 1", len(values))
	}
	return values[0].text()
}
