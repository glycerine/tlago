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
	"strings"
	"testing"
)

// Original ETest3.testSpec, VIOLATION_DEADLOCK constructor and
// checkDeadLock=true, with all inherited SuiteETestCase runner settings.
func TestJavaLegacySuiteETest3(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithDeadlock(t, "LegacySuiteETest3", "etest3", true, true, true, 1, true,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "ETest3.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationDeadlock {
		t.Fatalf("exit=%d, want VIOLATION_DEADLOCK; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCDeadlockReached)) == 0 {
		t.Fatal("TLC_DEADLOCK_REACHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original ETest4.testSpec and FAILURE_SPEC_EVAL constructor, with all inherited
// SuiteETestCase runner settings and the exact nested-expression stack.
func TestJavaLegacySuiteETest4(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteETest4", "etest4", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "ETest4.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FAILURE_SPEC_EVAL; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	stack := "0. Line 13, column 9 to line 16, column 51 in etest4\n" +
		"1. Line 13, column 12 to line 13, column 16 in etest4\n" +
		"2. Line 15, column 12 to line 15, column 26 in etest4\n" +
		"3. Line 15, column 12 to line 15, column 22 in etest4\n\n"
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCNestedExpression, stack)
	// This helper compares zero-count coverage records by their exact locations.
	requireJavaTLCUncovered(t, r, "line 18, col 9 to line 18, col 19 of module etest4")
}

// Original ETest6.testSpec and FAILURE_SPEC_EVAL constructor, with all inherited
// SuiteETestCase runner settings and both zero-count uncovered locations.
func TestJavaLegacySuiteETest6(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteETest6", "etest6", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "ETest6.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FAILURE_SPEC_EVAL; messages=%v", r.ExitStatus, r.Messages)
	}
	substring := "In evaluation, the identifier x is either undefined or not an operator.\nline 16, col 23 to line 16, col 23 of module etest6"
	// TestMPRecorder.recordedWithSubStringValue checks parameters of its first record.
	records := javaTLCRecords(r, tlc.ECGeneral)
	found := false
	if len(records) > 0 {
		for _, parameter := range records[0].Params {
			if strings.Contains(parameter, substring) {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("first GENERAL lacks %q: %v", substring, records)
	}
	requireJavaTLCUncovered(t, r,
		"line 13, col 14 to line 13, col 43 of module etest6",
		"line 19, col 13 to line 19, col 23 of module etest6")
}
