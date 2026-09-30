package tlc

import (
	"math/big"
	"testing"
)

func TestOffHeapIndexerSelectionAndIndexing(t *testing.T) {
	bitshift := NewOffHeapIndexer(16, 1)
	if bitshift.Kind != OffHeapIndexerBitshifting {
		t.Fatalf("power-of-two indexer kind = %v, want bitshifting", bitshift.Kind)
	}
	for _, fp := range []uint64{1, 16, 123, 1<<40 + 99, diskFPSetFlushedMask} {
		assertOffHeapIndex(t, bitshift, fp)
	}
	if got := bitshift.GetIdxProbe(123, 5); got != (bitshift.GetIdx(123)+5)%int64(bitshift.Positions) {
		t.Fatalf("bitshift probe idx = %d", got)
	}

	mult := NewOffHeapIndexer(3<<27, 1)
	if mult.Kind != OffHeapIndexerMult1024 {
		t.Fatalf("1024MiB-multiple indexer kind = %v, want mult1024", mult.Kind)
	}
	for _, fp := range []uint64{1, 1<<20 + 7, 1<<48 + 17, diskFPSetFlushedMask} {
		assertOffHeapIndex(t, mult, fp)
	}

	infinite := NewOffHeapIndexer(1000, 1)
	if infinite.Kind != OffHeapIndexerInfinitePrecision {
		t.Fatalf("non-special indexer kind = %v, want infinite precision", infinite.Kind)
	}
	for _, fp := range []uint64{1, 999, 1<<32 + 12345, diskFPSetFlushedMask} {
		assertOffHeapIndex(t, infinite, fp)
	}
}

func assertOffHeapIndex(t *testing.T, indexer *OffHeapIndexer, fp uint64) {
	t.Helper()
	got := indexer.GetIdx(fp)
	want := exactOffHeapIndex(indexer.Positions, indexer.FPBits, fp)
	if got != want {
		t.Fatalf("kind %v GetIdx(%d) = %d, want %d", indexer.Kind, fp, got, want)
	}
}

func exactOffHeapIndex(positions uint64, fpBits int, fp uint64) int64 {
	fp &= diskFPSetFlushedMask
	numerator := new(big.Int).SetUint64(fp)
	numerator.Mul(numerator, new(big.Int).SetUint64(positions))
	denominator := new(big.Int).Lsh(big.NewInt(1), uint(64-fpBits))
	numerator.Div(numerator, denominator)
	numerator.Mod(numerator, new(big.Int).SetUint64(positions))
	return int64(numerator.Uint64())
}
