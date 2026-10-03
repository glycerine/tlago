/*******************************************************************************
 * Copyright (c) 2018-2021 Microsoft Research. All rights reserved.
 * Copyright (c) 2026 NVIDIA Corporation. All rights reserved.
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
 *   loki der quaeler - initial API and implementation
 *   Markus Alexander Kuppe
 ******************************************************************************/
// Ports existing Java coverage model tests and CommonTestCase.assertCoverage after the
// coverage implementation. Fixtures and expected records retain upstream data.
package tlago

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// TestMPRecorder.Coverage equality deliberately ignores indentation and compares
// only location, count and cost. Keep that same assertion boundary here.
type javaCoverageRecord struct {
	location    string
	count, cost int64
}

func parseJavaCoverageRecord(t *testing.T, fields []string) javaCoverageRecord {
	t.Helper()
	r := javaCoverageRecord{location: strings.TrimSpace(strings.ReplaceAll(fields[0], "|", "")), count: -1, cost: -1}
	if len(fields) > 1 {
		var err error
		r.count, err = strconv.ParseInt(strings.TrimSpace(fields[1]), 10, 64)
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(fields) > 2 {
		var err error
		r.cost, err = strconv.ParseInt(strings.TrimSpace(fields[2]), 10, 64)
		if err != nil {
			t.Fatal(err)
		}
	}
	return r
}

func assertJavaCoverage(t *testing.T, result *tlc.Result, expected string) {
	t.Helper()
	expectedZero, actualZero := map[javaCoverageRecord]bool{}, map[javaCoverageRecord]bool{}
	var expectedValues, expectedCost, expectedActions []javaCoverageRecord
	for _, line := range strings.Split(expected, "\n") {
		r := parseJavaCoverageRecord(t, strings.Split(line, ":"))
		if strings.HasPrefix(r.location, "<") {
			expectedActions = append(expectedActions, r)
		} else {
			if r.count == 0 {
				expectedZero[r] = true
			}
			if r.cost >= 0 {
				expectedCost = append(expectedCost, r)
			} else {
				expectedValues = append(expectedValues, r)
			}
		}
	}
	var actualValues, actualCost, actualActions []javaCoverageRecord
	for _, m := range javaTLCRecords(result, tlc.ECTLCCoverageValue) {
		r := parseJavaCoverageRecord(t, m.Params)
		if r.count == 0 {
			actualZero[r] = true
		} else {
			actualValues = append(actualValues, r)
		}
	}
	for _, m := range javaTLCRecords(result, tlc.ECTLCCoverageValueCost) {
		r := parseJavaCoverageRecord(t, m.Params)
		if r.count != 0 {
			actualCost = append(actualCost, r)
		}
	}
	// Step B1 removes the actual zero-coverage set from the expected values.
	nonzero := expectedValues[:0]
	for _, r := range expectedValues {
		if !actualZero[r] {
			nonzero = append(nonzero, r)
		}
	}
	expectedValues = nonzero
	// The Java recorder groups actions by INIT, NEXT, PROPERTY, CONSTRAINT.
	for _, code := range []int{tlc.ECTLCCoverageInit, tlc.ECTLCCoverageNext, tlc.ECTLCCoverageProperty, tlc.ECTLCCoverageConstraint} {
		for _, m := range javaTLCRecords(result, code) {
			r := parseJavaCoverageRecord(t, m.Params)
			if strings.HasPrefix(r.location, "<") {
				actualActions = append(actualActions, r)
			}
		}
	}
	if !reflect.DeepEqual(actualZero, expectedZero) {
		t.Errorf("zero coverage=%v, want %v", actualZero, expectedZero)
	}
	if !slices.Equal(actualValues, expectedValues) {
		t.Errorf("coverage=%v, want %v", actualValues, expectedValues)
	}
	if !slices.Equal(actualCost, expectedCost) {
		t.Errorf("cost coverage=%v, want %v", actualCost, expectedCost)
	}
	if !slices.Equal(actualActions, expectedActions) {
		t.Errorf("action coverage=%v, want %v", actualActions, expectedActions)
	}
}

func TestJavaACoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "A", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"2"}) {
		t.Errorf("search depth=%v, want 2", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"7", "3", "0"}) {
		t.Errorf("stats=%v, want 7/3/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 7, col 1 to line 7, col 4 of module A>: 1:1
  line 7, col 12 to line 7, col 16 of module A: 1
  line 8, col 12 to line 8, col 21 of module A: 1
  |line 5, col 11 to line 5, col 49 of module A: 1
  ||line 5, col 31 to line 5, col 49 of module A: 131072
  ||line 5, col 20 to line 5, col 27 of module A: 1
  |line 8, col 16 to line 8, col 20 of module A: 1
<A line 13, col 1 to line 13, col 4 of module A>: 1:3
  line 13, col 9 to line 13, col 19 of module A: 3
  |line 13, col 14 to line 13, col 19 of module A: 3
  ||line 11, col 11 to line 11, col 61 of module A: 3
  |||line 11, col 31 to line 11, col 61 of module A: 30
  ||||line 11, col 57 to line 11, col 61 of module A: 162
  ||||line 11, col 41 to line 11, col 52 of module A: 30
  |||line 11, col 24 to line 11, col 27 of module A: 3
  ||line 13, col 18 to line 13, col 18 of module A: 3
<B line 15, col 1 to line 15, col 4 of module A>: 1:3
  line 15, col 9 to line 15, col 19 of module A: 3
  |line 15, col 14 to line 15, col 19 of module A: 3
  ||line 11, col 11 to line 11, col 61 of module A: 3
  |||line 11, col 31 to line 11, col 61 of module A: 9
  ||||line 11, col 57 to line 11, col 61 of module A: 15
  ||||line 11, col 41 to line 11, col 52 of module A: 9
  |||line 11, col 24 to line 11, col 27 of module A: 3
  ||line 15, col 18 to line 15, col 18 of module A: 3`)
}

// Port of CoverageStatisticsTest.testSpec with the original records and model.
func TestJavaCoverageStatistics(t *testing.T) {
	result := runJavaTLCModelTest(t, "CoverageStatistics", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"17"}) {
		t.Errorf("search depth=%v, want 17", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"98", "19", "0"}) {
		t.Errorf("stats=%v, want 98/19/0", stats)
	}
	if records := javaTLCRecords(result, tlc.ECGeneral); len(records) != 0 {
		t.Errorf("unexpected GENERAL: %v", records)
	}
	assertJavaCoverage(t, result, `<Init line 12, col 1 to line 12, col 4 of module CoverageStatistics>: 3:3
  line 12, col 12 to line 12, col 21 of module CoverageStatistics: 1
  line 13, col 12 to line 13, col 16 of module CoverageStatistics: 3
<A line 15, col 1 to line 15, col 1 of module CoverageStatistics>: 16:19
  line 15, col 9 to line 15, col 17 of module CoverageStatistics: 38
  |line 15, col 9 to line 15, col 9 of module CoverageStatistics: 19
  |line 15, col 15 to line 15, col 17 of module CoverageStatistics: 19
  line 16, col 9 to line 16, col 17 of module CoverageStatistics: 38
  |line 16, col 9 to line 16, col 9 of module CoverageStatistics: 19
  |line 16, col 15 to line 16, col 17 of module CoverageStatistics: 19
  line 17, col 9 to line 17, col 17 of module CoverageStatistics: 19
  line 18, col 9 to line 18, col 19 of module CoverageStatistics: 19
<B line 20, col 1 to line 20, col 1 of module CoverageStatistics>: 0:19
  line 20, col 9 to line 20, col 17 of module CoverageStatistics: 38
  |line 20, col 9 to line 20, col 9 of module CoverageStatistics: 19
  |line 20, col 15 to line 20, col 17 of module CoverageStatistics: 19
  line 21, col 9 to line 21, col 25 of module CoverageStatistics: 19
<C line 26, col 1 to line 26, col 1 of module CoverageStatistics>: 0:0
  line 26, col 9 to line 26, col 14 of module CoverageStatistics: 19
  line 27, col 9 to line 27, col 17 of module CoverageStatistics: 0
  line 28, col 9 to line 28, col 18 of module CoverageStatistics: 0
<U1 line 30, col 1 to line 30, col 2 of module CoverageStatistics>: 0:0
  line 30, col 7 to line 30, col 11 of module CoverageStatistics: 19
  line 30, col 16 to line 30, col 29 of module CoverageStatistics: 0
<U2 line 32, col 1 to line 32, col 2 of module CoverageStatistics>: 0:0
  line 32, col 7 to line 32, col 11 of module CoverageStatistics: 19
  line 32, col 16 to line 32, col 32 of module CoverageStatistics: 0
<U3 line 34, col 1 to line 34, col 2 of module CoverageStatistics>: 0:0
  line 34, col 7 to line 34, col 11 of module CoverageStatistics: 19
  line 34, col 16 to line 34, col 26 of module CoverageStatistics: 0
  line 34, col 31 to line 34, col 41 of module CoverageStatistics: 0
<U4 line 36, col 1 to line 36, col 2 of module CoverageStatistics>: 0:0
  line 36, col 7 to line 36, col 11 of module CoverageStatistics: 19
  line 36, col 16 to line 36, col 26 of module CoverageStatistics: 0
  line 36, col 31 to line 36, col 41 of module CoverageStatistics: 0
<UC1 line 38, col 1 to line 38, col 3 of module CoverageStatistics>: 0:19
  line 38, col 8 to line 38, col 16 of module CoverageStatistics: 38
  |line 38, col 8 to line 38, col 8 of module CoverageStatistics: 19
  |line 38, col 14 to line 38, col 16 of module CoverageStatistics: 19
  line 38, col 21 to line 38, col 37 of module CoverageStatistics: 19
<UC2 line 40, col 1 to line 40, col 3 of module CoverageStatistics>: 0:19
  line 40, col 8 to line 40, col 16 of module CoverageStatistics: 38
  |line 40, col 8 to line 40, col 8 of module CoverageStatistics: 19
  |line 40, col 14 to line 40, col 16 of module CoverageStatistics: 19
  line 40, col 21 to line 40, col 31 of module CoverageStatistics: 19
  line 40, col 36 to line 40, col 46 of module CoverageStatistics: 19
<UC3 line 42, col 1 to line 42, col 3 of module CoverageStatistics>: 0:19
  line 42, col 8 to line 42, col 16 of module CoverageStatistics: 38
  |line 42, col 8 to line 42, col 8 of module CoverageStatistics: 19
  |line 42, col 14 to line 42, col 16 of module CoverageStatistics: 19
  line 42, col 21 to line 42, col 34 of module CoverageStatistics: 19
<Constraint line 48, col 1 to line 48, col 10 of module CoverageStatistics>: 97:98
  line 48, col 15 to line 48, col 20 of module CoverageStatistics: 98`)
}

// Port of CCoverageTest.testSpec, including invariant coverage and set costs.
func TestJavaCCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "C", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"17"}) {
		t.Errorf("search depth=%v, want 17", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"253", "20", "0"}) {
		t.Errorf("stats=%v, want 253/20/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 14, col 1 to line 14, col 4 of module C>: 3:3
  line 14, col 12 to line 14, col 21 of module C: 1
  line 15, col 12 to line 15, col 16 of module C: 3
<A line 17, col 1 to line 17, col 1 of module C>: 16:20
  line 17, col 9 to line 17, col 17 of module C: 40
  |line 17, col 9 to line 17, col 9 of module C: 20
  |line 17, col 15 to line 17, col 17 of module C: 20
  line 18, col 9 to line 18, col 17 of module C: 40
  |line 18, col 9 to line 18, col 9 of module C: 20
  |line 18, col 15 to line 18, col 17 of module C: 20
  line 19, col 9 to line 19, col 17 of module C: 20
  line 20, col 9 to line 20, col 19 of module C: 20
<B line 22, col 1 to line 22, col 1 of module C>: 0:20
  line 22, col 9 to line 22, col 17 of module C: 40
  |line 22, col 9 to line 22, col 9 of module C: 20
  |line 22, col 15 to line 22, col 17 of module C: 20
  line 12, col 30 to line 12, col 48 of module C: 640
  line 12, col 19 to line 12, col 26 of module C: 20
  line 23, col 12 to line 23, col 15 of module C: 20
  line 24, col 9 to line 24, col 25 of module C: 20
<C line 26, col 1 to line 26, col 1 of module C>: 0:0
  line 26, col 9 to line 26, col 14 of module C: 20
  line 27, col 9 to line 27, col 17 of module C: 0
  line 28, col 9 to line 28, col 18 of module C: 0
<D line 30, col 1 to line 30, col 1 of module C>: 1:210
  line 30, col 6 to line 30, col 16 of module C: 210
  |line 30, col 13 to line 30, col 16 of module C: 20
  line 30, col 21 to line 30, col 31 of module C: 210
<U1 line 32, col 1 to line 32, col 2 of module C>: 0:0
  line 32, col 7 to line 32, col 11 of module C: 20
  line 32, col 16 to line 32, col 29 of module C: 0
<U2 line 34, col 1 to line 34, col 2 of module C>: 0:0
  line 34, col 7 to line 34, col 11 of module C: 20
  line 34, col 16 to line 34, col 32 of module C: 0
<U3 line 36, col 1 to line 36, col 2 of module C>: 0:0
  line 36, col 7 to line 36, col 11 of module C: 20
  line 36, col 16 to line 36, col 26 of module C: 0
  line 36, col 31 to line 36, col 41 of module C: 0
<Inv line 48, col 1 to line 48, col 3 of module C>
  line 48, col 8 to line 54, col 26 of module C: 21
  |line 48, col 11 to line 48, col 19 of module C: 21
  |line 49, col 11 to line 49, col 19 of module C: 21
  |line 50, col 11 to line 50, col 22 of module C: 21
  ||line 46, col 17 to line 46, col 64 of module C: 21
  |||line 46, col 42 to line 46, col 64 of module C: 5376
  ||||line 46, col 55 to line 46, col 64 of module C: 21504
  ||||line 46, col 51 to line 46, col 51 of module C: 5376
  |||line 46, col 26 to line 46, col 38 of module C: 21:26880
  ||||line 46, col 34 to line 46, col 37 of module C: 21
  |line 51, col 23 to line 51, col 26 of module C: 21
  |line 52, col 14 to line 54, col 26 of module C: 21
  ||line 52, col 17 to line 52, col 27 of module C: 21
  ||line 53, col 17 to line 53, col 40 of module C: 21
  |||line 53, col 32 to line 53, col 40 of module C: 105
  |||line 53, col 26 to line 53, col 29 of module C: 21
  ||line 54, col 17 to line 54, col 26 of module C: 21
<Inv2 line 56, col 1 to line 56, col 4 of module C>
  line 56, col 9 to line 62, col 27 of module C: 21
  |line 56, col 12 to line 56, col 20 of module C: 21
  |line 57, col 12 to line 57, col 20 of module C: 21
  |line 58, col 12 to line 58, col 23 of module C: 21
  ||line 46, col 17 to line 46, col 64 of module C: 21
  |||line 46, col 42 to line 46, col 64 of module C: 10752
  ||||line 46, col 55 to line 46, col 64 of module C: 48384
  ||||line 46, col 51 to line 46, col 51 of module C: 10752
  |||line 46, col 26 to line 46, col 38 of module C: 21:59136
  ||||line 46, col 34 to line 46, col 37 of module C: 21
  |line 59, col 24 to line 59, col 27 of module C: 21
  |line 60, col 15 to line 62, col 27 of module C: 21
  ||line 60, col 15 to line 60, col 28 of module C: 21
  ||line 61, col 17 to line 62, col 27 of module C: 21
  |||line 61, col 32 to line 62, col 27 of module C: 105
  |||line 61, col 26 to line 61, col 29 of module C: 21
<Constraint line 42, col 1 to line 42, col 10 of module C>: 252:253
  line 42, col 15 to line 42, col 20 of module C: 253`)
}

