/*******************************************************************************
 * Copyright (c) 2021 Microsoft Research. All rights reserved.
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
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original Github696Test.testSpec with all ModelCheckerTestCase defaults.
func TestJavaGithub696(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github696", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github696Test.json"))
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FAILURE_SPEC_EVAL", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitialState, "Attempted to compare the differently-typed model values A_31 and B_a")
}

// Original Github696bTest.testSpec with all ModelCheckerTestCase defaults.
func TestJavaGithub696b(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github696b", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github696bTest.json"))
	if r.ExitStatus != tlc.ExitStatusError {
		t.Fatalf("exit=%d, want ERROR", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECGeneral,
		"TLC threw an unexpected exception.\n"+
			"This was probably caused by an error in the spec or model.\n"+
			"See the User Output or TLC Console for clues to what happened.\n"+
			"The exception was a java.lang.RuntimeException\n"+
			": Attempted to compare the differently-typed model values B_a and A_a")
}
