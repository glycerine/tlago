/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
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

// Port of PrintTraceRaceTest.testSpec with the original four workers.
package tlago

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaPrintTraceRace(t *testing.T) {
	result := runJavaTLCModelTestWithRoot(t, "PrintTraceRace", "MC", true, true, true, 4)
	if result.ExitStatus != tlc.ExitStatusFailureSafetyEval {
		t.Fatalf("exit status=%d, want safety evaluation failure", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "2", "2", "0")
	if got := javaTLCRecords(result, tlc.ECGeneral); len(got) != 0 {
		t.Fatalf("unexpected GENERAL: %v", got)
	}
	if len(javaTLCRecords(result, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT not recorded")
	}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	expectedTrace := []string{"S = [q |-> <<>>, i |-> 1]", "S = [q |-> <<1>>, i |-> 2]"}
	if len(records) < len(expectedTrace) {
		t.Fatalf("trace length=%d, want at least two records", len(records))
	}
	for i, want := range expectedTrace {
		record := records[i]
		if record.StateInfo == nil {
			t.Fatalf("trace state %d has no TLCStateInfo", i+1)
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != want {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, want)
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
	// Java's Object[2] state payload is represented by StateInfo/StateNumber.
	requireJavaTLCUncovered(t, result, "line 15, col 12 to line 15, col 28 of module PrintTraceRace")
}
