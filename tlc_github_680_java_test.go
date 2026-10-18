package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Github680aTest.testSpec, including runWithDebugger=false and -nowarning.
func TestJavaGithub680a(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github680a", "Github680a", true, true, false, 1,
		"-config", "Github680a.tla", "-nowarning", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github680aTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCUnchangedVariableChanged, "x", "line 7, col 20 to line 7, col 20 of module Github680a")
}

// Original Github680bTest.testSpec, including runWithDebugger=false and -nowarning.
func TestJavaGithub680b(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github680b", "Github680b", true, true, false, 1,
		"-config", "Github680b.tla", "-nowarning", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github680bTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCUnchangedVariableChanged, "x", "line 7, col 22 to line 7, col 22 of module Github680b")
}

// Original Github680cTest.testSpec, including runWithDebugger=false and -nowarning.
func TestJavaGithub680c(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github680c", "Github680c", true, true, false, 1,
		"-config", "Github680c.tla", "-nowarning", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github680cTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCUnchangedVariableChanged, "x", "line 8, col 22 to line 8, col 22 of module Github680c")
}
