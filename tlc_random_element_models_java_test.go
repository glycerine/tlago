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
	"fmt"
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

func javaRandomElementTrace(xs []int) []string {
	trace := make([]string, len(xs))
	for y, x := range xs {
		trace[y] = fmt.Sprintf("/\\ x = %d\n/\\ y = %d", x, y)
	}
	return trace
}

// Original RandomElementTest.test, including inherited safety-violation exit.
func TestJavaRandomElement(t *testing.T) {
	r := runJavaTLCModelTest(t, "RandomElement", "-seed", "8006803340504660123", "-dumpTrace", "json", filepath.Join(t.TempDir(), "RandomElementTest.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG recorded")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "932", "855", "388")
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	requireJavaRandomSubsetTrace(t, r, javaRandomElementTrace([]int{843, 920, 483, 173, 590, 104, 785, 463, 443, 151, 767}), true)
	requireJavaTLCUncovered(t, r)
}

// Original RandomElementXandYTest.test with its exact seed and three-state trace.
func TestJavaRandomElementXandY(t *testing.T) {
	r := runJavaTLCModelTest(t, "RandomElementXandY", "-seed", "8006642976694192746", "-dumpTrace", "json", filepath.Join(t.TempDir(), "RandomElementXandYTest.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ x = 0\n/\\ y = 0", "/\\ x = 1\n/\\ y = 1", "/\\ x = 0\n/\\ y = 1"}, true)
	requireJavaTLCUncovered(t, r)
}

// Original RandomElementT4Test.test with its four-worker override.
func TestJavaRandomElementT4(t *testing.T) {
	r := runJavaTLCModelTestWithWorkers(t, "RandomElement", true, true, 4, "-seed", "15041980", "-dumpTrace", "json", filepath.Join(t.TempDir(), "RandomElementT4Test.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	records := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(records) != 11 {
		t.Fatalf("trace length=%d, want 11", len(records))
	}
	cnt := 0
	for _, record := range records {
		state := record.StateInfo.State
		y := state.Lookup(tlc.UniqueStringOf("y")).(*tlc.IntValue).Val
		if y != int32(cnt) {
			t.Fatalf("y=%d, want %d", y, cnt)
		}
		cnt++
		x := state.Lookup(tlc.UniqueStringOf("x")).(*tlc.IntValue).Val
		if x < 1 || x > 1000 {
			t.Fatalf("x=%d outside 1..1000", x)
		}
		if record.StateNumber != cnt {
			t.Fatalf("ordinal=%d, want %d", record.StateNumber, cnt)
		}
	}
}

// Original RandomElementSimulationTest.test, including debugger=false and
// CommonTestCase's action-aware trace assertion overload.
func TestJavaRandomElementSimulation(t *testing.T) {
	r := runJavaTLCModelTestWithDebugger(t, "RandomElement", true, true, false, 1, "-seed", "8006803340504660123", "-simulate", "num=1", "-dumpTrace", "json", filepath.Join(t.TempDir(), "RandomElementSimulationTest.json"))
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG recorded")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	expected := javaRandomElementTrace([]int{843, 843, 227, 227, 502, 535, 277, 327, 729, 550, 318})
	records := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(records) != len(expected) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expected))
	}
	for i, record := range records {
		action := "<Next line 9, col 9 to line 11, col 21 of module RandomElement>"
		if i == 0 {
			action = "<Init line 6, col 9 to line 7, col 16 of module RandomElement>"
		}
		if record.StateInfo.Info != action {
			t.Fatalf("action %d=%v, want %s", i, record.StateInfo.Info, action)
		}
		if got := replJavaTrim(record.StateInfo.String()); got != expected[i] {
			t.Fatalf("state %d=%q, want %q", i, got, expected[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
	requireJavaTLCUncovered(t, r)
}
