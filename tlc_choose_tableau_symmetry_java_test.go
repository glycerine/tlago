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
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original ChooseTableauSymmetryTestA.testSpec and constructor settings.
func TestJavaChooseTableauSymmetryA(t *testing.T) {
	runJavaChooseTableauSymmetryA(t)
}

func runJavaChooseTableauSymmetryA(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "ChooseTableauSymmetryTestA.json")}
	r := runJavaTLCModelTest(t, "ChooseTableauSymmetryMCa", append(args, extraArgs...)...)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "13", "6", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_144481269489929000", "AllReady")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"arr = (a :> \"ready\" @@ b :> \"ready\")", "arr = (a :> \"busy\" @@ b :> \"ready\")", "arr = (a :> \"busy\" @@ b :> \"busy\")", "arr = (a :> \"done\" @@ b :> \"busy\")", "arr = (a :> \"ready\" @@ b :> \"busy\")"}, true)
	actions := []string{
		"<Init line 5, col 9 to line 5, col 37 of module ChooseTableauSymmetry>",
		"<Ready(a) line 7, col 13 to line 8, col 47 of module ChooseTableauSymmetry>",
		"<Ready(b) line 7, col 13 to line 8, col 47 of module ChooseTableauSymmetry>",
		"<Busy(a) line 10, col 12 to line 11, col 46 of module ChooseTableauSymmetry>",
		"<Done(a) line 13, col 12 to line 14, col 47 of module ChooseTableauSymmetry>",
	}
	for i, record := range javaTLCRecords(r, tlc.ECTLCStatePrint2) {
		if record.StateInfo.Info != actions[i] {
			t.Fatalf("state%d action=%v, want %s", i+1, record.StateInfo.Info, actions[i])
		}
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "3", "<Ready(a) line 7, col 13 to line 8, col 47 of module ChooseTableauSymmetry>")
	requireJavaTLCUncovered(t, r)
	return r
}

// Original ChooseTableauSymmetryTestA_TTraceTest.testSpec and inherited settings.
func TestJavaChooseTableauSymmetryATTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "ChooseTableauSymmetryTestATTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaChooseTableauSymmetryA(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "ChooseTableauSymmetryMCa", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "6", "5", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaTTraceActions(t, r, []string{"arr = (a :> \"ready\" @@ b :> \"ready\")", "arr = (a :> \"busy\" @@ b :> \"ready\")", "arr = (a :> \"busy\" @@ b :> \"busy\")", "arr = (a :> \"done\" @@ b :> \"busy\")", "arr = (a :> \"ready\" @@ b :> \"busy\")"}, "<_init line 23, col 5 to line 23, col 28 of module ChooseTableauSymmetryTestATTrace>", "<_next line 27, col 5 to line 33, col 33 of module ChooseTableauSymmetryTestATTrace>")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "3", "<_next line 27, col 5 to line 33, col 33 of module ChooseTableauSymmetryTestATTrace>")
	requireJavaTLCUncovered(t, r)
}
