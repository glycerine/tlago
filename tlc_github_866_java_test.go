package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original Github866Test.testSpec, including its two-worker constructor settings.
func TestJavaGithub866(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github866", "Github866", false, false, false, 2, "-noGenerateSpecTE", "-config", "Github866.tla")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}
