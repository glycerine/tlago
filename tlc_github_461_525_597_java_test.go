package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"math"
	"path/filepath"
	"testing"
)

// Original Github461Test.testSpec with its complete assertion trace and call stack.
func TestJavaGithub461(t *testing.T) {
	runJavaGithub461(t)
}

func runJavaGithub461(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	args := append([]string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github461Test.json")}, extraArgs...)
	r := runJavaTLCModelTest(t, "Github461", args...)
	if r.ExitStatus != tlc.ExitStatusViolationAssert {
		t.Fatalf("exit=%d, want VIOLATION_ASSERT; messages=%v", r.ExitStatus, r.Messages)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCValueAssertFailed, `"Failure of assertion at line 8, column 4."`)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 2", "x = 3", "x = 4"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCNestedExpression, "0. Line 9, column 5 to line 10, column 17 in Github461\n1. Line 9, column 8 to line 9, column 65 in Github461\n\n")
	requireJavaTLCUncovered(t, r)
	return r
}

// Original Github525Test.testSpec with all runner defaults.
func TestJavaGithub525(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github525", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github525Test.json"))
	if r.ExitStatus != tlc.ExitStatusError {
		t.Fatalf("exit=%d, want ERROR; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCLiveCannotHandleFormula} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
}

// Original Github597Test.testSpec and its noRandomFPandSeed/doCoverage/doDump
// overrides. Preserve the runtime's random fp/seed by omitting both arguments.
func TestJavaGithub597(t *testing.T) {
	runJavaGithub597(t)
}

func runJavaGithub597(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	setJavaModelLivenessThreshold(t, math.MaxFloat64)
	r := runJavaTLCModelTestWithArguments(t, "Github597", "dekker", func(meta, traceDirectory string) []string {
		args := []string{"-metadir", meta, "-deadlock", "-debugger", "nosuspend,port=4712,nohalt",
			"-generateSpecTE", "-teSpecOutDir", traceDirectory, "-workers", "1", "-checkpoint", "0",
			"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github597Test.json"), "-config", "dekker.tla"}
		return append(args, extraArgs...)
	})
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "4356", "1500", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "Termination")
	for _, code := range []int{tlc.ECTLCCounterExample, tlc.ECTLCStatePrint2, tlc.ECTLCBackToState} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
	return r
}
