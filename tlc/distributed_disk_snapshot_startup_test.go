package tlc

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// No enabled original method covers corrupt named disk snapshots through
// coordinator startup. Expectations follow DiskFPSet.recover(String),
// MultiFPSet.recover(String) and FPSetManager.checkpoint's distinct catches.
func TestDistributedDiskSnapshotRecoveryStartupBoundary(t *testing.T) {
	for _, implementation := range []string{"tlc2.tool.fp.LSBDiskFPSet", "tlc2.tool.fp.MSBDiskFPSet"} {
		for _, nested := range []bool{false, true} {
			for _, remote := range []bool{false, true} {
				for _, input := range []string{"missing", "empty", "duplicate", "descending", "truncated"} {
					name := implementation + "/" + map[bool]string{false: "direct", true: "nested"}[nested] + "/" + map[bool]string{false: "local", true: "tcp"}[remote] + "/" + input
					t.Run(name, func(t *testing.T) {
						captureFailoverToolIO(t, ToolIOTool)
						metadata, failedDirectory, healthyDirectory := t.TempDir(), t.TempDir(), t.TempDir()
						trace := NewTLCTrace(metadata, "Spec")
						defer trace.Close()
						var savedTrace [16]byte
						binary.BigEndian.PutUint64(savedTrace[8:], 73)
						if err := os.WriteFile(trace.chkptName("chkpt"), savedTrace[:], 0600); err != nil {
							t.Fatal(err)
						}
						queue := NewMemStateQueue(metadata)
						if err := queue.BeginChkpt(); err != nil {
							t.Fatal(err)
						}
						if err := queue.CommitChkpt(); err != nil {
							t.Fatal(err)
						}
						queue.len = 5
						config := NewFPSetConfigurationWithRatioAndImplementation(1, implementation)
						config.SetMemory(1 << 20)
						var disk *DiskFPSet
						var store FPSet
						var sibling *DiskFPSet
						checkpointName := "Spec"
						if nested {
							config.SetFPBits(1)
							multi := NewMultiFPSet(config)
							store = multi
							// Use the actual source factory children, not test endpoints.
							if implementation == "tlc2.tool.fp.LSBDiskFPSet" {
								disk = multi.Sets[0].(*LSBDiskFPSet).DiskFPSet
								sibling = multi.Sets[1].(*LSBDiskFPSet).DiskFPSet
							} else {
								disk = multi.Sets[0].(*MSBDiskFPSet).DiskFPSet
								sibling = multi.Sets[1].(*MSBDiskFPSet).DiskFPSet
							}
							checkpointName += "_0"
						} else {
							if implementation == "tlc2.tool.fp.LSBDiskFPSet" {
								disk = NewLSBDiskFPSet(config).DiskFPSet
							} else {
								disk = NewMSBDiskFPSet(config).DiskFPSet
							}
							store = disk
						}
						store.Init(1, failedDirectory, "failed")
						defer store.Close()
						workers := append([]*BufferedRandomAccessFile(nil), disk.braf...)
						pool := append([]*BufferedRandomAccessFile(nil), disk.brafPool...)
						disk.poolIndex = 2
						values := []uint64{41, 43, 43, 97}
						if input == "descending" {
							values[2] = 42
						}
						var data [32]byte
						for i, fp := range values {
							binary.BigEndian.PutUint64(data[i*8:], fp)
						}
						length := len(data)
						if input == "empty" {
							length = 0
						} else if input == "truncated" {
							length = 19
						}
						if input != "missing" {
							if err := os.WriteFile(disk.chkptName(checkpointName, "chkpt"), data[:length], 0600); err != nil {
								t.Fatal(err)
							}
						}
						if nested {
							binary.BigEndian.PutUint64(data[:8], 61)
							if err := os.WriteFile(sibling.chkptName("Spec_1", "chkpt"), data[:8], 0600); err != nil {
								t.Fatal(err)
							}
						}
						healthy := NewMemFPSet()
						healthy.Init(1, healthyDirectory, "healthy")
						binary.BigEndian.PutUint64(data[:8], 101)
						if err := os.WriteFile(filepath.Join(healthyDirectory, "Spec.fp.chkpt"), data[:8], 0600); err != nil {
							t.Fatal(err)
						}
						var failedEndpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(store)
						var healthyEndpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(healthy)
						if remote {
							_, failedEndpoint = startFingerprintRPC(t, failedEndpoint)
							_, healthyEndpoint = startFingerprintRPC(t, healthyEndpoint)
						}
						manager := NewDistributedFPSetManager(failedEndpoint, healthyEndpoint)
						manager.fpSets[0].hostname = "corrupt-disk"
						server := NewTLCServer("Spec", "Spec", metadata, manager, queue, trace)
						server.SetTool(NewTool())
						server.app.fromCheckpoint = &metadata
						publicationCalls := 0
						publicationStop := errors.New("stop at native hostname boundary")
						server.ConfigurePublication(TLCServerPublication{LocalHostName: func() (string, error) {
							publicationCalls++
							return "", publicationStop
						}})
						recorder := &MemoryRecorder{}
						AddMessageRecorder(recorder)
						defer RemoveMessageRecorder(recorder)
						code, err := server.ModelCheck()
						continued := input == "truncated" || input == "missing" && !nested
						if code != ECGeneral || err == nil || server.IsDone() || trace.lastPtr != 73 || queue.Size() != 0 || manager.NumOfAliveServers() != 2 {
							t.Fatalf("recovery startup boundary = %d/%v", code, err)
						}
						if continued {
							if err != publicationStop || publicationCalls != 1 || healthy.Size() != 1 || !healthy.Contains(101) {
								t.Fatal("recoverable input stopped healthy recovery or publication")
							}
							wantRecovered := uint64(1)
							if input == "truncated" {
								wantRecovered = 3
								if nested {
									wantRecovered++
								}
							}
							if records := recorder.Records(ECTLCCheckpointRecoverEnd); len(records) != 1 || strings.Join(records[0].Params, ",") != strconv.FormatUint(wantRecovered, 10)+",0" {
								t.Fatal("recovery-end counts do not reflect actual disk/sibling storage")
							}
						} else {
							if publicationCalls != 0 || healthy.Size() != 0 || recorder.Recorded(ECTLCCheckpointRecoverEnd) {
								t.Fatal("failed recovery reached healthy registration or publication")
							}
							if input == "missing" {
								if isJavaIOException(err) || !strings.Contains(err.Error(), "recover checkpoint Spec_0") {
									t.Fatalf("nested I/O lost its operation failure boundary: %v", err)
								}
							} else if err.Error() != NewTLCRuntimeException(ECSystemIndexError).Error() {
								t.Fatalf("disk ordering/index failure lost source text: %v", err)
							}
						}
						wantCount, wantWrites := int64(0), uint64(0)
						if input == "duplicate" || input == "descending" {
							wantCount, wantWrites = 4, 3 // Header count includes the unread fourth record.
						} else if input == "truncated" {
							wantCount, wantWrites = 2, 2
						}
						if disk.GetFileCnt() != wantCount || disk.GetDiskWriteCnt() != wantWrites {
							t.Fatalf("partial disk reconstruction = %d/%d, want %d/%d", disk.GetFileCnt(), disk.GetDiskWriteCnt(), wantCount, wantWrites)
						}
						if nested && (sibling.Size() != 1 || !sibling.Contains(61)) {
							t.Fatal("nested failure returned before healthy sibling recovery completed")
						}
						wantSize := uint64(wantCount)
						if nested {
							wantSize++
						}
						if store.Size() != wantSize {
							t.Fatalf("recovered size = %d, want %d", store.Size(), wantSize)
						}
						currentReaders := append(append([]*BufferedRandomAccessFile(nil), disk.braf...), disk.brafPool...)
						for i, old := range append(workers, pool...) {
							current := currentReaders[i]
							if input == "truncated" {
								if current == old || !old.closed || current.closed {
									t.Fatal("successful recovery did not reopen each reader")
								}
							} else if current != old || old.closed {
								t.Fatal("failed reconstruction replaced a reader")
							}
						}
						if input == "truncated" {
							if disk.poolIndex != 0 || len(disk.index) != 2 || disk.index[0] != 41 || disk.index[1] != 43 || !disk.Contains(41) || !disk.Contains(43) || disk.Contains(97) {
								t.Fatal("trailing partial record changed completed disk recovery")
							}
						} else {
							if disk.poolIndex != 2 {
								t.Fatal("failed reconstruction reset reader pool cursor")
							}
							if input == "missing" {
								if disk.index != nil {
									t.Fatal("missing snapshot allocated an index")
								}
							} else {
								first := uint64(41)
								if input == "empty" {
									first = 0
								}
								if len(disk.index) != 2 || disk.index[0] != first || disk.index[1] != 0 {
									t.Fatal("failed reconstruction changed partial index mutation")
								}
							}
						}
						const warning = "Error: Failed to checkpoint the fingerprint server at corrupt-disk. This server might be down."
						wantWarnings := 0
						if input == "missing" && !nested {
							wantWarnings = 1
						}
						if strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning) != wantWarnings || !recorder.Recorded(ECTLCCheckpointRecoverStart) || recorder.Recorded(ECTLCComputingInit) || recorder.Recorded(ECTLCDistributedServerRunning) || recorder.Recorded(ECTLCFinished) || recorder.Recorded(ECGeneral) {
							t.Fatal("recovery diagnostics crossed the source startup boundary")
						}
					})
				}
			}
		}
	}
}
