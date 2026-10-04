/*******************************************************************************
 * Copyright (c) 2015, 2020 Microsoft Research. All rights reserved.
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

// Original ContinueTest.testSpec with three workers and source -continue.
func TestJavaContinue(t *testing.T) {
	r := runJavaTLCModelTestWithWorkers(t, "Continue", true, true, 3, "-continue", "-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ContinueTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "32", "29", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "29")
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTESpecGenerationComplete)) != 0 {
		t.Fatal("TE spec generated")
	}
	// Source checks concatenated length, not an ordering of concurrent traces.
	if n := len(javaTLCRecords(r, tlc.ECTLCStatePrint2)); n != 32 {
		t.Fatalf("trace states=%d, want32", n)
	}
	requireJavaTLCUncovered(t, r)
}

// Original EmptyTest.testSpec, including its explicit extra -coverage1 option.
func TestJavaEmpty(t *testing.T) {
	r := runJavaTLCModelTest(t, "Empty", "-coverage", "1", "-dumpTrace", "json", filepath.Join(t.TempDir(), "EmptyTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
}
