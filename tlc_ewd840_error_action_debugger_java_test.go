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
