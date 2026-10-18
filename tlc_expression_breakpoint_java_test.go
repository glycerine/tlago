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
