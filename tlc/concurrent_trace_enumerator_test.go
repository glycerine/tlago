package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

func concurrentEnumeratorReader(t *testing.T, fp *int64) *WorkerTraceEnumerator {
	t.Helper()
	path := filepath.Join(t.TempDir(), "worker.st")
	writer, err := NewBufferedRandomAccessFile(path, "rw")
	if err != nil {
		t.Fatal(err)
	}
	if fp != nil {
		if err := writer.WriteLongNat(1); err != nil {
			t.Fatal(err)
		}
		if err := writer.WriteShortNat(0); err != nil {
			t.Fatal(err)
		}
		if err := writer.WriteLong(*fp); err != nil {
			t.Fatal(err)
		}
	}
	length, err := writer.GetFilePointer()
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	reader, err := NewBufferedRandomAccessFile(path, "r")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reader.Close() })
	return &WorkerTraceEnumerator{raf: reader, length: length}
}

// No original methods directly cover these concurrent enumerator boundaries.
func TestConcurrentTraceEnumeratorExhaustedReadFails(t *testing.T) {
	for _, count := range []int{0, 1} {
		enumerator := &ConcurrentTraceEnumerator{}
		if count == 1 {
			enumerator.enums = []*WorkerTraceEnumerator{concurrentEnumeratorReader(t, nil)}
		}
		_, err := enumerator.NextFP()
		if _, ok := err.(*ArrayIndexOutOfBoundsException); !ok {
			t.Fatalf("exhausted %d-reader access: %T/%v", count, err, err)
		}
		if enumerator.idx != count {
			t.Fatalf("exhausted cursor %d, want %d", enumerator.idx, count)
		}
	}
}

func TestConcurrentTraceEnumeratorCursorFailureDoesNotAdvance(t *testing.T) {
	for _, operation := range []string{"position", "fingerprint"} {
		reader := concurrentEnumeratorReader(t, nil)
		if err := reader.Close(); err != nil {
			t.Fatal(err)
		}
		enumerator := &ConcurrentTraceEnumerator{enums: []*WorkerTraceEnumerator{reader, concurrentEnumeratorReader(t, nil)}}
		err := invokeDistributedServerOperation(func() error {
			if operation == "position" {
				enumerator.NextPos()
				return nil
			}
			_, err := enumerator.NextFP()
			return err
		})
		if !isJavaIOException(err) || enumerator.idx != 0 {
			t.Fatalf("closed %s cursor failure: %v, index %d", operation, err, enumerator.idx)
		}
	}
}

func TestConcurrentTraceEnumeratorCloseStopsAtFirstFailure(t *testing.T) {
	first, last := concurrentEnumeratorReader(t, nil), concurrentEnumeratorReader(t, nil)
	if err := first.raf.file.Close(); err != nil {
		t.Fatal(err)
	}
	enumerator := &ConcurrentTraceEnumerator{enums: []*WorkerTraceEnumerator{first, last}}
	if err := enumerator.Close(); !isJavaIOException(err) {
		t.Fatalf("close failure %v", err)
	}
	if first.raf == nil || !first.raf.closed || last.raf == nil || last.raf.closed {
		t.Fatal("close failure discarded first owner or closed later reader")
	}
}

func TestConcurrentTraceElementsRequiresWorker(t *testing.T) {
	trace := &ConcurrentTLCTrace{TLCTrace: NewTLCTrace(), Workers: []*Worker{nil}}
	err := invokeDistributedServerOperation(func() error { _, err := trace.Elements(); return err })
	if _, ok := err.(*NullPointerException); !ok {
		t.Fatalf("missing enumerated worker: %T/%v", err, err)
	}
	if !trace.mu.TryLock() {
		t.Fatal("failed elements creation retained monitor")
	}
	trace.mu.Unlock()
}

