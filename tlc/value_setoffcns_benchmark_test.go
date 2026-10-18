package tlc

import (
	"fmt"
	"testing"
)

var (
	javaSetOfFcnsBenchmarkPickCounts = [...]uint{10, 12, 14, 16}
	javaSetOfFcnsBenchmarkPairs      = [...][2]int32{{8, 16}, {16, 16}, {16, 32}, {46, 46}}
	javaSetOfFcnsBenchmarkResult     *SetEnumValue
)

func BenchmarkJavaSetOfFcnsRandomSubset(b *testing.B) {
	oldSeed, oldRNG, oldPoly := RandomEnumerableSeed(), ResetRandomEnumerableValues(), FP64IrredPoly()
	SetRandomEnumerableSeed(15041980)
	ResetRandomEnumerableValues()
	FP64Init()
	b.Cleanup(func() {
		SetRandomEnumerableSeed(oldSeed)
		SetRandomEnumerableGenerator(oldRNG)
		FP64InitPoly(oldPoly)
	})

	for _, numOfElements := range javaSetOfFcnsBenchmarkPickCounts {
		for _, pair := range javaSetOfFcnsBenchmarkPairs {
			sizeS, sizeT := pair[0], pair[1]
			name := fmt.Sprintf("numOfElements=%d/sizeS=%d/sizeT=%d", numOfElements, sizeS, sizeT)
			b.Run(name, func(b *testing.B) {
				domain := NewIntervalValue(1, sizeS)
				rangeValue := NewIntervalValue(1, sizeT)
				set := NewSetOfFcnsValue(domain, rangeValue).Normalize().(*SetOfFcnsValue)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err := set.GetRandomSubset(1 << numOfElements)
					if err != nil {
						b.Fatal(err)
					}
					javaSetOfFcnsBenchmarkResult = result
				}
			})
		}
	}
}
