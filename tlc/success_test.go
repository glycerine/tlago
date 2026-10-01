package tlc

import (
	"math"
	"testing"
)

func TestCalculateOptimisticProbabilityUsesJavaLongSubtractionBeforeDoubleConversion(t *testing.T) {
	distinct := int64(1<<54 + 1)
	generated := distinct + 1

	got := CalculateOptimisticProbability(uint64(distinct), generated)
	want := float64(distinct) / math.Pow(2, 64)
	if got != want {
		t.Fatalf("optimistic probability = %.17g, want %.17g", got, want)
	}
	if got == 0 {
		t.Fatalf("optimistic probability collapsed to zero; Java subtracts long counts before double conversion")
	}
}
