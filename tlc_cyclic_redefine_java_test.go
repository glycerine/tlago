/*******************************************************************************
 * Copyright (c) 2022 Microsoft Research. All rights reserved.
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
// Ports of the existing CyclicRedefine{Instance,Init,Next,Op,SubAction,Vars}Test
// testSpec methods, after the LET export and cyclic config-rewrite features.
package tlago

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func runJavaCyclicRedefine(t *testing.T, model, config string, status int) *tlc.Result {
	t.Helper()
	t.Setenv("tlc2.tool.impl.SpecProcessor.allowCyclicRedefinitions", "true")
	result := runJavaTLCModelTestWithDebugger(t, model, false, true, false, 1, "-config", config, "-noGenerateSpecTE")
	if result.ExitStatus != status {
		t.Fatalf("exit status=%d, want %d", result.ExitStatus, status)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	if got := javaTLCRecords(result, tlc.ECGeneral); len(got) != 0 {
		t.Fatalf("unexpected GENERAL: %v", got)
	}
	return result
}

func requireJavaCyclicRedefineTrace(t *testing.T, result *tlc.Result, action string) {
	t.Helper()
	expectedTrace := []string{"x = FALSE", "x = TRUE"}
	expectedActions := []string{tlc.InitialPredicate, action}
	records := javaTLCRecords(result, tlc.ECTLCStatePrint2)
	if len(records) != len(expectedTrace) {
		t.Fatalf("trace length=%d, want %d", len(records), len(expectedTrace))
	}
	for i, record := range records {
		if record.StateInfo == nil {
			t.Fatalf("trace state %d has no TLCStateInfo", i+1)
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if got := record.StateInfo.Info; got != expectedActions[i] {
			t.Fatalf("trace action %d=%v, want %q", i+1, got, expectedActions[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
}

func TestJavaCyclicRedefineInstance(t *testing.T) {
	result := runJavaCyclicRedefine(t, "CyclicRedefineInstance", "CyclicRedefineInstance.cfg", tlc.ExitStatusViolationSafety)
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "3", "2", "0")
	requireJavaCyclicRedefineTrace(t, result, "<ReDefNext line 7, col 15 to line 7, col 23 of module CyclicRedefineInstance>")
}

func TestJavaCyclicRedefineInit(t *testing.T) {
	result := runJavaCyclicRedefine(t, "CyclicRedefine", "CyclicRedefineInit.cfg", tlc.ExitStatusViolationSafety)
	requireJavaCyclicRedefineTrace(t, result, "<A(1) line 8, col 8 to line 8, col 21 of module Base>")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCInitGenerated1, "1")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "2", "2", "0")
}

func TestJavaCyclicRedefineNext(t *testing.T) {
	result := runJavaCyclicRedefine(t, "CyclicRedefine", "CyclicRedefineNext.cfg", tlc.ExitStatusSuccess)
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "4", "1", "0")
}

func TestJavaCyclicRedefineOp(t *testing.T) {
	result := runJavaCyclicRedefine(t, "CyclicRedefine", "CyclicRedefineOp.cfg", tlc.ExitStatusSuccess)
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "2")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "13", "2", "0")
}

func TestJavaCyclicRedefineSubAction(t *testing.T) {
	result := runJavaCyclicRedefine(t, "CyclicRedefine", "CyclicRedefineSubAction.cfg", tlc.ExitStatusSuccess)
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "1")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "4", "1", "0")
}

func TestJavaCyclicRedefineVars(t *testing.T) {
	result := runJavaCyclicRedefine(t, "CyclicRedefineVars", "CyclicRedefineVars.cfg", tlc.ExitStatusSuccess)
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "3")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "25", "4", "0")
}
