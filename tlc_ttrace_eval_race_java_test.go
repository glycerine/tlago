/*******************************************************************************
 * Copyright (c) 2021 Microsoft Research. All rights reserved.
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

// Original EvalExceptionTest_TTraceTest.testSpec. Recheck the generated error trace.
func TestJavaEvalExceptionTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "EvalExceptionTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaEvalException(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "DistBakery", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "6", "6", "0")
	expectedTrace := []string{
		"/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 0))\n/\\ net = (-2 :> (-1 :> <<>>) @@ -1 :> (-2 :> <<>>))\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"ncs\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"ncs\" )",
		"/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 0))\n/\\ net = (-2 :> (-1 :> <<>>) @@ -1 :> (-2 :> <<>>))\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"enter\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"ncs\" )",
		"/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 0))\n/\\ net = (-2 :> (-1 :> <<[type |-> \"write\", num |-> 1]>>) @@ -1 :> (-2 :> <<>>))\n/\\ acks = (-2 :> {-2} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"e1\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"ncs\" )",
		"/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 0))\n/\\ net = (-2 :> (-1 :> <<[type |-> \"write\", num |-> 1]>>) @@ -1 :> (-2 :> <<>>))\n/\\ acks = (-2 :> {-2} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"e1\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"enter\" )",
		"/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 1))\n/\\ net = (-2 :> (-1 :> <<>>) @@ -1 :> (-2 :> <<[type |-> \"ack\"]>>))\n/\\ acks = (-2 :> {-2} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"e1\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"enter\" )",
		"/\\ num = (-2 :> 1 @@ -1 :> (-1 :> 0))\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 1))\n/\\ net = ( -2 :> (-1 :> <<>>) @@\n  -1 :> (-2 :> <<[type |-> \"ack\"], [type |-> \"write\", num |-> (-1 :> 0)]>>) )\n/\\ acks = (-2 :> {-2} @@ -1 :> {-1})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"e1\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"e1\" )",
	}
	requireJavaRandomSubsetTrace(t, r, expectedTrace, true)
}

// Original PrintTraceRaceTest_TTraceTest.testSpec and its four-worker override.
func TestJavaPrintTraceRaceTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "PrintTraceRaceTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaPrintTraceRace(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheckWithWorkers(t, "PrintTraceRace", generated, false, true, false, 4)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	records := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	expectedTrace := []string{"S = [q |-> <<>>, i |-> 1]", "S = [q |-> <<1>>, i |-> 2]"}
	if len(records) < len(expectedTrace) {
		t.Fatalf("trace length=%d, want at least two records", len(records))
	}
	for i, want := range expectedTrace {
		record := records[i]
		if record.StateInfo == nil {
			t.Fatalf("trace state %d has no TLCStateInfo", i+1)
		}
		if got := replJavaTrim(record.StateInfo.String()); got != want {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, want)
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
	// Java's Object[2] payload is represented by StateInfo and StateNumber.
}
