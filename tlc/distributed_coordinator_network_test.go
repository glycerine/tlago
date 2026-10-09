package tlc

import (
	"net"
	"sync"
	"testing"
)

// No enabled upstream test covers the native publication boundary. These tests
// exercise the existing lifecycle callbacks through real Go TCP discovery.
func TestNativeCoordinatorPublicationLifecycle(t *testing.T) {
	network := NewDistributedCoordinatorNetwork("127.0.0.1", "127.0.0.1")
	defer network.Close()
	publication := network.Publication()
	if network.Address != "" {
		t.Fatal("listener opened before CreateRegistry")
	}
	if host, err := publication.LocalHostName(); err != nil || host != "127.0.0.1" {
		t.Fatalf("hostname = %q/%v", host, err)
	}
	registry, err := publication.CreateRegistry(0)
	if err != nil {
		t.Fatal(err)
	}
	if network.Address == "" {
		t.Fatal("listener address was not retained")
	}
	if _, err := publication.CreateRegistry(0); err == nil {
		t.Fatal("duplicate listener accepted")
	}
	if current, err := publication.GetRegistry(0); err != nil || current != registry {
		t.Fatalf("registry identity = %p/%v", current, err)
	}
	server := &TLCServer{FileName: "first", InternTable: NewInternTable(16)}
	if err := registry.Rebind(TLCServerName, server); err != nil {
		t.Fatal(err)
	}
	discovery := NewDistributedNetworkDiscovery()
	defer discovery.Close()
	location := "tcp://" + network.Address + "/"
	if _, err := discovery.Lookup(location + TLCServerWorkerName); err == nil || !err.(*DistributedOperationError).Reachable {
		t.Fatalf("workers published before initialization: %v", err)
	}
	if err := registry.Rebind(TLCServerWorkerName, server); err != nil {
		t.Fatal(err)
	}
	workerView, err := discovery.Lookup(location + TLCServerWorkerName)
	if err != nil {
		t.Fatal(err)
	}
	primaryView, err := discovery.Lookup(location + TLCServerName)
	if err != nil {
		t.Fatal(err)
	}
	if primaryView.(*NetworkServerEndpoint).Object != workerView.(*NetworkServerEndpoint).Object {
		t.Fatal("two names for the same coordinator returned different endpoint identities")
	}
	if name, err := workerView.GetSpecFileName(); err != nil || name != "first" {
		t.Fatalf("first binding = %q/%v", name, err)
	}
	replacement := &TLCServer{FileName: "second", InternTable: NewInternTable(16)}
	if err := registry.Rebind(TLCServerWorkerName, replacement); err != nil {
		t.Fatal(err)
	}
	if name, err := workerView.GetSpecFileName(); err != nil || name != "first" {
		t.Fatalf("rebind retargeted an existing coordinator reference: %q/%v", name, err)
	}
	replacementView, err := discovery.Lookup(location + TLCServerWorkerName)
	if err != nil {
		t.Fatal(err)
	}
	if name, err := replacementView.GetSpecFileName(); err != nil || name != "second" {
		t.Fatalf("fresh lookup did not resolve the replacement: %q/%v", name, err)
	}
	if local, err := registry.Lookup(TLCServerName); err != nil || local != server {
		t.Fatalf("local shutdown lookup = %p/%v", local, err)
	}
	if ok, err := publication.Unexport(server, false); err != nil || !ok {
		t.Fatalf("unpublish = %v/%v", ok, err)
	}
	if _, err := discovery.Lookup(location + TLCServerName); err == nil {
		t.Fatal("coordinator remained published")
	}
	if _, err := workerView.GetSpecFileName(); !isDistributedRemoteFailure(err) {
		t.Fatalf("removed coordinator reference remained callable: %v", err)
	}
	if name, err := replacementView.GetSpecFileName(); err != nil || name != "second" {
		t.Fatalf("removing the original coordinator affected its replacement: %q/%v", name, err)
	}
	if _, err := discovery.Lookup(location + TLCServerWorkerName); err != nil {
		t.Fatalf("unpublishing one server removed another: %v", err)
	}
	if err := registry.Unbind(TLCServerWorkerName); err != nil {
		t.Fatal(err)
	}
	if name, err := replacementView.GetSpecFileName(); err != nil || name != "second" {
		t.Fatalf("removing a discovery name removed the coordinator endpoint: %q/%v", name, err)
	}
	if err := registry.Unbind(TLCServerWorkerName); err == nil {
		t.Fatal("missing unbind succeeded")
	}
	if _, err := publication.Unexport(server, false); err == nil {
		t.Fatal("duplicate unpublish succeeded")
	}
	if err := network.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := publication.GetRegistry(0); !isDistributedRemoteFailure(err) {
		t.Fatalf("closed registry = %v", err)
	}
	if err := registry.Rebind(TLCServerName, replacement); err != net.ErrClosed {
		t.Fatalf("publication after close = %v", err)
	}
}

