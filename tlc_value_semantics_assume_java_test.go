package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaValueSemanticsAssume(t *testing.T) {
	result := runJavaTLCModelTestWithDebugger(t, "ValueSemanticsAssume", true, true, false, 1,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ValueSemanticsAssume.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECTLCAssumptionFalse, tlc.ECTLCAssumptionEvaluationError, tlc.ECGeneral} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected code %d records: %v", code, got)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "0", "0", "0")
}
