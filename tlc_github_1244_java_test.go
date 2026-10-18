package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

// Original Github1244Test.testSpec and source runner overrides.
func TestJavaGithub1244(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1244", "Github1244", true, true, true, 1, "-config", "Github1244.tla", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original Github1244bTest.testSpec and source runner overrides.
func TestJavaGithub1244B(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1244", "Github1244", true, true, true, 1, "-config", "Github1244b.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original Github1244cTest.testSpec and source runner overrides.
func TestJavaGithub1244C(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1244", "Github1244", true, true, true, 1, "-config", "Github1244c.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}
