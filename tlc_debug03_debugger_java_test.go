/*******************************************************************************
 * Copyright (c) 2022 Microsoft Research. All rights reserved.
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
// Translation of the complete Debug03Test.testSpec after source comparison.
package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaDebug03Debugger(t *testing.T) {
	const RM = "Debug03"
	h := startJavaDebuggerModel(t, "debug", RM, tlc.ExitStatusSuccess, "-config", "Debug03.tla")
	defer h.close()
	h.setSpecBreakpoint()

	stackFrames := h.continueFrames()
	debuggerAssertEqual(t, 2, len(stackFrames))
	assertTLCNextStatesFrame(t, stackFrames[0], 14, 16, 14, 19, RM, tlc.EmptyContext, 9)
	assertTLCSyntheticStateStackFrame(t, stackFrames[1], 7, 5, 7, 14, RM, tlc.EmptyContext, 1)

	// Remove all breakpoints and run the spec to completion.
	h.unsetBreakpoints()
	h.continueFrames()
}
