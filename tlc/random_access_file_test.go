package tlc

import (
	"path/filepath"
	"testing"
)

func TestRandomAccessFileCountsPrimitiveOperations(t *testing.T) {
	raf, err := NewRandomAccessFile(filepath.Join(t.TempDir(), "raf.bin"), "rw")
	if err != nil {
		t.Fatal(err)
	}
	if err := raf.WriteByteValue(7); err != nil {
		t.Fatal(err)
	}
	if err := raf.Seek(0); err != nil {
		t.Fatal(err)
	}
	got, err := raf.ReadByteValue()
	if err != nil {
		t.Fatal(err)
	}
	if got != 7 {
		t.Fatalf("ReadByteValue = %d, want 7", got)
	}
	if raf.SuperWriteCnt != 1 || raf.SuperSeekCnt != 1 || raf.SuperReadCnt != 1 {
		t.Fatalf("counters = write %d seek %d read %d, want 1 each", raf.SuperWriteCnt, raf.SuperSeekCnt, raf.SuperReadCnt)
	}
	if err := raf.Close(); err != nil {
		t.Fatal(err)
	}
}
