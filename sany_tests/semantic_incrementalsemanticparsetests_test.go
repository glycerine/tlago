package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/IncrementalSemanticParseTests.java.
// These are supplementary AST facade checks. All five complete original-method
// translations, including canonical graph and source-identity assertions, live
// in root sany_incremental_semantic_java_test.go.
func TestIncrementalSemanticParseTests_letInExpressionTest(t *testing.T) {
	spec := checkedSANYModuleBody(t, "op == LET M == INSTANCE Naturals IN M!+(1, 2)")
	if spec.Modules["Naturals"] == nil {
		t.Fatal("Naturals dependency was not loaded")
	}
	def := requireSANYDefinition(t, spec.Root, "op")
	let, ok := def.Expr.(*tlago.LetExpr)
	if !ok {
		t.Fatalf("op body = %#v, want LET expression", def.Expr)
	}
	if got := len(let.Instances); got != 1 {
		t.Fatalf("LET instance count = %d, want 1", got)
	}
	if inst := let.Instances[0]; inst.Name != "M" || inst.Module != "Naturals" {
		t.Fatalf("LET instance = %#v, want M == INSTANCE Naturals", inst)
	}
	call, ok := let.Body.(*tlago.CallExpr)
	if !ok {
		t.Fatalf("LET body = %#v, want operator call", let.Body)
	}
	if got := len(call.Args); got != 2 {
		t.Fatalf("M!+ argument count = %d, want 2", got)
	}
}

func TestIncrementalSemanticParseTests_letInExpressionWithTransitiveDepsTest(t *testing.T) {
	spec := checkedSANYModuleBody(t, "op == LET T == INSTANCE TLC IN T!JavaTime")
	for _, name := range []string{"TLC", "Naturals", "Sequences", "FiniteSets"} {
		if spec.Modules[name] == nil {
			t.Fatalf("%s dependency was not loaded", name)
		}
	}
	def := requireSANYDefinition(t, spec.Root, "op")
	let, ok := def.Expr.(*tlago.LetExpr)
	if !ok {
		t.Fatalf("op body = %#v, want LET expression", def.Expr)
	}
	if got := len(let.Instances); got != 1 {
		t.Fatalf("LET instance count = %d, want 1", got)
	}
	if inst := let.Instances[0]; inst.Name != "T" || inst.Module != "TLC" {
		t.Fatalf("LET instance = %#v, want T == INSTANCE TLC", inst)
	}
	if _, ok := let.Body.(*tlago.IdentExpr); !ok {
		t.Fatalf("LET body = %#v, want JavaTime operator reference", let.Body)
	}
}

func checkedSANYModuleBody(t *testing.T, body string) *tlago.Spec {
	t.Helper()
	spec, diags := tlago.CheckSanySource("Test.tla", wrapSANYTestModule(body))
	requireNoSANYDiagnostics(t, "check", diags)
	return spec
}

func requireSANYDefinition(t *testing.T, mod *tlago.Module, name string) tlago.Definition {
	t.Helper()
	if mod == nil {
		t.Fatal("missing module")
	}
	for _, def := range mod.Definitions {
		if def.Name == name {
			return def
		}
	}
	t.Fatalf("definition %s not found in %#v", name, mod.Definitions)
	return tlago.Definition{}
}
