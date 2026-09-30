package tlc

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBigIntPortedJavaWrapperBehaviors(t *testing.T) {
	bi := NewBigIntString("-129")
	if bi.String() != "-129" {
		t.Fatalf("String = %s, want -129", bi)
	}
	if got := bi.FingerPrint(); got != FP64NewBytes(BigIntToJavaBytes(bi.Value())) {
		t.Fatalf("FingerPrint = %d, want FP64 over Java bytes", got)
	}

	var buf bytes.Buffer
	if err := bi.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	roundTrip, err := ReadBigInt(&buf)
	if err != nil {
		t.Fatalf("ReadBigInt: %v", err)
	}
	if !bi.Equal(roundTrip) {
		t.Fatalf("round trip = %s, want %s", roundTrip, bi)
	}
}

func TestBigSetPortedDeprecatedWriteBehavior(t *testing.T) {
	prefix := filepath.Join(t.TempDir(), "bigset")
	set := NewBigSetWithSizes(10, 2, prefix)
	if err := set.Put(NewBigIntString("2")); err != nil {
		t.Fatalf("Put(2): %v", err)
	}
	if err := set.Put(NewBigIntString("1")); err != nil {
		t.Fatalf("Put(1): %v", err)
	}
	if !set.Contains(NewBigIntString("1")) {
		t.Fatalf("Contains(1) = false, want true")
	}
	if err := set.Write(); err != nil {
		t.Fatalf("Write: %v", err)
	}

	file, err := os.Open(prefix + "0")
	if err != nil {
		t.Fatalf("open written file: %v", err)
	}
	defer file.Close()
	count, err := ReadInt(file)
	if err != nil {
		t.Fatalf("ReadInt: %v", err)
	}
	if count != 2 {
		t.Fatalf("written count = %d, want 2", count)
	}
	if set.FilePtr != 1 {
		t.Fatalf("FilePtr = %d, want 1", set.FilePtr)
	}
}
