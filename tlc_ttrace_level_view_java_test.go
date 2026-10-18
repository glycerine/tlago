/*******************************************************************************
 * Copyright (c) 2017, 2019 Microsoft Research. All rights reserved.
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

// Original TLCGetLevelTest_TTraceTest.testSpec and inherited settings.
func TestJavaTLCGetLevelTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "TLCGetLevelTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaTLCGetLevel(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "TLCGetLevel", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "4", "4", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ yb = 0\n/\\ x = 0\n/\\ y = 0\n/\\ z = 1",
		"/\\ yb = 1\n/\\ x = 1\n/\\ y = 1\n/\\ z = 2",
		"/\\ yb = 2\n/\\ x = 2\n/\\ y = 2\n/\\ z = 3",
		"/\\ yb = 3\n/\\ x = 3\n/\\ y = 3\n/\\ z = 4",
	}, true)
	stutter := javaTLCRecords(r, tlc.ECTLCStatePrint3)
	if len(stutter) == 0 || stutter[0].StateNumber != 5 {
		t.Fatalf("stuttering=%v, want state5", stutter)
	}
}

// Original TraceWithLargeSetOfInitialStatesTest_TTraceTest.testSpec and maxSetSize.
func TestJavaTraceWithLargeSetOfInitialStatesTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "TraceWithLargeSetOfInitialStatesTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaTraceWithLargeSetOfInitialStates(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "TraceWithLargeSetOfInitialStatesTest", generated, false, true, "-maxSetSize", "10")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	requireJavaTTraceActions(t, r, []string{"/\\ x = 1\n/\\ y = FALSE", "/\\ x = 1\n/\\ y = TRUE"}, "<_init line 25, col 5 to line 26, col 24 of module TraceWithLargeSetOfInitialStatesTestTTrace>", "<_next line 30, col 5 to line 36, col 29 of module TraceWithLargeSetOfInitialStatesTestTTrace>")
	requireJavaTLCUncovered(t, r)
}

// Original ViewMapTest_TTraceTest.testSpec, with -view and all eight full states.
func TestJavaViewMapTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "ViewMapTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaViewMap(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "ViewMap", generated, false, true, "-view")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	requireJavaTTraceActions(t, r, []string{
		"/\\ buffer = <<>>\n/\\ waitset = {}\n/\\ pc = (c1 :> \"lbc\" @@ c2 :> \"lbc\" @@ p1 :> \"lbp\")",
		"/\\ buffer = <<>>\n/\\ waitset = {c1}\n/\\ pc = (c1 :> \"lbc\" @@ c2 :> \"lbc\" @@ p1 :> \"lbp\")",
		"/\\ buffer = <<>>\n/\\ waitset = {c1, c2}\n/\\ pc = (c1 :> \"lbc\" @@ c2 :> \"lbc\" @@ p1 :> \"lbp\")",
		"/\\ buffer = <<\"d\">>\n/\\ waitset = {c2}\n/\\ pc = (c1 :> \"lbc\" @@ c2 :> \"lbc\" @@ p1 :> \"lbp\")",
		"/\\ buffer = <<\"d\">>\n/\\ waitset = {c2, p1}\n/\\ pc = (c1 :> \"lbc\" @@ c2 :> \"lbc\" @@ p1 :> \"lbp\")",
		"/\\ buffer = <<>>\n/\\ waitset = {p1}\n/\\ pc = (c1 :> \"lbc\" @@ c2 :> \"lbc\" @@ p1 :> \"lbp\")",
		"/\\ buffer = <<>>\n/\\ waitset = {c1, p1}\n/\\ pc = (c1 :> \"lbc\" @@ c2 :> \"lbc\" @@ p1 :> \"lbp\")",
		"/\\ buffer = <<>>\n/\\ waitset = {c1, c2, p1}\n/\\ pc = (c1 :> \"lbc\" @@ c2 :> \"lbc\" @@ p1 :> \"lbp\")",
	}, "<_init line 27, col 5 to line 29, col 26 of module ViewMapTestTTrace>", "<_next line 33, col 5 to line 41, col 31 of module ViewMapTestTTrace>")
}
