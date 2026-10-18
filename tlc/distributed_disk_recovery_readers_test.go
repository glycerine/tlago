package tlc

import (
	"encoding/binary"
	"os"
	"testing"
)

// Source DiskFPSet.recover(String) replaces each reader immediately after its
// close, and resets poolIndex only after both loops. No enabled Java method
// checks failure partway through these loops.
func TestDistributedDiskRecoveryPreservesReaderReplacementOrder(t *testing.T) {
	for _, phase := range []string{"worker_close", "pool_close", "success"} {
		t.Run(phase, func(t *testing.T) {
			config := NewFPSetConfiguration()
			config.SetMemory(1 << 20)
			set := NewDiskFPSet(config)
			set.Init(3, t.TempDir(), "store")
			defer set.Close()
			workers := append([]*BufferedRandomAccessFile(nil), set.braf...)
			pool := append([]*BufferedRandomAccessFile(nil), set.brafPool...)
			set.poolIndex = 2
			var data [16]byte
			binary.BigEndian.PutUint64(data[:8], 41)
			binary.BigEndian.PutUint64(data[8:], 97)
			if err := os.WriteFile(set.chkptName("job", "chkpt"), data[:], 0600); err != nil {
				t.Fatal(err)
			}
			if phase == "worker_close" {
				if err := workers[1].file.Close(); err != nil {
					t.Fatal(err)
				}
			} else if phase == "pool_close" {
				if err := pool[1].file.Close(); err != nil {
					t.Fatal(err)
				}
			}
			err := set.RecoverFile("job")
			if phase == "success" {
				if err != nil {
					t.Fatal(err)
				}
			} else if !isJavaIOException(err) {
				t.Fatalf("reader close failure = %v", err)
			}
			if len(set.braf) != len(workers) || len(set.brafPool) != len(pool) {
				t.Fatal("recovery cleared or resized source reader arrays")
			}
			if set.GetFileCnt() != 2 || set.GetDiskWriteCnt() != 2 || len(set.index) != 2 || set.index[0] != 41 || set.index[1] != 97 {
				t.Fatal("reader failure changed completed file/index recovery")
			}
			for i, old := range workers {
				replaced := phase != "worker_close" || i < 1
				if replaced {
					if set.braf[i] == old || !old.closed || set.braf[i].closed {
						t.Fatalf("worker %d was not replaced in source order", i)
					}
				} else if set.braf[i] != old || old.closed != (i == 1) {
					t.Fatalf("worker %d changed after failure", i)
				}
			}
			for i, old := range pool {
				replaced := phase == "success" || phase == "pool_close" && i < 1
				if replaced {
					if set.brafPool[i] == old || !old.closed || set.brafPool[i].closed {
						t.Fatalf("pool %d was not replaced in source order", i)
					}
				} else if set.brafPool[i] != old || old.closed != (phase == "pool_close" && i == 1) {
					t.Fatalf("pool %d changed after failure", i)
				}
			}
			wantPoolIndex := 2
			if phase == "success" {
				wantPoolIndex = 0
			}
			if set.poolIndex != wantPoolIndex {
				t.Fatalf("pool cursor = %d, want %d", set.poolIndex, wantPoolIndex)
			}
			snapshot := set.readers.Load()
			if snapshot == nil || len(*snapshot) != len(set.braf) {
				t.Fatal("native reader snapshot lost partial recovery")
			}
			for i, reader := range *snapshot {
				if reader != set.braf[i] {
					t.Fatalf("native reader snapshot %d differs from source slot", i)
				}
			}
		})
	}
}

func TestDistributedDiskRecoveryReaderOpenFailurePreservesSlots(t *testing.T) {
	config := NewFPSetConfiguration()
	config.SetMemory(1 << 20)
	set := NewDiskFPSet(config)
	set.Init(2, t.TempDir(), "store")
	defer set.Close()
	workers := append([]*BufferedRandomAccessFile(nil), set.braf...)
	pool := append([]*BufferedRandomAccessFile(nil), set.brafPool...)
	set.poolIndex = 2
	set.fpFilename += ".missing"
	if err := set.reopenBRAFReaders(); !isJavaIOException(err) {
		t.Fatalf("reopen failure = %v", err)
	}
	if len(set.braf) != len(workers) || len(set.brafPool) != len(pool) || set.poolIndex != 2 {
		t.Fatal("failed open cleared reader arrays or reset pool cursor")
	}
	for i, old := range workers {
		if set.braf[i] != old || old.closed != (i == 0) {
			t.Fatalf("worker %d changed after failed open", i)
		}
	}
	for i, old := range pool {
		if set.brafPool[i] != old || old.closed {
			t.Fatalf("pool %d changed after failed open", i)
		}
	}
	snapshot := set.readers.Load()
	if snapshot == nil || len(*snapshot) != len(workers) {
		t.Fatal("native reader snapshot lost source slots")
	}
	for i, old := range workers {
		if (*snapshot)[i] != old {
			t.Fatalf("snapshot %d changed", i)
		}
	}
}
