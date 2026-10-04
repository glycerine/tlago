/*******************************************************************************
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software and to permit persons to whom the Software is furnished to do
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
