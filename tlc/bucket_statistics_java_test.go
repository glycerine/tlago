/*******************************************************************************
 * Copyright (c) 2015 Microsoft Research. All rights reserved.
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
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINbucketStatistic IN THE SOFTWARE.
 *
 * Contributors:
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/

package tlc

import (
	"math"
	"testing"
)

// Mirrors the original test's IBucketStatistics reference across every class row.
type javaBucketStatistics interface {
	AddSample(int)
	GetMean() float64
	GetMedian() int
	GetMin() int
	GetMax() int
	GetStdDev() float64
	GetPercentile(float64) float64
	GetObservations() int64
	String() string
}

// Double.compare(expected, actual) == 0 includes signed-zero and canonical NaN.
func javaBucketDoubleEquals(t *testing.T, expected, actual float64) {
	t.Helper()
	if math.IsNaN(expected) && math.IsNaN(actual) {
		return
	}
	if math.Float64bits(expected) != math.Float64bits(actual) {
		t.Fatalf("Double.compare(%v, %v) != 0", expected, actual)
	}
}
func javaBucketIntEquals(t *testing.T, expected, actual int64) {
	t.Helper()
	if expected != actual {
		t.Fatalf("got %d, want %d", actual, expected)
	}
}
func javaBucketIllegalArgument(t *testing.T, message string, body func()) {
	t.Helper()
	defer func() {
		failure := recover()
		if failure == nil {
			t.Fatal(message)
		}
		if _, ok := failure.(*IllegalArgumentException); !ok {
			panic(failure)
		}
	}()
	body()
}

// Complete BucketStatisticsTest: ten original methods on both parameter rows.
func TestJavaBucketStatistics(t *testing.T) {
	rows := []struct {
		name string
		make func() javaBucketStatistics
	}{
		{"ConcurrentBucketStatistics", func() javaBucketStatistics { return NewConcurrentBucketStatistics("BucketStatisticsTest") }},
		{"BucketStatistics", func() javaBucketStatistics { return NewBucketStatistics("BucketStatisticsTest") }},
	}
	tests := []struct {
		name string
		run  func(*testing.T, javaBucketStatistics)
	}{
		{"testInvalidArgument", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIllegalArgument(t, "fail()", func() { s.AddSample(-1) })
		}},
		{"testMean", func(t *testing.T, s javaBucketStatistics) {
			javaBucketDoubleEquals(t, -1, s.GetMean())
			s.AddSample(0)
			javaBucketDoubleEquals(t, 0, s.GetMean())
			s.AddSample(1)
			s.AddSample(2)
			javaBucketDoubleEquals(t, 1, s.GetMean())
			s.AddSample(2)
			s.AddSample(2)
			javaBucketDoubleEquals(t, 1.4, s.GetMean())
		}},
		{"testMedian", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIntEquals(t, -1, int64(s.GetMedian()))
			s.AddSample(0)
			javaBucketIntEquals(t, 0, int64(s.GetMedian()))
			s.AddSample(1)
			s.AddSample(2)
			javaBucketIntEquals(t, 1, int64(s.GetMedian()))
			s.AddSample(2)
			s.AddSample(2)
			javaBucketIntEquals(t, 2, int64(s.GetMedian()))
		}},
		{"testMin", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIntEquals(t, -1, int64(s.GetMin()))
			for _, amount := range []int{0, 0, 0, 1, 1, 2, 2, 2} {
				s.AddSample(amount)
			}
			javaBucketIntEquals(t, 0, int64(s.GetMin()))
		}},
		{"testMin2", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIntEquals(t, -1, int64(s.GetMin()))
			for _, amount := range []int{1, 1, 2, 2, 2} {
				s.AddSample(amount)
			}
			javaBucketIntEquals(t, 1, int64(s.GetMin()))
		}},
		{"testMax", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIntEquals(t, -1, int64(s.GetMax()))
			for _, amount := range []int{0, 0, 0, 1, 1, 2, 2, 2, 2, 3} {
				s.AddSample(amount)
			}
			javaBucketIntEquals(t, 3, int64(s.GetMax()))
		}},
		{"testStandardDeviation", func(t *testing.T, s javaBucketStatistics) {
			// Original JUnit assertEquals(-1.0, value, 0).
			if got := s.GetStdDev(); got != -1 {
				t.Fatalf("got %v, want -1.0 with delta 0", got)
			}
			for _, amount := range []int{0, 0, 0, 1, 1, 2, 2, 2, 2, 3} {
				s.AddSample(amount)
			}
			// Math.round returns long, converted to double by division by 10000d.
			rounded := float64(javaDoubleToLong(math.Floor(s.GetStdDev()*10000+0.5))) / 10000
			javaBucketDoubleEquals(t, 1.005, rounded)
		}},
		{"testGetPercentile", func(t *testing.T, s javaBucketStatistics) {
			if got := s.GetPercentile(1); got != -1 {
				t.Fatalf("got %v, want -1.0 with delta 0", got)
			}
			// The original catches any Exception and fails, so an escaping Go panic
			// also fails this test. No assertion is added to these two clamp calls.
			s.AddSample(1)
			s.GetPercentile(1.1)
			s.GetPercentile(-0.1)
			for _, amount := range []int{1, 1, 2, 2, 2, 3} {
				s.AddSample(amount)
			}
			javaBucketDoubleEquals(t, 2, s.GetPercentile(0.5))
			javaBucketDoubleEquals(t, 2, s.GetPercentile(0.5))
			javaBucketDoubleEquals(t, 2, s.GetPercentile(0.75))
			javaBucketDoubleEquals(t, 3, s.GetPercentile(0.999))
		}},
		{"testGetPercentileNaN", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIllegalArgument(t, "Parameter not a number", func() { s.GetPercentile(math.NaN()) })
		}},
		{"testToString", func(t *testing.T, s javaBucketStatistics) {
			// Invoke only, as in the original; no output assertions are added.
			s.String()
			for _, amount := range []int{0, 0, 0, 1, 1, 2, 2, 2, 2, 3} {
				s.AddSample(amount)
			}
			s.String()
		}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) { test.run(t, row.make()) })
			}
		})
	}
}

// Complete FixedSizedBucketStatisticsTest: all six methods and both size-8 rows.
func TestJavaFixedSizedBucketStatistics(t *testing.T) {
	rows := []struct {
		name string
		make func() javaBucketStatistics
	}{
		{"FixedSizedConcurrentBucketStatistics", func() javaBucketStatistics {
			return NewFixedSizedConcurrentBucketStatistics("FixedSizedBucketStatisticsTest", 8)
		}},
		{"FixedSizedBucketStatistics", func() javaBucketStatistics { return NewFixedSizedBucketStatistics("FixedSizedBucketStatisticsTest", 8) }},
	}
	tests := []struct {
		name string
		run  func(*testing.T, javaBucketStatistics)
	}{
		{"testMin", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIntEquals(t, -1, int64(s.GetMin()))
			for _, amount := range []int{0, 1, 1, 2, 2, 2} {
				s.AddSample(amount)
			}
			javaBucketIntEquals(t, 0, int64(s.GetMin()))
		}},
		{"testMin2", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIntEquals(t, -1, int64(s.GetMin()))
			for _, amount := range []int{1, 1, 2, 2, 2} {
				s.AddSample(amount)
			}
			javaBucketIntEquals(t, 1, int64(s.GetMin()))
		}},
		{"testMax", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIntEquals(t, -1, int64(s.GetMax()))
			for _, amount := range []int{0, 0, 0, 1, 1, 2, 2, 2, 2, 3} {
				s.AddSample(amount)
			}
			javaBucketIntEquals(t, 3, int64(s.GetMax()))
		}},
		{"testInvalidArgument", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIllegalArgument(t, "fail()", func() { s.AddSample(-1) })
		}},
		{"testGetPercentileNaN", func(t *testing.T, s javaBucketStatistics) {
			javaBucketIllegalArgument(t, "Parameter not a number", func() { s.GetPercentile(math.NaN()) })
		}},
		{"testMaximum", func(t *testing.T, s javaBucketStatistics) {
			s.AddSample(16)
			s.AddSample(16)
			s.AddSample(16)
			javaBucketIntEquals(t, 7, int64(s.GetMax()))
			javaBucketIntEquals(t, 7, int64(s.GetMedian()))
			javaBucketIntEquals(t, 3, s.GetObservations())
		}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) { test.run(t, row.make()) })
			}
		})
	}
}
