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
// Translation of the complete Debug02Test.testSpec after source comparison.
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

func TestJavaDebug02Debugger(t *testing.T) {
	const RM = "Debug02"
	h := startJavaDebuggerModel(t, "debug", RM, tlc.ExitStatusSuccess, "-config", "Debug02.tla")
	defer h.close()

	// xx in Init
	evaluated := h.evaluateHover(RM, "x", 6, 9, 6, 9)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "?", *evaluated.Result)
	// xx in Next
	evaluated = h.evaluateHover(RM, "x", 7, 14, 7, 14)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "?", *evaluated.Result)
	// xx' in Next
	evaluated = h.evaluateHover(RM, "x", 7, 9, 7, 9)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "line 7, col 9 to line 7, col 9 of module Debug02", *evaluated.Result)

	// outer-most frame of Init => xx and xx' still null
	h.stepIn()
	debuggerAssertTrue(t, h.debugger.TopFrame().State != nil)
	evaluated = h.evaluateHover(RM, "x", 6, 9, 6, 9)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "?", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 14, 7, 14)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "?", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 9, 7, 9)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "line 7, col 9 to line 7, col 9 of module Debug02", *evaluated.Result)

	// xx evaluated, xx' not yet
	h.stepIn()
	debuggerAssertTrue(t, h.debugger.TopFrame().State != nil)
	// xx in Init
	evaluated = h.evaluateHover(RM, "x", 6, 9, 6, 9)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "TRUE", *evaluated.Result)
	// xx in Next
	evaluated = h.evaluateHover(RM, "x", 7, 14, 7, 14)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "TRUE", *evaluated.Result)
	// xx' in Next
	evaluated = h.evaluateHover(RM, "x", 7, 9, 7, 9)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "line 7, col 9 to line 7, col 9 of module Debug02", *evaluated.Result)

	// outer-most frame of Next => xx evaluated, xx' not yet
	h.stepIn(2)
	debuggerAssertTrue(t, h.debugger.TopFrame().Action != nil)
	debuggerAssertTrue(t, !h.debugger.TopFrame().Action.GetT().AllAssigned())
	evaluated = h.evaluateHover(RM, "x", 6, 9, 6, 9)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "TRUE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 14, 7, 14)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "TRUE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 9, 7, 10)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "?", *evaluated.Result)

	// ... =~xx
	h.stepIn()
	debuggerAssertTrue(t, h.debugger.TopFrame().Action != nil)
	debuggerAssertTrue(t, !h.debugger.TopFrame().Action.GetT().AllAssigned())
	evaluated = h.evaluateHover(RM, "x", 6, 9, 6, 9)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "TRUE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 14, 7, 14)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "TRUE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 9, 7, 10)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "?", *evaluated.Result)

	// xx' =~xx
	h.stepIn()
	debuggerAssertTrue(t, h.debugger.TopFrame().Action != nil)
	peek := h.debugger.TopFrame().Action
	debuggerAssertTrue(t, !peek.GetT().AllAssigned())
	evaluated = h.evaluateHover(RM, "x", 6, 9, 6, 9)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "FALSE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 14, 7, 14)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "FALSE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 9, 7, 10)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "?", *evaluated.Result)

	h.stepIn()
	debuggerAssertTrue(t, h.debugger.TopFrame().Action != nil)
	debuggerAssertTrue(t, !h.debugger.TopFrame().Action.GetT().AllAssigned())
	evaluated = h.evaluateHover(RM, "x", 6, 9, 6, 9)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "FALSE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 14, 7, 14)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "FALSE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 9, 7, 10)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)
	debuggerAssertEqual(t, "?", *evaluated.Result)

	// Assert that constants of a single module spec (a spec without instantiation
	// and variables declared only in one module) gets flattened in the variable view.
	f := h.debugger.TopFrame().Action
	variables := f.GetVariables(f.GetConstantsID(), nil)
	debuggerAssertEqual(t, 2, len(variables))
	debuggerAssertEqual(t, "Debug02", variables[0].Name)
	variables = f.GetVariables(variables[0].VariablesReference, nil)
	debuggerAssertEqual(t, "val", variables[0].Name)
	debuggerAssertEqual(t, "42", variables[0].Value)

	h.setSpecBreakpoint()
	stepIn := h.continueFrames()
	assertTLCNextStatesFrame(t, stepIn[0], 8, 20, 8, 23, RM, tlc.EmptyContext, 1)
	debuggerAssertTrue(t, h.debugger.TopFrame().Next != nil)
	f2 := h.debugger.TopFrame().Next
	debuggerAssertTrue(t, f2.GetT().AllAssigned())
	evaluated = h.evaluateHover(RM, "x", 6, 9, 6, 9)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "FALSE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 14, 7, 14)
	debuggerAssertEqual(t, "BoolValue: "+tlc.BoolTrue.KindString(), *evaluated.Type)
	debuggerAssertEqual(t, "FALSE", *evaluated.Result)
	evaluated = h.evaluateHover(RM, "x", 7, 9, 7, 10)
	debuggerAssertEqual(t, (*string)(nil), evaluated.Type)

	// Remove all breakpoints and run the spec to completion.
	h.unsetBreakpoints()
	h.continueFrames()
}
