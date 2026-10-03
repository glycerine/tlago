/*******************************************************************************
 * Copyright (c) 2025, 2026 NVIDIA Corp. All rights reserved.
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
// Copyright (c) 2022, Oracle and/or its affiliates.
// Copyright (c) 2018, 2019 Microsoft Research. All rights reserved.
// Copyright (c) 2025 Microsoft Corp. All rights reserved.
// Ports of the existing checker testSpec methods, after their implementation.
package tlago

import (
	"reflect"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

func TestJavaConstLevelInvariant(t *testing.T) {
	result := runJavaTLCModelTestWithCoverage(t, "ConstLevelInvariant", false, "-config", "ConstLevelInvariant.tla")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	matched := false
	for _, m := range javaTLCRecords(result, tlc.ECTLCStats) {
		if reflect.DeepEqual(m.Params, []string{"3", "2", "0"}) {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("TLC_STATS=%v, want 3/2/0", javaTLCRecords(result, tlc.ECTLCStats))
	}
	matched = false
	for _, m := range javaTLCRecords(result, tlc.ECTLCSearchDepth) {
		if len(m.Params) > 0 && m.Params[0] == "2" {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("TLC_SEARCH_DEPTH=%v, want 2", javaTLCRecords(result, tlc.ECTLCSearchDepth))
	}
	if len(javaTLCRecords(result, tlc.ECTLCInvariantConstantLevel)) == 0 {
		t.Fatal("TLC_INVARIANT_CONSTANT_LEVEL not recorded")
	}
	for _, m := range javaTLCRecords(result, tlc.ECTLCCoverageValue) {
		if len(m.Params) > 1 && strings.TrimSpace(m.Params[1]) == "0" {
			t.Fatalf("unexpected uncovered line: %v", m.Params)
		}
	}
}

func TestJavaPostConditions(t *testing.T) {
	result := runJavaTLCModelTest(t, "PostConditionsTest")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCPostconditionFalse, tlc.ECTLCPostconditionEvaluationError} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	reg100 := tlc.Globals.MainChecker.GetAllValue(100)
	if len(reg100) == 0 {
		t.Fatal("Post1 was not evaluated: register 100 is empty")
	}
	if value, ok := reg100[0].(*tlc.IntValue); !ok || value.Val != 4 {
		t.Fatalf("Post1 did not write generated count: %v", reg100[0])
	}
	reg101 := tlc.Globals.MainChecker.GetAllValue(101)
	if len(reg101) == 0 {
		t.Fatal("Post2 was not evaluated: register 101 is empty")
	}
	if value, ok := reg101[0].(*tlc.IntValue); !ok || value.Val != 4 {
		t.Fatalf("Post2 did not write distinct count: %v", reg101[0])
	}
}

func TestJavaPostConditionsFail(t *testing.T) {
	result := runJavaTLCModelTest(t, "PostConditionsTest", "-config", "PostConditionsFailTest.cfg")
	if result.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit status=%d, want assumption violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	matched := false
	for _, m := range javaTLCRecords(result, tlc.ECTLCPostconditionFalse) {
		if len(m.Params) > 0 && m.Params[0] == "PostFail" {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("TLC_POSTCONDITION_FALSE=%v, want PostFail", javaTLCRecords(result, tlc.ECTLCPostconditionFalse))
	}
}

func TestJavaConstantOperatorConfiguration(t *testing.T) {
	result := runJavaTLCModelTest(t, "ConstantOperatorConfiguration")
	if result.ExitStatus != tlc.ExitStatusViolationSafety {
		t.Fatalf("exit status=%d, want safety violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	matched := false
	for _, m := range javaTLCRecords(result, tlc.ECTLCInvariantViolatedBehavior) {
		if len(m.Params) > 0 && m.Params[0] == "Inv" {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("TLC_INVARIANT_VIOLATED_BEHAVIOR=%v, want Inv", javaTLCRecords(result, tlc.ECTLCInvariantViolatedBehavior))
	}
}

func TestJavaPossibleCounts(t *testing.T) {
	result := runJavaTLCModelTest(t, "PossibleCountsTest")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	for _, code := range []int{tlc.ECTLCConfigIDMustNotBeConstant, tlc.ECTLCPostconditionFalse, tlc.ECTLCPostconditionEvaluationError} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCPossibleUnwitnessed} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
}

func TestJavaTLCSet(t *testing.T) {
	result := runJavaTLCModelTestWithSettings(t, "TLCSet", true, false, "-config", "TLCSetPost.cfg")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d, want success", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCPostconditionFalse, tlc.ECTLCPostconditionEvaluationError} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	for _, m := range javaTLCRecords(result, tlc.ECTLCCoverageValue) {
		if len(m.Params) > 1 && strings.TrimSpace(m.Params[1]) == "0" {
			t.Fatalf("unexpected uncovered line: %v", m.Params)
		}
	}
}

func TestJavaTLCGetNamedUndefined(t *testing.T) {
	result := runJavaTLCModelTest(t, "TLCGetNamedUndefined")
	if result.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit status=%d, want config parse error", result.ExitStatus)
	}
	for _, code := range []int{tlc.ECTLCFinished, tlc.ECTLCModuleTLCGetUndefined, tlc.ECTLCConfigSubstitutionNonConstant} {
		if len(javaTLCRecords(result, code)) == 0 {
			t.Fatalf("diagnostic %d not recorded", code)
		}
	}
}

func TestJavaGithub1109(t *testing.T) {
	result := runJavaTLCModelTest(t, "Github1109", "-config", "Github1109.tla")
	if result.ExitStatus != tlc.ExitStatusViolationAssumption {
		t.Fatalf("exit status=%d, want assumption violation", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCConfigSubstitutionNonConstant} {
		if got := javaTLCRecords(result, code); len(got) != 0 {
			t.Fatalf("unexpected diagnostic %d: %v", code, got)
		}
	}
	records := javaTLCRecords(result, tlc.ECTLCAssumptionEvaluationError)
	if len(records) == 0 {
		t.Fatal("TLC_ASSUMPTION_EVALUATION_ERROR not recorded")
	}
	matched := false
	for _, parameter := range records[0].Params {
		if strings.Contains(parameter, "Seq({\"s\"})\ncannot be enumerated") {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("TLC_ASSUMPTION_EVALUATION_ERROR=%v, want unenumerable Seq(S)", records)
	}
}

func TestJavaGithub1109a(t *testing.T) {
	result := runJavaTLCModelTest(t, "Github1109a", "-config", "Github1109a.tla")
	if result.ExitStatus != tlc.ExitStatusErrorConfigParse {
		t.Fatalf("exit status=%d, want config parse error", result.ExitStatus)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	matched := false
	for _, m := range javaTLCRecords(result, tlc.ECTLCConfigSubstitutionNonConstant) {
		if len(m.Params) >= 2 && m.Params[0] == "C" && m.Params[1] == "C2" {
			matched = true
		}
	}
	if !matched {
		t.Fatalf("TLC_CONFIG_SUBSTITUTION_NON_CONSTANT=%v, want C/C2", javaTLCRecords(result, tlc.ECTLCConfigSubstitutionNonConstant))
	}
}
