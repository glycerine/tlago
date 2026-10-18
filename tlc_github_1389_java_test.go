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

// Original Github1389Test.testSpec, retaining coverage and all other runner overrides.
func TestJavaGithub1389(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1389", "Github1389", true, false, false, 1, "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveCannotHandleFormula)) != 0 {
		t.Fatal("TLC_LIVE_CANNOT_HANDLE_FORMULA present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original Github1389CountingTest.testSpec, retaining coverage and all other runner overrides.
func TestJavaGithub1389Counting(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1389Counting", "Github1389Counting", true, false, false, 1, "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want ViolationSafety; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveCannotHandleFormula)) != 0 {
		t.Fatal("TLC_LIVE_CANNOT_HANDLE_FORMULA present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "CountAtMostFour")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original Github1389ViolatedTest.testSpec, retaining coverage and all other runner overrides.
func TestJavaGithub1389Violated(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1389Violated", "Github1389Violated", true, false, false, 1, "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ViolationLiveness; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveCannotHandleFormula)) != 0 {
		t.Fatal("TLC_LIVE_CANNOT_HANDLE_FORMULA present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "PropViolated")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1"}, false)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
}

// Original Github1389ViolatedBTest.testSpec, retaining coverage and all other runner overrides.
func TestJavaGithub1389ViolatedB(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1389ViolatedB", "Github1389ViolatedB", true, false, false, 1, "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ViolationLiveness; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveCannotHandleFormula)) != 0 {
		t.Fatal("TLC_LIVE_CANNOT_HANDLE_FORMULA present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "PropViolated")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1"}, false)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
}

// Original Github1389ViolatedCTest.testSpec, retaining coverage and all other runner overrides.
func TestJavaGithub1389ViolatedC(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1389ViolatedC", "Github1389ViolatedC", true, false, false, 1, "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ViolationLiveness; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveCannotHandleFormula)) != 0 {
		t.Fatal("TLC_LIVE_CANNOT_HANDLE_FORMULA present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "PropViolated")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}

// Original Github1389LoopsTest.testSpec and exact Error exit/runner overrides.
func TestJavaGithub1389Loops(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1389Loops", "Github1389Loops", true, false, false, 1, "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusError {
		t.Fatalf("exit=%d, want Error; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECSystemStackOverflow)) == 0 {
		t.Fatal("SYSTEM_STACK_OVERFLOW absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
}

// Original Github1389StateGuardTest.testSpec and exact Error exit/runner overrides.
func TestJavaGithub1389StateGuard(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1389StateGuard", "Github1389StateGuard", true, false, false, 1, "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusError {
		t.Fatalf("exit=%d, want Error; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECSystemStackOverflow)) == 0 {
		t.Fatal("SYSTEM_STACK_OVERFLOW absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
}