// Port of BCoverageTest.testSpec with the original model and records.
func TestJavaBCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "B", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"2"}) {
		t.Errorf("search depth=%v, want 2", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"5", "2", "0"}) {
		t.Errorf("stats=%v, want 5/2/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 4, col 1 to line 4, col 4 of module B>: 1:1
  line 4, col 9 to line 4, col 17 of module B: 1
<A line 8, col 1 to line 8, col 1 of module B>: 1:2
  line 8, col 6 to line 8, col 19 of module B: 2
<B line 10, col 1 to line 10, col 1 of module B>: 0:2
  line 10, col 6 to line 10, col 19 of module B: 2`)
}

// Port of DCoverageTest.testSpec with the original model and records.
func TestJavaDCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "D", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"2"}) {
		t.Errorf("search depth=%v, want 2", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"7", "3", "0"}) {
		t.Errorf("stats=%v, want 7/3/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 5, col 1 to line 5, col 4 of module D>: 1:1
  line 5, col 9 to line 5, col 13 of module D: 1
<A line 11, col 1 to line 11, col 1 of module D>: 1:3
  line 11, col 6 to line 11, col 17 of module D: 3
  |line 11, col 11 to line 11, col 17 of module D: 3
  ||line 9, col 12 to line 9, col 47 of module D: 9
  |||line 9, col 15 to line 9, col 19 of module D: 9
  |||line 9, col 33 to line 9, col 47 of module D: 6
<B line 13, col 1 to line 13, col 1 of module D>: 1:3
  line 13, col 6 to line 13, col 17 of module D: 3
  |line 13, col 11 to line 13, col 17 of module D: 3
  ||line 9, col 12 to line 9, col 47 of module D: 27
  |||line 9, col 15 to line 9, col 19 of module D: 27
  |||line 9, col 33 to line 9, col 47 of module D: 24`)
}

