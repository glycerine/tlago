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

// Original Github715Test.testSpec, including doCoverage=false.
func TestJavaGithub715(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github715", "Github715", false, true, true, 1,
		"-config", "Github715.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github715Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigOpIsEqual)) != 0 {
		t.Fatal("TLC_CONFIG_OP_IS_EQUAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaStateLevel)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_STATE_LEVEL absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original Github715bTest.testSpec, including doCoverage=false.
func TestJavaGithub715b(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github715", "Github715", false, true, true, 1,
		"-config", "Github715b.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github715bTest.json"))
	if r.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit=%d, want ExitStatusErrorConfigParse", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCConfigOpIsEqual, "x", "line 2, col 10 to line 2, col 10 of module Github715", "spec")
}

// Original Github715cTest.testSpec, including doCoverage=false.
func TestJavaGithub715c(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github715", "Github715", false, true, true, 1,
		"-config", "Github715c.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github715cTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaStateLevel)) == 0 {
		t.Fatal("TLC_LIVE_FORMULA_STATE_LEVEL absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}

// Original Github715dTest.testSpec, including doCoverage=false.
func TestJavaGithub715d(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github715", "Github715", false, true, true, 1,
		"-config", "Github715d.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github715dTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCLiveFormulaStateLevel)) != 0 {
		t.Fatal("TLC_LIVE_FORMULA_STATE_LEVEL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
}
