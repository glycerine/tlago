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
// Ports ACoverageTest.testSpec and CommonTestCase.assertCoverage after the
// coverage implementation. Fixtures and expected records retain upstream data.
package tlago

import (
	"reflect"
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
	if !reflect.DeepEqual(actualValues, expectedValues) {
		t.Errorf("coverage=%v, want %v", actualValues, expectedValues)
	}
	if !reflect.DeepEqual(actualCost, expectedCost) {
		t.Errorf("cost coverage=%v, want %v", actualCost, expectedCost)
	}
	if !reflect.DeepEqual(actualActions, expectedActions) {
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
