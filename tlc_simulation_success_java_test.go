package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Whole SuccessfulSimulationTestCase.testSpec inherited by each concrete class.
// Preserve the repeated STATE_PRINT2 assertion in the source method.
func requireJavaSuccessfulSimulation(t *testing.T, r *tlc.Result) {
	t.Helper()
	if len(javaTLCRecords(r, tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Fatal("A TE spec was generated, but it shouldn't")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCTemporalPropertyViolated, tlc.ECTLCCounterExample, tlc.ECTLCStatePrint2, tlc.ECTLCStatePrint3, tlc.ECTLCStatePrint2} {
		if len(javaTLCRecords(r, code)) != 0 {
			t.Fatalf("unexpected recorded code %d", code)
		}
	}
}

// Original SimulationTest2.testSpec, inherited from SuccessfulSimulationTestCase.
func TestJavaSimulationTest2(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Test2", "Test2", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "SimulationTest2.json"),
		"-simulate", "num=50", "-depth", "6")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	requireJavaSuccessfulSimulation(t, r)
}

// Original LiveCheckSimulationTest2 static initializer selects experimental
// liveness; retain all inherited runner flags and the source depth of ten.
func TestJavaLiveCheckSimulationTest2(t *testing.T) {
	r := runJavaTLCModelTestWithRootAndSetup(t, "Test2", "Test2", true, true, true, 1,
		func(*tlc.TLC) { tlc.SetSimulatorExperimentalLiveness(true) },
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "LiveCheckSimulationTest2.json"),
		"-simulate", "num=50", "-depth", "10")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	requireJavaSuccessfulSimulation(t, r)
}

// Whole SimulationTest2PostCondition.testSpec includes super.testSpec().
func TestJavaSimulationTest2PostCondition(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Test2", "Test2", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "SimulationTest2PostCondition.json"),
		"-config", "Test2PostCondition.cfg", "-simulate", "num=1", "-depth", "6")
	if r.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit=%d, want VIOLATION_ASSUMPTION", r.ExitStatus)
	}
	requireJavaSuccessfulSimulation(t, r)
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) == 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE absent")
	}
}

// Whole SimulationTestAssumption.testSpec overrides (does not call) the parent.
func TestJavaSimulationTestAssumption(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Test4", "Test4", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "SimulationTestAssumption.json"),
		"-simulate", "num=1")
	if r.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit=%d, want VIOLATION_ASSUMPTION", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCAssumptionFalse)) == 0 {
		t.Fatal("TLC_ASSUMPTION_FALSE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
}
