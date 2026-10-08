package tlc

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// Source MemStateQueue.recover retains untouched slots/start and stores each
// empty state before reading it. No enabled original method checks this failure.
func TestDistributedMemoryQueueRecoveryRetainsPartialMutation(t *testing.T) {
	for _, field := range []string{"uid", "level", "complete"} {
		t.Run(field, func(t *testing.T) {
			oldVariables, oldVarCount := stateVariables, UniqueStringVariableCount()
			stateVariables = nil
			SetUniqueStringVariableCount(0)
			defer func() { stateVariables = oldVariables; SetUniqueStringVariableCount(oldVarCount) }()
			directory := t.TempDir()
			queue := NewMemStateQueue(directory)
			oldFirst, oldSecond, untouched := &TLCStateMut{UID: 98}, &TLCStateMut{UID: 97}, &TLCStateMut{UID: 96}
			queue.states[0], queue.states[1], queue.states[3] = oldFirst, oldSecond, untouched
			queue.start = 3
			backing := &queue.states[0]
			file, err := os.Create(filepath.Join(directory, "queue.chkpt"))
			if err != nil {
				t.Fatal(err)
			}
			out := NewValueOutputStreamWithGlobalCompression(file)
			if err := out.WriteInt(2); err != nil {
				t.Fatal(err)
			}
			if err := (&TLCStateMut{WorkerID: 5, UID: 42, level: 3}).Write(out); err != nil {
				t.Fatal(err)
			}
			if field == "complete" {
				if err := (&TLCStateMut{WorkerID: 7, UID: 91, level: 4}).Write(out); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := out.WriteShortNat(7); err != nil {
					t.Fatal(err)
				}
				if field == "level" {
					if err := out.WriteLongNat(91); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := out.Close(); err != nil {
				t.Fatal(err)
			}
			err = queue.Recover()
			if field == "complete" {
				if err != nil {
					t.Fatal(err)
				}
			} else if !isJavaIOException(err) {
				t.Fatalf("partial recovery failure = %v", err)
			}
			if queue.start != 3 || queue.len != 2 || len(queue.states) != memStateQueueInitialSize || &queue.states[0] != backing || queue.states[3] != untouched {
				t.Fatal("recovery reset cursor, resized or cleared untouched slots")
			}
			first, second := queue.states[0], queue.states[1]
			if first == nil || first == oldFirst || first.WorkerID != 5 || first.UID != 42 || first.level != 3 {
				t.Fatal("completed record was lost")
			}
			wantUID, wantLevel := TLCStateInitUID, TLCStateInitLevel
			if field != "uid" {
				wantUID = 91
			}
			if field == "complete" {
				wantLevel = 4
			}
			if second == nil || second == oldSecond || second.WorkerID != 7 || second.UID != wantUID || second.level != wantLevel {
				t.Fatal("failed slot did not retain source partial state")
			}
		})
	}
}

func TestDistributedMemoryQueueRecoveryRetainsFixedCapacity(t *testing.T) {
	oldVariables, oldVarCount := stateVariables, UniqueStringVariableCount()
	stateVariables = nil
	SetUniqueStringVariableCount(0)
	defer func() { stateVariables = oldVariables; SetUniqueStringVariableCount(oldVarCount) }()
	queue := NewMemStateQueue(t.TempDir())
	file, err := os.Create(filepath.Join(queue.diskdir, "queue.chkpt"))
	if err != nil {
		t.Fatal(err)
	}
	out := NewValueOutputStreamWithGlobalCompression(file)
	if err := out.WriteInt(memStateQueueInitialSize + 1); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < memStateQueueInitialSize; i++ {
		if err := (&TLCStateMut{WorkerID: 5, UID: int64(i), level: 3}).Write(out); err != nil {
			t.Fatal(err)
		}
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	err = queue.Recover()
	if _, ok := err.(*ArrayIndexOutOfBoundsException); !ok {
		t.Fatalf("fixed-capacity failure = %T %v", err, err)
	}
	if len(queue.states) != memStateQueueInitialSize || queue.len != memStateQueueInitialSize+1 || queue.states[memStateQueueInitialSize-1].UID != memStateQueueInitialSize-1 {
		t.Fatal("capacity failure grew or rolled back source reconstruction")
	}
}

func TestDistributedMemoryQueueFailurePrecedesFingerprintRecovery(t *testing.T) {
	oldVariables, oldVarCount := stateVariables, UniqueStringVariableCount()
	stateVariables = nil
	SetUniqueStringVariableCount(0)
	defer func() { stateVariables = oldVariables; SetUniqueStringVariableCount(oldVarCount) }()
	directory := t.TempDir()
	queue := NewMemStateQueue(directory)
	old := &TLCStateMut{UID: 97}
	queue.states[0] = old
	data := []byte{0, 0, 0, 1, 7}
	if err := os.WriteFile(filepath.Join(directory, "queue.chkpt"), data, 0600); err != nil {
		t.Fatal(err)
	}
	trace := NewTLCTrace(directory, "Spec")
	defer trace.Close()
	metadata := make([]byte, 16)
	binary.BigEndian.PutUint64(metadata[8:], 73)
	if err := os.WriteFile(trace.chkptName("chkpt"), metadata, 0600); err != nil {
		t.Fatal(err)
	}
	fingerprints := &queueFailureFingerprintSet{MemFPSet: NewMemFPSet()}
	server := NewTLCServer("Spec", "Spec", directory, NewNonDistributedFPSetManager(fingerprints, "local", trace), queue, trace)
	server.SetTool(NewTool())
	server.app.fromCheckpoint = &directory
	published := 0
	server.ConfigurePublication(TLCServerPublication{LocalHostName: func() (string, error) { published++; return "unexpected", nil }})
	code, err := server.ModelCheck()
	if code != ECGeneral || !isJavaIOException(err) || trace.lastPtr != 73 || fingerprints.recoverCalls != 0 || published != 0 || server.IsDone() {
		t.Fatalf("memory queue failure order = %d/%v", code, err)
	}
	failed := queue.states[0]
	if failed == nil || failed == old || failed.WorkerID != 7 || failed.UID != TLCStateInitUID || queue.len != 1 {
		t.Fatal("coordinator lost partial queue slot")
	}
}
