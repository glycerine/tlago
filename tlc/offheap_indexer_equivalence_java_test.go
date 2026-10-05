/*******************************************************************************
 * Copyright (c) 2025 Microsoft Research. All rights reserved.
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
package tlc

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"
)

type javaOffHeapIndexerEquivalenceRow struct {
	positions  int64
	fpBits     int
	fpRangeBit int
}

// Whole original data() and roundHalfEven(), preserving duplicate rows and order.
func javaOffHeapIndexerEquivalenceRows() []javaOffHeapIndexerEquivalenceRow {
	rows := make([]javaOffHeapIndexerEquivalenceRow, 0, 21762)
	const multiple int64 = 8 * 1024 * 1024 * 1024
	add := func(positions int64, fpBits, b int) {
		rows = append(rows, javaOffHeapIndexerEquivalenceRow{positions, fpBits, b})
	}
	for fpBits := 1; fpBits > 0; fpBits-- {
		for i := 1; i < 8; i++ {
			for b := 62; b > 0; b-- {
				add(((int64(math.MaxInt64)>>2)/multiple)*multiple+int64(i)*(1<<30), fpBits, b)
			}
		}
		for k := 1 << 6; k > 0; k-- {
			for b := 62; b > 0; b-- {
				add(int64(k)*(1<<27), fpBits, b)
			}
		}
		for posBits := 63 - fpBits; posBits > 16; posBits-- {
			for b := 62; b > 0; b-- {
				add(int64(1)<<uint(posBits), fpBits, b)
			}
		}
	}
	return rows
}

// Whole original non-ignored testInfiniteInfMult and its complete doTest helper.
// The source ThreadLocalRandom is unseeded; Go's runtime-seeded bounded generator
// likewise supplies uniform fingerprints in the original [lowerBound, upperBound).
// testInfiniteBitshifting and testInfMultBitshifting remain @Ignore upstream.
func TestJavaOffHeapIndexerEquivalenceInfiniteInfMult(t *testing.T) {
	for row, input := range javaOffHeapIndexerEquivalenceRows() {
		t.Run(fmt.Sprintf("row%d/positions%d/fpBits%d/fpRangeBit%d", row, input.positions, input.fpBits, input.fpRangeBit), func(t *testing.T) {
			if !OffHeapMult1024IndexerIsSupported(input.positions) {
				t.Skip("source Assume: positions * 8 must be a multiple of 1024MB")
			}
			actual := NewMult1024OffHeapIndexer(input.positions, input.fpBits)
			expected := NewInfinitePrecisionOffHeapIndexer(input.positions, input.fpBits)
			javaOffHeapIndexerEquivalenceCheck(t, input.fpRangeBit, expected, actual)
		})
	}
}

func javaOffHeapIndexerEquivalenceCheck(t *testing.T, fpRangeBit int, expected, actual *OffHeapIndexer) {
	t.Helper()
	upperBound := int64(1) << uint(fpRangeBit)
	const n = 1 << 10
	lowerBound := int64(1) << uint(fpRangeBit-1)
	equal := func(fp int64) {
		t.Helper()
		want, got := expected.GetIdx(uint64(fp)), actual.GetIdx(uint64(fp))
		if got != want {
			t.Fatalf("getIdx(%d)=%d, want %d", fp, got, want)
		}
	}
	for i := int64(0); i < min(int64(n), upperBound-lowerBound); i++ {
		fp := lowerBound + rand.Int64N(upperBound-lowerBound)
		equal(fp)
		if actual.GetIdx(uint64(fp+1)) < actual.GetIdx(uint64(fp)) {
			t.Fatalf("getIdx(%d) < getIdx(%d)", fp+1, fp)
		}
	}
	step := (upperBound - lowerBound) / n
	for i := int64(0); i < min(int64(n), upperBound-lowerBound); i++ {
		fp := lowerBound + i*step
		equal(fp)
		equal(fp + 1)
		if actual.GetIdx(uint64(fp)) > actual.GetIdx(uint64(fp+1)) {
			t.Fatalf("getIdx(%d) > getIdx(%d)", fp, fp+1)
		}
	}
	for l := int64(1); l < n && upperBound+l <= math.MaxInt64; l++ {
		fp := upperBound + l
		equal(fp)
		if actual.GetIdx(uint64(fp)) < actual.GetIdx(uint64(fp-1)) {
			t.Fatalf("getIdx(%d) < getIdx(%d)", fp, fp-1)
		}
	}
	for l := int64(1); l < n && upperBound-l >= 0; l++ {
		fp := upperBound - l
		equal(fp)
		if actual.GetIdx(uint64(fp)) > actual.GetIdx(uint64(fp+1)) {
			t.Fatalf("getIdx(%d) > getIdx(%d)", fp, fp+1)
		}
	}
	equal(upperBound)
}
