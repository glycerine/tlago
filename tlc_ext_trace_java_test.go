/*******************************************************************************
 * Copyright (c) 2021, 2022 Microsoft Research. All rights reserved.
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
// Ports of TLCExtTraceTest.test and TLCExtTraceAliasTest.test.
package tlago

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaTLCExtTrace(t *testing.T) {
	result := runJavaTLCModelTest(t, "TLCExtTrace", "-config", "TLCExtTrace.tla")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "10")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "10", "10", "0")
	if got := javaTLCRecords(result, tlc.ECGeneral); len(got) != 0 {
		t.Fatalf("unexpected GENERAL: %v", got)
	}
}

func TestJavaTLCExtTraceAlias(t *testing.T) {
	result := runJavaTLCModelTest(t, "TLCExtTrace", "-config", "TLCExtTraceAlias.cfg")
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want safety violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "7")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "7", "7", "0")
	if got := javaTLCRecords(result, tlc.ECGeneral); len(got) != 0 {
		t.Fatalf("unexpected GENERAL: %v", got)
	}
	expectedTrace := []string{
		"/\\ x = 1\n/\\ t = <<[x |-> 1]>>",
		"/\\ x = 2\n/\\ t = <<[x |-> 1], [x |-> 2]>>",
		"/\\ x = 3\n/\\ t = <<[x |-> 1], [x |-> 2], [x |-> 3]>>",
		"/\\ x = 4\n/\\ t = <<[x |-> 1], [x |-> 2], [x |-> 3], [x |-> 4]>>",
		"/\\ x = 5\n/\\ t = <<[x |-> 1], [x |-> 2], [x |-> 3], [x |-> 4], [x |-> 5]>>",
		"/\\ x = 6\n/\\ t = <<[x |-> 1], [x |-> 2], [x |-> 3], [x |-> 4], [x |-> 5], [x |-> 6]>>",
		"/\\ x = 7\n/\\ t = <<[x |-> 1], [x |-> 2], [x |-> 3], [x |-> 4], [x |-> 5], [x |-> 6], [x |-> 7]>>",
	}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != len(expectedTrace) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expectedTrace))
	}
	for i, record := range records {
		if record.StateInfo == nil {
			t.Fatalf("trace state %d has no TLCStateInfo", i+1)
		}
		action := "<Action line 14, col 21 to line 14, col 40 of module TLCExtTrace>"
		if i == 0 {
			action = "<Initial predicate line 14, col 9 to line 14, col 13 of module TLCExtTrace>"
		}
		if info, ok := record.StateInfo.Info.(string); !ok || info != action {
			t.Fatalf("trace action %d=%v, want %s", i+1, record.StateInfo.Info, action)
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
}

// Port of TLCExtTraceSimTest.testSpec, using the same original embedded config.
func TestJavaTLCExtTraceSim(t *testing.T) {
	result := runJavaTLCModelTest(t, "TLCExtTrace", "-config", "TLCExtTrace.tla", "-simulate", "num=1")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCProgressSimu, "10", "1", "10", "0", "0")
	requireJavaTLCUncovered(t, result)
}
