package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Github1302Test.testSpec, including simulator arguments and overrides.
func TestJavaGithub1302(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1302", "Github1302", false, true, false, 1,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1302Test.json"),
		"-config", "Github1302.tla", "-simulate", "num=3", "-depth", "3")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStatsSimu, "10", "1", "0")
	requireJavaTLCUncovered(t, r)
}

// Original Github1302bTest.testSpec, retaining forced trace-spec generation.
func TestJavaGithub1302B(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1302b", "Github1302b", false, false, false, 1, "-config", "Github1302b.tla")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want ViolationSafety", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStats)) != 0 {
		t.Fatal("TLC_STATS present")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInvariantViolatedInitial, "Inv",
		"/\\ x = (<<1>> :> (<<1>> :> 10) @@ <<2>> :> 0)\n"+
			"/\\ y = (<<1>> :> (<<1>> :> 10) @@ <<2>> :> 99)\n")
}

// Original Github1302cTest.testSpec and source runner overrides.
func TestJavaGithub1302C(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1302c", "Github1302c", false, false, false, 1, "-config", "Github1302c.tla")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}
