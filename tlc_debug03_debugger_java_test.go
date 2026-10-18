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
