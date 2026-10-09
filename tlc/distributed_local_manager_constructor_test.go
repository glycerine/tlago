package tlc

import (
	"math"
	"testing"
)

// NonDistributedFPSetManager stores its reference without dereferencing it.
// No original Java method directly covers this constructor's failure timing.
func TestNonDistributedFingerprintMissingReferenceConstructor(t *testing.T) {
	captureFailoverToolIO(t, ToolIOTool)
	trace := &TLCTrace{}
	manager := NewNonDistributedFPSetManager(nil, "absent", trace)
	if manager.NumOfServers() != 1 || manager.NumOfAliveServers() != 1 || manager.GetFPSetIndex(7) != 0 || manager.GetMask() != math.MaxInt64 || manager.GetHostName() != "absent" || manager.Trace != trace {
		t.Fatal("local constructor changed reference-independent queries or captured trace")
	}
	for _, call := range []func([]*LongVec, ...*DistributedExecutor) []*BitVector{manager.PutBlock, manager.ContainsBlock} {
		if result := call([]*LongVec{}); result == nil || len(result) != 0 {
			t.Fatalf("empty batch with missing reference = %v, want non-null empty result", result)
		}
	}
	requireCoordinatorNullFailure(t, func() error { manager.Put(7); return nil })

	// A missing reference is transferred without trying to publish or connect
	// to a fabricated endpoint. The coordinator-only recovery trace is omitted.
	payload, err := encodeDistributedManager(manager, func(DistributedFingerprintEndpoint) (DistributedEndpointReference, error) {
		t.Fatal("missing local reference was published")
		return DistributedEndpointReference{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	worker, err := decodeDistributedManager(payload, func(DistributedEndpointReference) (DistributedFingerprintEndpoint, error) {
		t.Fatal("missing local reference was resolved")
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !worker.NonDistributed || worker.Trace != nil || worker.GetHostName() != "absent" || worker.NumOfAliveServers() != 1 || worker.GetMask() != math.MaxInt64 {
		t.Fatal("worker receipt changed local manager metadata or retained its recovery trace")
	}
	requireCoordinatorNullFailure(t, func() error { worker.Contains(7); return nil })
	// The ordinary distributed constructor evaluates fpSet.toString(), so its
	// missing-reference failure must remain immediate.
	requireCoordinatorNullFailure(t, func() error { NewDistributedFPSetManager(nil); return nil })
	if messages := ToolIOGetAllMessages(); len(messages) != 0 {
		t.Fatalf("missing local reference entered checked-I/O/failover reporting: %q", messages)
	}
}
