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
// Ports of AliasSafetyTest and AliasSafetySimuTest.testSpec.
package tlago

import (
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaAliasSafety(t *testing.T) {
	result := runJavaTLCModelTestWithDebugger(t, "Alias", true, true, false, 1, "-config", "Alias.tla")
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want safety violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCPostconditionFalse, tlc.ECTLCPostconditionEvaluationError} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCStats, "4", "4", "0")
	requireJavaTLCRecordedParams(t, result, tlc.ECTLCSearchDepth, "4")
	expectedTrace := []string{
		"/\\ yy = FALSE\n/\\ x = 1\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 1 e2: FALSE\"\n/\\ te = TRUE\n/\\ lvl = <<1, 1>>",
		"/\\ yy = TRUE\n/\\ x = 2\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 2 e2: TRUE\"\n/\\ te = TRUE\n/\\ lvl = <<2, 2>>",
		"/\\ yy = FALSE\n/\\ x = 3\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 3 e2: FALSE\"\n/\\ te = TRUE\n/\\ lvl = <<3, 3>>",
		"/\\ yy = TRUE\n/\\ x = 4\n/\\ a = 0\n/\\ b = TRUE\n/\\ anim = \"e1: 4 e2: TRUE\"\n/\\ te = TRUE\n/\\ lvl = <<4, 4>>",
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
		if !ok || (i == 0 && info != tlc.InitialPredicate) || (i > 0 && (info == tlc.InitialPredicate || strings.HasPrefix(info, "<Action"))) {
			t.Fatalf("trace action %d=%v, want initial predicate or named action", i+1, record.StateInfo.Info)
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
	values := tlc.Globals.MainChecker.GetAllValue(42)
	if len(values) == 0 {
		t.Fatal("register 42 is empty")
	}
	if value, ok := values[0].(*tlc.IntValue); !ok || value.Val != 4 {
		t.Fatalf("register 42=%v, want 4", values[0])
	}
}

func TestJavaAliasSafetySimu(t *testing.T) {
	result := runJavaTLCModelTest(t, "Alias", "-config", "Alias.tla", "-simulate", "num=1")
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want safety violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCPostconditionFalse, tlc.ECTLCPostconditionEvaluationError} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	expectedTrace := []string{
		"/\\ y = FALSE\n/\\ x = 1\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 1 e2: FALSE\"\n/\\ te = TRUE\n/\\ TLCGetAction = [ name |-> \"UnnamedAction\",\n  location |->\n      [ beginLine |-> 26,\n        beginColumn |-> 18,\n        endLine |-> 26,\n        endColumn |-> 26,\n        module |-> \"Alias\" ] ]\n/\\ lvl = <<1, 1>>",
		"/\\ y = TRUE\n/\\ x = 2\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 2 e2: TRUE\"\n/\\ te = TRUE\n/\\ TLCGetAction = [ name |-> \"A\",\n  location |->\n      [ beginLine |-> 15,\n        beginColumn |-> 1,\n        endLine |-> 17,\n        endColumn |-> 13,\n        module |-> \"Alias\" ] ]\n/\\ lvl = <<2, 2>>",
		"/\\ y = FALSE\n/\\ x = 3\n/\\ a = 1\n/\\ b = FALSE\n/\\ anim = \"e1: 3 e2: FALSE\"\n/\\ te = TRUE\n/\\ TLCGetAction = [ name |-> \"A\",\n  location |->\n      [ beginLine |-> 15,\n        beginColumn |-> 1,\n        endLine |-> 17,\n        endColumn |-> 13,\n        module |-> \"Alias\" ] ]\n/\\ lvl = <<3, 3>>",
		"/\\ y = TRUE\n/\\ x = 4\n/\\ a = 0\n/\\ b = TRUE\n/\\ anim = \"e1: 4 e2: TRUE\"\n/\\ te = TRUE\n/\\ TLCGetAction = [ name |-> \"A\",\n  location |->\n      [ beginLine |-> 15,\n        beginColumn |-> 1,\n        endLine |-> 17,\n        endColumn |-> 13,\n        module |-> \"Alias\" ] ]\n/\\ lvl = <<4, 4>>",
	}
	expectedActions := []string{
		"<Initial predicate line 26, col 18 to line 26, col 26 of module Alias>",
		"<A line 15, col 1 to line 17, col 13 of module Alias>",
		"<A line 15, col 1 to line 17, col 13 of module Alias>",
		"<A line 15, col 1 to line 17, col 13 of module Alias>",
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
		if !ok || info != expectedActions[i] {
			t.Fatalf("trace action %d=%v, want %s", i+1, record.StateInfo.Info, expectedActions[i])
		}
		if got := strings.TrimSpace(record.StateInfo.String()); got != expectedTrace[i] {
			t.Fatalf("trace state %d=%q, want %q", i+1, got, expectedTrace[i])
		}
		if record.StateNumber != i+1 {
			t.Fatalf("trace ordinal=%d, want %d", record.StateNumber, i+1)
		}
	}
}
