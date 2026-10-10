/*******************************************************************************
 * Copyright (c) 2021 Microsoft Research. All rights reserved.
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
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
// Translation of all four complete DebugTLCVariableTest methods.
package tlc

import (
	"math/rand"
	"testing"
	"time"
)

func TestJavaDebugTLCVariable(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	assertEquals := func(t *testing.T, want, got int) {
		t.Helper()
		if want != got {
			t.Fatalf("nested variables=%d, want %d", got, want)
		}
	}
	t.Run("testFiniteEmptySetValue", func(t *testing.T) {
		assertEquals(t, 0, len(NewDebugTLCVariableName("4711").SetInstance(NewSetEnumValue([]Value{}, false)).Nested(rnd)))
	})
	t.Run("testFiniteSetValue", func(t *testing.T) {
		outer := NewDebugTLCVariableName("4711").SetInstance(NewSetEnumValue([]Value{NewSetEnumValue([]Value{}, false)}, false)).Nested(rnd)
		assertEquals(t, 1, len(outer))
		assertEquals(t, 0, len(outer[0].Nested(rnd)))
	})
	t.Run("testFiniteNestedValue", func(t *testing.T) {
		vars := NewDebugTLCVariableName("4711").SetInstance(NewSetEnumValue([]Value{NewSetEnumValue([]Value{NewTupleValue([]Value{IntZero, IntOne})}, false)}, false)).Nested(rnd)
		assertEquals(t, 1, len(vars))
		vars = vars[0].Nested(rnd)
		assertEquals(t, 1, len(vars))
		vars = vars[0].Nested(rnd)
		assertEquals(t, 2, len(vars))
		for _, variable := range vars {
			assertEquals(t, 0, len(variable.Nested(rnd)))
		}
	})
	t.Run("testInfiniteValue", func(t *testing.T) {
		assertEquals(t, 0, len(NewDebugTLCVariableName("4711").SetInstance(STRING()).Nested(rnd)))
	})
}
