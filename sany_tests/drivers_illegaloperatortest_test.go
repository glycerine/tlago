package sany_tests

import (
	"bytes"
	"github.com/glycerine/tlago"
	"strings"
	"testing"
)

// Ported from drivers/IllegalOperatorTest.java: extensionless CLI input,
// combined streams, and the four original exact substring assertions.
func TestIllegalOperatorTest_test(t *testing.T) {
	var output bytes.Buffer
	tlago.RunCLI([]string{"check", sanyTestVectorPath("test-model", "IllegalOperatorTest")}, &output, &output)
	for _, want := range []string{
		"*** Errors: 1\n",
		"line 3, col 8 to line 3, col 11 of module IllegalOperatorTest\n",
		"Argument number 1 to operator 'D' \n",
		"should be a 1-parameter operator.",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("SANY output lacks %q:\n%s", want, output.String())
		}
	}
}
