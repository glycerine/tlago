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
// Port of DepthFirstErrorTraceTest.testSpec and its -dfid 9 constructor.
package tlago

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaDepthFirstErrorTrace(t *testing.T) {
	result := runJavaTLCModelTest(t, "DepthFirstErrorTrace", "-dfid", "9",
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "DepthFirstErrorTrace.json"))
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want VIOLATION_SAFETY", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCStatePrint1} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected code %d records: %v", code, got)
		}
	}
	expectedTrace := []string{"x = 0", "x = 1", "x = 2", "x = 3", "x = 4", "x = 5", "x = 6", "x = 7"}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != len(expectedTrace) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expectedTrace))
	}
	for i, record := range records {
		if record.StateInfo == nil || strings.TrimSpace(record.StateInfo.String()) != expectedTrace[i] {
			t.Fatalf("trace state %d=%v, want %q", i+1, record.StateInfo, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
		if got := record.StateInfo.Info; got != "" {
			t.Fatalf("trace action %d=%q, want empty label", i+1, got)
		}
	}
	requireJavaTLCUncovered(t, result)
}
