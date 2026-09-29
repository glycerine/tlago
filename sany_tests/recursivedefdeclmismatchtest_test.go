package sany_tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/RecursiveDefDeclMismatchTest.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestRecursiveDefDeclMismatchTest_test(t *testing.T) {
	t.Skip("tla2sany wip")

	var stdout, stderr bytes.Buffer
	code := tlago.RunCLI(
		[]string{"check", sanyTestVectorPath("test-model", "sany", "RecursiveDefDeclMismatch.tla")},
		&stdout,
		&stderr,
	)
	if code == tlago.ExitOK {
		t.Fatalf("check exit = %d, want failure", code)
	}
	got := stdout.String() + stderr.String()
	for _, want := range []string{
		"Definition of CountDown has different arity than its RECURSIVE declaration.",
		"The operator CountDown requires 2 arguments.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("SANY output missing %q\n%s", want, got)
		}
	}
}
