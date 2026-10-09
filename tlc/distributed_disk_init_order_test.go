package tlc

import (
	"os"
	"path/filepath"
	"testing"
)

// Source DiskFPSet.init assigns filenames, allocates reader arrays and resets
// poolIndex before opening/truncating storage. It does not reset membership or
// index/count metadata on repeated initialization. No original test directly
// checks these partial mutations or native handle retirement.
func TestDistributedDiskFingerprintInitPreservesSourceOrder(t *testing.T) {
	for _, backend := range []string{"lsb", "msb"} {
		for _, phase := range []string{"negative_workers", "missing_parent", "reinitialize"} {
			t.Run(backend+"/"+phase, func(t *testing.T) {
				config := NewFPSetConfiguration()
				config.SetMemory(1 << 20)
				var set *DiskFPSet
				if backend == "lsb" {
					set = NewLSBDiskFPSet(config).DiskFPSet
				} else {
					set = NewMSBDiskFPSet(config).DiskFPSet
				}
				set.Init(1, t.TempDir(), "old")
				t.Cleanup(set.Close)
				set.Put(41)
				if err := set.flushTable(); err != nil {
					t.Fatal(err)
				}
				set.Put(97)
				oldIndex := set.index
				oldWorkers, oldPool := set.braf, set.brafPool
				set.poolIndex = 3
				directory := t.TempDir()
				workers := 2
				if phase == "negative_workers" {
					workers = -1
				} else if phase == "missing_parent" {
					directory = filepath.Join(directory, "missing")
				}
				err := invokeDistributedServerOperation(func() error {
					set.Init(workers, directory, "new")
					return nil
				})
				base := directory + string(os.PathSeparator) + "new"
				if set.metadir != directory || set.tmpFilename != base+".tmp" || set.fpFilename != base+".fp" {
					t.Fatal("initialization failure preceded source filename assignments")
				}
				if set.GetFileCnt() != 1 || set.GetTblCnt() != 1 || len(set.index) != len(oldIndex) || &set.index[0] != &oldIndex[0] {
					t.Fatal("initialization reset source membership/index metadata")
				}
				if phase == "negative_workers" {
					if _, ok := err.(*NegativeArraySizeException); !ok {
						t.Fatalf("negative array failure = %T/%v", err, err)
					}
					if set.braf[0] != oldWorkers[0] || set.brafPool[0] != oldPool[0] || set.poolIndex != 3 || oldWorkers[0].closed {
						t.Fatal("negative array allocation changed prior reader owners")
					}
					if _, err := os.Stat(base + ".fp"); !os.IsNotExist(err) {
						t.Fatal("negative array failure opened backing storage")
					}
				} else {
					if len(set.braf) != 2 || len(set.brafPool) != 5 || set.poolIndex != 0 {
						t.Fatal("backing-file open preceded source reader allocation")
					}
					for _, reader := range append(append([]*BufferedRandomAccessFile{}, oldWorkers...), oldPool...) {
						if !reader.closed {
							t.Fatal("replacement leaked a native reader owner")
						}
					}
					if phase == "missing_parent" {
						if !isJavaIOException(err) {
							t.Fatalf("backing-file open failure = %T/%v", err, err)
						}
						for _, reader := range append(append([]*BufferedRandomAccessFile{}, set.braf...), set.brafPool...) {
							if reader != nil {
								t.Fatal("failed create populated a reader slot")
							}
						}
					} else {
						if err != nil {
							t.Fatal(err)
						}
						if info, err := os.Stat(base + ".fp"); err != nil || info.Size() != 0 {
							t.Fatalf("reinitialization did not create empty backing storage: %v", err)
						}
					}
				}
			})
		}
	}
}
