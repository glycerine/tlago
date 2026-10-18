package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

// Original Github1134a..fTest.testSpec methods. Each class has noGenerateSpec
// and doDumpTrace overrides; coverage, DOT, debugger and one worker stay enabled.
func TestJavaGithub1134A(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCa.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ViolationLiveness", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCPropertyViolatedInitial, "Init", "State = <<\"INIT\", \"INIT\">>\n")
	requireJavaTLCUncovered(t, r)
}

func TestJavaGithub1134B(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCb.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "4", "0")
	requireJavaTLCUncovered(t, r)
}

func TestJavaGithub1134C(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCc.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "4", "0")
	requireJavaTLCUncovered(t, r)
}

func TestJavaGithub1134D(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCd.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FailureSpecEval", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitialState,
		"Attempted to apply function:\n<<\"INIT\", \"INIT\">>\nto argument 3, which is not in the domain of the function.",
		"State = <<\"INIT\", \"INIT\">>\n")
}

func TestJavaGithub1134E(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCe.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "4", "0")
	requireJavaTLCUncovered(t, r)
}

func TestJavaGithub1134F(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCf.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FailureSpecEval", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitialState,
		"Attempted to apply function:\n<<\"INIT\", \"INIT\">>\nto argument 3, which is not in the domain of the function.",
		"State = <<\"INIT\", \"INIT\">>\n")
}
