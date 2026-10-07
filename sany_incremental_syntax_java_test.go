/*******************************************************************************
 * Copyright (c) 2025 Linux Foundation. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/
package tlago

import "testing"

// IncrementalSyntaxParseTests.parser: standalone SPEC-state parsing, with
// belchDEF initialization and no enclosing module or semantic generation.
func javaIncrementalSyntaxParser(input string) *SanyParser {
	manager := NewSanyTokenManager("", input)
	manager.SwitchTo(SanyLexSpec)
	parser := &SanyParser{tokenManager: manager}
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
	if heirs[0].GetHeirs()[0].Image != "op" {
		t.Fatalf("LHS name = %s", heirs[0].GetHeirs()[0].Image)
	}
	if heirs[1].Kind != SanyNodeKind(SanyTokenDef) {
		t.Fatalf("separator kind = %s", heirs[1].Kind.JavaName())
	}
	if heirs[2].Kind.JavaName() != "N_Number" {
		t.Fatalf("body kind = %s", heirs[2].Kind.JavaName())
	}
	if heirs[2].GetHeirs()[0].Image != "0" {
		t.Fatalf("number = %s", heirs[2].GetHeirs()[0].Image)
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
	if result.GetHeirs()[0].Image != "0" {
		t.Fatalf("number = %s", result.GetHeirs()[0].Image)
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
