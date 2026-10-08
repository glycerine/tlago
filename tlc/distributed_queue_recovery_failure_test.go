package tlc

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

type queueFailureFingerprintSet struct {
	*MemFPSet
	recoverCalls int
}

func (e *queueFailureFingerprintSet) RecoverTrace(*TLCTrace) error {
	e.recoverCalls++
	return NewIOException("fingerprint recovery must follow successful queue recovery")
}

// No enabled original test covers a failed queue state read during recovery.
// The source publishes the new empty slot before reading, preserves untouched
// buffer slots and propagates failure before fingerprint recovery/publication.
func TestDistributedQueueRecoveryPreservesPartialMutation(t *testing.T) {
	for _, phase := range []string{"enqueue", "dequeue"} {
		for _, field := range []string{"uid", "level"} {
			t.Run(phase+"/missing_"+field, func(t *testing.T) {
				queue := javaLongDiskStateQueueSetup(t)
				oldEnqueue, inactiveEnqueue := &TLCStateMut{UID: 98}, &TLCStateMut{UID: 97}
				oldDequeue, inactiveDequeue := &TLCStateMut{UID: 96}, &TLCStateMut{UID: 95}
				last := len(queue.deqBuf) - 1
				queue.enqBuf[0], queue.enqBuf[3] = oldEnqueue, inactiveEnqueue
				queue.deqBuf[last], queue.deqBuf[0] = oldDequeue, inactiveDequeue
				oldFile := queue.loFile
				file, err := os.Create(filepath.Join(queue.diskdir, "queue.chkpt"))
				if err != nil {
					t.Fatal(err)
				}
				out := NewValueOutputStreamWithGlobalCompression(file)
				for _, header := range []int{2, 5, 0, 1, last} {
					if err := out.WriteInt(int32(header)); err != nil {
						t.Fatal(err)
					}
				}
				if phase == "dequeue" {
					state := &TLCStateMut{WorkerID: 5, UID: 42, level: 3}
					if err := state.Write(out); err != nil {
						t.Fatal(err)
					}
				}
				// Retain each completed header field before the next missing field.
				if err := out.WriteShortNat(7); err != nil {
					t.Fatal(err)
				}
				wantUID := TLCStateInitUID
				if field == "level" {
					wantUID = 91
					if err := out.WriteLongNat(wantUID); err != nil {
						t.Fatal(err)
					}
				}
				if err := out.Close(); err != nil {
					t.Fatal(err)
				}
				trace := NewTLCTrace(queue.diskdir, "Spec")
				defer trace.Close()
				metadata := make([]byte, 16)
				binary.BigEndian.PutUint64(metadata[8:], 73)
				if err := os.WriteFile(filepath.Join(queue.diskdir, "Spec.st.chkpt"), metadata, 0600); err != nil {
					t.Fatal(err)
				}
				fingerprints := &queueFailureFingerprintSet{MemFPSet: NewMemFPSet()}
				manager := NewNonDistributedFPSetManager(fingerprints, "local", trace)
				server := NewTLCServer("Spec", "Spec", queue.diskdir, manager, queue, trace)
				server.SetTool(NewTool())
				server.app.fromCheckpoint = &queue.diskdir
				publicationCalls := 0
				server.ConfigurePublication(TLCServerPublication{LocalHostName: func() (string, error) { publicationCalls++; return "unexpected", nil }})
				code, err := server.ModelCheck()
				if code != ECGeneral || !isJavaIOException(err) {
					t.Fatalf("queue failure = %d/%v", code, err)
				}
				if trace.lastPtr != 73 || fingerprints.recoverCalls != 0 || publicationCalls != 0 || server.IsDone() {
					t.Fatal("queue failure changed recovery phase ordering")
				}
				if queue.len != 2 || queue.loPool != 5 || queue.hiPool != 0 || queue.enqIndex != 1 || queue.deqIndex != last || queue.lastLoPool != 4 || queue.loFile != oldFile {
					t.Fatal("partial queue header/restart mutation order changed")
				}
				if queue.enqBuf[3] != inactiveEnqueue || queue.deqBuf[0] != inactiveDequeue {
					t.Fatal("recovery cleared untouched buffer slots")
				}
				failed := queue.enqBuf[0]
				if phase == "dequeue" {
					if failed == nil || failed == oldEnqueue || failed.WorkerID != 5 || failed.UID != 42 || failed.level != 3 {
						t.Fatal("completed enqueue state was not preserved")
					}
					failed = queue.deqBuf[last]
				} else if queue.deqBuf[last] != oldDequeue {
					t.Fatal("unreached dequeue slot changed")
				}
				if failed == nil || failed == oldEnqueue || failed == oldDequeue || failed.WorkerID != 7 || failed.UID != wantUID || failed.level != TLCStateInitLevel {
					t.Fatal("failed state slot did not retain source partial header reads")
				}
			})
		}
	}
}
