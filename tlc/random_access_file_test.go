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

func TestRandomAccessFileModesMatchJava(t *testing.T) {
	for _, mode := range []string{"rw", "rws", "rwd"} {
		t.Run(mode, func(t *testing.T) {
			raf, err := NewRandomAccessFile(filepath.Join(t.TempDir(), "raf.bin"), mode)
			if err != nil {
				t.Fatal(err)
			}
			if err := raf.WriteByteValue(7); err != nil {
				t.Fatal(err)
			}
			if err := raf.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}

	_, err := NewRandomAccessFile(filepath.Join(t.TempDir(), "raf.bin"), "invalid")
	failure, ok := err.(*IllegalArgumentException)
	want := `Illegal mode "invalid" must be one of "r", "rw", "rws", or "rwd"`
	if !ok || failure.Error() != want {
		t.Fatalf("invalid mode failure = %T/%v, want IllegalArgumentException %q", err, err, want)
	}
}

func TestRandomAccessFileNegativeSeekMatchesJava(t *testing.T) {
	raf, err := NewRandomAccessFile(filepath.Join(t.TempDir(), "raf.bin"), "rw")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raf.Close() })

	err = raf.Seek(-1)
	failure, ok := err.(*IOException)
	if !ok || failure.Error() != "Negative seek offset" {
		t.Fatalf("negative seek failure = %T/%v, want IOException", err, err)
	}
	if raf.SuperSeekCnt != 1 {
		t.Fatalf("negative seek count = %d, want 1", raf.SuperSeekCnt)
	}
	if raf.SuperSeekTime != 0 {
		t.Fatalf("negative seek time = %d, want zero", raf.SuperSeekTime)
	}
}

func TestRandomAccessFileOpenFailuresMatchJava(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing", "raf.bin")
		_, err := NewRandomAccessFile(path, "r")
		failure, ok := err.(*FileNotFoundException)
		want := path + " (No such file or directory)"
		if !ok || failure.Error() != want {
			t.Fatalf("missing file failure = %T/%v, want FileNotFoundException %q", err, err, want)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		_, err := NewRandomAccessFile("bad\x00path", "r")
		failure, ok := err.(*FileNotFoundException)
		if !ok || failure.Error() != "Invalid file path" {
			t.Fatalf("invalid path failure = %T/%v, want FileNotFoundException", err, err)
		}
	})

	t.Run("mode precedes path validation", func(t *testing.T) {
		_, err := NewRandomAccessFile("bad\x00path", "invalid")
		failure, ok := err.(*IllegalArgumentException)
		want := `Illegal mode "invalid" must be one of "r", "rw", "rws", or "rwd"`
		if !ok || failure.Error() != want {
			t.Fatalf("invalid mode failure = %T/%v, want IllegalArgumentException %q", err, err, want)
		}
	})

	t.Run("directory", func(t *testing.T) {
		path := t.TempDir()
		_, err := NewRandomAccessFile(path, "r")
		failure, ok := err.(*FileNotFoundException)
		want := path + " (Is a directory)"
		if !ok || failure.Error() != want {
			t.Fatalf("directory failure = %T/%v, want FileNotFoundException %q", err, err, want)
		}
	})
}

func TestRandomAccessFileOperationFailuresMatchJava(t *testing.T) {
	path := filepath.Join(t.TempDir(), "raf.bin")
	raf, err := NewRandomAccessFile(path, "rw")
	if err != nil {
		t.Fatal(err)
	}
	if err := raf.Close(); err != nil {
		t.Fatal(err)
	}
	if err := raf.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}

	assertStreamClosed := func(operation string, err error) {
		t.Helper()
		failure, ok := err.(*IOException)
		if !ok || failure.Error() != "Stream Closed" {
			t.Fatalf("%s failure = %T/%v, want IOException %q", operation, err, err, "Stream Closed")
		}
	}
	assertStreamClosed("seek", raf.Seek(0))
	_, err = raf.Read(make([]byte, 1))
	assertStreamClosed("read", err)
	_, err = raf.Write([]byte{1})
	assertStreamClosed("write", err)
	if raf.SuperSeekCnt != 1 || raf.SuperReadCnt != 1 || raf.SuperWriteCnt != 1 {
		t.Fatalf("failed-operation counters = seek %d read %d write %d, want 1 each", raf.SuperSeekCnt, raf.SuperReadCnt, raf.SuperWriteCnt)
	}
	if raf.SuperSeekTime != 0 || raf.SuperReadTime != 0 || raf.SuperWriteTime != 0 {
		t.Fatalf("failed-operation times = seek %d read %d write %d, want zero", raf.SuperSeekTime, raf.SuperReadTime, raf.SuperWriteTime)
	}

	readOnly, err := NewRandomAccessFile(path, "r")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = readOnly.Close() })
	_, err = readOnly.Write([]byte{1})
	failure, ok := err.(*IOException)
	if !ok || failure.Error() != "Bad file descriptor" {
		t.Fatalf("read-only write failure = %T/%v, want IOException %q", err, err, "Bad file descriptor")
	}
	if readOnly.SuperWriteCnt != 1 || readOnly.SuperWriteTime != 0 {
		t.Fatalf("read-only write metrics = count %d time %d, want 1 and zero", readOnly.SuperWriteCnt, readOnly.SuperWriteTime)
	}
}
