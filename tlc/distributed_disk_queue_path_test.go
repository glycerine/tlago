package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

// No original queue method directly checks these checkpoint path boundaries.
func TestDistributedDiskQueueCheckpointDoesNotCreateParents(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "missing")
	directory := filepath.Join(parent, "nested")
	queue := NewDiskStateQueue(directory)
	t.Cleanup(queue.finishConstructionAndWait)
	if err := queue.BeginChkpt(); !isJavaIOException(err) {
		t.Fatalf("missing-parent begin = %v, want I/O failure", err)
	}
	if queue.diskdir != directory {
		t.Fatal("failed begin replaced the configured directory")
	}
	if _, err := os.Stat(parent); !os.IsNotExist(err) {
		t.Fatalf("checkpoint begin created source-missing parents: %v", err)
	}
}

func TestDistributedDiskQueueCheckpointRetainsLiteralPath(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, "target")
	child := filepath.Join(target, "child")
	if err := os.MkdirAll(child, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "link")
	if err := os.Symlink(child, link); err != nil {
		t.Fatal(err)
	}
	queue := NewDiskStateQueue(link + string(os.PathSeparator) + "..")
	t.Cleanup(queue.finishConstructionAndWait)
	if err := queue.BeginChkpt(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "queue.tmp")); err != nil {
		t.Fatalf("begin cleaned away the source symlink traversal: %v", err)
	}
	if err := queue.CommitChkpt(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "queue.chkpt")); err != nil {
		t.Fatalf("commit changed the source checkpoint location: %v", err)
	}
	queue.len = 5
	if err := queue.Recover(); err != nil || queue.Size() != 0 {
		t.Fatalf("recover did not use the literal checkpoint path: %v, size %d", err, queue.Size())
	}
	for _, name := range []string{"queue.tmp", "queue.chkpt"} {
		if _, err := os.Stat(filepath.Join(base, name)); !os.IsNotExist(err) {
			t.Fatalf("checkpoint touched lexically cleaned path %s: %v", name, err)
		}
	}
}

func TestDistributedDiskQueueEmptyDirectoryRetainsSourcePaths(t *testing.T) {
	queue := NewDiskStateQueue("")
	t.Cleanup(queue.finishConstructionAndWait)
	// Constructor workers wait for work; inspecting their configured paths
	// avoids performing filesystem operations against root checkpoint files.
	if queue.diskdir != "" || queue.loFile != string(os.PathSeparator)+"1" || queue.poolName(0) != string(os.PathSeparator)+"0" {
		t.Fatalf("empty-directory constructor substituted storage: %q/%q/%q", queue.diskdir, queue.loFile, queue.poolName(0))
	}
}
