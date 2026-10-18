package tlc

import (
	"math/big"
	"testing"
)

var (
	combinatoricsBenchmarkBincoef     = make([]*big.Int, 0, 187489)
	combinatoricsBenchmarkSlowBincoef = make([]*big.Int, 0, 187489)
)

// Port of CombinatoricsBenchmark.bigChoose.
func BenchmarkJavaCombinatoricsBigChoose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for n := MaxChooseNum + 1; n < MaxChooseNum<<3; n++ {
			for k := MaxChooseNum + 1; k < MaxChooseNum<<3; k++ {
				combinatoricsBenchmarkBincoef = append(combinatoricsBenchmarkBincoef, BigChoose(n, k))
			}
		}
	}
}

// Port of CombinatoricsBenchmark.slowBigChoose.
func BenchmarkJavaCombinatoricsSlowBigChoose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for n := MaxChooseNum + 1; n < MaxChooseNum<<3; n++ {
			for k := MaxChooseNum + 1; k < MaxChooseNum<<3; k++ {
				combinatoricsBenchmarkSlowBincoef = append(combinatoricsBenchmarkSlowBincoef, SlowBigChoose(n, k))
			}
		}
	}
}
