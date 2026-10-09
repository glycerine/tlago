package tlc

import (
	"bytes"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
)

// Native TCP accepts interface names containing hyphens. Exercise construction,
// publication and an actual callback via the listener's loopback address; this
// does not claim a link-local route on an interface supplied by the fixture.
func TestNativeWorkerScopedAddressPublication(t *testing.T) {
	for _, host := range []string{"fe80::1%br-tlc", "[fe80::1%br-tlc]", "fe80::1%veth.42"} {
		t.Run(host, func(t *testing.T) {
			network, err := NewDistributedWorkerNetwork("127.0.0.1:0", host)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := network.Close(); err != nil {
					t.Error(err)
				}
			})
			worker := NewDistributedWorker(7, nil, NewDistributedFPSetManager(), network.workerAddress)
			if err := network.Environment(DistributedWorkerEnvironment{}).PublishWorker(worker); err != nil {
				t.Fatal(err)
			}
			uri, err := url.Parse(worker.GetURI())
			if err != nil {
				t.Fatalf("invalid native worker URI %q: %v", worker.GetURI(), err)
			}
			if uri.Scheme != "tcp" || uri.Host != network.Address || strings.TrimPrefix(uri.Path, "/") != worker.networkReference.Object || worker.GetASCIIURI() != worker.GetURI() {
				t.Fatalf("native worker location lost its zone or endpoint: %q", worker.GetURI())
			}
			_, port, err := net.SplitHostPort(network.Address)
			if err != nil {
				t.Fatal(err)
			}
			callback, err := DialWorkerEndpoint(net.JoinHostPort("127.0.0.1", port), worker.networkReference.Object)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := callback.CloseConnection(); err != nil {
					t.Error(err)
				}
			})
			if alive, err := callback.IsAlive(); err != nil || !alive {
				t.Fatalf("callback not published: %v/%v", alive, err)
			}
			if reported, err := callback.GetURI(); err != nil || reported != worker.GetURI() {
				t.Fatalf("callback location changed: %q/%v", reported, err)
			}
		})
	}
}

// Carry the validated native address through main's asynchronous registration
// runnable, not just the direct worker constructor. Route the fixture callback
// over the actual IPv4 listener; no br-tlc interface is required or fabricated.
func TestNativeWorkerScopedAddressBootstrap(t *testing.T) {
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
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	previousPort := TLCServerPort()
	SetTLCServerPort(port)
	t.Cleanup(func() { SetTLCServerPort(previousPort) })
	network, err := NewDistributedWorkerNetwork("127.0.0.1:0", "fe80::1%br-tlc")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := network.Close(); err != nil {
			t.Error(err)
		}
	})
	process := NewDistributedWorkerProcess()
	var resolverDirectory string
	t.Cleanup(func() {
		if resolverDirectory != "" {
			if err := os.RemoveAll(resolverDirectory); err != nil {
				t.Error(err)
			}
		}
	})
	t.Cleanup(func() {
		if err := process.Shutdown(); err != nil {
			t.Error(err)
		}
	})
	var output bytes.Buffer
	env := network.Environment(DistributedWorkerEnvironment{
		ToolOut: &output, AvailableProcessors: func() int { return 1 },
		LoadApp: func(DistributedServerEndpoint, *DistributedFilenameToStreamResolver) (*TLCApp, error) {
			return nil, nil
		},
	})
	env.RegisterWorker = func(server DistributedServerEndpoint, worker *DistributedWorker) error {
		_, port, err := net.SplitHostPort(network.Address)
		if err != nil {
			return err
		}
		return server.(*NetworkServerEndpoint).RegisterWorkerReference(DistributedEndpointReference{
			Address: net.JoinHostPort("127.0.0.1", port), Object: worker.networkReference.Object,
		})
	}
	if err := process.Run([]string{"127.0.0.1"}, env); err != nil {
		t.Fatal(err)
	}
	if process.Resolver != nil {
		resolverDirectory = process.Resolver.tmpDir
	}
	if process.Group == nil {
		t.Fatalf("worker bootstrap failed: %s", output.String())
	}
	for _, err := range process.Group.WaitForRegistrations() {
		if err != nil {
			t.Fatal(err)
		}
	}
	fixture.mu.Lock()
	callback := fixture.worker
	fixture.mu.Unlock()
	if callback == nil {
		t.Fatal("native worker never registered")
	}
	uri, err := callback.GetURI()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(uri)
	if err != nil || parsed.Host != network.Address || !strings.Contains(uri, "%25br-tlc") {
		t.Fatalf("worker main lost the native zone: %q/%v", uri, err)
	}
	if alive, err := callback.IsAlive(); err != nil || !alive {
		t.Fatalf("registered callback = %v/%v", alive, err)
	}
	if err := callback.Exit(); err != nil {
		t.Fatal(err)
	}
	if _, err := callback.IsAlive(); !isDistributedRemoteFailure(err) {
		t.Fatalf("exited callback stayed available: %v", err)
	}
	if !process.Runtime.executor.IsShutdown() {
		t.Fatal("worker exit did not shut down its executor")
	}
}
