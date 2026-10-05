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

// Original Test50: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest50(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest50", "test50", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test50.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test51: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest51(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest51", "test51", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test51.json"))
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

// Original Test52: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest52(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest52", "test52", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test52.json"))
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
	if len(javaTLCRecords(r, tlc.ECTLCCoverageMismatch)) != 0 {
		t.Fatal("TLC_COVERAGE_MISMATCH present")
	}
	assertJavaCoverage(t, r, "<Init line 33, col 1 to line 33, col 4 of module test52>: 1:1\n  line 33, col 9 to line 34, col 20 of module test52: 1\n<Next line 35, col 1 to line 35, col 4 of module test52>: 1:2\n  line 35, col 12 to line 35, col 27 of module test52: 2\n  |line 35, col 16 to line 35, col 27 of module test52: 2\n  ||line 35, col 16 to line 35, col 16 of module test52: 2\n  ||line 35, col 23 to line 35, col 27 of module test52: 2:4\n  line 36, col 12 to line 36, col 27 of module test52: 2\n  |line 36, col 16 to line 36, col 27 of module test52: 2\n  ||line 36, col 16 to line 36, col 16 of module test52: 2\n  ||line 36, col 23 to line 36, col 27 of module test52: 2:4\n<Action line 50, col 15 to line 50, col 38 of module test52>\n  line 50, col 15 to line 50, col 38 of module test52: 2\n  |line 50, col 16 to line 50, col 28 of module test52: 2\n  ||line 47, col 2 to line 48, col 16 of module test52: 2\n  |||line 47, col 5 to line 47, col 28 of module test52: 2\n  ||||line 40, col 26 to line 40, col 31 of module test52: 2\n  |||||line 40, col 30 to line 40, col 30 of module test52: 4\n  |||line 48, col 5 to line 48, col 16 of module test52: 2")
}

// Original Test53: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest53(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest53", "test53", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test53.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "10", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test54: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest54(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest54", "test54", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test54.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "10", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Test55: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest55(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest55", "test55", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test55.json"))
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
	if len(javaTLCRecords(r, tlc.ECTLCCoverageMismatch)) != 0 {
		t.Fatal("TLC_COVERAGE_MISMATCH present")
	}
	assertJavaCoverage(t, r, "<Init line 9, col 1 to line 9, col 4 of module test55>: 1:1\n  line 9, col 9 to line 9, col 13 of module test55: 1\n<Next line 10, col 1 to line 10, col 4 of module test55>: 1:2\n  line 10, col 9 to line 10, col 18 of module test55: 2\n<Invariant line 13, col 1 to line 13, col 9 of module test55>\n  line 13, col 14 to line 14, col 20 of module test55: 2\n<Action line 7, col 1 to line 7, col 41 of module test55>\n  line 7, col 33 to line 7, col 33 of module test55: 1\n  line 8, col 10 to line 8, col 14 of module test55a: 1\n<Action line 7, col 1 to line 7, col 41 of module test55>\n  line 7, col 41 to line 7, col 41 of module test55: 1\n  line 8, col 21 to line 8, col 25 of module test55a: 1")
}

// Original Test56: whole inherited SuiteTestCase.testSpec and constructor.
func TestJavaLegacySuiteTest56(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest56", "test56", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Test56.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "6", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "3")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
	if len(javaTLCRecords(r, tlc.ECTLCCoverageMismatch)) != 0 {
		t.Fatal("TLC_COVERAGE_MISMATCH present")
	}
	assertJavaCoverage(t, r, "<Init line 10, col 1 to line 10, col 4 of module test56>: 3:3\n  line 10, col 25 to line 10, col 29 of module test56: 3\n  line 10, col 18 to line 10, col 21 of module test56: 1\n<Next line 11, col 1 to line 11, col 4 of module test56>: 3:6\n  line 11, col 9 to line 11, col 16 of module test56: 6\n<Action line 14, col 15 to line 14, col 33 of module test56>\n  line 14, col 15 to line 14, col 33 of module test56: 6\n  |line 14, col 16 to line 14, col 30 of module test56: 6\n  ||line 14, col 16 to line 14, col 18 of module test56: 6\n  ||line 14, col 22 to line 14, col 30 of module test56: 6\n  |||line 8, col 14 to line 8, col 37 of module test56: 6\n  ||||line 8, col 15 to line 8, col 18 of module test56: 6\n  ||||line 8, col 24 to line 8, col 37 of module test56: 12")
}