func TestNativeCoordinatorPublicationFailureCleanup(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	network := NewDistributedCoordinatorNetwork("127.0.0.1", "127.0.0.1")
	defer network.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	if _, err := network.Publication().CreateRegistry(port); err == nil {
		t.Fatal("occupied port accepted")
	}
	if network.registry != nil || network.Address != "" {
		t.Fatal("failed listen changed publication state")
	}
	if _, err := network.Publication().CreateRegistry(0); err != nil {
		t.Fatal(err)
	}
	// An early init error skips normal unbind/unexport in the original command.
	// The process owner must still be able to close the host and its listener.
	server := &TLCServer{InternTable: NewInternTable(16)}
	if err := network.registry.Rebind(TLCServerName, server); err != nil {
		t.Fatal(err)
	}
	if err := network.Close(); err != nil {
		t.Fatal(err)
	}
	if server.unexported.Load() {
		t.Fatal("network close substituted for source unexport")
	}
	if err := network.Close(); err != nil {
		t.Fatal(err)
	}
	unstarted := NewDistributedCoordinatorNetwork("127.0.0.1", "127.0.0.1")
	if err := unstarted.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := unstarted.Publication().CreateRegistry(0); err != net.ErrClosed {
		t.Fatalf("create after close = %v", err)
	}
}

func TestNativeCoordinatorConcurrentPublication(t *testing.T) {
	network := NewDistributedCoordinatorNetwork("127.0.0.1", "127.0.0.1")
	defer network.Close()
	registry, err := network.Publication().CreateRegistry(0)
	if err != nil {
		t.Fatal(err)
	}
	server := &TLCServer{FileName: "first", InternTable: NewInternTable(16)}
	replacement := &TLCServer{FileName: "second", InternTable: NewInternTable(16)}
	discovery := NewDistributedNetworkDiscovery()
	defer discovery.Close()
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			name := fmtInt(i)
			if err := registry.Rebind(name, server); err != nil {
				t.Error(err)
				return
			}
			if value, err := registry.Lookup(name); err != nil || value != server {
				t.Errorf("lookup = %p/%v", value, err)
			}
			location := "tcp://" + network.Address + "/" + name
			captured, err := discovery.Lookup(location)
			if err != nil {
				t.Error(err)
				return
			}
			if err := registry.Rebind(name, replacement); err != nil {
				t.Error(err)
				return
			}
			if spec, err := captured.GetSpecFileName(); err != nil || spec != "first" {
				t.Errorf("existing reference retargeted: %q/%v", spec, err)
			}
			current, err := discovery.Lookup(location)
			if err != nil {
				t.Error(err)
				return
			}
			if spec, err := current.GetSpecFileName(); err != nil || spec != "second" {
				t.Errorf("new reference = %q/%v", spec, err)
			}
			if err := registry.Unbind(name); err != nil {
				t.Error(err)
			}
		}(i)
	}
	group.Wait()
}
