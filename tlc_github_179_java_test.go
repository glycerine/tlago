/*******************************************************************************
 * Copyright (c) 2019 Microsoft Research. All rights reserved.
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

// Original Github179aTest.testSpec with all ModelCheckerTestCase defaults.
func TestJavaGithub179A(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github179a", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github179aTest.json"))
	if r.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit=%d, want ViolationAssumption; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCModuleValueJavaMethodOverride, "public static tlc2.value.impl.Value tlc2.module.TLC.PrintT(tlc2.value.impl.Value)", "Attempted to check equality of integer 1 with non-integer:\n{1}")
}

// Original Github179bTest.testSpec with all ModelCheckerTestCase defaults.
func TestJavaGithub179B(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github179b", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github179bTest.json"))
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FailureSpecEval; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCModuleValueJavaMethodOverride, "public static tlc2.value.impl.Value tlc2.module.TLC.Print(tlc2.value.impl.Value,tlc2.value.impl.Value)", "Attempted to check equality of integer 1 with non-integer:\n{1}")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCNestedExpression, "0. Line 13, column 1 to line 13, column 30 in Github179b\n1. Line 13, column 1 to line 13, column 23 in Github179b\n2. Line 9, column 27 to line 9, column 79 in Github179b\n3. Line 9, column 40 to line 9, column 79 in Github179b\n4. Line 9, column 43 to line 9, column 52 in Github179b\n\n")
}

// Original Github179cTest.testSpec with all ModelCheckerTestCase defaults.
func TestJavaGithub179C(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github179c", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github179cTest.json"))
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FailureSpecEval; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCModuleValueJavaMethodOverride, "public static tlc2.value.impl.Value tlc2.module.TLC.PrintT(tlc2.value.impl.Value)", "Attempted to check equality of integer 1 with non-integer:\n{1}")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCNestedExpression, "0. Line 16, column 9 to line 16, column 42 in Github179c\n1. Line 16, column 9 to line 16, column 33 in Github179c\n2. Line 16, column 9 to line 16, column 25 in Github179c\n3. Line 10, column 11 to line 12, column 39 in Github179c\n4. Line 10, column 24 to line 12, column 39 in Github179c\n5. Line 10, column 27 to line 10, column 36 in Github179c\n\n")
}
