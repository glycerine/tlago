package tlc

import (
	"reflect"
	"testing"
)

// DistributedFPSetTLCServer.registerFPSet registers the reference and then
// counts down its latch. DynamicFPSetManager.register does not probe the store.
// No enabled upstream method directly covers this native registration boundary.
func TestCoordinatorRPCFingerprintRegistrationDoesNotProbeEndpoint(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	recorder := &MemoryRecorder{}
	AddMessageRecorder(recorder)
	t.Cleanup(func() { RemoveMessageRecorder(recorder) })
	failedHost, fp := startFingerprintRPC(t, NewLocalFingerprintEndpoint(NewMemFPSet()))
	if err := failedHost.Close(); err != nil {
		t.Fatal(err)
	}
	manager := NewDynamicDistributedFPSetManager(1)
	registration := &distributedFPRegistration{expected: 1, remaining: 1, done: make(chan struct{})}
	server := &TLCServer{FPSetManager: manager, InternTable: NewInternTable(16), fpRegistration: registration}
	_, coordinator := startCoordinatorRPC(t, NewLocalServerEndpoint(server))
	reference := DistributedEndpointReference{Address: fp.Address, Object: fp.Object}
	if err := coordinator.RegisterFPSetReference(reference, "stopped-store"); err != nil {
		t.Fatalf("registration incorrectly required a reachable store: %v", err)
	}
	if manager.NumOfServers() != 1 || manager.NumOfAliveServers() != 1 || manager.entry(0).hostname != "stopped-store" {
		t.Fatal("registration changed source availability metadata or lost hostname")
	}
	select {
	case <-registration.done:
	default:
		t.Fatal("accepted reference did not release registration latch")
	}
	accepted := recorder.Records(ECTLCDistributedServerFPSetRegistered)
	if len(accepted) != 1 || !reflect.DeepEqual(accepted[0].Params, []string{"1", "1"}) {
		t.Fatalf("accepted registration diagnostic = %v", accepted)
	}
	// Even an unreachable extra reference must reach the manager's source
	// capacity rejection, not fail earlier while dialing its address.
	err := coordinator.RegisterFPSetReference(reference, "extra-store")
	if !isDistributedFPRegistrationRejected(err) || err.Error() != "Limit for FPset servers reached (1). Cannot handle additional servers" {
		t.Fatalf("extra registration lost source capacity rejection: %T/%v", err, err)
	}
	if manager.NumOfServers() != 1 || manager.NumOfAliveServers() != 1 || len(recorder.Records(ECTLCDistributedServerFPSetRegistered)) != 1 || recorder.Recorded(ECGeneral) {
		t.Fatal("rejected reference changed registration state or emitted acceptance/error")
	}
}

func TestCoordinatorRPCFingerprintRegistrationRejectsIncompleteReference(t *testing.T) {
	for _, reference := range []DistributedEndpointReference{{Address: "127.0.0.1:1"}, {Object: "primary"}, {}} {
		manager := NewDynamicDistributedFPSetManager(1)
		_, coordinator := startCoordinatorRPC(t, NewLocalServerEndpoint(&TLCServer{FPSetManager: manager, InternTable: NewInternTable(16)}))
		if err := coordinator.RegisterFPSetReference(reference, "invalid-store"); err == nil {
			t.Fatal("incomplete native endpoint reference registered")
		}
		if manager.NumOfServers() != 0 {
			t.Fatal("invalid native reference consumed a registration slot")
		}
	}
}
