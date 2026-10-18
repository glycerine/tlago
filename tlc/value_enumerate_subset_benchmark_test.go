package tlc

import (
	"fmt"
	"testing"
)

var (
	javaEnumerateSubsetBenchmarkSizes  = [...]int32{0, 1, 2, 3, 4, 8, 10, 12, 14, 16, 18, 19}
	javaEnumerateSubsetBenchmarkResult *SetEnumValue
)

func prepareJavaEnumerateSubset(b *testing.B, size int32) (*IntervalValue, *SubsetValue, int) {
	b.Helper()
	inner := NewIntervalValue(1, size)
	subset := NewSubsetValue(inner)
	count, err := subset.Size()
	if err != nil {
		b.Fatal(err)
	}
	return inner, subset, count
}

func collectJavaEnumerateSubset(b *testing.B, values *ValueVec, elements ValueEnumeration) {
	b.Helper()
	for elem := elements.NextElement(); elem != nil; elem = elements.NextElement() {
		values.Add(elem)
	}
	if err := elements.Err(); err != nil {
		b.Fatal(err)
	}
}

func benchmarkJavaEnumerateSubset(b *testing.B, run func(*testing.B, int32) *SetEnumValue) {
	b.Helper()
	for _, size := range javaEnumerateSubsetBenchmarkSizes {
		b.Run(fmt.Sprintf("numOfElements=%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				javaEnumerateSubsetBenchmarkResult = run(b, size)
			}
		})
	}
}

func BenchmarkJavaEnumerateSubsetElementsAlwaysNormalized(b *testing.B) {
	benchmarkJavaEnumerateSubset(b, func(b *testing.B, size int32) *SetEnumValue {
		_, subset, count := prepareJavaEnumerateSubset(b, size)
		values := NewValueVec(count)
		collectJavaEnumerateSubset(b, values, subset.ElementsNormalized())
		return NewSetEnumValueVec(values, true).Normalize().(*SetEnumValue)
	})
}

func BenchmarkJavaEnumerateSubsetKElementsNotNormalized(b *testing.B) {
	benchmarkJavaEnumerateSubset(b, func(b *testing.B, size int32) *SetEnumValue {
		inner, subset, count := prepareJavaEnumerateSubset(b, size)
		values := NewValueVec(count)
		innerSize, err := inner.Size()
		if err != nil {
			b.Fatal(err)
		}
		for k := 0; k <= innerSize; k++ {
			collectJavaEnumerateSubset(b, values, subset.KElements(k))
		}
		return NewSetEnumValueVec(values, false)
	})
}

func BenchmarkJavaEnumerateSubsetKElementsNormalized(b *testing.B) {
	benchmarkJavaEnumerateSubset(b, func(b *testing.B, size int32) *SetEnumValue {
		inner, subset, count := prepareJavaEnumerateSubset(b, size)
		values := NewValueVec(count)
		innerSize, err := inner.Size()
		if err != nil {
			b.Fatal(err)
		}
		for k := 0; k <= innerSize; k++ {
			collectJavaEnumerateSubset(b, values, subset.KElements(k))
		}
		return NewSetEnumValueVec(values, false).Normalize().(*SetEnumValue)
	})
}

func BenchmarkJavaEnumerateSubsetElementsNotNormalized(b *testing.B) {
	benchmarkJavaEnumerateSubset(b, func(b *testing.B, size int32) *SetEnumValue {
		_, subset, count := prepareJavaEnumerateSubset(b, size)
		values := NewValueVec(count)
		collectJavaEnumerateSubset(b, values, subset.ElementsLexicographic())
		return NewSetEnumValueVec(values, false)
	})
}

func BenchmarkJavaEnumerateSubsetElementsNormalized(b *testing.B) {
	benchmarkJavaEnumerateSubset(b, func(b *testing.B, size int32) *SetEnumValue {
		_, subset, count := prepareJavaEnumerateSubset(b, size)
		values := NewValueVec(count)
		collectJavaEnumerateSubset(b, values, subset.ElementsLexicographic())
		return NewSetEnumValueVec(values, false).Normalize().(*SetEnumValue)
	})
}
