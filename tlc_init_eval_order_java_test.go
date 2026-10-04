/*******************************************************************************
 * Copyright (c) 2018 Microsoft Research. All rights reserved.
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

// Original InitEvalOrder1Test.test, inherited from InitEvalOrderTest.
func TestJavaInitEvalOrder1(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrder", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrder1Test.json"), "-config", "InitEvalOrder1.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}

// Original InitEvalOrder2Test.test, inherited from InitEvalOrderTest.
func TestJavaInitEvalOrder2(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrder", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrder2Test.json"), "-config", "InitEvalOrder2.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}

// Original InitEvalOrder3Test.test, inherited from InitEvalOrderTest.
func TestJavaInitEvalOrder3(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrder", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrder3Test.json"), "-config", "InitEvalOrder3.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}

// Original InitEvalOrder4Test.test, inherited from InitEvalOrderTest.
func TestJavaInitEvalOrder4(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrder", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrder4Test.json"), "-config", "InitEvalOrder4.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
}

// Original InitEvalOrderBasicTest.test.
func TestJavaInitEvalOrderBasic(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "EvalOrder", "InitEvalOrderBasic", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "InitEvalOrderBasicTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCUncovered(t, r)
}
