package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original InitEvalOrder1Test.test, inherited from InitEvalOrderTest.
func TestJavaInitEvalOrder1(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrder", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrder1Test.json"), "-config", "InitEvalOrder1.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}

// Original InitEvalOrder2Test.test, inherited from InitEvalOrderTest.
func TestJavaInitEvalOrder2(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrder", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrder2Test.json"), "-config", "InitEvalOrder2.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}

// Original InitEvalOrder3Test.test, inherited from InitEvalOrderTest.
func TestJavaInitEvalOrder3(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrder", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrder3Test.json"), "-config", "InitEvalOrder3.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}

// Original InitEvalOrder4Test.test, inherited from InitEvalOrderTest.
func TestJavaInitEvalOrder4(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrder", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrder4Test.json"), "-config", "InitEvalOrder4.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}

// Original InitEvalOrderBasicTest.test.
func TestJavaInitEvalOrderBasic(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrderBasic", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrderBasicTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCUncovered(t, r)
}
