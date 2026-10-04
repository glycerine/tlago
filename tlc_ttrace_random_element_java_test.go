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

// Original RandomElementTest_TTraceTest.test and inherited settings.
func TestJavaRandomElementTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "RandomElementTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaRandomElement(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "RandomElement", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "11", "11", "0")
	expected := []string{"/\\ x = 843\n/\\ y = 0",
		"/\\ x = 920\n/\\ y = 1",
		"/\\ x = 483\n/\\ y = 2",
		"/\\ x = 173\n/\\ y = 3",
		"/\\ x = 590\n/\\ y = 4",
		"/\\ x = 104\n/\\ y = 5",
		"/\\ x = 785\n/\\ y = 6",
		"/\\ x = 463\n/\\ y = 7",
		"/\\ x = 443\n/\\ y = 8",
		"/\\ x = 151\n/\\ y = 9",
		"/\\ x = 767\n/\\ y = 10",
	}
	requireJavaRandomSubsetTrace(t, r, expected, true)
	requireJavaTLCUncovered(t, r)
}

// Original RandomElementXandYTest_TTraceTest.test and inherited settings.
func TestJavaRandomElementXandYTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "RandomElementXandYTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaRandomElementXandY(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "RandomElementXandY", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	expected := []string{"/\\ x = 0\n/\\ y = 0",
		"/\\ x = 1\n/\\ y = 1",
		"/\\ x = 0\n/\\ y = 1",
	}
	requireJavaRandomSubsetTrace(t, r, expected, true)
	requireJavaTLCUncovered(t, r)
}

// Original RandomElementSimulationTest_TTraceTest.test and inherited settings.
func TestJavaRandomElementSimulationTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "RandomElementSimulationTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaRandomElementSimulation(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "RandomElement", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	expected := []string{"/\\ x = 843\n/\\ y = 0",
		"/\\ x = 843\n/\\ y = 1",
		"/\\ x = 227\n/\\ y = 2",
		"/\\ x = 227\n/\\ y = 3",
		"/\\ x = 502\n/\\ y = 4",
		"/\\ x = 535\n/\\ y = 5",
		"/\\ x = 277\n/\\ y = 6",
		"/\\ x = 327\n/\\ y = 7",
		"/\\ x = 729\n/\\ y = 8",
		"/\\ x = 550\n/\\ y = 9",
		"/\\ x = 318\n/\\ y = 10",
	}
	requireJavaTTraceActions(t, r, expected, "<_init line 25, col 5 to line 26, col 24 of module RandomElementSimulationTestTTrace>", "<_next line 30, col 5 to line 36, col 29 of module RandomElementSimulationTestTTrace>")
	requireJavaTLCUncovered(t, r)
}

// Original RandomElementT4Test_TTraceTest.test and inherited settings.
func TestJavaRandomElementT4TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "RandomElementT4TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaRandomElementT4(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheckWithWorkers(t, "RandomElement", generated, false, true, false, 4)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	records := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(records) != 11 {
		t.Fatalf("trace length=%d, want 11", len(records))
	}
	cnt := 0
	for _, record := range records {
		vals := record.StateInfo.State.GetVals()
		y, _ := vals.Get2(tlc.UniqueStringOf("y"))
		if got := y.(*tlc.IntValue).Val; got != int32(cnt) {
			t.Fatalf("y=%d, want %d", got, cnt)
		}
		cnt++
		x, _ := record.StateInfo.State.GetVals().Get2(tlc.UniqueStringOf("x"))
		if got := x.(*tlc.IntValue).Val; got < 1 || got > 1000 {
			t.Fatalf("x=%d outside 1..1000", got)
		}
		if record.StateNumber != cnt {
			t.Fatalf("ordinal=%d, want %d", record.StateNumber, cnt)
		}
	}
}
