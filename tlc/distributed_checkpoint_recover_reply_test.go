package tlc

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type lostCheckpointRecoveryReply struct {
	*LocalFingerprintEndpoint
	host      *DistributedRPCServer
	calls     atomic.Int32
	completed chan struct{}
}

func (e *lostCheckpointRecoveryReply) RecoverFile(name string) error {
	e.calls.Add(1)
	if err := e.LocalFingerprintEndpoint.RecoverFile(name); err != nil {
		return err
	}
	defer close(e.completed)
	return e.host.Close()
}

// Source FPSetManager.Checkpoint.run catches recovery I/O failures without
// replaying or reassigning the registration. No original enabled method tests
// a recovery that completes its storage work but loses the reply over TCP.
func TestFingerprintRPCCompletedRecoveryReplyLoss(t *testing.T) {
	for _, backend := range []string{"mem", "lsb", "msb"} {
		t.Run(backend, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			newStore := func(directory string) FPSet {
				var store FPSet
				if backend == "mem" {
					store = NewMemFPSet()
				} else {
					implementation := "tlc2.tool.fp.LSBDiskFPSet"
					if backend == "msb" {
						implementation = "tlc2.tool.fp.MSBDiskFPSet"
					}
					config := NewFPSetConfigurationWithRatioAndImplementation(1, implementation)
					config.SetFPBits(1)
					config.SetMemory(1 << 20)
					store = NewFPSet(config)
				}
				store.Init(1, directory, "Spec")
				t.Cleanup(store.Close)
				return store
			}
			directories := []string{t.TempDir(), t.TempDir()}
			fingerprints := [][]uint64{{41, uint64(1)<<63 | 43}, {72, uint64(1)<<63 | 74}}
			stores := make([]FPSet, 2)
			for i, directory := range directories {
				producer := newStore(directory)
				for _, fp := range fingerprints[i] {
					producer.Put(fp)
				}
				if err := producer.BeginChkptFile("job"); err != nil {
					t.Fatal(err)
				}
				if err := producer.CommitChkptFile("job"); err != nil {
					t.Fatal(err)
				}
				producer.Close()
				stores[i] = newStore(directory)
				if stores[i].Size() != 0 {
					t.Fatal("fresh recovery storage is not empty")
				}
			}
			endpoint := &lostCheckpointRecoveryReply{LocalFingerprintEndpoint: NewLocalFingerprintEndpoint(stores[0]), completed: make(chan struct{})}
			host, failed := startFingerprintRPC(t, endpoint)
			endpoint.host = host
			_, healthy := startFingerprintRPC(t, NewLocalFingerprintEndpoint(stores[1]))
			manager := NewDistributedFPSetManager(failed, healthy)
			manager.fpSets[0].hostname = "lost-recovery"
			first, second := manager.entry(0), manager.entry(1)
			if err := manager.Recover("job"); err != nil {
				t.Fatal(err)
			}
			select {
			case <-endpoint.completed:
			case <-time.After(5 * time.Second):
				t.Fatal("completed recovery handler did not finish transport closure")
			}
			if endpoint.calls.Load() != 1 || manager.NumOfAliveServers() != 2 || manager.entry(0) != first || manager.entry(1) != second {
				t.Fatal("recovery reply loss was replayed or changed registrations")
			}
			const warning = "Error: Failed to checkpoint the fingerprint server at lost-recovery. This server might be down."
			if strings.Count(strings.Join(ToolIOGetAllMessages(), "\n"), warning) != 1 {
				t.Fatal("recovery reply loss did not report the exact source warning once")
			}
			for i, store := range stores {
				if store.Size() != 2 || !store.Contains(fingerprints[i][0]) || !store.Contains(fingerprints[i][1]) || store.Contains(fingerprints[1-i][0]) {
					t.Fatal("completed or subsequent healthy recovery lost partition membership")
				}
			}
			if err := failed.RecoverFile("job"); !isDistributedRemoteFailure(err) || endpoint.calls.Load() != 1 {
				t.Fatalf("failed connection replayed the completed recovery: %v", err)
			}
			// Publish the retained borrowed storage through a fresh native host.
			// Reading recovered membership must not repeat the recovery operation.
			_, observer := startFingerprintRPC(t, NewLocalFingerprintEndpoint(stores[0]))
			for _, fp := range fingerprints[0] {
				if present, err := observer.Contains(fp); err != nil || !present {
					t.Fatalf("recovered storage unavailable through fresh host: %v/%v", present, err)
				}
			}
		})
	}
}
