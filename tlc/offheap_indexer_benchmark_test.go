// Java source: test-benchmark/tlc2/tool/fp/OffHeapIndexerBenchmark.java
// (Microsoft Corporation, MIT license).
package tlc

import (
	"math"
	"testing"
)

type javaOffHeapIndexerBenchmark struct {
	rng      *JavaRandom
	infinite *OffHeapIndexer
	infMult  *OffHeapIndexer
	bitshift *OffHeapIndexer
}

var offHeapIndexerBenchmarkResult int64

func prepareJavaOffHeapIndexerBenchmark(b *testing.B) *javaOffHeapIndexerBenchmark {
	b.Helper()
	state := &javaOffHeapIndexerBenchmark{
		rng:      NewJavaRandom(0),
		infinite: NewInfinitePrecisionOffHeapIndexer(1<<30, 1),
		infMult:  NewMult1024OffHeapIndexer(1<<30, 1),
		bitshift: NewBitshiftingOffHeapIndexer(1<<30, 1),
	}
	state.rng.SetSeed(0) // @Setup(Level.Trial)
	seed := int64(0)
	seed++
	state.rng.SetSeed(seed) // @Setup(Level.Iteration)
	b.ResetTimer()
	return state
}

// Port of OffHeapIndexerBenchmark.InfMult.
func BenchmarkJavaOffHeapIndexerInfMult(b *testing.B) {
	state := prepareJavaOffHeapIndexerBenchmark(b)
	for i := 0; i < b.N; i++ {
		fp := uint64(state.rng.NextLong()) & uint64(math.MaxInt64)
		offHeapIndexerBenchmarkResult = state.infMult.GetIdx(fp)
	}
}

// Port of OffHeapIndexerBenchmark.Infinite.
func BenchmarkJavaOffHeapIndexerInfinite(b *testing.B) {
	state := prepareJavaOffHeapIndexerBenchmark(b)
	for i := 0; i < b.N; i++ {
		fp := uint64(state.rng.NextLong()) & uint64(math.MaxInt64)
		offHeapIndexerBenchmarkResult = state.infinite.GetIdx(fp)
	}
}

// Port of OffHeapIndexerBenchmark.Bitshift.
func BenchmarkJavaOffHeapIndexerBitshift(b *testing.B) {
	state := prepareJavaOffHeapIndexerBenchmark(b)
	for i := 0; i < b.N; i++ {
		fp := uint64(state.rng.NextLong()) & uint64(math.MaxInt64)
		offHeapIndexerBenchmarkResult = state.bitshift.GetIdx(fp)
	}
}
