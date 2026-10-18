package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original ActionLevelPropATest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropA(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropA.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropATest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevel)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevel absent")
	}
}

// Original ActionLevelPropBTest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropB(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropB.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropBTest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevel)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevel absent")
	}
}

// Original ActionLevelPropCTest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropC(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropC.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropCTest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevel)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevel absent")
	}
}

// Original ActionLevelPropDTest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropD(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropD.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropDTest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevelSquareASubV)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevelSquareASubV absent")
	}
}

// Original ActionLevelPropETest.testSpec; source coverage override and inherited exit.
func TestJavaActionLevelPropE(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "ActionLevelProp", "ActionLevelProp", false, true, true, 1, "-config", "ActionLevelPropE.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionLevelPropETest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigPropertyActionLevelAngleASubV)) == 0 {
		t.Fatal("ECTLCConfigPropertyActionLevelAngleASubV absent")
	}
}
