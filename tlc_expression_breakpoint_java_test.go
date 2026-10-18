/*******************************************************************************
 * Copyright (c) 2024 Linux Foundation. All rights reserved.
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
// Translation of the complete ExpressionBreakpointTest.testSpec.
package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaExpressionBreakpoint(t *testing.T) {
	const RM = "ExpressionBreakpointTest"
	h := startJavaDebuggerModel(t, RM, RM, tlc.ExitStatusSuccess, "-config", RM+".tla")
	defer h.close()
	h.setBreakpoints(debuggerBreakpoint(RM, 8, 5, 1, "(i + l + k) > j"))
	h.continueFrames()
	current := h.continueFrames()[0]
	debuggerAssertEqual(t, int32(2), current.State.GetT().Lookup(tlc.UniqueStringOf("i")).(*tlc.IntValue).Val)
	assertTLCStateFrame(t, current, 8, 8, RM, map[string]string{"k": "5", "l": "4"})
	h.unsetBreakpoints()
	h.continueFrames()
}
