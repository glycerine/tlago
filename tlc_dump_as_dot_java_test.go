/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
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
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original DumpAsDotTest.testSpec, including its exact DOT master-file bytes.
func TestJavaDumpAsDot(t *testing.T) {
	prefix := filepath.Join(t.TempDir(), "DumpAsDotTest")
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "MCa", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "DumpAsDotTest.json"),
		"-dump", "dot,colorize,actionlabels,stuttering", prefix)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "18", "11", "0")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 18 {
		t.Fatalf("register42=%v, want IntValue18", values[0])
	}
	dumpFile := prefix + ".dot"
	if _, err := os.Stat(dumpFile); err != nil {
		t.Fatalf("dump file absent: %v", err)
	}
	master, err := os.ReadFile(filepath.Join("tlc", "test_vectors", "models", "CodePlexBug08", "DumpAsDotTest.dot"))
	if err != nil {
		t.Fatal(err)
	}
	actual, err := os.ReadFile(dumpFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(master, actual) {
		t.Fatalf("DOT dump bytes differ from original master:\nactual:\n%s\nmaster:\n%s", actual, master)
	}
	requireJavaTLCUncovered(t, r)
}
