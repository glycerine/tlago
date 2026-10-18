package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original BugzillaBug279Test.testSpec and its deadlock/dump overrides.
func TestJavaBugzillaBug279(t *testing.T) {
	runJavaBugzillaBug279(t)
}

func runJavaBugzillaBug279(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	args := append([]string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "BugzillaBug279Test.json")}, extraArgs...)
	r := runJavaTLCModelTestWithDeadlock(t, "BugzillaBug279", "InitStateBug", true, false, true, 1, true,
		args...)
	if r.ExitStatus != tlc.ExitStatusViolationDeadlock {
		t.Fatalf("exit=%d, want VIOLATION_DEADLOCK", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCDeadlockReached)) == 0 {
		t.Fatal("TLC_DEADLOCK_REACHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	subset, err := tlc.NewSubsetValue(tlc.NewIntervalValue(1, 8)).ToSetEnum()
	if err != nil {
		t.Fatal(err)
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ set = {}\n/\\ pc = 0\n/\\ fun = {}",
		"/\\ set = SUBSET (1..20)\n/\\ pc = 1\n/\\ fun = {5}",
		"/\\ set = " + tlc.ValuesPPR(subset.Normalize()) + "\n/\\ pc = 2\n/\\ fun = {5}",
	}, true)
	requireJavaTLCUncovered(t, r)
	return r
}
