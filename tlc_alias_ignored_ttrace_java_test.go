package tlago

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func runIgnoredAliasTTrace(t *testing.T, config string, simulation, sourceDebugger bool, expectedExit int) *tlc.Result {
	t.Helper()
	generated := filepath.Join(t.TempDir(), "AliasIgnoredTTrace.tla")
	args := []string{"-config", config, "-teSpecOutDir", generated}
	if simulation {
		args = append(args, "-simulate", "num=1")
	}
	first := runJavaTLCModelTestWithDebugger(t, "Alias", true, true, sourceDebugger, 1, args...)
	if first.ExitStatus != expectedExit {
		t.Fatalf("source exit=%d, want %d", first.ExitStatus, expectedExit)
	}
	return runJavaTTraceRecheck(t, "Alias", generated, false, true)
}

func requireIgnoredAliasTTraceBase(t *testing.T, r *tlc.Result, expectedExit int) {
	t.Helper()
	if r.ExitStatus != expectedExit {
		t.Fatalf("exit=%d, want %d", r.ExitStatus, expectedExit)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
}

// Original AliasLivenessLassoTest_TTraceTest.testSpec. Upstream marks this
// method @Ignore("TESpec Bug"), which remains the first operation here.
func TestJavaAliasLivenessLassoTTraceIgnored(t *testing.T) {
	t.Skip("upstream @Ignore: TESpec Bug")
	r := runIgnoredAliasTTrace(t, "AliasLasso.cfg", false, true, tlc.ExitStatusViolationLiveness)
	requireIgnoredAliasTTraceBase(t, r, tlc.ExitStatusViolationLiveness)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "4")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 || len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("temporal violation or counterexample absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ y = FALSE\n/\\ x = 1\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 1 e2: FALSE\"\n/\\ te = TRUE",
		"/\\ y = TRUE\n/\\ x = 2\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 2 e2: TRUE\"\n/\\ te = TRUE",
		"/\\ y = FALSE\n/\\ x = 3\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 3 e2: FALSE\"\n/\\ te = TRUE",
		"/\\ y = TRUE\n/\\ x = 4\n/\\ a = -3\n/\\ b = FALSE\n/\\ anim = \"e1: 4 e2: TRUE\"\n/\\ te = TRUE",
	}, true)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCBackToState, "1", "<B line 20, col 1 to line 22, col 9 of module Alias>")
}

// Original AliasLivenessStutteringTest_TTraceTest.testSpec, retaining its
// upstream @Ignore("TESpec Bug").
func TestJavaAliasLivenessStutteringTTraceIgnored(t *testing.T) {
	t.Skip("upstream @Ignore: TESpec Bug")
	r := runIgnoredAliasTTrace(t, "AliasStuttering.cfg", false, true, tlc.ExitStatusViolationLiveness)
	requireIgnoredAliasTTraceBase(t, r, tlc.ExitStatusViolationLiveness)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "5", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "4")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) == 0 || len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("temporal violation or counterexample absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ y = FALSE\n/\\ x = 1\n/\\ a = 0\n/\\ b = TRUE\n/\\ anim = \"e1: 1 e2: FALSE\"\n/\\ te = TRUE",
	}, true)
	stuttering := javaTLCRecords(r, tlc.ECTLCStatePrint3)
	if len(stuttering) == 0 || stuttering[0].StateNumber != 2 {
		t.Fatalf("stuttering=%v, want ordinal 2", stuttering)
	}
}

// Original AliasSafetyTest_TTraceTest.testSpec, retaining its upstream
// @Ignore("TESpec Bug").
func TestJavaAliasSafetyTTraceIgnored(t *testing.T) {
	t.Skip("upstream @Ignore: TESpec Bug")
	r := runIgnoredAliasTTrace(t, "Alias.tla", false, false, tlc.ExitStatusViolationSafety)
	requireIgnoredAliasTTraceBase(t, r, tlc.ExitStatusViolationSafety)
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "4", "4", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "4")
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ y = FALSE\n/\\ x = 1\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 1 e2: FALSE\"\n/\\ te = TRUE",
		"/\\ y = TRUE\n/\\ x = 2\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 2 e2: TRUE\"\n/\\ te = TRUE",
		"/\\ y = FALSE\n/\\ x = 3\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 3 e2: FALSE\"\n/\\ te = TRUE",
		"/\\ y = TRUE\n/\\ x = 4\n/\\ a = 0\n/\\ b = TRUE\n/\\ anim = \"e1: 4 e2: TRUE\"\n/\\ te = TRUE",
	}, true)
}

// Original AliasSafetySimuTest_TTraceTest.testSpec, retaining its upstream
// @Ignore("TESpec Bug").
func TestJavaAliasSafetySimuTTraceIgnored(t *testing.T) {
	t.Skip("upstream @Ignore: TESpec Bug")
	r := runIgnoredAliasTTrace(t, "Alias.tla", true, true, tlc.ExitStatusViolationSafety)
	requireIgnoredAliasTTraceBase(t, r, tlc.ExitStatusViolationSafety)
	expected := []string{
		"/\\ y = FALSE\n/\\ x = 1\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 1 e2: FALSE\"\n/\\ te = TRUE\n/\\ TLCGetAction = [ name |-> \"UnnamedAction\",\n  location |->\n      [ beginLine |-> 26,\n        beginColumn |-> 18,\n        endLine |-> 26,\n        endColumn |-> 26,\n        module |-> \"Alias\" ] ]",
		"/\\ y = TRUE\n/\\ x = 2\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 2 e2: TRUE\"\n/\\ te = TRUE\n/\\ TLCGetAction = [ name |-> \"A\",\n  location |->\n      [ beginLine |-> 15,\n        beginColumn |-> 1,\n        endLine |-> 17,\n        endColumn |-> 13,\n        module |-> \"Alias\" ] ]",
		"/\\ y = FALSE\n/\\ x = 3\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 3 e2: FALSE\"\n/\\ te = TRUE\n/\\ TLCGetAction = [ name |-> \"A\",\n  location |->\n      [ beginLine |-> 15,\n        beginColumn |-> 1,\n        endLine |-> 17,\n        endColumn |-> 13,\n        module |-> \"Alias\" ] ]",
		"/\\ y = TRUE\n/\\ x = 4\n/\\ a = 0\n/\\ b = TRUE\n/\\ anim = \"e1: 4 e2: TRUE\"\n/\\ te = TRUE\n/\\ TLCGetAction = [ name |-> \"A\",\n  location |->\n      [ beginLine |-> 15,\n        beginColumn |-> 1,\n        endLine |-> 17,\n        endColumn |-> 13,\n        module |-> \"Alias\" ] ]",
	}
	actions := []string{
		"<Initial predicate line 26, col 18 to line 26, col 26 of module Alias>",
		"<A line 15, col 1 to line 17, col 13 of module Alias>",
		"<A line 15, col 1 to line 17, col 13 of module Alias>",
		"<A line 15, col 1 to line 17, col 13 of module Alias>",
	}
	records := javaTLCRecords(r, tlc.ECTLCStatePrint2)
	if len(records) != len(expected) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expected))
	}
	for i, record := range records {
		if record.StateInfo == nil || record.StateInfo.Info != actions[i] {
			t.Fatalf("trace action %d=%v, want %s", i+1, record.StateInfo, actions[i])
		}
		if got := replJavaTrim(record.StateInfo.String()); got != expected[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expected[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
}
