package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/parser/IncrementalSyntaxParseTests.java.
// Supplementary enclosing-module checks; the complete original standalone
// parser assertions live in root sany_incremental_syntax_java_test.go.
func TestIncrementalSyntaxNativeAPI_testParseBasicOpDef(t *testing.T) {
	root, diags := tlago.ParseSanySyntax("Test.tla", "---- MODULE Test ----\nop == 0\n====")
	requireNoSANYDiagnostics(t, "parse", diags)
	def := findSANYNodeByJavaKind(root, "N_OperatorDefinition")
	requireSANYNodeKind(t, def, "N_OperatorDefinition")
	heirs := def.GetHeirs()
	if len(heirs) < 3 {
		t.Fatalf("operator definition has %d heirs, want at least 3", len(heirs))
	}
	requireSANYNodeKind(t, heirs[0], "N_IdentLHS")
	lhsHeirs := heirs[0].GetHeirs()
	if len(lhsHeirs) == 0 || lhsHeirs[0].Image != "op" {
		t.Fatalf("operator definition lhs = %#v, want op", lhsHeirs)
	}
	if got := heirs[1].Kind.JavaName(); got != "DEF" {
		t.Fatalf("operator definition separator kind = %s, want DEF", got)
	}
	requireSANYNodeKind(t, heirs[2], "N_Number")
	numHeirs := heirs[2].GetHeirs()
	if len(numHeirs) == 0 || numHeirs[0].Image != "0" {
		t.Fatalf("operator definition body number = %#v, want 0", numHeirs)
	}
}

func TestIncrementalSyntaxNativeAPI_testParseBasicExpression(t *testing.T) {
	root, diags := tlago.ParseSanySyntax("Test.tla", "---- MODULE Test ----\nASSUME 0\n====")
	requireNoSANYDiagnostics(t, "parse", diags)
	num := findSANYNodeByJavaKind(root, "N_Number")
	requireSANYNodeKind(t, num, "N_Number")
	heirs := num.GetHeirs()
	if len(heirs) == 0 || heirs[0].Image != "0" {
		t.Fatalf("expression number = %#v, want 0", heirs)
	}
}

func TestIncrementalSyntaxNativeAPI_testParseConjunctionList(t *testing.T) {
	root, diags := tlago.ParseSanySyntax("Test.tla", "---- MODULE Test ----\nASSUME\n  /\\ TRUE\n  /\\ FALSE\n  /\\ TRUE\n====")
	requireNoSANYDiagnostics(t, "parse", diags)
	conj := findSANYNodeByJavaKind(root, "N_ConjList")
	requireSANYNodeKind(t, conj, "N_ConjList")
	if got := len(conj.GetHeirs()); got != 3 {
		t.Fatalf("conjunction list heirs = %d, want 3", got)
	}
}
