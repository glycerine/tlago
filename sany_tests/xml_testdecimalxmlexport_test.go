package sany_tests

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestDecimalXMLExport.java.
func TestTestDecimalXMLExport_test(t *testing.T) {
	spec, err := tlago.ParseSanyXMLSpec(sanyTestVectorPath("test-model", "Decimal.tla"), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	output, err := tlago.SanySpecToXMLString(spec, false, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"<integralPart>000123</integralPart>",
		"<fractionalPart>456000</fractionalPart>",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("XML output missing %q\n%s", want, output)
		}
	}
}
