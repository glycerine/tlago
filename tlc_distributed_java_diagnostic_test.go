//go:build tlago_disabled_distributed_tests

/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
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

package tlago

import "testing"

// These diagnostics deliberately bypass Java's unconditional setup assumption.
// They retain every original assertion and source Ant configuration. Passing a
// diagnostic is not an enabled original-test pass. The final off-heap flusher
// lifecycle bug is deliberately fixed in Go; see JAVA_BUG_FOUND.md.
func TestDiagnosticJavaDieHardDistributed(t *testing.T) {
	checkJavaDieHardDistributed(t)
}

func TestDiagnosticJavaEWD840Distributed(t *testing.T) {
	checkJavaEWD840Distributed(t)
}

func TestDiagnosticJavaEWD840DistributedWithFPSet(t *testing.T) {
	checkJavaEWD840DistributedWithFPSet(t)
}

func TestDiagnosticJavaTSnapShotDistributed(t *testing.T) {
	checkJavaTSnapShotDistributed(t)
}
