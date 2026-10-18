package tlc

import (
	"fmt"
	"testing"
)

var (
	javaSubsetBenchmarkPickCounts = [...]uint{10, 12, 14, 16, 20, 24}
	javaSubsetBenchmarkSizes      = [...]int32{32, 64, 128, 256}
	javaSubsetBenchmarkResult     *SetEnumValue
)

func benchmarkJavaSubset(b *testing.B, run func(*SubsetValue, int) (*SetEnumValue, error)) {
	b.Helper()
	oldSeed, oldRNG, oldPoly := RandomEnumerableSeed(), ResetRandomEnumerableValues(), FP64IrredPoly()
	SetRandomEnumerableSeed(15041980)
	ResetRandomEnumerableValues()
	FP64Init()
	b.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRNG)
		FP64InitPoly(oldPoly)
	})

	for _, numOfElements := range javaSubsetBenchmarkPickCounts {
		for _, size := range javaSubsetBenchmarkSizes {
			if size >= 128 && numOfElements > 20 {
				continue // Source setup calls System.exit(0) for this pair.
			}
			name := fmt.Sprintf("numOfElements=%d/size=%d", numOfElements, size)
			b.Run(name, func(b *testing.B) {
				subset := NewSubsetValue(NewIntervalValue(1, size)).Normalize().(*SubsetValue)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := run(subset, 1<<numOfElements)
					if err != nil {
						b.Fatal(err)
					}
					javaSubsetBenchmarkResult = result
				}
			})
		}
	}
}

// Port of SubsetBenchmark.randomSetOfSubsets.
func BenchmarkJavaSubsetRandomSetOfSubsets(b *testing.B) {
	benchmarkJavaSubset(b, func(subset *SubsetValue, picks int) (*SetEnumValue, error) {
		return subset.GetRandomSetOfSubsets(picks, 0.1)
	})
}

// Port of SubsetBenchmark.randomSetOfSubsetsExact.
func BenchmarkJavaSubsetRandomSetOfSubsetsExact(b *testing.B) {
	benchmarkJavaSubset(b, func(subset *SubsetValue, picks int) (*SetEnumValue, error) {
		return subset.GetRandomSetOfSubsetsUpTo(picks, 10)
	})
}
