package tlc

import (
	"bytes"
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

func fpNetworkTestEnvironment(t *testing.T, network *DistributedFPServerNetwork, coordinator *NetworkServerEndpoint, output *bytes.Buffer) DistributedFPServerEnvironment {
	t.Helper()
	t.Setenv(FPSetImplProperty, "tlc2.tool.fp.MemFPSet")
	t.Setenv("java.io.tmpdir", t.TempDir())
	previousRunning := distributedFPServerRunning.Load()
	distributedFPServerRunning.Store(true)
	t.Cleanup(func() { distributedFPServerRunning.Store(previousRunning) })
	env := network.FPEnvironment(DistributedFPServerEnvironment{ToolOut: output, SystemOut: output, SystemErr: output, CurrentTimeMillis: func() int64 { return 12345 }})
	// Keep the production native discovery call but direct it to this test's
	// ephemeral coordinator port, without changing the source port global.
	env.Lookup = func(string) (DistributedServerEndpoint, error) {
		return network.Discovery.Lookup("tcp://" + coordinator.Address + "/main")
	}
	return env
}

func TestNativeFPServerRegistrationReportingAndShutdown(t *testing.T) {
	manager := NewDynamicDistributedFPSetManager(1)
	registration := &distributedFPRegistration{expected: 1, remaining: 1, done: make(chan struct{})}
	coordinator := NewLocalServerEndpoint(&TLCServer{FPSetManager: manager, InternTable: NewInternTable(16), fpRegistration: registration})
	_, client := startCoordinatorRPC(t, coordinator)
	network, err := NewDistributedFPServerNetwork("127.0.0.1:0", "")
	if err != nil {
		t.Fatal(err)
	}
	defer network.Close()
	var output bytes.Buffer
	env := fpNetworkTestEnvironment(t, network, client, &output)
	waits := 0
	env.Wait = func(set FPSet, duration time.Duration) error {
		waits++
		if duration != 300000*time.Millisecond {
			return errors.New("source reporting wait changed")
		}
		_, monitor := fpSetLifecycleAndMonitor(set)
		depth := monitor.releaseForWait()
		defer monitor.reacquireAfterWait(depth)
		remote := manager.entry(0).set
		if found, err := remote.Put(math.MaxUint64); err != nil || found {
			return errors.New("remote fingerprint insertion failed")
		}
		if found, err := remote.Put(math.MaxUint64); err != nil || !found {
			return errors.New("repeat insertion answer changed")
		}
		if count, err := remote.Size(); err != nil || count != 1 {
			return errors.New("registered storage not shared")
		}
		return remote.Exit(false)
	}
	flush, err := runDistributedFPServer("coordinator", env)
	if err != nil || !flush || waits != 1 {
		t.Fatalf("FP command = %v/%v, waits %d, output %q", flush, err, waits, output.String())
	}
	if !strings.Contains(output.String(), "is ready.") || !strings.Contains(output.String(), "Progress: The number of fingerprints") || !strings.Contains(output.String(), "Exiting TLC Distributed FP Server") {
		t.Fatalf("missing command progress: %q", output.String())
	}
	if _, err := manager.entry(0).set.Size(); !isJavaIOException(err) {
		t.Fatalf("exited FP endpoint stayed published: %v", err)
	}
}

func TestNativeFPServerRejectionPreservesWorkerHost(t *testing.T) {
	manager := NewDynamicDistributedFPSetManager(1)
	if err := manager.RegisterFPSet(NewLocalFingerprintEndpoint(NewMemFPSet()), "first"); err != nil {
		t.Fatal(err)
	}
	registration := &distributedFPRegistration{expected: 1, done: make(chan struct{})}
	close(registration.done) // The first store has already satisfied startup registration.
	_, client := startCoordinatorRPC(t, NewLocalServerEndpoint(&TLCServer{FPSetManager: manager, InternTable: NewInternTable(16), fpRegistration: registration}))
	network, err := NewDistributedFPServerNetwork("127.0.0.1:0", "")
	if err != nil {
		t.Fatal(err)
	}
	defer network.Close()
	if err := network.Host.RegisterWorker("kept", &rpcTestWorker{}); err != nil {
		t.Fatal(err)
	}
	worker, err := DialWorkerEndpoint(network.Address, "kept")
	if err != nil {
		t.Fatal(err)
	}
	defer worker.CloseConnection()
	var output bytes.Buffer
	env := fpNetworkTestEnvironment(t, network, client, &output)
	env.Wait = func(FPSet, time.Duration) error { return errors.New("rejected server entered reporting") }
	flush, err := runDistributedFPServer("coordinator", env)
	if err != nil || flush || !strings.Contains(output.String(), "Limit for FPset servers reached (1)") {
		t.Fatalf("rejection = %v/%v, output %q", flush, err, output.String())
	}
	network.Host.mu.Lock()
	published := len(network.Host.fingerprints)
	network.Host.mu.Unlock()
	if published != 0 || manager.NumOfServers() != 1 || !distributedFPServerRunning.Load() {
		t.Fatal("rejection retained its publication or changed server/storage lifecycle")
	}
	if alive, err := worker.IsAlive(); err != nil || !alive {
		t.Fatalf("rejection stopped shared worker host: %v/%v", alive, err)
	}
}

func TestNativeFPRegistrationFailureCategories(t *testing.T) {
	for _, failure := range []error{NewFPSetManagerException("full"), NewRemoteException(javaString("connection"), nil), NewRuntimeException("runtime")} {
		payload, err := EncodeDistributedFailure(failure)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeDistributedFailure(payload)
		if err != nil {
			t.Fatal(err)
		}
		want := isDistributedFPRegistrationRejected(failure)
		if isDistributedFPRegistrationRejected(decoded) != want || decoded.Error() != failure.Error() {
			t.Fatalf("registration failure changed: %v", decoded)
		}
	}
}
