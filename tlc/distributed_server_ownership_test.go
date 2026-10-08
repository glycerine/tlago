package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

// The source has no direct tests for missing coordinator components. These
// checks preserve its failure boundary and the mutations before that boundary.
type ownershipCheckpointQueue struct {
	*MemStateQueue
	suspend   bool
	resumed   bool
	recovered bool
}

func (q *ownershipCheckpointQueue) SuspendAll() bool { return q.suspend }
func (q *ownershipCheckpointQueue) ResumeAll()       { q.resumed = true }
func (q *ownershipCheckpointQueue) Recover() error {
	q.recovered = true
	return q.MemStateQueue.Recover()
}

func requireCoordinatorNullFailure(t *testing.T, call func() error) {
	t.Helper()
	defer func() {
		failure := recover()
		if _, ok := failure.(*NullPointerException); !ok {
			t.Fatalf("missing component failure = %T/%v", failure, failure)
		}
	}()
	if err := call(); err != nil {
		t.Fatalf("missing component returned %v instead of failing at the access", err)
	}
}

func TestDistributedCoordinatorRequiredComponents(t *testing.T) {
	for _, call := range []func() error{
		(*TLCServer)(nil).Checkpoint,
		(*TLCServer)(nil).Recover,
		func() error { return (*TLCServer)(nil).Close(true) },
		(&TLCServer{}).Checkpoint,
		(&TLCServer{}).Recover,
		func() error { return (&TLCServer{}).Close(true) },
	} {
		requireCoordinatorNullFailure(t, call)
	}
}

func TestDistributedCheckpointMissingOwnerRetainsBeginOrder(t *testing.T) {
	for _, missing := range []string{"trace", "fingerprints"} {
		t.Run(missing, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			directory := t.TempDir()
			queue := &ownershipCheckpointQueue{MemStateQueue: NewMemStateQueue(directory), suspend: true}
			server := &TLCServer{StateQueue: queue, Metadir: directory, FileName: "Spec"}
			if missing == "fingerprints" {
				server.Trace = NewTLCTrace(directory, "Spec")
				defer server.Trace.Close()
			}
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			defer RemoveMessageRecorder(recorder)
			requireCoordinatorNullFailure(t, server.Checkpoint)
			if queue.resumed || !recorder.Recorded(ECTLCCheckpointStart) || recorder.Recorded(ECTLCCheckpointEnd) {
				t.Fatal("missing owner resumed the queue or changed checkpoint reporting")
			}
			if _, err := os.Stat(filepath.Join(directory, "queue.tmp")); err != nil {
				t.Fatal("queue begin did not precede the missing owner", err)
			}
			if missing == "fingerprints" {
				if _, err := os.Stat(server.Trace.chkptName("tmp")); err != nil {
					t.Fatal("trace begin did not precede the missing fingerprint manager", err)
				}
			}
			for _, path := range []string{"queue.chkpt", "Spec.st.chkpt", "vars.tmp", "vars.chkpt"} {
				if _, err := os.Stat(filepath.Join(directory, path)); !os.IsNotExist(err) {
					t.Fatalf("later checkpoint phase ran: %s/%v", path, err)
				}
			}
		})
	}
	queue := &ownershipCheckpointQueue{MemStateQueue: NewMemStateQueue(t.TempDir())}
	if err := (&TLCServer{StateQueue: queue}).Checkpoint(); err != nil || queue.resumed {
		t.Fatal("failed suspension reached later owners", err)
	}
}

func TestDistributedRecoveryMissingOwnerRetainsReadOrder(t *testing.T) {
	for _, missing := range []string{"trace", "queue", "fingerprints"} {
		t.Run(missing, func(t *testing.T) {
			directory := t.TempDir()
			trace := NewTLCTrace(directory, "Spec")
			defer trace.Close()
			trace.lastPtr = 73
			if err := trace.BeginChkpt(); err != nil {
				t.Fatal(err)
			}
			if err := trace.CommitChkpt(); err != nil {
				t.Fatal(err)
			}
			trace.lastPtr = 99
			queue := &ownershipCheckpointQueue{MemStateQueue: NewMemStateQueue(directory)}
			if err := queue.BeginChkpt(); err != nil {
				t.Fatal(err)
			}
			if err := queue.CommitChkpt(); err != nil {
				t.Fatal(err)
			}
			server := &TLCServer{Trace: trace, StateQueue: queue, Metadir: directory, FileName: "Spec"}
			if missing == "trace" {
				server.Trace = nil
			}
			if missing == "queue" {
				server.StateQueue = nil
			}
			requireCoordinatorNullFailure(t, server.Recover)
			wantPointer := int64(73)
			if missing == "trace" {
				wantPointer = 99
			}
			if trace.lastPtr != wantPointer || queue.recovered != (missing == "fingerprints") {
				t.Fatalf("recovery advanced out of order: pointer %d, queue recovered %v", trace.lastPtr, queue.recovered)
			}
		})
	}
}

func TestDistributedCloseMissingOwnerRetainsCleanupOrder(t *testing.T) {
	for _, missing := range []string{"trace", "fingerprints"} {
		t.Run(missing, func(t *testing.T) {
			directory := t.TempDir()
			trace := NewTLCTrace(directory, "Spec")
			defer trace.Close()
			server := &TLCServer{Trace: trace, Metadir: directory}
			if missing == "trace" {
				server.Trace = nil
			}
			requireCoordinatorNullFailure(t, func() error { return server.Close(true) })
			if trace.closed != (missing == "fingerprints") {
				t.Fatal("trace close did not precede missing fingerprint manager")
			}
			if _, err := os.Stat(directory); err != nil {
				t.Fatal("failed close deleted metadata", err)
			}
		})
	}
}
