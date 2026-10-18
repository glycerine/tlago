/*******************************************************************************
 * Copyright (c) 2017, 2021 Microsoft Research. All rights reserved.
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
 *   Ian Morris Nieves - initial design and implementation
 *   Markus Alexander Kuppe - initial design and implementation
 ******************************************************************************/

package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original FingerprintExceptionInitTest.testSpec.
func TestJavaFingerprintExceptionInit(t *testing.T) {
	r := runJavaTLCModelTest(t, "FingerprintExceptionInit", "-dumpTrace", "json", filepath.Join(t.TempDir(), "FingerprintExceptionInitTest.json"))
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FAILURE_SPEC_EVAL", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCFingerprintException,
		"1) line 7, col 20 to line 7, col 32 of module FingerprintExceptionInit\n0) line 7, col 13 to line 7, col 33 of module FingerprintExceptionInit\n",
		"Overflow when computing the number of elements in:\nSUBSET (1..36)")
	requireJavaTLCUncovered(t, r, "line 8, col 39 to line 8, col 64 of module FingerprintExceptionInit", "line 8, col 71 to line 8, col 76 of module FingerprintExceptionInit")
}

// Original FingerprintExceptionNextTest.testSpec.
func TestJavaFingerprintExceptionNext(t *testing.T) {
	r := runJavaTLCModelTest(t, "FingerprintExceptionNext", "-dumpTrace", "json", filepath.Join(t.TempDir(), "FingerprintExceptionNextTest.json"))
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FAILURE_SPEC_EVAL", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) == 0 {
		t.Fatal("GENERAL absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCFingerprintException,
		"1) line 8, col 51 to line 8, col 63 of module FingerprintExceptionNext\n0) line 8, col 44 to line 8, col 64 of module FingerprintExceptionNext\n",
		"Overflow when computing the number of elements in:\nSUBSET (1..32)")
	requireJavaTLCUncovered(t, r, "line 8, col 71 to line 8, col 76 of module FingerprintExceptionNext")
}

// Original FingerprintExceptionNextCallstackTest.testSpec.
func TestJavaFingerprintExceptionNextCallstack(t *testing.T) {
	r := runJavaTLCModelTest(t, "FingerprintExceptionNextCallstack", "-config", "FingerprintExceptionNext.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "FingerprintExceptionNextCallstackTest.json"))
	if r.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit=%d, want FAILURE_SPEC_EVAL", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	if len(javaTLCRecords(r, tlc.ECGeneral)) == 0 {
		t.Fatal("GENERAL absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCFingerprintException,
		"0. Line 9, column 9 to line 9, column 19 in FingerprintExceptionNextCallstack\n1. Line 9, column 14 to line 9, column 19 in FingerprintExceptionNextCallstack\n2. Line 7, column 10 to line 7, column 49 in FingerprintExceptionNextCallstack\n3. Line 7, column 43 to line 7, column 49 in FingerprintExceptionNextCallstack\n4. Line 7, column 25 to line 7, column 32 in FingerprintExceptionNextCallstack\n\n",
		"Attempted to check if the non-enumerable value\n42\nis element of\nSUBSET 42")
}
