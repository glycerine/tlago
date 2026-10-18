package tlc

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestBitVectorReadPreservesJavaNegativeNaturalLength(t *testing.T) {
	path := filepath.Join(t.TempDir(), "negative-nat.bin")
	if err := os.WriteFile(path, []byte{0x80, 0x00, 0x00, 0x00}, 0600); err != nil {
		t.Fatal(err)
	}
	raf, err := NewBufferedRandomAccessFile(path, "r")
	if err != nil {
		t.Fatal(err)
	}
	defer raf.Close()
	if value, err := raf.ReadNat(); err != nil || value != math.MinInt32 {
		t.Fatalf("readNat = %d/%v, want %d/nil", value, err, math.MinInt32)
	}
	if err := raf.Seek(0); err != nil {
		t.Fatal(err)
	}
	bv := NewBitVector(64)
	bv.Set(7)
	defer func() {
		failure := recover()
		negative, ok := failure.(*NegativeArraySizeException)
		if !ok || negative.GetMessage() == nil || *negative.GetMessage() != "-2147483648" {
			t.Fatalf("failure = %T: %v, want NegativeArraySizeException(-2147483648)", failure, failure)
		}
		if !bv.Get(7) || len(bv.word) != 1 {
			t.Fatal("failed bit-vector allocation replaced existing words")
		}
	}()
	_ = bv.ReadFromBufferedRandomAccessFile(raf)
}

func TestBitVectorSerializationRejectsNullCarriersBeforeMutation(t *testing.T) {
	bv := NewBitVector(64)
	bv.Set(7)
	var in dataInput
	var out dataOutput
	var raf *BufferedRandomAccessFile
	calls := map[string]func(){
		"read":      func() { _ = bv.Read(in) },
		"write":     func() { _ = bv.Write(out) },
		"read_raf":  func() { _ = bv.ReadFromBufferedRandomAccessFile(raf) },
		"write_raf": func() { _ = bv.WriteToBufferedRandomAccessFile(raf) },
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			defer func() {
				failure := recover()
				if _, ok := failure.(*NullPointerException); !ok {
					t.Fatalf("failure = %T: %v, want NullPointerException", failure, failure)
				}
				if !bv.Get(7) || len(bv.word) != 1 {
					t.Fatal("null carrier changed bit-vector words")
				}
			}()
			call()
		})
	}
}

// Upstream's enabled BitVector tests cover printing, not null iterator input.
// Init dereferences bv.word before changing any iterator position.
func TestBitVectorIteratorRejectsNullInput(t *testing.T) {
	bv := NewBitVector(8)
	bv.Set(1)
	bv.Set(7)
	iter := NewBitVectorIter(bv)
	if iter.Next() != 1 {
		t.Fatal("iterator did not reach the first bit")
	}
	for _, call := range []func(){func() { NewBitVectorIter(nil) }, func() { iter.Init(nil) }} {
		failure := invokeDistributedServerOperation(func() error { call(); return nil })
		if _, ok := failure.(*NullPointerException); !ok {
			t.Fatalf("null iterator input: %T %v", failure, failure)
		}
	}
	if iter.Next() != 7 || iter.Next() != -1 {
		t.Fatal("failed reinitialization discarded the existing iterator position")
	}
	iter.Init(NewBitVector(0))
	if iter.Next() != -1 {
		t.Fatal("empty initialized vector must still be iterable")
	}
	for _, uninitialized := range []*BitVectorIter{{}, NewBitVectorIter(&BitVector{})} {
		failure := invokeDistributedServerOperation(func() error { uninitialized.Next(); return nil })
		if _, ok := failure.(*NullPointerException); !ok {
			t.Fatalf("uninitialized iterator storage: %T %v", failure, failure)
		}
	}
}

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

	custom := NewBitVector(3)
	custom.Set(0)
	got := javaStringUTF16(custom.StringRangeChars(0, 3, 0xd800, '\u03a9'))
	want := []uint16{'[', '\u03a9', '\u03a9', 0xd800, ']'}
	if !equalUTF16Units(got, want) {
		t.Fatalf("custom StringRange = %#v, want %#v", got, want)
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
	if a.HashCode() != b.HashCode() {
		t.Fatalf("equal bit vectors have different hashes: %d vs %d", a.HashCode(), b.HashCode())
	}
}

func TestBitVectorHashUsesJavaIntNarrowing(t *testing.T) {
	bv := NewBitVector(64)
	bv.Set(63)

	if got, want := bv.HashCode(), -2147483648; got != want {
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
	if got := bv.TrueCnt(); got != 4 {
		t.Fatalf("TrueCount = %d, want 4", got)
	}

	bv.Clear()
	if got := bv.TrueCnt(); got != 0 {
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
	decoded := NewBitVector()
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
