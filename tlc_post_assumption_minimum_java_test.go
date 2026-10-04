/*******************************************************************************
 * Copyright (c) 2018, 2020 Microsoft Research. All rights reserved.
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

// Original PostAssumptionTest.testSpec and inherited assumption-violation exit.
func TestJavaPostAssumption(t *testing.T) {
	r := runJavaTLCModelTest(t, "DieHardTLA", "-config", "DieHardTLAPA", "-dumpTrace", "json", filepath.Join(t.TempDir(), "PostAssumptionTest.json"))
	if r.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit=%d, want VIOLATION_ASSUMPTION", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "97", "16", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCPostconditionFalse, "PostCondition")
	requireJavaTLCUncovered(t, r)
}

// Original MinimumDiameterTest.testSpec with source doDump=false.
func TestJavaMinimumDiameter(t *testing.T) {
	r := runJavaTLCModelTestWithSettings(t, "MinimumDiameter", true, false, "-dumpTrace", "json", filepath.Join(t.TempDir(), "MinimumDiameterTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCUncovered(t, r)
}
