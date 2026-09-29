package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/CanonicalOperatorTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestCanonicalOperatorTests_testCanonicalOperatorCorrectness(t *testing.T) {
	for _, op := range sanyOperatorFixtures() {
		expected := op.symbols[0]
		for _, symbol := range op.symbols {
			if _, ok := tlago.GetSanyOperator(symbol); !ok {
				t.Fatalf("operator %q is not registered", symbol)
			}
			if got := tlago.ResolveSanyOperatorSynonym(symbol); got != expected {
				t.Fatalf("canonical operator for %q = %q, want %q", symbol, got, expected)
			}
		}
	}
}
