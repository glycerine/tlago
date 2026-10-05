package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Whole original tlc2.tool.SimulatorTest.testPrintBehaviorShouldPrintErrorState.
func TestJavaSimulatorPrintBehaviorShouldPrintErrorState(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "Github726", "Github726"))
	if err != nil {
		t.Fatal(err)
	}
	tool := javaSimulationTool(t, "Github726", root, root)
	recorder := &tlc.MemoryRecorder{}
	tlc.AddMessageRecorder(recorder)
	defer tlc.RemoveMessageRecorder(recorder)
	simulator := tlc.NewSimulator(tool, false, -1, 0, 0,
		tlc.WithSimulatorMetaDir(""), tlc.WithSimulatorTraceFile(""),
		tlc.WithSimulatorTraceActions(""), tlc.WithSimulatorAril(0),
		tlc.WithSimulatorWorkerCount(0))
	stateVec := tlc.NewStateVec(1)
	stateVec.Add(tlc.EmptyState.CreateEmpty())
	simulator.PrintBehaviorException(tlc.NewTLCRuntimeExceptionMessage("TestException"), stateVec)
	if !recorder.Recorded(tlc.ECTLCErrorState) {
		t.Fatal("original assertTrue: recorder.recorded(TLC_ERROR_STATE)")
	}
}
