package tlago

import (
	"math"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Complete original SimulatorTest methods, each inherited unchanged by
// SimulatorMultiThreadTest. numWorkers() returns respectively 1 and 4.
func runJavaSimulatorCorrectness(t *testing.T, config string, deadlock bool, depth int, traces int64, workers int, continuation bool) *tlc.MemoryRecorder {
	t.Helper()
	recorder := &tlc.MemoryRecorder{}
	tlc.AddMessageRecorder(recorder)
	t.Cleanup(func() { tlc.RemoveMessageRecorder(recorder) })
	tool := javaSimulationTool(t, "BasicMultiTrace", "BasicMultiTrace", config, continuation)
	simulator := tlc.NewSimulator(tool, deadlock, depth, traces, 0, tlc.WithSimulatorWorkerCount(workers))
	if _, err := simulator.Simulate(); err != nil {
		t.Fatalf("original runSimulatorTest failed: %v", err)
	}
	return recorder
}

// TestMPRecorder.recordedWithStringValue checks parameter zero of the first
// record for the code; later records cannot satisfy this source assertion.
func javaSimulatorRecordedWithStringValue(recorder *tlc.MemoryRecorder, code int, value string) bool {
	records := recorder.Records(code)
	return len(records) > 0 && len(records[0].Params) > 0 && records[0].Params[0] == value
}

func TestJavaSimulatorCorrectnessSuccessfulSimulation(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			recorder := runJavaSimulatorCorrectness(t, "MC", false, 100, 100, context.workers, false)
			if recorder.Recorded(tlc.ECTLCInvariantViolatedInitial) {
				t.Fatal("original assertFalse failed")
			}
			if !(recorder.Recorded(tlc.ECTLCProgressSimu)) {
				t.Fatal("original assertTrue failed")
			}
			if !(recorder.Recorded(tlc.ECTLCStatsSimu)) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}

func TestJavaSimulatorCorrectnessInvariantViolationInitialState(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			recorder := runJavaSimulatorCorrectness(t, "MCInvInitState", false, 100, 100, context.workers, false)
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCInvariantViolatedInitial, "InvInitState")) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}

func TestJavaSimulatorCorrectnessInvariantViolation(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			recorder := runJavaSimulatorCorrectness(t, "MCInv", false, 100, 100, context.workers, false)
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCComputingInitProgress, "1")) {
				t.Fatal("original assertTrue failed")
			}
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCInvariantViolatedBehavior, "Inv")) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}

func TestJavaSimulatorCorrectnessInvariantBadEvalInitState(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			recorder := runJavaSimulatorCorrectness(t, "MCBadInvInitState", false, 100, 100, context.workers, false)
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCComputingInitProgress, "1")) {
				t.Fatal("original assertTrue failed")
			}
			if !(recorder.Recorded(tlc.ECTLCInitialState)) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}

func TestJavaSimulatorCorrectnessInvariantBadEvalNonInitState(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			recorder := runJavaSimulatorCorrectness(t, "MCBadInvNonInitState", false, 100, 100, context.workers, false)
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCComputingInitProgress, "1")) {
				t.Fatal("original assertTrue failed")
			}
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCInvariantEvaluationFailed, "InvBadEvalNonInitState")) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}

func TestJavaSimulatorCorrectnessUnderspecifiedInit(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			recorder := runJavaSimulatorCorrectness(t, "MCUnderspecInit", false, 100, 100, context.workers, false)
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCComputingInitProgress, "1")) {
				t.Fatal("original assertTrue failed")
			}
			if !(recorder.Recorded(tlc.ECTLCStateNotCompletelySpecifiedInitial)) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}

func TestJavaSimulatorCorrectnessInvariantViolationContinue(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			recorder := runJavaSimulatorCorrectness(t, "MCInv", false, 100, 100, context.workers, true)
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCComputingInitProgress, "1")) {
				t.Fatal("original assertTrue failed")
			}
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCInvariantViolatedBehavior, "Inv")) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}

func TestJavaSimulatorCorrectnessDontContinueOnRuntimeSpecError(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			recorder := runJavaSimulatorCorrectness(t, "MCBadInvNonInitState", false, 100, math.MaxInt64, context.workers, true)
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCComputingInitProgress, "1")) {
				t.Fatal("original assertTrue failed")
			}
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCInvariantEvaluationFailed, "InvBadEvalNonInitState")) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}

func TestJavaSimulatorCorrectnessLivenessViolation(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			tlc.FP64Init()
			recorder := runJavaSimulatorCorrectness(t, "MCLivenessProp", false, 100, 100, context.workers, false)
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCComputingInitProgress, "1")) {
				t.Fatal("original assertTrue failed")
			}
			if !(recorder.Recorded(tlc.ECTLCTemporalPropertyViolated)) {
				t.Fatal("original assertTrue failed")
			}
			if !(recorder.Recorded(tlc.ECTLCCounterExample)) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}

func TestJavaSimulatorCorrectnessLivenessViolationIgnoresContinue(t *testing.T) {
	for _, context := range []struct {
		name    string
		workers int
	}{{"SimulatorTest", 1}, {"SimulatorMultiThreadTest", 4}} {
		t.Run(context.name, func(t *testing.T) {
			tlc.FP64Init()
			recorder := runJavaSimulatorCorrectness(t, "MCLivenessProp", false, 100, math.MaxInt64, context.workers, true)
			if !(javaSimulatorRecordedWithStringValue(recorder, tlc.ECTLCComputingInitProgress, "1")) {
				t.Fatal("original assertTrue failed")
			}
			if !(recorder.Recorded(tlc.ECTLCTemporalPropertyViolated)) {
				t.Fatal("original assertTrue failed")
			}
			if !(recorder.Recorded(tlc.ECTLCCounterExample)) {
				t.Fatal("original assertTrue failed")
			}
		})
	}
}
