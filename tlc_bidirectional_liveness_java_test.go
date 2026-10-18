/*******************************************************************************
 * Copyright (c) 2018 Microsoft Research. All rights reserved.
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

// Original BidirectionalTransitions1Test.testSpec and inherited successful exit.
func TestJavaBidirectionalTransitions1(t *testing.T) {
	r := runJavaTLCModelTest(t, "BidirectionalTransitions", "-config", "BidirectionalTransitions1.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "BidirectionalTransitions1Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "13", "3", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCUncovered(t, r)
}

// Original BidirectionalTransitions2Test.testSpec and inherited successful exit.
func TestJavaBidirectionalTransitions2(t *testing.T) {
	r := runJavaTLCModelTest(t, "BidirectionalTransitions", "-config", "BidirectionalTransitions2.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "BidirectionalTransitions2Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "3")
	requireJavaTLCUncovered(t, r)
}

// Original inherited BidirectionalTransitions1BTest/2CTest assertions. The
// generated recheck bases omit the original property-name assertion.
func requireJavaBidirectionalCounterexample(t *testing.T, r *tlc.Result, stats []string, property string, expected []string) {
	t.Helper()
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, stats...)
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if property != "" {
		requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, property)
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, expected, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
}

func runJavaBidirectionalCounterexample(t *testing.T, class, config, property string, stats, expected []string, extraArgs ...string) *tlc.Result {
	t.Helper()
	args := []string{"-config", config, "-dumpTrace", "json", filepath.Join(t.TempDir(), class+".json")}
	r := runJavaTLCModelTest(t, "BidirectionalTransitions", append(args, extraArgs...)...)
	requireJavaBidirectionalCounterexample(t, r, stats, property, expected)
	return r
}

// Original BidirectionalTransitions1BxTest constructor and inherited testSpec.
func TestJavaBidirectionalTransitions1Bx(t *testing.T) {
	runJavaBidirectionalCounterexample(t, "BidirectionalTransitions1BxTest", "BidirectionalTransitions1Bx.cfg", "Prop1Bx", []string{"13", "3", "0"}, []string{"x = 0", "x = 2", "x = 1"})
}

// Original BidirectionalTransitions1BxTest_TTraceTest and inherited generated recheck testSpec.
func TestJavaBidirectionalTransitions1BxTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "BidirectionalTransitions1BxTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) {
		runJavaBidirectionalCounterexample(t, "BidirectionalTransitions1BxTest", "BidirectionalTransitions1Bx.cfg", "Prop1Bx", []string{"13", "3", "0"}, []string{"x = 0", "x = 2", "x = 1"}, "-teSpecOutDir", generated)
	}) {
		return
	}
	r := runJavaTTraceRecheck(t, "BidirectionalTransitions", generated, false, true)
	requireJavaBidirectionalCounterexample(t, r, []string{"4", "3", "0"}, "", []string{"x = 0", "x = 2", "x = 1"})
}

// Original BidirectionalTransitions1ByTest constructor and inherited testSpec.
func TestJavaBidirectionalTransitions1By(t *testing.T) {
	runJavaBidirectionalCounterexample(t, "BidirectionalTransitions1ByTest", "BidirectionalTransitions1By.cfg", "Prop1By", []string{"13", "3", "0"}, []string{"x = 0", "x = 2", "x = 1"})
}

// Original BidirectionalTransitions1ByTest_TTraceTest and inherited generated recheck testSpec.
func TestJavaBidirectionalTransitions1ByTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "BidirectionalTransitions1ByTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) {
		runJavaBidirectionalCounterexample(t, "BidirectionalTransitions1ByTest", "BidirectionalTransitions1By.cfg", "Prop1By", []string{"13", "3", "0"}, []string{"x = 0", "x = 2", "x = 1"}, "-teSpecOutDir", generated)
	}) {
		return
	}
	r := runJavaTTraceRecheck(t, "BidirectionalTransitions", generated, false, true)
	requireJavaBidirectionalCounterexample(t, r, []string{"4", "3", "0"}, "", []string{"x = 0", "x = 2", "x = 1"})
}

// Original BidirectionalTransitions2CxTest constructor and inherited testSpec.
func TestJavaBidirectionalTransitions2Cx(t *testing.T) {
	runJavaBidirectionalCounterexample(t, "BidirectionalTransitions2CxTest", "BidirectionalTransitions2Cx", "Prop2Cx", []string{"9", "4", "0"}, []string{"x = 0", "x = 1", "x = 2", "x = 3"})
}

// Original BidirectionalTransitions2CxTest_TTraceTest and inherited generated recheck testSpec.
func TestJavaBidirectionalTransitions2CxTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "BidirectionalTransitions2CxTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) {
		runJavaBidirectionalCounterexample(t, "BidirectionalTransitions2CxTest", "BidirectionalTransitions2Cx", "Prop2Cx", []string{"9", "4", "0"}, []string{"x = 0", "x = 1", "x = 2", "x = 3"}, "-teSpecOutDir", generated)
	}) {
		return
	}
	r := runJavaTTraceRecheck(t, "BidirectionalTransitions", generated, false, true)
	requireJavaBidirectionalCounterexample(t, r, []string{"5", "4", "0"}, "", []string{"x = 0", "x = 1", "x = 2", "x = 3"})
}

// Original BidirectionalTransitions2CyTest constructor and inherited testSpec.
func TestJavaBidirectionalTransitions2Cy(t *testing.T) {
	runJavaBidirectionalCounterexample(t, "BidirectionalTransitions2CyTest", "BidirectionalTransitions2Cy", "Prop2Cy", []string{"9", "4", "0"}, []string{"x = 0", "x = 1", "x = 2", "x = 3"})
}

// Original BidirectionalTransitions2CyTest_TTraceTest and inherited generated recheck testSpec.
func TestJavaBidirectionalTransitions2CyTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "BidirectionalTransitions2CyTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) {
		runJavaBidirectionalCounterexample(t, "BidirectionalTransitions2CyTest", "BidirectionalTransitions2Cy", "Prop2Cy", []string{"9", "4", "0"}, []string{"x = 0", "x = 1", "x = 2", "x = 3"}, "-teSpecOutDir", generated)
	}) {
		return
	}
	r := runJavaTTraceRecheck(t, "BidirectionalTransitions", generated, false, true)
	requireJavaBidirectionalCounterexample(t, r, []string{"5", "4", "0"}, "", []string{"x = 0", "x = 1", "x = 2", "x = 3"})
}
