package tlc

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// The original queue tests do not exercise raw queue filesystem boundaries.
func byteQueuePathFixture(directory string) *DiskByteArrayQueue {
	q := &DiskByteArrayQueue{diskdir: directory, loPool: 1, deqIndex: 2,
		deqBuf: make([][]byte, 2), enqBuf: make([][]byte, 2)}
	q.reader = NewByteArrayPoolReader(2, q.poolName(0))
	q.writer = NewByteArrayPoolWriter(2, nil)
	q.loFile = q.poolName(1)
	return q
}

func TestDistributedByteQueueDoesNotCreateStorageParents(t *testing.T) {
	for _, phase := range []string{"checkpoint", "pool_file", "spill"} {
		t.Run(phase, func(t *testing.T) {
			parent := filepath.Join(t.TempDir(), "missing")
			directory := filepath.Join(parent, "nested")
			q := byteQueuePathFixture(directory)
			var err error
			switch phase {
			case "checkpoint":
				err = q.BeginChkpt()
			case "pool_file":
				err = writeByteArrayPoolFile(q.poolName(0), [][]byte{{1}})
			case "spill":
				q.enqBuf[0], q.enqBuf[1], q.enqIndex = []byte{1}, []byte{2}, 2
				// Source doWork schedules the first pool without opening it.
				if err := q.spillEnqueueBuffer(); err != nil {
					t.Fatalf("initial pool scheduling failed: %v", err)
				}
				if _, err := os.Stat(parent); !os.IsNotExist(err) {
					t.Fatalf("scheduling created source-missing storage: %v", err)
				}
				q.enqBuf[0], q.enqBuf[1], q.enqIndex = []byte{3}, []byte{4}, 2
				err = q.spillEnqueueBuffer()
				if q.hiPool != 1 || q.enqIndex != 2 || !bytes.Equal(q.enqBuf[0], []byte{3}) {
					t.Fatal("failed pending pool write advanced spill state")
				}
			}
			if phase == "spill" {
				coded, ok := err.(*TLCError)
				if !ok || !coded.Runtime || coded.Code != ECSystemErrorWritingStates || coded.Cause != nil {
					t.Fatalf("missing-parent spill lost source coded assertion: %v", err)
				}
			} else if !isJavaIOException(err) {
				t.Fatalf("missing-parent %s = %v, want I/O failure", phase, err)
			}
			if _, err := os.Stat(parent); !os.IsNotExist(err) {
				t.Fatalf("%s created source-missing storage: %v", phase, err)
			}
		})
	}
}

func TestDistributedByteQueueRetainsLiteralStoragePaths(t *testing.T) {
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
	q := byteQueuePathFixture(link + string(os.PathSeparator) + "..")
	if err := q.BeginChkpt(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "queue.tmp")); err != nil {
		t.Fatalf("checkpoint begin lost source symlink traversal: %v", err)
	}
	if err := q.CommitChkpt(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "queue.chkpt")); err != nil {
		t.Fatalf("checkpoint commit changed source storage: %v", err)
	}
	q.len = 7
	if err := q.Recover(); err != nil || q.Size() != 0 {
		t.Fatalf("checkpoint recovery changed source storage: %v, size %d", err, q.Size())
	}
	if err := writeByteArrayPoolFile(q.poolName(0), [][]byte{{1, 2, 3}}); err != nil {
		t.Fatal(err)
	}
	encoded, err := os.ReadFile(filepath.Join(target, "0"))
	if err != nil || !bytes.Equal(encoded, []byte{0, 0, 0, 3, 1, 2, 3}) {
		t.Fatalf("pool path/content = %v/%v", encoded, err)
	}
	decoded := make([][]byte, 1)
	if err := readByteArrayPoolFile(q.poolName(0), decoded); err != nil || !bytes.Equal(decoded[0], []byte{1, 2, 3}) {
		t.Fatalf("pool recovery changed source storage: %v/%v", decoded, err)
	}
	for _, name := range []string{"queue.tmp", "queue.chkpt", "0"} {
		if _, err := os.Stat(filepath.Join(base, name)); !os.IsNotExist(err) {
			t.Fatalf("touched lexically cleaned path %s: %v", name, err)
		}
	}
}

func TestDistributedByteQueueEmptyDirectoryRetainsSourcePaths(t *testing.T) {
	q := NewDiskByteArrayQueue("")
	t.Cleanup(q.FinishAll)
	if q.diskdir != "" || q.loFile != string(os.PathSeparator)+"1" || q.poolName(0) != string(os.PathSeparator)+"0" {
		t.Fatalf("empty-directory constructor substituted storage: %q/%q/%q", q.diskdir, q.loFile, q.poolName(0))
	}
}
