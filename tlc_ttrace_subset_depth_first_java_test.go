/*******************************************************************************
 * Copyright (c) 2015, 2021 Microsoft Research. All rights reserved.
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
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original BugzillaBug279Test_TTraceTest.testSpec and its deadlock/no-DOT overrides.
func TestJavaBugzillaBug279TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "BugzillaBug279TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaBugzillaBug279(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheckWithDeadlock(t, "BugzillaBug279", generated, false, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	subset, err := tlc.NewSubsetValue(tlc.NewIntervalValue(1, 8)).ToSetEnum()
	if err != nil {
		t.Fatal(err)
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ set = {}\n/\\ pc = 0\n/\\ fun = {}",
		"/\\ set = SUBSET (1..20)\n/\\ pc = 1\n/\\ fun = {5}",
		"/\\ set = " + tlc.ValuesPPR(subset.Normalize()) + "\n/\\ pc = 2\n/\\ fun = {5}",
	}, true)
	requireJavaTLCUncovered(t, r)
}

// CommonTestCase.assertTraceWith with the original explicit action list.
func requireJavaTTraceActions(t *testing.T, r *tlc.Result, expectedTrace []string, init, next string) {
	t.Helper()
	requireJavaRandomSubsetTrace(t, r, expectedTrace, true)
	for i, record := range javaTLCRecords(r, tlc.ECTLCStatePrint2) {
		expected := next
		if i == 0 {
			expected = init
		}
		if got := record.StateInfo.Info; got != expected {
			t.Fatalf("trace action %d=%v, want %q", i+1, got, expected)
		}
	}
}

// Original DepthFirstDieHardTest_TTraceTest.testSpec. The first phase retains DFID.
func TestJavaDepthFirstDieHardTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "DepthFirstDieHardTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaDepthFirstDieHard(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "DieHard", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCStatePrint1} {
		if len(javaTLCRecords(r, code)) != 0 {
			t.Fatalf("unexpected diagnostic %d", code)
		}
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	expectedTrace := []string{
		"/\\ action = \"nondet\"\n/\\ smallBucket = 0\n/\\ bigBucket = 0\n/\\ water_to_pour = 0",
		"/\\ action = \"fill big\"\n/\\ smallBucket = 0\n/\\ bigBucket = 5\n/\\ water_to_pour = 0",
		"/\\ action = \"pour big to small\"\n/\\ smallBucket = 3\n/\\ bigBucket = 2\n/\\ water_to_pour = 3",
		"/\\ action = \"empty small\"\n/\\ smallBucket = 0\n/\\ bigBucket = 2\n/\\ water_to_pour = 3",
		"/\\ action = \"pour big to small\"\n/\\ smallBucket = 2\n/\\ bigBucket = 0\n/\\ water_to_pour = 2",
		"/\\ action = \"fill big\"\n/\\ smallBucket = 2\n/\\ bigBucket = 5\n/\\ water_to_pour = 2",
		"/\\ action = \"pour big to small\"\n/\\ smallBucket = 3\n/\\ bigBucket = 4\n/\\ water_to_pour = 1",
	}
	requireJavaTTraceActions(t, r, expectedTrace, "<_init line 29, col 5 to line 32, col 48 of module DepthFirstDieHardTestTTrace>", "<_next line 36, col 5 to line 46, col 53 of module DepthFirstDieHardTestTTrace>")
	requireJavaTLCUncovered(t, r)
}

// Original DepthFirstErrorTraceTest_TTraceTest.testSpec. The first phase retains DFID.
func TestJavaDepthFirstErrorTraceTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "DepthFirstErrorTraceTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaDepthFirstErrorTrace(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "DepthFirstErrorTrace", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCStatePrint1} {
		if len(javaTLCRecords(r, code)) != 0 {
			t.Fatalf("unexpected diagnostic %d", code)
		}
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	expectedTrace := []string{
		"x = 0",
		"x = 1",
		"x = 2",
		"x = 3",
		"x = 4",
		"x = 5",
		"x = 6",
		"x = 7",
	}
	requireJavaTTraceActions(t, r, expectedTrace, "<_init line 23, col 5 to line 23, col 24 of module DepthFirstErrorTraceTestTTrace>", "<_next line 27, col 5 to line 31, col 29 of module DepthFirstErrorTraceTestTTrace>")
	requireJavaTLCUncovered(t, r)
}
