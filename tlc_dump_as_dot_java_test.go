package tlago

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original DumpAsDotTest.testSpec, including its exact DOT master-file bytes.
func TestJavaDumpAsDot(t *testing.T) {
	prefix := filepath.Join(t.TempDir(), "DumpAsDotTest")
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "MCa", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "DumpAsDotTest.json"),
		"-dump", "dot,colorize,actionlabels,stuttering", prefix)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "18", "11", "0")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 18 {
		t.Fatalf("register42=%v, want IntValue18", values[0])
	}
	dumpFile := prefix + ".dot"
	if _, err := os.Stat(dumpFile); err != nil {
		t.Fatalf("dump file absent: %v", err)
	}
	master, err := os.ReadFile(filepath.Join("tlc", "test_vectors", "models", "CodePlexBug08", "DumpAsDotTest.dot"))
	if err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(dumpFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(master, actual) {
		t.Fatalf("DOT dump bytes differ from original master:\nactual:\n%s\nmaster:\n%s", actual, master)
	}
	requireJavaTLCUncovered(t, r)
}
