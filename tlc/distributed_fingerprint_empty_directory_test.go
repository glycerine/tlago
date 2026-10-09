package tlc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Keep every literal source path inside a temporary directory: an empty prefix
// followed by the separator and this root-relative name resolves there.
func TestDistributedFingerprintEmptyDirectoryRetainsLiteralPath(t *testing.T) {
	for _, backend := range []string{"mem", "mem1", "mem2", "lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			directory := t.TempDir()
			t.Chdir(directory)
			base := filepath.Join(directory, "Spec")
			name := strings.TrimPrefix(strings.TrimPrefix(base, filepath.VolumeName(base)), string(os.PathSeparator))
			newStore := func() FPSet {
				config := NewFPSetConfiguration()
				config.SetMemory(1 << 20)
				var store FPSet
				switch backend {
				case "mem":
					store = NewMemFPSet()
				case "mem1":
					store = NewMemFPSet1(config)
				case "mem2":
					store = &MemFPSet2{table: make([][]byte, 16), mask: 15}
				case "lsb":
					store = NewLSBDiskFPSet(config).DiskFPSet
				case "msb":
					store = NewMSBDiskFPSet(config).DiskFPSet
				}
				if err := invokeDistributedServerOperation(func() error {
					store.Init(1, "", name)
					return nil
				}); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(store.Close)
				return store
			}
			store := newStore()
			store.Put(1)
			if err := store.BeginChkptFile(name); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(base + ".fp.tmp"); err != nil {
				t.Fatalf("checkpoint ignored literal empty directory prefix: %v", err)
			}
			if err := store.CommitChkptFile(name); err != nil {
				t.Fatal(err)
			}
			store.Close()
			recovered := newStore()
			if err := recovered.RecoverFile(name); err != nil {
				t.Fatal(err)
			}
			count := uint64(1)
			if backend == "mem1" {
				count = 2 // Source restores count and then increments it through put.
			}
			if recovered.Size() != count || !recovered.Contains(1) {
				t.Fatal("literal empty-directory checkpoint lost recovered membership")
			}
		})
	}
}

func TestDFIDFingerprintEmptyDirectoryRetainsLiteralPath(t *testing.T) {
	directory := t.TempDir()
	t.Chdir(directory)
	base := filepath.Join(directory, "Spec")
	name := strings.TrimPrefix(strings.TrimPrefix(base, filepath.VolumeName(base)), string(os.PathSeparator))
	store := NewMemFPIntSetWithCapacity(4, 75).Init(1, "", name)
	if err := store.BeginChkptFile(name); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(base + ".fp.tmp"); err != nil {
		t.Fatal(err)
	}
	if err := store.CommitChkptFile(name); err != nil {
		t.Fatal(err)
	}
	if err := store.RecoverFile(name); err != nil {
		t.Fatal(err)
	}
}

func TestDFIDFingerprintCheckpointDoesNotCreateMissingParents(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "missing", "nested")
	store := NewMemFPIntSetWithCapacity(4, 75).Init(1, directory, "Spec")
	if err := store.BeginChkptFile("job"); err == nil || !os.IsNotExist(err) {
		t.Fatalf("checkpoint did not retain direct file-open failure: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(directory)); !os.IsNotExist(err) {
		t.Fatalf("failed checkpoint created parent directories: %v", err)
	}
	if store.Size() != 0 {
		t.Fatal("failed checkpoint changed fingerprint membership")
	}
}