// Port of ECoverageTest.testSpec with the original model and records.
func TestJavaECoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "E", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"2"}) {
		t.Errorf("search depth=%v, want 2", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"73", "9", "0"}) {
		t.Errorf("stats=%v, want 73/9/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 10, col 1 to line 10, col 4 of module E>: 1:1
  line 10, col 9 to line 10, col 13 of module E: 1
<Next line 12, col 1 to line 12, col 4 of module E>: 8:72
  line 12, col 9 to line 12, col 20 of module E: 72
  |line 12, col 16 to line 12, col 20 of module E: 9`)
}

// Port of FCoverageTest.testSpec with the original model and records.
func TestJavaFCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "F", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"1"}) {
		t.Errorf("search depth=%v, want 1", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"4", "2", "0"}) {
		t.Errorf("stats=%v, want 4/2/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 8, col 1 to line 8, col 4 of module F>: 2:2
  line 8, col 9 to line 8, col 79 of module F: 2
  |line 8, col 15 to line 8, col 79 of module F: 1
  ||line 6, col 25 to line 6, col 56 of module F: 1:3
  |||line 6, col 37 to line 6, col 54 of module F: 5
  ||||line 6, col 37 to line 6, col 40 of module F: 5
  ||||line 6, col 45 to line 6, col 54 of module F: 4
  |||||line 6, col 47 to line 6, col 47 of module F: 4
  |||||line 6, col 50 to line 6, col 53 of module F: 2
  |||||line 8, col 63 to line 8, col 78 of module F: 4
  ||||||line 8, col 63 to line 8, col 73 of module F: 4
  ||||||line 8, col 78 to line 8, col 78 of module F: 2
  |||line 6, col 33 to line 6, col 33 of module F: 1
  ||line 8, col 18 to line 8, col 28 of module F: 1:6
  ||line 8, col 41 to line 8, col 45 of module F: 5
  ||line 8, col 63 to line 8, col 78 of module F: 4
  |||line 8, col 63 to line 8, col 73 of module F: 4
  |||line 8, col 78 to line 8, col 78 of module F: 2
<Next line 10, col 1 to line 10, col 4 of module F>: 0:2
  line 10, col 9 to line 10, col 19 of module F: 2`)
}

