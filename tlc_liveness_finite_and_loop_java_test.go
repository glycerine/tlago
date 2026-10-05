/*******************************************************************************
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Github710aTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessGithub710a(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github710a", "Github710a", false, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github710aTest.json"), "-config", "Github710.cfg")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AtMostOnce")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = FALSE", "x = TRUE", "x = FALSE", "x = TRUE"}, true)
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 5 {
		t.Fatalf("register42=%v, want5", values[0])
	}
}

// Original Github710bTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessGithub710b(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github710b", "Github710b", false, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github710bTest.json"), "-config", "Github710.cfg")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "1")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AtMostOnce")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 1", "x = 2"}, true)
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 3 {
		t.Fatalf("register42=%v, want3", values[0])
	}
}

// Original Github710cTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessGithub710c(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github710c", "Github710a", false, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github710cTest.json"), "-config", "Github710c.cfg")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "2")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AtMostOnce")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = FALSE", "x = TRUE", "x = FALSE", "x = TRUE", "x = FALSE", "x = TRUE"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "5")
}

// Original Github710dTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessGithub710d(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github710d", "Github710d", false, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github710dTest.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "4", "4")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "NeverThree")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 2", "x = 3"}, true)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint3)) != 0 {
		t.Fatal("TLC_STATE_PRINT3 present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBackToState)) != 0 {
		t.Fatal("TLC_BACK_TO_STATE present")
	}
}

// Original Github710dFairTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessGithub710dFair(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github710dFair", "Github710d", false, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github710dFairTest.json"), "-config", "Github710dFair.cfg")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "4", "4")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "NeverThree")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 2", "x = 3"}, true)
	requireJavaCodePlexStuttering(t, r, 5)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint3)) == 0 {
		t.Fatal("TLC_STATE_PRINT3 absent")
	}
}

// Original Github710eTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessGithub710e(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github710e", "Github710e", false, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github710eTest.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "VisitsTwoThenStaysSmall")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 2"}, true)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint3)) == 0 {
		t.Fatal("TLC_STATE_PRINT3 absent")
	}
}

// Original Github710eFairTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessGithub710eFair(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github710eFair", "Github710e", false, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github710eFairTest.json"), "-config", "Github710eFair.cfg")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "VisitsTwoThenStaysSmall")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 2"}, true)
	requireJavaCodePlexStuttering(t, r, 4)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint3)) == 0 {
		t.Fatal("TLC_STATE_PRINT3 absent")
	}
}

// Original LoopTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessLoop(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Loop", "SystemLoop", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "LoopTest.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "Liveness")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 144, 64)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 2"}, true)
	requireJavaCodePlexStuttering(t, r, 4)
	requireJavaTLCUncovered(t, r)
}

// Original LoopTestForcedPartial.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessLoopTestForcedPartial(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	// Source static initializer sets AbstractChecker.LIVENESS_TESTING_IMPLEMENTATION.
	t.Setenv("tlc2.tool.liveness.ILiveCheck.testing", "true")
	r := runJavaTLCModelTestWithRoot(t, "LoopTestForcedPartial", "SystemLoop", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "LoopTestForcedPartial.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "Liveness")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0"}, true)
	requireJavaCodePlexStuttering(t, r, 2)
	requireJavaTLCUncovered(t, r)
}

// Original LoopTestWeakFair.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessLoopTestWeakFair(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LoopTestWeakFair", "SystemLoop", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "LoopTestWeakFair.json"), "-config", "SystemLoopWeakFair")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "4")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCCheckingTemporalProps, "complete", "4")
	requireJavaTLCUncovered(t, r)
}

// Original Test3.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessTest3(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Test3", "Test3", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test3.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "3", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "Prop1")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 0", "x = 2"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
	requireJavaTLCUncovered(t, r)
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 5 {
		t.Fatalf("register42=%v, want5", values[0])
	}
}
