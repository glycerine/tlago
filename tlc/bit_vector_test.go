package tlc

import (
	"path/filepath"
	"testing"
)

func TestBitVectorStringMatchesJavaFormatting(t *testing.T) {
	bv := NewBitVector(8)
	bv.Set(0)
	bv.Set(1)
	bv.Set(3)
	bv.Set(4)
	bv.Set(7)

	if got := bv.String(); got != "[10011011]" {
		t.Fatalf("String = %q, want [10011011]", got)
	}
	if got := bv.StringRange(4, 3); got != "[001]" {
		t.Fatalf("StringRange = %q, want [001]", got)
	}
}

func TestBitVectorEqualityIgnoresTrailingZeroWords(t *testing.T) {
	a := NewBitVector(1)
	b := NewBitVector(130)
	a.Set(0)
	b.Set(0)
	b.Set(129)
	b.Reset(129)

	if !a.Equal(b) {
		t.Fatalf("bit vectors with only trailing zero word differences should compare equal")
	}
	if a.Hash() != b.Hash() {
		t.Fatalf("equal bit vectors have different hashes: %d vs %d", a.Hash(), b.Hash())
	}
}

func TestBitVectorHashUsesJavaIntNarrowing(t *testing.T) {
	bv := NewBitVector(64)
	bv.Set(63)

	if got, want := bv.Hash(), -2147483648; got != want {
		t.Fatalf("Hash = %d, want Java signed int hash %d", got, want)
	}
}

func TestBitVectorIteratorReturnsSetBitsInAscendingOrder(t *testing.T) {
	bv := NewBitVector(0)
	for _, bit := range []int{70, 1, 64} {
		bv.Set(bit)
	}

	iter := NewBitVectorIter(bv)
	for _, want := range []int{1, 64, 70, -1} {
		if got := iter.Next(); got != want {
			t.Fatalf("Next = %d, want %d", got, want)
		}
	}
}

func TestBitVectorSetRangeAndTrueCount(t *testing.T) {
	bv := NewBitVector(0)
	bv.SetRange(62, 66)
	bv.SetBool(64, false)

	for _, bit := range []int{62, 63, 65, 66} {
		if !bv.Get(bit) {
			t.Fatalf("bit %d is false, want true", bit)
		}
	}
	if bv.Get(64) {
		t.Fatalf("bit 64 is true, want false")
	}
	if got := bv.TrueCount(); got != 4 {
		t.Fatalf("TrueCount = %d, want 4", got)
	}

	bv.Clear()
	if got := bv.TrueCount(); got != 0 {
		t.Fatalf("TrueCount after Clear = %d, want 0", got)
	}
}

func TestBitVectorBufferedRandomAccessFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bits.bin")
	raf, err := NewBufferedRandomAccessFile(path, "rw")
	if err != nil {
		t.Fatal(err)
	}
	original := NewBitVector(0)
	for _, bit := range []int{0, 63, 64, 127, 130} {
		original.Set(bit)
	}
	if err := original.WriteToBufferedRandomAccessFile(raf); err != nil {
		t.Fatal(err)
	}
	if err := raf.Seek(0); err != nil {
		t.Fatal(err)
	}
	decoded := NewBitVector(0)
	if err := decoded.ReadFromBufferedRandomAccessFile(raf); err != nil {
		t.Fatal(err)
	}
	if !decoded.Equal(original) {
		t.Fatalf("decoded bit vector did not match original")
	}
	if err := raf.Close(); err != nil {
		t.Fatal(err)
	}
}
