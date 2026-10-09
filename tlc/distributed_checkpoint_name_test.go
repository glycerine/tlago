package tlc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Source named checkpoint methods concatenate the supplied name literally.
// Empty names must not alias "fpset" or the initialized store's name.
func TestDistributedFingerprintCheckpointEmptyName(t *testing.T) {
	for _, backend := range []string{"mem", "mem1", "mem2", "lsb", "msb"} {
		for _, remote := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/remote=%v", backend, remote), func(t *testing.T) {
				directory := t.TempDir()
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
						// Only low fingerprints are needed for this file naming check.
						store = &MemFPSet2{table: make([][]byte, 16), mask: 15}
					case "lsb":
						store = NewLSBDiskFPSet(config).DiskFPSet
					case "msb":
						store = NewMSBDiskFPSet(config).DiskFPSet
					}
					store.Init(1, directory, "Spec")
					t.Cleanup(store.Close)
					return store
				}
				store := newStore()
				store.Put(1)
				var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(store)
				if remote {
					_, endpoint = startFingerprintRPC(t, endpoint)
				}
				if err := endpoint.BeginChkptFile("fpset"); err != nil {
					t.Fatal(err)
				}
				if err := endpoint.CommitChkptFile("fpset"); err != nil {
					t.Fatal(err)
				}
				prior, err := os.ReadFile(filepath.Join(directory, "fpset.fp.chkpt"))
				if err != nil {
					t.Fatal(err)
				}
				store.Put(2)
				if err := endpoint.BeginChkptFile(""); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(filepath.Join(directory, ".fp.tmp")); err != nil {
					t.Fatalf("empty named checkpoint did not use literal path: %v", err)
				}
				if err := endpoint.CommitChkptFile(""); err != nil {
					t.Fatal(err)
				}
				if data, err := os.ReadFile(filepath.Join(directory, "fpset.fp.chkpt")); err != nil || !bytes.Equal(data, prior) {
					t.Fatal("empty named checkpoint overwrote the fpset checkpoint")
				}
				if _, err := os.Stat(filepath.Join(directory, "Spec.fp.chkpt")); !os.IsNotExist(err) {
					t.Fatalf("explicit empty name selected the initialized name: %v", err)
				}
				store.Close()
				recovered := newStore()
				endpoint = NewLocalFingerprintEndpoint(recovered)
				if remote {
					_, endpoint = startFingerprintRPC(t, endpoint)
				}
				if err := endpoint.RecoverFile(""); err != nil {
					t.Fatal(err)
				}
				count := uint64(2)
				if backend == "mem1" {
					// Source SetOfLong.recover restores count, then calls put for
					// each serialized nonzero key, incrementing it again.
					count = 4
				}
				if recovered.Size() != count || !recovered.Contains(1) || !recovered.Contains(2) {
					t.Fatal("empty named recovery selected the wrong checkpoint")
				}
			})
		}
	}
}

func TestDFIDFingerprintCheckpointEmptyName(t *testing.T) {
	directory := t.TempDir()
	store := NewMemFPIntSetWithCapacity(4, 75).Init(1, directory, "Spec")
	if err := store.BeginChkptFile(""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, ".fp.tmp")); err != nil {
		t.Fatalf("empty DFID checkpoint did not use literal path: %v", err)
	}
	if err := store.CommitChkptFile(""); err != nil {
		t.Fatal(err)
	}
	if err := store.RecoverFile(""); err != nil {
		t.Fatal(err)
	}
}

func TestDistributedDiskFingerprintEmptyBackingName(t *testing.T) {
	for _, backend := range []string{"lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			directory := t.TempDir()
			config := NewFPSetConfiguration()
			config.SetMemory(1 << 20)
			var store *DiskFPSet
			if backend == "lsb" {
				store = NewLSBDiskFPSet(config).DiskFPSet
			} else {
				store = NewMSBDiskFPSet(config).DiskFPSet
			}
			store.Init(1, directory, "")
			t.Cleanup(store.Close)
			if _, err := os.Stat(filepath.Join(directory, ".fp")); err != nil {
				t.Fatalf("empty backing name did not use literal path: %v", err)
			}
			if _, err := os.Stat(filepath.Join(directory, "fpset.fp")); !os.IsNotExist(err) {
				t.Fatalf("empty backing name created an invented filename: %v", err)
			}
			store.Put(1)
			if err := store.BeginChkptFile(""); err != nil {
				t.Fatal(err)
			}
			if err := store.CommitChkptFile(""); err != nil {
				t.Fatal(err)
			}
			if data, err := os.ReadFile(filepath.Join(directory, ".fp.chkpt")); err != nil || len(data) != 8 {
				t.Fatalf("literal backing file was not checkpointed: %x/%v", data, err)
			}
		})
	}
}
