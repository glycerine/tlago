package sany_tests

import (
	"math/big"
	"strings"
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/IncrementalSemanticParseTests.java.
// Each test starts skipped until its Java assertions are ported and made green.
func TestIncrementalSemanticParseTests_basicOpDefTest(t *testing.T) {
	t.Skip("tla2sany wip")

	spec := checkedSANYModuleBody(t, "op == 0")
	def := requireSANYDefinition(t, spec.Root, "op")
	if got := len(def.Params); got != 0 {
		t.Fatalf("op arity = %d, want 0", got)
	}
	if def.Expr == nil {
		t.Fatal("op has nil body")
	}
	if lit, ok := def.Expr.(*tlago.LiteralExpr); !ok || lit.Kind != "number" || lit.Value != "0" {
		t.Fatalf("op body = %#v, want numeric literal 0", def.Expr)
	}
}

func TestIncrementalSemanticParseTests_basicExpressionTest(t *testing.T) {
	t.Skip("tla2sany wip")

	spec := checkedSANYModuleBody(t, "ASSUME 0")
	if got := len(spec.Root.Assumptions); got != 1 {
		t.Fatalf("assumption count = %d, want 1", got)
	}
	if lit, ok := spec.Root.Assumptions[0].Expr.(*tlago.LiteralExpr); !ok || lit.Kind != "number" || lit.Value != "0" {
		t.Fatalf("assumption expr = %#v, want numeric literal 0", spec.Root.Assumptions[0].Expr)
	}
}

func TestIncrementalSemanticParseTests_bigRadixNumeralTest(t *testing.T) {
	t.Skip("tla2sany wip")

	for _, tc := range []struct {
		literal string
		radix   int
	}{
		{literal: `\b` + strings.Repeat("1", 32), radix: 2},
		{literal: `\o` + strings.Repeat("7", 12), radix: 8},
		{literal: `\h` + strings.Repeat("f", 9), radix: 16},
	} {
		spec := checkedSANYModuleBody(t, "op == "+tc.literal)
		def := requireSANYDefinition(t, spec.Root, "op")
		lit, ok := def.Expr.(*tlago.LiteralExpr)
		if !ok || lit.Kind != "number" {
			t.Fatalf("%s body = %#v, want numeric literal", tc.literal, def.Expr)
		}
		expected, ok := new(big.Int).SetString(tc.literal[2:], tc.radix)
		if !ok {
			t.Fatalf("failed to parse expected integer %q", tc.literal)
		}
		actual, ok := new(big.Int).SetString(lit.Value[2:], tc.radix)
		if !ok {
			t.Fatalf("failed to parse actual integer %q", lit.Value)
		}
		if actual.Cmp(expected) != 0 {
			t.Fatalf("%s parsed as %s, want %s", tc.literal, actual, expected)
		}
	}
}

func TestIncrementalSemanticParseTests_letInExpressionTest(t *testing.T) {
	t.Skip("tla2sany wip")

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
	t.Skip("tla2sany wip")

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
