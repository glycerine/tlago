package sany_tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/Github723Test.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestGithub723Test_test(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tlago.RunCLI(
		[]string{"check", sanyTestVectorPath("test-model", "sany", "Github723.tla")},
		&stdout,
		&stderr,
	)
	if code == tlago.ExitOK {
		t.Fatalf("check exit = %d, want failure", code)
	}
	got := stdout.String() + stderr.String()
	want := "An operator must be substituted for symbol 'C', and it must have arity 1."
	if !strings.Contains(got, want) {
		t.Fatalf("SANY output missing %q\n%s", want, got)
	}
}
