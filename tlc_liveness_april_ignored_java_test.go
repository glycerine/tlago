/*******************************************************************************
 * Copyright (c) 2016 Microsoft Research. All rights reserved.
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
 *
 * The MIT License (MIT)
 *******************************************************************************/
package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

const ignoredLivenessSymmetryReason = "upstream @Ignore: Ignored for as long as symmetry is incorrectly handled by TLC with liveness checking."

func runIgnoredAprilSymmetry(t *testing.T, root, property, backAction string, expected []string, stats bool) {
	t.Helper()
	r := runJavaTLCModelTestWithRoot(t, "symmetry", root, true, true, true, 1)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	if stats {
		requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "3", "0")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, property)
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaRandomSubsetTrace(t, r, expected, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1", backAction)
}

// Original April20aTest.testSpec and constructor. Upstream ignores incorrect
// symmetry handling during liveness checking.
func TestJavaApril20AIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	runIgnoredAprilSymmetry(t, "April20aMC", "prop_146115895645522000",
		"<Action line 17, col 12 to line 19, col 20 of module April20a>", []string{
			"/\\ x = 0\n/\\ y = 0",
			"/\\ x = m2\n/\\ y = 1",
			"/\\ x = m1\n/\\ y = 2",
			"/\\ x = 0\n/\\ y = 0",
			"/\\ x = m1\n/\\ y = 1",
			"/\\ x = m2\n/\\ y = 2",
		}, true)
}

// Original April20bTest.testSpec and constructor, with the same upstream ignore.
func TestJavaApril20BIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	runIgnoredAprilSymmetry(t, "April20bMC", "prop_146115906602341000",
		"<Action line 17, col 12 to line 19, col 24 of module April20b>", []string{
			"/\\ x = <<>>\n/\\ y = 0",
			"/\\ x = <<m2>>\n/\\ y = 1",
			"/\\ x = <<m2, m1>>\n/\\ y = 2",
			"/\\ x = <<>>\n/\\ y = 0",
			"/\\ x = <<m1>>\n/\\ y = 1",
			"/\\ x = <<m1, m2>>\n/\\ y = 2",
		}, true)
}

// Original April21Test.testSpec and constructor, with the same upstream ignore.
func TestJavaApril21Ignored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	runIgnoredAprilSymmetry(t, "April21MC", "prop_146115895645522000",
		"<Action line 10, col 9 to line 10, col 16 of module April21>",
		[]string{"x = m1", "x = m2"}, false)
}

// Original April22Test.testSpec and constructor. This case expects symmetry
// reduction to finish without a temporal violation.
func TestJavaApril22Ignored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	r := runJavaTLCModelTestWithRoot(t, "symmetry", "April22MC", true, true, true, 1)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	if records := javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated); len(records) != 0 {
		t.Fatalf("temporal violation recorded: %v", records)
	}
	if records := javaTLCRecords(r, tlc.ECTLCCounterExample); len(records) != 0 {
		t.Fatalf("counterexample recorded: %v", records)
	}
}

// Original April25Test.testSpec and constructor, with the shared upstream
// incorrect-symmetry ignore retained.
func TestJavaApril25Ignored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	runIgnoredAprilSymmetry(t, "April25MC", "prop_146115895645522000",
		"<Action line 10, col 9 to line 10, col 16 of module April25>",
		[]string{"x = m1", "x = m2"}, false)
}

