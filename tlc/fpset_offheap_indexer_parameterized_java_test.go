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
	"math/bits"
	"testing"
)

// Original OffHeapIndexerParameterizedTest.roundHalfEven and data; duplicate
// rows are retained, and all five inherited methods run for each source row.
func javaOffHeapIndexerRows() []struct {
	positions int64
	fpBits    int
} {
	rows := make([]struct {
		positions int64
		fpBits    int
	}, 0, 21762)
	const multiple int64 = 8 * 1024 * 1024 * 1024
	add := func(positions int64, fpBits int) {
		rows = append(rows, struct {
			positions int64
			fpBits    int
		}{positions, fpBits})
	}
	for fpBits := 1; fpBits > 0; fpBits-- {
		for i := 1; i < 32; i++ {
			for j := 2; j < 32; j++ {
				add(((int64(math.MaxInt64)>>uint(j))/multiple)*multiple+int64(i)*(1<<30), fpBits)
			}
		}
		for k := 1 << 7; k > 0; k-- {
			add(int64(k)*(1<<27), fpBits)
		}
		for posBits := 63 - fpBits; posBits > 16; posBits-- {
			add(int64(1)<<uint(posBits), fpBits)
		}
	}
	return rows
}

// The source subclasses' getIndexer constructs a fresh instance at each call.
// Preserve their exact assumptions rather than selecting the automatic indexer.
func runJavaOffHeapIndexerParameterized(t *testing.T, kind OffHeapIndexerKind) {
	t.Helper()
	for row, input := range javaOffHeapIndexerRows() {
		t.Run(fmt.Sprintf("row%d/positions%d/fpBits%d", row, input.positions, input.fpBits), func(t *testing.T) {
			getIndexer := func(t *testing.T) *OffHeapIndexer {
				t.Helper()
				switch kind {
				case OffHeapIndexerBitshifting:
					if bits.OnesCount64(uint64(input.positions)) != 1 {
						t.Skip("source Assume: positions must have exactly one set bit")
					}
					return NewBitshiftingOffHeapIndexer(input.positions, input.fpBits)
				case OffHeapIndexerMult1024:
					if !OffHeapMult1024IndexerIsSupported(input.positions) {
						t.Skip("source Assume: Mult1024Indexer.isSupported(positions)")
					}
					if bits.TrailingZeros64(uint64(input.positions)) <= input.fpBits {
						t.Skip("source Assume: trailing zeros greater than fpBits")
					}
					return NewMult1024OffHeapIndexer(input.positions, input.fpBits)
				default:
					return NewInfinitePrecisionOffHeapIndexer(input.positions, input.fpBits)
				}
			}
			t.Run("testZero", func(t *testing.T) {
				if got := getIndexer(t).GetIdx(0); got != 0 {
					t.Fatalf("index=%d, want0", got)
				}
			})
			t.Run("testOne", func(t *testing.T) {
				if got := getIndexer(t).GetIdx(1); got != 0 {
					t.Fatalf("index=%d, want0", got)
				}
			})
			t.Run("testLongMin", func(t *testing.T) {
				highFP := ^uint64(0) >> uint(input.fpBits)
				if got := getIndexer(t).GetIdx(highFP); got != input.positions-1 {
					t.Fatalf("index=%d, want%d", got, input.positions-1)
				}
			})
			t.Run("testLongMax", func(t *testing.T) {
				highFP := uint64(math.MaxInt64) >> uint(input.fpBits)
				if got := getIndexer(t).GetIdx(highFP); got != input.positions/2-1 {
					t.Fatalf("index=%d, want%d", got, input.positions/2-1)
				}
			})
			t.Run("testSome", func(t *testing.T) {
				upper := int64(math.MaxInt64) >> uint(input.fpBits)
				const n = 1 << 10
				step := upper / n
				var l int64
				for i := int64(0); i < n && i < upper; i++ {
					h := i * step
					low := getIndexer(t).GetIdx(uint64(l))
					high := getIndexer(t).GetIdx(uint64(h))
					if low > high {
						t.Fatalf("index(%d)=%d > index(%d)=%d", l, low, h, high)
					}
					l = h
				}
			})
		})
	}
}

// Original OffHeapInfPrecisionIndexerParameterizedTest: five inherited methods.
func TestJavaOffHeapInfPrecisionIndexerParameterized(t *testing.T) {
	runJavaOffHeapIndexerParameterized(t, OffHeapIndexerInfinitePrecision)
}

// Original OffHeapMult1024IndexerParameterizedTest: five inherited methods.
func TestJavaOffHeapMult1024IndexerParameterized(t *testing.T) {
	runJavaOffHeapIndexerParameterized(t, OffHeapIndexerMult1024)
}

// Original OffHeapBitshiftingIndexerParameterizedTest: five inherited methods.
func TestJavaOffHeapBitshiftingIndexerParameterized(t *testing.T) {
	runJavaOffHeapIndexerParameterized(t, OffHeapIndexerBitshifting)
}
