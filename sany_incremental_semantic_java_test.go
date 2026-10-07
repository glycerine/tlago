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

import (
	"math/big"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Ported from IncrementalSemanticParseTests.bigRadixNumeralTest. The test
// stays in the root package to inspect the actual generated semantic node.
func TestIncrementalSemanticParseTests_bigRadixNumeralTest(t *testing.T) {
	for _, tc := range []struct {
		literal string
		radix   int
	}{
		{`\b` + strings.Repeat("1", 32), 2},
		{`\o` + strings.Repeat("7", 12), 8},
		{`\h` + strings.Repeat("f", 9), 16},
	} {
		manager := NewSanyTokenManager("", tc.literal)
		manager.SwitchTo(SanyLexSpec)
		parser := &SanyParser{tokenManager: manager}
		parser.belchDEF()
		syntax := parser.ExpressionUntil(func(token *SanyToken) bool { return token.Kind == SanyTokenEOF })
		expr, log := sanyExpr(syntax)
		generator := sanyExpressionGenerator(nil)
		generator.nodes = newSanyGeneratorNodes()
		generator.currentModule = &Module{semanticNode: newSanySemModuleNode("", nil, Position{})}
		log = append(log, generator.checkExpr(expr, nil, nil)...)
		if log.HasErrors() {
			t.Fatalf("%s: %v", tc.literal, log)
		}
		literal, ok := expr.(*LiteralExpr)
		if !ok {
			t.Fatalf("result = %T, want NumeralNode", expr)
		}
		var result tlc.SemanticNode = literal.numeralNode
		numeral, ok := result.(*tlc.NumeralNode)
		if !ok || numeral == nil {
			t.Fatalf("result = %T, want NumeralNode", result)
		}
		if numeral.UseVal() {
			t.Fatalf("%s: useVal = true, want false", tc.literal)
		}
		expected, valid := new(big.Int).SetString(tc.literal[2:], tc.radix)
		if !valid {
			t.Fatalf("invalid original expected numeral %q", tc.literal)
		}
		if numeral.BigVal() == nil || expected.Cmp(numeral.BigVal()) != 0 {
			t.Fatalf("%s: bigVal = %v, want %v", tc.literal, numeral.BigVal(), expected)
		}
	}
}
