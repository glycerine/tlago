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
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original RandomSubsetATest_TTraceTest and inherited RandomSubset_TTrace.testSpec.
func TestJavaRandomSubsetATTrace(t *testing.T) {
	runJavaRandomSubsetSeedTTrace(t, "RandomSubsetATest", 15041980, 129202, 100000008, "<<113, 155, 35>>", "<<2708, 3048, 3471>>")
}

// Original RandomSubsetBTest_TTraceTest and inherited RandomSubset_TTrace.testSpec.
func TestJavaRandomSubsetBTTrace(t *testing.T) {
	runJavaRandomSubsetSeedTTrace(t, "RandomSubsetBTest", 918347981374, 604, 100000000, "<<15, 152, 5>>", "<<86, 170, 1612>>")
}

func runJavaRandomSubsetSeedTTrace(t *testing.T, class string, seed int64, x, y int, p, q string) {
	t.Helper()
	generated := filepath.Join(t.TempDir(), class+"TTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaRandomSubsetSeed(t, seed, x, y, p, q, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "RandomSubset", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	state := func(z string) string {
		return fmt.Sprintf("/\\ p = %s\n/\\ q = %s\n/\\ x = %d\n/\\ y = %d\n/\\ z = %s", p, q, x, y, z)
	}
	requireJavaRandomSubsetTrace(t, r, []string{state("TRUE"), state("FALSE")}, true)
}

// Original RandomSubsetNextTest_TTraceTest.testSpec and inherited settings.
func TestJavaRandomSubsetNextTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "RandomSubsetNextTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaRandomSubsetNext(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "RandomSubsetNext", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "11", "11", "0")
	if len(javaTLCRecords(r, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ x = 23\n/\\ y = 0", "/\\ x = 26\n/\\ y = 1", "/\\ x = 18\n/\\ y = 2", "/\\ x = 29\n/\\ y = 3", "/\\ x = 189\n/\\ y = 4", "/\\ x = 19\n/\\ y = 5", "/\\ x = 92\n/\\ y = 6", "/\\ x = 250\n/\\ y = 7", "/\\ x = 41\n/\\ y = 8", "/\\ x = 52\n/\\ y = 9", "/\\ x = 78\n/\\ y = 10",
	}, true)
	requireJavaTLCUncovered(t, r)
}

// Original RandomSubsetNextT4Test_TTraceTest.testSpec and four-worker override.
func TestJavaRandomSubsetNextT4TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "RandomSubsetNextT4TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaRandomSubsetNextT4(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheckWithWorkers(t, "RandomSubsetNext", generated, false, true, false, 4)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG present")
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
		vals := record.StateInfo.State.GetVals()
		y, _ := vals.Get2(tlc.UniqueStringOf("y"))
		if got := y.(*tlc.IntValue).Val; got != int32(cnt) {
			t.Fatalf("y=%d, want %d", got, cnt)
		}
		cnt++
		x, _ := record.StateInfo.State.GetVals().Get2(tlc.UniqueStringOf("x"))
		if got := x.(*tlc.IntValue).Val; got < 1 || got > 1000 {
			t.Fatalf("x=%d outside 1..1000", got)
		}
		if record.StateNumber != cnt {
			t.Fatalf("ordinal=%d, want %d", record.StateNumber, cnt)
		}
	}
}

// Original RandomSubsetTest_TTraceTest.testSpec, including its literal x/y bound.
func TestJavaRandomSubsetModelTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "RandomSubsetTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaRandomSubsetModel(t, "-teSpecOutDir", generated) }) {
		return
	}
	r := runJavaTTraceRecheck(t, "RandomSubset", generated, false, true)
	if r.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit=%d, want VIOLATION_SAFETY; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated1, "1")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	records := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(records) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	if len(records) != 2 {
		t.Fatalf("trace length=%d, want 2", len(records))
	}
	first := records[0].StateInfo
	// Original inherited debugger selects extended states.
	if first.Info != "<_init line 31, col 5 to line 35, col 24 of module RandomSubsetTestTTrace>" {
		t.Fatalf("initial action=%v", first.Info)
	}
	firstState := first.State.GetVals()
	if firstState.Len() != 5 {
		t.Fatalf("first state size=%d, want 5", firstState.Len())
	}
	x, _ := firstState.Get2(tlc.UniqueStringOf("x"))
	firstX := x.(*tlc.IntValue)
	if firstX.Val < 1 || firstX.Val > 100000000 {
		t.Fatalf("first x=%d outside original bound", firstX.Val)
	}
	y, _ := firstState.Get2(tlc.UniqueStringOf("y"))
	firstY := y.(*tlc.IntValue)
	// Preserve firstX in the original y upper-bound assertion.
	if firstY.Val < 100000000 || firstX.Val > 100000010 {
		t.Fatalf("original y/x bound failed: y=%d x=%d", firstY.Val, firstX.Val)
	}
	z, _ := firstState.Get2(tlc.UniqueStringOf("z"))
	if eq, err := tlc.BoolTrue.Equal(z); err != nil || !eq {
		t.Fatalf("first z differs from TRUE: %v", err)
	}
	firstP, _ := firstState.Get2(tlc.UniqueStringOf("p"))
	javaRandomSubsetTupleIn(t, firstP, 200)
	firstQ, _ := firstState.Get2(tlc.UniqueStringOf("q"))
	javaRandomSubsetTupleIn(t, firstQ, 4000)
	second := records[1].StateInfo
	if !strings.HasPrefix(second.Info.(string), "<_next line 39, col 5 to line 51, col 29 of module RandomSubsetTestTTrace>") {
		t.Fatalf("second action=%v", second.Info)
	}
	secondState := second.State.GetVals()
	if secondState.Len() != 5 {
		t.Fatalf("second state size=%d, want 5", secondState.Len())
	}
	secondX, _ := secondState.Get2(tlc.UniqueStringOf("x"))
	if got := secondX.(*tlc.IntValue).Val; got != firstX.Val {
		t.Fatalf("x changed to %d", got)
	}
	secondY, _ := secondState.Get2(tlc.UniqueStringOf("y"))
	if got := secondY.(*tlc.IntValue).Val; got != firstY.Val {
		t.Fatalf("y changed to %d", got)
	}
	secondP, _ := secondState.Get2(tlc.UniqueStringOf("p"))
	if eq, err := firstP.Equal(secondP); err != nil || !eq {
		t.Fatalf("p changed: %v", err)
	}
	secondQ, _ := secondState.Get2(tlc.UniqueStringOf("q"))
	if eq, err := firstQ.Equal(secondQ); err != nil || !eq {
		t.Fatalf("q changed: %v", err)
	}
	secondZ, _ := secondState.Get2(tlc.UniqueStringOf("z"))
	if eq, err := tlc.BoolFalse.Equal(secondZ); err != nil || !eq {
		t.Fatalf("second z differs from FALSE: %v", err)
	}
	requireJavaTLCUncovered(t, r)
}
