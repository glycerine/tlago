/*******************************************************************************
 * Copyright (c) 2015, 2018 Microsoft Research. All rights reserved.
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

// Ports of the original DoInitFunctor checker testSpec methods.
package tlago

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaDoInitFunctorInvariant(t *testing.T) {
	result := runJavaTLCModelTest(t, "DoInitFunctorInvariant")
	requireJavaInitInvariantStop(t, result, "NotNine", "x = 9\n")
}

func TestJavaDoInitFunctorInvariantNoContinue(t *testing.T) {
	result := runJavaTLCModelTest(t, "DoInitFunctorInvariantContinue")
	requireJavaInitInvariantStop(t, result, "Inv", "x = 1\n")
}

func requireJavaInitInvariantStop(t *testing.T, result *tlc.Result, invariant, state string) {
	t.Helper()
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want safety violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECTLCStats, tlc.ECGeneral} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInvariantViolatedInitial, invariant, state)
	if got := javaTLCRecords(result, tlc.ECTLCInvariantViolatedInitial); len(got) != 1 {
		t.Fatalf("initial invariant violations=%v, want one", got)
	}
}

func TestJavaDoInitFunctorInvariantContinue(t *testing.T) {
	result := runJavaTLCModelTest(t, "DoInitFunctorInvariantContinue", "-continue")
	// Preserve Java's successful exit with continuation despite violations.
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "21", "11")
	if got := javaTLCRecords(result, tlc.ECGeneral); len(got) != 0 {
		t.Fatalf("unexpected GENERAL: %v", got)
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInvariantViolatedInitial, "Inv", "x = 1\n")
	violations := javaTLCRecords(result, tlc.ECTLCInvariantViolatedInitial)
	if len(violations) != 10 {
		t.Fatalf("initial invariant violations=%v, want ten", violations)
	}
	for j, m := range violations {
		want := []string{"Inv", fmt.Sprintf("x = %d\n", j+1)}
		if !reflect.DeepEqual(m.Params, want) {
			t.Fatalf("violation %d=%v, want %v", j+1, m.Params, want)
		}
	}
	for _, m := range javaTLCRecords(result, tlc.ECTLCCoverageValue) {
		if len(m.Params) > 1 && strings.TrimSpace(m.Params[1]) == "0" {
			t.Fatalf("unexpected uncovered line: %v", m.Params)
		}
	}
}

func TestJavaDoInitFunctorInvariantMinimalErrorStack(t *testing.T) {
	result := runJavaTLCModelTest(t, "DoInitFunctorMinimalErrorStack")
	if result.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit status=%d, want spec evaluation failure", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "1", "1", "1")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCModuleArgumentErrorAn, "first", ">=", "integer", "<<1, 1>>")
	errorStack := "0. Line 8, column 3 to line 9, column 13 in DoInitFunctorMinimalErrorStack\n" +
		"1. Line 8, column 6 to line 8, column 25 in DoInitFunctorMinimalErrorStack\n" +
		"2. Line 9, column 6 to line 9, column 13 in DoInitFunctorMinimalErrorStack\n" +
		"3. Line 14, column 8 to line 14, column 13 in DoInitFunctorMinimalErrorStack\n\n"
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCNestedExpression, errorStack)
	requireJavaTLCUncovered(t, result, "line 11, col 20 to line 11, col 33 of module DoInitFunctorMinimalErrorStack")
}

// TestMPRecorder.recordedWithStringValues compares a prefix of the first record.
func requireJavaTLCRecordedParams(t *testing.T, result *tlc.Result, code int, params ...string) {
	t.Helper()
	records := javaTLCRecords(result, code)
	if len(records) == 0 || len(records[0].Params) < len(params) || !reflect.DeepEqual(records[0].Params[:len(params)], params) {
		t.Fatalf("diagnostic %d=%v, want first-record prefix %v", code, records, params)
	}
}

func TestJavaDoInitFunctorProperty(t *testing.T) {
	result := runJavaTLCModelTest(t, "DoInitFunctorProperty")
	if result.ExitStatus != tlc.ExitStatusViolationLiveness {
		t.Fatalf("exit status=%d, want liveness violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECTLCStats, tlc.ECGeneral} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCPropertyViolatedInitial, "NotNine", "x = 9\n")
}

func TestJavaDoInitFunctorEvalException(t *testing.T) {
	result := runJavaTLCModelTest(t, "DoInitFunctorEvalException")
	if result.ExitStatus != tlc.ExitStatusFailureSpecEval {
		t.Fatalf("exit status=%d, want spec evaluation failure", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if len(javaTLCRecords(result, tlc.ECTLCStats)) == 0 {
		t.Fatal("TLC_STATS not recorded")
	}
	if got := javaTLCRecords(result, tlc.ECGeneral); len(got) != 0 {
		t.Fatalf("unexpected GENERAL: %v", got)
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInitialState,
		"TLC expected a boolean value, but did not find one. line 15, col 15 to line 15, col 18 of module DoInitFunctorEvalException",
		"x = 1\n")
	requireJavaTLCUncovered(t, result, "line 9, col 9 to line 9, col 14 of module DoInitFunctorEvalException")
}

func requireJavaTLCUncovered(t *testing.T, result *tlc.Result, locations ...string) {
	t.Helper()
	uncovered := map[string]bool{}
	for _, m := range javaTLCRecords(result, tlc.ECTLCCoverageValue) {
		if len(m.Params) > 1 && strings.TrimSpace(m.Params[1]) == "0" {
			uncovered[strings.TrimSpace(strings.ReplaceAll(m.Params[0], "|", ""))] = true
		}
	}
	want := map[string]bool{}
	for _, location := range locations {
		want[location] = true
	}
	if !reflect.DeepEqual(uncovered, want) {
		t.Fatalf("uncovered=%v, want %v", uncovered, want)
	}
}

// Original distributed-package DistributedDoInitFunctorInvariantContinueTest
// extends the ordinary ModelCheckerTestCase. It runs the NotNine fixture with
// continuation, not the separate Inv fixture in DoInitFunctorInvariantContinue.
func TestJavaDistributedDoInitFunctorInvariantContinue(t *testing.T) {
	r := runJavaTLCModelTest(t, "DoInitFunctorInvariant", "-continue")
	if r.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit=%d, want SUCCESS", r.ExitStatus)
	}
	if len(javaTLCRecords(r, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED absent")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCStats, "21", "11")
	if len(javaTLCRecords(r, tlc.ECGeneral)) != 0 {
		t.Fatal("GENERAL present")
	}
	requireJavaTLCRecordedParams(t, r, tlc.ECTLCInvariantViolatedInitial, "NotNine", "x = 9\n")
	requireJavaTLCUncovered(t, r)
}
