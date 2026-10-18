package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

// Original LetDef1BoxedTest.testSpec with its source overrides and inherited exit.
func TestJavaLetDef1Boxed(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "LetDef1Boxed", "LetDef1Boxed", true, true, false, 1, "-config", "LetDef1Boxed.tla", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ExitStatusViolationLiveness", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCActionPropertyViolatedBehavior)) == 0 {
		t.Fatal("TLC_ACTION_PROPERTY_VIOLATED_BEHAVIOR absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = TRUE", "x = FALSE"}, false)
}

// Original LetDef1BoxedbTest.testSpec with its source overrides and inherited exit.
func TestJavaLetDef1Boxedb(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "LetDef1Boxed", "LetDef1Boxed", true, true, false, 1, "-config", "LetDef1Boxedb.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ExitStatusViolationLiveness", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCActionPropertyViolatedBehavior)) == 0 {
		t.Fatal("TLC_ACTION_PROPERTY_VIOLATED_BEHAVIOR absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = TRUE", "x = FALSE"}, false)
}

// Original LetDef1BoxedcTest.testSpec with its source overrides and inherited exit.
func TestJavaLetDef1Boxedc(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "LetDef1Boxed", "LetDef1Boxed", true, true, true, 1, "-config", "LetDef1Boxedc.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original LetDef1BoxeddTest.testSpec with its source overrides and inherited exit.
func TestJavaLetDef1Boxedd(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "LetDef1Boxed", "LetDef1Boxed", true, true, true, 1, "-config", "LetDef1Boxedd.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original LetDef2BoxedTest.testSpec with its source overrides and inherited exit.
func TestJavaLetDef2Boxed(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "LetDef2Boxed", "LetDef2Boxed", true, true, false, 1, "-config", "LetDef2Boxed.tla", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ExitStatusViolationLiveness", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCActionPropertyViolatedBehavior)) == 0 {
		t.Fatal("TLC_ACTION_PROPERTY_VIOLATED_BEHAVIOR absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = TRUE", "x = FALSE"}, false)
}

// Original LetDef2BoxedbTest.testSpec with its source overrides and inherited exit.
func TestJavaLetDef2Boxedb(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "LetDef2Boxed", "LetDef2Boxed", true, true, false, 1, "-config", "LetDef2Boxedb.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ExitStatusViolationLiveness", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCActionPropertyViolatedBehavior)) == 0 {
		t.Fatal("TLC_ACTION_PROPERTY_VIOLATED_BEHAVIOR absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = TRUE", "x = FALSE"}, false)
}

// Original LetDef2BoxedcTest.testSpec with its source overrides and inherited exit.
func TestJavaLetDef2Boxedc(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "LetDef2Boxed", "LetDef2Boxed", true, true, true, 1, "-config", "LetDef2Boxedc.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original LetDef2BoxeddTest.testSpec with its source overrides and inherited exit.
func TestJavaLetDef2Boxedd(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "LetDef2Boxed", "LetDef2Boxed", true, true, true, 1, "-config", "LetDef2Boxedd.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}
