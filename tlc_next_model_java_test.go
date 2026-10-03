/*******************************************************************************
 * Copyright (c) 2015, 2018 Microsoft Research. All rights reserved.
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

// Ports of IncompleteNextTest and IncompleteNextMultipleActionsTest.testSpec.
package tlago

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaIncompleteNext(t *testing.T) {
	result := runJavaTLCModelTest(t, "IncompleteNext")
	requireJavaIncompleteNextResult(t, result)
	expectedTrace := []string{"/\\ x = 0\n/\\ y = 0", "/\\ x = 1\n/\\ y = null"}
	expectedActions := []string{
		"<Initial predicate line 6, col 19 to line 6, col 21 of module IncompleteNext>",
		"<Action line 6, col 30 to line 6, col 35 of module IncompleteNext>",
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
		if !ok || info != expectedActions[i] {
			t.Fatalf("trace action %d=%v, want %s", i+1, record.StateInfo.Info, expectedActions[i])
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStateNotCompletelySpecifiedNext, " is", "y")
	requireJavaTLCUncovered(t, result)
}

func TestJavaIncompleteNextMultipleActions(t *testing.T) {
	result := runJavaTLCModelTest(t, "IncompleteNextMultipleActions")
	requireJavaIncompleteNextResult(t, result)
	expectedTrace := []string{
		"/\\ x = 0\n/\\ y = 0\n/\\ z = 0",
		"/\\ x = 1\n/\\ y = null\n/\\ z = null",
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
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStateNotCompletelySpecifiedNext, "A1", "s are", "y, z")
	requireJavaTLCUncovered(t, result,
		"line 8, col 16 to line 8, col 21 of module IncompleteNextMultipleActions",
		"line 8, col 26 to line 8, col 31 of module IncompleteNextMultipleActions",
		"line 8, col 7 to line 8, col 11 of module IncompleteNextMultipleActions")
}

func requireJavaIncompleteNextResult(t *testing.T, result *tlc.Result) {
	t.Helper()
	if result.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit status=%d, want spec evaluation failure", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if got := javaTLCRecords(result, tlc.ECGeneral); len(got) != 0 {
		t.Fatalf("unexpected GENERAL: %v", got)
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "2", "1", "0")
}
