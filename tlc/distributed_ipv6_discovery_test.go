package tlc

import (
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"
	"syscall"
	"testing"
)

// Native addressing has no original Java test. Preserve each source binding
// and captured port while permitting Go IPv6 host operands and scoped hosts.
func TestNativeCoordinatorIPv6Locations(t *testing.T) {
	previousPort := TLCServerPort()
	SetTLCServerPort(12345)
	t.Cleanup(func() { SetTLCServerPort(previousPort) })
	endpoint := NewLocalServerEndpoint(&TLCServer{})
	for _, test := range []struct{ operand, host string }{
		{"worker.example", "worker.example"},
		{"127.0.0.1", "127.0.0.1"},
		{"::1", "::1"},
		{"[::1]", "::1"},
		{"fe80::1%eth0", "fe80::1%eth0"},
		{"[fe80::1%eth0]", "fe80::1%eth0"},
	} {
		t.Run(test.operand, func(t *testing.T) {
			for _, worker := range []bool{false, true} {
				binding := TLCServerName
				if worker {
					binding = TLCServerWorkerName
				}
				var captured string
				lookup := func(location string) (DistributedServerEndpoint, error) {
					captured = location
					parsed, err := url.Parse(location)
					if err != nil {
						t.Fatalf("invalid native discovery location %q: %v", location, err)
					}
					if parsed.Host != net.JoinHostPort(test.host, "12345") || parsed.Path != "/"+binding {
						t.Fatalf("location %q changed host, zone, port or binding", location)
					}
					return endpoint, nil
				}
				var found DistributedServerEndpoint
				var err error
				if worker {
					var retained string
					found, retained, err = DiscoverTLCWorkerServer(test.operand, lookup, nil, io.Discard)
					if retained != captured {
						t.Fatal("keepalive did not retain the actual discovery location")
					}
				} else {
					found, err = LookupDistributedFPServer(test.operand, lookup, nil, io.Discard)
				}
				if err != nil || found != endpoint {
					t.Fatalf("native discovery failed: %v/%v", found, err)
				}
			}
		})
	}
}

func TestNativeCoordinatorIPv6DiscoveryTCP(t *testing.T) {
	// This is a real IPv6 listener, rather than an IPv4 listener advertising
	// an IPv6 spelling. The network test requires local IPv6 support.
	network := NewDistributedCoordinatorNetwork("::1", "::1")
	t.Cleanup(func() {
		if err := network.Close(); err != nil {
			t.Error(err)
		}
	})
	registry, err := network.createRegistry(0)
	if errors.Is(err, syscall.EAFNOSUPPORT) || errors.Is(err, syscall.EADDRNOTAVAIL) {
		t.Skipf("IPv6 loopback unavailable: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	server := &TLCServer{}
	for _, binding := range []string{TLCServerName, TLCServerWorkerName} {
		if err := registry.Rebind(binding, server); err != nil {
			t.Fatal(err)
		}
	}
	_, portText, err := net.SplitHostPort(network.Address)
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
	discovery := NewDistributedNetworkDiscovery()
	t.Cleanup(func() {
		if err := discovery.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, host := range []string{"::1", "[::1]"} {
		worker, location, err := DiscoverTLCWorkerServer(host, discovery.Lookup, nil, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		fp, err := LookupDistributedFPServer(host, discovery.Lookup, nil, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		for _, endpoint := range []DistributedServerEndpoint{worker, fp} {
			if done, err := endpoint.IsDone(); err != nil || done {
				t.Fatalf("IPv6 coordinator status = %v/%v", done, err)
			}
		}
		server.SetDone()
		if done, err := NewTLCServerStatusLookup(discovery.Lookup)(location); err != nil || !done {
			t.Fatalf("captured IPv6 keepalive status = %v/%v", done, err)
		}
		server.Done.Store(false)
	}
	workerNetwork, err := NewDistributedWorkerNetwork("[::1]:0", "::1")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := workerNetwork.Close(); err != nil {
			t.Error(err)
		}
	})
	worker := NewDistributedWorker(0, nil, NewDistributedFPSetManager(), workerNetwork.workerAddress)
	if err := workerNetwork.Environment(DistributedWorkerEnvironment{}).PublishWorker(worker); err != nil {
		t.Fatal(err)
	}
	callback, err := DialWorkerEndpoint(workerNetwork.Address, worker.networkReference.Object)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = callback.CloseConnection() })
	if alive, err := callback.IsAlive(); err != nil || !alive {
		t.Fatalf("IPv6 worker callback = %v/%v", alive, err)
	}
	if uri, err := callback.GetURI(); err != nil || uri != "tcp://"+workerNetwork.Address+"/"+worker.networkReference.Object {
		t.Fatalf("IPv6 worker callback URI = %q/%v", uri, err)
	}
}
