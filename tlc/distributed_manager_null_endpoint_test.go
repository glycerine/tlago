package tlc

import (
	"bytes"
	"encoding/gob"
	"strings"
	"testing"
)

// No upstream test directly covers nullable endpoint fields inside transferred
// manager registrations. Preserve wrapper identity separately from endpoint nulls.
func TestDistributedManagerNullEndpointGraph(t *testing.T) {
	manager := NewDynamicDistributedFPSetManager(5)
	if err := manager.RegisterFPSet(nil, "empty-store"); err != nil {
		t.Fatal(err)
	}
	first := manager.entry(0)
	another := &distributedFPSets{hostname: "another-empty", available: false}
	endpoint := NewLocalFingerprintEndpoint(NewMemFPSet())
	live := &distributedFPSets{set: endpoint, hostname: "live", available: true}
	manager.fpSets = []*distributedFPSets{first, nil, first, another, live}
	manager.managerIsBroken = true
	referenceCalls := 0
	payload, err := encodeDistributedManager(manager, func(got DistributedFingerprintEndpoint) (DistributedEndpointReference, error) {
		referenceCalls++
		if got != endpoint {
			t.Fatalf("reference callback received %v, want live endpoint", got)
		}
		return DistributedEndpointReference{Address: "127.0.0.1:1234", Object: "store"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if referenceCalls != 1 || len(payload.Nodes) != 3 || !payload.Nodes[0].EndpointNil || !payload.Nodes[1].EndpointNil || payload.Nodes[2].EndpointNil {
		t.Fatalf("references %d, nodes %+v", referenceCalls, payload.Nodes)
	}
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(payload); err != nil {
		t.Fatal(err)
	}
	var received DistributedManagerPayload
	if err := gob.NewDecoder(&buffer).Decode(&received); err != nil {
		t.Fatal(err)
	}
	resolved := 0
	got, err := decodeDistributedManager(&received, func(ref DistributedEndpointReference) (DistributedFingerprintEndpoint, error) {
		resolved++
		if ref != payload.Nodes[2].Endpoint {
			t.Fatalf("unexpected endpoint reference %+v", ref)
		}
		return endpoint, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if resolved != 1 || got == manager || got.fpSets[0] == first || got.fpSets[0] != got.fpSets[2] || got.fpSets[0] == got.fpSets[3] || got.fpSets[1] != nil {
		t.Fatal("manager wrapper/null graph changed")
	}
	if got.entry(0).set != nil || got.entry(3).set != nil || got.entry(4).set != endpoint || got.entry(0).hostname != "empty-store" || !got.entry(0).available || got.entry(3).available || !got.managerIsBroken || got.Mask != manager.Mask || got.ExpectedNumServers != 5 {
		t.Fatal("null endpoint changed registration fields")
	}
	got.entry(0).available = false
	if !first.available {
		t.Fatal("receiver mutation changed coordinator wrapper")
	}
}

func TestDistributedManagerNullEndpointValidation(t *testing.T) {
	for _, ref := range []DistributedEndpointReference{{Address: "127.0.0.1:1234"}, {Object: "store"}, {Address: "127.0.0.1:1234", Object: "store"}} {
		payload := &DistributedManagerPayload{Partitions: []int{1}, Nodes: []DistributedManagerNode{{EndpointNil: true, Endpoint: ref}}}
		called := false
		if _, err := decodeDistributedManager(payload, func(DistributedEndpointReference) (DistributedFingerprintEndpoint, error) {
			called = true
			return nil, nil
		}); err == nil || called {
			t.Fatalf("contradictory null endpoint accepted or resolved: %+v, called=%v", ref, called)
		}
	}
}

func TestCoordinatorRPCManagerNullEndpointDefersFailover(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	manager := NewDynamicDistributedFPSetManager(4)
	if err := manager.RegisterFPSet(nil, "null-store"); err != nil {
		t.Fatal(err)
	}
	storage := NewMemFPSet()
	if err := manager.RegisterFPSet(NewLocalFingerprintEndpoint(storage), "live-store"); err != nil {
		t.Fatal(err)
	}
	first := manager.entry(0)
	manager.fpSets = append(manager.fpSets, first, &distributedFPSets{hostname: "another-null", available: true})
	fixture := coordinatorFixture()
	fixture.manager = manager
	_, coordinator := startCoordinatorRPC(t, fixture)
	got, err := coordinator.GetFPSetManager()
	if err != nil {
		t.Fatalf("null endpoint prevented manager receipt: %v", err)
	}
	if got.entry(0).set != nil || got.entry(0) != got.entry(2) || got.entry(0) == got.entry(3) || !got.entry(0).available || len(ToolIOGetAllMessages()) != 0 {
		t.Fatal("manager receipt changed null registration identity/availability or invoked failover")
	}
	if got.Put(4) || !storage.Contains(4) {
		t.Fatal("operation did not fail over from null endpoint to real remote storage")
	}
	if got.entry(0) != got.entry(1) || got.entry(2).available || got.entry(0).set == nil || manager.entry(0) != first || !first.available {
		t.Fatal("worker failover lost wrapper sharing or mutated coordinator")
	}
	messages := ToolIOGetAllMessages()
	if len(messages) != 1 || !strings.Contains(messages[0], "to the fp server at null-store.\nnull") {
		t.Fatalf("null endpoint failover diagnostic %q", messages)
	}
	// Each manager receipt is independent; prior worker failover is local.
	other, err := coordinator.GetFPSetManager()
	if err != nil || other.entry(0).set != nil || !other.entry(0).available || other.entry(0) != other.entry(2) {
		t.Fatalf("later manager receipt changed %v", err)
	}
}
