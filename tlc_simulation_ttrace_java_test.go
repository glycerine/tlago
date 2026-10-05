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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Common assertions of the whole original generated simulation test methods.
func requireJavaSimulationTTrace(t *testing.T, r *tlc.Result) {
	t.Helper()
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	for _, code := range []int{tlc.ECTLCTemporalPropertyViolated, tlc.ECTLCCounterExample, tlc.ECTLCStatePrint2} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("required code %d absent", code)
		}
	}
}

// Whole inherited AbstractExample_TTrace.testSpec for each concrete subclass.
func runJavaSimulationExampleTTrace(t *testing.T, class, name, property string, experimental bool) {
	t.Helper()
	generated := filepath.Join(t.TempDir(), class+"TTrace.tla")
	if !t.Run("original", func(t *testing.T) {
		runJavaSimulationExample(t, name, property, experimental, !experimental, "-teSpecOutDir", generated)
	}) {
		return
	}
	r := runJavaTTraceRecheck(t, name, generated, false, true)
	requireJavaSimulationTTrace(t, r)
	expected := make([]string, 10)
	for i := range expected {
		expected[i] = fmt.Sprintf("x = %d", i)
	}
	requireJavaRandomSubsetTrace(t, r, expected, true)
	// Whole CommonTestCase.assertBackToState(1).
	loop := javaTLCRecords(r, tlc.ECTLCBackToState)
	if len(loop) == 0 {
		t.Fatal("TLC_BACK_TO_STATE absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
}
func TestJavaSimulationExample1TTrace(t *testing.T) {
	runJavaSimulationExampleTTrace(t, "Example1Test", "Example1", "Liveness1", false)
}
func TestJavaSimulationExample2TTrace(t *testing.T) {
	runJavaSimulationExampleTTrace(t, "Example2Test", "Example2", "Liveness2", false)
}
func TestJavaLiveCheckSimulationExample1TTrace(t *testing.T) {
	runJavaSimulationExampleTTrace(t, "LiveCheckExample1Test", "Example1", "Liveness1", true)
}
func TestJavaLiveCheckSimulationExample2TTrace(t *testing.T) {
	runJavaSimulationExampleTTrace(t, "LiveCheckExample2Test", "Example2", "Liveness2", true)
}

// Whole SimulationTest2a_TTraceTest/LiveCheckSimulationTest2a_TTraceTest methods.
func runJavaSimulationTest2aTTrace(t *testing.T, experimental bool) {
	t.Helper()
	class := "SimulationTest2a"
	if experimental {
		class = "LiveCheckSimulationTest2a"
	}
	generated := filepath.Join(t.TempDir(), class+"TTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaSimulationTest2a(t, experimental, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "Test2a", generated, false, true)
	requireJavaSimulationTTrace(t, r)
	trace := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if trace[0].StateInfo == nil {
		t.Fatal("initial TLCStateInfo absent")
	}
	if got := replJavaTrim(trace[0].StateInfo.String()); got != "x = 0" {
		t.Fatalf("initial state=%q", got)
	}
	if trace[0].StateNumber != 1 {
		t.Fatalf("initial ordinal=%d, want 1", trace[0].StateNumber)
	}
	last := trace[len(trace)-1]
	if last.StateInfo == nil {
		t.Fatal("last TLCStateInfo absent")
	}
	if got := replJavaTrim(last.StateInfo.String()); got != "x = 4" {
		t.Fatalf("last state=%q", got)
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint3)) != 0 {
		t.Fatal("TLC_STATE_PRINT3 recorded")
	}
	trace = javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(trace) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	if trace[0].StateNumber != 1 {
		t.Fatalf("loop initial ordinal=%d, want 1", trace[0].StateNumber)
	}
}
func TestJavaSimulationTest2aTTrace(t *testing.T)          { runJavaSimulationTest2aTTrace(t, false) }
func TestJavaLiveCheckSimulationTest2aTTrace(t *testing.T) { runJavaSimulationTest2aTTrace(t, true) }

// Whole original StutteringTest_TTraceTest.testSpec.
func TestJavaSimulationStutteringTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "StutteringTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaSimulationStuttering(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "CodePlexBug08", generated, false, true)
	requireJavaSimulationTTrace(t, r)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint3)) == 0 {
		t.Fatal("TLC_STATE_PRINT3 absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBackToState)) != 0 {
		t.Fatal("Trace shows Back to state...")
	}
}
