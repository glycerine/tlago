/*******************************************************************************
 * Copyright (c) 2016 Microsoft Research. All rights reserved.
 * Copyright (c) 2026 NVIDIA Corporation. All rights reserved.
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
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"strings"
	"testing"
)

// Original ETest1.testSpec and ERROR_SPEC_PARSE constructor, with the
// SuiteETestCase TestPrintStream capture and full inherited runner settings.
func TestJavaLegacySuiteETest1(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	var output bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&output, &output)
	defer restore()
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteETest1", "etest1", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "ETest1.json"))
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, substring := range []string{
		"*** Errors: 1\n",
		"line 18, col 12 to line 18, col 17 of module etest1\n",
		"The operator Foo requires 1 arguments.",
	} {
		if !strings.Contains(output.String(), substring) {
			t.Fatalf("TestPrintStream output lacks %q:\n%s", substring, output.String())
		}
	}
}

// Original ETest2.testSpec and ERROR_SPEC_PARSE constructor, with the
// SuiteETestCase TestPrintStream capture and full inherited runner settings.
func TestJavaLegacySuiteETest2(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	var output bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&output, &output)
	defer restore()
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteETest2", "etest2", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "ETest2.json"))
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	for _, substring := range []string{
		"*** Errors: 1\n",
		"line 18, col 12 to line 18, col 14 of module etest2\n",
		"The operator Foo requires 2 arguments.",
	} {
		if !strings.Contains(output.String(), substring) {
			t.Fatalf("TestPrintStream output lacks %q:\n%s", substring, output.String())
		}
	}
}

// Original ETest5.testSpec and ERROR_SPEC_PARSE constructor, with the
// SuiteETestCase TestPrintStream capture and full inherited runner settings.
func TestJavaLegacySuiteETest5(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	var output bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&output, &output)
	defer restore()
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteETest5", "etest5", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "ETest5.json"))
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	for _, substring := range []string{
		"*** Errors: 1\n",
		"line 13, col 15 to line 13, col 20 of module etest5\n",
		"Unknown operator: `M!Init'.",
	} {
		if !strings.Contains(output.String(), substring) {
			t.Fatalf("TestPrintStream output lacks %q:\n%s", substring, output.String())
		}
	}
}
