package sany_tests

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/Github723Test.java.
// Preserve the void SANYmain output contract without an added exit assertion.
func TestGithub723Test_test(t *testing.T) {
	var output bytes.Buffer
	tlago.RunCLI(
		[]string{"check", sanyTestVectorPath("test-model", "sany", "Github723.tla")},
		&output,
		&output,
	)
	got := output.String()
	want := "An operator must be substituted for symbol 'C', and it must have arity 1."
	if !strings.Contains(got, want) {
		t.Fatalf("SANY output missing %q\n%s", want, got)
	}
}