// Port of GCoverageTest.testSpec with the original model and records.
func TestJavaGCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "G", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"1"}) {
		t.Errorf("search depth=%v, want 1", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"2", "1", "0"}) {
		t.Errorf("stats=%v, want 2/1/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 6, col 1 to line 6, col 4 of module G>: 1:1
  line 6, col 9 to line 6, col 20 of module G: 1
<Next line 8, col 1 to line 8, col 4 of module G>: 0:1
  line 8, col 12 to line 8, col 25 of module G: 1
  line 9, col 12 to line 9, col 27 of module G: 2
  line 10, col 12 to line 10, col 23 of module G: 2
<Prop line 12, col 1 to line 12, col 4 of module G>
  line 12, col 9 to line 12, col 20 of module G: 1
  |line 8, col 12 to line 8, col 25 of module G: 1
  |line 9, col 12 to line 9, col 27 of module G: 1
  |line 10, col 12 to line 10, col 23 of module G: 1`)
}

// Port of HCoverageTest.testSpec with the original model and records.
func TestJavaHCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "H", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"3"}) {
		t.Errorf("search depth=%v, want 3", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"26", "6", "0"}) {
		t.Errorf("stats=%v, want 26/6/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Inv line 22, col 1 to line 22, col 3 of module H>: 1:31
  line 22, col 11 to line 22, col 16 of module H: 1
  line 23, col 11 to line 23, col 15 of module H: 31
<A line 11, col 1 to line 11, col 1 of module H>: 1:6
  line 11, col 6 to line 11, col 12 of module H: 6
<BandC line 13, col 1 to line 13, col 5 of module H (13 13 14 28)>: 2:15
  line 13, col 16 to line 13, col 26 of module H: 11
  |line 13, col 16 to line 13, col 16 of module H: 6
  |line 13, col 22 to line 13, col 26 of module H: 6
  line 14, col 16 to line 14, col 28 of module H: 15
  |line 14, col 23 to line 14, col 28 of module H: 5
<BandC line 13, col 1 to line 13, col 5 of module H (15 13 16 22)>: 1:3
  line 15, col 16 to line 15, col 21 of module H: 3
  |line 15, col 16 to line 15, col 16 of module H: 6
  line 16, col 16 to line 16, col 22 of module H: 3
<DandE line 18, col 1 to line 18, col 5 of module H (18 11 18 27)>: 1:1
  line 18, col 11 to line 18, col 16 of module H: 7
  |line 18, col 11 to line 18, col 11 of module H: 6
  line 18, col 21 to line 18, col 27 of module H: 1
<DandE line 18, col 1 to line 18, col 5 of module H (18 34 18 53)>: 0:0
  line 18, col 34 to line 18, col 40 of module H: 6
  line 18, col 45 to line 18, col 53 of module H: 0`)
}

