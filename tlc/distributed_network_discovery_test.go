package tlc

import (
	"bytes"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestNativeDiscoveryBindingAndRetry(t *testing.T) {
	fixture := coordinatorFixture()
	host, client := startCoordinatorRPC(t, fixture)
	discovery := NewDistributedNetworkDiscovery()
	defer discovery.Close()
	location := "tcp://" + client.Address + "/late"
	var output bytes.Buffer
	var delays []time.Duration
	endpoint, err := lookupDistributedServerURL("coordinator", location, discovery.Lookup, func(delay time.Duration) error {
		delays = append(delays, delay)
		if len(delays) == 3 {
			return host.RegisterCoordinator("late", fixture, client.Address)
		}
		return nil
	}, &output)
	if err != nil || endpoint == nil || len(delays) != 3 || delays[0] != time.Second || delays[1] != time.Second || delays[2] != 2*time.Second {
		t.Fatalf("discovery retry = %v/%v/%v", endpoint, err, delays)
	}
	if strings.Count(output.String(), "reachable but not ready") != 3 {
		t.Fatalf("retry progress = %q", output.String())
	}
	// Naming lookup must not call a setting: the polynomial can fail after
	// discovery has succeeded and its original error still belongs to bootstrap.
	fixture.failure = NewRemoteException(javaString("polynomial failed"), nil)
	if _, err := discovery.Lookup(location); err != nil {
		t.Fatalf("setting failure polluted discovery: %v", err)
	}
	host.UnregisterCoordinator("late")
	if _, err := discovery.Lookup(location); err == nil || !err.(*DistributedOperationError).Reachable {
		t.Fatalf("unbound coordinator = %v", err)
	}
	for _, invalid := range []string{"rmi://host:1/name", "tcp://host:1", "tcp://host:1/name/extra", "tcp://user@host:1/name"} {
		if _, err := discovery.Lookup(invalid); err == nil {
			t.Fatalf("invalid location accepted: %s", invalid)
		}
	}
	if err := discovery.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := discovery.Lookup(location); !isDistributedRemoteFailure(err) {
		t.Fatalf("closed discovery = %v", err)
	}
}

func TestNativeDiscoveryUnreachableAndInterruption(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	discovery := NewDistributedNetworkDiscovery()
	defer discovery.Close()
	_, err = discovery.Lookup("//" + address + "/main")
	operation, ok := err.(*DistributedOperationError)
	if !ok || !operation.DiscoveryRetry || operation.Reachable || !errors.Is(operation.Cause, syscall.ECONNREFUSED) {
		t.Fatalf("connection refusal = %T/%v", err, err)
	}
	var output bytes.Buffer
	interrupted := NewInterruptedException("stop")
	_, err = lookupDistributedServerURL("coordinator", "//"+address+"/main", discovery.Lookup, func(time.Duration) error { return interrupted }, &output)
	if err != interrupted || !strings.Contains(output.String(), "unreachable") {
		t.Fatalf("interruption/progress = %v/%q", err, output.String())
	}
}

func TestNativeWorkerBootstrapAndCallback(t *testing.T) {
	previousIntern, previousPoly := internTable, FP64IrredPoly()
	t.Cleanup(func() {
		internTable = previousIntern
		FP64InitPoly(previousPoly)
		initBuiltInOPs()
		initCounterExampleUniqueStrings()
	})
	fixture := coordinatorFixture()
	fixture.manager = NewDistributedFPSetManager(NewLocalFingerprintEndpoint(NewMemFPSet()))
	host, coordinator := startCoordinatorRPC(t, fixture)
	if err := host.RegisterCoordinator(TLCServerWorkerName, fixture, coordinator.Address); err != nil {
		t.Fatal(err)
	}
	_, portText, err := net.SplitHostPort(coordinator.Address)
	if err != nil {
		t.Fatal(err)
	}
	previousPort := TLCServerPort()
	// Discovery uses the original coordinator-port property rather than a
	// different bootstrap path. Restore it after the process has shut down.
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	SetTLCServerPort(port)
	t.Cleanup(func() { SetTLCServerPort(previousPort) })
	network, err := NewDistributedWorkerNetwork("127.0.0.1:0", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := network.Close(); err != nil {
			t.Error(err)
		}
	})
	process := NewDistributedWorkerProcess()
	t.Cleanup(func() {
		if err := process.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	var output bytes.Buffer
	loaded := false
	env := network.Environment(DistributedWorkerEnvironment{
		ToolOut: &output, AvailableProcessors: func() int { return 1 }, ReadyDate: func() string { return "now" },
		LoadApp: func(server DistributedServerEndpoint, resolver *DistributedFilenameToStreamResolver) (*TLCApp, error) {
			if FP64IrredPoly() != ^uint64(0) || internTable != process.intern || resolver.server != server {
				return nil, errors.New("bootstrap initialized app before polynomial/intern/resolver")
			}
			loaded = true
			return nil, nil
		},
	})
	if err := process.Run([]string{"127.0.0.1"}, env); err != nil {
		t.Fatal(err)
	}
	if process.Group == nil || !loaded {
		t.Fatalf("bootstrap failed: %q", output.String())
	}
	for _, err := range process.Group.WaitForRegistrations() {
		if err != nil {
			t.Fatal(err)
		}
	}
	fixture.mu.Lock()
	callback := fixture.worker
	fixture.mu.Unlock()
	uri, err := callback.GetURI()
	if err != nil || !strings.HasPrefix(uri, "tcp://"+network.Address+"/worker-0-") {
		t.Fatalf("advertised callback URI = %q/%v", uri, err)
	}
	result, err := callback.GetNextStates([]*TLCStateMut{})
	if err != nil || result == nil || result.StatesComputed != 0 || len(result.NextStates) != 1 || result.NextStates[0].Size() != 0 {
		t.Fatalf("actual worker callback = %v/%v", result, err)
	}
	if err := callback.Exit(); err != nil {
		t.Fatal(err)
	}
	if _, err := callback.IsAlive(); !isDistributedRemoteFailure(err) {
		t.Fatalf("exited callback stayed published: %v", err)
	}
}

func TestNativeDiscoveryConcurrentLookup(t *testing.T) {
	_, client := startCoordinatorRPC(t, coordinatorFixture())
	discovery := NewDistributedNetworkDiscovery()
	defer discovery.Close()
	var group sync.WaitGroup
	results := make(chan DistributedServerEndpoint, 8)
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			endpoint, err := discovery.Lookup("tcp://" + client.Address + "/main")
			if err != nil {
				t.Error(err)
				return
			}
			results <- endpoint
		}()
	}
	group.Wait()
	close(results)
	var first DistributedServerEndpoint
	for endpoint := range results {
		if first == nil {
			first = endpoint
		}
		if first != endpoint {
			t.Fatal("concurrent lookup retained duplicate connections")
		}
	}
}
