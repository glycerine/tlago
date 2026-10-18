/*******************************************************************************
 * Copyright (c) 2015, 2020 Microsoft Research. All rights reserved.
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

// Original LoopTest_TTraceTest.testSpec and full original generation assertions.
func TestJavaLivenessLoopTTrace(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	generated := filepath.Join(t.TempDir(), "LoopTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaLivenessLoop(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "Loop", generated, false, true)
	retainJavaCodePlexGraphs(t)
	requireJavaSimulationTTrace(t, r)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "3", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	requireJavaNodeAndPtrSizes(t, 80, 48)
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 2"}, true)
	requireJavaCodePlexStuttering(t, r, 4)
	requireJavaTLCUncovered(t, r)
}

// Original LoopTestForcedPartial_TTraceTest static initializer and whole method.
func TestJavaLivenessLoopForcedPartialTTrace(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	t.Setenv("tlc2.tool.liveness.ILiveCheck.testing", "true")
	generated := filepath.Join(t.TempDir(), "LoopTestForcedPartialTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaLivenessLoopTestForcedPartial(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "LoopTestForcedPartial", generated, false, true)
	retainJavaCodePlexGraphs(t)
	requireJavaSimulationTTrace(t, r)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1", "1", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0"}, true)
	requireJavaCodePlexStuttering(t, r, 2)
}

// Whole original Test3_TTraceTest.testSpec, constructor and inherited settings.
func TestJavaLivenessTest3TTrace(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	generated := filepath.Join(t.TempDir(), "Test3TTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaLivenessTest3(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "Test3", generated, false, true)
	retainJavaCodePlexGraphs(t)
	requireJavaSimulationTTrace(t, r)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "4", "0")
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = 0", "x = 2"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
	requireJavaTLCUncovered(t, r)
}

// Whole original UnsymmetricModelCheckerTestA_TTraceTest.testSpec.
func TestJavaLivenessUnsymmetricModelCheckerTestATTrace(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	generated := filepath.Join(t.TempDir(), "UnsymmetricModelCheckerTestATTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaLivenessUnsymmetricModelCheckerTestA(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "UnsymmetricModelCheckerTestA", generated, false, true)
	retainJavaCodePlexGraphs(t)
	requireJavaSimulationTTrace(t, r)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCLiveImplied, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1", "")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTTraceActions(t, r, []string{"x = a", "x = 1"},
		"<_init line 23, col 5 to line 23, col 24 of module UnsymmetricModelCheckerTestATTrace>",
		"<_next line 27, col 5 to line 33, col 29 of module UnsymmetricModelCheckerTestATTrace>")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
}

// Whole original OneBitMutexNoSymmetryTest_TTraceTest.testSpec.
func TestJavaLivenessOneBitMutexNoSymmetryTTrace(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	generated := filepath.Join(t.TempDir(), "OneBitMutexNoSymmetryTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaLivenessOneBitMutexNoSymmetry(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "OneBitMutexNoSymmetry", generated, false, true)
	retainJavaCodePlexGraphs(t)
	requireJavaSimulationTTrace(t, r)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "14", "13", "0")
	requireJavaNodeAndPtrSizes(t, 380, 208)
	expected := javaOneBitMutexGeneratedTrace()
	requireJavaRandomSubsetTrace(t, r, expected, true)
	loop := javaTLCRecords(r, tlc.ECTLCBackToState)
	if len(loop) == 0 {
		t.Fatal("TLC_BACK_TO_STATE absent")
	}
	if len(loop[0].Params) < 2 {
		t.Fatal("back-loop record has fewer than two arguments")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "4",
		"<_next line 42, col 5 to line 54, col 37 of module OneBitMutexNoSymmetryTestTTrace>")
}

func javaOneBitMutexGeneratedTrace() []string {
	return []string{
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> FALSE @@ B :> FALSE)\n/\\ pc = (A :> \"ncs\" @@ B :> \"ncs\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> FALSE @@ B :> FALSE)\n/\\ pc = (A :> \"e1\" @@ B :> \"ncs\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {B} @@ B :> {})\n/\\ x = (A :> TRUE @@ B :> FALSE)\n/\\ pc = (A :> \"e2\" @@ B :> \"ncs\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {B} @@ B :> {})\n/\\ x = (A :> TRUE @@ B :> FALSE)\n/\\ pc = (A :> \"e2\" @@ B :> \"e1\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> TRUE @@ B :> FALSE)\n/\\ pc = (A :> \"e3\" @@ B :> \"e1\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {A})\n/\\ x = (A :> TRUE @@ B :> TRUE)\n/\\ pc = (A :> \"e3\" @@ B :> \"e2\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> TRUE @@ B :> TRUE)\n/\\ pc = (A :> \"e3\" @@ B :> \"e3\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> TRUE @@ B :> TRUE)\n/\\ pc = (A :> \"e3\" @@ B :> \"e4\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> TRUE @@ B :> TRUE)\n/\\ pc = (A :> \"e4\" @@ B :> \"e4\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> FALSE @@ B :> TRUE)\n/\\ pc = (A :> \"e5\" @@ B :> \"e4\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> FALSE @@ B :> FALSE)\n/\\ pc = (A :> \"e5\" @@ B :> \"e5\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> FALSE @@ B :> FALSE)\n/\\ pc = (A :> \"e1\" @@ B :> \"e5\")\n/\\ other = (A :> B @@ B :> A)",
		"/\\ unchecked = (A :> {} @@ B :> {})\n/\\ x = (A :> FALSE @@ B :> FALSE)\n/\\ pc = (A :> \"e1\" @@ B :> \"e1\")\n/\\ other = (A :> B @@ B :> A)",
	}
}
