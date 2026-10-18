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
// Translation of the complete EWD840ErrorActionDebuggerTest.testSpec.
package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaEWD840ErrorActionDebugger(t *testing.T) {
	const RM, MDL = "EWD840", "Error03"
	h := startJavaDebuggerModel(t, RM, MDL, -1, "-config", "Error03.tla")
	defer h.close()
	// This is just the assume that this test ignores.
	stackFrames := h.frames()
	debuggerAssertEqual(t, 1, len(stackFrames))

	// There are no breakpoints to unset, but so what...
	h.unsetBreakpoints()
	stackFrames = h.continueFrames()
	// Error occurs after TLC generated the first initial state.
	i := 17
	debuggerAssertEqual(t, i, len(stackFrames))
	for j := i - 1; j > 0; j-- {
		debuggerAssertTrue(t, stackFrames[j].Base.Exception == nil)
	}

	assertTLCActionFrame(t, stackFrames[0], 13, 69, 13, 77, MDL, (*tlc.Context)(nil))

	// Assert the exception variable.
	stackFrame := stackFrames[0].Base
	debuggerAssertTrue(t, stackFrame.Exception != nil)
	expVar := stackFrame.GetExceptionAsVariable()
	debuggerAssertEqual(t, 1, len(expVar))

	debuggerAssertEqual(t, "line 13, col 69 to line 13, col 77 of module Error03", expVar[0].Name)
	debuggerAssertEqual(t, "Attempted to check equality of integer 0 with non-integer:\n\"abc\"", expVar[0].Value)
	debuggerAssertEqual(t, "TLCRuntimeException", expVar[0].Type)

	debuggerAssertEqual(t, 0, len(stackFrame.GetVariables(expVar[0].VariablesReference, nil)))
}
