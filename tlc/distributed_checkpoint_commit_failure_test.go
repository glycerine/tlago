package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

// No enabled original method covers these checkpoint commit failure boundaries.
func TestDistributedCheckpointCommitRetainsIOAndFileMutation(t *testing.T) {
	for _, owner := range []string{"trace", "memory_queue", "disk_queue"} {
		for _, phase := range []string{"missing_temporary", "blocked_delete", "success"} {
			t.Run(owner+"/"+phase, func(t *testing.T) {
				directory := t.TempDir()
				var commit func() error
				var oldPath, newPath, message string
				switch owner {
				case "trace":
					trace := NewTLCTrace(directory, "Spec")
					defer trace.Close()
					commit = trace.CommitChkpt
					oldPath, newPath = trace.chkptName("chkpt"), trace.chkptName("tmp")
					message = "Trace.commitChkpt: cannot delete "
				case "memory_queue":
					queue := NewMemStateQueue(directory)
					commit = queue.CommitChkpt
					oldPath, newPath = filepath.Join(directory, "queue.chkpt"), filepath.Join(directory, "queue.tmp")
					message = "MemStateQueue.commitChkpt: cannot delete "
				case "disk_queue":
					queue := &DiskStateQueue{diskdir: directory}
					commit = queue.CommitChkpt
					oldPath, newPath = queue.queuePath("queue.chkpt"), queue.queuePath("queue.tmp")
					message = "DiskStateQueue.commitChkpt: cannot delete "
				}
				if phase == "blocked_delete" {
					if err := os.Mkdir(oldPath, 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(oldPath, "retained"), []byte("old"), 0600); err != nil {
						t.Fatal(err)
					}
				} else if err := os.WriteFile(oldPath, []byte("old"), 0600); err != nil {
					t.Fatal(err)
				}
				if phase != "missing_temporary" {
					if err := os.WriteFile(newPath, []byte("new"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				err := commit()
				if phase == "success" {
					if err != nil {
						t.Fatal(err)
					}
					data, err := os.ReadFile(oldPath)
					if err != nil || string(data) != "new" {
						t.Fatalf("promoted checkpoint = %q/%v", data, err)
					}
					if _, err := os.Stat(newPath); !os.IsNotExist(err) {
						t.Fatalf("temporary not promoted: %v", err)
					}
					return
				}
				if !isJavaIOException(err) || err.Error() != message+oldPath {
					t.Fatalf("commit failure = %T %v", err, err)
				}
				payload, encodeErr := EncodeDistributedFailure(err)
				if encodeErr != nil {
					t.Fatal(encodeErr)
				}
				decoded, decodeErr := DecodeDistributedFailure(payload)
				if decodeErr != nil || !isJavaIOException(decoded) {
					t.Fatalf("native I/O category = %v/%v", decoded, decodeErr)
				}
				if phase == "missing_temporary" {
					if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
						t.Fatalf("old checkpoint not deleted before failed promotion: %v", err)
					}
				} else {
					data, err := os.ReadFile(newPath)
					if err != nil || string(data) != "new" {
						t.Fatal("delete failure changed temporary checkpoint")
					}
					data, err = os.ReadFile(filepath.Join(oldPath, "retained"))
					if err != nil || string(data) != "old" {
						t.Fatal("delete failure changed old checkpoint")
					}
				}
			})
		}
	}
}

func TestDistributedDiskQueueCommitRetainsPartialPoolDeletion(t *testing.T) {
	directory := t.TempDir()
	queue := &DiskStateQueue{diskdir: directory, lastLoPool: 0, newLastLoPool: 3}
	for _, path := range []string{queue.poolName(0), queue.poolName(2), queue.queuePath("queue.chkpt"), queue.queuePath("queue.tmp")} {
		if err := os.WriteFile(path, []byte("retained"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	err := queue.CommitChkpt()
	if !isJavaIOException(err) || err.Error() != "DiskStateQueue.commitChkpt: cannot delete "+queue.poolName(1) {
		t.Fatalf("pool deletion failure = %v", err)
	}
	if queue.lastLoPool != 0 || queue.newLastLoPool != 3 {
		t.Fatal("failed deletion advanced pool checkpoint markers")
	}
	if _, err := os.Stat(queue.poolName(0)); !os.IsNotExist(err) {
		t.Fatal("earlier deletion was rolled back")
	}
	for _, path := range []string{queue.poolName(2), queue.queuePath("queue.chkpt"), queue.queuePath("queue.tmp")} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != "retained" {
			t.Fatalf("later file changed: %s", path)
		}
	}
}

type missingCommitTemporaryQueue struct {
	*MemStateQueue
	phase, traceTemporary string
}

func (q *missingCommitTemporaryQueue) BeginChkpt() error {
	if err := q.MemStateQueue.BeginChkpt(); err != nil {
		return err
	}
	if q.phase == "queue" {
		return os.Remove(filepath.Join(q.diskdir, "queue.tmp"))
	}
	return nil
}
func (q *missingCommitTemporaryQueue) CommitChkpt() error {
	if err := q.MemStateQueue.CommitChkpt(); err != nil {
		return err
	}
	if q.phase == "trace" {
		return os.Remove(q.traceTemporary)
	}
	return nil
}

func TestDistributedCoordinatorCommitFailureStopsLaterPhases(t *testing.T) {
	for _, phase := range []string{"queue", "trace"} {
		t.Run(phase, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			oldWorkers := NumWorkers()
			SetNumWorkers(0)
			defer SetNumWorkers(oldWorkers)
			directory := t.TempDir()
			trace := NewTLCTrace(directory, "Spec")
			defer trace.Close()
			queue := &missingCommitTemporaryQueue{MemStateQueue: NewMemStateQueue(directory), phase: phase, traceTemporary: trace.chkptName("tmp")}
			store := NewMemFPSet()
			store.Init(1, directory, "Spec")
			manager := NewNonDistributedFPSetManager(store, "local", trace)
			server := NewTLCServer("Spec", "Spec", directory, manager, queue, trace)
			recorder := &MemoryRecorder{}
			AddMessageRecorder(recorder)
			defer RemoveMessageRecorder(recorder)
			err := server.Checkpoint()
			if !isJavaIOException(err) {
				t.Fatalf("coordinator commit failure = %v", err)
			}
			if queue.stop || !recorder.Recorded(ECTLCCheckpointStart) || recorder.Recorded(ECTLCCheckpointEnd) {
				t.Fatal("failure changed queue resume or checkpoint completion output")
			}
			for _, path := range []string{store.chkptName("Spec", "tmp"), uniqueStringChkptName(directory, "tmp")} {
				if _, err := os.Stat(path); err != nil {
					t.Fatalf("earlier checkpoint begin missing: %s/%v", path, err)
				}
			}
			for _, path := range []string{store.chkptName("Spec", "chkpt"), uniqueStringChkptName(directory, "chkpt"), trace.chkptName("chkpt")} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("later commit ran: %s/%v", path, err)
				}
			}
			_, queueErr := os.Stat(filepath.Join(directory, "queue.chkpt"))
			if phase == "queue" {
				if !os.IsNotExist(queueErr) {
					t.Fatal("failed queue checkpoint was promoted")
				}
				if _, err := os.Stat(trace.chkptName("tmp")); err != nil {
					t.Fatal("queue failure changed earlier trace begin")
				}
			} else if queueErr != nil {
				t.Fatal("trace failure lost earlier queue commit")
			}
		})
	}
}
