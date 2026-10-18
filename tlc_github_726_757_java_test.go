/*******************************************************************************
 * Copyright (c) 2022 Microsoft Research. All rights reserved.
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
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original Github726Test.TestPrintWriter records each complete write(String).
type javaGithub726PrintWriter struct {
	output io.Writer
	log    []string
}

func (w *javaGithub726PrintWriter) Write(p []byte) (int, error) {
	w.log = append(w.log, string(p))
	return w.output.Write(p)
}

// Original Github726Test.testSpec and constructor output override.
func TestJavaGithub726(t *testing.T) {
	writer := &javaGithub726PrintWriter{output: os.Stdout}
	previous := tlc.TLCOutput
	tlc.TLCOutput = writer
	defer func() { tlc.TLCOutput = previous }()
	r := runJavaTLCModelTestWithRoot(t, "Github726", "Github726", true, true, false, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github726Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	if !slices.Contains(writer.log, "[data |-> 15, dst |-> \"a\"]\n") {
		t.Fatalf("correct record absent from writes: %q", writer.log)
	}
	if slices.Contains(writer.log, "[data |-> \"a\", dst |-> 15]\n") {
		t.Fatalf("incorrect record present in writes: %q", writer.log)
	}
}

// Original Github742Test.testSpec with all source runner defaults.
func TestJavaGithub742(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github742", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github742Test.json"))
	if r.ExitStatus != tlc.ExitStatusError {
		t.Fatalf("exit=%d, want ERROR", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveCannotHandleFormula)) == 0 {
		t.Fatal("TLC_LIVE_CANNOT_HANDLE_FORMULA absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
}

// Original Github743Test.testSpec, including the complete original trace.
func TestJavaGithub743(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github743", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github743Test.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ x = \"foo\"\n/\\ y = 0",
		"/\\ x = \"foo\"\n/\\ y = 1",
		"/\\ x = TRUE\n/\\ y = 3",
		"/\\ x = TRUE\n/\\ y = 5",
	}, true)
}

// Original Github746Test.testSpec with its exact initial-state diagnostic.
func TestJavaGithub746(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github746", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github746Test.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCComputingInit)) == 0 {
		t.Fatal("TLC_COMPUTING_INIT absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInvariantViolatedInitial, "Inv", "/\\ i = 1\n/\\ foo = \"bar\"\n/\\ j = TRUE\n")
}

// Original Github757Test.testSpec, noGenerateSpec=true, doDumpTrace=false,
// and its explicit JSON dump argument (path relocated to isolated test storage).
func TestJavaGithub757(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github757", "-noGenerateSpecTE", "-config", "Github757.tla",
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "states", "Github757.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
}
