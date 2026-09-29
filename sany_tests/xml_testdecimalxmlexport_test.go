package sany_tests

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestDecimalXMLExport.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestDecimalXMLExport_test(t *testing.T) {
	spec, diags := tlago.LoadSanySpec(sanyTestVectorPath("test-model", "Decimal.tla"), tlago.LoadOptions{})
	requireNoSANYDiagnostics(t, "parse", diags)
	requireNoSANYDiagnostics(t, "semantic", tlago.CheckSpec(spec))
	xmlText, xmlDiags := tlago.SanyXML(spec)
	requireNoSANYDiagnostics(t, "xml", xmlDiags)
	output := string(xmlText)
	for _, want := range []string{
		"<integralPart>000123</integralPart>",
		"<fractionalPart>456000</fractionalPart>",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("XML output missing %q\n%s", want, output)
		}
	}
}
