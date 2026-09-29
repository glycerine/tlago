package sany_tests

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestXMLExporterStringEscapes.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestXMLExporterStringEscapes_testSupportedStringEscapes(t *testing.T) {
	got := soleStringValueFromXML(t, checkedSANYXMLForPath(t, sanyTestVectorPath("test-model", "sany", "StringEscapes.tla")))
	if want := "\\ \n \r \t \""; got != want {
		t.Fatalf("StringValue = %q, want %q", got, want)
	}
}

func TestTestXMLExporterStringEscapes_testFormFeedStringEscapeIsRejected(t *testing.T) {
	_, diags := xmlForUnrepresentableStringFixture(sanyTestVectorPath("test-model", "sany", "StringFormFeedEscape.tla"))
	if !diags.HasErrors() {
		t.Fatal("expected XML export to reject form-feed string escape")
	}
	if got, want := diags.Error(), "U+000C"; !strings.Contains(got, want) {
		t.Fatalf("XML diagnostics missing %q\n%s", want, got)
	}
}

func soleStringValueFromXML(t *testing.T, xmlText string) string {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(xmlText))
	var values []string
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "StringValue" {
			continue
		}
		var value string
		if err := decoder.DecodeElement(&value, &start); err != nil {
			t.Fatalf("decode StringValue: %v", err)
		}
		values = append(values, value)
	}
	if len(values) != 1 {
		t.Fatalf("StringValue count = %d, want 1 in XML:\n%s", len(values), xmlText)
	}
	return values[0]
}

func xmlForUnrepresentableStringFixture(path string) (string, tlago.Diagnostics) {
	spec, diags := tlago.LoadSanySpec(path, tlago.LoadOptions{})
	if diags.HasErrors() {
		return "", diags
	}
	diags = append(diags, tlago.CheckSpec(spec)...)
	if diags.HasErrors() {
		return "", diags
	}
	xmlText, xmlDiags := tlago.SanyXML(spec)
	diags = append(diags, xmlDiags...)
	return string(xmlText), diags
}
