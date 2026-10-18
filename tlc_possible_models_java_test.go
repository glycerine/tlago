package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original PossibleTest.testSpec and inherited successful exit.
func TestJavaPossible(t *testing.T) {
	r := runJavaTLCModelTest(t, "PossibleTest", "-dumpTrace", "json", filepath.Join(t.TempDir(), "PossibleTest.json"))
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
	if len(javaTLCRecords(r, tlc.ECTLCPossibleUnwitnessed)) != 0 {
		t.Fatal("TLC_POSSIBLE_UNWITNESSED present")
	}
}

func runJavaPossibleFailure(t *testing.T, class, config string) *tlc.Result {
	t.Helper()
	return runJavaTLCModelTest(t, "PossibleFail", "-config", config, "-dumpTrace", "json", filepath.Join(t.TempDir(), class+".json"))
}

// Original PossibleFailActionTest.testSpec and constructor settings.
func TestJavaPossibleFailAction(t *testing.T) {
	r := runJavaPossibleFailure(t, "PossibleFailActionTest", "PossibleFailActionTest.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit=%d, want VIOLATION_ASSUMPTION; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	records := javaTLCRecords(r, tlc.ECTLCPossibleUnwitnessed)
	if len(records) == 0 {
		t.Fatal("TLC_POSSIBLE_UNWITNESSED absent")
	}
	// TestMPRecorder.recordedWithStringValueAt inspects the first record.
	if len(records[0].Params) == 0 || records[0].Params[0] != "BigJump" {
		t.Fatalf("unwitnessed=%v, want first parameter BigJump", records[0].Params)
	}
}

// Original PossibleFailMixedTest.testSpec and constructor settings.
func TestJavaPossibleFailMixed(t *testing.T) {
	r := runJavaPossibleFailure(t, "PossibleFailMixedTest", "PossibleFailMixedTest.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit=%d, want VIOLATION_ASSUMPTION; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	records := javaTLCRecords(r, tlc.ECTLCPossibleUnwitnessed)
	if len(records) == 0 {
		t.Fatal("TLC_POSSIBLE_UNWITNESSED absent")
	}
	// TestMPRecorder.recordedWithStringValueAt inspects the first record.
	if len(records[0].Params) == 0 || records[0].Params[0] != "Unreachable" {
		t.Fatalf("unwitnessed=%v, want first parameter Unreachable", records[0].Params)
	}
}

// Original PossibleFailNoTransTest.testSpec and constructor settings.
func TestJavaPossibleFailNoTrans(t *testing.T) {
	r := runJavaPossibleFailure(t, "PossibleFailNoTransTest", "PossibleFailNoTransTest.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit=%d, want VIOLATION_ASSUMPTION; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	records := javaTLCRecords(r, tlc.ECTLCPossibleUnwitnessed)
	if len(records) == 0 {
		t.Fatal("TLC_POSSIBLE_UNWITNESSED absent")
	}
	// TestMPRecorder.recordedWithStringValueAt inspects the first record.
	if len(records[0].Params) == 0 || records[0].Params[0] != "BigJump" {
		t.Fatalf("unwitnessed=%v, want first parameter BigJump", records[0].Params)
	}
}

// Original PossibleFailStateTest.testSpec and constructor settings.
func TestJavaPossibleFailState(t *testing.T) {
	r := runJavaPossibleFailure(t, "PossibleFailStateTest", "PossibleFailStateTest.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit=%d, want VIOLATION_ASSUMPTION; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	records := javaTLCRecords(r, tlc.ECTLCPossibleUnwitnessed)
	if len(records) == 0 {
		t.Fatal("TLC_POSSIBLE_UNWITNESSED absent")
	}
	// TestMPRecorder.recordedWithStringValueAt inspects the first record.
	if len(records[0].Params) == 0 || records[0].Params[0] != "Unreachable" {
		t.Fatalf("unwitnessed=%v, want first parameter Unreachable", records[0].Params)
	}
}
