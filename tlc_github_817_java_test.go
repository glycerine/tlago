// Copyright (c) 2023, Oracle and/or its affiliates.
// Copyright (c) 2026 NVIDIA Corp. All rights reserved.
/*******************************************************************************
 * Copyright (c) 2024 Microsoft. All rights reserved.
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
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original Github817Test.testSpec and constructor overrides.
func TestJavaGithub817(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github817", "-noGenerateSpecTE", "-config", "Github817.tla")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "0")
	requireJavaTLCUncovered(t, r)
}

// Original Github817bTest.testSpec and constructor overrides.
func TestJavaGithub817b(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github817", "-noGenerateSpecTE", "-config", "Github817b.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want ViolationLiveness; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCActionPropertyViolatedBehavior)) == 0 {
		t.Fatal("TLC_ACTION_PROPERTY_VIOLATED_BEHAVIOR absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{`x = "init"`, `x = "in-progress"`}, true)
}

// Original Github817cTest.testSpec and constructor overrides.
func TestJavaGithub817c(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github817", "-noGenerateSpecTE", "-config", "Github817c.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "0")
}

// Original Github817dTest.testSpec and constructor overrides.
func TestJavaGithub817d(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github817", "-noGenerateSpecTE", "-config", "Github817d.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "0")
}

// Original Github817eTest.testSpec and constructor overrides.
func TestJavaGithub817e(t *testing.T) {
	r := runJavaTLCModelTest(t, "Github817", "-noGenerateSpecTE", "-config", "Github817e.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want Success; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "0")
}
