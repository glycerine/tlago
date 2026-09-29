package sany_tests

import "testing"

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/OperatorAssociativityTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestOperatorAssociativityTests_testOperatorAssociativity(t *testing.T) {
	t.Skip("tla2sany wip")

	for _, op := range sanyOperatorFixtures() {
		if !op.op.IsInfix() {
			continue
		}
		for _, leftSymbol := range op.symbols {
			for _, rightSymbol := range op.symbols {
				expr := "A " + leftSymbol + " B " + rightSymbol + " C"
				_, diags := parseSANYOperatorExpression(expr)
				if op.associative && diags.HasErrors() {
					t.Fatalf("%s: associative operator should parse:\n%s", expr, diags.Error())
				}
				if !op.associative && !diags.HasErrors() {
					t.Fatalf("%s: non-associative operator should not parse", expr)
				}
			}
		}
	}
}
