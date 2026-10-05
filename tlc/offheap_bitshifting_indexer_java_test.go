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
	"math"
	"math/bits"
	"testing"
)

// Whole original OffHeapBitshiftingIndexerTest.testBitshifting.
func TestJavaOffHeapBitshiftingIndexerBitshifting(t *testing.T) {
	javaBitshiftingIndexerSweep(t, 1, 128, 8, NewBitshiftingOffHeapIndexer(128, 1))
}

// Whole original OffHeapBitshiftingIndexerTest.testBitshifting2.
func TestJavaOffHeapBitshiftingIndexerBitshifting2(t *testing.T) {
	javaBitshiftingIndexerSweep(t, 2, 128, 9, NewBitshiftingOffHeapIndexer(128, 2))
}

// Preserve every position and both boundary fingerprints from the Java doTest helper.
func javaBitshiftingIndexerSweep(t *testing.T, fpBits int, positions int64, logPos int, indexer *OffHeapIndexer) {
	t.Helper()
	if math.Pow(2, float64(logPos-fpBits)) != float64(positions) {
		t.Fatal("power of two does not equal positions")
	}
	if got := bits.LeadingZeros64((uint64(positions) << uint(64-logPos)) - 1); got != fpBits {
		t.Fatalf("leading zeros=%d, want %d", got, fpBits)
	}
	for l := int64(0); l < positions; l++ {
		fp := uint64(l) << uint(64-logPos)
		if got := indexer.GetIdx(fp); got != l {
			t.Fatalf("getIdx(%d)=%d, want %d", fp, got, l)
		}
		fpNext := (uint64(l+1) << uint(64-logPos)) - 1
		if got := indexer.GetIdx(fpNext); got != l {
			t.Fatalf("getIdx(%d)=%d, want %d", fpNext, got, l)
		}
	}
	if got := indexer.GetIdx(uint64(positions) << uint(64-logPos)); got != 0 {
		t.Fatalf("getIdx at wraparound=%d, want 0", got)
	}
}

// Whole original OffHeapBitshiftingIndexerTest.testShift1_268435456.
func TestJavaOffHeapBitshiftingIndexerShift1_268435456(t *testing.T) {
	const fpBits = 1
	const positions int64 = 268435456
	indexer := NewBitshiftingOffHeapIndexer(positions, fpBits)
	if got := indexer.GetIdx(1); got != 0 {
		t.Fatalf("getIdx(1)=%d, want 0", got)
	}
	maxFP := ^uint64(0) >> fpBits
	if got := indexer.GetIdx(maxFP); got != positions-1 {
		t.Fatalf("getIdx(maxFP)=%d, want %d", got, positions-1)
	}
	if got := indexer.GetIdxProbe(maxFP, 1); got != 0 {
		t.Fatalf("getIdx(maxFP, 1)=%d, want 0", got)
	}
	if got := indexer.GetIdxProbe(maxFP, int(int32(positions))+1); got != 0 {
		t.Fatalf("getIdx(maxFP, positions+1)=%d, want 0", got)
	}
}

// Whole original OffHeapBitshiftingIndexerTest.testBitshiftOvershoot.
func TestJavaOffHeapBitshiftingIndexerBitshiftOvershoot(t *testing.T) {
	indexer := NewBitshiftingOffHeapIndexer(536870912, 1)
	if got := indexer.GetIdxProbe(9223371952792813846, 5); got != 0 {
		t.Fatalf("getIdx(9223371952792813846, 5)=%d, want 0", got)
	}
}

// Whole original OffHeapBitshiftingIndexerTest.testNoOverflowErrorBitShifting.
// Catch only the original TLCRuntimeException; other failures must escape.
func TestJavaOffHeapBitshiftingIndexerNoOverflowErrorBitShifting(t *testing.T) {
	defer func() {
		if failure := recover(); failure != nil {
			if exception, ok := failure.(*TLCError); ok && exception.Runtime {
				t.Fatalf("Creation of BitshiftingIndexer threw an exception: %s", exception.Error())
			}
			panic(failure)
		}
	}()
	NewBitshiftingOffHeapIndexer(int64(math.MaxInt32)+1, 1)
}
