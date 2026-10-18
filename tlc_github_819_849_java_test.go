package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Github819Test.testSpec and constructor overrides.
func TestJavaGithub819(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github819", "-noGenerateSpecTE", "-config", "Github819.tla")
	if r.ExitStatus != tlc.ExitStatusFailureLivenessEval {
		t.Fatalf("exit=%d, want FAILURE_LIVENESS_EVAL; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_TAUTOLOGY absent")
	}
}

// Original Github849Test.testSpec and default trace/coverage/debugger settings.
func TestJavaGithub849(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github849", "-config", "Github849.tla", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github849Test.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCModuleValueJavaMethodOverride)) != 0 {
		t.Fatal("TLC_MODULE_VALUE_JAVA_METHOD_OVERRIDE present")
	}
}
