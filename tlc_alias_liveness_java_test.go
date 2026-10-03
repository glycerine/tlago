/*******************************************************************************
 * Copyright (c) 2020 Microsoft Research. All rights reserved.
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
// Ports of AliasLivenessLassoTest and AliasLivenessStutteringTest.testSpec.
package tlago

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaAliasLivenessLasso(t *testing.T) {
	result := runJavaTLCModelTest(t, "Alias", "-config", "AliasLasso.cfg")
	requireJavaAliasLivenessResult(t, result)
	expectedTrace := []string{
		"/\\ yy = FALSE\n/\\ x = 1\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 1 e2: FALSE\"\n/\\ te = TRUE\n/\\ lvl = <<1, 1>>\n/\\ trc = <<[x |-> 1, y |-> FALSE]>>",
		"/\\ yy = TRUE\n/\\ x = 2\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 2 e2: TRUE\"\n/\\ te = TRUE\n/\\ lvl = <<2, 2>>\n/\\ trc = <<[x |-> 1, y |-> FALSE], [x |-> 2, y |-> TRUE]>>",
		"/\\ yy = FALSE\n/\\ x = 3\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 3 e2: FALSE\"\n/\\ te = TRUE\n/\\ lvl = <<3, 3>>\n/\\ trc = <<[x |-> 1, y |-> FALSE], [x |-> 2, y |-> TRUE], [x |-> 3, y |-> FALSE]>>",
		"/\\ yy = TRUE\n/\\ x = 4\n/\\ a = -3\n/\\ b = FALSE\n/\\ anim = \"e1: 4 e2: TRUE\"\n/\\ te = TRUE\n/\\ lvl = <<4, 4>>\n/\\ trc = << [x |-> 1, y |-> FALSE],\n   [x |-> 2, y |-> TRUE],\n   [x |-> 3, y |-> FALSE],\n   [x |-> 4, y |-> TRUE] >>",
	}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != len(expectedTrace) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expectedTrace))
	}
	for i, record := range records {
		if record.StateInfo == nil {
			t.Fatalf("trace state %d has no TLCStateInfo", i+1)
		}
		info, ok := record.StateInfo.Info.(string)
		if !ok || info == tlc.InitialPredicate || strings.HasPrefix(info, "<Action") {
			t.Fatalf("trace action %d=%v, want a named action", i+1, record.StateInfo.Info)
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCBackToState, "1", "<B line 20, col 1 to line 22, col 9 of module Alias>")
}

func TestJavaAliasLivenessStuttering(t *testing.T) {
	result := runJavaTLCModelTest(t, "Alias", "-config", "AliasStuttering.cfg")
	requireJavaAliasLivenessResult(t, result)
	expectedTrace := []string{
		"/\\ yy = FALSE\n/\\ x = 1\n/\\ a = 0\n/\\ b = TRUE\n/\\ anim = \"e1: 1 e2: FALSE\"\n/\\ te = TRUE\n/\\ lvl = <<1, 1>>\n/\\ trc = <<[x |-> 1, y |-> FALSE]>>",
	}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != len(expectedTrace) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expectedTrace))
	}
	for i, record := range records {
		if record.StateInfo == nil {
			t.Fatalf("trace state %d has no TLCStateInfo", i+1)
		}
		info, ok := record.StateInfo.Info.(string)
		if !ok || info == tlc.InitialPredicate || strings.HasPrefix(info, "<Action") {
			t.Fatalf("trace action %d=%v, want a named action", i+1, record.StateInfo.Info)
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
	stuttering := javaTLCRecords(result, tlc.ECTLCStatePrint3)
	if len(stuttering) == 0 || stuttering[0].StateNumber != 2 {
		t.Fatalf("stuttering=%v, want ordinal 2", stuttering)
	}
}

func requireJavaAliasLivenessResult(t *testing.T, result *tlc.Result) {
	t.Helper()
	if result.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit status=%d, want liveness violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "5", "4", "0")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "4")
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCPostconditionFalse, tlc.ECTLCPostconditionEvaluationError} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCTemporalPropertyViolated, "Prop")
	if len(javaTLCRecords(result, tlc.ECTLCCounterExample)) == 0 {
		t.Fatal("TLC_COUNTER_EXAMPLE not recorded")
	}
}