// Port of ICoverageTest.testSpec with the original model and records.
func TestJavaICoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "I", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"1"}) {
		t.Errorf("search depth=%v, want 1", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"20", "5", "0"}) {
		t.Errorf("stats=%v, want 20/5/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Action line 11, col 9 to line 11, col 25 of module I>: 5:5
  line 11, col 9 to line 11, col 25 of module I: 5
  |line 11, col 15 to line 11, col 25 of module I: 1:6
<Action line 11, col 52 to line 11, col 55 of module I>: 0:15
  line 11, col 52 to line 11, col 52 of module I: 15
  |line 8, col 1 to line 9, col 22 of module I: 15
  ||line 9, col 8 to line 9, col 22 of module I: 15
  ||line 8, col 9 to line 8, col 15 of module I: 15
  line 11, col 54 to line 11, col 54 of module I: 15
<Inv line 13, col 1 to line 13, col 3 of module I>
  line 13, col 8 to line 13, col 34 of module I: 5
  |line 13, col 27 to line 13, col 34 of module I: 15
  |line 13, col 17 to line 13, col 24 of module I: 5`)
}

// Port of JCoverageTest.testSpec with the original model and records.
func TestJavaJCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "J", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"2"}) {
		t.Errorf("search depth=%v, want 2", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"11", "6", "0"}) {
		t.Errorf("stats=%v, want 11/6/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 5, col 1 to line 5, col 4 of module J>: 5:5
  line 5, col 12 to line 5, col 28 of module J: 1:6
  line 6, col 12 to line 6, col 40 of module J: 5
  |line 6, col 16 to line 6, col 40 of module J: 5:25
  ||line 6, col 23 to line 6, col 33 of module J: 5:30
<Next line 12, col 1 to line 12, col 4 of module J>: 1:6
  line 12, col 12 to line 12, col 22 of module J: 6
  |line 12, col 17 to line 12, col 22 of module J: 6
  ||line 9, col 2 to line 10, col 19 of module J: 6
  |||line 9, col 7 to line 9, col 11 of module J: 6
  |||line 9, col 16 to line 9, col 40 of module J: 2
  |||line 10, col 7 to line 10, col 19 of module J: 4
  ||line 12, col 21 to line 12, col 21 of module J: 6
  line 13, col 12 to line 13, col 22 of module J: 6`)
}

func TestJavaKCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "K", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"1"}) {
		t.Errorf("search depth=%v, want 1", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"2", "1", "0"}) {
		t.Errorf("stats=%v, want 2/1/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
}

func TestJavaLCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "L", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"1"}) {
		t.Errorf("search depth=%v, want 1", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"2", "1", "0"}) {
		t.Errorf("stats=%v, want 2/1/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
}

func TestJavaMCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "M", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"1"}) {
		t.Errorf("search depth=%v, want 1", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"37056", "192", "0"}) {
		t.Errorf("stats=%v, want 37056/192/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 16, col 1 to line 16, col 4 of module M>: 192:192
  line 17, col 17 to line 17, col 39 of module M: 1:24
  |line 17, col 18 to line 17, col 27 of module M: 1
  |line 17, col 32 to line 17, col 38 of module M: 1
  line 18, col 16 to line 18, col 31 of module M: 8:192
  |line 18, col 17 to line 18, col 21 of module M: 8
  |line 18, col 26 to line 18, col 30 of module M: 8
  line 19, col 6 to line 19, col 19 of module M: 64
  line 20, col 6 to line 20, col 21 of module M: 192
<Next line 22, col 1 to line 22, col 4 of module M>: 0:36864
  line 23, col 6 to line 23, col 40 of module M: 1536
  |line 23, col 18 to line 23, col 40 of module M: 192:4608
  ||line 23, col 19 to line 23, col 28 of module M: 192
  ||line 23, col 33 to line 23, col 39 of module M: 192
  line 24, col 6 to line 24, col 32 of module M: 12288
  |line 24, col 17 to line 24, col 32 of module M: 1536:36864
  ||line 24, col 18 to line 24, col 22 of module M: 1536
  ||line 24, col 27 to line 24, col 31 of module M: 1536
  line 25, col 6 to line 25, col 20 of module M: 36864
  |line 25, col 16 to line 25, col 20 of module M: 12288
  line 26, col 6 to line 26, col 22 of module M: 36864`)
}

func TestJavaOCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "O", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"4"}) {
		t.Errorf("search depth=%v, want 4", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"4", "4", "0"}) {
		t.Errorf("stats=%v, want 4/4/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 24, col 1 to line 24, col 4 of module O>: 1:1
  line 24, col 9 to line 24, col 13 of module O: 1
<I!Step line 22, col 1 to line 22, col 1 of module O>: 3:3
  line 22, col 31 to line 22, col 31 of module O: 7
  line 10, col 15 to line 10, col 19 of module O: 3
  |line 10, col 15 to line 10, col 15 of module O: 4
  line 11, col 15 to line 11, col 24 of module O: 3
<Inv line 28, col 1 to line 28, col 3 of module O>
  line 28, col 8 to line 28, col 14 of module O: 4`)
}

