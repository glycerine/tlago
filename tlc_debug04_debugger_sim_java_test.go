/*******************************************************************************
 * Copyright (c) 2025 NVIDIA Corp. All rights reserved.
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
// Translation of the complete Debug04SimTest.testSpec after source comparison.
package tlago

import (
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaDebug04DebuggerSim(t *testing.T) {
	const RM = "Debug04"
	h := startJavaDebuggerModel(t, "debug", RM, tlc.ExitStatusSuccess,
		"-config", "Debug04.tla", "-simulate", "num=4")
	defer h.close()
	expression := func(s string) *string { return &s }
	hasSuccessorB := func(next *tlc.TLCNextStatesStackFrame) bool {
		for _, s := range next.GetSuccessors().ToSlice() {
			equal, err := tlc.NewStringValue("B").Equal(s.Values().Get(tlc.UniqueStringOf("x")))
			if err != nil {
				t.Fatal(err)
			}
			if equal {
				return true
			}
		}
		return false
	}
	h.setSpecBreakpoint()

	// 88888888888888888888888888888 Step In 8888888888888888888888888888 //

	stackFrames := h.continueFrames()
	debuggerAssertEqual(t, 1, len(stackFrames))
	assertTLCInitStatesFrame(t, stackFrames[0], 6, 9, 6, 29, RM, tlc.EmptyContext, 3)
	init := stackFrames[0].Init
	debuggerAssertEqual(t, 3, init.GetStates().Size())

	// Debug Expressions initial frame (no action level)
	ea := tlc.TLCEvaluateArguments{}
	ea.FrameID = &stackFrames[0].Base.ID
	ea.Context = "repl"
	ea.Expression = expression("x")                                          // State-level
	debuggerAssertTrue(t, strings.HasPrefix(*h.debugger.Evaluate(ea).Result, // Only prefix bc the name of the on-the-fly
		// generated debug module is not known.
		"In evaluation, the identifier x is either undefined or not an operator."))
	ea.Expression = expression("x'") // Action-level
	debuggerAssertTrue(t, strings.HasPrefix(*h.debugger.Evaluate(ea).Result,
		"In evaluation, the identifier x is either undefined or not an operator."))

	// Goto the first A state (rely on the fact that states are sorted
	// lexicographically, making A the first state. However, assert it).
	statesVariables := init.GetStateVariables(nil)
	aState := statesVariables[0]
	debuggerAssertEqual(t, tlc.NewStringValue("A"), aState.TLCValue.(*tlc.RecordValue).Values[0])
	h.gotoState(new(tlc.GotoStateArgument).SetVariablesReference(aState.VariablesReference))
	stackFrames = h.frames()

	// idempotence check
	debuggerAssertEqual(t, stackFrames, h.frames())

	// Construct a trace x=A, x=A, ..., x=A by stepping into.
	for i := 0; i < 8; i++ {

		// i is number of states in the trace.
		debuggerAssertEqual(t, 2+i, len(stackFrames))
		assertTLCNextStatesFrame(t, stackFrames[0], 16, 20, 16, 23, RM, tlc.EmptyContext, 3)

		next := stackFrames[0].Next
		debuggerAssertTrue(t, next.GetS().AllAssigned())

		// Check that the TLC state variable 'x' has the expected value.
		debuggerAssertEqual(t, tlc.NewStringValue("A"), next.GetS().Values().Get(tlc.UniqueStringOf("x")))

		// Debug Expressions next frame (no action level)
		ea = tlc.TLCEvaluateArguments{}
		ea.FrameID = &stackFrames[0].Base.ID
		ea.Context = "repl"
		ea.Expression = expression("x") // State-level
		debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)
		ea.Expression = expression("x'") // Action-level
		debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)
		// assert that watch and repl are equivalent.
		ea.Context = "watch"
		ea.Expression = expression("x") // State-level
		debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)
		ea.Expression = expression("x'") // Action-level
		debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)

		ea.Expression = expression("TLCGet(\"level\")")
		debuggerAssertEqual(t, strconv.Itoa(i+1), *h.debugger.Evaluate(ea).Result)

		// Assert that the stack frames that have been manually pushed onto the
		// debugger's stack, represent the correct TLC states, i.e., the trace that is
		// constructed.
		for j := 0; j < len(stackFrames)-1; j++ {
			frame := stackFrames[len(stackFrames)-1-j]

			debuggerAssertTrue(t, frame.Synthetic != nil)
			f := frame.Synthetic
			debuggerAssertTrue(t, f.GetS().AllAssigned())

			// Check that the TLC state variable 'x' has the expected value.
			debuggerAssertEqual(t, tlc.NewStringValue("A"), f.GetS().Values().Get(tlc.UniqueStringOf("x")))
		}

		stackFrames = h.stepIn()
	}

	// Reverse direction: back to the initial state.
	for i := 8; i > 0; i-- {

		// i is number of states in the trace.
		debuggerAssertEqual(t, 2+i, len(stackFrames))
		assertTLCNextStatesFrame(t, stackFrames[0], 16, 20, 16, 23, RM, tlc.EmptyContext, 3)

		next := stackFrames[0].Next
		debuggerAssertTrue(t, next.GetS().AllAssigned())

		// Check that the TLC state variable 'x' has the expected value.
		debuggerAssertEqual(t, tlc.NewStringValue("A"), next.GetS().Values().Get(tlc.UniqueStringOf("x")))

		// Debug Expressions next frame (no action level)
		ea = tlc.TLCEvaluateArguments{}
		ea.FrameID = &stackFrames[0].Base.ID
		ea.Context = "repl"
		ea.Expression = expression("x") // State-level
		debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)
		ea.Expression = expression("x'") // Action-level
		debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)
		// assert that watch and repl are equivalent.
		ea.Context = "watch"
		debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)
		ea.Expression = expression("x'") // Action-level
		debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)

		ea.Expression = expression("TLCGet(\"level\")")
		debuggerAssertEqual(t, strconv.Itoa(i+1), *h.debugger.Evaluate(ea).Result)

		// Assert that the stack frames that have been manually pushed onto the
		// debugger's stack, represent the correct TLC states, i.e., the trace that is
		// deconstructed.
		for j := 0; j < len(stackFrames)-1; j++ {
			frame := stackFrames[len(stackFrames)-1-j]

			debuggerAssertTrue(t, frame.Synthetic != nil)
			f := frame.Synthetic
			debuggerAssertTrue(t, f.GetS().AllAssigned())

			// Check that the TLC state variable 'x' has the expected value.
			debuggerAssertEqual(t, tlc.NewStringValue("A"), f.GetS().Values().Get(tlc.UniqueStringOf("x")))

			// Debug Expressions (action level)
			ea = tlc.TLCEvaluateArguments{}
			ea.FrameID = &frame.Base.ID
			ea.Context = "repl"
			ea.Expression = expression("x") // State-level
			debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)
			ea.Expression = expression("x'") // Action-level
			debuggerAssertEqual(t, "A", *h.debugger.Evaluate(ea).Result)
		}

		// Select the predecessor state to continue the simulation back.
		// This also continues_ the debugger.
		stackFrames = h.stepOut()
	}

	// 88888888888888888888888888888 Step Over 8888888888888888888888888888 //

	// Construct a trace x=A, x=B, ..., x=A (no variable values are identical in
	// consecutive states) by stepping over.

	stackFrames = h.stepOut()
	debuggerAssertEqual(t, 1, len(stackFrames))
	assertTLCInitStatesFrame(t, stackFrames[0], 6, 9, 6, 29, RM, tlc.EmptyContext, 3)
	init = stackFrames[0].Init
	debuggerAssertEqual(t, 3, init.GetStates().Size())

	// Debug Expressions initial frame (no action level)
	ea = tlc.TLCEvaluateArguments{}
	ea.FrameID = &stackFrames[0].Base.ID
	ea.Context = "repl"
	ea.Expression = expression("x") // State-level
	debuggerAssertTrue(t, strings.HasPrefix(*h.debugger.Evaluate(ea).Result,
		"In evaluation, the identifier x is either undefined or not an operator."))
	ea.Expression = expression("x'") // Action-level
	debuggerAssertTrue(t, strings.HasPrefix(*h.debugger.Evaluate(ea).Result,
		"In evaluation, the identifier x is either undefined or not an operator."))
	// assert that watch and repl are equivalent.
	ea.Context = "watch"
	ea.Expression = expression("x") // State-level
	debuggerAssertTrue(t, strings.HasPrefix(*h.debugger.Evaluate(ea).Result,
		"In evaluation, the identifier x is either undefined or not an operator."))
	ea.Expression = expression("x'") // Action-level
	debuggerAssertTrue(t, strings.HasPrefix(*h.debugger.Evaluate(ea).Result,
		"In evaluation, the identifier x is either undefined or not an operator."))

	ea.Expression = expression("TLCGet(\"level\")")
	debuggerAssertEqual(t, strconv.Itoa(1), *h.debugger.Evaluate(ea).Result)

	h.gotoState(new(tlc.GotoStateArgument).SetVariablesReference(init.GetStateVariables(nil)[0].VariablesReference))
	stackFrames = h.frames()

	// idempotence check
	debuggerAssertEqual(t, stackFrames, h.frames())

	debuggerAssertTrue(t, stackFrames[0].State != nil)
	var oldVal tlc.Value

	for i := 0; i < 8; i++ {

		// i is number of states in the trace.
		debuggerAssertEqual(t, 2+i, len(stackFrames))
		assertTLCNextStatesFrame(t, stackFrames[0], 16, 20, 16, 23, RM, tlc.EmptyContext, 3)

		next := stackFrames[0].Next
		debuggerAssertTrue(t, next.GetS().AllAssigned())

		// Check that the TLC state variable 'x' has the expected value.
		debuggerAssertNotEqual(t, oldVal, next.GetS().Values().Get(tlc.UniqueStringOf("x")))
		oldVal = stackFrames[0].State.GetS().Values().Get(tlc.UniqueStringOf("x"))

		// Debug Expressions next frame (no action level)
		ea = tlc.TLCEvaluateArguments{}
		ea.FrameID = &stackFrames[0].Base.ID
		ea.Context = "repl"
		ea.Expression = expression("x") // State-level
		debuggerAssertEqual(t, oldVal.(*tlc.StringValue).Val.String(), *h.debugger.Evaluate(ea).Result)
		ea.Expression = expression("x'") // Action-level
		debuggerAssertEqual(t, oldVal.(*tlc.StringValue).Val.String(), *h.debugger.Evaluate(ea).Result)

		// assert that watch and repl are equivalent.
		ea.Context = "watch"
		ea.Expression = expression("x") // State-level
		debuggerAssertEqual(t, oldVal.(*tlc.StringValue).Val.String(), *h.debugger.Evaluate(ea).Result)
		ea.Expression = expression("x'") // Action-level
		debuggerAssertEqual(t, oldVal.(*tlc.StringValue).Val.String(), *h.debugger.Evaluate(ea).Result)

		ea.Expression = expression("TLCGet(\"level\")")
		debuggerAssertEqual(t, strconv.Itoa(i+1), *h.debugger.Evaluate(ea).Result)

		stackFrames = h.next()
	}

	// Back to the initial states
	stackFrames = h.stepOut(8)

	// 8888888888888888888888888 Action-Level Breakpoint Condition 888888888888888888888 //

	h.setSpecBreakpoint("x = \"A\" /\\ x' = \"B\"")
	stackFrames = h.continueFrames()
	assertTLCNextStatesFrame(t, stackFrames[0], 16, 20, 16, 23, RM, tlc.EmptyContext, 3)
	next := stackFrames[0].Next
	debuggerAssertTrue(t, next.GetS().AllAssigned())

	// Check that the TLC state variable 'x' has the expected value.
	debuggerAssertEqual(t, tlc.NewStringValue("A"), next.GetS().Values().Get(tlc.UniqueStringOf("x")))
	debuggerAssertTrue(t, hasSuccessorB(next))

	// 88888888888888888 ENABLED Next = FALSE with condition 88888888888888 //
	h.setSpecBreakpoint("~ENABLED Next")
	stackFrames = h.continueFrames()
	assertTLCNextStatesFrame(t, stackFrames[0], 16, 20, 16, 23, RM, tlc.EmptyContext, 0)

	// 88888888888888888 ENABLED Next = FALSE with no condition 88888888888888 //
	h.unsetBreakpoints()
	h.setSpecBreakpoint()
	// "manually" hit continue until we reach ~ENABLED Next. Compare Debug04.tla
	// where level is set to 50 (+1 for the init state).
	stackFrames = h.continueFrames(51)
	assertTLCNextStatesFrame(t, stackFrames[0], 16, 20, 16, 23, RM, tlc.EmptyContext, 0)

	// 88888888888888888 Run to next set of initial states 88888888888888 //
	h.unsetBreakpoints()
	h.setSpecBreakpoint("TLCGet(\"level\") = 1")
	stackFrames = h.continueFrames()
	assertTLCInitStatesFrame(t, stackFrames[0], 6, 9, 6, 29, RM, tlc.EmptyContext, 3)

	h.unsetBreakpoints()
	h.setSpecBreakpoint("TLCGet(\"level\") = 2")
	stackFrames = h.continueFrames()
	assertTLCNextStatesFrame(t, stackFrames[0], 16, 20, 16, 23, RM, tlc.EmptyContext, 3)
	next = stackFrames[0].Next
	debuggerAssertEqual(t, 2, next.GetS().Level())

	h.unsetBreakpoints()
	h.setSpecBreakpoint("TLCGet(\"level\") = 3")
	stackFrames = h.continueFrames()
	assertTLCNextStatesFrame(t, stackFrames[0], 16, 20, 16, 23, RM, tlc.EmptyContext, 3)
	next = stackFrames[0].Next
	debuggerAssertEqual(t, 3, next.GetS().Level())

	h.unsetBreakpoints()
	h.setSpecBreakpoint("TLCGet(\"level\") = 4")
	stackFrames = h.continueFrames()
	assertTLCNextStatesFrame(t, stackFrames[0], 16, 20, 16, 23, RM, tlc.EmptyContext, 3)
	next = stackFrames[0].Next
	debuggerAssertEqual(t, 4, next.GetS().Level())

	// 88888888888888888 FALSE should never fire 88888888888888 //
	h.unsetBreakpoints()
	h.setSpecBreakpoint("FALSE")
	h.continueFrames()
}
