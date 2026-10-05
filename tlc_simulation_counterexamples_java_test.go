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

// Whole AbstractExampleTestCase.testSpec, with its concrete constructor fields.
func runJavaSimulationExample(t *testing.T, name, property string, experimental, postcondition bool, extraArgs ...string) {
	t.Helper()
	if experimental {
		t.Setenv("TLAGO_SIMULATOR_EXPERIMENTAL_LIVENESS", "true")
	}
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), name+"Test.json"), "-simulate", "-depth", "11"}
	r := runJavaTLCModelTestWithRoot(t, name, name, true, true, true, 1, append(args, extraArgs...)...)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStatsSimu, "12")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, property)
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	trace := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(trace) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	expectedTrace := make([]string, 10)
	expectedActions := make([]string, 10)
	for i := range expectedTrace {
		expectedTrace[i] = fmt.Sprintf("/\\ x = %d\n/\\ l = %d", i, i+1)
		expectedActions[i] = "<Next line 8, col 9 to line 8, col 27 of module " + name + ">"
	}
	expectedActions[0] = "<Init line 6, col 9 to line 6, col 13 of module " + name + ">"
	if len(trace) != len(expectedTrace) {
		t.Fatalf("trace size=%d, want %d", len(trace), len(expectedTrace))
	}
	for i, record := range trace {
		if record.StateInfo == nil {
			t.Fatalf("state %d has no TLCStateInfo", i+1)
		}
		if record.StateInfo.Info != expectedActions[i] {
			t.Fatalf("action %d=%v, want %q", i+1, record.StateInfo.Info, expectedActions[i])
		}
		if got := replJavaTrim(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
	// Whole CommonTestCase.assertBackToState(1, action).
	loop := javaTLCRecords(r, tlc.ECTLCBackToState)
	if len(loop) == 0 {
		t.Fatal("TLC_BACK_TO_STATE absent")
	}
	if len(loop[0].Params) < 2 {
		t.Fatal("back-loop record has fewer than two arguments")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1", expectedActions[1])
	// Source assumeTrue(assertPostCondition) occurs after all counterexample checks.
	if !postcondition {
		t.Skip("original assumeTrue(assertPostCondition) is false")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR recorded")
	}
}

func TestJavaSimulationExample1(t *testing.T) {
	runJavaSimulationExample(t, "Example1", "Liveness1", false, true)
}
func TestJavaSimulationExample2(t *testing.T) {
	runJavaSimulationExample(t, "Example2", "Liveness2", false, true)
}
func TestJavaLiveCheckSimulationExample1(t *testing.T) {
	runJavaSimulationExample(t, "Example1", "Liveness1", true, false)
}
func TestJavaLiveCheckSimulationExample2(t *testing.T) {
	runJavaSimulationExample(t, "Example2", "Liveness2", true, false)
}

// Whole SimulationTest2a and LiveCheckSimulationTest2a.testSpec methods have the
// same assertions; retain their distinct constructors and experimental setting.
func runJavaSimulationTest2a(t *testing.T, experimental bool, extraArgs ...string) {
	t.Helper()
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "SimulationTest2a.json"), "-simulate"}
	if experimental {
		t.Setenv("TLAGO_SIMULATOR_EXPERIMENTAL_LIVENESS", "true")
		args = append(args, "num=100", "-depth", "10")
	} else {
		args = append(args, "-depth", "6")
	}
	r := runJavaTLCModelTestWithRoot(t, "Test2a", "Test2a", true, true, true, 1, append(args, extraArgs...)...)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS", r.ExitStatus)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCStatsSimu} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("required code %d absent", code)
		}
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "Prop1")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	trace := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(trace) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
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
	// Preserve the original repeated STATE_PRINT2/ordinal assertions exactly.
	trace = javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(trace) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	if trace[0].StateNumber != 1 {
		t.Fatalf("loop initial ordinal=%d, want 1", trace[0].StateNumber)
	}
}
func TestJavaSimulationTest2a(t *testing.T)          { runJavaSimulationTest2a(t, false) }
func TestJavaLiveCheckSimulationTest2a(t *testing.T) { runJavaSimulationTest2a(t, true) }

// Whole original liveness.simulation.StutteringTest.testSpec and constructor.
func TestJavaSimulationStuttering(t *testing.T) { runJavaSimulationStuttering(t) }

func runJavaSimulationStuttering(t *testing.T, extraArgs ...string) {
	t.Helper()
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "StutteringTest.json"), "-simulate"}
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "MC", true, true, true, 1, append(args, extraArgs...)...)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS", r.ExitStatus)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCStatsSimu} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("required code %d absent", code)
		}
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	for _, code := range []int{tlc.ECTLCTemporalPropertyViolated, tlc.ECTLCCounterExample, tlc.ECTLCStatePrint2, tlc.ECTLCStatePrint3} {
		if len(javaTLCRecords(r, code)) == 0 {
			t.Fatalf("required code %d absent", code)
		}
	}
	if len(javaTLCRecords(r, tlc.ECTLCBackToState)) != 0 {
		t.Fatal("Trace shows Back to state...")
	}
}
