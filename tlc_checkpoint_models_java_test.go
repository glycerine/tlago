/*******************************************************************************
 * Copyright (c) 2019 Microsoft Research. All rights reserved.
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
	"testing"
)

// Original CheckpointOnViolationTest.testSpec, including inherited exit status
// and doCheckpoint's Integer.MAX_VALUE / 60000 interval.
func TestJavaCheckpointOnViolation(t *testing.T) {
	r := runJavaTLCModelTest(t, "DieHard", "-checkpoint", "35791")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCStatePrint2, tlc.ECTLCCheckpointStart, tlc.ECTLCCheckpointEnd} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("event %d absent", code)
		}
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "252", "54", "11")
	if n := len(javaTLCRecords(r, tlc.ECTLCStatePrint2)); n != 7 {
		t.Fatalf("trace states=%d, want 7", n)
	}
	requireJavaTLCUncovered(t, r)
}

// Original CheckpointWhenTimeBoundTest.testSpec and its five-second stopAfter
// property. Preserve the full infinite-state model and base success exit.
func TestJavaCheckpointWhenTimeBound(t *testing.T) {
	t.Setenv("tlc2.TLC.stopAfter", "5")
	r := runJavaTLCModelTest(t, "InfiniteStateSpace", "-checkpoint", "35791")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCCheckpointStart, tlc.ECTLCCheckpointEnd} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("event %d absent", code)
		}
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCTESpecGenerationComplete, tlc.ECTLCStatePrint1, tlc.ECTLCStatePrint2} {
		if len(javaTLCRecords(r, code)) != 0 {
			t.Fatalf("unexpected event %d", code)
		}
	}
	requireJavaTLCUncovered(t, r)
}
