/*******************************************************************************
 * Copyright (c) 2023 Microsoft Research. All rights reserved.
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

// Original ActionCompositionATest.testSpec and doCoverage=false.
func TestJavaActionCompositionA(t *testing.T) {
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "ActionComposition", false, true, true, 1,
		"-config", "ActionCompositionA.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionCompositionATest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "10", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
}

// Original ActionCompositionBTest.testSpec and doCoverage=false.
func TestJavaActionCompositionB(t *testing.T) {
	t.Setenv("tlc2.tool.impl.Tool.cdot", "true")
	r := runJavaTLCModelTestWithRoot(t, "cdot", "ActionComposition", false, true, true, 1,
		"-config", "ActionCompositionB.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ActionCompositionBTest.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "10", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
	if len(javaTLCRecords(r, tlc.ECTLCInvariantViolatedBehavior)) == 0 {
		t.Fatal("TLC_INVARIANT_VIOLATED_BEHAVIOR absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 4", "x = 6"}, true)
}
