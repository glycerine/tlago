package tlc

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestTraceRecoveryClosedOwnerPreservesReadBeforeSeek(t *testing.T) {
	directory := t.TempDir()
	trace := NewTLCTrace(directory, "Spec")
	checkpoint := make([]byte, 16)
	binary.BigEndian.PutUint64(checkpoint[8:], 73)
	if err := os.WriteFile(filepath.Join(directory, "Spec.st.chkpt"), checkpoint, 0600); err != nil {
		t.Fatal(err)
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	if err := trace.Recover(); !isJavaIOException(err) {
		t.Fatalf("closed trace recovery reopened its owner: %v", err)
	}
	if trace.lastPtr != 73 {
		t.Fatal("source metadata read did not precede failed seek")
	}
}

type recoveryPublicationQueue struct {
	StateQueue
	recovered bool
}

func (q *recoveryPublicationQueue) Recover() error {
	q.recovered = true
	return q.StateQueue.Recover()
}

// Source modelCheck performs recovery before hostname resolution/publication
// and outside the init catch(Throwable). No enabled original method directly
// covers these corrupt trace-metadata startup boundaries.
func TestDistributedRecoveryFailurePrecedesPublication(t *testing.T) {
	for length := -1; length <= 16; length++ {
		t.Run(fmt.Sprintf("metadata_bytes_%d", length), func(t *testing.T) {
			directory := t.TempDir()
			trace := NewTLCTrace(directory, "Spec")
			defer trace.Close()
			if length >= 0 {
				metadata := make([]byte, length)
				if length == 16 {
					// Both fields are complete, but the saved cursor is negative.
					binary.BigEndian.PutUint64(metadata, ^uint64(0))
					binary.BigEndian.PutUint64(metadata[8:], 73)
				}
				if err := os.WriteFile(filepath.Join(directory, "Spec.st.chkpt"), metadata, 0600); err != nil {
					t.Fatal(err)
				}
			}
			queue := &recoveryPublicationQueue{StateQueue: NewStateQueue(directory)}
			server := NewTLCServer("Spec", "Spec", directory, nil, queue, trace)
			server.SetTool(NewTool())
			server.app.fromCheckpoint = &directory
			publicationCalls := 0
			server.ConfigurePublication(TLCServerPublication{LocalHostName: func() (string, error) { publicationCalls++; return "unexpected", nil }})
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			defer RemoveMessageRecorder(recorder)
			code, err := server.ModelCheck()
			if code != ECGeneral || !isJavaIOException(err) {
				t.Fatalf("recovery failure = %d/%v", code, err)
			}
			if publicationCalls != 0 || queue.recovered || server.IsDone() {
				t.Fatal("failed trace recovery reached a later startup phase")
			}
			if length == 16 && trace.lastPtr != 73 {
				t.Fatal("complete metadata did not update the pointer before invalid seek")
			}
			if !recorder.Recorded(ECTLCCheckpointRecoverStart) || recorder.Recorded(ECTLCCheckpointRecoverEnd) || recorder.Recorded(ECTLCComputingInit) || recorder.Recorded(ECTLCDistributedServerRunning) || recorder.Recorded(ECGeneral) {
				t.Fatal("recovery failure entered initialization/reporting instead of escaping")
			}
		})
	}
}
