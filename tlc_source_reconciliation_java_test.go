//go:build tlc_reconcile

package tlago

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original DistributedTrace.testSpec, including its source-failing expectations.
func TestJavaDistributedTraceReconciliation(t *testing.T) {
	r := runJavaTLCModelTestWithWorkers(t, "DistributedTrace", true, true, 4,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.DistributedTrace.json"))
	// JUnit runs inherited tearDown even when a body assertion fails.
	defer func() {
		if r.ExitStatus != tlc.ExitStatusSuccess {
			t.Errorf("inherited exit=%d, want SUCCESS", r.ExitStatus)
		}
	}()
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Error("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Error("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Error("A TE spec was generated, but it should not be")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"x = 10", "x = 11", "x = 12", "x = 13", "x = 14", "x = 15",
		"x = 16", "x = 17", "x = 18", "x = 19", "x = 20",
	}, false)
}

// Original AssertExpressionStack.testSpec, preserving the nested-stack check.
func TestJavaAssertExpressionStackReconciliation(t *testing.T) {
	r := runJavaTLCModelTest(t, "AssertExpressionStack",
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.AssertExpressionStack.json"))
	defer func() {
		if r.ExitStatus != tlc.ExitStatusSuccess {
			t.Errorf("inherited exit=%d, want SUCCESS", r.ExitStatus)
		}
	}()
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Error("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Error("TLC_BUG recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Error("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Error("A TE spec was generated, but it should not be")
	}
	// The source default one-worker debugger selects TLCStateMutExt.
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1"}, true)
	for _, record := range javaTLCRecords(r, tlc.ECTLCNestedExpression) {
		if len(record.Params) > 0 && record.Params[0] == "    The error call stack is empty.\n" {
			t.Error("empty nested expression stack recorded")
		}
	}
}

// Original LivenessSymmetryWarning.testSpec and inherited SUCCESS expectation.
func TestJavaLivenessSymmetryWarningReconciliation(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "LivenessSymmetryWarning", "April25MC", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.liveness.LivenessSymmetryWarning.json"))
	defer func() {
		if r.ExitStatus != tlc.ExitStatusSuccess {
			t.Errorf("inherited exit=%d, want SUCCESS", r.ExitStatus)
		}
	}()
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Error("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFeatureUnsupportedLivenessSymmetry)) == 0 {
		t.Error("TLC_FEATURE_UNSUPPORTED_LIVENESS_SYMMETRY absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Error("GENERAL recorded")
	}
}

// Original DepthFirstTerminate.testSpec and availableProcessors worker count.
func TestJavaDepthFirstTerminateReconciliation(t *testing.T) {
	r := runJavaTLCModelTestWithWorkers(t, "DepthFirstTerminate", true, true, runtime.NumCPU(),
		"-dfid", "50", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.DepthFirstTerminate.json"))
	defer func() {
		if r.ExitStatus != tlc.ExitStatusSuccess {
			t.Errorf("inherited exit=%d, want SUCCESS", r.ExitStatus)
		}
	}()
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Error("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Error("GENERAL recorded")
	}
}
