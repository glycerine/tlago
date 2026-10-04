/*******************************************************************************
 * Copyright (c) 2018, 2019, 2021 Microsoft Research. All rights reserved.
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
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
 *   Finn Hackett - initial implementation
 ******************************************************************************/
package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original ConstantContextTLCCacheTest.test, including inherited successful exit.
func TestJavaConstantContextTLCCache(t *testing.T) {
	result := runJavaTLCModelTest(t, "ConstantContextTLCCache", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.module.ConstantContextTLCCacheTest.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "2", "1", "0")
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("unexpected GENERAL: %v", records)
	}
}

// Original TLCExtTest.test, including its runWithDebugger=false override,
// embedded configuration and inherited successful exit.
func TestJavaTLCExtModel(t *testing.T) {
	result := runJavaTLCModelTestWithDebugger(t, "TLCExtTest", true, true, false, 1, "-config", "TLCExtTest.tla", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.module.TLCExtTest.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "2", "1", "0")
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("unexpected GENERAL: %v", records)
	}
}

// Original BagsTest.testSpec and inherited successful exit.
func TestJavaBagsModel(t *testing.T) {
	result := runJavaTLCModelTest(t, "BagsTest", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.BagsTest.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("unexpected GENERAL: %v", records)
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "0", "0", "0")
}

// Original ConstantRank1TLCEvalTest.testSpec; both source overrides retained.
func TestJavaConstantRank1TLCEval(t *testing.T) {
	result := runJavaTLCModelTest(t, "ConstantRank1TLCEval", "-config", "ConstantRank1TLCEval.tla", "-noGenerateSpecTE")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "3", "2", "0")
}

// Original ConstantRank2AssertErrorTest.testSpec; both source overrides retained.
func TestJavaConstantRank2AssertError(t *testing.T) {
	result := runJavaTLCModelTest(t, "ConstantRank2AssertError", "-config", "ConstantRank2AssertError.tla", "-noGenerateSpecTE")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "3", "2", "0")
}

// Original EmptySetEqAssumeTest.testSpec and noGenerateSpec override.
func TestJavaEmptySetEqAssume(t *testing.T) {
	result := runJavaTLCModelTest(t, "EmptySetEqAssume", "-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.EmptySetEqAssumeTest.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECTLCAssumptionFalse, tlc.ECTLCAssumptionEvaluationError, tlc.ECGeneral} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Fatalf("unexpected code %d: %v", code, records)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "0", "0", "0")
}

// Original EmptySetEqStatesTest.testSpec, including deadlock checking and
// the full zero-uncovered assertion; the original 34 forms stay in the fixture.
func TestJavaEmptySetEqStates(t *testing.T) {
	result := runJavaTLCModelTestWithDeadlock(t, "EmptySetEqStates", "EmptySetEqStates", true, true, true, 1, true,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.EmptySetEqStatesTest.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCDeadlockReached)) != 0 {
		t.Fatal("TLC_DEADLOCK_REACHED recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCInvariantViolatedBehavior)) != 0 {
		t.Fatal("TLC_INVARIANT_VIOLATED_BEHAVIOR recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "38", "4", "0")
	requireJavaTLCUncovered(t, result)
}

// Original EmptySetEqStatesRcdTest.testRcdSpec, including deadlock checking,
// record configuration and the full zero-uncovered assertion.
func TestJavaEmptySetEqStatesRcd(t *testing.T) {
	result := runJavaTLCModelTestWithDeadlock(t, "EmptySetEqStates", "EmptySetEqStates", true, true, true, 1, true,
		"-config", "EmptySetEqStatesRcd.cfg", "-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.EmptySetEqStatesRcdTest.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCDeadlockReached)) != 0 {
		t.Fatal("TLC_DEADLOCK_REACHED recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCInvariantViolatedBehavior)) != 0 {
		t.Fatal("TLC_INVARIANT_VIOLATED_BEHAVIOR recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "3", "1", "0")
	requireJavaTLCUncovered(t, result)
}

// Original KSubsetAssumeTest.testSpec; all original Cases assumptions and
// presentation/error assertions are evaluated from the unchanged models.
func TestJavaKSubsetAssume(t *testing.T) {
	// Java's tool-jar manifest supplies the original CommunityModules archive.
	modules, err := filepath.Abs(filepath.Join("test_vectors", "java-sany", "CommunityModules.jar"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLASSPATH", modules)
	result := runJavaTLCModelTest(t, "KSubsetAssume", "-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.KSubsetAssumeTest.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECTLCAssumptionFalse, tlc.ECTLCAssumptionEvaluationError, tlc.ECGeneral} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Fatalf("unexpected code %d: %v", code, records)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "0", "0", "0")
}
