package tlc

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// No original method directly covers cursor mutation after getTrace read failure.
func TestDistributedTraceFingerprintReadFailureRetainsCursor(t *testing.T) {
	for _, phase := range []string{"included-header", "excluded-header", "fingerprint", "later-record"} {
		t.Run(phase, func(t *testing.T) {
			data := []byte{0, 0, 0}
			included := phase != "excluded-header"
			if phase == "fingerprint" {
				data = []byte{0, 0, 0, 1, 7, 8, 9}
			}
			if phase == "later-record" {
				data = make([]byte, 14)
				binary.BigEndian.PutUint32(data, 12)
				binary.BigEndian.PutUint64(data[4:], 99)
			}
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "Spec.st"), data, 0600); err != nil {
				t.Fatal(err)
			}
			trace := NewTLCTrace(directory, "Spec")
			defer trace.Close()
			if err := trace.raf.Seek(1); err != nil {
				t.Fatal(err)
			}
			fps, err := trace.traceFPsFromDisk(0, included)
			if !isJavaIOException(err) || fps != nil {
				t.Fatalf("read failure published fingerprints: %v/%v", fps, err)
			}
			cursor, err := trace.raf.GetFilePointer()
			if err != nil || cursor != int64(len(data)) {
				t.Fatalf("failure cursor %d/%v, want %d", cursor, err, len(data))
			}
		})
	}
}

func TestDistributedTraceFingerprintReadRestoresCursorOnSuccess(t *testing.T) {
	for _, included := range []bool{false, true} {
		trace := NewTLCTrace(t.TempDir(), "Spec")
		if err := trace.raf.WriteLongNat(1); err != nil {
			t.Fatal(err)
		}
		if err := trace.raf.WriteLong(99); err != nil {
			t.Fatal(err)
		}
		saved, err := trace.raf.GetFilePointer()
		if err != nil {
			t.Fatal(err)
		}
		fps, err := trace.traceFPsFromDisk(0, included)
		wantLength := 0
		if included {
			wantLength = 1
		}
		if err != nil || len(fps) != wantLength || (included && fps[0] != 99) {
			t.Fatalf("normal fingerprints: %v/%v", fps, err)
		}
		if cursor, err := trace.raf.GetFilePointer(); err != nil || cursor != saved {
			t.Fatalf("normal cursor %d/%v, want %d", cursor, err, saved)
		}
		if err := trace.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDistributedTraceFingerprintCursorRestoreFailure(t *testing.T) {
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data, 1)
	binary.BigEndian.PutUint64(data[4:], 99)
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "Spec.st"), data, 0600); err != nil {
		t.Fatal(err)
	}
	trace := NewTLCTrace(directory, "Spec")
	defer trace.Close()
	// Keep the record cached, with the logical saved cursor outside that buffer.
	// Native descriptor closure allows cached reads but makes restoration's refill
	// fail. There is no production failure injection or replacement file interface.
	trace.raf.curr = BufferedRandomAccessFileBuffSz
	if err := trace.raf.file.Close(); err != nil {
		t.Fatal(err)
	}
	fps, err := trace.traceFPsFromDisk(0, true)
	if !isJavaIOException(err) || fps != nil {
		t.Fatalf("failed restoration published fingerprints: %v/%v", fps, err)
	}
	if cursor, err := trace.raf.GetFilePointer(); err != nil || cursor != BufferedRandomAccessFileBuffSz {
		t.Fatalf("restore attempt cursor: %d/%v", cursor, err)
	}
}
