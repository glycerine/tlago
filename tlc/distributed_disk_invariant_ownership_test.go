package tlc

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// No original method directly checks failure ownership in DiskFPSet.checkInvariant.
// Failures before its try/finally retain the table locks; scan failures close the
// file and release them. Inspect ownership without starting blocked storage work.
func TestDistributedDiskInvariantFailureOwnership(t *testing.T) {
	for _, backend := range []string{"lsb", "msb"} {
		for _, remote := range []bool{false, true} {
			for _, phase := range []string{"flush_io", "open_io", "scan_io", "invalid", "valid"} {
				name := backend + "/" + map[bool]string{false: "local", true: "tcp"}[remote] + "/" + phase
				t.Run(name, func(t *testing.T) {
					t.Setenv(DiskFPSetLogLockCntProperty, "1")
					config := NewFPSetConfiguration()
					config.SetMemory(1 << 20)
					var set *DiskFPSet
					if backend == "lsb" {
						set = NewLSBDiskFPSet(config).DiskFPSet
					} else {
						set = NewMSBDiskFPSet(config).DiskFPSet
					}
					set.Init(1, t.TempDir(), "Spec")
					held := make([]bool, set.rwLock.Size())
					t.Cleanup(func() {
						// Only fixture cleanup releases source-retained failure locks.
						for i, retained := range held {
							if retained {
								set.rwLock.GetAt(i).Unlock()
							}
						}
						set.Close()
					})
					switch phase {
					case "flush_io":
						set.Put(41)
						if err := os.Mkdir(set.tmpFilename, 0700); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(filepath.Join(set.tmpFilename, "retained"), nil, 0600); err != nil {
							t.Fatal(err)
						}
					case "open_io":
						if err := os.Remove(set.fpFilename); err != nil {
							t.Fatal(err)
						}
					default:
						var data [16]byte
						binary.BigEndian.PutUint64(data[:8], 41)
						binary.BigEndian.PutUint64(data[8:], 43)
						contents := data[:]
						if phase == "scan_io" {
							contents = contents[:9]
						} else if phase == "invalid" {
							binary.BigEndian.PutUint64(data[8:], 41)
						}
						if err := os.WriteFile(set.fpFilename, contents, 0600); err != nil {
							t.Fatal(err)
						}
					}
					var endpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(set)
					if remote {
						_, endpoint = startFingerprintRPC(t, endpoint)
					}
					valid, err := invokeFingerprintEndpoint(func() (bool, error) { return endpoint.CheckInvariant() })
					for i := range held {
						lock := set.rwLock.GetAt(i)
						if lock.TryLock() {
							lock.Unlock()
						} else {
							held[i] = true
						}
					}
					wantHeld := phase == "flush_io" || phase == "open_io"
					for i, retained := range held {
						if retained != wantHeld {
							t.Errorf("stripe %d retained = %v, want %v", i, retained, wantHeld)
						}
					}
					wantIO := wantHeld || phase == "scan_io"
					if isJavaIOException(err) != wantIO || (!wantIO && err != nil) || valid != (phase == "valid") {
						t.Fatalf("invariant result = %v/%v, want valid=%v, I/O=%v", valid, err, phase == "valid", wantIO)
					}
				})
			}
		}
	}
}
