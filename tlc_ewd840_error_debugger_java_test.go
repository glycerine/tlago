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
// Translation of the complete EWD840ErrorDebuggerTest.testSpec.
package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaEWD840ErrorDebugger(t *testing.T) {
	const RM, MDL = "EWD840", "Error02"
	h := startJavaDebuggerModel(t, RM, MDL, -1, "-config", "Error02.tla")
	defer h.close()
	// This is just the assume that this test ignores.
	stackFrames := h.frames()
	debuggerAssertEqual(t, 1, len(stackFrames))

	vars := tlc.StateVariables()

	// There are no breakpoints to unset, but so what...
	h.unsetBreakpoints()
	stackFrames = h.continueFrames()
	// Error occurs after TLC generated the first initial state.
	i := 9
	debuggerAssertEqual(t, i, len(stackFrames))
	i--
	i--
	assertTLCFrame(t, stackFrames[i], 20, 3, 23, 21, RM)
	i--
	assertTLCStateFrame(t, stackFrames[i], 20, 3, 23, 21, RM, tlc.EmptyContext, vars)
	debuggerAssertTrue(t, stackFrames[i].Base.Exception == nil)
	i--
	assertTLCStateFrame(t, stackFrames[i], 20, 6, 20, 34, RM, tlc.EmptyContext, vars)
	debuggerAssertTrue(t, stackFrames[i].Base.Exception == nil)
	i--
	assertTLCStateFrame(t, stackFrames[i], 21, 6, 21, 31, RM, tlc.EmptyContext, vars[0], vars[2], vars[3])
	debuggerAssertTrue(t, stackFrames[i].Base.Exception == nil)
	i--
	assertTLCStateFrame(t, stackFrames[i], 22, 6, 22, 13, RM, tlc.EmptyContext, vars[0], vars[2])
	debuggerAssertTrue(t, stackFrames[i].Base.Exception == nil)
	i--
	assertTLCStateFrame(t, stackFrames[i], 23, 6, 23, 21, RM, tlc.EmptyContext, vars[2])
	debuggerAssertTrue(t, stackFrames[i].Base.Exception == nil)
	i--
	assertTLCStateFrame(t, stackFrames[i], 14, 13, 14, 22, MDL, tlc.EmptyContext)
	debuggerAssertTrue(t, stackFrames[i].Base.Exception == nil)
	i--
	assertTLCStateFrame(t, stackFrames[i], 14, 13, 14, 22, MDL, tlc.EmptyContext)

	// Assert the exception variable.
	stackFrame := stackFrames[i].Base
	debuggerAssertTrue(t, stackFrame.Exception != nil)
	expVar := stackFrame.GetExceptionAsVariable()
	debuggerAssertEqual(t, 1, len(expVar))

	debuggerAssertEqual(t, "line 14, col 13 to line 14, col 22 of module Error02", expVar[0].Name)
	debuggerAssertEqual(t, "Attempted to check equality of integer 42 with non-integer:\n\"abc\"", expVar[0].Value)
	debuggerAssertEqual(t, "TLCRuntimeException", expVar[0].Type)

	debuggerAssertEqual(t, 0, len(stackFrame.GetVariables(expVar[0].VariablesReference, nil)))
}
