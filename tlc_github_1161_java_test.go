/*******************************************************************************
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
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

// Original Github1161Test.testSpec, preserving all original parameterized properties.
func TestJavaGithub1161(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1161", "Github1161", false, false, false, 1, "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveCannotHandleFormula)) != 0 {
		t.Fatal("TLC_LIVE_CANNOT_HANDLE_FORMULA present")
	}
}

// Original Github1161ViolatedTest.testSpec, including its safety exit and full trace.
func TestJavaGithub1161Violated(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1161", "Github1161", false, false, false, 1, "-config", "Github1161Violated.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want ViolationSafety", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveCannotHandleFormula)) != 0 {
		t.Fatal("TLC_LIVE_CANNOT_HANDLE_FORMULA present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "PropertyViolated")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"s = 0", "s = 1", "s = 2", "s = 3", "s = 4", "s = 5"}, false)
}
