// Java source: test-benchmark/tlc2/util/FP64Benchmark.java
// (Microsoft Research, MIT license).
package tlc

import "testing"

var fp64BenchmarkResult uint64

func prepareJavaFP64Benchmark(b *testing.B) {
	b.Helper()
	previous := FP64IrredPoly()
	FP64Init()
	b.Cleanup(func() { FP64InitPoly(previous) })
	b.ResetTimer()
}

// Port of FP64Benchmark.extendIntLoop.
func BenchmarkJavaFP64ExtendIntLoop(b *testing.B) {
	prepareJavaFP64Benchmark(b)
	for i := 0; i < b.N; i++ {
		fp64BenchmarkResult = FP64ExtendLoop(72316478964978, 3876421)
	}
}

// Port of FP64Benchmark.extendIntUnrolled.
func BenchmarkJavaFP64ExtendIntUnrolled(b *testing.B) {
	prepareJavaFP64Benchmark(b)
	for i := 0; i < b.N; i++ {
		fp64BenchmarkResult = FP64ExtendInt(72316478964978, 3876421)
	}
}
