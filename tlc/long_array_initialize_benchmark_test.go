package tlc

import (
	"fmt"
	"runtime"
	"testing"
)

var javaLongArrayInitializeSizes = [...]uint{8, 16, 24, 28, 31}

func benchmarkJavaLongArrayInitialize(b *testing.B, zero func(*LongArray) error) {
	b.Helper()
	for _, size := range javaLongArrayInitializeSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			array := NewLongArray(int64(1) << size)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := zero(array); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Port of LongArrayInitializeBenchmark.AputAddressSingle.
func BenchmarkJavaLongArrayInitializePutAddressSingle(b *testing.B) {
	benchmarkJavaLongArrayInitialize(b, func(array *LongArray) error {
		return array.ZeroMemory(1)
	})
}

// Port of LongArrayInitializeBenchmark.BputAddressConcurrent.
func BenchmarkJavaLongArrayInitializePutAddressConcurrent(b *testing.B) {
	availableProcessors := runtime.NumCPU()
	benchmarkJavaLongArrayInitialize(b, func(array *LongArray) error {
		return array.ZeroMemory(availableProcessors)
	})
}

// Port of LongArrayInitializeBenchmark.CsetMemory.
func BenchmarkJavaLongArrayInitializeSetMemory(b *testing.B) {
	benchmarkJavaLongArrayInitialize(b, func(array *LongArray) error {
		return array.ZeroMemory()
	})
}
