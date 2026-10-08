package tlc

import (
	"bytes"
	"errors"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"testing"
)

// New native transport tests: upstream has no enabled coordinator TCP harness.
type rpcCoordinatorFixture struct {
	*LocalServerEndpoint
	mu       sync.Mutex
	manager  *DistributedFPSetManager
	worker   DistributedWorkerEndpoint
	fp       DistributedFingerprintEndpoint
	hostname string
	failure  error
}

func (f *rpcCoordinatorFixture) GetFPSetManager() (*DistributedFPSetManager, error) {
	return f.manager.snapshotForWorker(), nil
}
func (f *rpcCoordinatorFixture) GetIrredPolyForFP() (uint64, error) { return math.MaxUint64, f.failure }
func (f *rpcCoordinatorFixture) GetFile(name string) ([]byte, error) {
	switch name {
	case "null":
		return nil, nil
	case "empty":
		return []byte{}, nil
	case "error":
		return nil, NewRemoteException(javaString("file unavailable"), nil)
	}
	return []byte{0, 255, 1}, nil
}
func (f *rpcCoordinatorFixture) RegisterWorker(worker DistributedWorkerEndpoint) error {
	if _, err := worker.GetURI(); err != nil {
		return err
	}
	f.mu.Lock()
	f.worker = worker
	f.mu.Unlock()
	return nil
}
func (f *rpcCoordinatorFixture) RegisterFPSet(fp DistributedFingerprintEndpoint, hostname string) error {
	f.mu.Lock()
	f.fp, f.hostname = fp, hostname
	f.mu.Unlock()
	return nil
}

func startCoordinatorRPC(t *testing.T, endpoint DistributedServerEndpoint) (*DistributedRPCServer, *NetworkServerEndpoint) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := NewDistributedRPCServer()
	if err := server.RegisterCoordinator("main", endpoint, listener.Addr().String()); err != nil {
		_ = listener.Close()
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Error(err)
		}
		if err := <-done; err != nil {
			t.Error(err)
		}
	})
	client, err := DialServerEndpoint(listener.Addr().String(), "main")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.CloseConnection() })
	return server, client
}
func coordinatorFixture() *rpcCoordinatorFixture {
	server := &TLCServer{InternTable: NewInternTable(16), FileName: "Spec", ConfigName: "Config"}
	server.SetCheckDeadlock(false)
	return &rpcCoordinatorFixture{LocalServerEndpoint: NewLocalServerEndpoint(server)}
}

func TestCoordinatorRPCSettingsFilesAndInterning(t *testing.T) {
	fixture := coordinatorFixture()
	original := fixture.Server.InternTable.Put("variable")
	original.loc = 3
	_, client := startCoordinatorRPC(t, fixture)
	if value, err := client.GetCheckDeadlock(); err != nil || value {
		t.Fatalf("deadlock = %v/%v", value, err)
	}
	if value, err := client.GetPreprocess(); err != nil || !value {
		t.Fatalf("preprocess = %v/%v", value, err)
	}
	if value, err := client.GetIrredPolyForFP(); err != nil || value != math.MaxUint64 {
		t.Fatalf("polynomial = %x/%v", value, err)
	}
	if value, err := client.IsDone(); err != nil || value {
		t.Fatalf("done = %v/%v", value, err)
	}
	if value, err := client.GetSpecFileName(); err != nil || value != "Spec" {
		t.Fatalf("spec = %q/%v", value, err)
	}
	if value, err := client.GetConfigFileName(); err != nil || value != "Config" {
		t.Fatalf("config = %q/%v", value, err)
	}
	for _, name := range []string{"null", "empty", "binary"} {
		got, err := client.GetFile(name)
		want, _ := fixture.GetFile(name)
		if err != nil || !bytes.Equal(got, want) || (got == nil) != (want == nil) {
			t.Fatalf("file %s = %v/%v", name, got, err)
		}
	}
	if _, err := client.GetFile("error"); !isDistributedRemoteFailure(err) {
		t.Fatalf("file failure = %v", err)
	}
	// The existing resolver catches coordinator connection failures and retains
	// its initial empty byte array, rather than propagating or returning null.
	resolver := &DistributedFilenameToStreamResolver{server: client}
	if data := resolver.fetch("error"); data == nil || len(data) != 0 {
		t.Fatal("resolver failed to catch native coordinator failure")
	}
	for i := 0; i < 2; i++ {
		got, err := client.Intern("variable")
		if err != nil || got == original || got.s != original.s || got.tok != original.tok || got.loc != 3 {
			t.Fatalf("intern metadata = %v/%v", got, err)
		}
		got.loc = 8
	}
	if original.loc != 3 {
		t.Fatal("worker changed coordinator metadata")
	}
}

