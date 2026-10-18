package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original EvaluatingValueTest.testSpec, with its source @Evaluation callback
// and inherited successful exit. The test index is shared with override tests.
func TestJavaEvaluatingValue(t *testing.T) {
	linkJavaNativeOverrideFixtures(t)
	r := runJavaTLCModelTest(t, "EvaluatingValueTest", "-dumpTrace", "json", filepath.Join(t.TempDir(), "EvaluatingValueTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE recorded")
	}
}
