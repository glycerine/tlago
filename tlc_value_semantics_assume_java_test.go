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
// Port of ValueSemanticsAssumeTest.testSpec and its constructor settings.
package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaValueSemanticsAssume(t *testing.T) {
	result := runJavaTLCModelTestWithDebugger(t, "ValueSemanticsAssume", true, true, false, 1,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ValueSemanticsAssume.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECTLCAssumptionFalse, tlc.ECTLCAssumptionEvaluationError, tlc.ECGeneral} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected code %d records: %v", code, got)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "0", "0", "0")
}
