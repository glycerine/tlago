package tlc

import (
	"net"
	"strings"
	"testing"
)

// Generated references name an object, not whichever process later occupies
// its TCP address. No original Java test directly covers native host reuse.
func TestFingerprintRPCStaleReferenceAfterAddressReuse(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	start := func(address string, storage FPSet) (*DistributedRPCServer, DistributedEndpointReference) {
		t.Helper()
		listener, err := net.Listen("tcp", address)
		if err != nil {
			t.Fatal(err)
		}
		host := NewDistributedRPCServer()
		reference, err := host.fingerprintReference(NewLocalFingerprintEndpoint(storage), listener.Addr().String())
		if err != nil {
			_ = listener.Close()
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- host.Serve(listener) }()
		t.Cleanup(func() {
			_ = host.Close()
			if err := <-done; err != nil && err != net.ErrClosed {
				t.Error(err)
			}
		})
		// Ensure Serve owns its listener before allowing the caller to close
		// the host and immediately bind the same address in another host.
		probe, err := DialFingerprintEndpoint(reference.Address, reference.Object)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := probe.Size(); err != nil {
			t.Fatal(err)
		}
		_ = probe.CloseConnection()
		return host, reference
	}
	oldHost, old := start("127.0.0.1:0", NewMemFPSet())
	if err := oldHost.Close(); err != nil {
		t.Fatal(err)
	}
	replacement := NewMemFPSet()
	_, fresh := start(old.Address, replacement)
	stale := &NetworkFingerprintEndpoint{Address: old.Address, Object: old.Object}
	t.Cleanup(func() { _ = stale.CloseConnection() })
	if seen, err := stale.Put(41); err == nil || !isDistributedRemoteFailure(err) || !isJavaIOException(err) {
		t.Fatalf("stale reference reached a replacement object: seen %v, failure %v", seen, err)
	}
	if replacement.Size() != 0 || old.Object == fresh.Object {
		t.Fatal("host replacement reused object identity or mutated replacement storage")
	}
	current := &NetworkFingerprintEndpoint{Address: fresh.Address, Object: fresh.Object}
	t.Cleanup(func() { _ = current.CloseConnection() })
	if seen, err := current.Put(43); err != nil || seen || !replacement.Contains(43) || replacement.Contains(41) {
		t.Fatalf("fresh reference cannot access its own storage: %v/%v", seen, err)
	}
	// Object disappearance must also reach TLC's existing partition failover,
	// rather than treating the replacement's empty membership as authoritative.
	healthy := NewMemFPSet()
	manager := NewDistributedFPSetManager(stale, NewLocalFingerprintEndpoint(healthy))
	answers := manager.PutBlock([]*LongVec{NewLongVecFrom([]int64{41}), NewLongVecFrom([]int64{45})})
	if len(answers) != 2 || !answers[0].Get(0) || !answers[1].Get(0) || !healthy.Contains(41) || !healthy.Contains(45) || replacement.Size() != 1 {
		t.Fatal("stale object reference bypassed fingerprint partition failover")
	}
	if manager.NumOfAliveServers() != 1 || manager.entry(0) != manager.entry(1) {
		t.Fatal("stale reference did not reassign to the healthy registration")
	}
}

func TestNativeGeneratedEndpointsAfterAddressReuse(t *testing.T) {
	for _, role := range []string{"fingerprint", "worker"} {
		t.Run(role, func(t *testing.T) {
			fixture := coordinatorFixture()
			_, coordinator := startCoordinatorRPC(t, fixture)
			publish := func(address string) (*DistributedWorkerNetwork, DistributedEndpointReference) {
				t.Helper()
				network, err := NewDistributedWorkerNetwork(address, "")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := network.Close(); err != nil {
						t.Error(err)
					}
				})
				if role == "fingerprint" {
					env := network.FPEnvironment(DistributedFPServerEnvironment{})
					if err := env.RegisterFPSet(coordinator, NewLocalFingerprintEndpoint(NewMemFPSet()), "host"); err != nil {
						t.Fatal(err)
					}
					fixture.mu.Lock()
					endpoint := fixture.fp.(*NetworkFingerprintEndpoint)
					reference := DistributedEndpointReference{Address: endpoint.Address, Object: endpoint.Object}
					fixture.mu.Unlock()
					return network, reference
				}
				worker := NewDistributedWorker(0, nil, nil)
				env := network.Environment(DistributedWorkerEnvironment{})
				if err := env.PublishWorker(worker); err != nil {
					t.Fatal(err)
				}
				if err := env.RegisterWorker(coordinator, worker); err != nil {
					t.Fatal(err)
				}
				return network, *worker.networkReference
			}
			original, old := publish("127.0.0.1:0")
			if err := original.Close(); err != nil {
				t.Fatal(err)
			}
			_, fresh := publish(old.Address)
			if old.Object == fresh.Object {
				t.Fatal("generated object identity survived host replacement")
			}
			if role == "fingerprint" {
				stale := &NetworkFingerprintEndpoint{Address: old.Address, Object: old.Object}
				t.Cleanup(func() { _ = stale.CloseConnection() })
				if _, err := stale.Put(41); !isDistributedRemoteFailure(err) {
					t.Fatalf("stale fingerprint reference did not fail: %v", err)
				}
				current, err := DialFingerprintEndpoint(fresh.Address, fresh.Object)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = current.CloseConnection() })
				if size, err := current.Size(); err != nil || size != 0 {
					t.Fatalf("replacement fingerprint storage mutated: %d/%v", size, err)
				}
			} else {
				stale, err := DialWorkerEndpoint(old.Address, old.Object)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = stale.CloseConnection() })
				if _, err := stale.GetURI(); !isDistributedRemoteFailure(err) {
					t.Fatalf("stale worker reference did not fail: %v", err)
				}
				current, err := DialWorkerEndpoint(fresh.Address, fresh.Object)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = current.CloseConnection() })
				if uri, err := current.GetURI(); err != nil || !strings.HasSuffix(uri, "/"+fresh.Object) {
					t.Fatalf("fresh worker reference lost its identity: %q/%v", uri, err)
				}
			}
		})
	}
}
