/*******************************************************************************
 * Copyright (c) 2024 Microsoft Corp.. All rights reserved.
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
	"github.com/glycerine/tlago/tlc"
	"math"
	"path/filepath"
	"sync/atomic"
	"testing"
)

// The source anonymous DotStateWriter subclass delegates before observing flags.
type javaDotConstrainedWriter struct {
	*tlc.StateWriter
	constrained *atomic.Bool
}

func (w *javaDotConstrainedWriter) WriteTransition(state, successor *tlc.TLCStateMut, flags tlc.StateVisitStatus, action *tlc.Action, pred ...tlc.SemanticNode) error {
	if err := w.StateWriter.WriteTransition(state, successor, flags, action, pred...); err != nil {
		return err
	}
	if flags&tlc.StateVisitNotInModel == tlc.StateVisitNotInModel {
		w.constrained.Store(true)
	}
	return nil
}

// Original DotConstrainedTest.testSpec, including its getStateWriter override.
func TestJavaDotConstrained(t *testing.T) {
	var constrained atomic.Bool
	setJavaModelLivenessThreshold(t, math.MaxFloat64)
	r := runJavaTLCModelTestWithRunnerSetup(t, "Github602", "Github602", func(meta, traceDirectory string) []string {
		return []string{"-metadir", meta, "-teSpecOutDir", traceDirectory, "-fp", "0", "-seed", "1", "-workers", "1", "-checkpoint", "0", "-deadlock", "-generateSpecTE", "-debugger", "nosuspend,port=4712,nohalt", "-coverage", "1", "-config", "DotConstrained.cfg", "-dump", "dot,constrained", filepath.Join(meta, "tlc2.tool.DotConstrainedTest.dot"), "-dumpTrace", "json", filepath.Join(t.TempDir(), "DotConstrainedTest.json")}
	}, func(runner *tlc.TLC) {
		name := runner.StateWriter.GetDumpFileName()
		// Close the parser-created writer before opening the overriding writer at
		// the same filename, so its buffered header cannot overwrite the new output.
		if err := runner.StateWriter.Close(); err != nil {
			t.Fatal(err)
		}
		writer, err := tlc.NewDotStateWriter(name, tlc.DotStateWriterOptions{StrictPrefix: true, Constrained: true})
		if err != nil {
			t.Fatal(err)
		}
		runner.StateWriter = &javaDotConstrainedWriter{StateWriter: writer, constrained: &constrained}
	})
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "4", "2", "0")
	if len(javaTLCRecords(r, tlc.ECTLCInvariantViolatedBehavior)) == 0 {
		t.Fatal("TLC_INVARIANT_VIOLATED_BEHAVIOR absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"x = 0", "x = 1", "x = -1"}, true)
	if !constrained.Load() {
		t.Fatal("writer did not receive IsNotInModel")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 4 {
		t.Fatalf("register42=%v, want IntValue4", values[0])
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR recorded")
	}
	requireJavaTLCUncovered(t, r)
}
