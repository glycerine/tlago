package tlc

import (
	"fmt"
	"testing"
)

var (
	javaIntervalBenchmarkSizes      = [...]uint{16, 18, 20, 22}
	javaIntervalBenchmarkPickCounts = [...]uint{10, 12, 14, 16}
	javaIntervalBenchmarkResult     *SetEnumValue
)

func BenchmarkJavaIntervalValueRandomSubset(b *testing.B) {
	oldSeed, oldRNG, oldPoly := RandomEnumerableSeed(), ResetRandomEnumerableValues(), FP64IrredPoly()
	SetRandomEnumerableSeed(15041980)
	ResetRandomEnumerableValues()
	FP64Init()
	b.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRNG)
		FP64InitPoly(oldPoly)
	})

	for _, size := range javaIntervalBenchmarkSizes {
		for _, numOfElements := range javaIntervalBenchmarkPickCounts {
			name := fmt.Sprintf("size=%d/numOfElements=%d", size, numOfElements)
			b.Run(name, func(b *testing.B) {
				interval := NewIntervalValue(1, int32(1<<size)).Normalize().(*IntervalValue)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := interval.GetRandomSubset(1 << numOfElements)
					if err != nil {
						b.Fatal(err)
					}
					javaIntervalBenchmarkResult = result
				}
			})
		}
	}
}
