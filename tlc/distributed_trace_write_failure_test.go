package tlc

import "testing"

// No original method directly covers TLCTrace's last-record pointer on partial
// record writes. Source assigns it before the predecessor/fingerprint writes.
func TestDistributedTraceWriteFailureRetainsRecordPointer(t *testing.T) {
	for _, kind := range []string{"initial", "successor", "single-process"} {
		for _, phase := range []string{"predecessor", "fingerprint"} {
			t.Run(kind+"/"+phase, func(t *testing.T) {
				trace := NewTLCTrace(t.TempDir(), "Spec")
				defer trace.Close()
				offset := int64(BufferedRandomAccessFileBuffSz - 2)
				if phase == "fingerprint" {
					offset = BufferedRandomAccessFileBuffSz - 8
				}
				if err := trace.raf.WriteFull(make([]byte, offset)); err != nil {
					t.Fatal(err)
				}
				if err := trace.raf.Flush(); err != nil {
					t.Fatal(err)
				}
				// The buffered owner's logical pointer remains available, but
				// crossing its buffer boundary now fails in the native file I/O.
				if err := trace.raf.file.Close(); err != nil {
					t.Fatal(err)
				}
				trace.lastPtr = 1
				state := &TLCStateMut{UID: 314, WorkerID: TLCStateInitWorkerID, level: 5}
				var predecessor *TLCStateMut
				if kind != "initial" {
					predecessor = state
				}
				var err error
				if kind == "single-process" {
					err = trace.WriteNextStateForWorker(3, predecessor, 0x0102030405060708, state, nil)
				} else {
					_, err = trace.WriteStateRecord(predecessor, 0x0102030405060708, state)
				}
				if !isJavaIOException(err) {
					t.Fatalf("partial record write failure: %T/%v", err, err)
				}
				if trace.lastPtr != offset {
					t.Fatalf("last-record pointer = %d, want attempted record at %d", trace.lastPtr, offset)
				}
				if pointer, err := trace.raf.GetFilePointer(); err != nil || pointer != BufferedRandomAccessFileBuffSz {
					t.Fatalf("partial write changed cursor: %d/%v", pointer, err)
				}
				if len(trace.Records()) != 0 || state.UID != 314 || state.WorkerID != TLCStateInitWorkerID || state.Level() != 5 {
					t.Fatal("failed record write published a record or changed state metadata")
				}
				if phase == "fingerprint" {
					wantPredecessor := byte(1)
					if kind != "initial" {
						wantPredecessor = 58 // Low byte of UID 314.
					}
					if trace.raf.buff[offset+3] != wantPredecessor || trace.raf.buff[offset+4] != 1 || trace.raf.buff[offset+7] != 4 {
						t.Fatal("failure discarded the completed header or partial fingerprint")
					}
				}
			})
		}
	}
}