func requireIgnoredApril29Graph(t *testing.T, successorTransitionCount int, checks ...[]int) {
	t.Helper()
	checker, ok := tlc.Globals.MainChecker.(*tlc.ModelChecker)
	if !ok || checker == nil || checker.LiveCheck == nil {
		t.Fatal("main checker/live check absent")
	}
	liveCheck := checker.LiveCheck
	if got := liveCheck.NumChecker(); got != 1 {
		t.Fatalf("live checker count=%d, want 1", got)
	}
	liveChecker := liveCheck.GetChecker(0)
	graph, ok := liveChecker.GetDiskGraph().(*tlc.DiskGraph)
	if !ok || graph == nil {
		t.Fatal("ordinary disk graph absent")
	}
	if err := graph.MakeNodePtrTbl(); err != nil {
		t.Fatal(err)
	}
	if got := graph.Size(); got != 2 {
		t.Fatalf("graph size=%d, want 2", got)
	}
	initNodes := graph.GetInitNodes()
	if got := initNodes.Size(); got != 2 {
		t.Fatalf("init-node vector size=%d, want 2", got)
	}
	solution := liveChecker.GetSolution()
	if solution == nil {
		t.Fatal("liveness solution absent")
	}
	slen := len(solution.GetCheckState())
	if slen != 0 {
		t.Fatalf("state checks=%d, want 0", slen)
	}
	alen := len(solution.GetCheckAction())
	if alen != 3 {
		t.Fatalf("action checks=%d, want 3", alen)
	}
	init, err := graph.GetNodeNoTableau(uint64(initNodes.ElementAt(0)))
	if err != nil {
		t.Fatal(err)
	}
	transitions := init.GetTransitions(slen, alen)
	if got := len(transitions); got != 2 {
		t.Fatalf("initial transitions=%d, want 2", got)
	}
	outs := make([]tlc.GraphTransition, 0, 1)
	for transition := range transitions {
		if transition.GetFP() != init.StateFP {
			outs = append(outs, transition)
		}
	}
	if got := len(outs); got != 1 {
		t.Fatalf("non-self initial transitions=%d, want 1", got)
	}
	successor, err := graph.GetNodeNoTableau(outs[0].GetFP())
	if err != nil {
		t.Fatal(err)
	}
	transitions = successor.GetTransitions(slen, alen)
	if got := len(transitions); got != successorTransitionCount {
		t.Fatalf("successor transitions=%d, want %d", got, successorTransitionCount)
	}
	for _, setBits := range checks {
		vector := tlc.NewBitVector(alen + slen)
		for _, bit := range setBits {
			vector.Set(bit)
		}
		transitions[tlc.NewGraphTransition(successor.StateFP, -1, vector)] = struct{}{}
	}
	if got := len(transitions); got != successorTransitionCount {
		t.Fatalf("successor transition set after equal additions=%d, want %d", got, successorTransitionCount)
	}
}

// Original April29Test.testSpec and constructor. The graph assertions preserve
// the source test's checks of ordinary liveness nodes and action predicates.
func TestJavaApril29Ignored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "symmetry", "April29MC", true, true, true, 1)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_146115895645522000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ x = a\n/\\ y = 0",
		"/\\ x = a\n/\\ y = 1",
	}, true)
	requireJavaCodePlexStuttering(t, r, 3)
	requireIgnoredApril29Graph(t, 1, []int{0, 2})
}

// Original April29dTest.testSpec and constructor, with the same upstream ignore.
func TestJavaApril29DIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "symmetry", "April29dMC", true, true, true, 1)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	requireIgnoredApril29Graph(t, 2, []int{0, 2}, []int{2})
	if records := javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated); len(records) != 0 {
		t.Fatalf("temporal violation recorded: %v", records)
	}
	if records := javaTLCRecords(r, tlc.ECTLCCounterExample); len(records) != 0 {
		t.Fatalf("counterexample recorded: %v", records)
	}
}

// Original May09Test.testSpec and constructor, with the shared upstream ignore.
func TestJavaMay09Ignored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	r := runJavaTLCModelTestWithRoot(t, "symmetry", "May09MC", true, true, true, 1)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_146183936178122000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ x = a\n/\\ y = 0",
		"/\\ x = a\n/\\ y = 1",
		"/\\ x = a\n/\\ y = 2",
		"/\\ x = a\n/\\ y = 3",
		"/\\ x = a\n/\\ y = 4",
		"/\\ x = a\n/\\ y = 5",
	}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "2",
		"<Action line 19, col 10 to line 21, col 18 of module May09>")
}

// Original May09dTest.testSpec and constructor, with the shared upstream ignore.
func TestJavaMay09DIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	r := runJavaTLCModelTestWithRoot(t, "symmetry", "May09dMC", true, true, true, 1)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if records := javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated); len(records) != 0 {
		t.Fatalf("temporal violation recorded: %v", records)
	}
	if records := javaTLCRecords(r, tlc.ECTLCCounterExample); len(records) != 0 {
		t.Fatalf("counterexample recorded: %v", records)
	}
}

// Original NQTest.testSpec and constructor, with the shared upstream ignore.
func TestJavaNQIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	r := runJavaTLCModelTestWithRoot(t, "symmetry/NQ", "MC", true, true, true, 1)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1049", "363", "0")
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	for _, code := range []int{
		tlc.ECTLCTemporalPropertyViolated,
		tlc.ECTLCCounterExample,
		tlc.ECTLCStatePrint2,
	} {
		if records := javaTLCRecords(r, code); len(records) != 0 {
			t.Fatalf("message code %d recorded: %v", code, records)
		}
	}
}

