package tlc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An empty directory is still a literal source path prefix. Build its absolute
// filename inside a temporary directory, never opening a filesystem-root file.
func emptyDirectoryDiskTrace(t *testing.T) *TLCTrace {
	t.Helper()
	directory := t.TempDir()
	t.Chdir(directory)
	full := filepath.Join(directory, "Spec")
	root := strings.TrimPrefix(strings.TrimPrefix(full, filepath.VolumeName(full)), string(os.PathSeparator))
	trace := NewTLCTrace("", root)
	t.Cleanup(func() { _ = trace.Close() })
	return trace
}

// No original method directly covers this prefix. Delete is native resource
// cleanup; checkpoint commit/recovery retain the reference file operations.
func TestDiskTraceEmptyDirectoryCheckpointAndCleanup(t *testing.T) {
	t.Run("commit-and-recover", func(t *testing.T) {
		trace := emptyDirectoryDiskTrace(t)
		if err := trace.WriteInitState(nil, 71); err != nil {
			t.Fatal(err)
		}
		position, err := trace.raf.GetFilePointer()
		if err != nil {
			t.Fatal(err)
		}
		last := trace.lastPtr
		if err := trace.BeginChkpt(); err != nil {
			t.Fatal(err)
		}
		pending, err := os.ReadFile(trace.chkptName("tmp"))
		if err != nil || len(pending) != 16 {
			t.Fatalf("checkpoint begin did not write source metadata: %x/%v", pending, err)
		}
		if err := trace.CommitChkpt(); err != nil {
			t.Fatal(err)
		}
		committed, err := os.ReadFile(trace.chkptName("chkpt"))
		if err != nil || !bytes.Equal(committed, pending) {
			t.Fatalf("commit skipped literal path: %x/%v", committed, err)
		}
		if _, err := os.Stat(trace.chkptName("tmp")); !os.IsNotExist(err) {
			t.Fatalf("commit retained temporary: %v", err)
		}
		if err := trace.WriteInitState(nil, 72); err != nil {
			t.Fatal(err)
		}
		if err := trace.Recover(); err != nil {
			t.Fatal(err)
		}
		if got, err := trace.raf.GetFilePointer(); err != nil || got != position || trace.lastPtr != last {
			t.Fatalf("recovery metadata/cursor = %d/%d/%v, want %d/%d", trace.lastPtr, got, err, last, position)
		}
	})
	t.Run("missing-temporary", func(t *testing.T) {
		trace := emptyDirectoryDiskTrace(t)
		if err := os.WriteFile(trace.chkptName("chkpt"), []byte("old checkpoint"), 0600); err != nil {
			t.Fatal(err)
		}
		err := trace.CommitChkpt()
		if !isJavaIOException(err) || err.Error() != "Trace.commitChkpt: cannot delete "+trace.chkptName("chkpt") {
			t.Fatalf("missing temporary commit failure: %T/%v", err, err)
		}
		if _, err := os.Stat(trace.chkptName("chkpt")); !os.IsNotExist(err) {
			t.Fatalf("commit did not delete old checkpoint before failed promotion: %v", err)
		}
	})
	t.Run("native-cleanup", func(t *testing.T) {
		trace := emptyDirectoryDiskTrace(t)
		owner := trace.raf
		if err := trace.BeginChkpt(); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(trace.chkptName("chkpt"), []byte("old checkpoint"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := trace.Delete(); err != nil {
			t.Fatal(err)
		}
		for _, path := range []string{trace.traceFileName(), trace.chkptName("tmp"), trace.chkptName("chkpt")} {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("cleanup skipped literal file %q: %v", path, err)
			}
		}
		if _, err := owner.GetFilePointer(); !isJavaIOException(err) {
			t.Fatalf("cleanup did not close existing owner: %v", err)
		}
	})
}
