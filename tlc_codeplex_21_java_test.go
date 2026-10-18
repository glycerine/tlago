package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original CodePlexBug21Test.testSpec, including the inherited SUCCESS exit.
func TestJavaCodePlexBug21(t *testing.T) {
	r := runJavaTLCModelTest(t, "CodePlexBug21", "-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug21Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
}
