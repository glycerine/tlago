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
 *   Markus Alexander Kuppe - initial API and implementation
 ******************************************************************************/
package tlc

import "testing"

// Complete tlc2.util.CombinatoricsTest translation. The original loop limits
// and all literal bit-length/exact-long/decimal assertions are retained.
func TestJavaCombinatorics(t *testing.T) {
	t.Run("testChoose", func(t *testing.T) {
		for n := 0; n < MaxChooseNum; n++ {
			for k := 0; k < MaxChooseNum; k++ {
				choose, binomial := Choose(n, k), Binomial(n, k)
				if choose != binomial {
					t.Fatalf("n=%d, k=%d: choose %d != binomial %d", n, k, choose, binomial)
				}
			}
		}
	})
	t.Run("testChooseBigChoose", func(t *testing.T) {
		for n := 0; n < MaxChooseNum+1; n++ {
			for k := 0; k < MaxChooseNum+1; k++ {
				choose := Choose(n, k)
				bigChoose := BigChoose(n, k)
				if !bigChoose.IsInt64() {
					t.Fatalf("n=%d, k=%d: longValueExact overflow: %s", n, k, bigChoose)
				}
				if choose != bigChoose.Int64() {
					t.Fatalf("n=%d, k=%d: choose %d != bigChoose %s", n, k, choose, bigChoose)
				}
			}
		}
	})
	t.Run("testSlowChooseBigChoose", func(t *testing.T) {
		for n := MaxChooseNum + 1; n < MaxChooseNum<<2; n++ {
			for k := MaxChooseNum + 1; k < MaxChooseNum<<2; k++ {
				slowBigChoose := SlowBigChoose(n, k)
				bigChoose := BigChoose(n, k)
				if slowBigChoose.Cmp(bigChoose) != 0 {
					t.Fatalf("n=%d, k=%d: slowBigChoose %s != bigChoose %s", n, k, slowBigChoose, bigChoose)
				}
			}
		}
	})
	for _, tc := range []struct {
		name            string
		n, k, bitLength int
		exactLong       int64
		decimal         string
	}{
		{"testBigChoose50c1", 50, 1, 6, 50, ""},
		{"testBigChoose50c10", 50, 10, 34, 10272278170, ""},
		{"testBigChoose50c20", 50, 20, 46, 47129212243960, ""},
		{"testBigChoose50c30", 50, 30, 46, 47129212243960, ""},
		{"testBigChoose400c1", 400, 1, 9, 400, ""},
		{"testBigChoose400c50", 400, 50, 214, 0, "17035900270730601418919867558071677342938596450600561760371485120"},
		{"testBigChoose400c100", 400, 100, 321, 0, "2241854791554337561923210387201698554845411177476295990399942258896013007429693894018935107174320"},
		{"testBigChoose400c200", 400, 200, 396, 0, "102952500135414432972975880320401986757210925381077648234849059575923332372651958598336595518976492951564048597506774120"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bigChoose := BigChoose(tc.n, tc.k)
			if bigChoose.BitLen() != tc.bitLength {
				t.Fatalf("bitLength: got %d, want %d", bigChoose.BitLen(), tc.bitLength)
			}
			if tc.decimal != "" {
				if bigChoose.String() != tc.decimal {
					t.Fatalf("decimal: got %s, want %s", bigChoose, tc.decimal)
				}
			} else {
				if !bigChoose.IsInt64() {
					t.Fatalf("longValueExact overflow: %s", bigChoose)
				}
				if bigChoose.Int64() != tc.exactLong {
					t.Fatalf("longValueExact: got %d, want %d", bigChoose.Int64(), tc.exactLong)
				}
			}
		})
	}
}
