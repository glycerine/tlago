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

// Original ETest7.testSpec and ERROR_CONFIG_PARSE constructor, with all inherited
// SuiteETestCase runner settings and exact nonconstant-substitution parameters.
func TestJavaLegacySuiteETest7(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteETest7", "etest7", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "ETest7.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ERROR_CONFIG_PARSE; messages=%v", r.ExitStatus, r.Messages)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCConfigSubstitutionNonConstant, "C", "Foo")
}

// Original ETest16.testSpec and ERROR_SPEC_PARSE constructor, preserving both
// duplicate-field ranges and messages with the full SuiteETestCase settings.
func TestJavaLegacySuiteETest16(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	var output bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&output, &output)
	defer restore()
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteETest16", "etest16", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "ETest16.json"))
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE; messages=%v", r.ExitStatus, r.Messages)
	}
	for _, substring := range []string{
		"*** Errors: 2\n",
		"line 5, col 27 to line 5, col 27 of module etest16\n",
		"Non-unique fields in constructor.\n",
		"line 7, col 27 to line 7, col 27 of module etest16\n",
		"Non-unique fields in constructor.",
	} {
		if !strings.Contains(output.String(), substring) {
			t.Fatalf("TestPrintStream output lacks %q:\n%s", substring, output.String())
		}
	}
}

// Original Test210.testSpec and ERROR_SPEC_PARSE constructor, preserving all
// ten exact TestPrintStream substrings and full SuiteETestCase runner settings.
func TestJavaLegacySuiteTest210(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	var output bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&output, &output)
	defer restore()
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest210", "test210", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Test210.json"))
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	for _, substring := range []string{
		"Semantic errors:\n\n*** Errors: 9\n",
		"line 33, col 16 to line 33, col 19 of module test210\n\nAccessing subexpression labeled `laby' of ASSUME/PROVE clause within the scope of a declaration\n from outside that declaration's scope.\n",
		"line 35, col 16 to line 35, col 16 of module test210\n\nAccessing ASSUME/PROVE clause within the scope of a declaration\n from outside that declaration's scope.\n",
		"line 42, col 41 to line 42, col 55 of module test210\n\nLabel not allowed within scope of declaration in nested ASSUME/PROVE.\n",
		"line 43, col 30 to line 43, col 44 of module test210\n\nLabel not allowed within scope of declaration in nested ASSUME/PROVE.\n",
		"line 55, col 16 to line 55, col 16 of module test210\n\nAccessing non-existent subexpression of a SUFFICES\n",
		"line 62, col 15 to line 62, col 18 of module test210\n\nAccessing subexpression labeled `laby' of ASSUME/PROVE clause within the scope of a declaration\n from outside that declaration's scope.\n",
		"line 63, col 15 to line 63, col 18 of module test210\n\nAccessing subexpression labeled `labi' of ASSUME/PROVE clause within the scope of a declaration\n from outside that declaration's scope.\n",
		"line 64, col 15 to line 64, col 18 of module test210\n\nAccessing subexpression labeled `labu' of ASSUME/PROVE clause within the scope of a declaration\n from outside that declaration's scope.\n",
		"line 65, col 15 to line 65, col 15 of module test210\n\nAccessing ASSUME/PROVE clause within the scope of a declaration\n from outside that declaration's scope.\n",
	} {
		if !strings.Contains(output.String(), substring) {
			t.Fatalf("TestPrintStream output lacks %q:\n%s", substring, output.String())
		}
	}
}

// Original Test212.testSpec and ERROR_SPEC_PARSE constructor, preserving all
// seven exact TestPrintStream substrings and full SuiteETestCase runner settings.
func TestJavaLegacySuiteTest212(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	var output bytes.Buffer
	restore := tlc.ToolIOSetSystemStreams(&output, &output)
	defer restore()
	r := runJavaTLCModelTestWithRoot(t, "LegacySuiteTest212", "test212", true, true, true, 1,
		"-dumpTrace", "json", filepath.Join(t.TempDir(), "Test212.json"))
	if r.ExitStatus != tlc.ExitStatusErrorSpecParse {
		t.Fatalf("exit=%d, want ERROR_SPEC_PARSE; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	for _, substring := range []string{
		"Semantic errors:\n\n*** Errors: 6\n",
		"line 27, col 1 to line 27, col 62 of module test212\n\nError in instantiating module 'test212a':\n A non-Leibniz operator substituted for 'Op2'.\n",
		"line 28, col 1 to line 28, col 62 of module test212\n\nError in instantiating module 'test212a':\n A non-Leibniz operator substituted for 'Op'.\n",
		"line 29, col 1 to line 29, col 62 of module test212\n\nError in instantiating module 'test212a':\n A non-Leibniz operator substituted for 'Op'.\n",
		"line 30, col 1 to line 30, col 64 of module test212\n\nError in instantiating module 'test212a':\n A non-Leibniz operator substituted for 'Op2'.\n",
		"line 31, col 1 to line 31, col 63 of module test212\n\nError in instantiating module 'test212a':\n A non-Leibniz operator substituted for 'Op2'.\n",
		"line 32, col 1 to line 32, col 63 of module test212\n\nError in instantiating module 'test212a':\n A non-Leibniz operator substituted for 'Op2'.\n",
	} {
		if !strings.Contains(output.String(), substring) {
			t.Fatalf("TestPrintStream output lacks %q:\n%s", substring, output.String())
		}
	}
}
