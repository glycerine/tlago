/*******************************************************************************
 * Copyright (c) 2025 NVIDIA Corp. All rights reserved.
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
	"path/filepath"
	"testing"
)

// Original Github1198aTest.testSpec and doCoverage override.
func TestJavaGithub1198A(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198a.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198aTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198bTest.testSpec and doCoverage override.
func TestJavaGithub1198B(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198b.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198bTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198cTest.testSpec and doCoverage override.
func TestJavaGithub1198C(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198c.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198cTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198dTest.testSpec and doCoverage override.
func TestJavaGithub1198D(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198d.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198dTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) != 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198fTest.testSpec and doCoverage override.
func TestJavaGithub1198F(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198f.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198fTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) != 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original Github1198hTest.testSpec and doCoverage override.
func TestJavaGithub1198H(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github1198", "Github1198", false, true, true, 1, "-config", "Github1198h.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1198hTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaAndFairnessTautology)) != 0 {
		t.Fatal("TLC_LIVE_FORMULA_AND_FAIRNESS_TAUTOLOGY present")
	}
	requireJavaTLCUncovered(t, r)
}
