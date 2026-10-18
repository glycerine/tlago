package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

// Original InvParameterizedATest.testSpec and inherited safety-violation exit.
func TestJavaInvParameterizedA(t *testing.T) {
	r := runJavaTLCModelTestWithSettings(t, "InvParameterizedTest", false, false, "-config", "InvParameterizedTest.tla", "-invlevel", "4", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "26", "7", "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "4")
	requireJavaTLCUncovered(t, r)
}

// Original InvParameterizedBTest.testSpec with the unmodified invariant text.
func TestJavaInvParameterizedB(t *testing.T) {
	r := runJavaTLCModelTestWithSettings(t, "InvParameterizedTest", false, false, "-config", "InvParameterizedTest.tla", "-inv", "~(small = 3 /\\ big = 0)", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCUncovered(t, r)
}

// Original InvParameterizedCTest.testSpec and inherited ERROR_SPEC_PARSE exit.
func TestJavaInvParameterizedC(t *testing.T) {
	r := runJavaTLCModelTestWithSettings(t, "InvParameterizedTest", false, false, "-config", "InvParameterizedTest.tla", "-inv", "garbled input to cause parse error", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
}
