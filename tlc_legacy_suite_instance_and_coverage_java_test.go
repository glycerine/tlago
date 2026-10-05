/*******************************************************************************
 * Copyright (c) 2016 Microsoft Research. All rights reserved.
 * Copyright (c) 2026 NVIDIA Corporation. All rights reserved.
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
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/

package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Test57.testSpec overrides SuiteTestCase.testSpec entirely.
func TestJavaLegacySuiteTest57(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest57", "test57", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test57.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCModeMC)) == 0 {
		t.Fatal("TLC_MODE_MC absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCComputingInit)) == 0 {
		t.Fatal("TLC_COMPUTING_INIT absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInvariantViolatedInitial, "Invariant1", "x = 1\n")
}

// Original Test58: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest58(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest58", "test58", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test58.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test59: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest59(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest59", "test59", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test59.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "6", "5", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test60: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest60(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest60", "test60", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test60.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test62: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest62(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest62", "test62", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test62.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test63: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest63(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest63", "test63", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test63.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "696", "216", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "72")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
	if len(javaTLCRecords(r, tlc.ECTLCCoverageMismatch)) != 0 {
		t.Fatal("TLC_COVERAGE_MISMATCH present")
	}
	assertJavaCoverage(t, r, "<BigInit line 37, col 1 to line 37, col 7 of module test63>: 72:72\n  line 37, col 15 to line 37, col 19 of module test63: 1\n  line 39, col 15 to line 39, col 26 of module test63: 72\n  |line 39, col 23 to line 39, col 26 of module test63: 12\n<BigNext line 24, col 1 to line 24, col 7 of module test63>: 144:624\n  line 21, col 15 to line 21, col 45 of module test63: 264\n  |line 21, col 24 to line 21, col 45 of module test63: 216:624\n  ||line 21, col 38 to line 21, col 44 of module test63: 1296\n  ||line 21, col 31 to line 21, col 34 of module test63: 216\n  line 22, col 15 to line 22, col 26 of module test63: 264\n  line 24, col 25 to line 24, col 27 of module test63: 216\n  line 5, col 12 to line 5, col 46 of module test63a: 144\n  |line 5, col 18 to line 5, col 46 of module test63a: 624\n  ||line 5, col 21 to line 5, col 27 of module test63a: 624\n  ||line 5, col 34 to line 5, col 39 of module test63a: 572\n  line 25, col 23 to line 25, col 24 of module test63: 480\n  line 14, col 13 to line 14, col 50 of module test63: 624\n  |line 14, col 18 to line 14, col 50 of module test63: 840\n  ||line 14, col 21 to line 14, col 25 of module test63: 840\n  |||line 5, col 12 to line 5, col 46 of module test63a: 840\n  ||||line 5, col 12 to line 5, col 14 of module test63a: 840\n  ||||line 5, col 18 to line 5, col 46 of module test63a: 840\n  |||||line 5, col 21 to line 5, col 27 of module test63a: 840\n  |||||line 5, col 34 to line 5, col 39 of module test63a: 770\n  ||line 14, col 39 to line 14, col 50 of module test63: 624\n  line 27, col 15 to line 27, col 50 of module test63: 840\n  |line 27, col 15 to line 27, col 19 of module test63: 840\n  ||line 5, col 12 to line 5, col 46 of module test63a: 840\n  |||line 5, col 12 to line 5, col 14 of module test63a: 840\n  |||line 5, col 18 to line 5, col 46 of module test63a: 840\n  ||||line 5, col 21 to line 5, col 27 of module test63a: 840\n  ||||line 5, col 34 to line 5, col 39 of module test63a: 770\n  |line 27, col 24 to line 27, col 50 of module test63: 216\n  line 28, col 15 to line 28, col 43 of module test63: 768\n<TypeOK line 49, col 1 to line 49, col 6 of module test63>\n  line 49, col 11 to line 50, col 25 of module test63: 216\n<Action line 52, col 1 to line 52, col 21 of module test63>\n  line 52, col 1 to line 52, col 21 of module test63: 72\n  line 4, col 12 to line 4, col 27 of module test63a: 72\n<Action line 52, col 1 to line 52, col 21 of module test63>\n  line 52, col 1 to line 52, col 21 of module test63: 2208\n  line 6, col 20 to line 6, col 29 of module test63a: 624\n  |line 6, col 21 to line 6, col 25 of module test63a: 624\n  ||line 5, col 12 to line 5, col 46 of module test63a: 624\n  |||line 5, col 12 to line 5, col 14 of module test63a: 624\n  |||line 5, col 18 to line 5, col 46 of module test63a: 624\n  ||||line 5, col 21 to line 5, col 27 of module test63a: 624\n  ||||line 5, col 34 to line 5, col 39 of module test63a: 572\n  |line 6, col 28 to line 6, col 29 of module test63a: 960")
}

// Original Test63a: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest63a(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest63a", "test63a", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test63a.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "24", "12", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "12")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test64: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest64(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest64", "test64", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test64.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test64a: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest64a(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest64a", "test64a", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test64a.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test65: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest65(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest65", "test65", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test65.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test65a: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest65a(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest65a", "test65a", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test65a.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}
