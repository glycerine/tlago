package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original Github715Test.testSpec, including doCoverage=false.
func TestJavaGithub715(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github715", "Github715", false, true, true, 1,
		"-config", "Github715.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github715Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigOpIsEqual)) != 0 {
		t.Fatal("TLC_CONFIG_OP_IS_EQUAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaStateLevel)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_STATE_LEVEL absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original Github715bTest.testSpec, including doCoverage=false.
func TestJavaGithub715b(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github715", "Github715", false, true, true, 1,
		"-config", "Github715b.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github715bTest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ExitStatusErrorConfigParse", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCConfigOpIsEqual, "x", "line 2, col 10 to line 2, col 10 of module Github715", "spec")
}

// Original Github715cTest.testSpec, including doCoverage=false.
func TestJavaGithub715c(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github715", "Github715", false, true, true, 1,
		"-config", "Github715c.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github715cTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaStateLevel)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_STATE_LEVEL absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original Github715dTest.testSpec, including doCoverage=false.
func TestJavaGithub715d(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github715", "Github715", false, true, true, 1,
		"-config", "Github715d.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github715dTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaStateLevel)) != 0 {
		t.Fatal("TLC_LIVE_FORMULA_STATE_LEVEL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}
