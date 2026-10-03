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
// Port of EvalExceptionTest.testSpec and its coverage-disabled constructor.
package tlago

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaEvalException(t *testing.T) {
	result := runJavaTLCModelTestWithCoverage(t, "DistBakery", false, "-config", "DistBakery.tla")
	if result.ExitStatus != tlc.ExitStatusError {
		t.Fatalf("exit status=%d, want ERROR", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "24", "17", "5")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCModuleArgumentErrorAn, "first", "<=", "integer", "(-1 :> 0)")
	expectedTrace := []string{
		"/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 0))\n/\\ net = (-2 :> (-1 :> <<>>) @@ -1 :> (-2 :> <<>>))\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"ncs\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"ncs\" )",
		"/\\ num = (-2 :> 0 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 0))\n/\\ net = (-2 :> (-1 :> <<>>) @@ -1 :> (-2 :> <<>>))\n/\\ acks = (-2 :> {} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"enter\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"ncs\" )",
		"/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 0))\n/\\ net = (-2 :> (-1 :> <<[type |-> \"write\", num |-> 1]>>) @@ -1 :> (-2 :> <<>>))\n/\\ acks = (-2 :> {-2} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"e1\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"ncs\" )",
		"/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 0))\n/\\ net = (-2 :> (-1 :> <<[type |-> \"write\", num |-> 1]>>) @@ -1 :> (-2 :> <<>>))\n/\\ acks = (-2 :> {-2} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"e1\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"enter\" )",
		"/\\ num = (-2 :> 1 @@ -1 :> 0)\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 1))\n/\\ net = (-2 :> (-1 :> <<>>) @@ -1 :> (-2 :> <<[type |-> \"ack\"]>>))\n/\\ acks = (-2 :> {-2} @@ -1 :> {})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"e1\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"enter\" )",
		"/\\ num = (-2 :> 1 @@ -1 :> (-1 :> 0))\n/\\ rnum = (-2 :> (-1 :> 0) @@ -1 :> (-2 :> 1))\n/\\ net = ( -2 :> (-1 :> <<>>) @@\n  -1 :> (-2 :> <<[type |-> \"ack\"], [type |-> \"write\", num |-> (-1 :> 0)]>>) )\n/\\ acks = (-2 :> {-2} @@ -1 :> {-1})\n/\\ pc = ( [node |-> -2, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -2, type |-> \"mutex\"] :> \"e1\" @@\n  [node |-> -1, type |-> \"msg\"] :> \"a\" @@\n  [node |-> -1, type |-> \"mutex\"] :> \"e1\" )",
	}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != len(expectedTrace) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expectedTrace))
	}
	for i, record := range records {
		if record.StateInfo == nil {
			t.Fatalf("trace state %d has no TLCStateInfo", i+1)
		}
		info, ok := record.StateInfo.Info.(string)
		if !ok || info == tlc.InitialPredicate || strings.HasPrefix(info, "<Action") {
			t.Fatalf("trace action %d=%v, want a named action", i+1, record.StateInfo.Info)
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCNestedExpression,
		"0. Line 209, column 5 to line 209, column 32 in DistBakery\n1. Line 209, column 22 to line 209, column 32 in DistBakery\n\n")
}
