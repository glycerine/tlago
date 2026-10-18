package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original Github297Test.testSpec and constructor settings. Upstream ignores
// the method for https://github.com/tlaplus/tlaplus/issues/297.
func TestJavaGithub297Ignored(t *testing.T) {
	t.Skip("upstream @Ignore: https://github.com/tlaplus/tlaplus/issues/297")
	r := runJavaTLCModelTestWithDebugger(t, "Github297", true, true, false, 1,
		"-config", "Github297.tla")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
}
