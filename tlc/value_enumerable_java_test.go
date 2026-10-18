/*******************************************************************************
 * Copyright (c) 20178 Microsoft Research. All rights reserved.
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
 *   Ian Morris Nieves - added support for fingerprint stack trace
 ******************************************************************************/
// Complete translation of EnumerableValueTest.test.
package tlc

import "testing"

func TestJavaEnumerableValue(t *testing.T) {
	oldSeed := RandomEnumerableSeed()
	oldRandom := ResetRandomEnumerableValues()
	SetRandomEnumerableSeed(15041980)
	t.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRandom)
	})

	indices := make(map[int]struct{})
	for n := 1; n < 10657; n++ {
		// Java DummyValue(n).elements(n) constructs SubsetEnumerator(n,
		// size() == n). Its other abstract-method stubs are never called.
		enumerator := newRandomSubsetIndices(n, n)
		for enumerator.hasNext() {
			index := enumerator.nextIndex()
			if !(0 <= index && index < n) {
				t.Fatalf("Index %d out of bounds for n=%d", index, n)
			}
			indices[index] = struct{}{}
		}
		if len(indices) != n {
			t.Fatalf("Missing indices for n=%d: got %d, want %d", n, len(indices), n)
		}
		clear(indices)
	}
}
