package tlc

import (
	"encoding/binary"
	"os"
	"reflect"
	"testing"
)

// No original method directly exercises worker read ownership or partial-record
// cursors. Record layout and mark/read/restore ordering follow Worker.readStateRecord.
func TestWorkerTraceRecordRequiresExistingOwner(t *testing.T) {
	for _, scenario := range []string{"healthy", "missing-owner", "closed-owner", "healthy-prior-error", "missing-owner-prior-error"} {
		t.Run(scenario, func(t *testing.T) {
			worker := NewWorker(0)
			worker.SetTraceContext(t.TempDir(), "Spec")
			defer worker.CloseTrace()
			record := make([]byte, 13)
			binary.BigEndian.PutUint32(record, 37)
			record[4] = 2
			binary.BigEndian.PutUint64(record[5:], 0x8000000000000001)
			if err := os.WriteFile(worker.traceFileBase+tlcTraceExt, record, 0600); err != nil {
				t.Fatal(err)
			}
			if scenario != "missing-owner" && scenario != "missing-owner-prior-error" {
				if err := worker.ensureTraceRAF(); err != nil {
					t.Fatal(err)
				}
				if err := worker.traceRAF.Seek(13); err != nil {
					t.Fatal(err)
				}
				worker.traceRAF.mark = 9
			}
			owner := worker.traceRAF
			if scenario == "closed-owner" {
				if err := worker.CloseTrace(); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "healthy-prior-error" || scenario == "missing-owner-prior-error" {
				worker.traceErr = NewIOException("earlier owner creation failure")
			}
			worker.lastPtr = 73
			var got ConcurrentTraceRecord
			err := invokeDistributedServerOperation(func() error {
				var err error
				got, err = worker.ReadStateRecord(0)
				return err
			})
			if worker.traceRAF != owner || worker.lastPtr != 73 {
				t.Fatal("record read replaced its owner or changed the writer's last pointer")
			}
			switch scenario {
			case "healthy", "healthy-prior-error":
				want := NewConcurrentTraceRecord(37, 2, 0x8000000000000001)
				if err != nil || !reflect.DeepEqual(got, want) || owner.curr != 13 || owner.mark != 13 {
					t.Fatalf("successful record/cursor/mark changed: %+v, %v", got, err)
				}
			case "closed-owner":
				if !isJavaIOException(err) || owner.mark != 13 || owner.curr != 13 {
					t.Fatalf("closed seek must follow mark publication: %v, mark %d", err, owner.mark)
				}
			default:
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing read owner: %T/%v", err, err)
				}
			}
			if !worker.mu.TryLock() {
				t.Fatal("record failure retained worker lock")
			}
			worker.mu.Unlock()
		})
	}
}

func TestWorkerTraceRecordFailureRetainsConsumedCursor(t *testing.T) {
	for length := 0; length <= 13; length++ {
		t.Run(fmtInt(length), func(t *testing.T) {
			worker := NewWorker(0)
			worker.SetTraceContext(t.TempDir(), "Spec")
			defer worker.CloseTrace()
			record := make([]byte, 13)
			binary.BigEndian.PutUint32(record, 37)
			record[4] = 2
			binary.BigEndian.PutUint64(record[5:], 73)
			if err := os.WriteFile(worker.traceFileBase+tlcTraceExt, record[:length], 0600); err != nil {
				t.Fatal(err)
			}
			if err := worker.ensureTraceRAF(); err != nil {
				t.Fatal(err)
			}
			owner := worker.traceRAF
			if err := owner.Seek(31); err != nil {
				t.Fatal(err)
			}
			owner.mark = 9
			worker.lastPtr = 500
			got, err := worker.ReadStateRecord(0)
			if length == 13 {
				if err != nil || !reflect.DeepEqual(got, NewConcurrentTraceRecord(37, 2, 73)) || owner.curr != 31 {
					t.Fatalf("complete record did not restore cursor: %+v/%v, cursor %d", got, err, owner.curr)
				}
			} else if !isJavaIOException(err) || owner.curr != int64(length) || !reflect.DeepEqual(got, ConcurrentTraceRecord{}) {
				t.Fatalf("partial record restored cursor or published result: %+v/%v, cursor %d", got, err, owner.curr)
			}
			if owner.mark != 31 || worker.lastPtr != 500 || worker.traceRAF != owner {
				t.Fatal("record read changed mark, last pointer or original owner")
			}
		})
	}
}
