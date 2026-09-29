package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/TestBuiltInOperatorInitialization.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestTestBuiltInOperatorInitialization_testInitAndReInit(t *testing.T) {
	for i := 0; i < 2; i++ {
		for _, symbol := range []string{`\lnot`, `\neg`, "SUBSET", "UNION", "DOMAIN", "[]", "<>", `\in`} {
			if _, ok := tlago.GetSanyOperator(symbol); !ok {
				t.Fatalf("built-in operator %q missing after initialization pass %d", symbol, i)
			}
		}
		for _, body := range []string{
			`op == \neg FALSE`,
			"op == SUBSET {1}",
			"op == UNION {{1}}",
			"op == DOMAIN [a |-> 1]",
			"op == []TRUE",
			"op == <>TRUE",
		} {
			checkedSANYModuleBody(t, body)
		}
	}
}
