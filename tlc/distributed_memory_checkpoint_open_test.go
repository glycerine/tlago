package tlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Source file-output constructors open the supplied checkpoint path without
// creating parent directories. No enabled original method covers this boundary.
func TestDistributedMemoryCheckpointDoesNotCreateMissingParents(t *testing.T) {
	for _, implementation := range []string{"memory", "memory1", "memory2"} {
		for _, remote := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/remote=%v", implementation, remote), func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				directory := filepath.Join(t.TempDir(), "missing", "nested")
				var set FPSet
				switch implementation {
				case "memory":
					set = NewMemFPSet()
				case "memory1":
					set = NewMemFPSet1(NewFPSetConfiguration())
				case "memory2":
					set = &MemFPSet2{}
				}
				set.Init(1, directory, "store")
				if remote {
					_, client := startFingerprintRPC(t, NewLocalFingerprintEndpoint(set))
					manager := NewDistributedFPSetManager(client)
					if err := manager.Checkpoint("job"); err != nil {
						t.Fatal(err)
					}
					messages := ToolIOGetAllMessages()
					if len(messages) != 1 || !strings.Contains(messages[0], "Failed to checkpoint the fingerprint server") {
						t.Fatalf("missing parent source I/O catch = %v", messages)
					}
				} else {
					if err := set.BeginChkptFile("job"); !isJavaIOException(err) || !os.IsNotExist(err) {
						t.Fatalf("missing parent open error = %v", err)
					}
				}
				if _, err := os.Stat(directory); !os.IsNotExist(err) {
					t.Fatalf("checkpoint created missing parent: %v", err)
				}
				if set.Size() != 0 {
					t.Fatal("failed checkpoint changed empty storage")
				}
			})
		}
	}
}

func TestDistributedMemoryBackingSetConstructorSize(t *testing.T) {
	zero := NewSetOfLong(0)
	if zero.length != 0 || len(zero.table) != 0 || zero.thresh != 0 || zero.count != 0 || zero.hasZero {
		t.Fatalf("empty source allocation = %#v", zero)
	}
	if zero.Put(41) || zero.length != 1 || zero.Size() != 1 || !zero.Contains(41) {
		t.Fatal("zero-sized set did not grow on first insertion")
	}
	defer func() {
		if failure := recover(); failure == nil {
			t.Fatal("negative constructor size accepted")
		} else if _, ok := failure.(*NegativeArraySizeException); !ok {
			t.Fatalf("negative constructor failure = %v", failure)
		}
	}()
	NewSetOfLong(-1)
}
