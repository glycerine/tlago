/*******************************************************************************
 * Copyright (c) 2019 Microsoft Research. All rights reserved.
 * Copyright (c) 2026 NVIDIA Corp. All rights reserved.
 *
 * The MIT License (MIT)
 *******************************************************************************/
package tlago

import (
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Original TLCGetNonDeterminismTest.testSpec, four-worker override and inherited
// SUCCESS exit. Upstream ignores the method as "No known fix/By design".
func TestJavaTLCGetNonDeterminismIgnored(t *testing.T) {
	t.Skip("upstream @Ignore: No known fix/By design")
	r := runJavaTLCModelTestWithWorkers(t, "TLCGetNonDeterminism", true, true, 4)
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	if records := javaTLCRecords(r, tlc.ECGeneral); len(records) != 0 {
		t.Fatalf("GENERAL recorded: %v", records)
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCTemporalPropertyViolated, "Prop")
	if len(javaTLCRecords(r, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE absent")
	}
	requireJavaRandomSubsetTrace(t, r, []string{
		"/\\ x = 0\n/\\ y = 0",
		"/\\ x = 1\n/\\ y = 1",
		"/\\ x = 2\n/\\ y = 2",
		"/\\ x = 3\n/\\ y = 3",
	}, false)
	stuttering := javaTLCRecords(r, tlc.ECTLCStatePrint3)
	if len(stuttering) == 0 || stuttering[0].StateNumber != 5 {
		t.Fatalf("stuttering=%v, want ordinal 5", stuttering)
	}
}
