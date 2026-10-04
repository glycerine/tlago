/*******************************************************************************
 * Copyright (c) 2018 Microsoft Research. All rights reserved.
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

// Original BugzillaBug279Test.testSpec and its deadlock/dump overrides.
func TestJavaBugzillaBug279(t *testing.T) {
	r := runJavaTLCModelTestWithDeadlock(t, "BugzillaBug279", "InitStateBug", true, false, true, 1, true,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "BugzillaBug279Test.json"))
	if r.ExitStatus != tlc.ExitStatusViolationDeadlock {
		t.Fatalf("exit=%d, want VIOLATION_DEADLOCK", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCDeadlockReached)) == 0 {
		t.Fatal("TLC_DEADLOCK_REACHED absent")
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
