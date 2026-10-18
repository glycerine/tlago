package tlc

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// No enabled original model method covers corrupt fingerprint input after
// trace/queue recovery but before coordinator publication.
func TestDistributedFingerprintRecoveryStartupBoundary(t *testing.T) {
	for _, packed := range []bool{false, true} {
		for _, remote := range []bool{false, true} {
			for _, duplicate := range []bool{false, true} {
				name := map[bool]string{false: "memory", true: "packed"}[packed] + "/" + map[bool]string{false: "local", true: "tcp"}[remote] + "/" + map[bool]string{false: "truncated", true: "duplicate"}[duplicate]
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
					var store FPSet = NewMemFPSet()
					if packed {
						store = NewMemFPSet2(NewFPSetConfiguration())
					}
					store.Init(1, failedDirectory, "failed")
					healthy := NewMemFPSet()
					healthy.Init(1, healthyDirectory, "healthy")
					var data [32]byte
					for i, fp := range []uint64{41, 43, 43, 97} {
						binary.BigEndian.PutUint64(data[i*8:], fp)
					}
					length := 19
					if duplicate {
						length = len(data)
					}
					for directory, contents := range map[string][]byte{failedDirectory: data[:length], healthyDirectory: data[24:]} {
						if err := os.WriteFile(filepath.Join(directory, "Spec.fp.chkpt"), contents, 0600); err != nil {
							t.Fatal(err)
						}
					}
					var failedEndpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(store)
					var healthyEndpoint DistributedFingerprintEndpoint = NewLocalFingerprintEndpoint(healthy)
					if remote {
						_, failedEndpoint = startFingerprintRPC(t, failedEndpoint)
						_, healthyEndpoint = startFingerprintRPC(t, healthyEndpoint)
					}
					manager := NewDistributedFPSetManager(failedEndpoint, healthyEndpoint)
					manager.fpSets[0].hostname = "corrupt-fingerprint"
					server := NewTLCServer("Spec", "Spec", metadata, manager, queue, trace)
					server.SetTool(NewTool())
					server.app.fromCheckpoint = &metadata
					publicationCalls := 0
					publicationStop := errors.New("stop after recovery at native hostname boundary")
					server.ConfigurePublication(TLCServerPublication{LocalHostName: func() (string, error) {
						publicationCalls++
						return "", publicationStop
					}})
					recorder := &MemoryRecorder{}
					AddMessageRecorder(recorder)
					defer RemoveMessageRecorder(recorder)
					code, err := server.ModelCheck()
					ignoredIO := packed && !duplicate
					if code != ECGeneral || err == nil || server.IsDone() || trace.lastPtr != 73 || queue.Size() != 0 {
						t.Fatalf("recovery startup lost earlier mutations or classification: %d/%v", code, err)
					}
					if store.Size() != 2 || !store.Contains(41) || !store.Contains(43) || store.Contains(97) || manager.NumOfAliveServers() != 2 {
						t.Fatal("fingerprint failure changed source prefix or availability")
					}
					if ignoredIO {
						if err != publicationStop || publicationCalls != 1 || healthy.Size() != 1 || !healthy.Contains(97) {
							t.Fatal("ignored I/O did not recover the next registration before publication")
						}
						if records := recorder.Records(ECTLCCheckpointRecoverEnd); len(records) != 1 || strings.Join(records[0].Params, ",") != "3,0" {
							t.Fatal("recovery-end statistics did not include partial and healthy storage")
						}
						const warning = "Error: Failed to checkpoint the fingerprint server at corrupt-fingerprint. This server might be down."
						if strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning) != 1 {
							t.Fatal("ignored I/O warning changed")
						}
					} else {
						expected := NewTLCRuntimeException(ECSystemDiskIOErrorForFile, "checkpoints")
						if duplicate {
							expected = NewTLCRuntimeException(ECTLCFPNotInSet)
						}
						if err.Error() != expected.Error() || publicationCalls != 0 || healthy.Size() != 0 || recorder.Recorded(ECTLCCheckpointRecoverEnd) {
							t.Fatal("runtime recovery failure reached a later phase or lost source text")
						}
					}
					if !recorder.Recorded(ECTLCCheckpointRecoverStart) || recorder.Recorded(ECTLCComputingInit) || recorder.Recorded(ECTLCDistributedServerRunning) || recorder.Recorded(ECTLCFinished) || recorder.Recorded(ECGeneral) {
						t.Fatal("recovery failure entered model initialization or completion")
					}
				})
			}
		}
	}
}
