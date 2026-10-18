package tlc

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// No original method directly covers null-state write ordering. Source requires
// the predecessor before depth/file access, but dereferences the target only
// after completing the record.
func TestWorkerTraceNullStateWriteOrder(t *testing.T) {
	for _, scenario := range []string{"initial", "initial-closed", "successor", "successor-closed", "predecessor", "predecessor-closed", "predecessor-missing-owner", "both"} {
		t.Run(scenario, func(t *testing.T) {
			worker := NewWorker(2)
			worker.SetTraceContext(t.TempDir(), "Spec")
			defer worker.CloseTrace()
			if scenario != "predecessor-missing-owner" {
				if err := worker.ensureTraceRAF(); err != nil {
					t.Fatal(err)
				}
				if err := worker.traceRAF.WriteFull(make([]byte, 9)); err != nil {
					t.Fatal(err)
				}
			}
			owner := worker.traceRAF
			closed := scenario == "initial-closed" || scenario == "successor-closed" || scenario == "predecessor-closed"
			if closed {
				if err := worker.CloseTrace(); err != nil {
					t.Fatal(err)
				}
			}
			mirror := NewTLCTrace()
			worker.Checker = &ModelChecker{Trace: mirror}
			worker.lastPtr = 500
			worker.SetLevel(3)
			worker.UnseenSuccessorStates = 17
			parent := &TLCStateMut{UID: 37, WorkerID: 1, level: 7}
			target := &TLCStateMut{UID: 314, WorkerID: 7, level: 5}
			initial := scenario == "initial" || scenario == "initial-closed"
			missingParent := scenario == "predecessor" || scenario == "predecessor-closed" || scenario == "predecessor-missing-owner" || scenario == "both"
			if missingParent {
				parent = nil
			}
			if !missingParent || scenario == "both" {
				target = nil
			}
			err := invokeDistributedServerOperation(func() error {
				if initial {
					return worker.WriteInitState(nil, 0x8000000000000001)
				}
				return worker.WriteNextState(parent, 0x8000000000000001, target, nil)
			})
			if closed && !missingParent {
				if !isJavaIOException(err) {
					t.Fatalf("closed cursor access must precede target dereference: %T/%v", err, err)
				}
			} else if _, ok := err.(*NullPointerException); !ok {
				t.Fatalf("missing required state was suppressed: %T/%v", err, err)
			}
			level := 3
			if !initial && !missingParent {
				level = 8
			}
			completed := !closed && !missingParent
			last := int64(500)
			if completed {
				last = 9
			}
			if worker.traceRAF != owner || worker.lastPtr != last || worker.GetMaxLevel() != level || worker.UnseenSuccessorStates != 17 || len(mirror.Records()) != 0 {
				t.Fatal("null failure changed owner, pointer, depth, count or mirror order")
			}
			if target != nil && (target.UID != 314 || target.WorkerID != 7 || target.Level() != 5) {
				t.Fatal("missing predecessor changed target metadata")
			}
			if completed {
				if owner.curr != 22 {
					t.Fatalf("record incomplete before target failure: cursor %d", owner.curr)
				}
				if err := owner.Flush(); err != nil {
					t.Fatal(err)
				}
				want := make([]byte, 22)
				previous, previousWorker := uint32(37), byte(1)
				if initial {
					previous, previousWorker = 1, 2
				}
				binary.BigEndian.PutUint32(want[9:], previous)
				want[13] = previousWorker
				binary.BigEndian.PutUint64(want[14:], 0x8000000000000001)
				got, err := os.ReadFile(worker.traceFileBase + tlcTraceExt)
				if err != nil || !bytes.Equal(got, want) {
					t.Fatalf("completed record bytes changed: %x/%v", got, err)
				}
			} else if owner != nil && owner.curr != 9 {
				t.Fatalf("failure before writing moved cursor to %d", owner.curr)
			}
			if owner == nil {
				if _, err := os.Stat(worker.traceFileBase + tlcTraceExt); !os.IsNotExist(err) {
					t.Fatalf("missing predecessor opened a trace: %v", err)
				}
			}
			if !worker.mu.TryLock() {
				t.Fatal("state failure retained worker lock")
			}
			worker.mu.Unlock()
		})
	}
}
