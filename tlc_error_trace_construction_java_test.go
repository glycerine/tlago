package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original EmptyOrderOfSolutionsTest.testSpec and inherited constructor settings.
func TestJavaEmptyOrderOfSolutions(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "EmptyOrderOfSolutionsTest.json")}
	r := runJavaTLCModelTest(t, "EmptyOrderOfSolutions", args...)
	if r.ExitStatus != tlc.ExitStatusFailureLivenessEval {
		t.Fatalf("exit=%d, want FAILURE_LIVENESS_EVAL; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_TAUTOLOGY absent")
	}
}

// Original ErrorTraceConstructionTest.testSpec, constructor and runner settings.
func TestJavaErrorTraceConstruction(t *testing.T) { runJavaErrorTraceConstruction(t) }
func runJavaErrorTraceConstruction(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "ErrorTraceConstructionTest.json")}
	r := runJavaTLCModelTest(t, "ErrorTraceConstructionMC", append(args, extraArgs...)...)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "8", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_14411937652505000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 288, 128)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ x = 0\n/\\ y = 0", "/\\ x = 0\n/\\ y = 1", "/\\ x = 0\n/\\ y = 2", "/\\ x = 0\n/\\ y = 3", "/\\ x = 0\n/\\ y = 4", "/\\ x = 1\n/\\ y = 5", "/\\ x = 0\n/\\ y = 6", "/\\ x = 0\n/\\ y = 7"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "4", "<N7 line 32, col 7 to line 34, col 15 of module ErrorTraceConstruction>")
	requireJavaTLCUncovered(t, r)
	return r
}

// Original ErrorTraceConstructionTest_TTraceTest.testSpec rechecks the generated artifact.
func TestJavaErrorTraceConstructionTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "ErrorTraceConstructionTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaErrorTraceConstruction(t, "-teSpecOutDir", generated) }) {
		return
	}
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTTraceRecheck(t, "ErrorTraceConstructionMC", generated, false, true)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "8", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 240, 128)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ x = 0\n/\\ y = 0", "/\\ x = 0\n/\\ y = 1", "/\\ x = 0\n/\\ y = 2", "/\\ x = 0\n/\\ y = 3", "/\\ x = 0\n/\\ y = 4", "/\\ x = 1\n/\\ y = 5", "/\\ x = 0\n/\\ y = 6", "/\\ x = 0\n/\\ y = 7"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "4", "<_next line 32, col 5 to line 40, col 29 of module ErrorTraceConstructionTestTTrace>")
	requireJavaTLCUncovered(t, r)
}
