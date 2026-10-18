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
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original Github362Test.testSpec, including its beforeSetUp output capture.
func TestJavaGithub362(t *testing.T) {
	var output bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&output, os.Stderr)
	defer restore()
	r := runJavaTLCModelTest(t, "Github362", "-config", "Github362.tla", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github362Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, substring := range []string{
		`<<"Evaluated initial state in A; overloadedName is: ", "fizzbuzz">>`,
		`<<"From A's perspective, B's overloadedName is: ", "x">>`,
		`<<"Evaluating initial state in B; overloadedName is ", "x">>`,
		`<<"Evaluated initial state in A; overloadedConst is: ", 4711>>`,
		`<<"From A's perspective, B's overloadedConst is: ", 42>>`,
		`<<"Evaluating initial state in B; overloadedConst is ", 42>>`,
	} {
		found := false
		for _, line := range strings.Split(output.String(), "\n") {
			if strings.Contains(line, substring) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("substring %q absent from output:\n%s", substring, output.String())
		}
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCSuccess} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("diagnostic %d absent", code)
		}
	}
}

// Original Github391Test.testSpec with all ModelCheckerTestCase defaults.
func TestJavaGithub391(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github391", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github391Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "0", "0", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "0")
}

// Original Github407Test.testSpec and doDump override, including every golden line and EOF.
func TestJavaGithub407(t *testing.T) {
	dumpPath := filepath.Join(t.TempDir(), "Github407.dump")
	r := runJavaTLCModelTestWithRoot(t, "Github407", "Github407", true, false, true, 1,
		"-dump", dumpPath, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github407Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
	if _, err := os.Stat(dumpPath); err != nil {
		t.Fatal(err)
	}
	expected, err := os.Open(filepath.Join("tlc", "test_vectors", "models", "Github407", "Github407.dump"))
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
			t.Fatalf("dump EOF differs at line %d", line)
		}
		if !hasExpected {
			break
		}
		if expectedReader.Text() != actualReader.Text() {
			t.Fatalf("dump line %d=%q, want %q", line, actualReader.Text(), expectedReader.Text())
		}
	}
	requireJavaTLCUncovered(t, r)
}
