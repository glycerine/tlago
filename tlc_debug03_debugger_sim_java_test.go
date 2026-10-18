package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaDebug03DebuggerSim(t *testing.T) {
	const RM = "Debug03"
	h := startJavaDebuggerModel(t, "debug", RM, tlc.ExitStatusSuccess,
		"-config", "Debug03.tla", "-simulate", "num=2")
	defer h.close()
	h.setSpecBreakpoint()

	stackFrames := h.continueFrames()
	debuggerAssertEqual(t, 1, len(stackFrames))
	assertTLCInitStatesFrame(t, stackFrames[0], 7, 5, 7, 14, RM, tlc.EmptyContext, 2)
	init := stackFrames[0].Init
	debuggerAssertEqual(t, 2, init.GetStates().Size())

	stackFrames = h.continueFrames()

	// Construct a trace x=0 .. x=8 by selecting each successor state in turn.
	for i := 0; i < 8; i++ {
		// i is number of states in the trace.
		debuggerAssertEqual(t, 2+i, len(stackFrames))
		assertTLCNextStatesFrame(t, stackFrames[0], 14, 16, 14, 19, RM, tlc.EmptyContext, 9)

		next := stackFrames[0].Next

		// Check that the TLC state variable 'x' has the expected value.
		debuggerAssertEqual(t, int32(i), next.GetS().Values().Get(tlc.UniqueStringOf("x")).(*tlc.IntValue).Val)

		// Assert that the manually pushed frames represent the constructed trace.
		for j := 0; j < len(stackFrames)-1; j++ {
			frame := stackFrames[len(stackFrames)-1-j]
			debuggerAssertTrue(t, frame.Synthetic != nil)
			f := frame.Synthetic
			debuggerAssertEqual(t, int32(j), f.GetS().Values().Get(tlc.UniqueStringOf("x")).(*tlc.IntValue).Val)
		}

		// Select the i-th successor state, which also continues the debugger.
		h.gotoState(new(tlc.GotoStateArgument).SetVariablesReference(next.GetSuccessorVariables(nil)[i].VariablesReference))
		stackFrames = h.frames()
	}

	// Reverse direction: back to the initial state.
	for i := 8; i > 0; i-- {
		// i is number of states in the trace.
		debuggerAssertEqual(t, 2+i, len(stackFrames))
		assertTLCNextStatesFrame(t, stackFrames[0], 14, 16, 14, 19, RM, tlc.EmptyContext, 9)

		next := stackFrames[0].Next

		// Check that the TLC state variable 'x' has the expected value.
		debuggerAssertEqual(t, int32(i), next.GetS().Values().Get(tlc.UniqueStringOf("x")).(*tlc.IntValue).Val)

		// Assert that the manually pushed frames represent the deconstructed trace.
		for j := 0; j < len(stackFrames)-1; j++ {
			frame := stackFrames[len(stackFrames)-1-j]
			debuggerAssertTrue(t, frame.Synthetic != nil)
			f := frame.Synthetic
			debuggerAssertEqual(t, int32(j), f.GetS().Values().Get(tlc.UniqueStringOf("x")).(*tlc.IntValue).Val)
		}

		// Select the predecessor state, which also continues the debugger.
		h.gotoState(new(tlc.GotoStateArgument).SetVariablesReference(next.TraceVariables(nil)[1].VariablesReference))
		stackFrames = h.frames()
	}

	// Step out of the first next-state frame to the two initial states.
	stackFrames = h.stepOut()
	assertTLCInitStatesFrame(t, stackFrames[0], 7, 5, 7, 14, RM, tlc.EmptyContext, 1)
	init = stackFrames[0].Init
	debuggerAssertEqual(t, 2, init.GetStates().Size())

	h.gotoState(new(tlc.GotoStateArgument).SetVariablesReference(init.GetStateVariables(nil)[1].VariablesReference))
	stackFrames = h.frames()
	debuggerAssertEqual(t, 2, len(stackFrames))
	assertTLCNextStatesFrame(t, stackFrames[0], 14, 16, 14, 19, RM, tlc.EmptyContext, 9)
	next := stackFrames[0].Next
	debuggerAssertEqual(t, int32(1), next.GetS().Values().Get(tlc.UniqueStringOf("x")).(*tlc.IntValue).Val)

	// Remove all breakpoints and run the spec to completion.
	h.unsetBreakpoints()
	h.continueFrames()
}
