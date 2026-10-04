/*******************************************************************************
 * Copyright (c) 2018 Microsoft Research. All rights reserved.
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
	"fmt"
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Original RandomSubsetEmptyTest.testSpec, with its noGenerateSpec override.
func TestJavaRandomSubsetEmpty(t *testing.T) {
	result := runJavaTLCModelTest(t, "RandomSubsetEmpty", "-noGenerateSpecTE", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.RandomSubsetEmptyTest.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECTLCAssumptionFalse, tlc.ECTLCAssumptionEvaluationError, tlc.ECGeneral} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Fatalf("unexpected code %d: %v", code, records)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "0", "0", "0")
}

// Original CommonTestCase.assertTraceWith(List<Object>, List<String>).
func requireJavaRandomSubsetTrace(t *testing.T, result *tlc.Result, expected []string, extended bool) {
	t.Helper()
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != len(expected) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expected))
	}
	for i, record := range records {
		if record.StateInfo == nil {
			t.Fatalf("state %d has no TLCStateInfo", i+1)
		}
		info, ok := record.StateInfo.Info.(string)
		if !ok {
			t.Fatalf("state %d action is %T", i+1, record.StateInfo.Info)
		}
		if i == 0 && !extended {
			if info != "<Initial predicate>" {
				t.Fatalf("initial action=%q", info)
			}
		} else if info == "<Initial predicate>" || strings.HasPrefix(info, "<Action") {
			t.Fatalf("state %d action=%q", i+1, info)
		}
		if got := replJavaTrim(record.StateInfo.String()); got != expected[i] {
			t.Fatalf("state %d=%q, want %q", i+1, got, expected[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("state ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
}

func runJavaRandomSubsetSeed(t *testing.T, seed int64, x, y int, p, q string) {
	t.Helper()
	result := runJavaTLCModelTest(t, "RandomSubset", "-seed", strconv.FormatInt(seed, 10), "-dumpTrace", "json", filepath.Join(t.TempDir(), "RandomSubset.json"))
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want VIOLATION_SAFETY", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("unexpected GENERAL: %v", records)
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInitGenerated1, "8008")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "8009", "8009", "8007")
	depthRecords := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depthRecords) == 0 || len(depthRecords[0].Params) == 0 {
		t.Fatal("search depth absent")
	}
	depth, err := strconv.ParseInt(depthRecords[0].Params[0], 10, 32)
	if err != nil || depth != 2 {
		t.Fatalf("search depth=%d/%v, want 2", depth, err)
	}
	if len(javaTLCRecords(result, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 not recorded")
	}
	state := func(z string) string {
		return fmt.Sprintf("/\\ p = %s\n/\\ q = %s\n/\\ x = %d\n/\\ y = %d\n/\\ z = %s", p, q, x, y, z)
	}
	// The original one-worker debugger setup selects TLCStateMutExt.
	requireJavaRandomSubsetTrace(t, result, []string{state("TRUE"), state("FALSE")}, true)
}

// Original RandomSubsetATest and RandomSubsetBTest constructors and inherited method.
func TestJavaRandomSubsetA(t *testing.T) {
	runJavaRandomSubsetSeed(t, 15041980, 129202, 100000008, "<<113, 155, 35>>", "<<2708, 3048, 3471>>")
}
func TestJavaRandomSubsetB(t *testing.T) {
	runJavaRandomSubsetSeed(t, 918347981374, 604, 100000000, "<<15, 152, 5>>", "<<86, 170, 1612>>")
}

func javaRandomSubsetTupleIn(t *testing.T, value tlc.Value, n int) {
	t.Helper()
	tuple := value.(*tlc.TupleValue)
	if len(tuple.Elems) != 3 {
		t.Fatalf("tuple length=%d, want 3", len(tuple.Elems))
	}
	for _, elem := range tuple.Elems {
		component := elem.(*tlc.IntValue).Val
		if component < 1 || component > int32(n) {
			t.Fatalf("tuple component=%d, want 1..%d", component, n)
		}
	}
}

// Original RandomSubsetTest.testSpec with every active trace assertion.
func TestJavaRandomSubsetModel(t *testing.T) {
	result := runJavaTLCModelTest(t, "RandomSubset", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.RandomSubsetTest.json"))
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want VIOLATION_SAFETY", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("unexpected GENERAL: %v", records)
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInitGenerated1, "8008")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "8009", "8009", "8007")
	depths := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depths) == 0 || len(depths[0].Params) == 0 {
		t.Fatal("search depth absent")
	}
	depth, err := strconv.ParseInt(depths[0].Params[0], 10, 32)
	if err != nil || depth != 2 {
		t.Fatalf("depth=%d/%v, want 2", depth, err)
	}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) == 0 {
		t.Fatal("TLC_STATE_PRINT2 not recorded")
	}
	if len(records) != 2 {
		t.Fatalf("trace length=%d, want 2", len(records))
	}
	first := records[0].StateInfo
	// Original source setup enables one-worker debugger and thus extended states.
	if first.Info != "<Init line 11, col 9 to line 15, col 44 of module RandomSubset>" {
		t.Fatalf("initial action=%v", first.Info)
	}
	firstState := first.State.Values()
	if firstState.Len() != 5 {
		t.Fatalf("first state size=%d, want 5", firstState.Len())
	}
	firstX := first.State.Lookup(tlc.UniqueStringOf("x")).(*tlc.IntValue)
	if firstX.Val < 1 || firstX.Val > 100000000 {
		t.Fatalf("first x=%d outside original bound", firstX.Val)
	}
	firstY := first.State.Lookup(tlc.UniqueStringOf("y")).(*tlc.IntValue)
	// Preserve the original firstX upper bound in this source assertion.
	if firstY.Val < 100000000 || firstX.Val > 100000010 {
		t.Fatalf("original y/x bound failed: y=%d x=%d", firstY.Val, firstX.Val)
	}
	if eq, err := tlc.BoolTrue.Equal(first.State.Lookup(tlc.UniqueStringOf("z"))); err != nil || !eq {
		t.Fatalf("first z differs from TRUE: %v", err)
	}
	firstP := first.State.Lookup(tlc.UniqueStringOf("p"))
	javaRandomSubsetTupleIn(t, firstP, 200)
	firstQ := first.State.Lookup(tlc.UniqueStringOf("q"))
	javaRandomSubsetTupleIn(t, firstQ, 4000)
	second := records[1].StateInfo
	if !strings.HasPrefix(second.Info.(string), "<Next line 17, col 9 to line 18, col 21 of module RandomSubset>") {
		t.Fatalf("second action=%v", second.Info)
	}
	if second.State.Values().Len() != 5 {
		t.Fatalf("second state size=%d, want 5", second.State.Values().Len())
	}
	if got := second.State.Lookup(tlc.UniqueStringOf("x")).(*tlc.IntValue).Val; got != firstX.Val {
		t.Fatalf("x changed to %d", got)
	}
	if got := second.State.Lookup(tlc.UniqueStringOf("y")).(*tlc.IntValue).Val; got != firstY.Val {
		t.Fatalf("y changed to %d", got)
	}
	if eq, err := firstP.Equal(second.State.Lookup(tlc.UniqueStringOf("p"))); err != nil || !eq {
		t.Fatalf("p changed: %v", err)
	}
	if eq, err := firstQ.Equal(second.State.Lookup(tlc.UniqueStringOf("q"))); err != nil || !eq {
		t.Fatalf("q changed: %v", err)
	}
	if eq, err := tlc.BoolFalse.Equal(second.State.Lookup(tlc.UniqueStringOf("z"))); err != nil || !eq {
		t.Fatalf("second z differs from FALSE: %v", err)
	}
	requireJavaTLCUncovered(t, result)
}

// Original RandomSubsetSetOfFcnsTest.testSpec (commented-out assertions omitted).
func TestJavaRandomSubsetSetOfFcns(t *testing.T) {
	result := runJavaTLCModelTest(t, "RandomSubsetSetOfFcns", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.RandomSubsetSetOfFcnsTest.json"))
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want SUCCESS", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("unexpected GENERAL: %v", records)
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInitGenerated1, "1000")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "2000", "1000", "0")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCUncovered(t, result)
}

// Original RandomSubsetNextTest.testSpec, exact source seed/counts/trace.
func TestJavaRandomSubsetNext(t *testing.T) {
	result := runJavaTLCModelTest(t, "RandomSubsetNext", "-seed", "15041980", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.RandomSubsetNextTest.json"))
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want VIOLATION_SAFETY", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "67291", "7729", "999")
	if len(javaTLCRecords(result, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT not recorded")
	}
	requireJavaRandomSubsetTrace(t, result, []string{
		"/\\ x = 23\n/\\ y = 0", "/\\ x = 26\n/\\ y = 1", "/\\ x = 18\n/\\ y = 2", "/\\ x = 29\n/\\ y = 3", "/\\ x = 189\n/\\ y = 4", "/\\ x = 19\n/\\ y = 5", "/\\ x = 92\n/\\ y = 6", "/\\ x = 250\n/\\ y = 7", "/\\ x = 41\n/\\ y = 8", "/\\ x = 52\n/\\ y = 9", "/\\ x = 78\n/\\ y = 10",
	}, true)
	requireJavaTLCUncovered(t, result)
}

// Original RandomSubsetNextT4Test.testSpec and its four-worker override.
func TestJavaRandomSubsetNextT4(t *testing.T) {
	result := runJavaTLCModelTestWithWorkers(t, "RandomSubsetNext", true, true, 4, "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.RandomSubsetNextT4Test.json"))
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want VIOLATION_SAFETY", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT not recorded")
	}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != 11 {
		t.Fatalf("trace length=%d, want 11", len(records))
	}
	cnt := 0
	for _, record := range records {
		state := record.StateInfo.State
		if y := state.Lookup(tlc.UniqueStringOf("y")).(*tlc.IntValue).Val; y != int32(cnt) {
			t.Fatalf("y=%d, want %d", y, cnt)
		}
		cnt++
		x := state.Lookup(tlc.UniqueStringOf("x")).(*tlc.IntValue).Val
		if x < 1 || x > 1000 {
			t.Fatalf("x=%d outside 1..1000", x)
		}
		if record.StateNumber != cnt {
			t.Fatalf("state ordinal=%d, want %d", record.StateNumber, cnt)
		}
	}
}

// Original RandomSubsetNextTuplesTest.testSpec, including every tuple component.
func TestJavaRandomSubsetNextTuples(t *testing.T) {
	result := runJavaTLCModelTest(t, "RandomSubsetNextTuples", "-dumpTrace", "json", filepath.Join(t.TempDir(), "tlc2.tool.RandomSubsetNextTuplesTest.json"))
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want VIOLATION_SAFETY", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCBug)) != 0 {
		t.Fatal("TLC_BUG recorded")
	}
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("unexpected GENERAL: %v", records)
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInitGenerated1, "4")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "5461", "5461", "4095")
	if len(javaTLCRecords(result, tlc.ECTLCBehaviorUpToThisPoint)) == 0 {
		t.Fatal("TLC_BEHAVIOR_UP_TO_THIS_POINT not recorded")
	}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != 7 {
		t.Fatalf("trace length=%d, want 7", len(records))
	}
	y := 0
	for _, record := range records {
		state := record.StateInfo.State
		if got := state.Lookup(tlc.UniqueStringOf("y")).(*tlc.IntValue).Val; got != int32(y) {
			t.Fatalf("y=%d, want %d", got, y)
		}
		y++
		if record.StateNumber != y {
			t.Fatalf("ordinal=%d, want %d", record.StateNumber, y)
		}
		javaRandomSubsetTupleIn(t, state.Lookup(tlc.UniqueStringOf("p")), 200)
		javaRandomSubsetTupleIn(t, state.Lookup(tlc.UniqueStringOf("q")), 4000)
	}
	requireJavaTLCUncovered(t, result)
}
