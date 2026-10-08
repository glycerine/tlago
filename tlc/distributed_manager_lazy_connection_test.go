package tlc

import (
	"strings"
	"sync"
	"testing"
)

// Source FPSets snapshots carry endpoint references, not a successful aliveness
// probe. A dead store must reach manager failover at operation time rather than
// prevent workers from receiving the snapshot.
func TestCoordinatorRPCSnapshotWithUnavailableFingerprintServer(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	failedHost, failed := startFingerprintRPC(t, NewLocalFingerprintEndpoint(NewMemFPSet()))
	healthyStorage := NewMemFPSet()
	_, healthy := startFingerprintRPC(t, NewLocalFingerprintEndpoint(healthyStorage))
	fixture := coordinatorFixture()
	fixture.manager = NewDistributedFPSetManager(failed, healthy)
	fixture.manager.fpSets[0].hostname = "snapshot-primary"
	fixture.manager.fpSets[1].hostname = "snapshot-survivor"
	_, client := startCoordinatorRPC(t, fixture)
	if err := failedHost.Close(); err != nil {
		t.Fatal(err)
	}
	manager, err := client.GetFPSetManager()
	if err != nil || manager == nil {
		t.Fatalf("dead fingerprint server prevented snapshot receipt: %v", err)
	}
	if manager.NumOfServers() != 2 || manager.NumOfAliveServers() != 2 || manager.entry(0) == manager.entry(1) {
		t.Fatal("snapshot performed availability detection or merged registrations")
	}
	answers := manager.PutBlock([]*LongVec{NewLongVecFrom([]int64{2}), NewLongVecFrom([]int64{3})})
	if len(answers) != 2 || !answers[0].Get(0) || !answers[1].Get(0) || healthyStorage.Size() != 2 || !healthyStorage.Contains(2) || !healthyStorage.Contains(3) {
		t.Fatal("operation-time failover did not publish both partitions at surviving store")
	}
	if manager.NumOfAliveServers() != 1 || manager.entry(0) != manager.entry(1) {
		t.Fatal("worker manager did not reassign the dead fingerprint partition")
	}
	if fixture.manager.NumOfAliveServers() != 2 || fixture.manager.entry(0) == fixture.manager.entry(1) {
		t.Fatal("worker failover mutated coordinator registrations")
	}
	messages := ToolIOGetAllMessages()
	if len(messages) != 1 || !strings.Contains(messages[0], "to the fp server at snapshot-primary.\n") {
		t.Fatalf("snapshot failure/failover warnings = %q", messages)
	}
	// Subsequent workers still receive the coordinator's unchanged references.
	second, err := client.GetFPSetManager()
	if err != nil || second == nil || second.NumOfAliveServers() != 2 {
		t.Fatalf("second worker snapshot = %v/%v", second, err)
	}
	if err := client.CloseConnection(); err != nil {
		t.Fatal(err)
	}
	if _, err := second.entry(1).set.Contains(3); !isJavaIOException(err) {
		t.Fatalf("owner closure allowed an unused fingerprint reference to dial: %v", err)
	}
	if _, err := manager.entry(1).set.Contains(3); !isJavaIOException(err) {
		t.Fatalf("owner closure retained an established fingerprint connection: %v", err)
	}
}

func TestCoordinatorRPCConcurrentFirstFingerprintCalls(t *testing.T) {
	storage := NewMemFPSet()
	_, endpoint := startFingerprintRPC(t, NewLocalFingerprintEndpoint(storage))
	fixture := coordinatorFixture()
	fixture.manager = NewDistributedFPSetManager(endpoint)
	_, client := startCoordinatorRPC(t, fixture)
	manager, err := client.GetFPSetManager()
	if err != nil {
		t.Fatal(err)
	}
	fingerprint := manager.entry(0).set
	start := make(chan struct{})
	var jobs sync.WaitGroup
	for i := range 8 {
		jobs.Add(1)
		go func() {
			defer jobs.Done()
			<-start
			if seen, err := fingerprint.Put(uint64(i)); err != nil || seen {
				t.Errorf("first concurrent put %d = %v/%v", i, seen, err)
			}
		}()
	}
	close(start)
	jobs.Wait()
	if storage.Size() != 8 {
		t.Fatal("concurrent first connections lost fingerprint operations")
	}
	if err := client.CloseConnection(); err != nil {
		t.Fatal(err)
	}
	if _, err := fingerprint.Contains(1); !isJavaIOException(err) {
		t.Fatalf("closed owner still permits fingerprint calls: %v", err)
	}
}
