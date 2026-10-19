package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"path/filepath"
	"testing"
)

// Original Github604Test.testSpec and inherited settings/exit.
func TestJavaLivenessGithub604(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github604", "Github604", false, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github604Test.json"), "-config", "Github604.tla")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "17", "4", "0")
}

// Original Github702Test.testSpec and inherited settings/exit.
// Copyright (c) 2022, Oracle and/or its affiliates.
func TestJavaLivenessGithub702(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "Github702", "Github702", false, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "Github702Test.json"), "-config", "Github702.cfg")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original IncompatibleTypesLiveTest.testSpec and inherited settings/exit.
func TestJavaLivenessIncompatibleTypesLive(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "IncompatibleTypesLive", "IncompatibleTypesLive", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "IncompatibleTypesLiveTest.json"), "-config", "IncompatibleTypesLive.tla")
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "10507", "103", "0")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
}

// Original TwoPhaseCommitTest.testSpec and inherited settings/exit.
func TestJavaLivenessTwoPhaseCommit(t *testing.T) {
	skipSlowTest(t)

	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "TwoPhaseCommit", "MC", true, true, true, 1, "-dumpTrace", "json", filepath.Join(t.TempDir(), "TwoPhaseCommitTest.json"))
	retainJavaCodePlexGraphs(t)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCUncovered(t, r)
}
