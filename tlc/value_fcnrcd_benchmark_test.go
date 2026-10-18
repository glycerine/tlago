// Java source: test-benchmark/tlc2/value/impl/FcnRcdBenchmark.java
// (Microsoft Research, MIT license).
package tlc

import (
	"fmt"
	"strconv"
	"testing"
)

var (
	javaFcnRcdBenchmarkSizes   = [...]int{0, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192}
	javaFcnRcdBenchmarkShuffle = NewJavaRandomDefault()
	javaFcnRcdBenchmarkResult  []Value
)

func prepareJavaFcnRcdBenchmark(b *testing.B, size int) *FcnRcdValue {
	b.Helper()
	domain := make([]Value, size)
	values := make([]Value, size)
	for i := range domain {
		domain[i] = NewStringValue("asdfghjkoiuytrewqzxcvbn" + strconv.Itoa(i))
		values[i] = NewIntValue(int32(i))
	}
	for i := len(domain); i > 1; i-- {
		j := int(javaFcnRcdBenchmarkShuffle.NextIntN(int32(i)))
		domain[i-1], domain[j] = domain[j], domain[i-1]
	}
	fcn := NewFcnRcdValue(domain, values, false).Normalize().(*FcnRcdValue)
	b.ResetTimer()
	return fcn
}

func benchmarkJavaFcnRcdSelect(b *testing.B, selectValue func(*FcnRcdValue, Value) (Value, error)) {
	b.Helper()
	for _, size := range javaFcnRcdBenchmarkSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			fcn := prepareJavaFcnRcdBenchmark(b, size)
			for iteration := 0; iteration < b.N; iteration++ {
				result := make([]Value, size)
				for i := range result {
					domain := NewStringValue("asdfghjkoiuytrewqzxcvbn" + strconv.Itoa(i))
					value, err := selectValue(fcn, domain)
					if err != nil {
						b.Fatal(err)
					}
					result[i] = value
				}
				javaFcnRcdBenchmarkResult = result
			}
		})
	}
}

// Port of FcnRcdBenchmark.fcnRcdValueSelectNoIndex.
func BenchmarkJavaFcnRcdValueSelectNoIndex(b *testing.B) {
	benchmarkJavaFcnRcdSelect(b, (*FcnRcdValue).SelectLinearSearch)
}

// Port of FcnRcdBenchmark.fcnRcdValueSelectBinarySearch.
func BenchmarkJavaFcnRcdValueSelectBinarySearch(b *testing.B) {
	benchmarkJavaFcnRcdSelect(b, (*FcnRcdValue).SelectBinarySearch)
}
