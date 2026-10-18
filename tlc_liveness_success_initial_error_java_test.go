package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

// Original Github1037Test.testSpec and constructor settings.
func TestJavaGithub1037(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github1037", "Github1037", false, false, false, 1, "-noGenerateSpecTE")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// Original Github790Test.testSpec and constructor settings.
func TestJavaGithub790(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github790", "Github790", false, false, false, 1, "-noGenerateSpecTE")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// Original InitialLivenessEvaluationErrorTest.testSpec and constructor settings.
func TestJavaInitialLivenessEvaluationError(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "InitialLivenessEvaluationError", "InitialLivenessEvaluationError", false, false, false, 1, "-noGenerateSpecTE")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse && r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("unexpected exit=%d; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) != 0 {
		t.Fatal("TLC_SUCCESS present")
	}
	if r.ExitStatus == tlc.ExitStatusViolationSafety {
		requireJavaTLCRecordedParams(t, r, tlc.ECTLCInvariantViolatedBehavior, "Inv")
		if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
			t.Fatal("TLC_STATE_PRINT2 absent")
		}
		requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1"}, false)
	} else {
		if len(javaTLCRecords(r, tlc.ECTLCInitialState)) == 0 && len(javaTLCRecords(r, tlc.ECTLCStateNotCompletelySpecifiedLive)) == 0 {
			t.Fatal("initial liveness evaluation diagnostic absent")
		}
		if len(javaTLCRecords(r, tlc.ECTLCInvariantViolatedBehavior)) != 0 {
			t.Fatal("TLC_INVARIANT_VIOLATED_BEHAVIOR present")
		}
		requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "1")
	}
}

// Original InitialLivenessEvaluationErrorInvariantOnlyTest.testSpec and constructor settings.
func TestJavaInitialLivenessEvaluationErrorInvariantOnly(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "InitialLivenessEvaluationErrorInvariantOnly", "InitialLivenessEvaluationError", false, false, false, 1, "-noGenerateSpecTE", "-config", "InitialLivenessEvaluationErrorInvariantOnly.cfg")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("unexpected exit=%d; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) != 0 {
		t.Fatal("TLC_SUCCESS present")
	}
	if r.ExitStatus == tlc.ExitStatusViolationSafety {
		requireJavaTLCRecordedParams(t, r, tlc.ECTLCInvariantViolatedBehavior, "Inv")
		if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
			t.Fatal("TLC_STATE_PRINT2 absent")
		}
		requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1"}, false)
	} else {
		if len(javaTLCRecords(r, tlc.ECTLCInitialState)) == 0 && len(javaTLCRecords(r, tlc.ECTLCStateNotCompletelySpecifiedLive)) == 0 {
			t.Fatal("initial liveness evaluation diagnostic absent")
		}
		if len(javaTLCRecords(r, tlc.ECTLCInvariantViolatedBehavior)) != 0 {
			t.Fatal("TLC_INVARIANT_VIOLATED_BEHAVIOR present")
		}
		requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "1")
	}
}
