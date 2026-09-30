package tlc

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestBufferedRandomAccessFileWriteReadAcrossBuffer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "braf.bin")
	raf, err := NewBufferedRandomAccessFile(path, "rw")
	if err != nil {
		t.Fatal(err)
	}
	count := BufferedRandomAccessFileBuffSz/8 + 3
	for i := 0; i < count; i++ {
		if err := raf.WriteLong(int64(i)); err != nil {
			t.Fatalf("write long %d: %v", i, err)
		}
	}
	if err := raf.Seek(0); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < count; i++ {
		got, err := raf.ReadLong()
		if err != nil {
			t.Fatalf("read long %d: %v", i, err)
		}
		if got != int64(i) {
			t.Fatalf("long %d = %d, want %d", i, got, i)
		}
	}
	if err := raf.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestBufferedRandomAccessFileWriteAfterSeekPastEndExtendsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "braf.bin")
	raf, err := NewBufferedRandomAccessFile(path, "rw")
	if err != nil {
		t.Fatal(err)
	}
	if err := raf.Seek(10); err != nil {
		t.Fatal(err)
	}
	if got, err := raf.ReadByteValue(); err != nil || got != -1 {
		t.Fatalf("read past EOF = %d, %v; want -1, nil", got, err)
	}
	if err := raf.WriteByteValue(0xff); err != nil {
		t.Fatal(err)
	}
	if length, err := raf.Length(); err != nil || length != 11 {
		t.Fatalf("length = %d, %v; want 11, nil", length, err)
	}
	if err := raf.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != 11 {
		t.Fatalf("disk size = %d, want 11", info.Size())
	}
}

func TestBufferedRandomAccessFileSetLengthConstrainsPointer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "braf.bin")
	raf, err := NewBufferedRandomAccessFile(path, "rw")
	if err != nil {
		t.Fatal(err)
	}
	if err := raf.Seek(10); err != nil {
		t.Fatal(err)
	}
	if err := raf.SetLength(5); err != nil {
		t.Fatal(err)
	}
	pointer, err := raf.GetFilePointer()
	if err != nil {
		t.Fatal(err)
	}
	if pointer != 5 {
		t.Fatalf("pointer = %d, want 5", pointer)
	}
	if _, err := raf.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("read at truncated EOF error = %v, want EOF", err)
	}
	if err := raf.Close(); err != nil {
		t.Fatal(err)
	}
}
