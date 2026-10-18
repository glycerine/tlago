package tlc

import (
	"fmt"
	"testing"
)

var (
	javaSetEnumBenchmarkPickCounts = [...]uint{10, 12, 14, 16}
	javaSetEnumBenchmarkSizes      = [...]uint{16, 18, 20, 22}
	javaSetEnumBenchmarkResult     *SetEnumValue
)

func javaSetEnumBenchmarkValues(from, to int) *ValueVec {
	values := NewValueVec(to - from)
	for i := from; i <= to; i++ {
		values.Add(NewIntValue(int32(i)))
	}
	return values
}

func BenchmarkJavaSetEnumValueRandomSubset(b *testing.B) {
	oldSeed, oldRNG, oldPoly := RandomEnumerableSeed(), ResetRandomEnumerableValues(), FP64IrredPoly()
	SetRandomEnumerableSeed(15041980)
	ResetRandomEnumerableValues()
	FP64Init()
	b.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRNG)
		FP64InitPoly(oldPoly)
	})

	for _, numOfElements := range javaSetEnumBenchmarkPickCounts {
		for _, size := range javaSetEnumBenchmarkSizes {
			name := fmt.Sprintf("numOfElements=%d/size=%d", numOfElements, size)
			b.Run(name, func(b *testing.B) {
				set := NewSetEnumValueVec(javaSetEnumBenchmarkValues(1, 1<<size), false).Normalize().(*SetEnumValue)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := set.GetRandomSubset(1 << numOfElements)
					if err != nil {
						b.Fatal(err)
					}
					javaSetEnumBenchmarkResult = result
				}
			})
		}
	}
}
