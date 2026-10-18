package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original Github798ITest.testSpec with noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub798I(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github798I", "-noGenerateSpecTE", "-config", "Github798I.tla")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
}

// Original Github798NTest.testSpec with noGenerateSpec=true and doDumpTrace=false.
func TestJavaGithub798N(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github798N", "-noGenerateSpecTE", "-config", "Github798N.tla")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
}

// Original Github807Test.testSpec, including exact actions and all overrides.
func TestJavaGithub807(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github807", "Github807", true, true, false, 1,
		"-noGenerateSpecTE", "-config", "Github807.tla")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 2"}, false)
	actions := []string{tlc.InitialPredicate, "<Add line 13, col 8 to line 13, col 15 of module Github807>", "<Add line 13, col 8 to line 13, col 15 of module Github807>"}
	for i, record := range javaTLCRecords(r, tlc.ECTLCStatePrint2) {
		action, ok := record.StateInfo.Info.(string)
		if !ok || action != actions[i] {
			t.Fatalf("trace action %d=%v, want %s", i+1, record.StateInfo.Info, actions[i])
		}
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}
