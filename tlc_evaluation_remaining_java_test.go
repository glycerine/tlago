package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original EmptyExistentialQuantifierTest.testSpec and inherited exit assertion.
func TestJavaEmptyExistentialQuantifier(t *testing.T) {
	r := runJavaTLCModelTestWithDeadlock(t, "EmptyExistentialQuantifier", "EmptyExistentialQuantifier", true, true, true, 1, true, "-config", "EmptyExistentialQuantifier.tla", "-dumpTrace", "json", filepath.Join(t.TempDir(), "EmptyExistentialQuantifierTest.json"))
	if r.ExitStatus != tlc.ExitStatusViolationDeadlock {
		t.Fatalf("exit=%d, want ExitStatusViolationDeadlock", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "0")
}

// Original MinimalSetOfInitStatesTest.testSpec and inherited exit assertion.
func TestJavaMinimalSetOfInitStates(t *testing.T) {
	r := runJavaTLCModelTest(t, "MinimalSetOfInitStates", "-dumpTrace", "json", filepath.Join(t.TempDir(), "MinimalSetOfInitStatesTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated2, "8", "s", "6")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "14", "6", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCUncovered(t, r)
}

// Original MinimalSetOfNextStatesTest.testSpec and inherited exit assertion.
func TestJavaMinimalSetOfNextStates(t *testing.T) {
	r := runJavaTLCModelTest(t, "MinimalSetOfNextStates", "-dumpTrace", "json", filepath.Join(t.TempDir(), "MinimalSetOfNextStatesTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "57", "7", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCUncovered(t, r, "line 34, col 10 to line 34, col 15 of module MinimalSetOfNextStates", "line 42, col 10 to line 42, col 15 of module MinimalSetOfNextStates")
}

// Original UndeclaredRecursionTest.testSpec and inherited exit assertion.
func TestJavaUndeclaredRecursion(t *testing.T) {
	r := runJavaTLCModelTestWithDebugger(t, "UndeclaredRecursion", false, true, false, 1, "-config", "UndeclaredRecursion.tla", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "0")
}
