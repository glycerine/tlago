package tlago

import (
	"github.com/glycerine/tlago/tlc"
	"testing"
)

// Original TemporalDoubleNegationAlwaysDoubleNegationTest.testValidProperty, inherited from AbstractTemporalDoubleNegationTest.
func TestJavaTemporalDoubleNegationAlwaysDoubleNegation(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "TemporalDoubleNegationAlwaysDoubleNegation", "TemporalDoubleNegation", false, false, false, 1, "-noGenerateSpecTE", "-config", "TemporalDoubleNegationAlwaysDoubleNegation")
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
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("A valid property must not produce a temporal violation")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("A valid property must not produce a counterexample")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original TemporalDoubleNegationAlwaysDualTest.testValidProperty, inherited from AbstractTemporalDoubleNegationTest.
func TestJavaTemporalDoubleNegationAlwaysDual(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "TemporalDoubleNegationAlwaysDual", "TemporalDoubleNegation", false, false, false, 1, "-noGenerateSpecTE", "-config", "TemporalDoubleNegationAlwaysDual")
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
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("A valid property must not produce a temporal violation")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("A valid property must not produce a counterexample")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original TemporalDoubleNegationEventuallyDoubleNegationTest.testValidProperty, inherited from AbstractTemporalDoubleNegationTest.
func TestJavaTemporalDoubleNegationEventuallyDoubleNegation(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "TemporalDoubleNegationEventuallyDoubleNegation", "TemporalDoubleNegation", false, false, false, 1, "-noGenerateSpecTE", "-config", "TemporalDoubleNegationEventuallyDoubleNegation")
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
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("A valid property must not produce a temporal violation")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("A valid property must not produce a counterexample")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original TemporalDoubleNegationFairEventuallyDoubleNegationTest.testValidProperty, inherited from AbstractTemporalDoubleNegationTest.
func TestJavaTemporalDoubleNegationFairEventuallyDoubleNegation(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "TemporalDoubleNegationFairEventuallyDoubleNegation", "TemporalDoubleNegation", false, false, false, 1, "-noGenerateSpecTE", "-config", "TemporalDoubleNegationFairEventuallyDoubleNegation")
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
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("A valid property must not produce a temporal violation")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("A valid property must not produce a counterexample")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original TemporalDoubleNegationFairLeadsToDualTest.testValidProperty, inherited from AbstractTemporalDoubleNegationTest.
func TestJavaTemporalDoubleNegationFairLeadsToDual(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "TemporalDoubleNegationFairLeadsToDual", "TemporalDoubleNegation", false, false, false, 1, "-noGenerateSpecTE", "-config", "TemporalDoubleNegationFairLeadsToDual")
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
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("A valid property must not produce a temporal violation")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("A valid property must not produce a counterexample")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original TemporalDoubleNegationImplicationDoubleNegationTest.testValidProperty, inherited from AbstractTemporalDoubleNegationTest.
func TestJavaTemporalDoubleNegationImplicationDoubleNegation(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "TemporalDoubleNegationImplicationDoubleNegation", "TemporalDoubleNegation", false, false, false, 1, "-noGenerateSpecTE", "-config", "TemporalDoubleNegationImplicationDoubleNegation")
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
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("A valid property must not produce a temporal violation")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("A valid property must not produce a counterexample")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original TemporalDoubleNegationImplicationTautologyTest.testValidProperty, inherited from AbstractTemporalDoubleNegationTest.
func TestJavaTemporalDoubleNegationImplicationTautology(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "TemporalDoubleNegationImplicationTautology", "TemporalDoubleNegation", false, false, false, 1, "-noGenerateSpecTE", "-config", "TemporalDoubleNegationImplicationTautology")
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
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("A valid property must not produce a temporal violation")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("A valid property must not produce a counterexample")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}

// Original TemporalDoubleNegationLeadsToDualTest.testValidProperty, inherited from AbstractTemporalDoubleNegationTest.
func TestJavaTemporalDoubleNegationLeadsToDual(t *testing.T) {
	t.Setenv("tlc2.tool.ModelChecker.vetoCleanup", "true")
	r := runJavaTLCModelTestWithRoot(t, "TemporalDoubleNegationLeadsToDual", "TemporalDoubleNegation", false, false, false, 1, "-noGenerateSpecTE", "-config", "TemporalDoubleNegationLeadsToDual")
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
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("A valid property must not produce a temporal violation")
	}
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) != 0 {
		t.Fatal("A valid property must not produce a counterexample")
	}
	if len(javaTLCRecords(r, tlc.ECTLCSuccess)) == 0 {
		t.Fatal("TLC_SUCCESS absent")
	}
}
