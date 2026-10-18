package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original ASTest.testSpec, including inherited FAILURE_SPEC_EVAL exit.
func TestJavaAS(t *testing.T) {
	r := runJavaTLCModelTest(t, "AS", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ASTest.json"))
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FAILURE_SPEC_EVAL", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatesAndNoNextAction)) == 0 {
		t.Fatal("TLC_STATES_AND_NO_NEXT_ACTION absent")
	}
}
