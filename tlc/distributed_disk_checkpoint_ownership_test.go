package tlc

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func diskCheckpointOwnershipSet(t *testing.T) *DiskFPSet {
	t.Helper()
	t.Setenv(DiskFPSetLogLockCntProperty, "1")
	config := NewFPSetConfiguration()
	config.SetMemory(1 << 20)
	set := NewDiskFPSet(config)
	set.Init(1, t.TempDir(), "store")
	t.Cleanup(func() {
		// Release the tested source failure state only for owned fixture cleanup.
		// Never attempt a storage operation while its retained locks are held.
		if set.flusherChosen.Load() {
			set.releaseTblWriteLock()
			set.flusherChosen.Store(false)
		}
		set.Close()
	})
	return set
}

func checkDiskCheckpointOwnership(t *testing.T, set *DiskFPSet, held bool) {
	t.Helper()
	if set.flusherChosen.Load() != held {
		t.Fatalf("flusher flag = %v, want %v", set.flusherChosen.Load(), held)
	}
	for i := 0; i < set.rwLock.Size(); i++ {
		lock := set.rwLock.GetAt(i)
		acquired := lock.TryLock()
		if acquired {
			lock.Unlock()
		}
		if acquired == held {
			t.Fatalf("stripe %d retained = %v, want %v", i, !acquired, held)
		}
	}
}

// Source beginChkpt(String) releases locks/resets the flusher only after
// flush/copy/marker succeed. No enabled original method covers failure ownership.
func TestDistributedDiskCheckpointRetainsFailureOwnership(t *testing.T) {
	for _, phase := range []string{"flush_io", "flush_runtime", "copy_io", "success"} {
		t.Run(phase, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			set := diskCheckpointOwnershipSet(t)
			set.Put(41)
			switch phase {
			case "flush_io":
				if err := os.Mkdir(set.tmpFilename, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(set.tmpFilename, "retained"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			case "flush_runtime":
				if err := set.flushTable(); err != nil {
					t.Fatal(err)
				}
				// Restore an unflushed duplicate in the table, as the existing
				// duplicate-merge check does; memInsert recognizes marked entries.
				set.tbl[set.getIndex(41)][0] = 41
				set.tblCnt = 1
			}
			name := "job"
			if phase == "copy_io" {
				name = filepath.Join("missing", "job")
			}
			err := set.BeginChkptFile(name)
			if phase == "success" {
				if err != nil {
					t.Fatal(err)
				}
				if set.checkPointMark != 1 {
					t.Fatal("successful copy did not advance marker")
				}
				data, err := os.ReadFile(set.chkptName(name, "tmp"))
				if err != nil || len(data) != 8 || binary.BigEndian.Uint64(data) != 41 {
					t.Fatalf("checkpoint data = %x/%v", data, err)
				}
				checkDiskCheckpointOwnership(t, set, false)
				return
			}
			if phase == "flush_runtime" {
				failure, ok := err.(*TLCError)
				if !ok || !failure.Runtime || failure.Code != ECTLCFPValueAlreadyOnDisk || isJavaIOException(err) {
					t.Fatalf("flush runtime failure = %#v", err)
				}
			} else if !isJavaIOException(err) {
				t.Fatalf("checkpoint I/O failure = %v", err)
			}
			if set.checkPointMark != 0 {
				t.Fatal("failed checkpoint advanced marker")
			}
			if phase == "copy_io" {
				if set.GetFileCnt() != 1 || set.GetTblCnt() != 0 || set.GetDiskWriteCnt() != 1 {
					t.Fatal("failed copy lost completed flush")
				}
			} else if set.GetTblCnt() != 1 {
				t.Fatal("failed flush reset table count")
			}
			checkDiskCheckpointOwnership(t, set, true)
		})
	}
}

func TestDistributedDiskCheckpointFailureContinuesToHealthyPartition(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	failed := diskCheckpointOwnershipSet(t)
	failed.Put(41)
	// Existing nonempty directory cannot be replaced by FileUtil.copyFile.
	target := failed.chkptName("job", "tmp")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "retained"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	healthy := NewMemFPSet()
	healthy.Init(1, t.TempDir(), "store")
	healthy.Put(97)
	_, first := startFingerprintRPC(t, NewLocalFingerprintEndpoint(failed))
	_, second := startFingerprintRPC(t, NewLocalFingerprintEndpoint(healthy))
	manager := NewDistributedFPSetManager(first, second)
	manager.fpSets[0].hostname = "failed-copy"
	firstEntry, secondEntry := manager.entry(0), manager.entry(1)
	if err := manager.Checkpoint("job"); err != nil {
		t.Fatal(err)
	}
	if manager.entry(0) != firstEntry || manager.entry(1) != secondEntry || !firstEntry.available || !secondEntry.available {
		t.Fatal("checkpoint failure reassigned partitions")
	}
	checkDiskCheckpointOwnership(t, failed, true)
	if failed.checkPointMark != 0 || failed.GetTblCnt() != 0 || failed.GetFileCnt() != 1 {
		t.Fatal("manager changed failed checkpoint mutation")
	}
	data, err := os.ReadFile(healthy.chkptName("job", "chkpt"))
	if err != nil || len(data) != 8 || binary.BigEndian.Uint64(data) != 97 {
		t.Fatalf("healthy partition commit = %x/%v", data, err)
	}
	messages := ToolIOGetAllMessages()
	if len(messages) != 1 || !strings.Contains(messages[0], "Failed to checkpoint the fingerprint server at failed-copy.") {
		t.Fatalf("manager diagnostic = %v", messages)
	}
}