func TestJavaGithub314Coverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "Github314", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"3"}) {
		t.Errorf("search depth=%v, want 3", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"3", "3", "0"}) {
		t.Errorf("stats=%v, want 3/3/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
}

func TestJavaGithub377Coverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "Github377", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"1"}) {
		t.Errorf("search depth=%v, want 1", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"2", "1", "0"}) {
		t.Errorf("stats=%v, want 2/1/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Action line 5, col 9 to line 5, col 16 of module Github377>: 1:1
  line 5, col 9 to line 5, col 16 of module Github377: 1
<Action line 5, col 24 to line 5, col 34 of module Github377>: 0:1
  line 5, col 24 to line 5, col 34 of module Github377: 1
<1Inv line 9, col 1 to line 9, col 4 of module Github377>
  line 10, col 6 to line 10, col 69 of module Github377: 1
  line 11, col 5 to line 11, col 13 of module Github377: 1
<1InvNonRec line 15, col 1 to line 15, col 10 of module Github377>
  line 16, col 6 to line 16, col 40 of module Github377: 1
  line 17, col 5 to line 17, col 13 of module Github377: 1
<2Inv line 21, col 1 to line 21, col 4 of module Github377>
  line 25, col 10 to line 25, col 47 of module Github377: 1
  |line 25, col 10 to line 25, col 34 of module Github377: 1
  |line 25, col 39 to line 25, col 47 of module Github377: 1
  ||line 25, col 39 to line 25, col 44 of module Github377: 1
  |||line 23, col 11 to line 23, col 70 of module Github377: 1
  ||||line 23, col 34 to line 23, col 70 of module Github377: 2
  ||||line 23, col 25 to line 23, col 27 of module Github377: 1
  line 26, col 7 to line 26, col 11 of module Github377: 1
<2aInv line 28, col 1 to line 28, col 5 of module Github377>
  line 32, col 10 to line 33, col 21 of module Github377: 1
  |line 32, col 13 to line 32, col 21 of module Github377: 1
  |line 33, col 13 to line 33, col 21 of module Github377: 1
  ||line 33, col 13 to line 33, col 18 of module Github377: 1
  |||line 31, col 11 to line 31, col 42 of module Github377: 1
  ||||line 31, col 34 to line 31, col 42 of module Github377: 1
  |||||line 31, col 34 to line 31, col 39 of module Github377: 1
  ||||||line 30, col 11 to line 30, col 70 of module Github377: 1
  |||||||line 30, col 34 to line 30, col 70 of module Github377: 2
  |||||||line 30, col 25 to line 30, col 27 of module Github377: 1
  |||||line 31, col 41 to line 31, col 41 of module Github377: 1
  ||||line 31, col 25 to line 31, col 27 of module Github377: 1
  line 34, col 7 to line 34, col 11 of module Github377: 1
<2bInv line 36, col 1 to line 36, col 5 of module Github377>
  line 40, col 10 to line 40, col 31 of module Github377: 1
  |line 40, col 10 to line 40, col 18 of module Github377: 1
  |line 40, col 23 to line 40, col 31 of module Github377: 1
  ||line 40, col 23 to line 40, col 28 of module Github377: 1
  |||line 39, col 11 to line 39, col 42 of module Github377: 1
  ||||line 39, col 34 to line 39, col 42 of module Github377: 1
  |||||line 39, col 34 to line 39, col 39 of module Github377: 1
  ||||||line 38, col 11 to line 38, col 70 of module Github377: 1
  |||||||line 38, col 34 to line 38, col 70 of module Github377: 2
  |||||||line 38, col 25 to line 38, col 27 of module Github377: 1
  |||||line 39, col 41 to line 39, col 41 of module Github377: 1
  ||||line 39, col 25 to line 39, col 27 of module Github377: 1
  line 41, col 7 to line 41, col 11 of module Github377: 1
<3aInv line 44, col 1 to line 44, col 5 of module Github377>
  line 48, col 10 to line 50, col 21 of module Github377: 1
  |line 49, col 16 to line 49, col 21 of module Github377: 1
  |line 50, col 13 to line 50, col 21 of module Github377: 1
  ||line 50, col 13 to line 50, col 18 of module Github377: 1
  |||line 47, col 11 to line 47, col 42 of module Github377: 1
  ||||line 47, col 34 to line 47, col 42 of module Github377: 1
  |||||line 47, col 34 to line 47, col 39 of module Github377: 1
  ||||||line 46, col 11 to line 46, col 70 of module Github377: 1
  |||||||line 46, col 34 to line 46, col 70 of module Github377: 2
  |||||||line 46, col 25 to line 46, col 27 of module Github377: 1
  ||||line 47, col 25 to line 47, col 27 of module Github377: 1
  line 51, col 7 to line 51, col 11 of module Github377: 1
<3bInv line 55, col 1 to line 55, col 5 of module Github377>
  line 58, col 11 to line 58, col 42 of module Github377: 1
  line 57, col 11 to line 57, col 70 of module Github377: 1
  |line 57, col 34 to line 57, col 70 of module Github377: 2
  |line 57, col 25 to line 57, col 27 of module Github377: 1
  line 60, col 13 to line 60, col 31 of module Github377: 1
  line 61, col 7 to line 61, col 11 of module Github377: 1
<4Inv line 66, col 1 to line 66, col 4 of module Github377>
  line 71, col 11 to line 71, col 18 of module Github377: 1
  line 72, col 7 to line 72, col 11 of module Github377: 1`)
}

func TestJavaGithub649Coverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "Github649", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"2"}) {
		t.Errorf("search depth=%v, want 2", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"3", "2", "0"}) {
		t.Errorf("stats=%v, want 3/2/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 14, col 1 to line 14, col 4 of module Github649>: 1:1
  line 12, col 21 to line 12, col 25 of module Github649: 2
  line 12, col 32 to line 12, col 37 of module Github649: 1
  line 12, col 44 to line 12, col 60 of module Github649: 1
  line 14, col 21 to line 14, col 25 of module Github649: 1
<Next line 16, col 1 to line 16, col 4 of module Github649>: 1:2
  line 16, col 9 to line 16, col 23 of module Github649: 2
<TypeOK line 9, col 1 to line 9, col 6 of module Github649>
  line 9, col 11 to line 9, col 24 of module Github649: 2
  |line 7, col 15 to line 7, col 62 of module Github649: 4
  ||line 7, col 18 to line 7, col 22 of module Github649: 4
  ||line 7, col 29 to line 7, col 41 of module Github649: 2
  ||line 7, col 48 to line 7, col 62 of module Github649: 2
  |line 9, col 17 to line 9, col 21 of module Github649: 2
<Constraint line 18, col 1 to line 18, col 10 of module Github649>: 3:3
  line 19, col 5 to line 19, col 26 of module Github649: 3
  |line 12, col 18 to line 12, col 60 of module Github649: 9
  ||line 12, col 21 to line 12, col 25 of module Github649: 9
  ||line 12, col 32 to line 12, col 37 of module Github649: 3
  ||line 12, col 44 to line 12, col 60 of module Github649: 6
  |line 19, col 10 to line 19, col 14 of module Github649: 3
  |line 19, col 17 to line 19, col 22 of module Github649: 3`)
}

