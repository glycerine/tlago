package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original PostAssumptionTest.testSpec and inherited assumption-violation exit.
func TestJavaPostAssumption(t *testing.T) {
	r := runJavaTLCModelTest(t, "DieHardTLA", "-config", "DieHardTLAPA", "-dumpTrace", "json", filepath.Join(t.TempDir(), "PostAssumptionTest.json"))
	if r.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit=%d, want VIOLATION_ASSUMPTION", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "97", "16", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCPostconditionFalse, "PostCondition")
	requireJavaTLCUncovered(t, r)
}

// Original MinimumDiameterTest.testSpec with source doDump=false.
func TestJavaMinimumDiameter(t *testing.T) {
	r := runJavaTLCModelTestWithSettings(t, "MinimumDiameter", true, false, "-dumpTrace", "json", filepath.Join(t.TempDir(), "MinimumDiameterTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCUncovered(t, r)
}