func TestCoordinatorRPCManagerSnapshotAndRegistration(t *testing.T) {
	fixture := coordinatorFixture()
	manager := NewDynamicDistributedFPSetManager(3)
	for _, name := range []string{"a", "b", "c"} {
		if err := manager.RegisterFPSet(NewLocalFingerprintEndpoint(NewMemFPSet()), name); err != nil {
			t.Fatal(err)
		}
	}
	manager.Reassign(0)
	manager.Trace = &TLCTrace{}
	fixture.manager = manager
	server, client := startCoordinatorRPC(t, fixture)
	worker := &rpcTestWorker{}
	if err := server.RegisterWorker("callback", worker); err != nil {
		t.Fatal(err)
	}
	callback, err := DialWorkerEndpoint(client.Address, "callback")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = callback.CloseConnection() })
	if err := client.RegisterWorker(callback); err != nil {
		t.Fatal(err)
	}
	fixture.mu.Lock()
	registeredWorker := fixture.worker
	fixture.mu.Unlock()
	if uri, err := registeredWorker.GetURI(); err != nil || uri != "tcp://worker:1234/primary" {
		t.Fatalf("worker callback = %q/%v", uri, err)
	}
	storage := NewMemFPSet()
	if err := server.RegisterFingerprint("remote", NewLocalFingerprintEndpoint(storage)); err != nil {
		t.Fatal(err)
	}
	fp, err := DialFingerprintEndpoint(client.Address, "remote")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = fp.CloseConnection() })
	if err := client.RegisterFPSet(fp, "remote-host"); err != nil {
		t.Fatal(err)
	}
	fixture.mu.Lock()
	registeredFP, registeredHost := fixture.fp, fixture.hostname
	fixture.mu.Unlock()
	if seen, err := registeredFP.Put(42); err != nil || seen || registeredHost != "remote-host" || !storage.Contains(42) {
		t.Fatalf("fingerprint callback = %v/%v", seen, err)
	}
	got, err := client.GetFPSetManager()
	if err != nil {
		t.Fatal(err)
	}
	if got == manager || got.Trace != nil || got.Mask != manager.Mask || got.ExpectedNumServers != 3 || got.NumOfAliveServers() != 2 || got.fpSets[0] != got.fpSets[1] {
		t.Fatal("manager state or wrapper identity changed")
	}
	if got.Reassign(1) != 2 || got.NumOfAliveServers() != 1 || manager.NumOfAliveServers() != 2 {
		t.Fatal("worker failover changed coordinator")
	}
	if got.Put(2) || !manager.Contains(2) {
		t.Fatal("remote snapshot copied fingerprint storage")
	}
	other, err := client.GetFPSetManager()
	if err != nil || other.NumOfAliveServers() != 2 {
		t.Fatalf("later worker inherited failover: %v", err)
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := registeredWorker.IsAlive(); !isDistributedRemoteFailure(err) {
		t.Fatalf("outbound worker connection survived host close: %v", err)
	}
}

func TestCoordinatorRPCFailureAndMissingName(t *testing.T) {
	fixture := coordinatorFixture()
	fixture.failure = NewRemoteException(javaString("polynomial unavailable"), nil)
	_, client := startCoordinatorRPC(t, fixture)
	if _, err := client.GetIrredPolyForFP(); !isDistributedRemoteFailure(err) || err.Error() != fixture.failure.Error() {
		t.Fatalf("coordinator failure changed: %v", err)
	}
	missing, err := DialServerEndpoint(client.Address, "missing")
	if err != nil {
		t.Fatal(err)
	}
	defer missing.CloseConnection()
	if _, err := missing.IsDone(); !isDistributedRemoteFailure(err) {
		t.Fatalf("missing coordinator = %v", err)
	}
	if err := client.RegisterWorker(&rpcTestWorker{}); err == nil {
		t.Fatal("unpublished local worker accepted")
	}
}

func TestDistributedManagerPayloadValidation(t *testing.T) {
	called := false
	resolve := func(DistributedEndpointReference) (DistributedFingerprintEndpoint, error) {
		called = true
		return nil, errors.New("unexpected dial")
	}
	for _, payload := range []*DistributedManagerPayload{nil, {Nil: true, Partitions: []int{0}}, {Partitions: []int{1}}, {Nodes: []DistributedManagerNode{{}}}} {
		if _, err := decodeDistributedManager(payload, resolve); err == nil {
			t.Fatal("malformed manager accepted")
		}
	}
	if called {
		t.Fatal("malformed manager attempted a dial")
	}
}

func TestCoordinatorRPCConcurrentSnapshotAndInterning(t *testing.T) {
	fixture := coordinatorFixture()
	fixture.manager = NewDistributedFPSetManager(NewLocalFingerprintEndpoint(NewMemFPSet()))
	_, client := startCoordinatorRPC(t, fixture)
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			value, err := client.Intern("shared")
			if err != nil || value == nil || value.s != "shared" {
				t.Errorf("concurrent intern = %v/%v", value, err)
				return
			}
			manager, err := client.GetFPSetManager()
			if err != nil || manager == nil || manager.NumOfServers() != 1 {
				t.Errorf("concurrent snapshot = %v/%v", manager, err)
				return
			}
			if _, err := manager.entry(0).set.Contains(42); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
}

type registrationContactWorker struct {
	rpcTestWorker
	contacts atomic.Int32
}

func (w *registrationContactWorker) GetURI() (string, error) {
	w.contacts.Add(1)
	return "tcp://worker:1234/registration", nil
}

func TestCoordinatorRPCRegistrationRequiresQueue(t *testing.T) {
	server := &TLCServer{}
	_, coordinator := startCoordinatorRPC(t, NewLocalServerEndpoint(server))
	worker := &registrationContactWorker{}
	_, endpoint := startWorkerRPC(t, worker)
	err := coordinator.RegisterWorker(endpoint)
	if !isDistributedNullFailure(err) {
		t.Fatalf("remote registration failure %T/%v, want missing queue failure", err, err)
	}
	if worker.contacts.Load() != 0 || len(server.GetServerThreads()) != 0 {
		t.Fatal("missing coordinator queue contacted or registered remote worker")
	}
	// Failure must leave the hosting coordinator usable.
	if done, err := coordinator.IsDone(); err != nil || done {
		t.Fatalf("coordinator status after rejected registration %v/%v", done, err)
	}
}
