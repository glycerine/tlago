package tlc

import "testing"

// Source base TLCServer rejects registration before accessing its manager.
// Only DistributedFPSetTLCServer registers stores and releases startup waiters.
// Upstream has no direct method for this distinction.
func TestDistributedBaseCoordinatorRejectsFingerprintRegistration(t *testing.T) {
	for _, kind := range []string{"local", "dynamic", "absent"} {
		t.Run(kind, func(t *testing.T) {
			captureFailoverToolIO(t, ToolIOTool)
			var manager *DistributedFPSetManager
			switch kind {
			case "local":
				manager = NewNonDistributedFPSetManager(NewMemFPSet(), "local", NewTLCTrace())
			case "dynamic":
				manager = NewDynamicDistributedFPSetManager(1)
			}
			server := &TLCServer{FPSetManager: manager}
			err := invokeDistributedServerOperation(func() error { return NewLocalServerEndpoint(server).RegisterFPSet(nil, "extra-store") })
			if _, ok := err.(*UnsupportedOperationException); !ok || err.Error() != "Not applicable for non-distributed TLCServer" {
				t.Fatalf("base coordinator registration failure: %T/%v", err, err)
			}
			if manager != nil && manager.NumOfServers() != map[string]int{"local": 1, "dynamic": 0}[kind] {
				t.Fatal("rejected registration modified manager")
			}
			if len(ToolIOGetAllMessages()) != 0 {
				t.Fatal("rejected registration emitted acceptance")
			}
		})
	}
}

func TestCoordinatorRPCBaseRejectsFingerprintRegistration(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	manager := NewDynamicDistributedFPSetManager(1)
	server := &TLCServer{FPSetManager: manager, InternTable: NewInternTable(16)}
	_, coordinator := startCoordinatorRPC(t, NewLocalServerEndpoint(server))
	// Registration transports a reference without probing the fingerprint host.
	err := coordinator.RegisterFPSetReference(DistributedEndpointReference{Address: "127.0.0.1:1", Object: "primary"}, "extra-store")
	failure, ok := err.(*DistributedOperationError)
	if !ok || failure.Class != "java.lang.UnsupportedOperationException" || failure.Error() != "Not applicable for non-distributed TLCServer" || failure.FingerprintRejected || failure.Remote || failure.IO || failure.Null {
		t.Fatalf("base registration lost boundary failure: %T/%v", err, err)
	}
	if manager.NumOfServers() != 0 || len(ToolIOGetAllMessages()) != 0 {
		t.Fatal("base registration altered manager or emitted acceptance")
	}
	if manager, err := coordinator.GetFPSetManager(); err != nil || manager.NumOfServers() != 0 {
		t.Fatalf("rejected registration damaged coordinator endpoint: %v", err)
	}
}
