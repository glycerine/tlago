/*******************************************************************************
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
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

// Ported from tlaplus/tlatools/org.lamport.tlatools/test/tla2sany/semantic/NestedModuleInstanceTest.java.
func TestNestedModuleInstanceTest_testTopLevelInstanceOfNestedModule(t *testing.T) {
	spec := checkedSANYSpecPath(t, sanyTestVectorPath("test-model", "sany", "NestedModuleTopLevelInstance.tla"))
	requireNoSANYDiagnostics(t, "semantic", tlago.CheckSpec(spec))
}

func TestNestedModuleInstanceTest_testLetInstanceOfNestedModule(t *testing.T) {
	// Preserve the original Java @Ignore reason and status.
	t.Skip("A nested module instantiated only from inside a LET is reported as multiply-defined")
	spec := checkedSANYSpecPath(t, sanyTestVectorPath("test-model", "sany", "NestedModuleLetInstance.tla"))
	requireNoSANYDiagnostics(t, "semantic", tlago.CheckSpec(spec))
}

func checkedSANYSpecPath(t *testing.T, path string) *tlago.Spec {
	t.Helper()
	spec, diags := tlago.LoadSanySpec(path, tlago.LoadOptions{})
	requireNoSANYDiagnostics(t, "parse", diags)
	return spec
}
