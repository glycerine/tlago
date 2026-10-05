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
	"runtime"
	"testing"
)

// Original Github317Test.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessGithub317(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github317", "Github317", true, true, true, runtime.NumCPU(), "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github317Test.json"), "-config", "Github317.tla", "-noGenerateSpecTE")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusError {
		t.Fatalf("exit=%d, want ERROR; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Fatal("A TE spec was generated, but it should not")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 records empty")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCNestedExpression, "0. Line 18, column 20 to line 18, column 23 in Github317\n1. Line 10, column 9 to line 11, column 23 in Github317\n2. Line 10, column 12 to line 10, column 21 in Github317\n3. Line 11, column 12 to line 11, column 23 in Github317\n4. Line 18, column 15 to line 18, column 18 in Github317\n5. Line 5, column 9 to line 5, column 15 in Github317\n6. Line 5, column 13 to line 5, column 13 in Github317\n\n")
	if got := len(javaTLCRecords(r, tlc.ECTLCNestedExpression)); got != 1 {
		t.Fatalf("TLC_NESTED_EXPRESSION count=%d, want1", got)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStateNotCompletelySpecifiedLive, "b", "line 5, col 13 to line 5, col 13 of module Github317")
}

// Original Github317aTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessGithub317a(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github317a", "Github317a", true, true, true, runtime.NumCPU(), "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github317aTest.json"), "-config", "Github317a.tla")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusError {
		t.Fatalf("exit=%d, want ERROR; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 records empty")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCNestedExpression, "0. Line 138, column 41 to line 138, column 48 in Github317a\n1. Line 117, column 13 to line 133, column 54 in Github317a\n2. Line 118, column 15 to line 133, column 54 in Github317a\n3. Line 128, column 23 to line 133, column 54 in Github317a\n4. Line 128, column 26 to line 128, column 67 in Github317a\n5. Line 129, column 26 to line 132, column 50 in Github317a\n6. Line 129, column 29 to line 129, column 59 in Github317a\n7. Line 129, column 29 to line 129, column 42 in Github317a\n\n")
	if got := len(javaTLCRecords(r, tlc.ECTLCNestedExpression)); got != 1 {
		t.Fatalf("TLC_NESTED_EXPRESSION count=%d, want1", got)
	}
}

// Original NoSymmetryTableauModelCheckerTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessNoSymmetryTableauModelChecker(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "NoSymmetryTableauModelChecker", "NoSymmetryLivenessTableauMC", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "NoSymmetryTableauModelCheckerTest.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCLiveImplied, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "8", "s")
	if len(javaTLCRecords(r, tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Fatal("A TE spec was generated, but it should not")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5492", "1272", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) != 0 {
		t.Fatal("TLC_STATE_PRINT2 present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original OneBitMutexNoSymmetryTest.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessOneBitMutexNoSymmetry(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "OneBitMutexNoSymmetry", "OneBitMutexNoSymmetryMC", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "OneBitMutexNoSymmetryTest.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "244", "127", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_144491423293819000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 11700, 3728)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaTLCUncovered(t, r, "line 80, col 38 to line 80, col 69 of module OneBitMutex", "line 96, col 16 to line 96, col 47 of module OneBitMutex", "line 97, col 16 to line 97, col 50 of module OneBitMutex")
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
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 244 {
		t.Fatalf("register42=%v, want244", values[0])
	}
}

// Original UnsymmetricModelCheckerTestA.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessUnsymmetricModelCheckerTestA(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "UnsymmetricModelCheckerTestA", "UnsymmetricMCA", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "UnsymmetricModelCheckerTestA.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCLiveImplied, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated2, "2", "s", "1")
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_1444366015116104000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = a", "x = 1"}, true)
	expectedActions := []string{"<Init line 5, col 10 to line 5, col 16 of module Unsymmetric>", "<NextA line 17, col 13 to line 19, col 36 of module Unsymmetric>"}
	for i, record := range javaTLCRecords(r, tlc.ECTLCStatePrint2) {
		if record.StateInfo.Info != expectedActions[i] {
			t.Fatalf("action%d=%v, want%q", i, record.StateInfo.Info, expectedActions[i])
		}
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
	requireJavaTLCUncovered(t, r, "line 19, col 31 to line 19, col 36 of module Unsymmetric")
}

// Original UnsymmetricModelCheckerTestB.testSpec, constructor, setup and inherited exit.
func TestJavaLivenessUnsymmetricModelCheckerTestB(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "UnsymmetricModelCheckerTestB", "UnsymmetricMCB", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "UnsymmetricModelCheckerTestB.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCLiveImplied, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated2, "2", "s", "1")
	if len(javaTLCRecords(r, tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Fatal("A TE spec was generated, but it should not")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "2", "0")
	requireJavaTLCUncovered(t, r, "line 37, col 31 to line 37, col 36 of module Unsymmetric")
}
