package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original TSnapShotTest.testSpec with its required four workers.
func TestJavaTSnapShot(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "TSnapShot", "MC", true, true, true, 4, "-dumpTrace", "json", filepath.Join(t.TempDir(), "TSnapShotTest.json"))
	if r.ExitStatus != tlc.ExitStatusFailureSafetyEval {
		t.Fatalf("exit=%d, want FAILURE_SAFETY_EVAL", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
}
