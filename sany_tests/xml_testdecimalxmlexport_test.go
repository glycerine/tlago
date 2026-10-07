/*******************************************************************************
 * Copyright (c) 2026 The Linux Foundation. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/
package sany_tests

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/xml/TestDecimalXMLExport.java.
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
