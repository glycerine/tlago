package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original ViewMapTest.testSpec, including every action label in the trace.
func TestJavaViewMap(t *testing.T) {
	runJavaViewMap(t)
}

func runJavaViewMap(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	args := []string{"-view", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ViewMapTest.json")}
	r := runJavaTLCModelTest(t, "ViewMap", append(args, extraArgs...)...)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
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
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ buffer = <<>>\n/\\ waitset = {}",
		"/\\ buffer = <<>>\n/\\ waitset = {c1}",
		"/\\ buffer = <<>>\n/\\ waitset = {c1, c2}",
		"/\\ buffer = <<\"d\">>\n/\\ waitset = {c2}",
		"/\\ buffer = <<\"d\">>\n/\\ waitset = {c2, p1}",
		"/\\ buffer = <<>>\n/\\ waitset = {p1}",
		"/\\ buffer = <<>>\n/\\ waitset = {c1, p1}",
		"/\\ buffer = <<>>\n/\\ waitset = {c1, c2, p1}",
	}, true)
	actions := []string{
		"<Init line 53, col 9 to line 56, col 63 of module ViewMap>",
		"<lbc(c1) line 73, col 14 to line 84, col 60 of module ViewMap>",
		"<lbc(c2) line 73, col 14 to line 84, col 60 of module ViewMap>",
		"<lbp(p1) line 58, col 14 to line 69, col 60 of module ViewMap>",
		"<lbp(p1) line 58, col 14 to line 69, col 60 of module ViewMap>",
		"<lbc(c1) line 73, col 14 to line 84, col 60 of module ViewMap>",
		"<lbc(c1) line 73, col 14 to line 84, col 60 of module ViewMap>",
		"<lbc(c2) line 73, col 14 to line 84, col 60 of module ViewMap>",
	}
	for i, m := range javaTLCRecords(r, tlc.ECTLCStatePrint2) {
		if m.StateInfo.Info != actions[i] {
			t.Fatalf("state%d action=%v, want %s", i+1, m.StateInfo.Info, actions[i])
		}
	}
	requireJavaTLCUncovered(t, r, "line 91, col 60 to line 91, col 73 of module ViewMap")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR recorded")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 43 {
		t.Fatalf("register42=%v, want IntValue43", values[0])
	}
	return r
}
