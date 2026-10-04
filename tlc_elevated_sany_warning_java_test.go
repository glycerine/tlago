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
 ******************************************************************************/
package tlago

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original ElevatedSanyWarning.testSpec and beforeSetUp, with its absolute
// corpus path and inherited ModelCheckerTestCase settings and exit assertion.
func TestJavaElevatedSanyWarning(t *testing.T) {
	specPath, err := filepath.Abs(filepath.Join("tlc", "test_vectors", "models", "ElevatedSanyWarning", "W4802_Pre_Test.tla"))
	if err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&stdout, os.Stderr)
	defer restore()
	r := runJavaTLCModelTestWithRoot(t, "ElevatedSanyWarning", specPath, true, true, true, 1,
		"-messagesAsErrors", "4802", "-dumpTrace", "json", filepath.Join(t.TempDir(), "ElevatedSanyWarning.json"))
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCParsingFailed)) == 0 {
		t.Fatal("TLC_PARSING_FAILED absent")
	}
	if !strings.Contains(stdout.String(), "Warning treated as error") {
		t.Fatalf("promoted warning absent: %s", stdout.String())
	}
}
