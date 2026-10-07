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
package sany_tests

import (
	"testing"

	"github.com/glycerine/tlago"
)

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/IncrementalSemanticParseTests.java.
// Remaining methods require reconciliation with the original canonical graph assertions.
// Complete basicOpDefTest, basicExpressionTest and bigRadixNumeralTest live in root
// sany_incremental_semantic_java_test.go.
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
