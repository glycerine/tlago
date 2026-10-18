package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Github1145Test.testSpec, with default ModelCheckerTestCase flags.
func TestJavaGithub1145(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github1145", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1145Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
}

// Original Github1145bTest.testSpec. KeepLazy must preserve the EXCEPT update
// through FcnLambdaValue.toTuple; the original disables all output instrumentation.
func TestJavaGithub1145B(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1145b", "Github1145b", false, false, false, 1, "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want ViolationSafety", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInvariantViolatedInitial, "Inv", "x = 1\n")
}
