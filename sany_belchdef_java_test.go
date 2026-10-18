package tlago

import (
	"strings"
	"testing"
)

// Port of BelchDefTests.runTestCase and its complete five-row matrix. Root
// placement lets the test inspect the actual lazy parser stream and belchDEF.
func TestBelchDefTests_runTestCase(t *testing.T) {
	for _, tc := range []struct {
		name   string
		inputs []string
	}{
		{name: "Empty module"},
		{name: "Single definition", inputs: []string{"op == 0"}},
		{name: "Multiple definitions", inputs: []string{"op == 0", "op2 == 1"}},
		{name: "Named theorem", inputs: []string{"op == 0 THEOREM T == TRUE"}},
		{name: "Named theorem after submodule", inputs: []string{"op == 0 ---- MODULE Inner ---- ==== THEOREM T == TRUE"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parser := &SanyParser{TokenSource: NewSanyTokenManager("", "---- MODULE Test ----\n"+strings.Join(tc.inputs, "\n")+"\n====")}
			// getNextToken advances even when consuming EOF. Access the actual
			// stream directly, since advance is the grammar's EOF-guarded helper.
			next := func() *SanyToken { token := parser.tokenAt(0); parser.at++; return token }
			munch := func(count int) {
				for i := 0; i < count; i++ {
					if next().Kind == SanyTokenDef {
						parser.belchDEF()
					}
				}
			}
			parser.belchDEF()
			// Java's initial EOF dummy is represented by the parser's unconsumed
			// cursor in Go, rather than a token occupying the stream.
			if parser.at != 0 || parser.previous() != nil {
				t.Fatal("belchDEF consumed the initial parser position")
			}
			if next().Kind == SanyTokenEOF {
				t.Fatal("first module token is EOF")
			}
			munch(3) // TokenizerTests.HEADER_TOKENS.size().
			for _, input := range tc.inputs {
				if parser.previous().Kind != SanyTokenDefbreak {
					t.Fatalf("current token = %s, want DEFBREAK", parser.previous().Kind.JavaName())
				}
				munch(len(strings.Split(input, " ")) + 1)
			}
			munch(2) // TokenizerTests.FOOTER_TOKENS.size().
		})
	}
}
