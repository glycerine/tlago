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
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original ExamplesACPNBTLCTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesACPNBTLC(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ACP_NB_TLC", "ACP_NB_TLC", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AllMustCommit")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesAbaAsynByzTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesAbaAsynByz(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "aba_asyn_byz", "aba_asyn_byz", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AllAccept")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesAsyncTerminationDetectionTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesAsyncTerminationDetection(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "AsyncTerminationDetection", "AsyncTerminationDetection", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "PendingInfOften")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesBcastByzTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesBcastByz(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "bcastByz", "bcastByz", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// Original ExamplesBcastFolkloreTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesBcastFolklore(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "bcastFolklore", "bcastFolklore", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "CrashCountInfOftenZero")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesBlockingQueuePoisonAppleTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesBlockingQueuePoisonApple(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "BlockingQueuePoisonApple", "BlockingQueuePoisonApple", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "BufferInfOftenNonEmpty")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesBufferedRandomAccessFileTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesBufferedRandomAccessFile(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "BufferedRandomAccessFile", "BufferedRandomAccessFile", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "DirtyLeadsToClean")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesCf1sFolkloreTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesCf1sFolklore(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "cf1s_folklore", "cf1s_folklore", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// Original ExamplesCoffeeCan100BeansTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesCoffeeCan100Beans(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "CoffeeCan100Beans", "CoffeeCan100Beans", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// The original tool-jar manifest supplies CommunityModules on the classpath.
func setJavaExampleClasspath(t *testing.T) {
	t.Helper()
	modules, err := filepath.Abs(filepath.Join("test_vectors", "java-sany", "CommunityModules.jar"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLASSPATH", modules)
}

// Original ExamplesEWD840Test.testSpec, constructor and all five settings overrides.
func TestJavaExamplesEWD840(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesEWD840", "EWD840", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "TokenPermanentlyBlack")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesEWD998ChanIDTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesEWD998ChanID(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesEWD998ChanID", "EWD998ChanID", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "PassesInfOftenLow")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesEWD998Test.testSpec, constructor and all five settings overrides.
func TestJavaExamplesEWD998(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesEWD998", "EWD998", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "TokenAwayFromZero")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesEnvironmentControllerTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesEnvironmentController(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesEnvironmentController", "EnvironmentController", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// Original ExamplesHuangTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesHuang(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesHuang", "Huang", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// Original ExamplesLiveHourClockTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesLiveHourClock(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesLiveHourClock", "LiveHourClock", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "ClockSettles")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesLockHSTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesLockHS(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesLockHS", "LockHS", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// Original ExamplesMCAlternatingBitTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesMCAlternatingBit(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesMCAlternatingBit", "MCAlternatingBit", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "PermanentBitDisagreement")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesMCEWD687aTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesMCEWD687a(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesMCEWD687a", "MCEWD687a", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "LeaderPermanentlyActive")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesMCLiveInternalMemoryTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesMCLiveInternalMemory(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesMCLiveInternalMemory", "MCLiveInternalMemory", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AlwaysBusy")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesMCLiveWriteThroughCacheTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesMCLiveWriteThroughCache(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesMCLiveWriteThroughCache", "MCLiveWriteThroughCache", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "PermanentlyBusy")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesMCWriteThroughCacheTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesMCWriteThroughCache(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesMCWriteThroughCache", "MCWriteThroughCache", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "BusyLeadsToRdy")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesMCYoYoNoPruningTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesMCYoYoNoPruning(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesMCYoYoNoPruning", "MCYoYoNoPruning", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AllPermanentlyDown")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesNbacgGuer01Test.testSpec, constructor and all five settings overrides.
func TestJavaExamplesNbacgGuer01(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesNbacgGuer01", "nbacg_guer01", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AllCommit")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesPrisonersTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesPrisoners(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesPrisoners", "Prisoners", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "CountInfOftenZero")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesSDPAttackNewSolutionTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesSDPAttackNewSolution(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesSDPAttackNewSolution", "MC", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AttackerEstablishes")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesSchedulingAllocatorTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesSchedulingAllocator(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesSchedulingAllocator", "SchedulingAllocator", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "PermanentAllocation")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesSimpleAllocatorTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesSimpleAllocator(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesSimpleAllocator", "SimpleAllocator", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AllocInfOftenEmpty")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesSingleLaneBridgeTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesSingleLaneBridge(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesSingleLaneBridge", "MC", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// Original ExamplesSpanTreeRandomTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesSpanTreeRandom(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesSpanTreeRandom", "SpanTreeRandom", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "NextAlwaysEnabled")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesSpanTreeTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesSpanTree(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesSpanTree", "SpanTree", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AllNodesDirectChildren")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original ExamplesSyncTerminationDetectionTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesSyncTerminationDetection(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesSyncTerminationDetection", "SyncTerminationDetection", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "DetectionLeadsToReactivation")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

func javaExamplesSDPAttackState(capDataMsg, sChannel, sTCPLinkSet, uState, dropPackets, fwDataChannel, aCounter, aChannel, aTCPLinkSet, sdpSucSession, dataKnowledge, spoofSession, agedRuleSet, fwCtlChannel, uTstamp, sniffCount, aclRuleSet, uTCPLinkSet, authKnowledge, uAuthSession, authChannel string) string {
	return "/\\ CapDataMsg = " +
		capDataMsg +
		"\n" +
		"/\\ sChannel = " +
		sChannel +
		"\n" +
		"/\\ sTCPLinkSet = " +
		sTCPLinkSet +
		"\n" +
		"/\\ uSvrInfo = [IP |-> 22, Port |-> 80]\n" +
		"/\\ uState = \"" +
		uState +
		"\"\n" +
		"/\\ DropPackets = " +
		dropPackets +
		"\n" +
		"/\\ FwState = \"Work\"\n" +
		"/\\ ReplayCount = 0\n" +
		"/\\ aState = \"Listen\"\n" +
		"/\\ CapAuthMsg = {}\n" +
		"/\\ FwDataChannel = " +
		fwDataChannel +
		"\n" +
		"/\\ SDPSvrState = \"Work\"\n" +
		"/\\ aCounter = " +
		aCounter +
		"\n" +
		"/\\ aChannel = " +
		aChannel +
		"\n" +
		"/\\ aTCPLinkSet = " +
		aTCPLinkSet +
		"\n" +
		"/\\ SDPSucSession = " +
		sdpSucSession +
		"\n" +
		"/\\ aIP = 11\n" +
		"/\\ uChannel = <<>>\n" +
		"/\\ Account = {[Key |-> 44, ClientID |-> 1]}\n" +
		"/\\ DataKnowledge = " +
		dataKnowledge +
		"\n" +
		"/\\ SpoofSession = " +
		spoofSession +
		"\n" +
		"/\\ uID = 1\n" +
		"/\\ ReplaySession = {}\n" +
		"/\\ sState = \"Listen\"\n" +
		"/\\ Key = 44\n" +
		"/\\ uIP = 11\n" +
		"/\\ SpoofCount = 0\n" +
		"/\\ AgedRuleSet = " +
		agedRuleSet +
		"\n" +
		"/\\ FwCtlChannel = " +
		fwCtlChannel +
		"\n" +
		"/\\ sSvrInfo = [IP |-> 22, Port |-> 80]\n" +
		"/\\ uTstamp = " +
		uTstamp +
		"\n" +
		"/\\ sniffCount = " +
		sniffCount +
		"\n" +
		"/\\ AclRuleSet = " +
		aclRuleSet +
		"\n" +
		"/\\ uSDPSvrInfo = [IP |-> 12, Port |-> 8000]\n" +
		"/\\ aSession = {}\n" +
		"/\\ uTCPLinkSet = " +
		uTCPLinkSet +
		"\n" +
		"/\\ AuthKnowledge = " +
		authKnowledge +
		"\n" +
		"/\\ SDPSvrInfo = [IP |-> 12, Port |-> 8000]\n" +
		"/\\ uAuthSession = " +
		uAuthSession +
		"\n" +
		"/\\ AuthChannel = " +
		authChannel
}

// Original ExamplesSDPAttackTest.testSpec, constructor and all five settings overrides.
func TestJavaExamplesSDPAttack(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesSDPAttack", "MC", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1103", "526", "175")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInvariantViolatedBehavior, "DataAccessSafeLaw")
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	const SPA_AUTH_SET = "{ [ MsgID |-> \"SPA_AUTH\",\n    sIP |-> 11,\n    sPort |-> 1024,\n    dIP |-> 12,\n    dPort |-> 8000,\n    ClientID |-> 1,\n    Tstamp |-> 0,\n    SvrIP |-> 58,\n    SvrPort |-> 124,\n    HMAC |-> 238,\n    Type |-> \"User\" ] }"
	const SPA_AUTH_SEQ = "<< [ MsgID |-> \"SPA_AUTH\",\n     sIP |-> 11,\n     sPort |-> 1024,\n     dIP |-> 12,\n     dPort |-> 8000,\n     ClientID |-> 1,\n     Tstamp |-> 0,\n     SvrIP |-> 58,\n     SvrPort |-> 124,\n     HMAC |-> 238,\n     Type |-> \"User\" ] >>"
	const USR_SYN_SET = "{ [ sIP |-> 11,\n    sPort |-> 1025,\n    dIP |-> 22,\n    dPort |-> 80,\n    Type |-> \"User\",\n    Flg |-> \"TCP_SYN\" ] }"
	const USR_SYN_SEQ = "<< [ sIP |-> 11,\n     sPort |-> 1025,\n     dIP |-> 22,\n     dPort |-> 80,\n     Type |-> \"User\",\n     Flg |-> \"TCP_SYN\" ] >>"
	const FW_CTL_ADD_SEQ = "<< [ op |-> \"Add\",\n     Rule |->\n         [ sIP |-> 11,\n           sPort |-> 65536,\n           dIP |-> 22,\n           dPort |-> 80,\n           protocol |-> \"TCP\",\n           action |-> \"Accept\" ] ] >>"
	const ACL_WILD = "{ [ sIP |-> 11,\n    sPort |-> 65536,\n    dIP |-> 22,\n    dPort |-> 80,\n    protocol |-> \"TCP\",\n    action |-> \"Accept\" ] }"
	const ACL_TWO = "{ [ sIP |-> 11,\n    sPort |-> 2024,\n    dIP |-> 22,\n    dPort |-> 80,\n    protocol |-> \"TCP\",\n    action |-> \"Accept\" ],\n  [ sIP |-> 11,\n    sPort |-> 65536,\n    dIP |-> 22,\n    dPort |-> 80,\n    protocol |-> \"TCP\",\n    action |-> \"Accept\" ] }"
	const ATK_SYN_SENT = "{ [ sIP |-> 11,\n    sPort |-> 2024,\n    dIP |-> 22,\n    dPort |-> 80,\n    State |-> \"SYN_SENT\",\n    AuthID |-> 65535 ] }"
	const ATK_SYN_SEQ = "<< [ sIP |-> 11,\n     sPort |-> 2024,\n     dIP |-> 22,\n     dPort |-> 80,\n     Type |-> \"Attacker\",\n     Flg |-> \"TCP_SYN\" ] >>"
	const ATK_ACK_SEQ = "<< [ sIP |-> 11,\n     sPort |-> 2024,\n     dIP |-> 22,\n     dPort |-> 80,\n     Type |-> \"Attacker\",\n     Flg |-> \"TCP_ACK\" ] >>"
	const USR_TCP = "{[sIP |-> 11, sPort |-> 1025, dIP |-> 22, dPort |-> 80, State |-> \"SYN_SENT\"]}"
	const sTCPLink10 = "{ [ sIP |-> 22,\n    sPort |-> 80,\n    dIP |-> 11,\n    dPort |-> 2024,\n    Type |-> \"Attacker\",\n    State |-> \"SYN_RCVD\" ] }"
	const aChannel10 = "<< [ sIP |-> 22,\n     sPort |-> 80,\n     dIP |-> 11,\n     dPort |-> 2024,\n     Type |-> \"Attacker\",\n     Flg |-> \"TCP_SYN_ACK\" ] >>"
	const atkEstablished = "{ [ sIP |-> 11,\n    sPort |-> 2024,\n    dIP |-> 22,\n    dPort |-> 80,\n    State |-> \"ESTABLISHED\",\n    AuthID |-> 65535 ] }"
	expectedTrace := []string{
		javaExamplesSDPAttackState("{}", "<<>>", "{}", "Start_Auth", "{}", "<<>>", "0", "<<>>", "{}", "{}", "{}", "{}", "{}", "<<>>", "0", "0", "{}", "{}", "{}", "{}", "<<>>"),
		javaExamplesSDPAttackState("{}", "<<>>", "{}", "Auth_End", "{}", "<<>>", "0", "<<>>", "{}", "{}", "{}", "{}", "{}", "<<>>", "1", "0", "{}", "{}", "{}", SPA_AUTH_SET, SPA_AUTH_SEQ),
		javaExamplesSDPAttackState("{}", "<<>>", "{}", "Connecting", "{}", USR_SYN_SEQ, "0", "<<>>", "{}", "{}", "{}", "{}", "{}", "<<>>", "1", "0", "{}", USR_TCP, "{}", SPA_AUTH_SET, SPA_AUTH_SEQ),
		javaExamplesSDPAttackState("{}", "<<>>", "{}", "Connecting", "{}", USR_SYN_SEQ, "0", "<<>>", "{}", SPA_AUTH_SET, "{}", "{}", "{}", FW_CTL_ADD_SEQ, "1", "0", "{}", USR_TCP, "{}", SPA_AUTH_SET, "<<>>"),
		javaExamplesSDPAttackState(USR_SYN_SET, "<<>>", "{}", "Connecting", "{}", USR_SYN_SEQ, "0", "<<>>", "{}", SPA_AUTH_SET, USR_SYN_SET, "{}", "{}", FW_CTL_ADD_SEQ, "1", "1", "{}", USR_TCP, "{}", SPA_AUTH_SET, "<<>>"),
		javaExamplesSDPAttackState(USR_SYN_SET, "<<>>", "{}", "Connecting", USR_SYN_SET, "<<>>", "0", "<<>>", "{}", SPA_AUTH_SET, USR_SYN_SET, "{}", "{}", FW_CTL_ADD_SEQ, "1", "1", "{}", USR_TCP, "{}", SPA_AUTH_SET, "<<>>"),
		javaExamplesSDPAttackState(USR_SYN_SET, "<<>>", "{}", "Connecting", USR_SYN_SET, "<<>>", "0", "<<>>", "{}", SPA_AUTH_SET, USR_SYN_SET, "{}", "{}", "<<>>", "1", "1", ACL_WILD, USR_TCP, "{}", SPA_AUTH_SET, "<<>>"),
		javaExamplesSDPAttackState(USR_SYN_SET, "<<>>", "{}", "Connecting", USR_SYN_SET, ATK_SYN_SEQ, "1", "<<>>", ATK_SYN_SENT, SPA_AUTH_SET, "{}", "{}", "{}", "<<>>", "1", "1", ACL_WILD, USR_TCP, "{}", SPA_AUTH_SET, "<<>>"),
		javaExamplesSDPAttackState(USR_SYN_SET, ATK_SYN_SEQ, "{}", "Connecting", USR_SYN_SET, "<<>>", "1", "<<>>", ATK_SYN_SENT, SPA_AUTH_SET, "{}", "{}", "{}", "<<>>", "1", "1", ACL_TWO, USR_TCP, "{}", SPA_AUTH_SET, "<<>>"),
		javaExamplesSDPAttackState(USR_SYN_SET, "<<>>", sTCPLink10, "Connecting", USR_SYN_SET, "<<>>", "1", aChannel10, ATK_SYN_SENT, SPA_AUTH_SET, "{}", "{}", "{}", "<<>>", "1", "1", ACL_TWO, USR_TCP, "{}", SPA_AUTH_SET, "<<>>"),
		javaExamplesSDPAttackState(USR_SYN_SET, "<<>>", sTCPLink10, "Connecting", USR_SYN_SET, ATK_ACK_SEQ, "1", "<<>>", atkEstablished, SPA_AUTH_SET, "{}", "{}", "{}", "<<>>", "1", "1", ACL_TWO, USR_TCP, "{}", SPA_AUTH_SET, "<<>>"),
	}
	requireJavaRandomSubsetTrace(t, r, expectedTrace, false)
}

// Original ExamplesMCDistributedReplicatedLogTest.testSpec. This is a local
// finite-state liveness model using ModelCheckerTestCase's single worker.
func TestJavaExamplesMCDistributedReplicatedLog(t *testing.T) {
	setJavaExampleClasspath(t)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "ExamplesMCDistributedReplicatedLog", "MCDistributedReplicatedLog", false, false, false, 1, "-noGenerateSpecTE", "-lncheck", "final")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "271", "37", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "InSync")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	expectedTrace := []string{
		"cLogs = (n1 :> <<>> @@ n2 :> <<>> @@ n3 :> <<>>)",
		"cLogs = (n1 :> <<>> @@ n2 :> <<v, v, v>> @@ n3 :> <<>>)",
		"cLogs = (n1 :> <<>> @@ n2 :> <<v, v, v>> @@ n3 :> <<v, v, v>>)",
		"cLogs = (n1 :> <<v, v>> @@ n2 :> <<v, v, v>> @@ n3 :> <<v, v, v>>)",
		"cLogs = (n1 :> <<v, v>> @@ n2 :> <<v, v, v, v, v>> @@ n3 :> <<v, v, v>>)",
		"cLogs = (n1 :> <<v, v, v>> @@ n2 :> <<v, v, v, v, v>> @@ n3 :> <<v, v, v>>)",
	}
	requireJavaRandomSubsetTrace(t, r, expectedTrace, false)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "2", "<Extend(n2) line 35, col 5 to line 38, col 49 of module DistributedReplicatedLog>")
}
