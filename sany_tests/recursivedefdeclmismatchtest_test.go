package sany_tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/RecursiveDefDeclMismatchTest.java.
// Preserve the void SANYmain output contract without an added exit assertion.
func TestRecursiveDefDeclMismatchTest_test(t *testing.T) {
	var output bytes.Buffer
	tlago.RunCLI(
		[]string{"check", sanyTestVectorPath("test-model", "sany", "RecursiveDefDeclMismatch.tla")},
		&output,
		&output,
	)
	got := output.String()
	for _, want := range []string{
		"Definition of CountDown has different arity than its RECURSIVE declaration.",
		"The operator CountDown requires 2 arguments.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("SANY output missing %q\n%s", want, got)
		}
	}
}
