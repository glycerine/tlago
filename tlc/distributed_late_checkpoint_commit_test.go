package tlc

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// Remove a pending file after the real queue commit, before the later owners
// commit. The checkpoint implementation itself has no injection hook.
type lateCheckpointFailureQueue struct {
	*MemStateQueue
	remove string
}

func (q *lateCheckpointFailureQueue) CommitChkpt() error {
	if err := q.MemStateQueue.CommitChkpt(); err != nil {
		return err
	}
	if q.remove != "" {
		return os.Remove(q.remove)
	}
	return nil
}

// No original Java method directly covers these late commit failures. Preserve
// TLCServer.checkpoint's order and the local/remote manager distinction: remote
// stores commit during checkpoint(), local storage commits last.
func TestDistributedLateCheckpointCommitFailure(t *testing.T) {
	for _, scenario := range []string{"local_intern", "remote_intern", "local_fingerprint"} {
		t.Run(scenario, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			oldWorkers := NumWorkers()
			SetNumWorkers(0)
			defer SetNumWorkers(oldWorkers)
			directory := t.TempDir()
			trace := NewTLCTrace(directory, "Spec")
			defer trace.Close()
			queue := &lateCheckpointFailureQueue{MemStateQueue: NewMemStateQueue(directory)}
			store := NewMemFPSet()
			store.Init(1, directory, "Spec")
			defer store.Close()
			store.Put(41)
			manager := NewNonDistributedFPSetManager(store, "local", trace)
			if scenario == "remote_intern" {
				_, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(store))
				manager = NewDistributedFPSetManager(client)
			}
			server := NewTLCServer("Spec", "Spec", directory, manager, queue, trace)
			server.InternTable = NewInternTable(16)
			server.InternTable.Put("before")
			if err := server.Checkpoint(); err != nil {
				t.Fatal(err)
			}
			read := func(path string) []byte {
				t.Helper()
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				return data
			}
			fpCommitted := store.chkptName("Spec", "chkpt")
			fpTemporary := store.chkptName("Spec", "tmp")
			internCommitted := uniqueStringChkptName(directory, "chkpt")
			internTemporary := uniqueStringChkptName(directory, "tmp")
			oldFP := read(fpCommitted)
			oldIntern := read(internCommitted)
			store.Put(97)
			server.InternTable.Put("after")
			queue.remove = internTemporary
			wantMessage := "InternTable.commitChkpt: cannot delete " + internCommitted
			if scenario == "local_fingerprint" {
				queue.remove = fpTemporary
				wantMessage = "MemFPSet.commitChkpt: cannot delete " + fpCommitted
			}
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			defer RemoveMessageRecorder(recorder)
			err := server.Checkpoint()
			if !isJavaIOException(err) || err.Error() != wantMessage {
				t.Fatalf("late commit failure = %T/%v, want %q", err, err, wantMessage)
			}
			if queue.stop || !recorder.Recorded(ECTLCCheckpointStart) || recorder.Recorded(ECTLCCheckpointEnd) {
				t.Fatal("late failure changed queue resume or emitted checkpoint completion")
			}
			absent := func(path string) {
				t.Helper()
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("expected absent checkpoint %s: %v", path, err)
				}
			}
			// Earlier queue and trace promotions remain committed on all paths.
			read(filepath.Join(directory, "queue.chkpt"))
			read(trace.chkptName("chkpt"))
			absent(filepath.Join(directory, "queue.tmp"))
			absent(trace.chkptName("tmp"))
			absent(internTemporary)
			if scenario == "local_fingerprint" {
				if bytes.Equal(read(internCommitted), oldIntern) {
					t.Fatal("fingerprint failure lost the earlier intern commit")
				}
				absent(fpCommitted) // Delete precedes the failed rename.
				absent(fpTemporary)
			} else {
				absent(internCommitted) // Delete precedes the failed rename.
				if scenario == "remote_intern" {
					if data := read(fpCommitted); bytes.Equal(data, oldFP) || len(data) != 16 {
						t.Fatal("intern failure rolled back the earlier remote fingerprint commit")
					}
					absent(fpTemporary)
				} else {
					if !bytes.Equal(read(fpCommitted), oldFP) || len(read(fpTemporary)) != 16 {
						t.Fatal("intern failure changed the older local fingerprint commit or pending generation")
					}
				}
			}
			if store.Size() != 2 || !store.Contains(41) || !store.Contains(97) {
				t.Fatal("checkpoint failure changed live fingerprint membership")
			}
		})
	}
}
