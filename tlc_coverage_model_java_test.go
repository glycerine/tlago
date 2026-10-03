/*******************************************************************************
 * Copyright (c) 2018 Microsoft Research. All rights reserved.
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