func TestWorkerTraceEnumeratorDoesNotFlushWriter(t *testing.T) {
	worker := NewWorker(0)
	worker.SetTraceContext(t.TempDir(), "Spec")
	defer worker.CloseTrace()
	if err := worker.WriteInitState(NewEmptyState(), 99); err != nil {
		t.Fatal(err)
	}
	enumerator, err := worker.Elements()
	if err != nil {
		t.Fatal(err)
	}
	defer enumerator.Close()
	if !worker.traceRAF.dirty {
		t.Fatal("enumerator construction flushed source writer")
	}
	if !enumerator.HasMoreFP() {
		t.Fatal("enumerator did not snapshot logical writer position")
	}
	if _, err := enumerator.NextFP(); !isJavaIOException(err) {
		t.Fatalf("fresh reader unexpectedly read unflushed bytes: %v", err)
	}
}

func TestConcurrentTraceEnumeratorSourceAdvanceAndReset(t *testing.T) {
	fp := int64(99)
	enumerator := &ConcurrentTraceEnumerator{enums: []*WorkerTraceEnumerator{concurrentEnumeratorReader(t, nil), concurrentEnumeratorReader(t, &fp)}}
	if pos := enumerator.NextPos(); pos != 42 || enumerator.idx != 1 {
		t.Fatalf("position %d/index %d", pos, enumerator.idx)
	}
	if got, err := enumerator.NextFP(); err != nil || got != 99 {
		t.Fatalf("fingerprint %d/%v", got, err)
	}
	enumerator.Reset(123)
	if enumerator.idx != 0 {
		t.Fatal("reset did not reset selector")
	}
	if pos := enumerator.NextPos(); pos != -1 || enumerator.idx != 1 {
		t.Fatal("reset rewound child reader")
	}
	if err := enumerator.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkerTraceEnumeratorRequiresOriginalWriterAndReader(t *testing.T) {
	for _, scenario := range []string{"missing-writer", "closed-writer", "missing-file"} {
		t.Run(scenario, func(t *testing.T) {
			worker := NewWorker(0)
			worker.SetTraceContext(t.TempDir(), "Spec")
			defer worker.CloseTrace()
			if scenario != "missing-writer" {
				if err := worker.ensureTraceRAF(); err != nil {
					t.Fatal(err)
				}
				if scenario == "closed-writer" {
					if err := worker.traceRAF.Close(); err != nil {
						t.Fatal(err)
					}
				} else if err := os.Remove(worker.traceFileBase + tlcTraceExt); err != nil {
					t.Fatal(err)
				}
			}
			err := invokeDistributedServerOperation(func() error { _, err := worker.Elements(); return err })
			if scenario == "missing-writer" {
				if _, ok := err.(*NullPointerException); !ok {
					t.Fatalf("missing writer failure: %T/%v", err, err)
				}
			} else if !isJavaIOException(err) {
				t.Fatalf("%s failure: %T/%v", scenario, err, err)
			}
		})
	}
}

func TestConcurrentTraceEnumeratorDoesNotScanPastEmptyNeighbor(t *testing.T) {
	fp := int64(0)
	enumerator := &ConcurrentTraceEnumerator{enums: []*WorkerTraceEnumerator{concurrentEnumeratorReader(t, nil), concurrentEnumeratorReader(t, nil), concurrentEnumeratorReader(t, &fp)}}
	if pos := enumerator.NextPos(); pos != -1 || enumerator.idx != 1 {
		t.Fatalf("neighbor-only advancement: position %d, index %d", pos, enumerator.idx)
	}
	if pos := enumerator.NextPos(); pos != 42 || enumerator.idx != 2 {
		t.Fatalf("next invocation: position %d, index %d", pos, enumerator.idx)
	}
	if got, err := enumerator.NextFP(); err != nil || got != 0 {
		t.Fatalf("actual zero fingerprint: %d/%v", got, err)
	}
	if err := enumerator.Close(); err != nil {
		t.Fatal(err)
	}
}
