package tlc

import (
	"math"
	"testing"
)

func prepareJavaLongArrayBenchmark(b *testing.B) *LongArray {
	b.Helper()
	const elements = int64(1 << 10)
	array := NewLongArray(elements)
	if err := array.ZeroMemory(1); err != nil {
		b.Fatal(err)
	}
	for i := int64(0); i < elements; i++ {
		array.Set(i, math.MaxInt64-i)
	}
	b.ResetTimer()
	return array
}

// Port of LongArrayBenchmark.AswapWithCopy.
func BenchmarkJavaLongArraySwapWithCopy(b *testing.B) {
	array := prepareJavaLongArrayBenchmark(b)
	for i := 0; i < b.N; i++ {
		array.SwapCopy(0, array.Size()-1)
	}
}

// Port of LongArrayBenchmark.BswapGetSet.
func BenchmarkJavaLongArraySwapGetSet(b *testing.B) {
	array := prepareJavaLongArrayBenchmark(b)
	for i := 0; i < b.N; i++ {
		array.Swap(0, array.Size()-1)
	}
}
