package tlc

import (
	"encoding/binary"
	"os"
	"testing"
)

// No original method directly exercises recovered metadata publication before
// a missing/closed trace owner or a failed seek.
func TestWorkerTraceRecoveryPublishesBeforeSeek(t *testing.T) {
	for _, scenario := range []string{"healthy", "missing-owner", "closed-owner", "prior-open-error", "negative-seek"} {
		t.Run(scenario, func(t *testing.T) {
			worker := NewWorker(0)
			worker.SetTraceContext(t.TempDir(), "Spec")
			defer worker.CloseTrace()
			position := int64(13)
			if scenario == "negative-seek" {
				position = -1
			}
			checkpoint := make([]byte, 16)
			binary.BigEndian.PutUint64(checkpoint, uint64(position))
			binary.BigEndian.PutUint64(checkpoint[8:], 73)
			if err := os.WriteFile(worker.traceFileBase+".chkpt", checkpoint, 0600); err != nil {
				t.Fatal(err)
			}
			if scenario != "missing-owner" && scenario != "prior-open-error" {
				if err := worker.ensureTraceRAF(); err != nil {
					t.Fatal(err)
				}
			}
			owner := worker.traceRAF
			if scenario == "closed-owner" {
				if err := worker.CloseTrace(); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "prior-open-error" {
				worker.traceErr = NewIOException("earlier open failure")
			}
			worker.lastPtr = 500
			err := invokeDistributedServerOperation(worker.RecoverTrace)
			if worker.lastPtr != 73 {
				t.Fatalf("metadata read was not published before owner/seek failure: ptr %d, error %v", worker.lastPtr, err)
			}
			if worker.traceRAF != owner {
				t.Fatal("recovery replaced/reopened original trace owner")
			}
			switch scenario {
			case "healthy":
				if err != nil {
					t.Fatal(err)
				}
				if cursor, err := owner.GetFilePointer(); err != nil || cursor != 13 {
					t.Fatalf("recovered cursor %d/%v", cursor, err)
				}
			case "missing-owner", "prior-open-error":
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing original owner: %T/%v", err, err)
				}
			default:
				if !isJavaIOException(err) {
					t.Fatalf("failed source seek: %T/%v", err, err)
				}
			}
			if !worker.mu.TryLock() {
				t.Fatal("recovery failure retained worker monitor")
			}
			worker.mu.Unlock()
		})
	}
}

func TestWorkerTraceRecoveryReadPrecedesOwnerAccess(t *testing.T) {
	for length := 0; length < 16; length++ {
		worker := NewWorker(0)
		worker.SetTraceContext(t.TempDir(), "Spec")
		checkpoint := make([]byte, 16)
		binary.BigEndian.PutUint64(checkpoint, 13)
		binary.BigEndian.PutUint64(checkpoint[8:], 73)
		if err := os.WriteFile(worker.traceFileBase+".chkpt", checkpoint[:length], 0600); err != nil {
			t.Fatal(err)
		}
		worker.lastPtr = 500
		err := invokeDistributedServerOperation(worker.RecoverTrace)
		if !isJavaIOException(err) || worker.lastPtr != 500 || worker.traceRAF != nil {
			t.Fatalf("truncated %d-byte checkpoint changed publication/owner: %v, ptr %d", length, err, worker.lastPtr)
		}
	}
}
