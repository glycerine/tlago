package tlc

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// No original method directly exercises missing checkpoint trace owners.
func TestDistributedTraceCheckpointUsesExistingOwner(t *testing.T) {
	for _, operation := range []string{"begin", "recover"} {
		for _, scenario := range []string{"missing", "missing-prior-error", "closed", "healthy-prior-error"} {
			t.Run(operation+"/"+scenario, func(t *testing.T) {
				trace := NewTLCTrace(t.TempDir(), "Spec")
				defer trace.Close()
				if err := trace.raf.WriteFull(make([]byte, 9)); err != nil {
					t.Fatal(err)
				}
				owner := trace.raf
				if scenario == "missing" || scenario == "missing-prior-error" {
					if err := owner.Close(); err != nil {
						t.Fatal(err)
					}
					trace.raf, owner = nil, nil
				} else if scenario == "closed" {
					if err := trace.Close(); err != nil {
						t.Fatal(err)
					}
				}
				if scenario == "missing-prior-error" || scenario == "healthy-prior-error" {
					trace.traceErr = NewIOException("earlier creation failure")
				}
				trace.lastPtr = 500
				prior := []byte("prior temporary checkpoint")
				if err := os.WriteFile(trace.chkptName("tmp"), prior, 0600); err != nil {
					t.Fatal(err)
				}
				checkpoint := make([]byte, 16)
				binary.BigEndian.PutUint64(checkpoint, 13)
				binary.BigEndian.PutUint64(checkpoint[8:], 73)
				if err := os.WriteFile(trace.chkptName("chkpt"), checkpoint, 0600); err != nil {
					t.Fatal(err)
				}
				err := invokeDistributedServerOperation(func() error {
					if operation == "begin" {
						return trace.BeginChkpt()
					}
					return trace.Recover()
				})
				if trace.raf != owner {
					t.Fatal("checkpoint operation replaced its file owner")
				}
				if scenario == "healthy-prior-error" {
					if err != nil {
						t.Fatal(err)
					}
				} else if scenario == "closed" {
					if !isJavaIOException(err) {
						t.Fatalf("closed owner failure: %T/%v", err, err)
					}
				} else if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing owner failure: %T/%v", err, err)
				}
				wantPtr := int64(500)
				if operation == "recover" {
					wantPtr = 73
				}
				if trace.lastPtr != wantPtr {
					t.Fatalf("last pointer = %d, want %d", trace.lastPtr, wantPtr)
				}
				if operation == "begin" {
					got, err := os.ReadFile(trace.chkptName("tmp"))
					if err != nil {
						t.Fatal(err)
					}
					if scenario == "healthy-prior-error" {
						binary.BigEndian.PutUint64(checkpoint, 9)
						binary.BigEndian.PutUint64(checkpoint[8:], 500)
						prior = checkpoint
					}
					if !bytes.Equal(got, prior) {
						t.Fatal("checkpoint bytes or failure-before-truncation ordering changed")
					}
				} else if scenario == "healthy-prior-error" && owner.curr != 13 {
					t.Fatal("recovery did not seek the existing owner")
				}
				if !trace.mu.TryLock() {
					t.Fatal("checkpoint operation retained trace lock")
				}
				trace.mu.Unlock()
			})
		}
	}
}
