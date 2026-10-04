/*******************************************************************************
 * Copyright (c) 2025 NVIDIA. All rights reserved.
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

// Original InvParameterizedATest.testSpec and inherited safety-violation exit.
func TestJavaInvParameterizedA(t *testing.T) {
	r := runJavaTLCModelTestWithSettings(t, "InvParameterizedTest", false, false, "-config", "InvParameterizedTest.tla", "-invlevel", "4", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "26", "7", "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "4")
	requireJavaTLCUncovered(t, r)
}

// Original InvParameterizedBTest.testSpec with the unmodified invariant text.
func TestJavaInvParameterizedB(t *testing.T) {
	r := runJavaTLCModelTestWithSettings(t, "InvParameterizedTest", false, false, "-config", "InvParameterizedTest.tla", "-inv", "~(small = 3 /\\ big = 0)", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCUncovered(t, r)
}

// Original InvParameterizedCTest.testSpec and inherited ERROR_SPEC_PARSE exit.
func TestJavaInvParameterizedC(t *testing.T) {
	r := runJavaTLCModelTestWithSettings(t, "InvParameterizedTest", false, false, "-config", "InvParameterizedTest.tla", "-inv", "garbled input to cause parse error", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
}