// Original NQaTest.testSpec and constructor. The final fatal assertion is part
// of the ignored source body, which requires a future trace-completeness check.
func TestJavaNQAIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	r := runJavaTLCModelTestWithRoot(t, "symmetry/NQ", "MCa", true, true, true, 1)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "1049", "363", "0")
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_14467710210701039000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	if len(javaTLCRecords(r, tlc.ECTLCStatePrint2)) == 0 {
		t.Fatal("TLC_STATE_PRINT2 absent")
	}
	t.Fatal("Check actual error trace and its completeness")
}

// Original OneBitMutexTest.testSpec and constructor, with the shared upstream
// symmetry/liveness ignore retained.
func TestJavaOneBitMutexIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "symmetry/OneBitMutex", "OneBitMutexMC", true, true, true, 1)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "133", "68", "0")
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_144491423293819000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 7380, 2368)
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> FALSE @@ B :> FALSE)\n" +
			"/\\ pc = (A :> \"ncs\" @@ B :> \"ncs\")",
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> FALSE @@ B :> FALSE)\n" +
			"/\\ pc = (A :> \"e1\" @@ B :> \"ncs\")",
		"/\\ unchecked = (A :> {B} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> TRUE @@ B :> FALSE)\n" +
			"/\\ pc = (A :> \"e2\" @@ B :> \"ncs\")",
		"/\\ unchecked = (A :> {B} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> TRUE @@ B :> FALSE)\n" +
			"/\\ pc = (A :> \"e2\" @@ B :> \"e1\")",
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> TRUE @@ B :> FALSE)\n" +
			"/\\ pc = (A :> \"e3\" @@ B :> \"e1\")",
		"/\\ unchecked = (A :> {} @@ B :> {A})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> TRUE @@ B :> TRUE)\n" +
			"/\\ pc = (A :> \"e3\" @@ B :> \"e2\")",
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> TRUE @@ B :> TRUE)\n" +
			"/\\ pc = (A :> \"e3\" @@ B :> \"e3\")",
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> TRUE @@ B :> TRUE)\n" +
			"/\\ pc = (A :> \"e4\" @@ B :> \"e3\")",
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> TRUE @@ B :> TRUE)\n" +
			"/\\ pc = (A :> \"e4\" @@ B :> \"e4\")",
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> FALSE @@ B :> TRUE)\n" +
			"/\\ pc = (A :> \"e5\" @@ B :> \"e4\")",
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> FALSE @@ B :> FALSE)\n" +
			"/\\ pc = (A :> \"e5\" @@ B :> \"e5\")",
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> FALSE @@ B :> FALSE)\n" +
			"/\\ pc = (A :> \"e1\" @@ B :> \"e5\")",
		"/\\ unchecked = (A :> {} @@ B :> {})\n" +
			"/\\ other = (A :> B @@ B :> A)\n" +
			"/\\ x = (A :> FALSE @@ B :> FALSE)\n" +
			"/\\ pc = (A :> \"e1\" @@ B :> \"e1\")",
	}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "4",
		"<Action line 60, col 13 to line 64, col 29 of module OneBitMutex>")
}

// Original ChooseTableauSymmetryTest.testSpec and constructor, with the shared
// upstream symmetry/liveness ignore retained.
func TestJavaChooseTableauSymmetryIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	r := runJavaTLCModelTestWithRoot(t, "symmetry", "ChooseTableauSymmetryMC", true, true, true, 1)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "13", "6", "0")
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_143263724208534000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"arr = (a :> \"ready\" @@ b :> \"ready\")",
		"arr = (a :> \"busy\" @@ b :> \"ready\")",
		"arr = (a :> \"done\" @@ b :> \"ready\")",
		"arr = (a :> \"done\" @@ b :> \"busy\")",
		"arr = (a :> \"done\" @@ b :> \"done\")",
		"arr = (a :> \"done\" @@ b :> \"ready\")",
	}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "4",
		"<Action line 7, col 13 to line 8, col 47 of module ChooseTableauSymmetry>")
}

func runIgnoredSymmetryModelCheckerTrace(t *testing.T, root string, stats []string, nodeSize, ptrSize int64, trace []string, backAction string) {
	t.Helper()
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "symmetry", root, true, true, true, 1)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, stats...)
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_143263724208534000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, nodeSize, ptrSize)
	requireJavaRandomSubsetTrace(t, r, trace, true)
	if backAction == "" {
		requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1")
	} else {
		requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1", backAction)
	}
}

