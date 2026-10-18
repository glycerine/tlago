package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original ContinueTest.testSpec with three workers and source -continue.
func TestJavaContinue(t *testing.T) {
	r := runJavaTLCModelTestWithWorkers(t, "Continue", true, true, 3, "-continue", "-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ContinueTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "32", "29", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "29")
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Fatal("TE spec generated")
	}
	// Source checks concatenated length, not an ordering of concurrent traces.
	if n := len(javaTLCRecords(r, tlc.ECTLCStatePrint2)); n != 32 {
		t.Fatalf("trace states=%d, want32", n)
	}
	requireJavaTLCUncovered(t, r)
}

// Original EmptyTest.testSpec, including its explicit extra -coverage1 option.
func TestJavaEmpty(t *testing.T) {
	r := runJavaTLCModelTest(t, "Empty", "-coverage", "1", "-dumpTrace", "json", filepath.Join(t.TempDir(), "EmptyTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
}
