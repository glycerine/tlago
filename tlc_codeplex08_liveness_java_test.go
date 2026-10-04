/*******************************************************************************
 * Copyright (c) 2015, 2016 Microsoft Research. All rights reserved.
 * Copyright (c) 2023, 2024, Oracle and/or its affiliates.
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
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// CommonTestCase.assertNodeAndPtrSizes: flush and inspect the actual graph files.
func requireJavaNodeAndPtrSizes(t *testing.T, nodesSize, ptrsSize int64) {
	t.Helper()
	checker := tlc.Globals.MainChecker
	if checker == nil || checker.LiveCheck == nil {
		t.Fatal("main checker/live check absent")
	}
	if err := checker.LiveCheck.FlushWritesToDiskFiles(); err != nil {
		t.Fatal(err)
	}
	if checker.Metadir == "" {
		t.Fatal("metadir absent")
	}
	for _, file := range []struct {
		name string
		size int64
	}{{"nodes_0", nodesSize}, {"ptrs_0", ptrsSize}} {
		info, err := os.Stat(filepath.Join(checker.Metadir, file.name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() != file.size {

			t.Fatalf("%s size=%d, want %d", file.name, info.Size(), file.size)
		}
	}
}

// ModelCheckerTestCase sets VETO_CLEANUP to retain graphs for its assertions.
func retainJavaCodePlexGraphs(t *testing.T) {
	t.Helper()
	live := tlc.Globals.MainChecker.LiveCheck
	t.Cleanup(func() {
		for _, checker := range live.Checkers {
			var err error
			if checker.TableauDiskGraph != nil {
				err = checker.TableauDiskGraph.Close()
			} else if checker.DiskGraph != nil {
				err = checker.DiskGraph.Close()
			}
			if err != nil {
				t.Error(err)
			}
		}
	})
}

// CommonTestCase.assertStuttering checks the first stuttering ordinal.
func requireJavaCodePlexStuttering(t *testing.T, r *tlc.Result, ordinal int) {
	t.Helper()
	records := javaTLCRecords(r, tlc.ECTLCStatePrint3)
	if len(records) == 0 {
		t.Fatal("TLC_STATE_PRINT3 absent")
	}
	if records[0].StateNumber != ordinal {
		t.Fatalf("stuttering=%d, want %d", records[0].StateNumber, ordinal)
	}
}

// Original CodePlexBug08Test.testSpec, constructor and inherited settings.
func TestJavaCodePlexBug08(t *testing.T) { runJavaCodePlexBug08(t) }
func runJavaCodePlexBug08(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug08Test.json")}
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "MC", true, true, true, 1, append(args, extraArgs...)...)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "18", "11", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_14235897021223000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 744, 320)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ b = FALSE\n/\\ x = 2", "/\\ b = TRUE\n/\\ x = 3", "/\\ b = FALSE\n/\\ x = 3", "/\\ b = TRUE\n/\\ x = 4", "/\\ b = FALSE\n/\\ x = 4", "/\\ b = TRUE\n/\\ x = 5"}, true)
	requireJavaCodePlexStuttering(t, r, 7)
	requireJavaTLCUncovered(t, r)
	return r
}

// Original CodePlexBug08Test_TTraceTest.testSpec, rechecking the actual generated artifact.
func TestJavaCodePlexBug08TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "CodePlexBug08TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaCodePlexBug08(t, "-teSpecOutDir", generated) }) {
		return
	}
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTTraceRecheck(t, "CodePlexBug08", generated, false, true)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "6", "6", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 164, 96)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ b = FALSE\n/\\ x = 2", "/\\ b = TRUE\n/\\ x = 3", "/\\ b = FALSE\n/\\ x = 3", "/\\ b = TRUE\n/\\ x = 4", "/\\ b = FALSE\n/\\ x = 4", "/\\ b = TRUE\n/\\ x = 5"}, true)
	requireJavaCodePlexStuttering(t, r, 7)
	requireJavaTLCUncovered(t, r)
}

// Original CodePlexBug08aTest.testSpec, constructor and inherited settings.
func TestJavaCodePlexBug08a(t *testing.T) { runJavaCodePlexBug08a(t) }
func runJavaCodePlexBug08a(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug08aTest.json")}
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "MCa", true, true, true, 1, append(args, extraArgs...)...)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "18", "11", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_14273156623745000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 816, 352)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ b = FALSE\n/\\ x = 1", "/\\ b = TRUE\n/\\ x = 2", "/\\ b = FALSE\n/\\ x = 2", "/\\ b = TRUE\n/\\ x = 3", "/\\ b = FALSE\n/\\ x = 3", "/\\ b = TRUE\n/\\ x = 4", "/\\ b = FALSE\n/\\ x = 4", "/\\ b = TRUE\n/\\ x = 5"}, true)
	actions := []string{"<Init line 15, col 9 to line 15, col 35 of module CodeplexBug8>", "<B line 11, col 6 to line 13, col 18 of module CodeplexBug8>", "<A line 6, col 6 to line 9, col 19 of module CodeplexBug8>", "<B line 11, col 6 to line 13, col 18 of module CodeplexBug8>", "<A line 6, col 6 to line 9, col 19 of module CodeplexBug8>", "<B line 11, col 6 to line 13, col 18 of module CodeplexBug8>", "<A line 6, col 6 to line 9, col 19 of module CodeplexBug8>", "<B line 11, col 6 to line 13, col 18 of module CodeplexBug8>"}
	for i, record := range javaTLCRecords(r, tlc.ECTLCStatePrint2) {
		if record.StateInfo.Info != actions[i] {
			t.Fatalf("state%d action=%v, want %s", i+1, record.StateInfo.Info, actions[i])
		}
	}
	requireJavaCodePlexStuttering(t, r, 9)
	requireJavaTLCUncovered(t, r)
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 18 {
		t.Fatalf("register42=%v, want IntValue18", values[0])
	}
	return r
}

// Original CodePlexBug08aTest_TTraceTest.testSpec, rechecking the actual generated artifact.
func TestJavaCodePlexBug08aTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "CodePlexBug08aTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaCodePlexBug08a(t, "-teSpecOutDir", generated) }) {
		return
	}
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTTraceRecheck(t, "CodePlexBug08", generated, false, true)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "8", "8", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 220, 128)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ b = FALSE\n/\\ x = 1", "/\\ b = TRUE\n/\\ x = 2", "/\\ b = FALSE\n/\\ x = 2", "/\\ b = TRUE\n/\\ x = 3", "/\\ b = FALSE\n/\\ x = 3", "/\\ b = TRUE\n/\\ x = 4", "/\\ b = FALSE\n/\\ x = 4", "/\\ b = TRUE\n/\\ x = 5"}, true)
	requireJavaCodePlexStuttering(t, r, 9)
	requireJavaTLCUncovered(t, r)
}

// Original CodePlexBug08AgentRingTest.testSpec, constructor and inherited settings.
func TestJavaCodePlexBug08AgentRing(t *testing.T) { runJavaCodePlexBug08AgentRing(t) }
func runJavaCodePlexBug08AgentRing(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug08AgentRingTest.json")}
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "AgentRingMC", true, true, true, 1, append(args, extraArgs...)...)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "361", "120")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_142745588773041000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 8496, 2880)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ Agent = [Loc |-> 0, LastLoad |-> 0, ReadyToMove |-> TRUE, Task |-> 0]\n/\\ CanCreate = TRUE\n/\\ Nodes = (0 :> [Load |-> 0] @@ 1 :> [Load |-> 0])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 0, ReadyToMove |-> TRUE, Task |-> 0]\n/\\ CanCreate = TRUE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 0])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 0, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = TRUE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 0])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 0, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = TRUE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 0, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 1, ReadyToMove |-> TRUE, Task |-> 1]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 1])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 1, ReadyToMove |-> FALSE, Task |-> 1]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 1])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 2, ReadyToMove |-> TRUE, Task |-> 1]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 1])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 2, ReadyToMove |-> FALSE, Task |-> 1]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 1])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 2, ReadyToMove |-> TRUE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 2, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 2, ReadyToMove |-> TRUE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 2, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])"}, true)
	actions := []string{"<Init line 50, col 3 to line 53, col 12 of module AgentRing>", "<CreateTasks line 82, col 3 to line 84, col 35 of module AgentRing>", "<Move line 58, col 3 to line 60, col 35 of module AgentRing>", "<CreateTasks line 82, col 3 to line 84, col 35 of module AgentRing>", "<Stop line 78, col 3 to line 79, col 31 of module AgentRing>", "<LookAndAct line 63, col 3 to line 75, col 24 of module AgentRing>", "<Move line 58, col 3 to line 60, col 35 of module AgentRing>", "<LookAndAct line 63, col 3 to line 75, col 24 of module AgentRing>", "<Move line 58, col 3 to line 60, col 35 of module AgentRing>", "<LookAndAct line 63, col 3 to line 75, col 24 of module AgentRing>", "<Move line 58, col 3 to line 60, col 35 of module AgentRing>", "<LookAndAct line 63, col 3 to line 75, col 24 of module AgentRing>", "<Move line 58, col 3 to line 60, col 35 of module AgentRing>"}
	for i, record := range javaTLCRecords(r, tlc.ECTLCStatePrint2) {
		if record.StateInfo.Info != actions[i] {
			t.Fatalf("state%d action=%v, want %s", i+1, record.StateInfo.Info, actions[i])
		}
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "10")
	requireJavaTLCUncovered(t, r)
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionFalse)) != 0 {
		t.Fatal("TLC_POSTCONDITION_FALSE present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCPostconditionEvaluationError)) != 0 {
		t.Fatal("TLC_POSTCONDITION_EVALUATION_ERROR present")
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 361 {
		t.Fatalf("register42=%v, want IntValue361", values[0])
	}
	return r
}

// Original CodePlexBug08AgentRingTest_TTraceTest.testSpec, rechecking the actual generated artifact.
func TestJavaCodePlexBug08AgentRingTTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "CodePlexBug08AgentRingTestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaCodePlexBug08AgentRing(t, "-teSpecOutDir", generated) }) {
		return
	}
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTTraceRecheck(t, "CodePlexBug08", generated, false, true)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "14", "13")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 380, 208)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ Agent = [Loc |-> 0, LastLoad |-> 0, ReadyToMove |-> TRUE, Task |-> 0]\n/\\ CanCreate = TRUE\n/\\ Nodes = (0 :> [Load |-> 0] @@ 1 :> [Load |-> 0])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 0, ReadyToMove |-> TRUE, Task |-> 0]\n/\\ CanCreate = TRUE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 0])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 0, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = TRUE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 0])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 0, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = TRUE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 0, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 1, ReadyToMove |-> TRUE, Task |-> 1]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 1])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 1, ReadyToMove |-> FALSE, Task |-> 1]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 1])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 2, ReadyToMove |-> TRUE, Task |-> 1]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 1])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 2, ReadyToMove |-> FALSE, Task |-> 1]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 1])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 2, ReadyToMove |-> TRUE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 2, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 0, LastLoad |-> 2, ReadyToMove |-> TRUE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])", "/\\ Agent = [Loc |-> 1, LastLoad |-> 2, ReadyToMove |-> FALSE, Task |-> 0]\n/\\ CanCreate = FALSE\n/\\ Nodes = (0 :> [Load |-> 2] @@ 1 :> [Load |-> 2])"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "10")
	requireJavaTLCUncovered(t, r)
}

// Original CodePlexBug08EWD840FL1Test.testSpec, constructor and inherited settings.
func TestJavaCodePlexBug08EWD840FL1(t *testing.T) { runJavaCodePlexBug08EWD840FL1(t) }
func runJavaCodePlexBug08EWD840FL1(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug08EWD840FL1Test.json")}
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "EWD840MC1", true, true, true, 1, append(args, extraArgs...)...)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "15986", "1566", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "FalseLiveness")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 7560068, 279616)
	requireJavaTLCUncovered(t, r)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 0\n/\\ tcolor = \"black\"", "/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> TRUE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> TRUE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> TRUE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> FALSE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> FALSE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 2\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> FALSE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 1\n/\\ tcolor = \"black\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 1\n/\\ tcolor = \"black\"", "/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 1\n/\\ tcolor = \"black\""}, true)
	actions := []string{"<Init line 21, col 3 to line 24, col 21 of module EWD840>", "<InitiateProbe line 30, col 3 to line 35, col 43 of module EWD840>", "<SendMsg(3) line 61, col 3 to line 65, col 31 of module EWD840>", "<SendMsg(2) line 61, col 3 to line 65, col 31 of module EWD840>", "<SendMsg(1) line 61, col 3 to line 65, col 31 of module EWD840>", "<SendMsg(1) line 61, col 3 to line 65, col 31 of module EWD840>", "<SendMsg(2) line 61, col 3 to line 65, col 31 of module EWD840>", "<Deactivate(0) line 69, col 3 to line 71, col 38 of module EWD840>", "<Deactivate(1) line 69, col 3 to line 71, col 38 of module EWD840>", "<SendMsg(2) line 61, col 3 to line 65, col 31 of module EWD840>", "<Deactivate(2) line 69, col 3 to line 71, col 38 of module EWD840>", "<Deactivate(3) line 69, col 3 to line 71, col 38 of module EWD840>", "<PassToken(3) line 47, col 3 to line 52, col 43 of module EWD840>", "<PassToken(2) line 47, col 3 to line 52, col 43 of module EWD840>", "<SendMsg(1) line 61, col 3 to line 65, col 31 of module EWD840>", "<Deactivate(1) line 69, col 3 to line 71, col 38 of module EWD840>"}
	for i, record := range javaTLCRecords(r, tlc.ECTLCStatePrint2) {
		if record.StateInfo.Info != actions[i] {
			t.Fatalf("state%d action=%v, want %s", i+1, record.StateInfo.Info, actions[i])
		}
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1", "<PassToken(1) line 47, col 3 to line 52, col 43 of module EWD840>")
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register42 empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 15986 {
		t.Fatalf("register42=%v, want IntValue15986", values[0])
	}
	return r
}

// Original CodePlexBug08EWD840FL1Test_TTraceTest.testSpec, rechecking the actual generated artifact.
func TestJavaCodePlexBug08EWD840FL1TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "CodePlexBug08EWD840FL1TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaCodePlexBug08EWD840FL1(t, "-teSpecOutDir", generated) }) {
		return
	}
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTTraceRecheck(t, "CodePlexBug08", generated, false, true)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "17", "16", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 464, 256)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 0\n/\\ tcolor = \"black\"", "/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> TRUE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> TRUE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> TRUE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> FALSE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 3\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> FALSE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"black\" @@ 3 :> \"white\")\n/\\ tpos = 2\n/\\ tcolor = \"white\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> FALSE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 1\n/\\ tcolor = \"black\"", "/\\ active = (0 :> FALSE @@ 1 :> TRUE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 1\n/\\ tcolor = \"black\"", "/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ color = (0 :> \"white\" @@ 1 :> \"black\" @@ 2 :> \"white\" @@ 3 :> \"white\")\n/\\ tpos = 1\n/\\ tcolor = \"black\""}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
	requireJavaTLCUncovered(t, r)
}

// Original CodePlexBug08EWD840FL2Test.testSpec, constructor and inherited settings.
func TestJavaCodePlexBug08EWD840FL2(t *testing.T) { runJavaCodePlexBug08EWD840FL2(t) }
func runJavaCodePlexBug08EWD840FL2(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug08EWD840FL2Test.json")}
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "EWD840MC2", true, true, true, 1, append(args, extraArgs...)...)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "15986", "1566", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "FalseLiveness2")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 54037212, 831296)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ tpos = 0\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> TRUE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> TRUE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"black\" @@ 3 :> \"white\")", "/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")"}, true)
	actions := []string{"<Init line 21, col 3 to line 24, col 21 of module EWD840>", "<InitiateProbe line 30, col 3 to line 35, col 43 of module EWD840>", "<SendMsg(3) line 61, col 3 to line 65, col 31 of module EWD840>", "<SendMsg(2) line 61, col 3 to line 65, col 31 of module EWD840>", "<Deactivate(3) line 69, col 3 to line 71, col 38 of module EWD840>", "<Deactivate(0) line 69, col 3 to line 71, col 38 of module EWD840>", "<PassToken(3) line 47, col 3 to line 52, col 43 of module EWD840>", "<SendMsg(2) line 61, col 3 to line 65, col 31 of module EWD840>", "<PassToken(2) line 47, col 3 to line 52, col 43 of module EWD840>", "<Deactivate(2) line 69, col 3 to line 71, col 38 of module EWD840>"}
	for i, record := range javaTLCRecords(r, tlc.ECTLCStatePrint2) {
		if record.StateInfo.Info != actions[i] {
			t.Fatalf("state%d action=%v, want %s", i+1, record.StateInfo.Info, actions[i])
		}
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1", "<PassToken(1) line 47, col 3 to line 52, col 43 of module EWD840>")
	return r
}

// Original CodePlexBug08EWD840FL2Test_TTraceTest.testSpec, rechecking the actual generated artifact.
func TestJavaCodePlexBug08EWD840FL2TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "CodePlexBug08EWD840FL2TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaCodePlexBug08EWD840FL2(t, "-teSpecOutDir", generated) }) {
		return
	}
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTTraceRecheck(t, "CodePlexBug08", generated, false, true)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "11", "10", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 296, 160)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ tpos = 0\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> TRUE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> TRUE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"black\" @@ 3 :> \"white\")", "/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
}

// Original CodePlexBug08EWD840FL3Test.testSpec, constructor and inherited settings.
func TestJavaCodePlexBug08EWD840FL3(t *testing.T) { runJavaCodePlexBug08EWD840FL3(t) }
func runJavaCodePlexBug08EWD840FL3(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug08EWD840FL3Test.json")}
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "EWD840MC3", true, true, true, 1, append(args, extraArgs...)...)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "15986", "1566", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "FalseLiveness3")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 143808, 25040)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ tpos = 0\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"black\" @@ 3 :> \"white\")", "/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
	requireJavaTLCUncovered(t, r)
	return r
}

// Original CodePlexBug08EWD840FL3Test_TTraceTest.testSpec, rechecking the actual generated artifact.
func TestJavaCodePlexBug08EWD840FL3TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "CodePlexBug08EWD840FL3TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaCodePlexBug08EWD840FL3(t, "-teSpecOutDir", generated) }) {
		return
	}
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTTraceRecheck(t, "CodePlexBug08", generated, false, true)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "9", "8", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 240, 128)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ tpos = 0\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> FALSE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 2\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"black\" @@ 3 :> \"white\")", "/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> TRUE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 1\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")"}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
	requireJavaTLCUncovered(t, r)
}

// Original CodePlexBug08EWD840FL4Test.testSpec, constructor and inherited settings.
func TestJavaCodePlexBug08EWD840FL4(t *testing.T) { runJavaCodePlexBug08EWD840FL4(t) }
func runJavaCodePlexBug08EWD840FL4(t *testing.T, extraArgs ...string) *tlc.Result {
	t.Helper()
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	args := []string{"-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug08EWD840FL4Test.json")}
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "EWD840MC4", true, true, true, 1, append(args, extraArgs...)...)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "15986", "1566", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "AllNodesTerminateIfNoMessages")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 135540, 23456)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ tpos = 0\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")"}, true)
	requireJavaCodePlexStuttering(t, r, 3)
	requireJavaTLCUncovered(t, r)
	return r
}

// Original CodePlexBug08EWD840FL4Test_TTraceTest.testSpec, rechecking the actual generated artifact.
func TestJavaCodePlexBug08EWD840FL4TTrace(t *testing.T) {
	generated := filepath.Join(t.TempDir(), "CodePlexBug08EWD840FL4TestTTrace.tla")
	if !t.Run("original", func(t *testing.T) { runJavaCodePlexBug08EWD840FL4(t, "-teSpecOutDir", generated) }) {
		return
	}
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTTraceRecheck(t, "CodePlexBug08", generated, false, true)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "2", "2", "0")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 52, 32)
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{"/\\ tpos = 0\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"black\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")", "/\\ tpos = 3\n/\\ active = (0 :> FALSE @@ 1 :> FALSE @@ 2 :> FALSE @@ 3 :> TRUE)\n/\\ tcolor = \"white\"\n/\\ color = (0 :> \"white\" @@ 1 :> \"white\" @@ 2 :> \"white\" @@ 3 :> \"white\")"}, true)
	requireJavaCodePlexStuttering(t, r, 3)
	requireJavaTLCUncovered(t, r)
}

// Original CodePlexBug08AgentRing790Test.testSpec and coverage=false constructor settings.
func TestJavaCodePlexBug08AgentRing790(t *testing.T) {
	r := runJavaTLCModelTestWithRoot(t, "CodePlexBug08", "AgentRingMC", false, true, true, 1, "-config", "AgentRing790MC.cfg", "-dumpTrace", "json", filepath.Join(t.TempDir(), "CodePlexBug08AgentRing790Test.json"))
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}
