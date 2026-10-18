package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

const ignoredGithub607Reason = "upstream @Ignore: Fix eager evaluation of TBGraphNode#isConsistent in TableauLiveChecker#addNextState discussed in Github issue #607"

// Original Github607Test.testSpec, constructor arguments and coverage override.
func TestJavaGithub607Ignored(t *testing.T) {
	t.Skip(ignoredGithub607Reason)
	r := runJavaTLCModelTestWithRoot(t, "Github607", "Github607", false, true, true, 1,
		"-config", "Github607.tla")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}
