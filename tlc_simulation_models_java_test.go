package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Whole original simulation.Github1191Test.testSpec, its constructor and
// ModelCheckerTestCase runner/exit settings. runWithDebugger returns false.
func TestJavaSimulationGithub1191(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1191", "Github1191", true, true, false, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1191Test.json"),
		"-config", "Github1191.tla", "-simulate", "num=1")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
}

// Whole original simulation.Github1191aTest.testSpec, including zero uncovered.
func TestJavaSimulationGithub1191a(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1191a", "Github1191a", true, true, false, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1191aTest.json"),
		"-config", "Github1191a.tla", "-simulate", "num=1")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCUncovered(t, r)
}

// Whole original simulation.Github602Test.testSpec and disabled debugger override.
func TestJavaSimulationGithub602(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github602", "Github602", true, true, false, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github602Test.json"),
		"-config", "Github602.tla", "-simulate", "num=1")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCProgressSimu, "156", "1", "100", "0", "0")
}

// Whole original simulation.NQSpecTest.testSpec. Preserve the inherited debugger
// option and all ModelCheckerTestCase settings; no new debugger test is added.
func TestJavaSimulationNQSpec(t *testing.T) {
	skipSlowTest(t)

	r := runJavaTLCModelTestWithRoot(t, "SimulationNQSpec", "MC", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "NQSpecTest.json"),
		"-simulate", "num=100")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
}
