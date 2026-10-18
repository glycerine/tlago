package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Test207: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest207(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest207", "test207", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test207.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test208: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest208(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest208", "test208", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test208.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test209: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest209(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest209", "test209", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test209.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test219: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest219(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest219", "test219", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test219.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}