func TestJavaImpliedCoverage(t *testing.T) {
	result := runJavaTLCModelTest(t, "Implied", "-coverage", "9999")
	if result.ExitStatus != tlc.ExitStatusSuccess {
		t.Fatalf("exit status=%d; messages=%+v", result.ExitStatus, result.Messages)
	}
	if len(javaTLCRecords(result, tlc.ECTLCFinished)) == 0 {
		t.Fatal("TLC_FINISHED not recorded")
	}
	depth := javaTLCRecords(result, tlc.ECTLCSearchDepth)
	if len(depth) == 0 || !reflect.DeepEqual(depth[0].Params, []string{"3"}) {
		t.Errorf("search depth=%v, want 3", depth)
	}
	stats := javaTLCRecords(result, tlc.ECTLCStats)
	if len(stats) == 0 || !reflect.DeepEqual(stats[0].Params, []string{"3", "3", "0"}) {
		t.Errorf("stats=%v, want 3/3/0", stats)
	}
	for _, code := range []int{tlc.ECGeneral, tlc.ECTLCCoverageMismatch} {
		if records := javaTLCRecords(result, code); len(records) != 0 {
			t.Errorf("unexpected records: %v", records)
		}
	}
	assertJavaCoverage(t, result, `<Init line 6, col 1 to line 6, col 4 of module Implied>: 1:1
  line 6, col 9 to line 6, col 13 of module Implied: 1
<Next line 7, col 1 to line 7, col 4 of module Implied>: 2:2
  line 7, col 9 to line 7, col 13 of module Implied: 2
  |line 7, col 9 to line 7, col 9 of module Implied: 3
  line 7, col 18 to line 7, col 27 of module Implied: 2
<Action line 10, col 17 to line 10, col 25 of module Implied>
  line 10, col 17 to line 10, col 25 of module Implied: 1
<Action line 12, col 31 to line 12, col 48 of module Implied>
  line 12, col 31 to line 12, col 48 of module Implied: 2`)
}
