package tlc

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// No original method directly covers these worker checkpoint owner boundaries.
// Source flushes its existing owner before opening/truncating the temporary file.
func TestWorkerTraceCheckpointRequiresExistingOwner(t *testing.T) {
	for _, scenario := range []string{"healthy", "missing-owner", "missing-owner-prior-error", "closed-owner", "flush-error", "healthy-prior-error", "temporary-open-error"} {
		t.Run(scenario, func(t *testing.T) {
			worker := NewWorker(0)
			worker.SetTraceContext(t.TempDir(), "Spec")
			defer worker.CloseTrace()
			if scenario != "missing-owner" && scenario != "missing-owner-prior-error" {
				if err := worker.ensureTraceRAF(); err != nil {
					t.Fatal(err)
				}
				if _, err := worker.traceRAF.Write(make([]byte, 13)); err != nil {
					t.Fatal(err)
				}
			}
			owner := worker.traceRAF
			if scenario == "closed-owner" {
				if err := worker.CloseTrace(); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "flush-error" {
				if err := owner.file.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "healthy-prior-error" || scenario == "missing-owner-prior-error" {
				worker.traceErr = NewIOException("earlier owner creation failure")
			}
			worker.lastPtr = 73
			temporary := worker.traceFileBase + ".tmp"
			if scenario == "temporary-open-error" {
				if err := os.Mkdir(temporary, 0700); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(temporary, []byte("prior temporary"), 0600); err != nil {
				t.Fatal(err)
			}
			err := invokeDistributedServerOperation(worker.BeginChkpt)
			if worker.traceRAF != owner || worker.lastPtr != 73 {
				t.Fatal("checkpoint changed its original trace owner or last pointer")
			}
			healthy := scenario == "healthy" || scenario == "healthy-prior-error"
			if healthy {
				if err != nil {
					t.Fatal(err)
				}
				want := make([]byte, 16)
				binary.BigEndian.PutUint64(want, 13)
				binary.BigEndian.PutUint64(want[8:], 73)
				got, err := os.ReadFile(temporary)
				if err != nil || !bytes.Equal(got, want) {
					t.Fatalf("checkpoint metadata bytes changed: %x/%v", got, err)
				}
			} else if scenario == "missing-owner" || scenario == "missing-owner-prior-error" {
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing owner must fail before temporary access: %T/%v", err, err)
				}
			} else if !isJavaIOException(err) {
				t.Fatalf("checkpoint owner/file failure lost I/O category: %T/%v", err, err)
			}
			if !healthy && scenario != "temporary-open-error" {
				got, err := os.ReadFile(temporary)
				if err != nil || string(got) != "prior temporary" {
					t.Fatalf("owner failure truncated temporary metadata: %q/%v", got, err)
				}
			}
			if healthy || scenario == "temporary-open-error" {
				info, err := os.Stat(worker.traceFileBase + tlcTraceExt)
				if err != nil || info.Size() != 13 {
					t.Fatalf("metadata access preceded trace flush: %v/%v", info, err)
				}
			}
			if !worker.mu.TryLock() {
				t.Fatal("checkpoint failure retained worker lock")
			}
			worker.mu.Unlock()
		})
	}
}
