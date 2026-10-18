package tlc

import (
	"bytes"
	"testing"
	"time"
)

// No original method directly tests these publication/discovery boundaries.
func TestNativeCoordinatorFailureCategories(t *testing.T) {
	for _, transport := range []string{"local", "tcp"} {
		t.Run(transport, func(t *testing.T) {
			publication := NewTLCRegistryNamespace().Publication()
			port := TLCServerPort()
			if transport == "tcp" {
				network := NewDistributedCoordinatorNetwork("127.0.0.1", "127.0.0.1")
				t.Cleanup(func() { _ = network.Close() })
				publication, port = network.Publication(), 0
			}
			registry, err := publication.CreateRegistry(port)
			if err != nil {
				t.Fatal(err)
			}
			checkMissing := func(err error) {
				t.Helper()
				failure, ok := err.(*DistributedOperationError)
				if !ok || !isDistributedCoordinatorBindingMissing(failure) || failure.Class != "tlc.CoordinatorBindingMissing" || !failure.DiscoveryRetry || !failure.Reachable || failure.Remote || failure.IO || failure.Recoverable {
					t.Fatalf("missing binding must use Go discovery category: %T/%v", err, err)
				}
				payload, err := EncodeDistributedFailure(failure)
				if err != nil {
					t.Fatal(err)
				}
				decoded, err := DecodeDistributedFailure(payload)
				if err != nil || !isDistributedCoordinatorBindingMissing(decoded) {
					t.Fatalf("native binding category lost: %v/%v", decoded, err)
				}
				var output bytes.Buffer
				delays := []time.Duration{}
				server := NewLocalServerEndpoint(&TLCServer{})
				attempt := 0
				found, err := LookupTLCWorkerServer("coordinator", func(string) (DistributedServerEndpoint, error) {
					attempt++
					if attempt < 4 {
						return nil, decoded
					}
					return server, nil
				}, func(delay time.Duration) error { delays = append(delays, delay); return nil }, &output)
				if err != nil || found != server || len(delays) != 3 || delays[0] != time.Second || delays[1] != time.Second || delays[2] != 2*time.Second {
					t.Fatalf("native missing binding changed retry order: %v/%v", delays, err)
				}
			}
			_, err = registry.Lookup(TLCServerName)
			checkMissing(err)
			checkMissing(registry.Unbind(TLCServerName))
			if _, err := publication.CreateRegistry(port); err == nil {
				t.Fatal("duplicate registry creation succeeded")
			} else if failure, ok := err.(*DistributedOperationError); !ok || !failure.Remote || !failure.IO {
				t.Fatalf("duplicate publication must use Go operation category: %T/%v", err, err)
			}
			server := &TLCServer{}
			server.ConfigurePublication(publication)
			worker := &orderedExitWorker{rpcTestWorker: &rpcTestWorker{}}
			registerCompletedExitThread(server, worker, "tcp://worker:1/primary")
			if err := server.RunWorkerShutdownHook(); err != nil || worker.exitCalls != 0 || server.GetWorkerCount() != 1 {
				t.Fatalf("missing coordinator must suppress hook traversal: %v", err)
			}
			if err := registry.Rebind(TLCServerName, server); err != nil {
				t.Fatal(err)
			}
			if err := server.RunWorkerShutdownHook(); err != nil || worker.exitCalls != 1 {
				t.Fatalf("published coordinator suppressed hook traversal: %v", err)
			}
			remove := publication.Unexport
			if remove == nil {
				remove = server.publicationBoundaries().Unexport
			}
			if removed, err := remove(server, false); err != nil || !removed {
				t.Fatalf("coordinator removal failed: %v/%v", removed, err)
			}
			_, err = remove(server, false)
			failure, ok := err.(*DistributedOperationError)
			if !ok || failure.Class != "tlc.CoordinatorEndpointRemoved" || !failure.Remote || !failure.IO || failure.WorkerUnavailable || failure.ExitIgnorable {
				t.Fatalf("coordinator removal must use its own Go category: %T/%v", err, err)
			}
		})
	}
}

func TestNativeCoordinatorCatalogReferenceIsLazy(t *testing.T) {
	namespace := NewTLCRegistryNamespace()
	port := TLCServerPort()
	reference, err := namespace.GetRegistry(port)
	if err != nil || reference == nil {
		t.Fatalf("reference creation must precede catalog availability: %v", err)
	}
	_, err = reference.Lookup(TLCServerName)
	failure, ok := err.(*DistributedOperationError)
	if !ok || !failure.Remote || !failure.IO || !failure.DiscoveryRetry || failure.Reachable || failure.BindingMissing {
		t.Fatalf("absent catalog must differ from missing binding: %T/%v", err, err)
	}
	registry, err := namespace.CreateRegistry(port)
	if err != nil {
		t.Fatal(err)
	}
	_, err = reference.Lookup(TLCServerName)
	if !isDistributedCoordinatorBindingMissing(err) {
		t.Fatalf("lazy reference did not observe catalog creation: %v", err)
	}
	server := &TLCServer{}
	if err := registry.Rebind(TLCServerName, server); err != nil {
		t.Fatal(err)
	}
	if found, err := reference.Lookup(TLCServerName); err != nil || found != server {
		t.Fatalf("lazy reference did not observe publication: %p/%v", found, err)
	}
}
