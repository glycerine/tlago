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
// Translation of the complete EWD840DebuggerSimTest.testSpec.
package tlago

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaEWD840DebuggerSim(t *testing.T) {
	const RM, MDL = "EWD840", "MC02Sim"
	h := startJavaDebuggerModel(t, RM, MDL, tlc.ExitStatusViolationSafety,
		"-config", "MC02.cfg", "-simulate", "num=1", "-seed", "111123", "-fp", "1")
	defer h.close()
	stackFrames := h.frames()
	debuggerAssertEqual(t, 1, len(stackFrames))

	assertTLCFrame(t, stackFrames[0], 5, 5, RM)
	// prefix depends on where the tests execute.
	debuggerAssertTrue(t, strings.HasSuffix(strings.ReplaceAll(h.sourcePath(stackFrames[0]), "\\", "/"), "test_vectors/models/EWD840/EWD840.tla"))
	stackFrames = h.stepIn()
	debuggerAssertEqual(t, 2, len(stackFrames))
	assertTLCFrame(t, stackFrames[1], 5, 5, RM)
	assertTLCFrame(t, stackFrames[0], 5, 5, RM)

	// Assert the constants of EWD840 and MC02.
	f := stackFrames[0]
	variables := f.GetVariables(f.Base.GetConstantsID(), nil)
	debuggerAssertEqual(t, 3, len(variables))
	debuggerAssertEqual(t, RM, variables[0].Name)
	nested := f.GetVariables(variables[0].VariablesReference, nil)
	debuggerAssertEqual(t, 3, len(nested))
	debuggerAssertEqual(t, "Color", nested[0].Name)
	debuggerAssertEqual(t, "{\"white\", \"black\"}", nested[0].Value)
	debuggerAssertEqual(t, "Nodes", nested[1].Name)
	debuggerAssertEqual(t, "{0, 1}", nested[1].Value)
	debuggerAssertEqual(t, "const_143073460396411000", nested[2].Name)
	debuggerAssertEqual(t, "2", nested[2].Value)

	debuggerAssertEqual(t, MDL, variables[1].Name)
	nested = f.GetVariables(variables[1].VariablesReference, nil)
	debuggerAssertEqual(t, 1, len(nested))
	debuggerAssertEqual(t, "const_143073460396411000", nested[0].Name)
	debuggerAssertEqual(t, "2", nested[0].Value)

	vars := tlc.StateVariables()

	h.replaceBreakpoints(RM, 23)

	// The spec has 16 initial states over which we will continue each time checking
	// the stack frames:
	for i := 0; i < 16; i++ { //64
		stackFrames = h.continueFrames()

		debuggerAssertEqual(t, 7, len(stackFrames))
		assertTLCStateFrame(t, stackFrames[4], 20, 23, RM, vars)
		assertTLCStateFrame(t, stackFrames[3], 20, 20, RM, vars)
		assertTLCStateFrame(t, stackFrames[2], 21, 21, RM, vars[0], vars[2], vars[3])
		assertTLCStateFrame(t, stackFrames[1], 22, 22, RM, vars[0], vars[2])
		assertTLCStateFrame(t, stackFrames[0], 23, 23, RM, vars[2])
	}

	// Debug the InitiateProbe action of the next-state relation.
	h.replaceBreakpoints(RM, 26)
	stackFrames = h.continueFrames()

	// First frame captures the complete action.
	debuggerAssertEqual(t, 3, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[0], 26, 31, RM, vars)

	// Second frame captures the first line.
	stackFrames = h.stepIn()
	debuggerAssertEqual(t, 4, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[1], 26, 31, RM, vars)
	assertTLCActionFrame(t, stackFrames[0], 26, 26, RM, vars)

	// Third frame.
	stackFrames = h.stepIn()
	debuggerAssertEqual(t, 5, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[2], 26, 31, RM, vars)
	assertTLCActionFrame(t, stackFrames[1], 26, 26, RM, vars)
	assertTLCActionFrame(t, stackFrames[0], 26, 26, RM, vars)

	// Fourth frame.
	stackFrames = h.stepIn(2)
	debuggerAssertEqual(t, 5, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[2], 26, 31, RM, vars)
	assertTLCActionFrame(t, stackFrames[1], 26, 26, RM, vars)
	assertTLCActionFrame(t, stackFrames[0], 27, 27, RM, vars)

	// Debug the SendMsg action of the next-state relation.
	h.replaceBreakpoints(RM, 46)
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 6, len(stackFrames))
	context := tlc.EmptyContext.Cons(nil, tlc.IntZero).Cons(nil, tlc.IntZero)
	/*
	  /\ active[i]
	  /\ \E j \in Nodes \ {i} :
	        /\ active' = [active EXCEPT ![j] = TRUE]
	        /\ color' = [color EXCEPT ![i] = IF j>i THEN "black" ELSE @]
	  /\ UNCHANGED <<tpos, tcolor>>
	*/
	assertTLCActionFrame(t, stackFrames[3], 44, 48, RM, context, vars)
	/*
	  /\ active[i]
	*/
	assertTLCActionFrame(t, stackFrames[2], 44, 44, RM, context, vars)
	/*
	  /\ \E j \in Nodes \ {i} :
	        /\ active' = [active EXCEPT ![j] = TRUE]
	        /\ color' = [color EXCEPT ![i] = IF j>i THEN "black" ELSE @]
	*/
	assertTLCActionFrame(t, stackFrames[1], 45, 47, RM, context, vars)
	/*
	   /\ active' = [active EXCEPT ![j] = TRUE]
	   /\ color' = [color EXCEPT ![i] = IF j>i THEN "black" ELSE @]
	*/
	context = context.Cons(nil, tlc.IntOne)
	assertTLCActionFrame(t, stackFrames[0], 46, 47, RM, context, vars)

	/*
		/\ active' = [active EXCEPT ![j] = TRUE]
	*/
	stackFrames = h.stepIn()
	debuggerAssertEqual(t, 7, len(stackFrames))
	context = tlc.EmptyContext.Cons(nil, tlc.IntZero).Cons(nil, tlc.IntZero)
	assertTLCActionFrame(t, stackFrames[4], 44, 48, RM, context, vars)
	assertTLCActionFrame(t, stackFrames[3], 44, 44, RM, context, vars)
	assertTLCActionFrame(t, stackFrames[2], 45, 47, RM, context, vars)
	context = context.Cons(nil, tlc.IntOne)
	assertTLCActionFrame(t, stackFrames[1], 46, 47, RM, context, vars)
	assertTLCActionFrame(t, stackFrames[0], 46, 46, RM, context, vars)

	/*
		[active EXCEPT ![j] = TRUE]
		The breakpoint on this line (46) means that step in/out/over
		takes precedence.
	*/
	stackFrames = h.stepIn()
	debuggerAssertEqual(t, 8, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[0], 46, 46, RM, context, vars)
	/*
	   /\ color' = [color EXCEPT ![i] = IF j>i THEN "black" ELSE @]
	*/
	stackFrames = h.stepIn(5)
	debuggerAssertEqual(t, 9, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[0], 47, 47, RM, context, vars[0], vars[2], vars[3])

	/*
		/\ UNCHANGED <<tpos, tcolor>>
	*/
	stackFrames = h.stepIn(7)
	debuggerAssertEqual(t, 9, len(stackFrames))
	context = tlc.EmptyContext.Cons(nil, tlc.IntZero).Cons(nil, tlc.IntZero)
	assertTLCActionFrame(t, stackFrames[0], 48, 48, RM, context, vars[0], vars[2])

	// 8888888888888888888 State Constraint 8888888888888888888 //
	h.replaceBreakpoints(MDL, 16)
	stackFrames = h.continueFrames()
	stackFrames = h.stepIn(13)
	debuggerAssertEqual(t, 13, len(stackFrames))
	assertTLCStateFrame(t, stackFrames[0], 16, 54, 16, 64, MDL, tlc.EmptyContext.Cons(nil, tlc.IntZero))
	contextVariables := stackFrames[0].GetVariables(stackFrames[0].Base.ContextID, nil)
	debuggerAssertTrue(t, contextVariables != nil)
	debuggerAssertEqual(t, 1, len(contextVariables))
	variable := contextVariables[0]
	debuggerAssertEqual(t, "node", variable.Name)
	debuggerAssertEqual(t, "IntValue: "+tlc.IntZero.KindString(), variable.Type)
	debuggerAssertEqual(t, "0", variable.Value)

	// 8888888888888888888 Action Constraint 8888888888888888888 //
	h.replaceBreakpoints(MDL, 19)
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 10, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[0], 19, 21, MDL)

	// 8888888888888888888 Invariant Inv 8888888888888888888 //
	h.replaceBreakpoints(RM, 94)
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 8, len(stackFrames))
	assertTLCStateFrame(t, stackFrames[0], 94, 3, 96, 26, RM, tlc.EmptyContext)

	// 8888888888888888888 Spec Breakpoint 8888888888888888888 //
	h.unsetBreakpoints()
	h.setSpecBreakpoint("TLCGet(\"level\") > 3")
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 5, len(stackFrames))

	for j := 1; j < 5; j++ {
		debuggerAssertTrue(t, stackFrames[j].Synthetic != nil)
		debuggerAssertEqual(t, 5-j, stackFrames[j].Synthetic.GetS().Level())
	}

	// 8888888888888888888 ALIAS Alias 8888888888888888888 //

	h.replaceBreakpoints(MDL, 33)

	// Continue to when TLC finds a violation of the Stop invariant.
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 15, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[0], 31, 18, 31, 23, RM, tlc.EmptyContext)
	actionStackFrame := stackFrames[0].Action
	debuggerAssertTrue(t, tlc.IsInvariantViolatedException(actionStackFrame.Exception))
	debuggerAssertEqual(t, "Invariant Stop is violated.", *actionStackFrame.Exception.(interface{ GetMessage() *string }).GetMessage())

	for i := 0; i < 5; i++ {
		stackFrames = h.continueFrames()
		// i is number of states in the trace.
		debuggerAssertEqual(t, 3+i, len(stackFrames))
		assertTLCActionFrame(t, stackFrames[0], 33, 20, 33, 49, MDL)
		assertTLCActionFrame(t, stackFrames[1], 28, 5, 34, 5, MDL)
	}
	stackFrames = h.continueFrames()
	debuggerAssertEqual(t, 7, len(stackFrames))
	assertTLCActionFrame(t, stackFrames[0], 33, 20, 33, 49, MDL)
	assertTLCActionFrame(t, stackFrames[1], 28, 5, 34, 5, MDL)

	// Remove all breakpoints and run the spec to completion.
	h.unsetBreakpoints()
	h.continueFrames()
}
