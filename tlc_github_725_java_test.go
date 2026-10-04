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

// Original Github725Test.testSpec with its original runner overrides.
func TestJavaGithub725(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github725", "Github725", false, true, false, 1,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github725Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original Github725bTest.testSpec with its original runner overrides.
func TestJavaGithub725b(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github725b", "Github725b", false, true, false, 1,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github725bTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original Github725cTest.testSpec with its original runner overrides.
func TestJavaGithub725c(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github725c", "Github725c", false, true, false, 1,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github725cTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original Github725dTest.testSpec with its original runner overrides.
func TestJavaGithub725d(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github725d", "Github725d", false, true, false, 1,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github725dTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original Github725eTest.testSpec with its original runner overrides.
func TestJavaGithub725e(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github725e", "Github725e", false, true, false, 1,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github725eTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original Github725fTest.testSpec with its original runner overrides.
func TestJavaGithub725f(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github725f", "Github725f", false, true, false, 1,
		"-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github725fTest.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want ExitStatusSuccess; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
}

// Original Github725gTest.testSpec with its original runner overrides.
func TestJavaGithub725g(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github725g", "Github725g", false, true, false, 1,
		"-noGenerateSpecTE")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ExitStatusViolationLiveness; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "1")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "Prop")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ b = 0\n/\\ q = 0\n/\\ s = <<1, 2, 3>>"}, false)
	stutter := javaTLCRecords(r, tlc.ECTLCStatePrint3)
	if len(stutter) == 0 || stutter[0].StateNumber != 2 {
		t.Fatalf("stuttering=%v, want ordinal2", stutter)
	}
}

// Original Github725hTest.testSpec with its original runner overrides.
func TestJavaGithub725h(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "Github725g", "Github725g", false, true, false, 1,
		"-noGenerateSpecTE", "-config", "Github725h.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want ExitStatusViolationSafety; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInvariantViolatedInitial, "Inv", "/\\ b = 0\n/\\ q = 0\n/\\ s = <<1, 2, 3>>\n")
}
