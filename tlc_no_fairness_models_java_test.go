/*******************************************************************************
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
	"path/filepath"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original ModelCheckerTestCase setup, with each source class's coverage override.
func runJavaNoFairnessButLiveProp(t *testing.T, class, config string) *tlc.Result {
	t.Helper()
	return runJavaTLCModelTestWithCoverage(t, "NoFairnessButLiveProp", false, "-config", config, "-dumpTrace", "json", filepath.Join(t.TempDir(), class+".json"))
}

// Original NoFairnessButLivePropATest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropA(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropATest", "NoFairnessButLivePropA.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) == 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original NoFairnessButLivePropBTest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropB(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropBTest", "NoFairnessButLivePropB.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original NoFairnessButLivePropDTest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropD(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropDTest", "NoFairnessButLivePropD.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) == 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original NoFairnessButLivePropGTest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropG(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropGTest", "NoFairnessButLivePropG.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoSpecButProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_SPEC_BUT_PROPERTY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original NoFairnessButLivePropHTest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropH(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropHTest", "NoFairnessButLivePropH.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoSpecButProperty)) == 0 {
		t.Fatal("TLC_CONFIG_NO_SPEC_BUT_PROPERTY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original NoFairnessButLivePropNoWarningCustomFairTest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropNoWarningCustomFair(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropNoWarningCustomFairTest", "NoFairnessButLivePropNoWarningCustomFair.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoSpecButProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_SPEC_BUT_PROPERTY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original NoFairnessButLivePropNoWarningNoFairTest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropNoWarningNoFair(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropNoWarningNoFairTest", "NoFairnessButLivePropNoWarningNoFair.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoSpecButProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_SPEC_BUT_PROPERTY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original NoFairnessButLivePropNoWarningWithFairTest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropNoWarningWithFair(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropNoWarningWithFairTest", "NoFairnessButLivePropNoWarningWithFair.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoSpecButProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_SPEC_BUT_PROPERTY present")
	}
	requireJavaTLCUncovered(t, r)
}

// Original NoFairnessButLivePropOTest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropO(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropOTest", "NoFairnessButLivePropO.cfg")
	if r.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit=%d, want VIOLATION_LIVENESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoSpecButProperty)) == 0 {
		t.Fatal("TLC_CONFIG_NO_SPEC_BUT_PROPERTY absent")
	}
	requireJavaTLCUncovered(t, r)
}

// Original NoFairnessButLivePropPTest.testSpec and constructor settings.
func TestJavaNoFairnessButLivePropP(t *testing.T) {
	r := runJavaNoFairnessButLiveProp(t, "NoFairnessButLivePropPTest", "NoFairnessButLivePropP.cfg")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS; messages=%v", r.ExitStatus, r.Messages)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "3", "2", "0")
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCSearchDepth, "2")
	if len(javaTLCRecords(r, tlc.ECTLCTemporalPropertyViolated)) != 0 {
		t.Fatal("TLC_TEMPORAL_PROPERTY_VIOLATED present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoFairnessButLiveProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_FAIRNESS_BUT_LIVE_PROPERTY present")
	}
	if len(javaTLCRecords(r, tlc.ECTLCConfigNoSpecButProperty)) != 0 {
		t.Fatal("TLC_CONFIG_NO_SPEC_BUT_PROPERTY present")
	}
	requireJavaTLCUncovered(t, r)
}