var ignoredSymmetryModelCheckerShortTrace = []string{
	"/\\ x = a\n/\\ y = 0",
	"/\\ x = a\n/\\ y = 1",
	"/\\ x = b\n/\\ y = 0",
	"/\\ x = b\n/\\ y = 1",
}

var ignoredSymmetryModelCheckerLongTrace = []string{
	"/\\ x = a\n/\\ y = 0",
	"/\\ x = a\n/\\ y = 1",
	"/\\ x = a\n/\\ y = 2",
	"/\\ x = a\n/\\ y = 3",
	"/\\ x = a\n/\\ y = 4",
	"/\\ x = b\n/\\ y = 0",
	"/\\ x = b\n/\\ y = 1",
	"/\\ x = b\n/\\ y = 2",
	"/\\ x = b\n/\\ y = 3",
	"/\\ x = b\n/\\ y = 4",
}

// Original SymmetryModelCheckerTest3.testSpec and constructor.
func TestJavaSymmetryModelChecker3Ignored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	runIgnoredSymmetryModelCheckerTrace(t, "MC", []string{"5", "2", "0"}, 84, 32,
		ignoredSymmetryModelCheckerShortTrace,
		"<Action line 27, col 12 to line 29, col 27 of module SymmetryLiveness3>")
}

// Original SymmetryModelCheckerTest3a.testSpec and constructor.
func TestJavaSymmetryModelChecker3AIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	runIgnoredSymmetryModelCheckerTrace(t, "MCa", []string{"4", "2", "0"}, 72, 32,
		ignoredSymmetryModelCheckerShortTrace, "")
}

// Original SymmetryModelCheckerTestLong.testSpec and constructor.
func TestJavaSymmetryModelCheckerLongIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	runIgnoredSymmetryModelCheckerTrace(t, "LongMC", []string{"8", "5", "0"}, 192, 80,
		ignoredSymmetryModelCheckerLongTrace,
		"<Action line 47, col 12 to line 49, col 27 of module SymmetryLivenessLong>")
}

// Original SymmetryModelCheckerTestLonga.testSpec and constructor.
func TestJavaSymmetryModelCheckerLongAIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	runIgnoredSymmetryModelCheckerTrace(t, "LongMCa", []string{"7", "5", "0"}, 180, 80,
		ignoredSymmetryModelCheckerLongTrace,
		"<Action line 25, col 12 to line 27, col 27 of module SymmetryLivenessLong>")
}

// Original SymmetryTableauModelCheckerTest.testSpec and constructor.
func TestJavaSymmetryTableauModelCheckerIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	r := runJavaTLCModelTestWithRoot(t, "symmetry", "SymmetryLivenessTableauMC", true, true, true, 1)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCLiveImplied, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInitGenerated2, "8", "s", "2")
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "779", "168", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_144172195414092000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECGeneral,
		"TLC threw an unexpected exception.\n"+
			"This was probably caused by an error in the spec or model.\n"+
			"The error occurred when TLC was checking liveness.\n"+
			"The exception was a tlc2.tool.EvalException\n"+
			": Failed to recover the next state from its fingerprint during\n"+
			"liveness error trace re-construction. This indicates that the\n"+
			"spec is in fact not symmetric (Please report a TLC bug if the\n"+
			"spec is known to be symmetric).")
}

// Original TableauSymmetryTest.testSpec and constructor.
func TestJavaTableauSymmetryIgnored(t *testing.T) {
	t.Skip(ignoredLivenessSymmetryReason)
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "symmetry", "TableauSymmetryMC", true, true, true, 1)
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want inherited SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "13", "6", "0")
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "prop_144481269489929000")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaNodeAndPtrSizes(t, 624, 224)
	requireJavaRandomSubsetTrace(t, r, []string{
		"arr = (a :> \"ready\" @@ b :> \"ready\")",
		"arr = (a :> \"busy\" @@ b :> \"ready\")",
		"arr = (a :> \"done\" @@ b :> \"ready\")",
		"arr = (a :> \"done\" @@ b :> \"busy\")",
		"arr = (a :> \"done\" @@ b :> \"done\")",
		"arr = (a :> \"done\" @@ b :> \"ready\")",
	}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "4",
		"<Action line 7, col 13 to line 8, col 47 of module TableauSymmetry>")
}
