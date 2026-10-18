package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"strconv"
	"testing"
)

// Original DiameterTest.testSpec, four workers and source lower-bound assertion.
func TestJavaDiameter(t *testing.T) {
	r := runJavaTLCModelTestWithWorkers(t, "DieHardTLA", true, true, 4, "-dumpTrace", "json", filepath.Join(t.TempDir(), "DiameterTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "97", "16", "0")
	records := javaTLCRecords(r, tlc.ECTLCSearchDepth)
	if len(records) == 0 || len(records[0].Params) == 0 {
		t.Fatal("search depth absent")
	}
	level, err := strconv.ParseInt(records[0].Params[0], 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	if level < 8 {
		t.Fatalf("Level below threshold: %d", level)
	}
	requireJavaTLCUncovered(t, r)
}
