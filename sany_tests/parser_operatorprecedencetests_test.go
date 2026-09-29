package sany_tests

import (
	"fmt"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/OperatorPrecedenceTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestOperatorPrecedenceTests_testOperatorCombination(t *testing.T) {
	operators := sanyOperatorFixtures()
	for _, op1 := range operators {
		for _, op2 := range operators {
			if op1.op.IsPostfix() && op2.op.IsPrefix() {
				continue
			}
			for _, op1Symbol := range op1.symbols {
				for _, op2Symbol := range op2.symbols {
					if op1.op.IsInfix() && op2.op.IsPrefix() &&
						op2.op.Symbol == "-" && sanyOperatorLowerPrecThan(op2, op1) {
						continue
					}
					expr := deriveSANYOperatorExpression(op1, op1Symbol, op2, op2Symbol)
					root, diags := parseSANYOperatorExpression(expr)
					expectParseSuccess := !sanyOperatorConflictsWith(op1, op2)
					if expectParseSuccess && diags.HasErrors() {
						t.Fatalf("%s: expected parse success:\n%s", expr, diags.Error())
					}
					if !expectParseSuccess && !diags.HasErrors() {
						t.Fatalf("%s: expected parse failure", expr)
					}
					if expectParseSuccess {
						checkSANYParsePrecedence(t, root, op1, op1Symbol, op2, op2Symbol)
					}
				}
			}
		}
	}
}

func deriveSANYOperatorExpression(op1 sanyOperatorFixture, op1Symbol string, op2 sanyOperatorFixture, op2Symbol string) string {
	switch {
	case op1.op.IsPrefix() && op2.op.IsPrefix():
		return fmt.Sprintf("%s %s A", op1Symbol, op2Symbol)
	case op1.op.IsPrefix() && op2.op.IsInfix():
		return fmt.Sprintf("%s A %s B", op1Symbol, op2Symbol)
	case op1.op.IsPrefix() && op2.op.IsPostfix():
		return fmt.Sprintf("%s A %s", op1Symbol, op2Symbol)
	case op1.op.IsInfix() && op2.op.IsPrefix():
		return fmt.Sprintf("A %s %s B", op1Symbol, op2Symbol)
	case op1.op.IsInfix() && op2.op.IsInfix():
		return fmt.Sprintf("A %s B %s C", op1Symbol, op2Symbol)
	case op1.op.IsInfix() && op2.op.IsPostfix():
		return fmt.Sprintf("A %s B %s", op1Symbol, op2Symbol)
	case op1.op.IsPostfix() && op2.op.IsInfix():
		return fmt.Sprintf("A %s %s B", op1Symbol, op2Symbol)
	case op1.op.IsPostfix() && op2.op.IsPostfix():
		return fmt.Sprintf("A %s %s", op1Symbol, op2Symbol)
	default:
		panic("unsupported operator fixture combination")
	}
}

func checkSANYParsePrecedence(t *testing.T, root *tlago.SanySyntaxNode, op1 sanyOperatorFixture, op1Symbol string, op2 sanyOperatorFixture, op2Symbol string) {
	t.Helper()
	lowerPrecOp := sanyOperatorExpressionInAssume(root)
	lowerPrecOpSymbol := sanyOperatorImage(lowerPrecOp)
	higherPrecOp := sanyHigherPrecOperator(lowerPrecOp)
	higherPrecOpSymbol := sanyOperatorImage(higherPrecOp)
	parsedOp1Symbol := sanyParsedOperatorSymbol(op1, op1Symbol)
	parsedOp2Symbol := sanyParsedOperatorSymbol(op2, op2Symbol)
	switch {
	case sanySameOperatorFixture(op1, op2) && op1.associative:
		requireSANYOperatorSymbols(t, lowerPrecOpSymbol, parsedOp2Symbol, higherPrecOpSymbol, parsedOp1Symbol)
	case sanySameOperatorFixture(op1, op2):
		requireSANYOperatorSymbols(t, lowerPrecOpSymbol, parsedOp1Symbol, higherPrecOpSymbol, parsedOp2Symbol)
	case sanyOperatorLowerPrecThan(op1, op2) || op2.op.IsPrefix():
		requireSANYOperatorSymbols(t, lowerPrecOpSymbol, parsedOp1Symbol, higherPrecOpSymbol, parsedOp2Symbol)
	default:
		requireSANYOperatorSymbols(t, lowerPrecOpSymbol, parsedOp2Symbol, higherPrecOpSymbol, parsedOp1Symbol)
	}
}

func sanyParsedOperatorSymbol(op sanyOperatorFixture, symbol string) string {
	if op.op.IsPrefix() && symbol == "-" {
		return "-."
	}
	return symbol
}

func requireSANYOperatorSymbols(t *testing.T, gotLower, wantLower, gotHigher, wantHigher string) {
	t.Helper()
	if gotLower != wantLower || gotHigher != wantHigher {
		t.Fatalf("operator precedence = lower %q higher %q, want lower %q higher %q", gotLower, gotHigher, wantLower, wantHigher)
	}
}

func sanyOperatorExpressionInAssume(root *tlago.SanySyntaxNode) *tlago.SanySyntaxNode {
	module := root.GetHeirs()
	body := module[2]
	units := body.GetHeirs()
	assume := units[0]
	return assume.GetHeirs()[1]
}

func sanyHigherPrecOperator(lowerPrecOp *tlago.SanySyntaxNode) *tlago.SanySyntaxNode {
	heirs := lowerPrecOp.GetHeirs()
	switch lowerPrecOp.Kind.JavaName() {
	case "N_PrefixExpr":
		return heirs[1]
	case "N_InfixExpr":
		if sanyIsOperatorExpression(heirs[0]) {
			return heirs[0]
		}
		return heirs[2]
	case "N_PostfixExpr":
		return heirs[0]
	default:
		return nil
	}
}

func sanyIsOperatorExpression(node *tlago.SanySyntaxNode) bool {
	if node == nil {
		return false
	}
	switch node.Kind.JavaName() {
	case "N_PrefixExpr", "N_InfixExpr", "N_PostfixExpr":
		return true
	default:
		return false
	}
}

func sanyOperatorImage(op *tlago.SanySyntaxNode) string {
	if op == nil {
		return ""
	}
	heirs := op.GetHeirs()
	switch op.Kind.JavaName() {
	case "N_PrefixExpr":
		return heirs[0].GetHeirs()[1].Image
	case "N_InfixExpr":
		return heirs[1].GetHeirs()[1].Image
	case "N_PostfixExpr":
		return heirs[1].GetHeirs()[1].Image
	default:
		return ""
	}
}
