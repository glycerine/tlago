/*******************************************************************************
 * Copyright (c) 2025 Microsoft. All rights reserved.
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

// Original Github1134a..fTest.testSpec methods. Each class has noGenerateSpec
// and doDumpTrace overrides; coverage, DOT, debugger and one worker stay enabled.
func TestJavaGithub1134A(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCa.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ViolationLiveness", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCPropertyViolatedInitial, "Init", "State = <<\"INIT\", \"INIT\">>\n")
	requireJavaTLCUncovered(t, r)
}

func TestJavaGithub1134B(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCb.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "4", "0")
	requireJavaTLCUncovered(t, r)
}

func TestJavaGithub1134C(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCc.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "4", "0")
	requireJavaTLCUncovered(t, r)
}

func TestJavaGithub1134D(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCd.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FailureSpecEval", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitialState,
		"Attempted to apply function:\n<<\"INIT\", \"INIT\">>\nto argument 3, which is not in the domain of the function.",
		"State = <<\"INIT\", \"INIT\">>\n")
}

func TestJavaGithub1134E(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCe.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "4", "0")
	requireJavaTLCUncovered(t, r)
}

func TestJavaGithub1134F(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1134", "MC.tla", true, true, true, 1,
		"-config", "MCf.cfg", "-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FailureSpecEval", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitialState,
		"Attempted to apply function:\n<<\"INIT\", \"INIT\">>\nto argument 3, which is not in the domain of the function.",
		"State = <<\"INIT\", \"INIT\">>\n")
}
