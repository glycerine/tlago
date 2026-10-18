package tlc

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Source MemFPSet/MemFPSet2 recovery inserts records before checking the next
// record. No enabled original method exercises partial input through a manager.
func TestDistributedMemoryRecoveryRetainsPrefixAndManagerBoundary(t *testing.T) {
	for _, remote := range []bool{false, true} {
		for _, packed := range []bool{false, true} {
			t.Run(fmt.Sprintf("tcp_%t/packed_%t", remote, packed), func(t *testing.T) {
				captureFailoverToolIO(t, ToolIOTool)
				directory, healthyDirectory := t.TempDir(), t.TempDir()
				var store FPSet = NewMemFPSet()
				if packed {
					store = NewMemFPSet2(NewFPSetConfiguration())
				}
				store.Init(1, directory, "failed")
				store.Put(41)
				healthy := NewMemFPSet()
				healthy.Init(1, healthyDirectory, "healthy")
				var failedEndpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(store)
				var healthyEndpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(healthy)
				if remote {
					_, failedEndpoint = startFingerprintRPC(t, failedEndpoint)
					_, healthyEndpoint = startFingerprintRPC(t, healthyEndpoint)
				}
				for phase := 1; phase <= 8; phase++ {
					t.Run(fmt.Sprintf("phase_%d", phase), func(t *testing.T) {
						first := uint64(100 + phase*10)
						second, later := first+2, first+4
						var data [32]byte
						for i, fp := range []uint64{first, second, second, later} {
							binary.BigEndian.PutUint64(data[i*8:], fp)
						}
						length := 16 + phase
						if phase == 8 {
							length = len(data) // duplicate, followed by an unvisited record
						}
						name := fmt.Sprintf("job_%d", phase)
						if err := os.WriteFile(filepath.Join(directory, name+".fp.chkpt"), data[:length], 0600); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(filepath.Join(healthyDirectory, name+".fp.chkpt"), data[24:], 0600); err != nil {
							t.Fatal(err)
						}
						manager := NewDistributedFPSetManager(failedEndpoint, healthyEndpoint)
						manager.fpSets[0].hostname = "partial-input"
						registration := manager.entry(0)
						priorWarnings := len(ToolIOGetAllMessages())
						before, healthyBefore := store.Size(), healthy.Size()
						err := manager.Recover(name)
						ignoredIO := packed && phase != 8
						if ignoredIO {
							if err != nil || healthy.Size() != healthyBefore+1 || !healthy.Contains(later) {
								t.Fatalf("source I/O catch did not continue healthy recovery: %v", err)
							}
							messages := ToolIOGetAllMessages()[priorWarnings:]
							const warning = "Error: Failed to checkpoint the fingerprint server at partial-input. This server might be down."
							if len(messages) != 1 || strings.TrimSpace(messages[0]) != warning {
								t.Fatalf("source recovery warning = %q", messages)
							}
						} else if err == nil || isJavaIOException(err) || healthy.Size() != healthyBefore || len(ToolIOGetAllMessages()) != priorWarnings {
							t.Fatalf("runtime failure was swallowed or reached healthy recovery: %v", err)
						}
						if !ignoredIO {
							expected := NewTLCRuntimeException(ECSystemDiskIOErrorForFile, "checkpoints")
							if phase == 8 {
								expected = NewTLCRuntimeException(ECTLCFPNotInSet)
							}
							if err.Error() != expected.Error() {
								t.Fatalf("recovery failure text = %q, want %q", err.Error(), expected.Error())
							}
						}
						if store.Size() != before+2 || !store.Contains(41) || !store.Contains(first) || !store.Contains(second) || store.Contains(later) {
							t.Fatal("failed recovery rolled back prefix, cleared prior storage, or consumed later input")
						}
						if manager.entry(0) != registration || manager.NumOfAliveServers() != 2 {
							t.Fatal("file corruption changed source endpoint availability or identity")
						}
					})
				}
			})
		}
	}
}
