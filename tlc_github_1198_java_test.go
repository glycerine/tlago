package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Github1198aTest.testSpec and doCoverage override.
func TestJavaGithub1198A(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198a.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198aTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198bTest.testSpec and doCoverage override.
func TestJavaGithub1198B(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198b.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198bTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198cTest.testSpec and doCoverage override.
func TestJavaGithub1198C(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198c.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198cTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198dTest.testSpec and doCoverage override.
func TestJavaGithub1198D(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198d.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198dTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) != 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198fTest.testSpec and doCoverage override.
func TestJavaGithub1198F(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198f.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198fTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) != 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198hTest.testSpec and doCoverage override.
func TestJavaGithub1198H(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198h.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198hTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) != 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198eTest.testSpec. Upstream ignores the unsupported
// S /\ L1 /\ L2 => L1 /\ L2 warning-composition case.
func TestJavaGithub1198EIgnored(t *testing.T) {
	t.Skip("upstream @Ignore: S /\\ L1 /\\ L2 => L1 /\\ L2 is not yet supported because composing the warning requires some thought, especially reporting correct locations.")
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1,
		"-config", "Github1198e.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198eTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198gTest.testSpec. Upstream ignores the unsupported
// S /\ L1 /\ L2 => L1 case.
func TestJavaGithub1198GIgnored(t *testing.T) {
	t.Skip("upstream @Ignore: S /\\ L1 /\\ L2 => L1 is not yet supported.")
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1,
		"-config", "Github1198g.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198gTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY absent")
	}
	requireJavaTLCUncovered(t, r)
}
