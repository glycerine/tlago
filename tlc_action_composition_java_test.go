package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original ActionCompositionATest.testSpec and doCoverage=false.
func TestJavaActionCompositionA(t *testing.T) {
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "ActionComposition", false, true, true, 1,
		"-config", "ActionCompositionA.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionCompositionATest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "10", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
}

// Original ActionCompositionBTest.testSpec and doCoverage=false.
func TestJavaActionCompositionB(t *testing.T) {
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "ActionComposition", false, true, true, 1,
		"-config", "ActionCompositionB.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionCompositionBTest.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "10", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
	if len(javaTLCRecords(r, tlc.ECTLCInvariantViolatedBehavior)) == 0 {
		t.Fatal("TLC_INVARIANT_VIOLATED_BEHAVIOR absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 4", "x = 6"}, true)
}

// Original ActionCompositionDistributiveATest.testSpec. Upstream marks the
// method @Ignore, so retain that status while keeping the complete body visible.
func TestJavaActionCompositionDistributiveA(t *testing.T) {
	t.Skip("upstream @Ignore")
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "ActionCompositionDistributive", false, true, false, 1,
		"-config", "ActionCompositionDistributiveA.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
}

// Original ActionCompositionDistributiveBTest.testSpec. Upstream marks the
// method @Ignore, so retain that status while keeping the complete body visible.
func TestJavaActionCompositionDistributiveB(t *testing.T) {
	t.Skip("upstream @Ignore")
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "ActionCompositionDistributive", false, true, false, 1,
		"-config", "ActionCompositionDistributiveB.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
}
