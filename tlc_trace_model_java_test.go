/*******************************************************************************
 * Copyright (c) 2017 Microsoft Research. All rights reserved.
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
// Port of TraceWithLargeSetOfInitialStatesTest.testSpec.
package tlago

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaTraceWithLargeSetOfInitialStates(t *testing.T) {
	runJavaTraceWithLargeSetOfInitialStates(t)
}

func runJavaTraceWithLargeSetOfInitialStates(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	args := []string{"-maxSetSize", "10", "-dumpTrace", "json", filepath.Join(t.TempDir(), "TraceWithLargeSetOfInitialStatesTest.json")}
	result := runJavaTLCModelTest(t, "TraceWithLargeSetOfInitialStatesTest", append(args, extraArgs...)...)
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want safety violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCBug} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	if len(javaTLCRecords(result, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT not recorded")
	}
	expectedTrace := []string{"/\\ x = 1\n/\\ y = FALSE", "/\\ x = 1\n/\\ y = TRUE"}
	expectedActions := []string{
		"<Initial predicate line 6, col 25 to line 6, col 33 of module TraceWithLargeSetOfInitialStatesTest>",
		"<Action line 6, col 42 to line 6, col 60 of module TraceWithLargeSetOfInitialStatesTest>",
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
	requireJavaTLCUncovered(t, result)
	return result
}
