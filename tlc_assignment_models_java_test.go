package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

func runJavaAssignmentModel(t *testing.T, name string, stats []string, depth bool) {
	t.Helper()
	r := runJavaTLCModelTest(t, name, "-dumpTrace", "json", filepath.Join(t.TempDir(), name+"Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, stats...)
	if depth {
		requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	}
	requireJavaTLCUncovered(t, r)
}

// Original AssignmentInitTest.test, including inherited successful exit.
func TestJavaAssignmentInit(t *testing.T) {
	runJavaAssignmentModel(t, "AssignmentInit", []string{"6", "5", "0"}, false)
}

// Original AssignmentInitNegTest.test, including inherited successful exit.
func TestJavaAssignmentInitNeg(t *testing.T) {
	runJavaAssignmentModel(t, "AssignmentInitNeg", []string{"2", "1", "0"}, false)
}

// Original AssignmentInitExpensiveTest.test, including inherited successful exit.
func TestJavaAssignmentInitExpensive(t *testing.T) {
	runJavaAssignmentModel(t, "AssignmentInitExpensive", []string{"10002", "1", "0"}, false)
}

// Original AssignmentNextTest.test, including inherited successful exit.
func TestJavaAssignmentNext(t *testing.T) {
	runJavaAssignmentModel(t, "AssignmentNext", []string{"3", "2", "0"}, true)
}

// Original AssignmentNext2Test.test, including inherited successful exit.
func TestJavaAssignmentNext2(t *testing.T) {
	runJavaAssignmentModel(t, "AssignmentNext2", []string{"3", "2", "0"}, false)
}

// Original AssignmentNext3Test.test, including inherited successful exit.
func TestJavaAssignmentNext3(t *testing.T) {
	runJavaAssignmentModel(t, "AssignmentNext3", []string{"26", "5", "0"}, false)
}
