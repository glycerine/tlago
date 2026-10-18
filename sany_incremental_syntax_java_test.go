package tlago

import "testing"

// IncrementalSyntaxParseTests.parser: standalone SPEC-state parsing, with
// belchDEF initialization and no enclosing module or semantic generation.
func javaIncrementalSyntaxParser(input string) *SanyParser {
	manager := NewSanyTokenManager("", input)
	manager.SwitchTo(SanyLexSpec)
	parser := &SanyParser{TokenSource: manager}
	parser.belchDEF()
	return parser
}
func TestIncrementalSyntaxParseTests_testParseBasicOpDef(t *testing.T) {
	result := javaIncrementalSyntaxParser("op == 0").OperatorOrFunctionDefinition()
	if result == nil {
		t.Fatal("null result")
	}
	if result.Kind.JavaName() != "N_OperatorDefinition" {
		t.Fatalf("kind = %s", result.Kind.JavaName())
	}
	heirs := result.GetHeirs()
	if heirs[0].Kind.JavaName() != "N_IdentLHS" {
		t.Fatalf("LHS kind = %s", heirs[0].Kind.JavaName())
	}
	if heirs[0].GetHeirs()[0].GetImage() != "op" {
		t.Fatalf("LHS name = %s", heirs[0].GetHeirs()[0].GetImage())
	}
	if heirs[1].Kind != SanyNodeKind(SanyTokenDef) {
		t.Fatalf("separator kind = %s", heirs[1].Kind.JavaName())
	}
	if heirs[2].Kind.JavaName() != "N_Number" {
		t.Fatalf("body kind = %s", heirs[2].Kind.JavaName())
	}
	if heirs[2].GetHeirs()[0].GetImage() != "0" {
		t.Fatalf("number = %s", heirs[2].GetHeirs()[0].GetImage())
	}
}
func TestIncrementalSyntaxParseTests_testParseBasicExpression(t *testing.T) {
	result := javaIncrementalSyntaxParser("0").ExpressionUntil(func(token *SanyToken) bool { return token.Kind == SanyTokenEOF })
	if result == nil {
		t.Fatal("null result")
	}
	if result.Kind.JavaName() != "N_Number" {
		t.Fatalf("kind = %s", result.Kind.JavaName())
	}
	if result.GetHeirs()[0].GetImage() != "0" {
		t.Fatalf("number = %s", result.GetHeirs()[0].GetImage())
	}
}
func TestIncrementalSyntaxParseTests_testParseConjunctionList(t *testing.T) {
	result := javaIncrementalSyntaxParser("  /\\ TRUE\n  /\\ FALSE\n  /\\ TRUE").ExpressionUntil(func(token *SanyToken) bool { return token.Kind == SanyTokenEOF })
	if result == nil {
		t.Fatal("null result")
	}
	if result.Kind.JavaName() != "N_ConjList" {
		t.Fatalf("kind = %s", result.Kind.JavaName())
	}
	if len(result.GetHeirs()) != 3 {
		t.Fatalf("heirs = %d, want 3", len(result.GetHeirs()))
	}
}
