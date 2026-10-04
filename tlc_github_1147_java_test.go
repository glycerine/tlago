/*******************************************************************************
 * Copyright (c) 2025 Microsoft. All rights reserved.
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
	"bufio"
	"github.com/glycerine/tlago/tlc"
	"os"
	"path/filepath"
	"testing"
)

// Original Github1147Test.testSpec, including complete golden DOT/EOF comparison.
func TestJavaGithub1147(t *testing.T) {
	dumpPath := filepath.Join(t.TempDir(), "Github1147.dot")
	r := runJavaTLCModelTestWithRoot(t, "Github1147", "Github1147", true, false, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Github1147Test.json"),
		"-dump", "dot,actionLabels,colorize", dumpPath)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want ViolationSafety", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "51", "50", "47")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
	if _, err := os.Stat(dumpPath); err != nil {
		t.Fatal(err)
	}
	expected, err := os.Open(filepath.Join("tlc", "test_vectors", "models", "Github1147", "Github1147.dot"))
	if err != nil {
		t.Fatal(err)
	}
	defer expected.Close()
	actual, err := os.Open(dumpPath)
	if err != nil {
		t.Fatal(err)
	}
	defer actual.Close()
	expectedReader, actualReader := bufio.NewScanner(expected), bufio.NewScanner(actual)
	for line := 1; ; line++ {
		hasExpected, hasActual := expectedReader.Scan(), actualReader.Scan()
		if expectedReader.Err() != nil {
			t.Fatal(expectedReader.Err())
		}
		if actualReader.Err() != nil {
			t.Fatal(actualReader.Err())
		}
		if hasExpected != hasActual {
			t.Fatalf("DOT EOF differs at line %d", line)
		}
		if !hasExpected {
			break
		}
		if expectedReader.Text() != actualReader.Text() {
			t.Fatalf("DOT line %d=%q, want %q", line, actualReader.Text(), expectedReader.Text())
		}
	}
	requireJavaTLCUncovered(t, r)
}
